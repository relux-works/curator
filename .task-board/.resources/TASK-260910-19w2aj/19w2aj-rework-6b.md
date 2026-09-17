# Rework 6 (re-anchored) — TASK-260910-19w2aj

Trunk moved (aa46ecd) while the rev6 candidate sat uncommitted; the orchestrator captured it as precondition resource TASK-260910-19w2aj_rev6-candidate.patch (git diff --binary HEAD, 9 paths) and cleaned the workspace; this spawn replayed the lock checkpoint onto the current trunk (which now also contains the landed transport revision 2 story — re-run your narrow tests after applying).
1. `git apply --binary` the patch in the Story worktree; `git status --short` must list the rev6 paths.
2. Apply rework 6 exactly as written in 19w2aj-rework-6.md (expand selections/collections over the resolved commit's frozen tree, not the checkout; CLI regressions; mutant).
3. Narrow tests, tool calls under 2 minutes, evidence, checklist, handoff rev7 (story_final).
