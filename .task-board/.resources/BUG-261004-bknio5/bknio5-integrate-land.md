# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261004-3oognx via BUG-261004-bknio5 (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261004-3oognx --cr BUG-261004-bknio5 --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-bknio5-land.log

Attach the log as `BUG-261004-bknio5_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
