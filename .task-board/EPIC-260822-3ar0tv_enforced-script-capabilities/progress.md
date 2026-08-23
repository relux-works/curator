## Status
development

## Review
required

## Task Class
code

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
Origin: skill-project-management install work (its board, TASK-260822-1gs27d) surfaced that script commands make capability declarations unenforceable — a script with network:none can still download and execute new code at run time. Core.md 4.3 states capabilities are an audit surface, not a runtime sandbox; shims are bare exec launchers. Draft decision written to curator-spec working tree: decisions/0007-enforced-script-capabilities.md (uncommitted, awaiting maintainer review). Doctrine deliberately mirrors decision 0006: manager-owned worker, mandatory portable controls vs versioned native inventory, capability-evidence records, no false kernel-guarantee claims, explicit versioned opt-in (schema 8) with declared-only labeling for legacy.

## Precondition Resources
(none)

## Outcome Resources
(none)

## Created
2026-08-22T14:51:20Z

## Last Update
2026-08-23T10:08:09Z
