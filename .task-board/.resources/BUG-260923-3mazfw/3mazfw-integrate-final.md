# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-3vwgy4 via BUG-260923-3mazfw (bound developer run, curator)

Revision 5 is accepted (carry-forward without CHANGELOG). No board writes before or during. From
/Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-3vwgy4 --cr BUG-260923-3mazfw --revision 5 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3mazfw-final.log

Attach the log as `BUG-260923-3mazfw_integration-final.md` and stop. If it refuses (write boundary / delivery / anything), attach the exact refusal
and stop — the orchestrator delivers. Change no file.
