package swiftpmsource

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
)

// SourceKind preserves SwiftPM's source-control location kind. The two kinds
// must never be silently transformed into each other while generating mirrors.
type SourceKind string

// Supported SwiftPM source-control and contained path kinds.
const (
	SourceRemote SourceKind = "remoteSourceControl"
	SourceLocal  SourceKind = "localSourceControl"
	SourcePath   SourceKind = "fileSystem"
)

// Pin is one exact top-level Package.resolved source-control instance.
type Pin struct {
	Identity          string
	Kind              SourceKind
	RawLocation       string
	CanonicalLocation string
	Revision          string
	Version           string
	Branch            string
}

// Lock is the frozen top-level lock. Dependency lockfiles are ordinary
// captured text and never contribute pins to this value.
type Lock struct {
	Schema int
	Digest closuregraph.ID
	Bytes  []byte
	Pins   []Pin
}

type resolvedEnvelope struct {
	Version int           `json:"version"`
	Pins    []resolvedPin `json:"pins"`
}

type resolvedPin struct {
	Identity      string        `json:"identity"`
	Package       string        `json:"package"`
	Kind          string        `json:"kind"`
	Location      string        `json:"location"`
	RepositoryURL string        `json:"repositoryURL"`
	State         resolvedState `json:"state"`
}

type resolvedState struct {
	Revision string `json:"revision"`
	Version  string `json:"version"`
	Branch   string `json:"branch"`
}

// ParseResolved parses the supported v2/v3 lock schemas without accepting
// unknown trailing JSON, mutable pins, duplicate identities, or kindless
// locations.
func ParseResolved(payload []byte) (Lock, error) {
	if len(payload) == 0 {
		return Lock{}, fail(CodeResolutionUnfrozen, "root Package.resolved is absent")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return Lock{}, fail(CodeResolutionUnfrozen, "root Package.resolved is malformed: %v", err)
	}
	if len(raw) != 2 || raw["version"] == nil || raw["pins"] == nil {
		return Lock{}, fail(CodeResolutionUnfrozen, "root Package.resolved has an unsupported shape")
	}
	var envelope resolvedEnvelope
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil || (envelope.Version != 2 && envelope.Version != 3) {
		return Lock{}, fail(CodeResolutionUnfrozen, "root Package.resolved schema is unsupported")
	}
	pins := make([]Pin, 0, len(envelope.Pins))
	for _, item := range envelope.Pins {
		identity := strings.ToLower(item.Identity)
		if identity == "" {
			identity = strings.ToLower(item.Package)
		}
		location := item.Location
		if location == "" {
			location = item.RepositoryURL
		}
		kind := SourceKind(item.Kind)
		if envelope.Version == 2 && kind == "" {
			kind = SourceRemote
		}
		if identity == "" || location == "" || !validSourceKind(kind) || !validRevision(item.State.Revision) {
			return Lock{}, failFields(CodeResolutionUnfrozen, map[string]string{"identity": identity}, "lock pin is incomplete or mutable")
		}
		canonical, err := canonicalLocation(kind, location)
		if err != nil {
			return Lock{}, err
		}
		pins = append(pins, Pin{Identity: identity, Kind: kind, RawLocation: location, CanonicalLocation: canonical, Revision: strings.ToLower(item.State.Revision), Version: item.State.Version, Branch: item.State.Branch})
	}
	sort.Slice(pins, func(i, j int) bool { return pins[i].Identity < pins[j].Identity })
	for index := range pins {
		if index > 0 && pins[index-1].Identity == pins[index].Identity {
			return Lock{}, failFields(CodeDependencyPinMismatch, map[string]string{"identity": pins[index].Identity}, "duplicate lock identity")
		}
	}
	digest := sha256.Sum256(payload)
	return Lock{Schema: envelope.Version, Digest: closuregraph.ID("sha256:" + hex.EncodeToString(digest[:])), Bytes: append([]byte(nil), payload...), Pins: pins}, nil
}

func validRevision(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validSourceKind(kind SourceKind) bool { return kind == SourceRemote || kind == SourceLocal }

func canonicalLocation(kind SourceKind, location string) (string, error) {
	if strings.ContainsAny(location, "\x00\r\n") {
		return "", fail(CodeDependencyOriginUnsupported, "source-control location is non-portable")
	}
	switch kind {
	case SourceRemote:
		if !strings.Contains(location, "://") || strings.HasPrefix(location, "file://") {
			return "", fail(CodeDependencyOriginUnsupported, "remote source-control location is not remote")
		}
		return strings.TrimSuffix(location, "/"), nil
	case SourceLocal:
		if !strings.HasPrefix(location, "/") {
			return "", fail(CodeDependencyOriginUnsupported, "local source-control location is not absolute")
		}
		return strings.TrimSuffix(location, "/"), nil
	default:
		return "", fail(CodeDependencyOriginUnsupported, "unsupported source-control kind %q", kind)
	}
}

func lockRecordID(lock Lock) (closuregraph.ID, error) {
	pins := make([]any, len(lock.Pins))
	for index, pin := range lock.Pins {
		pins[index] = map[string]any{"branch": pin.Branch, "canonical_location": pin.CanonicalLocation, "identity": pin.Identity, "kind": string(pin.Kind), "raw_location": pin.RawLocation, "revision": pin.Revision, "version": pin.Version}
	}
	return closuregraph.DomainID("swiftpm-root-lock-v1", map[string]any{"raw_sha256": string(lock.Digest), "schema": int64(lock.Schema), "pins": pins})
}
