package contextstore

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

func TestExistsDistinguishesAbsentFromUnreadable(t *testing.T) {
	home := t.TempDir()
	const (
		kind   = "skill"
		name   = "skill-a"
		pinKey = "pin"
	)
	entry := EntryDir(home, kind, name, pinKey)
	if exists, err := Exists(home, kind, name, pinKey); err != nil || exists {
		t.Fatalf("absent context entry = (%v, %v), want false and nil", exists, err)
	}
	if err := os.MkdirAll(entry, 0o755); err != nil {
		t.Fatal(err)
	}
	parent := filepath.Dir(entry)
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	if err := os.Chmod(parent, 0o000); err != nil {
		t.Skipf("this host cannot create mode-000 manager state: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
	if _, err := os.Stat(entry); err == nil {
		t.Skip("root ignores directory permissions; unreadability is untestable here")
	}
	if exists, err := Exists(home, kind, name, pinKey); exists || err == nil ||
		!strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), entry) {
		t.Fatalf("unreadable context entry = (%v, %v), want typed unreadable state for %s", exists, err, entry)
	}
}
