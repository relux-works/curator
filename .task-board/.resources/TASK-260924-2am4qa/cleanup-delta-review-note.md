# Review note — cleanup revision (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

The previous revision of this task was ACCEPTED on content except for one stray repository file (a task document at the repo root).
The new revision must differ from it ONLY by removing that file. Verify by diffing the two Change Request patches (both are resources on
this task): per-file `git patch-id --stable` identity for every other path, the stray path absent, no other path added or removed; the
runtime validation log for the new revision green. accept_cr on the new revision, or changes requested naming the unexpected path.
No LOGBOOK.md.
