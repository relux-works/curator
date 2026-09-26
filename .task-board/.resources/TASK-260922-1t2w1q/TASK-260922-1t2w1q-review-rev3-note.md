# Review note for TASK-260922-1t2w1q revision 3 (orchestrator, binding) — F-C2

Revision 3 = revision 2 unchanged (rev2's gate failed only on an unrelated Windows managerlock
timing flake); revision 2 = rework 1 for your revision-1 verdict (TASK-260922-1t2w1q_review-verdict-rev1.md;
brief TASK-260922-1t2w1q-rework-1.md): P1-a every apply requires a prior complete plan identity and
refuses before any mutation when missing/stale (the no-expect apply row replaced); P1-b the locked,
revalidated plan is printed BEFORE the first mutation with a temporal CLI assertion; P1-c durable
migration journal written before each mutation (intent + rollback data), temp-link + atomic rename
replacement (no remove-then-create window), recovery of a partial apply on the next inspect/apply,
rollback covering a syscall failure mid-op and a failed marker publication, rows exercising
injected Symlink/Rename failure and process interruption; P2 the plan hash includes a digest of the
inventoried marker (no credential reads). Gate green: run 35705009860 — verify the gate commit
resolves to the exact revision-3 tree and that rev1→rev3 is this scope. Rerun your four probes
(`TestReviewerMigrationCLI/{missing-plan,print-before-write}`, `TestReviewerFailedRelinkRollback`,
`TestReviewerMarkerDrift`) against rev3 — all must pass; inspect the journal format and the recovery
path (journal present + half-done links → deterministic restore/complete); confirm no secret bytes
are read/copied; mutants (drop plan requirement; print after; remove-then-create; hash without
marker; skip journal) killed. Everything accepted at rev1 stays. Record exactly one verdict:
accept_cr(TASK-260922-1t2w1q, revision=3, evidence=<your outcome resource>) on ACCEPT, or
changes_requested with file:line and reproduction. Do not write into the control root's LOGBOOK.md.
