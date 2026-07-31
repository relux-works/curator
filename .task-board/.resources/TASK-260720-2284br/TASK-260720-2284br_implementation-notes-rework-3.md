# TASK-260720-2284br — rework cycle 3 implementation notes

Input: `TASK-260720-2284br_review-verdict-cycle-3.md` (changes requested, R4).

Where the work lives: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
(base HEAD / `origin/main` `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`). Nothing
committed or staged. This document supplements the earlier notes
(`..._implementation-notes.md`, `..._implementation-notes-rework-1.md`,
`..._implementation-notes-rework-2.md`); everything they describe still holds
except where the closure-input audit is corrected below.

Scope of this cycle: close R4 only. R1, R2 and R3 were accepted closed in the
cycle-3 verdict and their behaviour is untouched.

## R4 — root cause

The cycle-2 audit classified the project and global manifests by **where they
live** rather than by **who writes them**, and both classifications were wrong:

- The project `Skillfile.json` sits in the checkout the run holds the canonical
  project operation lock on. But `curator add` and `curator remove` call
  `manifest.AddDecl` / `manifest.RemoveDecl` straight from
  `cmd/curator/main.go:432,460` and take no lock at all, so the lock witnesses
  nothing about that file.
- The global `Skillfile.json` sits under `GlobalRoot(home)`, which *is* the
  global scope's own operation identity. Same problem: `curator global add` and
  `curator global remove` write it lock-free (`cmd/curator/main.go:872,883`).

Both were read before private staging, both could move before the manager-home
lock, and neither was ever rechecked. Both scopes then committed a stale
closure — installing a declaration that had been removed, or omitting one that
had been added, together with its shims and adapter mirrors.

The same re-audit surfaced a third input in the same class, which the previous
cycles had not called out at all: `Skillfile.dev.json`. It has *no* Curator
writer, so nothing serializes it against an installation whatsoever, and a
substitution redirects a declaration at a local checkout or another ref — it
selects installed content exactly like the manifest, and it also decides the
strict-audit refusal. It is now observed too.

## Fix

`internal/install/install.go` (`projectAttempt`)

- The observation set is created at the top of the attempt, before the first
  declaration input is read, instead of at step 5.
- `manifest/project` is observed before `manifest.Load` (step 1).
- `substitutions/project` is observed before `devsub.Load` (step 4).
- `activation/hybrid-manifest` is observed before `LoadHybridDecls` (step 5), as
  before.
- Ordering is unchanged and deliberate: **digest, then read**. A writer that
  races the read leaves a recorded digest that no longer matches the file, so
  the recheck restarts. The reverse order would record the writer's bytes while
  the closure was derived from the bytes actually read — a stale commit the
  recheck could not see. The cost is a bounded spurious restart, which is always
  the safe direction.

`internal/install/global.go` (`globalAttempt`)

- The observation set is created at the top and `manifest/global` is observed
  before `manifest.Load(GlobalRoot(home))`.
- The old comment claiming the operation lock covered it is replaced with the
  reason it does not.

`internal/install/commit.go`

- New keys `projectManifestKey`, `globalManifestKey`, `substitutionsKey`
  alongside `hybridActivationKey`, documented by their actual writers.
- The `observations` doc comment carries the corrected audit (below).
- Simplification: `observations` now owns the path to re-read each key.
  `observe` writes `generations` and `paths` together, `recheck` iterates its
  own keys, and the parallel `generationPaths` map — plus the
  `commitRequest.generationPaths` field and both call sites' bookkeeping — is
  gone. The invariant that used to be conventional ("every observed key must
  also be registered in the paths map, or it is silently never rechecked") is
  now structural: it is impossible to capture an observation without the
  location that revalidates it.

No change was needed to the restart routing: `recheck` already maps any
generation difference to `restartClosure`, and `runCommit` already calls it
after journal recovery and before both cache publication and target staging
(verified again this cycle at `commit.go` — acquire home lock → `Journal.Recover`
→ `observed.recheck` → stage root → `publishWinners` → `stageTargets` → consumer
ledger last).

Restart re-derivation was verified rather than assumed: `runWithRestarts`
re-enters `projectAttempt` / `globalAttempt`, both of which re-read their
manifest from disk, so a restart produces a genuinely fresh closure. A manifest
that disappears between attempts correctly yields `skipped`.

## Corrected closure-input audit (verdict item 4)

Re-derived from the actual writers, not from directory location. `Yes` in the
"Serialized?" column means some lock or write-once property genuinely prevents
the input from moving under a run.

| Input | Real writers | Serialized? | Treatment |
| --- | --- | --- | --- |
| Project `Skillfile.json` | `curator add` / `curator remove` → `manifest.AddDecl` / `RemoveDecl`, no lock | **No** | **Observed** `manifest/project` |
| Global `Skillfile.json` | `curator global add` / `remove` → same functions, no lock | **No** | **Observed** `manifest/global` |
| `Skillfile.dev.json` | none — hand-edited | **No** | **Observed** `substitutions/project` |
| Hybrid activation manifest | `curator hybrid add` / `rm`, no lock | **No** | **Observed** `activation/hybrid-manifest` |
| Installed markers (project / hybrid / global) | other installs | Under the home lock only | **Observed**, one key per closure node |
| Protected build-cache verdicts | other installs publishing winners | Under the home lock only | **Observed** per planned build as outcomes; re-derived through the publishing store |
| Staged target owners and preimages | other installs | Under the home lock | Digested inside `journalPlan` under the home lock; no optimistic key needed |
| Skill snapshots | written once per commit | **Yes**, commit-addressed | Not observed. A moved tag leaves this run committing the commit it pinned — complete and self-consistent; the next run resolves the new commit and the moved-tag gate reports it. A `path` substitution is covered too: `closure.resolveNode` resolves the local checkout to a concrete HEAD commit and snapshots it under that commit like any other source, so only *which* checkout it points at needs observing, and that is the substitution manifest above |
| Registry attestations, MCP verification results | registry service; agent configuration files | **No** | Not observed, and no lock claimed. They select nothing — the closure is already resolved when they run — and each is recorded into the marker as evidence of what this operation found. Later drift leaves this run committing the evidence it pinned, exactly like a snapshot |
| Audit trust state | `curator audit pin`, outside any operation | **No** | Not observed, and no lock claimed. It is a gate on the resolved closure, not an input to it: it can refuse an installation but cannot change which skills it contains. An approval revoked after the gate ran leaves a complete installation the next run's gate refuses. The gate memoizes its own verdicts beside that state; that cache is derived from content hashes and is not an installation target |
| Managed `.gitignore` | `curator init`, and **this operation itself** when `--fix-gitignore` is set | **No** | Deliberately not observed, and no lock claimed. It is a hygiene precondition, not a closure input — its content selects nothing that gets installed — so a concurrent edit cannot make this run commit the wrong closure, and the next run re-runs the gate. Observing a file the run itself writes would only manufacture restart loops |
| User configuration | the user, between runs | Loaded once by the entrypoint | Not observed. **Documented boundary, not a proof of freshness:** it is passed through as one immutable value, so every phase and every restart attempt reads the same snapshot and the run is internally consistent. Observing it would not help, because a restart re-reads no configuration and a concurrent edit could only become a restart loop. Picking up a mid-run configuration edit needs per-attempt reloading, which is a separate decision from this revalidation contract |

## Tests

New: `internal/install/revalidation_test.go` (6 cases), placed in
`internal/install` beside the closest sibling regression
(`TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan`) rather than
in `internal/install/atomicity`. That is a deliberate response to the cycle-2
reviewer note about the atomicity package's runtime margin: adding six more
complete installations there would have pushed a package that already runs 411s
plain and 427s under `-race` towards the 600s default timeout. Every mutation
goes through the **real** writer named in the finding (`manifest.AddDecl` /
`manifest.RemoveDecl`), not a hand-written file, and fires from `OnStaged` —
after every private build succeeded, before the commit phase takes the home
lock.

- `TestStableDeclarationInputsCommitWithoutRestarting` — the control, covering
  both scopes. Without it the five mutation cases would pass even if every
  install restarted unconditionally.
- `TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState`
- `TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure` —
  the opposite direction, so the fix cannot be a one-way "did anything vanish"
  check.
- `TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState`
- `TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure`
- `TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution`

Each mutation case asserts the restart was reported, named closure resolution,
named its exact observation key, and that all three target classes a stale
closure would corrupt are correct afterwards: the context, the shim, and the
adapter mirror (project mirrors in the checkout, global mirrors in the user
home). Absence is checked with `Lstat`, so a mirror link counts as present
instead of resolving into the tree behind it. Each also asserts no transaction
journal survived the discarded attempt.

### Negative control

The regressions were proven to fail without the fix. With the three
`observed.observe(...)` lines for `projectManifestKey`, `globalManifestKey` and
`substitutionsKey` removed and everything else identical (read-only `-overlay`,
no product file touched), all five mutation cases fail with
`a manifest/project|manifest/global|substitutions/project change did not restart
closure resolution` and the control still passes:
`gates-cycle4/negative-control.log`, exit 1.

### Reviewer's own repro

The cycle-3 overlay was run unmodified against the fix and now passes:

```text
go test -overlay .temp/TASK-260720-2284br/review/global-overlay.json ./internal/install \
  -run '^TestReviewerStale(Project|Global)ManifestRestartsClosure$' -count=1 -v
--- PASS: TestReviewerStaleGlobalManifestRestartsClosure (2.68s)
--- PASS: TestReviewerStaleProjectManifestRestartsClosure (2.74s)
```

## Gates

Two independent final-tree passes exist. The first (`gates-final-r4/`, 18:22-18:38)
completed every gate at exit 0, but its session was killed before it could
package evidence or hand off. The run that hands off therefore re-ran **every**
gate first-hand rather than inheriting that archive: `gates-r4-verify/`,
started 18:42:38, finished 19:03:41.

Same protocol as before: each gate is a standalone process, no `tee`, no pipe
chain, real exit code written to `<name>.exit` as the last action, so a missing
`.exit` reads as "killed", never as "passed". Newest product source is 18:15
(`install.go`, `global.go`); the run started at 18:42, so it covers the final
tree.

| Gate | Exit | Sec |
| --- | --- | --- |
| `gofmt -l .` | 0 | 0 |
| `git diff --check` | 0 | 0 |
| `go build ./...` | 0 | 1 |
| `go vet ./...` | 0 | 0 |
| `test-revalidation` — the 6 new R4 cases | 0 | 51 |
| `overlay-r4-manifests` — reviewer's cycle-3 repro, unmodified | 0 | 23 |
| `overlay-r3-hybrid` — reviewer's cycle-2 repro, unmodified | 0 | 14 |
| `test-install` | 0 | 369 |
| `test-atomicity` | 0 | 388 |
| `test-godriver` | **1** | 60 |
| `test-rest` — the remaining 37 packages | 0 | 28 |
| `race-revalidation` | 0 | 119 |
| `race-concurrency` | 0 | 61 |
| `race-activation` | 0 | 95 |
| `race-staging` — `staging` + `transaction` | 0 | 52 |
| `golangci-lint` v2.4.0 `run ./...` | 0 | 1 |

### `test-godriver` exited 1 — reported as red, diagnosed as machine contention

Not hidden and not claimed as passing in the driver run. Diagnosis:

- Every failure is `process_timeout: Go probe exceeded its deadline` at exactly
  ~15.02s. That is the hard wall-clock cap in `internal/godriver/session.go:34`
  (`defaultProbeTimeout = 15s`), which `session.go:136-137` also clamps as a
  maximum — a caller cannot raise it. No assertion failed anywhere.
- `internal/godriver` is sibling-owned (`TASK-260720-6i3cya`) and untouched by
  this task: its newest file is 08:01, hours before this cycle's 18:14-18:15
  edits. Nothing in this cycle's diff reaches it.
- The same tree passed this same gate at exit 0 in 62s at 18:32 in the earlier
  `gates-final-r4` pass.
- The machine was saturated by unrelated processes during the run —
  `gramdrive-agent` 99% CPU, `suggestd` 98%, `iTerm2` 74%, `fileproviderd` 61%,
  `duetexpertd` 37%, load average 9-12.

Three follow-up runs, all recorded in `gates-r4-verify/`:

| Run | Command | Exit | Sec |
| --- | --- | --- | --- |
| `test-godriver-rerun1` | default parallelism, still under load | 1 | 304 |
| `test-godriver-rerun2-parallel1` | `-parallel 1` | **0** | 26 |
| `test-godriver-rerun3-default` | default parallelism, load subsided | **0** | 25 |

Concurrent probe subprocesses competing for CPU against a fixed 15s wall-clock
deadline is sufficient and necessary to explain all of it: serializing the
package fixes it, and so does waiting for the machine to quiet down.

**Not claimed:** that the gate is green under load with default parallelism. It
is not. `go test ./...` on a loaded machine can fail this package. That is a
pre-existing property of a sibling-owned package, not a regression from this
cycle, and raising a security-relevant probe bound to make a test pass would be
exactly the kind of forced fit this task must not make. Flagged for the reviewer
and for `TASK-260720-6i3cya`'s owner rather than patched here.

### Lint

`golangci-lint` v2.4.0 `run ./...` exits 0 with `0 issues.` repo-wide, so
"Lint clean" is honestly true this cycle and the earlier deferred-lint
checklist items stay superseded.

The 1s wall clock is a cache hit, so it was re-verified: after
`golangci-lint cache clean`, `run ./...` still exits 0 with `0 issues.` in 3s
(`lint-repo-nocache.log`). Verbose output reports `Active 8 linters` and
`Issues before processing: 79, after processing: 0` — that filtering is **not**
this task's doing: `.golangci.yml` is byte-identical to the base commit
(`git diff 17804ce -- .golangci.yml` is empty, file dated Jul 21), its gosec
excludes (G306, G301, G122, G703, plus gosec-off-in-tests) predate this story,
and the 104 `#nosec` directives repo-wide each carry a per-site justification.

## Independent re-verification before handoff

The cycle-3 code was written by a session that was killed mid-gate-run. The
handing-off run re-checked the load-bearing claims first-hand instead of
trusting the notes above:

- **Reviewer's R4 repro passes.** Both `TestReviewerStaleProjectManifestRestartsClosure`
  and `...Global...` PASS against the unmodified cycle-3 overlay. The cycle-2
  hybrid repro also still passes.
- **`observe` really precedes the read at all four sites** —
  `install.go:177` before `manifest.Load` at 178, `217` before `devsub.Load` at
  218, `254` before `LoadHybridDecls` at 255, `global.go:79` before
  `manifest.Load` at 81. Digest-then-read, as documented.
- **The negative control is an honest minimal diff.** `diff` of the overlay
  sources against the real files is exactly three deleted `observed.observe`
  lines and nothing else. With them removed all five mutation cases fail and the
  control still passes (exit 1). The regressions therefore have teeth.
- **The writer claim is exact.** `manifest.AddDecl` / `RemoveDecl` have exactly
  four non-test callers — `cmd/curator/main.go:432,460,872,883` — and none of
  them takes any lock; each just resolves a target root and writes.
- **`OnStaged` fires at the boundary the finding describes** — `install.go:501`,
  immediately after `stageBuilds` and immediately before `runCommit`, which is
  where the home lock is acquired. The tests mutate in the right window.
- **Restart routing and ordering** — `runCommit` is home lock → `Journal.Recover`
  → `observed.recheck` → stage root → `publishWinners` → `stageTargets` →
  consumer merged last (`commit.go:454-504`). Restarts are bounded at
  `defaultMaxRestarts = 3`, so digest-then-read cannot spin forever.
- **Commit classes match the AC** — context 10, runtime 20, canonical shim 30,
  forwarding shim 40, env file 50, adapter ledger 60, mirror ledger 70, removal
  80, consumer 90 last (`internal/staging/staging.go:24-43`).
- **The closure-input audit is complete for the code as written.** The only
  pre-home-lock shared reads in `install.go` are the three observed ones, and in
  `global.go` the one observed one. `agents` and `aliases` — the other values
  feeding activation — derive from `projectManifest` (observed) and
  `cfg.Projects[alias]`, the immutable per-run configuration snapshot documented
  as a boundary rather than defended as fresh.
- **No reviewer scratch leaked into the tree.** The overlays map virtual paths;
  `internal/install/reviewer_*` does not exist on disk.

## Reviewer attention

- The `.gitignore`, audit-trust, registry/MCP and user-configuration rows above
  are documented boundaries, not freshness proofs. They are stated as such
  rather than defended with a lock that does not exist — that mis-statement is
  exactly what R4 was.
- `internal/install/atomicity` still runs ~388-411s plain against the 600s
  default timeout. This cycle deliberately added nothing to it; the next case
  that belongs there should still come with a splitting decision.
- `internal/godriver` exited 1 in the driver run under machine contention and
  passes in isolation. See the gates section — it is sibling-owned, untouched by
  this task, and deliberately not patched here. Worth its own task.
- Darwin/arm64 only. No Windows runtime evidence is claimed.
