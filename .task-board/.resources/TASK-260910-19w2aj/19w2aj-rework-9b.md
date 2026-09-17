# Rework 9 (re-anchored) — TASK-260910-19w2aj

Trunk moved (0945447816cb) while the rev9 candidate sat uncommitted; the orchestrator captured it as precondition resource TASK-260910-19w2aj_rev9-candidate.patch (git diff --binary HEAD) and cleaned the workspace; this spawn replayed the lock checkpoint onto the current trunk.
1. `git apply --binary` the patch in the Story worktree; `git status --short` must list the rev9 paths.
2. Apply rework 9 exactly as written in 19w2aj-rework-9.md (Windows-portable script command path in the refresh-failure fixture; product unchanged).
3. Narrow tests, tool calls under 2 minutes, evidence, checklist, handoff rev10 (story_final).
