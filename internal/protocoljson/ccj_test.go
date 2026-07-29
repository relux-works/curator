package protocoljson

import (
	"fmt"
	"testing"
)

// TestMarshalCanonicalEscapesEveryControlCharacter pins the exact bytes CCJ-1
// emits for the whole C0 range. The escape digits are produced by indexing a
// hexadecimal table with the rune itself, so this fixes the contract the
// encoder has to keep: the five named escapes stay named, every other control
// character becomes a lower-case \u00xx pair, and the first printable
// character above the range is still written through untouched.
func TestMarshalCanonicalEscapesEveryControlCharacter(t *testing.T) {
	named := map[rune]string{
		'\b': `\b`,
		'\t': `\t`,
		'\n': `\n`,
		'\f': `\f`,
		'\r': `\r`,
	}
	for character := rune(0); character < 0x20; character++ {
		escape, ok := named[character]
		if !ok {
			escape = fmt.Sprintf(`\u%04x`, character)
		}
		want := `{"s":"` + escape + `"}`
		got, err := MarshalCanonical(map[string]any{"s": string(character)})
		if err != nil {
			t.Fatalf("MarshalCanonical(U+%04X): %v", character, err)
		}
		if string(got) != want {
			t.Fatalf("U+%04X encoded as %s, CCJ-1 requires %s", character, got, want)
		}
	}
	// The guard is exclusive: a space is the first character the encoder must
	// write literally, and the two structural characters keep their own
	// escapes rather than falling into the \u00xx form.
	for _, testCase := range []struct {
		value string
		want  string
	}{
		{value: " ", want: `{"s":" "}`},
		{value: `"`, want: `{"s":"\""}`},
		{value: `\`, want: `{"s":"\\"}`},
		{value: "\x7f", want: "{\"s\":\"\x7f\"}"},
		{value: "é", want: `{"s":"é"}`},
	} {
		got, err := MarshalCanonical(map[string]any{"s": testCase.value})
		if err != nil {
			t.Fatalf("MarshalCanonical(%q): %v", testCase.value, err)
		}
		if string(got) != testCase.want {
			t.Fatalf("%q encoded as %s, CCJ-1 requires %s", testCase.value, got, testCase.want)
		}
	}
}

// TestRequireCanonicalAcceptsEveryControlCharacterEscape closes the loop: the
// bytes the encoder produces for the C0 range must be exactly the bytes the
// canonical-form check accepts, and re-encoding them must be a fixed point.
func TestRequireCanonicalAcceptsEveryControlCharacterEscape(t *testing.T) {
	for character := rune(0); character < 0x20; character++ {
		payload, err := MarshalCanonical(map[string]any{"s": string(character)})
		if err != nil {
			t.Fatalf("MarshalCanonical(U+%04X): %v", character, err)
		}
		if err := RequireCanonical(payload); err != nil {
			t.Fatalf("U+%04X encoding %s is not canonical: %v", character, payload, err)
		}
		var decoded map[string]any
		if err := UnmarshalCanonical(payload, &decoded); err != nil {
			t.Fatalf("U+%04X encoding %s did not round-trip: %v", character, payload, err)
		}
		if decoded["s"] != string(character) {
			t.Fatalf("U+%04X round-tripped to %q", character, decoded["s"])
		}
	}
}
