# TASK-260910-2vnjej — review verdict rev4: ACCEPTED

CR-TASK-260910-2vnjej-4, base 213a53e5, tree 2742e187 (the worktree's tracked diff and untracked files match the 26 CR paths). Reviewer: claude-opus-5-5 low. Shell zsh, `set -o pipefail`, real exit codes.

## Independent runs
- `go vet` and `GOOS=windows go vet` on ./internal/registry ./internal/install ./internal/pathboundary → 0
- `go test ./internal/registry ./internal/pathboundary ./internal/config` → 0
- `go test ./internal/install -run 'Checkpoint|Bootstrap|Registry|Mirror|Diverg'` → 0. Without CURATOR_CONFORMANCE_ROOT the vector tests SKIP, so I reran them below.
- With CURATOR_CONFORMANCE_ROOT set to curator-spec 23435129 conformance/v1 (rc.13 pin), `TestRegistryBootstrapVectorsDriveInstall` and `TestRegistryMirrorComparisonVectorsDriveInstall` PASS → 0. Both drive the install entry: `e.install` → marker, messages and persisted posture.
- `internal/envprofile TestManagerOwnedAbsenceReadsAreGuarded` → 0

## Mutants (in a disposable copy under $TMPDIR; the worktree was not touched)
| # | Rule | Mutant | Result |
|---|---|---|---|
| M1 | mirror-group root comparison (mirror.go Observe) | `len(roots) < 2` → `true` (never diverge) | KILLED: TestRegistryMirrorComparisonVectorsDriveInstall, rc=1. It survives when the conformance root is unset (the test skips); the hosted gate sets the root. |
| M2 | high-water enforced (checkpoint.go reconcile) | lower/equal-conflict check → `false &&` | KILLED: TestReconcileBootstrapCheckpointPreservesAndAdvancesHighWater, rc=1 |
| M3 | unverifiable ≠ absent | swallow every read error when state exists | KILLED: TestReconcileBootstrapCheckpointFailsClosedForMalformedRebootstrap and TestReconcileBootstrapCheckpointRequiresValidSignatureOnlyOnFirstUse, rc=1 |
| M4 | private-file / DACL check | CheckPrivateFile → nil | KILLED: TestCheckPrivateFileRefusesForeignMutation, rc=1 |

## Findings
- The checkpoint read goes through stateread.Lstat, then a no-follow open with a SameFile recheck, then a signature check. Present-but-bad input returns UnusableError, which fails closed. The one exception, when state already exists and the signature is invalid, is documented. On Windows, CheckPrivateFile reuses checkNode/checkMutationPermissions, which is the platform owner/DACL model.
- Gap ledger: 2 rows owned by STORY-260910-6bo7ej were removed (manager-config-v2 bootstrap members).
- No CHANGELOG or LOGBOOK edits.
- Residuals (bounds, not blockers):
  - (a) The Windows DACL branch is verified only by the hosted windows lane, not locally.
  - (b) An invalid signature on a rebootstrap is ignored silently rather than diagnosed. This is per the code comment; its spec citation should be confirmed in release prep.
