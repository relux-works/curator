package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Umbrella provider trust-root vectors (environments §11, §13). Every
// case of vectors/umbrella-provider-resolution.json runs through the
// production resolveProvider entry point — the function backing both
// dispatch and the env status posture rows — for BOTH rollout
// revisions: the case's directories and executables are materialized in
// a temp tree, the case's PATH is searched, and the resolved path, the
// closed diagnostic, and the named roots are compared. A coverage row
// is driven only by this production entry point, never by a helper.

// umbrellaVectorCase is one umbrella-provider-resolution.json case.
type umbrellaVectorCase struct {
	Name                string   `json:"name"`
	Subcommand          string   `json:"subcommand"`
	ExecutableName      string   `json:"executable_name"`
	InstallDir          string   `json:"install_dir"`
	ProviderDirectories []string `json:"provider_directories"`
	PathEntries         []string `json:"path_entries"`
	PublishedDirs       []string `json:"published_dirs"`
	ManagedDirs         []string `json:"managed_dirs"`
	UnreadableDirs      []string `json:"unreadable_dirs"`
	Present             []struct {
		Path       string `json:"path"`
		Executable bool   `json:"executable"`
	} `json:"present"`
	RevisionA umbrellaVectorExpect `json:"revision_a"`
	RevisionB umbrellaVectorExpect `json:"revision_b"`
}

// umbrellaVectorExpect is one revision's expected outcome. Pointers
// distinguish an expected null from an absent member: silent successes
// carry no trust_roots_consulted member at all.
type umbrellaVectorExpect struct {
	Resolved               *string  `json:"resolved"`
	Diagnostic             *string  `json:"diagnostic"`
	TrustRootsConsulted    []string `json:"trust_roots_consulted"`
	TrustRootsPresent      bool
	UntrustedPath          *string `json:"untrusted_path"`
	MigrationHintDirectory *string `json:"migration_hint_directory"`
	UnreadableDirectory    *string `json:"unreadable_directory"`
}

// TestUmbrellaProviderResolutionVectors executes the §11 vector family
// for both revisions. The committed CI pin serves the family, so a root
// that publishes no such vector — or publishes it with no cases — fails
// instead of skipping or passing empty.
func TestUmbrellaProviderResolutionVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", "umbrella-provider-resolution.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var family struct {
		Cases []umbrellaVectorCase `json:"cases"`
	}
	raw := map[string]any{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(payload, &family); err != nil {
		t.Fatal(err)
	}
	if len(family.Cases) == 0 {
		t.Fatalf("vectors/umbrella-provider-resolution.json publishes no cases")
	}
	markPresentRoots(t, raw, family.Cases)
	for _, tc := range family.Cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			sandbox := materializeUmbrellaCase(t, tc)
			checkUmbrellaRevision(t, sandbox, tc, providerRevisionA, tc.RevisionA)
			checkUmbrellaRevision(t, sandbox, tc, providerRevisionB, tc.RevisionB)
		})
	}
}

// markPresentRoots records which expectations carry an explicit
// trust_roots_consulted member, since its absence on silent successes
// is significant: the member is compared only when present.
func markPresentRoots(t *testing.T, raw map[string]any, cases []umbrellaVectorCase) {
	t.Helper()
	rawCases, _ := raw["cases"].([]any)
	if len(rawCases) != len(cases) {
		t.Fatalf("vector case count %d does not match decoded %d", len(rawCases), len(cases))
	}
	for i, entry := range rawCases {
		obj, _ := entry.(map[string]any)
		if obj == nil {
			t.Fatalf("case %d is not an object", i)
		}
		for _, key := range []string{"revision_a", "revision_b"} {
			rev, _ := obj[key].(map[string]any)
			if rev == nil {
				t.Fatalf("case %q carries no %s", cases[i].Name, key)
			}
			_, present := rev["trust_roots_consulted"]
			if key == "revision_a" {
				cases[i].RevisionA.TrustRootsPresent = present
			} else {
				cases[i].RevisionB.TrustRootsPresent = present
			}
		}
	}
}

// umbrellaSandbox is one materialized vector case: the remapped trust
// roots, search domain, and refused directories.
type umbrellaSandbox struct {
	inputs providerInputs
	remap  func(string) string
}

// materializeUmbrellaCase builds the case's filesystem in a temp tree.
// Every absolute vector path is remapped below the sandbox root, so the
// lookup runs against real directories the test owns. Unreadable roots
// are left uncreated: a trust root that does not exist is still a read
// failure, never absence (§8.4). On Windows an executable file carries
// the .exe suffix the platform lookup requires, and a non-executable
// file keeps its bare name; the expected paths are translated the same
// way.
func materializeUmbrellaCase(t *testing.T, tc umbrellaVectorCase) umbrellaSandbox {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	sandbox := filepath.Join(tmp, "root")
	remap := func(vectorPath string) string {
		rel := strings.TrimPrefix(filepath.Clean(vectorPath), "/")
		return filepath.Join(sandbox, filepath.FromSlash(rel))
	}
	unreadable := map[string]bool{}
	for _, dir := range tc.UnreadableDirs {
		unreadable[remap(dir)] = true
	}
	mkdirs := func(dirs ...string) {
		for _, dir := range dirs {
			if unreadable[dir] {
				continue
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	installDir := remap(tc.InstallDir)
	var providerDirs []string
	for _, dir := range tc.ProviderDirectories {
		providerDirs = append(providerDirs, remap(dir))
	}
	var pathEntries []string
	for _, entry := range tc.PathEntries {
		pathEntries = append(pathEntries, remap(entry))
	}
	var published []labeledDir
	for _, dir := range tc.PublishedDirs {
		published = append(published, labeledDir{path: remap(dir), label: "a manager-published directory"})
	}
	var managed []labeledDir
	for _, dir := range tc.ManagedDirs {
		managed = append(managed, labeledDir{path: remap(dir), label: "a managed directory"})
	}
	mkdirs(installDir)
	mkdirs(providerDirs...)
	mkdirs(pathEntries...)
	for _, entry := range published {
		mkdirs(entry.path)
	}
	for _, entry := range managed {
		mkdirs(entry.path)
	}
	for _, file := range tc.Present {
		full := remap(file.Path)
		mkdirs(filepath.Dir(full))
		name := full
		if runtime.GOOS == "windows" && file.Executable {
			name += ".exe"
		}
		mode := os.FileMode(0o644)
		if file.Executable {
			mode = 0o755
		}
		if err := os.WriteFile(name, []byte("#!/bin/sh\nexit 0\n"), mode); err != nil {
			t.Fatal(err)
		}
	}
	return umbrellaSandbox{
		inputs: providerInputs{
			installDir:    installDir,
			providerDirs:  providerDirs,
			publishedDirs: published,
			managedRoots:  managed,
			pathEntries:   pathEntries,
			platform:      runtime.GOOS,
			pathExt:       append([]string{}, defaultWindowsPathExt...),
		},
		remap: func(vectorPath string) string {
			out := remap(vectorPath)
			if runtime.GOOS == "windows" && isVectorExecutableFile(tc, vectorPath) {
				out += ".exe"
			}
			return out
		},
	}
}

// isVectorExecutableFile reports whether the vector's present list
// carries path as an executable file, for the Windows .exe translation
// of expected file paths. Directories never translate.
func isVectorExecutableFile(tc umbrellaVectorCase, vectorPath string) bool {
	for _, file := range tc.Present {
		if file.Path == vectorPath {
			return file.Executable
		}
	}
	return false
}

// checkUmbrellaRevision runs one revision of one case through the
// production entry point and compares the resolved path, the closed
// diagnostic, and every named root and path.
func checkUmbrellaRevision(t *testing.T, sandbox umbrellaSandbox, tc umbrellaVectorCase, revision providerRevision, want umbrellaVectorExpect) {
	t.Helper()
	in := sandbox.inputs
	in.revision = revision
	label := "revision A"
	if revision == providerRevisionB {
		label = "revision B"
	}
	outcome := resolveProvider(tc.Subcommand, in)
	if tc.ExecutableName != "curator-"+tc.Subcommand {
		t.Fatalf("%s: executable_name %q does not name the subcommand", label, tc.ExecutableName)
	}
	// The outcomes are disjoint: a lookup either selects (silently or
	// with the revision-A warning) or fails with exactly one diagnostic.
	selected := outcome.resolved != ""
	failed := outcome.diagnostic != "" && !outcome.warned
	if selected == failed {
		t.Fatalf("%s: resolved=%q warned=%v diagnostic=%q is not disjoint", label, outcome.resolved, outcome.warned, outcome.diagnostic)
	}
	if want.Resolved == nil {
		if selected {
			t.Fatalf("%s: resolved %q, want no selection", label, outcome.resolved)
		}
	} else {
		if sandbox.remap(*want.Resolved) != outcome.resolved {
			t.Fatalf("%s: resolved %q, want %q", label, outcome.resolved, sandbox.remap(*want.Resolved))
		}
	}
	if want.Diagnostic == nil {
		if outcome.warned || outcome.diagnostic != "" {
			t.Fatalf("%s: diagnostic=%q warned=%v, want a silent success", label, outcome.diagnostic, outcome.warned)
		}
	} else if *want.Diagnostic != outcome.diagnostic {
		t.Fatalf("%s: diagnostic %q, want %q", label, outcome.diagnostic, *want.Diagnostic)
	}
	if want.TrustRootsPresent {
		var wantRoots []string
		for _, root := range want.TrustRootsConsulted {
			wantRoots = append(wantRoots, sandbox.remap(root))
		}
		if strings.Join(wantRoots, "\x00") != strings.Join(outcome.roots, "\x00") {
			t.Fatalf("%s: roots consulted %q, want %q", label, outcome.roots, wantRoots)
		}
	}
	if want.UntrustedPath != nil && sandbox.remap(*want.UntrustedPath) != outcome.untrustedPath {
		t.Fatalf("%s: untrusted path %q, want %q", label, outcome.untrustedPath, sandbox.remap(*want.UntrustedPath))
	}
	if want.MigrationHintDirectory != nil && sandbox.remap(*want.MigrationHintDirectory) != outcome.hintDir {
		t.Fatalf("%s: migration hint %q, want %q", label, outcome.hintDir, sandbox.remap(*want.MigrationHintDirectory))
	}
	if want.UnreadableDirectory != nil && sandbox.remap(*want.UnreadableDirectory) != outcome.unreadableDir {
		t.Fatalf("%s: unreadable directory %q, want %q", label, outcome.unreadableDir, sandbox.remap(*want.UnreadableDirectory))
	}
	// A warning selects under revision A only; revision B never warns.
	if outcome.warned && revision != providerRevisionA {
		t.Fatalf("%s: a warning fired outside revision A", label)
	}
	if revision == providerRevisionA && outcome.diagnostic == providerDiagnosticOutsideRoots && !outcome.warned {
		t.Fatalf("revision A: the outside-roots diagnostic fired without a selection")
	}
}
