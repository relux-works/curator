package projectlease

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
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
		{"gc-exact-home-witness", "internal/projectlease/manager.go", "if err := witness.Held().Validate(home); err != nil {", "if err := witness.Held().Validate(home); false && err != nil {", "scopes", "TestCollectProjectLeaseForeignLockInterleaving", "foreign GC interleaving:"},
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
				args := []string{"test", "-count=1", "-timeout=90s", "-v", "./internal/" + tc.pkg, "-run", "^" + tc.run + "$"}
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
