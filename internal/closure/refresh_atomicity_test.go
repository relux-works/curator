package closure

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/sourcelock"
)

// advanceDraftGitTag commits one more file in repo and force-moves tag to
// the new commit, so the next explicit attempt resolves a new generation.
func advanceDraftGitTag(t *testing.T, repo, tag string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		full := append([]string{"-c", "commit.gpgsign=false", "-c", "tag.gpgSign=false"}, args...)
		cmd := exec.Command("git", full...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(repo, "skills", "review", "references", "v2.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-qm", "second")
	run("tag", "-f", tag)
}

// TestRefreshDraftFirstWriteFailurePreservesPriorFiles injects a failure at
// the first publication step (the lock rename cannot stage its temp file)
// and proves the prior lock and bindings survive byte-identically with no
// temp files left behind.
func TestRefreshDraftFirstWriteFailurePreservesPriorFiles(t *testing.T) {
	project := t.TempDir()
	home := t.TempDir()
	writeDraftSkill(t, filepath.Join(project, "skills", "review"), "review", false, nil, nil)
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	m, raw := parseDraftManifest(t, project, payload)
	cfg := DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	bindingsPath := filepath.Join(home, "bindings.json")
	plan, err := RefreshDraft(cfg, lockPath, bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	priorLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	priorBindings, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	// Change the live bytes so a successful refresh would publish a new
	// generation: without a mutation this test cannot tell rollback apart
	// from a no-op write.
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Make the lock rename fail at temp-file staging. A host that ignores
	// directory permissions cannot run this vector.
	if err := os.Chmod(project, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(project, 0o755) })
	probe, probeErr := os.CreateTemp(project, ".probe-*")
	if probeErr == nil {
		_ = probe.Close()
		_ = os.Remove(probe.Name())
		t.Skip("this process can write through a read-only directory")
	}
	if _, err := RefreshDraft(cfg, lockPath, bindingsPath); err == nil {
		t.Fatalf("refresh with an unwritable lock path succeeded")
	}
	restored, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != string(priorLock) {
		t.Fatalf("failed refresh replaced the lock")
	}
	restoredLock, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if restoredLock.LockSHA256 != plan.Lock.LockSHA256 {
		t.Fatalf("lock generation moved %s -> %s", plan.Lock.LockSHA256, restoredLock.LockSHA256)
	}
	restoredBindings, err := os.ReadFile(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restoredBindings) != string(priorBindings) {
		t.Fatalf("failed refresh replaced the bindings")
	}
	entries, err := os.ReadDir(project)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".lock-") {
			t.Fatalf("failed refresh left a temp file behind: %s", entry.Name())
		}
	}
	// The previously published snapshot stays usable.
	if _, err := OpenDraftFrozen(home, plan.Lock, FrozenOptions{}); err != nil {
		t.Fatalf("prior snapshot unusable after failed refresh: %v", err)
	}
}

// TestRefreshDraftCrashedGenerationFailsClosedAndRetryRecovers simulates a
// crash between the two publication renames (the new lock beside the old
// bindings) and proves the mismatch fails closed with source_lock_stale
// instead of reinterpreting the lock — and that the next explicit attempt
// re-publishes both files and recovers.
func TestRefreshDraftCrashedGenerationFailsClosedAndRetryRecovers(t *testing.T) {
	project := t.TempDir()
	home := t.TempDir()
	repo := draftGitRepo(t, map[string]string{"skills/review": "review"}, "v1")
	payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v1"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	m, raw := parseDraftManifest(t, project, payload)
	resolveCfg := func() DraftResolveConfig {
		return DraftResolveConfig{ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: raw,
			Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}}}
	}
	lockPath := filepath.Join(project, "Skillfile.lock.json")
	bindingsPath := filepath.Join(home, "bindings.json")
	first, err := RefreshDraft(resolveCfg(), lockPath, bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	// Advance the tag and resolve the next generation without publishing
	// the bindings: exactly the crash window between the two renames.
	advanceDraftGitTag(t, repo, "v1")
	second, err := ResolveDraft(resolveCfg())
	if err != nil {
		t.Fatal(err)
	}
	if second.Lock.LockSHA256 == first.Lock.LockSHA256 {
		t.Fatalf("advanced tag resolved the same lock generation")
	}
	if err := sourcelock.Write(lockPath, second.Lock); err != nil {
		t.Fatal(err)
	}
	staleBindings, err := sourcelock.ReadBindings(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := staleBindings.CheckFresh(second.Lock); err == nil || !strings.Contains(err.Error(), "source_lock_stale") {
		t.Fatalf("CheckFresh = %v, want source_lock_stale", err)
	}
	// The next explicit attempt re-publishes both files; the recovered
	// generation is fresh and its snapshot opens.
	recovered, err := RefreshDraft(resolveCfg(), lockPath, bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	freshBindings, err := sourcelock.ReadBindings(bindingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := freshBindings.CheckFresh(recovered.Lock); err != nil {
		t.Fatalf("recovered bindings are not fresh: %v", err)
	}
	stored, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if stored.LockSHA256 != recovered.Lock.LockSHA256 {
		t.Fatalf("stored %s != recovered %s", stored.LockSHA256, recovered.Lock.LockSHA256)
	}
}
