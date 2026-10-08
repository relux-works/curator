# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261008-11e9ay via TASK-261008-2wx7vm (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261008-11e9ay --cr TASK-261008-2wx7vm --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2wx7vm-land.log

Attach the log as `TASK-261008-2wx7vm_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
