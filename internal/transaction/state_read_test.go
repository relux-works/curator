package transaction

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

func TestJournalInventoryDistinguishesAbsentAndUnreadable(t *testing.T) {
	engine := mustEngine(t, t.TempDir())
	if ids, err := engine.journalIDs(); err != nil || len(ids) != 0 {
		t.Fatalf("journalIDs on absent journal directory = (%v, %v), want empty, nil", ids, err)
	}
	if err := os.MkdirAll(engine.journalRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		t.Skip("platform-control: POSIX mode-bit unreadability")
	}
	if err := os.Chmod(engine.journalRoot, 0); err != nil {
		t.Skipf("this environment cannot set unreadable mode: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(engine.journalRoot, 0o700) })
	if _, err := os.ReadDir(engine.journalRoot); err == nil {
		t.Skip("this environment can read a mode-000 directory; unreadability is untestable here")
	}
	if ids, err := engine.journalIDs(); err == nil || ids != nil || !strings.Contains(err.Error(), stateread.DiagUnreadable) || !strings.Contains(err.Error(), filepath.Clean(engine.journalRoot)) {
		t.Fatalf("journalIDs on unreadable journal directory = (%v, %v), want typed unreadable error", ids, err)
	}
}
