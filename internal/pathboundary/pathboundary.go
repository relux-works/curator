// Package pathboundary verifies operator-owned source directories without
// changing or adopting their contents. It is shared by path-source admission
// and the profile-store boundary work described by environments §4.
package pathboundary

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// CheckOwnership names the owner-identity boundary check.
	CheckOwnership = "ownership"
	// CheckPermissions names the private-mutation boundary check.
	CheckPermissions = "permissions"
	// CheckContainment names the source-root containment boundary check.
	CheckContainment = "containment"
	// CheckRegular names the regular-file-type boundary check.
	CheckRegular = "regular_types"
	// CheckLinkSafety names the no-symlink boundary check.
	CheckLinkSafety = "link_safety"
)

// Failure identifies the first boundary check that could not be proven.
// Root is the operator-declared directory; Path is the component that failed.
type Failure struct {
	Root   string
	Path   string
	Check  string
	Reason string
	Cause  error
}

func (f *Failure) Error() string {
	if f.Cause != nil {
		return fmt.Sprintf("%s: %s boundary check failed at %s: %v", f.Check, f.Root, f.Path, f.Cause)
	}
	return fmt.Sprintf("%s: %s boundary check failed at %s: %s", f.Check, f.Root, f.Path, f.Reason)
}

func (f *Failure) Unwrap() error { return f.Cause }

// OwnerIdentity is the platform-specific identity used to compare a path's
// owner with the effective operator.
type OwnerIdentity string

// OwnerLookup reads the identity that owns a path without following a link.
// Tests may inject an identity for a selected path while production uses
// DefaultOwnerLookup.
type OwnerLookup func(path string, info os.FileInfo) (OwnerIdentity, error)

// DefaultOwnerLookup returns the platform implementation that reads owner
// identity without following a symbolic link.
func DefaultOwnerLookup() OwnerLookup { return lookupOwner }

// Validate proves the protected-boundary contract for a path source
// directory (environments §4). It never follows links, repairs permissions,
// changes ownership, or copies source bytes.
func Validate(root string) error {
	return ValidateWithOwner(root, DefaultOwnerLookup())
}

// ValidateWithOwner proves the same contract as Validate while allowing the
// caller to supply owner identities. Production supplies DefaultOwnerLookup;
// the seam lets production-entry tests cover foreign-owner paths without
// requiring a second host user or chown privilege.
func ValidateWithOwner(root string, ownerLookup OwnerLookup) error {
	return validateWithOwnerAndEntryInfo(root, ownerLookup, func(_ string, entry fs.DirEntry) (os.FileInfo, error) {
		return entry.Info()
	})
}

// entryInfoLookup is the metadata read performed after a parent directory
// has been listed. validateWithOwnerAndEntryInfo keeps that boundary explicit
// so tests can reproduce a replacement or removal in the readdir/lstat window.
type entryInfoLookup func(path string, entry fs.DirEntry) (os.FileInfo, error)

func validateWithOwnerAndEntryInfo(root string, ownerLookup OwnerLookup, entryInfo entryInfoLookup) error {
	if ownerLookup == nil {
		return &Failure{Root: root, Path: root, Check: CheckOwnership, Reason: "owner lookup is unavailable"}
	}
	operator, err := effectiveOwner()
	if err != nil {
		return &Failure{Root: root, Path: root, Check: CheckOwnership, Cause: err}
	}
	absolute, rootInfo, err := inspectRoot(root)
	if err != nil {
		return err
	}
	if err := checkNode(absolute, absolute, rootInfo, operator, ownerLookup); err != nil {
		return err
	}

	return filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, walkErr error) error {
		// vanished applies environments §4 to every per-entry probe (lstat,
		// link, ownership/DACL, and the directory open for recursion): an
		// entry that no longer exists when examined was not part of the
		// directory at examination time. Any other probe error fails closed.
		vanished := func(err error) (error, bool) {
			if path == absolute || !IsAbsent(err) {
				return nil, false
			}
			if entry != nil && entry.IsDir() {
				// Keep WalkDir from descending into the vanished path.
				return filepath.SkipDir, true
			}
			return nil, true
		}
		if walkErr != nil {
			if result, gone := vanished(walkErr); gone {
				return result
			}
			return &Failure{Root: absolute, Path: path, Check: CheckRegular, Cause: walkErr}
		}
		if path == absolute {
			return nil
		}
		relative, err := filepath.Rel(absolute, path)
		if err != nil || escapes(relative) {
			return &Failure{Root: absolute, Path: path, Check: CheckContainment, Reason: "component does not resolve below the declared directory", Cause: err}
		}
		info, err := entryInfo(path, entry)
		if err != nil {
			if result, gone := vanished(err); gone {
				return result
			}
			return &Failure{Root: absolute, Path: path, Check: CheckRegular, Cause: err}
		}
		link, err := isLink(path, info)
		if err != nil {
			if result, gone := vanished(err); gone {
				return result
			}
			return &Failure{Root: absolute, Path: path, Check: CheckLinkSafety, Cause: err}
		}
		if link {
			// A link is refused whether or not it still exists when its
			// target is read: lstat already observed it inside the tree.
			if escapes, readErr := symlinkEscapes(absolute, path); readErr != nil {
				return &Failure{Root: absolute, Path: path, Check: CheckLinkSafety, Cause: readErr}
			} else if escapes {
				return &Failure{Root: absolute, Path: path, Check: CheckContainment, Reason: "symbolic-link target escapes the declared directory"}
			}
			return &Failure{Root: absolute, Path: path, Check: CheckLinkSafety, Reason: "symbolic link below the declared directory"}
		}
		if err := checkNode(absolute, path, info, operator, ownerLookup); err != nil {
			var failure *Failure
			if errors.As(err, &failure) && failure.Cause != nil {
				if result, gone := vanished(failure.Cause); gone {
					return result
				}
			}
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return &Failure{Root: absolute, Path: path, Check: CheckRegular, Reason: "component is neither a directory nor a regular file"}
		}
		return nil
	})
}

// ValidateRoot checks only the exact directory node. It is useful for a
// protected parent whose children are independent declared source roots.
func ValidateRoot(root string) error {
	absolute, info, err := inspectRoot(root)
	if err != nil {
		return err
	}
	operator, err := effectiveOwner()
	if err != nil {
		return &Failure{Root: absolute, Path: absolute, Check: CheckOwnership, Cause: err}
	}
	return checkNode(absolute, absolute, info, operator, DefaultOwnerLookup())
}

func inspectRoot(root string) (string, os.FileInfo, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", nil, &Failure{Root: root, Path: root, Check: CheckContainment, Cause: err}
	}
	absolute = filepath.Clean(absolute)
	rootInfo, err := os.Lstat(absolute)
	if err != nil {
		return "", nil, &Failure{Root: absolute, Path: absolute, Check: CheckRegular, Cause: err}
	}
	rootLink, err := isLink(absolute, rootInfo)
	if err != nil {
		return "", nil, &Failure{Root: absolute, Path: absolute, Check: CheckLinkSafety, Cause: err}
	}
	if rootLink {
		return "", nil, &Failure{Root: absolute, Path: absolute, Check: CheckLinkSafety, Reason: "declared directory is a symbolic link"}
	}
	if !rootInfo.IsDir() {
		return "", nil, &Failure{Root: absolute, Path: absolute, Check: CheckRegular, Reason: "declared path is not a directory"}
	}
	return absolute, rootInfo, nil
}

// ProtectTree establishes private mutation permissions for a tree the
// manager just created. It is only for manager-created staging such as an
// onboarding import; an existing operator path is never repaired this way.
func ProtectTree(root string) error { return protectTree(root) }

func checkNode(root, path string, info os.FileInfo, operator OwnerIdentity, ownerLookup OwnerLookup) error {
	owner, err := ownerLookup(path, info)
	if err != nil {
		return &Failure{Root: root, Path: path, Check: CheckOwnership, Cause: err}
	}
	if owner != operator {
		return &Failure{Root: root, Path: path, Check: CheckOwnership, Reason: fmt.Sprintf("%s is not owned by the effective operator", path)}
	}
	if err := checkMutationPermissions(path, info); err != nil {
		return &Failure{Root: root, Path: path, Check: CheckPermissions, Cause: err}
	}
	return nil
}

func escapes(relative string) bool {
	return relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

// symlinkEscapes resolves only the target spelling. It reads the link
// metadata and target bytes, but never follows the link. This lets the
// boundary diagnostic distinguish an escaping link from a link wholly
// within the declared tree; both are then refused.
func symlinkEscapes(root, path string) (bool, error) {
	target, err := os.Readlink(path)
	if err != nil {
		return false, err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return true, err
	}
	relative, err := filepath.Rel(root, filepath.Clean(target))
	if err != nil {
		return true, err
	}
	return escapes(relative), nil
}

// IsAbsent reports a missing declared path without conflating other read
// failures with absence.
func IsAbsent(err error) bool { return errors.Is(err, fs.ErrNotExist) }
