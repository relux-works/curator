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
- [ ] env resolve --repair --dry-run reports a pending identity migration (lock schema/hash bump and sibling-env marker rewrites) with a distinct diag token instead of exiting 0; managed.go:2523 currently skips that branch under --dry-run, so a dry-run cannot gate a repair

## Notes

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-10-12T05:48:46Z

## Last Update
2026-10-12T10:36:58Z
