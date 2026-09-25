# TASK-260922-18ex37 — rework 6: remove the merge artefact (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 5 CHANGES_REQUESTED with ONE blocking finding (verdict rev5 F1): the candidate adds `.github/workflows/ci.yml.merged.tmp`.
Everything else is verified — change nothing else.
1. `task-board m 'set_status(TASK-260922-18ex37, status=development)'`.
2. Delete `.github/workflows/ci.yml.merged.tmp` from the Story worktree; `git status --porcelain` must show no other untracked artefact
   (`*.tmp`, `*.orig`, `*.rej`, `*.merged*`) — remove any such file too and list it.
3. VERIFY `git diff --name-only HEAD -- . ':!.task-board'` equals revision 5's paths minus that file (61 paths).
4. Append "Revision 6 — artefact removed" to results, `resource update`, `task-board handoff TASK-260922-18ex37 --role developer`; stay in
   the turn while the gate runs. A write-boundary `policy warn` block is a warning.
