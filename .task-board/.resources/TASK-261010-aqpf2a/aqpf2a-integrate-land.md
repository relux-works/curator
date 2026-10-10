# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261010-14k25n via TASK-261010-aqpf2a (bound developer run, curator)

Revision 2 is ACCEPTED (delta review RUN-261010-4ba650). No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261010-14k25n --cr TASK-261010-aqpf2a --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-aqpf2a-land.log

Attach the log as `TASK-261010-aqpf2a_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file. Do not build or run tests on this host.
