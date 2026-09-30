# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260930-15hioz via TASK-260930-1kylpg (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260930-15hioz --cr TASK-260930-1kylpg --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1kylpg-land.log

Attach the log as `TASK-260930-1kylpg_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
