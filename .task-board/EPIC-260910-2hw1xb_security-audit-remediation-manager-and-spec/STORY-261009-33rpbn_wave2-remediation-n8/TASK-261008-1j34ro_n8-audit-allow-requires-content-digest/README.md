# N8: audit --allow parses a supported content identity before any filesystem access and refuses everything else; containment check before writing pin state

## Description
See .research/261004_inline-audit-wave-2.md for the reproduction, the production entry and the expected fix. Use the matching wave2 probe test as the regression (adapt it from expected-red to a passing assertion of the invariant).

## Scope
production fix plus regression test through the production entry

## Acceptance Criteria
The invariant holds; the wave2 probe for it passes; negative controls of the report still pass; hosted gate green.
