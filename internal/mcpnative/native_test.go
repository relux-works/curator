package mcpnative

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/contextmaterialize"
)

// Each editor inserts absent records while preserving foreign entry bytes and
// unrelated settings; quotes, LF and DEL arguments round-trip through parsers.
// Equal and different same-NAME records are both conflicts.
func TestAppendPreservesForeignAndRefusesExisting(t *testing.T) {
	for _, tc := range []struct{ env, doc string }{
		{"claude_code", " {\n \"setting\": [1,2], \"mcpServers\": { \"foreign\" : {\"command\":\"external\", \"args\": []} } }\n"},
		{"claude_code", "{\"setting\":true}\n"},
		{"codex_cli", "setting=true\n[mcp_servers.foreign]\ncommand='external'"},
		{"codex_cli", "# keep\nsetting = \"value\"\n[mcp_servers.\"foreign\"] # untouched\ncommand = 'external'\nargs = [\"x\", # comment\n \"y\"]\nnote = '''\n[mcp_servers.fake]\n'''\n"},
	} {
		t.Run(tc.env+tc.doc[:1], func(t *testing.T) {
			input := []byte(tc.doc)
			before, err := Entries(tc.env, input)
			if err != nil {
				t.Fatal(err)
			}
			server := contextmaterialize.MCPServer{Name: "fresh", Transport: "stdio", Command: "fixture-mcp", Args: []string{"a\"b", "line\nnext", "del\x7f"}}
			out, entry, err := Add(tc.env, input, server)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(out, []byte("setting")) {
				t.Fatal("foreign setting lost")
			}
			after, err := Entries(tc.env, out)
			if err != nil {
				t.Fatal(err)
			}
			if len(entry) == 0 || !bytes.Equal(after["fresh"], entry) {
				t.Fatal("new entry span is absent or unstable")
			}
			var decoded struct {
				Args []string `json:"args" toml:"args"`
			}
			if tc.env == "codex_cli" {
				var document struct {
					Servers map[string]struct{ Args []string } `toml:"mcp_servers"`
				}
				if _, err = toml.Decode(string(out), &document); err != nil {
					t.Fatal(err)
				}
				decoded.Args = document.Servers["fresh"].Args
			} else if err = json.Unmarshal(entry, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded.Args, server.Args) {
				t.Fatalf("argument round-trip = %q, want %q", decoded.Args, server.Args)
			}
			for name, raw := range before {
				if !bytes.Equal(raw, after[name]) {
					t.Fatalf("foreign entry %s changed", name)
				}
			}
			if tc.env == "codex_cli" && !bytes.HasPrefix(out, input) {
				t.Fatal("foreign TOML bytes changed")
			}
			more, _, err := Add(tc.env, out, contextmaterialize.MCPServer{Name: "second", Transport: "stdio", Command: "fixture-mcp"})
			if err != nil {
				t.Fatal(err)
			}
			again, _ := Entries(tc.env, more)
			if !bytes.Equal(entry, again["fresh"]) {
				t.Fatal("adding another record changed installed entry span")
			}
			for _, command := range []string{"fixture-mcp", "different"} {
				server.Command = command
				if _, _, err = Add(tc.env, out, server); err == nil || !strings.Contains(err.Error(), Conflict) {
					t.Fatalf("same-NAME accepted: %v", err)
				}
			}
		})
	}
}

// Malformed native containers, duplicate keys and invalid table definitions
// refuse inspection; the adjacent valid controls above establish non-reject-all.
func TestUnreadableNativeDocuments(t *testing.T) {
	for _, tc := range []struct{ env, doc string }{
		{"claude_code", `{"mcpServers":{},"mcpServers":{}}`}, {"claude_code", `{"mcpServers":{"x":{},"x":{}}}`},
		{"claude_code", `{"mcpServers":[]}`}, {"claude_code", `[]`}, {"claude_code", `{"x":"\ud800"}`},
		{"codex_cli", "[mcp_servers.x]\ncommand='a'\n[mcp_servers.x]\ncommand='b'\n"},
		{"codex_cli", "mcp_servers='bad'\n"},
	} {
		if _, err := Entries(tc.env, []byte(tc.doc)); err == nil || !strings.Contains(err.Error(), Unreadable) {
			t.Fatalf("%s accepted %q: %v", tc.env, tc.doc, err)
		}
	}
}

// Complete canonical markers bind owner, profile, environment and selected path;
// no hash-only, open-shape or cross-home record confers ownership.
func TestOwnershipClosedAndBound(t *testing.T) {
	o := NewOwnership("sample", "claude_code")
	o.Records = append(o.Records, Record{Name: "fresh", Owner: "curator", ServerSHA256: Digest([]byte("s")), ContentSHA256: Digest([]byte("c")), InstalledSHA256: Digest([]byte("entry"))})
	valid, err := o.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = DecodeOwnership(valid, "sample", "claude_code"); err != nil {
		t.Fatal(err)
	}
	for _, edit := range []func(map[string]any){
		func(m map[string]any) { m["extra"] = true }, func(m map[string]any) { m["profile_id"] = "another" }, func(m map[string]any) { m["environment"] = "codex_cli" }, func(m map[string]any) { m["path"] = "other.json" },
		func(m map[string]any) { delete(m["records"].([]any)[0].(map[string]any), "owner") },
		func(m map[string]any) { delete(m["records"].([]any)[0].(map[string]any), "preimage_sha256") },
		func(m map[string]any) { m["records"].([]any)[0].(map[string]any)["owner"] = "other" },
		func(m map[string]any) { m["records"].([]any)[0].(map[string]any)["installed_sha256"] = "sha256:bad" },
		func(m map[string]any) { m["records"] = append(m["records"].([]any), m["records"].([]any)[0]) },
	} {
		var value map[string]any
		_ = json.Unmarshal(valid, &value)
		edit(value)
		body, _ := projectmcpCanonical(value)
		if _, err = DecodeOwnership(body, "sample", "claude_code"); err == nil || !strings.Contains(err.Error(), Unreadable) {
			t.Fatalf("invalid marker accepted: %s", body)
		}
	}
	if _, err = DecodeOwnership(append(valid, '\n'), "sample", "claude_code"); err == nil || !strings.Contains(err.Error(), Unreadable) {
		t.Fatal("noncanonical ownership accepted")
	}
}

// Inline/dotted TOML records retain foreign names for inspection but have no
// certified table span. First install refuses even when TOML syntax could
// accept appending, so successful syntax alone cannot expand qualification.
func TestSealedInlineContainerRefusesLossyConversion(t *testing.T) {
	for _, document := range []string{"mcp_servers = { foreign = { command = 'external', args = [] } }\n", "mcp_servers.foreign.command = 'external'\n"} {
		input := []byte(document)
		entries, err := Entries("codex_cli", input)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := entries["foreign"]; !ok {
			t.Fatal("inline/dotted foreign name lost")
		}
		_, _, err = Add("codex_cli", input, contextmaterialize.MCPServer{Name: "fresh", Transport: "stdio", Command: "fixture-mcp"})
		if err == nil || !strings.Contains(err.Error(), Unreadable) {
			t.Fatalf("uncertified native table converted: %v", err)
		}
	}
}

// Legal three-, four- and five-quote TOML endings retain their semantic value
// and foreign span across append; installed spans must be nonempty and stable.
func TestTOMLMultilineClosingRuns(t *testing.T) {
	for _, quote := range []string{"\"", "'"} {
		for count := 3; count <= 5; count++ {
			t.Run(fmt.Sprintf("%s/%d", quote, count), func(t *testing.T) {
				input := []byte("[mcp_servers.foreign]\ncommand='external'\nnote=" + strings.Repeat(quote, 3) + "hello" + strings.Repeat(quote, count) + "\n")
				var before map[string]any
				if _, err := toml.Decode(string(input), &before); err != nil {
					t.Fatal(err)
				}
				entries, err := Entries("codex_cli", input)
				if err != nil {
					t.Fatal(err)
				}
				output, span, err := Add("codex_cli", input, contextmaterialize.MCPServer{Name: "fresh", Transport: "stdio", Command: "fixture-mcp"})
				if err != nil {
					t.Fatal(err)
				}
				if len(span) == 0 {
					t.Fatal("INSERTED_SPAN_ABSENT")
				}
				after, err := Entries("codex_cli", output)
				if err != nil || !bytes.Equal(after["fresh"], span) || !bytes.Equal(entries["foreign"], after["foreign"]) || !bytes.HasPrefix(output, input) {
					t.Fatalf("foreign/span changed: %v", err)
				}
				var decoded map[string]any
				if _, err = toml.Decode(string(output), &decoded); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(before["mcp_servers"].(map[string]any)["foreign"], decoded["mcp_servers"].(map[string]any)["foreign"]) {
					t.Fatal("foreign semantics changed")
				}
				if _, _, err = Add("codex_cli", output, contextmaterialize.MCPServer{Name: "fresh", Transport: "stdio", Command: "other"}); err == nil || !strings.Contains(err.Error(), Conflict) {
					t.Fatalf("conflict: %v", err)
				}
			})
		}
	}
}
