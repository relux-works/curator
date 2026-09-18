# TASK-260910-hwxr26 review verdict revision 5

CHANGES_REQUESTED — route to-dev.
Candidate 5453d0cb345f94b32e12b36284afc5ee44651088, base 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90. Independently compared all 12 changed files byte-for-byte with candidate blobs: 12/12 match. No production code modified.

## Findings

F1 (P2) internal/audit/sourceaudit.go:203-213: ParseSourceAudit decodes only the first JSON value and never requires EOF. Appending `{}` to an otherwise valid persisted binding is admitted through install.Project. This is a malformed source-audit document, but the source gate passes. Require exactly one complete JSON document (including rejection of trailing garbage), with production regression tests.

F2 (P2) internal/audit/sourceaudit.go:428-441: presence checks followed by decoding into bool accept JSON null as false for pinned/revoked. On an unpinned, non-revoked package, replace either member with null and recompute evidence_sha256: install.Project admits the incomplete/malformed report. Validate required scalar types/non-null values before trusted comparisons; add production regressions and a narrowing mutant for null admission. Audit other required report members for the same zero-value coercion, rather than patching only these two names.

Both violate the task requirement to validate malformed evidence and the complete-report requirement in skillfile-sources section 4. The live audit still runs; these findings prove malformed record admission, not a hostile-content bypass.

## Independent evidence

Shell zsh. `go test -overlay .temp/review-hwxr26-r5/overlay.json -p 1 ./internal/install -run '^TestReviewMalformedRecordRev5$' -count=1 -timeout=90s` exited 1; 3/3 refusal assertions failed (trailing-object, null-pinned, null-revoked). Each result had only `review: install marker is invalid for schema 2`, proving progression past the source gate to the known sibling marker boundary. Runtime reported 12.451s. Probe attached as TASK-260910-hwxr26_review-probe-rev5.go; append to package tests using a Go overlay to reproduce without altering candidate files.

`go test -p 1 ./internal/install -run '^TestDraftAudit' -count=1 -timeout=90s` exited 1 from the 90-second timeout while executing TestDraftAuditRenewalMarkersNeverRenewable/package-stale-marker-stale-plus-drift in fixture Git setup. This is incomplete verification, not a passing suite or a product failure. No full local suite run; no independent mutant run or cross-platform execution claimed.

Read producer results-rev5 and attached rev5-validation.log. Hosted gate reports success, exit 0, run 35279584641, commit 0cf831161022340f6174bf7b63d251f13efdfd78. Independently resolved that commit tree to exact candidate 5453d0cb345f94b32e12b36284afc5ee44651088. Hosted results and producer mutant/legacy results are attached evidence, not independently rerun here; rose-air is skipped.

The rev4 diagnostic-injection fix uses typed errors correctly by inspection; the new findings are parser/shape gaps. Preserve typed renewal and positive valid-renewal controls during rework. No acceptance granted. Run is not goal-bound (spawn goal queried). No LOGBOOK.md edit per campaign prohibition; findings persisted in this task-scoped verdict and board notes.
