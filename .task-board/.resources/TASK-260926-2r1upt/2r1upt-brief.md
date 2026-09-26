# TASK-260926-2r1upt — per-guard rows for declaredDependencyReplaySources (THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md` and the task description. In internal/install/draftsources.go declaredDependencyReplaySources (11burj):
add one production-entry row per guard (directory mismatch → source_snapshot_changed, identity mismatch → source_snapshot_changed, and every
other branch), each with a per-guard mutant that survives before and is killed after (real exit codes). No production change unless a row
exposes a defect (then fix and say so). No CHANGELOG/LOGBOOK edit. Focused bounded runs. Attach results, check DoD,
`task-board handoff TASK-260926-2r1upt --role developer`. A write-boundary `policy warn` block is a warning.
