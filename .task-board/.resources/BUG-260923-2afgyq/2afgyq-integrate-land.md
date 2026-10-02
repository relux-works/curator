# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260822-2lvw0e via BUG-260923-2afgyq (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260822-2lvw0e --cr BUG-260923-2afgyq --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2afgyq-land.log

Attach the log as `BUG-260923-2afgyq_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
