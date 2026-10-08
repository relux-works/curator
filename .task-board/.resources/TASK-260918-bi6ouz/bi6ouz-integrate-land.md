# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260916-12lbww via TASK-260918-bi6ouz (bound developer run, curator)

Revision 5 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260916-12lbww --cr TASK-260918-bi6ouz --revision 10 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-bi6ouz-land.log

Attach the log as `TASK-260918-bi6ouz_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
