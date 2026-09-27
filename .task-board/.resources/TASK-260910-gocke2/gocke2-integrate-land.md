# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260910-1lf0m5 via TASK-260910-gocke2 (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260910-1lf0m5 --cr TASK-260910-gocke2 --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-gocke2-land.log

Attach the log as `TASK-260910-gocke2_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
