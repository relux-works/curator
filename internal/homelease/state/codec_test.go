package state

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

// The exact accepted schema corpus rejects each malformed carrier while its
// paired control remains valid. This proves decoding, not runtime authority.
func TestAcceptedSchemaCorpus(t *testing.T) {
	raw, err := os.ReadFile("testdata/schema-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Pairs []struct {
			Name, Schema, Positive, Negative string
			NegativeHex                      string `json:"negative_hex"`
		}
	}
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.Pairs) != 131 {
		t.Fatalf("unexpected accepted corpus size: %d", len(corpus.Pairs))
	}
	for _, pair := range corpus.Pairs {
		t.Run(pair.Name, func(t *testing.T) {
			if err := Validate(pair.Schema, []byte(pair.Positive)); err != nil {
				t.Fatalf("positive: %v", err)
			}
			bad := []byte(pair.Negative)
			if pair.NegativeHex != "" {
				bad, err = hex.DecodeString(pair.NegativeHex)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := Validate(pair.Schema, bad); err == nil {
				t.Fatal("accepted forbidden carrier")
			}
		})
	}
}
