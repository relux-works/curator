# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260928-1xu5sf via BUG-260928-2bfhgj (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260928-1xu5sf --cr BUG-260928-2bfhgj --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2bfhgj-land.log

Attach the log as `BUG-260928-2bfhgj_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
