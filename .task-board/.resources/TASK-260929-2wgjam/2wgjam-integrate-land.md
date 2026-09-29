# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260929-1s4r14 via TASK-260929-2wgjam (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260929-1s4r14 --cr TASK-260929-2wgjam --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2wgjam-land.log

Attach the log as `TASK-260929-2wgjam_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
