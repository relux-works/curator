# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261006-1fpobd via TASK-261006-3ptm2x (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261006-1fpobd --cr TASK-261006-3ptm2x --revision 7 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3ptm2x-land.log

Attach the log as `TASK-261006-3ptm2x_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
