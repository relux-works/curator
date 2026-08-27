## Status
backlog

## Assigned To
(none)

## Created
2026-07-20T02:09:20Z

## Last Update
2026-07-30T14:35:00Z

## Blocked By
- TASK-260720-12r55p

## Blocks
- TASK-260720-3s27te
- TASK-260720-31zeo2
- TASK-260728-1e6811

## Checklist
- [ ] Record the base SHA and create the task worktree only after the clean local main clone is fast-forwarded and dependency handoffs are present.
- [ ] Exercise real build and launch plus cache, dry-run, rollback, recovery, concurrency, and platform shim behavior.
- [ ] Run the cross-platform CI matrix or attach equivalent runner evidence plus strict mypy results.
- [ ] Keep the default curator-spec pin on the previous release and record the caller-supplied candidate suite digest as non-release evidence

## Notes
Cross-story release boundary from STORY-260720-21bsr2: cross-platform candidate evidence may use an explicitly supplied suite root, but this task must not commit an unreleased curator-spec ref. Official released-suite pin provenance is audited by TASK-260720-1utsx8 after TASK-260720-25d05o.
Solution-architecture audit 2026-07-20: scope and AC now keep candidate-suite validation out of the committed release pin. See outcome TASK-260720-3pemm6_release-boundary-plan.md; TASK-260720-1utsx8 alone advances the pin after TASK-260720-25d05o.

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260720-3pemm6_release-boundary-plan.md](file://TASK-260720-3pemm6/TASK-260720-3pemm6_release-boundary-plan.md) — Candidate-suite testing and qualified release-pin ownership sequence

## Estimate
estimated(fibonacci(8))
