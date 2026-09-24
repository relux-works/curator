# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-vkxt08 via BUG-260922-6chzf9 (bound developer run, curator)

Revision 4 is accepted (carry-forward without CHANGELOG). No board writes before or during. From
/Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-vkxt08 --cr BUG-260922-6chzf9 --revision 4 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-6chzf9-final.log

Attach the log as `BUG-260922-6chzf9_integration-final.md` and stop. If it refuses (write boundary / delivery / anything), attach the exact refusal
and stop — the orchestrator delivers. Change no file.
