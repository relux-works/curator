// Package privatedir creates and validates owner-only private directories in
// a platform-faithful way.
//
// The portable execution and protected-output boundaries in
// internal/closureexec require directories that only the effective user can
// read, enter, or mutate, and they must be able to prove that property later
// rather than assume it. On Unix that property is the 0o700 permission bits.
// On Windows permission bits are synthesized from the read-only attribute and
// say nothing about who may access a directory, so the same property is an
// owner-only, inheritance-protected DACL — the exact shape the reviewed
// protected-root backend in internal/buildcache already creates and validates
// for the build cache. This package gives the closure boundaries that shape
// without importing internal/buildcache (which itself depends on
// internal/closureexec).
//
// The four operations deliberately mirror the call sites they serve:
//
//   - Make creates one directory that is private the instant it becomes
//     observable, like os.Mkdir(path, 0o700) is on Unix.
//   - MakeAll creates a directory and any missing parents, each private at
//     creation, like os.MkdirAll(path, 0o700) is on Unix.
//   - Validate proves an existing path is a real, unlinked, owner-only
//     directory this manager could have created.
//   - Protect makes an existing directory the caller just created private,
//     for the one case (os.MkdirTemp) where creation cannot carry the shape.
//
// Validate returns plain errors; each boundary wraps them in its own
// diagnostic code, because the same property failure is
// artifact_local_output_drift at the protected store and
// closure_input_undeclared at the portable output root.
package privatedir

import "os"

// Make creates path as an owner-only private directory. It fails with
// os.ErrExist (wrapped in *os.PathError) when path already exists, matching
// os.Mkdir.
func Make(path string) error { return makePrivate(path) }

// MakeAll creates path and any missing parents as owner-only private
// directories. Levels that already exist are left untouched and are not
// validated; callers that need the leaf proven private call Validate.
func MakeAll(path string) error { return makeAllPrivate(path) }

// Validate reports whether path is a real directory, not a symlink or
// reparse point, owned by the effective user, and accessible to that user
// only. The nil return is the proof the protected boundaries rely on.
func Validate(path string) error { return validatePrivate(path) }

// Protect makes an existing directory private. It is for directories whose
// creation API cannot attach the private shape atomically (os.MkdirTemp); a
// directory created by Make or MakeAll never needs it.
func Protect(path string) error { return protectPrivate(path) }

// CreateTemp creates an owner-only regular file with a unique name in dir.
// Unlike os.CreateTemp on Windows, it attaches the owner-only DACL at file
// creation so the temporary state is never published with inherited access.
func CreateTemp(dir, pattern string) (*os.File, error) {
	return createPrivateTempFile(dir, pattern)
}

// ValidateFile proves that path is a regular, non-link owner-only file.
func ValidateFile(path string) error { return validatePrivateFile(path) }

// ProtectFile makes an existing regular file owner-only. It rejects links and
// special files instead of following or changing them.
func ProtectFile(path string) error { return protectPrivateFile(path) }

// ProtectTree makes every directory and regular file in a newly created tree
// owner-only. It rejects links and special files and is intended for private
// staging immediately before publication.
func ProtectTree(root string) error { return protectPrivateTree(root) }
