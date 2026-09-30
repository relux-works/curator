package atomicity

import (
	"os"
	"testing"

	"github.com/relux-works/curator/internal/testgitenv"
)

// TestMain keeps every git this package's tests run off the operator's
// global and system git config; see internal/testgitenv.
func TestMain(m *testing.M) {
	restoreGitEnv := testgitenv.Isolate()
	code := m.Run()
	restoreGitEnv()
	os.Exit(code)
}
