# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260915-3w11un via BUG-260922-306v4m (bound developer run, curator)

Revision 8 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260915-3w11un --cr BUG-260922-306v4m --revision 8 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-306v4m-land.log

Attach the log as `BUG-260922-306v4m_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
