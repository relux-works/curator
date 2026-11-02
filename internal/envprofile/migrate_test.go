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
	"github.com/relux-works/curator/internal/envorigin"
	"github.com/relux-works/curator/internal/envregistry"
	"github.com/relux-works/curator/internal/transaction"
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
// relinks (a symlink, never a copy) and records migration provenance in
// the schema-2 marker.
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
	if payload, err := os.ReadFile(markerPath); err != nil || string(payload) == string(markerBefore) {
		t.Fatal("a schema-2 relink records migration provenance in the marker")
	}
	migratedMarker := readManagedMarker(t, fx, "pi")
	if migratedMarker.Passthrough == nil || len(*migratedMarker.Passthrough) != 1 || (*migratedMarker.Passthrough)[0].Provenance != "migrated" {
		t.Fatalf("migration publishes a complete credential record: %+v", migratedMarker.Passthrough)
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
	changedSet := map[string]bool{}
	for _, path := range changed {
		changedSet[path] = true
	}
	if len(changed) != 2 || !changedSet[link] || !changedSet[markerPath] {
		t.Fatalf("the link target and migration provenance marker change: %v", changed)
	}
}

// TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker proves that a
// schema-1 marker stays old until migration already requires a marker update
// for an unlink; that replacement upgrades once and records the complete v2
// credential set for the new isolated state.
func TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker(t *testing.T) {
	enableV1WritersForTest(t)
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	markerPath := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), envmarker.Name)
	marker := readManagedMarker(t, fx, "pi")
	legacyEntries := []envmarker.Passthrough{}
	if marker.Passthrough != nil {
		for _, entry := range *marker.Passthrough {
			if entry.Path != "" {
				legacyEntries = append(legacyEntries, envmarker.Passthrough{Path: entry.Path, Strategy: entry.Strategy})
			}
		}
	}
	legacy := *marker
	legacy.Version = envmarker.VersionV1
	legacy.Passthrough = &legacyEntries
	legacyBytes, err := legacy.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(markerPath, legacyBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("schema-1 marker's recorded link is present: %v", err)
	}
	config := envregistry.DefaultMachineConfig()
	config.Isolation = map[string]map[string]string{"acme": {"pi": envregistry.IsolationIsolated}}
	req := fx.migrateRequest()
	req.Machine = config
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 || report.Ops()[0].Kind != MigrateOpUnlink || report.Ops()[0].Path != "auth.json" {
		t.Fatalf("migration requires an unlink marker update: %+v", report.Ops())
	}
	req.Expect = report.Hash
	if _, err := ApplyMigration(req); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("the migration unlinks the recorded passthrough: %v", err)
	}
	upgraded := readManagedMarker(t, fx, "pi")
	if upgraded.Version != envmarker.VersionV2 || upgraded.Passthrough == nil || len(*upgraded.Passthrough) != 0 {
		t.Fatalf("marker replacement upgrades with the complete isolated record set: %+v", upgraded)
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
	piMarker := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), envmarker.Name)
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
	if len(changed) != 3 || !changedSet[piLink] || !changedSet[piMarker] || !changedSet[codexMarker] {
		t.Fatalf("the relinked link and each migrated home's marker change: %v", changed)
	}
	piRecord := readManagedMarker(t, fx, "pi")
	if piRecord.Passthrough == nil || len(*piRecord.Passthrough) != 1 || (*piRecord.Passthrough)[0].Provenance != "migrated" {
		t.Fatalf("the relinked Pi credential records migration provenance: %+v", piRecord.Passthrough)
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
		hit := false
		crash.transactionOptions = []transaction.Option{transaction.WithHooks(transaction.Hooks{Fault: func(event transaction.Event) error {
			if !hit && event.Point == transaction.PointAfterInstall && event.Class == "credential-entry" {
				hit = true
				panic("no-copy-unit-stop")
			}
			return nil
		}})}
		func() {
			defer func() {
				if v := recover(); v != "no-copy-unit-stop" {
					t.Fatalf("wrong stop: %v", v)
				}
			}()
			_, _ = ApplyMigration(crash)
		}()
		if !hit {
			t.Fatal("actual credential owner did not reach stop")
		}
		journals, err := filepath.Glob(filepath.Join(gx.home, "state", "transactions", "v1", "*.json"))
		if err != nil || len(journals) == 0 {
			t.Fatalf("standing transaction journal: %v %v", journals, err)
		}
		for _, path := range journals {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, sentinel := range []string{gxPiCredential, gxCodexCredential} {
				if strings.Contains(string(raw), sentinel) {
					t.Fatal("journal contains fixture credential payload")
				}
			}
		}
		assertNoCopy(t, "while complete unit stands")
		op, err := beginOperation(gx.home)
		if err != nil {
			t.Fatal(err)
		}
		if err = op.close(); err != nil {
			t.Fatal(err)
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
		before := nativeSnapshot(t, fx.home)
		req := fx.migrateRequest()
		if report, err := PlanMigration(req); err == nil || !strings.Contains(err.Error(), envmarker.DiagMarkerInvalid) || report != nil {
			t.Fatalf("invalid marker common planning refusal: %v", err)
		}
		req.Expect = strings.Repeat("a", 64)
		if result, err := ApplyMigration(req); err == nil || !strings.Contains(err.Error(), envmarker.DiagMarkerInvalid) || result != nil {
			t.Fatalf("invalid marker common apply refusal: %v", err)
		}
		if nativeSnapshot(t, fx.home) != before {
			t.Fatal("invalid marker refusal changed state")
		}

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
// uninspectable distinction for recorded Pi authority: a re-point planned
// through a substituted non-directory ancestor fails physical origin validation
// before credential inspection — never silence, and never the absence
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
	before := snapshotCredentialScope(t, fx)
	req := fx.migrateRequest()
	report, err := PlanMigration(req)
	if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || !strings.Contains(err.Error(), "ancestor is not a readable directory") || report != nil {
		t.Fatalf("physical origin rejection: %v", err)
	}
	req.Expect = strings.Repeat("a", 64)
	if result, err := ApplyMigration(req); err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || result != nil && len(result.Applied) > 0 {
		t.Fatalf("physical origin apply refusal: %v", err)
	}
	added, removed, changed := snapshotDiff(before, snapshotCredentialScope(t, fx))
	if len(added)+len(removed)+len(changed) != 0 {
		t.Fatalf("invalid origin refusal changed preimages: %v %v %v", added, removed, changed)
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
	if result != nil && len(result.Applied) != 0 {
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

// Returned faults after credential or marker install restore both homes,
// captured origins and native fixtures; the healthy Pi migration is nearby.
func TestMigrateRollbackRestoresPriorState(t *testing.T) {
	for _, mode := range []string{"after-credential", "after-marker"} {
		t.Run(mode, func(t *testing.T) { migrationTransactionCase(t, mode) })
	}
}

// TestMigrateOldRootBytesWarnAndProceed proves bytes at the old Pi root
// alone (agent root empty) warn loudly but do not block: the relink
// leaves the old bytes untouched, the manager never moves them, and the
// re-pointed link dangles until the operator reconciles native bytes out of
// band. Explicit migration may correct that recorded pointer, but Pi Resolve
// still refuses the absent native leaf without a fragment.
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
	auth := filepath.Join(fx.native["pi"], "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(auth), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auth, nil, 0600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	if err := os.Remove(auth); err != nil {
		t.Fatal(err)
	}
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
	if err == nil || !strings.Contains(err.Error(), envregistry.DiagCredentialConflict) || !strings.Contains(err.Error(), "reason=dangling") || current != nil && len(current.Document) > 0 {
		t.Fatalf("Pi absent native leaf must visibly refuse launch: %v", err)
	}
}

// TestMigratePendingNeedsNothing proves dangling-to-declared entries
// get no operation: the plan is empty, apply succeeds applying nothing,
// and the pending link is untouched.
func TestMigratePendingNeedsNothing(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	auth := filepath.Join(fx.native["pi"], "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(auth), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auth, nil, 0600); err != nil {
		t.Fatal(err)
	}
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	if err := os.Remove(auth); err != nil {
		t.Fatal(err)
	}
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
	if result != nil && len(result.Applied) != 0 {
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
	if result != nil && len(result.Applied) != 0 {
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

// A substituted marker lock binding refuses before mutation, regardless of
// a previously printed plan hash. Credential payloads are never read.
func TestMigrateMarkerDriftRefuses(t *testing.T) {
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	auth := filepath.Join(fx.native["pi"], "agent", "auth.json")
	if err := os.MkdirAll(filepath.Dir(auth), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(auth, nil, 0600); err != nil {
		t.Fatal(err)
	}
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
	if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || !strings.Contains(err.Error(), "marker/installed lock binding mismatch") {
		t.Fatalf("a substituted marker binding must refuse before plan execution, got %v", err)
	}
	if result != nil && len(result.Applied) != 0 {
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

// The real staging syscall rejects an oversized target before replacement;
// whole-unit compensation preserves both original links and home preimages.
func TestMigrateFailedRelinkKeepsOldLink(t *testing.T) {
	migrationTransactionCase(t, "oversized-symlink")
}

// Raw staging Symlink failure and a fault immediately before credential
// install are separate boundaries. The latter is not a Rename syscall witness.
// Both preserve complete original homes, not merely return an error.
func TestMigrateSyscallFailureRollsBack(t *testing.T) {
	for _, mode := range []string{"symlink-error", "before-install"} {
		t.Run(mode, func(t *testing.T) { migrationTransactionCase(t, mode) })
	}
}

// A fresh engine completes genuine credential-prefix intent using captured
// origins. This models an in-process restart, not an OS kill.
func TestMigrateInterruptedApplyRecovers(t *testing.T) { migrationTransactionCase(t, "restart") }

// A real marker publication fault after credential installation restores
// complete prior marker/origin/home state and preserves native fixtures.
func TestMigratePublishFailureRevertsLinks(t *testing.T) { migrationTransactionCase(t, "after-marker") }

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
// A malformed legacy record blocks read-only planning before backend lookup
// as well as Apply; it is preserved exactly rather than guessed away.
func TestMigrateBrokenJournalRefuses(t *testing.T) {
	fx, _, _, _ := mistargetedPiFixture(t)
	if err := os.MkdirAll(filepath.Dir(migrationJournalPath(fx.home)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(migrationJournalPath(fx.home), []byte("{not a journal"), 0600); err != nil {
		t.Fatal(err)
	}
	before := nativeSnapshot(t, fx.home)
	req := fx.migrateRequest()
	req.Expect = strings.Repeat("a", 64)
	for _, apply := range []bool{false, true} {
		var err error
		if apply {
			_, err = ApplyMigration(req)
		} else {
			_, err = PlanMigration(req)
		}
		if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || !strings.Contains(err.Error(), "legacy migration journal") {
			t.Fatalf("malformed journal refusal apply=%t: %v", apply, err)
		}
		if nativeSnapshot(t, fx.home) != before {
			t.Fatal("malformed journal refusal changed state")
		}
	}
}

// A substituted link after interrupted publication refuses before recovery
// continuation, preserving the unexpected entry and entire durable state.
func TestMigrateRecoveryRefusesUnexpectedTarget(t *testing.T) {
	migrationTransactionCase(t, "wrong-link")
}

// An unsealed legacy journal cannot authorize temp cleanup or recovery.
// Temp and journal remain exact; completed-receipt cleanup has a nearby control.
func TestMigrateUnsealedLegacyTempRefuses(t *testing.T) {
	fx, link, old, target := mistargetedPiFixture(t)
	report, err := PlanMigration(fx.migrateRequest())
	if err != nil {
		t.Fatal(err)
	}
	temp := filepath.Join(filepath.Dir(link), ".migrate-owned-test.tmp")
	if err = os.Symlink(target, temp); err != nil {
		t.Fatal(err)
	}
	j := &migrationJournal{Version: 1, Plan: report.Hash, Ops: []migrationJournalOp{{Profile: fx.profile, EnvID: "pi", Kind: MigrateOpRelink, Path: "auth.json", From: old, To: target, Temp: temp, TempTarget: target}}, Markers: []migrationJournalMarker{}}
	if err = writeMigrationJournal(fx.home, j); err != nil {
		t.Fatal(err)
	}
	before := nativeSnapshot(t, fx.home)
	req := fx.migrateRequest()
	req.Expect = report.Hash
	result, err := ApplyMigration(req)
	if err == nil || !strings.Contains(err.Error(), envorigin.DiagInvalid) || result != nil && len(result.Applied) > 0 {
		t.Fatalf("unsealed legacy continuation refusal: %v", err)
	}
	if nativeSnapshot(t, fx.home) != before {
		t.Fatal("unsealed legacy refusal changed journal, home or claimed temp")
	}
}

// TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode is the
// v2 writer-mode counterpart of
// TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker: with the writer
// switch explicitly on, the required marker replacement upgrades a schema-1
// marker straight to v3 carrying hash_version 2.
func TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode(t *testing.T) {
	enableV1WritersForTest(t)
	requireLinkCapability(t)
	fx := writeManagedFixture(t, "acme")
	seedLiveNativeCredentials(t, fx)
	provision(t, fx, "pi", envregistry.DefaultMachineConfig())
	markerPath := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), envmarker.Name)
	marker := readManagedMarker(t, fx, "pi")
	legacyEntries := []envmarker.Passthrough{}
	if marker.Passthrough != nil {
		for _, entry := range *marker.Passthrough {
			if entry.Path != "" {
				legacyEntries = append(legacyEntries, envmarker.Passthrough{Path: entry.Path, Strategy: entry.Strategy})
			}
		}
	}
	legacy := *marker
	legacy.Version = envmarker.VersionV1
	legacy.HashVersion = 0
	legacy.Passthrough = &legacyEntries
	legacyBytes, err := legacy.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(markerPath, legacyBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ManagedHomeDir(fx.home, "acme", "pi"), "auth.json")
	if _, err := os.Lstat(link); err != nil {
		t.Fatalf("schema-1 marker's recorded link is present: %v", err)
	}
	enableV2WritersForTest(t)
	config := envregistry.DefaultMachineConfig()
	config.Isolation = map[string]map[string]string{"acme": {"pi": envregistry.IsolationIsolated}}
	req := fx.migrateRequest()
	req.Machine = config
	report, err := PlanMigration(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Ops()) != 1 || report.Ops()[0].Kind != MigrateOpUnlink || report.Ops()[0].Path != "auth.json" {
		t.Fatalf("migration requires an unlink marker update: %+v", report.Ops())
	}
	req.Expect = report.Hash
	if _, err := ApplyMigration(req); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Fatalf("the migration unlinks the recorded passthrough: %v", err)
	}
	upgraded := readManagedMarker(t, fx, "pi")
	if upgraded.Version != envmarker.VersionV3 || upgraded.HashVersion != 2 || upgraded.Passthrough == nil || len(*upgraded.Passthrough) != 0 {
		t.Fatalf("marker replacement upgrades with the complete isolated record set: %+v", upgraded)
	}
}
