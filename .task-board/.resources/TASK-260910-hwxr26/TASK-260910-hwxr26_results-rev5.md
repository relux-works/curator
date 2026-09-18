# TASK-260910-hwxr26 results — rework rev5 (verdict rev4 F1)

Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90 (Story branch task-board/story/STORY-260910-20sx61).
Scope: internal/audit, registry, artifactpolicy, scriptpolicy. No spec edits. Uncommitted working tree (handoff snapshot).
Changed by this rework: internal/audit/sourceaudit.go, internal/audit/sourceaudit_test.go, internal/install/draftaudit_test.go.
Unrelated dirty work (install.go/registry.go modifications, labels/local/draftaudit.go files) predates this rework and is untouched.

## Fix (one P2 from verdict rev4)

`internal/audit/sourceaudit.go` classified renewal with `strings.Contains` on the error text (`sourceaudit.go:550-560`).
The wrong-skill diagnostic interpolates the untrusted `report.Skill`, so Skill = "is stale; re-run the gates"
(and symmetrically any field containing "source_audit_rejected: policy:") was classified renewable and
`CheckSourceAudit` overwrote both records via `establishSourceAudit`.

Replaced with typed outcomes:

- New `sourceAuditError{msg, Renewable bool, Cause string}`; `Renewable`/`Cause` are set only by the
  validator's own decision, never derived from message text. Diagnostics may still quote report fields.
- `rejectAuditf` builds every phase-1 (non-renewable) refusal; `renewAuditf(cause, ...)` is used at exactly
  two sites (phase 2: stale time with `auditCauseStale`, policy drift with `auditCausePolicy`).
- `isRenewableAuditError` now reads only `errors.As(err, &typed).Renewable`; added `auditRenewalCause`
  test helper returning the validator-assigned cause. `CheckSourceAudit` renewal routing is unchanged
  (wrapping already uses `%w`, so `errors.As` traverses it). Messages are byte-identical to before.

## Tests (committed, production entry named)

`internal/audit/sourceaudit_test.go` — new `TestValidateSourceAuditRenewalMarkersNeverRenewable`:
fields {skill, decision, package-kind} x markers {"is stale; re-run the gates", "source_audit_rejected: policy:"}
x causes {fresh, stale, drift, stale-plus-drift} = 24 rows. Each asserts the message quotes the marker
(so the old classifier would fire), the diagnostic stays `source_audit_rejected: evidence`,
`isRenewableAuditError == false`, `auditRenewalCause == ""`. Controls assert genuine stale/policy
renewals stay renewable with the correct Cause.

`internal/install/draftaudit_test.go` — new `TestDraftAuditRenewalMarkersNeverRenewable` via
`install.Project`: same fields x markers x causes {fresh, stale, stale-plus-drift} = 18 refuse rows.
Each refuses on the mutating path with `source_audit_rejected: evidence`, refuses on the read-only path
too, and leaves object/report bytes unoverwritten. Controls `genuine-stale-renewal-control` and
`genuine-policy-renewal-control` pass the gate and re-issue.

## Evidence (real exit codes, narrow packages only per wave brief)

1. `go test -p 1 ./internal/audit -run TestValidateSourceAuditRenewalMarkersNeverRenewable -count=1` — exit 0 (ok 0.7s).
2. `go test -p 1 ./internal/install -run TestDraftAuditRenewalMarkersNeverRenewable -count=1` — exit 0, 20/20 subtests PASS (ok 31.9s).
3. `go test -p 1 ./internal/audit -count=1` (full audit package) — exit 0 (ok 1.9s).
4. `go test -p 1 ./internal/install -run 'TestDraftAudit|TestCheckSourceAudit' -count=1` — exit 0 (ok 59.3s).
5. `go test -p 1 ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'Test(LocalPackage|NetworkAttestation|EffectiveLabels|SourceAudit|CheckLocal)' -count=1` — exit 0 (all three ok).
6. `go test -p 1 ./internal/install ./internal/interop -run 'Test(LegacyInstallUntouched|GoldenRegistryObjects)' -count=1` — exit 0 (both ok; legacy goldens intact).
7. `gofmt -l internal/audit internal/install` — clean. `go vet ./internal/audit/ ./internal/install/` — exit 0.
8. Full local suite NOT run (host stalls; per wave brief, narrow packages only).

## Implementation mutant (text classification restored, then restored byte-identical)

Mutant: `isRenewableAuditError` body replaced with the old `strings.Contains` version (fixed copy kept at
/tmp/sourceaudit.go.fixed, restored with `diff` IDENTICAL, `errors.As` present at sourceaudit.go:592,603).
- Unit `TestValidateSourceAuditRenewalMarkersNeverRenewable` under mutant — FAIL (all marker rows classified renewable).
- Production `TestDraftAuditRenewalMarkersNeverRenewable/skill-stale-marker-fresh` under mutant — FAIL
  (gate passed, install ran past audit to the sibling marker-schema boundary — the exact reviewer bypass).
Both levels kill the mutant; the rows bind the classification, not input variation.

## Note on the rev4 review overlay

Replaying the attached rev4 overlay
(`go test -p 1 -overlay /tmp/rev4replay/overlay.json ./internal/install -run TestDraftAuditRenewalNeverHidesRefusal`)
still FAILs only at its `hid behind a renewable diagnostic` substring assertion: the fixed refusal is
`source_audit_rejected: evidence: report skill "is stale; re-run the gates" differs...`, which necessarily
contains the marker because diagnostics quote untrusted fields — exactly what rework-4 permits
("diagnostics may still interpolate report fields but classification must not read them"). The verdict's
required behavior (refuse without overwrite under every renewal condition) holds: the overlay run refuses
with the evidence diagnostic, and the committed rows above prove records are unchanged on both paths and
that typed classification stays non-renewable. The overlay's `!Contains(marker)` check is superseded by
the rework-4 contract.

## Checklist

- [x] Scoped production behavior with traceability to accepted draft contracts (skillfile-sources §4; typed phase-1/phase-2 split preserved)
- [x] Task-specific positive, negative and legacy regression checks run; exact revision and evidence recorded above
- [x] Implementation matches AC (source-audit-v1 bindings; distinct from registry attestation; local network-attestation refusal intact per item 4 mask)
- [x] Solution fits project architecture (audit owns bindings; install calls CheckSourceAudit; no new interfaces)
- [x] Tests green (items 1–6)
- [x] Relevant tests written for new/changed behavior and passing (unit 24 rows + production 18 rows + controls)
- [x] Lint clean (gofmt, go vet)
- [x] Build/validation commands run after changes; build not broken (all test binaries built and ran)
- [x] Outcome artifact attached (this file, TASK-260910-hwxr26_results-rev5.md)
- [x] No live credential export or runtime-home modification; unrelated dirty work untouched
