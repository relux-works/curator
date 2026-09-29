//go:build unix

package registry

import (
	"errors"
	"os"
	"testing"
)

func TestReadBootstrapCheckpointRefusesGroupWritableFile(t *testing.T) {
	signer := newSigner(t)
	path := writeCheckpointFixture(t, t.TempDir(), signer, 9, false, true)
	if err := os.Chmod(path, 0o620); err != nil {
		t.Fatal(err)
	}
	_, err := readBootstrapCheckpoint(path, []string{signer.pinned})
	if err == nil {
		t.Fatal("group-writable checkpoint was accepted")

	}
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("group-writable checkpoint reported as absent: %v", err)
	}
}
