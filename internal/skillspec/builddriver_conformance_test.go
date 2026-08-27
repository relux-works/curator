package skillspec

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/relux-works/curator/internal/verr"
)

// buildDriverRejection is one authoritative rejection vector. Only the fields
// Curator has to honour are decoded; the portable expected values stay in the
// suite and are never copied into this repository.
type buildDriverRejection struct {
	Name     string `json:"name"`
	Boundary string `json:"boundary"`
	Expected struct {
		Result           string `json:"result"`
		Error            string `json:"error"`
		Reuse            bool   `json:"reuse"`
		ArtifactExecuted bool   `json:"artifact_executed"`
	} `json:"expected"`
}

// loadBuildDriverRejections returns the authoritative rejection vectors indexed
// by case name.
func loadBuildDriverRejections(t *testing.T) map[string]buildDriverRejection {
	t.Helper()
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	path := filepath.Join(root, "vectors", "build-drivers.json")
	payload, err := os.ReadFile(path) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no build-drivers vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		RejectionCases []buildDriverRejection `json:"rejection_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	indexed := make(map[string]buildDriverRejection, len(vectors.RejectionCases))
	for _, testCase := range vectors.RejectionCases {
		indexed[testCase.Name] = testCase
	}
	return indexed
}

// manifestRejection is the Curator-owned half of one mapping: the snapshot that
// reproduces the published condition and the stable validation path Curator
// reports for it.
type manifestRejection struct {
	// boundary is the authoritative boundary this Curator seam owns.
	boundary string
	// manifest is the schema v6 (or explicitly older) manifest document.
	manifest string
	// files are extra snapshot members beyond the shared build fixture.
	files map[string]string
	// prepare mutates the snapshot after the shared fixture is written.
	prepare func(t *testing.T, dir string)
	// wantPath is the stable Curator validation path for this rejection.
	wantPath string
}

const conformanceBuildCommand = `"tool":{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool"}`

func conformanceManifest(body string) string {
	return `{"schema_version":6,"capabilities":{},` + body + `}`
}

// linkConformancePath creates target -> source inside the snapshot. The special
// file helper is defined per platform so the "special file" vectors exercise a
// real device-like member wherever the host supports one.
func linkConformancePath(t *testing.T, source, target string) {
	t.Helper()
	if err := os.Symlink(source, target); err != nil {
		t.Skipf("this host cannot create the symbolic link the vector needs: %v", err)
	}
}

func writeConformanceTree(t *testing.T, extra map[string]string, prepare func(*testing.T, string), manifest string) string {
	t.Helper()
	files := buildFiles()
	for name, content := range extra {
		files[name] = content
	}
	dir := writeSkill(t, manifest, files)
	if prepare != nil {
		prepare(t, dir)
	}
	return dir
}

func manifestBoundaryRejections() map[string]manifestRejection {
	with := func(entry string) string {
		return conformanceManifest(`"build_roots":["build"],"commands":{"tool":` + entry + `}`)
	}
	return map[string]manifestRejection{
		"schema-5-build-command": {
			boundary: "manifest",
			manifest: `{"schema_version":5,"capabilities":{},"commands":{` + conformanceBuildCommand + `}}`,
			wantPath: "commands.tool",
		},
		"unknown-driver": {
			boundary: "manifest",
			manifest: with(`{"type":"build","driver":"custom-v1","source_dir":"build/cmd/tool"}`),
			wantPath: "commands.tool.driver",
		},
		"forbidden-args": {
			boundary: "manifest",
			manifest: with(`{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","args":["-x"]}`),
			wantPath: "commands.tool.args",
		},
		"forbidden-env": {
			boundary: "manifest",
			manifest: with(`{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","env":{"KEY":"value"}}`),
			wantPath: "commands.tool.env",
		},
		"forbidden-output": {
			boundary: "manifest",
			manifest: with(`{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","output":"bin/tool"}`),
			wantPath: "commands.tool.output",
		},
		"forbidden-toolchain": {
			boundary: "manifest",
			manifest: with(`{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","toolchain":"go1.24.0"}`),
			wantPath: "commands.tool.toolchain",
		},
		"forbidden-hooks": {
			boundary: "manifest",
			manifest: with(`{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","hooks":{"pre_build":"scripts/helper"}}`),
			wantPath: "commands.tool.hooks",
		},
		"mixed-script-build-shape": {
			boundary: "manifest",
			manifest: with(`{"type":"build","driver":"go-v1","source_dir":"build/cmd/tool","unix_path":"scripts/helper"}`),
			wantPath: "commands.tool.unix_path",
		},
	}
}

func filesystemBoundaryRejections() map[string]manifestRejection {
	return map[string]manifestRejection{
		"missing-build-roots": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"commands":{` + conformanceBuildCommand + `}`),
			wantPath: "commands.tool.source_dir",
		},
		"missing-build-root-directory": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["absent"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"absent/cmd/tool"}}`),
			wantPath: "build_roots[0]",
		},
		"unused-build-root": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["build","spare"],"commands":{` + conformanceBuildCommand + `}`),
			files:    map[string]string{"spare/go.mod": "module example.com/spare\n\ngo 1.23\n"},
			wantPath: "build_roots[1]",
		},
		"overlapping-build-roots": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["build","build/cmd"],"commands":{` + conformanceBuildCommand + `}`),
			wantPath: "build_roots",
		},
		"runtime-overlapping-build-root": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"runtime_roots":["build/cmd"],"build_roots":["build"],"commands":{` + conformanceBuildCommand + `}`),
			wantPath: "build_roots",
		},
		"root-build-root": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["."],"commands":{` + conformanceBuildCommand + `}`),
			wantPath: "build_roots[0]",
		},
		"build-root-symlink": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["linked"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"linked/cmd/tool"}}`),
			prepare: func(t *testing.T, dir string) {
				linkConformancePath(t, "build", filepath.Join(dir, "linked"))
			},
			wantPath: "build_roots[0]",
		},
		"build-root-special-file": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["special"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"special/cmd/tool"}}`),
			prepare: func(t *testing.T, dir string) {
				makeConformanceSpecialFile(t, filepath.Join(dir, "special"))
			},
			wantPath: "build_roots[0]",
		},
		"root-source-dir": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["build"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"."}}`),
			wantPath: "commands.tool.source_dir",
		},
		"escaped-source-dir": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["build"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"build/../../escape"}}`),
			wantPath: "commands.tool.source_dir",
		},
		"source-outside-root": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["build"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"outside/cmd/tool"}}`),
			files:    map[string]string{"outside/cmd/tool/main.go": "package main\n"},
			wantPath: "commands.tool.source_dir",
		},
		"source-link": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["build"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"build/cmd/linked"}}`),
			prepare: func(t *testing.T, dir string) {
				linkConformancePath(t, "tool", filepath.Join(dir, "build", "cmd", "linked"))
			},
			wantPath: "commands.tool.source_dir",
		},
		"source-special-file": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["build"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"build/cmd/special"}}`),
			prepare: func(t *testing.T, dir string) {
				makeConformanceSpecialFile(t, filepath.Join(dir, "build", "cmd", "special"))
			},
			wantPath: "commands.tool.source_dir",
		},
		"source-not-directory": {
			boundary: "filesystem",
			manifest: conformanceManifest(`"build_roots":["build"],"commands":{"tool":{"type":"build","driver":"go-v1","source_dir":"build/cmd/plain"}}`),
			files:    map[string]string{"build/cmd/plain": "not a directory\n"},
			wantPath: "commands.tool.source_dir",
		},
	}
}

// TestManifestAndFilesystemRejectionVectors proves every authoritative manifest
// and filesystem rejection cluster reaches a stable Curator validation error
// before any package code, module graph, or Go toolchain is consulted.
func TestManifestAndFilesystemRejectionVectors(t *testing.T) {
	published := loadBuildDriverRejections(t)
	owned := map[string]manifestRejection{}
	for name, mapping := range manifestBoundaryRejections() {
		owned[name] = mapping
	}
	for name, mapping := range filesystemBoundaryRejections() {
		owned[name] = mapping
	}

	for _, testCase := range publishedNamesForBoundaries(published, "manifest", "filesystem") {
		mapping, ok := owned[testCase]
		if !ok {
			t.Errorf("authoritative rejection %q has no Curator mapping", testCase)
			continue
		}
		vector := published[testCase]
		t.Run(testCase, func(t *testing.T) {
			if vector.Expected.Result != "reject" || vector.Expected.Reuse || vector.Expected.ArtifactExecuted {
				t.Fatalf("vector %q no longer fails closed: %+v", testCase, vector.Expected)
			}
			if mapping.boundary != vector.Boundary {
				t.Fatalf("Curator owns %q at the %s boundary, suite publishes %s", testCase, mapping.boundary, vector.Boundary)
			}
			dir := writeConformanceTree(t, mapping.files, mapping.prepare, mapping.manifest)
			_, err := Load(dir)
			if err == nil {
				t.Fatalf("%s was accepted, want the %s rejection", testCase, vector.Expected.Error)
			}
			var validation *verr.Error
			if !errors.As(err, &validation) {
				t.Fatalf("%s produced %v, want a stable validation error", testCase, err)
			}
			if validation.Path != mapping.wantPath {
				t.Fatalf("%s validation path = %q, want %q (message: %s)", testCase, validation.Path, mapping.wantPath, validation.Message)
			}
			if validation.Message == "" {
				t.Fatalf("%s carries no diagnostic message", testCase)
			}
		})
	}

	for name := range owned {
		if _, ok := published[name]; !ok {
			t.Errorf("Curator maps %q, which the authoritative suite no longer publishes", name)
		}
	}
}

// TestSchemaSixMixedScriptAndBuildCommandsVector is the executable assertion for
// the authoritative positive manifest vector.
func TestSchemaSixMixedScriptAndBuildCommandsVector(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "build-drivers.json")) // #nosec G304 -- explicit conformance input
	if os.IsNotExist(err) {
		t.Skipf("%s publishes no build-drivers vector", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		PositiveCases []struct {
			Name     string          `json:"name"`
			Result   string          `json:"result"`
			Manifest json.RawMessage `json:"manifest"`
		} `json:"positive_cases"`
	}
	if err := json.Unmarshal(payload, &vectors); err != nil {
		t.Fatal(err)
	}
	var manifest json.RawMessage
	var result string
	for _, testCase := range vectors.PositiveCases {
		if testCase.Name == "schema-6-mixed-script-and-build-commands" {
			manifest, result = testCase.Manifest, testCase.Result
		}
	}
	if len(manifest) == 0 {
		t.Fatal("authoritative suite publishes no schema-6-mixed-script-and-build-commands manifest")
	}
	if result != "accepted" {
		t.Fatalf("vector result = %q, want accepted", result)
	}

	// The published vector carries only the shape under test. The remaining
	// schema-required envelope is taken from the authoritative agent-skill-v6
	// schema case, so no expected value is invented here.
	merged := map[string]any{}
	if err := json.Unmarshal(manifest, &merged); err != nil {
		t.Fatal(err)
	}
	envelopePayload, err := os.ReadFile(filepath.Join(root, "schema-cases", "agent-skill-v6", "valid.json")) // #nosec G304 -- explicit conformance input
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]any
	if err := json.Unmarshal(envelopePayload, &envelope); err != nil {
		t.Fatal(err)
	}
	for key, value := range envelope {
		if _, present := merged[key]; !present {
			merged[key] = value
		}
	}
	completed, err := json.Marshal(merged)
	if err != nil {
		t.Fatal(err)
	}

	// The published manifest names the suite fixture layout, so the snapshot is
	// materialised from the manifest itself rather than from a private copy.
	var document struct {
		SchemaVersion int      `json:"schema_version"`
		BuildRoots    []string `json:"build_roots"`
		RuntimeRoots  []string `json:"runtime_roots"`
		Commands      map[string]struct {
			Type      string `json:"type"`
			Driver    string `json:"driver"`
			SourceDir string `json:"source_dir"`
			UnixPath  string `json:"unix_path"`
			WinPath   string `json:"win_path"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(manifest, &document); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, buildRoot := range document.BuildRoots {
		files[buildRoot+"/go.mod"] = "module example.com/golden\n\ngo 1.23\n"
	}
	for _, command := range document.Commands {
		if command.SourceDir != "" {
			files[command.SourceDir+"/main.go"] = "package main\n\nfunc main() {}\n"
		}
		if command.UnixPath != "" {
			files[command.UnixPath] = "#!/bin/sh\n"
		}
		if command.WinPath != "" {
			files[command.WinPath] = "@echo off\n"
		}
	}
	dir := writeSkill(t, string(completed), files)
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# golden\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec, err := Load(dir)
	if err != nil {
		t.Fatalf("authoritative positive manifest was rejected: %v", err)
	}
	if spec.SchemaVersion != 6 {
		t.Fatalf("schema version = %d", spec.SchemaVersion)
	}
	var builds, scripts int
	for name, command := range document.Commands {
		parsed, ok := spec.Commands[name]
		if !ok {
			t.Fatalf("command %q was dropped", name)
		}
		if parsed.Type != command.Type {
			t.Fatalf("command %q type = %q, want %q", name, parsed.Type, command.Type)
		}
		switch command.Type {
		case "build":
			builds++
			if parsed.Driver != command.Driver || parsed.SourceDir != command.SourceDir {
				t.Fatalf("build command %q = %+v", name, parsed)
			}
		case "script":
			scripts++
			if parsed.UnixPath != command.UnixPath || parsed.WinPath != command.WinPath {
				t.Fatalf("script command %q = %+v", name, parsed)
			}
		}
	}
	if builds == 0 || scripts == 0 {
		t.Fatalf("mixed vector resolved %d build and %d script commands", builds, scripts)
	}
}

// publishedNamesForBoundaries returns the published case names for the given
// boundaries in the suite's own order.
func publishedNamesForBoundaries(published map[string]buildDriverRejection, boundaries ...string) []string {
	wanted := make(map[string]bool, len(boundaries))
	for _, boundary := range boundaries {
		wanted[boundary] = true
	}
	names := make([]string, 0, len(published))
	for name, testCase := range published {
		if wanted[testCase.Boundary] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
