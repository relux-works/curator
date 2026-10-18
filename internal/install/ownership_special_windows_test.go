package install

import (
	"os"
	"testing"
)

// Windows has no filesystem FIFO. A non-regular directory marker exercises
// the common refusal branch; FIFO-specific evidence comes from Unix lanes.
func ownershipSpecial(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
