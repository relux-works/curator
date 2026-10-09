# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261009-33rpbn via TASK-261008-1j34ro (bound developer run, curator)

Revision 4 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261009-33rpbn --cr TASK-261008-1j34ro --revision 4 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1j34ro-land.log

Attach the log as `TASK-261008-1j34ro_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file. Do not build or run tests on this host.
