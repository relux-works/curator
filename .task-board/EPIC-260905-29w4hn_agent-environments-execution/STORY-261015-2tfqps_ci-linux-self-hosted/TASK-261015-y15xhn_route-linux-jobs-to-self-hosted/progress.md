## Status
done

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
Implemented inline by the orchestrator (small CI routing change). Landed in curator main 9e671d78 via PR #121 (fast-forward of the exact reviewed head). Orchestrator security review recorded on the PR: self-hosted routing only for push, workflow_dispatch and same-repository pull requests; fork pull requests stay hosted; no pull_request_target or secrets. All 8 Linux jobs green on the self-hosted runners.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-10-15T03:35:18Z

## Last Update
2026-10-15T10:36:30Z
