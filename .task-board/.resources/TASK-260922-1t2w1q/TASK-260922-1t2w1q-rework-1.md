# TASK-260922-1t2w1q rework 1 (orchestrator, binding) — F-C2

Verdict rev1: CHANGES_REQUESTED (TASK-260922-1t2w1q_review-verdict-rev1.md; reviewer probes
attached: `TestReviewerMigrationCLI/{missing-plan,print-before-write}`,
`TestReviewerFailedRelinkRollback`, `TestReviewerMarkerDrift`). Continue from the revision-1 tree
(no checkout/clean/stash); fix all four, commit the reviewer probes as rows:

P1-a apply without a plan: `migrate.go:790` checks drift only when `Expect` is set;
`envmigrate.go:57` accepts `--apply` without `--expect`. Rule: every apply entry REQUIRES a prior
complete plan identity (hash) and refuses before any mutation when it is missing or stale; replace
the test that blesses no-expect apply; update repair hints/docs.
P1-b plan printed after mutation: `envmigrate.go:58-60` applies then prints. Print the locked,
revalidated plan BEFORE the first mutation; if the output cannot be delivered, fail before
mutating; add a temporal CLI assertion (writer hook observing the link state on first write), not
final-string containment.
P1-c relink not atomic / not journaled: `migrate.go:859-872` removes the old link then creates the
new one; a `Symlink` failure leaves the old link gone and outside `applied`; relink-only plans
never publish a journal entry; marker bytes are journaled only after link changes; a failed marker
publication leaves link changes standing. Implement a durable migration journal written BEFORE
each mutation (intent + rollback data), temp-link + atomic rename replacement (`symlink` to a temp
name then `rename` over the old link — never a remove-then-create window), recovery of a partial
apply from the journal on the next inspect/apply, rollback that covers a syscall failure mid-op and
a failed marker publication (revert links), and rows that exercise the syscall boundary (injected
`Symlink`/`Rename` failure) and process interruption (journal present, links half-done → recovery
restores or completes deterministically).
P2 plan hash omits the marker identity: include a digest of the inventoried marker (raw bytes
digest, no credential reads) in the plan hash so a marker edit between plan and apply refuses.
results.md "Revision 2": the journal format, the atomic replacement sequence, recovery rules,
rows + mutants (drop the plan requirement; print after; remove-then-create; hash without marker).
Republish only on a green gate.
