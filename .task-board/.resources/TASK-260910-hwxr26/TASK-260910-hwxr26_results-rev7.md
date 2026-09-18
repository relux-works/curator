# TASK-260910-hwxr26 results — rework rev7 (verdict rev6 F1)

Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90 (Story branch task-board/story/STORY-260910-20sx61).
Scope: internal/audit, registry, artifactpolicy, scriptpolicy. No spec edits. Uncommitted working tree (handoff snapshot).
Changed by this rework: internal/audit/sourceaudit.go, internal/audit/sourceaudit_test.go, internal/install/draftaudit_test.go, internal/audit/testdata/draft-sources-v1/{source-audit-v1.schema.json,source-types-v1.schema.json} (vendored verbatim).
Unrelated dirty work (install.go/registry.go modifications, labels/local/draftaudit.go files) predates this rework and is untouched.

Spec: curator-spec main 4a2fa3ec428b5021e9e8a80f29fa7fc5a88e4990; schemas sha256 source-audit-v1 265e2ed6a010f9df95a3c95c89d3528092b8c1bd612659c6caa1ef5c01256b26, source-types-v1 fbb389bbfe5e2d05293e669e196d150f4fbaef1601ae239f8e7054c200621c15. Vendored copies byte-identical (cmp AUDIT-OK/TYPES-OK).

## Fix (verdict rev6 F1: closed package arms accept forbidden zero-valued members)

Root cause: SourcePackage combined every union arm into one Go struct; DisallowUnknownFields recognized Git fields in a local package, and validate() checked decoded nonzero values. commit:null, repository/source/directory null/empty disappeared through zero-value decoding and object() projection.

Fix (sourceaudit.go): schema-driven closed-shape validation on RAW bytes BEFORE Go decoding, for both object and evidence report:
- sourcePackageArms table (local-snapshot: kind+snapshot; network-git: kind+repository+commit+directory; configured-git: kind+source+commit+directory), each additionalProperties:false, with declared JSON kinds; TestSourcePackageArmsMatchAcceptedSchema reads the vendored schemas and asserts arms/required/closedness match.
- validateSourcePackageRaw: kind-selected arm, presence alone refuses (foreign null/empty/valid), required missing/null/wrong-type refuses, commit {object_format,hex} closed non-null strings.
- validateSourceAuditObjectRaw: top-level 7 required non-null typed, no extras.
- validateEvidenceReportRaw: top-level 11 required non-null typed, no extras, plus package union; trailing tokens refused.
- ParseSourceAudit: raw map decode + EOF check + raw top-level + raw package, then struct decode + value patterns.
- ValidateSourceAudit: evidence digest, then validateEvidenceReportRaw, then struct decode + completeness/trusted bindings. Typed renewal ordering (phase-1 non-renewable before phase-2 stale/policy) and all accepted controls preserved byte-identical.

## Tests (committed, production entry named)

- internal/audit/sourceaudit_test.go: TestSourcePackageArmsMatchAcceptedSchema (vendored schema vs table); TestSourceAuditClosedPackageShape (all 3 arms x foreign null/empty/valid, required missing/null/wrongtype, unknown, trailing, commit closedness for object via ParseSourceAudit and report via ValidateSourceAudit with recomputed digest; valid controls per arm admit).
- internal/install/draftaudit_test.go: TestDraftAuditClosedPackageShapeRefuses via install.Project (production call site). One fixture established once and restored per row (single Git setup). Local arm (the only arm install produces; git arms covered arm-for-arm at audit level): foreign commit/repository/source/directory x null/empty/valid, required kind/snapshot missing/null/wrongtype, unknown package/top members, commit extra/null-hex, trailing — for object and report (report with recomputed evidence_sha256). Each row refuses on mutating and read-only paths with source_audit and leaves both records unoverwritten; valid controls pass both paths.

## Evidence (shell bash with set -o pipefail, real exit codes, narrow packages only)

1. go vet ./internal/audit — exit 0.
2. go vet ./internal/install — exit 0. gofmt -l on 3 touched files — clean. git diff --check — exit 0.
3. go test -p 1 ./internal/audit -run 'TestSourceAuditClosedPackageShape|TestSourcePackageArmsMatchAcceptedSchema|TestParseSourceAudit' -count=1 -timeout=90s — exit 0 (ok 0.519s).
4. go test -p 1 ./internal/install -run '^TestDraftAuditClosedPackageShapeRefuses$' -count=1 -timeout=110s — exit 0 (ok 11.448s, 46/46 PASS).
5. Reviewer overlay: go test -overlay /tmp/hwxr26-overlay-rev6.json -p 1 ./internal/install -run '^TestReviewClosedPackageShape$' -count=1 -timeout=90s — exit 0, 14/14 PASS (ok 7.747s; was 0/28 exit 1 at rev6).
6. go test -p 1 ./internal/audit -count=1 -timeout=100s (full audit) — exit 0 (ok 1.648s).
7. go test -p 1 ./internal/install -run '^TestDraftAuditMalformedShapeRefuses$' -count=1 -timeout=110s — exit 0 (ok 8.736s; rev5 preserved).
8. go test -p 1 ./internal/install -run '^TestDraftAudit(RenewalNeverHidesRefusal|RenewalMarkersNeverRenewable)$' -count=1 -timeout=115s — exit 0 (ok 18.505s; rework-3/4 preserved).
9. go test -p 1 ./internal/install -run '^TestDraftAudit(PolicyRenewalBrokenEvidence|BrokenReportRefuses|FirstIssuanceVsBrokenRecord|BindingLifecycle)$' -count=1 -timeout=115s — exit 0 (ok 6.403s; rev1/2 preserved).
10. go test -p 1 ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'Test(LocalPackage|EffectiveLabels)' -count=1 -timeout=100s — exit 0 (registry no tests matched; artifactpolicy ok; scriptpolicy ok).
11. Full local suite NOT run (host stalls; per wave brief narrow packages only).

## Implementation mutant (arm check removed, then restored)

Mutant: validateSourcePackageRaw returns nil immediately (restores zero-value cross-arm admission); fixed copy kept at /tmp/hwxr26-sourceaudit.fixed.go, restored with cp, grep shows no MUTANT, go vet exit 0.
- go test -p 1 ./internal/audit -run 'TestSourceAuditClosedPackageShape/object-local-snapshot-foreign-commit-null|...report...' — exit 1, 2/2 FAIL (admitted: <nil>).
- go test -p 1 ./internal/install -run 'TestDraftAuditClosedPackageShapeRefuses/object-foreign-commit-null|...report...' — exit 1, 2/2 FAIL, each shows only 'review: install marker is invalid for schema 2' (gate bypassed to sibling marker boundary) — the exact rev6 reviewer signature.
The new tables kill the ordering/zero-value mutant; the fix binds the shape, not input variation.

## Reviewer probe rev6 disposition

TASK-260910-hwxr26_review-probe-rev6.go TestReviewClosedPackageShape (14/14: object/report x commit null, repository/source/directory null/empty) is a strict subset of the committed production table (which adds valid variants, required/unknown/trailing/commit-closedness). Overlay replay above refuses all 14 with source_audit on both paths without overwrite. No other probe claims.

## Checklist

- [x] Scoped production behavior with traceability to accepted draft contracts (skillfile-sources §4; source-audit-v1 + source-types-v1 closed arms; distinct from registry attestation)
- [x] Task-specific positive, negative and legacy regression checks run; exact revision and evidence recorded above
- [x] Implementation matches AC (identity/context/policy/evidence/time/decision bindings; local network-attestation refusal intact; pins/revocation/assurance before cache/compiler; absent/malformed/stale/wrong evidence refused without weakening currentness)
- [x] Solution fits project architecture (audit owns raw shape/validation; install calls CheckSourceAudit; no new interfaces; no json-schema dep)
- [x] Tests green (items 1–10)
- [x] Relevant tests written for new/changed behavior and passing (audit arm table + schema conformance + production 46 rows + controls)
- [x] Lint clean (gofmt, go vet, git diff --check)
- [x] Build/validation commands run after changes; build not broken
- [x] Outcome artifact attached (this file, TASK-260910-hwxr26_results-rev7.md)
- [x] No live credential export or runtime-home modification; unrelated dirty work untouched
