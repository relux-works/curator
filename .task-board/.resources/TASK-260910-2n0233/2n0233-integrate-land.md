# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260910-25yc0h via TASK-260910-2n0233 (bound developer run, curator)

Revision 4 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260910-25yc0h --cr TASK-260910-2n0233 --revision 4 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2n0233-land.log

Attach the log as `TASK-260910-2n0233_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
