# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261001-2peo2f via TASK-261001-3bsyvh (bound developer run, curator)

Revision 4 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261001-2peo2f --cr TASK-261001-3bsyvh --revision 4 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3bsyvh-land.log

Attach the log as `TASK-261001-3bsyvh_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
