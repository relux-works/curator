# TASK-260917-8vfgxf results

## Implementation

- Added `internal/opaquescan.NULPaths`, a recursive snapshot scan that reads manager-owned entries through `internal/stateread`. It examines regular files at every depth and extension, does not follow symlinks, ignores directories and other non-regular entries, and returns unreadable/disappearing paths as errors.
- `contextaudit.Detect` now always adds an unwaivable `opaque-file` / `nul-byte` blocking finding for each matching file. `DetectFiles` applies the same rule to all supplied files. The secret detector keeps its prior scope and waiver behavior.
- The production install path is `envprofile.Install` → `contextresolve.Resolve` → `auditAndStore` → `auditMembers` → `auditMember`; `auditMember` calls `contextaudit.Detect` before storing/publishing the member. The update path is `UpdateWithOptions` / `UpdateWithPolicy` → `updateLocked`; new members are checked by `auditNewUpdateMember`, and candidate members also flow through `auditCandidates` → `auditMembers` → `auditMember`. Global skill migration uses the same context detector.
- `audit.auditSubject` scans before consulting the verdict cache and blocks NUL-bearing snapshots without using the content-hash cache. This prevents a clean v1-colliding twin's cached allow from authorizing the NUL-bearing tree. The audit ruleset cache version was raised from 1 to 2.
- Refusal diagnostics now retain the existing “blocking finding” phrase and identify the finding class and relative file path.

## Acceptance evidence

- `TestInstallBlocksBothV1CollidingSkillTrees` reaches the real `Install` path. It constructs the v1 single-record and split-record trees, verifies their `hashing.ContentSHA256` values collide, then verifies each install is refused with `opaque-file` and the respective file path.
- The same test verifies a NUL-free tree still installs. `TestInstallBlocksDeepNULFileInContextPathSnapshot` refuses a deep `assets/deep/image.unsupported` file, and `TestUpdateBlocksDeepNULFileInNewContextMember` refuses `docs/deep/opaque.unknown` while preserving the old lock. Context detector tests cover an out-of-scope deep file, an unwaivable finding, and symlink non-following.
- The rc.13 spec checkout had no matching interim opaque/NUL blocking wording or vector in `protocol/` or `conformance/v1/vectors/context-detectors.json` (search exit code 1, no matches). The task brief assigns spec wording/vectors to TASK-260917-2vapkz.

## Mutant evidence

Both mutants survived the existing manager-test selection (exit 0) and were killed by the new production-entry rows (exit 1):

1. **Directory-limited scan.** Mutant limited NUL detection to `context/`. Existing selection `go test -count=1 ./internal/envprofile -run '^(TestInstallPathAndList|TestInstallRejectsSecretMember|TestUpdateBlocksSecretOverlayMember)$'` exited 0. New selection `go test -count=1 ./internal/envprofile -run '^(TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember)$'` exited 1: both collision rows and the deep assets install were admitted.
2. **Warning demotion.** Mutant demoted context findings and audit findings/decision to warning. The existing manager selection above exited 0. The new production-entry selection above exited 1: both collision rows and the deep install/update were admitted.

During the first directory-mutant baseline attempt, the same existing selection exited 1 because the new diagnostic wording had dropped an existing test's “blocking finding” substring. The wording was adjusted to preserve that text, and the repeated selection under the mutant exited 0 as recorded above.

## Verification

- `go test -count=1 ./internal/opaquescan ./internal/contextaudit ./internal/audit` — exit 0.
- `go test -count=1 ./internal/envprofile -run '^(TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember)$'` — exit 0.
- `CURATOR_CONFORMANCE_ROOT` pointed to `conformance/v1`; `go test -count=1 ./internal/interop/environments -run '^TestConformanceContextDetectors$'` — exit 0.
- `go build ./...` — exit 0.
- `go vet ./...` — exit 0.
- `golangci-lint run` — exit 0, 0 issues.
- `gofmt -l cmd internal` — exit 0, no files listed.
- `git diff --check` — exit 0.
- An initial run of `go test -count=1 ./internal/opaquescan ./internal/contextaudit ./internal/audit` exited 1 on an unused local; that was corrected and the exact command passed afterward.
- A package-wide `go test -count=1 ./internal/envprofile` run was manually interrupted after more than four minutes; it returned exit 1 and is not counted as passing. The focused install/update production-entry selection above passed on the final tree.

## Gate fix 1 (rev2)

1. Renamed `internal/opaquescan/nul.go` -> `internal/opaquescan/nulopaque.go` (NUL is a Windows reserved device name). `git ls-files -c --others --exclude-standard | grep -iE '(^|/)(nul|con|prn|aux|com[1-9]|lpt[1-9])(\.|/|$)'` printed nothing (grep exit 1 = no match).
2. The absence-sensitive `os.Stat`/`os.IsNotExist` (agent-skill.json vs csk-skill.json manifest name) moved from `audit.detect` into `audit.detectWithOpaquePaths` in the refactor. It inspects the caller-selected source snapshot, not manager-owned state, so the existing reviewed allowlist entry was renamed `internal/audit/audit.go:detect` -> `internal/audit/audit.go:detectWithOpaquePaths` under its unchanged reason ("Inspects a caller-selected source snapshot for declared content..."). No entries were added, so the allowlist is no wider.

Commands (real exit codes):
- `go build ./...` -> 0
- `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded` -> 0
- `go test ./internal/audit ./internal/opaquescan ./internal/contextaudit` -> 0 (opaquescan has no test files)
- `go test ./internal/envprofile -run 'NUL|Nul|Opaque'` -> 0 (TestInstallBlocksDeepNULFileInContextPathSnapshot, TestUpdateBlocksDeepNULFileInNewContextMember)
- `go vet ./internal/audit ./internal/opaquescan ./internal/envprofile ./internal/contextaudit` -> 0
- I did not rerun the mutants or the full suite in rev2. The rename only moves a file, and the allowlist change only renames an entry, so the rev1 mutant evidence above still holds.


## Rework 1 — default-config admission and file-naming refusal

### Production paths

- `audit.Gate` and `GateReadOnly` share `gate` (`internal/audit/audit.go:155-205`). It calls `opaquescan.NULPaths` before checking `!cfg.Audit.Enabled` at line 195. Scan failures refuse admission; every NUL path becomes a critical `audit.opaque.nul-byte` finding with `DecisionBlock`, and `auditFindingMessage` includes `Finding.File` (`audit.go:571-598`).
- The project-scope entry is `install.Project` (`internal/install/install.go:231`), whose production audit closure calls `Gate` or `GateReadOnly` at lines 565-583. The global entry is `install.Global` (`internal/install/global.go:36`), with the same shared gate at lines 184-200.
- CLI build-source auditing reaches the shared gate from `productionExternalDeps` (`cmd/curator/main.go:1866-1902`); the CLI audit command reaches it through `auditTarget` (`main.go:2168-2199`). Tests drive both CLI call paths with audit disabled.
- Context/environment admission uses `envprofile.Install` → `auditAndStore` → `auditMembers` → `auditMember`; `auditMember` calls `contextaudit.Detect` before storing/publishing (`internal/envprofile/envprofile.go:687, 1492-1542`). Update admission uses `Update`, `UpdateWithPolicy`, or `UpdateWithOptions` → `auditNewUpdateMember` (`envprofile.go:1123-1163, 1567-1582`). Global skill migration also calls `contextaudit.Detect` before recording the member (`envprofile.go:2001-2012`). `contextaudit.Detect` scans every regular file with the same `opaquescan.NULPaths` rule (`internal/contextaudit/contextaudit.go:141-192`).
- A production call-site search found no other `audit.Gate` or `contextaudit.Detect` admission path. The explicit `AuditGate` option on install is the existing test seam; normal project/global production calls use the shared gate.

### Regression rows

- `TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled` (`internal/install/nul_opaque_test.go:12-118`) asserts the default config has audit disabled, proves the two distinct v1-framed trees have equal hashes, and drives both project and global install for both trees. Each refusal must carry `critical audit.opaque.nul-byte` and the relative filename. It also refuses a deep `docs/deep/opaque.unsupported` file and confirms a NUL-free tree still installs in both lanes.
- The existing environment rows drive `envprofile.Install` for both colliding trees, a NUL-free skill, and a deep unsupported file, plus `UpdateWithPolicy` for a deep NUL file while preserving the prior lock (`internal/envprofile/nul_opaque_test.go`). The context detector tests also confirm an opaque finding cannot be waived and that symlink targets are not followed (`internal/contextaudit/contextaudit_test.go`).
- `TestGateBlocksDeepOpaqueNULWhenAuditIsDisabledAndNamesFile` checks the shared disabled-config gate directly (`internal/audit/audit_test.go:130-149`). CLI regressions cover the `productionExternalDeps` normal/dry-run paths and the `curator audit` path (`cmd/curator/nul_opaque_test.go:14-70`).

### Mutants

For each mutant, the existing clean `TestGlobalInstall` control passed with exit 0; the new default-config negative install test failed with exit 1. Mutants were restored from saved source copies and `cmp` verified the restoration.

1. **Audit-toggle bypass:** returned early when `Audit.Enabled` was false, before the scan. Control `go test ./internal/install -run '^TestGlobalInstall$' -count=1` → 0; regression `go test ./internal/install -run 'TestOpaqueNULIsBlockedAtProjectAndGlobalInstallWithAuditDisabled' -count=1` → 1.
2. **Directory-limited scan:** restricted `NULPaths` to `assets/`. The control command → 0; the same regression command → 1 because the `docs/` collision and deep-docs rows were admitted.
3. **Warning demotion:** changed the opaque report decision from `DecisionBlock` to `DecisionWarn`. The control command → 0; the same regression command → 1 because project/global installs were admitted with warnings.

### Spec and read requirements

At `v1.0.0-rc.13`, `protocol/core.md` §8 still uses the NUL-delimited v1 framing. `profiles/manager.md` §7 describes the optional audit pipeline but has no interim opaque-NUL rule. The rc.13 `conformance/v1` tree has no opaque/NUL hash vector. The task brief assigns that spec addition to TASK-260917-2vapkz. The new recursive byte reads are through `internal/stateread` in `opaquescan.NULPaths`; the unchanged reviewed allowlist entry remains at `detectWithOpaquePaths`, and `TestManagerOwnedAbsenceReadsAreGuarded` passes. No Windows-reserved path names were added. No CHANGELOG or LOGBOOK file was edited.

Board inspection found no `README.md` resource attached to this task or either story; I read the project README, the task's four attached instructions, the epic's `remediation-manager-producer-rules.md`, and the task/story descriptions and AC through the board CLI.

### Rework validation

- `go test ./internal/audit` → 0.
- `go test ./internal/audit ./internal/opaquescan` → 0 (`opaquescan` has no standalone test files).
- `go test ./internal/contextaudit` → 0.
- `go test ./internal/install -run 'NUL|Opaque'` → 0; final uncached `go test ./internal/install -run 'NUL|Opaque' -count=1` → 0.
- `go test ./internal/envprofile -run 'NUL|Opaque'` → 0.
- `go test ./cmd/curator -run 'NUL|Opaque'` → 0.
- `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded` → 0.
- `go test ./internal/interop/environments -run TestConformanceContextDetectors` → 0.
- `go build ./internal/audit ./internal/opaquescan ./internal/contextaudit ./internal/install ./internal/envprofile ./cmd/curator` → 0.
- `go vet ./internal/audit ./internal/opaquescan ./internal/contextaudit ./internal/install ./internal/envprofile ./cmd/curator` → 0.
- `golangci-lint run ./internal/audit ./internal/opaquescan ./internal/contextaudit ./internal/install ./internal/envprofile ./cmd/curator` → 0 issues, exit 0.
- `gofmt -l` over changed Go files → exit 0, no output; `git diff --check` → 0.
- Required reserved-name grep over `git ls-files` printed nothing (grep exit 1 = no match). The same check over tracked and untracked candidate paths also printed nothing (exit 1).
- A broad `go test ./internal/install` diagnostic run before this rework exited 1 at the 10-minute test timeout while `TestDraftGitBuildsPublishReceipt3` was still running. The full package suite is not counted as green; the focused NUL/Opaque selection above passed.

## Review branch evidence and handoff

The attached rev2 verdict, `TASK-260917-8vfgxf_review-verdict-rev2.md`, was CHANGES REQUESTED. Its evidence remains attached, and this rework addresses the reported default-config bypass and missing filename. The task was routed back to `development` for this rework. The prior developer handoff attempt returned exit 1 because checklist item 15 was still unchecked; the reject branch evidence and development routing already exist, so that item is now being marked handled before retrying handoff.

## CHANGELOG entry (for release prep)

Reject skill and context snapshots containing NUL bytes in any regular file with a blocking opaque finding that names the file, even when optional skill auditing is disabled.
