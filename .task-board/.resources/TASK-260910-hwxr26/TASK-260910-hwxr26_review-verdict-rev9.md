# TASK-260910-hwxr26 review verdict — revision 9

CHANGES_REQUESTED — one P2 finding. Route to `to-dev`.

Reviewed base 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90, candidate tree 51ee352a02405ee68360ea2fb03a2754798e9d71. All 14 changed worktree files byte-match candidate blobs. No repository code changed; independent tests used a Go overlay.

## F1 — P2: unreadable existing binding is treated as first issuance

Location: internal/audit/sourceaudit.go:888–894 and :1068–1071 (write effects :958–961).

`sourceAuditObjectExists` collapses every os.Stat error to false. Stat follows symlinks: an existing dangling binding or symlink loop therefore looks absent. LoadSourceAudit classifies both missing and unreadable bindings as source_audit_unavailable, and CheckSourceAudit then calls establishSourceAudit on this false absence. This bypasses validation of the existing report and can overwrite it. It violates the accepted §4 report refusal and rework-1 requirement to distinguish first issuance from a broken existing record.

Concrete reproduction at install.Project (attached TestReviewRev9):
1. Establish a valid local binding through mutating Project.
2. Replace its object path with a dangling symlink to a nonexistent file; leave the report in place.
3. Dry-run refuses, but mutating Project passes the source-audit gate and creates the symlink target. It fails only downstream at the known sibling marker-schema boundary.
4. Alternatively replace the object with a self-referencing symlink and replace the report with `{}`. Mutating Project overwrites the malformed report with newly issued evidence, then fails writing the object with ELOOP. This is a read failure, not legitimate absence.

Required correction: preserve typed absence/read-failure/present state from loading/probing. Establish only after positively proving no binding entry exists; any unreadable or existing broken entry refuses before either record is written. Do not collapse stat errors into absence; use non-following entry inspection where appropriate and propagate all non-absence errors. Add production regression rows for dangling and looping binding links with unchanged report/target assertions, plus genuine first-issuance controls. Narrowing mutant should restore error-to-absence collapse and be killed. This is ordinary implementation rework, not a human decision or external blocker.

## Independent evidence

Commands ran via zsh, direct log redirection (no pipeline masking); exit codes observed from completed processes:
- `go test -p 1 ./internal/install -run '^TestDraftAudit(DuplicateKeysRefuse|MalformedShapeRefuses|ClosedPackageShapeRefuses)$' -count=1 -timeout=90s` — exit 0, 19.439s.
- `go test -p 1 ./internal/install -run '^TestDraftAudit(RenewalNeverHidesRefusal|RenewalMarkersNeverRenewable|PolicyRenewalBrokenEvidence|BrokenReportRefuses|PinAdmits|RevocationBlocks|StrictLocalRequiresAttestation)$' -count=1 -timeout=100s` — exit 0, 22.007s.
- `go test -overlay /tmp/hwxr26-overlay.json -p 1 ./internal/install -run '^TestReviewRev9$' -count=1 -v -timeout=90s` — exit 1, 1.872s. Four independent shape variants × two production paths = 8/8 refusals with unchanged records (escaped duplicate object/report keys, foreign-arm null, null revoked). Two storage failure variants reproduce F1 (2/2). Valid initial issuance succeeds at the audit gate. Overlay maps virtual internal/install/review_rev9_test.go to attached review-probe-rev9.go; no candidate source changes needed.
- `git diff --check` — exit 0.
- `task-board spawn goal "$TASK_BOARD_RUN_ID"` — no goal bound; checked before verdict. No directives.

Hosted evidence reused, not locally replayed: attached revision-9 validation log and live `gh run view 35296850550` both report success. Gate commit 514777d054f94e4f38097e3e14bbc6033967582d resolves via git rev-parse to exactly candidate tree 51ee352a02405ee68360ea2fb03a2754798e9d71. macOS/Linux/Windows tests, macOS/Linux race, lint and interop passed; rose-air skipped, unverified. Full legacy/lint/platform suite accepted from that exact hosted run; no local full suite claim. Producer's rev8 narrowing-mutant evidence was inspected, not independently rerun because F1 already requires rework.

## Review conclusion / logbook note

Rev7 duplicate-key fix is effective on original object and report bytes before lossy decode, including escaped equivalent keys. Closed shapes, strict scalar presence, two-phase renewal and typed renewable outcomes remain effective in targeted checks. Local network-attestation refusal, pins and revocation checks passed. Remaining blind spot is filesystem availability classification: a bool existence probe erases the distinction between an absent record and failure to inspect an existing record. Record this finding in the task notes; repository LOGBOOK.md is intentionally untouched under campaign restrictions. No acceptance issued.
