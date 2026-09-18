# TASK-260910-hwxr26 results — rework rev3 (verdict rev2 F1)

Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90 (Story branch task-board/story/STORY-260910-20sx61).
Scope: internal/audit, registry, artifactpolicy, scriptpolicy. No spec edits. Uncommitted working tree (handoff snapshot).

## Fix (one P2 from verdict rev2)

`internal/audit/sourceaudit.go` `ValidateSourceAudit`: moved the renewable policy check
after evidence integrity/completeness, trusted bindings, time, and decision.
New order: identity, context, evidence, time, decision, policy.
A corrupt, incomplete, or mismatching existing record now refuses with its
non-renewable `source_audit_rejected: evidence|time|decision` diagnostic even
when the trusted policy also drifted; a renewable policy/stale error never hides
another refusal. Genuine policy renewal over valid evidence still returns the
renewable policy error and re-issues on the mutating path. `CheckSourceAudit`
needed no change: it already routes renewal through `ValidateSourceAudit` +
`isRenewableAuditError`, and the Load/unavailable split from rework-1 is intact.

## Tests (committed, production entry named in AC)

`internal/audit/sourceaudit_test.go`:
- New `TestValidateSourceAuditPolicyDriftNeverHidesRefusal`: each row combines a
  policy-digest drift with one broken binding (corrupt report, digest mismatch,
  wrong-skill recomputed, findings mismatch, future timestamp, decision drift);
  asserts the non-policy diagnostic and non-renewable classification; plus a
  genuine-renewal row asserting the renewable policy error.

`internal/install/draftaudit_test.go`:
- New `TestDraftAuditPolicyRenewalBrokenEvidence` (5 subtests via
  `install.Project`, FailOn high -> critical drift): corrupt ({}), wrong-skill
  recomputed with updated digest, future-time object, decision-drift
  (object+report warn vs live allow) all refuse with `source_audit` (never the
  renewable policy wording) and leave object/report bytes unoverwritten;
  genuine-renewal (valid evidence + drift) passes the gate and re-issues with a
  new policy digest.

## Evidence (shell: bash; real exit codes via `set -o pipefail`)

1. `go test -p 1 ./internal/audit -run 'TestValidateSourceAuditPolicyDriftNeverHidesRefusal' -count=1 -timeout=60s -v` — exit 0 (7/7 subtests PASS).
2. `go test -p 1 ./internal/audit -run 'TestValidateSourceAudit|TestCheckSourceAudit' -count=1 -timeout=60s` — exit 0.
3. `go test -p 1 ./internal/install -run 'TestDraftAuditPolicyRenewalBrokenEvidence' -count=1 -timeout=100s -v` — exit 0 (5/5 subtests PASS).
4. `go test -p 1 ./internal/install -run 'TestDraftAuditPolicyRenewalBrokenEvidence|TestDraftAuditBrokenReport|TestDraftAuditFirstIssuance' -count=1 -timeout=100s` — exit 0.
5. `go test -p 1 ./internal/audit ./internal/install ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'Test(ParseSourceAudit|ValidateSourceAudit|SourceAudit|CheckSourceAudit|PolicyDigest|DraftAudit|Local|EffectiveLabels)' -count=1 -timeout=100s` — exit 0 (registry: no tests in mask).
6. `go test -p 1 ./internal/registry ./internal/install ./internal/interop -run 'Test(CheckLocal|ResolveWithoutIdentity|LegacyInstallUntouched|GoldenRegistryObjects)' -count=1 -timeout=100s` — exit 0.
7. `go test -p 1 ./internal/audit -count=1 -timeout=60s` (full audit package) — exit 0.
8. Reviewer overlay probe `go test -p 1 -overlay /tmp/hwx-overlay.json ./internal/install -run '^TestReviewPolicyRenewalBrokenEvidence$' -count=1` (board resource `TASK-260910-hwxr26_review-overlay-rev2.go`) — exit 0 PASS (was exit 1 at rev2).
9. `go vet ./internal/audit ./internal/install` — exit 0. `git diff --check` — exit 0. `gofmt -l` on touched files — clean.
10. Full local suite NOT run (host stalls; per wave brief, narrow packages only).

## Narrowing implementation mutant (reorder-back)

Mutant: policy check moved back before evidence (exact revert of this fix),
applied to a scratch copy of `sourceaudit.go`, then restored byte-identical
(`diff /tmp/sourceaudit.fixed.go internal/audit/sourceaudit.go` identical).
- `go test -p 1 ./internal/audit -run TestValidateSourceAuditPolicyDriftNeverHidesRefusal` under mutant — FAIL (6/6 broken rows return renewable `source_audit_rejected: policy:` instead of evidence/time/decision).
- `go test -p 1 ./internal/install -run 'TestDraftAuditPolicyRenewalBrokenEvidence/corrupt'` under mutant — FAIL (corrupt+drift admitted past source audit, reaches downstream `install marker is invalid for schema 2` instead of `source_audit` refusal).
Both new rows kill the implementation mutant; they are not input variations.

## Checklist

- Scoped production behavior with traceability to skillfile-sources §4: done (ordering fix only; no other behavior changed).
- Positive, negative, legacy regression checks with exact revision: done (see Evidence).
- Implementation matches AC; fits architecture (draft lane only, frozen v1 untouched, source-audit distinct from registry attestation): done.
- Relevant tests written and passing; lint clean; build not broken: done.
- Review rev2 F1 addressed; reviewer probe adopted-passing plus new narrowing rows: done.

## Files

- internal/audit/sourceaudit.go (ValidateSourceAudit reorder + comment)
- internal/audit/sourceaudit_test.go (TestValidateSourceAuditPolicyDriftNeverHidesRefusal)
- internal/install/draftaudit_test.go (TestDraftAuditPolicyRenewalBrokenEvidence)
- Retained rev1/rev2 wiring: internal/install/install.go, internal/registry/registry.go, internal/registry/local.go, internal/artifactpolicy/labels.go, internal/scriptpolicy/labels.go and tests.
