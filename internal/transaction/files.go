package transaction

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func copyTarget(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("transaction staging path already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return err
	}
	if info.Mode().IsRegular() {
		return copyRegular(source, destination, info.Mode().Perm())
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("unsafe staged transaction source: %s", source)
	}
	if err := os.Mkdir(destination, info.Mode().Perm()); err != nil {
		return err
	}
	if err := copyDirectory(source, destination); err != nil {
		_ = os.RemoveAll(destination)
		return err
	}
	if err := os.Chmod(destination, info.Mode().Perm()); err != nil {
		_ = os.RemoveAll(destination)
		return err
	}
	return syncTree(destination)
}

func copyDirectory(source, destination string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())
		info, err := os.Lstat(sourcePath)
		if err != nil {
			return err
		}
		switch {
		case info.Mode().IsRegular():
			if err := copyRegular(sourcePath, destinationPath, info.Mode().Perm()); err != nil {
				return err
			}
		case info.IsDir() && info.Mode()&os.ModeSymlink == 0:
			if err := os.Mkdir(destinationPath, info.Mode().Perm()); err != nil {
				return err
			}
			if err := copyDirectory(sourcePath, destinationPath); err != nil {
				return err
			}
			if err := os.Chmod(destinationPath, info.Mode().Perm()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsafe staged transaction entry: %s", sourcePath)
		}
	}
	return syncTree(destination)
}

func copyRegular(source, destination string, mode os.FileMode) (result error) {
	input, err := os.Open(source) // #nosec G304 -- caller-selected private staging
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, input.Close()) }()
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode) // #nosec G304 -- deterministic sibling staging
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = output.Close()
		if !ok {
			_ = os.Remove(destination)
		}
	}()
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	if err := output.Chmod(mode); err != nil {
		return err
	}
	if err := output.Sync(); err != nil {
		return err
	}
	if err := output.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

// syncTree flushes one target path. A symbolic link is accepted only as the
// target itself: it holds no data of its own, so its durability is the
// durability of the directory entry that names it. A link found inside a tree is
// still refused, exactly as when digesting or copying one.
func syncTree(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return syncDirectory(filepath.Dir(path))
	}
	return syncTreeEntry(path, info)
}

func syncTreeEntry(path string, info os.FileInfo) error {
	if info.Mode().IsRegular() {
		return syncRegular(path)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("unsafe transaction target while syncing: %s", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		child := filepath.Join(path, entry.Name())
		childInfo, err := os.Lstat(child)
		if err != nil {
			return err
		}
		if err := syncTreeEntry(child, childInfo); err != nil {
			return err
		}
	}
	return syncDirectory(path)
}
