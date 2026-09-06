# Review brief: stage (c) cycle 3 (rework 2)

## Subject

- Branch `feat/agent-environments-stage-c` at `4df4d507` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, 21 signed commits past `origin/main`.
  PR https://github.com/relux-works/curator/pull/61 — pushed, so hosted lanes run on the exact head
  you review. Rework 2 is `git diff 0bcea201..HEAD`.
- Read `TASK-260906-1uf713_review-findings-stage-c-2.md` (your predecessor's cycle-2 findings),
  `producer-brief-stage-c-rework-2.md`, and `TASK-260906-1uf713_rework-report-2.md`. The cycle-1
  findings and probes archive are attached too if you need the history.
- Authority: curator-spec `550579d` — §1, §6, §8.4, §9.1, §9.5, §9.6, §12.1, §12.2;
  `profiles/manager.md`; `cli/curator.md`.
- Change Request revision to accept: the one recorded for TASK-260906-1uf713.

## Standing

Cycle 2 confirmed all four cycle-1 blocking findings fixed and driven, the B1 seam enumeration true,
the B2 write sweep clean, and every hosted lane green on `0bcea201` including Windows. Its own
mutants R1–R5 were killed. Do not re-derive that.

Two findings remained, and both were invisible to every lane — which is the standard this cycle is
held to as well: **a green suite is not evidence.**

## Review dimensions

1. **C2-B1 — the corrupt marker, as a class.** A failed marker read used to fall through the
   managed-home skip, so `import` re-detected curator's own generated root context as native. The
   producer was told to fix the *class*, not the call site: every place in this stage that treats a
   failed read of manager-owned state as "not present". Verify it did. Drive `import` against a
   corrupt marker, an unreadable marker, and an absent marker, and confirm three distinct outcomes
   with the §8.4 distinction intact. Then hunt for the same shape elsewhere in the stage yourself —
   this defect has now appeared twice (`readSkillsLedger`, `readRootSurface`) and a third site is the
   likeliest place for it to hide. Apply the producer's own mutant and confirm the kill.

2. **C2-M1 — is the gate actually derived now?** The lockable-subset test was supposed to stop being a
   hand-written list of knob names and start enumerating §12.2 from the same source the code does.
   Verify that is what happened rather than a fifth and sixth case being added. Re-apply all four
   knob mutants — cycle 1's `forms`, plus `secret_material_waivers`, `xdg_seed_allowlist`,
   `in_place_mode` — and confirm each dies. Then invent a fifth knob mutant of your own choosing and
   confirm it dies too; if it survives, the gate is still a list. The `secret_material_waivers` path
   is the one your predecessor drove end to end into §9.1 secret material, so weight it accordingly.

3. **C2-m2 — `Config.CheckMachineUse` was dropped.** The producer removed it rather than routing the
   seam through it. Verify nothing lost coverage: the §12.2 refusal must still be driven from
   `run()`, and the comment that used to lie about reachability must be gone rather than merely
   moved. Confirm no other function in this stage carries a comment claiming a production path it
   does not have — that is how cycle-1 M1 happened.

4. **C2-m1 — the stated bound.** The §9.5 dotfile list is POSIX-only, which makes the heuristic inert
   on Windows. The orchestrator ruled this a spec question, filed as TASK-260906-vlrjo1, and required
   an explicit bound naming that task rather than invented Windows paths or a silenced case. Verify
   the bound is stated in both the code and the report, that it is truthful, and that the case still
   runs and asserts something real on all three runners.

5. **C2-m3, C2-m4.** Ledger row 304 must describe what its test actually asserts. The `applyPlan`
   seed and marker writes should now carry the remove-first hardening; confirm the change is inert
   with respect to behaviour outside the manager tree, since your predecessor was explicit this is
   not the B2 data-loss class.

6. **The whole-stage question.** This is the last cycle before landing unless you find something.
   Beyond the six findings, spend your remaining effort where the stage is weakest rather than where
   it is best documented: composition's four weight rules in order, the two precedence primitives
   driving real materialized bytes, and the import's loss list. Say plainly whether stage (c) is safe
   to land.

7. **The hosted lanes.** Read `gh pr checks 61`. For any lane not green, extract failing cases from
   that run's uploaded evidence artifacts, not from `--log-failed`. A red lane on this head is
   blocking. **Run the two `test-gate.sh` lanes sequentially**, never concurrently and never alongside
   a `-race` suite.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root.

## Verdict contract

Attach `TASK-260906-1uf713_review-findings-stage-c-3.md` with a `repeat-of:` line naming any class
that recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review` with `accept_cr` on the recorded revision, and state plainly whether every hosted lane was
green on the exact head you accepted. Do not mark the task done. Then
`task-board handoff TASK-260906-1uf713 --role reviewer`.
