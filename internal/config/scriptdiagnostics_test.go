package config

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestParseScriptDiagnosticsDir proves the operator-selected diagnostics
// destination: absent disables reporting, an absolute directory is kept,
// and anything else — relative, empty, non-string — is rejected, because
// a relative path would resolve against a working directory the package
// controls.
func TestParseScriptDiagnosticsDir(t *testing.T) {
	load := func(t *testing.T, text string) (*Config, error) {
		t.Helper()
		return Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
	}
	base := `{"schema_version": 1, "skills_root": "/tmp/skills", "projects": {}`
	absolute := filepath.Join(t.TempDir(), "script-diagnostics")

	config, err := load(t, base+`, "script_diagnostics_dir": `+strconv.Quote(absolute)+`}`)
	if err != nil {
		t.Fatalf("an absolute diagnostics dir was rejected: %v", err)
	}
	if config.ScriptDiagnosticsDir != absolute {
		t.Fatalf("ScriptDiagnosticsDir = %q, want %q", config.ScriptDiagnosticsDir, absolute)
	}

	absent, err := load(t, base+`}`)
	if err != nil {
		t.Fatalf("an absent diagnostics dir was rejected: %v", err)
	}
	if absent.ScriptDiagnosticsDir != "" {
		t.Fatalf("ScriptDiagnosticsDir = %q, want empty", absent.ScriptDiagnosticsDir)
	}

	cases := []struct {
		name  string
		value string
		want  string
	}{
		{"relative path", `"diagnostics"`, "absolute"},
		{"empty path", `""`, "script_diagnostics_dir"},
		{"non-string path", `42`, "script_diagnostics_dir"},
	}
	if runtime.GOOS == "windows" {
		cases = append(cases,
			struct {
				name  string
				value string
				want  string
			}{"drive-relative path", `"C:diagnostics"`, "absolute"},
			struct {
				name  string
				value string
				want  string
			}{"rooted path without a volume", `"\\diagnostics"`, "absolute"},
		)
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := load(t, base+`, "script_diagnostics_dir": `+testCase.value+`}`)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("err = %v, want mention of %q", err, testCase.want)
			}
		})
	}
}
