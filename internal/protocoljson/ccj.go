package protocoljson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// MaxSafeInteger is the largest integer magnitude admitted by CCJ-1. The
// bound preserves exact cross-language representation in IEEE-754 doubles.
const MaxSafeInteger = int64(9007199254740991)

// MarshalCanonical encodes a JSON-domain value as exact Curator Canonical JSON
// 1 (CCJ-1). Callers are responsible for applying their schema projection,
// such as removing the top-level registry signature member before encoding.
func MarshalCanonical(value any) ([]byte, error) {
	var buffer strings.Builder
	if err := appendCanonical(&buffer, value); err != nil {
		return nil, err
	}
	return []byte(buffer.String()), nil
}

// CanonicalEqual reports whether payload is byte-for-byte equal to the CCJ-1
// encoding of value. It intentionally does not normalize payload.
func CanonicalEqual(payload []byte, value any) (bool, error) {
	canonical, err := MarshalCanonical(value)
	if err != nil {
		return false, err
	}
	return bytes.Equal(payload, canonical), nil
}

// RequireCanonical validates payload and requires exact byte equality with its
// CCJ-1 encoding. Whitespace, alternate escapes, a BOM, and a terminal newline
// are therefore rejected even when they decode to the same JSON value.
func RequireCanonical(payload []byte) error {
	if err := Validate(payload); err != nil {
		return fmt.Errorf("invalid CCJ-1: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("invalid CCJ-1: %w", err)
	}
	canonical, err := MarshalCanonical(value)
	if err != nil {
		return fmt.Errorf("invalid CCJ-1: %w", err)
	}
	if !bytes.Equal(payload, canonical) {
		return fmt.Errorf("invalid CCJ-1: bytes are not canonical")
	}
	return nil
}

// UnmarshalCanonical requires exact CCJ-1 bytes, retains integer precision,
// rejects unknown struct fields, and decodes into destination.
func UnmarshalCanonical(payload []byte, destination any) error {
	if err := RequireCanonical(payload); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode canonical JSON: %w", err)
	}
	return nil
}

func appendCanonical(buffer *strings.Builder, value any) error {
	switch typed := value.(type) {
	case map[string]any:
		return appendObject(buffer, typed)
	case []any:
		buffer.WriteByte('[')
		for index, item := range typed {
			if index > 0 {
				buffer.WriteByte(',')
			}
			if err := appendCanonical(buffer, item); err != nil {
				return err
			}
		}
		buffer.WriteByte(']')
		return nil
	case map[string]string:
		object := make(map[string]any, len(typed))
		for key, item := range typed {
			object[key] = item
		}
		return appendObject(buffer, object)
	case []string:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = item
		}
		return appendCanonical(buffer, items)
	case string:
		return appendString(buffer, typed)
	case json.Number:
		integer, err := strconv.ParseInt(string(typed), 10, 64)
		if err != nil || strconv.FormatInt(integer, 10) != string(typed) {
			return fmt.Errorf("CCJ-1 numbers must be shortest-form base-10 integers: %q", typed)
		}
		return appendInteger(buffer, integer)
	case int:
		return appendInteger(buffer, int64(typed))
	case int8:
		return appendInteger(buffer, int64(typed))
	case int16:
		return appendInteger(buffer, int64(typed))
	case int32:
		return appendInteger(buffer, int64(typed))
	case int64:
		return appendInteger(buffer, typed)
	case uint:
		return appendUnsigned(buffer, uint64(typed))
	case uint8:
		return appendUnsigned(buffer, uint64(typed))
	case uint16:
		return appendUnsigned(buffer, uint64(typed))
	case uint32:
		return appendUnsigned(buffer, uint64(typed))
	case uint64:
		return appendUnsigned(buffer, typed)
	case bool:
		if typed {
			buffer.WriteString("true")
		} else {
			buffer.WriteString("false")
		}
		return nil
	case nil:
		buffer.WriteString("null")
		return nil
	default:
		return fmt.Errorf("CCJ-1 does not support %T", value)
	}
}

func appendObject(buffer *strings.Builder, value map[string]any) error {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	buffer.WriteByte('{')
	for index, key := range keys {
		if index > 0 {
			buffer.WriteByte(',')
		}
		if err := appendString(buffer, key); err != nil {
			return err
		}
		buffer.WriteByte(':')
		if err := appendCanonical(buffer, value[key]); err != nil {
			return err
		}
	}
	buffer.WriteByte('}')
	return nil
}

func appendInteger(buffer *strings.Builder, value int64) error {
	if value < -MaxSafeInteger || value > MaxSafeInteger {
		return fmt.Errorf("CCJ-1 integer outside safe range: %d", value)
	}
	buffer.WriteString(strconv.FormatInt(value, 10))
	return nil
}

func appendUnsigned(buffer *strings.Builder, value uint64) error {
	if value > uint64(MaxSafeInteger) {
		return fmt.Errorf("CCJ-1 integer outside safe range: %d", value)
	}
	buffer.WriteString(strconv.FormatUint(value, 10))
	return nil
}

func appendString(buffer *strings.Builder, value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("CCJ-1 string is not valid UTF-8")
	}
	const hexadecimal = "0123456789abcdef"
	buffer.WriteByte('"')
	for _, character := range value {
		switch character {
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
			if character < 0x20 {
				// Ranging a string never yields a negative rune, so a control
				// character here is one of 0x00..0x1f. Both nibbles therefore
				// index the sixteen-entry table directly, and no narrowing
				// conversion stands between the rune and the escape digits.
				buffer.WriteString(`\u00`)
				buffer.WriteByte(hexadecimal[character>>4])
				buffer.WriteByte(hexadecimal[character&0x0f])
			} else {
				buffer.WriteRune(character)
			}
		}
	}
	buffer.WriteByte('"')
	return nil
}
