# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261002-1w63le via TASK-261002-1zk11s (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261002-1w63le --cr TASK-261002-1zk11s --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1zk11s-land.log

Attach the log as `TASK-261002-1zk11s_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
