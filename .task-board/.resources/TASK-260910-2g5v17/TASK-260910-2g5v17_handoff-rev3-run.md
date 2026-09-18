# Handoff-only run — TASK-260910-2g5v17 (P4): publish revision 3 (the exact landed tree)

**This is NOT an integration run** (do not run `integrate`, `complete`,
`close-landed`, `prepare-landed-review` or `start-landed-rework` — all done).
Your predecessor RUN-260918-a36f78 exported the landed review, opened the
landed rework and applied the export's `update_patch`, so the Story worktree
`<control-root>/.temp/STORY-260910-stz5f0/worktree` now holds the exact
landed tree `9b7d33115bd3ffe44c34c5a340de72aaab06d0cb` (see
`TASK-260910-2g5v17_results-rev3.md`). Its `task-board handoff` moved the
task to `to-review` but, being bound as an integration run, published NO
Change Request revision (the current revision is still 2, accepted). A normal
producer handoff must publish revision 3.

Do exactly this:
1. Verify the snapshot tree with a temporary index exactly like
   `scripts/remote-gate.sh` (`GIT_INDEX_FILE=$tmp git read-tree HEAD; git add -A .; git write-tree`)
   — it MUST print `9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`; the worktree
   must not contain `TASK-260910-2g5v17_results.md`. If either fails, quote
   it and stop.
2. Write `/tmp/2g5v17/handoff-note-rev3.md` (outside the worktree) — two lines:
   this run's id and "tree verified 9b7d331…; revision 3 = landed tree" — and
   attach it as `TASK-260910-2g5v17_handoff-note-rev3.md` (type=outcome).
3. `task-board handoff TASK-260910-2g5v17 --role developer`. Quote the output.
   The runtime publishes revision 3 and runs the hosted gate. If the handoff
   refuses, quote the typed error verbatim and stop.
