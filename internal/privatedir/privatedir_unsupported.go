//go:build !unix && !windows

package privatedir

import (
	"fmt"
	"os"
)

// A platform with neither Unix permission bits nor Windows DACLs cannot
// express the private shape, so every operation fails closed.

func makePrivate(string) error {
	return fmt.Errorf("private directories are unsupported on this platform")
}

func makeAllPrivate(string) error {
	return fmt.Errorf("private directories are unsupported on this platform")
}

func validatePrivate(string) error {
	return fmt.Errorf("private directories are unsupported on this platform")
}

func protectPrivate(string) error {
	return fmt.Errorf("private directories are unsupported on this platform")
}

func createPrivateTempFile(string, string) (*os.File, error) {
	return nil, fmt.Errorf("private files are unsupported on this platform")
}

func validatePrivateFile(string) error {
	return fmt.Errorf("private files are unsupported on this platform")
}

func protectPrivateFile(string) error {
	return fmt.Errorf("private files are unsupported on this platform")
}

func protectPrivateTree(string) error {
	return fmt.Errorf("private trees are unsupported on this platform")
}
