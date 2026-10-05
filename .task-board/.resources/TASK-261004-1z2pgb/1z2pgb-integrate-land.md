# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261004-7fglii via TASK-261004-1z2pgb (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261004-7fglii --cr TASK-261004-1z2pgb --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1z2pgb-land.log

Attach the log as `TASK-261004-1z2pgb_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
