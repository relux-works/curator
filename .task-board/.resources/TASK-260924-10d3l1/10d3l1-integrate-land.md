# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260924-1oyh2m via TASK-260924-10d3l1 (bound developer run, curator)

Revision 4 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260924-1oyh2m --cr TASK-260924-10d3l1 --revision 4 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-10d3l1-land.log

Attach the log as `TASK-260924-10d3l1_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
