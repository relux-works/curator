# N7: install.Global reconciles a recognized manager-owned legacy forwarding link (exact target through backup/rollback; foreign-file and changed-preimage refusals kept)

## Description
See .research/261004_inline-audit-wave-2.md for the reproduction, the production entry and the expected fix. Use the matching wave2 probe test as the regression (adapt it from expected-red to a passing assertion of the invariant).

## Scope
production fix plus regression test through the production entry

## Acceptance Criteria
The invariant holds; the wave2 probe for it passes; negative controls of the report still pass; hosted gate green.
