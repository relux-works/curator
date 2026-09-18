# TASK-260910-hwxr26 review verdict — revision 1

Verdict: CHANGES_REQUESTED. Route: to-dev.
CR: CR-TASK-260910-hwxr26-1 revision 1.
Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90.
Candidate tree: 96d14b020bc9ee51d6aeb637df3daf767c63efd2.

## Findings

F1 (P2): Existing missing or mismatching evidence is silently replaced on installation.
Locations: internal/audit/sourceaudit.go:568–584 (LoadSourceAudit classification at :496–517).
The accepted skillfile-sources §4 explicitly says a missing, unreadable or mismatching report fails. CheckSourceAudit instead treats every source_audit_unavailable error as permission to establish on persist=true, and every validation failure as permission to overwrite the pair. This conflates first-time issuance with a broken existing record. Reproduction through install.Project: establish a clean local package binding; delete its .report.json, or replace its report with {}; call Project with DraftSourcesV1=true and DryRun=false. Both pass source audit and reach the known downstream marker-schema error. Neither emits source_audit refusal. Fix by distinguishing initial issuance from incomplete/corrupt existing evidence, refusing broken reports before downstream work; keep any authorized renewal separately classified rather than catching all validation failures. Add production-path negative tests for missing, unreadable, mismatching and malformed reports on both paths.

F2 (P2): An incomplete, wrong-skill report with a caller-recomputed digest passes read-only validation.
Locations: internal/audit/sourceaudit.go:368–381 and :575–578.
Only nonempty skill/content/decision and package/content/decision equality are validated; required report completeness and trusted findings/pin/revocation/policy-label bindings are not checked. The expected evidence is the same untrusted stored bytes whose hash the object supplies. Reproduction through install.Project: establish binding; remove findings, pinned, revoked, script_policy, assurance_policy and schema_version from its report, change skill to different-skill, recompute evidence_sha256 in the object, then dry-run. Result is Status:ok. This certifies incomplete evidence as valid, contrary to the complete existing audit report contract. A fresh live decision still runs, so this reproduction does NOT claim hostile execution bypass. Validate report completeness and bind its identity and security-relevant contents to independently recomputed trusted state, rather than merely checking the supplied pair against itself. Add production-entry forged-evidence tests.

## Independent evidence

Repository code unchanged. All 12 changed files were byte-compared to git show of the exact candidate and matched. Working-tree status remained identical after review; test additions exist only in an ignored temporary Go overlay.

Shell: zsh; no pipelines in test invocations.

1. `go test -p 1 ./internal/audit ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy ./internal/install -run 'Test(ParseSourceAudit|ValidateSourceAudit|SourceAudit|PolicyDigest|CheckSourceAudit|DraftAudit|EffectiveLabels|LocalPackage|NetworkAttestation)' -count=1 -timeout=75s` — exit 0. Registry had no matching tests in this invocation, corrected below.
2. `go test -p 1 ./internal/registry ./internal/install ./internal/interop -run 'Test(CheckLocal|ResolveWithoutIdentity|LegacyInstallUntouched|GoldenRegistryObjects)' -count=1 -timeout=75s` — exit 0, all three packages ran matching tests.
3. `go test -overlay .temp/review-hwxr26/overlay.json -p 1 ./internal/install -run '^TestReviewSourceEvidenceRefusals$' -count=1 -timeout=75s` — exit 1: all 3 refusal expectations failed. Observed refusal coverage for these attack rows: 0/3. Attached log and complete overlay test file retain the exact probes. To replay, materialize the test artifact and map the absolute internal/install/draftaudit_test.go path to it via Go's JSON Replace overlay, then run the command above. The artifact includes the original tests plus TestReviewSourceEvidenceRefusals.
4. `git diff --check` — exit 0.

Hosted gate accepted as attached evidence, not rerun: TASK-260910-hwxr26_change-request_rev1-validation.log records GitHub run 35255725745 success, exit 0; ubuntu/macos/windows test jobs, race, lint and interop passed. `git rev-parse '36e1dad42371eb66e1e5c06829855017e316e41d^{tree}'` independently resolves to the exact candidate tree above. Candidate suite and rose-air were skipped; no claim for those lanes.

Producer evidence reports a delete-only mutant; no narrowing-mutant result is attached. No reviewer mutation campaign was needed to establish rejection: unchanged production code already fails these three adversarial negative expectations. Full acceptance-clause production coverage, complete legacy goldens and assurance narrowing coverage are not established by this review. Existing positive/negative targeted tests pass, but do not cover the broken-report mutation path or self-minted incomplete evidence above.

Goal queried before verdict: run is not goal-bound. No directives recorded. No commit, integration, runtime-home changes, or source edits. Findings recorded in board notes and this artifact; no logbook CLI/tool is available, and campaign rules prohibit LOGBOOK.md edits.

Next producer: fix both findings, retain the repro cases as production-entry regression tests, add the requested missing refusal coverage and narrowing evidence, and hand off a new revision for independent review.
