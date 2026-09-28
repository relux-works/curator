# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260910-2qmrb8 via TASK-260927-4pv4au (bound developer run, curator)

Revision 3 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260910-2qmrb8 --cr TASK-260927-4pv4au --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-4pv4au-land.log

Attach the log as `TASK-260927-4pv4au_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
