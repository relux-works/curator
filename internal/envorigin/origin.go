// Package envorigin validates private provisioning-origin/1 records. It owns
// no native selection or mutation: callers must validate physical boundaries,
// exact current binding and transaction authority before using a record.
package envorigin

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	Name         = ".agent-provisioning-origin.json"
	Version      = "provisioning-origin/1"
	DiagRequired = "environment_provisioning_origin_required"
	DiagInvalid  = "environment_provisioning_origin_invalid"
)

// Selector preserves absence separately from an explicitly empty value.
type Selector struct {
	Name  string  `json:"name"`
	State string  `json:"state"`
	Value *string `json:"value,omitempty"`
}

// Capture identifies the authority that supplied the original selection.
type Capture struct {
	Method    string `json:"method"`
	Authority string `json:"authority"`
	Time      string `json:"time"`
}

// Binding is exact local authority, never a prefix or caller-derived default.
type Binding struct {
	Profile       string `json:"profile"`
	LockSHA256    string `json:"lock_sha256"`
	ManagedTarget string `json:"managed_target"`
}

// Record is the full separate carrier; published home markers stay unchanged.
type Record struct {
	Record       string   `json:"record"`
	Adapter      string   `json:"adapter"`
	OperatorHome string   `json:"operator_home"`
	Selector     Selector `json:"selector"`
	NativeRoot   string   `json:"native_root"`
	Capture      Capture  `json:"capture"`
	Binding      Binding  `json:"binding"`
}

//go:embed schemas/provisioning-origin-v1/*.json
var schemas embed.FS

var compiled = sync.OnceValues(func() (*jsonschema.Schema, error) {
	raw, err := schemas.ReadFile("schemas/provisioning-origin-v1/provisioning-origin-v1.schema.json")
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err = json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	c.UseLoader(noNetwork{})
	c.AssertFormat()
	c.UseRegexpEngine(compilePattern)
	id := doc["$id"].(string)
	if err = c.AddResource(id, doc); err != nil {
		return nil, err
	}
	return c.Compile(id)
})

type noNetwork struct{}

func (noNetwork) Load(string) (any, error) {
	return nil, fmt.Errorf("unbundled provisioning-origin schema")
}

type schemaPattern struct{ re *regexp2.Regexp }

func (r schemaPattern) String() string { return r.re.String() }
func (r schemaPattern) MatchString(s string) bool {
	ok, err := r.re.MatchString(s)
	return err == nil && ok
}
func compilePattern(pattern string) (jsonschema.Regexp, error) {
	re, err := regexp2.Compile(pattern, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	re.MatchTimeout = time.Second
	return schemaPattern{re}, nil
}

// Parse rejects unknown/duplicate members, unsupported revisions, malformed
// selector states and invalid bindings without echoing private record contents.
func Parse(raw []byte) (*Record, error) {
	invalid := func() (*Record, error) { return nil, fmt.Errorf("%s: invalid private record", DiagInvalid) }
	if err := protocoljson.Validate(raw); err != nil {
		return invalid()
	}
	schema, err := compiled()
	if err != nil {
		return invalid()
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err = decoder.Decode(&value); err != nil {
		return invalid()
	}
	if err = schema.Validate(value); err != nil {
		return invalid()
	}
	var record Record
	if err = json.Unmarshal(raw, &record); err != nil {
		return invalid()
	}
	adapter, err := envregistry.ByID(record.Adapter)
	if err != nil || adapter.EnvVar != record.Selector.Name || !identifiers.Valid(record.Binding.Profile) {
		return invalid()
	}
	for _, value := range []string{record.OperatorHome, record.NativeRoot, record.Binding.ManagedTarget} {
		if !cleanAbsolute(value) {
			return invalid()
		}
	}
	return &record, nil
}

// Marshal validates the complete carrier, rather than trusting constructed
// typed values that may omit required fields or violate cross-field constraints.
func (r *Record) Marshal() ([]byte, error) {
	raw, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	if _, err = Parse(raw); err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

// Digest scopes an operation-local cache to the full validated authority.
func (r *Record) Digest() (string, error) {
	raw, err := r.Marshal()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// Match refuses rebinding based on current caller state. Authorized profile
// update must explicitly commit the new binding with its home transaction.
func (r *Record) Match(adapter string, binding Binding) error {
	if r == nil {
		return fmt.Errorf("%s: explicit trusted-input env adopt required", DiagRequired)
	}
	if r.Adapter != adapter || r.Binding != binding {
		return fmt.Errorf("%s: authority binding mismatch", DiagInvalid)
	}
	return nil
}

var windowsVolume = regexp.MustCompile(`^[A-Za-z]:\\`)

func cleanAbsolute(value string) bool {
	if strings.HasPrefix(value, "/") {
		return !strings.HasPrefix(value, "//") && path.Clean(value) == value
	}
	// Parse cross-platform specification records without treating a Windows
	// drive path as a POSIX relative path; live admission remains host-specific.
	if !windowsVolume.MatchString(value) && !strings.HasPrefix(value, `\\`) {
		return false
	}
	normalized := strings.ReplaceAll(value, `\`, "/")
	for _, component := range strings.Split(normalized, "/") {
		if component == "." || component == ".." {
			return false
		}
	}
	if strings.Contains(normalized[2:], "//") {
		return false
	}
	return !strings.HasSuffix(normalized, "/") || windowsVolume.MatchString(value) && len(value) == 3
}
