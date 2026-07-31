## Status
backlog

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- BUG-260731-33v6zz

## Blocks
- (none)

## Checklist
- [ ] Parse the current combined Windows raw go-test evidence and name all six failing top-level cases.
- [ ] Fix the transaction and fingerprint platform defects without skips, tolerances, or ledger relaxation.
- [ ] Add focused Windows regression coverage and preserve Linux/macOS behavior.
- [ ] Publish a signed Curator PR and prove the six cases green on native windows-latest.
- [ ] Attach evidence, obtain independent Opus review, and land the accepted PR to main.

## Notes
Created from RUN-260731-1b588e round-8 evidence. Scheduling dependency is BUG-260731-33v6zz / PR 13 so the residual fix is developed and validated against the combined Windows base and does not collide in internal/godriver. PR 12 is already merged as b30773b8.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-07-31T14:13:39Z

## Last Update
2026-07-31T14:13:51Z
