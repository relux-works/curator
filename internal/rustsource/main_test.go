package rustsource

import (
	"os"
	"testing"

	"github.com/relux-works/curator/internal/testgitenv"
)

func requireNativeCargoDescriptor(t *testing.T) {
	t.Helper()
	if reason := NativeCargoUnavailableReason(); reason != "" {
		t.Skip(reason)
	}
}

func TestMain(m *testing.M) {
	if handled, code := DispatchInternalWorker(os.Args[1:], os.Stdin, os.Stdout); handled {
		os.Exit(code)
	}
	restoreGitEnv := testgitenv.Isolate()
	code := m.Run()
	restoreGitEnv()
	os.Exit(code)
}
