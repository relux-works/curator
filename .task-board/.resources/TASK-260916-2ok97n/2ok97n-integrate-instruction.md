# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260822-2h0v9j via TASK-260916-2ok97n (bound developer run, curator)

Revision 7 (story_final) is accepted; R1-R4 are checkpointed on the Story branch (refreshed onto 1511b345). Trunk is frozen at 1511b345
for this landing. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260822-2h0v9j --cr TASK-260916-2ok97n --revision 7 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2h0v9j.log

Attach the log as `TASK-260916-2ok97n_integration-results.md` and stop. If it refuses (write boundary, anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
