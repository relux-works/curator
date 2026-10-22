package projectmcp

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func decode(t *testing.T, text string) any {
	t.Helper()
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

// Valid pinned siblings admit only the closed npx/uvx operand forms. Near misses
// cannot gain a different runtime selection by a flag, alias, source or range.
func TestRuntimePins(t *testing.T) {
	for _, tc := range []struct {
		command string
		args    []string
		valid   bool
	}{
		{"npx", []string{"tool@1.2.3"}, true}, {"npx", []string{"-y", "@scope/tool@0.0.1", "--stdio"}, true}, {"npx", []string{"--yes", "tool@1.2.3"}, true},
		{"npx", []string{"tool"}, false}, {"npx", []string{"tool@latest"}, false}, {"npx", []string{"tool@01.2.3"}, false}, {"npx", []string{"tool@1.2.3-beta"}, false}, {"npx", []string{"tool@^1.2.3"}, false},
		{"npx", []string{"--package=tool@1.2.3", "sh"}, false}, {"npx", []string{"-y", "--yes", "tool@1.2.3"}, false}, {"npx", []string{"https://tools.example/tool.tgz"}, false},
		{"NPX", []string{"tool@1.2.3"}, false}, {"npx.cmd", []string{"tool@1.2.3"}, false}, {"uvx.exe", []string{"tool@1.2.3"}, false},
		{"uvx", []string{"tool@1.2.3", "--stdio"}, true}, {"uvx", []string{"--from", "example-tools==1.2.3", "example-mcp"}, true},
		{"uvx", []string{"tool==1.2.3"}, false}, {"uvx", []string{"--from", "tool>=1", "tool"}, false}, {"uvx", []string{"--with", "tool@1.2.3"}, false}, {"uvx", []string{"--from", "tool==1.2.3", "-c"}, false}, {"uvx", []string{"--from", "tool==1.2.3", "./tool"}, false},
		{"tool", []string{}, true},
	} {
		t.Run(tc.command+strings.Join(tc.args, "_"), func(t *testing.T) {
			if got := validPin(tc.command, tc.args); got != tc.valid {
				t.Fatalf("pin accepted=%v want %v", got, tc.valid)
			}
		})
	}
}

// Strict URL/transport/environment/option/package validation has valid neighbors;
// secret findings identify their class without exposing literal token bytes.
func TestDeclarationValidation(t *testing.T) {
	token := "AKIA1234567890ABCDEF"
	for _, tc := range []struct{ name, text, code string }{
		{"url", `"https://docs.example/mcp"`, ""},
		{"http", `"http://docs.example/mcp"`, Invalid}, {"query", `"https://docs.example/mcp?token=x"`, Invalid}, {"fragment", `"https://docs.example/mcp#x"`, Invalid}, {"userinfo", `"https://user@docs.example/mcp"`, Invalid},
		{"space", `"https://docs.example/a b"`, Invalid}, {"format", `"https://docs.example/\u200b"`, Invalid}, {"expansion", `"https://docs.example/$TOKEN"`, Invalid}, {"unicode-host", `"https://例.example/mcp"`, Invalid},
		{"unknown", `{"server":{"transport":"http","url":"https://docs.example/mcp","auth":{}}}`, Invalid},
		{"environment", `{"server":{"transport":"http","url":"https://docs.example/mcp","environments":["pi","muse"]}}`, ""},
		{"alias", `{"server":{"transport":"http","url":"https://docs.example/mcp","environments":["codex"]}}`, Invalid},
		{"empty-environments", `{"server":{"transport":"http","url":"https://docs.example/mcp","environments":[]}}`, Invalid},
		{"duplicate-environments", `{"server":{"transport":"http","url":"https://docs.example/mcp","environments":["pi","pi"]}}`, Invalid},
		{"env-reference", `{"server":{"transport":"stdio","command":"tool","args":[],"env_names":["TOOLS_TOKEN"]}}`, ""},
		{"reserved-env", `{"server":{"transport":"stdio","command":"tool","args":[],"env_names":["PATH"]}}`, Invalid},
		{"env-value", `{"server":{"transport":"stdio","command":"tool","args":[],"env":{"TOKEN":"value"}}}`, Invalid},
		{"secret-url", `"https://docs.example/` + token + `"`, "context-secret-material"},
		{"secret-arg", `{"server":{"transport":"stdio","command":"tool","args":["` + token + `"]}}`, "context-secret-material"},
		{"secret-option-value", `{"server":{"transport":"http","url":"https://docs.example/mcp"},"install":{"headers":{"value":"` + token + `"}}}`, "context-secret-material"},
		{"secret-option-key", `{"server":{"transport":"http","url":"https://docs.example/mcp"},"install":{"` + token + `":true}}`, "context-secret-material"},
		{"ignored-options", `{"server":{"transport":"http","url":"https://docs.example/mcp"},"install":{"scope":"user","forged_approval":{"state":"active"}}}`, ""},
		{"package", `{"package":{"git":"https://code.example/tools.git","tag":"v1.2.3"}}`, ""},
		{"mixed", `{"package":{"git":"https://code.example/tools.git","tag":"v1.2.3"},"server":{"transport":"http","url":"https://docs.example/mcp"}}`, Invalid},
		{"directory-dot", `{"package":{"git":"https://code.example/tools.git","tag":"v1.2.3","directory":"."}}`, Invalid},
		{"local-package", `{"package":{"git":"file:///tmp/tool","tag":"v1.2.3"}}`, Invalid},
		{"branch-package", `{"package":{"git":"https://code.example/tools.git","branch":"main"}}`, Invalid},
		{"option-null", `{"server":{"transport":"http","url":"https://docs.example/mcp"},"install":null}`, Invalid},
		{"null", `null`, Invalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := ParseDeclaration(decode(t, tc.text))
			if tc.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				if d.Server != nil {
					var refusal *Error
					if !errors.As(d.RequirePackage(), &refusal) || refusal.Code != InlineUnsupported {
						t.Fatal("inline granted authority")
					}
				}
				return
			}
			var refusal *Error
			if !errors.As(err, &refusal) || refusal.Code != tc.code {
				t.Fatalf("error=%v want %s", err, tc.code)
			}
			if strings.Contains(err.Error(), token) {
				t.Fatal("secret echoed")
			}
		})
	}
}

// Shorthand, expanded defaults and reordered sets have one normalized identity.
// Args, explicit selectors and install options retain their separate bindings.
func TestDeclarationDigests(t *testing.T) {
	parse := func(text string) Declaration {
		d, err := ParseDeclaration(decode(t, text))
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	a := parse(`"https://docs.example/mcp"`)
	b := parse(`{"server":{"env_names":[],"url":"https://docs.example/mcp","transport":"http"},"install":{}}`)
	da, _ := Digest(a.Model())
	db, _ := Digest(b.Model())
	if da != db {
		t.Fatal("shorthand identity differs")
	}
	c := parse(`{"server":{"transport":"http","url":"https://docs.example/mcp"},"install":{"sig":"retained"}}`)
	dc, _ := Digest(c.Model())
	sa, _ := Digest(a.Server)
	sc, _ := Digest(c.Server)
	if dc == da || sa != sc {
		t.Fatal("install identity conflated")
	}
	d := parse(`{"server":{"transport":"stdio","command":"tool","args":["a","b"],"env_names":["Z_TOKEN","A_TOKEN"],"environments":["pi","muse"]}}`)
	e := parse(`{"server":{"transport":"stdio","command":"tool","args":["a","b"],"env_names":["A_TOKEN","Z_TOKEN"],"environments":["muse","pi"]}}`)
	sd, _ := Digest(d.Server)
	se, _ := Digest(e.Server)
	if sd != se {
		t.Fatal("set order changed identity")
	}
	e.Server["args"] = []string{"b", "a"}
	se, _ = Digest(e.Server)
	if sd == se {
		t.Fatal("argument order omitted")
	}
	b.Server["environments"] = []string{"claude_code"}
	sb, _ := Digest(b.Server)
	if sa == sb {
		t.Fatal("selector omitted")
	}
}

// A failed set exposes valid siblings only as refusal evidence; it never
// returns a partially admitted map. Valid sets retain their ordinary result.
func TestSetRefusalSiblings(t *testing.T) {
	valid := map[string]any{"a-valid": "https://safe.example/mcp", "z-valid": "https://safe.example/mcp"}
	if admitted, err := ParseSet(valid); err != nil || len(admitted) != 2 {
		t.Fatalf("valid set: %v %v", admitted, err)
	}
	valid["bad"] = map[string]any{"server": nil}
	admitted, err := ParseSet(valid)
	var partial *SetError
	if admitted != nil || !errors.As(err, &partial) || len(partial.Valid) != 2 {
		t.Fatalf("partial admission or lost siblings: %v %v", admitted, err)
	}
	var refusal *Error
	if !errors.As(err, &refusal) || refusal.Code != Invalid {
		t.Fatalf("lost refusal class: %v", err)
	}
}
