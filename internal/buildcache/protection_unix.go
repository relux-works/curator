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

	artifactParts := strings.Split(filepath.Clean(filepath.FromSlash(artifactRel)), string(filepath.Separator))
	if len(artifactParts) != 2 || artifactParts[0] != "bin" || artifactParts[1] == "" || artifactParts[1] == "." || artifactParts[1] == ".." {
		opened.close()
		return nil, untrustedf("artifact path is not a direct bin child")
	}
	artifact, err := openUnixProtectedFile(binFD, artifactParts[1], true)
	if err != nil {
		opened.close()
		return nil, err
	}
	opened.artifact = artifact
	return opened, nil
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
	defer unix.Close(rootFD)
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
	defer unix.Close(fd)
	return unix.Fsync(fd)
}

func renameDirectoryNoReplace(from, to string) error {
	return os.Rename(from, to)
}
