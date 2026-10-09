# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261009-2kq6u5 via TASK-261008-eaqg8a (bound developer run, curator)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261009-2kq6u5 --cr TASK-261008-eaqg8a --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-eaqg8a-land.log

Attach the log as `TASK-261008-eaqg8a_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file. Do not build or run tests on this host.
