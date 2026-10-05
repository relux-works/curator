# TASK-261003-1uzji7 — developer handoff policy conflict

Implementation and targeted local verification are ready for review, but no
Change Request was published and no to-review transition is claimed.

The required command ran directly:

`task-board handoff TASK-261003-1uzji7 --role developer`

Actual exit: **1**. The CLI waited without output; it was not retried or
backgrounded across the end of a turn. Its terminal refusal was:

```
cannot hand off TASK-261003-1uzji7: unchecked checklist items [2 5 8]
(Every acceptance criterion proven with real exit codes; hosted gate green;
Lint clean; Important findings, decisions, anomalies, or regressions recorded
in logbook when relevant): handoff evidence missing
```

It also warned that board content comparison could not record progress for
the command. That warning is not a successful handoff.

## Constraint and failed assumption

The current, binding produce-mode brief requires ONLY targeted local tests,
uses the hosted Change Request gate as the arbiter, and explicitly forbids
LOGBOOK.md and CHANGELOG.md edits. The generic role checklist requires hosted
green and a logbook entry before the CLI will publish that Change Request.
Assuming that the current brief would permit deferred developer checklist
items at handoff was disproven by the real exit-1 refusal.

All local test/build evidence is already attached. Scoped gofmt and vet are
passing; full lint and hosted/cross-platform/race acceptance are pending, so
the full-lint checklist was also deliberately not represented as green.
No unrun gate was checked, no failing gate was represented as passing, and
no checklist gate was removed or weakened to get around this refusal.
LOGBOOK.md and CHANGELOG.md remain untouched.

## Options and recommendation

1. Recommended: the orchestrator reconciles this task's developer handoff
   checklist with the binding produce mode, explicitly assigning hosted
   acceptance/full lint to review and recording that logbook edits are
   prohibited for this task. Then route the persisted candidate to its
   hosted Change Request gate and review.
2. Change the produce-mode instruction and authorize the conflicting work.
   This still needs a supported route to hosted validation before the
   current handoff gate, and must not be inferred by the developer.

Exact external action needed: reconcile the handoff checklist/role policy so
that a producer can publish the Change Request with targeted evidence while
hosted acceptance remains pending and LOGBOOK.md is excluded, or provide a
replacement binding instruction and a valid hosted pre-handoff route.

The separate landing prerequisite remains unchanged: integrate only after
curator v0.15.0-rc.3. No branch commit, switch, merge, rebase, or release tag was
performed in this run. Code and task-scoped research/evidence remain
uncommitted in the assigned Story worktree.
