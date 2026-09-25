# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-3qwrnl via BUG-260923-krcm6m (bound developer run, curator)

Revision 4 is accepted (carry-forward without CHANGELOG). No board writes before or during. From
/Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-3qwrnl --cr BUG-260923-krcm6m --revision 4 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-krcm6m-final.log

Attach the log as `BUG-260923-krcm6m_integration-final.md` and stop. If it refuses (write boundary / delivery / anything), attach the exact refusal
and stop — the orchestrator delivers. Change no file.
