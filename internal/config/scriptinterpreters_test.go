package config

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestParseScriptInterpreters proves the operator-trusted interpreter
// bindings parse, and that the mapping is closed and absolute: an unknown
// identifier, a relative path, or a non-object is rejected rather than
// carried into a launch.
func TestParseScriptInterpreters(t *testing.T) {
	load := func(t *testing.T, text string) (*Config, error) {
		t.Helper()
		return Load(writeConfig(t, t.TempDir(), "config.json", text), nil)
	}
	base := `{"schema_version": 1, "skills_root": "/tmp/skills", "projects": {}`

	config, err := load(t, base+`}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(config.ScriptInterpreters) != 0 {
		t.Fatalf("absent bindings = %v, want empty", config.ScriptInterpreters)
	}

	// The accepted bindings are built from the host temp dir so they are
	// absolute on every OS: a POSIX literal such as /opt/... is not
	// absolute on Windows, where filepath.IsAbs needs a volume. On
	// Windows this row is therefore also the volume-absolute acceptance
	// row (e.g. C:\Users\...\node).
	binDir := t.TempDir()
	nodePath := filepath.Join(binDir, "node")
	pythonPath := filepath.Join(binDir, "python3")
	config, err = load(t, base+`, "script_interpreters": {"node-v1": `+strconv.Quote(nodePath)+`, "python3-v1": `+strconv.Quote(pythonPath)+`}}`)
	if err != nil {
		t.Fatal(err)
	}
	if config.ScriptInterpreters["node-v1"] != nodePath || config.ScriptInterpreters["python3-v1"] != pythonPath {
		t.Fatalf("bindings = %v", config.ScriptInterpreters)
	}
	if !managerKeys["script_interpreters"] {
		t.Fatal("script_interpreters is not a known manager key")
	}
	if LockableKeys["script_interpreters"] {
		t.Fatal("interpreter bindings are operator-owned executable selections and must never be lockable")
	}

	type rejection struct {
		name  string
		value string
		want  string
	}
	cases := []rejection{
		{"unknown identifier", `{"bash-v1": "/bin/bash"}`, "script_interpreters.bash-v1"},
		{"empty identifier set is fine but null is not an object", `null`, ""},
		{"non-object", `["/bin/python3"]`, "script_interpreters"},
		{"empty path", `{"node-v1": ""}`, "script_interpreters.node-v1"},
		{"relative path", `{"node-v1": "bin/node"}`, "absolute"},
		{"non-string path", `{"node-v1": 42}`, "script_interpreters.node-v1"},
	}
	// Windows has two more not-absolute shapes the production check
	// (filepath.IsAbs) must reject exactly like a relative POSIX path: a
	// drive-relative path and a rooted path without a volume.
	if runtime.GOOS == "windows" {
		cases = append(cases,
			rejection{"drive-relative path", `{"node-v1": "C:tools\\node.exe"}`, "absolute"},
			rejection{"rooted path without a volume", `{"node-v1": "\\tools\\node.exe"}`, "absolute"},
		)
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			config, err := load(t, base+`, "script_interpreters": `+testCase.value+`}`)
			if testCase.want == "" {
				if err != nil {
					t.Fatalf("null bindings were rejected: %v", err)
				}
				if len(config.ScriptInterpreters) != 0 {
					t.Fatalf("null bindings = %v, want empty", config.ScriptInterpreters)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("err = %v, want mention of %q", err, testCase.want)
			}
		})
	}
}
