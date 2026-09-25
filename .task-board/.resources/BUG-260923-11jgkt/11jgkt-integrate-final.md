# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-laeycm via BUG-260923-11jgkt (bound developer run, curator)

Revision 6 is accepted (carry-forward without CHANGELOG). No board writes before or during. From
/Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-laeycm --cr BUG-260923-11jgkt --revision 6 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-11jgkt-final.log

Attach the log as `BUG-260923-11jgkt_integration-final.md` and stop. If it refuses (write boundary / delivery / anything), attach the exact refusal
and stop — the orchestrator delivers. Change no file.
