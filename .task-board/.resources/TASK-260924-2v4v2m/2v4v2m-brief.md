# TASK-260924-2v4v2m — (THE ONLY CURRENT INSTRUCTION)

Control root: curator; your Story worktree only. Read `campaign-producer-rules.md`. Landing gate = hosted CI at handoff.
Source: the gap matrix `TASK-260924-3re9jo_results.md` (attached to TASK-260924-3re9jo; read its "Implementation leaf list" row for
this task and the "Prior review evidence" section) and curator-spec main (`protocol/skillfile-sources.md`, `protocol/repository-transport.md`,
`schemas/draft-sources-v1/`). Do exactly this task's description and AC; production-entry rows; narrowing mutants in a disposable
copy (table in results); no weakening of any existing assertion. Note: TASK-260924-1aa9wb is removing the CURATOR_DRAFT_SOURCES_V1
switch in parallel — enable the lane in your tests the way existing draft tests do today; do not touch the switch yourself.
Bounded runs (`go test ./internal/marker`, ≤10 min/call), CHANGELOG. Attach `TASK-260924-2v4v2m_results.md`, check the DoD item, `task-board handoff TASK-260924-2v4v2m --role
developer`. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
