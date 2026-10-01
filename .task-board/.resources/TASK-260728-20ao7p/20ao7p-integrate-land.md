# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260728-1ojb1p via TASK-260728-20ao7p (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260728-1ojb1p --cr TASK-260728-20ao7p --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-20ao7p-land.log

Attach the log as `TASK-260728-20ao7p_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
