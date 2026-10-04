# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261002-2327ef via TASK-261004-31fcvu (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261002-2327ef --cr TASK-261004-31fcvu --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-31fcvu-land.log

Attach the log as `TASK-261004-31fcvu_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
