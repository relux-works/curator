//go:build unix

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Drive run(), including install and takeover: the mode must survive both
// backup creation and unmanage restoration, not just a helper invocation.
// Windows FileMode does not express DACLs; these are Unix permission checks.
func TestEnvUnmanageTakeoverPreservesFileMode(t *testing.T) {
	for _, mode := range []os.FileMode{0o600, 0o644, 0o751, 0o400} {
		t.Run(fmt.Sprintf("%04o", mode), func(t *testing.T) {
			source, _ := profileHome(t)
			pkg := t.TempDir()
			writeContextPackage(t, pkg, "acme", "1.0.0", "managed context\n")
			if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
				t.Fatalf("profile install exit %d: %s", code, stderr)
			}
			// Install may activate the first profile. Unmanage before placing the
			// hand-maintained file so --takeover sees an operator-owned entry.
			if code, _, stderr := runProfile(t, source, "env", "unmanage", "--env", "claude_code"); code != exitOK {
				t.Fatalf("initial unmanage exit %d: %s", code, stderr)
			}
			native := os.Getenv("CLAUDE_CONFIG_DIR")
			full := filepath.Join(native, "CLAUDE.md")
			original := "operator-owned private context\n"
			if err := os.WriteFile(full, []byte(original), 0o600); err != nil {
				t.Fatal(err)
			}
			// Set exact bits independently of the process umask.
			if err := os.Chmod(full, mode); err != nil {
				t.Fatal(err)
			}
			if code, _, stderr := runProfile(t, source, "profile", "use", "acme", "--takeover", "--env", "claude_code"); code != exitOK {
				t.Fatalf("profile use --takeover exit %d: %s", code, stderr)
			}
			backup := filepath.Join(native, ".agent-environment-backup", "1", "CLAUDE.md")
			assertUnmanageFileMode(t, backup, mode)
			if code, _, stderr := runProfile(t, source, "env", "unmanage", "--restore-backups", "--env", "claude_code"); code != exitOK {
				t.Fatalf("env unmanage --restore-backups exit %d: %s", code, stderr)
			}
			assertUnmanageFileMode(t, full, mode)
			assertUnmanageFileMode(t, backup, mode)
			if got := string(readFileForUnmanageTest(t, full)); got != original {
				t.Fatalf("restored bytes = %q, want %q", got, original)
			}
		})
	}
}

func assertUnmanageFileMode(t *testing.T, full string, want os.FileMode) {
	t.Helper()
	info, err := os.Lstat(full)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != want {
		t.Fatalf("%s: type/mode = %v, want regular file %04o", filepath.Base(full), info.Mode(), want)
	}
}
