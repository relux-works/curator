# Rework 2 (re-anchored) — TASK-260910-5nrmtt

Trunk moved again (board-state commit b7633b7) while the rev2 candidate sat uncommitted, so the orchestrator captured it as precondition resource TASK-260910-5nrmtt_rev2-candidate.patch (git diff --binary HEAD, 12 paths, identical to the reviewed rev2 tree c65570b0 minus the checkpoint) and cleaned the workspace; this spawn replayed the story checkpoint onto the current trunk.

1. `git apply --binary` the patch in the Story worktree; `git status --short` must show exactly the 12 rev2 paths.
2. Then apply rework 2 exactly as written in 5nrmtt-rework-2.md (P1 positive diagnostic grammar; P2 Windows Job Object or typed refusal before process creation).
3. Narrow tests (-p 1), evidence, mutants, checklist, handoff rev3.
