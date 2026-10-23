// Package mcpnative implements inert, append-only first installs. It neither
// resolves packages nor issues ownership-action authority.
package mcpnative

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/projectmcp"
	"github.com/relux-works/curator/internal/protocoljson"
)

const Capability = "profile-mcp-native-writer-v1"
const AuthorityUnavailable = "mcp_native_writer_authority_unavailable"
const Conflict = "mcp_native_entry_conflict"
const Unreadable = "environment_surface_unreadable"

func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func Key(profile, environment, path string) string {
	data, _ := protocoljson.MarshalCanonical([]string{profile, environment, path})
	return strings.TrimPrefix(Digest(data), "sha256:")
}
func Path(environment string) string {
	switch environment {
	case "claude_code":
		return contextmaterialize.MCPClaudePath
	case "codex_cli":
		return contextmaterialize.MCPCodexPath
	}
	return ""
}

type Record struct {
	Name            string  `json:"name"`
	Owner           string  `json:"owner"`
	ServerSHA256    string  `json:"server_sha256"`
	ContentSHA256   string  `json:"content_sha256"`
	PreimageSHA256  *string `json:"preimage_sha256"`
	InstalledSHA256 string  `json:"installed_sha256"`
}
type Ownership struct {
	SchemaVersion int      `json:"schema_version"`
	Capability    string   `json:"capability"`
	Scope         string   `json:"scope"`
	Profile       string   `json:"profile_id"`
	Environment   string   `json:"environment"`
	Path          string   `json:"path"`
	Records       []Record `json:"records"`
}

func NewOwnership(profile, environment string) *Ownership {
	return &Ownership{1, Capability, "profile", profile, environment, Path(environment), []Record{}}
}
func (o *Ownership) Bytes() ([]byte, error) { return projectmcpCanonical(o) }
func projectmcpCanonical(value any) ([]byte, error) {
	b, e := json.Marshal(value)
	if e != nil {
		return nil, e
	}
	var v any
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if e = d.Decode(&v); e != nil {
		return nil, e
	}
	return protocoljson.MarshalCanonical(v)
}
func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, c := range value[7:] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func closed(data []byte, names ...string) bool {
	var v map[string]json.RawMessage
	if json.Unmarshal(data, &v) != nil || len(v) != len(names) {
		return false
	}
	for _, n := range names {
		if _, ok := v[n]; !ok {
			return false
		}
	}
	return true
}
func DecodeOwnership(data []byte, profile, environment string) (*Ownership, error) {
	bad := func() (*Ownership, error) { return nil, fmt.Errorf("%s: invalid protected MCP ownership", Unreadable) }
	if protocoljson.RequireCanonical(data) != nil || !closed(data, "schema_version", "capability", "scope", "profile_id", "environment", "path", "records") {
		return bad()
	}
	var o Ownership
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(&o) != nil {
		return bad()
	}
	if o.SchemaVersion != 1 || o.Capability != Capability || o.Scope != "profile" || o.Profile != profile || !identifiers.Valid(profile) || o.Environment != environment || o.Path == "" || o.Path != Path(environment) || o.Records == nil {
		return bad()
	}
	var raw struct{ Records []json.RawMessage }
	_ = json.Unmarshal(data, &raw)
	for i, r := range o.Records {
		if !closed(raw.Records[i], "name", "owner", "server_sha256", "content_sha256", "preimage_sha256", "installed_sha256") || !projectmcp.ValidName(r.Name) || r.Owner != "curator" || !validDigest(r.ServerSHA256) || !validDigest(r.ContentSHA256) || !validDigest(r.InstalledSHA256) || r.PreimageSHA256 != nil && !validDigest(*r.PreimageSHA256) || i > 0 && o.Records[i-1].Name >= r.Name {
			return bad()
		}
	}
	return &o, nil
}

// Entries returns exact selected-entry bytes. A nil span is an existing foreign
// inline TOML record; it is never absence and cannot be enrolled implicitly.
func Entries(environment string, data []byte) (map[string][]byte, error) {
	if len(data) == 0 {
		return map[string][]byte{}, nil
	}
	var entries map[string][]byte
	var err error
	switch environment {
	case "claude_code":
		entries, err = jsonEntries(data)
	case "codex_cli":
		entries, err = tomlEntries(data)
	default:
		err = fmt.Errorf("unsupported adapter")
	}
	if err != nil {
		return nil, fmt.Errorf("%s: native MCP document cannot be inspected", Unreadable)
	}
	return entries, nil
}

type member struct {
	name              string
	start, end, value int
}

func object(data []byte, start int) ([]member, int, error) {
	d := json.NewDecoder(bytes.NewReader(data[start:]))
	if tok, e := d.Token(); e != nil || tok != json.Delim('{') {
		return nil, 0, fmt.Errorf("not object")
	}
	var out []member
	for d.More() {
		begin := start + int(d.InputOffset())
		for begin < len(data) && (data[begin] == ' ' || data[begin] == '\n' || data[begin] == '\r' || data[begin] == '\t' || data[begin] == ',') {
			begin++
		}
		k, e := d.Token()
		if e != nil {
			return nil, 0, e
		}
		n, ok := k.(string)
		if !ok {
			return nil, 0, fmt.Errorf("key")
		}
		v := start + int(d.InputOffset())
		for v < len(data) && (data[v] == ':' || data[v] == ' ' || data[v] == '\n' || data[v] == '\r' || data[v] == '\t') {
			v++
		}
		var raw json.RawMessage
		if e = d.Decode(&raw); e != nil {
			return nil, 0, e
		}
		out = append(out, member{n, begin, start + int(d.InputOffset()), v})
	}
	end := start + int(d.InputOffset())
	for end < len(data) && data[end] != '}' {
		end++
	}
	if _, e := d.Token(); e != nil {
		return nil, 0, e
	}
	return out, end, nil
}
func jsonEntries(data []byte) (map[string][]byte, error) {
	if e := protocoljson.Validate(data); e != nil {
		return nil, e
	}
	members, _, e := object(data, 0)
	if e != nil {
		return nil, e
	}
	out := map[string][]byte{}
	for _, m := range members {
		if m.name == "mcpServers" {
			rows, _, e := object(data, m.value)
			if e != nil {
				return nil, e
			}
			for _, r := range rows {
				out[r.name] = data[r.value:r.end]
			}
		}
	}
	return out, nil
}

// TOML is validated by the native TOML parser first. The lexer only locates
// table boundaries outside strings, comments and arrays; it never rewrites them.
func tomlEntries(data []byte) (map[string][]byte, error) {
	var document map[string]any
	if _, e := toml.Decode(string(data), &document); e != nil {
		return nil, e
	}
	out := map[string][]byte{}
	if raw, ok := document["mcp_servers"]; ok {
		table, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("mcp table")
		}
		for n := range table {
			out[n] = nil
		}
	}
	type header struct {
		start int
		name  string
	}
	var headers []header
	quote := byte(0)
	triple := false
	escaped := false
	depth := 0
	comment := false
	line := true
	for i := 0; i < len(data); i++ {
		c := data[i]
		if comment {
			if c == '\n' {
				comment = false
				line = true
			}
			continue
		}
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if quote == '"' && c == '\\' {
				escaped = true
				continue
			}
			if c == quote {
				if triple {
					if i+2 < len(data) && data[i+1] == quote && data[i+2] == quote {
						// TOML permits one or two quote characters immediately
						// before the closing delimiter. Consume the entire run;
						// otherwise a fourth quote starts a phantom string.
						for i+1 < len(data) && data[i+1] == quote {
							i++
						}
						quote = 0
						triple = false
					}
				} else {
					quote = 0
				}
			}
			continue
		}
		if line && depth == 0 {
			j := i
			for j < len(data) && (data[j] == ' ' || data[j] == '\t') {
				j++
			}
			if j < len(data) && data[j] == '[' {
				end := bytes.IndexByte(data[j:], '\n')
				if end < 0 {
					end = len(data) - j
				}
				lineData := data[j : j+end]
				var probe map[string]any
				md, e := toml.Decode(string(lineData)+"\n__curator_probe=1\n", &probe)
				if e != nil || len(md.Keys()) == 0 {
					return nil, fmt.Errorf("ambiguous table")
				}
				key := md.Keys()[0]
				name := ""
				if len(key) >= 2 && key[0] == "mcp_servers" {
					name = key[1]
				}
				headers = append(headers, header{i, name})
				i = j + end - 1
				line = false
				continue
			}
			line = false
		}
		switch c {
		case '#':
			comment = true
		case '\n':
			line = true
		case '"', '\'':
			quote = c
			if i+2 < len(data) && data[i+1] == c && data[i+2] == c {
				triple = true
				i += 2
			}
		case '[', '{':
			depth++
		case ']', '}':
			depth--
		}
	}
	for i, h := range headers {
		if h.name == "" {
			continue
		}
		end := len(data)
		if i+1 < len(headers) {
			end = headers[i+1].start
		}
		// Trailing line separators belong to the container edit, not the
		// preceding record. This keeps spans stable when an EOF needs a newline.
		out[h.name] = append(out[h.name], bytes.TrimRight(data[h.start:end], "\r\n \t")...)
	}
	return out, nil
}

// Add only inserts a new NAME. Existing native entries, including equal ones,
// are always conflicts; no action-shaped data can turn this into replacement.
func Add(environment string, data []byte, server contextmaterialize.MCPServer) ([]byte, []byte, error) {
	entries, e := Entries(environment, data)
	if e != nil {
		return nil, nil, e
	}
	if _, exists := entries[server.Name]; exists {
		return nil, nil, fmt.Errorf("%s: foreign same-NAME entry", Conflict)
	}
	_, encoded, _, e := contextmaterialize.MCPFile(environment, []contextmaterialize.MCPServer{server})
	if e != nil {
		return nil, nil, e
	}
	if len(encoded) == 0 {
		return nil, nil, fmt.Errorf("unsupported adapter")
	}
	var result []byte
	if environment == "codex_cli" {
		for _, span := range entries {
			if span == nil {
				return nil, nil, fmt.Errorf("%s: inline or dotted native records have no certified table span", Unreadable)
			}
		}
		result = append([]byte{}, data...)
		if len(result) > 0 && result[len(result)-1] != '\n' {
			result = append(result, '\n')
		}
		result = append(result, encoded...)
	} else {
		if len(data) == 0 {
			data = []byte("{\"mcpServers\":{}}\n")
		}
		root, end, e := object(data, 0)
		if e != nil {
			return nil, nil, e
		}
		key := []byte{}
		var shape map[string]json.RawMessage
		_ = json.Unmarshal(encoded, &shape)
		rows, _, _ := object(shape["mcpServers"], 0)
		value := shape["mcpServers"][rows[0].start:rows[0].end]
		key = append(key, value...)
		found := false
		for _, m := range root {
			if m.name == "mcpServers" {
				found = true
				old, close, e := object(data, m.value)
				if e != nil {
					return nil, nil, e
				}
				if len(old) > 0 {
					key = append([]byte(","), key...)
				}
				result = append(result, data[:close]...)
				result = append(result, key...)
				result = append(result, data[close:]...)
				break
			}
		}
		if !found {
			insert := append([]byte("\"mcpServers\":{"), key...)
			insert = append(insert, '}')
			if len(root) > 0 {
				insert = append([]byte(","), insert...)
			}
			result = append(result, data[:end]...)
			result = append(result, insert...)
			result = append(result, data[end:]...)
		}
	}
	final, e := Entries(environment, result)
	if e != nil {
		return nil, nil, e
	}
	span := final[server.Name]
	if len(span) == 0 {
		return nil, nil, fmt.Errorf("%s: inserted native entry has no certified span", Unreadable)
	}
	return result, span, nil
}
func Sort(o *Ownership) {
	sort.Slice(o.Records, func(i, j int) bool { return o.Records[i].Name < o.Records[j].Name })
}
