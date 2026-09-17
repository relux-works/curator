package identity

import (
	"strings"
	"testing"
)

func TestDraftCanonicalKey(t *testing.T) {
	for _, key := range []string{
		"example.org/kit",
		"example.org/Kit",
		"example.org/a/b/c",
		"e.co/x",
		"example.org/a_b.c-d/e",
		"0x9.example-1/x",
	} {
		if !DraftCanonicalKey(key) {
			t.Errorf("DraftCanonicalKey(%q) = false", key)
		}
	}
	for _, key := range []string{
		"",
		" ",
		"example.org",
		"Example.org/kit",
		"example.org/kit.git",
		"example.org/kit/",
		"/example.org/kit",
		"example.org//kit",
		"example.org/../kit",
		"example.org/./kit",
		"example.org:8443/kit",
		"user@example.org/kit",
		"example.org/ki t",
		"example.org/kit?x",
		"example.org/kit#x",
		"example.org/ki%74",
		"https://example.org/kit",
		"git@example.org:kit.git",
		"ssh://example.org/kit",
		strings.Repeat("a", 5000),
	} {
		if DraftCanonicalKey(key) {
			t.Errorf("DraftCanonicalKey(%q) = true", key)
		}
	}
}
