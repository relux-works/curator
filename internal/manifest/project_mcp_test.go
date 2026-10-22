package manifest

import (
	"strings"
	"testing"
)

// Schema 3 inherits source selection without upgrading legacy inputs. Schemas
// 1/2 still reject the extension, and machine-global dispatch remains separate.
func TestProjectMCPSchemaDispatch(t *testing.T) {
	for _, version := range []string{"1", "2"} {
		if _, err := ParseBytes([]byte(`{"schema_version":`+version+`,"skills":[],"mcp_servers":{}}`), "Skillfile.json"); err == nil {
			t.Fatalf("schema %s admitted MCP", version)
		}
	}
	m, err := ParseBytes([]byte(`{"schema_version":3,"sources":{"s":{"path":"."}},"skills":[{"name":"review","from":"s","directory":"skills/review"}],"mcp_servers":{}}`), "Skillfile.json")
	if err != nil || m.SchemaVersion != 3 || m.Skills[0].Selector == nil {
		t.Fatalf("inheritance: %+v %v", m, err)
	}
	if _, err := ParseBytes([]byte(`{"schema_version":4,"skills":[]}`), "Skillfile.json"); err == nil || !strings.Contains(err.Error(), "requires a newer tool") {
		t.Fatal(err)
	}
}

// Strict schema-3 JSON rejects duplicate/scalar/CCJ violations before any
// declaration is usable. Diagnostics do not expose user-supplied keys or values.
func TestProjectMCPStrictJSON(t *testing.T) {
	for _, text := range []string{
		`{"schema_version":3.0,"skills":[]}`,
		`{"schema_version":3,"skills":[],"mcp_servers":{"x":"https://safe.example/mcp","x":"https://other.example/mcp"}}`,
		`{"schema_version":3,"skills":[],"mcp_servers":{"x":{"server":{"transport":"stdio","command":"tool","args":["\ud800"]}}}}`,
		`{"schema_version":3,"skills":[],"mcp_servers":{"x":{"server":{"transport":"http","url":"https://safe.example/mcp"},"install":{"value":1.5}}}}`,
		`{"schema_version":3,"skills":[],"mcp_servers":{"x":{"server":{"transport":"http","url":"https://safe.example/mcp"},"install":{"value":9007199254740992}}}}`,
		`{"schema_version":3,"skills":[],"private-unrecognized-key":true}`,
	} {
		if _, err := ParseBytes([]byte(text), "Skillfile.json"); err == nil || !strings.Contains(err.Error(), "mcp_declaration_invalid") || strings.Contains(err.Error(), "private-unrecognized-key") {
			t.Fatalf("input %s: %v", text, err)
		}
	}
	if _, err := ParseBytes([]byte(`{"schema_version":3,"skills":[],"mcp_servers":{"x":{"server":{"transport":"http","url":"https://safe.example/mcp"},"install":{"value":9007199254740991}}}}`), "Skillfile.json"); err != nil {
		t.Fatal(err)
	}
}
