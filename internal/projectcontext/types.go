// Package projectcontext owns CIP-0002 project registrations and inert item
// admission. It does not compose profiles, create leases, or launch providers.
// Admitted root context is additive: the composition consumer must append it
// after the selected profile document, never replace that document.
package projectcontext

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
)

const (
	MaxRecord          = 1 << 20
	MaxFiles           = 10000
	MaxFileBytes       = 16 << 20
	MaxTotalBytes      = 64 << 20
	AppendAfterProfile = "append-after-profile"
)

var idRE = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
var hexRE = regexp.MustCompile(`^[0-9a-f]{64}$`)

func environment(s string) bool      { return s == "claude_code" || s == "codex_cli" || s == "muse" }
func fail(code, detail string) error { return fmt.Errorf("%s: %s", code, detail) }

// Profile is the selected immutable base profile, not a replacement profile.
type Profile struct {
	Name       string `json:"name"`
	LockSHA256 string `json:"lock_sha256"`
}
type Registration struct {
	Kind         string  `json:"kind"`
	CheckoutID   string  `json:"checkout_id"`
	ProjectRoot  string  `json:"project_root"`
	RootIdentity string  `json:"root_identity"`
	Profile      Profile `json:"profile"`
	Environment  string  `json:"environment"`
	HomeRef      string  `json:"home_ref"`
	WorktreeRef  string  `json:"worktree_ref"`
}
type Item struct {
	Kind        string  `json:"kind"`
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Environment string  `json:"environment"`
	Required    bool    `json:"required"`
	Payload     Payload `json:"payload"`
}

// Payload is the closed context/skill arm. Inline MCP has no admitted arm in
// this stage; discovery reports it as unsupported until CIP-0022 stage B.
type Payload struct {
	SnapshotRef    string `json:"snapshot_ref"`
	SnapshotSHA256 string `json:"snapshot_sha256"`
	LockRef        string `json:"lock_ref,omitempty"`
	LockSHA256     string `json:"lock_sha256,omitempty"`
}
type Approval struct {
	Kind          string `json:"kind"`
	CheckoutID    string `json:"checkout_id"`
	ItemID        string `json:"item_id"`
	ItemSHA256    string `json:"item_sha256"`
	Environment   string `json:"environment"`
	ApproverRef   string `json:"approver_ref"`
	CeilingSHA256 string `json:"ceiling_sha256"`
}
type File struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
	Mode   string `json:"mode"`
}
type Snapshot struct {
	Version int    `json:"version"`
	Files   []File `json:"files"`
}

// Declaration is an explicit manager-operation request, never authority. Paths
// are relative to the registered root; dependencies name local Git repositories
// used only as offline object sources for the owning context resolver.
type Declaration struct {
	Version int       `json:"version"`
	Items   []Request `json:"items"`
}
type Request struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Environment  string            `json:"environment"`
	Required     bool              `json:"required"`
	Source       string            `json:"source"`
	Dependencies map[string]string `json:"dependencies"`
	Packaging    string            `json:"packaging,omitempty"`
}
type Candidate struct {
	Item         Item     `json:"item"`
	ItemSHA256   string   `json:"item_sha256"`
	Request      Request  `json:"request"`
	SourceSHA256 string   `json:"source_sha256"`
	Members      []Member `json:"members"`
	Rendering    string   `json:"rendering"`
}
type Member struct {
	Name           string `json:"name"`
	SnapshotRef    string `json:"snapshot_ref"`
	SnapshotSHA256 string `json:"snapshot_sha256"`
	Pin            string `json:"pin"`
}
type FileChange struct {
	Path   string `json:"path"`
	Effect string `json:"effect"`
	Before string `json:"before"`
	After  string `json:"after"`
}
type Change struct {
	ItemID string       `json:"item_id"`
	Effect string       `json:"effect"`
	Before string       `json:"before"`
	After  string       `json:"after"`
	Files  []FileChange `json:"files"`
}
type Review struct {
	Version            int         `json:"version"`
	RegistrationRef    string      `json:"registration_ref"`
	RegistrationSHA256 string      `json:"registration_sha256"`
	Items              []Candidate `json:"items"`
	Changes            []Change    `json:"changes"`
	Omitted            []Omission  `json:"omitted"`
}
type Omission struct {
	ItemID       string `json:"item_id"`
	Reason       string `json:"reason"`
	Source       string `json:"source,omitempty"`
	SourceSHA256 string `json:"source_sha256,omitempty"`
}
type Preview struct {
	Review        Review `json:"review"`
	PackageSHA256 string `json:"package_sha256"`
}
type Discovery struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	SHA256 string `json:"sha256"`
	Status string `json:"status"`
}
type Inspection struct {
	Registration Registration `json:"registration"`
	Mode         string       `json:"mode"`
	Memory       string       `json:"memory"`
	Rendering    string       `json:"rendering"`
	Candidates   []Discovery  `json:"candidates"`
}

func canonical(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var x any
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if err = d.Decode(&x); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(x)
}
func digest(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func identity(v any) string {
	b, err := canonical(v)
	if err != nil {
		panic(err)
	}
	return digest(b)
}
func decode(b []byte, v any) error {
	if len(b) > MaxRecord {
		return fail("project_context_invalid", "record exceeds 1 MiB")
	}
	if err := protocoljson.Validate(b); err != nil {
		return fail("project_context_invalid", err.Error())
	}
	// Bound nesting independently of the destination's type.
	depth := 0
	d := json.NewDecoder(bytes.NewReader(b))
	for {
		tok, e := d.Token()
		if e != nil {
			break
		}
		if t, ok := tok.(json.Delim); ok {
			if t == '{' || t == '[' {
				depth++
				if depth > 16 {
					return fail("project_context_invalid", "record depth exceeds 16")
				}
			} else {
				depth--
			}
		}
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return fail("project_context_invalid", err.Error())
	}
	return nil
}
func ParseDeclaration(b []byte) (Declaration, error) {
	var d Declaration
	if err := decode(b, &d); err != nil {
		return d, err
	}
	// Required booleans must be present, never null/absent (Go zero values
	// alone cannot enforce the closed reader contract).
	var raw struct {
		Version json.RawMessage              `json:"version"`
		Items   []map[string]json.RawMessage `json:"items"`
	}
	_ = json.Unmarshal(b, &raw)
	if d.Version != 1 || d.Items == nil || len(d.Items) > 256 {
		return d, fail("project_context_invalid", "expected version 1 and at most 256 explicit items")
	}
	seen := map[string]bool{}
	for i, r := range d.Items {
		if _, present := raw.Items[i]["packaging"]; present {
			return d, fail("project_context_invalid", "packaging is an explicit manager operation")
		}
		if !idRE.MatchString(r.ID) || seen[r.ID] || !environment(r.Environment) || (r.Type != "context" && r.Type != "skill" && r.Type != "mcp") {
			return d, fail("project_context_invalid", "invalid or duplicate item identity/type/environment")
		}
		seen[r.ID] = true
		for _, k := range []string{"id", "type", "environment", "required", "source", "dependencies"} {
			v, ok := raw.Items[i][k]
			if !ok || bytes.Equal(v, []byte("null")) {
				return d, fail("project_context_invalid", "missing or null item member: "+k)
			}
		}
		if !relative(r.Source) || len(r.Dependencies) > 256 {
			return d, fail("project_context_invalid", "source must be one bounded relative path")
		}
		for k, v := range r.Dependencies {
			if !idRE.MatchString(k) || !relative(v) {
				return d, fail("project_context_invalid", "invalid dependency source")
			}
		}
	}
	return d, nil
}
func relative(p string) bool {
	if !identifiers.PortablePath(p) || len(p) > 240 || strings.ContainsAny(p, "\\:*?[]\x00") || strings.HasPrefix(p, "/") {
		return false
	}
	for _, c := range strings.Split(p, "/") {
		if c == "" || c == "." || c == ".." || strings.HasSuffix(c, ".") || strings.HasSuffix(c, " ") {
			return false
		}
	}
	return len(strings.Split(p, "/")) <= 32
}

// ReadDeclaration reads an explicit operator request as bounded inert data.
// It refuses links and non-regular files before opening the candidate.
func ReadDeclaration(path string) (Declaration, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Declaration{}, err
	}
	root, err := os.OpenRoot(filepath.Dir(abs))
	if err != nil {
		return Declaration{}, err
	}
	defer root.Close()
	b, _, err := read(root, filepath.Base(abs), MaxRecord)
	if err != nil {
		return Declaration{}, err
	}
	return ParseDeclaration(b)
}
