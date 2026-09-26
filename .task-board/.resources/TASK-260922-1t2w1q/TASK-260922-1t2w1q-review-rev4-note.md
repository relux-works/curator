# Review note for TASK-260922-1t2w1q revision 4 (orchestrator, binding) — F-C2 recovery path

Revision 4 = rework 2 for your revision-3 verdict (TASK-260922-1t2w1q_review-verdict-rev3.md; brief
TASK-260922-1t2w1q-rework-2.md): (1) recovery validates the entire inventory (marker bytes vs
journal prior/intended; every link's current target vs recorded prior/intended) BEFORE any write,
refuses unknown states naming the operator action while preserving them and the journal,
reconciles only states proven to belong to the transaction, then re-checks the plan hash;
(2) cleanup removes only the journal-recorded temporary link path, and only if it is a symlink with
the recorded target — foreign regular files/dirs/symlinks preserved and reported; your two rows
committed; mutants (glob cleanup; recovery before validation) killed. Gate green: run 35709733177 —
verify the gate commit resolves to the exact revision-4 tree and that rev3→rev4 is this scope.
Rerun `TestReviewerRecoveryMarkerDrift` and `TestReviewerRecoveryPreservesRegularTemp` against rev4
(must pass), the ordinary interrupted-apply row, and probe one more shape of your choice (e.g. an
unexpected symlink target during recovery). Everything accepted at rev1/rev3 stays. Record exactly
one verdict: accept_cr(TASK-260922-1t2w1q, revision=4, evidence=<your outcome resource>) on ACCEPT,
or changes_requested with file:line and reproduction. Do not write into the control root's LOGBOOK.md.
