package crossconformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/audit"
	"github.com/relux-works/curator/internal/capabilities"
	"github.com/relux-works/curator/internal/closure"
	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/install"
	"github.com/relux-works/curator/internal/snapshot"
	"github.com/relux-works/curator/internal/sourcelock"
	"github.com/relux-works/curator/internal/staging"
)

// Capture and audit-state semantic rows (§3–§4): mutation detection,
// dirty-git capture, missing snapshots, refresh identity, audit report
// binding, and the local attestation gate.

func init() {
	registerDraftSemantic("capture-mutation", driveCaptureMutation)
	registerDraftSemantic("frozen-copy-mutation", driveFrozenCopyMutation)
	registerDraftSemantic("local-git-dirty", driveLocalGitDirty)
	registerDraftSemantic("runtime-only-refresh", driveRuntimeOnlyRefresh)
	registerDraftSemantic("build-only-refresh", driveBuildOnlyRefresh)
	registerDraftSemantic("missing-audit-report", driveMissingAuditReport)
	registerDraftSemantic("strict-network-attestation-local", driveStrictNetworkAttestationLocal)
}

// driveCaptureMutation exercises the live mutation window through the
// production closure resolver. The acquisition-aware test hook scopes the
// mutation to this source while semantic rows execute concurrently.
func driveCaptureMutation(t *testing.T, c draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"src"}},"skills":[{"name":"review","from":"s","directory":"skills/review"}]}`
	project, home := draftProject(t, payload, map[string]string{"src/skills/review": "review"})
	packageRoot := filepath.Join(project, "src", "skills", "review")
	if err := os.WriteFile(filepath.Join(packageRoot, "SKILL.md"), []byte("---\nname: review\ndescription: captured A\n---\n# review\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	canonicalPackageRoot, err := staging.Canonicalize(packageRoot)
	if err != nil {
		t.Fatal(err)
	}
	var matched bool
	var hookErr error
	restoreHook := snapshot.SetCaptureAfterCopyHookForTesting(func(acquisition *snapshot.LocalAcquisition) {
		if acquisition.Physical != canonicalPackageRoot {
			return
		}
		matched = true
		hookErr = os.WriteFile(filepath.Join(packageRoot, "SKILL.md"), []byte("---\nname: review\ndescription: captured B\n---\n# review\n"), 0o644)
	})
	defer restoreHook()

	_, resolveErr := closure.ResolveDraft(closure.DraftResolveConfig{
		ProjectRoot:     project,
		Home:            home,
		Manifest:        draftManifest(t, project, payload),
		ManifestPayload: []byte(payload),
	})
	if !matched {
		t.Fatal("capture hook did not observe this production acquisition")
	}
	if hookErr != nil {
		t.Fatalf("capture hook mutation: %v", hookErr)
	}
	if resolveErr == nil || !strings.Contains(resolveErr.Error(), c.Expected) {
		t.Fatalf("ResolveDraft err = %v, want %q", resolveErr, c.Expected)
	}
	if _, statErr := os.Stat(sourcelock.PathIn(project)); !os.IsNotExist(statErr) {
		t.Fatalf("failed capture wrote a lock: stat err = %v", statErr)
	}
}

// driveFrozenCopyMutation mutates the staged frozen copy between audit
// and publication and proves PublishLocal refuses with
// source_snapshot_changed and stores nothing.
func driveFrozenCopyMutation(t *testing.T, _ draftSemanticCase) {
	base := t.TempDir()
	home := t.TempDir()
	pkg := filepath.Join(base, "pkg")
	writeDraftSkill(t, pkg, "review")
	extra := filepath.Join(pkg, "a.md")
	if err := os.WriteFile(extra, []byte("audited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	acquisition, err := snapshot.PrepareLocalAcquisition(pkg, ".", base, "s", []string{filepath.Join(base, ".agents")}, nil, map[string]bool{"s": true}, "")
	if err != nil {
		t.Fatalf("PrepareLocalAcquisition: %v", err)
	}
	t.Cleanup(func() { _ = acquisition.Close() })
	inventory, err := snapshot.Capture(acquisition)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(acquisition.Staging, "a.md"), []byte("mutated\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot.PublishLocal(home, acquisition.Staging, inventory); err == nil || !strings.HasPrefix(err.Error(), "source_snapshot_changed") {
		t.Fatalf("PublishLocal err = %v, want leading source_snapshot_changed", err)
	}
	if entries, statErr := os.ReadDir(snapshot.LocalStoreDir(home)); statErr == nil && len(entries) != 0 {
		t.Fatalf("refused publication left store state behind: %v", entries)
	}
}

// driveLocalGitDirty proves a local path means admitted filesystem
// bytes even inside Git: dirty tracked bytes and untracked files enter
// the inventory (snapshot-B-and-C), .git metadata does not.
func driveLocalGitDirty(t *testing.T, _ draftSemanticCase) {
	requireGit(t)
	repo := t.TempDir()
	writeDraftSkill(t, filepath.Join(repo, "pkg"), "review")
	tracked := filepath.Join(repo, "pkg", "tracked.md")
	if err := os.WriteFile(tracked, []byte("head-A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDraftGit(t, repo, "init", "-q")
	runDraftGit(t, repo, "add", ".")
	runDraftGit(t, repo, "commit", "-qm", "head")
	if err := os.WriteFile(tracked, []byte("filesystem-B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	untracked := filepath.Join(repo, "pkg", "untracked.md")
	if err := os.WriteFile(untracked, []byte("untracked-C\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	pkg := filepath.Join(repo, "pkg")
	acquisition, err := snapshot.PrepareLocalAcquisition(pkg, ".", repo, "s", []string{filepath.Join(repo, ".agents")}, nil, map[string]bool{"s": true}, "")
	if err != nil {
		t.Fatalf("PrepareLocalAcquisition: %v", err)
	}
	t.Cleanup(func() { _ = acquisition.Close() })
	inventory, err := snapshot.Capture(acquisition)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	byPath := map[string]string{}
	for _, entry := range inventory.Files {
		byPath[entry.Path] = entry.SHA256
		if strings.HasPrefix(entry.Path, ".git/") || entry.Path == ".git" {
			t.Fatalf("inventory admits git metadata: %s", entry.Path)
		}
	}
	for rel, content := range map[string]string{"tracked.md": "filesystem-B\n", "untracked.md": "untracked-C\n"} {
		want := "sha256:" + sha256Hex(content)
		if byPath[rel] != want {
			t.Fatalf("%s sha = %s, want %s (snapshot-B-and-C)", rel, byPath[rel], want)
		}
	}
}

// driveRuntimeOnlyRefresh edits one runtime script, refreshes, and
// proves the package identity changes while context bytes stay frozen.
func driveRuntimeOnlyRefresh(t *testing.T, c draftSemanticCase) {
	driveSingleInputRefresh(t, c, "scripts/run.sh", "echo two\n")
}

// driveBuildOnlyRefresh edits one build input, refreshes, and proves
// the package and cache identity change.
func driveBuildOnlyRefresh(t *testing.T, c draftSemanticCase) {
	driveSingleInputRefresh(t, c, "build/main.go", "package changed\n")
}

func driveSingleInputRefresh(t *testing.T, _ draftSemanticCase, rel, after string) {
	t.Helper()
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home := draftProject(t, payload, map[string]string{"skills/review": "review"})
	pkg := filepath.Join(project, "skills", "review")
	for file, content := range map[string]string{"scripts/run.sh": "echo one\n", "build/main.go": "package main\n"} {
		full := filepath.Join(pkg, filepath.FromSlash(file))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	first := resolveDraftPlan(t, project, home, payload)
	before, ok := first.Lock.Find("review")
	if !ok {
		t.Fatal("lock misses review")
	}
	target := filepath.Join(pkg, filepath.FromSlash(rel))
	beforeBytes, err := os.ReadFile(filepath.Join(pkg, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(after), 0o644); err != nil {
		t.Fatal(err)
	}
	second := refreshDraftLocked(t, project, home, payload)
	refreshed, ok := second.Lock.Find("review")
	if !ok {
		t.Fatal("refreshed lock misses review")
	}
	if refreshed.Package.Snapshot == before.Package.Snapshot {
		t.Fatalf("refresh kept package %s after a %s edit", before.Package.Snapshot, rel)
	}
	afterBytes, err := os.ReadFile(filepath.Join(pkg, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(afterBytes) != string(beforeBytes) {
		t.Fatal("refresh changed SKILL.md bytes")
	}
	if result := draftInstall(t, project, home, install.Options{}); result.Status != "ok" {
		t.Fatalf("install after refresh = %+v", result)
	}
}

// driveMissingAuditReport establishes an allow binding, removes the
// evidence report, and proves both the read-only and the mutating
// CheckSourceAudit paths reject without re-establishing.
func driveMissingAuditReport(t *testing.T, _ draftSemanticCase) {
	cfg, subject := draftAuditWitness(t, "advisory", "high", "echo ok\n")
	now := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	if _, err := audit.CheckSourceAudit(cfg, subject, true, now); err != nil {
		t.Fatalf("establish: %v", err)
	}
	objectPath, err := audit.SourceAuditPath(cfg.Home(), subject.Package)
	if err != nil {
		t.Fatal(err)
	}
	evidencePath, err := audit.SourceEvidencePath(cfg.Home(), subject.Package)
	if err != nil {
		t.Fatal(err)
	}
	objectBefore, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(evidencePath); err != nil {
		t.Fatal(err)
	}
	if _, err := audit.CheckSourceAudit(cfg, subject, false, now); err == nil || !strings.Contains(err.Error(), "source_audit") {
		t.Fatalf("read-only err = %v, want a source_audit refusal", err)
	}
	if _, err := audit.CheckSourceAudit(cfg, subject, true, now); err == nil || !strings.Contains(err.Error(), "source_audit") {
		t.Fatalf("mutating err = %v, want a source_audit refusal", err)
	}
	objectAfter, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(objectBefore) != string(objectAfter) {
		t.Fatal("broken existing binding was overwritten")
	}
	if _, statErr := os.Stat(evidencePath); !os.IsNotExist(statErr) {
		t.Fatal("broken report was silently re-established")
	}
}

func draftAuditWitness(t *testing.T, mode, failOn, script string) (*config.Config, audit.SourceSubject) {
	t.Helper()
	home := t.TempDir()
	cfg := &config.Config{Path: filepath.Join(home, "config.json"), Audit: config.Audit{Enabled: true, Mode: mode, FailOn: failOn, Backend: "null"}}
	frozen := t.TempDir()
	if err := os.MkdirAll(filepath.Join(frozen, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(frozen, "scripts", "tool"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	subject := audit.SourceSubject{
		Name: "review", Source: "review", Snapshot: frozen, SchemaVersion: 4,
		Capabilities:  capabilities.ImplicitNone(),
		Package:       audit.SourcePackage{Kind: "local-snapshot", Snapshot: "sha256:" + strings.Repeat("1", 64)},
		ContentSHA256: "sha256:" + strings.Repeat("2", 64),
		ScriptLabels:  []string{"script-worker-v1:unsupported"},
		AssuranceLabels: []string{
			"policy:curator-artifact-policy-v1",
			"policy-version:1",
			"detectors:curator-artifact-detectors-v1",
			"limits:curator-artifact-limits-v1",
		},
	}
	return cfg, subject
}

// driveStrictNetworkAttestationLocal proves a strict registry policy
// refuses a local-snapshot package at install planning time, before any
// cache or compiler work, and publishes nothing.
func driveStrictNetworkAttestationLocal(t *testing.T, _ draftSemanticCase) {
	payload := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`
	project, home := draftProject(t, payload, map[string]string{"skills/review": "review"})
	resolveDraftPlan(t, project, home, payload)
	before := treeDigest(t, project) + treeDigest(t, home)
	cfg := draftInstallConfig(home)
	cfg.Audit.RegistryPolicy = "strict"
	result := install.Project(cfg, project, "test", install.Options{Platform: draftPlatform()})
	if result.Status != "failed" || !strings.Contains(strings.Join(result.Errors, ";"), "no network attestation identity") {
		t.Fatalf("install = %+v, want the local attestation refusal", result)
	}
	if after := treeDigest(t, project) + treeDigest(t, home); after != before {
		t.Fatal("refused install published state")
	}
}
