# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260922-1cenbr via TASK-260922-cww1ov (bound developer run, curator)

Revision 12 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260922-1cenbr --cr TASK-260922-cww1ov --revision 12 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-cww1ov-land.log

Attach the log as `TASK-260922-cww1ov_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
