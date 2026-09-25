# TASK-260922-3bbvrs: gap-ledger-for-published-case-consumers

## Description
Generalise the draft-sources classification harness (internal/crossconformance/draftsources_semantic_test.go: driven|known-gap|bound|skipped with a tallied total asserted against the published case count) to the consumers that render every published case — internal/config TestManagerConfigV2Vectors and the system-config-v2 sibling first, then any other family that iterates a published case list. Add a committed ledger (.github/ci/conformance-gaps.tsv or the repository's chosen shape) whose rows are: family, case id, owning board element, one-line reason. The ratchet is the point: a known-gap row that starts passing MUST fail the gate until it is removed, so the ledger can only shrink.

## Scope
curator: the published-case consumers and their harness, the ledger file, the gate wiring that pins the counts, docs. Does not move SPEC_PIN (that is the sibling leaf) and does not change any implemented behaviour.

## Acceptance Criteria
1) each consumer classifies every published case and asserts driven+known-gap+bound+skipped equals the published count; 2) the ledger carries family, case id, owner and reason, and is the only place a gap is declared; 3) a passing known-gap row fails the gate (test proves the ratchet), an unlisted failing case fails the gate, a vanished case fails the gate; 4) narrowing mutants: drop the tally assertion, accept an unlisted gap, let a passing gap stay listed — each killed by a named test; 5) CHANGELOG and a docs paragraph stating a gap row is an owed implementation, never an accepted deviation
