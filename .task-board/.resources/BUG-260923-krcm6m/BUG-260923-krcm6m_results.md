# BUG-260923-krcm6m results

## Hosted root cause evidence

Downloaded the specified artifact from hosted run 35901867517 with `gh run download 35901867517 -n test-evidence-windows-latest` (exit 0). The task-scoped `BUG-260923-krcm6m_hosted_windows_evidence.json` contains the relevant rows extracted from `test/go-test.json`: the Windows `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/0s/one_second_past_the_bound` row failed at `registry_test.go:734` with `refused=false want true` and elapsed 1.55 seconds.

The test supplied a fixed, whole-second `now`, an instant fetch, and a fresh `t.TempDir()` for each row. `CheckSnapshotsWithPolicy` starts its latency timer before initializing persistent rollback state. On an empty cache, `loadSnapshotStateCatalog` creates the catalog through `writeSnapshotStateCatalog` and `writeProtectedStateFile` (including `File.Sync`) before the future-date comparison. The comparison adds `time.Since(start)` to `now + clockSkew`. The hosted failure and 1.55-second row duration show that this cache-initialization path let the elapsed term cross the one-second edge, so the test's nominal `now+1s` timestamp was admitted. The fresh per-row directories rule out cache reuse between subtests.

## Change

The exact-bound test now seeds each isolated cache with an empty rollback catalog before calling `CheckSnapshotsWithPolicy`. This keeps the persistent production entry point and all inside, exact, one-second-past, and one-hour-past assertions, while excluding first-use catalog I/O from the elapsed allowance under test. The production threshold is unchanged. `CHANGELOG.md` records the test stabilization.

## Verification

Commands were run directly in the Story worktree shell; exit codes are recorded exactly.

- `go test ./internal/registry -run '^TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew$' -count=50` — exit 0; 93.644s.
- Threshold-plus-one-second mutant (`clockSkew + time.Second`) with `go test ./internal/registry -run '^TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew$' -count=1` — expected failure, exit 1; the one-second-past row failed at 0s, 30s, and 5m skew. The production condition was restored afterward.
- `go test ./internal/registry` — exit 0; 13.519s.
- `golangci-lint run ./internal/registry/...` — exit 0; 0 issues.
- `go build ./internal/registry` — exit 0.

The hosted Windows run is the failure evidence above; no new hosted run was launched locally. The hosted CI gate will run at handoff.

## Revision 2 (carry-forward republish)

Trunk moved to `1511b345`; the accepted revision-1 delta was carried uncommitted into the Story worktree by three-way merge. Verification per `krcm6m-carry-2.md`:

- `internal/registry/registry_test.go`: trunk did NOT touch this path (HEAD base still the pre-fix `t.TempDir()` inline form); the working-tree hunk is byte-identical to `BUG-260923-krcm6m_change-request_rev1.patch`. Nothing else changed.
- `CHANGELOG.md`: intersecting path (trunk added E4 umbrella-provider and other Unreleased entries). Both sides present: the carried `### Fixed` future-bound entry sits at the top of `## Unreleased`, trunk's `### Added` entries follow intact. No conflict markers (`git diff --check` exit 0; marker grep empty), nothing dropped or duplicated.
- Working tree contains exactly the two rev-1 paths; no other edits.

Focused bounded run (this turn, standalone process, real exit code):

- `go test ./internal/registry -run 'FutureBound' -count=20` — exit 0; ok 194.765s. All `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew` rows (0s/30s/5m x inside/exact/+1s/+1h) and `TestSnapshotFutureBoundToleratesCheckerLatency` pass.

Revision-1 evidence (hosted artefact root cause, -count=50 stability, threshold+1s mutant killed, full package suite, lint, build) is unchanged and carries forward; see above.

## Revision 2 re-verification (this headless run, trunk `1511b345`)

Per-path carry check (standalone commands, this turn):

- `internal/registry/registry_test.go`: non-intersecting (rev-1 base index `3e2b9fd7` == HEAD version); working-tree diff minus index lines is byte-identical to the rev-1 patch hunk (`diff` empty). Change nothing else: file holds only the seeded-catalog block.
- `CHANGELOG.md`: intersecting (trunk base moved `2518c29a` -> `17773d62`). Both sides present: carried `### Fixed` future-bound entry at top of `## Unreleased`, trunk's `### Added` (E4 umbrella-provider) entries intact below. No conflict markers (marker grep empty).
- Working tree holds exactly the two rev-1 paths (`git status --short`: `M CHANGELOG.md`, `M internal/registry/registry_test.go`).

Focused bounded run (standalone process, no pipe, real exit code):

- `go test ./internal/registry -run FutureBound -count=20` — exit 0; ok 231.025s.

## Revision 4 (carry-forward republish)

Trunk is now `948ae7c9`; the accepted rev-3 delta was carried uncommitted into the Story worktree by `worktree converge`. Verification per `krcm6m-carry-4.md` step 2 (standalone commands, this turn, before the CHANGELOG revert):

- `internal/registry/registry_test.go`: trunk did NOT touch this path (HEAD still shows the pre-fix inline `t.TempDir()` form at lines 731-732; `git log` for the path shows no trunk-side change since the rev-3 base). The working-tree hunk is byte-identical to `BUG-260923-krcm6m_change-request_rev3.patch` (working diff minus `index`/`diff --git` lines vs rev-3 registry hunk: `diff` empty, exit 0). Nothing else in the file changed.
- `CHANGELOG.md`: intersecting path (trunk moved since the rev-3 base `17773d62`: HEAD now opens `## Unreleased` with an `R5 script-worker-v1` entry above the `E4: umbrella provider` entry; rev-3 context showed `E4` first). Both sides were present before the revert: the carried `### Fixed` future-bound block sat at the top of `## Unreleased` with lines identical to the rev-3 added lines, and trunk's `R5` plus `E4` `### Added` entries followed intact in HEAD order. No conflict markers (`git diff --check` exit 0; marker grep count 0), nothing dropped or duplicated.
- Working tree at verify time held exactly the two rev-3 paths (`git status --short`: `M CHANGELOG.md`, `M internal/registry/registry_test.go`); no stray root `TASK-*`/`BUG-*.md` and no `test/` or `ledger/` paths (`ls` confirms absent).

CHANGELOG POLICY (orchestrator, 2026-09-24), applied after verification, change nothing else:

- Reverted this task's `CHANGELOG.md` hunk entirely: `git checkout HEAD -- CHANGELOG.md` (exit 0). The file now equals trunk's (`cmp` against `git show HEAD:CHANGELOG.md` exit 0; `git diff HEAD -- CHANGELOG.md` empty; `git diff --check` exit 0).
- Working tree now holds exactly one path: `M internal/registry/registry_test.go` (`git diff HEAD --stat`: 1 file, 11 insertions, 1 deletion).
- The entry text is copied verbatim into the CHANGELOG entry section below; the release-prep leaf writes all entries.

Focused bounded run (this turn, standalone process, no pipe, real exit code, after the revert):

- `go test ./internal/registry -run FutureBound -count=20` — exit 0; ok 238.345s. All `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew` rows and `TestSnapshotFutureBoundToleratesCheckerLatency` pass.

Revision-3 acceptance and evidence carry forward unchanged: `BUG-260923-krcm6m_review-verdict-rev3.md` ACCEPTED the content (root cause at `snapshot.go:178`/`89` vs catalog init at `300/407/453/475`, fix seeds the empty catalog before the checker, bound unwidened, narrowing mutant killed, `-count=50` stable, remote gate run 35935995625 passed). No product code changed in any revision (test-only delta); production threshold and exact-edge assertions keep their strength.

## CHANGELOG entry (for release prep)

### Fixed

- Stabilized the audit-registry exact future-bound test on Windows by creating
  its empty rollback-state catalog before calling the checker. The checker's
  elapsed-time allowance includes cache initialization, whose durable file
  sync could take longer than the one-second test edge. The test still drives
  `CheckSnapshotsWithPolicy` with an isolated cache and retains the same
  inside, exact-bound, one-second-past, and far-future assertions.

## Revision 5 (carry-forward republish)

Trunk is now `a48f584c`; the accepted rev-4 delta was carried uncommitted into the Story worktree by `worktree converge`. Verification per `krcm6m-carry-5.md` step 2 (standalone commands, this turn):

- `internal/registry/registry_test.go`: the ONLY path in `BUG-260923-krcm6m_change-request_rev4.patch`. Trunk did NOT touch it relative to the rev-4 base (working diff header `index 3e2b9fd7..301e5782` matches the patch exactly; HEAD blob still shows the pre-fix inline `t.TempDir()` form). Working-tree diff body is byte-identical to the rev-4 patch (`diff` empty). Change nothing else: the file holds only the seeded-catalog block.
- Intersecting paths: none carried. `CHANGELOG.md` has no rev-4 hunk (rev-4 patch is test-only) and equals trunk (`git diff HEAD -- CHANGELOG.md` empty) — CHANGELOG POLICY already satisfied, nothing to revert. The entry text stays verbatim in the "CHANGELOG entry (for release prep)" section above; the release-prep leaf writes all entries.
- Working tree holds exactly one path (`git status --short`: `M internal/registry/registry_test.go`); no stray root `TASK-*`/`BUG-*.md`, no `test/` or `ledger/` paths. No conflict markers (`git diff --check` exit 0; scoped marker grep over `internal/`, `CHANGELOG.md`, `go.mod` empty).

Focused bounded run (this turn, standalone processes, no pipes, real exit codes):

- Attempt 1: `go test ./internal/registry -run FutureBound -count=20` — exit 1. Failures were confined to `TestSnapshotFutureBoundToleratesCheckerLatency`, a test this task does not touch: one `a_slow_sibling's_latency_is_inside_the_bound` logic miss plus `rollback state could not be persisted: open /var/folders/...` errors. The host Data volume was at 100% capacity (3.8Gi avail) with a concurrent agent `go test` running — environment/resource pressure, not the carried delta. That test passes in isolation (`-run TestSnapshotFutureBoundToleratesCheckerLatency -count=1` — exit 0, 11.393s, verified this turn).
- Remediation: `go clean -cache` (regenerable tool output only; exited 1 on one busy subdir but freed the 12G build cache to 40M).
- Attempt 2 (after reclaim): same command — exit 0; ok 226.544s. The captured log (1 line, zero `--- FAIL`) shows the full `FutureBound` selection — all `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew` rows and `TestSnapshotFutureBoundToleratesCheckerLatency` — green across all 20 iterations.

Revision-4 acceptance and evidence carry forward unchanged: content ACCEPTED (root cause at the cache-init `File.Sync` inside the checker's latency allowance, fix seeds the empty catalog before the checker, bound unwidened, narrowing mutant killed, `-count=50` stable). No product code changed in any revision (test-only delta); production threshold and exact-edge assertions keep their strength.
