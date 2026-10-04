# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261004-o4s9aq via BUG-261004-2v9pbz (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261004-o4s9aq --cr BUG-261004-2v9pbz --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2v9pbz-land.log

Attach the log as `BUG-261004-2v9pbz_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
