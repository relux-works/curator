# TASK-261015-258cwf developer handoff

Candidate remains uncommitted on the managed Story branch. Base: `3f5de6dff9aab4c037ba069323604b166bf0414e`; fetched origin/main and local main matched that base before implementation (0 commits behind).

## Implementation and brief accounting

1. Full `env config show` retains the environments knobs and raw permissions map, adding `permissions_posture` keyed by installed profiles. No adapter or provisioning is needed. Both interactive and machine use retain the command's existing JSON form; the command has no separate human renderer or --json flag.
2. `env config show permissions` prints the machine-current profile's `{mode, locked, source}`. Without a current profile it refuses with an unavailable diagnostic. Failed inventory/current reads refuse as unreadable, with no posture output. Configuration load errors also refuse before output, preserving their existing diagnostics.
3. `envprofile.PermissionsForProfile` is the single semantic implementation used by `buildFragment` and config show. Global lock always means native/global/true; explicit entries are profile/false; absence is native/default/false. CLI tests use production `run`, including real file-backed machine/system reads. Additional property coverage enumerates lock state, absent/native/yolo entries, and configured/absent profile identities.
4. docs/cli.md, docs/environment-config.md and Unreleased Added line updated. README Tools now documents the existing lint command and local log location as required by repository instructions.

The new envprofile helper/tests, its buildFragment caller, envfragment JSON tags, and CLI test file are necessary supporting files for the brief's shared-code and production-test requirements. No unrelated source changes. This is an initial implementation, not rework. Previously committed reviewer tests are retained under their original names.

## Coverage map and bounds

Brief gap: the supplied brief contains no attack-surface table. This report supplies behavioral accounting, but does not claim catalog-complete attack coverage or silently waive the missing-table precondition.

Behavioral AC coverage: **7 of 7 rows driven** (named candidate tests listed below).

| AC row | Named test | Production call site |
| --- | --- | --- |
| Global native/global/true for every installed profile, including absence from system map | TestEnvConfigShowPermissionsAtProductionEntry/global | run -> cmdEnv -> cmdEnvConfigShow -> InstalledPermissions -> PermissionsForProfile |
| Profile yolo/profile/false and configured-map stability | TestEnvConfigShowPermissionsAtProductionEntry/profile | same full-show entry |
| Absent entry native/default/false; uninstalled profile excluded | TestEnvConfigShowPermissionsAtProductionEntry/profile | same full-show entry |
| Current-profile single-knob posture | TestEnvConfigShowPermissionsAtProductionEntry | run -> cmdEnvConfigShow -> Current -> PermissionsForProfile |
| Machine/system configuration read failure emits no guessed posture | TestEnvConfigShowPermissionsRejectsUnreadableConfig/machine and /system | run -> cmdEnv -> loadConfig -> config.Load |
| Inventory/current read failure emits no guessed posture | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent | run -> cmdEnvConfigShow -> InstalledPermissions or Current |
| Empty installation stays read-only; missing current is unavailable | TestEnvConfigShowPermissionsWithoutProfiles | run -> cmdEnvConfigShow |

Fragment compatibility uses the retained TestEnvResolveEmitsPermissionsV2AtProductionEntry (four lattice cases) through run -> cmdEnvResolve -> envprofile.Resolve -> buildFragment -> PermissionsForProfile. The exact fragment-object tests retain their original names.

Out-of-contract rows: separate --json flag and human-only renderer, per brief item 1's explicit allowance for the command's existing JSON form; runtime launch/application, because this task only reports posture. No other supplied behavior row is excluded. Read-failure fixtures use deterministic directory/file shape errors, rather than relying on chmod under a privileged runner. The suite does not claim every OS error string or runtime-launch behavior. No source-text gate is introduced, so token-preserving source-checker mutation is inapplicable.

## Logbook

- Reused the existing lock-free installedProfiles inventory rather than ListWithPolicy: the latter may ensure/provision the default profile. Config show remains a read-only command.
- CLI TestMain holds the host GOROOT lock for the entire package. Initial parallel mutant attempts queued behind the broad suite; these were terminated (exit 143) and are unrun evidence. Mutants must be replayed sequentially after that suite releases the lock.
- The first requested development transition failed (exit 1) because estimate was absent. Setting estimate 3 and retrying development each exited 0.

## Validation

Native CLI target: darwin/arm64, Go 1.25.5; golangci-lint 2.12.2 built with Go 1.25.5. No externally supplied conformance root; tests use repository fixture/default selection. Relevant existing tests and all new tests run locally; this is not hosted cross-platform matrix evidence.

| Command | Exit | Result |
| --- | ---: | --- |
| go build ./... | 0 | Repository compiles |
| make lint | 0 | Both initial and updated-test candidate runs green |
| go test ./internal/envprofile -run '^TestPermissionsForProfileProperties$' -count=1 | 0 | All 12 permission-policy combinations green |
| go test ./cmd/curator -run '^TestEnvConfigShowPermissions' -count=1 (initial tests) | 0 | Initial CLI behavior green; updated tests rerun separately below |
| git diff --check | 0 | Whitespace clean |
| go test ./cmd/curator/... ./internal/envprofile/... ./internal/envfragment/... -timeout=8m | 143 | Not green: cmd child failed status checks from ambient installed provider and timed out at 480s; top-level process terminated after the timeout. Other package output missing, so not accepted as evidence |
| go test ./cmd/curator -run '^Test(Env\|Profile)' -timeout=8m -count=1 (ambient PATH attempt) | 143 | Stopped to remove the evidenced ambient provider pollution; unrun |

The broad run's failures named TestDraftStatusReportsUpToDateThroughCLI, TestDraftStatusNeedsInstallAfterRefreshWithoutInstall, TestDraftStatusLiveMutationWithoutRefreshStaysCurrent, TestEnvStatusMatrix, TestEnvStatusRegistryBoundaryPostureAndCheck, TestEnvStatusCheckCurrentScopeOnly, TestEnvStatusMissingAndUnreadableKeepRecord, TestEnvStatusUnreadableApprovalStateSurfaced, TestEnvStatusReportsShellHookTrustPosture, and TestLegacyPackageStatusBindingAndRepair. Status reported the ambient curator-engines executable as subcommand_provider_untrusted. The timeout named TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt. Those results do not establish a repository regression or a passing full suite. The environment/profile rerun uses the same toolchain with the user binary directory excluded from PATH; no production trust policy is weakened.

Full repository and hosted landing/race matrix are not claimed here. The orchestrator owns integration and the configured hosted landing gate. Raw logs stay task-scoped under `.temp/TASK-261015-258cwf/`.

## Incremental validation and mutants

The broad environment/profile candidate suite (`PATH=<Go toolchain bin>:<package-manager bin>:<system bins> go test ./cmd/curator -run '^Test(Env|Profile)' -timeout=8m -count=1`) exited **0**, reporting 336.035s. The final test delta strengthened the exact installed-profile assertion (default, acme, other; uninstalled excluded), checked the empty-installation refusal, and gave unreadable-config cases an installed current profile. Production source was unchanged after the broad run. The complete changed CLI tests plus the retained fragment schema-case test were rerun with `go test ./cmd/curator -run '^Test(EnvConfigShowPermissions|CuratorSpecMainLaunchEnvFragmentV2SchemaCases)' -count=1 -v`: **exit 0**, 2.015s. This reuses the broad unchanged-test evidence and directly reruns the changed test scope. `make lint` after these test changes exited **0**; `git diff --check` exited **0**.

The baseline Go overlay restores every changed tracked Go file to HEAD and removes the newly added Go files from the build, without touching the worktree. `go test -overlay .temp/TASK-261015-258cwf/baseline.json ./cmd/curator -run '^TestDraftStatusReportsUpToDateThroughCLI$' -count=1 -timeout=2m` exited **1** with ambient PATH (same pre-existing status failure); with the clean PATH it exited **0**, 1.173s. This demonstrates the ambient-provider baseline failure rather than declaring it unrelated by inference. No full-suite success is inferred from that one baseline check.

Each mutant below retains the gate and weakens exactly the stated boundary. They use Go build overlays, leaving candidate files untouched, and run the behavioral production-entry suite. All listed mutant commands exited **1**, as expected-red gates; they are failing commands, not passing suites. No mutant survived.

| Mutant | What it narrows the gate to | Named failing test | Exit / surviving bound |
| --- | --- | --- | --- |
| global-lock | Honor a global lock only for profiles explicitly in the system permissions map; admit absent-profile unlocked fallback | TestEnvConfigShowPermissionsAtProductionEntry/global | 1; killed |
| system-read-failure | Preserve machine reads and system overlay, but treat a failed system read as an empty schema-2 system object; installed profiles become native/default/unlocked | TestEnvConfigShowPermissionsRejectsUnreadableConfig/system | 1; killed |
| inventory-read-failure | Preserve inventory enumeration but convert only its error to an empty successful posture map | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent | 1; killed |
| current-read-failure | Preserve current-pointer read/check but convert only its error to native/default/unlocked output | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent | 1; killed |

Commands (all with clean PATH, each directly executed):

```sh
go test -overlay .temp/TASK-261015-258cwf/global-lock.json ./cmd/curator -run '^TestEnvConfigShowPermissionsAtProductionEntry$' -count=1
go test -overlay .temp/TASK-261015-258cwf/system-read-failure.json ./cmd/curator -run '^TestEnvConfigShowPermissionsRejectsUnreadableConfig$' -count=1
go test -overlay .temp/TASK-261015-258cwf/inventory-read-failure.json ./cmd/curator -run '^TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent$' -count=1
go test -overlay .temp/TASK-261015-258cwf/current-read-failure.json ./cmd/curator -run '^TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent$' -count=1
```

The system-read mutant was first killed before the installed-profile fixture was strengthened (exit 1), then replayed against that stronger fixture (exit 1). Initial parallel overlay attempts were cancelled with exit 143 and do not count as mutant kills.

The second library attempt lacked any completed package output after its child test binary exited; it was terminated (exit 143) rather than counted green. A diagnostic SIGQUIT request produced no output before termination. The fresh verbose `-count=1` attempt below supplies accepted evidence; no cache explanation is asserted without evidence.

## Candidate file identity

```json
{
  "CHANGELOG.md": "ebf0c2c76fa716ffcba5d49ce178285733337b1bfa0ba5ff1aa769434e504b75",
  "README.md": "17bedec1e22ceecba14a2e532cc8b69f89c1a78ffee132a16a78171825185356",
  "cmd/curator/envconfig.go": "8ad4100aeea9b00d69ede53c505d4f3ee9b7a12d4eec43a45649d36f691f3882",
  "cmd/curator/envconfig_permissions_test.go": "6a3d8d01e0b23f230ee8d072050c12da38f314359c6457c11fa53426fc35fafc",
  "docs/cli.md": "a3650b707261233d5e0138eed9704a436eaa026743e3c5c7eb5143dec780a638",
  "docs/environment-config.md": "95be70aac3809a0c78126c84644e426cb44303b43f4fdaae6182be54898a7e15",
  "internal/envfragment/envfragment.go": "c6df17cc43d751c9c7c6299ecf87e12adfeb28e89953af29534025878b3244b9",
  "internal/envprofile/managed.go": "b48eae6bf1c25b527160915954954a17925d6250e5f21f88e74e0b6ebb92a9e8",
  "internal/envprofile/permissions.go": "3cd06856b3f6b2f47078e757ea20326dc63e4d5c8ecb370d64c6813e809ab350",
  "internal/envprofile/permissions_test.go": "2f198dc708046acbac42328151c7d16fc15335f6acca51fed29290579e3d980a"
}
```

## Reproducible narrowing deltas

Apply each independently to the candidate via a Go overlay.

### global-lock

```diff
--- internal/envprofile/permissions.go
+++ internal/envprofile/permissions.go
@@ -8,7 +8,7 @@
 // PermissionsForProfile resolves the fragment permission policy (§10.2).
 // A global lock forces native even for a profile absent from the system map.
 func PermissionsForProfile(machine envregistry.MachineConfig, profile string) envfragment.Permissions {
-	if machine.PermissionsLocked {
+	if machine.PermissionsLocked && machine.Permissions[profile] != "" {
 		return envfragment.Permissions{Mode: "native", Locked: true, Source: "global"}
 	}
 	if mode, configured := machine.Permissions[profile]; configured {
```

### system-read-failure

```diff
--- internal/config/config.go
+++ internal/config/config.go
@@ -309,7 +309,7 @@
 		systemPath = resolved
 		systemData, err := readObject(systemPath)
 		if err != nil {
-			return nil, err
+            systemData = map[string]any{"schema_version": float64(2)}
 		}
 		systemDataForSources = systemData
 		merged, mergedLocked, err := applySystem(systemData, userData, systemPath, warn)
```

### inventory-read-failure

```diff
--- internal/envprofile/permissions.go
+++ internal/envprofile/permissions.go
@@ -22,7 +22,7 @@
 func InstalledPermissions(home string, machine envregistry.MachineConfig) (map[string]envfragment.Permissions, error) {
 	profiles, err := installedProfiles(home)
 	if err != nil {
-		return nil, err
+        return map[string]envfragment.Permissions{}, nil
 	}
 	result := make(map[string]envfragment.Permissions, len(profiles))
 	for _, profile := range profiles {
```

### current-read-failure

```diff
--- cmd/curator/envconfig.go
+++ cmd/curator/envconfig.go
@@ -52,8 +52,7 @@
 	if args[0] == "permissions" {
 		profile, err := envprofile.Current(cfg.Home())
 		if err != nil {
-			_, _ = fmt.Fprintln(c.stderr, "curator: permissions posture unreadable:", err)
-			return exitFail
+            return c.printJSON(envprofile.PermissionsForProfile(machineFromConfig(cfg), ""))
 		}
 		if profile == "" {
 			_, _ = fmt.Fprintln(c.stderr, "curator: permissions posture unavailable: no current profile")
```

## Changed CLI test output

```text
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-absent.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-global-unlocked.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-locked-profile-source.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-missing-mode.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-unknown-field.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-unknown-mode.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-unknown-source.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-yolo-locked.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-yolo-silent.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/valid-permissions-native-locked.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/valid-permissions-native-silent.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/valid-permissions-yolo-unlocked.json
=== RUN   TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/valid.json
--- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases (0.01s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-absent.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-global-unlocked.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-locked-profile-source.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-missing-mode.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-unknown-field.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-unknown-mode.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-unknown-source.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-yolo-locked.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid-permissions-yolo-silent.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/invalid.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/valid-permissions-native-locked.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/valid-permissions-native-silent.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/valid-permissions-yolo-unlocked.json (0.00s)
    --- PASS: TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases/valid.json (0.00s)
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry/profile
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry/global
--- PASS: TestEnvConfigShowPermissionsAtProductionEntry (1.22s)
    --- PASS: TestEnvConfigShowPermissionsAtProductionEntry/profile (0.61s)
    --- PASS: TestEnvConfigShowPermissionsAtProductionEntry/global (0.61s)
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig/machine
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig/system
--- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig (0.42s)
    --- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig/machine (0.00s)
    --- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig/system (0.00s)
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent
--- PASS: TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent (0.00s)
=== RUN   TestEnvConfigShowPermissionsWithoutProfiles
--- PASS: TestEnvConfigShowPermissionsWithoutProfiles (0.00s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	2.015s
```

Full fragment package: `go test ./internal/envfragment/... -count=1 -timeout=2m` with clean PATH exited **0**, 0.272s. This separately establishes the unchanged fragment-byte/schema tests.

## Library result and review handoff

`PATH=<Go toolchain bin>:<package-manager bin>:<system bins> go test ./internal/envprofile/... ./internal/envfragment/... -count=1 -timeout=8m -v` exited **0**: envprofile 247.941s; envfragment 0.263s. TestPermissionsForProfileProperties is present under its original name and green. The verbose log contains 19 broader conditional skips: 18 require CURATOR_CONFORMANCE_ROOT, which is unset, and TestFoldedPathsAreSourceInvalid requires a case-sensitive filesystem. These do not cover the new permissions rows; none of the seven task AC rows or new named CLI/property tests was skipped. Full published-root conformance remains a stated integration bound, not a claimed green gate. The successful no-cache run establishes its own result; it does not prove a cache root cause for the earlier stalled attempt.

All changed production code was built; all changed tests were run after their last edits; repository lint and whitespace checks passed. The original permissions fragment CLI tests and the exact fragment package tests remain intact and green. No prior leaf rework/reviewer-test removal occurred. Final HEAD remains the recorded base and every candidate change remains uncommitted for the Change Request snapshot. Status requested: ready for review, not accepted integration.
