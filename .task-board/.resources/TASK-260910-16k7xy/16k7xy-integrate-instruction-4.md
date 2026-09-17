# Integration instruction (rollback + retry) — TASK-260910-16k7xy (bound developer run)

The prepared integration transaction of STORY-260910-24nyb1 (candidate 3532b1bd on expected trunk c64ceaf) is stale: the shared control-root main advanced to a foreign board commit before the fast-forward, and a resumed integrate refuses `integration_indeterminate`. Rules: no board writes (no resources, notes, checklist, set_status) before or during; run in the FOREGROUND from the control root /Users/administrator/Developer/ReluxWorks/curator/curator; a single foreground call may stay open ~45 minutes (remote gate).

Step 1 — roll the prepared transaction back:

    task-board worktree integrate STORY-260910-24nyb1 --cr TASK-260910-16k7xy --revision 6 --rollback 2>&1 | tee .temp/integrate-16k7xy-4a.log

Step 2 — only if step 1 succeeded (exit 0), integrate afresh (reparent onto the current trunk + revalidation):

    task-board worktree integrate STORY-260910-24nyb1 --cr TASK-260910-16k7xy --revision 6 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-16k7xy-4b.log

ONLY AFTER both exit: attach `.temp/integrate-16k7xy-4a.log` and `.temp/integrate-16k7xy-4b.log` as outcome resources `TASK-260910-16k7xy_integration-rollback.md` and `TASK-260910-16k7xy_integration-results-4.md`, then stop. If either refuses, attach the exact refusal and stop; do not retry, do not handoff, no code changes.
