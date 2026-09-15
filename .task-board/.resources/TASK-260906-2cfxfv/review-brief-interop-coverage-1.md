# Review brief: make a dropped environments family fail the candidate lane (cycle 1)

## Subject

- Branch `feat/interop-root-artifacts` at `31106aa8` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage`, two signed commits past
  `origin/main` (`fb916acd`). PR https://github.com/relux-works/curator/pull/63. Diff:
  `git diff origin/main..HEAD` — 11 files, +560/−64.
- Read `producer-brief-interop-coverage.md`, `producer-brief-interop-finish.md`, and
  `TASK-260906-2cfxfv_drafting-report.md`. The producer's raw evidence is in the worktree under
  `.temp/TASK-260906-2cfxfv/` — materialized roots, logs, and a mutant harness.
- Authority: curator-spec main `87a0d006`; the gate scripts themselves are the contract here —
  `suite-plan.sh`, `test-gate.sh`, `platform-case-gate.sh`, `skip-classes.tsv`, `root-artifacts.tsv`.
- Change Request revision to accept: the one recorded for TASK-260906-2cfxfv.

## What this leaf is for

A candidate root that stopped publishing `vectors/environments.json` would have passed the candidate
lane green with every environments case silently skipped, because `internal/interop` declared none of
its families in `root-artifacts.tsv` and `root-content` is policy `allow` in every lane. Nothing was
broken — the gate simply could not tell. This is the last package carrying that hole.

## Review dimensions

1. **Re-run the negatives yourself; they are the whole point.** The producer reports six — each of
   the four vector families and the two byte-exact trees removed in turn, every one `exit=1` naming
   the missing artefact. Reproduce at least three from scratch, and reproduce the positive: the
   unmodified candidate root green. Do not read the logs; run them. Materialize roots as **plain
   checkouts verified against `manifest.json`** — `git archive` corrupts a root because
   `export-subst` is set on `conformance/v1/fixtures/byte-exact/subst.txt`.

2. **Is the declaration complete, and is it honest?** The row claims every case in the new package
   reads one of the declared artefacts without a per-family guard. Verify that by reading the moved
   tests, not the row: if a case reads an artefact the row omits, the hole moved rather than closed.
   Conversely, if the row declares something no case reads, the package defers on a root that could
   have served it — say so.

3. **The split's blast radius.** Confirm the default lane keeps every non-environments
   `internal/interop` case it runs today: compare the served case list on the `SPEC_PIN` root before
   and after the split, by name. Confirm `internal/interop/golden_test.go`'s change does not weaken
   what the pinned Go manager consumes — curator-spec's own Implementations lane runs the pin against
   that package, and this repository has lost a lane to exactly that before.

4. **The contract test.** `TestNoCaseHereSkipsForAnythingButTheDeferredRoot` claims the only skip this
   package can take is the deferred-root one. Attack it: add a case to the package that skips for a
   different reason and confirm the contract test fails. If it can be satisfied while a silent skip
   exists, it is decoration.

5. **`gate-selftest.sh` grew by 71 lines.** Read what those assertions actually assert, and mutate the
   thing each is meant to protect to confirm each can fail. A self-test that cannot fail is the same
   defect this leaf exists to close, one level up.

6. **The ledger.** Three rows were added to `platform-cases.tsv`. Confirm each describes what its test
   asserts, and that the tolerated classes are right for a package that is now deferred rather than
   `root-content`-skipped.

7. **Hosted lanes.** Read `gh pr checks 63`; a red lane is blocking. Then say whether a candidate
   dispatch is needed to prove this leaf, and if so, say which artefact you would remove — the
   orchestrator will run it.

## Method

- Drive the gate scripts, not descriptions of them.
- Anchor each `-run` level separately and count `=== RUN` lines: `go test -run '^(Parent/child)$'`
  splits on the slash, matches nothing and **exits 0**, which reads as a survivor.
- Run the two `test-gate` lanes sequentially, never alongside a `-race` suite.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root.

## Verdict contract

Attach `TASK-260906-2cfxfv_review-findings-1.md`. Blocking or major → set the task to `development`.
Otherwise an explicit ACCEPT at `to-review` with `accept_cr` on the recorded revision, stating what
the hosted lanes reported. Do not mark the task done. Then
`task-board handoff TASK-260906-2cfxfv --role reviewer`.
