# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260928-9vt338 via TASK-260917-8vfgxf (bound developer run, curator)

Revision 3 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260928-9vt338 --cr TASK-260917-8vfgxf --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-8vfgxf-land.log

Attach the log as `TASK-260917-8vfgxf_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
