package permissionsmemory

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Claim: public collection keeps request digests and fixed marker booleans,
// never the synthetic native prompt; empty intake and a private locator in
// identity data refuse without publishing an output file.
func TestCollector(t *testing.T) {
	for _, kind := range []string{"bounded", "empty", "private"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			input := filepath.Join(root, "input")
			out := filepath.Join(root, "public.json")
			if err := os.MkdirAll(input, 0700); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "bounded":
				write(t, filepath.Join(input, "case", "evidence", "requests.json"), `[{"input":"NEVER_PUBLISH_NATIVE_PROMPT QUAL_CODEX_MEMORY_CANARY"}]`)
			case "private":
				write(t, filepath.Join(input, "case", "evidence", "identity.json"), `{"version":"/home/fixture/private-location"}`)
			}
			cmd := exec.Command("/usr/bin/python3", "collect.py", out, "control="+input)
			body, err := cmd.CombinedOutput()
			if kind != "bounded" {
				if err == nil {
					t.Fatal("collector accepted forbidden intake")
				}
				expected := "no native records"
				if kind == "private" {
					expected = "private locator"
				}
				if !strings.Contains(string(body), expected) {
					t.Fatalf("unrelated collection failure: %s", body)
				}
				if _, err := os.Stat(out); !os.IsNotExist(err) {
					t.Fatal("refused collection published output")
				}
				return
			}
			if err != nil {
				t.Fatalf("valid collection failed: %s", body)
			}
			raw := read(t, out)
			var projected map[string]any
			if json.Unmarshal([]byte(raw), &projected) != nil || !strings.Contains(raw, `"QUAL_CODEX_MEMORY_CANARY": true`) || !strings.Contains(raw, `"sha256"`) {
				t.Fatal("missing bounded witnesses")
			}
			if strings.Contains(raw, "NEVER_PUBLISH_NATIVE_PROMPT") {
				t.Fatal("native prompt leaked")
			}
		})
	}
}

// Claim: a canary in ordinary input or tool schemas never counts as a tool
// result. A real function_call_output remains available for exact matching.
func TestFunctionResults(t *testing.T) {
	rows := []map[string]any{{"input": []any{map[string]any{"type": "message", "content": "NON_RESULT_CANARY"}, map[string]any{"type": "function_call_output", "output": "ACTUAL_RESULT"}}, "tools": []any{map[string]any{"name": "NON_RESULT_TOOL"}}}}
	got := functionResults(rows)
	if !strings.Contains(got, "ACTUAL_RESULT") || strings.Contains(got, "NON_RESULT") {
		t.Fatal("tool-result attribution lost")
	}
}
