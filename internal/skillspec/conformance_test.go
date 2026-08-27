package skillspec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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

func TestSchemaV6ConformanceCases(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Instance string `json:"instance"`
		Schema   string `json:"schema"`
		Valid    bool   `json:"valid"`
	}
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatal(err)
	}
	consumed := map[string]int{}
	for _, testCase := range cases {
		manifestName := ""
		switch testCase.Schema {
		case "agent-skill-v6.schema.json":
			manifestName = CanonicalManifestName
		case "csk-skill-v6.schema.json":
			manifestName = LegacyManifestName
		default:
			continue
		}
		consumed[testCase.Schema]++
		t.Run(testCase.Instance, func(t *testing.T) {
			dir := t.TempDir()
			manifest, err := os.ReadFile(filepath.Join(root, "schema-cases", filepath.FromSlash(testCase.Instance)))
			if err != nil {
				t.Fatal(err)
			}
			writeConformanceFile(t, dir, manifestName, manifest)
			writeConformanceFile(t, dir, "build/go.mod", []byte("module example.com/tool\n\ngo 1.23\n"))
			writeConformanceFile(t, dir, "build/cmd/tool/main.go", []byte("package main\n"))
			writeConformanceFile(t, dir, "scripts/tool", []byte("#!/bin/sh\n"))
			_, err = Load(dir)
			if (err == nil) != testCase.Valid {
				t.Fatalf("valid = %v, error = %v", testCase.Valid, err)
			}
		})
	}
	for _, schema := range []string{"agent-skill-v6.schema.json", "csk-skill-v6.schema.json"} {
		if consumed[schema] == 0 {
			t.Fatalf("authoritative index contains no %s cases", schema)
		}
	}
}

func TestSchemaV6FixtureParsingDoesNotExecuteGo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell sentinel is Unix-only")
	}
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	fakeBin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "go-invoked")
	script := "#!/bin/sh\nprintf invoked > " + marker + "\nexit 97\n"
	if strings.ContainsAny(marker, " \t\n\"'") {
		t.Fatalf("unexpected temp path requiring shell quoting: %q", marker)
	}
	if err := os.WriteFile(filepath.Join(fakeBin, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fakeBin)
	spec, err := Load(filepath.Join(root, "fixtures", "go-build-skill"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Commands["golden-tool"].Driver != "go-v1" {
		t.Fatalf("fixture build command = %+v", spec.Commands["golden-tool"])
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("parsing invoked go; marker stat = %v", err)
	}
}

func writeConformanceFile(t *testing.T, root, rel string, content []byte) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}
