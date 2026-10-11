# R-CS2 candidate refresh refusal

Command: `task-board worktree refresh-candidate TASK-261011-2xdx9u`

Exit code: 1

Exact output:
```text
candidate refresh requires a rework revision; TASK-261011-2xdx9u is ready
```

The refresh brief explicitly requires attaching the exact output and stopping on refusal. No repository changes, tests, builds, or handoff were performed in this run. The existing candidate edits remain untouched. A rework revision is required before the requested refresh can be attempted again.
