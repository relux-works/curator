package gitops

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// extractWatchdog bounds the deadlock class the review found: a refusal must
// return, not leave git blocked on a full stdout pipe. 20 s is far beyond
// any healthy extraction of the few-MiB trees these tests commit.
const extractWatchdog = 20 * time.Second

// extractBounded runs Extract under the watchdog and fails the test if it
// does not return in time.
func extractBounded(t *testing.T, repo, commit, dest string) (error, time.Duration) {
	t.Helper()
	start := time.Now()
	done := make(chan error, 1)
	go func() { done <- Extract(repo, commit, dest) }()
	select {
	case err := <-done:
		return err, time.Since(start)
	case <-time.After(extractWatchdog):
		t.Fatalf("Extract did not return within %s: cat-file deadlock", extractWatchdog)
		return nil, 0
	}
}

// largeBlobOID writes a blob larger than any pipe buffer (8 MiB) so a
// refusal ahead of it would deadlock a reader that stops draining.
func largeBlobOID(t *testing.T, repo string) string {
	t.Helper()
	return blobOID(t, repo, strings.Repeat("0123456789abcdef", 8<<16))
}

// commitRawTreeMissing is commitRawTree for trees that reference objects the
// repository does not hold (mktree --missing).
func commitRawTreeMissing(t *testing.T, repo string, lines []string) string {
	t.Helper()
	cmd := exec.Command("git", "mktree", "-z", "--missing")
	cmd.Dir = repo
	cmd.Env = gitEnv()
	cmd.Stdin = strings.NewReader(strings.Join(lines, "\x00") + "\x00")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mktree --missing: %v\n%s", err, out)
	}
	return gitRun(t, repo, "commit-tree", strings.TrimSpace(string(out)), "-m", "missing")
}

// assertNothingWritten fails when dest exists with any entry, or when a
// refused extraction left a sibling behind in parent.
func assertNothingWritten(t *testing.T, parent, dest string) {
	t.Helper()
	if entries, err := os.ReadDir(dest); err == nil && len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("refused extraction wrote %v under %s", names, dest)
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if filepath.Join(parent, entry.Name()) != dest {
			t.Fatalf("refused extraction left %s beside the destination", entry.Name())
		}
	}
}

func TestExtractRefusesOversizeBlobWithoutStreaming(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, strings.Repeat("x", 1<<20))
	commit := commitRawTree(t, repo, []string{"100644 blob " + oid + "\tbig.bin"})
	saved := maxSnapshotFileBytes
	t.Cleanup(func() { maxSnapshotFileBytes = saved })
	maxSnapshotFileBytes = 512 << 10
	parent := t.TempDir()
	dest := filepath.Join(parent, "snap")
	err, took := extractBounded(t, repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("err = %v, want size refusal", err)
	}
	t.Logf("single oversize blob refused in %s", took)
	assertNothingWritten(t, parent, dest)
}

func TestExtractRefusesDuplicatePlatformPathBeforeLargeBlob(t *testing.T) {
	probe := filepath.Join(t.TempDir(), "probe")
	if err := os.WriteFile(probe, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(probe), "PROBE")); err != nil {
		t.Skip("test filesystem is case-sensitive: A.txt and a.txt coexist, no platform collision to refuse")
	}
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	small := blobOID(t, repo, "x")
	commit := commitRawTree(t, repo, []string{
		"100644 blob " + small + "\tA.txt",
		"100644 blob " + small + "\ta.txt",
		"100644 blob " + largeBlobOID(t, repo) + "\tbig.bin",
	})
	parent := t.TempDir()
	dest := filepath.Join(parent, "snap")
	err, took := extractBounded(t, repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "duplicate platform path") {
		t.Fatalf("err = %v, want platform-path collision refusal", err)
	}
	t.Logf("collision ahead of an 8 MiB blob refused in %s", took)
	assertNothingWritten(t, parent, dest)
}

func TestExtractRefusesEscapeBeforeLargeBlob(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	commit := commitLiteralTree(t, repo, []literalEntry{
		{name: "..", oid: blobOID(t, repo, "x")},
		{name: "big.bin", oid: largeBlobOID(t, repo)},
	})
	parent := t.TempDir()
	dest := filepath.Join(parent, "snap")
	err, took := extractBounded(t, repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "unsafe path") {
		t.Fatalf("err = %v, want escape refusal", err)
	}
	t.Logf("escape ahead of an 8 MiB blob refused in %s", took)
	assertNothingWritten(t, parent, dest)
}

// TestExtractRefusesMissingObjectsFromListing: ls-tree -l reports "BAD" for
// an object the repository cannot read, so an unreadable blob is refused
// from the listing and never reaches cat-file.
func TestExtractRefusesMissingObjectsFromListing(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	commit := commitRawTreeMissing(t, repo, []string{
		"100644 blob " + strings.Repeat("b", 40) + "\tgone.txt",
		"100644 blob " + largeBlobOID(t, repo) + "\tbig.bin",
	})
	parent := t.TempDir()
	dest := filepath.Join(parent, "snap")
	err, took := extractBounded(t, repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "missing or unreadable object") {
		t.Fatalf("err = %v, want missing-object refusal", err)
	}
	t.Logf("missing object refused from the listing in %s", took)
	assertNothingWritten(t, parent, dest)
}

// truncateLooseObject corrupts a loose object in place so ls-tree -l still
// reports its size but cat-file dies while streaming its body.
func truncateLooseObject(t *testing.T, repo, oid string) {
	t.Helper()
	path := filepath.Join(repo, ".git", "objects", oid[:2], oid[2:])
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload[:len(payload)/2], 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestExtractRemovesWrittenFilesOnMidStreamError: a good entry precedes a
// blob whose loose object is truncated, so cat-file dies mid-body. The file
// already written must be removed and the caller's pre-existing destination
// left otherwise intact.
func TestExtractRemovesWrittenFilesOnMidStreamError(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	big := largeBlobOID(t, repo)
	// Tree entries are sorted by name: a-first.txt streams (and is written)
	// before zz-big.bin fails.
	commit := commitRawTree(t, repo, []string{
		"100644 blob " + blobOID(t, repo, "first") + "\ta-first.txt",
		"100644 blob " + big + "\tzz-big.bin",
	})
	truncateLooseObject(t, repo, big)
	dest := t.TempDir()
	keep := filepath.Join(dest, "keep.txt")
	if err := os.WriteFile(keep, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	err, took := extractBounded(t, repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "reading blob "+big) {
		t.Fatalf("err = %v, want a streaming failure naming the corrupt blob", err)
	}
	t.Logf("mid-stream failure after one written file returned in %s", took)
	for _, name := range []string{"a-first.txt", "zz-big.bin"} {
		if _, err := os.Lstat(filepath.Join(dest, name)); !os.IsNotExist(err) {
			t.Fatalf("%s survived a failed extraction: %v", name, err)
		}
	}
	if got, _ := os.ReadFile(keep); string(got) != "keep" {
		t.Fatalf("pre-existing file disturbed: %q", got)
	}
}

// TestPlanWritesRefusesCaseCollisionBeforeStreaming pins the pre-pass
// itself: on a case-folding destination the A.txt/a.txt pair is refused by
// planWrites, before Extract starts cat-file, and the probe file it uses is
// removed again. Extract's end-to-end cleanup would hide a pre-pass that
// silently fell back to the mid-stream check.
func TestPlanWritesRefusesCaseCollisionBeforeStreaming(t *testing.T) {
	dest := t.TempDir()
	folds, err := destinationFoldsCase(dest)
	if err != nil {
		t.Fatal(err)
	}
	if !folds {
		t.Skip("test filesystem is case-sensitive: A.txt and a.txt coexist, no platform collision to refuse")
	}
	entries := []treeEntry{
		{mode: "100644", kind: "blob", oid: strings.Repeat("a", 40), size: 1, path: "A.txt"},
		{mode: "100644", kind: "blob", oid: strings.Repeat("a", 40), size: 1, path: "a.txt"},
	}
	plan, err := planWrites(dest, entries)
	if err == nil || !strings.Contains(err.Error(), "duplicate platform path") {
		t.Fatalf("plan = %v, err = %v; want collision refusal from the pre-pass", plan, err)
	}
	if left, _ := os.ReadDir(dest); len(left) != 0 {
		t.Fatalf("pre-pass left entries behind: %v", left)
	}
	plan, err = planWrites(dest, entries[:1])
	if err != nil || len(plan) != 1 || plan[0].target != filepath.Join(dest, "A.txt") {
		t.Fatalf("single entry plan = %v, %v", plan, err)
	}
}

// TestExtractTerminatesCatFileOnMidStreamFramingError puts a git shim on
// PATH whose cat-file answers a header for the wrong object and then 8 MiB
// of junk, so the child is alive and blocked on the pipe when the framing
// check fails. The child must be killed and drained before Wait.
func TestExtractTerminatesCatFileOnMidStreamFramingError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake git shim executes a POSIX launcher; the kill-and-drain path is asserted on the unix runners")
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	commit := commitRawTree(t, repo, []string{"100644 blob " + blobOID(t, repo, "x") + "\tsmall.txt"})
	shimDir := t.TempDir()
	shim := "#!/bin/sh\nif [ \"$3\" = cat-file ]; then\n  printf '%s blob 1\\n' " + strings.Repeat("0", 40) +
		"\n  head -c 8388608 /dev/zero\n  exit 0\nfi\nexec '" + realGit + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(shimDir, "git"), []byte(shim), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	parent := t.TempDir()
	dest := filepath.Join(parent, "snap")
	err, took := extractBounded(t, repo, commit, dest)
	if err == nil || !strings.Contains(err.Error(), "unexpected git cat-file --batch response") {
		t.Fatalf("err = %v, want framing refusal", err)
	}
	t.Logf("framing error with 8 MiB queued terminated the child in %s", took)
	assertNothingWritten(t, parent, dest)
}

// TestExtractTerminatesCatFileOnWriteFailure reproduces the deadlock shape
// with the real git: the first entry cannot be opened for writing (its
// parent directory is read-only) while git already has an 8 MiB blob queued
// behind it. The old code waited on a child blocked on the pipe.
func TestExtractTerminatesCatFileOnWriteFailure(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	gitRun(t, repo, "update-index", "--add", "--cacheinfo", "100644,"+blobOID(t, repo, "x")+",ro/blocked.txt")
	gitRun(t, repo, "update-index", "--add", "--cacheinfo", "100644,"+largeBlobOID(t, repo)+",zz-big.bin")
	commit := gitRun(t, repo, "commit-tree", gitRun(t, repo, "write-tree"), "-m", "ro")
	dest := t.TempDir()
	ro := filepath.Join(dest, "ro")
	if err := os.Mkdir(ro, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(ro, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ro, 0o755) })
	if probe, err := os.Create(filepath.Join(ro, "probe")); err == nil {
		_ = probe.Close()
		t.Skip("this process can write through a read-only directory")
	}
	err, took := extractBounded(t, repo, commit, dest)
	if err == nil || !os.IsPermission(err) {
		t.Fatalf("err = %v, want a permission failure on ro/blocked.txt", err)
	}
	t.Logf("write failure with 8 MiB queued terminated cat-file in %s", took)
	if _, err := os.Lstat(filepath.Join(dest, "zz-big.bin")); !os.IsNotExist(err) {
		t.Fatalf("zz-big.bin written after the failure: %v", err)
	}
}

func TestExtractRefusesDotGitComponents(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, "[core]\n")
	parent := t.TempDir()
	dest := filepath.Join(parent, "snap")
	for _, name := range []string{".git", ".git/config", "sub/.git/HEAD", ".GIT/config", "sub/.Git"} {
		commit := commitLiteralTree(t, repo, []literalEntry{{name: name, oid: oid}})
		err := Extract(repo, commit, dest)
		if err == nil || !strings.Contains(err.Error(), "unsafe path") {
			t.Fatalf("%q: err = %v, want .git refusal", name, err)
		}
		assertNothingWritten(t, parent, dest)
	}
	// A name that merely contains the token stays admitted.
	commit := commitLiteralTree(t, repo, []literalEntry{{name: ".gitattributes", oid: oid}, {name: "git/.gitkeep", oid: oid}})
	if err := Extract(repo, commit, dest); err != nil {
		t.Fatalf(".gitattributes and git/.gitkeep must be admitted: %v", err)
	}
	if err := EnsureRepo(dest); err == nil {
		t.Fatal("a snapshot must never look like a git repository")
	}
}

func TestExtractPreservesNestedExecutableBit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose portable executable permission bits")
	}
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	oid := blobOID(t, repo, "#!/bin/sh\n")
	for path, mode := range map[string]string{"sub/run.sh": "100755", "sub/deep/plain.sh": "100644", "sub/deep/exec.sh": "100755"} {
		gitRun(t, repo, "update-index", "--add", "--cacheinfo", mode+","+oid+","+path)
	}
	commit := gitRun(t, repo, "commit-tree", gitRun(t, repo, "write-tree"), "-m", "nested")
	dest := filepath.Join(t.TempDir(), "snap")
	if err := Extract(repo, commit, dest); err != nil {
		t.Fatal(err)
	}
	for path, exec := range map[string]bool{"sub/run.sh": true, "sub/deep/plain.sh": false, "sub/deep/exec.sh": true} {
		info, err := os.Stat(filepath.Join(dest, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm()&0o111 != 0; got != exec {
			t.Fatalf("%s: executable = %v, want %v (mode %v)", path, got, exec, info.Mode())
		}
		payload, _ := os.ReadFile(filepath.Join(dest, filepath.FromSlash(path)))
		if !bytes.Equal(payload, []byte("#!/bin/sh\n")) {
			t.Fatalf("%s: bytes %q", path, payload)
		}
	}
}
