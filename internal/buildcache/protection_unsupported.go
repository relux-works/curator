//go:build !unix && !windows

package buildcache

import (
	"fmt"
	"io"
	"os"
)

func protectionSupported() bool { return false }

func openProtectedEntry(string, string, string) (*openedEntry, error) {
	return nil, fmt.Errorf("persistent build cache protection is unsupported")
}

func openProtectedEntryFrom(*protectedDir, string) (*openedEntry, error) {
	return nil, fmt.Errorf("persistent build cache protection is unsupported")
}

func openProtectedDir(string, string) (*protectedDir, error) {
	return nil, fmt.Errorf("persistent build cache protection is unsupported")
}

func openProtectedChildFile(*os.File, string) (*os.File, error) {
	return nil, fmt.Errorf("persistent build cache protection is unsupported")
}

func ensureProtectedBase(string, string) error {
	return fmt.Errorf("persistent build cache protection is unsupported")
}

func makeProtectedTempDir(string, string) (string, error) {
	return "", fmt.Errorf("persistent build cache protection is unsupported")
}

func createProtectedDir(string) error {
	return fmt.Errorf("persistent build cache protection is unsupported")
}

func writeProtectedFile(string, os.FileMode, io.Reader) error {
	return fmt.Errorf("persistent build cache protection is unsupported")
}

func syncDirectory(string) error {
	return fmt.Errorf("persistent build cache protection is unsupported")
}

func syncDirHandle(*os.File) error {
	return fmt.Errorf("persistent build cache protection is unsupported")
}

func renameDirectoryNoReplace(string, string) error {
	return fmt.Errorf("persistent build cache protection is unsupported")
}
