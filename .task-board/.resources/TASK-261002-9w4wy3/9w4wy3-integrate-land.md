# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260925-1v7pvn via TASK-261002-9w4wy3 (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260925-1v7pvn --cr TASK-261002-9w4wy3 --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-9w4wy3-land.log

Attach the log as `TASK-261002-9w4wy3_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
