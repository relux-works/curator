# TASK-260919-2cmg0y brief (orchestrator, binding)

Story STORY-260919-37szes — LAST leaf of the Story (tip carries 3ukdk4, 2wyzde, 3v7x6j,
2eg8nv, 1sbj7o, 3ccq6b, 2d9gfv). Origin: TASK-260910-3eu4cy review F-W1 (verdict rev3 §F-W1):
the Windows `internal/install` lane spends 1257–1529 s in
`TestDraftFailureAtEveryTargetClassRestoresPriorState` because the trunk transaction engine
walks `filepath.EvalSymlinks` once per target per journal save
(`saveJournal → validateJournal → validateIndependentTargetNamespaces → canonicalNamespacePath`,
internal/transaction/journal.go:71, namespace.go:87/201); a late-class rollback costs 5–6 min
on Windows. The CI budget was raised to 120m meanwhile (TASK-3ux95w), so this is a performance
leaf, not a landing blocker.

## Rulings

R1 Cache canonical namespace paths per transaction (engine/journal scoped, keyed by the exact
input path, value = resolved path or the resolution error class), populated on first use and
INVALIDATED by the per-write boundary recheck (`runCommitRecheck`/boundary recheck seam —
find the exact seam and cite it) so that the recheck always re-walks the filesystem: the
independence proof must never rely on a stale resolution across a recheck boundary. No
cache across transactions or engines; no cache when the recheck is disabled/absent
(fall back to the current per-save walk). Every existing namespace/boundary negative row
(`TestRunCommitRecheckRefusesSwappedParent`, `…RefusesDestinationInsideAdmitted`,
`…PassesUnchangedBoundaries`, staging/namespace refusals) stays green and a NEW negative row
proves the invalidation: a symlink swapped between two saves inside one recheck epoch is
still caught at the next recheck (and the transaction refuses as today).
R2 fsync policy: do NOT weaken durability. A change is admissible only if (a) it removes a
provably redundant sync (same file synced twice, or a directory fsync repeated for the same
replace) with the durability rows (`durability_unix.go`, crash-consistency/replace rows)
unchanged and green, and (b) the measurement shows it matters. Otherwise record the fsync
policy as a bound with the measurement, not a change.
R3 No rollback proof weakened, no test deleted or skipped: the sweep keeps injecting a
failure at every publication step with the same assertions; no new skip class (ledger
vocabulary unchanged).
R4 Measurement is the AC: the Windows hosted-gate elapsed of
`TestDraftFailureAtEveryTargetClassRestoresPriorState` (from the gate's
`test-evidence-windows-latest` go-test.json) must be ≥ 2× faster than the baseline. Baseline =
the same test's elapsed in the most recent green gate of this Story (run 35555033097,
BUG-2d9gfv rev2) — extract both numbers and the runner-speed proxy the F-W1 verdict used
(sum of top-level install tests present in both runs) so a fast/slow runner does not fake
the ratio. Also report the macOS/Linux elapsed and the local per-class timings before/after.
If 2× is not reached with the cache alone, say what the remaining cost is (profile) and stop
at the proof-preserving change — do not chase the number with proof-weakening changes.

## Evidence

- Unit rows in `internal/transaction`: cache hit/miss counts through a seam (EvalSymlinks
  call counter or an injected resolver) — one save with N targets walks each path once; K
  saves within one epoch still walk once; a recheck resets the counter (mutant: drop the
  invalidation → the swapped-symlink row fails; mutant: cache across transactions → the
  cross-transaction row fails).
- Production-entry: the sweep itself unchanged; `install.Project` rollback rows unchanged.
- Windows: rows run there (this is the lane that matters); no POSIX-only shim for the new
  rows unless a sibling row already skips for the same reason.
- CHANGELOG `## Unreleased` → `### Changed` (performance, behaviour-preserving) entry.
results.md: hot-path description, cache design + invalidation seam, measurement table
(baseline vs candidate on the same lanes with the runner-speed proxy), fsync decision with
evidence, mutant table, ratio line (no corpus change expected). Publish the Change Request
only when the configured gate is green.
