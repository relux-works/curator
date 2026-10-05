# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261004-3v2zm2 via TASK-261004-2iewnz (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261004-3v2zm2 --cr TASK-261004-2iewnz --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2iewnz-land.log

Attach the log as `TASK-261004-2iewnz_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
