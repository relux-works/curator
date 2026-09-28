# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-11vn9k via BUG-260922-k6eypp (bound developer run, curator)

Revision 3 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-11vn9k --cr BUG-260922-k6eypp --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-k6eypp-land.log

Attach the log as `BUG-260922-k6eypp_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
