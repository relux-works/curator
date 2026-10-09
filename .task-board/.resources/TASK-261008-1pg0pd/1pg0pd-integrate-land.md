# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261009-3246nu via TASK-261008-1pg0pd (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261009-3246nu --cr TASK-261008-1pg0pd --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1pg0pd-land.log

Attach the log as `TASK-261008-1pg0pd_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file. Do not build or run tests on this host.
