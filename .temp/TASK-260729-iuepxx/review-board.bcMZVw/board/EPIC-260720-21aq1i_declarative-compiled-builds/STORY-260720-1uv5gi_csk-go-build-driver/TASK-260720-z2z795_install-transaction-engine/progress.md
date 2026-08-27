## Status
backlog

## Assigned To
(none)

## Created
2026-07-20T02:09:19Z

## Last Update
2026-07-20T03:01:20Z

## Blocked By
- TASK-260720-1pvfj5

## Blocks
- TASK-260720-2x6mjn

## Checklist
- [ ] Record the base SHA and create the task worktree only after the clean local main clone is fast-forwarded and dependency handoffs are present.
- [ ] Exercise deterministic ordering, crash recovery, reverse rollback, concurrent consumers, stale-preimage defense, and lock-order failures.
- [ ] Run focused pytest plus python -m mypy and attach task-scoped evidence.

## Notes

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260720-z2z795_csk-build-lifecycle.puml](file://TASK-260720-z2z795/TASK-260720-z2z795_csk-build-lifecycle.puml) — PlantUML source for audit, dry-run, build, commit, rollback, and GC ordering
- [TASK-260720-z2z795_csk-build-lifecycle.svg](file://TASK-260720-z2z795/TASK-260720-z2z795_csk-build-lifecycle.svg) — Rendered csk compiled-build lifecycle diagram
