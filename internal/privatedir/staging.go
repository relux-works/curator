package privatedir

import (
	"os"
)

// TempStaging creates an operation-private staging directory for local
// snapshot capture: a fresh temporary directory that is made private and
// proven private before use. The caller owns removal. A failure leaves no
// directory behind.
func TempStaging(parent, prefix string) (string, error) {
	directory, err := os.MkdirTemp(parent, prefix)
	if err != nil {
		return "", err
	}
	if err := Protect(directory); err != nil {
		_ = os.RemoveAll(directory)
		return "", err
	}
	if err := Validate(directory); err != nil {
		_ = os.RemoveAll(directory)
		return "", err
	}
	return directory, nil
}
