package skillintegrity

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/skillspec"
	"github.com/relux-works/curator/internal/whitelist"
)

func fixture(t *testing.T) (string, Identity) {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, root, "SKILL.md", []byte("---\nname: fixture-skill\ndescription: Fixture skill.\n---\nFixture context.\n"), 0o644)
	writeFixture(t, root, "references/guide.md", []byte("Fixture reference."), 0o644)
	return root, Identity{Name: "fixture-skill", Source: "https://example.org/skill.git", Commit: strings.Repeat("a", 40)}
}
func writeFixture(t *testing.T, root, path string, data []byte, mode os.FileMode) {
	t.Helper()
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, data, mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(full, mode); err != nil {
		t.Fatal(err)
	}
	// Windows permission bits do not establish the owner/DACL boundary.
	// These fixture writes are new test payloads, never pre-existing inputs.
	if runtime.GOOS == "windows" {
		if err := privatedir.ProtectTree(root); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBudgetPauseResumeExcludesNonVerificationTime(t *testing.T) {
	started := time.Now()
	pausedAt := started.Add(VerificationTimeout - time.Second)
	resumedAt := pausedAt.Add(2 * VerificationTimeout)
	budget := &Budget{started: started}
	budget.pauseAt(pausedAt)
	if got, want := budget.elapsedAt(resumedAt), VerificationTimeout-time.Second; got != want {
		t.Fatalf("elapsed while paused = %s, want %s", got, want)
	}
	budget.resumeAt(resumedAt)
	if got, want := budget.elapsedAt(resumedAt.Add(2*time.Second)), VerificationTimeout+time.Second; got != want {
		t.Fatalf("elapsed after resuming = %s, want %s", got, want)
	}
}

// Native canonical/legacy/runtime fallback and manifestless skills all receive
// complete evidence, without writing a synthesized declaration into the source.
// Canonical encoding binds the returned evidence and survives the strict reader.
func TestBuildAndVerifyEvidence(t *testing.T) {
	for _, path := range []string{"", "agent-skill.json", "csk-skill.json", "agents/runtime.json"} {
		t.Run(path, func(t *testing.T) {
			root, id := fixture(t)
			if path != "" {
				data := []byte(syntheticDeclaration)
				if path == "agents/runtime.json" {
					data = []byte(`{"commands":{}}`)
				}
				writeFixture(t, root, path, data, 0o644)
			}
			e, err := Build(context.Background(), root, id, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(e.Manifest.Files) != 2+btoi(path != "") || e.Manifest.DeclarationPath != path {
				t.Fatalf("unexpected roster/declaration: %+v", e.Manifest)
			}
			if path == "" {
				if e.Manifest.ManifestKind != "synthesized" || !bytes.Equal(e.Manifest.Declaration, []byte(syntheticDeclaration)) {
					t.Fatal("manifestless declaration gained authority")
				}
				if _, err := os.Lstat(filepath.Join(root, "agent-skill.json")); !os.IsNotExist(err) {
					t.Fatal("build wrote a native manifest")
				}
			}
			raw, err := protocoljson.MarshalCanonical(e.Object())
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeEntry(raw, id)
			if err != nil {
				t.Fatal(err)
			}
			if err := Verify(context.Background(), root, id, decoded, nil); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProfileSelectionFiltersBeforeCaptureAndKeepsBuildInputs(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "SKILL.md", []byte("---\nname: fixture-skill\ndescription: Fixture skill.\n---\nFixture context.\n"), 0o644)
	writeFixture(t, root, "agent-skill.json", []byte(`{"schema_version":8,"capabilities":{},"build_roots":["tools/runner"],"commands":{"runner":{"type":"build","driver":"go-v1","source_dir":"tools/runner/cmd/runner","modules":["modules/support"]}}}`), 0o644)
	writeFixture(t, root, "references/guide.md", []byte("Guide.\n"), 0o644)
	writeFixture(t, root, "data/public.json", []byte("{}\n"), 0o644)
	writeFixture(t, root, "tools/runner/go.mod", []byte("module example.invalid/runner\n\ngo 1.24\n"), 0o644)
	writeFixture(t, root, "tools/runner/cmd/runner/main.go", []byte("package main\nfunc main() {}\n"), 0o644)
	writeFixture(t, root, "tools/runner/vendor/modules.txt", []byte("# example.invalid/support v0.1.0\n"), 0o644)
	writeFixture(t, root, "tools/runner/vendor/example.invalid/support/support.go", []byte("package support\n"), 0o644)
	writeFixture(t, root, "modules/support/go.mod", []byte("module example.invalid/support\n\ngo 1.24\n"), 0o644)
	writeFixture(t, root, "modules/support/support.go", []byte("package support\n"), 0o644)
	large := filepath.Join(root, ".excluded-cache", "over-limit.bin")
	if err := os.MkdirAll(filepath.Dir(large), 0o700); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(large)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(MaxBytes + 1); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	identity := Identity{Name: "fixture-skill", Source: "https://example.org/fixture-skill.git", Commit: strings.Repeat("a", 40)}
	selection := &whitelist.FileSelection{Include: []string{"data/public.json"}, Exclude: []string{".excluded-cache/**"}}
	entry, err := BuildWithProfileSelection(context.Background(), root, identity, selection, nil)
	if err != nil {
		t.Fatalf("filtered capture failed: %v", err)
	}
	wantContext := []string{"SKILL.md", "data/public.json", "references/guide.md"}
	if !slices.Equal(entry.Manifest.ContextFiles, wantContext) {
		t.Fatalf("context files = %v, want %v", entry.Manifest.ContextFiles, wantContext)
	}
	if entry.Manifest.FileSelection == nil || !slices.Equal(entry.Manifest.FileSelection.Include, selection.Include) || !slices.Equal(entry.Manifest.FileSelection.Exclude, selection.Exclude) {
		t.Fatalf("selection rules were not recorded: %+v", entry.Manifest.FileSelection)
	}
	for _, required := range []string{"agent-skill.json", "tools/runner/go.mod", "tools/runner/cmd/runner/main.go", "tools/runner/vendor/modules.txt", "modules/support/go.mod", "modules/support/support.go"} {
		if !hasEvidenceFile(entry.Manifest.Files, required) {
			t.Errorf("declared build input %q was not retained in lock evidence", required)
		}
	}
	if hasEvidenceFile(entry.Manifest.Files, ".excluded-cache/over-limit.bin") {
		t.Fatal("excluded payload entered integrity capture")
	}
	if err := Verify(context.Background(), root, identity, entry, nil); err != nil {
		t.Fatalf("filtered evidence did not verify: %v", err)
	}
	if err := os.Truncate(large, MaxBytes+2); err != nil {
		t.Fatal(err)
	}
	if err := Verify(context.Background(), root, identity, entry, nil); err != nil {
		t.Fatalf("excluded file drift affected filtered evidence: %v", err)
	}
	writeFixture(t, root, "data/public.json", []byte("{\"updated\":true}\n"), 0o644)
	if err := Verify(context.Background(), root, identity, entry, nil); !IsCode(err, "skill_integrity_mismatch") {
		t.Fatalf("selected file drift did not refuse: %v", err)
	}

	if _, err := Build(context.Background(), root, identity, nil); !IsCode(err, "skill_integrity_manifest_invalid") || !strings.Contains(err.Error(), "capture budget exceeded") {
		t.Fatalf("unfiltered capture did not refuse at the existing limit: %v", err)
	}
}

func TestProfileSelectionRefusesRequiredInputsAndProtectedFiles(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "SKILL.md", []byte("---\nname: fixture-skill\ndescription: Fixture skill.\n---\nFixture context.\n"), 0o644)
	writeFixture(t, root, "agent-skill.json", []byte(`{"schema_version":8,"capabilities":{},"build_roots":["tools/runner"],"commands":{"runner":{"type":"build","driver":"go-v1","source_dir":"tools/runner/cmd/runner","modules":["modules/support"]}}}`), 0o644)
	writeFixture(t, root, "tools/runner/go.mod", []byte("module example.invalid/runner\n\ngo 1.24\n"), 0o644)
	writeFixture(t, root, "tools/runner/cmd/runner/main.go", []byte("package main\nfunc main() {}\n"), 0o644)
	writeFixture(t, root, "modules/support/go.mod", []byte("module example.invalid/support\n\ngo 1.24\n"), 0o644)
	writeFixture(t, root, "modules/support/support.go", []byte("package support\n"), 0o644)
	identity := Identity{Name: "fixture-skill", Source: "https://example.org/fixture-skill.git", Commit: strings.Repeat("a", 40)}
	for _, pattern := range []string{"tools/runner/cmd/runner/main.go", "tools/runner/cmd/runner/**", "modules/support/**"} {
		t.Run(pattern, func(t *testing.T) {
			_, err := BuildWithProfileSelection(context.Background(), root, identity, &whitelist.FileSelection{Exclude: []string{pattern}}, nil)
			if !IsCode(err, "profile_skill_files_build_source_excluded") {
				t.Fatalf("required build input exclusion did not refuse: %v", err)
			}
		})
	}
	for _, file := range []string{"SKILL.md", "agent-skill.json"} {
		t.Run(file, func(t *testing.T) {
			_, err := BuildWithProfileSelection(context.Background(), root, identity, &whitelist.FileSelection{Exclude: []string{file}}, nil)
			if !IsCode(err, "profile_skill_files_invalid") {
				t.Fatalf("protected file exclusion did not refuse: %v", err)
			}
		})
	}
}

func TestSelectionIntegrityEntryRoundTripsSingleRuleList(t *testing.T) {
	for _, selection := range []*whitelist.FileSelection{
		{Include: []string{"extras/data.json"}},
		{Exclude: []string{"references/private/**"}},
	} {
		root, id := fixture(t)
		writeFixture(t, root, "extras/data.json", []byte("{}\n"), 0o644)
		entry, err := BuildWithProfileSelection(context.Background(), root, id, selection, nil)
		if err != nil {
			t.Fatalf("build with selection %+v: %v", selection, err)
		}
		raw, err := protocoljson.MarshalCanonical(entry.Object())
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeEntry(raw, id)
		if err != nil {
			t.Fatalf("decode selection %+v: %v", selection, err)
		}
		if !equalFileSelection(decoded.Manifest.FileSelection, selection) {
			t.Fatalf("decoded selection = %+v, want %+v", decoded.Manifest.FileSelection, selection)
		}
	}
}

func hasEvidenceFile(files []File, path string) bool {
	for _, file := range files {
		if file.Path == path {
			return true
		}
	}
	return false
}

func parseTestCapabilities(t *testing.T, raw string) capabilities.Manifest {
	t.Helper()
	var object map[string]any
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		t.Fatal(err)
	}
	parsed, err := capabilities.Parse(object)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

// TestProfileCapabilitiesSynthesizeStableEvidence proves explicit all-none
// capabilities keep today's bytes and key/list order cannot change lock evidence.
func TestProfileCapabilitiesSynthesizeStableEvidence(t *testing.T) {
	root, identity := fixture(t)
	without, err := Build(context.Background(), root, identity, nil)
	if err != nil {
		t.Fatal(err)
	}
	allNone := parseTestCapabilities(t, `{"network":"none","filesystem":"none","exec":"none","secrets":"none","env_read":"none"}`)
	withAllNone, err := BuildWithProfileCapabilities(context.Background(), root, identity, &allNone, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(withAllNone.Manifest.Declaration, []byte(syntheticDeclaration)) || withAllNone.ManifestSHA256 != without.ManifestSHA256 {
		t.Fatal("explicit all-none declaration changed legacy synthetic evidence")
	}
	emptyObject := parseTestCapabilities(t, `{}`)
	withDefaultFilesystem, err := BuildWithProfileCapabilities(context.Background(), root, identity, &emptyObject, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(withDefaultFilesystem.Manifest.Declaration, []byte(syntheticDeclaration)) {
		t.Fatal("empty capability object lost its default repository filesystem scope")
	}
	var defaultDeclaration map[string]any
	if err := json.Unmarshal(withDefaultFilesystem.Manifest.Declaration, &defaultDeclaration); err != nil {
		t.Fatal(err)
	}
	defaultCaps, ok := defaultDeclaration["capabilities"].(map[string]any)
	if !ok || defaultCaps["filesystem"] != "repo" {
		t.Fatalf("empty object did not preserve the parser's repo default: %#v", defaultDeclaration)
	}

	firstCaps := parseTestCapabilities(t, `{"network":["api.example.net","docs.example.net"],"filesystem":"repo"}`)
	secondCaps := parseTestCapabilities(t, `{"filesystem":"repo","network":["docs.example.net","api.example.net"]}`)
	first, err := BuildWithProfileCapabilities(context.Background(), root, identity, &firstCaps, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildWithProfileCapabilities(context.Background(), root, identity, &secondCaps, nil)
	if err != nil {
		t.Fatal(err)
	}
	firstLock, err := protocoljson.MarshalCanonical(first.Object())
	if err != nil {
		t.Fatal(err)
	}
	secondLock, err := protocoljson.MarshalCanonical(second.Object())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstLock, secondLock) || !bytes.Equal(first.Manifest.Declaration, second.Manifest.Declaration) {
		t.Fatal("semantic capability order changed canonical lock evidence")
	}
	upperHost := parseTestCapabilities(t, `{"network":["API.EXAMPLE.NET"]}`)
	lowerHost := parseTestCapabilities(t, `{"network":["api.example.net"]}`)
	upperEntry, err := BuildWithProfileCapabilities(context.Background(), root, identity, &upperHost, nil)
	if err != nil {
		t.Fatal(err)
	}
	lowerEntry, err := BuildWithProfileCapabilities(context.Background(), root, identity, &lowerHost, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(upperEntry.Manifest.Declaration, lowerEntry.Manifest.Declaration) {
		t.Fatal("case-equivalent audit host patterns changed canonical evidence")
	}
	if err := Verify(context.Background(), root, identity, first, nil); err != nil {
		t.Fatalf("dynamic synthesized evidence did not verify: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(root, skillspec.CanonicalManifestName)); !os.IsNotExist(err) {
		t.Fatal("profile capabilities were written into the skill source")
	}
}

// TestSynthesizedEvidenceShapeStaysClosed proves recomputing a forged entry
// digest cannot authorize commands, runtime roots, extra fields or bad caps.
func TestSynthesizedEvidenceShapeStaysClosed(t *testing.T) {
	root, identity := fixture(t)
	caps := parseTestCapabilities(t, `{"network":["api.example.net"]}`)
	base, err := BuildWithProfileCapabilities(context.Background(), root, identity, &caps, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range []string{
		`{"schema_version":3,"runtime_roots":["runtime"],"commands":{},"capabilities":{"network":["api.example.net"]}}`,
		`{"schema_version":3,"runtime_roots":[],"commands":{"run":"tool"},"capabilities":{"network":["api.example.net"]}}`,
		`{"schema_version":3,"runtime_roots":[],"commands":{},"capabilities":{"network":["api.example.net"]},"extra":true}`,
		`{"schema_version":3,"runtime_roots":[],"commands":{},"capabilities":null}`,
	} {
		t.Run(declaration, func(t *testing.T) {
			forged := base
			forged.Manifest.Declaration = json.RawMessage(declaration)
			raw, err := protocoljson.MarshalCanonical(forged.Manifest.Object())
			if err != nil {
				t.Fatal(err)
			}
			forged.ManifestSHA256 = digest(raw)
			if err := forged.Validate(identity); !IsCode(err, "skill_integrity_manifest_invalid") {
				t.Fatalf("forged declaration with recomputed digest was accepted: %v", err)
			}
		})
	}
}

// TestProfileCapabilitiesCannotReplaceNativeEvidence proves the synthesis
// helper refuses a valid native declaration instead of overwriting it.
func TestProfileCapabilitiesCannotReplaceNativeEvidence(t *testing.T) {
	root, identity := fixture(t)
	writeFixture(t, root, skillspec.CanonicalManifestName, []byte(`{"schema_version":3,"capabilities":{"network":["api.example.net"]}}`), 0o644)
	caps := parseTestCapabilities(t, `{"network":["api.example.net"]}`)
	if _, err := BuildWithProfileCapabilities(context.Background(), root, identity, &caps, nil); !IsCode(err, "profile_skill_capability_native_manifest") {
		t.Fatalf("profile declaration replaced native evidence: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(root, skillspec.CanonicalManifestName))
	if err != nil || string(after) != `{"schema_version":3,"capabilities":{"network":["api.example.net"]}}` {
		t.Fatalf("native declaration changed: %s, %v", after, err)
	}
}
func btoi(value bool) int {
	if value {
		return 1
	}
	return 0
}

// Native runtime commands remain native and their exact declaration bytes are
// bound beside reference assets and executable metadata. Matching dual native
// manifests retain the canonical precedence rather than becoming synthetic.
func TestNativeDeclarationsPreserveAuthority(t *testing.T) {
	root, id := fixture(t)
	declaration := []byte(`{"commands":{"hello":"scripts/hello.sh"}}`)
	writeFixture(t, root, "agents/runtime.json", declaration, 0o644)
	writeFixture(t, root, "scripts/hello.sh", []byte("#!/bin/sh\nexit 0\n"), 0o755)
	e, err := Build(context.Background(), root, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	if e.Manifest.ManifestKind != "native" || e.Manifest.DeclarationSHA256 != digest(declaration) || len(e.Manifest.Files) != 4 {
		t.Fatal("lost native declaration or complete roster")
	}
	spec, err := skillspec.Load(root)
	if err != nil || spec.Commands["hello"].UnixPath != "scripts/hello.sh" {
		t.Fatalf("native commands changed: %+v,%v", spec, err)
	}
	after, _ := os.ReadFile(filepath.Join(root, "agents/runtime.json"))
	if !bytes.Equal(after, declaration) {
		t.Fatal("build changed native declaration")
	}
	root, id = fixture(t)
	writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0o644)
	writeFixture(t, root, "csk-skill.json", []byte(syntheticDeclaration), 0o644)
	e, err = Build(context.Background(), root, id, nil)
	if err != nil || e.Manifest.DeclarationPath != "agent-skill.json" {
		t.Fatalf("matching native manifests: %+v,%v", e, err)
	}
}

// A present unreadable manifest is unavailable evidence, never a synthesized
// declaration. Root execution cannot establish the Unix mode-denial witness.
func TestUnreadableNativeIsNotAbsence(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("fixture requires Unix permission denial without root bypass")
	}
	root, id := fixture(t)
	path := filepath.Join(root, "agent-skill.json")
	writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0o000)
	defer func() { _ = os.Chmod(path, 0o644) }()
	e, err := Build(context.Background(), root, id, nil)
	if !IsCode(err, "skill_integrity_unavailable") || e.Manifest.ManifestKind != "" {
		t.Fatalf("unreadable native became evidence: %+v,%v", e, err)
	}
}

// Roster additions/removals, content changes, execute-bit drift and an added
// native declaration each refuse against previously valid evidence.
func TestEvidenceRejectsEveryDriftDirection(t *testing.T) {
	for _, change := range []string{"add", "remove", "edit", "execute", "native-substitution"} {
		t.Run(change, func(t *testing.T) {
			root, id := fixture(t)
			e, err := Build(context.Background(), root, id, nil)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "add":
				writeFixture(t, root, "references/extra.md", []byte("extra"), 0o644)
			case "remove":
				if err := os.Remove(filepath.Join(root, "references/guide.md")); err != nil {
					t.Fatal(err)
				}
			case "edit":
				writeFixture(t, root, "references/guide.md", []byte("changed"), 0o644)
			case "execute":
				if err := os.Chmod(filepath.Join(root, "references/guide.md"), 0o755); err != nil {
					t.Fatal(err)
				}
				observed, err := os.Stat(filepath.Join(root, "references/guide.md"))
				if err != nil {
					t.Fatal(err)
				}
				if observed.Mode().Perm()&0o111 == 0 {
					t.Skip("filesystem cannot materialize the executable-bit mutation")
				}
			case "native-substitution":
				writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0o644)
			}
			if err := Verify(context.Background(), root, id, e, nil); !IsCode(err, "skill_integrity_mismatch") {
				t.Fatalf("%s drift: %v", change, err)
			}
		})
	}
}

// Malformed/present metadata cannot become an absent synthetic declaration;
// canonical/legacy disagreement and every payload link boundary refuse.
func TestEvidenceStructuralRefusals(t *testing.T) {
	for _, kind := range []string{"bad-manifest", "dual-conflict", "symlink", "hardlink", "nested-git", "case-collision"} {
		t.Run(kind, func(t *testing.T) {
			root, id := fixture(t)
			want := "skill_integrity_boundary_invalid"
			switch kind {
			case "bad-manifest":
				writeFixture(t, root, "agent-skill.json", []byte(`{"schema_version":999}`), 0o644)
				want = "skill_integrity_manifest_invalid"
			case "dual-conflict":
				writeFixture(t, root, "agent-skill.json", []byte(syntheticDeclaration), 0o644)
				writeFixture(t, root, "csk-skill.json", []byte(`{"schema_version":1}`), 0o644)
				want = "skill_integrity_manifest_invalid"
			case "symlink":
				if err := os.Symlink("references/guide.md", filepath.Join(root, "linked.md")); err != nil {
					t.Fatal(err)
				}
			case "hardlink":
				if err := os.Link(filepath.Join(root, "references/guide.md"), filepath.Join(root, "linked.md")); err != nil {
					t.Fatal(err)
				}
			case "nested-git":
				writeFixture(t, root, "nested/.git/config", []byte("fixture"), 0o644)
			case "case-collision":
				writeFixture(t, root, "references/GUIDE.md", []byte("other"), 0o644)
				// A case-insensitive filesystem cannot materialize the forbidden
				// pair. Its independent evidence-parser control still runs below.
				entries, _ := os.ReadDir(filepath.Join(root, "references"))
				if len(entries) == 1 {
					t.Skip("filesystem cannot represent the colliding pair")
				}
			}
			if _, err := Build(context.Background(), root, id, nil); !IsCode(err, want) {
				t.Fatalf("%s: %v", kind, err)
			}
		})
	}
}

// Evidence JSON is closed, required and bound to the original identity/hash.
// Missing/nullable booleans and identity fields cannot acquire zero defaults.
func TestEvidenceJSONRefusals(t *testing.T) {
	root, id := fixture(t)
	e, err := Build(context.Background(), root, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(e)
	for _, kind := range []string{"unknown", "missing-executable", "null-executable", "null-source", "changed-identity", "changed-hash", "case-collision", "nonportable-path"} {
		t.Run(kind, func(t *testing.T) {
			var object map[string]any
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&object); err != nil {
				t.Fatal(err)
			}
			m := object["manifest"].(map[string]any)
			files := m["files"].([]any)
			first := files[0].(map[string]any)
			switch kind {
			case "nonportable-path":
				first["path"] = "notes:local.txt"
			case "unknown":
				object["extra"] = true
			case "missing-executable":
				delete(first, "executable")
			case "null-executable":
				first["executable"] = nil
			case "null-source":
				m["source"] = nil
			case "changed-identity":
				m["commit"] = strings.Repeat("b", 40)
			case "changed-hash":
				object["manifest_sha256"] = "sha256:" + strings.Repeat("0", 64)
			case "case-collision":
				second := files[1].(map[string]any)
				second["path"] = "skill.md"
			}
			if kind != "changed-hash" {
				canonical, err := protocoljson.MarshalCanonical(m)
				if err != nil {
					t.Fatal(err)
				}
				object["manifest_sha256"] = digest(canonical)
			}
			if kind == "null-executable" {
				object["manifest_sha256"] = e.ManifestSHA256
			}
			changed, _ := json.Marshal(object)
			if _, err := DecodeEntry(changed, id); !IsCode(err, "skill_integrity_manifest_invalid") {
				t.Fatalf("%s: %v", kind, err)
			}
		})
	}
}

// A real empty file supplies the valid zero-size control. Alternate numeric
// spellings must refuse even when coercion would reproduce its valid hash.
func TestEvidenceSizeTokenGrammar(t *testing.T) {
	root, id := fixture(t)
	writeFixture(t, root, "references/guide.md", nil, 0o644)
	e, err := Build(context.Background(), root, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeEntry(raw, id); err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"-0", "0.0", "0e0"} {
		changed := bytes.Replace(raw, []byte(`"size":0`), []byte(`"size":`+token), 1)
		if bytes.Equal(raw, changed) {
			t.Fatal("control plant did not apply")
		}
		if _, err := DecodeEntry(changed, id); !IsCode(err, "skill_integrity_manifest_invalid") {
			t.Fatalf("size %s: %v", token, err)
		}
	}
}

// Shared closure budgets cannot reset at each skill. Cancellation does not
// synthesize evidence, and an observed mismatch survives later unavailability.
func TestEvidenceBudgetsAndInterruptedMismatch(t *testing.T) {
	root, id := fixture(t)
	budget := NewBudget()
	budget.files = MaxFiles
	if _, err := Build(context.Background(), root, id, budget); !IsCode(err, "skill_integrity_manifest_invalid") {
		t.Fatalf("object budget: %v", err)
	}
	budget = NewBudget()
	budget.bytes = MaxBytes
	if _, err := Build(context.Background(), root, id, budget); !IsCode(err, "skill_integrity_manifest_invalid") {
		t.Fatalf("byte budget: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Build(ctx, root, id, nil); !IsCode(err, "skill_integrity_unavailable") {
		t.Fatalf("cancellation: %v", err)
	}
	e, err := Build(context.Background(), root, id, nil)
	if err != nil {
		t.Fatal(err)
	}
	writeFixture(t, root, "SKILL.md", []byte("changed first file"), 0o644)
	// This fixture context cancels at the fifth deadline observation: after
	// SKILL.md has been captured, before the next directory is inspected.
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	controlled := &cancelAfterChecks{Context: ctx2, cancel: cancel2, after: 5}
	err = Verify(controlled, root, id, e, nil)
	if !IsCode(err, "skill_integrity_mismatch") || !IsCode(err, "skill_integrity_unavailable") {
		t.Fatalf("lost prior mismatch/unavailability: %v", err)
	}
}

type cancelAfterChecks struct {
	context.Context
	cancel       context.CancelFunc
	calls, after int
}

func (c *cancelAfterChecks) Err() error {
	c.calls++
	if c.calls == c.after {
		c.cancel()
	}
	return c.Context.Err()
}

// The resolver's admitted legacy root form permits bare SKILL.md. Applying a
// directory selector invokes the existing stricter metadata contract instead.
func TestLegacyRootMetadataCompatibility(t *testing.T) {
	root, id := fixture(t)
	writeFixture(t, root, "SKILL.md", []byte("# Existing legacy context\n"), 0o644)
	if _, err := Build(context.Background(), root, id, nil); err != nil {
		t.Fatalf("legacy valid form rejected: %v", err)
	}
	id.Directory = "skills/fixture-skill"
	if _, err := Build(context.Background(), root, id, nil); !IsCode(err, "skill_integrity_manifest_invalid") {
		t.Fatalf("selected invalid metadata: %v", err)
	}
	writeFixture(t, root, "SKILL.md", []byte("---\nname: fixture-skill\ndescription: Valid selected context.\n---\nContext.\n"), 0o644)
	if _, err := Build(context.Background(), root, id, nil); err != nil {
		t.Fatalf("selected valid form rejected: %v", err)
	}
}
