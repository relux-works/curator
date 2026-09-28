# TASK-260922-1zfqq0 rework 1 → revision 2 (orchestrator brief, binding)

Revision 1's gate failed for a REAL reason (not a flake), and the cause is the orchestrator's brief, which
said "SPEC.md + README + CHANGELOG only, no Go code". This repository has a conformance gate that ties the
SPEC §6 diagnostics table to the code's registry: `internal/diagnostics` `gate_conformance_test.go`
(`TestGateOwnerFormPositives`: `Codes()` must list every normative code; `TestGateCoverageCounts`:
"normative codes = 21, want 18"). Your SPEC revision added three normative codes; the registry still has 18.

Fix (minimal, declaration only):
1. Register the three new codes in `internal/diagnostics` exactly as SPEC §6 names them, with the exit codes
   and owner forms the SPEC table states — DECLARATION ONLY. Do not implement resolution, refusal or
   transport behaviour: that is F-L1b (TASK-260922-2u5jzw). If a code cannot be declared without behaviour,
   say so and stop.
2. Update the gate's expected counts/positives so the gate proves SPEC ↔ registry equality at 21, and keep
   every existing row. Run `make check` (build, fmt-check, vet, test, race) with a real exit code.
3. Everything else in revision 1 stays as is (SPEC/README/CHANGELOG, `specVersion`/help golden bump).
Continue from the revision-1 tree (no checkout/clean/stash), append a "Revision 2" section to
`TASK-260922-1zfqq0_results.md`, then `task-board handoff TASK-260922-1zfqq0 --role developer`.
