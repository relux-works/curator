# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260928-rp2r1j via TASK-260910-3i6vod (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260928-rp2r1j --cr TASK-260910-3i6vod --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3i6vod-land.log

Attach the log as `TASK-260910-3i6vod_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
