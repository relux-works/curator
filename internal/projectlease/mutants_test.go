package projectlease

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Every bounded overlay must change its named behavioral assertion, with an
// unchanged positive control first. A crash, setup error, missing anchor or
// unapplied plant never counts as a killed mutant. No live user state is used.
func TestLeaseNarrowingMutants(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	goTool := os.Getenv("CURATOR_GO")
	if goTool == "" {
		goTool = "go"
	}
	cases := []struct{ name, file, before, after, pkg, run, message string }{
		{"acquisition-returns-before-lock", "internal/projectlease/manager.go", "func (m *Manager) Acquire(ctx context.Context, r Request) (Binding, error) {", "func (m *Manager) Acquire(ctx context.Context, r Request) (Binding, error) {\n if r.LeaseID == \"two\" { return Binding{}, nil }", "projectlease", "TestPublicationSerializesLeaseAcquisition", "acquisition passed an in-flight update:"},
		{"busy-update-does-not-retry", "internal/projectlease/manager.go", "case <-timer.C:\n\t\t}", "case <-timer.C: return lastBusy\n\t\t}", "projectlease", "TestBusyWaitAndCancellation/reconcile", "busy update did not wait for reconcile:"},
		{"consumer-pruning-acquisition", "internal/scopes/consumers.go", "if err := check(); err != nil {\n\t\treturn err\n\t}\n\treturn os.WriteFile(filepath.Join(home, ConsumersName), payload, 0o644)", "return os.WriteFile(filepath.Join(home, ConsumersName), payload, 0o644)", "scopes", "TestGCPruningRetainsOwnership", "stale pruning overwrote registry"},

		{"internal-build-late-acquisition", "internal/buildcache/collect.go", "if err := requireHomeLock(lock); err != nil {\n\t\treturn err\n\t}\n\tif err := retireEntryChecked(root, name, func() error { return requireHomeLock(lock) }); err != nil {", "if err := retireEntryChecked(root, name, func() error { return nil }); err != nil {", "buildcache", "TestSweepRetainsOwnershipAfterClassification", "late stale sweep removed candidate"},

		{"ownership-during-rollback", "internal/transaction/engine.go", "return witness.Held().Validate(filepath.Dir(filepath.Dir(filepath.Dir(engine.journalRoot))))", "_ = witness; return nil", "projectlease", "TestAcquireCannotUndoRecoveredOwner", "stale owner damaged recovered state"},
		{"gc-retained-acquisition", "internal/scopes/gc.go", "func (lock maintenanceLock) AssertHeld() error { return lock.check() }", "func (lock maintenanceLock) AssertHeld() error { return nil }", "scopes", "TestCollectCannotSweepNewOwnerLease", "stale GC removed newly leased build"},
		{"runtime-removal-acquisition", "internal/scopes/gc.go", "if err := check(); err != nil {\n\t\t\t\t\treturn removed, err\n\t\t\t\t}\n\t\t\t\tif err := os.RemoveAll", "if err := os.RemoveAll", "scopes", "TestGCSweepBoundariesRetainOwnership", "runtime stale sweep removed state"},
		{"external-artifact-acquisition", "internal/buildrepo/gc.go", "if err := check(); err != nil {\n\t\t\t\treturn removed, err\n\t\t\t}\n\t\t\tif err := os.RemoveAll(path)", "if err := os.RemoveAll(path)", "scopes", "TestGCSweepBoundariesRetainOwnership", "external-artifact stale sweep removed state"},
		{"external-snapshot-acquisition", "internal/buildrepo/gc.go", "if err := check(); err != nil {\n\t\t\t\treturn removed, err\n\t\t\t}\n\t\t\tif err := os.RemoveAll(filepath.Join(snapshotsRoot, entry.Name()))", "if err := os.RemoveAll(filepath.Join(snapshotsRoot, entry.Name()))", "scopes", "TestGCSweepBoundariesRetainOwnership", "external-snapshot stale sweep removed state"},
		{"gc-recovered-lifecycle", "cmd/curator/main.go", "projectlease.RecoverUnderLock(home, lock, engine)", "engine.RecoverWithOutcome(lock)", "cmd/curator", "TestGCRecoversProjectLeaseArtifacts", "GC recovered lifecycle remained prepared"},
		{"profile-recovered-lifecycle", "internal/envprofile/lock.go", "projectlease.RecoverUnderLock(home, lock, engine)", "engine.RecoverWithOutcome(lock)", "envprofile", "TestRecoveredLeaseLifecycle", "recovered lifecycle remained prepared/active or released"},
		{"install-recovered-lifecycle", "internal/install/commit.go", "projectlease.RecoverUnderLock(journal.home, lock, journal.Engine)", "journal.Engine.RecoverWithOutcome(lock)", "install", "TestRecoveredLeaseLifecycle", "recovered lifecycle remained prepared/active or released"},

		{"profile-recovery-policy", "internal/envprofile/lock.go", "transaction.WithCommitCheck(projectlease.TransactionCheck(home))", "transaction.WithCommitCheck(func() transaction.CommitCheck { _ = projectlease.TransactionCheck(home); return nil }())", "envprofile", "TestDefaultRecoveryChecksLeaseJournal", "default lease recovery policy missing"},
		{"install-recovery-policy", "internal/install/commit.go", "transaction.WithCommitCheck(projectlease.TransactionCheck(home))", "transaction.WithCommitCheck(func() transaction.CommitCheck { _ = projectlease.TransactionCheck(home); return nil }())", "install", "TestDefaultRecoveryChecksLeaseJournal", "default lease recovery policy missing"},
		{"pending-gc-preserves-other-owners", "internal/projectlease/manager.go", "engine.PendingRecoveryFor(lock, \"project-leases\")", "engine.PendingRecovery(lock)", "scopes", "TestCollectWithUnrelatedPendingTransaction", "unrelated pending transaction blocked legacy GC"},
		{"pending-owner-filter", "internal/transaction/pending.go", "journal.ProjectIdentity == projectIdentity", "false && journal.ProjectIdentity == projectIdentity", "transaction", "TestPendingRecoveryForOwner", "owner-filtered pending"},
		{"gc-cli-recovery-policy", "cmd/curator/main.go", "transaction.WithCommitCheck(projectlease.TransactionCheck(home))", "transaction.WithCommitCheck(func() transaction.CommitCheck { _ = projectlease.TransactionCheck(home); return nil }())", "cmd/curator", "TestGCRecoversProjectLeaseArtifacts", "GC recovery committed missing artifact"},
		{"prepared-lock-identity", "internal/projectlease/transaction_check.go", "if err := witness.Held().Validate(home); err != nil {", "if err := witness.Held().Validate(home); false && err != nil {", "projectlease", "TestLeaseTransactionPolicyCurrentLock", "stale transaction policy accepted"},
		{"pending-recovery-gc", "internal/projectlease/manager.go", "if pending {\n\t\treturn Roots{}, ErrPendingRecovery", "if false && pending {\n\t\treturn Roots{}, ErrPendingRecovery", "projectlease", "TestAcquirePreparedRecoveryArtifacts", "pending GC did not refuse safely"},
		{"recovery-artifact-validation", "internal/projectlease/transaction_check.go", "return m.verifyTransactionArtifacts(journal)", "_ = m; return nil", "projectlease", "TestAcquirePreparedRecoveryArtifacts", "missing recovery artifact committed or wrong refusal"},
		{"binding-member-case", "internal/projectlease/types.go", "if err != nil || string(exact) != string(raw) {", "if false && (err != nil || string(exact) != string(raw)) {", "projectlease", "TestBindingExactMembers", "case-alias binding accepted"},
		{"transaction-commit-check", "internal/transaction/engine.go", "return engine.commitCheck(engine.commitLock, *snapshot)", "_ = snapshot; return nil", "transaction", "TestCommitCheck", "commit check refusal"},
		{"gc-exact-home-witness", "internal/projectlease/manager.go", "if err := witness.Held().Validate(home); err != nil {", "if err := witness.Held().Validate(home); false && err != nil {", "projectlease", "TestGCReferenceWitness", "foreign GC witness ="},
		{"acquire-artifact-presence", "internal/projectlease/manager.go", "if err = m.verifyArtifacts(c); err != nil {", "if err = m.verifyArtifacts(c); false && err != nil {", "projectlease", "TestAcquireArtifactsBeforeCommit", "missing artifact admission"},
		{"idle-repair-allows-drift", "internal/projectlease/manager.go", "err = m.verifyRepairView(old)", "err = m.verifyView(old.Composition, old)", "projectlease", "TestIdleRepairGeneratedDrift", "idle repair refused changed generated state"},
		{"idle-repair-native-directory", "internal/projectlease/manager.go", "if !observed.Info.Mode().IsRegular() {", "if false && !observed.Info.Mode().IsRegular() {", "projectlease", "TestIdleRepairRefusesDirectoryTakeover", "refusal <nil>, want not a regular file"},
		{"receipt-alias", "internal/projectlease/types.go", "if aliases[key] {", "if false && aliases[key] {", "projectlease", "TestReceiptPortablePaths", "invalid receipt paths accepted"},
		{"receipt-file-ancestor", "internal/projectlease/types.go", `if aliases[strings.Join(parts[:i], "/")] {`, `if false && aliases[strings.Join(parts[:i], "/")] {`, "projectlease", "TestReceiptPortablePaths", "invalid receipt paths accepted"},
		{"pid-only", "internal/projectlease/manager.go", "} else if o.ProcessGeneration != e.ProcessGeneration {", "} else if false && o.ProcessGeneration != e.ProcessGeneration {", "projectlease", "TestDetachDescendantsUnknownAndPIDReuse", "process incarnation mismatch released"},
		{"release-on-launcher-return", "internal/projectlease/manager.go", "e.State = Active\n\t\te.ProcessGeneration = observation.ProcessGeneration", "e.State = Released\n\t\te.ProcessGeneration = observation.ProcessGeneration", "projectlease", "TestLauncherExitRetainsLease", "launcher exit released transferred lease"},
		{"collect-on-ttl-alone", "internal/projectlease/manager.go", "o, err := m.observe(ctx, e.Binding)\n\t\tif err != nil {\n\t\t\te.State = Unknown", "o, err := m.observe(ctx, e.Binding)\n\t\tif err != nil {\n\t\t\te.State = Unknown\n if info, er:=os.Stat(statePath(m.home)); er==nil && time.Since(info.ModTime())>time.Hour {e.State=Released}", "projectlease", "TestDetachDescendantsUnknownAndPIDReuse", "TTL or missing liveness released lease"},
		{"ignore-descendants", "internal/projectlease/manager.go", "o.Terminated && o.OwnershipReconciled && o.DescendantsTerminated", "o.Terminated && o.OwnershipReconciled", "projectlease", "TestDetachDescendantsUnknownAndPIDReuse", "live descendants released"},
		{"abort-without-no-child-proof", "internal/projectlease/manager.go", "o.NoChildStarted && !o.Running && o.OwnershipReconciled", "!o.Running && o.OwnershipReconciled", "projectlease", "TestPreparedReservationAndTransferGates", "missing no-child proof released"},
		{"ignore-busy-update", "internal/projectlease/manager.go", "if len(sessions) == 0 {", "if true || len(sessions) == 0 {", "projectlease", "TestBusyUpdateAndIdleReuse", "refusal <nil>, want project_view_busy"},
		{"ignore-concurrency-capability", "internal/projectlease/manager.go", "(!c.ConcurrentState || !e.Composition.ConcurrentState ||", "(false ||", "projectlease", "TestStableHomesAndConcurrentSessions", "refusal <nil>, want project_view_busy"},
		{"ignore-receipt", "internal/projectlease/manager.go", "if !reflect.DeepEqual(expected, c.Generated) {", "if false && !reflect.DeepEqual(expected, c.Generated) {", "projectlease", "TestPublicationReceiptAndRollback", "refusal <nil>, want receipt"},
		{"ignore-revocation", "internal/projectlease/manager.go", "if err = m.authority.Authorize(ctx, held, c, owner); err != nil {", "if err = m.authority.Authorize(ctx, held, c, owner); false && err != nil {", "projectlease", "TestRecoveryResumeAndRevocation", "revoked acquire=<nil>"},
		{"ignore-mode-drift", "internal/projectlease/manager.go", "if runtime.GOOS != \"windows\" && uint32(info.Mode().Perm()) != surface.Mode {", "if false && runtime.GOOS != \"windows\" && uint32(info.Mode().Perm()) != surface.Mode {", "projectlease", "TestGeneratedModesAndInventory", "refusal <nil>, want mode drift"},
		{"omit-runtime-gc-roots", "internal/scopes/gc.go", "marked.runtime[entry] = true", "marked.runtime[entry] = false", "scopes", "TestCollectProjectLeaseReferencesAndUnknown", "GC="},
		{"omit-build-gc-roots", "internal/scopes/gc.go", "referenced = append(referenced, leaseRoots.BuildKeys...)", "referenced = append(referenced, []string{}...)", "scopes", "TestCollectProjectLeaseReferencesAndUnknown", "GC="},
		{"missing-ledger-is-absence", "internal/projectlease/store.go", "if HasState(home) {", "if false && HasState(home) {", "projectlease", "TestLinksDriftAndMissingState", "missing ledger became no leases"},
		{"zero-held-witness", "internal/managerlock/managerlock.go", "if h.lock == nil {\n\t\treturn ErrNotHeld\n\t}", "if h.lock == nil {return nil}", "managerlock", "TestHeldWitnessScopeAndLifetime", "zero witness accepted"},
		{"stale-held-witness", "internal/managerlock/managerlock.go", "if err := h.lock.AssertCurrent(); err != nil {", "if err := h.lock.AssertCurrent(); false && err != nil {", "managerlock", "TestHeldWitnessScopeAndLifetime", "stale witness accepted"},
		{"wrong-home-held-witness", "internal/managerlock/managerlock.go", "if canonical != h.lock.home {", "if false && canonical != h.lock.home {", "managerlock", "TestHeldWitnessScopeAndLifetime", "wrong-home witness accepted"},
		{"missing-authority-witness", "internal/projectlease/manager.go", "m.authority.Resolve(ctx, held, ref, sha)", "m.authority.Resolve(ctx, managerlock.Held{}, ref, sha)", "projectlease", "TestAuthorityRequiresHeldWitness", "manager lock is not held"},
		{"callback-before-witness-validation", "internal/projectlease/manager.go", "if err := held.Validate(m.home); err != nil {", "if err := held.Validate(m.home); false && err != nil {", "projectlease", "TestAuthorityRequiresHeldWitness", "zero witness invoked callbacks"},
		{"omit-empty-home-creation", "internal/projectlease/manager.go", "if err = privatedir.MakeAll(projectHome); err != nil {", "if err = error(nil); err != nil {", "projectlease", "TestEmptyGeneratedHome", "empty generated publication:"},
		{"ignore-missing-empty-home", "internal/projectlease/manager.go", "if err := m.verifyHome(c); err != nil {", "if err := m.verifyHome(c); false && err != nil {", "projectlease", "TestEmptyGeneratedHome", "missing empty home admitted resume"},
	}
	controls := map[string]bool{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "ignore-mode-drift" && os.PathSeparator == '\\' {
				t.Skip("native receipt mode bits are Unix-specific")
			}
			run := func(overlay string) ([]byte, error) {
				pkgPath := "./internal/" + tc.pkg
				if tc.pkg == "cmd/curator" {
					pkgPath = "./cmd/curator"
				}
				deadline := "90s"
				if runtime.GOOS == "windows" {
					// The positive recovery matrix passed in 65s in the full
					// stream; its nested control hit 90s during a durable rename.
					// Timeout or panic still cannot qualify as an intended kill.
					deadline = "5m"
				}
				args := []string{"test", "-count=1", "-timeout=" + deadline, "-v", pkgPath, "-run", "^" + tc.run + "$"}
				if overlay != "" {
					args = append([]string{"test", "-overlay", overlay}, args[1:]...)
				}
				cmd := exec.Command(goTool, args...)
				cmd.Dir = root
				for _, kv := range os.Environ() {
					if !strings.HasPrefix(kv, "GIT_DIR=") && !strings.HasPrefix(kv, "GIT_WORK_TREE=") {
						cmd.Env = append(cmd.Env, kv)
					}
				}
				if tc.pkg == "cmd/curator" {
					// All CLI witnesses here recover synthetic journals without
					// fingerprinting or mutating the host toolchain. Helper mode
					// keeps its lock from consuming either control's watchdog.
					cmd.Env = append(cmd.Env, "CURATOR_TEST_CLI_HELPER=1")
				}
				return cmd.CombinedOutput()
			}
			key := tc.pkg + "/" + tc.run
			if !controls[key] {
				out, err := run("")
				if err != nil || !strings.Contains(string(out), "--- PASS: "+tc.run) {
					t.Fatalf("positive control failed: %v\n%s", err, out)
				}
				controls[key] = true
			}
			path := filepath.Join(root, tc.file)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.ReplaceAll(string(raw), "\r\n", "\n")
			if strings.Count(source, tc.before) != 1 {
				t.Fatal("mutation anchor not unique")
			}
			scratch := t.TempDir()
			replacement := filepath.Join(scratch, "replacement.go")
			if err = os.WriteFile(replacement, []byte(strings.Replace(source, tc.before, tc.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			raw, err = json.Marshal(map[string]any{"Replace": map[string]string{path: replacement}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(scratch, "overlay.json")
			if err = os.WriteFile(overlay, raw, 0600); err != nil {
				t.Fatal(err)
			}
			out, err := run(overlay)
			exit, ok := err.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(out), "--- FAIL: "+tc.run) || !strings.Contains(string(out), tc.message) || strings.Contains(string(out), "panic:") {
				t.Fatalf("mutant survived or failed for another reason: %v\n%s", err, out)
			}
			t.Logf("intended expected-red exit 1:\n%s", out)
		})
	}
}
