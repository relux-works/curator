# Review brief: stage (a) core — cycle 5 (rework of F13, F14, F15)

Subject: curator worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`, branch
`feat/agent-environments-stage-a`, head `b6f00e1a` (one signed commit on the cycle-4 head `dea3f5ac`,
5 files, +428/−50). Cycle-4 findings: `TASK-260905-30zs8t_review-findings-stage-a-4.md`; author
decisions: `producer-brief-stage-a-rework-4.md`; rework report:
`TASK-260905-30zs8t_rework-report-4.md`.

Verify by reproducing, with your own scripts:
1. **F14** — classification is syntactic and never stats the operand: an operand whose spelling is a
   git identity but which also names an existing directory in the working directory resolves as `git`
   and is refused by a locked allowlist (the cycle-4 `a9.sh` shape); `/tmp/<absent>` and `./<absent>`
   with `--range` give `profile_install_ref_conflict`; the Windows spellings are classified as `path`.
   Grep for any remaining `Stat`/`Lstat` on the classification path.
2. **F13** — `profile install --use` performs the §9.2 switch: every entry attempted, per-adapter
   results, the current recorded **only** when the whole scope materialized, `profile_use_partial`
   otherwise. Drive it on a machine already current on another profile with four real adapter homes;
   force one adapter to fail and confirm the previous current survives, the marker and materialized
   bytes agree, and the exit code and message say what happened. Confirm the non-activating branch is
   unchanged.
3. **F15** — the policy table covers `list`, `use`, `sync` and `update`; re-run the M-D mutant (drop the
   threading from one of the three new paths) and confirm a named test dies.
4. **Regression** — the full "verified, held" set from cycles 1–4: directory-addressed detector cases,
   fresh-home default profile, interrupted switch under the lock, F8 three spellings, F9 gates with
   `audit.enabled` false, F10 system-locked migration, F12 `file://` rejection. This rework touched the
   install and activation paths, so re-do them rather than trusting the report.
5. **Gates** — the full set including `bash .github/ci/gate-selftest.sh`, the platform-case gate for the
   three GOOS values, and the vector families through `CURATOR_CONFORMANCE_ROOT`; confirm two rows of
   the producer's mutant table yourself.
6. Commits signed by the repository's human identity; eleven commits on curator main `a2406dfe`.

This is the fifth cycle. If the fixes hold and no new blocking or major finding appears, ACCEPT so the
stage can land; keep a strict bar for anything that would ship a wrong artifact or a bypassed gate, and
record everything smaller as a follow-up rather than another cycle. Read-only (scratch under the
worktree's `.temp/`). Never write into the control root. Findings resource
`TASK-260905-30zs8t_review-findings-stage-a-5.md`. Blocking/major → `development`; else explicit ACCEPT
at `to-review` with `accept_cr`. Do not mark done.
`task-board handoff TASK-260905-30zs8t --role reviewer`.
