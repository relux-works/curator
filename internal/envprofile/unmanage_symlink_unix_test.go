//go:build unix

package envprofile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUnmanageRestoreSymlinkReplacesDestinationEntry(t *testing.T) {
	native, outside := t.TempDir(), t.TempDir()
	backup := filepath.Join(native, ".agent-environment-backup", "1", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(backup), 0o700); err != nil {
		t.Fatal(err)
	}
	// Preserve the literal relative target, even though it dangles at the
	// backup location. Restoring must not read or copy target bytes.
	want := "../operator-context.md"
	if err := os.Symlink(want, backup); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(outside, "managed-target.md")
	if err := os.WriteFile(external, []byte("external managed context\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(external)
	if err != nil {
		t.Fatal(err)
	}
	full := filepath.Join(native, "CLAUDE.md")
	if err := os.Symlink(external, full); err != nil {
		t.Fatal(err)
	}
	req := UnmanageRequest{
		Home: t.TempDir(), EnvID: "claude_code", RestoreBackups: true,
		NativeHomeOf: func(string) (string, error) { return native, nil },
	}
	plans, err := planUnmanageHomes(req)
	if err != nil {
		t.Fatal(err)
	}
	entry := plans[0].restores["CLAUDE.md"]
	if entry.kind != os.ModeSymlink || entry.linkTarget != want || len(entry.payload) != 0 {
		t.Fatalf("restore entry = %+v; want symlink with literal target", entry)
	}
	result, err := Unmanage(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Homes) != 1 || len(result.Homes[0].Restored) != 1 || result.Homes[0].Restored[0] != "CLAUDE.md" {
		t.Fatalf("restore report = %+v", result)
	}
	payload, err := os.ReadFile(external)
	if err != nil || string(payload) != "external managed context\n" {
		t.Fatalf("restore changed destination link target: %q (%v)", payload, err)
	}
	after, err := os.Stat(external)
	if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatalf("restore changed destination target metadata: %v", err)
	}
	for _, name := range []string{full, backup} {
		got, err := os.Readlink(name)
		if err != nil || got != want {
			t.Fatalf("link target = %q (%v), want %q", got, err, want)
		}
	}
}
