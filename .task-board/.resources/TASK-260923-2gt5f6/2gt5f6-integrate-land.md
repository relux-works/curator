# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-1v3no2 via TASK-260923-2gt5f6 (bound developer run, curator)

Revision 3 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-1v3no2 --cr TASK-260923-2gt5f6 --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2gt5f6-land.log

Attach the log as `TASK-260923-2gt5f6_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
