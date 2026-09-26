// Production entry points under test: PlanMigration and ApplyMigration
// for the Decision 0017 explicit credential migration (environments
// §7.4, §10.1; manager §12.4). Every row drives a production entry on a
// temporary store; helper-direct assertions are bounds, not rows.
package envprofile

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/envmarker"
	"github.com/relux-works/curator/internal/envregistry"
)

func (fx *managedFixture) migrateRequest() MigrateRequest {
	return MigrateRequest{
		Home:    fx.home,
		Machine: envregistry.DefaultMachineConfig(),
		Detect:  func(envregistry.Adapter) string { return "unknown" },
		NativeHomeOf: func(id string) (string, error) {
			if id == "opencode" {
				return fx.native["opencode-native"], nil
			}
			return fx.native[id], nil
		},
	}
}

func findMigrateHome(report *MigrateReport, profile, env string) *MigrateHome {
	for i := range report.Homes {
		if report.Homes[i].Profile == profile && report.Homes[i].EnvID == env {
			return &report.Homes[i]
		}
	}
	return nil
}

// fileSnapshot records every non-directory entry below the credential
// scope: regular files by sha256, symlinks by target. The manager
// state directory (locks, journals) is excluded: the mutation lock and
// the journal legitimately write there on every mutating operation.
type fileSnapshot map[string]string

func snapshotCredentialScope(t *testing.T, fx *managedFixture) fileSnapshot {
	t.Helper()
	snap := fileSnapshot{}
	roots := []string{filepath.Join(fx.home, "environments"), filepath.Join(fx.home, "profiles")}
	for _, dir := range fx.native {
		roots = append(roots, dir)
	}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				target, err := os.Readlink(path)
				if err != nil {
					return err
				}
				snap[path] = "link->" + target
				return nil
			}
			payload, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(payload)
			snap[path] = "file:" + hex.EncodeToString(sum[:])
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	return snap
}

// snapshotDiff partitions the before/after snapshots into added,
// removed, and changed paths.
func snapshotDiff(before, after fileSnapshot) (added, removed, changed []string) {
	for path, digest := range after {
		prior, ok := before[path]
		if !ok {
			added = append(added, path)
		} else if prior != digest {
			changed = append(changed, path)
		}
	}
	for path := range before {
		if _, ok := after[path]; !ok {
			removed = append(removed, path)
		}
	}
	return added, removed, changed
}

// credentialHolders returns the regular files under the ENTIRE
// temporary manager home plus every native root whose bytes contain
// the sentinel. The home walk has no path allowlist: environments,
// profiles, state, the journal, backups, and temp paths are all
// covered, so a copy planted anywhere the migration writes is found.
// Allowance: none — lock/journal metadata holds only hashes, paths,
// and digests, never credential content, so any holder below the home
// is a secret copy and fails the row. Native roots keep their
// expected holders (the live credential files themselves).
func credentialHolders(t *testing.T, fx *managedFixture, sentinel string) map[string]bool {
	t.Helper()
	holders := map[string]bool{}
	roots := []string{fx.home}
	for _, dir := range fx.native {
		roots = append(roots, dir)
	}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			payload, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(payload), sentinel) {
				holders[path] = true
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	return holders
}

func sha256Of(t *testing.T, path string) string {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// TestMigratePiWrongTargetToAgentRoot is the operator-case row: a Pi
// home linked at the pre-0017 native root migrates to ~/.pi/agent with
// bytes preserved and mode intact, printing the plan before applying.
// The plan is deterministic (same hash and bytes across runs); apply
// relinks (a symlink, never a copy) and rewrites no marker.
func TestMigratePiWrongTargetToAgentRoot(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	native := fx.native["pi"]
	agentAuth := filepath.Join(native, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	credential := []byte("{\"t\":\"operator-pi-credential\"}\n")
	if err := os.WriteFile(agentAuth, credential, 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(native, "auth.json")
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	markerPath := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), envmarker.Name)
	markerBefore, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	wantDigest := sha256Of(t, agentAuth)

	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	inspect := report.Text(MigratePhaseInspect)
	for _, want := range []string{"mis-targeted", oldTarget, agentAuth, "effective isolation shared", "native Pi roots", "absent", "present"} {
		if !strings.Contains(inspect, want) {
			t.Fatalf("inspect inventories the mis-targeted link and both Pi roots, want %q in:\n%s", want, inspect)
		}
	}
	home := findMigrateHome(report, "acme", "pi")
	if home == nil || home.Isolation != envregistry.IsolationShared || home.Mode != envmarker.ModeManagedHome {
		t.Fatalf("the Pi home inventories shared/managed-home: %+v", home)
	}
	if home.PiOld != MigrateRootAbsent || home.PiNew != MigrateRootPresent {
		t.Fatalf("Pi roots inventoried old=%s new=%s", home.PiOld, home.PiNew)
	}
	if len(report.Ops()) != 1 || len(report.Conflicts()) != 0 {
		t.Fatalf("one relink op, no conflicts: ops=%+v conflicts=%+v", report.Ops(), report.Conflicts())
	}
	op := report.Ops()[0]
	if op.Kind != MigrateOpRelink || op.Path != "auth.json" || op.From != oldTarget || op.To != agentAuth {
		t.Fatalf("relink op %+v", op)
	}
	plan := report.Text(MigratePhasePlan)
	matched, err := regexp.MatchString(`credential migration plan [0-9a-f]{64}`, plan)
	if err != nil || !matched {
		t.Fatalf("plan prints its hash:\n%s", plan)
	}
	for _, want := range []string{"relink auth.json", oldTarget, agentAuth, "--expect " + report.Hash} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan prints the exact operation, want %q in:\n%s", want, plan)
		}
	}
	again, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if again.Hash != report.Hash || again.Text(MigratePhasePlan) != plan {
		t.Fatal("the plan is deterministic: same hash and bytes across runs")
	}

	before := snapshotCredentialScope(t, fx)
	req := fx.migrateRequest()
	req.Expect = report.Hash
	result, err := ApplyMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Applied) != 1 || result.Applied[0] != op {
		t.Fatalf("apply executes exactly the printed plan: %+v", result.Applied)
	}
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the migrated link is a symlink, never a copy: %v", err)
	}
	if target, err := os.Readlink(link); err != nil || target != agentAuth {
		t.Fatalf("the link now targets the agent root: %q (%v)", target, err)
	}
	if digest := sha256Of(t, agentAuth); digest != wantDigest {
		t.Fatal("native credential bytes are byte-identical after migration")
	}
	if _, err := os.Lstat(oldTarget); !os.IsNotExist(err) {
		t.Fatalf("migration creates nothing at the old target: %v", err)
	}
	if payload, err := os.ReadFile(markerPath); err != nil || string(payload) != string(markerBefore) {
		t.Fatal("a relink rewrites no marker bytes")
	}
	after, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	healed := findMigrateHome(after, "acme", "pi")
	if healed == nil || healed.Isolation != envregistry.IsolationShared || healed.Mode != envmarker.ModeManagedHome {
		t.Fatalf("the effective mode is preserved on upgrade: %+v", healed)
	}
	if len(after.Ops()) != 0 || len(after.Conflicts()) != 0 {
		t.Fatalf("the migrated home plans nothing: %+v %+v", after.Ops(), after.Conflicts())
	}
	if _, err := Resolve(fx.request("pi")); err != nil {
		t.Fatalf("the migrated home is current: %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 {
		t.Fatalf("migration adds and removes no files: added=%v removed=%v", added, removed)
	}
	if len(changed) != 1 || changed[0] != link {
		t.Fatalf("only the link target changes: %v", changed)
	}
}

// TestMigrateNoSecretCopies is the no-copy row: across a relink and an
// unlink in one apply, native credential bytes stay byte-identical, no
// new file appears anywhere in scope, and no file newly contains
// credential content — the only changed bytes are the link itself and
// the marker record that drops the unlinked entry. The content scan
// covers the ENTIRE manager home (state, journal, backups, temps — no
// path allowlist) plus every native root: any copy the migration
// plants anywhere fails the row. A second phase crashes mid-apply and
// scans while the journal stands, so transient journal content is
// covered too (a deleted journal cannot be inspected after success).
func TestMigrateNoSecretCopies(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	piNative := fx.native["pi"]
	agentAuth := filepath.Join(piNative, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	piCredential := "CRED-SENTINEL-PI-7f3a\n"
	if err := os.WriteFile(agentAuth, []byte(piCredential), 0o600); err != nil {
		t.Fatal(err)
	}
	codexAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
	codexCredential := "CRED-SENTINEL-CODEX-9e1b\n"
	if err := os.WriteFile(codexAuth, []byte(codexCredential), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	piLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(piNative, "auth.json")
	_ = os.Remove(piLink)
	if err := os.Symlink(oldTarget, piLink); err != nil {
		t.Fatal(err)
	}
	codexLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	codexMarker := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)
	isolated := envregistry.DefaultMachineConfig()
	isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}

	wantPi := sha256Of(t, agentAuth)
	wantCodex := sha256Of(t, codexAuth)
	before := snapshotCredentialScope(t, fx)
	piHoldersBefore := credentialHolders(t, fx, piCredential)
	codexHoldersBefore := credentialHolders(t, fx, codexCredential)
	if len(piHoldersBefore) != 1 || !piHoldersBefore[agentAuth] {
		t.Fatalf("the Pi credential lives in exactly its native file: %v", piHoldersBefore)
	}
	if len(codexHoldersBefore) != 1 || !codexHoldersBefore[codexAuth] {
		t.Fatalf("the codex credential lives in exactly its native file: %v", codexHoldersBefore)
	}

	req := fx.migrateRequest()
	req.Machine = isolated
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 2 || len(report.Conflicts()) != 0 {
		t.Fatalf("one relink plus one unlink, no conflicts: %+v %+v", report.Ops(), report.Conflicts())
	}
	req.Expect = report.Hash
	result, err := ApplyMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Applied) != 2 {
		t.Fatalf("both operations applied: %+v", result.Applied)
	}
	if digest := sha256Of(t, agentAuth); digest != wantPi {
		t.Fatal("Pi native bytes are byte-identical after migration")
	}
	if digest := sha256Of(t, codexAuth); digest != wantCodex {
		t.Fatal("codex native bytes are byte-identical after migration")
	}
	if _, err := os.Lstat(codexLink); !os.IsNotExist(err) {
		t.Fatalf("the stale link is unlinked: %v", err)
	}
	marker := readManagedMarker(t, fx, "codex_cli")
	if marker.Passthrough == nil || len(*marker.Passthrough) != 0 {
		t.Fatalf("the record drops the unlinked entry: %+v", marker.Passthrough)
	}
	if marker.Mode != envmarker.ModeManagedHome {
		t.Fatalf("the marker mode is intact: %s", marker.Mode)
	}
	piHoldersAfter := credentialHolders(t, fx, piCredential)
	codexHoldersAfter := credentialHolders(t, fx, codexCredential)
	if len(piHoldersAfter) != 1 || !piHoldersAfter[agentAuth] {
		t.Fatalf("no new file carries Pi credential content: %v", piHoldersAfter)
	}
	if len(codexHoldersAfter) != 1 || !codexHoldersAfter[codexAuth] {
		t.Fatalf("no new file carries codex credential content: %v", codexHoldersAfter)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 {
		t.Fatalf("migration adds no files: %v", added)
	}
	if len(removed) != 1 || removed[0] != codexLink {
		t.Fatalf("only the stale link is removed: %v", removed)
	}
	changedSet := map[string]bool{}
	for _, path := range changed {
		changedSet[path] = true
	}
	if len(changed) != 2 || !changedSet[piLink] || !changedSet[codexMarker] {
		t.Fatalf("only the relinked link and the codex marker change: %v", changed)
	}
	// The migrated homes are current under their preserved modes.
	if _, err := Resolve(fx.request("pi")); err != nil {
		t.Fatalf("the migrated Pi home is current: %v", err)
	}
	check := fx.request("codex_cli")
	check.Machine = isolated
	if _, err := Resolve(check); err != nil {
		t.Fatalf("the migrated codex home is current: %v", err)
	}
	after, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if got := findMigrateHome(after, "acme", "pi"); got == nil || got.Isolation != envregistry.IsolationShared {
		t.Fatalf("Pi stays shared: %+v", got)
	}
	if got := findMigrateHome(after, "acme", "codex_cli"); got == nil || got.Isolation != envregistry.IsolationIsolated {
		t.Fatalf("codex stays isolated: %+v", got)
	}
	// Interrupted-apply transient: the journal is deleted on success,
	// so prove it never carries credential bytes while it stands. A
	// fresh fixture crashes after the first operation through the F-C2
	// crash-injection point; the whole-home scan (which includes the
	// standing journal, state, and temps) plus an explicit journal
	// read must find no credential content, and the recovered apply
	// must converge clean as well.
	t.Run("interrupted journal holds no credential bytes", func(t *testing.T) {
		gx := writeManagedFixture(t, "acme")
		gxPiNative := gx.native["pi"]
		gxAgentAuth := filepath.Join(gxPiNative, "agent", "auth.json")
		if err := os.MkdirAll(filepath.Dir(gxAgentAuth), 0o755); err != nil {
			t.Fatal(err)
		}
		gxPiCredential := "CRED-SENTINEL-PI-7f3a\n"
		if err := os.WriteFile(gxAgentAuth, []byte(gxPiCredential), 0o600); err != nil {
			t.Fatal(err)
		}
		gxCodexAuth := filepath.Join(gx.native["codex_cli"], "auth.json")
		gxCodexCredential := "CRED-SENTINEL-CODEX-9e1b\n"
		if err := os.WriteFile(gxCodexAuth, []byte(gxCodexCredential), 0o600); err != nil {
			t.Fatal(err)
		}
		provision(t, gx, "pi", envregistry.DefaultMachineConfig())
		provision(t, gx, "codex_cli", envregistry.DefaultMachineConfig())
		gxPiLink := filepath.Join(ManagedHomeDir(gx.home, "acme", "pi"), "auth.json")
		gxOldTarget := filepath.Join(gxPiNative, "auth.json")
		_ = os.Remove(gxPiLink)
		if err := os.Symlink(gxOldTarget, gxPiLink); err != nil {
			t.Fatal(err)
		}
		gxCodexLink := filepath.Join(ManagedHomeDir(gx.home, "acme", "codex_cli"), "auth.json")
		gxIsolated := envregistry.DefaultMachineConfig()
		gxIsolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
		gxReq := gx.migrateRequest()
		gxReq.Machine = gxIsolated
		gxReport, err := PlanMigration(gxReq)
		if err != nil {
			t.Fatal(err)
		}
		if len(gxReport.Ops()) != 2 {
			t.Fatalf("one relink plus one unlink: %+v", gxReport.Ops())
		}
		gxReq.Expect = gxReport.Hash
		assertNoCopy := func(t *testing.T, why string) {
			t.Helper()
			piHolders := credentialHolders(t, gx, gxPiCredential)
			if len(piHolders) != 1 || !piHolders[gxAgentAuth] {
				t.Fatalf("%s: no file may carry Pi credential content: %v", why, piHolders)
			}
			codexHolders := credentialHolders(t, gx, gxCodexCredential)
			if len(codexHolders) != 1 || !codexHolders[gxCodexAuth] {
				t.Fatalf("%s: no file may carry codex credential content: %v", why, codexHolders)
			}
		}
		assertNoCopy(t, "before the crash")
		crash := gxReq
		crash.InjectFault = func(point string, applied int) error {
			if point == MigrateFaultCrash && applied >= 1 {
				return errors.New("kill -9")
			}
			return nil
		}
		interrupted, err := ApplyMigration(crash)
		if err == nil || !strings.Contains(err.Error(), "interrupted") {
			t.Fatalf("the crash must interrupt, got %v", err)
		}
		if interrupted == nil || len(interrupted.Applied) != 1 {
			t.Fatalf("one operation stands at the kill: %+v", interrupted)
		}
		journal, err := readMigrationJournal(gx.home)
		if err != nil || journal == nil {
			t.Fatalf("the journal stands after the kill: %+v (%v)", journal, err)
		}
		journalBytes, err := os.ReadFile(migrationJournalPath(gx.home))
		if err != nil {
			t.Fatal(err)
		}
		for _, sentinel := range []string{gxPiCredential, gxCodexCredential} {
			if strings.Contains(string(journalBytes), sentinel) {
				t.Fatalf("the standing journal holds credential bytes: %q", sentinel)
			}
		}
		assertNoCopy(t, "while the journal stands")
		retry := gxReq
		retry.InjectFault = nil
		result, err := ApplyMigration(retry)
		if err != nil {
			t.Fatal(err)
		}
		if result.Recovered == nil || result.Recovered.Plan != gxReport.Hash {
			t.Fatalf("the apply reports the recovery: %+v", result.Recovered)
		}
		if len(result.Applied) != 2 {
			t.Fatalf("the recovered apply executes the whole plan: %+v", result.Applied)
		}
		if _, err := os.Lstat(migrationJournalPath(gx.home)); !os.IsNotExist(err) {
			t.Fatalf("the recovered apply deletes the journal: %v", err)
		}
		if _, err := os.Lstat(gxCodexLink); !os.IsNotExist(err) {
			t.Fatalf("the codex link is unlinked: %v", err)
		}
		if target, err := os.Readlink(gxPiLink); err != nil || target != gxAgentAuth {
			t.Fatalf("the Pi link migrated: %q (%v)", target, err)
		}
		assertNoCopy(t, "after recovery converges")
	})
}

// TestMigrateConflictsRefuse drives every migration refusal through the
// production entries: the plan lists the conflict with the exact
// operator choice, apply refuses with environment_credential_conflict
// naming it, and zero bytes change anywhere in scope. The empty-file
// subtest is the narrowing killer for a mutant that refuses only
// non-empty files.
func TestMigrateConflictsRefuse(t *testing.T) {
	requireLinkCapability(t)
	for _, tc := range []struct {
		name       string
		bytes      []byte
		wantDetail []string
		wantChoice []string
	}{
		{"regular file non-empty", []byte("managed-credential\n"),
			[]string{"holds a regular file", "auth.json"},
			[]string{"decide which credential bytes win", "move the loser aside"}},
		{"regular file empty", []byte{},
			[]string{"holds a regular file", "auth.json"},
			[]string{"decide which credential bytes win", "move the loser aside"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := writeManagedFixture(t, "acme")
			nativeAuth := []byte("{\"t\":\"operator-credential\"}\n")
			if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), nativeAuth, 0o600); err != nil {
				t.Fatal(err)
			}
			provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
			link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
			_ = os.Remove(link)
			if err := os.WriteFile(link, tc.bytes, 0o600); err != nil {
				t.Fatal(err)
			}
			assertMigrateConflict(t, fx, fx.migrateRequest(), link, tc.wantDetail, tc.wantChoice)
			if payload, err := os.ReadFile(link); err != nil || string(payload) != string(tc.bytes) {
				t.Fatalf("managed bytes untouched: %q (%v)", payload, err)
			}
			if payload, err := os.ReadFile(filepath.Join(fx.native["codex_cli"], "auth.json")); err != nil || string(payload) != string(nativeAuth) {
				t.Fatalf("native bytes untouched: %q (%v)", payload, err)
			}
		})
	}
	t.Run("foreign symlink", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte("cli_auth_credentials_store = \"keyring\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte("cli_auth_credentials_store = \"file\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
		foreign := filepath.Join(t.TempDir(), "auth.json")
		if err := os.WriteFile(foreign, []byte("foreign\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(foreign, link); err != nil {
			t.Fatal(err)
		}
		assertMigrateConflict(t, fx, fx.migrateRequest(), link,
			[]string{"links to", foreign}, []string{"remove the unrecorded link out of band"})
		if target, err := os.Readlink(link); err != nil || target != foreign {
			t.Fatalf("the foreign link is untouched: %q (%v)", target, err)
		}
	})
	t.Run("stale foreign symlink", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "auth.json"), []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
		foreign := filepath.Join(t.TempDir(), "auth.json")
		if err := os.WriteFile(foreign, []byte("foreign\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		_ = os.Remove(link)
		if err := os.Symlink(foreign, link); err != nil {
			t.Fatal(err)
		}
		isolated := envregistry.DefaultMachineConfig()
		isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
		req := fx.migrateRequest()
		req.Machine = isolated
		assertMigrateConflict(t, fx, req, link,
			[]string{"not the recorded credential target", foreign}, []string{"remove the link out of band"})
		if target, err := os.Readlink(link); err != nil || target != foreign {
			t.Fatalf("the foreign link is untouched: %q (%v)", target, err)
		}
	})
	t.Run("two live pi credentials", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		native := fx.native["pi"]
		agentAuth := filepath.Join(native, "agent", "auth.json")
		if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
			t.Fatal(err)
		}
		oldAuth := filepath.Join(native, "auth.json")
		if err := os.WriteFile(agentAuth, []byte("{\"t\":\"agent-credential\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(oldAuth, []byte("{\"t\":\"old-credential\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		provision(t, fx, "pi", envregistry.DefaultMachineConfig())
		link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
		_ = os.Remove(link)
		if err := os.Symlink(oldAuth, link); err != nil {
			t.Fatal(err)
		}
		assertMigrateConflict(t, fx, fx.migrateRequest(), "",
			[]string{"two live Pi credentials", oldAuth, agentAuth},
			[]string{"decide which of", oldAuth, agentAuth, "never copies, moves, or deletes"})
		if target, err := os.Readlink(link); err != nil || target != oldAuth {
			t.Fatalf("the link is untouched: %q (%v)", target, err)
		}
		if payload, err := os.ReadFile(oldAuth); err != nil || string(payload) != "{\"t\":\"old-credential\"}\n" {
			t.Fatalf("old-root bytes untouched: %q (%v)", payload, err)
		}
		if payload, err := os.ReadFile(agentAuth); err != nil || string(payload) != "{\"t\":\"agent-credential\"}\n" {
			t.Fatalf("agent-root bytes untouched: %q (%v)", payload, err)
		}
	})
	t.Run("non-empty directory", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		seedLiveNativeCredentials(t, fx)
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		link := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
		_ = os.Remove(link)
		if err := os.Mkdir(link, 0o755); err != nil {
			t.Fatal(err)
		}
		kept := filepath.Join(link, "kept")
		if err := os.WriteFile(kept, []byte("kept\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		assertMigrateConflict(t, fx, fx.migrateRequest(), link,
			[]string{"non-empty directory"}, []string{"clear the directory out of band"})
		if payload, err := os.ReadFile(kept); err != nil || string(payload) != "kept\n" {
			t.Fatalf("directory contents untouched: %q (%v)", payload, err)
		}
	})
	t.Run("invalid marker", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		seedLiveNativeCredentials(t, fx)
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		markerPath := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), envmarker.Name)
		if err := os.WriteFile(markerPath, []byte("{not a marker"), 0o644); err != nil {
			t.Fatal(err)
		}
		assertMigrateConflict(t, fx, fx.migrateRequest(), "",
			[]string{"marker is invalid"}, []string{"back the marker up"})
	})
	t.Run("unknown credential store", func(t *testing.T) {
		fx := writeManagedFixture(t, "acme")
		seedLiveNativeCredentials(t, fx)
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		if err := os.WriteFile(filepath.Join(fx.native["codex_cli"], "config.toml"), []byte("cli_auth_credentials_store = \"ephemeral\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		assertMigrateConflict(t, fx, fx.migrateRequest(), "",
			[]string{"cannot establish the credential strategy"}, []string{"fix the native credential configuration out of band"})
	})
}

// assertMigrateConflict proves one migration refusal end to end: the
// plan lists the conflict with its operator choice and prints blocked,
// apply refuses with environment_credential_conflict naming them, and
// the filesystem is byte-identical before and after both entries.
func assertMigrateConflict(t *testing.T, fx *managedFixture, req MigrateRequest, link string, wantDetail, wantChoice []string) {
	t.Helper()
	before := snapshotCredentialScope(t, fx)
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Conflicts()) == 0 {
		t.Fatalf("the plan lists the conflict: %+v", report.Ops())
	}
	plan := report.Text(MigratePhasePlan)
	if !strings.Contains(plan, "blocked:") {
		t.Fatalf("a conflicted plan prints blocked:\n%s", plan)
	}
	for _, want := range append(append([]string{}, wantDetail...), wantChoice...) {
		if !strings.Contains(plan, want) {
			t.Fatalf("the plan names %q in:\n%s", want, plan)
		}
	}
	if link != "" && !strings.Contains(plan, link) {
		t.Fatalf("the plan names the path %s in:\n%s", link, plan)
	}
	req.Expect = report.Hash
	result, err := ApplyMigration(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
		t.Fatalf("apply must refuse with %s, got %v", envregistry.DiagCredentialConflict, err)
	}
	if result == nil || result.Report == nil {
		t.Fatal("a refused apply still returns the printed plan")
	}
	for _, want := range append(append([]string{}, wantDetail...), wantChoice...) {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the refusal names %q: %v", want, err)
		}
	}
	if len(result.Applied) != 0 {
		t.Fatalf("a refused apply executes nothing: %+v", result.Applied)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("a refused apply writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
}

// TestMigratePiRootInspectionFailure pins the absent-versus-
// uninspectable distinction for the Pi roots inventory: a re-point
// planned while a native root cannot be statted is a conflict carrying
// the inspection diagnostic — never silence, and never the absence
// wording. A regular file where the agent directory should be fails the
// target stat deterministically on Unix, even as root; Windows maps a
// stat through a file to path-not-found, so the row is Unix-only.
func TestMigratePiRootInspectionFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this host cannot create an uninspectable Pi root: Windows maps a stat through a file to path-not-found, so the inspection diagnostic runs on the unix runners")
	}
	fx := writeManagedFixture(t, "acme")
	native := fx.native["pi"]
	agentAuth := filepath.Join(native, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(native, "auth.json")
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	// Break the new root's parent: a file where the agent directory was.
	if err := os.RemoveAll(filepath.Join(native, "agent")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(native, "agent"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assertMigrateConflict(t, fx, fx.migrateRequest(), "",
		[]string{"cannot be inspected", agentAuth}, []string{"restore access to the native roots out of band"})
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if plan := report.Text(MigratePhasePlan); strings.Contains(plan, "does not exist") {
		t.Fatalf("an inspection failure is not absence:\n%s", plan)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the link is untouched: %q (%v)", target, err)
	}
}

// TestMigratePlanDriftRefuses proves apply refuses when the inventory
// changed since the plan: a stale --expect hash fails closed with zero
// writes, and a truncated hash is not accepted (full-hash narrowing
// killer). The missing-plan half lives in TestMigrateApplyRequiresPlan.
func TestMigratePlanDriftRefuses(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	native := fx.native["pi"]
	agentAuth := filepath.Join(native, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(native, "auth.json")
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	stale := report.Hash
	// The world moves after the plan: the link is fixed by hand.
	_ = os.Remove(link)
	if err := os.Symlink(agentAuth, link); err != nil {
		t.Fatal(err)
	}
	before := snapshotCredentialScope(t, fx)
	req := fx.migrateRequest()
	req.Expect = stale
	result, err := ApplyMigration(req)
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
		t.Fatalf("a drifted apply must refuse with %s, got %v", envregistry.DiagCredentialConflict, err)
	}
	for _, want := range []string{"plan drift", stale, "re-run"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the drift refusal names %q: %v", want, err)
		}
	}
	if result == nil || result.Report == nil || len(result.Applied) != 0 {
		t.Fatalf("a drifted apply executes nothing: %+v", result)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("a drifted apply writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
	// A truncated hash is not the plan hash: prefix matching would
	// admit it, so it must refuse too.
	fresh, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	prefix := fx.migrateRequest()
	prefix.Expect = fresh.Hash[:16]
	if _, err := ApplyMigration(prefix); err == nil || !strings.Contains(err.Error(), "plan drift") {
		t.Fatalf("a truncated --expect must refuse with plan drift, got %v", err)
	}
	// The hand-fixed home converges: a fresh full hash applies cleanly
	// with zero operations.
	clean := fx.migrateRequest()
	clean.Expect = fresh.Hash
	result, err = ApplyMigration(clean)
	if err != nil {
		t.Fatalf("the fresh hash applies on the converged plan: %v", err)
	}
	if len(result.Applied) != 0 {
		t.Fatalf("the hand-fixed home needs no operations: %+v", result.Applied)
	}
}

// TestMigrateRollbackRestoresPriorState injects apply failures and
// proves every outcome is the prior state, never a mix: links move
// back, prior markers restore byte-identical, native bytes are
// untouched. The link-phase fault fires before any marker publish;
// the post-publish fault fires after, exercising the marker rollback.
func TestMigrateRollbackRestoresPriorState(t *testing.T) {
	requireLinkCapability(t)
	setup := func(t *testing.T) (*managedFixture, MigrateRequest, map[string]string, map[string]string) {
		t.Helper()
		fx := writeManagedFixture(t, "acme")
		piNative := fx.native["pi"]
		agentAuth := filepath.Join(piNative, "agent", "auth.json")
		if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		codexAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
		if err := os.WriteFile(codexAuth, []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		provision(t, fx, "pi", envregistry.DefaultMachineConfig())
		provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
		piLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
		_ = os.Remove(piLink)
		if err := os.Symlink(filepath.Join(piNative, "auth.json"), piLink); err != nil {
			t.Fatal(err)
		}
		isolated := envregistry.DefaultMachineConfig()
		isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
		req := fx.migrateRequest()
		req.Machine = isolated
		report, err := PlanMigration(req)
		if err != nil {
			t.Fatal(err)
		}
		if len(report.Ops()) != 2 {
			t.Fatalf("the rollback plan carries two operations: %+v", report.Ops())
		}
		req.Expect = report.Hash
		markers := map[string]string{}
		for _, env := range []string{"pi", "codex_cli"} {
			payload, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, env), envmarker.Name))
			if err != nil {
				t.Fatal(err)
			}
			markers[env] = string(payload)
		}
		natives := map[string]string{
			agentAuth: sha256Of(t, agentAuth),
			codexAuth: sha256Of(t, codexAuth),
		}
		return fx, req, markers, natives
	}
	assertRolledBack := func(t *testing.T, fx *managedFixture, markers, natives map[string]string) {
		t.Helper()
		piLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
		if target, err := os.Readlink(piLink); err != nil || target != filepath.Join(fx.native["pi"], "auth.json") {
			t.Fatalf("the Pi link moves back to its prior target: %q (%v)", target, err)
		}
		codexLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
		if target, err := os.Readlink(codexLink); err != nil || target != filepath.Join(fx.native["codex_cli"], "auth.json") {
			t.Fatalf("the codex link is re-created at its prior target: %q (%v)", target, err)
		}
		for _, env := range []string{"pi", "codex_cli"} {
			payload, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, env), envmarker.Name))
			if err != nil || string(payload) != markers[env] {
				t.Fatalf("the prior %s marker restores byte-identical", env)
			}
		}
		for path, want := range natives {
			if digest := sha256Of(t, path); digest != want {
				t.Fatalf("native bytes untouched: %s", path)
			}
		}
		if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
			t.Fatalf("a clean rollback deletes the journal: %v", err)
		}
		for _, env := range []string{"pi", "codex_cli"} {
			strays, err := filepath.Glob(filepath.Join(ManagedHomeDir(fx.home, fx.profile, env), ".migrate-*.tmp"))
			if err != nil {
				t.Fatal(err)
			}
			if len(strays) != 0 {
				t.Fatalf("no temporary links survive rollback: %v", strays)
			}
		}
	}
	t.Run("link phase", func(t *testing.T) {
		fx, req, markers, natives := setup(t)
		req.InjectFault = func(point string, applied int) error {
			if point == MigrateFaultLink && applied >= 1 {
				return os.ErrInvalid
			}
			return nil
		}
		if _, err := ApplyMigration(req); err == nil || !strings.Contains(err.Error(), "rolled back") {
			t.Fatalf("the injected fault fails rolled back, got %v", err)
		}
		assertRolledBack(t, fx, markers, natives)
	})
	t.Run("post publish", func(t *testing.T) {
		fx, req, markers, natives := setup(t)
		req.InjectFault = func(point string, _ int) error {
			if point == MigrateFaultPublish {
				return os.ErrInvalid
			}
			return nil
		}
		if _, err := ApplyMigration(req); err == nil || !strings.Contains(err.Error(), "rolled back") {
			t.Fatalf("the injected fault fails rolled back, got %v", err)
		}
		assertRolledBack(t, fx, markers, natives)
	})
}

// TestMigrateOldRootBytesWarnAndProceed proves bytes at the old Pi root
// alone (agent root empty) warn loudly but do not block: the relink
// leaves the old bytes untouched, the manager never moves them, and the
// re-pointed link dangles pending until the operator moves the live
// bytes out of band. The narrowing killer for a mutant that refuses
// whenever the old root holds bytes.
func TestMigrateOldRootBytesWarnAndProceed(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	native := fx.native["pi"]
	if err := os.MkdirAll(filepath.Join(native, "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldAuth := filepath.Join(native, "auth.json")
	oldBytes := []byte("{\"t\":\"old-credential\"}\n")
	if err := os.WriteFile(oldAuth, oldBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	agentAuth := filepath.Join(native, "agent", "auth.json")
	_ = os.Remove(link)
	if err := os.Symlink(oldAuth, link); err != nil {
		t.Fatal(err)
	}
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 || len(report.Conflicts()) != 0 {
		t.Fatalf("one relink op, no conflicts: %+v %+v", report.Ops(), report.Conflicts())
	}
	plan := report.Text(MigratePhasePlan)
	for _, want := range []string{"native bytes exist at the old Pi root", oldAuth, "never copies"} {
		if !strings.Contains(plan, want) {
			t.Fatalf("the plan warns about the orphaned old bytes, want %q in:\n%s", want, plan)
		}
	}
	apply := fx.migrateRequest()
	apply.Expect = report.Hash
	if _, err := ApplyMigration(apply); err != nil {
		t.Fatalf("old-root bytes alone do not block: %v", err)
	}
	if target, err := os.Readlink(link); err != nil || target != agentAuth {
		t.Fatalf("the link targets the declared root: %q (%v)", target, err)
	}
	if payload, err := os.ReadFile(oldAuth); err != nil || string(payload) != string(oldBytes) {
		t.Fatalf("old-root bytes untouched: %q (%v)", payload, err)
	}
	current, err := Resolve(fx.request("pi"))
	if err != nil {
		t.Fatalf("the re-pointed home succeeds pending: %v", err)
	}
	if warnings := strings.Join(current.Warnings, "; "); !strings.Contains(warnings, "detached-pending") {
		t.Fatalf("the pending finding reports loudly: %q", warnings)
	}
}

// TestMigratePendingNeedsNothing proves dangling-to-declared entries
// get no operation: the plan is empty, apply succeeds applying nothing,
// and the pending link is untouched.
func TestMigratePendingNeedsNothing(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	want := filepath.Join(fx.native["pi"], "agent", "auth.json")
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	home := findMigrateHome(report, "acme", "pi")
	if home == nil || len(home.Ops) != 0 || len(home.Conflicts) != 0 {
		t.Fatalf("pending plans nothing: %+v", home)
	}
	found := false
	for _, entry := range home.Entries {
		if entry.Path == "auth.json" && entry.State == MigratePending {
			found = true
		}
	}
	if !found {
		t.Fatalf("the entry classifies dangling-to-declared: %+v", home.Entries)
	}
	if plan := report.Text(MigratePhasePlan); !strings.Contains(plan, "nothing to migrate") {
		t.Fatalf("an empty plan says so:\n%s", plan)
	}
	apply := fx.migrateRequest()
	apply.Expect = report.Hash
	result, err := ApplyMigration(apply)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Applied) != 0 {
		t.Fatalf("apply executes nothing: %+v", result.Applied)
	}
	if target, err := os.Readlink(link); err != nil || target != want {
		t.Fatalf("the pending link is untouched: %q (%v)", target, err)
	}
}

// TestMigrateUnknownOperands covers the scope gates.
func TestMigrateUnknownOperands(t *testing.T) {
	fx := writeManagedFixture(t, "acme")
	req := fx.migrateRequest()
	req.Profile = "ghost"
	if _, err := PlanMigration(req); err == nil || !strings.Contains(err.Error(), DiagProfileUnknown) {
		t.Fatalf("an uninstalled profile must fail with %s, got %v", DiagProfileUnknown, err)
	}
	req = fx.migrateRequest()
	req.EnvID = "cursor"
	if _, err := PlanMigration(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagUnknown) {
		t.Fatalf("an unregistered environment must fail with %s, got %v", envregistry.DiagUnknown, err)
	}
	if _, err := ApplyMigration(req); err == nil || !strings.Contains(err.Error(), envregistry.DiagUnknown) {
		t.Fatalf("apply gates the environment too, got %v", err)
	}
}

// TestMigrateInspectIsReadOnly proves planning writes nothing: no file
// changes anywhere in scope, and no default profile is created (unlike
// List, planning never ensures anything).
func TestMigrateInspectIsReadOnly(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	before := snapshotCredentialScope(t, fx)
	profilesBefore, err := os.ReadDir(ProfilesDir(fx.home))
	if err != nil {
		t.Fatal(err)
	}
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	_ = report.Text(MigratePhaseInspect)
	_ = report.Text(MigratePhasePlan)
	if _, err := PlanMigration(fx.migrateRequest()); err != nil {
		t.Fatal(err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("planning writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
	profilesAfter, err := os.ReadDir(ProfilesDir(fx.home))
	if err != nil {
		t.Fatal(err)
	}
	if len(profilesAfter) != len(profilesBefore) {
		t.Fatalf("planning creates no profile: %v vs %v", profilesBefore, profilesAfter)
	}
	if _, err := os.Stat(sourcePath(fx.home, "default")); !os.IsNotExist(err) {
		t.Fatalf("planning creates no default profile: %v", err)
	}
}

// mistargetedPiFixture provisions the operator's Pi shape: a recorded
// link still aimed at the pre-0017 native root while the live bytes
// sit at the agent root. It returns the managed link, the old target,
// and the agent-root target.
func mistargetedPiFixture(t *testing.T) (*managedFixture, string, string, string) {
	t.Helper()
	fx := writeManagedFixture(t, "acme")
	native := fx.native["pi"]
	agentAuth := filepath.Join(native, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(native, "auth.json")
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	return fx, link, oldTarget, agentAuth
}

// TestMigrateApplyRequiresPlan is the committed reviewer missing-plan
// probe: apply without a prior complete plan identity refuses before
// any mutation — no link moves, no journal appears — and names the
// --plan invocation that prints the required hash. A mutant that drops
// the plan requirement applies cleanly and fails the refusal.
func TestMigrateApplyRequiresPlan(t *testing.T) {
	requireLinkCapability(t)
	fx, link, oldTarget, _ := mistargetedPiFixture(t)
	before := snapshotCredentialScope(t, fx)
	result, err := ApplyMigration(fx.migrateRequest())
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) {
		t.Fatalf("apply without a plan must refuse with %s, got %v", envregistry.DiagCredentialConflict, err)
	}
	for _, want := range []string{"needs a prior plan", "--plan", "--expect"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the refusal names %q: %v", want, err)
		}
	}
	if result == nil || result.Report == nil || len(result.Applied) != 0 {
		t.Fatalf("a plan-less apply executes nothing: %+v", result)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the link is untouched: %q (%v)", target, err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("a plan-less apply journals nothing: %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("a plan-less apply writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
}

// failWriter refuses every write: the print-before-mutation row proves
// an undeliverable plan fails the apply before anything changes.
type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) { return 0, errors.New("plan output refused") }

// TestMigratePrintFailureRefusesBeforeMutation proves the locked plan
// prints before the first mutation: when the output cannot be
// delivered, apply fails with zero writes and no journal. The temporal
// CLI half is TestEnvMigratePrintBeforeWrite.
func TestMigratePrintFailureRefusesBeforeMutation(t *testing.T) {
	requireLinkCapability(t)
	fx, link, oldTarget, _ := mistargetedPiFixture(t)
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	req := fx.migrateRequest()
	req.Expect = report.Hash
	req.Print = failWriter{}
	before := snapshotCredentialScope(t, fx)
	result, err := ApplyMigration(req)
	if err == nil || !strings.Contains(err.Error(), "cannot print") {
		t.Fatalf("an undeliverable plan must refuse before mutating, got %v", err)
	}
	if result == nil || result.Report == nil || len(result.Applied) != 0 {
		t.Fatalf("an unprinted apply executes nothing: %+v", result)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the link is untouched: %q (%v)", target, err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("an unprinted apply journals nothing: %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("an unprinted apply writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
}

// TestMigrateMarkerDriftRefuses is the committed reviewer marker-drift
// probe: a marker edit between plan and apply (here the valid
// profile.lock_sha256 field) re-hashes the plan, so the stale hash
// refuses with zero writes. The check digests raw marker bytes —
// credential bytes are never read for it. A mutant hashing without the
// marker accepts the stale plan and fails the refusal.
func TestMigrateMarkerDriftRefuses(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(fx.native["pi"], "auth.json")
	_ = os.Remove(link)
	if err := os.Symlink(oldTarget, link); err != nil {
		t.Fatal(err)
	}
	req := fx.migrateRequest()
	before, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	markerPath := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), envmarker.Name)
	raw, err := os.ReadFile(markerPath)
	if err != nil {
		t.Fatal(err)
	}
	marker, err := envmarker.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	marker.Profile.LockSHA256 = strings.Repeat("a", 64)
	updated, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(markerPath, updated, 0o600); err != nil {
		t.Fatal(err)
	}
	snap := snapshotCredentialScope(t, fx)
	req.Expect = before.Hash
	result, err := ApplyMigration(req)
	if err == nil || !strings.Contains(err.Error(), "plan drift") {
		t.Fatalf("a marker edit must refuse with plan drift, got %v", err)
	}
	if result == nil || result.Report == nil || len(result.Applied) != 0 {
		t.Fatalf("a drifted apply executes nothing: %+v", result)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the link is untouched: %q (%v)", target, err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("a drifted apply journals nothing: %v", err)
	}
	added, removed, changed := snapshotDiff(snap, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("a drifted apply writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
}

// TestMigrateFailedRelinkKeepsOldLink is the committed reviewer
// syscall-failure probe, adapted to the journaled executor: an
// OS-rejected oversized target fails the relink with the old link
// standing (never a remove-then-create window), and the journal-driven
// reconciliation keeps the prior link. Helper-level by construction —
// the production-entry syscall rows are
// TestMigrateSyscallFailureRollsBack — so this pins the OS boundary,
// not the entry.
func TestMigrateFailedRelinkKeepsOldLink(t *testing.T) {
	home := t.TempDir()
	full := filepath.Join(ManagedHomeDir(home, "acme", "pi"), "auth.json")
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("old-target", full); err != nil {
		t.Fatal(err)
	}
	oversized := strings.Repeat("x", 100000)
	journal := &migrationJournal{
		Version: migrationJournalVersion,
		Plan:    strings.Repeat("0", 64),
		Ops: []migrationJournalOp{{
			Profile: "acme", EnvID: "pi", Kind: MigrateOpRelink,
			Path: "auth.json", From: "old-target", To: oversized,
		}},
		Markers: []migrationJournalMarker{},
	}
	ops := []MigrateOp{{
		Profile: "acme", EnvID: "pi", Kind: MigrateOpRelink,
		Path: "auth.json", From: "old-target", To: oversized,
	}}
	applied, err := executeMigrationOps(MigrateRequest{Home: home}, ops, journal)
	if err == nil {
		t.Fatal("expected OS rejection of the oversized target")
	}
	if len(applied) != 0 {
		t.Fatalf("a failed relink applies nothing: %+v", applied)
	}
	if target, err := os.Readlink(full); err != nil || target != "old-target" {
		t.Fatalf("the failed relink leaves the old link standing: %q (%v)", target, err)
	}
	if err := reconcileJournalLinks(home, journal, os.Symlink, os.Rename); err != nil {
		t.Fatalf("reconciliation keeps the prior link: %v", err)
	}
	if target, err := os.Readlink(full); err != nil || target != "old-target" {
		t.Fatalf("the prior link stands after reconciliation: %q (%v)", target, err)
	}
	strays, err := filepath.Glob(filepath.Join(filepath.Dir(full), ".migrate-*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(strays) != 0 {
		t.Fatalf("no temporary links survive a failed relink: %v", strays)
	}
}

// TestMigrateSyscallFailureRollsBack drives the syscall boundary
// through the production entry: an injected Symlink failure and an
// injected Rename failure each fail the apply with the old link
// standing, the prior state restored, the journal deleted, and no
// temporary links or credential copies left behind. A mutant restoring
// remove-then-create loses the old link on the symlink half and fails
// the standing-link assertion.
func TestMigrateSyscallFailureRollsBack(t *testing.T) {
	requireLinkCapability(t)
	setup := func(t *testing.T) (*managedFixture, MigrateRequest, string, string, string) {
		t.Helper()
		fx := writeManagedFixture(t, "acme")
		native := fx.native["pi"]
		agentAuth := filepath.Join(native, "agent", "auth.json")
		if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
			t.Fatal(err)
		}
		credential := "CRED-SENTINEL-SYSCALL-4b2d\n"
		if err := os.WriteFile(agentAuth, []byte(credential), 0o600); err != nil {
			t.Fatal(err)
		}
		provision(t, fx, "pi", envregistry.DefaultMachineConfig())
		link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
		oldTarget := filepath.Join(native, "auth.json")
		_ = os.Remove(link)
		if err := os.Symlink(oldTarget, link); err != nil {
			t.Fatal(err)
		}
		req := fx.migrateRequest()
		report, err := PlanMigration(req)
		if err != nil {
			t.Fatal(err)
		}
		if len(report.Ops()) != 1 {
			t.Fatalf("one relink op: %+v", report.Ops())
		}
		req.Expect = report.Hash
		return fx, req, link, oldTarget, credential
	}
	assertRolledBack := func(t *testing.T, fx *managedFixture, link, oldTarget, credential string, before fileSnapshot, err error, want string) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("the injected failure must surface, got %v", err)
		}
		if !strings.Contains(err.Error(), "rolled back") {
			t.Fatalf("the failure rolls back, got %v", err)
		}
		if target, err := os.Readlink(link); err != nil || target != oldTarget {
			t.Fatalf("the old link stands through the failure: %q (%v)", target, err)
		}
		agentAuth := filepath.Join(fx.native["pi"], "agent", "auth.json")
		holders := credentialHolders(t, fx, credential)
		if len(holders) != 1 || !holders[agentAuth] {
			t.Fatalf("no file newly carries credential content: %v", holders)
		}
		if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
			t.Fatalf("the rollback deletes the journal: %v", err)
		}
		strays, err := filepath.Glob(filepath.Join(filepath.Dir(link), ".migrate-*.tmp"))
		if err != nil {
			t.Fatal(err)
		}
		if len(strays) != 0 {
			t.Fatalf("no temporary links survive the failure: %v", strays)
		}
		added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
		if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
			t.Fatalf("the failure writes nothing durable: added=%v removed=%v changed=%v", added, removed, changed)
		}
	}
	t.Run("symlink failure", func(t *testing.T) {
		fx, req, link, oldTarget, credential := setup(t)
		before := snapshotCredentialScope(t, fx)
		req.symlink = func(_, _ string) error {
			return errors.New("injected symlink failure")
		}
		_, err := ApplyMigration(req)
		assertRolledBack(t, fx, link, oldTarget, credential, before, err, "injected symlink failure")
	})
	t.Run("rename failure", func(t *testing.T) {
		fx, req, link, oldTarget, credential := setup(t)
		before := snapshotCredentialScope(t, fx)
		req.rename = func(_, _ string) error {
			return errors.New("injected rename failure")
		}
		_, err := ApplyMigration(req)
		assertRolledBack(t, fx, link, oldTarget, credential, before, err, "injected rename failure")
	})
}

// TestMigrateInterruptedApplyRecovers proves crash recovery through the
// production entries: a simulated kill after the first operation leaves
// the journal standing with links half-done; inspect and plan report
// the interruption read-only; the next apply with the same plan hash
// recovers to the prior state and then executes the whole plan,
// ending migrated with the journal gone.
func TestMigrateInterruptedApplyRecovers(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	piNative := fx.native["pi"]
	agentAuth := filepath.Join(piNative, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	codexAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
	if err := os.WriteFile(codexAuth, []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	piLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(piNative, "auth.json")
	_ = os.Remove(piLink)
	if err := os.Symlink(oldTarget, piLink); err != nil {
		t.Fatal(err)
	}
	codexLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	isolated := envregistry.DefaultMachineConfig()
	isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
	req := fx.migrateRequest()
	req.Machine = isolated
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 2 {
		t.Fatalf("the plan carries two operations: %+v", report.Ops())
	}
	req.Expect = report.Hash
	// Kill after the first operation: registry order puts the codex
	// unlink first and the Pi relink second.
	crash := req
	crash.InjectFault = func(point string, applied int) error {
		if point == MigrateFaultCrash && applied >= 1 {
			return errors.New("kill -9")
		}
		return nil
	}
	interrupted, err := ApplyMigration(crash)
	if err == nil || !strings.Contains(err.Error(), "interrupted") {
		t.Fatalf("the crash must interrupt, got %v", err)
	}
	if interrupted == nil || len(interrupted.Applied) != 1 {
		t.Fatalf("one operation stands at the kill: %+v", interrupted)
	}
	if _, err := os.Lstat(codexLink); !os.IsNotExist(err) {
		t.Fatalf("the first operation landed before the kill: %v", err)
	}
	if target, err := os.Readlink(piLink); err != nil || target != oldTarget {
		t.Fatalf("the second operation never ran: %q (%v)", target, err)
	}
	journal, err := readMigrationJournal(fx.home)
	if err != nil || journal == nil {
		t.Fatalf("the journal stands after the kill: %+v (%v)", journal, err)
	}
	if journal.Plan != report.Hash || len(journal.Ops) != 2 || !journal.Ops[0].Done || journal.Ops[1].Done {
		t.Fatalf("the journal records one of two done: %+v", journal)
	}
	// The read-only phases report the interruption without recovering.
	stale, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if stale.Recovery == nil || stale.Recovery.Plan != report.Hash || stale.Recovery.Done != 1 {
		t.Fatalf("the plan surfaces the interruption: %+v", stale.Recovery)
	}
	for _, phase := range []string{MigratePhaseInspect, MigratePhasePlan} {
		text := stale.Text(phase)
		if !strings.Contains(text, "interrupted migration apply") || !strings.Contains(text, report.Hash) {
			t.Fatalf("%s banners the interruption:\n%s", phase, text)
		}
	}
	if plan := stale.Text(MigratePhasePlan); !strings.Contains(plan, "recover:") {
		t.Fatalf("the interrupted plan points at recovery:\n%s", plan)
	}
	if _, err := os.Lstat(codexLink); !os.IsNotExist(err) {
		t.Fatal("planning recovers nothing")
	}
	// The next apply with the same hash recovers, then executes.
	var printed strings.Builder
	retry := req
	retry.Print = &printed
	retry.InjectFault = nil
	result, err := ApplyMigration(retry)
	if err != nil {
		t.Fatal(err)
	}
	if result.Recovered == nil || result.Recovered.Plan != report.Hash {
		t.Fatalf("the apply reports the recovery: %+v", result.Recovered)
	}
	if len(result.Applied) != 2 {
		t.Fatalf("the recovered apply executes the whole plan: %+v", result.Applied)
	}
	out := printed.String()
	announced := strings.Index(out, "recovering interrupted migration apply")
	planned := strings.Index(out, "credential migration plan")
	if announced < 0 || planned < 0 || announced > planned {
		t.Fatalf("the recovery announces before the locked plan prints:\n%s", out)
	}
	if target, err := os.Readlink(piLink); err != nil || target != agentAuth {
		t.Fatalf("the Pi link migrated: %q (%v)", target, err)
	}
	if _, err := os.Lstat(codexLink); !os.IsNotExist(err) {
		t.Fatalf("the codex link is unlinked: %v", err)
	}
	marker := readManagedMarker(t, fx, "codex_cli")
	if marker.Passthrough == nil || len(*marker.Passthrough) != 0 {
		t.Fatalf("the record drops the unlinked entry: %+v", marker.Passthrough)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("the recovered apply deletes the journal: %v", err)
	}
}

// TestMigratePublishFailureRevertsLinks proves a failed marker
// publication reverts the links: the apply fails naming the
// publication, every link is back at its prior target, prior markers
// restore byte-identical, native bytes are untouched, and the journal
// is deleted.
func TestMigratePublishFailureRevertsLinks(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	piNative := fx.native["pi"]
	agentAuth := filepath.Join(piNative, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	codexAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
	if err := os.WriteFile(codexAuth, []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	piLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(piNative, "auth.json")
	_ = os.Remove(piLink)
	if err := os.Symlink(oldTarget, piLink); err != nil {
		t.Fatal(err)
	}
	codexLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	isolated := envregistry.DefaultMachineConfig()
	isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
	req := fx.migrateRequest()
	req.Machine = isolated
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 2 {
		t.Fatalf("the plan carries two operations: %+v", report.Ops())
	}
	markers := map[string]string{}
	for _, env := range []string{"pi", "codex_cli"} {
		payload, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, env), envmarker.Name))
		if err != nil {
			t.Fatal(err)
		}
		markers[env] = string(payload)
	}
	natives := map[string]string{
		agentAuth: sha256Of(t, agentAuth),
		codexAuth: sha256Of(t, codexAuth),
	}
	req.Expect = report.Hash
	req.InjectFault = func(point string, _ int) error {
		if point == MigrateFaultPublishBegin {
			return errors.New("injected publish failure")
		}
		return nil
	}
	_, err = ApplyMigration(req)
	if err == nil || !strings.Contains(err.Error(), "marker publish failed") {
		t.Fatalf("the publish failure must surface, got %v", err)
	}
	if !strings.Contains(err.Error(), "rolled back") {
		t.Fatalf("the publish failure reverts the links, got %v", err)
	}
	if target, err := os.Readlink(piLink); err != nil || target != oldTarget {
		t.Fatalf("the Pi link moves back to its prior target: %q (%v)", target, err)
	}
	if target, err := os.Readlink(codexLink); err != nil || target != codexAuth {
		t.Fatalf("the codex link is re-created at its prior target: %q (%v)", target, err)
	}
	for _, env := range []string{"pi", "codex_cli"} {
		payload, err := os.ReadFile(filepath.Join(ManagedHomeDir(fx.home, fx.profile, env), envmarker.Name))
		if err != nil || string(payload) != markers[env] {
			t.Fatalf("the prior %s marker restores byte-identical", env)
		}
	}
	for path, want := range natives {
		if digest := sha256Of(t, path); digest != want {
			t.Fatalf("native bytes untouched: %s", path)
		}
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("the rollback deletes the journal: %v", err)
	}
}

// TestMigrateSealedJournalDeletesWithoutTouching proves a sealed
// (fully applied) journal never rolls back completed work: the next
// apply deletes it without touching a link or marker byte — here the
// retry honestly drift-refuses, because the world is migrated and no
// longer matches the pre-migration plan — and a re-plan converges.
func TestMigrateSealedJournalDeletesWithoutTouching(t *testing.T) {
	requireLinkCapability(t)
	fx, piLink, oldTarget, agentAuth := mistargetedPiFixture(t)
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	apply := fx.migrateRequest()
	apply.Expect = report.Hash
	if _, err := ApplyMigration(apply); err != nil {
		t.Fatal(err)
	}
	// The crash shape: a sealed journal over completed work.
	sealed := &migrationJournal{
		Version:     migrationJournalVersion,
		Plan:        report.Hash,
		Ops:         []migrationJournalOp{{Profile: "acme", EnvID: "pi", Kind: MigrateOpRelink, Path: "auth.json", From: oldTarget, To: agentAuth, Done: true}},
		Markers:     []migrationJournalMarker{},
		MarkersDone: true,
		Complete:    true,
	}
	if err := writeMigrationJournal(fx.home, sealed); err != nil {
		t.Fatal(err)
	}
	stale, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if stale.Recovery == nil || stale.Recovery.Plan != report.Hash {
		t.Fatalf("the plan surfaces the sealed journal: %+v", stale.Recovery)
	}
	if plan := stale.Text(MigratePhasePlan); !strings.Contains(plan, "recover:") {
		t.Fatalf("the sealed plan points at recovery:\n%s", plan)
	}
	before := snapshotCredentialScope(t, fx)
	retry := fx.migrateRequest()
	retry.Expect = report.Hash
	result, err := ApplyMigration(retry)
	if err == nil || !strings.Contains(err.Error(), "plan drift") {
		t.Fatalf("the migrated world honestly drift-refuses the pre-migration hash, got %v", err)
	}
	if result == nil || result.Recovered == nil || result.Recovered.Plan != report.Hash {
		t.Fatalf("the apply reports the sealed recovery: %+v", result)
	}
	if target, err := os.Readlink(piLink); err != nil || target != agentAuth {
		t.Fatalf("completed work is never rolled back: %q (%v)", target, err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("the sealed journal deletes: %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("the sealed recovery writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
	fresh, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh.Ops()) != 0 || len(fresh.Conflicts()) != 0 {
		t.Fatalf("the re-plan converges: %+v %+v", fresh.Ops(), fresh.Conflicts())
	}
	converge := fx.migrateRequest()
	converge.Expect = fresh.Hash
	if _, err := ApplyMigration(converge); err != nil {
		t.Fatalf("the converged apply succeeds: %v", err)
	}
}

// TestMigrateBrokenJournalRefuses proves an unusable journal fails
// closed: inspect and plan banner it as blocked, and apply refuses
// naming the backup-first operator choice with zero writes and the
// journal left for inspection.
func TestMigrateBrokenJournalRefuses(t *testing.T) {
	requireLinkCapability(t)
	fx, link, oldTarget, _ := mistargetedPiFixture(t)
	if err := os.MkdirAll(filepath.Dir(migrationJournalPath(fx.home)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(migrationJournalPath(fx.home), []byte("{not a journal"), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if report.Recovery == nil || report.Recovery.Broken == "" {
		t.Fatalf("the plan marks the journal broken: %+v", report.Recovery)
	}
	if plan := report.Text(MigratePhasePlan); !strings.Contains(plan, "unusable") || !strings.Contains(plan, "blocked:") {
		t.Fatalf("the plan banners the broken journal:\n%s", plan)
	}
	if inspect := report.Text(MigratePhaseInspect); !strings.Contains(inspect, "unusable") {
		t.Fatalf("the inventory banners the broken journal:\n%s", inspect)
	}
	before := snapshotCredentialScope(t, fx)
	req := fx.migrateRequest()
	req.Expect = report.Hash
	result, err := ApplyMigration(req)
	if err == nil || !strings.Contains(err.Error(), "unusable") {
		t.Fatalf("a broken journal must refuse, got %v", err)
	}
	if !strings.Contains(err.Error(), "back it up out of band") {
		t.Fatalf("the refusal names the operator choice: %v", err)
	}
	if result == nil || result.Report == nil || len(result.Applied) != 0 {
		t.Fatalf("a broken journal applies nothing: %+v", result)
	}
	if target, err := os.Readlink(link); err != nil || target != oldTarget {
		t.Fatalf("the link is untouched: %q (%v)", target, err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); err != nil {
		t.Fatalf("the broken journal is left for inspection: %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("a broken journal writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
}

// TestMigrateRecoveryRefusesUnexpectedTarget proves the recovery
// validates the entire inventory before any write: an interrupted apply
// whose pending link was re-pointed out of band to an unexpected target
// refuses naming the exact operator choice, preserving the unexpected
// link, the half-done state, and the journal. A mutant that recovers
// before validating silently moves the link back and accepts the stale
// plan.
func TestMigrateRecoveryRefusesUnexpectedTarget(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	piNative := fx.native["pi"]
	agentAuth := filepath.Join(piNative, "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(agentAuth), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(agentAuth, []byte("{\"t\":\"operator-pi\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	codexAuth := filepath.Join(fx.native["codex_cli"], "auth.json")
	if err := os.WriteFile(codexAuth, []byte("{\"t\":\"operator\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	provision(t, fx, "codex_cli", envregistry.DefaultMachineConfig())
	piLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	oldTarget := filepath.Join(piNative, "auth.json")
	_ = os.Remove(piLink)
	if err := os.Symlink(oldTarget, piLink); err != nil {
		t.Fatal(err)
	}
	codexLink := filepath.Join(ManagedHomeDir(fx.home, "acme", "codex_cli"), "auth.json")
	isolated := envregistry.DefaultMachineConfig()
	isolated.Isolation = map[string]map[string]string{"acme": {"codex_cli": "isolated"}}
	req := fx.migrateRequest()
	req.Machine = isolated
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 2 {
		t.Fatalf("the plan carries two operations: %+v", report.Ops())
	}
	req.Expect = report.Hash
	crash := req
	crash.InjectFault = func(point string, applied int) error {
		if point == MigrateFaultCrash && applied >= 1 {
			return errors.New("kill -9")
		}
		return nil
	}
	if _, err := ApplyMigration(crash); err == nil || !strings.Contains(err.Error(), "interrupted") {
		t.Fatalf("the crash must interrupt, got %v", err)
	}
	unexpected := filepath.Join(t.TempDir(), "operator-target")
	_ = os.Remove(piLink)
	if err := os.Symlink(unexpected, piLink); err != nil {
		t.Fatal(err)
	}
	req.InjectFault = nil
	before := snapshotCredentialScope(t, fx)
	result, err := ApplyMigration(req)
	if err == nil {
		t.Fatal("an unexpected link target during recovery must refuse")
	}
	for _, want := range []string{"recovery found", unexpected, "out of band", "re-run"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("the refusal names %q: %v", want, err)
		}
	}
	if result == nil || len(result.Applied) != 0 {
		t.Fatalf("a refused recovery applies nothing: %+v", result)
	}
	if target, err := os.Readlink(piLink); err != nil || target != unexpected {
		t.Fatalf("the unexpected link is preserved: %q (%v)", target, err)
	}
	if _, err := os.Lstat(codexLink); !os.IsNotExist(err) {
		t.Fatalf("the half-done state is preserved: %v", err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); err != nil {
		t.Fatalf("the journal stands for the operator fix: %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added) != 0 || len(removed) != 0 || len(changed) != 0 {
		t.Fatalf("a refused recovery writes nothing: added=%v removed=%v changed=%v", added, removed, changed)
	}
}

// TestMigrateRecoveryCleansOwnedTemp proves the journal owns its
// temporary links exactly: a leftover owned temp (symlink to the
// recorded target at the recorded path, from a kill between the symlink
// and the rename) is removed by the next recovery, which then executes
// the plan. Foreign paths are never touched (see
// TestReviewerRecoveryPreservesRegularTemp).
func TestMigrateRecoveryCleansOwnedTemp(t *testing.T) {
	requireLinkCapability(t)
	fx, piLink, oldTarget, agentAuth := mistargetedPiFixture(t)
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 {
		t.Fatalf("one relink op: %+v", report.Ops())
	}
	owned := filepath.Join(filepath.Dir(piLink), ".migrate-owned-test.tmp")
	if err := os.Symlink(agentAuth, owned); err != nil {
		t.Fatal(err)
	}
	journal := &migrationJournal{
		Version: migrationJournalVersion,
		Plan:    report.Hash,
		Ops: []migrationJournalOp{{
			Profile: "acme", EnvID: "pi", Kind: MigrateOpRelink,
			Path: "auth.json", From: oldTarget, To: agentAuth,
			Temp: owned, TempTarget: agentAuth,
		}},
		Markers: []migrationJournalMarker{},
	}
	if err := writeMigrationJournal(fx.home, journal); err != nil {
		t.Fatal(err)
	}
	retry := fx.migrateRequest()
	retry.Expect = report.Hash
	result, err := ApplyMigration(retry)
	if err != nil {
		t.Fatal(err)
	}
	if result.Recovered == nil || result.Recovered.Plan != report.Hash {
		t.Fatalf("the apply reports the recovery: %+v", result.Recovered)
	}
	if _, err := os.Lstat(owned); !os.IsNotExist(err) {
		t.Fatalf("the owned temp is cleaned: %v", err)
	}
	if target, err := os.Readlink(piLink); err != nil || target != agentAuth {
		t.Fatalf("the Pi link migrated: %q (%v)", target, err)
	}
	if _, err := os.Lstat(migrationJournalPath(fx.home)); !os.IsNotExist(err) {
		t.Fatalf("the recovered apply deletes the journal: %v", err)
	}
}
