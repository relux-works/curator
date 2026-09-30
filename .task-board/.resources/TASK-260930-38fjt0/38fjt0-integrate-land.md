# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260930-1soi12 via TASK-260930-38fjt0 (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260930-1soi12 --cr TASK-260930-38fjt0 --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-38fjt0-land.log

Attach the log as `TASK-260930-38fjt0_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
