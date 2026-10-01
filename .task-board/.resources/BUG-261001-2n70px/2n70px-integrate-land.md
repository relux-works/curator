# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261001-1rrh6z via BUG-261001-2n70px (bound developer run, curator)

Revision 3 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261001-1rrh6z --cr BUG-261001-2n70px --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2n70px-land.log

Attach the log as `BUG-261001-2n70px_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
