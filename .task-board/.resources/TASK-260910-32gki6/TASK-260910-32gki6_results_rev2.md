# TASK-260910-32gki6 — developer results

## Scope and implementation

Implemented protected profile-store verification in the manager `env resolve` production path. It checks enclosing environment/store/lock/marker boundaries before named entries, then recomputes entry pins, then checks home currency/surface records. Boundary checks cover ownership, private mutation permissions, containment, regular file types, and `lstat` link safety. Enclosing-boundary failures refuse the operation; an untrusted entry is reported as `environment_store_untrusted` and emits no fragment. An unreadable lock remains untrusted and is never rebuilt.

For `git` members, resolve reads only the local Git object database to map the locked commit to its tree ID, hashes the store entry bytes as a Git tree, and compares the IDs. Resolve does not fetch or re-extract the snapshot. Missing local commit objects fail closed with the member named and `pinned commit object unavailable`. Repair revalidates the entry and uses the managed no-follow write helpers. Recorded system-prompt and root-context surface hashes are checked against the current plan and drift is reported.

The command acceptance test exercises `curator env resolve --repair --dry-run` and verifies a swapped store entry is rejected without emitting a fragment or mutating the entry.

## Spec clauses and production vectors

Normative source: curator-spec v1.0.0-rc.13, SPEC_PIN `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`.

| Rule / rc.13 clause | Production entry and conformance rows | Result |
|---|---|---|
| §4 protected state and §9 env resolve ordering: enclosing boundaries, entry boundaries, pin recomputation, then home currency; fail-closed classes | `Resolve`: `intact-resolve-emits-fragment`, `swapped-system-prompt-bytes-untrusted`, `swapped-root-context-bytes-untrusted`, `symlinked-entry-root-untrusted`, `wrong-ownership-untrusted`, `wrong-permissions-untrusted`, `containment-escape-untrusted`, `non-regular-component-untrusted`, `lock-file-wrong-owner-untrusted`, `marker-symlink-untrusted`, `environments-root-wrong-owner-untrusted`, `store-root-symlinked-untrusted`, `intact-updated-store-old-marker-stale`, `swapped-updated-store-old-marker-untrusted`, `unprovisioned-intact-stale`, `unprovisioned-swapped-untrusted`, `unreadable-marker-non-current`, `swapped-bytes-emits-fragment`, `untrusted-reported-current`, `untrusted-without-diagnostic` | 20/20 driven; 0 gaps, 0 bounds, 0 skipped |
| §4 failure classes and §9 repair behavior: dry run reports without mutation; enclosing failure refuses repair | `Resolve --repair --dry-run`: `dry-run-untrusted-reports-would-rebuild`, `dry-run-intact-plans-nothing`, `dry-run-enclosing-no-rebuild`, `dry-run-mutates` | 4/4 driven; 0 gaps, 0 bounds, 0 skipped |
| §4 entry repair rules: rebuild only a revalidated Git snapshot; path/local entries cannot be rebuilt; never adopt swapped bytes | `Resolve --repair`: `repair-rebuilds-git-entry-from-snapshot`, `repair-path-entry-cannot-rebuild`, `repair-local-entry-cannot-rebuild`, `repair-rebuilds-entry-boundary-failure`, `repair-enclosing-refuses-no-rebuild`, `repair-stale-old-marker-succeeds`, `repair-swapped-old-marker-never-adopted`, `repair-unprovisioned-intact-provisions`, `repair-unprovisioned-swapped-rebuilds`, `repair-reapplies-untrusted` | 10/10 driven; 0 gaps, 0 bounds, 0 skipped |
| §4/§9 status reporting for intact, entry and enclosing failures | `StatusOf`: `status-intact-current`, `status-names-failing-check`, `status-enclosing-names-boundary`, `status-hides-failing-check` | 4/4 driven; 0 gaps, 0 bounds, 0 skipped |
| §4 Git pin identity and §5.1/§5.6 recorded surface integrity | Resolve regressions: matching local commit/tree passes; tampered sidecar and swapped entry fail; absent locked commit object fails even with an ambient alternate `GIT_DIR`; recorded context surface hash drift is reported | 4 focused production-entry regressions pass |
| §8.4.1 absence vs unreadable manager state | Resolve regression for unreadable manifest/lock diagnostics plus `TestManagerOwnedAbsenceReadsAreGuarded` | Pass |
| E5 managed-write no-follow behavior used by entry repair | `TestEnvironmentWriteNofollowVectors` | 10/11 driven; 0 gaps, 1 bound (`backup-symlinked-target-refused`: no production entry requests backup of an unauthorized unowned link); adjacent takeover paths are driven |

The rc.13 vector root was extracted to `$TMPDIR/curator-rc13-JOPTCA/conformance/v1`. The case-count manifest records 20 resolve, 4 dry-run, 10 repair, and 4 status vectors (38 total).

## Verification and mutation evidence

All commands below were run as standalone processes. Local test/build/lint evidence is from macOS; hosted platform gates run after handoff.

| Command | Exit | Result |
|---|---:|---|
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-JOPTCA/conformance/v1 go test -v ./internal/envprofile -run 'TestStoreBoundary(ResolveVectorsDriveResolve|DryRunVectorsDriveResolveRepairPlan|RepairVectorsDriveResolveRepair|StatusVectorsDriveStatus)'` | 0 | 38/38 rc.13 store-boundary vectors driven |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-JOPTCA/conformance/v1 go test -v ./internal/envprofile -run '^TestEnvironmentWriteNofollowVectors$'` | 0 | 10 driven, 1 documented bound |
| `CURATOR_CONFORMANCE_ROOT=$TMPDIR/curator-rc13-JOPTCA/conformance/v1 go test ./internal/envprofile -run 'Test(ResolveVerifiesGitStoreEntryAgainstPinnedTree|ResolveRejectsGitPinMissingFromLocalObjectDatabase|ResolveReportsRecordedContextSurfaceHashDrift|ReservedSkillName|SurfacingUnreadableManifest|ManagerOwnedAbsenceReadsAreGuarded)'` | 0 | Focused pin, surface drift and diagnostic regressions |
| `go test ./cmd/curator -run '^TestEnvResolveRejectsSwappedStoreEntry$'` | 0 | CLI acceptance criterion |
| `go build ./...` | 0 | Build succeeds |
| `golangci-lint run` | 0 | 0 issues |
| `gofmt -l` over changed Go files | 0 | No output |
| `git diff --check` | 0 | Clean |

Four narrowing mutants were each rejected by their targeted test (expected red, exit 1); each mutation was restored before the final green runs:

- Skip protected store boundary validation → `wrong-permissions-untrusted` vector failed because resolve returned no error (exit 1).
- Skip pin verification → `TestResolveVerifiesGitStoreEntryAgainstPinnedTree` failed because the fragment was emitted (exit 1).
- Treat missing local commit object as a pass → `TestResolveRejectsGitPinMissingFromLocalObjectDatabase` failed because the fragment was emitted (exit 1).
- Ignore the recorded `ContentSHA256` surface hash → `TestResolveReportsRecordedContextSurfaceHashDrift` failed because drift was not reported (exit 1).

## Conformance gap ledger

At `HEAD` before this change: 0 rows owned by `STORY-260910-148pj1` or `TASK-260910-32gki6` (13 total ledger rows). After this change: 0 owned rows (13 total). The Story-owned gap count remains 0→0, so no owned gap row required removal; `.github/ci/conformance-gaps.tsv` is unchanged. The 38 passing vectors are production-entry vectors and are recorded in `.github/ci/conformance-case-counts.tsv`.

## CHANGELOG entry (for release prep)

`env resolve` now checks protected store boundaries and recomputes Git and state pins; it reports recorded surface hash drift and fails closed when the local pinned commit object is unavailable.

No CHANGELOG.md or LOGBOOK.md file was edited. Findings and test evidence are recorded here for the task handoff.
