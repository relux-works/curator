# Rework 1 (re-anchored) — TASK-260916-2bwfli

Trunk moved while the rev2 candidate sat uncommitted; the orchestrator captured it as precondition resource TASK-260916-2bwfli_rev2-candidate.patch (git diff --binary HEAD, the 7 leaf files) and cleaned the workspace; this spawn replayed the two checkpoints onto the current trunk.
1. `git apply --binary` the patch in the Story worktree; `git status --short` must list exactly the rev2 leaf files.
2. Apply rework 1 exactly as written in 2bwfli-rework-1.md (named provider admission through the accepted AuthProvider contract; unknown/missing → ErrProviderUnavailable; CLI regressions; fix TestDraftTransportAuth).
3. Narrow tests, tool calls under 2 minutes, evidence, mutant, checklist, handoff rev3 (story_final).
