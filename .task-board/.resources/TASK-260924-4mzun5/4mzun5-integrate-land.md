# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260924-1ckno7 via TASK-260924-4mzun5 (bound developer run, curator)

Revision 5 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260924-1ckno7 --cr TASK-260924-4mzun5 --revision 5 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-4mzun5-land.log

Attach the log as `TASK-260924-4mzun5_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file. Do not build or run tests on this host.
