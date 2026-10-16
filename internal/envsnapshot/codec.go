// Package envsnapshot implements the closed snapshot metadata and read-only
// actual-tree identity boundary. Parsing metadata never verifies a directory.
package envsnapshot

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"reflect"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Snapshot and descriptor wire versions are independent of launch fragments.
const (
	ManifestVersion   = "managed-home-snapshot-v1"
	DescriptorVersion = "snapshot-launch-v1"
	// Mismatch is deliberately content-free, including for failed filesystem reads.
	Mismatch = "environment_materialize_snapshot_mismatch"
)

// Error exposes a stable refusal code and a bounded reason, never raw input,
// file contents, local paths, or operating-system error strings.
type Error struct{ Code, Reason string }

func (e *Error) Error() string   { return e.Code + ": " + e.Reason }
func refuse(reason string) error { return &Error{Mismatch, reason} }

// Profile binds the selected profile name and verified lock digest.
type Profile struct {
	Name       string `json:"name"`
	LockSHA256 string `json:"lock_sha256"`
}

// File describes one regular content file below home/.
type File struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
	Mode   string `json:"mode"`
}

// Manifest is the closed managed-home-snapshot-v1 inventory and provenance.
type Manifest struct {
	Snapshot       string  `json:"snapshot"`
	Adapter        string  `json:"adapter"`
	Profile        Profile `json:"profile"`
	CuratorVersion string  `json:"curator_version"`
	FragmentSHA256 string  `json:"fragment_sha256"`
	LaunchSHA256   string  `json:"launch_sha256"`
	Files          []File  `json:"files"`
}

// Precedence retains the source fragment precedence policy.
type Precedence struct {
	Winner    string `json:"winner"`
	Placement string `json:"placement"`
}

// Permissions retains the source fragment permission policy.
type Permissions struct {
	Mode   string `json:"mode"`
	Locked bool   `json:"locked"`
	Source string `json:"source"`
}

// Channel is one closed adapter-defined launch channel.
type Channel struct {
	Kind      string   `json:"kind"`
	Semantics string   `json:"semantics,omitempty"`
	Flag      string   `json:"flag,omitempty"`
	Argument  string   `json:"argument,omitempty"`
	Name      string   `json:"name,omitempty"`
	With      []string `json:"with,omitempty"`
	Key       string   `json:"key,omitempty"`
	Variable  string   `json:"variable,omitempty"`
	Filename  string   `json:"filename,omitempty"`
}

// SystemPrompt references inert prompt content and its available channels.
type SystemPrompt struct {
	Path     string    `json:"path"`
	Channels []Channel `json:"channels"`
}

// MCP references the portable MCP layer and variable names, never values.
type MCP struct {
	Path     string    `json:"path"`
	EnvNames []string  `json:"env_names"`
	Channels []Channel `json:"channels"`
}

// RootContext carries the explicit Muse exclusive-context fact.
type RootContext struct {
	Exclusive bool `json:"exclusive"`
}

// Descriptor is a portable snapshot-launch-v1 object, not a launch fragment.
type Descriptor struct {
	Descriptor   string            `json:"descriptor"`
	Environment  string            `json:"environment"`
	Profile      Profile           `json:"profile"`
	Precedence   Precedence        `json:"precedence"`
	Env          map[string]string `json:"env"`
	Permissions  Permissions       `json:"permissions"`
	SystemPrompt *SystemPrompt     `json:"system_prompt,omitempty"`
	MCP          *MCP              `json:"mcp,omitempty"`
	RootContext  *RootContext      `json:"root_context,omitempty"`
}

// The exact published schemas, including their referenced definitions, are
// embedded: validation never downloads schemas or consults ambient state.
//
//go:embed schemas/*/*
var schemas embed.FS
var validators = sync.OnceValues(func() (map[string]*jsonschema.Schema, error) {
	c := jsonschema.NewCompiler()
	c.UseRegexpEngine(compilePattern)
	c.UseLoader(closedLoader{})
	paths, err := fs.Glob(schemas, "schemas/*/*.json")
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		raw, err := schemas.ReadFile(p)
		if err != nil {
			return nil, err
		}
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		if err := c.AddResource(doc["$id"].(string), doc); err != nil {
			return nil, err
		}
	}
	base := "https://relux-works.github.io/curator-spec/schemas/managed-home-snapshot-v1/"
	m, err := c.Compile(base + "managed-home-snapshot-v1.schema.json")
	if err != nil {
		return nil, err
	}
	d, err := c.Compile(base + "snapshot-launch-v1.schema.json")
	return map[string]*jsonschema.Schema{"manifest": m, "descriptor": d}, err
})

func decode(raw []byte, manifest bool, dst any) error {
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return refuse("metadata-canonical")
	}
	payload := raw[:len(raw)-1]
	if err := protocoljson.UnmarshalCanonical(payload, dst); err != nil {
		return refuse("metadata-canonical")
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return refuse("metadata-canonical")
	}
	checks, err := validators()
	if err != nil {
		return refuse("schema-unavailable")
	}
	schema := checks["descriptor"]
	reason := "launch-schema"
	if manifest {
		schema = checks["manifest"]
		reason = "manifest-schema"
	}
	if err := schema.Validate(value); err != nil {
		return refuse(reason)
	}
	return nil
}

// DecodeManifest validates closed schema and exact CCJ-1 plus one LF framing.
// Cross-file and actual-tree checks are exclusively the job of Verify.
func DecodeManifest(raw []byte) (Manifest, error) {
	var m Manifest
	err := decode(raw, true, &m)
	return m, err
}

// DecodeDescriptor validates the closed descriptor schema and exact wire frame.
func DecodeDescriptor(raw []byte) (Descriptor, error) {
	var d Descriptor
	err := decode(raw, false, &d)
	return d, err
}

func encode(value any, manifest bool) ([]byte, error) {
	if !validStrings(reflect.ValueOf(value)) {
		return nil, refuse("metadata-canonical")
	}
	// Convert typed structs to the JSON domain without losing integer precision.
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, refuse("metadata-schema")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var object any
	if err := decoder.Decode(&object); err != nil {
		return nil, refuse("metadata-schema")
	}
	canonical, err := protocoljson.MarshalCanonical(object)
	if err != nil {
		return nil, refuse("metadata-canonical")
	}
	canonical = append(canonical, '\n')
	if manifest {
		_, err = DecodeManifest(canonical)
	} else {
		_, err = DecodeDescriptor(canonical)
	}
	if err != nil {
		return nil, err
	}
	return canonical, nil
}

// EncodeManifest emits validated CCJ-1 followed by exactly one LF.
func EncodeManifest(m Manifest) ([]byte, error) { return encode(m, true) }

// EncodeDescriptor emits validated CCJ-1 followed by exactly one LF.
func EncodeDescriptor(d Descriptor) ([]byte, error) { return encode(d, false) }

// Digest hashes exact bytes; manifest and descriptor callers include their LF.
func Digest(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }

// FragmentDigest requires source CCJ-1 WITHOUT its presentation LF. Passing a
// rendered fragment with LF is refused rather than hashed in the wrong domain.
func FragmentDigest(canonical []byte) (string, error) {
	if err := protocoljson.RequireCanonical(canonical); err != nil {
		return "", refuse("fragment-canonical")
	}
	return Digest(canonical), nil
}

type closedLoader struct{}

func (closedLoader) Load(string) (any, error) { return nil, errors.New("unbundled schema") }

type schemaPattern struct{ re *regexp2.Regexp }

func (p schemaPattern) MatchString(s string) bool {
	matched, err := p.re.MatchString(s)
	return err == nil && matched
}
func (p schemaPattern) String() string { return p.re.String() }
func compilePattern(s string) (jsonschema.Regexp, error) {
	re, err := regexp2.Compile(s, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	re.MatchTimeout = time.Second
	return schemaPattern{re}, nil
}

// encoding/json replaces invalid UTF-8. Refuse it before converting the typed
// model to the JSON domain so encoding cannot silently change path identity.
func validStrings(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return utf8.ValidString(v.String())
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !validStrings(v.Field(i)) {
				return false
			}
		}
	case reflect.Pointer, reflect.Interface:
		return v.IsNil() || validStrings(v.Elem())
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !validStrings(v.Index(i)) {
				return false
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			if !validStrings(iter.Key()) || !validStrings(iter.Value()) {
				return false
			}
		}
	}
	return true
}
