# Republish brief — TASK-260916-dv7xv5 (rev3 is complete; hand off only)

Supersedes `dv7xv5-rework-rev3-brief.md` for this run. The rev4 artifact is
already attached (`verify-e-findings-rev4.md` = rev3 plus the two transcript
corrections of review verdict rev3, applied by the orchestrator; sibling READMEs
already corrected). The checklist is fully checked (13/13); the
"Tests green" baseline item is checked by the orchestrator on the recorded
not-applicable basis — do not uncheck or discuss it.

Your whole job:
1. Confirm with `task-board q 'get(TASK-260916-dv7xv5) { overview }'` that
   `verify-e-findings-rev4.md` is listed as an outcome and the checklist has no
   unchecked item.
   Do NOT edit any resource, README, code or test; do not re-run research.
2. Run `task-board handoff TASK-260916-dv7xv5 --role researcher` so the runtime
   publishes the Change Request revision (empty repository delta is expected and
   admissible for this research task) and the task reaches `to-review`.
3. If the handoff refuses, record the exact refusal text in the task notes
   (`task-board m 'set_notes(TASK-260916-dv7xv5, text="…")'`) and stop; do not
   work around it.
