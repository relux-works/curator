# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261004-1lk8e9 via TASK-261004-3pvg2k (bound developer run, curator-spec)

Revision 1 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator-spec:

    task-board worktree integrate STORY-261004-1lk8e9 --cr TASK-261004-3pvg2k --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3pvg2k-land.log

Attach the log as `TASK-261004-3pvg2k_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
