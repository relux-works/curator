# TASK-260824-1kzj22 — review verdict: ACCEPTED (with two orchestrator dispositions)

Reviewer run: RUN-260824-69b8fd (claude, claude-opus-5). Read-only; no file was
staged, committed, reset, or cleaned.

**Verdict: ACCEPTED.** The parallelization is correct, safe, maximal for its
tree, and provably regression-free. The "≤ 4 min is unreachable in scope" claim
is not an excuse — I reproduced it arithmetically from my own measurements, and
it is in fact *stronger* than the producer stated. Two items need an
Orchestrator decision (they are not producer rework): the unmet AC target, and
an integration/branch-base problem described in §D.

---

## A. The parallelization is correct and safe — CONFIRMED

### A1. Every parallelized case is genuinely independent

I did not spot-check. I built the **transitive call closure** of each of the 36
parallelized tests over every helper defined in `cmd/curator/*_test.go` and
scanned the whole closure for `t.Setenv`, `os.Setenv`, `os.Chdir`, `capture(`,
`os.Stdout`, `os.Stderr`, and assignment to `resolveCLIProvider`.

**Result: zero hits across all 36 closures.** No parallelized case reaches
process-global state directly or through a helper. Every filesystem fixture is
`t.TempDir()`. No test in the package binds a port.

The producer's claim that Go panics on `t.Setenv` + `t.Parallel()` is correct in
both directions (`t.Setenv` panics if the test or an ancestor is parallel;
`t.Parallel()` panics if `t.Setenv` was already called), so the env column is
fatal, not merely risky. Verified no parallelized case sets env.

Two cases worth naming because they are the least obvious and both check out:

- `TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery` shells
  out twice. It builds into `t.TempDir()` and drives the subprocess with a fully
  explicit `command.Env = []string{...}` — it never mutates the parent process
  environment. Concurrent `go build` is safe (the Go build cache is
  concurrency-safe).
- The six cases that call `run()` and let it write to the *real* `os.Stdout`
  (`TestRunVersionExitsZero`, `TestRunNoArgsPrintsUsage`, `TestRunUnknownCommand`,
  `TestShellInitPrintsHooks`, `TestSkillCheckOnTempDir`,
  `TestHiddenWorkerModeIsNotAUserVisibleCommand`). The load-bearing invariant the
  producer documented on `capture` is **correct**: `t.Parallel()` signals the
  parent `tRunner` and blocks on the parent's barrier, which is only released
  after the parent's body — i.e. the entire serial pass of top-level tests — has
  returned. A parallel case therefore can never overlap a serial case holding a
  swapped stream. Documenting this at `capture` rather than leaving it implicit
  is the right call; it is the one thing a future contributor would get wrong.

The single package-level mutable seam, `resolveCLIProvider`
(`cmd/curator/assurance.go:17`), is written only by the two serial cases the
producer named. There is no `flag.CommandLine`/`flag.Parse` usage in the package,
so no hidden global parser state either.

### A2. Zero new flaky failures — CONFIRMED, and I reran everything myself

Rather than rerun the ~400 s monolith inside a 10-minute call cap, I split the
package into three bounded runs that together cover **all 69 test cases**:

| My independent run | Cases | Wall-clock | Result |
| --- | ---: | ---: | --- |
| `go test -race -count=1 -run '^(<the 36>)$'` | 36 | 6.8 s | ok, **0 `DATA RACE`** |
| `go test -count=1 -run '^(<32 remaining serial>)$'` | 32 | 153.9 s | ok |
| `go test -count=1 -run '^TestCompiledProjectStatusRepairRollbackRecovery$'` | 1 | 219.4 s | ok |

I verified the `-run` regex selects **exactly 36** top-level cases (`=== RUN`
count on a `-v` pass), so the green race run is not an empty selection.

Accepted without rerunning: the producer's three consecutive full runs
(543.7 / 374.0 / 470.3 s, all exit 0) and the three baseline runs
(326.6 / 305.6 / 522.5 s, all exit 0). Independently corroborated by the
orchestrator monolith `TASK-260824-1kzj22_full-go-01.log` — I recomputed its
SHA-256 as `f78708506bf91110e8d5faaa149d179bf74e4668c3fe0aee39df9506b3d4edd3`,
**matching the attached description**; it shows 54 `ok` packages, 0 `FAIL`, and
`cmd/curator 399.683s` on the patched tree.

### A3. Coverage unchanged, and no assertion weakened — CONFIRMED

I did not take the 57.1 % figure on trust. I parsed both profiles from the
evidence bundle and diffed them block by block myself:

```
blocks before 778   after 778
only-before 0       only-after 0
blocks differing in covered-state: 0
stmts covered before 695/1218   after 695/1218
```

Exact agreement. And the mechanism check that makes this airtight: the diff is
**45 insertions, 0 deletions** across 4 `_test.go` files. No assertion was
removed, no `t.Skip` was added, no case was dropped from the run set. Speed was
not bought with weakened evidence.

### A4. Gates — CONFIRMED, rerun by me

| Gate | My result |
| --- | --- |
| `gofmt -l cmd/curator/` | clean, no output |
| `go vet ./cmd/curator/` | exit 0 |
| `golangci-lint run ./cmd/curator/...` | `0 issues.` — binary reports `version 2.12.2 built with go1.25.5`, i.e. the pinned version |
| `git diff --check` | clean |
| `task-board validate` | `Board is valid. No issues found.` |
| production files touched | **none** — diffstat is 4 `_test.go` files only |

---

## B. The "≤ 4 min unreachable in scope" claim is HONEST — and understated

I checked each named claim against the source rather than the narrative.

| Claim | Verification |
| --- | --- |
| `run()` reads process-global `CURATOR_CONFIG` | `cmd/curator/main.go:94` — `func run(args []string) int`, takes only `args`; `loadConfig()` at `main.go:151` calls `config.Load("", ...)`; `internal/config/config.go:180` reads `os.Getenv("CURATOR_CONFIG")`. **TRUE.** |
| `capture` swaps process-global streams | `cmd/curator/status_test.go:40` — replaces `os.Stdout`/`os.Stderr` for the call. **TRUE.** |
| All 33 serial cases are serial *by construction* | I ran the same transitive-closure scan over the serial set. **All 33 hit `t.Setenv` and/or `capture`/`os.Stdout`. Zero of them lack a justification.** The producer did not leave safe parallelism on the table, and did not over-claim a single row. |
| The heavy test's 5 subtests share one mutated fixture | `cmd/curator/status_test.go:433` — one `fixture := newInstalledCompiledProject(t)` is passed to all five `t.Run` closures, which corrupt and repair it (`assertCorruptCompiledStateIsRepaired`, `assertUntrustedCompiledStateIsRepaired`). **TRUE — it cannot be split.** |
| Cost is cold Go compilation, not test overhead | `internal/godriver/session.go` `bootstrapEnvironment` pins `GOPATH`/`GOMODCACHE`/`GOCACHE`/`GOTMPDIR`/`HOME` to a per-session hermetic layout; `internal/godriver/guards_test.go:60` rejects `GOCACHE=os.TempDir()` with `CodeWorkerProtocolInvalid`. **TRUE** — the hermetic cache is a deliberately enforced guard, so the compile cost is real work, not slack. |

### The arithmetic proof the producer did not state

From my own three bounded runs, the package floor with `t.Parallel()` alone is:

```
219.4 s  (TestCompiledProjectStatusRepairRollbackRecovery, unsplittable)
+ 153.9 s  (32 remaining serial cases, t.Setenv-bound)
= 373.3 s  floor   >   240 s target
```

**Even if the 36 parallelized cases cost literally zero, the package cannot go
below ~373 s.** The AC is not "hard on a loaded machine" — it is arithmetically
out of reach for any test-only change. My independent 373 s floor also lands
almost exactly on the orchestrator's measured 399.7 s, so the decomposition is
sound and not an artifact of host contention.

### One correction to the producer's option table

The producer marks Option A (injectable config path + streams) as "Reaches
≤ 4 min? **Yes**". That is too confident. Even with every case parallelizable,
the package floor becomes the single heaviest test, which cannot be split:
**219 s by my measurement, 230–268 s by the producer's.** The producer's own
numbers therefore put the post-refactor floor *above* the 240 s target. The
target sits inside the measurement noise of one irreducible test.

This does not weaken the acceptance — it strengthens the case that the AC number
itself is the defect.

---

## C. Recommendation on the unmet target

**Do both, and the AC amendment is the load-bearing one.**

1. **Amend the AC** to the measured floor rather than a number that no test-only
   change can satisfy. Suggested wording: *"cmd/curator wall-clock reduced as far
   as `t.Parallel()` allows, with the residual serial bottleneck named and
   measured; ~373 s floor recorded, dominated by
   `TestCompiledProjectStatusRepairRollbackRecovery` (219–268 s) whose five
   subtests share one mutated fixture."* Record that ≤ 4 min is below the
   single-test floor and is therefore not a meaningful target for this task class.
2. **File a follow-up task** for the production refactor — Option A, injectable
   config path and output streams on `run()`. It is ordinary Go CLI design, it
   removes both couplings at once, and it leaves coverage attribution intact.
   Scope it honestly: it unblocks parallelism for ~33 more cases but will land
   around 220–270 s, not under 240 s. Getting genuinely under 4 min additionally
   requires splitting `TestCompiledProjectStatusRepairRollbackRecovery`'s shared
   fixture, which is a separate design question.

I am **not** holding acceptance hostage to either. The delivered work is the
correct, complete, maximal test-only change.

---

## D. Integration finding the commit-owning mover MUST handle

This is not producer rework, but it will silently lose the work if nobody acts.

1. **The change is uncommitted, and it is not in the assigned story worktree.**
   The story workspace named in the assignment
   (`.temp/STORY-260811-2epsp4/worktree`, branch
   `task-board/story/STORY-260811-2epsp4`) is at `903af23` (main head), is
   **clean**, and contains **zero** `t.Parallel()` in `cmd/curator`. Its reflog
   shows it was created and reset to HEAD at 03:32, ~1 min before this review run.
   The delivered change lives as **uncommitted working-tree modifications** in
   `/Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-skill` on branch
   `codex/legacy-board-repair` at `6f93b51` — which is also where `TASK_BOARD_DIR`
   points and where the STORY-260811-2epsp4 adapter delivery commits (`f8b7cc7`,
   `6f93b51`) actually live. The producer worked in the right place; the assigned
   story worktree is the stale artifact. Reviewed there accordingly.

2. **The patch does not apply cleanly to main.** `codex/legacy-board-repair` and
   main have diverged (14 main-only commits, 6 codex-only). Dry-run on the story
   worktree at `903af23` (`git apply --check`, non-destructive, nothing written):

   ```
   builds_test.go          applies cleanly
   status_test.go          applies cleanly
   toolchain_host_test.go  applies cleanly
   main_test.go            DOES NOT APPLY — 3-way applies with conflicts
   ```

   `cmd/curator/main_test.go` differs by 691 lines between the two bases. Whoever
   lands this needs a 3-way apply plus a manual resolution on that one file.

3. **Seven `cmd/curator` tests on main were never inventoried.** The producer's
   isolation review is complete and correct for its tree (69 cases + `TestMain`;
   the task description's "70 Test functions" counts `TestMain`). Main carries
   76 cases, including `TestInstallPrintsTheRemedyAVersionManagerSelectionEarns`
   in `toolchain_remedy_test.go` (a file that does not exist on the codex branch)
   and six additional cases in `main_test.go`. If the branches merge, those seven
   need a parallelization/isolation pass. Fold this into the follow-up task.

---

## E. Definition-of-Done disposition

| DoD item | Status |
| --- | --- |
| Independent cases marked `t.Parallel()` with correct isolation; test code only | **MET** — verified by transitive closure scan, not spot-check |
| Isolation review documented, both directions | **MET** — all 69 cases accounted for; serial justifications independently confirmed, zero over-claims |
| Wall-clock ≤ 4 min (or closest reachable with bottleneck named) | **Target NOT met; the "closest reachable, bottleneck named" branch IS met.** Floor proved at ~373 s. Needs Orchestrator disposition (§C). |
| Zero new flaky failures across 3 consecutive runs; coverage unchanged | **MET** — 3 producer runs + orchestrator monolith all exit 0; coverage diffed block-for-block by me, exact |
| Focused `-race` run green | **MET** — rerun by me: 36 cases, 6.8 s, 0 `DATA RACE` |
| gofmt, vet, golangci-lint v2.12.2, `git diff --check`, `task-board validate` | **MET** — all rerun by me |
| Evidence logs + developer outcome attached | **MET** — bundle complete; monolith SHA-256 verified against its description |
| Implementation matches AC / fits architecture / tests green | **MET**, except the AC's numeric target per §C |

**Accepted.** Hand to the commit-owning mover with §D as a blocking integration
checklist and §C as the target disposition.
