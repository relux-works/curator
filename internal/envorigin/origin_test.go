package envorigin

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// The exact proposed canonical corpus is exercised by the Go consumer as well
// as the Python publication checker. No filesystem authority is claimed here.
func TestCanonicalSchemaCases(t *testing.T) {
	raw, err := os.ReadFile("testdata/schema-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name   string          `json:"name"`
		Record json.RawMessage `json:"record"`
		Valid  bool            `json:"valid"`
	}
	if err = json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			record, err := Parse(c.Record)
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v error=%v", c.Valid, err)
			}
			if c.Valid {
				encoded, e := record.Marshal()
				if e != nil {
					t.Fatal(e)
				}
				if _, e = Parse(encoded); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}

// Duplicate keys and an extra JSON document cannot redefine trusted fields;
// a valid record controls the same parser and errors exclude input bytes.
func TestTransportRefusals(t *testing.T) {
	raw, err := os.ReadFile("testdata/schema-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Record json.RawMessage `json:"record"`
	}
	if err = json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	valid := cases[0].Record
	if _, err = Parse(valid); err != nil {
		t.Fatal(err)
	}
	for _, value := range [][]byte{append(append([]byte{}, valid...), []byte(" {}")...), []byte(`{"record":"private-value","record":"provisioning-origin/1"}`)} {
		if _, err = Parse(value); err == nil || !strings.Contains(err.Error(), DiagInvalid) || strings.Contains(err.Error(), "private-value") {
			t.Fatalf("transport refusal=%v", err)
		}
	}
}
