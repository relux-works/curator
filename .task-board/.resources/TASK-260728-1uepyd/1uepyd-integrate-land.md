# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260930-3feoy5 via TASK-260728-1uepyd (bound developer run, curator)

Revision 3 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260930-3feoy5 --cr TASK-260728-1uepyd --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1uepyd-land.log

Attach the log as `TASK-260728-1uepyd_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
