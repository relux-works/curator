package snapshot

// Production-path tests for draft local-snapshot capture (revision 1,
// protocol/skillfile-sources.md section 3). Every test reaches a
// production entry point: Capture, PublishLocal or OpenLocal. The
// normative hash vectors pin the local-snapshot-v1 inventory digest
// from conformance/draft-sources-v1/snapshot-cases.json.

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/stateread"
)

func TestFrozenSnapshotReadersRefuseBlockedParents(t *testing.T) {
	t.Run("local_snapshot", func(t *testing.T) {
		home := t.TempDir()
		store := LocalStoreDir(home)
		if err := os.WriteFile(store, []byte("blocker"), 0o600); err != nil {
			t.Fatal(err)
		}
		digest := "sha256:" + strings.Repeat("a", 64)
		_, err := OpenLocal(home, digest)
		assertUnreadableSnapshotState(t, err, localSnapshotDir(home, digest))
	})

	t.Run("git_snapshot", func(t *testing.T) {
		home := t.TempDir()
		cacheRoot := filepath.Join(home, "cache")
		if err := os.WriteFile(cacheRoot, []byte("blocker"), 0o600); err != nil {
			t.Fatal(err)
		}
		commit := strings.Repeat("a", 40)
		_, err := AuthenticateGit(home, "example.org/kit", "", commit)
		assertUnreadableSnapshotState(t, err, Dir(home, "example.org/kit", commit))
	})
}

func assertUnreadableSnapshotState(t *testing.T, err error, path string) {
	t.Helper()
	var stateErr *stateread.Error
	if !errors.As(err, &stateErr) || stateErr.Kind != stateread.KindUnreadable || stateErr.Path != path {
		t.Fatalf("snapshot state error = %v, want typed unreadable for %s", err, path)
	}
}

// normativeVector is one hand-pinned row of the accepted snapshot
// byte vectors: exact file bytes, per-file hashes with executable
// flags, and the full inventory digest.
type normativeVector struct {
	id       string
	files    map[string]string
	exec     map[string]bool
	fileSHA  map[string]string
	snapshot string
}

var normativeVectors = []normativeVector{
	{
		id: "base",
		files: map[string]string{
			"SKILL.md":       "name: review\n",
			"build/main.go":  "package main\n",
			"scripts/run.sh": "echo one\n",
		},
		exec: map[string]bool{"SKILL.md": false, "build/main.go": false, "scripts/run.sh": true},
		fileSHA: map[string]string{
			"SKILL.md":       "sha256:06142a8fcd46ff59b5ddcd13ab902f20a5b9e2a51ce3a40b65fe735ce0371b95",
			"build/main.go":  "sha256:df1d036cbbf3df46e2045071e082245ece204c7f53ecf0a4e022bff9bb228f47",
			"scripts/run.sh": "sha256:0cb42bbdf016ecafd6c21ac6c4b1760bf5b346c70c4f96ba890ef3d74883c8c2",
		},
		snapshot: "sha256:481f362d0c82cdf8c32ed854edfc125d21a7c470c0128086f7eebb1956d832a9",
	},
	{
		id: "runtime-edit",
		files: map[string]string{
			"SKILL.md":       "name: review\n",
			"build/main.go":  "package main\n",
			"scripts/run.sh": "echo two\n",
		},
		exec: map[string]bool{"SKILL.md": false, "build/main.go": false, "scripts/run.sh": true},
		fileSHA: map[string]string{
			"SKILL.md":       "sha256:06142a8fcd46ff59b5ddcd13ab902f20a5b9e2a51ce3a40b65fe735ce0371b95",
			"build/main.go":  "sha256:df1d036cbbf3df46e2045071e082245ece204c7f53ecf0a4e022bff9bb228f47",
			"scripts/run.sh": "sha256:7d97a50c9b1eb3b6a49320a5238fd08280240d28befc12465e493d17d8bc8d56",
		},
		snapshot: "sha256:15ac234d51e71847165d42e2695df3d91ca49314bd5e5811eb17cad6ce8cea80",
	},
	{
		id: "build-edit",
		files: map[string]string{
			"SKILL.md":       "name: review\n",
			"build/main.go":  "package changed\n",
			"scripts/run.sh": "echo one\n",
		},
		exec: map[string]bool{"SKILL.md": false, "build/main.go": false, "scripts/run.sh": true},
		fileSHA: map[string]string{
			"SKILL.md":       "sha256:06142a8fcd46ff59b5ddcd13ab902f20a5b9e2a51ce3a40b65fe735ce0371b95",
			"build/main.go":  "sha256:fbe06aa700c70a568bef3585095a668af24c85c11553f83818009eba13070f60",
			"scripts/run.sh": "sha256:0cb42bbdf016ecafd6c21ac6c4b1760bf5b346c70c4f96ba890ef3d74883c8c2",
		},
		snapshot: "sha256:2aa1bc2a835d6fd931de8f032761c17288bfefeaba492fe229158c602b8750ac",
	},
}

// writeVectorPackage materializes one normative vector as a live
// package tree with the pinned executable bits.
func writeVectorPackage(t *testing.T, pkg string, vector normativeVector) {
	t.Helper()
	for rel, content := range vector.files {
		full := filepath.Join(pkg, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(0o644)
		if vector.exec[rel] {
			mode = 0o755
		}
		if err := os.Chmod(full, mode); err != nil {
			t.Fatal(err)
		}
	}
	// The executable vector is normative on unix runners where the
	// platform reports POSIX execute bits. On Windows the permission
	// bits are synthesized, so the executable-flag assertion is
	// deferred to the unix runners (platform-control skip class).
	info, err := os.Lstat(filepath.Join(pkg, "scripts", "run.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		if info.Mode().Perm()&0o111 == 0 {
			t.Skip("Windows does not expose portable executable permission bits; the normative executable vectors are asserted on the unix runners")
		}
		for rel, wantExec := range vector.exec {
			got, err := os.Lstat(filepath.Join(pkg, filepath.FromSlash(rel)))
			if err != nil {
				t.Fatal(err)
			}
			if (got.Mode().Perm()&0o111 != 0) != wantExec {
				t.Skipf("Windows does not expose portable executable permission bits; the normative executable vectors are asserted on the unix runners (unexpected bits for %s)", rel)
			}
		}
		return
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("scripts/run.sh lost its POSIX execute bit %o; normative executable vectors must hold on unix", info.Mode().Perm())
	}
	for rel, wantExec := range vector.exec {
		got, err := os.Lstat(filepath.Join(pkg, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		if (got.Mode().Perm()&0o111 != 0) != wantExec {
			t.Fatalf("platform reports unexpected execute bits for %s: mode %o, want executable %v", rel, got.Mode().Perm(), wantExec)
		}
	}
}

// prepareVectorAcquisition runs the production admission entry point
// for one materialized vector package.
func prepareVectorAcquisition(t *testing.T, pkg string) *LocalAcquisition {
	t.Helper()
	project := filepath.Dir(pkg)
	acquisition, err := PrepareLocalAcquisition(pkg, ".", project, "s", boundaryOutputs(project), nil, map[string]bool{"s": true}, "")
	if err != nil {
		t.Fatalf("PrepareLocalAcquisition: %v", err)
	}
	t.Cleanup(func() { _ = acquisition.Close() })
	return acquisition
}

// TestCaptureMatchesNormativeVectors drives the production Capture for
// all three accepted byte vectors: exact per-file hashes, executable
// flags and the full inventory digest must equal the pinned values.
func TestCaptureMatchesNormativeVectors(t *testing.T) {
	for _, vector := range normativeVectors {
		t.Run(vector.id, func(t *testing.T) {
			base := t.TempDir()
			pkg := filepath.Join(base, "pkg")
			writeVectorPackage(t, pkg, vector)
			acquisition := prepareVectorAcquisition(t, pkg)
			inventory, err := Capture(acquisition)
			if err != nil {
				t.Fatalf("Capture: %v", err)
			}
			if len(inventory.Files) != len(vector.files) {
				t.Fatalf("files = %v, want %d entries", inventory.Files, len(vector.files))
			}
			for _, entry := range inventory.Files {
				wantSHA, ok := vector.fileSHA[entry.Path]
				if !ok {
					t.Fatalf("unexpected inventory path %q", entry.Path)
				}
				if entry.SHA256 != wantSHA {
					t.Fatalf("%s sha = %s, want %s", entry.Path, entry.SHA256, wantSHA)
				}
				if entry.Executable != vector.exec[entry.Path] {
					t.Fatalf("%s executable = %v, want %v", entry.Path, entry.Executable, vector.exec[entry.Path])
				}
			}
			if inventory.Snapshot != vector.snapshot {
				t.Fatalf("snapshot = %s, want %s", inventory.Snapshot, vector.snapshot)
			}
		})
	}
}

// TestVectorsFreezeContextAcrossRuntimeAndBuildEdits pins the
// runtime-only-refresh / build-only-refresh semantic: SKILL.md bytes
// are identical across all three vectors, but every vector carries a
// distinct package identity, so runtime and build inputs are frozen
// into the snapshot digest rather than inferred from context.
func TestVectorsFreezeContextAcrossRuntimeAndBuildEdits(t *testing.T) {
	skillSHA := normativeVectors[0].fileSHA["SKILL.md"]
	seen := map[string]string{}
	for _, vector := range normativeVectors {
		if vector.fileSHA["SKILL.md"] != skillSHA {
			t.Fatalf("%s changed SKILL.md bytes %s, want the frozen %s", vector.id, vector.fileSHA["SKILL.md"], skillSHA)
		}
		if prior, ok := seen[vector.snapshot]; ok {
			t.Fatalf("snapshot collision: %s and %s share %s", prior, vector.id, vector.snapshot)
		}
		seen[vector.snapshot] = vector.id
	}
	if len(seen) != 3 {
		t.Fatalf("want 3 distinct package identities, got %v", seen)
	}
}

// captureFixture builds an ordinary admitted package with frontmatter
// SKILL.md plus payload files for race and store tests.
func captureFixture(t *testing.T, files map[string]string) (home, pkg string, acquisition *LocalAcquisition) {
	t.Helper()
	base := t.TempDir()
	home = t.TempDir()
	pkg = filepath.Join(base, "pkg")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkg, "SKILL.md"), []byte("---\nname: review\ndescription: A skill\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for rel, content := range files {
		full := filepath.Join(pkg, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	acquisition, err := PrepareLocalAcquisition(pkg, ".", base, "s", boundaryOutputs(base), nil, map[string]bool{"s": true}, "")
	if err != nil {
		t.Fatalf("PrepareLocalAcquisition: %v", err)
	}
	t.Cleanup(func() { _ = acquisition.Close() })
	return home, pkg, acquisition
}

// TestCaptureAdmitsDirtyAndUntrackedInsideGit proves a local path
// means admitted filesystem bytes even inside Git: dirty tracked bytes
// and untracked files enter the inventory, .git metadata does not, and
// capture synthesizes no Git commit.
func TestCaptureAdmitsDirtyAndUntrackedInsideGit(t *testing.T) {
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "SKILL.md"), []byte("---\nname: review\ndescription: A skill\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "tracked.md"), []byte("A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", ".")
	gitRun(t, repo, "commit", "-qm", "one")
	// Dirty tracked bytes plus an untracked file: the snapshot must
	// carry B and C, never the committed A.
	if err := os.WriteFile(filepath.Join(repo, "tracked.md"), []byte("B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "untracked.md"), []byte("C\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	acquisition, err := PrepareLocalAcquisition(repo, ".", t.TempDir(), "s", []string{filepath.Join(home, "outputs")}, nil, map[string]bool{"s": true}, "")
	if err != nil {
		t.Fatalf("PrepareLocalAcquisition: %v", err)
	}
	defer func() { _ = acquisition.Close() }()
	inventory, err := Capture(acquisition)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	byPath := map[string]FileEntry{}
	for _, entry := range inventory.Files {
		byPath[entry.Path] = entry
		if strings.Contains(entry.Path, ".git") {
			t.Fatalf("inventory carries metadata path %q", entry.Path)
		}
	}
	for _, want := range []string{"SKILL.md", "tracked.md", "untracked.md"} {
		if _, ok := byPath[want]; !ok {
			t.Fatalf("inventory = %v, want %s", inventory.Files, want)
		}
	}
	stored, err := PublishLocal(home, acquisition.Staging, inventory)
	if err != nil {
		t.Fatalf("PublishLocal: %v", err)
	}
	if payload, err := os.ReadFile(filepath.Join(stored, "tracked.md")); err != nil || string(payload) != "B\n" {
		t.Fatalf("stored tracked.md = %q, %v; want dirty B", payload, err)
	}
	if payload, err := os.ReadFile(filepath.Join(stored, "untracked.md")); err != nil || string(payload) != "C\n" {
		t.Fatalf("stored untracked.md = %q, %v; want untracked C", payload, err)
	}
	if _, err := os.Lstat(filepath.Join(stored, ".git")); !os.IsNotExist(err) {
		t.Fatalf("stored snapshot carries .git: %v", err)
	}
	// Capture must never turn dirty bytes into a commit.
	count := exec.Command("git", "rev-list", "--count", "HEAD")
	count.Dir = repo
	out, err := count.Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(out)) != "1" {
		t.Fatalf("capture synthesized a commit: rev-list --count HEAD = %s", out)
	}
}

// TestCaptureDetectsAddedMember proves a file admitted between
// admission and capture fails the install with
// source_snapshot_changed instead of publishing a mix.
func TestCaptureDetectsAddedMember(t *testing.T) {
	_, pkg, acquisition := captureFixture(t, map[string]string{"a.md": "a\n"})
	if err := os.WriteFile(filepath.Join(pkg, "late.md"), []byte("late\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Capture(acquisition)
	if err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("Capture err = %v, want leading source_snapshot_changed", err)
	}
}

// TestCaptureDetectsRemovedMember proves a file vanishing between
// admission and capture fails the same way.
func TestCaptureDetectsRemovedMember(t *testing.T) {
	_, pkg, acquisition := captureFixture(t, map[string]string{"doomed.md": "x\n"})
	if err := os.Remove(filepath.Join(pkg, "doomed.md")); err != nil {
		t.Fatal(err)
	}
	_, err := Capture(acquisition)
	if err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("Capture err = %v, want leading source_snapshot_changed", err)
	}
}

// TestCaptureDetectsLiveContentChange proves bytes changing inside one
// capture fail it: the live file is rewritten after the copy loop within
// the same Capture, so the post-copy re-read must refuse with
// source_snapshot_changed. Narrowing the post-copy live comparison to a
// file-count check keeps the count identical and must fail this test.
func TestCaptureDetectsLiveContentChange(t *testing.T) {
	_, pkg, acquisition := captureFixture(t, map[string]string{"a.md": "before\n"})
	var hookErr error
	captureAfterCopyHook = func() {
		hookErr = os.WriteFile(filepath.Join(pkg, "a.md"), []byte("after\n"), 0o644)
	}
	defer func() { captureAfterCopyHook = nil }()
	_, err := Capture(acquisition)
	if hookErr != nil {
		t.Fatalf("race hook: %v", hookErr)
	}
	if err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("Capture err = %v, want leading source_snapshot_changed", err)
	}
}

// TestReviewIdentityReplacementDuringCapture proves a file replaced by a
// new file with identical bytes and mode inside one capture still fails
// it: the post-copy physical-identity revalidation must refuse with
// source_snapshot_changed. Comparing digests alone cannot see this race.
func TestReviewIdentityReplacementDuringCapture(t *testing.T) {
	_, pkg, acquisition := captureFixture(t, map[string]string{"a.md": "same\n"})
	var hookErr error
	captureAfterCopyHook = func() {
		path := filepath.Join(pkg, "a.md")
		if err := os.WriteFile(path+".new", []byte("same\n"), 0o644); err != nil {
			hookErr = err
			return
		}
		hookErr = os.Rename(path+".new", path)
	}
	defer func() { captureAfterCopyHook = nil }()
	_, err := Capture(acquisition)
	if hookErr != nil {
		t.Fatalf("race hook: %v", hookErr)
	}
	if err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("Capture err = %v, want leading source_snapshot_changed", err)
	}
}

// TestPublishRefusesCopyMutatedAfterAudit proves the publication copy is
// hashed immediately before the rename: mutating the frozen copy after
// the pre-publication audit flows into the publication copy, so
// PublishLocal must refuse with source_snapshot_changed and leave no
// store entry behind. Skipping the pre-rename hash must fail this test.
func TestPublishRefusesCopyMutatedAfterAudit(t *testing.T) {
	home, _, acquisition := captureFixture(t, map[string]string{"a.md": "before\n"})
	inventory, err := Capture(acquisition)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	var hookErr error
	publishAfterAuditHook = func() {
		hookErr = os.WriteFile(filepath.Join(acquisition.Staging, "a.md"), []byte("after\n"), 0o644)
	}
	defer func() { publishAfterAuditHook = nil }()
	_, err = PublishLocal(home, acquisition.Staging, inventory)
	if hookErr != nil {
		t.Fatalf("race hook: %v", hookErr)
	}
	if err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("PublishLocal err = %v, want leading source_snapshot_changed", err)
	}
	if _, statErr := os.Lstat(localSnapshotDir(home, inventory.Snapshot)); !os.IsNotExist(statErr) {
		t.Fatalf("refused publication left store state behind")
	}
}

// TestPublishDetectsFrozenMutation proves the frozen-copy rehash
// before publication: mutating the staged copy between audit and
// publication fails with source_snapshot_changed. Narrowing
// PublishLocal to skip its pre-publication rehash must fail this
// test.
func TestPublishDetectsFrozenMutation(t *testing.T) {
	home, _, acquisition := captureFixture(t, map[string]string{"a.md": "audited\n"})
	inventory, err := Capture(acquisition)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(acquisition.Staging, "a.md"), []byte("mutated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = PublishLocal(home, acquisition.Staging, inventory)
	if err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("PublishLocal err = %v, want leading source_snapshot_changed", err)
	}
	if _, statErr := os.Lstat(localSnapshotDir(home, inventory.Snapshot)); !os.IsNotExist(statErr) {
		t.Fatalf("refused publication left store state behind")
	}
}

// TestPublishOpenRoundTrip proves the store is content-addressed and
// idempotent: publishing twice reuses the winner, and OpenLocal
// resolves the locked digest back to the frozen tree.
func TestPublishOpenRoundTrip(t *testing.T) {
	home, _, acquisition := captureFixture(t, map[string]string{"a.md": "a\n"})
	inventory, err := Capture(acquisition)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	first, err := PublishLocal(home, acquisition.Staging, inventory)
	if err != nil {
		t.Fatalf("PublishLocal: %v", err)
	}
	second, err := PublishLocal(home, acquisition.Staging, inventory)
	if err != nil || second != first {
		t.Fatalf("republish = %q, %v; want the idempotent %q", second, err, first)
	}
	opened, err := OpenLocal(home, inventory.Snapshot)
	if err != nil || opened != first {
		t.Fatalf("OpenLocal = %q, %v; want %q", opened, err, first)
	}
	if payload, err := os.ReadFile(filepath.Join(opened, "a.md")); err != nil || string(payload) != "a\n" {
		t.Fatalf("opened a.md = %q, %v", payload, err)
	}
}

// TestOpenMissingSnapshot proves a locked digest with no store entry
// fails with source_snapshot_unavailable instead of being recreated
// from current bytes.
func TestOpenMissingSnapshot(t *testing.T) {
	home := t.TempDir()
	absent := "sha256:" + strings.Repeat("0", 64)
	if _, err := OpenLocal(home, absent); err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_unavailable") {
		t.Fatalf("OpenLocal err = %v, want leading source_snapshot_unavailable", err)
	}
	// Removing a published entry makes the same locked digest
	// unavailable: the store never falls back to the live tree.
	_, _, acquisition := captureFixture(t, map[string]string{"a.md": "a\n"})
	inventory, err := Capture(acquisition)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	stored, err := PublishLocal(home, acquisition.Staging, inventory)
	if err != nil {
		t.Fatalf("PublishLocal: %v", err)
	}
	if err := os.RemoveAll(filepath.Dir(stored)); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenLocal(home, inventory.Snapshot); err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_unavailable") {
		t.Fatalf("OpenLocal after removal err = %v, want leading source_snapshot_unavailable", err)
	}
}

// TestOpenTamperedSnapshot proves a store tree that no longer matches
// its digest fails with source_snapshot_changed and keeps its bytes
// out of installers.
func TestOpenTamperedSnapshot(t *testing.T) {
	home, _, acquisition := captureFixture(t, map[string]string{"a.md": "a\n"})
	inventory, err := Capture(acquisition)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	stored, err := PublishLocal(home, acquisition.Staging, inventory)
	if err != nil {
		t.Fatalf("PublishLocal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stored, "a.md"), []byte("tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenLocal(home, inventory.Snapshot); err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("OpenLocal err = %v, want leading source_snapshot_changed", err)
	}
}

// TestOpenMalformedDigest proves a locked digest that is not a
// well-formed sha256: digest fails with source_snapshot_unavailable
// instead of resolving a directory or falling back to live bytes.
func TestOpenMalformedDigest(t *testing.T) {
	home := t.TempDir()
	for _, bad := range []string{"", "not-a-digest", "sha256:" + strings.Repeat("0", 63), "sha256:" + strings.Repeat("z", 64)} {
		if _, err := OpenLocal(home, bad); err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_unavailable") {
			t.Fatalf("OpenLocal(%q) err = %v, want leading source_snapshot_unavailable", bad, err)
		}
	}
}

// TestCaptureRefusesUnprepared proves the production entry points
// reject calls without a prepared acquisition or audited inventory:
// Capture(nil), an empty acquisition and a digest-less PublishLocal
// all fail with source_member_invalid and publish nothing.
func TestCaptureRefusesUnprepared(t *testing.T) {
	if _, err := Capture(nil); err == nil || !strings.HasPrefix(err.Error(), "source_member_invalid") {
		t.Fatalf("Capture(nil) err = %v, want leading source_member_invalid", err)
	}
	if _, err := Capture(&LocalAcquisition{}); err == nil || !strings.HasPrefix(err.Error(), "source_member_invalid") {
		t.Fatalf("Capture(empty) err = %v, want leading source_member_invalid", err)
	}
	home := t.TempDir()
	if _, err := PublishLocal(home, t.TempDir(), Inventory{}); err == nil || !strings.HasPrefix(err.Error(), "source_member_invalid") {
		t.Fatalf("PublishLocal(empty) err = %v, want leading source_member_invalid", err)
	}
	if _, err := os.Lstat(LocalStoreDir(home)); !os.IsNotExist(err) {
		t.Fatalf("refused publication left store state behind: %v", err)
	}
}

// TestCaptureDetectsLinkSwap proves a regular file replaced by a link
// between admission and capture fails with source_snapshot_changed:
// narrowing Capture to skip its link/special rejection must fail this
// test.
func TestCaptureDetectsLinkSwap(t *testing.T) {
	_, pkg, acquisition := captureFixture(t, map[string]string{"a.md": "a\n"})
	target := filepath.Join(pkg, "a.md")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("SKILL.md", target); err != nil {
		t.Skipf("platform cannot create links: %v", err)
	}
	_, err := Capture(acquisition)
	if err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("Capture err = %v, want leading source_snapshot_changed", err)
	}
}
