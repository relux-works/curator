# Review note — TASK-260923-xq4pjj shim adoption (orchestrator, binding; THE ONLY CURRENT INSTRUCTION; read-only review)

Host-only metadata leaf. Review the evidence in TASK-260923-xq4pjj_global-adopt-results.md against `xq4pjj-adopt-1.md` and the task AC,
READ-ONLY on the host (never write to ~/.curator, ~/.local/bin or any daemon state):
1. ~/.local/bin/task-board bytes equal the canonical shim (cmp) and its sha256 equals the recorded pre-adopt value.
2. The Curator ownership marker lists task-board (it was written by `curator global adopt`, not by hand — compare with the pre-adopt
   backup under ~/.curator/backups/20260928T043627Z-shim-global-adopt/ and the command's own backup under ~/.curator/backups/global-bins/).
3. A fresh login shell resolves task-board and `task-board --version` exits 0; the backups exist with the recorded sha256.
4. The results honestly report the tb-sessiond deleted-binary observation and task-board-tui state.
Check the DoD items that the evidence satisfies; the "via curator global install" wording of item 2 is superseded by the operator decision
(`global adopt`). accept (review verdict + checklist) or changes requested. No LOGBOOK.md.
