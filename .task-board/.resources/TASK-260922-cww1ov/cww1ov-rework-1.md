# TASK-260922-cww1ov rework 1 → revision 2 (orchestrator brief, binding)

Verdict: CHANGES_REQUESTED on revision 1 (`TASK-260922-cww1ov_review-verdict-rev1.md`, one P2
coverage defect + evidence bookkeeping). Everything else in revision 1 is ACCEPTED as reviewed —
do not re-open, re-design or re-run the accepted rows beyond what R1 requires.

## R1 (the only substantive fix) — the no-copy row cannot see manager state
`internal/envprofile/migrate_test.go:114`: `credentialHolders` covers only environments, profiles
and native roots; the snapshot helper at :55 excludes state as well. So `TestMigrateNoSecretCopies`
(:278) cannot prove the no-secret-copy property for the journal/state area migration itself writes
— the very area F-C2's reviews had already flagged as a bound. The reviewer's attached
`TASK-260922-cww1ov_review-mutants.py`, mutation `review-state-secret-copy`, plants a copy of the
credential bytes at `<dir of migrationJournalPath(home)>/credential-backup` (0600) inside the relink
arm of `executeMigrationOps` and the row still exits 0.

Fix the TEST, not production:
- scan the ENTIRE temporary manager home for credential content — state, journal, backup and temp
  paths included — while keeping the native-root scans; permit only legitimate non-secret
  lock/journal metadata (name the allowance explicitly and keep it narrow, e.g. by asserting that
  no file anywhere under the home contains the credential bytes, rather than allowlisting paths);
- a deleted journal cannot be inspected after success: also scan at the existing interrupted-apply
  hook (the crash-injection point F-C2 already uses), so transient journal content is covered;
- add the narrowing mutant that copies ONLY into manager state and require this named
  production-entry row to kill it (executed, in the mutant table);
- keep the scope and residual bounds explicit in results.md.

## Bookkeeping (secondary, but required)
- results.md claims lane shapes 39/4/3; the actual ledger additions are 37 all-platform rows with
  Windows host-capability tolerance, 5 all-platform/no-tolerance rows and 4 Unix-only rows (46
  total). Recount from `.github/ci/platform-cases.tsv` yourself and state the true numbers.
- `.github/ci/platform-cases.tsv:475` says three ENOTDIR rows; there are four. Fix the comment.
- The attached results.md ends in a literal truncation marker — attach a COMPLETE results resource
  (write the file, then `task-board resource add`), not a truncated one.

## Rules
No product change unless R1's scan exposes a real defect — then fix it minimally and name it. No
accepted test weakened or deleted; F-C1/F-C2 reviewer rows stay as committed. Re-run the focused
tests and your mutants with real exit codes (state the shell, `-count=1`, bounded timeouts), publish
only on a GREEN hosted gate for the revised candidate, then
`task-board handoff TASK-260922-cww1ov --role developer`. This is the FINAL leaf of
STORY-260922-1cenbr: on acceptance the orchestrator integrates the Story.
