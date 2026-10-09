# THE ONLY CURRENT INSTRUCTION — republish TASK-260924-4mzun5 on current trunk (developer; no code changes)

Your rework-3 candidate (Change Request revision 3) was never reviewed: trunk advanced on CHANGELOG.md (the N6 landing a443dfb5, then the N7 landing 97822069; trunk is now 97822069) and the reviewer spawn was refused with `revision_base_superseded`. The orchestrator ran `task-board worktree converge STORY-260924-1ckno7` twice: the workspace now sits on 97822069 with your whole delta carried over, and revision 3 is stale. Verified by the orchestrator: 45 of the 46 changed tracked files and all 11 untracked files are byte-identical to your rework-3 tree, and CHANGELOG.md is trunk's file plus your entry (a pure addition). The index was reset for the converge, so the new files you had staged are untracked again; they belong to the candidate.

Do exactly this:
1. In the Story worktree, check `git status` and `git diff 97822069 -- CHANGELOG.md` (only your entry is added). Do not edit any file, except to repair an obvious merge defect in CHANGELOG.md (there should be none).
2. Do NOT build or run tests on this host: no `go test`, no compiled test binaries, no `go run`. They are mechanically refused and they harm the host. The runner's hosted gate validates the Change Request at handoff.
3. Do not touch LOGBOOK.md or the board files directly.
4. Republish with `task-board handoff TASK-260924-4mzun5 --role developer`. If the handoff requires the task to be in development first, run `set_status(TASK-260924-4mzun5, status="development")` and retry once. If the handoff refuses for any other reason, record the exact refusal in the task notes and stop; do not work around it.
5. END YOUR TURN.
