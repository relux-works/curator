# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260916-8ql03k via BUG-260916-3aco9f (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260916-8ql03k --cr BUG-260916-3aco9f --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3aco9f-land.log

Attach the log as `BUG-260916-3aco9f_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
