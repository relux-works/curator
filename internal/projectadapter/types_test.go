package projectadapter

import "testing"

// Claim: the shared typed refusal preserves the MCP-specific diagnostic and
// uses the project diagnostic for other/unknown surfaces, with the reason in
// the rendered error. This classifies refusal data; it grants no capability.
func TestUnsupportedDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		surface SurfaceName
		code    string
	}{
		{MCP, "mcp_contribution_policy_refused"},
		{Permissions, "project_context_unsupported"},
		{"unknown", "project_context_unsupported"},
	} {
		err := &Unsupported{Surface: tc.surface, Reason: "unqualified native behavior"}
		if err.Code() != tc.code || err.Error() != tc.code+": "+string(tc.surface)+": unqualified native behavior" {
			t.Fatalf("wrong shared refusal: %v", err)
		}
	}
}
