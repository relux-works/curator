//go:build unix

package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnmanageRestorePlanRetainsMetadata(t *testing.T) {
	native := t.TempDir()
	backup := filepath.Join(native, ".agent-environment-backup", "1", "CLAUDE.md")
	if err := os.MkdirAll(filepath.Dir(backup), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backup, []byte("private context\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	plans, err := planUnmanageHomes(UnmanageRequest{
		EnvID: "claude_code", RestoreBackups: true,
		NativeHomeOf: func(string) (string, error) { return native, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := plans[0].restores["CLAUDE.md"]
	if !ok || entry.kind != 0 || entry.mode != 0o600 || string(entry.payload) != "private context\n" {
		t.Fatalf("restore entry = %+v, present = %v; want regular 0600 with backup bytes", entry, ok)
	}
}

// The production Unmanage entry point must refuse a parent link before any
// restore writes, even though the final regular-file replacement is atomic.
func TestUnmanageRestoreRefusesLinkedParent(t *testing.T) {
	native, outside := t.TempDir(), t.TempDir()
	backup := filepath.Join(native, ".agent-environment-backup", "1", "context", "private.md")
	if err := os.MkdirAll(filepath.Dir(backup), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backup, []byte("backup context\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(outside, "private.md")
	if err := os.WriteFile(external, []byte("external context\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(native, "context")); err != nil {
		t.Fatal(err)
	}
	_, err := Unmanage(UnmanageRequest{
		Home: t.TempDir(), EnvID: "claude_code", RestoreBackups: true,
		NativeHomeOf: func(string) (string, error) { return native, nil },
	})
	if err == nil || !strings.Contains(err.Error(), DiagWriteWouldFollowLink) {
		t.Fatalf("Unmanage error = %v, want parent-link refusal", err)
	}
	payload, err := os.ReadFile(external)
	if err != nil || string(payload) != "external context\n" {
		t.Fatalf("external context changed: %q (%v)", payload, err)
	}
}
