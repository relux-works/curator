package envprofile

// Production-path proof for draft local-snapshot capture (revision 1):
// every test drives the public Install entry with DraftSourcesV1 on.
// The draft install must freeze the admitted bytes at capture and
// install from the frozen snapshot, never the live directory.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/snapshot"
)

// draftGit runs git inside dir, skipping when git is unavailable.
func draftGit(t *testing.T, dir string, args ...string) {
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

// TestDraftInstallFreezesLiveMutations proves the install consumes the
// captured snapshot, not the live directory: the live source is mutated
// after the snapshot freezes but before the store consumes it, inside a
// single Install, and the installed entry must still carry the frozen
// bytes. Installing the live directory instead of the snapshot serves
// the mutated bytes, so that narrowing must fail this test.
func TestDraftInstallFreezesLiveMutations(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeSeedPackage(t, root, "acme")
	var hookErr error
	admitPathSourceAfterFreezeHook = func() {
		hookErr = os.WriteFile(filepath.Join(root, "context", "a.md"), []byte("mutated\n"), 0o644)
	}
	defer func() { admitPathSourceAfterFreezeHook = nil }()
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: Policy{DraftSourcesV1: true}})
	if hookErr != nil {
		t.Fatalf("race hook: %v", hookErr)
	}
	if err != nil {
		t.Fatalf("draft install: %v", err)
	}
	member, ok := lockMember(info.Lock, "acme")
	if !ok {
		t.Fatalf("lock members %+v carry no root", info.Lock.Members)
	}
	entry := contextstore.EntryDir(home, member.Kind, member.Name, member.StateHash)
	installed, err := os.ReadFile(filepath.Join(entry, "context", "a.md"))
	if err != nil || string(installed) != "acme\n" {
		t.Fatalf("installed a.md = %q, %v; want frozen acme bytes, not the live mutation", installed, err)
	}
	// Sanity: the live mutation really landed before consumption, so a
	// live-directory install would have served it.
	if live, err := os.ReadFile(filepath.Join(root, "context", "a.md")); err != nil || string(live) != "mutated\n" {
		t.Fatalf("live a.md = %q, %v; want the raced mutated bytes", live, err)
	}
	// The capture must have published exactly one immutable snapshot
	// below the manager home.
	store, err := os.ReadDir(snapshot.LocalStoreDir(home))
	if err != nil || len(store) != 1 {
		t.Fatalf("local-snapshots = %v, %v; want one immutable entry", store, err)
	}
}

// TestDraftInstallAdmitsDirtyUntrackedInsideGit proves the draft path
// install snapshots dirty and untracked bytes even when the source is
// a Git checkout: the installed entry carries the live bytes, not
// HEAD.
func TestDraftInstallAdmitsDirtyUntrackedInsideGit(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeSeedPackage(t, root, "acme")
	draftGit(t, root, "init", "-q", "-b", "main")
	draftGit(t, root, "add", ".")
	draftGit(t, root, "commit", "-qm", "one")
	if err := os.WriteFile(filepath.Join(root, "context", "a.md"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "context", "untracked.md"), []byte("untracked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, _, _, err := Install(home, InstallOptions{Operand: root, Policy: Policy{DraftSourcesV1: true}})
	if err != nil {
		t.Fatalf("draft install of a dirty checkout: %v", err)
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

// TestDraftInstallRefusesSnapshotInsideStore proves the snapshot
// store itself is a managed output: selecting a package inside it is
// refused with the boundary diagnostic and publishes nothing.
func TestDraftInstallRefusesSnapshotInsideStore(t *testing.T) {
	home := t.TempDir()
	pinHomes(t)
	root := filepath.Join(t.TempDir(), "root")
	writeSeedPackage(t, root, "acme")
	if _, _, _, err := Install(home, InstallOptions{Operand: root, Policy: Policy{DraftSourcesV1: true}}); err != nil {
		t.Fatalf("seed install: %v", err)
	}
	entries, err := os.ReadDir(snapshot.LocalStoreDir(home))
	if err != nil || len(entries) != 1 {
		t.Fatalf("local-snapshots = %v, %v; want the seeded snapshot", entries, err)
	}
	// A package rooted inside the snapshot store is managed output.
	inner := filepath.Join(snapshot.LocalStoreDir(home), entries[0].Name(), "snapshot")
	_, _, _, err = Install(home, InstallOptions{Operand: inner, Policy: Policy{DraftSourcesV1: true}})
	if err == nil || !strings.Contains(err.Error(), "source_output_overlap") {
		t.Fatalf("install inside the snapshot store err = %v, want source_output_overlap", err)
	}
}
