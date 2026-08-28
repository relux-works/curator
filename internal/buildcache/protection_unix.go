//go:build unix

package buildcache

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func protectionSupported() bool { return true }

func openProtectedEntry(home, entryPath, artifactRel string) (*openedEntry, error) {
	rel, err := filepath.Rel(home, entryPath)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, untrustedf("cache entry crosses the manager-home boundary")
	}
	rootFD, err := unix.Open(home, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, untrustedf("open manager home without following links: %v", err)
	}
	root := os.NewFile(uintptr(rootFD), home)
	if err := validateUnixDir(rootFD, "manager home"); err != nil {
		_ = root.Close()
		return nil, err
	}

	opened := &openedEntry{extra: []io.Closer{root}}
	currentFD := rootFD
	parts := strings.Split(rel, string(filepath.Separator))
	for index, part := range parts {
		if part == "" || part == "." || part == ".." {
			opened.close()
			return nil, untrustedf("cache entry contains an invalid path component")
		}
		fd, openErr := unix.Openat(currentFD, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			opened.close()
			if errors.Is(openErr, unix.ENOENT) {
				return nil, fmt.Errorf("cache entry is incomplete: %w", openErr)
			}
			return nil, untrustedf("open cache directory %q without following links: %v", part, openErr)
		}
		if err := validateUnixDir(fd, part); err != nil {
			_ = unix.Close(fd)
			opened.close()
			return nil, err
		}
		file := os.NewFile(uintptr(fd), filepath.Join(home, filepath.Join(parts[:index+1]...)))
		if index == len(parts)-1 {
			opened.entryDir = file
		} else {
			opened.extra = append(opened.extra, file)
		}
		currentFD = fd
	}

	receipt, err := openUnixProtectedFile(currentFD, ReceiptFilename, false)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.receipt = receipt
	executionReceipt, err := openUnixProtectedFile(currentFD, ExecutionReceiptFilename, false)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.executionReceipt = executionReceipt

	binFD, err := unix.Openat(currentFD, "bin", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		opened.close()
		if errors.Is(err, unix.ENOENT) {
			return nil, fmt.Errorf("cache entry is incomplete: artifact directory is absent")
		}
		return nil, untrustedf("open artifact directory without following links: %v", err)
	}
	if err := validateUnixDir(binFD, "artifact directory"); err != nil {
		_ = unix.Close(binFD)
		opened.close()
		return nil, err
	}
	opened.binDir = os.NewFile(uintptr(binFD), filepath.Join(entryPath, "bin"))

	artifactName, err := artifactChildName(artifactRel)
	if err != nil {
		opened.close()
		return nil, err
	}
	artifact, err := openUnixProtectedFile(binFD, artifactName, true)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.artifact = artifact
	return opened, nil
}

// openProtectedEntryFrom opens the receipt, artifact directory, and artifact of
// an already-proven cache entry relative to that entry's own descriptor.
//
// Nothing here resolves a pathname. Every child is opened with openat from the
// directory the caller proved, without following a link, so an exchange of the
// cache-root or entry pathname cannot substitute a different tree for the one
// being classified. The entry descriptor stays owned by the caller.
func openProtectedEntryFrom(entry *protectedDir, artifactRel string) (*openedEntry, error) {
	if entry == nil || entry.dir == nil {
		return nil, untrustedf("protected cache entry handle is missing")
	}
	entryFD := int(entry.dir.Fd())
	if err := validateUnixDir(entryFD, "cache entry"); err != nil {
		return nil, err
	}
	artifactName, err := artifactChildName(artifactRel)
	if err != nil {
		return nil, err
	}

	opened := &openedEntry{entryDir: entry.dir, borrowedEntryDir: true}
	receipt, err := openUnixProtectedFile(entryFD, ReceiptFilename, false)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.receipt = receipt
	executionReceipt, err := openUnixProtectedFile(entryFD, ExecutionReceiptFilename, false)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.executionReceipt = executionReceipt

	binFD, err := unix.Openat(entryFD, "bin", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		opened.close()
		if errors.Is(err, unix.ENOENT) {
			return nil, fmt.Errorf("cache entry is incomplete: artifact directory is absent")
		}
		return nil, untrustedf("open artifact directory without following links: %v", err)
	}
	if err := validateUnixDir(binFD, "artifact directory"); err != nil {
		_ = unix.Close(binFD)
		opened.close()
		return nil, err
	}
	opened.binDir = os.NewFile(uintptr(binFD), filepath.Join(entry.dir.Name(), "bin"))

	artifact, err := openUnixProtectedFile(binFD, artifactName, true)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.artifact = artifact
	return opened, nil
}

// openProtectedDir resolves dirPath below home without following a link at any
// component and validates the type, owner, and permissions of each one. It is
// the traversal boundary maintenance revalidates before reading the cache root
// or classifying one entry, and it never creates or repairs anything.
func openProtectedDir(home, dirPath string) (*protectedDir, error) {
	rel, err := filepath.Rel(home, dirPath)
	if err != nil || rel == "." || rel == ".." ||
		strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return nil, untrustedf("protected directory crosses the manager-home boundary")
	}
	rootFD, err := unix.Open(home, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		if errors.Is(err, unix.ENOENT) {
			return nil, fmt.Errorf("open manager home: %w", os.ErrNotExist)
		}
		return nil, untrustedf("open manager home without following links: %v", err)
	}
	opened := &protectedDir{parents: []*os.File{os.NewFile(uintptr(rootFD), home)}}
	if err := validateUnixDir(rootFD, "manager home"); err != nil {
		opened.close()
		return nil, err
	}
	currentFD := rootFD
	parts := strings.Split(rel, string(filepath.Separator))
	for index, part := range parts {
		if part == "" || part == "." || part == ".." {
			opened.close()
			return nil, untrustedf("protected directory contains an invalid path component")
		}
		fd, openErr := unix.Openat(currentFD, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if openErr != nil {
			opened.close()
			if errors.Is(openErr, unix.ENOENT) {
				return nil, fmt.Errorf("open protected directory %q: %w", part, os.ErrNotExist)
			}
			return nil, untrustedf("open protected directory %q without following links: %v", part, openErr)
		}
		if err := validateUnixDir(fd, part); err != nil {
			_ = unix.Close(fd)
			opened.close()
			return nil, err
		}
		file := os.NewFile(uintptr(fd), filepath.Join(home, filepath.Join(parts[:index+1]...)))
		if index == len(parts)-1 {
			opened.dir = file
		} else {
			opened.parents = append(opened.parents, file)
		}
		currentFD = fd
	}
	return opened, nil
}

// openProtectedChildFile opens one non-executable regular file inside an
// already-validated protected directory handle, without following a link.
func openProtectedChildFile(dir *os.File, name string) (*os.File, error) {
	if dir == nil {
		return nil, untrustedf("protected directory handle is missing")
	}
	return openUnixProtectedFile(int(dir.Fd()), name, false)
}

func openUnixProtectedFile(parentFD int, name string, executable bool) (*os.File, error) {
	fd, err := unix.Openat(parentFD, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		if errors.Is(err, unix.ENOENT) {
			return nil, fmt.Errorf("cache entry is incomplete: %s is absent", name)
		}
		return nil, untrustedf("open protected file %q without following links: %v", name, err)
	}
	if err := validateUnixFile(fd, name, executable); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	return os.NewFile(uintptr(fd), name), nil
}

func validateUnixDir(fd int, label string) error {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return untrustedf("stat %s: %v", label, err)
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFDIR {
		return untrustedf("%s is not a directory", label)
	}
	if int(stat.Uid) != os.Geteuid() {
		return untrustedf("%s owner does not match the effective user", label)
	}
	if stat.Mode&0o022 != 0 {
		return untrustedf("%s is writable by group or other", label)
	}
	return nil
}

func validateUnixFile(fd int, label string, executable bool) error {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return untrustedf("stat %s: %v", label, err)
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG {
		return untrustedf("%s is not a regular file", label)
	}
	if stat.Nlink != 1 {
		return untrustedf("%s has multiple hard links", label)
	}
	if int(stat.Uid) != os.Geteuid() {
		return untrustedf("%s owner does not match the effective user", label)
	}
	if stat.Mode&0o022 != 0 {
		return untrustedf("%s is writable by group or other", label)
	}
	if executable && stat.Mode&0o100 == 0 {
		return untrustedf("artifact is not executable by its owner")
	}
	return nil
}

func ensureProtectedBase(home, base string) error {
	rel, err := filepath.Rel(home, base)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return untrustedf("cache root crosses the manager-home boundary")
	}
	rootFD, err := unix.Open(home, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return untrustedf("open manager home without following links: %v", err)
	}
	defer func() { _ = unix.Close(rootFD) }()
	if err := validateUnixDir(rootFD, "manager home"); err != nil {
		return err
	}
	currentFD := rootFD
	var opened []int
	defer func() {
		for index := len(opened) - 1; index >= 0; index-- {
			_ = unix.Close(opened[index])
		}
	}()
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		fd, openErr := unix.Openat(currentFD, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if errors.Is(openErr, unix.ENOENT) {
			if err := unix.Mkdirat(currentFD, part, 0o700); err != nil && !errors.Is(err, unix.EEXIST) {
				return fmt.Errorf("create protected cache directory %q: %w", part, err)
			}
			fd, openErr = unix.Openat(currentFD, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		}
		if openErr != nil {
			return untrustedf("open protected cache directory %q: %v", part, openErr)
		}
		if err := validateUnixDir(fd, part); err != nil {
			_ = unix.Close(fd)
			return err
		}
		opened = append(opened, fd)
		currentFD = fd
	}
	return nil
}

func makeProtectedTempDir(parent, pattern string) (string, error) {
	path, err := os.MkdirTemp(parent, pattern)
	if err != nil {
		return "", err
	}
	// #nosec G302 -- a protected cache directory is owner-only by design; the
	// linter's 0600 ceiling is for files and would make the directory
	// unenterable.
	if err := os.Chmod(path, 0o700); err != nil {
		_ = os.RemoveAll(path)
		return "", err
	}
	return path, nil
}

func createProtectedDir(path string) error {
	if err := os.Mkdir(path, 0o700); err != nil {
		return err
	}
	// #nosec G302 -- owner-only directory; see makeProtectedTempDir.
	return os.Chmod(path, 0o700)
}

func writeProtectedFile(path string, mode os.FileMode, source io.Reader) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode) // #nosec G304 -- path is manager-derived staging
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = file.Close()
		if !ok {
			_ = os.Remove(path)
		}
	}()
	if err := file.Chmod(mode); err != nil {
		return err
	}
	if _, err := io.Copy(file, source); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func syncDirectory(path string) error {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return err
	}
	defer func() { _ = unix.Close(fd) }()
	return unix.Fsync(fd)
}

// syncDirHandle durably records a directory change through an already-proven
// handle, so the sweep never has to re-resolve the pathname it just mutated.
func syncDirHandle(dir *os.File) error {
	if dir == nil {
		return fmt.Errorf("directory handle is missing")
	}
	return unix.Fsync(int(dir.Fd()))
}

func renameDirectoryNoReplace(from, to string) error {
	return os.Rename(from, to)
}
