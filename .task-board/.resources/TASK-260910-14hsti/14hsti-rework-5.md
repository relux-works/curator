# Rework 5 — TASK-260910-14hsti: one finding from verdict rev7 (TASK-260910-14hsti_review-verdict-rev7.md)

F1 HIGH — Snapshot.Durable (internal/staging/boundaries.go:527–552, FileIdentity calls at 533 and 545) re-stats each pathname at conversion time and records whatever occupies that spelling NOW instead of the identities captured at planning; so the journal-persisted proof can differ from the in-memory guard and Recover verifies the wrong identity (reviewer regression TestReviewRecoveryUsesOriginalSnapshotIdentity through real Engine.Prepare + fresh Engine.Recover).

Required: Durable must serialise the identities already held by the Snapshot (the values captured at planning), never re-inspect the filesystem; if a stored identity cannot be serialised, fail closed at planning (no journal, no publication). Add the reviewer's regression as a committed test (swap the parent between planning and Durable conversion; Recover must refuse), plus a control where nothing changed. Everything else in rev7 stays.

Narrow tests only (-p 1, -run filters, tool calls under 2 minutes), mutant (make Durable re-stat → the regression must fail), evidence with exit codes, checklist, handoff rev8.
