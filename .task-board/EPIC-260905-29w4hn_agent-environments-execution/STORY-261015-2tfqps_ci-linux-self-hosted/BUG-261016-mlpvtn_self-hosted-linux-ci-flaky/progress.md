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
(empty)

## Notes
Fix direction (incident lane advice, 2026-10-16): workflow-side isolation without a host change: job-level GOCACHE=${{ runner.temp }}/go-build, GOMODCACHE=${{ runner.temp }}/gomod, GOPATH=${{ runner.temp }}/gopath, GOTOOLCHAIN=local; key any setup-go/cache entries by Go version. Host load: a workflow concurrency group, or fewer runners by the host owner (host changes need the owner OK; runner owner to be confirmed with Ivan or tb-keeper).

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-10-16T00:14:03Z

## Last Update
2026-10-16T00:20:10Z
