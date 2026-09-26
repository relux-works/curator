# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260924-3gd2d6 via TASK-260924-11burj (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260924-3gd2d6 --cr TASK-260924-11burj --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-11burj-land.log

Attach the log as `TASK-260924-11burj_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
