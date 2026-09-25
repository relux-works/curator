package skillspec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/identifiers"
)

func TestPortablePathConformanceVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "portable-paths.json"))
	if err != nil {
		t.Fatal(err)
	}
	var published []struct {
		Input string `json:"input"`
		Valid bool   `json:"valid"`
	}
	if err := json.Unmarshal(payload, &published); err != nil {
		t.Fatal(err)
	}
	type portablePathCase struct {
		ID    string
		Input string
		Valid bool
	}
	cases := make([]portablePathCase, len(published))
	for i, testCase := range published {
		cases[i] = portablePathCase{ID: "case-" + strconv.Itoa(i+1), Input: testCase.Input, Valid: testCase.Valid}
	}
	conformancecoverage.Run(t, "portable-paths/vectors", cases,
		func(tc portablePathCase) string { return tc.ID }, func(t *testing.T, testCase portablePathCase) {
			_, err := validateRelativePath(testCase.Input, "path", true)
			if (err == nil) != testCase.Valid {
				t.Errorf("path %q valid=%v, error=%v", testCase.Input, testCase.Valid, err)
			}
		})
}

// TestReleasedSchemaCases runs the published schema-case families of the
// manifest schemas this build claims. The schema-8 families are named
// explicitly: a root that stops publishing them makes this test fail on the
// missing directory rather than quietly narrowing what was checked.
func TestReleasedSchemaCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	for _, suite := range []struct{ directory, manifest string }{
		{"agent-skill-v7", CanonicalManifestName},
		{"csk-skill-v7", LegacyManifestName},
		{"agent-skill-v8", CanonicalManifestName},
		{"csk-skill-v8", LegacyManifestName},
	} {
		entries, err := os.ReadDir(filepath.Join(root, "schema-cases", suite.directory))
		if err != nil {
			t.Fatal(err)
		}
		type schemaCase struct{ Name string }
		var cases []schemaCase
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			cases = append(cases, schemaCase{Name: entry.Name()})
		}
		conformancecoverage.Run(t, suite.directory+"/schema-cases", cases,
			func(tc schemaCase) string { return tc.Name }, func(t *testing.T, testCase schemaCase) {
				payload, err := os.ReadFile(filepath.Join(root, "schema-cases", suite.directory, testCase.Name))
				if err != nil {
					t.Fatal(err)
				}
				snapshot := materializeManifestFixture(t, payload, suite.manifest)
				_, err = Load(snapshot)
				wantValid := strings.HasPrefix(testCase.Name, "valid")
				if (err == nil) != wantValid {
					t.Fatalf("valid=%v, error=%v", wantValid, err)
				}
			})
	}
}

func TestDraftManifestDependencyDirectoryGrammar(t *testing.T) {
	vectorPath := filepath.Join("testdata", "draft-sources-v1", "manifest-dependency-directories.json")
	payload, err := os.ReadFile(vectorPath)
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		DirectoryGrammarCases []struct {
			Name  string `json:"name"`
			Input string `json:"input"`
			Valid bool   `json:"valid"`
		} `json:"directory_grammar_cases"`
	}
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatal(err)
	}
	if len(vector.DirectoryGrammarCases) != 9 {
		t.Fatalf("directory grammar vector has %d cases, want 9", len(vector.DirectoryGrammarCases))
	}
	for _, testCase := range vector.DirectoryGrammarCases {
		t.Run(testCase.Name, func(t *testing.T) {
			manifest := map[string]any{
				"schema_version": 9,
				"capabilities":   map[string]any{},
				"dependencies": map[string]any{"skills": map[string]any{
					"developer": map[string]any{
						"git":       "https://github.com/example/role-skills.git",
						"ref":       map[string]any{"kind": "revision", "value": strings.Repeat("a", 40)},
						"directory": testCase.Input,
					},
				}},
			}
			raw, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Load(writeSkill(t, string(raw), nil))
			if (err == nil) != testCase.Valid {
				t.Fatalf("valid=%v, error=%v", testCase.Valid, err)
			}
			if err != nil && !strings.Contains(err.Error(), "dependencies.skills.developer.directory") {
				t.Fatalf("error %q does not identify the directory field", err)
			}
		})
	}
}

func TestDraftManifestDependencyDirectorySchemaCases(t *testing.T) {
	root := filepath.Join("testdata", "draft-sources-v1", "schema-cases")
	for _, suite := range []struct{ directory, manifest string }{
		{"agent-skill-v9", CanonicalManifestName},
		{"csk-skill-v9", LegacyManifestName},
	} {
		entries, err := os.ReadDir(filepath.Join(root, suite.directory))
		if err != nil {
			t.Fatal(err)
		}
		consumed := 0
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			if entry.Name() == "valid.json" || entry.Name() == "invalid.json" {
				continue // These pin schema-level version acceptance, not the parser's directory gate.
			}
			consumed++
			t.Run(suite.directory+"/"+entry.Name(), func(t *testing.T) {
				payload, err := os.ReadFile(filepath.Join(root, suite.directory, entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				snapshot := materializeManifestFixture(t, payload, suite.manifest)
				_, err = Load(snapshot)
				wantValid := strings.HasPrefix(entry.Name(), "valid")
				if (err == nil) != wantValid {
					t.Fatalf("valid=%v, error=%v", wantValid, err)
				}
			})
		}
		if consumed != 9 {
			t.Fatalf("schema-cases/%s contains %d directory cases, want 9", suite.directory, consumed)
		}
	}
}

// materializeManifestFixture lays out the snapshot a schema case describes:
// build roots with their go.mod, declared runtime roots, command source
// directories and script files, and the schema-8 declared module directories
// with their own go.mod. A path the case deliberately makes non-portable is
// left unmaterialised, because the parser must reject it on its spelling
// alone.
//
// Runtime roots are materialised so an invalid case fails for the rule it was
// written to test. Without them every schema-8 script case that declares a
// runtime root is rejected first by "runtime root does not exist", which is
// true of the valid cases too and therefore proves nothing about the invalid
// ones.
func materializeManifestFixture(t *testing.T, payload []byte, manifestName string) string {
	t.Helper()
	dir := t.TempDir()
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	if roots, ok := object["build_roots"].([]any); ok {
		for _, raw := range roots {
			root, ok := raw.(string)
			if !ok || root == "." || strings.Contains(root, "..") {
				continue
			}
			if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(root)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(root), "go.mod"), []byte("module fixture\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if roots, ok := object["runtime_roots"].([]any); ok {
		for _, raw := range roots {
			root, ok := raw.(string)
			if !ok || !materializablePath(root) {
				continue
			}
			if err := os.MkdirAll(filepath.Join(dir, filepath.FromSlash(root)), 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	if commands, ok := object["commands"].(map[string]any); ok {
		for _, raw := range commands {
			command, _ := raw.(map[string]any)
			modules, _ := command["modules"].([]any)
			for _, rawModule := range modules {
				module, ok := rawModule.(string)
				if !ok || !materializablePath(module) {
					continue
				}
				directory := filepath.Join(dir, filepath.FromSlash(module))
				if err := os.MkdirAll(directory, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(directory, "go.mod"), []byte("module fixture\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			for _, field := range []string{"source_dir", "unix_path", "win_path"} {
				value, ok := command[field].(string)
				if !ok || value == "." || strings.Contains(value, "..") {
					continue
				}
				path := filepath.Join(dir, filepath.FromSlash(value))
				if field == "source_dir" {
					if err := os.MkdirAll(path, 0o755); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, []byte("fixture"), 0o755); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
	if err := os.WriteFile(filepath.Join(dir, manifestName), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// materializablePath reports whether a declared path can be laid out on this
// host without escaping the fixture directory or colliding with a platform
// rule the parser is supposed to reject on its own.
func materializablePath(value string) bool {
	if value == "" || value == "." || strings.HasPrefix(value, "/") ||
		strings.Contains(value, "..") || strings.Contains(value, `\`) {
		return false
	}
	for _, component := range strings.Split(value, "/") {
		if !identifiers.PortableComponent(component) {
			return false
		}
	}
	return true
}
