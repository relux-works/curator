package main

// Draft source currentness at the CLI production entry.
//
// These tests drive `project resolve`, `install`, and `status` through
// run() with the draft switch set in-process. Status consumes the frozen
// lock only: changed package, lock, attestation, substitution, or
// declared ref is non-current, live-byte mutation without refresh stays
// current, and every check stays read-only with a nonzero --check exit
// on drift. None of these tests run in parallel: they mutate process
// environment per run.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/marker"
	"github.com/relux-works/curator/internal/sourcelock"
)

const draftStatusPayload = `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review"]}]}`

// setupDraftStatusProject resolves one local skill through the CLI and
// returns the config path and project root.
func setupDraftStatusProject(t *testing.T) (configPath, project string) {
	t.Helper()
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	configPath, project = setupCLIProject(t, root)
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(draftStatusPayload), 0o644); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(project, "skills", "review"), "review")
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve = %d\nstderr:\n%s", code, stderr)
	}
	return configPath, project
}

func runInstall(t *testing.T, configPath string) {
	t.Helper()
	if code, _, stderr := capture(t, configPath, "install", "app"); code != exitOK {
		t.Fatalf("install = %d\nstderr:\n%s", code, stderr)
	}
}

func runStatus(t *testing.T, configPath string, args ...string) (int, string, string) {
	t.Helper()
	return capture(t, configPath, append([]string{"status"}, args...)...)
}

// digestTree summarizes one root for the read-only proof: every regular
// file by bytes, links by destination, directories by entry.
func digestTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == ".git" || strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			dest, err := os.Readlink(path)
			if err != nil {
				return err
			}
			out[rel] = "link:" + dest
			return nil
		}
		if !info.IsDir() {
			payload, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			out[rel] = "file:" + string(payload)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func assertTreesEqual(t *testing.T, before, after map[string]string) {
	t.Helper()
	for key, want := range before {
		if after[key] != want {
			t.Fatalf("status changed %s", key)
		}
	}
	for key := range after {
		if _, ok := before[key]; !ok {
			t.Fatalf("status created %s", key)
		}
	}
}

// TestDraftStatusReportsUpToDateThroughCLI is the positive row: a
// resolved and installed draft skill reports up-to-date, --check exits
// zero, the JSON document carries the same verdict, and every check
// leaves the project and manager home untouched.
func TestDraftStatusReportsUpToDateThroughCLI(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	runInstall(t, configPath)
	home := filepath.Dir(configPath)
	beforeProject := digestTree(t, project)
	beforeHome := digestTree(t, home)

	code, stdout, stderr := runStatus(t, configPath, "app")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "app: review up-to-date") {
		t.Fatalf("stdout lacks the up-to-date row:\n%s", stdout)
	}
	if code, _, stderr := runStatus(t, configPath, "--check", "app"); code != exitOK {
		t.Fatalf("status --check = %d, want 0\nstderr:\n%s", code, stderr)
	}
	code, stdout, stderr = runStatus(t, configPath, "--json", "app")
	if code != exitOK {
		t.Fatalf("status --json = %d\nstderr:\n%s", code, stderr)
	}
	var document struct {
		Alias  string            `json:"alias"`
		Skills map[string]string `json:"skills"`
	}
	if err := json.Unmarshal([]byte(stdout), &document); err != nil {
		t.Fatalf("status --json is not a document: %v\n%s", err, stdout)
	}
	if document.Skills["review"] != stateUpToDate {
		t.Fatalf("JSON skills = %v, want review up-to-date", document.Skills)
	}
	assertTreesEqual(t, beforeProject, digestTree(t, project))
	assertTreesEqual(t, beforeHome, digestTree(t, home))
}

// TestDraftStatusNeedsInstallAfterRefreshWithoutInstall proves declared
// currency through the lock: refreshing onto new bytes without
// installing is needs-install with a nonzero --check, and installing
// binds the new generation.
func TestDraftStatusNeedsInstallAfterRefreshWithoutInstall(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	runInstall(t, configPath)
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitOK {
		t.Fatalf("project refresh = %d\nstderr:\n%s", code, stderr)
	}
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "app: review needs-install") {
		t.Fatalf("stdout lacks the needs-install row:\n%s", stdout)
	}
	if code, _, _ := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0, want nonzero for needs-install")
	}
	runInstall(t, configPath)
	if code, stdout, stderr := runStatus(t, configPath, "--check", "app"); code != exitOK {
		t.Fatalf("status --check after install = %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}

// TestDraftStatusLockOnlyChangeNeedsInstall proves the binding lock
// comparison through the CLI: a second member joins the selection while
// review's package is unchanged, so after refresh-without-install review
// is needs-install — its marker binds the previous lock generation.
func TestDraftStatusLockOnlyChangeNeedsInstall(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	runInstall(t, configPath)
	writeCLISkill(t, filepath.Join(project, "skills", "other"), "other")
	widened := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review","other"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(widened), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, stderr := capture(t, configPath, "project", "refresh", "app"); code != exitOK {
		t.Fatalf("project refresh = %d\nstderr:\n%s", code, stderr)
	}
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "app: review needs-install") {
		t.Fatalf("review must be needs-install after a lock-only change:\n%s", stdout)
	}
	if code, _, _ := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0 after a lock-only change")
	}
}

// TestDraftStatusLegacyMarkerNeedsInstall proves a legacy marker in a
// draft project is never current through the CLI: the installation
// predates the lock and must be installed again.
func TestDraftStatusLegacyMarkerNeedsInstall(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	dir := filepath.Join(project, ".agents", "skills", "review")
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := hashing.ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	legacy := &marker.Marker{
		Name: "review", Source: "review", RefKind: "tag", Ref: "v1",
		Commit: strings.Repeat("1", 40), ContentSHA256: hash, Locale: "en",
		Agents: []string{"codex_cli"}, Commands: []string{}, Dependencies: []string{},
		SkillSchemaVersion: 4, RuntimeRoots: []string{}, BuildRoots: []string{},
		InstalledAt: "2026-09-18T00:00:00Z", Files: []string{"SKILL.md", "references/info.md"},
		Activation: &marker.Activation{Context: true, Commands: []string{}},
		Builds:     map[string]marker.Build{},
	}
	if err := marker.Write(dir, legacy); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "app: review needs-install") {
		t.Fatalf("stdout lacks the needs-install row for a legacy marker:\n%s", stdout)
	}
	if code, _, _ := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0, want nonzero for a legacy marker")
	}
}

// TestDraftStatusStaleManifestRefuses proves the lock binds the
// declaration: editing the manifest without refresh is not a per-member
// verdict at all — status prints the source_lock_stale refusal and exits
// nonzero, exactly like every other failed dry run.
func TestDraftStatusStaleManifestRefuses(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	runInstall(t, configPath)
	stale := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review","other"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code == exitOK {
		t.Fatalf("status = 0, want nonzero for a stale lock\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "source_lock_stale") {
		t.Fatalf("stderr lacks the stale-lock refusal:\n%s", stderr)
	}
	if strings.Contains(stdout, "needs-install") {
		t.Fatalf("a stale lock must not yield per-member rows:\n%s", stdout)
	}
	if code, _, stderr := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0, want nonzero for a stale lock\nstderr:\n%s", stderr)
	}
}

// TestDraftStatusStaleEmptyLockRefuses is the zero-member row: a lock
// resolved from an empty selection, then a manifest that declares a skill
// without refresh. The lock is stale and the declared skill is neither
// resolved nor installed, so status prints the refusal and --check is
// nonzero — never an empty, silent success.
func TestDraftStatusStaleEmptyLockRefuses(t *testing.T) {
	withDraftSourcesSwitch(t)
	root := t.TempDir()
	configPath, project := setupCLIProject(t, root)
	empty := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(empty), 0o644); err != nil {
		t.Fatal(err)
	}
	writeCLISkill(t, filepath.Join(project, "skills", "review"), "review")
	if code, _, stderr := capture(t, configPath, "project", "resolve", "app"); code != exitOK {
		t.Fatalf("project resolve(empty) = %d\nstderr:\n%s", code, stderr)
	}
	lock, err := sourcelock.Read(sourcelock.PathIn(project))
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Members) != 0 {
		t.Fatalf("lock members = %d, want the zero-member lock", len(lock.Members))
	}
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(draftStatusPayload), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code == exitOK {
		t.Fatalf("status = 0, want nonzero for a stale zero-member lock\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "source_lock_stale") {
		t.Fatalf("stderr lacks the stale-lock refusal:\n%s", stderr)
	}
	if code, _, stderr := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0, want nonzero for a stale zero-member lock\nstderr:\n%s", stderr)
	}
}

// TestDraftStatusStaleSwappedSelectionRefuses proves the stale refusal
// never describes the wrong skill set: with lock {review} and a manifest
// swapped to include ["other"], the rows must not say "review
// needs-install" — the lock no longer binds the declaration at all.
func TestDraftStatusStaleSwappedSelectionRefuses(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	runInstall(t, configPath)
	writeCLISkill(t, filepath.Join(project, "skills", "other"), "other")
	swapped := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["other"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(swapped), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code == exitOK {
		t.Fatalf("status = 0, want nonzero for a stale lock\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
	}
	if !strings.Contains(stderr, "source_lock_stale") {
		t.Fatalf("stderr lacks the stale-lock refusal:\n%s", stderr)
	}
	if strings.Contains(stdout, "review needs-install") {
		t.Fatalf("a stale lock must not report the undeclared member:\n%s", stdout)
	}
	if code, _, _ := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0, want nonzero for a stale lock")
	}
}

// TestDraftStatusContentDrift proves installed-byte tampering surfaces as
// content-drift with a nonzero --check.
func TestDraftStatusContentDrift(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	runInstall(t, configPath)
	installed := filepath.Join(project, ".agents", "skills", "review", "SKILL.md")
	payload, err := os.ReadFile(installed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(installed, append(payload, []byte("drift")...), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "app: review content-drift") {
		t.Fatalf("stdout lacks the content-drift row:\n%s", stdout)
	}
	if code, _, _ := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0, want nonzero for content-drift")
	}
}

// TestDraftStatusNotInstalled proves a resolved-but-never-installed skill
// reports not-installed with a nonzero --check.
func TestDraftStatusNotInstalled(t *testing.T) {
	configPath, _ := setupDraftStatusProject(t)
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "app: review not-installed") {
		t.Fatalf("stdout lacks the not-installed row:\n%s", stdout)
	}
	if code, _, _ := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0, want nonzero for not-installed")
	}
}

// TestDraftStatusLiveMutationWithoutRefreshStaysCurrent proves status
// consumes frozen inputs: live bytes that moved without an explicit
// refresh never rescan into the verdict, so the installation stays
// up-to-date until refresh rebinds the lock.
func TestDraftStatusLiveMutationWithoutRefreshStaysCurrent(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	runInstall(t, configPath)
	if err := os.WriteFile(filepath.Join(project, "skills", "review", "references", "extra.md"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, stdout, stderr := runStatus(t, configPath, "--check", "app"); code != exitOK {
		t.Fatalf("status --check = %d, want 0 without a refresh\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
}

// TestDraftStatusInvalidMarker proves an unreadable installed marker is
// never current.
func TestDraftStatusInvalidMarker(t *testing.T) {
	configPath, project := setupDraftStatusProject(t)
	runInstall(t, configPath)
	if err := os.WriteFile(filepath.Join(project, ".agents", "skills", "review", marker.Name), []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runStatus(t, configPath, "app")
	if code != exitOK {
		t.Fatalf("status = %d\nstderr:\n%s", code, stderr)
	}
	if !strings.Contains(stdout, "app: review invalid-marker") {
		t.Fatalf("stdout lacks the invalid-marker row:\n%s", stdout)
	}
	if code, _, _ := runStatus(t, configPath, "--check", "app"); code == exitOK {
		t.Fatalf("status --check = 0, want nonzero for invalid-marker")
	}
}

// writeDraftStatusMarker stages one installed v5 marker whose content
// hash matches its bytes, so the classifier rows below decide on
// identity, never on drift.
func writeDraftStatusMarker(t *testing.T, skillsDir string, m *marker.Marker) {
	t.Helper()
	dir := filepath.Join(skillsDir, m.Name)
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
		t.Fatal(err)
	}
	hash, err := hashing.ContentSHA256(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	m.ContentSHA256 = hash
	m.Files = []string{"SKILL.md", "references/info.md"}
	if err := marker.Write(dir, m); err != nil {
		t.Fatalf("Write = %v", err)
	}
}

func draftStatusMember(snapshot string) sourcelock.Member {
	return sourcelock.Member{
		Name:          "review",
		Package:       sourcelock.Package{Kind: "local-snapshot", Snapshot: snapshot},
		Directory:     "skills/review",
		ContentSHA256: "sha256:" + strings.Repeat("e", 64),
	}
}

func draftStatusGitMember() sourcelock.Member {
	return sourcelock.Member{
		Name: "review",
		Package: sourcelock.Package{Kind: "network-git", Repository: "example.org/kit",
			Commit: sourcelock.Commit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}, Directory: "skills/review"},
		Directory:     "skills/review",
		ContentSHA256: "sha256:" + strings.Repeat("e", 64),
	}
}

func draftStatusGitRecorded(lockSHA string) *marker.Marker {
	m := draftStatusRecorded("sha256:"+strings.Repeat("a", 64), lockSHA)
	m.Package = &marker.Package{Kind: "network-git", Repository: "example.org/kit",
		Commit: &marker.Commit{ObjectFormat: "sha1", Hex: strings.Repeat("1", 40)}, Directory: "skills/review"}
	return m
}

func draftStatusRecorded(snapshot, lockSHA string) *marker.Marker {
	return &marker.Marker{
		Name:               "review",
		Package:            &marker.Package{Kind: "local-snapshot", Snapshot: snapshot},
		LockSHA256:         lockSHA,
		Locale:             "en",
		Agents:             []string{"codex_cli"},
		Commands:           []string{},
		Dependencies:       []string{},
		SkillSchemaVersion: 4,
		RuntimeRoots:       []string{},
		BuildRoots:         []string{},
		InstalledAt:        "2026-09-18T00:00:00Z",
		Activation:         &marker.Activation{Context: true, Commands: []string{}},
		Requirers:          []string{"<project>"},
		Builds:             map[string]marker.Build{},
	}
}

// TestDraftStatusAppliesKeepsLegacyProjectsUntouched pins the diversion
// guard: only a schema-2 manifest with the draft switch takes the draft
// lane. Every other shape keeps the legacy surface byte-identically.
func TestDraftStatusAppliesKeepsLegacyProjectsUntouched(t *testing.T) {
	write := func(t *testing.T, payload string) string {
		t.Helper()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "Skillfile.json"), []byte(payload), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	schema2 := write(t, `{"schema_version":2,"sources":{},"skills":[]}`)
	schema1 := write(t, `{"schema_version":1,"skills":[]}`)

	if draftStatusApplies(schema2) {
		t.Fatalf("schema-2 without the switch takes the draft lane")
	}
	if draftStatusApplies(schema1) {
		t.Fatalf("schema-1 takes the draft lane")
	}
	if draftStatusApplies(t.TempDir()) {
		t.Fatalf("a missing manifest takes the draft lane")
	}
	withDraftSourcesSwitch(t)
	if !draftStatusApplies(schema2) {
		t.Fatalf("schema-2 with the switch keeps the legacy lane")
	}
	if draftStatusApplies(schema1) {
		t.Fatalf("schema-1 with the switch takes the draft lane")
	}
}

// TestDraftStatusDriftFailsClosedWithoutFrozenInputs pins the race the
// dry run cannot close: a schema-2 manifest whose lock vanished before
// classification yields no verdict at all, so the caller exits nonzero
// read-only instead of reporting an empty skill set as current.
func TestDraftStatusDriftFailsClosedWithoutFrozenInputs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Skillfile.json"), []byte(draftStatusPayload), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := draftStatusDrift(dir, filepath.Join(dir, ".agents", "skills"), nil); err == nil {
		t.Fatalf("drift without a lock returned a verdict")
	}
}

// TestDraftStatusDriftFailsClosedOnStaleLock pins the drift-level stale
// branch: a lock that no longer binds the declaration is an error, never
// rows — even though the CLI normally refuses before classification.
func TestDraftStatusDriftFailsClosedOnStaleLock(t *testing.T) {
	_, project := setupDraftStatusProject(t)
	stale := `{"schema_version":2,"sources":{"s":{"path":"."}},"skills":[{"from":"s","directory":"skills","include":["review","other"]}]}`
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	if rows, err := draftStatusDrift(project, filepath.Join(project, ".agents", "skills"), nil); err == nil {
		t.Fatalf("drift with a stale lock returned rows: %v", rows)
	}
}

// TestClassifyDraftMemberComparesExactCurrentness pins every comparison
// of the per-member verdict: package, lock, attestation (registry,
// status, key, and presence in both directions), and substitution must
// all match the frozen plan, or the member is needs-install.
func TestClassifyDraftMemberComparesExactCurrentness(t *testing.T) {
	snapshot := "sha256:" + strings.Repeat("a", 64)
	lockSHA := "sha256:" + strings.Repeat("b", 64)
	attested := func() *marker.Attestation {
		return &marker.Attestation{Registry: "example.org/kit", Status: "audited", KeyID: "0123456789abcdef"}
	}
	cases := []struct {
		name   string
		git    bool
		mutate func(member *sourcelock.Member, lock *string, recorded *marker.Marker, effective **marker.Attestation)
		want   string
	}{
		{"current", false, func(*sourcelock.Member, *string, *marker.Marker, **marker.Attestation) {}, stateUpToDate},
		{"current-attested", true, func(_ *sourcelock.Member, _ *string, recorded *marker.Marker, effective **marker.Attestation) {
			recorded.Attestation = attested()
			*effective = attested()
		}, stateUpToDate},
		{"package-mismatch", false, func(member *sourcelock.Member, _ *string, _ *marker.Marker, _ **marker.Attestation) {
			member.Package.Snapshot = "sha256:" + strings.Repeat("f", 64)
		}, stateNeedsInstall},
		{"package-mismatch-git", true, func(member *sourcelock.Member, _ *string, _ *marker.Marker, _ **marker.Attestation) {
			member.Package.Commit.Hex = strings.Repeat("2", 40)
		}, stateNeedsInstall},
		{"lock-mismatch", false, func(_ *sourcelock.Member, lock *string, _ *marker.Marker, _ **marker.Attestation) {
			*lock = "sha256:" + strings.Repeat("c", 64)
		}, stateNeedsInstall},
		{"attestation-registry", true, func(_ *sourcelock.Member, _ *string, recorded *marker.Marker, effective **marker.Attestation) {
			recorded.Attestation = attested()
			changed := attested()
			changed.Registry = "other.test/kit"
			*effective = changed
		}, stateNeedsInstall},
		{"attestation-status", true, func(_ *sourcelock.Member, _ *string, recorded *marker.Marker, effective **marker.Attestation) {
			recorded.Attestation = attested()
			changed := attested()
			changed.Status = "deprecated"
			*effective = changed
		}, stateNeedsInstall},
		{"attestation-key", true, func(_ *sourcelock.Member, _ *string, recorded *marker.Marker, effective **marker.Attestation) {
			recorded.Attestation = attested()
			changed := attested()
			changed.KeyID = "fedcba9876543210"
			*effective = changed
		}, stateNeedsInstall},
		{"attestation-recorded-absent", true, func(_ *sourcelock.Member, _ *string, _ *marker.Marker, effective **marker.Attestation) {
			*effective = attested()
		}, stateNeedsInstall},
		{"attestation-effective-absent", true, func(_ *sourcelock.Member, _ *string, recorded *marker.Marker, _ **marker.Attestation) {
			recorded.Attestation = attested()
		}, stateNeedsInstall},
		{"substituted", true, func(_ *sourcelock.Member, _ *string, recorded *marker.Marker, _ **marker.Attestation) {
			recorded.Substituted = "dev-here"
		}, stateNeedsInstall},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			skillsDir := t.TempDir()
			member := draftStatusMember(snapshot)
			recorded := draftStatusRecorded(snapshot, lockSHA)
			if row.git {
				member = draftStatusGitMember()
				recorded = draftStatusGitRecorded(lockSHA)
			}
			lock := lockSHA
			var effective *marker.Attestation
			row.mutate(&member, &lock, recorded, &effective)
			writeDraftStatusMarker(t, skillsDir, recorded)
			if got := classifyDraftMember(skillsDir, member, lock, effective); got != row.want {
				t.Fatalf("classifyDraftMember = %q, want %q", got, row.want)
			}
		})
	}
}

// TestClassifyDraftMemberPresenceRows pins the non-identity verdicts: a
// legacy marker can never describe a locked package, drifted bytes are
// content-drift, a missing directory is not-installed, and an unreadable
// one is unresolvable — never current, never absent.
func TestClassifyDraftMemberPresenceRows(t *testing.T) {
	snapshot := "sha256:" + strings.Repeat("a", 64)
	lockSHA := "sha256:" + strings.Repeat("b", 64)
	member := draftStatusMember(snapshot)

	t.Run("legacy-marker", func(t *testing.T) {
		skillsDir := t.TempDir()
		dir := filepath.Join(skillsDir, "review")
		if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "references", "info.md"), []byte("context"), 0o644); err != nil {
			t.Fatal(err)
		}
		hash, err := hashing.ContentSHA256(dir, nil)
		if err != nil {
			t.Fatal(err)
		}
		legacy := &marker.Marker{
			Name: "review", Source: "review", RefKind: "tag", Ref: "v1",
			Commit: strings.Repeat("1", 40), ContentSHA256: hash, Locale: "en",
			Agents: []string{"codex_cli"}, Commands: []string{}, Dependencies: []string{},
			SkillSchemaVersion: 4, RuntimeRoots: []string{}, BuildRoots: []string{},
			InstalledAt: "2026-09-18T00:00:00Z", Files: []string{"SKILL.md", "references/info.md"},
			Activation: &marker.Activation{Context: true, Commands: []string{}},
			Builds:     map[string]marker.Build{},
		}
		if err := marker.Write(dir, legacy); err != nil {
			t.Fatal(err)
		}
		if got := classifyDraftMember(skillsDir, member, lockSHA, nil); got != stateNeedsInstall {
			t.Fatalf("classifyDraftMember = %q, want needs-install", got)
		}
	})

	t.Run("content-drift", func(t *testing.T) {
		skillsDir := t.TempDir()
		writeDraftStatusMarker(t, skillsDir, draftStatusRecorded(snapshot, lockSHA))
		drifted := filepath.Join(skillsDir, "review", "SKILL.md")
		payload, err := os.ReadFile(drifted)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(drifted, append(payload, []byte("drift")...), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := classifyDraftMember(skillsDir, member, lockSHA, nil); got != stateContentDrift {
			t.Fatalf("classifyDraftMember = %q, want content-drift", got)
		}
	})

	t.Run("not-installed", func(t *testing.T) {
		if got := classifyDraftMember(t.TempDir(), member, lockSHA, nil); got != stateNotInstalled {
			t.Fatalf("classifyDraftMember = %q, want not-installed", got)
		}
	})

	t.Run("unreadable", func(t *testing.T) {
		skillsDir := t.TempDir()
		writeDraftStatusMarker(t, skillsDir, draftStatusRecorded(snapshot, lockSHA))
		// Deny the parent: Lstat of the child then fails with
		// something other than absence, which must stay
		// unresolvable rather than collapsing into not-installed.
		if err := os.Chmod(skillsDir, 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(skillsDir, 0o755) })
		if _, err := os.Lstat(filepath.Join(skillsDir, "review")); err == nil {
			t.Skip("this environment can read a mode-000 directory")
		}
		if got := classifyDraftMember(skillsDir, member, lockSHA, nil); got != stateUnresolvable {
			t.Fatalf("classifyDraftMember = %q, want unresolvable", got)
		}
	})
}
