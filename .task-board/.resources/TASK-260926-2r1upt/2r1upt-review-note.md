# Review note — TASK-260926-2r1upt per-guard replay rows (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review against `2r1upt-brief.md`: one production-entry row per guard of internal/install/draftsources.go declaredDependencyReplaySources
(directory mismatch → source_snapshot_changed, identity mismatch → source_snapshot_changed, every other branch); for each guard reproduce
yourself that its mutant survives before and is killed by the new row (bounded, focused go test runs). Only draftsources_test.go changes;
gate green. accept_cr or changes requested with file:line. No LOGBOOK.md.
