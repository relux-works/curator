//go:build windows

package closureexec

import (
	"io/fs"
)

// markTreeDirImmutable is a no-op on Windows, deliberately. The read-only
// attribute Unix mode bits map to does not restrict writing INTO a Windows
// directory — it only breaks MoveFileEx, so setting it would forbid the
// store's atomic tmp-to-digest rename while protecting nothing. The
// immutability of an admitted tree on Windows rests on what the platform can
// actually enforce and prove: every member file carries the read-only
// attribute (set and validated through the shared file checks), and the tree
// content is re-hashed against its digest at every use.
func markTreeDirImmutable(string) error { return nil }

// treeDirIsImmutable accepts every real directory on Windows: synthesized
// directory permission bits carry no write semantics here, so the bits can
// prove nothing either way. Member files remain individually validated.
func treeDirIsImmutable(fs.FileInfo) bool { return true }

// restoreTreeDirMutable mirrors markTreeDirImmutable and is a no-op.
func restoreTreeDirMutable(string) error { return nil }
