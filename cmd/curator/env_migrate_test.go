package main

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
)

// These tests drive the production run() entry point for the env
// migrate rows: the explicit credential migration step on temporary
// stores (environments §7.4, §10.1; manager §12.4).

// requireLinkCapability skips the row when the host cannot create the
// symlinks the fixture or the production repair needs — the Windows
// privilege pattern shared with the envprofile 0017 rows. The reason
// classifies host-capability in .github/ci/skip-classes.tsv via the
// existing "this host cannot create" entry; ledger rows tolerate it on
// Windows only, so a skip on a unix runner still fails the gate. A
// symlink failure after this probe passed is a real defect and fails,
// never skips: direct os.Symlink sites keep t.Fatal on purpose.
func requireLinkCapability(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("probe\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "link")); err != nil {
		t.Skipf("this host cannot create symlinks: %v", err)
	}
}

// installPiProfile installs the acme profile and provisions its Pi
// home through the CLI, returning the managed link, the old native
// target, and the agent-root target.
func installPiProfile(t *testing.T, source stubConfigSource) (link, oldTarget, agentAuth string) {
	t.Helper()
	writeNativeCredentials(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "pi", "--repair"); code != exitOK {
		t.Fatalf("resolve pi --repair = %d\nstderr:\n%s", code, stderr)
	}
	managed := filepath.Join(source.cfg.Home(), "environments", "acme", "pi")
	link = filepath.Join(managed, "auth.json")
	native := os.Getenv("PI_CODING_AGENT_DIR")
	oldTarget = filepath.Join(native, "auth.json")
	agentAuth = filepath.Join(native, "agent", "auth.json")
	if target, err := os.Readlink(link); err != nil || target != agentAuth {
		t.Fatalf("provisioned Pi link targets %q (%v), want %q", target, err, agentAuth)
	}
	return link, oldTarget, agentAuth
}

func planHash(t *testing.T, plan string) string {
	t.Helper()
	match := regexp.MustCompile(`credential migration plan ([0-9a-f]{64})`).FindStringSubmatch(plan)
	if match == nil {
		t.Fatalf("plan prints its hash:\n%s", plan)
	}
	return match[1]
}

// TestEnvMigratePlanApplyPi migrates the operator's Pi case through the
// CLI: the mis-targeted home plans a printed relink, apply executes
// exactly that plan with bytes preserved, and a second plan/apply round
// proves drift refusal on a stale hash.
func TestEnvMigratePlanApplyPi(t *testing.T) {
	requireLinkCapability(t)
	source, _ := profileHome(t)
	link, oldTarget, agentAuth := installPiProfile(t, source)
	want, err := os.ReadFile(agentAuth)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	code, inspect, _ := runProfile(t, source, "env", "migrate", "--inspect")
	if code != exitOK {
		t.Fatalf("migrate --inspect = %d\n%s", code, inspect)
	}
	for _, want := range []string{"mis-targeted", oldTarget, agentAuth, "native Pi roots"} {
		if !strings.Contains(inspect, want) {
			t.Fatalf("inspect inventories the case, want %q in:\n%s", want, inspect)
		}
	}
	code, plan, _ := runProfile(t, source, "env", "migrate", "--plan")
	if code != exitOK {
		t.Fatalf("migrate --plan = %d\n%s", code, plan)
	}
	hash := planHash(t, plan)
	for _, want := range []string{"relink auth.json", oldTarget, agentAuth, "--expect " + hash} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan prints the exact operation, want %q in:\n%s", want, plan)
		}
	}
	if code, again, _ := runProfile(t, source, "env", "migrate", "--plan"); code != exitOK || again != plan {
		t.Fatalf("planning is deterministic: %d, identical bytes: %v", code, again == plan)
	}
	code, applied, stderr := runProfile(t, source, "env", "migrate", "--apply", "--expect", hash)
	if code != exitOK {
		t.Fatalf("migrate --apply = %d\nstdout:\n%s\nstderr:\n%s", code, applied, stderr)
	}
	if !strings.Contains(applied, plan) {
		t.Fatalf("apply prints the plan before applying:\n%s", applied)
	}
	if !strings.Contains(applied, "migration complete") {
		t.Fatalf("apply reports completion:\n%s", applied)
	}
	if target, err := os.Readlink(link); err != nil || target != agentAuth {
		t.Fatalf("the link now targets the agent root: %q (%v)", target, err)
	}
	if payload, err := os.ReadFile(agentAuth); err != nil || string(payload) != string(want) {
		t.Fatalf("native bytes preserved: %q (%v)", payload, err)
	}
	markerPath := filepath.Join(source.cfg.Home(), "environments", "acme", "pi", envmarker.Name)
	markerBytes, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := envmarker.Parse(markerBytes)
	if err != nil {
		t.Fatal(err)
	}
	if marker.Version != envmarker.VersionV2 || marker.Passthrough == nil || len(*marker.Passthrough) != 1 {
		t.Fatalf("migration marker carries one schema-2 record: %+v", marker)
	}
	credential := (*marker.Passthrough)[0]
	if credential.Path != "auth.json" || credential.Isolation != "shared" || credential.Strategy != "file-link" ||
		credential.SourceRole != "native" || credential.Backend != "file" || credential.BackendVersion != "0.84.2" || credential.Provenance != "migrated" {
		t.Fatalf("migration credential record: %+v", credential)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "pi"); code != exitOK {
		t.Fatalf("the migrated home resolves bare: %d\n%s", code, stderr)
	}
	// A second round proves drift refusal through the CLI: plan clean,
	// break the link, apply with the stale hash.
	_, cleanPlan, _ := runProfile(t, source, "env", "migrate", "--plan")
	cleanHash := planHash(t, cleanPlan)
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	code, _, stderr = runProfile(t, source, "env", "migrate", "--apply", "--expect", cleanHash)
	if code != exitFail || !strings.Contains(stderr, "plan drift") {
		t.Fatalf("a stale --expect refuses with plan drift: %d\n%s", code, stderr)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("a drifted apply writes nothing: %q (%v)", target, err)
	}
}

// TestEnvMigrateConflictRefuses proves a regular file at a link path
// refuses through the CLI naming the operator choice, with bytes
// untouched.
func TestEnvMigrateConflictRefuses(t *testing.T) {
	requireLinkCapability(t)
	source, _ := profileHome(t)
	writeNativeCredentials(t)
	pkg := t.TempDir()
	writeContextPackage(t, pkg, "acme", "1.0.0", "hello\n")
	if code, _, stderr := runProfile(t, source, "profile", "install", pkg); code != exitOK {
		t.Fatalf("install stderr:\n%s", stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "resolve", "codex_cli", "--repair"); code != exitOK {
		t.Fatalf("resolve codex_cli --repair = %d\nstderr:\n%s", code, stderr)
	}
	link := filepath.Join(source.cfg.Home(), "environments", "acme", "codex_cli", "auth.json")
	_ = os.Remove(link)
	managed := []byte("managed-credential\n")
	if err := os.WriteFile(link, managed, 0o600); err != nil {
		t.Fatal(err)
	}
	nativeAuth := filepath.Join(os.Getenv("CODEX_HOME"), "auth.json")
	native, err := os.ReadFile(nativeAuth)
	if err != nil {
		t.Fatal(err)
	}
	code, plan, _ := runProfile(t, source, "env", "migrate", "--plan")
	if code != exitOK || !strings.Contains(plan, "blocked:") {
		t.Fatalf("a conflicted plan prints blocked: %d\n%s", code, plan)
	}
	code, _, stderr := runProfile(t, source, "env", "migrate", "--apply", "--expect", planHash(t, plan))
	if code != exitFail {
		t.Fatalf("apply = %d, want %d", code, exitFail)
	}
	combined := plan + "\n" + stderr
	for _, want := range []string{"environment_credential_conflict", "decide which credential bytes win", link} {
		if !strings.Contains(combined, want) {
			t.Fatalf("the refusal names %q in:\n%s", want, combined)
		}
	}
	if payload, err := os.ReadFile(link); err != nil || string(payload) != string(managed) {
		t.Fatalf("managed bytes untouched: %q (%v)", payload, err)
	}
	if payload, err := os.ReadFile(nativeAuth); err != nil || string(payload) != string(native) {
		t.Fatalf("native bytes untouched: %q (%v)", payload, err)
	}
}

// TestEnvResolveRepairNeedsMigration proves repair never migrates
// through the CLI: a mis-targeted Pi home refuses pointing at the
// explicit step, with the link untouched.
func TestEnvResolveRepairNeedsMigration(t *testing.T) {
	requireLinkCapability(t)
	source, _ := profileHome(t)
	link, oldTarget, agentAuth := installPiProfile(t, source)
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runProfile(t, source, "env", "resolve", "pi", "--repair")
	if code != exitFail {
		t.Fatalf("resolve --repair = %d, want %d", code, exitFail)
	}
	for _, want := range []string{"environment_credential_conflict", "migration needed", "env migrate", link, oldTarget, agentAuth} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("the refusal points at the explicit step, want %q in:\n%s", want, stderr)
		}
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the refused link is untouched: %q (%v)", target, err)
	}
}

// TestEnvMigrateUsage covers the operand gates through the CLI.
func TestEnvMigrateUsage(t *testing.T) {
	source, _ := profileHome(t)
	if code, _, _ := runProfile(t, source, "env", "migrate"); code != exitUsage {
		t.Fatalf("migrate without a phase = %d, want %d", code, exitUsage)
	}
	if code, _, _ := runProfile(t, source, "env", "migrate", "--plan", "--apply"); code != exitUsage {
		t.Fatalf("migrate with two phases = %d, want %d", code, exitUsage)
	}
	if code, _, stderr := runProfile(t, source, "env", "migrate", "--plan", "--expect", "abc"); code != exitUsage || !strings.Contains(stderr, "--expect applies only with --apply") {
		t.Fatalf("migrate --plan --expect = %d\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "migrate", "--apply"); code != exitUsage || !strings.Contains(stderr, "--apply needs --expect") {
		t.Fatalf("migrate --apply without --expect = %d\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "migrate", "--plan", "--env", "cursor"); code != exitFail || !strings.Contains(stderr, "environment_unknown") {
		t.Fatalf("migrate --env cursor = %d\n%s", code, stderr)
	}
	if code, _, stderr := runProfile(t, source, "env", "migrate", "--plan", "--profile", "ghost"); code != exitFail || !strings.Contains(stderr, "profile_unknown") {
		t.Fatalf("migrate --profile ghost = %d\n%s", code, stderr)
	}
	if code, stdout, _ := runProfile(t, source, "env", "migrate", "--inspect"); code != exitOK || !strings.Contains(stdout, "credential migration inventory") {
		t.Fatalf("migrate --inspect = %d\n%s", code, stdout)
	}
}

// TestEnvMigrateApplyRequiresPlan is the committed reviewer missing-plan
// probe: --apply without a prior --plan hash refuses (usage) with the
// link untouched. The library half is TestMigrateApplyRequiresPlan.
func TestEnvMigrateApplyRequiresPlan(t *testing.T) {
	requireLinkCapability(t)
	source, _ := profileHome(t)
	link, oldTarget, _ := installPiProfile(t, source)
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runProfile(t, source, "env", "migrate", "--apply")
	if code != exitUsage {
		t.Fatalf("apply without a prior plan = %d, want %d\n%s", code, exitUsage, stderr)
	}
	if !strings.Contains(stderr, "--apply needs --expect") {
		t.Fatalf("the refusal names the missing hash:\n%s", stderr)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the refused link is untouched: %q (%v)", target, err)
	}
}

// planOrderWriter observes the link state on the first stdout write:
// the committed reviewer print-before-write probe. The plan must print
// before the first mutation, so the link still aims at the old target.
type planOrderWriter struct {
	t    *testing.T
	link string
	old  string
	seen bool
}

func (w *planOrderWriter) Write(p []byte) (int, error) {
	if !w.seen {
		w.seen = true
		got, err := os.Readlink(w.link)
		if err != nil || got != w.old {
			w.t.Errorf("first stdout write occurs AFTER mutation: target=%q err=%v; wanted prior %q", got, err, w.old)
		}
	}
	return len(p), nil
}

// TestEnvMigratePrintBeforeWrite proves the temporal order through the
// CLI: the locked plan reaches stdout before the first link mutation,
// not as a post-hoc echo. A final-string containment check cannot prove
// this; the writer hook can. A mutant printing after the apply moves
// the link first and fails the hook.
func TestEnvMigratePrintBeforeWrite(t *testing.T) {
	requireLinkCapability(t)
	source, _ := profileHome(t)
	link, oldTarget, agentAuth := installPiProfile(t, source)
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	_, plan, _ := runProfile(t, source, "env", "migrate", "--plan")
	writer := &planOrderWriter{t: t, link: link, old: oldTarget}
	var stderr strings.Builder
	code := run([]string{"env", "migrate", "--apply", "--expect", planHash(t, plan)}, source, writer, &stderr)
	if code != exitOK {
		t.Fatalf("migrate --apply = %d\n%s", code, stderr.String())
	}
	if !writer.seen {
		t.Fatal("apply prints the plan")
	}
	if target, err := os.Readlink(link); err != nil || target != agentAuth {
		t.Fatalf("the link migrated after the print: %q (%v)", target, err)
	}
}

// TestEnvResolveRepairFailedOnUninspectable proves the CLI surfaces the
// repair_failed refusal: a correctly targeted Pi link whose native
// target cannot be inspected fails env resolve --repair with
// environment_repair_failed carrying the inspection reason, with the
// link untouched. Unix-only: the uninspectable-target fixture is the
// ENOTDIR shape Windows maps to path-not-found. The library half is
// TestCredentialLinkInspectionRepairFails.
func TestEnvResolveRepairFailedOnUninspectable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this host cannot create an uninspectable link target: Windows maps a stat through a file to path-not-found, so the inspection diagnostic runs on the unix runners")
	}
	source, _ := profileHome(t)
	link, _, agentAuth := installPiProfile(t, source)
	native := os.Getenv("PI_CODING_AGENT_DIR")
	// Break the target's parent: a file where the agent directory was.
	if err := os.RemoveAll(filepath.Join(native, "agent")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "agent"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runProfile(t, source, "env", "resolve", "pi", "--repair")
	if code != exitFail {
		t.Fatalf("resolve --repair = %d, want %d\n%s", code, exitFail, stderr)
	}
	for _, want := range []string{"environment_repair_failed", "cannot be inspected", "environment_credential_conflict", "auth.json"} {
		if !strings.Contains(stderr, want) {
			t.Fatalf("the refusal carries the inspection reason, want %q in:\n%s", want, stderr)
		}
	}
	if target, err := os.Readlink(link); err != nil || target != agentAuth {
		t.Fatalf("repair touches nothing: %q (%v)", target, err)
	}
}
