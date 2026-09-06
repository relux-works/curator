// Package contextmaterialize MCP launch-channel materialization (environments §5.8): the resolved MCP
// set of a profile for an adapter renders as one inert, hashed file per
// adapter, in a managed home only. No env member, no value, and no
// operator-supplied byte ever enters the file: the fragment carries the
// env_names union and the operator's environment supplies values at launch.
// Where the resolved set is empty no file is written, and pi declares no
// file and no fragment mcp section in revision 1.
package contextmaterialize

import (
	"fmt"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

// MCP transports (environments §2.2).
const (
	MCPTransportStdio = "stdio"
	MCPTransportHTTP  = "http"
)

// MCP file locations below a managed home (environments §5.8). The codex
// layer path is fixed by the tool: -p layers
// $CODEX_HOME/<name>.config.toml, so curator-mcp.config.toml is a reserved
// name inside every managed codex_cli home.
const (
	MCPClaudePath   = ".agent-context/mcp/claude_code.json"
	MCPCodexPath    = "curator-mcp.config.toml"
	MCPOpenCodePath = ".agent-context/mcp/opencode.json"
)

// MCPCodexLayerName is the fixed -p layer name of the codex channel
// (environments §7.8).
const MCPCodexLayerName = "curator-mcp"

// MCPServer is one resolved MCP declaration.
type MCPServer struct {
	Name string
	// Transport is exactly stdio or http.
	Transport string
	// Command and Args carry a stdio server; URL carries an http server.
	Command string
	Args    []string
	URL     string
	// EnvNames lists the variable names the server expects at run time.
	EnvNames []string
	// Environments is the adapter selector; nil means every adapter.
	Environments []string
}

// MCPApplies reports whether the server's selector admits the environment.
func MCPApplies(server MCPServer, environment string) bool {
	if server.Environments == nil {
		return true
	}
	for _, id := range server.Environments {
		if id == environment {
			return true
		}
	}
	return false
}

// SupportsMCPFile reports whether the adapter materializes an MCP channel
// file (environments §5.8): every revision-1 adapter except pi.
func SupportsMCPFile(environment string) bool {
	switch environment {
	case "claude_code", "codex_cli", "opencode":
		return true
	default:
		return false
	}
}

// MCPPath names the adapter's MCP file home-relative path (environments
// §5.8), or "" where the adapter declares no file.
func MCPPath(environment string) string {
	switch environment {
	case "claude_code":
		return MCPClaudePath
	case "codex_cli":
		return MCPCodexPath
	case "opencode":
		return MCPOpenCodePath
	default:
		return ""
	}
}

// MCPSet resolves the adapter's MCP set: the servers whose environments
// selector applies, in sorted name order (environments §5.8).
func MCPSet(servers []MCPServer, environment string) ([]MCPServer, error) {
	var set []MCPServer
	set = []MCPServer{}
	seen := map[string]bool{}
	for _, server := range servers {
		if !MCPApplies(server, environment) {
			continue
		}
		if !identifiers.Valid(server.Name) {
			return nil, fmt.Errorf("mcp server name %q is not a portable identifier", server.Name)
		}
		if seen[server.Name] {
			return nil, fmt.Errorf("duplicate mcp server %q", server.Name)
		}
		seen[server.Name] = true
		switch server.Transport {
		case MCPTransportStdio:
			if server.Command == "" {
				return nil, fmt.Errorf("mcp server %q carries no command", server.Name)
			}
		case MCPTransportHTTP:
			if server.URL == "" {
				return nil, fmt.Errorf("mcp server %q carries no url", server.Name)
			}
		default:
			return nil, fmt.Errorf("mcp server %q transport %q is not stdio or http", server.Name, server.Transport)
		}
		set = append(set, server)
	}
	sort.Slice(set, func(i, j int) bool { return set[i].Name < set[j].Name })
	return set, nil
}

// MCPEnvNames returns the sorted union of the env_names of the servers in
// the set (environments §10.2). Values never appear: names only.
func MCPEnvNames(set []MCPServer) []string {
	union := map[string]bool{}
	for _, server := range set {
		for _, name := range server.EnvNames {
			union[name] = true
		}
	}
	names := make([]string, 0, len(union))
	for name := range union {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// MCPFile renders the adapter's MCP channel file: its home-relative path,
// its exact document bytes, and written=false where the resolved set is
// empty or the adapter declares no file (environments §5.8).
func MCPFile(environment string, servers []MCPServer) (path string, document []byte, written bool, err error) {
	set, err := MCPSet(servers, environment)
	if err != nil {
		return "", nil, false, err
	}
	if len(set) == 0 || !SupportsMCPFile(environment) {
		return "", nil, false, nil
	}
	switch environment {
	case "claude_code":
		document, err = claudeMCP(set)
	case "codex_cli":
		document, err = codexMCP(set)
	case "opencode":
		document, err = openCodeMCP(set)
	default:
		return "", nil, false, nil
	}
	if err != nil {
		return "", nil, false, err
	}
	return MCPPath(environment), document, true, nil
}

// claudeMCP renders the CCJ-1 bytes of the object whose single member
// mcpServers maps each server name to its shape, followed by exactly one
// LF (environments §5.8).
func claudeMCP(set []MCPServer) ([]byte, error) {
	table := map[string]any{}
	for _, server := range set {
		switch server.Transport {
		case MCPTransportStdio:
			args := server.Args
			if args == nil {
				args = []string{}
			}
			table[server.Name] = map[string]any{
				"args":    args,
				"command": server.Command,
				"type":    "stdio",
			}
		case MCPTransportHTTP:
			table[server.Name] = map[string]any{
				"type": "http",
				"url":  server.URL,
			}
		}
	}
	document, err := protocoljson.MarshalCanonical(map[string]any{"mcpServers": table})
	if err != nil {
		return nil, err
	}
	return append(document, '\n'), nil
}

// openCodeMCP renders the CCJ-1 bytes of the object whose single member mcp
// maps each server name to its shape, followed by exactly one LF
// (environments §5.8).
func openCodeMCP(set []MCPServer) ([]byte, error) {
	table := map[string]any{}
	for _, server := range set {
		switch server.Transport {
		case MCPTransportStdio:
			command := []string{server.Command}
			command = append(command, server.Args...)
			table[server.Name] = map[string]any{
				"command": command,
				"type":    "local",
			}
		case MCPTransportHTTP:
			table[server.Name] = map[string]any{
				"type": "remote",
				"url":  server.URL,
			}
		}
	}
	document, err := protocoljson.MarshalCanonical(map[string]any{"mcp": table})
	if err != nil {
		return nil, err
	}
	return append(document, '\n'), nil
}

// codexMCP renders the TOML layer document whose only table is mcp_servers,
// one [mcp_servers.<name>] table per server in sorted name order, keys in
// command/args (stdio) or url (http) order, one key per line, LF line
// endings, exactly one trailing LF, and no other bytes (environments §5.8).
// <name> is emitted as a TOML bare key, which the core §2 identifier
// grammar guarantees needs no quoting.
func codexMCP(set []MCPServer) ([]byte, error) {
	var out strings.Builder
	for _, server := range set {
		fmt.Fprintf(&out, "[mcp_servers.%s]\n", server.Name)
		switch server.Transport {
		case MCPTransportStdio:
			fmt.Fprintf(&out, "command = %s\n", tomlString(server.Command))
			fmt.Fprintf(&out, "args = %s\n", tomlStringList(server.Args))
		case MCPTransportHTTP:
			fmt.Fprintf(&out, "url = %s\n", tomlString(server.URL))
		}
	}
	return []byte(out.String()), nil
}

// tomlString renders a TOML basic string.
func tomlString(value string) string {
	var out strings.Builder
	out.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\b':
			out.WriteString(`\b`)
		case '\t':
			out.WriteString(`\t`)
		case '\n':
			out.WriteString(`\n`)
		case '\f':
			out.WriteString(`\f`)
		case '\r':
			out.WriteString(`\r`)
		default:
			if character < 0x20 || character == 0x7f {
				fmt.Fprintf(&out, `\u%04X`, character)
			} else {
				out.WriteRune(character)
			}
		}
	}
	out.WriteByte('"')
	return out.String()
}

// tomlStringList renders the args list: elements as TOML basic strings
// separated by exactly ", ", the empty list as [].
func tomlStringList(args []string) string {
	if len(args) == 0 {
		return "[]"
	}
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, tomlString(arg))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
