# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260930-2o0ybs via TASK-260930-2mtgv7 (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260930-2o0ybs --cr TASK-260930-2mtgv7 --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2mtgv7-land.log

Attach the log as `TASK-260930-2mtgv7_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
