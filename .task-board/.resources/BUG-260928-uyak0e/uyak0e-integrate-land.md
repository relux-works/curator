# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260928-7eowfl via BUG-260928-uyak0e (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260928-7eowfl --cr BUG-260928-uyak0e --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-uyak0e-land.log

Attach the log as `BUG-260928-uyak0e_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
