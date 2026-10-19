# TASK-261018-70enqt: candidate refresh blocked by CR state

The requested production command ran with actual exit 1:

```text
$ task-board worktree refresh-candidate TASK-261018-70enqt
candidate refresh requires a rework revision; TASK-261018-70enqt is ready
```

Task status was moved to development (exit 0). `worktree status` (exit 0)
reports CR revision 2 ready, 118 changed paths, unchanged checkpoint
`402fdbc9b77803cb2d6b11f4f1cfd83961e6897b`, and the live producer lease.
Local main is `9c1605a7`; no fresh-authority qualification is claimed.

The refresh command rejects ready before preparing Git changes. Its admitted
states are changes_requested, stale, conflicted, aborted and draft (or absent
initial CR). Converge is explicitly orchestrator-only; withdraw also checks
orchestrator authority. No identity removal, direct ledger edit, manual
branch movement, or invented review verdict was used.

The orchestrator must authorize the existing unreviewed CR as rework through
the supported lifecycle before resuming this producer. If no such transition
is supported while CHANGELOG convergence conflicts, fix that shared lifecycle
contract in its owning repository first. The recommendation is to preserve the
candidate and repair admission, then refresh and resolve only CHANGELOG.
Validation of the old checkpoint would not satisfy the brief.

Only the requested refresh section was appended to the research gaps file.
Qualification code, tests, root README and real HOME were not edited by this
attempt. `git diff --exit-code -- README.md` and `git diff --check` both exited
0. No refreshed-tree full local suite, hosted gate, build, vet, lint, replay,
or handoff ran. Existing evidence is historical and not a refreshed-tree pass.
The updated gaps resource retains its original AC and mutant coverage maps.

Task-scoped logbook: the r3 assumption that this CR was rework-admissible was
false. The production refusal is persisted; status is blocked pending the
orchestrator's lifecycle action. No control-root file was directly edited.
