# Integration refusal — CR revision 2

Run: RUN-260906-18b608 (developer / implementer)
Accepted CR: CR-TASK-260906-1xitqi-2 revision 2
Candidate base: 7320bc2adbd15aa5ade4e78ef7ef9008274e7478
Protected main: b056e5dae73be4dc92f2a948992f0941d7283f89

## Attempts and exact outcomes

1. Initial integrate exited 1 with integration_base_moved because control HEAD was 7320bc2 while freshly observed protected main was b056e5d. No transaction/ref movement.
2. Independently verified refs/heads/main by ls-remote and exact-ref fetch; both advertised and fetched b056e5d. Preserved authoritative board state through a path-scoped Git stash, fast-forwarded control main, restored the exact pre-refresh ledger (SHA-256 a50579c48912d71443f2026c9b1a62e1f5c8a989f88d30e897034998780a76fc), and confirmed ledger mirror 130/130 with 0 pending, pre-boundary, unowned, or journaled failures.
3. Retried identical integrate command; exited 1 with integration_base_moved: upstream and accepted revision both change LOGBOOK.md, so the combination is unreviewed. Revision 2 is now stale. No integration transaction exists and no new commit was created. Control HEAD remains b056e5d.

## Constraint and required route

The typed overlap refusal must not be bypassed. Preserve both LOGBOOK.md histories and all completed b056e5d stage-b/policy changes. Parent must route a tracked developer rework on refreshed main, publish a new CR revision, and obtain fresh reviewer acceptance and exact-tree validation before rerunning producer-bound integration. Do not push, tag, or publish before that succeeds.

## Exact input needed

A newly reviewed CR revision based on b056e5d that carries both LOGBOOK.md entries without replacing upstream content, plus its configured validation evidence.
