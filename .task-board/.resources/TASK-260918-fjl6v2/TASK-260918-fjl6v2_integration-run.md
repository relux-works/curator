# Integration run — STORY-260918-2yvd86 (rc.12 conformance-pin promotion union landing), final leaf TASK-260918-fjl6v2 revision 1 accepted

You are the tracked developer run bound to the accepted revision 1 of
`TASK-260918-fjl6v2`, the FINAL leaf of story `STORY-260918-2yvd86`
(the story's only leaf; the union it lands was accepted at TASK-260917-16l2md
revision 5 and re-applied here on the current trunk). The reviewer accepted revision 1
(`TASK-260918-fjl6v2_review-verdict-rev1.md`). The curator repository is the
board owner, so `worktree integrate` is the delivery path here.

Do exactly this, nothing else:
1. `task-board worktree obligations` — confirm the row
   `TASK-260918-fjl6v2  1  accepted  checkpoint` (a final leaf is integrated,
   not checkpointed).
2. From the control root `/Users/administrator/Developer/ReluxWorks/curator/curator`
   (main working tree on `main`), run
   `task-board worktree integrate STORY-260918-2yvd86 --cr TASK-260918-fjl6v2 --revision 1`
   — it lands the Story candidate as ONE signed squash commit on the local
   `main` (never pushed by you), performs the Story's done transition and
   commits the board delta as a second board-only commit. Quote the full
   output including `post_landing_steps`. If it refuses with `integration_base_moved` (the trunk moved onto the
   candidate's paths since the review), quote it and stop — the orchestrator
   routes the refresh. For any other refusal quote the typed
   error verbatim in your outcome and stop: no repair, no reset, no manual
   commit, no board status change, no `--rollback` unless the output says
   the transaction is at phase `prepared` with trunk unmoved AND names a
   cause a re-run cannot fix — re-run once first, then report.
3. `git -C <control root> log --oneline -4`, `git -C <control root> status --short`,
   `git -C <control root> verify-commit HEAD~1` and `HEAD` (both signed);
   `task-board worktree obligations` again; `task-board q 'get(STORY-260918-2yvd86) { status children }'`.
4. Attach `TASK-260918-fjl6v2_integration.md` (task outcome) with the
   transcripts, the squash commit id and the board commit id, then exit.
   Do NOT push anything, do NOT run `reconcile-trunk` (the orchestrator
   publishes the branch, lands the PR and reconciles), no code changes.
