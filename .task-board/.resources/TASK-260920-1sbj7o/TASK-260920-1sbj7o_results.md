# TASK-260920-1sbj7o results — install-level refusal preserves build_repository_identity_invalid

## Outcome

`install.Project` and `curator install` / `status` now report
`build_repository_identity_invalid` (not `build_repository_source_unavailable`)
for an invalid identity plan. The true class from `ValidateTransportPlan` is
preserved end to end, and the CLI remediation table gained a sanitized row for
it. Ready for review.

## Root cause

`buildrepo.RunPipeline` (`internal/buildrepo/pipeline.go`) collapsed **every**
`Acquire` failure to `build_repository_source_unavailable: exact external
source is unavailable`. The install external-planning path
(`planExternalBuilds` / `stageExternalBuilds` in `internal/install/external.go`)
routes the draft transport acquisition through `RunPipeline`, so a
`ValidateTransportPlan` refusal (`build_repository_identity_invalid` for a
port/alias endpoint outside the strict external-build lane grammar) was masked
before `install.Project` ever saw it.

Scope note: the task scope names "internal/install external planning error
classification". The classification (collapse) physically lives in
`RunPipeline`, which is only reached from install external planning for this
lane; the fix is there (7 lines incl. comment), at the root cause, covering
both the plan and stage phases. No other fix location can recover the class
once collapsed.

## Changes (6 files, +85/-9)

- `internal/buildrepo/pipeline.go` — `RunPipeline` returns an `Acquire` error
  whose `ErrorCode` is `CodeIdentityInvalid` unchanged, skipping the offline
  snapshot fallback (a deterministic refusal is not an availability failure,
  so no snapshot may substitute and no audit runs past it). All other
  `Acquire` failures keep the existing collapse + offline behavior.
- `cmd/curator/draft_diagnostics.go` — new `draftRemediations` row for
  `build_repository_identity_invalid`: "fix the endpoint entry in machine
  source-policy.json: the strict external-build lane admits no explicit port
  and no host alias, then retry the explicit attempt". Static prose, no URL,
  scheme, or secret. Applies to both `install` and `status` via the shared
  `printFailures` → `withDraftRemediation` path.
- `internal/crossconformance/draftsources_semantic_v2_test.go` —
  `driveV2ExternalRefused` (rows `v2-external-build-port-refused`,
  `v2-external-build-alias-refused`) now asserts the exact class in
  `install.Project` errors and fails if the masked class appears (was:
  logged only).
- `internal/buildrepo/pipeline_test.go` — new
  `TestPipelinePreservesIdentityInvalidFromAcquisition`: preservation,
  byte-identical diagnostic, and no `snapshot-load` / `audit-call` /
  cache / compiler events past the refusal.
- `cmd/curator/draft_transport_test.go` — new
  `TestDraftTransportIdentityInvalidSurfacesThroughCLI`: `install --dry-run`
  through `run()` with a schema-2 port policy for both fixture identities
  fails with the true class + remediation, zero fetches, and no masked class.
  (`status` shares the identical `install.Project(DryRun)` + `printFailures`
  path, so it reports the same text; the existing
  `TestDraftExternalStatusFailsClosedWithoutSource` still pins the
  `source_unavailable` fail-closed half.)
- `cmd/curator/draft_diagnostics_test.go` — positive remediation row for the
  class; the legacy list now pins `build_repository_source_unavailable` as
  still-unremediated instead (other classes unchanged at the remediation
  layer).

## Narrow evidence (all `-p 1 -count=1`, real exit codes)

| check | exit |
|---|---|
| `go build ./...` | 0 |
| `go vet` on touched packages | 0 |
| `gofmt -l` on touched files (empty) + `golangci-lint run` on touched trees (0 issues) | 0 |
| buildrepo: `TestPipelinePreservesIdentityInvalidFromAcquisition`, `TestPipelineFailuresStopBeforeCacheAndCompiler`, `TestExternalPipelineOfflineProtectedSnapshotAndTagFailure`, `TestProtectedSnapshotCorruptionQuarantinesBeforeAudit` | 0 |
| cmd/curator: `TestWithDraftRemediationTable`, `TestDraftAttemptTransportTable`, `TestDraftAttemptClauseShape`, `TestEndpointExhaustionCarriesRemediation` | 0 |
| cmd/curator: `TestDraftTransportIdentityInvalidSurfacesThroughCLI` (new, 8.3s) | 0 |
| cmd/curator: `TestDraftRemediationThroughCLI` (full suite, no regression) | 0 |
| cmd/curator: `TestDraftTransportResolvedMatrix`, `TestDraftTransportProviderAdmission` (guards: fail-closed TLS, invalid policy, unknown providers still `source_unavailable`) | 0 |
| crossconformance: `TestDraftExternalStatusFailsClosedWithoutSource` (guard) | 0 |
| buildrepo: `TestValidateTransportPlan*`, `TestClassify*`, `TestAllowSecondAttempt*`, `TestExhaustion*`, `TestAcquireNetworkResolved*` (gate + exhaustion shape) | 0 |
| install: `TestRevision2*`, `TestDraftTransport*`, `TestUserConfigIgnored*`, `TestResolvedLane*` | 0 |
| crossconformance rows `v2-external-build-port-refused` + `v2-external-build-alias-refused` via `-run` filter | subtests PASS (parent tally exit 1 is the harness's by-design "run the full matrix without -run subtest filters" artifact; ratio line: `2 driven, 0 known-gap, 0 bound, 0 skipped`) |

Full-matrix run (94 semantic cases, ~41 min per rev3) is out of scope for this
headless turn; the two affected rows pass in isolation and the tally
mechanism is untouched.

## Narrowing mutant: re-mask the class → KILLED at the production entry

Method: `git stash push -- internal/buildrepo/pipeline.go` (production fix
reverted, tests kept), rerun, `git stash pop`, re-verify green.

| mutant check (fix reverted) | exit | observed |
|---|---|---|
| `TestPipelinePreservesIdentityInvalidFromAcquisition` | 1 | `error = build_repository_source_unavailable: exact external source is unavailable, want ... identity_invalid` |
| crossconformance `v2-external-build-port-refused` + `v2-external-build-alias-refused` (`install.Project` entry) | 1 | both subtests FAIL |
| `TestDraftTransportIdentityInvalidSurfacesThroughCLI` | 1 | `stderr: error: skill-a.https-cmd: build_repository_source_unavailable: exact external source is unavailable` |

After `git stash pop`: pipeline + remediation checks re-run green (exit 0).

## Unchanged (verified, not assumed)

- Other refusal classes: generic/`offline` acquire errors, `repository_policy_invalid`,
  `repository_endpoint_unavailable` exhaustion, unknown-provider skips, and
  fail-closed fetch classes still collapse to `source_unavailable` at install
  level (guards above all exit 0; the `invalid policy` CLI test documents the
  intentional masking for `repository_policy_invalid`).
- Exhaustion shape: `repository_endpoint_unavailable` construction untouched;
  transport gate/exhaustion tests exit 0.
- Remediation layer: `build_repository_source_unavailable` stays unremediated
  (legacy pin); idempotence + already-guided rows still pass.
- Legacy v1: no frozen schema/lane file touched; change set is 6 files, none
  overlapping the checkpointed siblings (3ukdk4, 2wyzde, 3v7x6j, 2eg8nv).
- Sanitization: preserved diagnostic names only the endpoint index
  ("transport plan endpoint 1 carries an explicit port or host alias..."),
  never the URL; remediation carries no URL, scheme, or secret; `RedactDiagnostic`
  still bounds the install error text.

## Checklist mapping

- [x] install.Project + install/status report `build_repository_identity_invalid` for an invalid identity plan; `ValidateTransportPlan` class preserved end to end
- [x] crossconformance rows assert the exact class; remediation row exists; re-mask mutant killed
- [x] narrow evidence with exit codes here; handoff via `task-board handoff`
- [x] code per task description and AC
- [x] tests written for new/changed behavior and passing
- [x] lint clean (`golangci-lint run`, 0 issues; `gofmt` clean)
- [x] build + validation commands run after changes; build not broken
- [x] this outcome artifact attached as `TASK-260920-1sbj7o_results.md`
- [x] scope note (root cause in `RunPipeline` on the install path) recorded here

revision 2 = revision 1 unchanged; gate rerun after the macOS git flake BUG-260920-3vfwch

---

# Revision 3 (rework-1 after rev2 CHANGES_REQUESTED)

## Ruling on F2 (scope, not declare)

The `RunPipeline` preservation is now scoped to the §7 refusal's own
static diagnostic, `build_repository_identity_invalid: transport plan
endpoint` — the only identity_invalid diagnostic that names a
transport plan endpoint (produced solely by `ValidateTransportPlan`,
resolved lane only). This is the reviewer's preferred option in its
strictest form: the frozen lane is byte-identical, and no other
draft-lane refusal changes class either. No CHANGELOG entry: the only
newly-surfaced class is the AC itself on the unreleased draft lane, so
no released behavior changes.

Single source of truth: new exported
`buildrepo.TransportPlanRefusalPrefix` (`internal/buildrepo/transport.go`),
used both by the `RunPipeline` preservation predicate and by the CLI
remediation row key, so the two cannot drift apart.

## Changes vs rev2 (candidate now 8 files)

- `internal/buildrepo/transport.go` — new `TransportPlanRefusalPrefix`
  const beside `ValidateTransportPlan`.
- `internal/buildrepo/pipeline.go` — preservation predicate now also
  requires the §7 prefix (class + static diagnostic).
- `cmd/curator/draft_diagnostics.go` — remediation row re-keyed from
  the bare class to `TransportPlanRefusalPrefix`; table comment notes
  the substring keying.
- `internal/buildrepo/pipeline_test.go` — new
  `TestPipelineCollapsesNonTransportPlanIdentityInvalid`: a non-§7
  identity refusal (frozen SSH-wrapper message) still collapses to the
  byte-identical legacy diagnostic via the legacy snapshot branch.
- `cmd/curator/draft_diagnostics_test.go` — stbg4d pin restored
  verbatim (`build_repository_identity_invalid: wrong identity`) plus
  a real lane message pin (`...: SSH requires the exact manager
  wrapper`); docs-pin class list and remedy-keyword map extended.
- `cmd/curator/draft_transport_test.go` — `runDraftInstall`
  refactored to delegate to a new `runDraftCLI` (args parameter; all
  12 existing call sites untouched); `status` assertions added to the
  CLI row test; committed `TestReviewProbeLegacyLaneIdentityInvalid`.
- `docs/troubleshooting.md` — new `### build_repository_identity_invalid`
  section, scoped to the transport-plan refusal on the resolved lane.
- `internal/crossconformance/draftsources_semantic_v2_test.go` —
  unchanged from rev2 (exact-class assertions stand).

## Finding mapping

- F1: row keys on the §7 static diagnostic; stbg4d pin restored;
  legacy probe committed as a negative row — frozen output is the
  base diagnostic `error: skill-a.ssh-cmd:
  build_repository_source_unavailable: exact external source is
  unavailable` for both `install --dry-run` and `status`, with no
  `identity_invalid` and no appended guidance.
- F2: scoped per the ruling above; pinned by the committed legacy
  probe and the pipeline collapse test.
- F3: docs section + pin entries added. Section intro unchanged: with
  scoped preservation the class still appears at install/status level
  only on the draft lane, so "appear only on the draft Skillfile lane"
  stays true.
- F4: `status` row committed — exact class line
  (`error: skill-a.https-cmd: build_repository_identity_invalid:
  transport plan endpoint 1 carries ...`), remediation, no masked
  class, zero fetches.
- R1 (bound, not fixed): `admissionError` is unexported in buildrepo
  with no exported constructor, so install cannot build one in a
  single line; the install-layer `fmt.Errorf`-with-class matches local
  convention (cf. `drafttransport.go:203`). Behavior-neutral under
  scoped preservation: those messages lack the §7 fragment, so they
  still collapse exactly as on base.

## Narrow evidence, revision 3 (bash, `set -o pipefail`, `-p 1 -count=1`)

| check | exit |
|---|---|
| `go build ./...` | 0 |
| `go vet` buildrepo/install/cmd/curator/crossconformance | 0 |
| `gofmt -l` on touched files (empty) | 0 |
| `golangci-lint run ./internal/buildrepo/... ./cmd/curator/...` (0 issues) | 0 |
| buildrepo `TestPipelinePreservesIdentityInvalidFromAcquisition`, `TestPipelineCollapsesNonTransportPlanIdentityInvalid`, `TestPipelineFailuresStopBeforeCacheAndCompiler` | 0 |
| buildrepo `-run 'Pipeline\|Admission'` (full subset) | 0 |
| cmd/curator `TestWithDraftRemediationTable`, `TestDraftDocsPinExamples` (stbg4d pins + docs pin) | 0 |
| cmd/curator `-run 'TestDraftDiagnostics\|TestDraftDocsPin\|TestDraftTransport\|TestReviewProbe'` (DocsPin, LegacyGolden, ResolvedMatrix, ProviderAdmission, IdentityInvalid CLI incl. status, ReviewProbe) | 0 |
| cmd/curator guards: AttemptTransportTable, AttemptClauseShape, EndpointExhaustionCarriesRemediation, RemediationThroughCLI | 0 |
| crossconformance `TestDraftExternalStatusFailsClosedWithoutSource` (fail-closed guard) | 0 |
| install `-run 'TestDraftTransport\|TestUserConfigIgnored\|TestResolvedLane'` | 0 |
| crossconformance `v2-external-build-port-refused` + `v2-external-build-alias-refused` | subtests 2/2 PASS, `2 driven, 0 known-gap, 0 bound, 0 skipped` (parent exit 1 is the by-design subtest-filter tally artifact, as in rev1) |
| final-bytes re-runs: remediation+pin, pipeline pair, CLI row, legacy probe, xconf rows | 0 (xconf 2/2 PASS) |

## Narrowing mutants (revision 3)

Method M1: `git stash push -- internal/buildrepo/pipeline.go`
(production fix reverted; `TransportPlanRefusalPrefix` const stays,
unused const compiles), rerun, `git stash pop`, re-verify green.

| M1 re-collapse (fix reverted) | exit | observed |
|---|---|---|
| `TestPipelinePreservesIdentityInvalidFromAcquisition` | 1 | `error = build_repository_source_unavailable: ...`, want preserved class — KILLED |
| xconf port+alias rows (`install.Project` entry) | 1 | both subtests FAIL — KILLED |
| `TestDraftTransportIdentityInvalidSurfacesThroughCLI` | 1 | FAIL — KILLED |
| `TestReviewProbeLegacyLaneIdentityInvalid` + `TestWithDraftRemediationTable` | 0 | still pass — mutant correctly scoped (frozen lane + presentation untouched) |

Method M2: re-widen the remediation key to the bare class (one-line
sed), rerun, restore. `TestWithDraftRemediationTable` exit 1 —
KILLED by the restored stbg4d pins (both the synthetic and the real
lane message fire the mis-scoped row).

## Unchanged (verified, not assumed)

- Other refusal classes: generic/offline acquire errors, non-§7
  identity refusals, `repository_policy_invalid`, exhaustion shape,
  unknown-provider skips, fail-closed fetch classes — all still
  collapse/guard exactly as on base (guards above exit 0; the new
  pipeline collapse test pins the non-§7 identity path explicitly).
- Remediation layer: `build_repository_source_unavailable` stays
  unremediated; idempotence + already-guided rows still pass.
- Legacy v1: frozen lane is byte-identical, pinned end to end by the
  committed probe (base diagnostic reproduced verbatim); no
  `internal/install` production file touched; sibling checkpoints
  (3ukdk4, 2wyzde, 3v7x6j, 2eg8nv) untouched — `git status --short`
  lists exactly the 8 candidate paths above.
- Sanitization: preserved diagnostic names only the endpoint index;
  remediation carries no URL, scheme, or secret.

## Revision 3 checklist mapping

- [x] F1: remediation row scoped to the §7 refusal; stbg4d pin restored; legacy-lane negative row committed
- [x] F2: preservation scoped (ruling recorded here); frozen lane byte-identical and pinned
- [x] F3: troubleshooting section + docs-pin entries; intro judgment recorded
- [x] F4: committed `status` row with the exact class line
- [x] R1 recorded as a bound with cause (no exported constructor; behavior-neutral)
- [x] AC holds: install.Project + install/status report the true class; xconf rows assert it; remediation row present
- [x] re-mask mutant + re-widen mutant killed; scope survivors explained
- [x] narrow evidence with exit codes here; lint clean; build not broken
- [x] this outcome artifact updated as `TASK-260920-1sbj7o_results.md`
- [x] ready for review; handoff via `task-board handoff`
