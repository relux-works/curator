# TASK-260713-c7a18d: harden-shell-activation

## Description
Implement the seamless command and optional shell activation contract, including regression tests and user-facing guidance.

## Acceptance Criteria
- See parent story acceptance criteria
- No shell profile mutation is performed automatically
- Existing unmanaged commands in a user bin are never overwritten
- Installation remains deterministic when no safe user bin exists
