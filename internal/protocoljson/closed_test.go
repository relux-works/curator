package protocoljson

import (
	"encoding/json"
	"testing"
)

// Proves exact, recursive field presence for internal authority carriers. It
// does not change the acceptance rules of any existing public protocol reader.
func TestClosedAuthorityDecode(t *testing.T) {
	type nested struct {
		Value string `json:"value"`
	}
	type carrier struct {
		Required string          `json:"required"`
		Nested   nested          `json:"nested"`
		Optional *string         `json:"optional,omitempty"`
		Items    []nested        `json:"items"`
		Raw      json.RawMessage `json:"raw"`
	}
	valid := `{"required":"","nested":{"value":"x"},"items":[],"raw":{"separate":"schema"}}`
	var got carrier
	if err := DecodeClosed([]byte(valid), &got); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		`{"Required":"","nested":{"value":"x"},"items":[],"raw":{}}`,
		`{"required":"","nested":{"Value":"x"},"items":[],"raw":{}}`,
		`{"required":"","nested":{"value":"x","extra":0},"items":[],"raw":{}}`,
		`{"required":"","nested":null,"items":[],"raw":{}}`,
		`{"required":"","nested":{"value":"x"},"items":null,"raw":{}}`,
		`{"required":"","nested":{"value":"x"},"items":[],"optional":null,"raw":{}}`,
		`{"required":"","nested":{"value":"x"},"items":[{"Value":"x"}],"raw":{}}`,
		`{"required":"","required":"","nested":{"value":"x"},"items":[],"raw":{}}`,
		`{"required":"","nested":{"value":"x"},"items":[]}`,
	} {
		if err := DecodeClosed([]byte(bad), &got); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
}
