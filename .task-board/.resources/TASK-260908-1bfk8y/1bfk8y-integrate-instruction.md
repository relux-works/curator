# Integration instruction — TASK-260908-1bfk8y (bound developer run, curator)

Revision 1 is accepted and it is the last live leaf of STORY-260907-2bddfc (its sibling
TASK-260906-284db9 is already done), so this integrate closes the Story. Control root and board are
both the curator repository. If trunk moved since the review, let the integrate reparent and
revalidate (the gate is `sh scripts/remote-gate.sh`, hosted). Rules: no board writes (no resources,
notes, checklist, set_status) before or during the integrate. Run exactly, from the control root
/Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260907-2bddfc --cr TASK-260908-1bfk8y --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1bfk8y.log

ONLY AFTER it exits: attach `.temp/integrate-1bfk8y.log` as outcome resource
`TASK-260908-1bfk8y_integration-results.md` and stop. If it refuses, the exact refusal is in the
log; do not retry, do not handoff, no code changes.
