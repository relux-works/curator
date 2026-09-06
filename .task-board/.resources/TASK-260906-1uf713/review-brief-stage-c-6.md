# Review brief: stage (c) cycle 6 (rework 5, landing decision)

## Subject

- Branch `feat/agent-environments-stage-c` at `8b8aa041` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, 32 signed commits past
  `origin/main`. PR https://github.com/relux-works/curator/pull/61. Rework 5 is
  `git diff 71e6baec..HEAD`.
- **All eleven hosted lanes are green on this exact head**; the orchestrator waited for them before
  spawning you. A candidate dispatch against the **task authority** `550579d` was started at the same
  time (run 34058365116) — read it, since that is the root this stage is written against and the one
  its AC names.
- Read `TASK-260906-1uf713_review-findings-stage-c-5.md` and its probes archive, then
  `producer-brief-stage-c-rework-5.md` and `TASK-260906-1uf713_rework-report-5.md`.
- Authority: curator-spec **`550579d`** for this stage's conformance subset, with `87a0d006` (main
  today) read only for the known consumption gap described below.
- Change Request revision to accept: the one recorded for TASK-260906-1uf713.

## Standing, and the one thing already decided

Five cycles, twenty findings. Cycle 5 confirmed every cycle-4 finding fixed and this the first fully
green head. Do not re-derive that.

**The candidate lane against curator-spec main is red, and that is a decided, filed matter, not a
finding for this cycle.** All three of its jobs fail only in `internal/config`, on exactly the 30
subcases the `path`-overlay reconciliation added after this stage's authority was frozen. The
orchestrator landed that reconciliation mid-flight; consuming it means porting the spelling
discriminator into the Go reader and reaching the feature from three surfaces. It is filed as
**TASK-260906-19gjyw**, and stage (c) will not be reported complete against curator-spec main until
that lands. Your job here is to confirm the *bound* is truthfully worded and the ledger rows say what
their tests assert — not to reopen the decision.

## Review dimensions

1. **C5-M1 — the reinstall retry.** `install <same path> --as <same name> --use --takeover` used to
   drop both flags and report success, which is the exact retry §9.5's stop invites. Verify all four
   combinations through `run()` — `--use` alone, `--takeover` alone, both, neither — on a `path` root
   that is current and one that is not, and that the corrected doc comment now describes the real
   path. This is the third cycle to find a comment claiming a production path the code does not have;
   check the rest of the stage's comments for the same shape while you are in there.

2. **C5-m1 — the addressing mode again.** The regression test should now pin the bare
   `profile install <same path>` form as well as the `--as` one. Re-apply cycle 5's mutant
   (`options.As != ""`) and confirm it dies. Then ask what other addressing mode of this command is
   unpinned — every blocking finding in this stage lived in one.

3. **C5-m4 — the two blocking-audit gates.** Both `reinstallPathLocked`'s copied gate and the
   `updateLocked` twin should now have narrowing mutants and named tests. Cycle 5 reported
   exploitability as *unknown* because a second independent gate sits in the same loop; the producer
   was asked to say whether a blocking-but-not-strict finding is reachable at all. Judge that answer —
   "not reachable, and here is why" is acceptable; an unexamined gate is not.

4. **C5-m2 — the skip enumeration.** Every skip this stage introduces should now be classified, found
   by enumeration rather than by waiting for a lane to redden. Verify the enumeration was actually
   performed and is complete: walk the stage's skips yourself against `skip-classes.tsv`.

5. **C5-m3 — the reissued bound.** The bound should now name TASK-260906-19gjyw and describe an
   implementation gap rather than a spec contradiction, and ledger rows 303–304 must describe what
   their tests assert. Confirm nothing untruthful lands.

6. **The candidate lane against the authority.** Read run 34058365116. Say whether the environments
   families ran and passed on all three runners against `550579d`, which is what this stage's AC
   requires.

7. **The landing question.** Sixth cycle, first green head, one filed consumption gap. Say plainly
   whether stage (c) is safe to land. If it is, say so without hedging. If it is not, name the one
   thing that must change — and weigh whether it genuinely must change *before* landing rather than in
   the follow-up that already exists.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root. **Run the two `test-gate.sh` lanes sequentially.**

## Verdict contract

Attach `TASK-260906-1uf713_review-findings-stage-c-6.md` with a `repeat-of:` line naming any class
that recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review` with `accept_cr` on the recorded revision, stating that every hosted lane was green on the
exact head you accepted and what both candidate dispatches reported. Do not mark the task done. Then
`task-board handoff TASK-260906-1uf713 --role reviewer`.
