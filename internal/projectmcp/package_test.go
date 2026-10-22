package projectmcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Stage-A inline bounds count Unicode scalars; package admission also retains
// the legacy UTF-8 byte bounds. The package gate is intentionally stricter.
func TestPackageRetainsLegacyByteBounds(t *testing.T) {
	for _, n := range []int{4096, 4097} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			server := map[string]any{"transport": "stdio", "command": "npx", "args": []any{"tool@1.0.0", strings.Repeat("é", n)}}
			if _, err := ParseDeclaration(map[string]any{"server": server}); err != nil {
				t.Fatalf("inline scalar bound: %v", err)
			}
			root := t.TempDir()
			payload, err := json.Marshal(map[string]any{"schema_version": 1, "name": "tool", "version": "1.0.0", "server": server})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "agent-mcp.json"), payload, 0600); err != nil {
				t.Fatal(err)
			}
			_, err = LoadPackage(root)
			if n == 4096 && err != nil {
				t.Fatal(err)
			}
			if n == 4097 && (err == nil || !strings.Contains(err.Error(), Invalid)) {
				t.Fatalf("legacy byte bound: %v", err)
			}
		})
	}
}
