# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261003-3bx9x0 via TASK-261003-1uzji7 (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261003-3bx9x0 --cr TASK-261003-1uzji7 --revision 5 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1uzji7-land.log

Attach the log as `TASK-261003-1uzji7_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
