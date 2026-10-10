# THE ONLY CURRENT INSTRUCTION — rework TASK-261010-aqpf2a to rev2 (researcher, minimal and mechanical)

The review of rev1 (`TASK-261010-aqpf2a_review-verdict-rev1.md`) accepted the content and requires exactly one change: **drop the `LOGBOOK.md` hunk; keep `.research/261010_pi-opencode-tool-lockdown.md` byte-identical to rev1.** Campaign rule: producers never edit `LOGBOOK.md`.

Do exactly this in the task's Story worktree:
1. Revert only the `LOGBOOK.md` change, so the file equals its content at the Change Request base. Touch no other path; no broad reset, no discarding of other work.
2. Verify before publishing: the diff from the base touches exactly one path, `.research/261010_pi-opencode-tool-lockdown.md`, and that file is byte-identical to the rev1 file (506 lines).
3. Publish Change Request rev2 through the normal producer flow the runner gives you, then `task-board handoff TASK-261010-aqpf2a --role researcher` and END YOUR TURN.

No other edits, no new research, no tests or builds on this host, no hand edits of board files.
