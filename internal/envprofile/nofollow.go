package envprofile

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/stateread"
)

// managedPath joins a path below root after checking each existing parent
// component with lstat semantics. Components at or above root are allowed to
// resolve normally; components below it may not be symlinks.
func managedPath(root, rel string, createParents bool) (string, error) {
	parts, err := managedPathParts(rel)
	if err != nil {
		return "", err
	}
	root = filepath.Clean(root)
	if createParents {
		if err := privatedir.MakeAll(root); err != nil {
			return "", err
		}
	}
	rootState, err := stateread.Stat(root)
	if err != nil {
		return "", err
	}
	if rootState.Kind == stateread.KindAbsent {
		return "", stateread.AbsentError(root)
	}
	if !rootState.Info.IsDir() {
		return "", fmt.Errorf("managed root %s is not a directory", root)
	}
	current := root
	for _, component := range parts[:len(parts)-1] {
		current = filepath.Join(current, component)
		state, statErr := stateread.Lstat(current)
		if statErr != nil {
			return "", statErr
		}
		if state.Kind == stateread.KindAbsent && createParents {
			if mkdirErr := privatedir.Make(current); mkdirErr != nil && !os.IsExist(mkdirErr) {
				return "", mkdirErr
			}
			state, statErr = stateread.Lstat(current)
		}
		if statErr != nil {
			return "", statErr
		}
		if state.Kind == stateread.KindAbsent {
			return filepath.Join(append([]string{root}, parts...)...), nil
		}
		info := state.Info
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("%s: %s", DiagWriteWouldFollowLink, current)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("managed path component %s is not a directory", current)
		}
	}
	return filepath.Join(append([]string{root}, parts...)...), nil
}

// managedDirectory creates each component below root separately and refuses
// any pre-existing symbolic link on the route.
func managedDirectory(root, rel string) (string, error) {
	parts, err := managedPathParts(rel)
	if err != nil {
		return "", err
	}
	root = filepath.Clean(root)
	if err := privatedir.MakeAll(root); err != nil {
		return "", err
	}
	rootState, err := stateread.Stat(root)
	if err != nil {
		return "", err
	}
	if rootState.Kind == stateread.KindAbsent {
		return "", stateread.AbsentError(root)
	}
	if !rootState.Info.IsDir() {
		return "", fmt.Errorf("managed root %s is not a directory", root)
	}
	current := root
	for _, component := range parts {
		current = filepath.Join(current, component)
		state, err := stateread.Lstat(current)
		if err != nil {
			return "", err
		}
		if state.Kind == stateread.KindAbsent {
			if mkdirErr := privatedir.Make(current); mkdirErr != nil && !os.IsExist(mkdirErr) {
				return "", mkdirErr
			}
			state, err = stateread.Lstat(current)
		}
		if err != nil {
			return "", err
		}
		if state.Kind == stateread.KindAbsent {
			return "", stateread.AbsentError(current)
		}
		info := state.Info
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("%s: %s", DiagWriteWouldFollowLink, current)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("managed path component %s is not a directory", current)
		}
	}
	return current, nil
}

func managedPathParts(rel string) ([]string, error) {
	slash := filepath.ToSlash(rel)
	if strings.Contains(rel, "\\") || !fs.ValidPath(slash) || slash == "." {
		return nil, fmt.Errorf("invalid managed relative path %q", rel)
	}
	return strings.Split(slash, "/"), nil
}

func managedJoin(base, rel string) string {
	if base == "" {
		return filepath.ToSlash(filepath.FromSlash(rel))
	}
	return filepath.ToSlash(filepath.Join(filepath.FromSlash(base), filepath.FromSlash(rel)))
}

func managedRelative(root, full string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("managed path %s is not below root %s", full, root)
	}
	rel = filepath.ToSlash(rel)
	if _, err := managedPathParts(rel); err != nil {
		return "", err
	}
	return rel, nil
}

func checkManagedPrivateTarget(root, rel string) error {
	rootState, err := stateread.Stat(root)
	if err != nil {
		return err
	}
	if rootState.Kind == stateread.KindAbsent {
		return nil
	}
	full, err := managedPath(root, rel, false)
	if err != nil {
		return err
	}
	targetState, err := stateread.Lstat(full)
	if err != nil {
		return err
	}
	if targetState.Kind == stateread.KindAbsent {
		return nil
	}
	info := targetState.Info
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s: %s", DiagWriteWouldFollowLink, full)
	}
	return nil
}

// atomicManagedFile replaces a file entry with bytes staged under an
// operation-private name in the same directory (§8.3.1). Rename replaces a
// target symlink itself rather than opening the path through it.
func atomicManagedFile(root, rel string, payload []byte, mode fs.FileMode) error {
	full, err := managedPath(root, rel, true)
	if err != nil {
		return err
	}
	file, err := privatedir.CreateTemp(filepath.Dir(full), ".curator-write-*.tmp")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer func() { _ = os.Remove(temp) }()
	if err := file.Chmod(mode); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(payload); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return replaceManagedEntry(temp, full)
}

// errSymlinkUnavailable marks a failure to create the private symlink itself:
// the only condition under which a caller may fall back to a copy (manager
// §5). A failure to rename over an existing entry is not that condition.
type errSymlinkUnavailable struct{ err error }

func (e *errSymlinkUnavailable) Error() string { return e.err.Error() }
func (e *errSymlinkUnavailable) Unwrap() error { return e.err }

// replaceManagedEntry renames source over target. Where the platform refuses
// to rename over an existing link/reparse-point or directory entry (Windows),
// the target ENTRY is removed with lstat semantics — never opened or
// traversed (§8.3.1) — and the rename is retried. os.Remove on a directory
// symlink or junction removes the link itself; a non-empty real directory is
// not removed and the original error is returned. On unix the first rename
// replaces the entry atomically and the fallback is not reached.
func replaceManagedEntry(source, target string) error {
	err := renameManagedEntry(source, target)
	if err == nil {
		return nil
	}
	info, lerr := os.Lstat(target)
	if lerr != nil {
		return err
	}
	if info.Mode()&(os.ModeSymlink|os.ModeIrregular) == 0 && !info.IsDir() {
		return err
	}
	if rerr := os.Remove(target); rerr != nil {
		return err
	}
	return renameManagedEntry(source, target)
}

// atomicManagedLink creates an operation-private symlink beside the target
// and renames that entry into place (§8.3.1). The target text is never opened
// by this operation.
func atomicManagedLink(root, rel, linkTarget string) error {
	full, err := managedPath(root, rel, true)
	if err != nil {
		return err
	}
	dir := filepath.Dir(full)
	for range 8 {
		var nonce [16]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			return err
		}
		temp := filepath.Join(dir, ".curator-link-"+hex.EncodeToString(nonce[:]))
		if err := os.Symlink(linkTarget, temp); err != nil {
			if os.IsExist(err) {
				continue
			}
			return &errSymlinkUnavailable{err: err}
		}
		if err := replaceManagedEntry(temp, full); err != nil {
			_ = os.Remove(temp)
			return err
		}
		return nil
	}
	return fmt.Errorf("could not allocate a private symlink name beside %s", full)
}

// removeManagedEntry removes a target entry without resolving its final
// component and refuses traversal through any parent link.
func removeManagedEntry(root, rel string) error {
	full, err := managedPath(root, rel, false)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// readManagedRegular opens a regular file without following a final symlink.
// It is used for backup input; links are handled with Readlink and preserved
// as links instead.
func readManagedRegular(root, rel string) (payload []byte, present bool, err error) {
	full, err := managedPath(root, rel, false)
	if err != nil {
		return nil, false, err
	}
	state, err := stateread.Lstat(full)
	if err != nil {
		return nil, false, err
	}
	if state.Kind == stateread.KindAbsent {
		return nil, false, nil
	}
	info := state.Info
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, false, fmt.Errorf("%s: %s", DiagWriteWouldFollowLink, full)
	}
	if !info.Mode().IsRegular() {
		return nil, false, nil
	}
	file, err := openReadNoFollow(full)
	if err != nil {
		return nil, false, err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	opened, err := file.Stat()
	if err != nil {
		return nil, false, err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return nil, false, fmt.Errorf("%s: %s changed while being backed up", DiagWriteWouldFollowLink, full)
	}
	payload, err = io.ReadAll(file)
	return payload, true, err
}
