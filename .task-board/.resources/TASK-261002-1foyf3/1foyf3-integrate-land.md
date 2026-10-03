# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261002-1r1hz7 via TASK-261002-1foyf3 (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261002-1r1hz7 --cr TASK-261002-1foyf3 --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1foyf3-land.log

Attach the log as `TASK-261002-1foyf3_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
