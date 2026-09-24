# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260924-2tyzhh via BUG-260924-5p8b0z (bound developer run, curator)

Revision 1 is accepted (no CHANGELOG, no trunk intersection since base 5b326aa3). No board writes before or during. From
/Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260924-2tyzhh --cr BUG-260924-5p8b0z --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-5p8b0z-final.log

Attach the log as `BUG-260924-5p8b0z_integration-final.md` and stop. If it refuses (write boundary / delivery / anything), attach the exact refusal
and stop — the orchestrator delivers. Change no file.
