# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261001-38ijl9 via TASK-261001-1klixs (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261001-38ijl9 --cr TASK-261001-1klixs --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1klixs-land.log

Attach the log as `TASK-261001-1klixs_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
