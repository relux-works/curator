# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261002-1pd460 via BUG-261002-ot3ea1 (bound developer run, curator)

Revision 5 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261002-1pd460 --cr BUG-261002-ot3ea1 --revision 5 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-ot3ea1-land.log

Attach the log as `BUG-261002-ot3ea1_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
