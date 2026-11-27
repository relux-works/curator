package skillspec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchema10ToolsDependency(t *testing.T) {
	dir := t.TempDir()
	manifest := `{
  "schema_version": 10,
  "capabilities": {
    "network": ["colab.research.google.com"],
    "exec": ["colab", "ssh"]
  },
  "dependencies": {
    "tools": {
      "colab": {"id": "pypi/google-colab-cli", "version": ">=0.8.0"},
      "uv": {"id": "aqua/astral-sh/uv", "version": ">=0.10.0"}
    }
  }
}`
	if err := os.WriteFile(filepath.Join(dir, CanonicalManifestName), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := Load(dir)
	if err != nil {
		t.Fatalf("Load schema 10 failed: %v", err)
	}
	if spec.SchemaVersion != 10 {
		t.Fatalf("SchemaVersion = %d, want 10", spec.SchemaVersion)
	}
	if len(spec.Tools) != 2 {
		t.Fatalf("len(Tools) = %d, want 2", len(spec.Tools))
	}
	if got := spec.Tools["colab"]; got.ID != "pypi/google-colab-cli" || got.Version != ">=0.8.0" || got.Alias != "colab" {
		t.Fatalf("unexpected colab tool: %+v", got)
	}
	if got := spec.Tools["uv"]; got.ID != "aqua/astral-sh/uv" || got.Version != ">=0.10.0" {
		t.Fatalf("unexpected uv tool: %+v", got)
	}
}

func TestSchema10ToolsRejections(t *testing.T) {
	cases := []struct {
		name     string
		manifest string
		wantErr  string
	}{
		{
			name:     "schema 9 rejects dependencies.tools",
			manifest: `{"schema_version":9,"capabilities":{},"dependencies":{"tools":{"colab":{"id":"pypi/google-colab-cli","version":">=0.8.0"}}}}`,
			wantErr:  "requires schema_version 10",
		},
		{
			name:     "rejects arbitrary url in id",
			manifest: `{"schema_version":10,"capabilities":{},"dependencies":{"tools":{"colab":{"id":"https://evil.example/colab.tgz","version":"1.0.0"}}}}`,
			wantErr:  "closed catalog id",
		},
		{
			name:     "rejects extra url field",
			manifest: `{"schema_version":10,"capabilities":{},"dependencies":{"tools":{"colab":{"id":"pypi/google-colab-cli","version":"0.8.0","url":"https://evil.example"}}}}`,
			wantErr:  "unsupported",
		},
		{
			name:     "rejects curator- prefix alias",
			manifest: `{"schema_version":10,"capabilities":{},"dependencies":{"tools":{"curator-evil":{"id":"pypi/google-colab-cli","version":"0.8.0"}}}}`,
			wantErr:  "curator-",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, CanonicalManifestName), []byte(tc.manifest), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(dir)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Load() err = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}
