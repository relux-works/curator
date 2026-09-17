package sourcelock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/protocoljson"
	"github.com/relux-works/curator/internal/verr"
)

// BindingsSchemaVersion is the only machine-bindings schema this reader accepts.
const BindingsSchemaVersion = 1

// SourceBinding is the machine-private record for one source alias: the
// canonical physical source location plus a snapshot of the admitted
// root-input configuration consumed at lock time. The operator-owned
// source-policy.json remains the configuration source of truth; this is
// the per-lock-generation record refresh compares against.
type SourceBinding struct {
	Location   string
	RootInputs []string
}

// Bindings records the machine side of a frozen selection, keyed by source
// alias and bound to one lock generation. It is never portable content:
// it must not be copied into locks, manifests, receipts, or markers.
type Bindings struct {
	LockSHA256 string
	Sources    map[string]SourceBinding
}

// NewBindings builds a bindings record over a lock digest.
func NewBindings(lockSHA256 string, sources map[string]SourceBinding) (*Bindings, error) {
	bindings := &Bindings{LockSHA256: lockSHA256, Sources: sources}
	if err := bindings.Validate(); err != nil {
		return nil, err
	}
	return bindings, nil
}

// Validate enforces identifier aliases, absolute canonical-stable locations,
// and unique disjoint portable root inputs.
func (b *Bindings) Validate() error {
	if b == nil {
		return verr.New("bindings", "source_selection_invalid: bindings are nil")
	}
	if !digestRE.MatchString(b.LockSHA256) {
		return verr.New("lock_sha256", "source_selection_invalid: lock_sha256 must be sha256:<64 lowercase hex>")
	}
	for alias, binding := range b.Sources {
		path := "sources." + alias
		if !identifiers.Valid(alias) {
			return verr.New(path, "source_alias_unknown: invalid source alias %q", alias)
		}
		if err := validLocation(binding.Location); err != nil {
			return verr.New(path+".location", "source_selection_invalid: %s", err)
		}
		seen := map[string]bool{}
		for _, input := range binding.RootInputs {
			if !identifiers.PortablePath(input) {
				return verr.New(path+".root_inputs", "source_selection_invalid: root input %q must be a portable source-relative path", input)
			}
			if seen[input] {
				return verr.New(path+".root_inputs", "source_selection_invalid: duplicate root input %q", input)
			}
			for previous := range seen {
				if isPathPrefix(previous, input) || isPathPrefix(input, previous) {
					return verr.New(path+".root_inputs", "source_selection_invalid: overlapping root inputs %q and %q", previous, input)
				}
			}
			seen[input] = true
		}
	}
	return nil
}

func validLocation(value string) error {
	if value == "" {
		return fmt.Errorf("location must be a canonical absolute path")
	}
	for index := 0; index < len(value); index++ {
		if value[index] == 0 {
			return fmt.Errorf("location must not contain NUL")
		}
	}
	if !filepath.IsAbs(value) || filepath.Clean(value) != value {
		return fmt.Errorf("location must be a canonical absolute path")
	}
	return nil
}

func isPathPrefix(parent, child string) bool {
	return strings.HasPrefix(child, parent+"/")
}

// CheckFresh requires the bindings to belong to the given lock generation.
// Changed machine bindings require explicit refresh and never reinterpret
// an existing lock.
func (b *Bindings) CheckFresh(lock *Lock) error {
	if b == nil {
		return verr.New("bindings", "source_selection_invalid: bindings are nil")
	}
	if lock == nil {
		return verr.New("lock", "source_selection_invalid: lock is nil")
	}
	if b.LockSHA256 != lock.LockSHA256 {
		return verr.New("lock_sha256", "source_lock_stale: machine bindings belong to another lock generation; explicit refresh required")
	}
	return nil
}

func (b *Bindings) object() map[string]any {
	sources := make(map[string]any, len(b.Sources))
	for alias, binding := range b.Sources {
		entry := map[string]any{"location": binding.Location}
		if binding.RootInputs != nil {
			inputs := make([]any, 0, len(binding.RootInputs))
			for _, input := range binding.RootInputs {
				inputs = append(inputs, input)
			}
			entry["root_inputs"] = inputs
		}
		sources[alias] = entry
	}
	return map[string]any{
		"schema_version": BindingsSchemaVersion,
		"lock_sha256":    b.LockSHA256,
		"sources":        sources,
	}
}

// ParseBindings decodes one machine-bindings payload under the strict reader
// discipline: protocol JSON, no unknown fields, full validation.
func ParseBindings(payload []byte) (*Bindings, error) {
	if err := protocoljson.Validate(payload); err != nil {
		return nil, fmt.Errorf("source_selection_invalid: malformed bindings JSON: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var raw any
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("source_selection_invalid: malformed bindings JSON: %w", err)
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil, verr.New("bindings", "source_selection_invalid: bindings must contain a JSON object")
	}
	if unknown := unknownFields(obj, "schema_version", "lock_sha256", "sources"); len(unknown) > 0 {
		return nil, verr.New("bindings", "source_selection_invalid: unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	version, ok := obj["schema_version"].(json.Number)
	if !ok || version.String() != strconv.Itoa(BindingsSchemaVersion) {
		return nil, verr.New("schema_version", "source_selection_invalid: unsupported bindings schema_version")
	}
	digest, _ := obj["lock_sha256"].(string)
	rawSources, ok := obj["sources"].(map[string]any)
	if !ok {
		return nil, verr.New("sources", "source_selection_invalid: bindings require field 'sources' as an object")
	}
	sources := make(map[string]SourceBinding, len(rawSources))
	for alias, entry := range rawSources {
		binding, err := parseSourceBinding(entry, "sources."+alias)
		if err != nil {
			return nil, err
		}
		sources[alias] = binding
	}
	bindings := &Bindings{LockSHA256: digest, Sources: sources}
	if err := bindings.Validate(); err != nil {
		return nil, err
	}
	return bindings, nil
}

func parseSourceBinding(entry any, path string) (SourceBinding, error) {
	obj, ok := entry.(map[string]any)
	if !ok {
		return SourceBinding{}, verr.New(path, "source_selection_invalid: must be an object")
	}
	if unknown := unknownFields(obj, "location", "root_inputs"); len(unknown) > 0 {
		return SourceBinding{}, verr.New(path, "source_selection_invalid: unsupported field(s): %s", strings.Join(unknown, ", "))
	}
	location, _ := obj["location"].(string)
	var inputs []string
	if raw, present := obj["root_inputs"]; present {
		list, ok := raw.([]any)
		if !ok {
			return SourceBinding{}, verr.New(path+".root_inputs", "source_selection_invalid: root inputs must be a list")
		}
		for _, item := range list {
			text, ok := item.(string)
			if !ok {
				return SourceBinding{}, verr.New(path+".root_inputs", "source_selection_invalid: root inputs must be a list of strings")
			}
			inputs = append(inputs, text)
		}
	}
	return SourceBinding{Location: location, RootInputs: inputs}, nil
}

// ReadBindings loads and validates the machine bindings at path. The path
// is caller-chosen machine configuration; it is never the portable lock.
func ReadBindings(path string) (*Bindings, error) {
	payload, err := os.ReadFile(path) // #nosec G304 -- caller-supplied machine path
	if err != nil {
		return nil, err
	}
	bindings, err := ParseBindings(payload)
	if err != nil {
		return nil, fmt.Errorf("bindings %s: %w", path, err)
	}
	return bindings, nil
}

// WriteBindings validates the bindings and stores their canonical bytes at
// path atomically with owner-only permissions.
func WriteBindings(path string, bindings *Bindings) error {
	if bindings == nil {
		return verr.New("bindings", "source_selection_invalid: bindings are nil")
	}
	if err := bindings.Validate(); err != nil {
		return err
	}
	canonical, err := protocoljson.MarshalCanonical(bindings.object())
	if err != nil {
		return err
	}
	return writeFileAtomic(path, canonical, 0o600, 0o700)
}
