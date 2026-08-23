package skillspec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	var cases []struct {
		Input string `json:"input"`
		Valid bool   `json:"valid"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range cases {
		_, err := validateRelativePath(testCase.Input, "path", true)
		if (err == nil) != testCase.Valid {
			t.Errorf("path %q valid=%v, error=%v", testCase.Input, testCase.Valid, err)
		}
	}
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
		consumed := 0
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			consumed++
			t.Run(suite.directory+"/"+entry.Name(), func(t *testing.T) {
				payload, err := os.ReadFile(filepath.Join(root, "schema-cases", suite.directory, entry.Name()))
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
		if consumed == 0 {
			t.Fatalf("the root publishes schema-cases/%s but it contains no cases", suite.directory)
		}
	}
}

// materializeManifestFixture lays out the snapshot a schema case describes:
// build roots with their go.mod, command source directories and script files,
// and the schema-8 declared module directories with their own go.mod. A path
// the case deliberately makes non-portable is left unmaterialised, because the
// parser must reject it on its spelling alone.
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
