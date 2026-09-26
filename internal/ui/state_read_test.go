package ui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/stateread"
)

func TestSkillsUnderDistinguishesAbsentAndUnreadableMarkers(t *testing.T) {
	dir := t.TempDir()
	installed := filepath.Join(dir, "skill-a")
	if err := os.Mkdir(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	if rows := skillsUnder(dir); len(rows) != 0 {
		t.Fatalf("absent marker rows = %+v, want no installed skills", rows)
	}
	path := filepath.Join(installed, marker.Name)
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
		t.Skip("host-capability: this environment can read a mode-000 file; permission-only marker row unavailable")
	}
	rows := skillsUnder(dir)
	if len(rows) != 1 || rows[0].Name != "skill-a" || !strings.Contains(rows[0].Diagnostic, stateread.DiagUnreadable) || !strings.Contains(rows[0].Diagnostic, path) {
		t.Fatalf("unreadable marker rows = %+v, want a typed diagnostic for %s", rows, path)
	}
}

func TestSkillsUnderReportsInvalidMarkerSeparately(t *testing.T) {
	dir := t.TempDir()
	installed := filepath.Join(dir, "skill-a")
	if err := os.Mkdir(installed, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(installed, marker.Name)
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rows := skillsUnder(dir)
	if len(rows) != 1 || rows[0].Name != "skill-a" || !strings.Contains(rows[0].Diagnostic, marker.DiagInvalid) || strings.Contains(rows[0].Diagnostic, stateread.DiagUnreadable) {
		t.Fatalf("invalid marker rows = %+v, want one typed invalid marker diagnostic", rows)
	}
}
