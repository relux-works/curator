# TASK-260922-1t2w1q rework 2 (orchestrator, binding) — F-C2 recovery path

Verdict rev3: CHANGES_REQUESTED (TASK-260922-1t2w1q_review-verdict-rev3.md; reviewer rows
`TestReviewerRecoveryMarkerDrift`, `TestReviewerRecoveryPreservesRegularTemp`). The straight-line
path is accepted (plan required, print-before-mutation, temp-link+rename, journal intent, marker
digest in the hash, syscall-failure/drift/no-copy/no-silent-repair rows). Two P1 defects in RECOVERY:

1. `migrate.go:914` runs recovery before the plan-hash comparison (:930), and recovery (:1510-1515)
   unconditionally republishes the journal's Prior marker bytes without checking that the current
   bytes equal the recorded prior or the intended new state — a marker edited between an
   interrupted apply and the retry is overwritten and the stale plan is accepted. Rule: validate
   the ENTIRE recovery inventory (marker bytes vs journal prior/intended; every link's current
   target vs the journal's recorded prior/intended targets) BEFORE any recovery write; refuse
   unknown states naming the exact operator action, preserving them and the journal; reconcile
   only states proven to belong to this transaction; then re-check the plan hash. Keep the
   ordinary interrupted-apply success row; add refusal rows for a marker edit and for an
   unexpected symlink target during recovery.
2. `migrate.go:1259-1265` globs `.migrate-*.tmp` and removes every match for every journal
   operation directory (also for operations not yet executed) — a foreign regular file
   (operator backup) is deleted. Rule: the journal records the exact temporary-link path it
   created; cleanup removes ONLY that path and only if it is a symlink with the recorded target;
   a regular file, directory or foreign symlink at any candidate path is preserved and reported
   (refusal or finding, never deletion). Rows: preservation of a foreign regular file during
   recovery; positive owned-temp cleanup; mutants (glob cleanup restored; recovery before
   validation) killed.
Commit the reviewer's two rows. Continue from the revision-3 tree (no checkout/clean/stash);
append "Revision 4" to results.md; republish only on a green gate.
