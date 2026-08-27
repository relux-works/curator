# TASK-260720-2284br — rework cycle 2 implementation notes

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
(base `origin/main` 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8; nothing staged or committed).

Scope of this cycle: close R3 from
`TASK-260720-2284br_review-verdict-cycle-2.md`. R1 and R2 were accepted closed in
that verdict and are untouched here.

## R3 — root cause

`projectAttempt` reads the machine-level hybrid activation manifest at step 5
(`scopes.LoadHybridDecls(cfg.Home())`) and merges the applicable declarations
into the effective closure. The optimistic observation set was built at step 16
and only ever contained installed marker generations plus per-build cache
outcomes. The hybrid manifest was therefore consulted outside the home lock and
never rechecked inside it.

That path is genuinely unlocked on both sides: `curator hybrid add` and
`curator hybrid rm` call `scopes.AddHybridDecl` / `scopes.RemoveHybridDecl`
directly from `cmd/curator/main.go` (lines 985 and 994) and take no project,
key, or home lock at all. So a declaration could be added, removed, or
retargeted at any point between closure resolution and the commit, and the run
would commit the closure it planned against — installing or omitting a hybrid
context and its adapter mirror against current machine truth.

## Fix

`internal/install/install.go`

- The observation set and its path map are now created at step 5 instead of
  step 16, and `scopes.HybridManifestPath(cfg.Home())` is observed under the key
  `activation/hybrid-manifest` before the manifest is read.
- Ordering is deliberate: **digest, then read**. If a writer races the read, the
  recorded digest no longer matches the file and the recheck restarts. The
  reverse order would record the writer's bytes while the closure was derived
  from the bytes actually read — a stale commit that the recheck could not see.
  The cost is a bounded spurious restart, which is always the safe direction.
- Step 16 keeps appending marker generations to the same set; only its comment
  changed.

`internal/install/commit.go`

- New `hybridActivationKey` constant.
- The `observations` doc comment now carries the complete closure-input audit
  (below). No behavioural change: `recheck` already maps any generation
  difference to `restartClosure`, and `runCommit` already calls it after journal
  recovery and before both cache publication and target staging.

`internal/install/global.go`

- Comment only: the machine-wide scope consults no hybrid activation manifest
  (hybrid declarations activate against a project), and its own Skillfile lives
  under `GlobalRoot(home)`, which is the identity it already holds the canonical
  operation lock on.

Restart re-derivation was verified rather than assumed: `runWithRestarts`
re-enters `projectAttempt`, which re-reads `LoadHybridDecls`, so a restart
produces a genuinely fresh closure.

## Closure-input audit (verdict item 4)

Every input consulted outside the home lock, and why it is safe:

| Input | Treatment |
| --- | --- |
| Hybrid activation manifest | **Observed** (`activation/hybrid-manifest`) — mutated under no lock by `curator hybrid add/rm` |
| Installed markers (project / hybrid / global stores) | Observed, one key per closure node |
| Protected build-cache verdicts | Observed per planned build as outcomes, re-derived through the publishing store |
| Staged target owners and preimages | Digested under the home lock in `journalPlan`; no optimistic key needed |
| Project `Skillfile.json`, `Skillfile.dev.json`, `.gitignore` | In the checkout, covered by the canonical project operation lock held from before planning until after handoff |
| Global `Skillfile.json` | Under `GlobalRoot(home)`, which is the identity `Global` takes that same operation lock on |
| Skill snapshots | Commit-addressed and written once, so the resolved tree cannot change under the run. A tag that moves after resolution leaves this run committing the commit it pinned — a complete, self-consistent install; the next run resolves the new commit and the moved-tag gate reports it |
| User configuration | Loaded once by the process entrypoint and passed through as one immutable value, so every phase and every restart attempt reads the same snapshot and the run is internally consistent. **Documented boundary, not a proof of freshness:** observing it would not help, because a restart re-reads no configuration and a concurrent edit could only become a restart loop. Picking up a mid-run configuration edit needs per-attempt reloading, which is a separate decision from this revalidation contract |

The two other `LoadHybridDecls` callers (`cmd/curator/main.go:1000,1010`) are
the read-only `hybrid list` / `hybrid status` commands and are not install paths.

## Tests

New: `internal/install/atomicity/activation_test.go` (4 cases). The suite was
chosen over `internal/install` because each case drives a complete real
installation through the exported API only, which is exactly what
`internal/install/atomicity/doc.go` states the package is for.

- `TestStableHybridActivationCommitsWithoutRestarting` — the control. An
  unchanged manifest must **not** restart; without it the three mutation cases
  would pass even if every install restarted unconditionally.
- `TestHybridDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleContext`
- `TestHybridDeclarationRetargetedBeforeHomeLockRestartsAndCommitsNoStaleContext`
  — the declaration still exists and only its target list moved, so a mere
  "is it still declared" check cannot pass this one.
- `TestHybridDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure` —
  the opposite direction, and it also covers the absent-to-present transition of
  the manifest file itself.

Each mutation case asserts the restart was reported, named closure resolution,
named `activation/hybrid-manifest`, that no stale hybrid context and no stale
adapter mirror committed (Lstat/Readlink based via `adapterEntryState`, so a
link is never dereferenced into looking like the tree behind it), that the
project's own declaration still installed, and that no transaction journal
survived.

`internal/install/atomicity/fixture_test.go` gained `hybridDeclareTargeting`;
the existing `hybridDeclare` now delegates to it.

### Negative control

The regression was proven to fail without the fix. With the single
`observed.observe(hybridActivationKey, ...)` line removed and everything else
identical, all three mutation cases fail with
`a hybrid activation change did not restart closure resolution` and the control
still passes (`negative-control.log`). The line was then restored and the build
re-verified.

The reviewer's own overlay repro was also run unmodified against the fix and now
passes (`reviewer-overlay-now-passes.log`):
`go test -overlay .temp/TASK-260720-2284br/review/overlay.json ./internal/install
-run TestReviewerStaleHybridActivationRestarts` — PASS.

## Reviewer attention

`internal/install/atomicity` now runs 405.4s, up from 360.1s, against the
default 600s per-package test timeout. Still green with no override, but the
margin is down to roughly a third. The next case added to that package should
come with a decision about splitting it rather than another four installs.
