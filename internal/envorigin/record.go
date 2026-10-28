// Package envorigin reads private provisioning-origin/1 authority. It never
// reads credential payloads, repairs a home, or derives history from a link.
package envorigin

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/relux-works/curator/internal/pathboundary"
	"github.com/relux-works/curator/internal/protocoljson"
)

// Name and the associated constants define the closed private origin carrier.
const (
	Name           = ".agent-provisioning-origin.json"
	Version        = "provisioning-origin/1"
	DiagRequired   = "environment_provisioning_origin_required"
	DiagInvalid    = "environment_provisioning_origin_invalid"
	MaxRecordBytes = 64 << 10
)

var profilePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
var lockPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// Selector preserves original environment-variable presence independently of its value.
type Selector struct {
	Name  string  `json:"name"`
	State string  `json:"state"`
	Value *string `json:"value,omitempty"`
}

// Capture records how original provisioning authority was established.
type Capture struct {
	Method    string `json:"method"`
	Authority string `json:"authority"`
	Time      string `json:"time"`
}

// Binding ties authority to one installed profile, lock and managed target.
type Binding struct {
	Profile       string `json:"profile"`
	LockSHA256    string `json:"lock_sha256"`
	ManagedTarget string `json:"managed_target"`
}

// Record is the private immutable provisioning-origin capture and current binding.
type Record struct {
	Record       string   `json:"record"`
	Adapter      string   `json:"adapter"`
	OperatorHome string   `json:"operator_home"`
	Selector     Selector `json:"selector"`
	NativeRoot   string   `json:"native_root"`
	Capture      Capture  `json:"capture"`
	Binding      Binding  `json:"binding"`
}

func invalid(format string, args ...any) error { return fmt.Errorf(DiagInvalid+": "+format, args...) }
func cleanAbsolute(s string) bool {
	return s != "" && filepath.IsAbs(s) && filepath.Clean(s) == s && !strings.ContainsAny(s, "\x00\r\n") && !hasControl(s)
}
func hasControl(s string) bool {
	for _, r := range s {
		if r < 32 || r == 127 {
			return true
		}
	}
	return false
}
func selectorName(id string) string {
	switch id {
	case "codex_cli":
		return "CODEX_HOME"
	case "claude_code":
		return "CLAUDE_CONFIG_DIR"
	case "pi":
		return "PI_CODING_AGENT_DIR"
	case "opencode", "muse":
		return "XDG_CONFIG_HOME"
	}
	return ""
}

// ManagedTarget returns the fixed registry namespace; an arbitrary supplied path
// can never widen a binding or adoption's scope.
func ManagedTarget(manager, profile, adapter string) string {
	p := ManagedRoot(manager, profile, adapter)
	if adapter == "opencode" {
		p = filepath.Join(p, "opencode")
	}
	return p
}

// ManagedRoot is the whole owned adapter directory. OpenCode's marker/origin
// target is its child; XDG seed siblings remain in this parent custody unit.
func ManagedRoot(manager, profile, adapter string) string {
	return filepath.Join(manager, "environments", profile, adapter)
}

func (r *Record) validate() error {
	if r == nil || r.Record != Version || selectorName(r.Adapter) == "" {
		return invalid("unsupported or missing record/adapter")
	}
	if !cleanAbsolute(r.OperatorHome) || !cleanAbsolute(r.NativeRoot) || !cleanAbsolute(r.Binding.ManagedTarget) {
		return invalid("origin paths must be clean absolute paths")
	}
	if !profilePattern.MatchString(r.Binding.Profile) || !lockPattern.MatchString(r.Binding.LockSHA256) {
		return invalid("invalid profile or canonical lock binding")
	}
	if r.Selector.Name != selectorName(r.Adapter) {
		return invalid("selector does not match adapter")
	}
	switch r.Selector.State {
	case "absent":
		if r.Selector.Value != nil {
			return invalid("absent selector carries a value")
		}
	case "present-empty":
		if r.Selector.Value == nil || *r.Selector.Value != "" {
			return invalid("empty selector requires an explicit empty value")
		}
	case "present-value":
		if r.Selector.Value == nil || *r.Selector.Value == "" || hasControl(*r.Selector.Value) {
			return invalid("value selector requires a nonempty control-free value")
		}
	default:
		return invalid("unknown selector state")
	}
	if !(r.Capture.Method == "provisioning" && r.Capture.Authority == "native-selection" || r.Capture.Method == "adoption" && r.Capture.Authority == "explicit-operator-input") {
		return invalid("invalid capture provenance")
	}
	if _, err := time.Parse(time.RFC3339Nano, r.Capture.Time); err != nil || !strings.HasSuffix(r.Capture.Time, "Z") {
		return invalid("capture time must be UTC RFC3339")
	}
	return nil
}

// Marshal serializes a new capture only. Rebinding must use Rebind to retain
// original capture byte slices rather than reconstructing their values.
func (r *Record) Marshal() ([]byte, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	return json.Marshal(r)
}

// Parse enforces the closed transport/semantic schema and exact live binding.
// Physical origin inspection is metadata-only and never creates missing tails.
func Parse(raw []byte, want Binding, adapter, manager string) (*Record, error) {
	r, err := ParseCaptured(raw, want, adapter, manager)
	if err != nil {
		return nil, err
	}
	if err := ValidateExternal(manager, r.OperatorHome); err != nil {
		return nil, err
	}
	if err := ValidateExternal(manager, r.NativeRoot); err != nil {
		return nil, err
	}
	return r, nil
}

// ParseCaptured verifies saved grammar and bindings without a fresh origin
// route lookup. It is not forward authority. Exact rollback/cleanup must be
// able to inspect saved facts even if native A has become inaccessible.
func ParseCaptured(raw []byte, want Binding, adapter, manager string) (*Record, error) {
	r, err := decode(raw)
	if err != nil {
		return nil, err
	}
	if !cleanAbsolute(manager) || want.ManagedTarget != ManagedTarget(manager, want.Profile, adapter) || r.Adapter != adapter || r.Binding != want {
		return nil, invalid("adapter/profile/lock/target binding mismatch")
	}
	return r, nil
}
func decode(raw []byte) (*Record, error) {
	if len(raw) == 0 || len(raw) > MaxRecordBytes {
		return nil, invalid("invalid record size")
	}
	if err := protocoljson.Validate(raw); err != nil {
		return nil, invalid("closed JSON: %v", err)
	}
	var r Record
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&r); err != nil {
		return nil, invalid("closed record: %v", err)
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil || !exactMembers(members, "record", "adapter", "operator_home", "selector", "native_root", "capture", "binding") {
		return nil, invalid("missing record members")
	}
	for key, names := range map[string][]string{"capture": {"method", "authority", "time"}, "binding": {"profile", "lock_sha256", "managed_target"}} {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(members[key], &nested); err != nil || !exactMembers(nested, names...) {
			return nil, invalid("missing %s members", key)
		}
	}
	var sel map[string]json.RawMessage
	if err := json.Unmarshal(members["selector"], &sel); err != nil {
		return nil, invalid("invalid selector")
	}
	value, exists := sel["value"]
	if r.Selector.State == "absent" {
		if exists || !exactMembers(sel, "name", "state") {
			return nil, invalid("absent selector must omit value")
		}
	} else if !exists || bytes.Equal(bytes.TrimSpace(value), []byte("null")) || !exactMembers(sel, "name", "state", "value") {
		return nil, invalid("selector value presence is required")
	}
	if err := r.validate(); err != nil {
		return nil, err
	}
	return &r, nil
}

// encoding/json matches field names case-insensitively; authority grammar does
// not. Exact membership also rejects null objects and missing zero-value fields.
func exactMembers(m map[string]json.RawMessage, names ...string) bool {
	if len(m) != len(names) {
		return false
	}
	for _, name := range names {
		if _, ok := m[name]; !ok {
			return false
		}
	}
	return true
}

// Read pins the exact private no-follow file across open/read/stat. The caller
// must re-read under its mutation lock and before each admitted publication.
func Read(manager string, want Binding, adapter string) (*Record, []byte, error) {
	path := filepath.Join(want.ManagedTarget, Name)
	if !cleanAbsolute(manager) || !profilePattern.MatchString(want.Profile) || !lockPattern.MatchString(want.LockSHA256) || selectorName(adapter) == "" || want.ManagedTarget != ManagedTarget(manager, want.Profile, adapter) {
		return nil, nil, invalid("unexpected target namespace")
	}
	raw, present, err := ReadPrivateAuthority(manager, path, MaxRecordBytes)
	if err != nil {
		return nil, nil, err
	}
	if !present {
		return nil, nil, fmt.Errorf("%s: legacy managed home has no verified provisioning origin; use env adopt with verified original anchor, selector state/value, native root and exact installed binding; ordinary resolve/repair cannot infer these facts", DiagRequired)
	}
	r, err := Parse(raw, want, adapter, manager)
	return r, raw, err
}

// ReadPrivateAuthority reads only a known manager-owned authority file. It
// never follows a link or turns a parent/permission/read error into absence.
// Lifecycle carriers reuse this transport, with their own closed decoder.
func ReadPrivateAuthority(manager, path string, limit int64) ([]byte, bool, error) {
	return readOwnedRecord(manager, path, limit, true)
}

// ReadOwnedRecord preserves compatibility with readable installed profile
// records. They are not private origin/lifecycle carriers: other identities
// may read them, but may not mutate them or supply a link/fork as authority.
func ReadOwnedRecord(manager, path string, limit int64) ([]byte, bool, error) {
	return readOwnedRecord(manager, path, limit, false)
}

func readOwnedRecord(manager, path string, limit int64, private bool) (data []byte, exists bool, resultErr error) {
	if !cleanAbsolute(manager) || !cleanAbsolute(path) || !below(manager, path) || path == manager || limit <= 0 || limit > 16<<20 {
		return nil, false, invalid("private authority namespace/size")
	}
	parentPath := filepath.Dir(path)
	if err := pathboundary.ValidateRouteWithOwner(manager, parentPath, pathboundary.DefaultOwnerLookup()); err != nil {
		if errors.Is(err, os.ErrNotExist) && absentAuthorityParent(manager, parentPath) {
			return nil, false, nil
		}
		return nil, false, invalid("authority parent boundary: %v", err)
	}
	if parent, err := os.Lstat(parentPath); err != nil || !parent.IsDir() {
		return nil, false, invalid("authority parent is not a directory")
	}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, invalid("authority metadata: %v", err)
	}
	if !info.Mode().IsRegular() || info.Size() > limit || !singleLink(path, info) || (private && !privateMode(path, info)) {
		if private {
			return nil, false, invalid("authority must be a private regular single-link file")
		}
		return nil, false, invalid("installed record must be a regular single-link file")
	}
	if err := pathboundary.ValidateRouteWithOwner(manager, path, pathboundary.DefaultOwnerLookup()); err != nil {
		return nil, false, invalid("authority parent/owner boundary: %v", err)
	}
	f, err := pathboundary.OpenReadNoFollow(path)
	if err != nil {
		return nil, false, invalid("authority open: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			data = nil
			exists = false
			resultErr = errors.Join(resultErr, err)
		}
	}()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) || !singleLink(path, opened) || (private && !privateMode(path, opened)) {
		return nil, false, invalid("authority entry changed while opening")
	}
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, false, invalid("authority read: %v", err)
	}
	after, err := os.Lstat(path)
	if err != nil || !os.SameFile(info, after) || info.Size() != after.Size() || info.ModTime() != after.ModTime() || !singleLink(path, after) || (private && !privateMode(path, after)) {
		return nil, false, invalid("authority changed while reading")
	}
	if err := pathboundary.ValidateRouteWithOwner(manager, path, pathboundary.DefaultOwnerLookup()); err != nil {
		return nil, false, invalid("authority boundary changed while reading: %v", err)
	}
	if int64(len(raw)) > limit || int64(len(raw)) != info.Size() {
		return nil, false, invalid("authority read size changed")
	}
	return raw, true, nil
}

// An ENOENT wrapped by a boundary check may be resolution of a broken link,
// rather than a missing entry. Walk lexical entries without following links;
// only a genuinely missing component after validated directory prefixes can
// authorize absence. An unreadable, linked or wrong-kind prefix is invalid.
func absentAuthorityParent(manager, parent string) bool {
	rel, err := filepath.Rel(manager, parent)
	if err != nil {
		return false
	}
	paths := []string{manager}
	current := manager
	if rel != "." {
		for _, part := range strings.Split(rel, string(filepath.Separator)) {
			current = filepath.Join(current, part)
			paths = append(paths, current)
		}
	}
	for _, path := range paths {
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			return true
		}
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return false
		}
		if pathboundary.ValidateRouteWithOwner(manager, path, pathboundary.DefaultOwnerLookup()) != nil {
			return false
		}
	}
	return false
}

// Digest identifies only noncredential authority bytes for operation/cache pins.
func Digest(raw []byte) string { h := sha256.Sum256(raw); return "sha256:" + hex.EncodeToString(h[:]) }

// CaptureFacts preserves the exact original JSON value slices in a stable
// order. It is private comparison evidence, never portable profile material.
func CaptureFacts(raw []byte) ([]byte, error) {
	if _, err := decode(raw); err != nil {
		return nil, err
	}
	var m map[string]json.RawMessage
	_ = json.Unmarshal(raw, &m)
	var out []byte
	for _, k := range []string{"operator_home", "selector", "native_root", "capture"} {
		out = append(out, m[k]...)
		out = append(out, 0)
	}
	return out, nil
}

// Rebind replaces only binding's JSON value; every capture byte is unchanged.
func Rebind(raw []byte, old, next Binding, adapter, manager string) ([]byte, error) {
	if _, err := Parse(raw, old, adapter, manager); err != nil {
		return nil, err
	}
	if old.Profile != next.Profile || old.ManagedTarget != next.ManagedTarget || !lockPattern.MatchString(next.LockSHA256) {
		return nil, invalid("rebind may change only the installed lock")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, err := d.Token(); err != nil {
		return nil, err
	}
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return nil, err
		}
		start := int(d.InputOffset())
		var v json.RawMessage
		if err := d.Decode(&v); err != nil {
			return nil, err
		}
		end := int(d.InputOffset())
		if token == "binding" {
			for start < end && (raw[start] == ':' || raw[start] == ' ' || raw[start] == '\n' || raw[start] == '\t' || raw[start] == '\r') {
				start++
			}
			b, _ := json.Marshal(next)
			out := append([]byte{}, raw[:start]...)
			out = append(out, b...)
			out = append(out, raw[end:]...)
			if _, err := Parse(out, next, adapter, manager); err != nil {
				return nil, err
			}
			return out, nil
		}
	}
	return nil, invalid("binding absent")
}
