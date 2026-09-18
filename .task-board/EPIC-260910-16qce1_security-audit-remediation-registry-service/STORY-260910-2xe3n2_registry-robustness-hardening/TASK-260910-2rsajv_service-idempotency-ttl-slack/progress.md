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
- [ ] IDEMPOTENCY_TTL_SECONDS = 26h with a comment citing the profile minimum; docs say 26h
- [ ] Tests: retry between 24h and 26h deduplicated; > 26h expired; existing idempotency tests green
- [ ] CHANGELOG Unreleased entry R8
- [ ] pytest (with CURATOR_CONFORMANCE_ROOT) and mypy strict exit 0, transcripts in TASK-260910-2rsajv_results.md

## Notes

## Precondition Resources
- [TASK-260910-2rsajv_brief.md](file://TASK-260910-2rsajv/TASK-260910-2rsajv_brief.md) — Producer brief
- [remediation-registry-producer-rules.md](file://TASK-260910-2rsajv/remediation-registry-producer-rules.md) — Campaign rules for service tasks

## Outcome Resources
(none)

## Created
2026-09-10T14:47:01Z

## Last Update
2026-09-18T04:47:25Z
