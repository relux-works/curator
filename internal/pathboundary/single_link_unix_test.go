//go:build unix

package pathboundary

import (
	"os"
	"path/filepath"
	"testing"
)

type noLinkMetadata struct{ os.FileInfo }

func (noLinkMetadata) Sys() any { return nil }

// Unsupported metadata is not evidence of a single link; a real lstat of the
// same ordinary file remains a valid positive control.
func TestSingleLinkRequiresUnixMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload")
	if err := os.WriteFile(path, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := singleLink(path, info); err != nil {
		t.Fatal(err)
	}
	if err := singleLink(path, noLinkMetadata{info}); err == nil || err.Error() != "regular-file link metadata is unavailable" {
		t.Fatalf("missing link metadata accepted or misclassified: %v", err)
	}
}
