package observation

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// The producer's exact closed vectors are the positive/negative oracle. A
// valid observation remains readable here without supplying lease authority.
func TestPublishedObservationVectors(t *testing.T) {
	raw, err := os.ReadFile("testdata/observation-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string
		Valid    bool
		Document json.RawMessage
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := Parse(c.Document)
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v: %v", c.Valid, err)
			}
		})
	}
}

// Duplicate keys, unsafe transport and invented proof fields cannot pass the
// parser even when the unaffected producer vector is valid.
func TestObservationTransportAndClosedProofs(t *testing.T) {
	raw, err := os.ReadFile("testdata/observation-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct{ Document json.RawMessage }
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	good := string(cases[0].Document)
	if _, err := Parse([]byte(good)); err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string]string{
		"duplicate":     strings.Replace(good, `"schema":`, `"schema":"curator-run-session-observation-v1","schema":`, 1),
		"process-proof": strings.Replace(good, `"session": {`, `"process":{"pid":42,"started_at":"2026-01-01T00:00:00Z"},"session": {`, 1),
		"trailing":      good + `{}`, "bom": "\xef\xbb\xbf" + good,
		"oversize": strings.Repeat(" ", MaxBytes) + good,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(bad)); err == nil {
				t.Fatal("unsafe observation accepted")
			}
		})
	}
}
