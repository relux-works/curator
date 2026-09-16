# TASK-260909-2vy977 lifecycle blocker

Implementation and narrow validation are preserved uncommitted. Current evidence is attached as TASK-260909-2vy977_current-results.md and TASK-260909-2vy977_mutants-current.tar.gz; historical recovery evidence is retained. Additional direct commands in zsh with set -o pipefail: GOWORK=off go vet ./internal/defaults ./internal/diagnostics ./cmd/curator-run exited 0; make fmt-check exited 0.

## Exact refusal
`task-board handoff TASK-260909-2vy977 --role developer` exited 1:
`cannot hand off TASK-260909-2vy977: unchecked checklist items [3 4] (... final publication works without a local workspace override.; ... configured validation pass ...): handoff evidence missing`

No CR or runtime validation was published. No manual make check was run. Rows 3 and 4 remain unchecked because their future publication/validation assertions cannot honestly be made yet. Row 3 was previously checked in historical state; this recovery unchecked it, preserving history. Other implementation checks are satisfied within the explicit leaf scope; generic planning clauses were evaluated as inapplicable (no decomposition/research/new dependency work).

## Constraint and failed assumption
The campaign mandates runtime-only make check exactly once at CR publication. Public handoff requires every checklist item already checked before moving to the state that triggers publication. Inspected source confirms cmd/runHandoff calls board.Handoff -> validateHandoffEvidence, which rejects any unchecked item before status mutation. The runtime cannot generate the required future evidence through this refused handoff. This is an ordering/ownership conflict, not a product-code defect or failed Go test.

## Options and exact input
Recommended: orchestrator separates post-handoff publication/runtime-validation attestations from producer pre-handoff checklist requirements, keeping them mandatory before independent reviewer acceptance. Then this exact candidate can use ordinary handoff; the runtime executes make check once and its log supplies the evidence.
Alternative: source-owned runtime provides a supported prepublication validation step with exact-result reuse, followed by handoff. This is broader tooling work and outside launcher scope.
Not acceptable: precheck unrun validation; substitute historical suite from another candidate; manually duplicate make check; edit private board records; bypass handoff.

Exact external decision needed: authorize and apply the producer-versus-runtime attestation ownership split (or supply an existing supported runtime validation operation). No requirements may be waived and no source workaround is proposed. Asked through the session clarification tool. Candidate and all evidence preserved; task routes blocked until lifecycle ownership is resolved. Host execution itself is now responsive; old outages are historical.
