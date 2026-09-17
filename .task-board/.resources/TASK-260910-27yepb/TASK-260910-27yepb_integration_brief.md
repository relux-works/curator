# Integration run — STORY-260910-3rvvxh (R1/P1 service half), final leaf TASK-260910-27yepb revision 1 accepted

You are the tracked developer run bound to the accepted revision 1 of
`TASK-260910-27yepb`, the FINAL leaf of story `STORY-260910-3rvvxh`
(`TASK-260910-14dnb7` is checkpointed on the Story branch at `feecd4b`). The
reviewer accepted it (`TASK-260910-27yepb_review-verdict-rev1.md`).

Do exactly this, nothing else:
1. `task-board worktree obligations` — confirm the row
   `TASK-260910-27yepb  1  accepted  checkpoint` (a final leaf is integrated,
   not checkpointed).
2. From the control root
   `/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry`
   (the repository's main working tree on `main`), run
   `task-board worktree integrate STORY-260910-3rvvxh --cr TASK-260910-27yepb --revision 1`
   — it lands the Story candidate as ONE signed squash commit on the local
   `main` (never pushed by you), performs the Story's done transition and
   writes the board-only commit in the board repository. Quote the full
   output including `post_landing_steps`. If it refuses, quote the typed
   error verbatim in your outcome and stop: no repair, no reset, no manual
   commit, no board status change, no `--rollback` unless the output tells
   you the transaction is at phase `prepared` with trunk unmoved AND the
   refusal names a cause you cannot fix by re-running — then re-run once
   before considering rollback, and report.
3. `git -C <control root> log --oneline -3`, `git -C <control root> status --short`,
   `git -C <control root> verify-commit HEAD`; `task-board worktree obligations`
   again.
4. Attach `TASK-260910-27yepb_integration.md` (task outcome) with the
   transcripts, the squash commit id and the board commit id, then exit.
   Do NOT push anything, do NOT run `reconcile-trunk` (the orchestrator
   publishes the branch, lands the PR and reconciles), no code changes.
