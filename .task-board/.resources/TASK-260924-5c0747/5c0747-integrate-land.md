# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260924-2go2bz via TASK-260924-5c0747 (bound developer run, curator)

Revision 3 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260924-2go2bz --cr TASK-260924-5c0747 --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-5c0747-land.log

Attach the log as `TASK-260924-5c0747_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
