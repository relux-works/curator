# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260916-2otjbn via TASK-260916-3oh0u8 (bound developer run, curator)

Revision 3 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260916-2otjbn --cr TASK-260916-3oh0u8 --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3oh0u8-land.log

Attach the log as `TASK-260916-3oh0u8_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
