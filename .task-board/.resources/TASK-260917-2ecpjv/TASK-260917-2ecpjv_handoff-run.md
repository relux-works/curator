# Handoff run — TASK-260917-2ecpjv (rc.12 qualification, outcome already attached)

Your predecessor RUN-260918-ba4ea7 completed the read-only qualification
(`TASK-260917-2ecpjv_qualification.md`, verdict **qualified**, checklist
11/11) but its handoff failed at Change Request construction with
`stale-anchor: change_request_base_authority_mismatch` because the task was
the story's last open leaf while the curator trunk had moved. The
orchestrator added the sibling leaf `TASK-260918-11f9l1` (the trunk replay),
so this task is no longer story-final and its Change Request can be
constructed.

Do exactly this, nothing else:
1. `task-board q 'get(TASK-260917-2ecpjv) { status }'` and confirm the
   outcome resource `TASK-260917-2ecpjv_qualification.md` exists (read it;
   do not redo the qualification, do not touch any repository).
2. Append a two-line "Handoff note (RUN <your id>)" to a NEW task-scoped
   outcome `TASK-260917-2ecpjv_handoff-note.md` stating that the outcome is
   unchanged and why this run exists (so the handoff sees a new task-scoped
   outcome), attach it with `task-board m 'add_resource(TASK-260917-2ecpjv, name="TASK-260917-2ecpjv_handoff-note.md", path=<file>, type=outcome, description="…")'`.
3. `task-board handoff TASK-260917-2ecpjv --role researcher`. Quote the
   output. If it refuses again with stale-anchor, quote the error verbatim
   in the note and stop — do not retry, do not mutate anything else.
