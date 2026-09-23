# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260908-g7o5zw via TASK-260908-2kqa77 (bound developer run, curator)

Revision 6 (base refresh onto fad88136, refresh-reviewed) is accepted and it is the last live leaf of
STORY-260908-g7o5zw (goreleaser config value gate). Other Story integrations are landing on curator main right now;
the integrate serialises on the machine semaphore and reparents onto the moved trunk. No board writes before or
during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260908-g7o5zw --cr TASK-260908-2kqa77 --revision 6 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2kqa77-6.log

Attach the log as `TASK-260908-2kqa77_integration-results.md` and stop. If it refuses as a stale CR (trunk changed
CHANGELOG.md too), attach the refusal and stop — the orchestrator runs invalidate-acceptance + refresh. Change no file.
