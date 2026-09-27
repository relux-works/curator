# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260916-2d9coh via TASK-260916-55g9dg (bound developer run, curator)

Revision 5 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260916-2d9coh --cr TASK-260916-55g9dg --revision 5 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-55g9dg-land.log

Attach the log as `TASK-260916-55g9dg_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
