# TASK-260906-2b3nar base refresh → revision 2 (orchestrator brief, binding)

Revision 1 was ACCEPTED (`TASK-260906-2b3nar_review-verdict-rev1.md`, no blocking findings) but the
integration refused: trunk advanced to `48da2690` (the rc.12 pin promotion, PR #79) with a
`CHANGELOG.md` change that revision 1 also makes, so the record went stale and the acceptance has
been released. The content is NOT in question — this is a base refresh only.

1. Refresh the candidate onto the current trunk:

       task-board worktree refresh-candidate TASK-260906-2b3nar

   `CHANGELOG.md` carries `merge=union` in `.git/info/attributes`, so the two Fixed/Added sections
   should combine without a conflict. On a real conflict follow the command's `--replay-resolutions`
   instructions; never hand-commit the replay worktree.
2. Prove the refreshed candidate is byte-identical to revision 1 except for the combined
   `CHANGELOG.md` (report `git status --short` plus a per-file diff summary in results.md).
3. Re-run your own rows on the refreshed tree with real exit codes: `go test ./internal/gitops
   ./internal/snapshot -count=1` plus the four new fold rows and, if the ledger demands it,
   `ledger-consistency.sh`. The refreshed base carries the rc.12 conformance pin (`SPEC_PIN` =
   `dced9b8`), so a vector-consuming row that previously SKIPPED may now execute — name any row
   whose behaviour changes. If one fails, fix it minimally and say so.
4. Append a "Revision 2" section to `TASK-260906-2b3nar_results.md` (refresh only), then
   `task-board handoff TASK-260906-2b3nar --role developer`. Publish only on a green gate.

No product change beyond a minimal fix demanded by item 3; no test weakened. This is the last live
leaf of STORY-260905-2qvzwk: on re-acceptance the orchestrator integrates the Story.
