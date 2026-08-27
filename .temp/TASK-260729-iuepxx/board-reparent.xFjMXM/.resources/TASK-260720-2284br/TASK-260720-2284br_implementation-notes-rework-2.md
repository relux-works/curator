# TASK-260720-2284br — rework cycle 2 implementation notes

Input: `TASK-260720-2284br_review-verdict-cycle-2.md` (changes requested, R3) and
the cycle-2 rework + completion directives of 2026-07-28.

Where the work lives: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
(base HEAD / `origin/main` `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`). Nothing
committed or staged. This document supplements
`TASK-260720-2284br_implementation-notes.md` and
`..._implementation-notes-rework-1.md`; everything they describe still holds.

Scope of this cycle: close R3 only. R1 and R2 were accepted closed in the cycle-2
verdict and are untouched.

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
key, or home lock at all. A declaration could be added, removed, or retargeted at
any point between closure resolution and the commit, and the run would commit the
closure it planned against — installing or omitting a hybrid context and its
adapter mirror against stale machine truth.

## Fix

`internal/install/install.go`

- The observation set and its path map are now created at step 5 instead of
  step 16, and `scopes.HybridManifestPath(cfg.Home())` is observed under the key
  `activation/hybrid-manifest` before the manifest is read.
- Ordering is deliberate: **digest, then read**. If a writer races the read, the
  recorded digest no longer matches the file and the recheck restarts. The
  reverse order would record the writer's bytes while the closure was derived
  from the bytes actually read — a stale commit the recheck could not see. The
  cost is a bounded spurious restart, which is always the safe direction.
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

Verified ordering inside the home lock (`commit.go:402-445`):
acquire home lock → `Journal.Recover` → `observed.recheck` → create stage root →
`publishWinners` → `stageTargets` → consumer ledger last. The recheck therefore
precedes both cache publication and target staging, as the verdict required.

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

Independently re-verified this cycle: `LoadHybridDecls` / `HybridManifestPath`
have exactly one non-test install-path caller (`install.go:237,239`). The other
two callers (`cmd/curator/main.go:1000,1010`) are the read-only `hybrid list`
and `hybrid status` commands. `internal/install/global.go` reads no hybrid
manifest and acquires its operation lock on `GlobalRoot(home)` (line 46) before
any manifest read (line 61+).

## Tests

New: `internal/install/atomicity/activation_test.go` (4 cases). The suite was
chosen over `internal/install` because each case drives a complete real
installation through the exported API only, which is what
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
adapter mirror committed (Lstat/Readlink based via `adapterEntryState`, so a link
is never dereferenced into looking like the tree behind it), that the project's
own declaration still installed, and that no transaction journal survived.

`internal/install/atomicity/fixture_test.go` gained `hybridDeclareTargeting`;
the existing `hybridDeclare` now delegates to it.

### Negative control

The regression was proven to fail without the fix. With the single
`observed.observe(hybridActivationKey, ...)` line removed and everything else
identical, all three mutation cases fail with
`a hybrid activation change did not restart closure resolution` and the control
still passes (`gates-cycle2/negative-control.log`). The line was then restored
and the build re-verified.

The reviewer's own overlay repro was also run unmodified against the fix and now
passes (`gates-cycle2/reviewer-overlay-now-passes.log`):
`go test -overlay .temp/TASK-260720-2284br/review/overlay.json ./internal/install
-run TestReviewerStaleHybridActivationRestarts` — PASS.

## Gates — all re-run on the final tree, real exit codes

Every gate below ran as a standalone process; no `tee`, no pipe chain. Each
records its real exit code to a durable `<name>.exit` file written as the last
action, so a missing `.exit` means "killed or still running", never "passed".
Archive: `TASK-260720-2284br_gate-evidence-rework-2.tar.gz` (`gates-cycle3/`).

| Gate | Command | Exit |
| --- | --- | --- |
| gofmt | `gofmt -l .` | **0** (no output) |
| diff check | `git diff --check` | **0** (no output) |
| build | `go build ./...` | **0** |
| vet | `go vet ./...` | **0** |
| tests — atomicity | `go test -count=1 ./internal/install/atomicity` | **0** (411.3s) |
| tests — install | `go test -count=1 ./internal/install` | **0** (236.0s) |
| tests — other 38 pkgs | `go test -count=1 <all remaining>` | **0** (38/38 ok) |
| race — activation | `go test -race -count=1` over the 4 new activation cases | **0** (427.0s) |
| race — restart/concurrency | `go test -race -count=1` over concurrent-consumers, recovery-before-mutation, stale-generation-restart | **0** (56.3s) |
| race — staging/transaction | `go test -race -count=1 ./internal/staging ./internal/transaction` | **0** |
| lint | pinned `golangci-lint` v2.4.0 `run ./...` | **0** — `0 issues.` |

Together the three test chunks are the complete `go test ./...` package set (40
packages), split only so each command completes inside the harness limit.

Lint is genuinely clean repo-wide this cycle: `0 issues`, exit 0. This supersedes
every earlier deferred-lint note on this task; the 45 inherited findings were
cleared during rework cycle 1 and have not returned.

### Honesty notes

- An earlier attempt to run `go test ./...` as one shell was **killed at exit
  144** before finishing. It is not evidence and is not claimed. The file is kept
  as `gates-cycle3/KILLED-not-evidence_full-suite-exit144.log` and is replaced by
  the three recorded chunks above.
- The cycle-2 archive's `gate-test-race.log` is empty for the same reason
  (externally killed shell) and is likewise not claimed. The race evidence for
  this handoff is the three recorded race gates above.
- `gates-cycle2/gate-test-full.log` shows every package green, but it finished at
  17:02:45 while `commit.go` was last modified at 17:15:33, so it does **not**
  cover the final tree. That is exactly why the suite was re-run here rather than
  inherited.
- `setsid` does not exist on macOS, so the first detached-runner attempt produced
  nothing at all. Gates were re-run in bounded foreground commands instead.

## Reviewer attention

- `internal/install/atomicity` runs 411.3s plain and 427.0s for the four
  activation cases alone under `-race`, against the default 600s per-package test
  timeout. Still green with no override, but the plain-run margin is down to
  roughly a third and CI runners are slower. The next case added to that package
  should come with a decision about splitting it rather than another four
  installs.
- The user-configuration row in the audit table is a documented boundary, not a
  freshness proof. Per-attempt configuration reloading is deliberately out of
  scope for this revalidation contract.
- Darwin/arm64 only. No Windows runtime evidence is claimed.
- Hardened containment (STORY-260728-327soo) remains separate and non-gating; no
  kernel/container isolation is claimed here.
