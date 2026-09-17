# Integration run — STORY-260910-2awkzu (S6 shell-hook trust gate), final leaf TASK-260910-3ungjy revision 5 accepted

You are the tracked developer run bound to the accepted revision 5 of
`TASK-260910-3ungjy`, the FINAL leaf of story `STORY-260910-2awkzu`
(`TASK-260910-1952mz` is checkpointed on the Story branch; `TASK-260910-1wjst3`
was the spec leaf, integrated). The reviewer accepted revision 5
(`TASK-260910-3ungjy_review-verdict-rev5.md`). The curator repository is the
board owner, so `worktree integrate` is the delivery path here.

Do exactly this, nothing else:
1. `task-board worktree obligations` — confirm the row
   `TASK-260910-3ungjy  5  accepted  checkpoint` (a final leaf is integrated,
   not checkpointed).
2. From the control root `/Users/administrator/Developer/ReluxWorks/curator/curator`
   (main working tree on `main`), run
   `task-board worktree integrate STORY-260910-2awkzu --cr TASK-260910-3ungjy --revision 5`
   — it lands the Story candidate as ONE signed squash commit on the local
   `main` (never pushed by you), performs the Story's done transition and
   commits the board delta as a second board-only commit. Quote the full
   output including `post_landing_steps`. If it refuses, quote the typed
   error verbatim in your outcome and stop: no repair, no reset, no manual
   commit, no board status change, no `--rollback` unless the output says
   the transaction is at phase `prepared` with trunk unmoved AND names a
   cause a re-run cannot fix — re-run once first, then report.
3. `git -C <control root> log --oneline -4`, `git -C <control root> status --short`,
   `git -C <control root> verify-commit HEAD~1` and `HEAD` (both signed);
   `task-board worktree obligations` again; `task-board q 'get(STORY-260910-2awkzu) { status children }'`.
4. Attach `TASK-260910-3ungjy_integration.md` (task outcome) with the
   transcripts, the squash commit id and the board commit id, then exit.
   Do NOT push anything, do NOT run `reconcile-trunk` (the orchestrator
   publishes the branch, lands the PR and reconciles), no code changes.
