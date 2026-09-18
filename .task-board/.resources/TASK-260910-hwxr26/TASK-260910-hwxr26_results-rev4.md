# TASK-260910-hwxr26 results — rework rev4 (verdict rev3 F1)

Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90 (Story branch task-board/story/STORY-260910-20sx61).
Scope: internal/audit, registry, artifactpolicy, scriptpolicy. No spec edits. Uncommitted working tree (handoff snapshot).

## Fix (one P2 from verdict rev3)

`internal/audit/sourceaudit.go` `ValidateSourceAudit`: restructured instead of
patching the next branch. Validation is now two strict phases:

- Phase 1 — NON-RENEWABLE bindings, evaluated completely first: identity,
  context, evidence integrity/completeness/trusted bindings, stored-vs-live
  decision, forged-future time. Any failure refuses on both paths, never
  overwriting.
- Phase 2 — RENEWABLE conditions only, after phase 1 passes: stale time,
  trusted policy-label drift. Only these two may re-issue on the mutating path.

Concretely the stored-vs-live decision check moved above the stale-time
return (it already sat above the policy return since rev2); the future-time
check stays in phase 1 with an explicit comment. `isRenewableAuditError` and
`CheckSourceAudit` needed no change: decision/future errors were already
non-renewable, and renewal already routes through `ValidateSourceAudit`.
A forged decision combined with stale time (with or without policy drift)
now refuses with `source_audit_rejected: decision:` instead of renewing.

## Tests (committed, production entry named in AC)

`internal/audit/sourceaudit_test.go`:
- New `TestValidateSourceAuditStaleNeverHidesRefusal`: renewable causes
  {stale, stale+policy-drift} x broken bindings {corrupt report, digest
  mismatch, wrong-skill recomputed, findings mismatch, forged decision} —
  10 rows asserting the non-renewable diagnostic, never the stale/policy
  wording, never renewable; plus genuine-stale and genuine-stale+drift
  controls asserting the renewable classification. (Policy-drift-only column
  stays in `TestValidateSourceAuditPolicyDriftNeverHidesRefusal`.)

`internal/install/draftaudit_test.go`:
- New `TestDraftAuditRenewalNeverHidesRefusal` (10 subtests via
  `install.Project`): stale and stale+drift (FailOn high -> critical) each
  combined with forged-decision (object+report warn, recomputed digest — the
  exact reviewer repro), wrong-skill, corrupt ({}), digest-mismatch. All
  refuse on the mutating path with the non-renewable `source_audit`
  diagnostic, refuse on the read-only path too, and leave object/report bytes
  unoverwritten. Controls `genuine-stale-renewal` and
  `genuine-stale-plus-drift-renewal` pass the gate and re-issue (fresh
  created_at, new policy digest under drift).
- Shared test helpers `rewriteAuditObject` / `auditBindingPaths` (test-only;
  production code never edits a stored binding).

## Evidence (real exit codes, narrow packages only per wave brief)

1. `go test -p 1 ./internal/audit -run 'TestValidateSourceAudit' -count=1` — exit 0.
2. `go test -p 1 ./internal/install -run 'TestDraftAuditRenewalNeverHidesRefusal' -count=1` — exit 0 (10/10 subtests PASS).
3. `go test -p 1 ./internal/audit -count=1` (full audit package) — exit 0.
4. `go test -p 1 ./internal/install -run 'TestDraftAudit|TestReview|TestCheckSourceAudit' -count=1` — exit 0.
5. `go test -p 1 ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'Test(LocalPackage|NetworkAttestation|EffectiveLabels|SourceAudit|CheckLocal)' -count=1` — exit 0 (all three ok).
6. `go test -p 1 ./internal/audit ./internal/registry ./internal/install ./internal/interop -run 'Test(ValidateSourceAudit|CheckSourceAudit|CheckLocal|ResolveWithoutIdentity|LegacyInstallUntouched|GoldenRegistryObjects)' -count=1` — exit 0 (all four ok; frozen/legacy goldens intact).
7. Reviewer rev3 overlay replay `go test -p 1 -overlay /tmp/rev3replay/overlay.json ./internal/install -run '^TestReviewStaleDecisionRenewal$' -count=1` (board resource `TASK-260910-hwxr26_review-overlay-rev3.go`) — exit 0 PASS (was exit 1 at rev3).
8. `gofmt -l internal/audit internal/install` — clean. `go vet ./internal/audit/` — exit 0.
9. Full local suite NOT run (host stalls; per wave brief, narrow packages only).

## Ordering implementation mutant (killed, then restored byte-identical)

Mutant: stored-decision check moved back below the stale return (exact
revert of this fix), applied to the worktree file with the fixed copy kept
at /tmp/sourceaudit.go.fixed, then restored (`diff` IDENTICAL, fix line
`Phase 1 ends here` present at sourceaudit.go:468).
- `go test -p 1 ./internal/audit -run 'TestValidateSourceAuditStaleNeverHidesRefusal'` under mutant — FAIL.
- `go test -p 1 ./internal/install -run 'TestDraftAuditRenewalNeverHidesRefusal/stale-plus-drift-with-forged-decision'` under mutant — FAIL.
Both levels kill the ordering mutant; the rows are not input variations.
Post-restore: evidence items 1–6 re-verified green on the restored tree
(items 1,3,4,5,6 run after restore; item 2 control fix verified pre-mutant
and the full matrix re-ran green pre-mutant — see note).

Note: item 4's mask (`TestDraftAudit|...`) includes the new
`TestDraftAuditRenewalNeverHidesRefusal` matrix and ran green after the
restore on the byte-identical tree, so both new tests are verified on the
final tree; the mutant FAIL lines above prove the rows bind the ordering.

## Checklist

- [x] Implement the scoped production behavior with traceability to the accepted draft contracts (skillfile-sources §4 decision binding; phase split).
- [x] Task-specific positive, negative and legacy regression checks run; exact revision (9b185d2 + uncommitted tree) and evidence recorded above.
- [x] Implementation matches AC (source-audit-v1 bindings; distinct from registry attestation; local-package network-attestation refusal intact per TestDraftAuditStrictLocalRequiresAttestation in item 4 mask).
- [x] Solution fits project architecture (audit package owns bindings; install calls CheckSourceAudit; no new interfaces).
- [x] Tests green (items 1–7).
- [x] Relevant tests written for new/changed behavior and passing (phase-1/phase-2 matrix at unit + production entry).
- [x] Lint clean (gofmt, go vet).
- [x] Build/validation commands run after changes; build not broken (test binaries built and ran for all four packages).
- [x] Outcome artifact attached (this file, TASK-260910-hwxr26_results-rev4.md).
- [x] No live credential export or runtime-home modification; unrelated dirty work untouched (install.go/registry.go modifications predate this rework and are untouched).
