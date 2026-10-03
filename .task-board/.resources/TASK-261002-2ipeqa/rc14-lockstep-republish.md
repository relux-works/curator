# THE ONLY CURRENT INSTRUCTION — TASK-261002-2ipeqa republish → revision 2 (orchestrator, binding)

The orchestrator converged your workspace onto trunk 64345d71, and your delta was carried without conflict. Trunk now includes:
- BUG-260923-2afgyq: the five marker cross-field cases now PASS, and their gap rows are removed for the existing digests;
- TASK-261002-9w4wy3: the two Windows hard-link cases now pass, and their rows are removed.

Do:
1. Recompute the rc.14 digest (6f832d81…) count and gap rows on THIS trunk. The rc.14 gap rows must NOT list those seven now-passing cases: derive them from the actual production behaviour, not from the old candidate template. Exact counts. CHANGELOG keeps both sides.
2. Run, with real exit codes and following host-rules (-work):
   - `CURATOR_CONFORMANCE_ROOT=<curator-spec candidate-rc14>/conformance/v1 go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance|Coverage' -count=1`;
   - the default-pin run;
   - `go test ./internal/marker ./internal/scriptworker -count=1`.
3. Add a "Revision 2" section to the results. Then run `task-board handoff TASK-261002-2ipeqa --role developer` and END YOUR TURN.

The writer stays OFF and SPEC_PIN stays unchanged. Never edit LOGBOOK.md.

## Decision (binding, ~10:40Z): the hosted gate is the arbiter
The local stalls come from the host's syspolicyd crash-loop, not from the change. The CR validation suite runs the full matrix on hosted GitHub runners (scripts/remote-gate.sh), including the exact count and coverage rows.

Do:
- Record what ran locally and what stalled, as you did.
- Do not retry stalled local runs more than once.
- Hand off so the hosted gate validates revision 2. Do not set the task to blocked for this reason.
