# TASK-261015-258cwf successor developer handoff

Candidate stays uncommitted at base `3f5de6dff9aab4c037ba069323604b166bf0414e` for the managed Change Request flow. No branch, index or commit changes and no direct control-root edits were made by this successor. All source, tests and docs remain in the managed worktree.

## Behavior and brief accounting

1. Full `env config show` retains existing knob fields and the configured permissions map. It adds `permissions_posture` keyed by installed profile. Global/native/locked, profile/unlocked and native/default/unlocked are derived by `envprofile.PermissionsForProfile`, also called by fragment `buildFragment`.
2. `env config show permissions` uses the same resolver for the machine-current profile. Missing current is unavailable. Config, inventory and current-pointer read failures exit 1 with a diagnostic and no guessed posture. Inventory metadata must establish a directory before its listing can establish an empty installation.
3. Tests drive production `run`; machine/system failure tests use real file-backed configuration. Existing reviewer/fragment tests retain their original names. The successor retains the Windows-failing regression and adds `TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor` for failed metadata inspection.
4. CLI docs, environment-config docs and Unreleased Added entry are updated. README documents the repository lint entry point and log location. Both interactive and machine use keep the command's existing JSON renderer; there is no separate --json flag.

The leaf's ten candidate files are the CLI implementation/test, shared envprofile resolver/test and fragment caller, envfragment JSON field tags, two documentation files, README and CHANGELOG. Supporting files implement the brief's shared-code and production-test requirements. Successor rework is bounded to `internal/envprofile/permissions.go` and `cmd/curator/envconfig_permissions_test.go`; no out-of-module rework delta.

## Logbook: recovered gate failure

The prior candidate's configured gate exited 1. Hosted run https://github.com/relux-works/curator/actions/runs/38680796445 reported the Windows Test job failing; all other executed jobs reported success. Inspection of its downloaded raw `go-test.json` identifies exactly one failing test: `TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent`. It returned successful empty `permissions_posture` when `profiles` was a regular file. Every other permissions test in that stream passed. This is a repository behavior failure, not an external-CI failure or a reason to weaken the gate.

All ten candidate hashes initially matched the predecessor's recorded identity. This successor added a domain type check before inventory enumeration; the existing test now rejects the bad inventory independently of directory-enumeration behavior. A second production-entry regression exercises a blocked ancestor and metadata failure. No generic state-reader contract or runtime launch behavior was changed.

## Coverage map and bounds

Brief gap: no surface table was supplied. No catalog-complete attack-surface claim is made. Behavioral AC accounting: **8 of 8 rows driven**.

| AC row | Named test | Production call site |
| --- | --- | --- |
| Global native/global/true for every installed profile, including absent system-map entry | TestEnvConfigShowPermissionsAtProductionEntry/global | run -> cmdEnv -> cmdEnvConfigShow -> InstalledPermissions -> PermissionsForProfile |
| Profile yolo/profile/false and raw configured-map stability | TestEnvConfigShowPermissionsAtProductionEntry/profile | same full-show entry |
| Missing entry native/default/false; uninstalled entry excluded | TestEnvConfigShowPermissionsAtProductionEntry/profile | same full-show entry |
| Current-profile single-knob output | TestEnvConfigShowPermissionsAtProductionEntry | run -> cmdEnvConfigShow -> Current -> PermissionsForProfile |
| Machine/system config read failure | TestEnvConfigShowPermissionsRejectsUnreadableConfig/machine and /system | run -> cmdEnv -> loadConfig -> config.Load |
| Invalid inventory type/current-pointer failure | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent | run -> cmdEnvConfigShow -> InstalledPermissions / Current |
| Failed inventory metadata/current ancestor read | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor | run -> cmdEnvConfigShow -> InstalledPermissions -> stateread.Stat; Current |
| Empty installation remains read-only and missing current refuses | TestEnvConfigShowPermissionsWithoutProfiles | run -> cmdEnvConfigShow |

Retained fragment coverage: `TestEnvResolveEmitsPermissionsV2AtProductionEntry` drives run -> cmdEnvResolve -> envprofile.Resolve -> buildFragment -> PermissionsForProfile across the four permissions lattice cases. `TestCuratorSpecMainLaunchEnvFragmentV2SchemaCases` retains the exact schema cases. `TestPermissionsForProfileProperties` enumerates 12 combinations and checks the lock/source/mode invariant.

Out-of-contract rows: separate human renderer/--json flag, allowed by brief item 1's explicit existing-JSON-form option; runtime launch/application, outside the reporting behavior described in brief items 1–2. No other supplied AC row is excluded. Fixtures cover deterministic type/ancestor failures rather than all OS error codes or concurrent filesystem replacements. Published conformance-root tests and the hosted platform matrix are separate bounds; no full current-candidate matrix pass is claimed before the configured Change Request rerun. No source-text gate is introduced.

## Candidate checks run by this successor

Local platform is darwin/arm64, Go 1.25.5, golangci-lint 2.12.2. Broader tests exclude the previously diagnosed ambient provider binary directory from PATH, preserving the same Go/package-manager/system tools. CURATOR_CONFORMANCE_ROOT is unset. Relevant conditional external-root tests retain their normal skips; they do not cover any newly claimed row.

| Command | Exit | Result |
| --- | ---: | --- |
| `go build ./...` | 0 | Repository builds after source repair |
| `make lint` | 0 | 0 issues, including rerun after the added ancestor test |
| `gofmt -d internal/envprofile/permissions.go cmd/curator/envconfig_permissions_test.go` | 0 | Empty formatting diff |
| `git diff --check` | 0 | Whitespace clean |
| `go test ./internal/envprofile/... ./internal/envfragment/... -count=1 -timeout=8m` | 0 | Full touched library packages; 289.651s and 0.260s |
| `go test ./cmd/curator -run '^Test(Env\|Profile\|CuratorSpecMainLaunchEnvFragmentV2SchemaCases)' -timeout=8m -count=1` | 0 | Environment/profile/schema regression scope, 320.927s |
| `go test ./cmd/curator -run '^TestEnvConfigShowPermissions' -count=1 -v -timeout=2m` | 0 | Every new permissions entry test, after all test edits |

Every command ran directly, with its real exit status preserved. Full CLI package/repository and current Windows execution are left to the configured hosted Change Request gate; the local CLI command intentionally selects the relevant command scopes. No whole-repository local pass is claimed.


Prior evidence reused only where source/test/config identity is unchanged: historical exact fragment/schema behavior and the initial failure investigation. All impacted behavior and all changed tests are rerun here. The predecessor's cancelled/truncated broad attempts remain unrun evidence; they are not promoted to green. The current full hosted gate is owned and rerun automatically by the Change Request flow after role handoff.

## Narrowing mutants on the successor source

All executed overlays keep the candidate untouched and execute behavioral production-entry tests. Expected-red commands exit 1 and are reported as failures, never as green validation. No executed mutant survived.

| Mutant | What it narrows the gate to | Named failing test | Exit / survivor bound |
| --- | --- | --- | --- |
| global-lock | Honor a global lock only when the profile appears in the system map, admitting absent-profile unlocked output | TestEnvConfigShowPermissionsAtProductionEntry/global | 1; killed |
| system-read | Convert only failed system-config reads to an empty schema-2 system object, admitting default/unlocked output | TestEnvConfigShowPermissionsRejectsUnreadableConfig/system | 1; killed |
| current-read | Convert only failed current-pointer reads to native/default/unlocked output | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent | 1; killed |
| inventory-type | Keep inventory metadata/type inspection but admit its known non-directory case as empty successful posture | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent | 1; killed |
| inventory-metadata | Keep inventory metadata inspection but admit its failed-read case as empty successful posture | TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor | 1; killed |

The predecessor's historical inventory-enumeration-error mutant kill belongs only to its source identity. It is not counted as successor evidence. The five executed mutants above establish the claimed successor bounds.

### global-lock

```diff
--- internal/envprofile/permissions.go
+++ internal/envprofile/permissions.go
@@ -11,7 +11,7 @@
 // PermissionsForProfile resolves the fragment permission policy (§10.2).
 // A global lock forces native even for a profile absent from the system map.
 func PermissionsForProfile(machine envregistry.MachineConfig, profile string) envfragment.Permissions {
-	if machine.PermissionsLocked {
+	if machine.PermissionsLocked && machine.Permissions[profile] != "" {
 		return envfragment.Permissions{Mode: "native", Locked: true, Source: "global"}
 	}
 	if mode, configured := machine.Permissions[profile]; configured {
```

### system-read

```diff
--- internal/config/config.go
+++ internal/config/config.go
@@ -309,7 +309,7 @@
 		systemPath = resolved
 		systemData, err := readObject(systemPath)
 		if err != nil {
-			return nil, err
+			systemData = map[string]any{"schema_version": float64(2)}
 		}
 		systemDataForSources = systemData
 		merged, mergedLocked, err := applySystem(systemData, userData, systemPath, warn)
```

### current-read

```diff
--- cmd/curator/envconfig.go
+++ cmd/curator/envconfig.go
@@ -52,8 +52,7 @@
 	if args[0] == "permissions" {
 		profile, err := envprofile.Current(cfg.Home())
 		if err != nil {
-			_, _ = fmt.Fprintln(c.stderr, "curator: permissions posture unreadable:", err)
-			return exitFail
+			return c.printJSON(envprofile.PermissionsForProfile(machineFromConfig(cfg), ""))
 		}
 		if profile == "" {
 			_, _ = fmt.Fprintln(c.stderr, "curator: permissions posture unavailable: no current profile")
```

### inventory-type

```diff
--- internal/envprofile/permissions.go
+++ internal/envprofile/permissions.go
@@ -1,8 +1,6 @@
 package envprofile
 
 import (
-	"fmt"
-
 	"github.com/relux-works/curator/internal/envfragment"
 	"github.com/relux-works/curator/internal/envregistry"
 	"github.com/relux-works/curator/internal/stateread"
@@ -31,7 +29,7 @@
 		return nil, err
 	}
 	if metadata.Kind == stateread.KindPresent && !metadata.Info.IsDir() {
-		return nil, stateread.UnusableError(inventory, fmt.Errorf("profile inventory must be a directory"))
+		return map[string]envfragment.Permissions{}, nil
 	}
 	profiles, err := installedProfiles(home)
 	if err != nil {
```

### inventory-metadata

```diff
--- internal/envprofile/permissions.go
+++ internal/envprofile/permissions.go
@@ -28,7 +28,7 @@
 	inventory := ProfilesDir(home)
 	metadata, err := stateread.Stat(inventory)
 	if err != nil {
-		return nil, err
+		return map[string]envfragment.Permissions{}, nil
 	}
 	if metadata.Kind == stateread.KindPresent && !metadata.Info.IsDir() {
 		return nil, stateread.UnusableError(inventory, fmt.Errorf("profile inventory must be a directory"))
```


## Candidate file identity

```json
{
  "CHANGELOG.md": "ebf0c2c76fa716ffcba5d49ce178285733337b1bfa0ba5ff1aa769434e504b75",
  "README.md": "17bedec1e22ceecba14a2e532cc8b69f89c1a78ffee132a16a78171825185356",
  "cmd/curator/envconfig.go": "8ad4100aeea9b00d69ede53c505d4f3ee9b7a12d4eec43a45649d36f691f3882",
  "cmd/curator/envconfig_permissions_test.go": "b986580d0564689858992d84e30e19963a54c15ff79e638aa5e16e0d687d4181",
  "docs/cli.md": "a3650b707261233d5e0138eed9704a436eaa026743e3c5c7eb5143dec780a638",
  "docs/environment-config.md": "95be70aac3809a0c78126c84644e426cb44303b43f4fdaae6182be54898a7e15",
  "internal/envfragment/envfragment.go": "c6df17cc43d751c9c7c6299ecf87e12adfeb28e89953af29534025878b3244b9",
  "internal/envprofile/managed.go": "b48eae6bf1c25b527160915954954a17925d6250e5f21f88e74e0b6ebb92a9e8",
  "internal/envprofile/permissions.go": "be96047869eb43f0e412384e3e25dbfd5faa7db90d32079b9b5ce6ada6723b6d",
  "internal/envprofile/permissions_test.go": "2f198dc708046acbac42328151c7d16fc15335f6acca51fed29290579e3d980a"
}
```

## Focused entry-test output

```text
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry/profile
=== RUN   TestEnvConfigShowPermissionsAtProductionEntry/global
--- PASS: TestEnvConfigShowPermissionsAtProductionEntry (1.57s)
    --- PASS: TestEnvConfigShowPermissionsAtProductionEntry/profile (0.74s)
    --- PASS: TestEnvConfigShowPermissionsAtProductionEntry/global (0.83s)
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig/machine
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableConfig/system
--- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig (0.50s)
    --- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig/machine (0.00s)
    --- PASS: TestEnvConfigShowPermissionsRejectsUnreadableConfig/system (0.01s)
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent
--- PASS: TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent (0.01s)
=== RUN   TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor
--- PASS: TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor (0.00s)
=== RUN   TestEnvConfigShowPermissionsWithoutProfiles
--- PASS: TestEnvConfigShowPermissionsWithoutProfiles (0.00s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	2.625s
```

## Mutant commands

Each command below exited **1** (expected-red). All use `-count=1 -timeout=2m` and the clean test PATH described above.

```sh
go test -overlay .temp/TASK-261015-258cwf/successor-global-lock.json ./cmd/curator -run '^TestEnvConfigShowPermissionsAtProductionEntry$' -count=1 -timeout=2m
go test -overlay .temp/TASK-261015-258cwf/successor-system-read.json ./cmd/curator -run '^TestEnvConfigShowPermissionsRejectsUnreadableConfig$' -count=1 -timeout=2m
go test -overlay .temp/TASK-261015-258cwf/successor-current-read.json ./cmd/curator -run '^TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent$' -count=1 -timeout=2m
go test -overlay .temp/TASK-261015-258cwf/successor-inventory-type.json ./cmd/curator -run '^TestEnvConfigShowPermissionsRejectsUnreadableInventoryAndCurrent$' -count=1 -timeout=2m
go test -overlay .temp/TASK-261015-258cwf/successor-inventory-metadata.json ./cmd/curator -run '^TestEnvConfigShowPermissionsRejectsUnreadableInventoryAncestor$' -count=1 -timeout=2m
```
