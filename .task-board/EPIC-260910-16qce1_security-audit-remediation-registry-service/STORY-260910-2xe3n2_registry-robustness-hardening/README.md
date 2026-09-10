# STORY-260910-2xe3n2: registry-robustness-hardening

## Description
Findings R4, R5, R7, R8 (Low): pathological JSON yields RecursionError as 500 instead of 400; idempotency retention is exactly the 24h minimum with no slack; verify-backup defaults to keys from the live home; proxy deployments share one rate-limit bucket.

## Scope
curator-skill-registry protocol.py/app.py/cli.py + deployment docs

## Acceptance Criteria
RecursionError maps to 400 invalid_json; idempotency retention has >=26h slack; verify-backup requires or loudly warns on implicit key resolution; deployment docs cover proxy rate-limit bucketing
