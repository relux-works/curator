//go:build !windows

package buildrepo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

type nativeProtectedDirGuard struct{}

func openNativeProtectedDirGuard(_ []string, _ func(string)) (*nativeProtectedDirGuard, error) {
	return &nativeProtectedDirGuard{}, nil
}

func (*nativeProtectedDirGuard) validate() error { return nil }
func (*nativeProtectedDirGuard) close()          {}

func regularFileIdentity(info os.FileInfo, directory bool) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Geteuid() {
		return false
	}
	return directory || uint64(stat.Nlink) == 1
}

func protectedFileIdentity(info os.FileInfo, directory bool) bool {
	return regularFileIdentity(info, directory)
}

func nativeProtectedDir(name string, create bool, hook func(string)) error {
	if create {
		if err := os.MkdirAll(name, 0o700); err != nil {
			return err
		}
	}
	info, err := os.Lstat(name)
	if err != nil {
		return err
	}
	if hook != nil {
		hook("directory-proof")
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o022 != 0 || !protectedFileIdentity(info, true) {
		return admissionError(CodeProtectedBoundaryUntrusted, "protected directory cannot be proved private")
	}
	return nil
}

func readNativeProtectedFile(name string, maxBytes int64, hook func(string)) ([]byte, error) {
	info, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxBytes || !protectedFileIdentity(info, false) {
		return nil, fmt.Errorf("protected file shape invalid")
	}
	if hook != nil {
		hook("file-proof")
	}
	file, err := os.Open(name) // #nosec G304 -- manager-derived protected path was lstat and ownership-proved above.
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("protected file shape invalid")
	}
	return data, nil
}

func secureProtectedTree(root string) error {
	return filepath.WalkDir(root, func(_ string, _ os.DirEntry, err error) error { return err })
}
