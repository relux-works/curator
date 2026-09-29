# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260929-1rfp7e via TASK-260929-jup8re (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260929-1rfp7e --cr TASK-260929-jup8re --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-jup8re-land.log

Attach the log as `TASK-260929-jup8re_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
