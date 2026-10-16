//go:build !windows

package envprofile

import (
	"os"
	"testing"
)

func replaceCapturedFixture(t *testing.T, source, target string) error {
	t.Helper()
	return os.Rename(source, target)
}
