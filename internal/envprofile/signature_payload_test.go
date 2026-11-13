package envprofile

import (
	"bytes"
	"testing"
)

// Removing a folded signature preserves the remaining header/message bytes;
// malformed/ambiguous headers refuse instead of verifying different bytes.
func TestSplitGitSignedPayload(t *testing.T) {
	for _, name := range []string{"gpgsig", "gpgsig-sha256"} {
		raw := []byte("tree abc\n" + name + " -----BEGIN SSH SIGNATURE-----\n line\n -----END SSH SIGNATURE-----\ncommitter fixture <fixture@example.test> 0 +0000\n\nmessage\n\x00")
		data, signature, err := splitGitSignedPayload("commit", raw)
		if err != nil || !bytes.Equal(data, []byte("tree abc\ncommitter fixture <fixture@example.test> 0 +0000\n\nmessage\n\x00")) || !bytes.Equal(signature, []byte("-----BEGIN SSH SIGNATURE-----\nline\n-----END SSH SIGNATURE-----\n")) {
			t.Fatalf("split %s: error=%v", name, err)
		}
	}
	for _, raw := range []string{"no separator", "tree abc\n\nmessage", "gpgsig one\ngpgsig-sha256 two\n\nmessage"} {
		if _, _, err := splitGitSignedPayload("commit", []byte(raw)); err == nil {
			t.Fatal("malformed header accepted")
		}
	}
	for _, armor := range []string{"SSH", "PGP"} {
		prefix := []byte("object abc\ntype commit\n\nmessage\n")
		sig := []byte("-----BEGIN " + armor + " SIGNATURE-----\nvalue\n")
		data, signature, err := splitGitSignedPayload("tag", append(append([]byte{}, prefix...), sig...))
		if err != nil || !bytes.Equal(data, prefix) || !bytes.Equal(signature, sig) {
			t.Fatalf("tag split: %v", err)
		}
	}
}
