# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261004-3m2pt7 via BUG-261004-33fnzw (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261004-3m2pt7 --cr BUG-261004-33fnzw --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-33fnzw-land.log

Attach the log as `BUG-261004-33fnzw_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
