# Candidate refresh blocked before validation

## Requested scope
Refresh the existing uncommitted SNAP-LOCK candidate onto current trunk, retain both README additions exactly once, rerun the focused tests and full remote gate, then hand off for review. No implementation changes are requested in this refresh.

## Observed constraint
The live producer owns the Story lease, but current Change Request revision 1 remains `ready`, not a rework revision. The Story checkpoint is still `0d00a109ad1da77a5506747c89f24af9442eb341`. The inspected trunk is `dccac3199aa5c543c589c129f5ef2bc437523a2a` (five commits ahead). No branch, index or candidate source was changed by this run.

| Command | Actual exit | Result |
| --- | ---: | --- |
| `task-board m 'set_status(TASK-261016-4zofp0, status=development)'` | 0 | Producer lifecycle entered |
| `task-board worktree refresh-candidate STORY-261016-1wj1v9` | 1 | `candidate refresh producer binding mismatch`; corrected the argument to the assigned leaf |
| `task-board worktree refresh-candidate TASK-261016-4zofp0` | 1 | `candidate refresh requires a rework revision; TASK-261016-4zofp0 is ready` |
| `task-board worktree status STORY-261016-1wj1v9 --json` | 0 | Confirms revision 1 ready and the unchanged old checkpoint |
| `task-board worktree converge --help` | 0 | Convergence is operator-only; tracked runs are expressly refused |
| `task-board worktree invalidate-acceptance --help` | 0 | Requires integrating plus stale/accepted evidence; this ready revision is ineligible |

`withdraw_cr` is also operator-only, per the project-management change-request lifecycle reference. No attempt was made to impersonate an operator, remove the run binding, invent a reviewer verdict, or edit board records.

## README inspection
The current candidate retains its internal capture/metadata-status validation paragraph. Current trunk additionally moves the existing managed-login warning outside its details block, adds the agent-logins reference, and adds PlantUML/Lychee tool documentation. Those trunk additions are absent from the old candidate. A successful refresh must retain all of them plus the capture paragraph exactly once. Editing README alone now would not update the checkpoint or register a valid refresh, so no speculative file replacement was made.

## Validation identity and limits
The previous handoff remains historical evidence: signed snapshot `1e6be490fee5768d0aa6808b825a20437058ef1c`, tree `b91a63aa36e9e266a7514d9a4d92356181bbfc6f`, full gate run 38845304296 exit 0. It does not qualify a refreshed candidate that does not yet exist.

Focused tests and `sh scripts/remote-gate.sh` were NOT rerun in this run: the prerequisite refresh refused. No new green result, mutation kill, runtime qualification, or review readiness is claimed. Earlier coverage/mutation artifacts remain bound to their recorded old source/test/configuration/environment identity. No new behavior was implemented, so new product negative tests are inapplicable to this blocked refresh.

## Recovery choices and required external input
1. Recommended: the board operator restores a supported rework transition for revision 1 while preserving its candidate and history, then resumes this producer so refresh, README reconciliation, focused tests, and the full gate can execute.
2. If the board offers no supported transition for a ready revision with a conflicting trunk delta, repair that reusable recovery contract in its owning tool repository. This producer must not work around it by manually merging the managed Story or rewriting board state.

The exact required external input is a supported operator-side disposition that makes this candidate eligible for producer refresh. No product or implementation decision is needed. Existing source and tests remain uncommitted and preserved.

## Logbook
Refresh discovered a lifecycle mismatch: the brief expects rework eligibility, while persisted revision 1 is ready. The documented operator recovery routes cannot be executed by this tracked producer. The task is routed to blocked for that ownership boundary, with validation deferred until refresh succeeds.
