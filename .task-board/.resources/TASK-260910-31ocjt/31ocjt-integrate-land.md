# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260928-1t6bto via TASK-260910-31ocjt (bound developer run, curator)

Revision 6 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260928-1t6bto --cr TASK-260910-31ocjt --revision 6 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-31ocjt-land.log

Attach the log as `TASK-260910-31ocjt_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
