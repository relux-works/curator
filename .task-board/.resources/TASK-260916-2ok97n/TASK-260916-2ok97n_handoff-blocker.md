# TASK-260916-2ok97n handoff blocker

Date: 2026-09-23

## Constraint and evidence

Local R5 implementation and the current Darwin checks are recorded in
`TASK-260916-2ok97n_results.md` and the attached Darwin evidence bundle. The
task requires hosted Ubuntu/macOS/Windows evidence and rose-air evidence, but
campaign-producer-rules assign signed integration, hosted checks, and the
rose-air lane to the parent orchestrator. This producer run has no candidate
integration/landing run to observe those remote lanes. Windows real-interpreter
execution is therefore unverified; only the Windows test binary was
cross-compiled here.

The board handoff also enforces checklist completion before launching its own
landing runtime:

```text
task-board m 'set_status(TASK-260916-2ok97n, status=to-review)'
exit 1: spawned producer cannot move the task directly to to-review; use task-board handoff

task-board handoff TASK-260916-2ok97n --role developer
exit 1: unchecked checklist items [1 2 8]
```

Item 1 requires the unobserved hosted lanes and rose-air. Item 2 includes
“landing suite runs once via the handoff runtime”, but the handoff refuses
before it starts that runtime while the item is unchecked. Item 8 requests a
logbook record; no `logbook` executable is available and campaign rules
explicitly prohibit editing `LOGBOOK.md`. The findings and anomaly record are
attached to this task as outcomes instead.

## Attempts and options

No checklist item was checked by proxy. The production tests, narrowing mutants,
task-scoped results, Darwin platform gate, and three-platform ledger
consistency check are complete and attached. The handoff was attempted after
those artifacts were attached and was rejected by the board's precondition.

1. The orchestrator publishes/integrates the candidate through its signed path,
   runs hosted Ubuntu/macOS/Windows plus the landing suite and rose-air, then
   records those outcomes and completes the checklist. This follows the
   campaign ownership boundary but needs a supported path to start the landing
   run before the producer checklist item is marked.
2. The board owner changes the handoff checklist contract so producer handoff
   can leave explicitly delegated hosted/landing/logbook work unchecked, with
   attached evidence and owner recorded. This preserves the role handoff while
   keeping the external checks visible, but requires a workflow decision.

Recommended: the orchestrator/board owner provide an audited path that permits
the developer handoff with items 1 and 2 delegated to integration (and item 8
recorded as unavailable under the campaign rule), then have the parent run and
attach the hosted/landing evidence. Until that path or evidence is supplied,
the task cannot honestly be marked `to-review` by this run.

## Exact external input needed

The parent orchestrator or board owner must either (a) initiate the signed
candidate integration/landing workflow and own the hosted/rose-air results, or
(b) authorize the delegated-checklist handoff path described above. No product
or runtime design decision is outstanding.
