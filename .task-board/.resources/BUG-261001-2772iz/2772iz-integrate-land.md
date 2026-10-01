# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-261001-17ali8 via BUG-261001-2772iz (bound developer run, curator)

Revision 2 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-261001-17ali8 --cr BUG-261001-2772iz --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2772iz-land.log

Attach the log as `BUG-261001-2772iz_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
