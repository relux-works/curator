package harnessparity

import (
	"github.com/relux-works/curator/internal/testgitenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	cleanup := testgitenv.Isolate()
	root, err := os.MkdirTemp("", "original-capture-system-")
	if err != nil {
		os.Exit(90)
	}
	path := filepath.Join(root, "system.json")
	if os.WriteFile(path, []byte(`{"schema_version":2,"locked":[]}`), 0600) != nil {
		os.Exit(91)
	}
	if os.Setenv("CURATOR_SYSTEM_CONFIG", path) != nil {
		os.Exit(92)
	}
	code := m.Run()
	_ = os.RemoveAll(root)
	cleanup()
	os.Exit(code)
}

// Both exact positive trees must pass and both actual one-byte plants must fail.
func TestOriginalComparatorControls(t *testing.T) {
	for _, c := range []fixtureCase{{Name: "claude_code-mcp", ID: "claude_code", Operation: "managed", State: "user-settings"}, {Name: "claude-launcher-typed-projection", ID: "claude_code", Operation: "launcher", State: "typed-projection"}} {
		t.Run(c.Name, func(t *testing.T) {
			original := runCase(t, c)
			if original.Refusal != "" {
				t.Fatal("positive control refused")
			}
			if err := compareResult(original, original); err != nil {
				t.Fatal(err)
			}
			altered := original
			nodes := append([]treeNode(nil), original.Rendered...)
			target := ".agent-context/mcp/claude_code.json"
			if c.Operation == "launcher" {
				nodes = append([]treeNode(nil), original.Tree...)
				target = "projection.json"
			}
			found := false
			want, got := "", ""
			for i, node := range nodes {
				if node.Path == target && node.Kind == "file" && len(node.Bytes) > 0 {
					nodes[i].Bytes = append([]byte(nil), node.Bytes...)
					nodes[i].Bytes[0] ^= 1
					want = node.SHA256
					got = hashBytes(nodes[i].Bytes)
					nodes[i].SHA256 = got
					found = true
					break
				}
			}
			if !found {
				t.Fatal("mutation not applied")
			}
			if c.Operation == "launcher" {
				altered.Tree = nodes
			} else {
				altered.Rendered = nodes
			}
			err := compareResult(original, altered)
			if err == nil || !strings.Contains(err.Error(), target) || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), got) || want == got {
				t.Fatalf("wrong mutation verdict: %v", err)
			}
		})
	}
}

// Literal and JSON-escaped Windows fixture-root parameters normalize equally;
// unrelated values remain unchanged. This does not claim native Windows output.
func TestWindowsRootParameter(t *testing.T) {
	root := `C:\fixture\root`
	raw := []byte(`{"path":"C:\\fixture\\root\\child","other":"unchanged"}`)
	got := string(fixtureTemplate(raw, root))
	want := `{"path":"${FIXTURE_ROOT}\\child","other":"unchanged"}`
	if got != want {
		t.Fatal("escaped fixture parameter mismatch")
	}
	if string(fixtureTemplate([]byte("unrelated"), root)) != "unrelated" {
		t.Fatal("unrelated bytes changed")
	}
}
