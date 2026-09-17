package envprofile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextstore"
	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/hookapproval"
)

// Production entry points under test: StatusOf, Remove (orphan retention),
// and the seed, fallback, collision, and repair-failed edges of Resolve.

func statusRequest(fx *managedFixture) StatusRequest {
	return StatusRequest{
		Home:    fx.home,
		Machine: envregistry.DefaultMachineConfig(),
		Detect:  func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: func(id string) (string, error) {
			if id == "opencode" {
				return fx.native["opencode-native"], nil
			}
			return fx.native[id], nil
		},
		OperatorXDG: fx.xdg,
		LaunchDir:   fx.launch,
	}
}

func provision(t *testing.T, fx *managedFixture, envID string, machine envregistry.MachineConfig) {
	t.Helper()
	req := fx.request(envID)
	req.Machine = machine
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatalf("provision %s: %v", envID, err)
	}
}

// TestStatusOrphanRetention proves removal without purge retains the
// managed homes and status reports each orphan by path.
func TestStatusOrphanRetention(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	managed := ManagedHomeDir(fx.home, "acme", "codex_cli")
	if err := SetCurrent(fx.home, "default"); err != nil {
		t.Fatal(err)
	}
	if err := Remove(fx.home, "acme", false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(managed, envmarker.Name)); err != nil {
		t.Fatalf("a retained home keeps its marker: %v", err)
	}
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, orphan := range status.Orphans {
		if orphan == managed {
			found = true
		}
	}
	if !found {
		t.Fatalf("orphans %v name no %s", status.Orphans, managed)
	}
	if !status.NonCurrent {
		t.Fatal("an orphaned home is non-current")
	}
	// A purge removes the homes with markers and backups.
	fx2 := writeManagedFixture(t, "acme2")
	provision2 := fx2.request("codex_cli")
	provision2.Repair = true
	if _, err := Resolve(provision2); err != nil {
		t.Fatal(err)
	}
	if err := Remove(fx2.home, "acme2", true); err != nil {
		// acme2 is current, so removal refuses with profile_in_use; clear
		// first like an operator would.
		if !strings.Contains(err.Error(), DiagInUse) {
			t.Fatal(err)
		}
		if err := SetCurrent(fx2.home, "default"); err != nil {
			t.Fatal(err)
		}
		if err := Remove(fx2.home, "acme2", true); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Lstat(filepath.Join(EnvRoot(fx2.home), "acme2")); !os.IsNotExist(err) {
		t.Fatal("a purge removes the profile's environments directory")
	}
}

// TestStatusShadowAcknowledgment proves the pi shadowing row is
// non-current by default and a current warning under
// shadow_acknowledged.
func TestStatusShadowAcknowledgment(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	homeDir := ManagedHomeDir(fx.home, "acme", "pi")
	if err := os.WriteFile(filepath.Join(homeDir, "AGENTS.override.md"), []byte("override\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := StatusOf(statusRequest(fx))
	if err != nil {
		t.Fatal(err)
	}
	row := findHome(status, "acme", "pi")
	if row == nil || row.Current {
		t.Fatal("a shadowed surface is non-current by default")
	}
	acked := statusRequest(fx)
	acked.Machine.ShadowAcknowledged = []envregistry.ShadowAck{{Env: "pi", Path: "AGENTS.override.md"}}
	status, err = StatusOf(acked)
	if err != nil {
		t.Fatal(err)
	}
	row = findHome(status, "acme", "pi")
	if row == nil || !row.Current {
		t.Fatalf("an acknowledged shadow downgrades to a warning: %+v", row)
	}
	// Resolve stays out of the shadowing row: the recorded surfaces verify.
	if _, err := Resolve(fx.request("pi")); err != nil {
		t.Fatalf("resolve warns but stays current: %v", err)
	}
}

// TestStatusProfileMembersAndUnregistered narrows the §12 rows: the
// matrix carries the lock's context members with weights and the
// precedence primitives per activation, and unregistered env-ids named in
// machine configuration.
func TestStatusProfileMembersAndUnregistered(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	provision(t, fx, "claude_code", envregistry.DefaultMachineConfig())
	req := statusRequest(fx)
	req.Machine.Forms = map[string]string{"cursor": "monolithic", "claude_code": "monolithic"}
	req.Machine.Isolation = map[string]map[string]string{"acme": {"cursor": "shared"}}
	req.Machine.ShadowAcknowledged = []envregistry.ShadowAck{{Env: "ghost", Path: "AGENTS.override.md"}}
	status, err := StatusOf(req)
	if err != nil {
		t.Fatal(err)
	}
	var profile *ProfileState
	for i := range status.Profiles {
		if status.Profiles[i].Profile == "acme" {
			profile = &status.Profiles[i]
		}
	}
	if profile == nil {
		t.Fatalf("profiles %+v name no acme", status.Profiles)
	}
	if len(profile.Members) == 0 {
		t.Fatal("the profile row carries no lock members")
	}
	found := false
	for _, member := range profile.Members {
		if member.Name == "acme" && member.Kind == contextlock.KindContext && member.Weight == 100 {
			found = true
		}
	}
	if !found {
		t.Fatalf("members %+v name no acme context weight 100", profile.Members)
	}
	if profile.Precedence.Winner == "" || profile.Precedence.Placement == "" {
		t.Fatalf("precedence %+v", profile.Precedence)
	}
	if len(status.UnregisteredEnvironments) != 2 || status.UnregisteredEnvironments[0] != "cursor" || status.UnregisteredEnvironments[1] != "ghost" {
		t.Fatalf("unregistered %+v", status.UnregisteredEnvironments)
	}
	row := findHome(status, "acme", "claude_code")
	if row == nil || row.Mode == "" || row.Form == "" {
		t.Fatalf("home row %+v carries no mode/form", row)
	}
	if len(row.SeededProjects) == 0 {
		t.Fatal("the claude_code row carries no seeded-projects")
	}
}

func findHome(status *Status, profile, env string) *HomeState {
	for i := range status.Homes {
		if status.Homes[i].Profile == profile && status.Homes[i].Environment == env {
			return &status.Homes[i]
		}
	}
	return nil
}

// TestSeedUnreadableStopsBeforeFirstWrite narrows the seed gate: a seed
// that exists but cannot be read fails provisioning with
// environment_seed_unreadable before any managed path is written.
func TestSeedUnreadableStopsBeforeFirstWrite(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	// A directory where the seed file should be reads as existing but
	// unreadable-as-a-file: absence and unreadability stay different.
	if err := os.MkdirAll(filepath.Join(fx.native["codex_cli"], "config.toml"), 0o755); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	_, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagSeedUnreadable) {
		t.Fatalf("an unreadable seed must fail with %s, got %v", envregistry.DiagSeedUnreadable, err)
	}
	if _, stat := os.Lstat(filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)); !os.IsNotExist(stat) {
		t.Fatal("provisioning stops before the first write")
	}
}

// TestReservedSkillName narrows the §9.4 gate through Resolve: a skill
// colliding with the curator-* reservation is refused, never materialized
// onto a PATH the §11 dispatch trusts.
func TestReservedSkillName(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	lock, _, err := readLock(fx.home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	pin := strings.Repeat("d", 40)
	writeStoreEntry(t, fx.home, "skill", "curator-evil", pin, map[string]string{"SKILL.md": "# evil\n"})
	lock.Members = append(lock.Members, contextlock.Member{
		Kind: "skill", Name: "curator-evil", Source: "github.com/example/evil",
		Commit: pin, RequiredBy: []string{"acme"},
	})
	lock.Sort()
	if _, err := contextlock.Write(lockPath(fx.home, "acme"), lock); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	_, err = Resolve(req)
	if err == nil || !strings.Contains(err.Error(), DiagReservedCommand) {
		t.Fatalf("a curator-* skill must fail with %s, got %v", DiagReservedCommand, err)
	}
}

// TestProfilePathCollision narrows the §5 gate: two profile names folding
// to one platform path below the environments root fail provisioning.
func TestProfilePathCollision(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	if err := os.MkdirAll(filepath.Join(EnvRoot(fx.home), "ACME"), 0o755); err != nil {
		t.Fatal(err)
	}
	req := fx.request("codex_cli")
	req.Repair = true
	_, err := Resolve(req)
	if err == nil || !strings.Contains(err.Error(), "environment_path_collision") {
		t.Fatalf("a folding profile name must fail, got %v", err)
	}
}

// TestOpencodeFormFallback proves the §5.3 rule: an unmanaged opencode.json
// blocks the referenced form, warns environment_form_unavailable, and
// materializes monolithic without editing the unmanaged file.
func TestOpencodeFormFallback(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	homeDir := ManagedHomeDir(fx.home, "acme", "opencode")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	unmanaged := []byte("{\"instructions\": [\"operator\"]}\n")
	if err := os.WriteFile(filepath.Join(homeDir, "opencode.json"), unmanaged, 0o644); err != nil {
		t.Fatal(err)
	}
	machine := envregistry.DefaultMachineConfig()
	machine.Forms = map[string]string{"opencode": "referenced"}
	req := fx.request("opencode")
	req.Machine = machine
	req.Repair = true
	result, err := Resolve(req)
	if err != nil {
		t.Fatalf("the fallback warns instead of failing: %v", err)
	}
	warned := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "environment_form_unavailable") {
			warned = true
		}
	}
	if !warned {
		t.Fatalf("no form-unavailable warning: %v", result.Warnings)
	}
	marker := readManagedMarker(t, fx, "opencode")
	if surface := marker.Surfaces["root-context"]; surface.Form != "monolithic" {
		t.Fatalf("fallback form %q", surface.Form)
	}
	if payload, _ := os.ReadFile(filepath.Join(homeDir, "opencode.json")); string(payload) != string(unmanaged) {
		t.Fatal("the unmanaged file was edited")
	}
}

// TestRepairFailedMissingStore proves repair fails with
// environment_repair_failed — not stale, not silent — when the store
// cannot restore the home.
func TestRepairFailedMissingStore(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	homeDir := ManagedHomeDir(fx.home, "acme", "codex_cli")
	_ = os.Remove(filepath.Join(homeDir, "AGENTS.md"))
	lock, _, err := readLock(fx.home, "acme")
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range lock.Members {
		if member.Kind == "context" {
			_ = os.RemoveAll(contextstore.EntryDir(fx.home, "context", member.Name, member.PinKey()))
		}
	}
	req := fx.request("codex_cli")
	req.Repair = true
	_, err = Resolve(req)
	if err == nil || !strings.Contains(err.Error(), DiagRepairFailed) {
		t.Fatalf("an unrestorable home must fail with %s, got %v", DiagRepairFailed, err)
	}
}

// TestOpencodeXDGSeeds proves the XDG seed class: allowlisted operator
// entries seed as parent links recorded in the marker, and an unrecorded
// entry shadowing an allowlisted one warns and is never touched.
func TestOpencodeXDGSeeds(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	if err := os.MkdirAll(filepath.Join(fx.xdg, "git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fx.xdg, "git", "config"), []byte("[user]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	req := fx.request("opencode")
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	marker := readManagedMarker(t, fx, "opencode")
	if len(marker.SeedLinks) != 1 || marker.SeedLinks[0] != "git" {
		t.Fatalf("seed links %v", marker.SeedLinks)
	}
	parent := ManagedParent(fx.home, "acme", "opencode")
	if target, err := os.Readlink(filepath.Join(parent, "git")); err != nil || target != filepath.Join(fx.xdg, "git") {
		t.Fatalf("xdg seed targets %q (%v)", target, err)
	}
	// A shadow: an unrecorded parent entry over an allowlisted operator
	// entry warns and is left as it is.
	if err := os.MkdirAll(filepath.Join(fx.xdg, "gh"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "gh"), []byte("operator file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bare, err := Resolve(fx.request("opencode"))
	if err != nil {
		t.Fatalf("a shadow warns, never blocks: %v", err)
	}
	shadowed := false
	for _, warning := range bare.Warnings {
		if strings.Contains(warning, envregistry.DiagSeedShadowed) {
			shadowed = true
		}
	}
	if !shadowed {
		t.Fatalf("no seed-shadowed warning: %v", bare.Warnings)
	}
	if payload, _ := os.ReadFile(filepath.Join(parent, "gh")); string(payload) != "operator file\n" {
		t.Fatal("the shadowing entry was touched")
	}
}

// TestClaudeSeedMergePreservesToolState proves the .claude.json seed is a
// merge: tool-written members survive repair, and the launch entry gains
// the trust keys.
func TestClaudeSeedMergePreservesToolState(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	machine := envregistry.DefaultMachineConfig()
	machine.Forms = map[string]string{"claude_code": "referenced"}
	provision(t, fx, "claude_code", machine)
	homeDir := ManagedHomeDir(fx.home, "acme", "claude_code")
	seeded, err := os.ReadFile(filepath.Join(homeDir, ".claude.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(seeded), `"hasCompletedOnboarding":true`) {
		t.Fatalf("seed shape: %s", seeded)
	}
	// The tool writes its own members around the seed.
	toolState := strings.Replace(string(seeded), `"projects":`, `"firstRun":true,"projects":`, 1)
	if err := os.WriteFile(filepath.Join(homeDir, ".claude.json"), []byte(toolState), 0o644); err != nil {
		t.Fatal(err)
	}
	// Force a repair through another surface and prove the merge kept the
	// tool's members.
	_ = os.Remove(filepath.Join(homeDir, "CLAUDE.md"))
	req := fx.request("claude_code")
	req.Machine = machine
	req.Repair = true
	if _, err := Resolve(req); err != nil {
		t.Fatal(err)
	}
	merged, err := os.ReadFile(filepath.Join(homeDir, ".claude.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(merged), `"firstRun":true`) {
		t.Fatalf("repair rewrote tool state: %s", merged)
	}
	if !strings.Contains(string(merged), `"hasClaudeMdExternalIncludesApproved":true`) {
		t.Fatalf("the launch entry lacks the includes key: %s", merged)
	}
}

// TestStatusShellHookTrustFromLaunchProject proves the §8.6 posture rows
// of StatusOf: the launch directory's project contributes its env files,
// recorded paths are always known, and malformed state lines warn without
// breaking the valid rows. The NonCurrent verdict over these rows (changed
// fails, unapproved warns) runs through the same field the CLI --check
// tests prove end to end; this level pins the rows, where the launch
// directory is injectable.
func TestStatusShellHookTrustFromLaunchProject(t *testing.T) {
	home := t.TempDir()
	project := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(filepath.Join(project, ".agents"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "Skillfile.json"), []byte(`{"schema_version":1,"skills":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(project, ".agents", "env.sh")
	if err := os.WriteFile(envPath, []byte("export CURATOR_PROJECT_ENV=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(envPath)
	if err != nil {
		t.Fatal(err)
	}
	request := StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), LaunchDir: project}

	status, err := StatusOf(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrust) != 1 {
		t.Fatalf("trust rows = %+v, want the unapproved project file", status.ShellHookTrust)
	}
	row := status.ShellHookTrust[0]
	if row.Path != resolved || row.State != hookapproval.PostureUnapproved || row.Diagnostic != hookapproval.DiagnosticEnvUnapproved {
		t.Fatalf("trust row = %+v", row)
	}

	if _, err := hookapproval.ApproveFile(home, envPath, hookapproval.ApprovedByOperator, parseTrustStamp(t, "2026-09-17T00:00:00Z")); err != nil {
		t.Fatal(err)
	}
	status, err = StatusOf(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrust) != 1 {
		t.Fatalf("trust rows = %+v", status.ShellHookTrust)
	}
	row = status.ShellHookTrust[0]
	if row.Path != resolved || row.State != hookapproval.PostureApproved || row.ApprovedBy != hookapproval.ApprovedByOperator {
		t.Fatalf("trust row = %+v", row)
	}

	if err := os.WriteFile(envPath, []byte("export CURATOR_PROJECT_ENV=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err = StatusOf(request)
	if err != nil {
		t.Fatal(err)
	}
	row = status.ShellHookTrust[0]
	if row.State != hookapproval.PostureChanged || row.Diagnostic != hookapproval.DiagnosticEnvChanged || row.ApprovedBy != hookapproval.ApprovedByOperator {
		t.Fatalf("trust row = %+v", row)
	}

	// Outside any project the recorded path stays known.
	status, err = StatusOf(StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), LaunchDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrust) != 1 || status.ShellHookTrust[0].Path != resolved {
		t.Fatalf("trust rows outside a project = %+v, want the recorded path", status.ShellHookTrust)
	}

	// A malformed line warns and is skipped; the valid row survives.
	state, err := os.ReadFile(hookapproval.ApprovalsPath(home))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hookapproval.ApprovalsPath(home), append(state, []byte("malformed-line\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err = StatusOf(request)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrustWarnings) != 1 || !strings.Contains(status.ShellHookTrustWarnings[0], "line 2") {
		t.Fatalf("trust warnings = %v, want the malformed line", status.ShellHookTrustWarnings)
	}
	if len(status.ShellHookTrust) != 1 || status.ShellHookTrust[0].State != hookapproval.PostureChanged {
		t.Fatalf("trust rows = %+v, want the valid changed row", status.ShellHookTrust)
	}
}

func parseTrustStamp(t *testing.T, value string) time.Time {
	t.Helper()
	stamp, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return stamp
}

// TestStatusShellHookTrustMissingUnreadableAndStateFailures proves the
// R1–R3 rows of StatusOf with the launch directory outside any project: a
// recorded file that disappears or becomes unreadable keeps its row and
// record, an unreadable approval state warns, and a truly absent state
// stays the quiet case. This level pins the rows and each row's own
// NonCurrent verdict; the matrix-level NonCurrent flag over these rows
// (and the CLI --check exits) runs through an otherwise-current matrix in
// the cmd/curator tests, where the exit code is evidence of the trust
// posture alone.
func TestStatusShellHookTrustMissingUnreadableAndStateFailures(t *testing.T) {
	home := t.TempDir()
	envPath := filepath.Join(t.TempDir(), "project", ".agents", "env.sh")
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(envPath, []byte("export CURATOR_PROJECT_ENV=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hookapproval.ApproveFile(home, envPath, hookapproval.ApprovedByOperator, parseTrustStamp(t, "2026-09-17T00:00:00Z")); err != nil {
		t.Fatal(err)
	}
	outside := StatusRequest{Home: home, Machine: envregistry.DefaultMachineConfig(), LaunchDir: t.TempDir()}

	status, err := StatusOf(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrust) != 1 || status.ShellHookTrust[0].State != hookapproval.PostureApproved {
		t.Fatalf("trust rows = %+v, want the approved recorded path", status.ShellHookTrust)
	}
	if status.ShellHookTrust[0].NonCurrent() {
		t.Fatal("an approved recorded path is current")
	}

	// R1: the approved file disappears; its row stays and is non-current.
	if err := os.Remove(envPath); err != nil {
		t.Fatal(err)
	}
	status, err = StatusOf(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrust) != 1 {
		t.Fatalf("trust rows = %+v, want the missing recorded path", status.ShellHookTrust)
	}
	row := status.ShellHookTrust[0]
	if row.Path != resolved || row.State != hookapproval.PostureApproved ||
		row.ApprovedBy != hookapproval.ApprovedByOperator || row.File != hookapproval.PostureFileMissing {
		t.Fatalf("trust row = %+v", row)
	}
	if !row.NonCurrent() {
		t.Fatal("a recorded-but-missing file is non-current")
	}

	// R2: the approved file becomes unreadable-as-bytes; its record stays.
	if err := os.Mkdir(envPath, 0o700); err != nil {
		t.Fatal(err)
	}
	status, err = StatusOf(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrust) != 1 {
		t.Fatalf("trust rows = %+v, want the unreadable recorded path", status.ShellHookTrust)
	}
	row = status.ShellHookTrust[0]
	if row.Path != resolved || row.ApprovedBy != hookapproval.ApprovedByOperator || row.File != hookapproval.PostureFileUnreadable {
		t.Fatalf("trust row = %+v", row)
	}
	if !row.NonCurrent() {
		t.Fatal("an unreadable recorded file is non-current")
	}

	// R3: the approval state itself becomes unreadable; the failure warns
	// and is non-current, never an empty set.
	if err := os.Remove(hookapproval.ApprovalsPath(home)); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(hookapproval.ApprovalsPath(home), 0o755); err != nil {
		t.Fatal(err)
	}
	status, err = StatusOf(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrustWarnings) != 1 || !strings.Contains(status.ShellHookTrustWarnings[0], "cannot read") {
		t.Fatalf("trust warnings = %v, want the state read failure", status.ShellHookTrustWarnings)
	}
	// The matrix-level NonCurrent flag for this failure is proven at the
	// CLI level, where the matrix is otherwise current.

	// Paired control: a truly absent state carries no rows and no
	// warnings when no project candidate exists.
	if err := os.Remove(hookapproval.ApprovalsPath(home)); err != nil {
		t.Fatal(err)
	}
	status, err = StatusOf(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.ShellHookTrust) != 0 || len(status.ShellHookTrustWarnings) != 0 {
		t.Fatalf("trust rows = %+v warnings = %v, want none", status.ShellHookTrust, status.ShellHookTrustWarnings)
	}
}
