# TASK-260924-2y5w1w results

## Implementation

- Added `PermissionMode`, `ToolRelease`, and `NativeArgs` to `vendorplugin.SpawnRequest`. The vendor receives a private copy of `NativeArgs`; the permission members are projected through the shared passthrough request and all three are checked after `Vendor.Spawn` before `agentic.BuildPlanWithEnvironment` runs.
- Added `agentic.StoredPolicyInspection` and capability dispatch from an admitted plan. Claude and Codex inspectors use only the supplied plan environment, home, and work directory. Pi reports `unsupported`.
- Claude reads the selected user `settings.json`, the work directory's `.claude/settings.json`, and `.claude/settings.local.json`. Codex reads the selected user `config.toml`, the work directory's `.codex/config.toml`, and the selected profile file under the Codex config root. Relaxations retain selector, value, and source path. Absent, unreadable, unparseable, unavailable-root, and unknown-profile sources are reported separately and never treated as clean. Managed policy outside the launch inputs is reported as not provided.
- Documented inspected sources and bounds in README's **Permission-mode** section. Added one bullet under `## Unreleased`; released changelog entries were not changed. Release v0.5.21 remains for the orchestrator to cut.

## Selector × source coverage

Each row below is a production inspector test with one known selector placed in one source. Each plugin covers 3 selectors × 3 sources = 9 rows; total: 18/18 rows.

| Plugin | Source | Selector | Fixture relaxation |
|---|---|---|---|
| Claude | User settings | `permissions.defaultMode` | `bypassPermissions` |
| Claude | User settings | `permissions.allow` | `Bash(git *)` |
| Claude | User settings | `permissions.additionalDirectories` | `/shared` |
| Claude | Project shared settings | `permissions.defaultMode` | `bypassPermissions` |
| Claude | Project shared settings | `permissions.allow` | `Bash(git *)` |
| Claude | Project shared settings | `permissions.additionalDirectories` | `/shared` |
| Claude | Project local settings | `permissions.defaultMode` | `bypassPermissions` |
| Claude | Project local settings | `permissions.allow` | `Bash(git *)` |
| Claude | Project local settings | `permissions.additionalDirectories` | `/shared` |
| Codex | User config | `approval_policy` | `never` |
| Codex | User config | `sandbox_mode` | `danger-full-access` |
| Codex | User config | `sandbox_permissions` | `["disk-write-access"]` |
| Codex | Project config | `approval_policy` | `never` |
| Codex | Project config | `sandbox_mode` | `danger-full-access` |
| Codex | Project config | `sandbox_permissions` | `["disk-write-access"]` |
| Codex | Selected profile | `approval_policy` | `never` |
| Codex | Selected profile | `sandbox_mode` | `danger-full-access` |
| Codex | Selected profile | `sandbox_permissions` | `["disk-write-access"]` |

Separate tests cover Claude relaxed default modes, restrictive Codex values, absent files, malformed files, and POSIX mode-000 unreadable files. The tests use temporary fixture trees only; no real provider settings were read. The mode-000 tests ran as uid 501 on macOS and were not skipped.

## Vendor admission coverage

Tests call `vendorplugin.BuildLaunchWithEnvironment` with a fake vendor and a fake `claude` executable. The test rows prove:

| Request | Result |
|---|---|
| Native mode, unpinned release, raw native args | Fields reach `Vendor.Spawn`; native args remain verbatim in the plan |
| Yolo mode, verified release | The pinned bypass mapping appears before native args |
| Yolo plus `--permission-mode plan` | Refused with `ErrNativePolicyConflict` after vendor admission |
| Yolo plus release `2.1.274` | Refused with `ErrPermissionModeUnverifiedRelease` after vendor admission |
| Vendor changes permission mode, release, or native args | Refused with `vendorplugin.ErrVendorContract` |
| Vendor mutates its received `NativeArgs` slice | Refused; the caller's slice remains unchanged |

## Narrowing mutants

All mutants ran in a disposable copy under `.temp/TASK-260924-2y5w1w-mutants`, which was removed after the runs. Each non-zero result below is expected and means the test suite killed the mutant.

| Mutant | Direct test command | Exit | Evidence |
|---|---|---:|---|
| Claude and Codex mark an unreadable mode-000 settings file inspected/clean | `go test ./pkg/agentic/systems/claude ./pkg/agentic/systems/codex -run TestStoredPolicyInspectorReportsUnreadableMode000 -count=1` | 1 | Both unreadable-source tests failed on the path appearing in `SourcesInspected`. |
| Claude drops `permissions.additionalDirectories` | `go test ./pkg/agentic/systems/claude -run '^TestStoredPolicyInspectorReportsEveryKnownSelectorAtEveryLocalSource$' -count=1` | 1 | All three source rows for that selector failed to find their relaxation. |
| Vendor fidelity checks for the three permission members are removed | `go test ./pkg/agentic/systems/claude -run '^TestBuildLaunchWithEnvironmentRefusesVendorPermissionRequestMutation$' -count=1` | 1 | The mode downgrade, release rewrite, and native-argument drop subtests each received no contract refusal. |

## Validation history

| Command | Exit | Result |
|---|---:|---|
| `go test ./pkg/vendorplugin ./pkg/agentic ./pkg/agentic/systems/claude ./pkg/agentic/systems/codex ./pkg/agentic/systems/pi` | 1, then 0 | First run found the Pi fixture's missing reader argument and the Codex argv guard catching a sandbox-value spelling. Both were fixed; the direct rerun passed. |
| `go test ./...` | 0 twice | The final direct run passed every package after the final code and test changes. |
| `go vet ./...` | 0 twice | Final run passed. |
| `go build ./...` | 0 twice | Final run passed. |
| `golangci-lint run --new` | 1, 1, then 0 twice | Intermediate runs found newly added lint issues; those were corrected. Final and post-adversarial-test runs report `0 issues`. |
| `golangci-lint run` | 1 twice | Initial run reported 27 findings, including the new issues that were corrected. Final run reports 20 findings, all in unchanged files outside this task. |
| `go test ./pkg/agentic/systems/codex -run '^TestCodexArgvHasExactlyOneConstructionSite$' -count=1` | 0 | Provider argv spelling guard passed. |
| `gofmt -d` over changed Go files | 0 | No formatting diff. |
| `git diff --check` | 0 | No whitespace errors in tracked diff. |

The final full-lint findings are pre-existing `errcheck`, `staticcheck`, and `unused` findings in `pkg/providerlimits`, `pkg/vendorplugin`, `pkg/agentic/systems/{claude,codex,gemini}`, `pkg/inferenceengine`, and `tools/agents-management/cmd`. The changed-code lint gate is clean.

## Worktree and handoff

- Worktree base and checkpoint: `1fef4b247ec74012fd454ce7c75093d4862cfc43` (`main` authority verified by the board).
- Changes remain uncommitted in the assigned Story worktree.
- No `LOGBOOK.md` was created, per the task brief. Findings and validation history are recorded here.
- This artifact is attached to task `TASK-260924-2y5w1w` as an outcome resource.
