# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260906-1a2i5a via TASK-260925-h4syhu (bound developer run, curator)

Revision 7 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260906-1a2i5a --cr TASK-260925-h4syhu --revision 7 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-h4syhu-land.log

Attach the log as `TASK-260925-h4syhu_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
