# Review brief: stage (c) cycle 5 (rework 4, landing decision)

## Subject

- Branch `feat/agent-environments-stage-c` at `71e6baec` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, 29 signed commits past
  `origin/main`. PR https://github.com/relux-works/curator/pull/61. Rework 4 is
  `git diff 4a8a1d42..HEAD`.
- **All eleven hosted lanes are green on this exact head, Windows included** — the orchestrator waited
  for them before spawning you, which it had not done for the two previous cycles. So a red lane is
  not this cycle's job; finding what green lanes cannot see is.
- A candidate-suite dispatch against curator-spec `87a0d006` was started at the same time
  (run 34052590291). Read it before you finish: it is the only lane that serves the environments
  families, since `SPEC_PIN` still predates them.
- Read `TASK-260906-1uf713_review-findings-stage-c-4.md`, unpack its probes archive,
  then `producer-brief-stage-c-rework-4.md` and `TASK-260906-1uf713_rework-report-4.md`.
- Authority: curator-spec `87a0d006` — §1, §6, §8.4, §9.1, §9.2, §9.5, §9.6, §12.1, §12.2;
  `profiles/manager.md`; `cli/curator.md`. Note the `path`-overlay reconciliation landed there since
  cycle 4, so the M1 bound stage (c) carries may now be stale — check it.
- Change Request revision to accept: the one recorded for TASK-260906-1uf713.

## Standing

Four cycles have found sixteen findings in this stage. Every blocking one lived in an addressing mode
nobody tested: stage (a) F1, cycle-1 B3, cycle-2 C2-B1, cycle-3 C3-B2, cycle-4 C4-B1. Two of them were
regressions introduced by the previous cycle's fix. Weight your effort accordingly.

## Review dimensions

1. **C4-B1 — both halves of §1 at once.** The paired assertions should now hold together: `update`
   must not move the pin after a source edit, and `install <same path> --as <same name>` must, with
   the new bytes reaching the materialized surface. Drive both through `run()`. Then attack the seam
   between them: a same-path install under a *different* name, a different path under the same name,
   an install whose source vanished after the first install, and `install --use` on a `path` root.
   Cycle 4's finding was a regression created by cycle 3's fix in exactly this neighbourhood.

2. **C4-M1 — the three refusals, and the §8.4 class.** All three should now be driven. Re-apply
   cycle 4's three mutants and confirm each dies, especially the one that turned a failed snapshot
   read into `unchanged`. The producer's answer to the structural question is honest and negative:
   nothing prevents a fourth site, because the three read three different manager-owned shapes through
   three different helpers, and the class is closed by vigilance plus per-site mutants. Judge whether
   that answer is right — if a shared read helper or an enumerating test is in fact feasible, say so;
   if not, confirm the per-site pins are complete by finding every read of manager-owned state in the
   stage yourself and checking each has one.

3. **C4-B2 — the skip registry.** The mode-000 **file** reason is now classified. Verify the fix
   widened the registry rather than making the skip text lie about the artefact, that the three
   sibling "directory" skips still classify, and that no other skip reason introduced by this stage is
   unregistered. Reproduce the classifier locally against the observed reason text, since no Unix lane
   exercises the skip.

4. **C4-m1, C4-m2.** The builtin `default` is now refused under a lock at both levels with the mutant
   dying at each. For m2 — `profile update` on a machine already violating the lock moving the lock
   before the resync fails — check the producer either fixed it or documented the answer; cycle 4 said
   an undocumented answer is not acceptable.

5. **The M1 bound may be stale.** Stage (c) carries an explicit bound saying the `path` overlay is
   unreachable pending a spec fix. That fix landed on curator-spec main as `87a0d006`. Decide whether
   the bound still holds against the landed spec, or whether stage (c) can now reach the feature and
   the bound must be retired. Say which; do not require the producer to implement it in this cycle if
   it is real work, but do not let a stale bound land unexamined.

6. **The candidate lane.** `SPEC_PIN` predates the environments artifacts, so the default lanes defer
   the packages that read them. Run 34052590291 serves them. Read its three Candidate suite jobs and
   say whether the environments families actually ran and passed on all three runners.

7. **The landing question.** This is the fifth cycle. Say plainly whether stage (c) is safe to land.
   If it is, say so without hedging. If it is not, name the one thing that must change.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root. **Run the two `test-gate.sh` lanes sequentially**, never concurrently and never
alongside a `-race` suite.

## Verdict contract

Attach `TASK-260906-1uf713_review-findings-stage-c-5.md` with a `repeat-of:` line naming any class
that recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review` with `accept_cr` on the recorded revision, stating that every hosted lane was green on the
exact head you accepted and what the candidate lane reported. Do not mark the task done. Then
`task-board handoff TASK-260906-1uf713 --role reviewer`.
