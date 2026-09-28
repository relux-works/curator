# TASK-260927-4pv4au results

## Implementation

Implemented security posture revision A in the Story worktree. Schema-2 config admits `security_posture`; schema-1 stays permissive. Hardened defaults fill absent policy knobs, explicit values retain precedence, and a system-locked hardened posture wins over machine config. System config may carry posture only as a locked `hardened` value.

Permissive operations emit `security_posture_permissive` once with the migration hint. Curator status and schema-2 env status report the ordered posture inventory and provenance. Hardened profile/install paths refuse an empty source allowlist, MCP declarations with an empty MCP allowlist, and explicit null passthrough; the empty MCP list without declarations warns and proceeds. Environment status `--check` is non-current for hardened contradictions. The effective-posture API is available to the install resolver. Hardened defaults also set `require_source_signers` when that knob is absent.

The pinned posture vectors have 11/17 cases driven and 6/17 bounded: the two unreachable-registry cases belong to TASK-260910-1sapuy; revision-B flip cases belong to TASK-260927-25hk87. No `security-posture/vectors` gap rows remain. The direct config vector test compares the exact revision tuple supplied by each vector. The production CLI test checks the manager's compiled shipped rows; this worktree currently reports S4 `s4-warn` and Codex seed B, so those independently shipped row values are normalized to the running manager in that CLI comparison. Its install fixture uses canonical source identities and includes MCP dependencies in the broad source allowlist used by the current resolver.

## Local verification

Commands were run as standalone processes; exit codes below are captured process results.

- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/config -count=1` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^TestSecurityPostureVectorsThroughCLI$' -count=1 -v` — exit 0; 11 driven, 6 bounded, 17 total.
- `go test ./cmd/curator -run '^(TestSecurityPostureWarningAndEnvStatusRows|TestSecurityPostureWarningForEnforcedShim|TestSecurityPostureHardenedContradictionFailsEnvStatusCheck|TestSecurityPostureSchema1EnvStatusHasNoPostureInventory|TestSecurityPostureCuratorStatusRows|TestSecurityPostureProfileInstallRefusesEmptySourceAllowlistBeforeWriting|TestSecurityPostureProfileInstallRefusesMCPDeclarations|TestSecurityPostureHardenedMCPContradictionFailsEnvStatusCheck|TestSecurityPostureHardenedEmptyMCPWithoutDeclarationsWarnsAndInstalls|TestSecurityPostureResolveRefusesExplicitUnboundedPassthrough|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands)$' -count=1 -v` — exit 0.
- `go test ./cmd/curator -run '^TestSecurityPostureHardenedMCPContradictionFailsEnvStatusCheck$' -count=1 -v` — exit 0.
- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -v` — exit 0; 377/377 reads covered (263 via stateread, 114 allowlisted).
- `go test ./internal/envprofile -count=1` — exit 0 (8m24s, from the earlier implementation run; envprofile production code was not changed afterward).
- `go test ./internal/scriptworker -run '^TestProductionBinaryLaunchesWhenHostProvides$' -count=1 -v` — exit 0.
- `go test ./internal/install -run '^TestEnforcedInstallAndLaunchAtCLIEntry$' -count=1 -v` — exit 0.
- `go build -o "$TMPDIR/curator-security-posture-build" ./cmd/curator` — exit 0.
- `golangci-lint run` — exit 0, 0 issues.
- `git diff --check` — exit 0.

Intermediate command outcomes: an early invocation of the CLI vector test exited 1 when the test compared pinned A rows to this manager's compiled S4-warn/Codex-B rows; another exited 1 because the CLI fixture's MCP source was outside the broad source allowlist; a subsequent invocation exited 1 because the fixture helper dropped normalized `[]string` values and reported an empty source allowlist. Both test-fixture issues were corrected; the final invocation above passes. The first focused posture/status test group exited 1 because its new MCP contradiction fixture omitted the MCP declaration; the fixture was corrected, and both the focused MCP test and complete behavior group above pass. Three curator test attempts exited 1 after 301 seconds with `acquire package host GOROOT test lock: context deadline exceeded` while another Story's curator tests held the shared lock; final focused curator runs acquired the lock and passed. The first lint invocation exited 1 with two findings (ineffectual assignment and unused parameter), corrected before the green lint run.

The prior full `go test ./cmd/curator -count=1` attempt exited 1 at the approximately 10-minute command limit, stalled in `TestEnvStatusRegistryBoundaryPostureAndCheck` while probing the host adapter release command. The full curator package suite was not rerun; the changed-behavior focused tests passed. The first hosted Change Request validation (revision 1) was red: three failures were `TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands`, `TestProductionBinaryLaunchesWhenHostProvides`, and `TestEnforcedInstallAndLaunchAtCLIEntry`. The status assertion and test configs were updated; the next handoff validation log is authoritative for this revision.

## Mutant evidence

Each temporary mutant was applied, its test command was run directly, and the mutant was restored. All four gates were killed with real exit code 1:

- Flip the shipped posture revision A to B: `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/config -run '^TestSecurityPostureVectors/revision-A-default-permissive-status$' -count=1` — exit 1, effective-default mismatch.
- Ignore the locked posture: `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/config -run '^TestSecurityPostureVectors/locked-posture-beats-explicit-permissive$' -count=1` — exit 1, locked defaults replaced by explicit values.
- Drop the permissive warning: `go test ./cmd/curator -run '^TestSecurityPostureWarningAndEnvStatusRows$' -count=1` — exit 1, expected one warning but observed zero.
- Remove the explicit-null refusal: `go test ./cmd/curator -run '^TestSecurityPostureResolveRefusesExplicitUnboundedPassthrough$' -count=1` — exit 1, required refusal absent.

No `CHANGELOG.md` or `LOGBOOK.md` edits were made.

## CHANGELOG entry (for release prep)

Add revision-A manager `security_posture` support with the permissive migration warning, hardened effective defaults and refusals, and status inventory rows.

The revision-1 hosted validation was red and its three failures are listed above. The Change Request validation log produced by this handoff is the authoritative hosted result; this artifact does not infer green status from local runs.

## Revision 3 — no warning on the launch path

Addressed review F1 from `TASK-260927-4pv4au_review-verdict-rev2.md`. `runEnforcedShim` no longer writes `SecurityPostureWarning` to the launcher's stderr; the launched command continues to own stdout and stderr. Restored the original schema-1 fixtures in `internal/scriptworker/derive_test.go` and `internal/install/scriptpolicy_test.go`. The new production-entry regression `TestProductionEnforcedShimDoesNotLeakPermissiveWarning` launches the built curator binary through its native shim under both explicit schema-2 permissive and schema-1 default-permissive config, captures stdout and stderr separately, verifies neither contains `security_posture_permissive`, and confirms the interpreter ran.

### Revision 3 local verification

Commands were run as standalone processes with captured exit codes:

- `go test ./internal/scriptworker -run '^(TestProductionBinaryLaunchesWhenHostProvides|TestProductionEnforcedShimDoesNotLeakPermissiveWarning)$' -count=1` — exit 0.
- `go test ./internal/install -run '^TestEnforcedInstallAndLaunchAtCLIEntry$' -count=1` — exit 0; this exercises the restored schema-1 install-and-launch fixture.
- `go test ./internal/config ./internal/envprofile -run 'Posture|Security|Hardened|Permissive' -count=1` — exit 0.
- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` — exit 0.
- `go test ./cmd/curator -run '^(TestSecurityPostureWarningAndEnvStatusRows|TestSecurityPostureWarningIsNotEmittedByEnforcedShim|TestSecurityPostureHardenedContradictionFailsEnvStatusCheck|TestSecurityPostureSchema1EnvStatusHasNoPostureInventory|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands)$' -count=1` — exit 0.
- `go build -o /tmp/curator-security-posture-f1 ./cmd/curator` — exit 0.
- `golangci-lint run` — exit 0, 0 issues.
- `git diff --check` — exit 0.
- Narrowing mutant: temporarily re-added the `SecurityPostureWarning` stderr emission inside `runEnforcedShim`, then ran `go test ./internal/scriptworker -run '^TestProductionEnforcedShimDoesNotLeakPermissiveWarning$' -count=1` — exit 1 as expected; both explicit-permissive and schema-1 subtests failed because the launched command's stderr contained the warning. Restored `main.go` from a pre-mutant copy and verified byte-for-byte with `cmp` (exit 0).
- Re-ran `go test ./internal/scriptworker -run '^TestProductionEnforcedShimDoesNotLeakPermissiveWarning$' -count=1` on the restored source — exit 0.

An initial attempt at the combined scriptworker test command exited 1 at compile time because the fixture refactor left an unused `curator` binding. Removed that binding and reran the same command successfully (exit 0). No `CHANGELOG.md` or `LOGBOOK.md` edits were made. The hosted gate is queued by handoff after this turn; no hosted result is claimed here.

### Handoff checklist routing note

The first `task-board handoff TASK-260927-4pv4au --role developer` attempt exited 1 because checklist item 13 was unchecked. Its condition was already discharged for the previous review round: the attached rev2 verdict is changes-requested, and the task was routed back to `development` for this rework. This evidence was recorded before checking item 13 and retrying handoff.
