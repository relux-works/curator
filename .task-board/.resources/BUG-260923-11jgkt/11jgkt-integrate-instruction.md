# Integration instruction — STORY-260923-laeycm via BUG-260923-11jgkt (bound developer run, curator)

Revision 1 (story_final) is accepted and it is the only leaf of
STORY-260923-laeycm (Windows snapshot sharing-violation fix). Other Story integrations are landing on curator main right now;
the integrate serialises on the machine semaphore and reparents onto the moved trunk. No board writes before or
during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-laeycm --cr BUG-260923-11jgkt --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-11jgkt-1.log

Attach the log as `BUG-260923-11jgkt_integration-results.md` and stop. If it refuses as a stale CR (trunk changed
CHANGELOG.md too), attach the refusal and stop — the orchestrator runs invalidate-acceptance + refresh. Change no file.
