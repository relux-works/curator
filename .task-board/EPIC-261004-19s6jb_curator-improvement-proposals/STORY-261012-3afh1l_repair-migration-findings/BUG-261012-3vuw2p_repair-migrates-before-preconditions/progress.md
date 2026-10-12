## Status
backlog

## Review
required

## Task Class
code

## Estimate
notEstimated

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [ ] All target-env preconditions (credential targets, private targets, passthrough) are evaluated before the identity migration publishes any byte
- [ ] Or: a failure after the migration rolls the lock and every rewritten marker back in the same transaction
- [ ] Negative row: a schema-1 lock plus an absent muse auth target; --repair refuses and lock.json plus every sibling marker are byte-identical afterwards

## Notes

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-10-12T13:16:55Z

## Last Update
2026-10-12T13:17:34Z
