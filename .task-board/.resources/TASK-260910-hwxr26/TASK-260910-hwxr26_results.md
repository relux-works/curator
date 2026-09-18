# TASK-260910-hwxr26 results \u2014 bind source audit to existing assurance gates

Status: ready for review (developer role complete, uncommitted working tree).
Base: story branch `task-board/story/STORY-260910-20sx61` at
`9b185d2503d7bfb4fe66a33ad591a50abe9f8c90`. Work left UNCOMMITTED per
handoff policy. Shell for all commands below: `bash`, `set -o pipefail`
not needed (no pipes); every command quoted with a real exit code.

## Scope kept

Only `internal/audit`, `internal/registry`, `internal/artifactpolicy`,
`internal/scriptpolicy`, plus a draft-scoped call site in
`internal/install` (step 12b, runs only when a schema-2 lock is consumed;
frozen v1 lane never sets `draftLock`). No spec edits, no marker/status
changes, no frozen-lane behavior change.

## What was built

- `internal/audit/sourceaudit.go` (new): source-audit-v1 record
  (skillfile-sources \u00a74). Strict closed-shape parse; policy digest over
  audit mode/threshold, detector generations, revocations, and effective
  script/assurance labels; canonical evidence report (findings, pins,
  revocations, script/assurance labels, verdict); six-binding validation in
  fixed order identity/context/policy/evidence/time/decision; machine store
  under `<home>/source-audit/` (never the project tree); `CheckSourceAudit`
  gate \u2014 no-op when audit disabled, establishes the binding through a live
  run on the mutating path, fails unavailable/stale on the read-only path
  (missing evidence is unknown, never current). The stored object is never
  authority: the live decision is always recomputed and enforced.
- `internal/registry/local.go` (new): `CheckLocalPackage` \u2014 strict
  registry policy + local-snapshot fails ("no network attestation");
  advisory passes; Git kinds resolve on the normal exact-match path.
- `internal/registry/registry.go`: `Resolve` with empty identity AND empty
  commit returns unknown without calling fetch \u2014 no content-hash-only
  coincidence can forge an attestation for local content. All existing
  callers pass non-empty identities (verified by test + grep).
- `internal/artifactpolicy/labels.go`, `internal/scriptpolicy/labels.go`
  (new): `EffectiveLabels()` \u2014 closed assurance/script posture labels bound
  into the policy digest and evidence report.
- `internal/install/draftaudit.go` + step-12b hook in `install.go`: per
  draft member runs the registry local check then the source-audit check,
  after validation/skillcheck and before registry resolution, build
  planning, cache, and compiler. Pins, revocation, canary, scriptpolicy,
  and assurance preflight all still run in their existing positions.

## Acceptance traceability

- identity/context/policy/evidence/time/decision bindings:
  `ValidateSourceAudit` + `TestValidateSourceAuditBindings` (10 negative
  rows, one per binding aspect, each asserting its binding substring).
- Distinct from registry attestation:
  `TestSourceAuditNeverRegistryAttestation` (attestation members rejected
  by closed-shape parse; evidence report carries none) and
  `TestResolveWithoutIdentityIsUnknown`.
- Required network attestation for local packages fails:
  `TestDraftAuditStrictLocalRequiresAttestation` (production `Project`
  dry-run, strict \u2192 `no network attestation`); narrowing twin
  `TestDraftAuditAdvisoryLocalPasses` (same fixture, advisory \u2192 ok).
- Authorized pins / revocation / assurance before cache-compiler:
  `TestDraftAuditHostileContentBlocked` (strict, hostile \u2192 `audit
  blocked`, and no source-audit state stored), `TestDraftAuditPinAdmits`
  (same hostile content + operator pin \u2192 gate passes, binding stored),
  `TestDraftAuditRevocationBlocks` (revocation dominates pin \u2192 refused),
  `TestDraftAuditEnforcedScriptRefused` (schema-8 enforced command \u2192
  `script_execution_policy_unsupported` at validation, before the new
  gate), unit rows `TestCheckSourceAuditPinPreserved`,
  `TestCheckSourceAuditRevocationPreserved`,
  `TestCheckSourceAuditNotSelfAuthorizing` (stored allow + hostile live
  bytes \u2192 refused on mutating AND read-only paths).
- Absent/malformed/stale/wrong evidence without weakening currentness:
  `TestParseSourceAuditMalformed` (9 rows incl. the spec's
  invalid-decision/invalid-no-evidence shapes),
  `TestCheckSourceAuditEstablishThenValidate` (read-only missing \u2192
  `source_audit_unavailable`; report deleted \u2192 unavailable, never ok),
  `TestCheckSourceAuditStaleReestablishesOnMutate` (read-only stale \u2192
  refused; mutating path re-runs live gates and re-issues),
  `TestDraftAuditBindingLifecycle` (production-path lifecycle).

## Evidence (real exit codes)

- `go test ./internal/audit/ ./internal/scriptpolicy/ -count=1` \u2192 ok (exit 0)
- `go test ./internal/registry/ -count=1` \u2192 ok (exit 0)
- `go test ./internal/artifactpolicy/ -count=1` \u2192 ok (exit 0, 267s)
- `go test ./internal/install/ -run
  'TestDraft|TestLegacyInstallUntouched|TestStrictRegistry|TestRegistry|TestEnforcedScript|TestScript'
  -count=1` \u2192 ok (exit 0, 80s)
- `go test ./internal/install/ -run
  'TestAssurance|TestBuildAuthority|TestAudit|TestAttest|TestMarker'
  -count=1` \u2192 ok (exit 0, 31s)
- `go test ./internal/interop/ -run 'TestGoldenRegistryObjects' -count=1`
  \u2192 ok (exit 0)
- `golangci-lint run internal/audit/... internal/registry/...
  internal/scriptpolicy/... internal/artifactpolicy/...
  internal/install/...` \u2192 0 issues (exit 0)
- `gofmt -l cmd internal` \u2192 empty; `go vet` on the five packages \u2192 clean.
- Delete-mutant probe (temporary `draftLock != nil && false`, reverted
  byte-exact afterwards): `TestDraftAuditStrictLocalRequiresAttestation`
  FAILS with the gate removed (install passes unattested) and passes with
  it \u2014 the tests drive the production gate, not helpers.

## Known boundary (not this task)

Real (non-dry-run) draft installs fail downstream at marker publication
with `review: install marker is invalid for schema 2` \u2014 verified
pre-existing with audit disabled (gate no-ops, same failure), caused by
draft members lacking v4 marker identity fields. Marker v5
(`package`/`lock_sha256`) belongs to the sibling build-receipts/runtime
tasks (TASK-260910-dufdai / TASK-260910-17ps6u). This task's
establishment tests therefore assert gate behavior + stored binding rather
than end-to-end `ok`. Status-currentness for draft markers is likewise
sibling scope; this task never reports current on missing/unreadable
evidence in the paths it owns.

## Files

Modified: `internal/install/install.go` (+21/-1: draftLock capture, step
12b, import), `internal/registry/registry.go` (+9: identity-less guard).
Added: `internal/audit/sourceaudit.go`,
`internal/audit/sourceaudit_test.go`, `internal/registry/local.go`,
`internal/registry/local_test.go`, `internal/artifactpolicy/labels.go`,
`internal/artifactpolicy/labels_test.go`,
`internal/scriptpolicy/labels.go`, `internal/scriptpolicy/labels_test.go`,
`internal/install/draftaudit.go`, `internal/install/draftaudit_test.go`.

revision 9 = revision 8 unchanged; gate rerun after a Windows runner flake
