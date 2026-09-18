# TASK-260917-16l2md: promote-spec-pin-rc12-with-wave1-manager-union

## Description
curator: move SPEC_PIN in .github/workflows/ci.yml to dced9b8317e0e8af79edf2d0539b32bd22b6c85b (v1.0.0-rc.12) in one candidate together with the union of the three held wave-1 manager candidates (E2 transitive system modules TASK-260916-55g9dg rev3 worktree state, E4 provider trust roots TASK-260916-3oh0u8 rev1, S4 MCP env passthrough bounds TASK-260910-gocke2 rev1; attached as candidate patches), resolving their overlaps in internal/config and CHANGELOG, so that TestManagerConfigV2Vectors and every other vector-driven test pass on the hosted gate at the new root; the ledger/platform-cases rows the candidates added for root-content skips are removed where the new root publishes the family.

## Scope
(define task scope)

## Acceptance Criteria
Hosted gate green on all lanes at SPEC_PIN dced9b8 with no root-content skip for a family the root publishes and no weakened test; each of E2/E4/S4 behaves per its landed spec (closed diagnostics, knob defaults, warn-first profiles as default, env status posture); the E2 rev2 review corrections stay closed; CHANGELOG carries the three entries and the pin note
