# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260910-148pj1 via TASK-260910-32gki6 (bound developer run, curator)

Revision 5 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260910-148pj1 --cr TASK-260910-32gki6 --revision 5 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-32gki6-land.log

Attach the log as `TASK-260910-32gki6_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
