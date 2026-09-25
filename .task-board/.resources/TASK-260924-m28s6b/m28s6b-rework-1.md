# TASK-260924-m28s6b — rework 1 (THE ONLY CURRENT INSTRUCTION)

Revision 1 (lock replay) failed ONLY on the race lanes (run 35978346497):
    cmd/curator TestDraftDocsPinExamples: draft_diagnostics_test.go:1152: docs/troubleshooting.md misses the source_snapshot_unavailable
    remedy keyword "run the explicit attempt first"
With replay, `source_snapshot_unavailable` now means "the declared source cannot be reached" (a missing snapshot is re-materialized). Make the
three places agree on the NEW meaning: the remedy row for source_snapshot_unavailable / source_snapshot_changed in
`cmd/curator/draft_diagnostics.go` (remediation table), `docs/troubleshooting.md`, and the pinned keywords in `TestDraftDocsPinExamples` — the pin
must keep pinning a real remedy keyword (update it to the new remedy text, don't delete it). Nothing else changes (no CHANGELOG edit). Bounded:
`go test ./cmd/curator -run 'TestDraftDocsPinExamples|Diagnostic'` + your replay rows. Append "Revision 2", `resource update`, `task-board handoff
TASK-260924-m28s6b --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
