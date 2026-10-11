# Candidate refresh refusal

Command:
```text
task-board worktree refresh-candidate TASK-261011-2xdx9u
```

Actual exit code: 1.

Exact output:
```text
candidate refresh requires a rework revision; TASK-261011-2xdx9u is ready
```

The current refresh brief requires attaching the exact output and stopping if refresh-candidate refuses. No repository edits, tests, builds, or handoff were performed after this refusal. Existing worktree changes were preserved. The board lifecycle must admit a rework revision before candidate refresh can proceed.
