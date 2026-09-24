# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260923-3vwgy4 via BUG-260923-3mazfw (bound developer run, curator)

Revision 2 (story_final) is accepted and it is the only leaf of
STORY-260923-3vwgy4 (macOS git fork/exec EACCES retry). Other Story integrations are landing on curator main right now;
the integrate serialises on the machine semaphore and reparents onto the moved trunk. No board writes before or
during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-3vwgy4 --cr BUG-260923-3mazfw --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3mazfw-2.log

Attach the log as `BUG-260923-3mazfw_integration-results.md` and stop. If it refuses as a stale CR (trunk changed
CHANGELOG.md too), attach the refusal and stop — the orchestrator runs invalidate-acceptance + refresh. Change no file.
