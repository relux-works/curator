# TASK-261010-1992si rev2 refresh-candidate refusal

Command: `task-board worktree refresh-candidate TASK-261010-1992si` — exit 1

```
change_request_candidate_drift: candidate refresh would leave tracked board paths outside its file transaction stale; synchronize the board with the prepared checkpoint before retrying (paths=.task-board/.activity/EPIC-261004-19s6jb/events.ndjson, .task-board/.activity/STORY-261009-33rpbn/events.ndjson, .task-board/.activity/STORY-261010-14k25n/events.ndjson, .task-board/.activity/TASK-261010-aqpf2a/events.ndjson, .task-board/EPIC-261004-19s6jb_curator-improvement-proposals/STORY-261010-14k25n_harness-lockdown-pi-opencode/TASK-261010-aqpf2a_research-pi-opencode-tool-disable/progress.md, .task-board/EPIC-261004-19s6jb_curator-improvement-proposals/progress.md, prepared_tree_oid=e69805f534e7922e87baee096bd40414fb5ec903, workspace_tree_oid=2793eb6e0b393d8c6495a6e694f5d47f5cf7281b)
```

Study file unchanged: sha256 2dabff8504032b0c68439cd7b3a6990874684449a48a7f525a6db69fae363239. Nothing resolved by hand; no handoff run. Orchestrator must synchronize the worktree board copy with the prepared checkpoint, then retry.
