# N9: audit pins record a defined creation time per the manager source-audit contract, with a validity test

## Description
See .research/261004_inline-audit-wave-2.md for the reproduction, the production entry and the expected fix. Use the matching wave2 probe test as the regression (adapt it from expected-red to a passing assertion of the invariant).

## Scope
production fix plus regression test through the production entry

## Acceptance Criteria
The invariant holds; the wave2 probe for it passes; negative controls of the report still pass; hosted gate green.
