## Status
done

## Review
none

## Task Class
metadata

## Estimate
estimated(fibonacci(2))

## Blocked By
- TASK-260822-1so0ym

## Blocks
- (none)

## Checklist
(empty)

## Notes
RE-SCOPED per TASK-260823-omp8zt impact-analysis: module-roots does not land separately — schema 8 is one shared bump, so the module-roots prose and vectors enter the SAME schema-8 candidate that TASK-260822-c0rxj7 creates and qualify together. This task reduces to: verify the module-roots family is complete inside the candidate (prose, schema, vectors, double regeneration) and record the evidence; the single landing PR with pin advances happens after both manager implementations qualify.
CARRY-FORWARD FROM REVIEWER OF TASK-260822-1so0ym (accepted, RUN-260823-82df18): before this lands, one normative module-root diagnostic still has zero conformance-vector coverage.

build_module_root_declaration_invalid is the fifth diagnostic in the profiles/manager.md module-root table; grepping the whole candidate tree at 6001dc3 hits only the prose. The agent-skill-v8 / csk-skill-v8 schema-cases cover its syntactic subset (dot, absolute, backslash, duplicate, parent, windows-device) but not the two clauses only a vector can express: a declared module directory that is not a real link-free directory in the snapshot, and one with no go.mod directly inside it. That is exactly the stated purpose of conformance/v1/vectors/module-roots.json (its doc comment: filesystem and build-graph cases JSON Schema cannot express).

Two smaller bijection branches are also unvectored: two effective directives resolving to the same declaration, and an unreadable annotation shape such as a three-token side.

This was not named by TASK-260822-1so0ym AC, so that task was accepted rather than reopened. Decide here whether to close the gap in this landing scope or open a follow-up under STORY-260822-1pm1c9 — either way it should be settled before the EPIC-260822-18ylpq consumer implementation stories build against these vectors. Details in TASK-260822-1so0ym_review-verdict.md and LOGBOOK 0053.

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260822-10udu1_family-completeness.md](file://TASK-260822-10udu1/TASK-260822-10udu1_family-completeness.md)

## Created
2026-08-22T16:01:00Z

## Last Update
2026-08-23T19:42:04Z

## Assigned To
orchestrator-inline
