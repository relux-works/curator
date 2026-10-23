package manifest

import (
	"errors"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/projectmcp"
)

// Schema-3 dispatch precedes float conversion regardless of member order.
// Oversized signed integers in nested ignored options produce bounded typed
// diagnostics without echoing the literal; safe integer neighbors still parse.
func TestProjectMCPOversizedIntegerDiagnostics(t *testing.T) {
	for _, last := range []bool{false, true} {
		for _, token := range []string{"0", "9007199254740991", "-9007199254740991", "-0", "9007199254740992", "-9007199254740992", strings.Repeat("9", 17), strings.Repeat("9", 400), "-" + strings.Repeat("9", 400)} {
			valid := token == "0" || token == "9007199254740991" || token == "-9007199254740991"
			body := `"skills":[],"mcp_servers":{"tool":{"server":{"transport":"http","url":"https://service.example/mcp"},"install":{"nested":[{"value":` + token + `}]}}}`
			payload := `{"schema_version":3,` + body + `}`
			if last {
				payload = `{` + body + `,"schema_version":3}`
			}
			_, err := ParseBytes([]byte(payload), "Skillfile.json")
			if valid {
				if err != nil {
					t.Fatal(err)
				}
				continue
			}
			var diagnostic *projectmcp.Error
			if !errors.As(err, &diagnostic) || diagnostic.Code != projectmcp.Invalid || len(err.Error()) > 140 || (len(token) > 16 && strings.Contains(err.Error(), token)) {
				t.Fatalf("unbounded or untyped diagnostic for %d-character integer: %v", len(token), err)
			}
		}
	}
}
