# TASK-260910-hwxr26 review verdict revision 3

Verdict: CHANGES_REQUESTED. Route to to-dev.
Candidate: d8cc65ff88366b880b84ca357f4b99c5bb9a21ea; base 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90. All 12 changed paths byte-match the candidate. No production code modified.

## F1 — P2: stale renewal still hides a mismatching decision

internal/audit/sourceaudit.go:471-475 returns the renewable stale error before comparing the stored decision with the live decision. CheckSourceAudit then establishes a replacement object/report. This violates the decision binding and the rev2 rework requirement to validate all non-renewable bindings before choosing renewal.

Concrete production reproduction: issue a valid clean local binding via install.Project; change report decision and object decision from allow to warn; recompute evidence_sha256; set created_at to 2020-01-01T00:00:00Z; change FailOn from high to critical; call install.Project again. The gate accepts and overwrites BOTH the forged object and report. The later unrelated marker-stage failure is not a source-audit refusal.

Independent TestReviewStaleDecisionRenewal (Go overlay derived from the producer production test) fails with exit 1, reporting no source_audit refusal and both overwrite assertions. Coverage: 0/1 additional combined stale/decision refusal rows enforced. Existing policy-renewal rows pass; they do not combine stale time with decision mismatch.

Required fix: complete all non-renewable validation, including the live decision binding, before returning either renewable policy OR stale-time outcomes. Add committed install.Project regression for stale + forged decision (with and without policy drift), assert no overwrite, retain valid stale and valid policy renewal controls, and kill an implementation ordering mutant. Avoid another single-condition patch: enumerate both renewal causes against the non-renewable bindings.

## Validation and bounds

Independent zsh command: go test -p 1 ./internal/install ./internal/audit ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'Test(DraftAudit|SourceAudit|LocalPackage|NetworkAttestation|EffectiveLabels)' -count=1 — exit 0; registry had no matching tests. Install package included all five new policy-renewal rows.

Independent adversarial command: set -o pipefail; go test -p 1 -overlay .review-hwxr26-rev3/overlay.json ./internal/install -run '^TestReviewStaleDecisionRenewal$' -count=1 — exit 1 (expected refusal assertions fail). Attached overlay replaces internal/install/draftaudit_test.go; reconstruct a Go overlay JSON with absolute original/replacement paths to replay.

Read producer results-rev3 and rev3-validation.log. Hosted run 35268046836 reports exit 0; git rev-parse a66285966fdbc4edabdef0a5d1cba330f6fa6850^{tree} independently equals the reviewed candidate. Hosted Linux/macOS/Windows test lanes report success; rose-air and Candidate suite skipped. These are accepted attached logs, not independently rerun hosted checks. Producer reorder-back mutant evidence was read, not independently repeated. Full local suite deliberately not run. No cross-platform probe claim.

Run goal query: no active goal (not goal-bound). Finding also recorded in board notes; LOGBOOK.md is forbidden by campaign instructions.

Additional independent legacy/validation command: go test -p 1 ./internal/audit ./internal/registry ./internal/install ./internal/interop -run 'Test(ValidateSourceAudit|CheckSourceAudit|CheckLocal|ResolveWithoutIdentity|LegacyInstallUntouched|GoldenRegistryObjects)' -count=1 -timeout=60s — exit 0, all four packages passed.
