// Package crossconformance owns Curator's cross-adapter source-closure proof.
//
// It publishes three reusable things and nothing else:
//
//   - an independent Curator Canonical JSON 1 scanner, canonicalizer, and
//     domain-separated identity function that never calls the production
//     encoder it is meant to check;
//   - the accepted 53-record CGP05/CGP10 golden corpus plus the structural
//     claims the accepted decision requires of it; and
//   - the normative semantic obligations and rejection vectors every adapter
//     path must satisfy, expressed so that Rust, npm, pnpm, Yarn Classic,
//     modern Yarn, and SwiftPM/C-family all run the same suite.
//
// The package deliberately implements no adapter, starts no process, reads no
// package-manager state, and re-derives no security decision. It is the
// integration proof over the already-accepted adapter contracts.
package crossconformance

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// MaxSafeInteger is the largest integer magnitude CCJ-1 admits. It is restated
// here rather than imported so that this oracle disagrees with the production
// encoder when the production encoder changes.
const MaxSafeInteger = int64(9007199254740991)

// Value is a decoded CCJ-1 value. Objects keep decoded keys and their values;
// canonical ordering is applied at emission, not at parse time, so a payload
// that arrives out of order is still detected as noncanonical.
type Value struct {
	Kind    ValueKind
	Str     string
	Int     int64
	Bool    bool
	Items   []Value
	Members []Member
}

// Member is one decoded object member in the order it appeared.
type Member struct {
	Key   string
	Value Value
}

// ValueKind enumerates the closed CCJ-1 value domain.
type ValueKind int

const (
	// KindNull is the JSON null literal.
	KindNull ValueKind = iota
	// KindBool is a JSON boolean literal.
	KindBool
	// KindInt is a CCJ-1 integer; CCJ-1 admits no other number form.
	KindInt
	// KindString is a JSON string.
	KindString
	// KindArray is a JSON array.
	KindArray
	// KindObject is a JSON object.
	KindObject
)

// Parse decodes exactly one CCJ-1 value from payload and rejects trailing
// bytes, duplicate object members, non-integer numbers, invalid UTF-8, and
// every escape form CCJ-1 does not emit.
func Parse(payload []byte) (Value, error) {
	scanner := &scanner{input: payload}
	value, err := scanner.value()
	if err != nil {
		return Value{}, err
	}
	if scanner.offset != len(payload) {
		return Value{}, fmt.Errorf("trailing bytes at offset %d", scanner.offset)
	}
	return value, nil
}

// Canonical re-emits value as exact CCJ-1 bytes.
func Canonical(value Value) ([]byte, error) {
	var builder strings.Builder
	if err := emit(&builder, value); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

// RequireCanonical decodes payload and requires that its own canonical
// re-emission is byte-identical. Whitespace, reordered members, alternate
// escapes, and a terminal newline are therefore all rejected.
func RequireCanonical(payload []byte) (Value, error) {
	value, err := Parse(payload)
	if err != nil {
		return Value{}, err
	}
	canonical, err := Canonical(value)
	if err != nil {
		return Value{}, err
	}
	if string(canonical) != string(payload) {
		return Value{}, errors.New("payload bytes are not canonical CCJ-1")
	}
	return value, nil
}

// DomainID derives ID(label, payload) = SHA256(label || 0x00 || CCJ(payload))
// from bytes that must already be canonical.
func DomainID(label string, payload []byte) (string, error) {
	if label == "" || strings.ContainsRune(label, 0) {
		return "", errors.New("identity label must be non-empty and NUL-free")
	}
	if _, err := RequireCanonical(payload); err != nil {
		return "", fmt.Errorf("identity payload: %w", err)
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte(label))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write(payload)
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

// Member returns the named member of an object value.
func (value Value) Member(key string) (Value, bool) {
	if value.Kind != KindObject {
		return Value{}, false
	}
	for _, member := range value.Members {
		if member.Key == key {
			return member.Value, true
		}
	}
	return Value{}, false
}

// String returns the string member at path, or "" when it is absent or has
// another kind. Callers that need the distinction use Member.
func (value Value) String(path ...string) string {
	current := value
	for _, key := range path {
		next, ok := current.Member(key)
		if !ok {
			return ""
		}
		current = next
	}
	if current.Kind != KindString {
		return ""
	}
	return current.Str
}

// Strings returns the string array at path, or nil when it is absent or has
// another kind.
func (value Value) Strings(path ...string) []string {
	current := value
	for _, key := range path {
		next, ok := current.Member(key)
		if !ok {
			return nil
		}
		current = next
	}
	if current.Kind != KindArray {
		return nil
	}
	items := make([]string, 0, len(current.Items))
	for _, item := range current.Items {
		if item.Kind != KindString {
			return nil
		}
		items = append(items, item.Str)
	}
	return items
}

type scanner struct {
	input  []byte
	offset int
}

func (s *scanner) value() (Value, error) {
	if s.offset >= len(s.input) {
		return Value{}, fmt.Errorf("unexpected end of input at offset %d", s.offset)
	}
	switch character := s.input[s.offset]; character {
	case '{':
		return s.object()
	case '[':
		return s.array()
	case '"':
		text, err := s.text()
		return Value{Kind: KindString, Str: text}, err
	case 't':
		return Value{Kind: KindBool, Bool: true}, s.literal("true")
	case 'f':
		return Value{Kind: KindBool}, s.literal("false")
	case 'n':
		return Value{Kind: KindNull}, s.literal("null")
	default:
		if character == '-' || (character >= '0' && character <= '9') {
			return s.number()
		}
		return Value{}, fmt.Errorf("unexpected byte %q at offset %d", character, s.offset)
	}
}

func (s *scanner) literal(word string) error {
	if !strings.HasPrefix(string(s.input[s.offset:]), word) {
		return fmt.Errorf("invalid literal at offset %d", s.offset)
	}
	s.offset += len(word)
	return nil
}

func (s *scanner) object() (Value, error) {
	s.offset++
	value := Value{Kind: KindObject}
	if s.peek() == '}' {
		s.offset++
		return value, nil
	}
	seen := map[string]bool{}
	for {
		if s.peek() != '"' {
			return Value{}, fmt.Errorf("object member name expected at offset %d", s.offset)
		}
		key, err := s.text()
		if err != nil {
			return Value{}, err
		}
		if seen[key] {
			return Value{}, fmt.Errorf("duplicate object member %q", key)
		}
		seen[key] = true
		if s.peek() != ':' {
			return Value{}, fmt.Errorf("object member separator expected at offset %d", s.offset)
		}
		s.offset++
		member, err := s.value()
		if err != nil {
			return Value{}, err
		}
		value.Members = append(value.Members, Member{Key: key, Value: member})
		switch s.peek() {
		case ',':
			s.offset++
		case '}':
			s.offset++
			return value, nil
		default:
			return Value{}, fmt.Errorf("object continuation expected at offset %d", s.offset)
		}
	}
}

func (s *scanner) array() (Value, error) {
	s.offset++
	value := Value{Kind: KindArray, Items: []Value{}}
	if s.peek() == ']' {
		s.offset++
		return value, nil
	}
	for {
		item, err := s.value()
		if err != nil {
			return Value{}, err
		}
		value.Items = append(value.Items, item)
		switch s.peek() {
		case ',':
			s.offset++
		case ']':
			s.offset++
			return value, nil
		default:
			return Value{}, fmt.Errorf("array continuation expected at offset %d", s.offset)
		}
	}
}

func (s *scanner) number() (Value, error) {
	start := s.offset
	if s.peek() == '-' {
		s.offset++
	}
	digits := s.offset
	for s.offset < len(s.input) && s.input[s.offset] >= '0' && s.input[s.offset] <= '9' {
		s.offset++
	}
	if s.offset == digits {
		return Value{}, fmt.Errorf("number requires at least one digit at offset %d", digits)
	}
	if s.offset < len(s.input) {
		switch s.input[s.offset] {
		case '.', 'e', 'E':
			return Value{}, fmt.Errorf("CCJ-1 admits integers only, at offset %d", s.offset)
		}
	}
	text := string(s.input[start:s.offset])
	parsed, err := strconv.ParseInt(text, 10, 64)
	if err != nil || strconv.FormatInt(parsed, 10) != text {
		return Value{}, fmt.Errorf("CCJ-1 integers must be shortest-form base 10: %q", text)
	}
	if parsed < -MaxSafeInteger || parsed > MaxSafeInteger {
		return Value{}, fmt.Errorf("CCJ-1 integer outside the safe range: %s", text)
	}
	return Value{Kind: KindInt, Int: parsed}, nil
}

func (s *scanner) text() (string, error) {
	s.offset++
	var builder strings.Builder
	for {
		if s.offset >= len(s.input) {
			return "", errors.New("unterminated string")
		}
		character := s.input[s.offset]
		switch {
		case character == '"':
			s.offset++
			if !utf8.ValidString(builder.String()) {
				return "", errors.New("string is not valid UTF-8")
			}
			return builder.String(), nil
		case character == '\\':
			decoded, err := s.escape()
			if err != nil {
				return "", err
			}
			builder.WriteString(decoded)
		case character < 0x20:
			return "", fmt.Errorf("unescaped control byte %#02x at offset %d", character, s.offset)
		default:
			decoded, size := utf8.DecodeRune(s.input[s.offset:])
			if decoded == utf8.RuneError && size <= 1 {
				return "", fmt.Errorf("invalid UTF-8 at offset %d", s.offset)
			}
			builder.Write(s.input[s.offset : s.offset+size])
			s.offset += size
		}
	}
}

func (s *scanner) escape() (string, error) {
	if s.offset+1 >= len(s.input) {
		return "", errors.New("truncated escape")
	}
	switch s.input[s.offset+1] {
	case '"':
		s.offset += 2
		return `"`, nil
	case '\\':
		s.offset += 2
		return `\`, nil
	case '/':
		s.offset += 2
		return "/", nil
	case 'b':
		s.offset += 2
		return "\b", nil
	case 'f':
		s.offset += 2
		return "\f", nil
	case 'n':
		s.offset += 2
		return "\n", nil
	case 'r':
		s.offset += 2
		return "\r", nil
	case 't':
		s.offset += 2
		return "\t", nil
	case 'u':
		return s.unicodeEscape()
	default:
		return "", fmt.Errorf("unsupported escape \\%c at offset %d", s.input[s.offset+1], s.offset)
	}
}

func (s *scanner) unicodeEscape() (string, error) {
	first, err := s.hexQuad()
	if err != nil {
		return "", err
	}
	if !utf16.IsSurrogate(rune(first)) {
		return string(rune(first)), nil
	}
	if s.offset+1 >= len(s.input) || s.input[s.offset] != '\\' || s.input[s.offset+1] != 'u' {
		return "", errors.New("unpaired UTF-16 surrogate escape")
	}
	second, err := s.hexQuad()
	if err != nil {
		return "", err
	}
	combined := utf16.DecodeRune(rune(first), rune(second))
	if combined == utf8.RuneError {
		return "", errors.New("invalid UTF-16 surrogate pair")
	}
	return string(combined), nil
}

func (s *scanner) hexQuad() (uint16, error) {
	if s.offset+6 > len(s.input) {
		return 0, errors.New("truncated \\u escape")
	}
	digits := string(s.input[s.offset+2 : s.offset+6])
	parsed, err := strconv.ParseUint(digits, 16, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid \\u escape %q", digits)
	}
	s.offset += 6
	return uint16(parsed), nil
}

func (s *scanner) peek() byte {
	if s.offset >= len(s.input) {
		return 0
	}
	return s.input[s.offset]
}

func emit(builder *strings.Builder, value Value) error {
	switch value.Kind {
	case KindNull:
		builder.WriteString("null")
		return nil
	case KindBool:
		if value.Bool {
			builder.WriteString("true")
		} else {
			builder.WriteString("false")
		}
		return nil
	case KindInt:
		if value.Int < -MaxSafeInteger || value.Int > MaxSafeInteger {
			return fmt.Errorf("CCJ-1 integer outside the safe range: %d", value.Int)
		}
		builder.WriteString(strconv.FormatInt(value.Int, 10))
		return nil
	case KindString:
		return emitString(builder, value.Str)
	case KindArray:
		builder.WriteByte('[')
		for index, item := range value.Items {
			if index > 0 {
				builder.WriteByte(',')
			}
			if err := emit(builder, item); err != nil {
				return err
			}
		}
		builder.WriteByte(']')
		return nil
	case KindObject:
		return emitObject(builder, value)
	default:
		return fmt.Errorf("unsupported CCJ-1 value kind %d", value.Kind)
	}
}

func emitObject(builder *strings.Builder, value Value) error {
	ordered := make([]Member, len(value.Members))
	copy(ordered, value.Members)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Key < ordered[j].Key })
	builder.WriteByte('{')
	for index, member := range ordered {
		if index > 0 {
			builder.WriteByte(',')
		}
		if err := emitString(builder, member.Key); err != nil {
			return err
		}
		builder.WriteByte(':')
		if err := emit(builder, member.Value); err != nil {
			return err
		}
	}
	builder.WriteByte('}')
	return nil
}

func emitString(builder *strings.Builder, text string) error {
	if !utf8.ValidString(text) {
		return errors.New("CCJ-1 string is not valid UTF-8")
	}
	const hexadecimal = "0123456789abcdef"
	builder.WriteByte('"')
	for _, character := range text {
		switch character {
		case '"':
			builder.WriteString(`\"`)
		case '\\':
			builder.WriteString(`\\`)
		case '\b':
			builder.WriteString(`\b`)
		case '\f':
			builder.WriteString(`\f`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		default:
			if character < 0x20 {
				builder.WriteString(`\u00`)
				builder.WriteByte(hexadecimal[(character>>4)&0x0f])
				builder.WriteByte(hexadecimal[character&0x0f])
				continue
			}
			builder.WriteRune(character)
		}
	}
	builder.WriteByte('"')
	return nil
}

// Text builds a CCJ string value.
func Text(value string) Value { return Value{Kind: KindString, Str: value} }

// Integer builds a CCJ integer value.
func Integer(value int64) Value { return Value{Kind: KindInt, Int: value} }

// Boolean builds a CCJ boolean value.
func Boolean(value bool) Value { return Value{Kind: KindBool, Bool: value} }

// Array builds a CCJ array value.
func Array(items ...Value) Value {
	if items == nil {
		items = []Value{}
	}
	return Value{Kind: KindArray, Items: items}
}

// TextArray builds a CCJ array of strings.
func TextArray(values []string) Value {
	items := make([]Value, 0, len(values))
	for _, value := range values {
		items = append(items, Text(value))
	}
	return Array(items...)
}

// Object builds a CCJ object value. Emission sorts members, so callers may
// list them in whatever order reads best.
func Object(members ...Member) Value {
	return Value{Kind: KindObject, Members: members}
}

// Field is shorthand for one object member.
func Field(key string, value Value) Member { return Member{Key: key, Value: value} }
