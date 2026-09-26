package contextlock

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

func TestReadDistinguishesAbsentAndUnreadableLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lock.json")
	_, _, absentErr := Read(path)
	var absentState *stateread.Error
	if absentErr == nil || !errors.As(absentErr, &absentState) || absentState.Kind != stateread.KindAbsent || !strings.Contains(absentErr.Error(), stateread.DiagAbsent) {
		t.Fatalf("absent lock error = %v, want typed absence", absentErr)
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Skipf("host-capability: this environment can read a mode-000 file: chmod failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("host-capability: this environment can read a mode-000 file; permission-only lock row unavailable")
	}
	_, _, err := Read(path)
	var stateErr *stateread.Error
	if err == nil || !errors.As(err, &stateErr) || stateErr.Kind != stateread.KindUnreadable || !strings.Contains(err.Error(), path) {
		t.Fatalf("unreadable lock error = %v, want typed unreadable diagnostic for %s", err, path)
	}
}
