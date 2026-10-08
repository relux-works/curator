# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260928-oflbe1 via TASK-260927-25hk87 (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260928-oflbe1 --cr TASK-260927-25hk87 --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-25hk87-land.log

Attach the log as `TASK-260927-25hk87_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
