# TASK-260923-em42lw — rework 2 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.
(em42lw-brief / rework-1 are history: the brief's scope still defines the deliverable; rework-1's ledger fix held.)

Revision 2's hosted gate (run 35892109178) failed on every OS in `internal/config TestManagerConfigV2Vectors`
(schema2-minimal-defaults, -empty-environments-defaults, -partial-knobs-fill-defaults, -every-knob,
-overlay-path-source, -overlay-path-windows-backslash, …). The ONLY difference, got vs want:

    only in got: {"permissions": {}}      (everything else identical)

Your normalized manager config now emits the decision-0018 `environments.permissions` default. curator-spec
main's `conformance/v1/vectors/manager-config-v2.json` expects exactly that (`"permissions": {}`), but CI pins
rc.12 (`SPEC_PIN` dced9b8), whose vectors predate the member. Moving the pin is STORY-260922-2goxjs (18ex37),
not this leaf. Do NOT drop the member and do NOT loosen the comparison globally.

Fix narrowly in the vector harness: when the SUPPLIED root's expected `environments` object has no
`permissions` member (a pre-0018 root), accept exactly `permissions == {}` (the default) as the declared forward
difference and compare every other key byte-for-byte; a NON-default `permissions` against such a root must still
FAIL. Log the accommodation once per vector (`t.Logf`) naming the root revision. Rows:
- the harness accommodation against a pre-0018 fixture root (default passes, non-default fails);
- the same vectors against a local curator-spec main root (`CURATOR_CONFORMANCE_ROOT=<local
  curator-spec>/conformance/v1`) pass WITHOUT the accommodation (report counts).
Narrowing mutant (disposable copy): accept any `permissions` value against the old root → killed.
Bounded local runs of `./internal/config/...` (+ `-race`) and the gate scripts you touch. Append "Revision 3" to
results, then `task-board handoff TASK-260923-em42lw --role developer`. A `run_wrote_outside_worktree … policy
warn` block is a warning — verify the status moved to `to-review`. No LOGBOOK.md.
