# CHANGES_REQUESTED — TASK-260910-hwxr26 revision 7

Candidate tree: 22fe0497b260576da5b6d64d54fd971057009d91. Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90. All 14 candidate paths match working bytes; both vendored schemas match the accepted curator-spec checkout byte-for-byte. No repository code modified.

## F1 — P2: duplicate JSON keys bypass closed raw shapes

Locations: internal/audit/sourceaudit.go:282-285 and :444-447 (raw decoders); :463-466 (second typed decode).

Both entry points decode into maps without protocoljson.Validate. Go drops earlier duplicate members before raw shape validation. Consequently a document prefixed with `"package":{"kind":"local-snapshot","commit":null},` before its ordinary package member passes: the forbidden foreign-arm member disappears from the map, and the later typed decode also admits the null. Duplicate schema_version (999 followed by 1) and report revoked (true followed by false) likewise pass. Core section 1 explicitly requires duplicate-key rejection; skillfile-sources introductory paragraph preserves core requirements. This is malformed evidence admission, not a claim that a real live revocation can be bypassed.

Independent production probe TestReviewDuplicateAuditJSON: 4/4 malformed variants admitted on BOTH install.Project dry-run and mutating paths (0/8 required refusals). Dry-run returns Status:ok; mutating proceeds past this gate and only fails at the known sibling schema-2 marker boundary. The report cases recompute evidence_sha256 as an attacker can. Command exit 1, 3.489 seconds. The valid initial issuance succeeds through the audit gate.

Reproduction: append the attached probe to the candidate draftaudit_test.go via a Go overlay (the attachment contains the full replacement); map the absolute internal/install/draftaudit_test.go path to that file in {"Replace":{...}}. Run:
`go test -overlay /tmp/hwxr26-review7-overlay.json -p 1 ./internal/install -run '^TestReviewDuplicateAuditJSON$' -count=1 -v -timeout=70s`
The attached log records all eight admissions.

Required rework: call the existing shared protocoljson.Validate on original object and report bytes before any lossy map or struct decoding, preserving source_audit diagnostics. Add production duplicate-key regressions at top level and nested package/commit/report members, with intact-record controls and no-overwrite assertions; preserve the shape/renewal checks. Reuse the existing validator for the full transport requirement instead of implementing another decoder. Kill a narrowing mutant that leaves duplicate rejection active for objects but omits it for reports (or nested members).

## Independent verification

All commands used zsh; exit codes observed directly after completion, no background jobs left running.
- Narrow audit tests: `go test -p 1 ./internal/audit -run 'Test(Source|ParseSource|ValidateSource|CheckSource|PolicyDigest)' -count=1 -timeout=60s`: exit 0, 0.504s.
- Production shape/renewal: `go test -p 1 ./internal/install -run '^TestDraftAudit(ClosedPackageShapeRefuses|MalformedShapeRefuses|RenewalNeverHidesRefusal|RenewalMarkersNeverRenewable)$' -count=1 -timeout=100s`: exit 0, 63.327s. Attached review-tests log.
- Legacy/local-registry/pin/revocation/label controls: `go test -p 1 ./internal/install ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'TestLegacyInstallUntouchedWhenDraftOff|TestCheckLocalPackage|TestEffectiveLabels|TestDraftAuditStrictLocalRequiresAttestation|TestDraftAuditPinAdmits|TestDraftAuditRevocationBlocks' -count=1 -timeout=65s`: exit 0 for all four packages. Attached controls log.
- git diff --check: exit 0.
- Hosted evidence accepted from attached rev7-validation.log: run 35289443484 success; gate commit 52682c6562c0fce93f14e70c2e41da873be64601 resolves locally to EXACT candidate tree 22fe0497b260576da5b6d64d54fd971057009d91. Full suite not rerun locally. Hosted ubuntu/macOS/Windows jobs reported success; rose-air and candidate suite skipped, not claimed passing.

Bounds: the new committed production shape table drives only local-snapshot (1/3 package arms); network/configured arms are helper-level coverage. Producer arm-gate mutant evidence was inspected, not independently replayed; it removes the full raw package guard rather than narrowing one arm. No claim of exhaustive mutation coverage. F1 independently demonstrates the remaining admission gap despite green existing tests.

Run goal queried immediately before verdict: not goal-bound. Verdict branch: changes_requested; route to to-dev. No accept_cr and no commit_ack. Finding also recorded in task notes; campaign rules forbid LOGBOOK.md edits.
