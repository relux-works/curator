# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261003-3m6kh1 via TASK-261003-1kcv6v (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261003-3m6kh1 --cr TASK-261003-1kcv6v --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1kcv6v-land.log

Attach the log as `TASK-261003-1kcv6v_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
