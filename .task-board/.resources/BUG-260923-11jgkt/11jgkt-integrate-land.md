# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-laeycm via BUG-260923-11jgkt (bound developer run, curator)

Revision 7 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-laeycm --cr BUG-260923-11jgkt --revision 7 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-11jgkt-land.log

Attach the log as `BUG-260923-11jgkt_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
