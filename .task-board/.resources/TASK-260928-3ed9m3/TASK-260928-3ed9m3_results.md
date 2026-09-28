# TASK-260928-3ed9m3 results — revision 2 re-apply

## Re-apply on trunk

Re-applied accepted rev1 from `refs/campaign/lpnvkn-rev1-20260928` onto trunk `d8e87bac`. Resolved `internal/envprofile/managed.go` and `switch.go` by carrying rev1's skill-surface behavior over trunk's E5 nofollow helpers and preserving E6 path-source preflight. Materialization and directory-backup paths added here use `managedPath`, `atomicManagedFile`, `atomicManagedLink`, and guarded tree traversal; existing E5 helpers were preserved.

`git diff --name-only origin/main -- . ':!.task-board'` exited 0 and listed exactly these seven paths: `cmd/curator/env.go`, `cmd/curator/global_lock_publication_test.go`, `cmd/curator/main.go`, `cmd/curator/profile.go`, `internal/envprofile/global.go`, `internal/envprofile/managed.go`, and `internal/envprofile/switch.go`. No CHANGELOG or LOGBOOK changes; no stray files. The worktree is uncommitted.

## Behavior and production-entry rows

Curator-spec main `4ad8042`, environments.md §9.4 and manager transaction rules §§2.5/12.3, require global add/install to publish the resolved lock and immutable entries before in-place materialization, retain the published lock on `environment_surface_unmanaged_conflict`, and recover through `profile sync --takeover` or `profile use --takeover`, then retry. The global commands remain outside the takeover carrier set.

The shared production path in `internal/envprofile/global.go:44` publishes the extended lock before managed/in-place preflight and materialization. The CLI rows in `cmd/curator/global_lock_publication_test.go` exercise:

- `TestGlobalAddPublishesProfileLockBeforeNativeMaterialization`
- `TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers`
- `TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization`
- `TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers`

The conflict rows verify the error names `.agent-context/skills/new-skill/SKILL.md`, the extended lock and store entry remain, the operator-owned bytes remain until takeover, `profile sync --takeover` backs them up and materializes the updated skill set, and retry succeeds. These rows must switch to `conformance/v1/vectors/environments-global-lock-publication.json` when `SPEC_PIN` moves past `4ad8042`. No gap-ledger rows were added because that vector family is not in the pinned suite.

## Local validation (zsh with pipefail)

- `go test ./internal/envprofile -run 'Global|Takeover|Lock|Nofollow|Path|Guarded'` — exit 0 (`ok`, 104.719s).
- `go test ./cmd/curator -run '^(TestGlobalAddPublishesProfileLockBeforeNativeMaterialization|TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers|TestGlobalInstallPublishesMigratedLockBeforeNativeMaterialization|TestGlobalInstallConflictRetainsMigratedLockAndTakeoverRecovers)$' -count=1` — exit 0 (`ok`, 90.253s).
- `go vet ./internal/envprofile ./cmd/curator` — exit 0.
- `golangci-lint run ./internal/envprofile ./cmd/curator` — exit 0, 0 issues.
- `git diff --check origin/main -- . ':!.task-board'` — exit 0.
- The broader `go test ./cmd/curator -run 'Global|Takeover|Profile'` was stopped after it selected the 36-case profile-install matrix; it returned exit 1 and is not counted as passing. The four task production-entry rows above were then run directly with a narrow mask and passed.
- Hosted gate: not run locally. The board runner publishes the Change Request and runs the hosted gate after developer handoff.

## Mutant evidence

Each mutant was run as a fresh `go test -count=1` process against the CLI production-entry row; exit 1 is the expected killed-mutant result.

- **Lock published after in-place surfaces:** `go test ./cmd/curator -run '^(TestGlobalAddPublishesProfileLockBeforeNativeMaterialization|TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers)$' -count=1` — exit 1. The ordering row found the skill absent; the conflict row observed success instead of the required conflict.
- **Published lock rolled back on conflict:** `go test ./cmd/curator -run '^TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers$' -count=1` — exit 1. The regression row found no `new-skill` lock member.
- **Narrowed conflict inventory (adapter root file only; skill directories omitted):** `go test ./cmd/curator -run '^TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers$' -count=1` — exit 1. The row observed global add succeed over the operator-owned skill directory.

All mutated files were restored from saved candidate copies; `cmp` returned exit 0 for each restoration. The named regression and narrowing attack requested for this revision are the production CLI row `TestGlobalAddConflictRetainsLockAndProfileSyncTakeoverRecovers` and the narrowed-inventory mutant above.

The rev1 verdict resource is marked ACCEPTED and records the install-specific mutant bound: global install's rows can be non-discriminating when the skill is already in the lock. The install production-entry rows passed here; ordering and rollback mutants were killed through the shared global operation using the add rows.

## CHANGELOG entry (for release prep)

Global skill add/install now publish the resolved profile lock before native surface materialization and retain it on unmanaged-surface conflicts, enabling takeover and retry.
