# TASK-260910-hwxr26 results — rework rev6 (verdict rev5 F1/F2)

Base: 9b185d2503d7bfb4fe66a33ad591a50abe9f8c90 (Story branch task-board/story/STORY-260910-20sx61).
Scope: internal/audit, registry, artifactpolicy, scriptpolicy. No spec edits. Uncommitted working tree (handoff snapshot).
Changed by this rework: internal/audit/sourceaudit.go, internal/audit/sourceaudit_test.go, internal/install/draftaudit_test.go.
Unrelated dirty work (install.go/registry.go modifications, labels/local/draftaudit.go files) predates this rework and is untouched.

## Fixes (two P2 parser/shape gaps from verdict rev5)

F1 — `ParseSourceAudit` (sourceaudit.go:203) decoded only the first JSON value and never required EOF:
appending `{}` to an otherwise valid persisted binding was admitted through `install.Project`.
Fixed to require exactly one complete JSON document: `json.Decoder` with `DisallowUnknownFields`
(closed shape preserved), then a second `Decode` must return `io.EOF` — any trailing token refuses
`source_audit_rejected: malformed`. Trailing whitespace/newline (stored files end with `\n`) still admits.

F2 — evidence presence checks followed by decoding into `bool` accepted JSON null as `false` for
`pinned`/`revoked` (and symmetrically for every other required member via Go zero values).
On an unpinned, non-revoked package, replacing either member with null and recomputing
`evidence_sha256` was admitted. Fixed structurally, not per-field: after `json.Unmarshal` into
`map[string]json.RawMessage`, every required member (`schema_version`, `skill`, `package`,
`content_sha256`, `findings`, `pinned`, `revoked`, `revocation`, `script_policy`,
`assurance_policy`, `decision`) must be present AND non-null — `bytes.TrimSpace(raw[key]) == "null"`
refuses `source_audit_rejected: evidence: report member %q is null` before any trusted comparison.
Wrong types are refused by the existing strict `Decoder` with `DisallowUnknownFields` (now also
EOF-checked for the report as defense-in-depth). Typed renewal (`sourceAuditError`/`Renewable`/`Cause`
via `errors.As`, phase-1 before phase-2) is preserved byte-identical; valid stale/policy renewals unchanged.

## Tests (committed, production entry named)

`internal/audit/sourceaudit_test.go` — `TestParseSourceAuditMalformed` gains three rows:
trailing object (`s + "{}"`), trailing newline plus object (`s + "\n{}"`), trailing garbage
(`s + " trailing"`) — all must refuse `source_audit_rejected`. Valid single document (with and
without trailing newline) still admits.

`internal/install/draftaudit_test.go` — new `TestDraftAuditMalformedShapeRefuses` via
`install.Project` (production call site). One fixture is established once
(`strictLocalProject` + mutating `Project` + `auditBindingPaths`) and restored per row, so the
table stays fast with a single Git setup. Rows (24): `trailing-object` (append `{}` to the stored
object), `trailing-report` (append `{}` to the report + recomputed digest), `null-<member>` for
each of the 11 required report members, `wrongtype-<member>` for each of the 11 members
(schema_version:"1", skill:123, package:"bad", content_sha256:123, findings:"bad",
pinned:"true", revoked:1, revocation:123, script_policy:"bad", assurance_policy:123,
decision:123). Each row refuses on the mutating path with `source_audit`, refuses on the
read-only path too, and leaves object/report bytes unoverwritten. Final controls restore the
valid binding and assert the gate passes on both paths.

## Evidence (real exit codes, narrow packages only per wave brief; shell bash with `set -o pipefail`)

1. `go test -p 1 ./internal/audit -run 'TestParseSourceAudit' -count=1 -timeout=90s` — exit 0 (ok 0.700s).
2. `go test -p 1 ./internal/install -run '^TestDraftAuditMalformedShapeRefuses$' -count=1 -timeout=110s -v` — exit 0, 24/24 subtests PASS (ok 18.138s; test 17.45s).
3. `go test -p 1 ./internal/audit -count=1 -timeout=100s` (full audit package) — exit 0 (ok 1.284s).
4. `go test -p 1 ./internal/install -run '^TestDraftAudit(BindingLifecycle|HostileContentBlocked|PinAdmits)$' -count=1 -timeout=100s` — exit 0 (ok 6.884s).
5. `go test -p 1 ./internal/install -run '^TestDraftAuditRenewalNeverHidesRefusal$' -count=1 -timeout=110s` — exit 0 (ok 9.537s; rework-3 ordering preserved).
6. `go test -p 1 ./internal/install -run '^TestDraftAuditRenewalMarkersNeverRenewable$' -count=1 -timeout=110s` — exit 0 (ok 31.018s; rework-4 typed renewal preserved).
7. `go test -p 1 ./internal/install -run '^TestDraftAudit(RevocationBlocks|BrokenReportRefuses|FirstIssuanceVsBrokenRecord)$' -count=1 -timeout=110s` — exit 0 (ok 12.026s).
8. `go test -p 1 ./internal/install -run '^TestDraftAuditPolicyRenewalBrokenEvidence$' -count=1 -timeout=110s` — exit 0 (ok 7.896s).
9. `go test -p 1 ./internal/registry ./internal/artifactpolicy ./internal/scriptpolicy -run 'Test(LocalPackage|EffectiveLabels|SourceAudit|DraftAudit)' -count=1 -timeout=100s` — exit 0 (registry no tests matched; artifactpolicy ok; scriptpolicy ok).
10. `go vet ./internal/audit ./internal/install` — exit 0. `gofmt -l` on the three touched files — clean.
11. Full local suite NOT run (host stalls; per wave brief, narrow packages only). No cross-platform run claimed.

## Implementation mutant (lenient decoding restored, then restored)

Mutant: removed both EOF trailing checks and the generic null-rejection block (restoring the exact
rev5 lenient shape; fixed copy kept at /tmp/hwxr26/sourceaudit.fixed.go, restored with `cp` and
verified `grep` shows `var extra`, `reportExtra`, `is null` present plus `TestParseSourceAudit` exit 0).
- `go test -p 1 ./internal/install -run 'TestDraftAuditMalformedShapeRefuses/(trailing-object|null-pinned|null-revoked|null-revocation)' -count=1 -timeout=100s` under mutant — exit 1, 4/4 FAIL. Each failure shows only `review: install marker is invalid for schema 2` (gate bypassed to the sibling marker boundary) — the exact reviewer rev5 signature.
The new table kills the lenient-decoding mutant; the fix binds the shape, not input variation.

## Reviewer probe rev5 disposition

The attached `TASK-260910-hwxr26_review-probe-rev5.go` rows (`trailing-object`, `null-pinned`,
`null-revoked`) are a strict subset of the committed table and use the same production entry and
helpers. The committed run above refuses all three with `source_audit` on both paths without
overwrite; the probe's `Contains(result.Errors, "source_audit")` assertion now passes. No overlay
replay claimed beyond the mutant reproduction above.

## Checklist

- [x] Scoped production behavior with traceability to accepted draft contracts (skillfile-sources §4; single-document + complete-report bindings)
- [x] Task-specific positive, negative and legacy regression checks run; exact revision and evidence recorded above
- [x] Implementation matches AC (source-audit-v1 identity/context/policy/evidence/time/decision bindings; distinct from registry attestation; local network-attestation refusal intact per preserved lifecycle tests)
- [x] Solution fits project architecture (audit owns parsing/validation; install calls CheckSourceAudit; no new interfaces)
- [x] Tests green (items 1–9)
- [x] Relevant tests written for new/changed behavior and passing (audit 3 rows + production 24 rows + controls)
- [x] Lint clean (gofmt, go vet)
- [x] Build/validation commands run after changes; build not broken (all test binaries built and ran)
- [x] Outcome artifact attached (this file, TASK-260910-hwxr26_results-rev6.md)
- [x] No live credential export or runtime-home modification; unrelated dirty work untouched
