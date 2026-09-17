# Rollback instruction — TASK-260910-19w2aj (bound developer run)

The prepared integration transaction of STORY-260910-3vxe3y (candidate 57925471 on expected trunk a511835) went stale because the shared control-root main advanced before the fast-forward. The orchestrator restored the local main ref to the expected trunk. Rules: no board writes; run in the FOREGROUND from the control root /Users/administrator/Developer/ReluxWorks/curator/curator exactly:

    task-board worktree integrate STORY-260910-3vxe3y --cr TASK-260910-19w2aj --revision 11 --rollback 2>&1 | tee .temp/integrate-19w2aj-rollback.log

Then attach `.temp/integrate-19w2aj-rollback.log` as outcome resource `TASK-260910-19w2aj_integration-rollback.md` and stop. Do NOT run a normal integrate afterwards; the orchestrator re-runs it on the current trunk. If it refuses, attach the exact refusal and stop.
