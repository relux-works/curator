package registry

import (
	"os"
	"testing"
)

// TestCrossCheckCanonicalBytes compares CanonicalBytes with a reference
// canonical form supplied through env vars; skipped unless set. Used by
// out-of-repo interoperability harnesses.
func TestCrossCheckCanonicalBytes(t *testing.T) {
	refFile := os.Getenv("CURATOR_CROSSCHECK_CANONICAL")
	if refFile == "" {
		t.Skip("cross-check env not set")
	}
	want, err := os.ReadFile(refFile)
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{
		"b": "два", "a": 1.0, "sig": map[string]any{"signature": "x"},
		"z": map[string]any{"y": []any{"м", 2.0}, "x": true},
		"nested": map[string]any{
			"esc": "quote\"and\\slash", "num": 3.5, "null": nil, "tab": "a\tb",
		},
	}
	got := CanonicalBytes(payload)
	if string(got) != string(want) {
		t.Fatalf("canonical bytes diverge:\n got %s\nwant %s", got, want)
	}
}
