// Package protocoljson validates the common JSON transport requirements used
// by portable Curator Protocol objects before a package-specific schema parser
// decodes them.
package protocoljson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

// Validate rejects byte-order marks, invalid UTF-8, duplicate object keys,
// lone Unicode surrogates, malformed JSON, and trailing non-whitespace data.
func Validate(payload []byte) error {
	if bytes.HasPrefix(payload, []byte{0xef, 0xbb, 0xbf}) {
		return fmt.Errorf("protocol JSON must not contain a byte-order mark")
	}
	if !utf8.Valid(payload) {
		return fmt.Errorf("protocol JSON is not valid UTF-8")
	}
	if err := validateSurrogates(payload); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := consumeValue(decoder); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("protocol JSON has trailing data")
		}
		return fmt.Errorf("protocol JSON trailing data: %w", err)
	}
	return nil
}

func consumeValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("malformed protocol JSON: %w", err)
	}
	delim, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return fmt.Errorf("malformed protocol JSON object: %w", err)
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("protocol JSON object key is not a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate JSON object key %q", key)
			}
			seen[key] = true
			if err := consumeValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim('}') {
			return fmt.Errorf("malformed protocol JSON object")
		}
	case '[':
		for decoder.More() {
			if err := consumeValue(decoder); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil || end != json.Delim(']') {
			return fmt.Errorf("malformed protocol JSON array")
		}
	default:
		return fmt.Errorf("unexpected protocol JSON delimiter %q", delim)
	}
	return nil
}

func validateSurrogates(payload []byte) error {
	inString := false
	for index := 0; index < len(payload); index++ {
		switch payload[index] {
		case '"':
			inString = !inString
		case '\\':
			if !inString || index+1 >= len(payload) {
				continue
			}
			if payload[index+1] != 'u' {
				index++
				continue
			}
			value, ok := hex16(payload, index+2)
			if !ok {
				continue // the JSON decoder reports malformed escapes
			}
			if value >= 0xdc00 && value <= 0xdfff {
				return fmt.Errorf("protocol JSON contains a lone low surrogate")
			}
			if value >= 0xd800 && value <= 0xdbff {
				if index+11 >= len(payload) || payload[index+6] != '\\' || payload[index+7] != 'u' {
					return fmt.Errorf("protocol JSON contains a lone high surrogate")
				}
				low, valid := hex16(payload, index+8)
				if !valid || low < 0xdc00 || low > 0xdfff {
					return fmt.Errorf("protocol JSON contains a lone high surrogate")
				}
				index += 11
				continue
			}
			index += 5
		}
	}
	return nil
}

func hex16(payload []byte, start int) (uint16, bool) {
	if start+4 > len(payload) {
		return 0, false
	}
	var value uint16
	for _, character := range payload[start : start+4] {
		value <<= 4
		switch {
		case character >= '0' && character <= '9':
			value += uint16(character - '0')
		case character >= 'a' && character <= 'f':
			value += uint16(character-'a') + 10
		case character >= 'A' && character <= 'F':
			value += uint16(character-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
