# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-2btaia via TASK-260923-2elcdc (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-2btaia --cr TASK-260923-2elcdc --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2elcdc-land.log

Attach the log as `TASK-260923-2elcdc_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
