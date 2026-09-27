// Package stateread provides a shared boundary for state and surface reads.
// It classifies a missing path as absent only when its ancestor chain remains
// traversable; errors caused by a non-directory or unreadable ancestor remain
// unreadable.
package stateread

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Kind is the observed state of a manager-owned file or directory.
type Kind string

const (
	// KindPresent means the requested state was present and readable.
	KindPresent Kind = "present"
	// KindAbsent means the requested path is missing under a traversable parent.
	KindAbsent Kind = "absent"
	// KindUnreadable means the path could not be inspected or read reliably.
	KindUnreadable Kind = "unreadable"
)

const (
	// DiagAbsent identifies proven absence of manager-owned state.
	DiagAbsent = "manager_state_absent"
	// DiagUnreadable identifies state whose read or inspection failed.
	DiagUnreadable = "manager_state_unreadable"
)

// File is the result of a state-file read. Bytes is populated only when the
// path was present and readable.
type File struct {
	Kind  Kind
	Bytes []byte
}

// Directory is the result of a manager-owned directory listing. Entries is
// populated only when the path was present and readable.
type Directory struct {
	Kind    Kind
	Entries []os.DirEntry
}

// Metadata is the result of a state-path stat. Info is populated only when
// the path was present and readable.
type Metadata struct {
	Kind Kind
	Info os.FileInfo
}

// Link is the result of reading a symbolic link's target text. Target is
// populated only when the link was present and readable.
type Link struct {
	Kind   Kind
	Target string
}

// Error preserves the path and failed-read class for callers that need to
// attach an operation-specific diagnostic.
type Error struct {
	Kind  Kind
	Path  string
	Cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	code := DiagUnreadable
	if e.Kind == KindAbsent {
		code = DiagAbsent
	}
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", code, e.Path)
	}
	return fmt.Sprintf("%s: %s: %v", code, e.Path, e.Cause)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	if e.Kind == KindAbsent && e.Cause == nil {
		return fs.ErrNotExist
	}
	return e.Cause
}

// AbsentError constructs a typed absence outcome for an operation that
// requires the path to exist.
func AbsentError(path string) *Error {
	return &Error{Kind: KindAbsent, Path: path}
}

// UnusableError classifies present bytes or metadata that cannot be decoded
// or validated as unreadable state.
func UnusableError(path string, cause error) *Error {
	return &Error{Kind: KindUnreadable, Path: path, Cause: cause}
}

// ReadFile reads state or surface data and keeps absence separate from all
// other read failures.
func ReadFile(path string) (File, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- caller supplies the exact state or surface path.
	if err == nil {
		return File{Kind: KindPresent, Bytes: data}, nil
	}
	if missingPath(path, err) {
		return File{Kind: KindAbsent}, nil
	}
	return File{Kind: KindUnreadable}, &Error{Kind: KindUnreadable, Path: path, Cause: err}
}

// ReadRegularFile reads a present regular file (environments §8.4.1). A missing leaf remains
// absent; inspection, type, and read failures are unreadable. This is used
// for manager records and seed files whose boundary requires regular bytes.
func ReadRegularFile(path string) (File, error) {
	metadata, err := Lstat(path)
	if err != nil {
		return File{Kind: KindUnreadable}, err
	}
	if metadata.Kind == KindAbsent {
		return File{Kind: KindAbsent}, nil
	}
	if metadata.Kind != KindPresent || metadata.Info == nil || !metadata.Info.Mode().IsRegular() {
		return File{Kind: KindUnreadable}, UnusableError(path, fmt.Errorf("path must be a regular file"))
	}
	return ReadFile(path)
}

// ReadDir lists state or surface inventory and keeps absence separate from
// all other listing failures.
func ReadDir(path string) (Directory, error) {
	entries, err := os.ReadDir(path)
	if err == nil {
		return Directory{Kind: KindPresent, Entries: entries}, nil
	}
	if missingPath(path, err) {
		return Directory{Kind: KindAbsent}, nil
	}
	return Directory{Kind: KindUnreadable}, &Error{Kind: KindUnreadable, Path: path, Cause: err}
}

// Stat inspects state or surface metadata and keeps absence separate from
// all other stat failures.
func Stat(path string) (Metadata, error) {
	info, err := os.Stat(path)
	if err == nil {
		return Metadata{Kind: KindPresent, Info: info}, nil
	}
	if missingPath(path, err) {
		return Metadata{Kind: KindAbsent}, nil
	}
	return Metadata{Kind: KindUnreadable}, &Error{Kind: KindUnreadable, Path: path, Cause: err}
}

// Lstat inspects the exact state path without following a symbolic link.
func Lstat(path string) (Metadata, error) {
	return LstatWith(path, os.Lstat)
}

// LstatWith classifies an lstat result using the same absent-versus-unreadable
// rule as the other state readers. readPath is an injectable syscall boundary
// for tests that need to exercise errors unavailable on the host filesystem.
func LstatWith(path string, readPath func(string) (os.FileInfo, error)) (Metadata, error) {
	info, err := readPath(path)
	if err == nil {
		return Metadata{Kind: KindPresent, Info: info}, nil
	}
	if missingPath(path, err) {
		return Metadata{Kind: KindAbsent}, nil
	}
	return Metadata{Kind: KindUnreadable}, &Error{Kind: KindUnreadable, Path: path, Cause: err}
}

// Readlink reads a symbolic link's target text and keeps absence separate
// from all other inspection failures.
func Readlink(path string) (Link, error) {
	return ReadlinkWith(path, os.Readlink)
}

// ReadlinkWith classifies a readlink result using the shared absent-versus-
// unreadable rule. readPath is an injectable syscall boundary for tests that
// need to exercise read errors unavailable on the host filesystem.
func ReadlinkWith(path string, readPath func(string) (string, error)) (Link, error) {
	target, err := readPath(path)
	if err == nil {
		return Link{Kind: KindPresent, Target: target}, nil
	}
	if missingPath(path, err) {
		return Link{Kind: KindAbsent}, nil
	}
	return Link{Kind: KindUnreadable}, &Error{Kind: KindUnreadable, Path: path, Cause: err}
}

// missingPath recognizes a missing leaf only when no existing ancestor makes
// the path unaddressable. Windows can report a child beneath a regular file as
// ERROR_PATH_NOT_FOUND; treating that as absence would let callers take an
// absence fallback after a failed read.
func missingPath(path string, cause error) bool {
	return missingPathWith(path, cause, os.IsNotExist)
}

// missingPathWith accepts the platform absence predicate so tests can model
// platform mappings such as Windows ENOTDIR also satisfying IsNotExist.
func missingPathWith(path string, cause error, isNotExist func(error) bool) bool {
	if !isNotExist(cause) {
		return false
	}
	parent := filepath.Dir(path)
	for {
		info, err := os.Lstat(parent)
		if err == nil {
			if info.IsDir() {
				return true
			}
			if info.Mode()&os.ModeSymlink == 0 {
				return false
			}
			// A symlink ancestor is traversable only when it resolves to a
			// directory. Broken links, loops, and links to files stay unreadable.
			target, err := os.Stat(parent)
			return err == nil && target.IsDir()
		}
		if !os.IsNotExist(err) {
			return false
		}
		next := filepath.Dir(parent)
		if next == parent {
			return true
		}
		parent = next
	}
}
