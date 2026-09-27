# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260916-1ll22r via TASK-260918-ryh3kw (bound developer run, curator)

Revision 6 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260916-1ll22r --cr TASK-260918-ryh3kw --revision 6 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-ryh3kw-land.log

Attach the log as `TASK-260918-ryh3kw_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
