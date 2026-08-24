package crossconformance_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/crossconformance"
	"github.com/relux-works/curator/internal/protocoljson"
)

// The independent scanner must accept exactly the canonical form and reject
// every near miss. A validator that normalizes its input silently would sign
// off on records whose published bytes are not the bytes that were hashed.
func TestRequireCanonicalRejectsEveryNoncanonicalSpelling(t *testing.T) {
	for _, testCase := range []struct{ name, payload string }{
		{"leading-space", ` {"a":1}`},
		{"inner-space", `{"a": 1}`},
		{"trailing-newline", "{\"a\":1}\n"},
		{"unsorted-members", `{"b":1,"a":1}`},
		{"duplicate-member", `{"a":1,"a":2}`},
		{"solidus-escape", `{"a":"\/"}`},
		{"uppercase-hex-escape", `{"a":"` + `\u001F` + `"}`},
		{"redundant-escape", `{"a":"` + `\u0041` + `"}`},
		{"float", `{"a":1.0}`},
		{"exponent", `{"a":1e3}`},
		{"negative-zero", `{"a":-0}`},
		{"leading-zero", `{"a":01}`},
		{"unsafe-integer", `{"a":9007199254740992}`},
		{"trailing-comma", `{"a":1,}`},
		{"trailing-value", `{"a":1}{}`},
		{"raw-control-byte", "{\"a\":\"\x01\"}"},
		{"unterminated", `{"a":1`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := crossconformance.RequireCanonical([]byte(testCase.payload)); err == nil {
				t.Fatalf("noncanonical payload %q was accepted", testCase.payload)
			}
		})
	}
}

// Canonical bytes must survive the round trip untouched, including the exact
// control-character escape spellings CCJ-1 fixes.
func TestCanonicalFormIsAFixedPointForEveryControlCharacter(t *testing.T) {
	named := map[rune]string{'\b': `\b`, '\f': `\f`, '\n': `\n`, '\r': `\r`, '\t': `\t`}
	for character := rune(0); character < 0x20; character++ {
		escape, ok := named[character]
		if !ok {
			escape = fmt.Sprintf(`\u%04x`, character)
		}
		payload := `{"s":"` + escape + `"}`
		value, err := crossconformance.RequireCanonical([]byte(payload))
		if err != nil {
			t.Fatalf("U+%04X: %v", character, err)
		}
		if value.String("s") != string(character) {
			t.Fatalf("U+%04X decoded to %q", character, value.String("s"))
		}
		canonical, err := crossconformance.Canonical(value)
		if err != nil {
			t.Fatalf("U+%04X: %v", character, err)
		}
		if string(canonical) != payload {
			t.Fatalf("U+%04X re-emitted as %s, want %s", character, canonical, payload)
		}
	}
}

// The independent implementation and the production encoder must agree byte
// for byte on the whole shared value domain, otherwise the oracle above is
// only checking itself.
func TestIndependentEncoderMatchesProductionEncoderOnSharedValues(t *testing.T) {
	for _, value := range []any{
		map[string]any{},
		map[string]any{"b": "two", "a": "one", "A": "upper"},
		map[string]any{"nested": map[string]any{"z": []any{1, 2, 3}, "a": true, "n": nil}},
		map[string]any{"unicode": "é 中\U0001F600", "quote": `"`, "backslash": `\`},
		map[string]any{"bounds": []any{crossconformance.MaxSafeInteger, -crossconformance.MaxSafeInteger, 0}},
		map[string]any{"empty": []any{}, "tabs": "\t\n\r\b\f\x00\x1f"},
	} {
		production, err := protocoljson.MarshalCanonical(value)
		if err != nil {
			t.Fatalf("production encoder: %v", err)
		}
		decoded, err := crossconformance.RequireCanonical(production)
		if err != nil {
			t.Fatalf("independent scanner rejected production bytes %s: %v", production, err)
		}
		independent, err := crossconformance.Canonical(decoded)
		if err != nil {
			t.Fatal(err)
		}
		if string(independent) != string(production) {
			t.Fatalf("independent %s != production %s", independent, production)
		}
	}
}

// Domain separation is the whole point of the identity rule: the same payload
// under two labels must never produce one identity.
func TestDomainSeparationChangesIdentity(t *testing.T) {
	payload := []byte(`{"a":1}`)
	first, err := crossconformance.DomainID("curator-node-v1", payload)
	if err != nil {
		t.Fatal(err)
	}
	second, err := crossconformance.DomainID("curator-edge-v1", payload)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("two domain labels produced one identity")
	}
	if !crossconformance.ValidIdentity(first) || !crossconformance.ValidIdentity(second) {
		t.Fatalf("identities are malformed: %s %s", first, second)
	}
	if _, err = crossconformance.DomainID("", payload); err == nil {
		t.Fatal("an empty domain label was accepted")
	}
	if _, err = crossconformance.DomainID("bad\x00label", payload); err == nil {
		t.Fatal("a NUL-bearing domain label was accepted")
	}
	if _, err = crossconformance.DomainID("curator-node-v1", []byte(`{"a": 1}`)); err == nil {
		t.Fatal("a noncanonical payload was hashed instead of rejected")
	}
}

// Surrogate pairs are the one escape form CCJ-1 never emits but a peer
// implementation may hand us; decoding must be exact and re-emission must
// return to the literal form.
func TestSurrogateEscapesDecodeAndReemitLiterally(t *testing.T) {
	value, err := crossconformance.Parse([]byte(`{"s":"\ud83d\ude00"}`))
	if err != nil {
		t.Fatal(err)
	}
	if value.String("s") != "\U0001F600" {
		t.Fatalf("surrogate pair decoded to %q", value.String("s"))
	}
	canonical, err := crossconformance.Canonical(value)
	if err != nil {
		t.Fatal(err)
	}
	if string(canonical) != "{\"s\":\"\U0001F600\"}" {
		t.Fatalf("surrogate pair re-emitted as %s", canonical)
	}
	for _, broken := range []string{`{"s":"\ud83d"}`, `{"s":"\ud83dA"}`, `{"s":"\udc00\ud83d"}`} {
		if _, err = crossconformance.Parse([]byte(broken)); err == nil {
			t.Fatalf("broken surrogate escape %s was accepted", broken)
		}
	}
}

// Accessors must not silently answer for the wrong kind; a validator built on
// them would then pass a record whose field is a number where an identity is
// required.
func TestTypedAccessorsRefuseWrongKinds(t *testing.T) {
	value, err := crossconformance.RequireCanonical([]byte(`{"a":1,"b":["x","y"],"c":[1],"d":{"e":"f"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if value.String("a") != "" || value.String("missing") != "" || value.String("d", "missing") != "" {
		t.Fatal("String returned a value for a non-string or absent member")
	}
	if value.String("d", "e") != "f" {
		t.Fatal("String did not walk into a nested object")
	}
	if got := strings.Join(value.Strings("b"), ","); got != "x,y" {
		t.Fatalf("Strings(b) = %q", got)
	}
	if value.Strings("c") != nil || value.Strings("a") != nil {
		t.Fatal("Strings returned items for a non-string array")
	}
}
