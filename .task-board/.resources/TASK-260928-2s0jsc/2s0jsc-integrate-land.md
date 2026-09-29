# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260928-16hi30 via TASK-260928-2s0jsc (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260928-16hi30 --cr TASK-260928-2s0jsc --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2s0jsc-land.log

Attach the log as `TASK-260928-2s0jsc_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
