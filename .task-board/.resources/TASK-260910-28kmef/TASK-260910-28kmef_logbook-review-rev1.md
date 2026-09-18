# R5 review logbook — TASK-260910-28kmef revision 1

2026-09-18: Independent review found direct canonicalization depth errors are CanonicalDepthError/ValueError, not the required ProtocolError. Both actual deep input and forced RecursionError reproduce it. HTTP handling passes four 20,000-depth probes. Changes requested for exception compatibility and direct regression tests.

Exact dced9b8 conformance run: 171 pass, one failure (missing checkpoint_cases); base c7ef32c reproduces the same failure. Actual CI pin is 47c3c8c. This is a campaign instruction/evidence mismatch, not introduced by R5; no pin or spec edits made. Current-spec 5146c7b suite: 172 pass; mypy strict passes. Two narrowing mutants killed. Full evidence: TASK-260910-28kmef_review-verdict-rev1.md.

No logbook executable or connector available in this session, so this logbook entry is persisted as a task-scoped board outcome via the authorized resource API.
