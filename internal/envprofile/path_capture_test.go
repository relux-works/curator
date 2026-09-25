package envprofile

// Profile path-source fixtures for the accepted environments snapshot
// contract; these do not exercise the project Skillfile source protocol.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/relux-works/curator/internal/contextstore"
)

// pathSourceGit runs git inside dir, skipping when git is unavailable.
func pathSourceGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git on PATH")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// TestPathInstallCapturesDirtyUntrackedInsideGit proves profile path
// installs preserve the environment contract when the source is a Git
// checkout: the installed entry carries dirty and untracked bytes, not
// only the Git HEAD tree.
func TestPathInstallCapturesDirtyUntrackedInsideGit(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeSeedPackage(t, root, "acme")
	pathSourceGit(t, root, "init", "-q", "-b", "main")
	pathSourceGit(t, root, "add", ".")
	pathSourceGit(t, root, "commit", "-qm", "one")
	if err := os.WriteFile(filepath.Join(root, "context", "a.md"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "context", "untracked.md"), []byte("untracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _, _, err := Install(home, InstallOptions{Operand: root})
	if err != nil {
		t.Fatalf("path install of a dirty checkout: %v", err)
	}
	member, ok := lockMember(info.Lock, "acme")
	if !ok {
		t.Fatalf("lock members %+v carry no root", info.Lock.Members)
	}
	entry := contextstore.EntryDir(home, member.Kind, member.Name, member.StateHash)
	if payload, err := os.ReadFile(filepath.Join(entry, "context", "a.md")); err != nil || string(payload) != "dirty\n" {
		t.Fatalf("installed a.md = %q, %v; want dirty bytes, not HEAD", payload, err)
	}
	if payload, err := os.ReadFile(filepath.Join(entry, "context", "untracked.md")); err != nil || string(payload) != "untracked\n" {
		t.Fatalf("installed untracked.md = %q, %v; want untracked bytes", payload, err)
	}
	if _, err := os.Lstat(filepath.Join(entry, ".git")); !os.IsNotExist(err) {
		t.Fatalf("installed entry carries .git metadata: %v", err)
	}
}
