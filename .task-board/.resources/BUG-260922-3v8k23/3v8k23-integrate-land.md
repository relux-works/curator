# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-2mla0q via BUG-260922-3v8k23 (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-2mla0q --cr BUG-260922-3v8k23 --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3v8k23-land.log

Attach the log as `BUG-260922-3v8k23_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
