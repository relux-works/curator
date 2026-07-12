package capabilities

import (
	"encoding/json"
	"strings"
	"testing"
)

func parseJSON(t *testing.T, text string) any {
	t.Helper()
	var raw any
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		t.Fatalf("fixture JSON: %v", err)
	}
	return raw
}

func TestParseNilIsImplicitNone(t *testing.T) {
	m, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Network) != 0 || len(m.Exec) != 0 || m.PromptScope != "" {
		t.Fatalf("implicit none not empty: %+v", m)
	}
}

func TestParseDefaults(t *testing.T) {
	m, err := Parse(parseJSON(t, `{}`))
	if err != nil {
		t.Fatal(err)
	}
	if m.Filesystem.Keyword != "repo" {
		t.Fatalf("filesystem default = %+v, want repo", m.Filesystem)
	}
	if len(m.Network) != 0 || len(m.Exec) != 0 || len(m.Secrets) != 0 || len(m.EnvRead) != 0 {
		t.Fatalf("defaults not none/empty: %+v", m)
	}
}

func TestParseFull(t *testing.T) {
	m, err := Parse(parseJSON(t, `{
		"network": ["api.example.com", "*.internal"],
		"filesystem": "home-config",
		"exec": ["python3"],
		"secrets": ["tracker-cli"],
		"env_read": ["HOME", "MY_VAR1"],
		"prompt_scope": "  Do the thing.  "
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Network) != 2 || m.Filesystem.Keyword != "home-config" || m.Exec[0] != "python3" {
		t.Fatalf("parsed: %+v", m)
	}
	if m.PromptScope != "Do the thing." {
		t.Fatalf("prompt_scope not trimmed: %q", m.PromptScope)
	}
}

func TestParseFilesystemPaths(t *testing.T) {
	m, err := Parse(parseJSON(t, `{"filesystem": ["~/notes", "/var/data", "relative/ok"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Filesystem.Paths) != 3 || m.Filesystem.Keyword != "" {
		t.Fatalf("filesystem: %+v", m.Filesystem)
	}
}

func TestParseRejections(t *testing.T) {
	cases := []struct {
		name string
		text string
		path string
	}{
		{"not object", `[]`, "capabilities"},
		{"unknown field", `{"nets": []}`, "capabilities"},
		{"network url", `{"network": ["https://x"]}`, "capabilities.network[0]"},
		{"network space", `{"network": ["a b"]}`, "capabilities.network[0]"},
		{"network dup", `{"network": ["a", "a"]}`, "capabilities.network"},
		{"network type", `{"network": "all"}`, "capabilities.network"},
		{"exec path", `{"exec": ["bin/tool"]}`, "capabilities.exec[0]"},
		{"exec dash", `{"exec": ["-rf"]}`, "capabilities.exec[0]"},
		{"env digit", `{"env_read": ["1BAD"]}`, "capabilities.env_read[0]"},
		{"env symbol", `{"env_read": ["A-B"]}`, "capabilities.env_read[0]"},
		{"fs dotdot", `{"filesystem": ["a/../b"]}`, "capabilities.filesystem[0]"},
		{"fs dash", `{"filesystem": ["-x"]}`, "capabilities.filesystem[0]"},
		{"scope empty", `{"prompt_scope": "  "}`, "capabilities.prompt_scope"},
		{"scope type", `{"prompt_scope": 5}`, "capabilities.prompt_scope"},
		{"empty item", `{"secrets": [""]}`, "capabilities.secrets"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(parseJSON(t, tc.text))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.HasPrefix(err.Error(), tc.path) {
				t.Fatalf("error %q does not start with path %q", err, tc.path)
			}
		})
	}
}
