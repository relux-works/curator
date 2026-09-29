# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260910-6bo7ej via TASK-260910-2vnjej (bound developer run, curator)

Revision 7 is ACCEPTED. No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260910-6bo7ej --cr TASK-260910-2vnjej --revision 7 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2vnjej-land.log

Attach the log as `TASK-260910-2vnjej_integration-land.md` and stop. If it refuses (write boundary / delivery / stale / anything), attach the exact
refusal and stop — the orchestrator delivers. Change no file.
