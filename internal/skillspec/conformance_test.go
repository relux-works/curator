package skillspec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestSchema7ReleasedCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	for _, suite := range []struct{ directory, manifest string }{
		{"agent-skill-v7", CanonicalManifestName},
		{"csk-skill-v7", LegacyManifestName},
	} {
		entries, err := os.ReadDir(filepath.Join(root, "schema-cases", suite.directory))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			t.Run(suite.directory+"/"+entry.Name(), func(t *testing.T) {
				payload, err := os.ReadFile(filepath.Join(root, "schema-cases", suite.directory, entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				snapshot := materializeSchema7Fixture(t, payload, suite.manifest)
				_, err = Load(snapshot)
				wantValid := strings.HasPrefix(entry.Name(), "valid")
				if (err == nil) != wantValid {
					t.Fatalf("valid=%v, error=%v", wantValid, err)
				}
			})
		}
	}
}

func materializeSchema7Fixture(t *testing.T, payload []byte, manifestName string) string {
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
