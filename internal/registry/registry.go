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
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
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
		if err != nil {
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
	if !knownStatuses[record.Status] {
		return Record{}, fmt.Errorf("audit record status %q is not recognized", record.Status)
	}
	if audit, present := payload["audit"]; present {
		if _, ok := audit.(map[string]any); !ok {
			return Record{}, fmt.Errorf("audit record field 'audit' must be an object")
		}
	}
	return record, nil
}

// VerifySigned checks the signature of a record or snapshot object against
// any pinned key.
func VerifySigned(payload map[string]any, pinnedKeys []string) bool {
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
