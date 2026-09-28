# TASK-260923-14df7m — CHANGELOG correction before v0.5.20 (THE ONLY CURRENT INSTRUCTION)

F-M1c (TASK-260923-3e4i3n) landed on main as 4dfacfb. The orchestrator found two CHANGELOG problems (no code change allowed):
1. The heading was changed to `## Unreleased — v0.5.20`. This repository keeps `## Unreleased` at release tags
   (v0.5.18 and v0.5.19 were both tagged with it). Restore `## Unreleased`.
2. The EXISTING F-M1b entry (released in v0.5.18) was rewritten ("expected release v0.5.18" → "v0.5.20",
   grammar sentence changed). Released history must not be rewritten: restore that entry byte-for-byte to its
   4e229cc text, and add a NEW bullet for this leaf (Decision 0018 choice 4 known-conflict refusal under yolo,
   `ErrNativePolicyConflict` / `*NativePolicyConflictError{Selector, Placement}`, `permission-grammar-v2` for
   Claude and Codex while Pi stays v1, release v0.5.20 cut by the orchestrator).
Verify: `git diff origin/main -- CHANGELOG.md` shows only the new bullet; every non-CHANGELOG path is byte-identical to
revision 2 (tree da3001cc). `go test ./...` bounded. Attach `TASK-260923-14df7m_results.md`, check the DoD item citing it, then
`task-board handoff TASK-260923-14df7m --role developer`. A `run_wrote_outside_worktree … policy warn` block is a
warning — verify status `to-review`.
