# CHANGES_REQUESTED — revision 6

Task: TASK-260910-hwxr26. CR-TASK-260910-hwxr26-6.
Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90.
Exact candidate tree: f5dc8df4c01b39e5d0a9ea6751cda666b01faa33.
All 12 changed working-tree files independently compared byte-for-byte with candidate blobs: equal. No repository code modified.

## F1 — P2: closed package arms accept forbidden zero-valued members

Location: internal/audit/sourceaudit.go:78-95 (also ParseSourceAudit:203 and report decoder:456-484).
SourcePackage combines every union arm into one Go struct. DisallowUnknownFields consequently recognizes Git fields in a local-snapshot package, and validate() only checks decoded nonzero values. commit:null, repository:null/empty-string, source:null/empty-string and directory:null/empty-string disappear through zero-value decoding and object() projection. Thus malformed identity/evidence passes trusted comparisons. The accepted source-types-v1.schema.json #/$defs/package local arm allows ONLY kind and snapshot, with additionalProperties:false; source-audit-v1 uses this package definition. These records must refuse even though the extra fields do not change the projected identity.

Concrete reproduction at install.Project:
1. Establish the ordinary strictLocalProject binding with draftAuditConfig.
2. Add package.commit = null to the stored binding. Invoke Project with DryRun true and false.
3. Alternatively add that member to the report and recompute the binding evidence_sha256.
Actual: dry-run Status:ok, no errors; mutating path reaches the known sibling marker failure (install marker is invalid for schema 2), with no source_audit refusal. Both paths admit the malformed source-audit record.
The attached overlay TestReviewClosedPackageShape expands this to all seven forbidden field/value combinations in both object and report: 14/14 malformed cases admitted on BOTH paths, 0/28 expected refusals. Exit 1. This is not a claim that the downstream install completes.

Required rework: enforce package wire shape by selected kind before discarding presence/type information, for both object and evidence decoding. Reject properties belonging to other arms even if null or empty; preserve required non-null/type constraints recursively. Derive permitted/required arms from the accepted schema and cover all three arms rather than adding a special-case check for commit:null. Add production regression rows and a narrowing mutant that restores zero-value cross-arm admission. Preserve valid arm controls and typed renewal ordering.

## Independent verification

Shell: zsh; each go invocation recorded its real exit code through rc=$? and exit $rc. All used -p 1 and -count=1.
- go test ./internal/install -run ^TestDraftAuditMalformedShapeRefuses$ -timeout=90s: exit 0, new 24-row shape table passes.
- go test ./internal/install -run ^TestDraftAudit(RenewalNeverHidesRefusal|RenewalMarkersNeverRenewable|StrictLocalRequiresAttestation|AdvisoryLocalPasses)$ -timeout=100s: exit 0.
- go test ./internal/audit ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -timeout=90s: overall exit 1. audit, registry, scriptpolicy pass; artifactpolicy times out in pre-existing TestDirectoryEntryLimitStopsTheLiveWalker while creating fixture files (containers_test.go:174). No claim of full local green.
- Independent overlay: go test -overlay /tmp/TASK-260910-hwxr26-review-overlay-rev6.json -p 1 ./internal/install -run ^TestReviewClosedPackageShape$ -count=1 -timeout=90s: exit 1, 14/14 failing subtests as described above.
- Focused follow-up: go test -p 1 ./internal/artifactpolicy ./internal/scriptpolicy -run ^TestEffectiveLabels -count=1 -timeout=60s: exit 0, both affected label tests pass.
- git diff --check: exit 0.
- Hosted evidence accepted from attached rev6 validation log: run 35284672850 successful, exit 0; git rev-parse 26f265005c83e6037576456e891b9d7b56e55867^{tree} independently equals exact candidate tree. Ubuntu/macOS/Windows checks green in that log; rose-air skipped. No independent hosted rerun.
- Producer lenient-decoding mutant evidence inspected, not independently rerun. No broad clause-completeness or platform coverage claimed. Legacy coverage relies on attached hosted gate and prior review evidence; not independently replayed in this round.

Reproduction artifact: TASK-260910-hwxr26_review-probe-rev6.go is an overlay replacement for internal/install/draftaudit_test.go. Map its downloaded absolute path to that source path using Go overlay Replace. Probe and logs attached before verdict.

Reviewer goal queried: this run is not goal-bound. Verdict branch: changes_requested; route to to-dev. No accept_cr, commit_ack, commits or code changes. Finding persisted on the board; campaign forbids LOGBOOK.md edits.
