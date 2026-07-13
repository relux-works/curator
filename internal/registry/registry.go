// Package registry implements the audit registry client (Spec §13):
// canonical bytes, Ed25519 verification against pinned keys, deny-wins
// federation, snapshot verification with persisted monotonic versions,
// record caching with TTL and offline grace, and record submission.
package registry

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/identity"
)

// Record statuses (Spec §13.1).
const (
	StatusAudited    = "audited"
	StatusRevoked    = "revoked"
	StatusDeprecated = "deprecated"
	StatusPending    = "pending"
)

var knownStatuses = map[string]bool{
	StatusAudited: true, StatusRevoked: true, StatusDeprecated: true, StatusPending: true,
}

var (
	commitRE = regexp.MustCompile(`^[0-9a-f]{40}(?:[0-9a-f]{24})?$`)
	sha256RE = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	keyIDRE  = regexp.MustCompile(`^[0-9a-f]{16}$`)
)

// Federation results (Spec §13.3).
const (
	ResultRevoked    = "revoked"
	ResultAudited    = "audited"
	ResultDeprecated = "deprecated"
	ResultUnknown    = "unknown"
)

// Registry is one pinned trusted registry.
type Registry struct {
	Name       string
	URL        string
	PublicKeys []string
}

// Record is a parsed audit record.
type Record struct {
	Name           string
	SourceIdentity string
	Commit         string
	ContentSHA256  string
	Status         string
	Raw            map[string]any
}

// KeyID returns sig.key_id when present.
func (r Record) KeyID() string {
	sig, _ := r.Raw["sig"].(map[string]any)
	keyID, _ := sig["key_id"].(string)
	return keyID
}

// Attestation is the authorizing verified record summary.
type Attestation struct {
	Registry string
	Status   string
	KeyID    string
	Record   map[string]any
}

// Resolution combines every trusted registry for one artifact.
type Resolution struct {
	Result      string
	Attestation *Attestation
	Warnings    []string
}

// FetchFn returns raw record payloads for an artifact query.
type FetchFn func(url, sourceIdentity, commit, contentSHA256 string) ([]map[string]any, error)

// CanonicalBytes is the signed form of any registry object: compact sorted
// CCJ-1 JSON of every field except the top-level "sig" (Spec Registry §1).
// Invalid CCJ-1 input returns nil; verification uses CanonicalBytesChecked so
// malformed signed objects cannot be mistaken for valid bytes.
func CanonicalBytes(record map[string]any) []byte {
	payload, _ := CanonicalBytesChecked(record)
	return payload
}

// CanonicalBytesChecked validates the CCJ-1 value domain and returns the
// canonical signed bytes.
func CanonicalBytesChecked(record map[string]any) ([]byte, error) {
	body := map[string]any{}
	for key, value := range record {
		if key != "sig" {
			body[key] = value
		}
	}
	return compactSortedJSON(body)
}

func compactSortedJSON(value any) ([]byte, error) {
	switch typed := value.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var buffer strings.Builder
		buffer.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				buffer.WriteByte(',')
			}
			keyJSON, err := bytesNoEscape(key)
			if err != nil {
				return nil, err
			}
			buffer.Write(keyJSON)
			buffer.WriteByte(':')
			item, err := compactSortedJSON(typed[key])
			if err != nil {
				return nil, err
			}
			buffer.Write(item)
		}
		buffer.WriteByte('}')
		return []byte(buffer.String()), nil
	case []any:
		var buffer strings.Builder
		buffer.WriteByte('[')
		for index, item := range typed {
			if index > 0 {
				buffer.WriteByte(',')
			}
			encoded, err := compactSortedJSON(item)
			if err != nil {
				return nil, err
			}
			buffer.Write(encoded)
		}
		buffer.WriteByte(']')
		return []byte(buffer.String()), nil
	case string:
		return bytesNoEscape(typed)
	case nil:
		return []byte("null"), nil
	case bool:
		if typed {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	case int:
		return safeInteger(int64(typed))
	case int64:
		return safeInteger(typed)
	case json.Number:
		value, err := strconv.ParseInt(string(typed), 10, 64)
		if err != nil || strconv.FormatInt(value, 10) != string(typed) {
			return nil, fmt.Errorf("CCJ-1 numbers must be base-10 integers: %q", typed)
		}
		return safeInteger(value)
	default:
		return nil, fmt.Errorf("CCJ-1 does not support %T", value)
	}
}

func safeInteger(value int64) ([]byte, error) {
	const maximum = int64(9007199254740991)
	if value < -maximum || value > maximum {
		return nil, fmt.Errorf("CCJ-1 integer outside safe range: %d", value)
	}
	return []byte(strconv.FormatInt(value, 10)), nil
}

// bytesNoEscape implements the CCJ-1 minimal string escaping rules.
func bytesNoEscape(value string) ([]byte, error) {
	if !utf8.ValidString(value) {
		return nil, fmt.Errorf("CCJ-1 string is not valid UTF-8")
	}
	var buffer strings.Builder
	buffer.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"':
			buffer.WriteString(`\"`)
		case '\\':
			buffer.WriteString(`\\`)
		case '\b':
			buffer.WriteString(`\b`)
		case '\f':
			buffer.WriteString(`\f`)
		case '\n':
			buffer.WriteString(`\n`)
		case '\r':
			buffer.WriteString(`\r`)
		case '\t':
			buffer.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&buffer, `\u%04x`, r)
			} else {
				buffer.WriteRune(r)
			}
		}
	}
	buffer.WriteByte('"')
	return []byte(buffer.String()), nil
}

// ParsePublicKey decodes a pinned key "ed25519:<base64>" (prefix optional)
// into 32 raw bytes (Spec §13.2).
func ParsePublicKey(value string) (ed25519.PublicKey, error) {
	text := strings.TrimSpace(value)
	text = strings.TrimPrefix(text, "ed25519:")
	raw, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, fmt.Errorf("invalid public key encoding: %q", value)
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("ed25519 public key must be 32 bytes, got %d", len(raw))
	}
	return ed25519.PublicKey(raw), nil
}

// KeyID derives the key id: first 16 hex chars of SHA-256 over the raw key.
func KeyID(publicKey ed25519.PublicKey) string {
	sum := sha256.Sum256(publicKey)
	return hex.EncodeToString(sum[:])[:16]
}

// ParseRecord validates a raw record payload (Spec §13.1).
func ParseRecord(payload map[string]any) (Record, error) {
	if unknown := unknownKeys(payload, "schema_version", "name", "source_identity", "commit", "content_sha256", "status", "audit", "endorsements", "sig"); len(unknown) > 0 {
		return Record{}, fmt.Errorf("audit record has unknown fields: %s", strings.Join(unknown, ", "))
	}
	if schema, present := payload["schema_version"]; present && !integerEquals(schema, 1) {
		return Record{}, fmt.Errorf("audit record schema_version must be 1")
	}
	record := Record{Raw: payload}
	for _, field := range []struct {
		key string
		dst *string
	}{
		{"name", &record.Name},
		{"source_identity", &record.SourceIdentity},
		{"commit", &record.Commit},
		{"content_sha256", &record.ContentSHA256},
		{"status", &record.Status},
	} {
		value, _ := payload[field.key].(string)
		if value == "" {
			return Record{}, fmt.Errorf("audit record requires a non-empty string %q", field.key)
		}
		*field.dst = value
	}
	if !identifiers.Valid(record.Name) {
		return Record{}, fmt.Errorf("audit record name is not a portable identifier")
	}
	if utf8.RuneCountInString(record.SourceIdentity) > 4096 || !identity.ValidCanonical(record.SourceIdentity) {
		return Record{}, fmt.Errorf("audit record source_identity is not canonical")
	}
	if !commitRE.MatchString(record.Commit) {
		return Record{}, fmt.Errorf("audit record commit must be a full lowercase object id")
	}
	if !sha256RE.MatchString(record.ContentSHA256) {
		return Record{}, fmt.Errorf("audit record content_sha256 is malformed")
	}
	if !knownStatuses[record.Status] {
		return Record{}, fmt.Errorf("audit record status %q is not recognized", record.Status)
	}
	if audit, present := payload["audit"]; present {
		if _, ok := audit.(map[string]any); !ok {
			return Record{}, fmt.Errorf("audit record field 'audit' must be an object")
		}
	}
	if err := validateSignatureEnvelope(payload["sig"]); err != nil {
		return Record{}, fmt.Errorf("audit record signature: %w", err)
	}
	if rawEndorsements, present := payload["endorsements"]; present {
		endorsements, ok := rawEndorsements.([]any)
		if !ok {
			return Record{}, fmt.Errorf("audit record endorsements must be an array")
		}
		for index, rawEndorsement := range endorsements {
			endorsement, ok := rawEndorsement.(map[string]any)
			if !ok || len(unknownKeys(endorsement, "endorser", "sig")) > 0 || len(endorsement) != 2 {
				return Record{}, fmt.Errorf("audit record endorsement %d is malformed", index)
			}
			endorser, _ := endorsement["endorser"].(string)
			if endorser == "" || utf8.RuneCountInString(endorser) > 8192 {
				return Record{}, fmt.Errorf("audit record endorsement %d has an invalid endorser", index)
			}
			if err := validateSignatureEnvelope(endorsement["sig"]); err != nil {
				return Record{}, fmt.Errorf("audit record endorsement %d signature: %w", index, err)
			}
		}
	}
	if _, err := CanonicalBytesChecked(payload); err != nil {
		return Record{}, fmt.Errorf("audit record is not valid CCJ-1: %w", err)
	}
	return record, nil
}

// VerifySigned checks the signature of a record or snapshot object against
// any pinned key.
func VerifySigned(payload map[string]any, pinnedKeys []string) bool {
	if err := validateSignatureEnvelope(payload["sig"]); err != nil {
		return false
	}
	sig, _ := payload["sig"].(map[string]any)
	algorithm, _ := sig["algorithm"].(string)
	keyID, _ := sig["key_id"].(string)
	signatureB64, _ := sig["signature"].(string)
	if algorithm != "ed25519" || keyID == "" || signatureB64 == "" {
		return false
	}
	signature, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return false
	}
	message, err := CanonicalBytesChecked(payload)
	if err != nil {
		return false
	}
	for _, pinned := range pinnedKeys {
		publicKey, err := ParsePublicKey(pinned)
		if err != nil {
			continue
		}
		if KeyID(publicKey) != keyID {
			continue
		}
		if ed25519.Verify(publicKey, message, signature) {
			return true
		}
	}
	return false
}

func validateSignatureEnvelope(raw any) error {
	envelope, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("must be an object")
	}
	if unknown := unknownKeys(envelope, "algorithm", "key_id", "signature"); len(unknown) > 0 || len(envelope) != 3 {
		return fmt.Errorf("must contain exactly algorithm, key_id, and signature")
	}
	algorithm, _ := envelope["algorithm"].(string)
	keyID, _ := envelope["key_id"].(string)
	signatureText, _ := envelope["signature"].(string)
	if algorithm != "ed25519" || !keyIDRE.MatchString(keyID) {
		return fmt.Errorf("has an unsupported algorithm or malformed key id")
	}
	signature, err := base64.StdEncoding.DecodeString(signatureText)
	if err != nil || len(signature) != ed25519.SignatureSize || base64.StdEncoding.EncodeToString(signature) != signatureText {
		return fmt.Errorf("has a malformed Ed25519 signature")
	}
	return nil
}

func unknownKeys(object map[string]any, allowed ...string) []string {
	known := make(map[string]bool, len(allowed))
	for _, field := range allowed {
		known[field] = true
	}
	var unknown []string
	for field := range object {
		if !known[field] {
			unknown = append(unknown, field)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func integerEquals(raw any, expected int64) bool {
	switch value := raw.(type) {
	case int:
		return int64(value) == expected
	case int64:
		return value == expected
	case json.Number:
		parsed, err := strconv.ParseInt(string(value), 10, 64)
		return err == nil && strconv.FormatInt(parsed, 10) == string(value) && parsed == expected
	default:
		return false
	}
}

// Matches reports whether a record names the artifact: equal content hash,
// or equal source identity plus commit (Spec §13.3).
func Matches(record Record, sourceIdentity, commit, contentSHA256 string) bool {
	if record.ContentSHA256 == contentSHA256 {
		return true
	}
	return record.SourceIdentity == sourceIdentity && record.Commit == commit
}

// Resolve combines verified records from every trusted registry under
// deny-wins (Spec §13.3).
func Resolve(registries []Registry, sourceIdentity, commit, contentSHA256 string, fetch FetchFn) Resolution {
	var warnings []string
	var audited, deprecated *Attestation
	for _, reg := range registries {
		if len(reg.PublicKeys) == 0 {
			warnings = append(warnings, fmt.Sprintf("registry %s has no pinned keys; its records are not trusted", reg.Name))
			continue
		}
		payloads, err := fetch(reg.URL, sourceIdentity, commit, contentSHA256)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("registry %s unavailable: %v", reg.Name, err))
			continue
		}
		for _, payload := range payloads {
			record, err := ParseRecord(payload)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("registry %s returned a malformed record: %v", reg.Name, err))
				continue
			}
			if !Matches(record, sourceIdentity, commit, contentSHA256) {
				continue
			}
			if !VerifySigned(payload, reg.PublicKeys) {
				warnings = append(warnings, fmt.Sprintf("registry %s record for %s failed signature verification", reg.Name, record.Name))
				continue
			}
			attestation := &Attestation{Registry: reg.Name, Status: record.Status, KeyID: record.KeyID(), Record: payload}
			switch record.Status {
			case StatusRevoked:
				return Resolution{Result: ResultRevoked, Attestation: attestation, Warnings: warnings}
			case StatusAudited:
				if audited == nil {
					audited = attestation
				}
			case StatusDeprecated:
				if deprecated == nil {
					deprecated = attestation
				}
			}
		}
	}
	if audited != nil {
		return Resolution{Result: ResultAudited, Attestation: audited, Warnings: warnings}
	}
	if deprecated != nil {
		return Resolution{Result: ResultDeprecated, Attestation: deprecated, Warnings: warnings}
	}
	return Resolution{Result: ResultUnknown, Warnings: warnings}
}

// parseISO8601 accepts RFC 3339 with a Z or offset suffix.
func parseISO8601(value any) (time.Time, bool) {
	text, ok := value.(string)
	if !ok || text == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}
