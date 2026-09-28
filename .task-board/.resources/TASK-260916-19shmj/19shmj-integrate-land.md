# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260916-73a5zg via TASK-260916-19shmj (bound developer run, curator)

Revision 6 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260916-73a5zg --cr TASK-260916-19shmj --revision 6 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-19shmj-land.log

Attach the log as `TASK-260916-19shmj_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
