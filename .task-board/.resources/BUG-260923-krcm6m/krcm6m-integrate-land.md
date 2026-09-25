# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-3qwrnl via BUG-260923-krcm6m (bound developer run, curator)

Revision 5 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-3qwrnl --cr BUG-260923-krcm6m --revision 5 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-krcm6m-land.log

Attach the log as `BUG-260923-krcm6m_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
