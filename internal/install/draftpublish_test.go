package install

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/gitops"
	"github.com/relux-works/curator/internal/manifest"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/transaction"
)

// This suite proves the draft repair, refresh, and boundary contract at
// the production entry (install.Project): status-relevant currentness
// through the moved-tag gate, repair that revalidates the exact locked
// source and evidence without touching the lock, refresh that replaces
// lock and marker only after success, write-time boundary rechecks that
// refuse a retargeted destination or a replaced admitted input with full
// rollback, and unmanaged adapter destinations that refuse takeover.

// TestDraftDeclaredTagBumpIsNotAMovedTag is the negative row for the
// install-time moved-tag gate on the draft lane: a declared bump from v1
// to v2 (v1 untouched) installs cleanly under StrictTags with no
// moved-tag warning, because a schema-5 marker carries no ref and a
// commit difference is not evidence that a tag moved. The installed
// marker binds the newly declared commit.
func TestDraftDeclaredTagBumpIsNotAMovedTag(t *testing.T) {
	project, home, repo, _ := setupGitInstall(t)
	if result := draftRealInstall(t, project, home, Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	first, err := gitops.Resolve(repo, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	skillMD := filepath.Join(repo, "skills", "review", "SKILL.md")
	skillPayload, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillMD, append(skillPayload, []byte("\n<!-- v2 -->\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "second")
	testGit(t, repo, "tag", "v2") // new tag; v1 stays where it was
	still, err := gitops.Resolve(repo, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	if still.Commit != first.Commit {
		t.Fatalf("v1 moved: %s -> %s", first.Commit, still.Commit)
	}
	second, err := gitops.Resolve(repo, "tag", "v2")
	if err != nil {
		t.Fatal(err)
	}
	payload := `{"schema_version":2,"sources":{"s":{"git":"https://example.org/kit.git","tag":"v2"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	refreshDraftLocked(t, project, home, payload, map[string]string{"s": repo})

	strict := draftRealInstall(t, project, home, Options{StrictTags: true})
	if strict.Status != "ok" {
		t.Fatalf("declared tag bump v1->v2 refused under StrictTags: %+v (v1 still %s, v2 %s)", strict.Errors, first.Commit, second.Commit)
	}
	if joined := strings.Join(strict.Messages, "\n"); strings.Contains(joined, "moved tag") {
		t.Fatalf("declared tag bump reported as a moved tag: %s", joined)
	}
	advisory := draftRealInstall(t, project, home, Options{})
	if advisory.Status != "ok" {
		t.Fatalf("install = %+v", advisory)
	}
	if joined := strings.Join(advisory.Messages, "\n"); strings.Contains(joined, "moved tag") {
		t.Fatalf("declared tag bump reported as a moved tag: %s", joined)
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "review"))
	if recorded == nil || recorded.Package == nil || recorded.Package.Commit == nil || recorded.Package.Commit.Hex != second.Commit {
		t.Fatalf("marker binds %+v, want %s", recorded, second.Commit)
	}
}

// TestDraftSameTagMoveInstallsWithoutWarning is the second negative row:
// on the draft lane install never re-resolves refs, so even a genuine
// same-tag move surfaces only as a new lock at explicit refresh — the
// install itself warns about nothing and binds the refreshed commit.
func TestDraftSameTagMoveInstallsWithoutWarning(t *testing.T) {
	project, home, repo, _ := setupGitInstall(t)
	if result := draftRealInstall(t, project, home, Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	first, err := gitops.Resolve(repo, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	skillMD := filepath.Join(repo, "skills", "review", "SKILL.md")
	skillPayload, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillMD, append(skillPayload, []byte("\n<!-- moved -->\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "second")
	testGit(t, repo, "tag", "-f", "v1")
	second, err := gitops.Resolve(repo, "tag", "v1")
	if err != nil {
		t.Fatal(err)
	}
	if second.Commit == first.Commit {
		t.Fatalf("tag still resolves %s", first.Commit)
	}
	payload, err := os.ReadFile(filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	// The declaration is unchanged, so the refresh is what observes the
	// move: the lock must now bind the new commit.
	plan := refreshDraftLocked(t, project, home, string(payload), map[string]string{"s": repo})
	member, ok := plan.Lock.Find("review")
	if !ok || member.Package.Commit.Hex != second.Commit {
		t.Fatalf("refreshed lock binds %+v, want the moved commit %s", member, second.Commit)
	}

	strict := draftRealInstall(t, project, home, Options{StrictTags: true})
	if strict.Status != "ok" {
		t.Fatalf("install after a refreshed tag move refused under StrictTags: %+v", strict)
	}
	if joined := strings.Join(strict.Messages, "\n"); strings.Contains(joined, "moved tag") {
		t.Fatalf("install after a refreshed tag move warns: %s", joined)
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "review"))
	if recorded == nil || recorded.Package == nil || recorded.Package.Commit == nil ||
		recorded.Package.Commit.Hex != second.Commit {
		t.Fatalf("installed marker binds %+v, want the moved commit %s", recorded, second.Commit)
	}
}

// TestLegacyDeclaredTagBumpIsNotAMovedTag is the frozen-v1 control for
// the draft negative rows above: the same declared bump under a v4
// marker passes StrictTags on the legacy lane, which the v5 bound leaves
// untouched.
func TestLegacyDeclaredTagBumpIsNotAMovedTag(t *testing.T) {
	e := newEnv(t)
	e.skill("skill-a")
	e.declare("skill-a")
	if result := e.install(Options{}); result.Status != "ok" {
		t.Fatalf("first: %+v", result)
	}
	dir := filepath.Join(e.skillsRoot, "skill-a")
	e.write(dir, "SKILL.md", "---\nname: skill-a\ndescription: d2\n---\n# v2\n")
	e.git(dir, "commit", "-qam", "two")
	e.git(dir, "tag", "v2")
	payload, _ := json.MarshalIndent(map[string]any{
		"schema_version": 1,
		"agents":         []string{"claude_code"},
		"skills":         []map[string]any{{"name": "skill-a", "tag": "v2"}},
	}, "", "  ")
	e.write(e.project, "Skillfile.json", string(payload))
	strict := e.install(Options{StrictTags: true})
	if strict.Status != "ok" {
		t.Fatalf("legacy declared bump refused under StrictTags: %+v", strict)
	}
	if strings.Contains(strings.Join(strict.Messages, "\n"), "moved tag") {
		t.Fatalf("legacy declared bump reported as moved: %v", strict.Messages)
	}
}

// TestDraftDryRunExposesResolvedAttestations proves install.Project
// surfaces the registry evidence the effective plan selected: a dry run
// carries the injected ResolveAttest map on Result.Attestations, and a
// failed resolution leaves it nil. Status compares this map against the
// recorded markers (skillfile-sources §4: changed registry, status, or
// key is non-current); without this row the wiring has no
// production-entry coverage.
func TestDraftDryRunExposesResolvedAttestations(t *testing.T) {
	project, home := setupDraftSweep(t)
	injected := map[string]*marker.Attestation{
		"review": {Registry: "example.org/kit", Status: "audited", KeyID: "0123456789abcdef"},
		"scan":   {Registry: "example.org/kit", Status: "audited", KeyID: "0123456789abcdef"},
	}
	result := draftSweepInstall(t, home, project, Options{DryRun: true,
		ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
			return injected, nil, nil
		},
	})
	if result.Status != "ok" {
		t.Fatalf("dry run = %+v", result)
	}
	if !reflect.DeepEqual(result.Attestations, injected) {
		t.Fatalf("Attestations = %+v, want the injected map %+v", result.Attestations, injected)
	}

	failed := draftSweepInstall(t, home, project, Options{DryRun: true,
		ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
			return nil, nil, errors.New("registry unreachable")
		},
	})
	if failed.Status != "failed" {
		t.Fatalf("dry run = %+v, want failed", failed)
	}
	if failed.Attestations != nil {
		t.Fatalf("Attestations = %+v, want nil when resolution fails", failed.Attestations)
	}
}

// TestDraftRepairRevalidatesLockedSource proves a repair revalidates the
// exact locked bytes: a tampered frozen snapshot fails the reinstall and
// the prior lock and marker survive byte-identically.
func TestDraftRepairRevalidatesLockedSource(t *testing.T) {
	project, home := setupDraftSweep(t)
	if result := draftSweepInstall(t, home, project, Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	lockPath := sourcelock.PathIn(project)
	priorLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(project, ".agents", "skills", "review")
	priorMarker, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	store, err := snapshot.OpenLocal(home, member.Package.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	staged, err := os.ReadFile(filepath.Join(store, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store, "SKILL.md"), append(staged, []byte("tampered")...), 0o644); err != nil {
		t.Fatal(err)
	}
	again := draftSweepInstall(t, home, project, Options{})
	if again.Status != "failed" || !strings.Contains(strings.Join(again.Errors, ";"), "source_snapshot_changed") {
		t.Fatalf("reinstall = %+v, want source_snapshot_changed", again)
	}
	afterLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterLock) != string(priorLock) {
		t.Fatalf("refused repair rewrote the lock")
	}
	afterMarker, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	if string(afterMarker) != string(priorMarker) {
		t.Fatalf("refused repair rewrote the marker")
	}
}

// TestDraftRepairRestoresDriftedContent proves a repair reinstalls drifted
// installed bytes from the frozen lock without touching the lock itself:
// the package identity is unchanged, only the live bytes are restored.
func TestDraftRepairRestoresDriftedContent(t *testing.T) {
	project, home := setupDraftSweep(t)
	if result := draftSweepInstall(t, home, project, Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	lockPath := sourcelock.PathIn(project)
	priorLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(project, ".agents", "skills", "review")
	contextFile := filepath.Join(installed, "SKILL.md")
	staged, err := os.ReadFile(contextFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(contextFile, append(staged, []byte("drift")...), 0o644); err != nil {
		t.Fatal(err)
	}
	again := draftSweepInstall(t, home, project, Options{})
	if again.Status != "ok" {
		t.Fatalf("repair = %+v", again)
	}
	restored, err := os.ReadFile(contextFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != string(staged) {
		t.Fatalf("repair left drifted bytes behind")
	}
	afterLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterLock) != string(priorLock) {
		t.Fatalf("repair rewrote the lock")
	}
	lock, err := sourcelock.Read(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	member, ok := lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	recorded := marker.Read(installed)
	if recorded == nil || recorded.LockSHA256 != lock.LockSHA256 {
		t.Fatalf("repaired marker binds %v, want lock %s", recorded, lock.LockSHA256)
	}
	recordedDigest, err := recorded.Package.Digest()
	if err != nil {
		t.Fatal(err)
	}
	memberDigest, err := member.Package.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if recordedDigest != memberDigest {
		t.Fatalf("repaired package %s != locked %s", recordedDigest, memberDigest)
	}
}

// TestDraftRefusedRepairPreservesLock proves a repair refused by required
// evidence changes neither the previous lock nor the installed marker:
// the recorded attestation summary never authorizes, and the refusal
// preserves prior state exactly.
func TestDraftRefusedRepairPreservesLock(t *testing.T) {
	project, home, _, _ := setupGitInstall(t)
	attested := &marker.Attestation{Registry: "example.org/kit", Status: "audited"}
	first := draftRealInstall(t, project, home, Options{ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
		return map[string]*marker.Attestation{"review": attested}, nil, nil
	}})
	if first.Status != "ok" {
		t.Fatalf("install = %+v", first)
	}
	lockPath := sourcelock.PathIn(project)
	priorLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(project, ".agents", "skills", "review")
	priorMarker, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	revoked := errors.New("review is revoked by example.org/kit")
	again := draftRealInstall(t, project, home, Options{ResolveAttest: func([]*closure.Node) (map[string]*marker.Attestation, []string, error) {
		return nil, nil, revoked
	}})
	if again.Status != "failed" || !strings.Contains(strings.Join(again.Errors, ";"), "revoked") {
		t.Fatalf("reinstall = %+v, want the revocation refusal", again)
	}
	afterLock, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterLock) != string(priorLock) {
		t.Fatalf("refused repair rewrote the lock")
	}
	afterMarker, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	if string(afterMarker) != string(priorMarker) {
		t.Fatalf("refused repair rewrote the attested marker")
	}
}

// TestDraftInstallFailsStaleBindingsGeneration proves frozen-input
// enforcement across the crash window: a new lock beside old bindings
// (a refresh interrupted between its two publication writes) fails
// closed with source_lock_stale instead of reinterpreting the lock, the
// install itself writes no lock bytes, and the next explicit refresh
// recovers the generation.
func TestDraftInstallFailsStaleBindingsGeneration(t *testing.T) {
	project, home, repo, _ := setupGitInstall(t)
	if result := draftRealInstall(t, project, home, Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	first, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	skillMD := filepath.Join(repo, "skills", "review", "SKILL.md")
	skillPayload, err := os.ReadFile(skillMD)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skillMD, append(skillPayload, []byte("\n<!-- v2 -->\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	testGit(t, repo, "add", ".")
	testGit(t, repo, "commit", "-qm", "second")
	testGit(t, repo, "tag", "-f", "v1")
	payload, err := os.ReadFile(filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := manifest.ParseBytes(payload, filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot: project, Home: home, Manifest: m, ManifestPayload: payload,
		Expansion: manifest.ExpansionOptions{GitRoots: map[string]string{"s": repo}},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Simulate the crash: publish the new lock without its bindings.
	if err := sourcelock.Write(sourcelock.PathIn(project), second.Lock); err != nil {
		t.Fatal(err)
	}
	crashedLock, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	denied := draftRealInstall(t, project, home, Options{})
	if denied.Status != "failed" || !strings.Contains(strings.Join(denied.Errors, ";"), "source_lock_stale") {
		t.Fatalf("install = %+v, want source_lock_stale", denied)
	}
	afterLock, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(afterLock) != string(crashedLock) {
		t.Fatalf("denied install rewrote the lock")
	}
	recorded := marker.Read(filepath.Join(project, ".agents", "skills", "review"))
	if recorded == nil || recorded.LockSHA256 != first.LockSHA256 {
		t.Fatalf("denied install moved the marker off %s", first.LockSHA256)
	}
	// The next explicit attempt re-publishes both files and the install
	// completes onto the recovered generation.
	recovered := refreshDraftLocked(t, project, home, string(payload), map[string]string{"s": repo})
	done := draftRealInstall(t, project, home, Options{})
	if done.Status != "ok" {
		t.Fatalf("install after recovery = %+v", done)
	}
	recorded = marker.Read(filepath.Join(project, ".agents", "skills", "review"))
	if recorded == nil || recorded.LockSHA256 != recovered.Lock.LockSHA256 {
		t.Fatalf("recovered marker binds %v, want lock %s", recorded, recovered.Lock.LockSHA256)
	}
}

// TestDraftRefreshThenInstallReplacesLockAndMarker is the refresh
// integration row: the explicit attempt publishes the new lock, a failed
// install preserves the previous marker beside it, and the successful
// install binds the new generation. Neither file is replaced before its
// own success.
func TestDraftRefreshThenInstallReplacesLockAndMarker(t *testing.T) {
	project, home := setupDraftSweep(t)
	if result := draftSweepInstall(t, home, project, Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	installed := filepath.Join(project, ".agents", "skills", "review")
	priorMarker, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	priorLock, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan := refreshDraftLocked(t, project, home, string(payload), nil)
	refreshedLock, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(refreshedLock) == string(priorLock) {
		t.Fatalf("refresh published the same lock bytes")
	}
	probe := &draftCommitProbe{failClass: "10-context", failErr: errors.New("injected failure in class 10-context")}
	failed := draftSweepInstall(t, home, project, Options{Commit: CommitDeps{Hooks: probe.hooks(), MaxRestarts: 1}})
	if failed.Status != "failed" {
		t.Fatalf("install status = %q, want failed", failed.Status)
	}
	keptMarker, err := os.ReadFile(filepath.Join(installed, marker.Name))
	if err != nil {
		t.Fatal(err)
	}
	if string(keptMarker) != string(priorMarker) {
		t.Fatalf("failed install replaced the previous marker")
	}
	keptLock, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(keptLock) != string(refreshedLock) {
		t.Fatalf("failed install moved the published lock")
	}
	assertNoDraftJournalRemains(t, home)
	done := draftSweepInstall(t, home, project, Options{})
	if done.Status != "ok" {
		t.Fatalf("install = %+v", done)
	}
	recorded := marker.Read(installed)
	if recorded == nil || recorded.LockSHA256 != plan.Lock.LockSHA256 {
		t.Fatalf("installed marker binds %v, want lock %s", recorded, plan.Lock.LockSHA256)
	}
}

// onceHook runs mutate exactly once at the given transaction point and
// lets the commit continue, so a test can change live state after the
// pre-journal recheck and prove the per-write guard catches it.
type onceHook struct {
	mu     sync.Mutex
	point  transaction.Point
	fired  bool
	mutate func()
}

func (hook *onceHook) hooks() transaction.Hooks {
	return transaction.Hooks{
		Fault: func(event transaction.Event) error {
			hook.mu.Lock()
			defer hook.mu.Unlock()
			if hook.fired || event.Point != hook.point {
				return nil
			}
			hook.fired = true
			hook.mutate()
			return nil
		},
	}
}

func (hook *onceHook) assertFired(t *testing.T) {
	t.Helper()
	hook.mu.Lock()
	defer hook.mu.Unlock()
	if !hook.fired {
		t.Fatal("the mutation hook never fired")
	}
}

// copyTreeDuplicates duplicates the link-free tree at src onto dst. Draft
// context trees hold regular files only; a link fails loudly instead of
// being copied blind.
func copyTreeDuplicates(t *testing.T, src, dst string) {
	t.Helper()
	info, err := os.Lstat(src)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("cannot duplicate a link: %s", src)
	}
	if !info.IsDir() {
		payload, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, payload, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		copyTreeDuplicates(t, filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()))
	}
}

// TestDraftRetargetFailsAtWriteTimeAndRollsBack replaces a destination
// parent with a byte-identical copy after the pre-journal recheck — the
// same spelling, a new identity — and proves the per-write guard refuses
// the publication and rolls every target back.
func TestDraftRetargetFailsAtWriteTimeAndRollsBack(t *testing.T) {
	project, home := setupDraftSweep(t)
	if result := draftSweepInstall(t, home, project, Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	refreshDraftLocked(t, project, home, string(payload), nil)
	before := snapshotDraftState(t, home, project)
	skillsDir := filepath.Join(project, ".agents", "skills")
	stagedAway := skillsDir + ".orig"
	hook := &onceHook{point: transaction.PointPrepared, mutate: func() {
		if err := os.Rename(skillsDir, stagedAway); err != nil {
			panic(err)
		}
		copyTreeDuplicates(t, stagedAway, skillsDir)
	}}
	result := draftSweepInstall(t, home, project, Options{Commit: CommitDeps{Hooks: hook.hooks(), MaxRestarts: 1}})
	hook.assertFired(t)
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_output_overlap") {
		t.Fatalf("install = %+v, want the source_output_overlap refusal", result)
	}
	// Undo the test's own retarget, then prove the failed run itself
	// changed nothing: every target rolled back to its preimage.
	if err := os.RemoveAll(skillsDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(stagedAway, skillsDir); err != nil {
		t.Fatal(err)
	}
	// The rename above moved the journal's own sidecars
	// (.curator-txn-*) out of the engine's view and back around the
	// rollback that already cleaned its own view of them. They are
	// operation-private staging, never installed state — the stale
	// scan skips dotfiles — so the test asserts they are the ONLY
	// difference, sweeps them, and proves the live state is exact.
	sweepRetargetSidecars(t, skillsDir)
	if changed := before.diff(snapshotDraftState(t, home, project)); len(changed) > 0 {
		t.Fatalf("rollback did not restore the prior state: %v", changed)
	}
	assertNoDraftJournalRemains(t, home)
}

// sweepRetargetSidecars removes the journal-owned staging sidecars a
// parent-swap test moved aside and back. Only that exact prefix is swept;
// the whole-state diff after it proves everything else is byte-identical.
func sweepRetargetSidecars(t *testing.T, skillsDir string) {
	t.Helper()
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".curator-txn-") {
			continue
		}
		if err := os.RemoveAll(filepath.Join(skillsDir, name)); err != nil {
			t.Fatal(err)
		}
	}
}

// TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack replaces a
// frozen admitted input after staging copied it and proves the per-write
// guard refuses the publication: the staged copy may no longer describe
// the input it was taken from, so nothing is published.
func TestDraftAdmittedReplacementFailsAtWriteTimeAndRollsBack(t *testing.T) {
	project, home := setupDraftSweep(t)
	if result := draftSweepInstall(t, home, project, Options{}); result.Status != "ok" {
		t.Fatalf("install = %+v", result)
	}
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(project, "Skillfile.json"))
	if err != nil {
		t.Fatal(err)
	}
	plan := refreshDraftLocked(t, project, home, string(payload), nil)
	member, ok := plan.Lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	store, err := snapshot.OpenLocal(home, member.Package.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	before := snapshotDraftState(t, home, project)
	stagedAway := store + ".orig"
	hook := &onceHook{point: transaction.PointPrepared, mutate: func() {
		if err := os.Rename(store, stagedAway); err != nil {
			panic(err)
		}
		if err := os.Mkdir(store, 0o755); err != nil {
			panic(err)
		}
	}}
	result := draftSweepInstall(t, home, project, Options{Commit: CommitDeps{Hooks: hook.hooks(), MaxRestarts: 1}})
	hook.assertFired(t)
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "source_output_overlap") {
		t.Fatalf("install = %+v, want the source_output_overlap refusal", result)
	}
	if err := os.Remove(store); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(stagedAway, store); err != nil {
		t.Fatal(err)
	}
	if changed := before.diff(snapshotDraftState(t, home, project)); len(changed) > 0 {
		t.Fatalf("rollback did not restore the prior state: %v", changed)
	}
	assertNoDraftJournalRemains(t, home)
}

// TestDraftUnmanagedAdapterDestinationRefuses proves the adapter lane
// never takes over an unmanaged destination: a foreign file at a mirror
// path fails the install closed before any write, the foreign bytes
// survive, and the lock is untouched.
func TestDraftUnmanagedAdapterDestinationRefuses(t *testing.T) {
	project, home := setupDraftSweep(t)
	foreign := filepath.Join(project, ".claude", "skills", "review")
	if err := os.MkdirAll(filepath.Dir(foreign), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(foreign, []byte("not ours"), 0o644); err != nil {
		t.Fatal(err)
	}
	priorLock, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	result := draftSweepInstall(t, home, project, Options{})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "adapter target already exists and is not managed") {
		t.Fatalf("install = %+v, want the unmanaged refusal", result)
	}
	kept, err := os.ReadFile(foreign)
	if err != nil {
		t.Fatal(err)
	}
	if string(kept) != "not ours" {
		t.Fatalf("refused install modified the unmanaged file")
	}
	afterLock, err := os.ReadFile(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if string(afterLock) != string(priorLock) {
		t.Fatalf("refused install rewrote the lock")
	}
	if _, err := os.Lstat(filepath.Join(project, ".agents", "skills", "review")); !os.IsNotExist(err) {
		t.Fatalf("refused install staged a live installation: %v", err)
	}
	assertNoDraftJournalRemains(t, home)
}
