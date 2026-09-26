# TASK-260922-cww1ov base refresh → revision 3 (orchestrator brief, binding)

Revision 2 was ACCEPTED (astra verdict `TASK-260922-cww1ov_review-verdict-rev2.md`) but the
integration refused: trunk advanced to `48da2690` (the rc.12 pin promotion, PR #79) with a
`CHANGELOG.md` change that revision 2 also makes, so the record went stale and the acceptance has
been released (`worktree invalidate-acceptance`). The task is back at `to-dev` with revision 2
marked stale. NOTHING about the content is in question — this is a base refresh only.

## What to do
1. Refresh the candidate onto the current trunk:

       task-board worktree refresh-candidate TASK-260922-cww1ov

   `CHANGELOG.md` carries `merge=union` in this repository's `.git/info/attributes`, so the two
   Added sections should combine without a conflict. If a regular-file conflict is reported anyway,
   follow the command's own instructions (`--replay-resolutions` bound to REBASE_HEAD, the unmerged
   path and the SHA-256 of the replacement content) — never hand-commit the replay worktree.
2. Verify the refreshed candidate is byte-identical to revision 2 EXCEPT for the combined
   `CHANGELOG.md` (the F-C1/F-C2 checkpoints and your 22 paths must be unchanged): report the exact
   diff of the refresh in results.md — `git status --short` plus a per-file `git diff` summary.
3. Re-run the narrow rows you own on the refreshed tree with real exit codes (the 0017 hazards, the
   migration/recovery rows, the no-copy scan over the manager home) — the refreshed base carries the
   rc.12 conformance pin (`SPEC_PIN` is now `dced9b8`, v1.0.0-rc.12), so a vector-consuming row that
   previously SKIPPED may now execute. If any row changes behaviour because of the new root, say so
   explicitly with its name; if one fails, fix it minimally and name the fix.
4. Append a "Revision 3" section to `TASK-260922-cww1ov_results.md` (refresh only: what combined,
   what stayed identical, which rows executed that previously skipped, exit codes), then
   `task-board handoff TASK-260922-cww1ov --role developer`. Publish only on a green gate.

No product change beyond a minimal fix demanded by item 3. No test weakened. This is still the FINAL
leaf of STORY-260922-1cenbr: on re-acceptance the orchestrator integrates the Story.
