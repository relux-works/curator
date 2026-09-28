package globalbins

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/identifiers"
	"github.com/relux-works/curator/internal/runtimestore"
	"github.com/relux-works/curator/internal/stateread"
)

// Adoption describes a verified forwarding shim that was added to the
// user-bin ownership ledger.
type Adoption struct {
	Path           string
	Backup         string
	AlreadyManaged bool
	DryRun         bool
}

// Adopt takes an existing user-bin forwarding shim under Curator management
// only when its bytes equal the canonical forwarding shim for an installed
// global command. The source entry is never replaced. A caller performing a
// real adoption must hold the manager-home mutation lock; dry runs are read-only.
func Adopt(home, name, platform string, environment map[string]string, userHome string, dryRun bool) (Adoption, error) {
	if platform == "" {
		platform = runtimestore.Platform()
	}
	if userHome == "" {
		userHome, _ = os.UserHomeDir()
	}
	selection := Select(home, platform, environment, userHome)
	if selection.Path == "" {
		return Adoption{}, fmt.Errorf("global: cannot adopt command %q: %s", name, selection.Warning)
	}

	path := filepath.Join(selection.Path, name)
	if identifiers.Valid(name) {
		path = shimPath(selection.Path, name, platform)
	}
	if !identifiers.Valid(name) {
		return Adoption{}, adoptionRefusal(path, "command name is not a valid Curator global command")
	}

	canonical := shimPath(filepath.Join(home, "global", "bin"), name, platform)
	canonicalState, err := stateread.Lstat(canonical)
	if err != nil {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("canonical target %s cannot be inspected: %v", canonical, err))
	}
	if canonicalState.Kind == stateread.KindAbsent {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("no Curator global command has a canonical target at %s", canonical))
	}
	if canonicalState.Info == nil || !canonicalState.Info.Mode().IsRegular() {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("canonical target %s is not a regular file", canonical))
	}

	entryState, err := stateread.Lstat(path)
	if err != nil {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("entry cannot be inspected: %v", err))
	}
	if entryState.Kind == stateread.KindAbsent {
		return Adoption{}, adoptionRefusal(path, "entry does not exist")
	}
	if entryState.Info == nil {
		return Adoption{}, adoptionRefusal(path, "entry metadata is unavailable")
	}
	if entryState.Info.Mode()&os.ModeSymlink != 0 {
		return Adoption{}, adoptionRefusal(path, "entry is a symbolic link")
	}
	if !entryState.Info.Mode().IsRegular() {
		return Adoption{}, adoptionRefusal(path, "entry is not a regular file")
	}
	entry, err := stateread.ReadRegularFile(path)
	if err != nil {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("entry cannot be read as a regular file: %v", err))
	}

	want := runtimestore.UnixShimContent(canonical, nil)
	if platform == "windows" {
		want = runtimestore.WindowsShimContent(canonical, nil)
	}
	if !bytes.Equal(entry.Bytes, []byte(want)) {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("bytes differ from the canonical Curator shim for %s", canonical))
	}

	managed, err := readLedger(selection.Path)
	if err != nil {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("ownership marker %s cannot be read: %v", filepath.Join(selection.Path, managedFile), err))
	}
	if managed[name] {
		return Adoption{Path: path, AlreadyManaged: true}, nil
	}
	if dryRun {
		return Adoption{Path: path, DryRun: true}, nil
	}

	if err := confirmAdoptionEntry(path, entryState.Info, entry.Bytes); err != nil {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("entry changed while it was being adopted: %v", err))
	}
	backup, err := copyAdoptionBackup(home, path, entry.Bytes, entryState.Info)
	if err != nil {
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("backup could not be created under %s: %v", filepath.Join(home, "backups", "global-bins"), err))
	}
	if err := confirmAdoptionEntry(path, entryState.Info, entry.Bytes); err != nil {
		_ = os.Remove(backup)
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("entry changed while it was being adopted: %v", err))
	}

	managed[name] = true
	if err := writeLedger(selection.Path, managed); err != nil {
		_ = os.Remove(backup)
		return Adoption{}, adoptionRefusal(path, fmt.Sprintf("ownership marker %s could not be written: %v", filepath.Join(selection.Path, managedFile), err))
	}
	return Adoption{Path: path, Backup: backup}, nil
}

func adoptionRefusal(path, reason string) error {
	return fmt.Errorf("global: cannot adopt %s: %s", path, reason)
}

func confirmAdoptionEntry(path string, original os.FileInfo, originalBytes []byte) error {
	state, err := stateread.Lstat(path)
	if err != nil {
		return err
	}
	if state.Kind != stateread.KindPresent || state.Info == nil || !state.Info.Mode().IsRegular() {
		return fmt.Errorf("entry is no longer a regular file")
	}
	if !os.SameFile(original, state.Info) {
		return fmt.Errorf("entry identity changed")
	}
	if original.Mode() != state.Info.Mode() || !original.ModTime().Equal(state.Info.ModTime()) {
		return fmt.Errorf("entry metadata changed")
	}
	current, err := stateread.ReadRegularFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(current.Bytes, originalBytes) {
		return fmt.Errorf("entry bytes changed")
	}
	return nil
}

func copyAdoptionBackup(home, source string, payload []byte, sourceInfo os.FileInfo) (string, error) {
	backupBase := filepath.Join(home, "backups")
	if state, err := stateread.Lstat(backupBase); err != nil {
		return "", err
	} else if state.Kind == stateread.KindPresent && (state.Info == nil || !state.Info.IsDir() || state.Info.Mode()&os.ModeSymlink != 0) {
		return "", fmt.Errorf("backup root is not a real directory")
	}
	if err := os.MkdirAll(backupBase, 0o700); err != nil {
		return "", err
	}
	baseState, err := stateread.Lstat(backupBase)
	if err != nil {
		return "", err
	}
	if baseState.Kind != stateread.KindPresent || baseState.Info == nil || !baseState.Info.IsDir() || baseState.Info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("backup root is not a real directory")
	}

	backupRoot := filepath.Join(home, "backups", "global-bins")
	if state, err := stateread.Lstat(backupRoot); err != nil {
		return "", err
	} else if state.Kind == stateread.KindPresent && (state.Info == nil || !state.Info.IsDir() || state.Info.Mode()&os.ModeSymlink != 0) {
		return "", fmt.Errorf("backup directory is not a real directory")
	}
	if err := os.MkdirAll(backupRoot, 0o700); err != nil {
		return "", err
	}
	rootState, err := stateread.Lstat(backupRoot)
	if err != nil {
		return "", err
	}
	if rootState.Kind != stateread.KindPresent || rootState.Info == nil || !rootState.Info.IsDir() || rootState.Info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("backup directory is not a real directory")
	}

	pattern := filepath.Base(source) + "-*.bak"
	backup, err := os.CreateTemp(backupRoot, pattern)
	if err != nil {
		return "", err
	}
	path := backup.Name()
	keep := false
	defer func() {
		_ = backup.Close()
		if !keep {
			_ = os.Remove(path)
		}
	}()
	if _, err := backup.Write(payload); err != nil {
		return "", err
	}
	if err := backup.Chmod(sourceInfo.Mode().Perm()); err != nil {
		return "", err
	}
	if err := backup.Sync(); err != nil {
		return "", err
	}
	if err := backup.Close(); err != nil {
		return "", err
	}
	if err := os.Chtimes(path, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		return "", err
	}
	keep = true
	return path, nil
}
