//go:build windows

package snapshot

import (
	"github.com/relux-works/curator/internal/staging"
)

// captureFileToken pins the physical identity of one admitted file at
// inventory time through the staging identity primitive: a single
// backup-semantics open whose volume serial plus file index feeds the
// token. FileInfo carries no public file index on Windows and
// os.SameFile resolves it lazily from the stored path, so comparing
// two Lstat FileInfo values misses an atomic same-byte replacement;
// the single-handle token cannot name a swapped object.
func captureFileToken(path string) (string, error) {
	return staging.FileIdentity(path)
}
