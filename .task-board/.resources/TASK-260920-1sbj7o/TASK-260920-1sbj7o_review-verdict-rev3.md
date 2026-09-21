# Review verdict — TASK-260920-1sbj7o revision 3 (CR-TASK-260920-1sbj7o-3)

**Verdict: ACCEPTED** (`accept_cr(TASK-260920-1sbj7o, revision=3, evidence=this resource)`).
Reviewer: claude-opus-5 (RUN-260920-28636b), 2026-09-20. Independent evidence bundle:
`TASK-260920-1sbj7o_review-rev3-evidence.tar.gz` (drivers, raw logs, mutant outputs, base-tree probe, hosted ledgers).

## Tree and gate identity (verified, not assumed)

- Story worktree `git status --short` = exactly the 8 candidate paths; temp-index `read-tree HEAD && add -u && write-tree`
  = `158987e16fa9f9d93c9abb3caad45c1510b6b1b4` = CR candidate tree; `HEAD` = base `6a6e2a14`; index untouched.
- Patch resource `TASK-260920-1sbj7o_change-request_rev3.patch` sha256 `a69278b2…` matches the CR record.
- Hosted gate (rev3 validation log) run 35518913747: `gh run view --json headSha` = `bb8cb11a…`; locally
  `git rev-parse bb8cb11a^{tree}` = `158987e1…` (candidate), parent = base. Conclusion `success` on every lane
  (Test ubuntu/macos/windows, Race ubuntu/macos, Lint, Naming gate, Interop conformance, Gate self-test ×3; rose-air and
  candidate suite skipped by design).
- Review ran on disposable clones of the control root checked out at the gate commit (`/tmp/1sbj7o-rev3-review/{cand,probe}`,
  `HEAD^{tree}` = candidate before, between and after every mutant) and at the base (`base`). The Story worktree was never written.

## What revision 3 does (read against the rev2 verdict)

- `internal/buildrepo/transport.go:437` — exported `TransportPlanRefusalPrefix = CodeIdentityInvalid + ": transport plan endpoint"`,
  the static class+diagnostic prefix of the §7 refusal (`transport.go:499`, the only `build_repository_identity_invalid`
  producer that names a transport plan endpoint; every other "transport plan endpoint" message carries
  `repository_policy_invalid`/`repository_mirror_undeclared`).
- `internal/buildrepo/pipeline.go:211-221` — `RunPipeline` preserves an `Acquire` error only when
  `ErrorCode == CodeIdentityInvalid && strings.Contains(err, TransportPlanRefusalPrefix)`; every other acquire failure
  (including every other identity refusal) keeps the legacy offline-snapshot branch and the `source_unavailable` collapse.
  `parseTransportPlan` is reachable only from `ValidateTransportPlan` (`internal/install/drafttransport.go:212`) and
  `AcquireNetworkResolved` (`:171`), both behind `deps.DraftTransportResolution` with a non-nil policy; the frozen
  `AcquireNetwork` path never produces the prefix, so the frozen lane is byte-identical by construction.
- `cmd/curator/draft_diagnostics.go:101` — remediation row keyed on `buildrepo.TransportPlanRefusalPrefix` (single
  source of truth with the pipeline predicate); static prose, no URL/scheme/secret. Reaches `install` and `status` through
  the shared `printFailures` (`main.go:658-660`); the install-level text is `skill.cmd: build_repository_identity_invalid:
  transport plan endpoint N carries …` (`external.go:196`, `failBuild`→`RedactDiagnostic`, index only, never the URL).
- Tests: `TestPipelinePreservesIdentityInvalidFromAcquisition` (+ no snapshot-load/audit/cache/compiler past the refusal);
  `TestPipelineCollapsesNonTransportPlanIdentityInvalid` (frozen SSH-wrapper text still collapses byte-identically, via the
  legacy snapshot branch); `TestWithDraftRemediationTable` positive row + restored stbg4d pin
  `build_repository_identity_invalid: wrong identity` + real lane pin `skill-a.ssh-cmd: …: SSH requires the exact manager
  wrapper` + `source_unavailable` pin; `TestDraftDocsPinExamples` class list + remedy keyword;
  `TestDraftTransportIdentityInvalidSurfacesThroughCLI` (install --dry-run: true class, no masked class, remediation, zero
  fetches; **status**: exact line `error: skill-a.https-cmd: build_repository_identity_invalid: transport plan endpoint 1
  carries …`, no masked class, remediation, zero fetches); committed `TestReviewProbeLegacyLaneIdentityInvalid` (switch off,
  no policy, `GIT_SSH` unset → both `install --dry-run` and `status` print the frozen
  `error: skill-a.ssh-cmd: build_repository_source_unavailable: exact external source is unavailable`, no
  `identity_invalid`, no guidance); `runDraftInstall` is a pure extraction into `runDraftCLI` (body unchanged, 12 call
  sites untouched). Crossconformance rows `v2-external-build-{port,alias}-refused` assert the exact class through
  `install.Project` and fail on the masked class; no known-gap/bound marker, so they count as driven.
- `docs/troubleshooting.md:529-543` — `### build_repository_identity_invalid` scoped to the transport-plan refusal on the
  resolved lane; remedy wording identical to the table row.

## Independent reruns (bash, `set -o pipefail`, real exit codes; clone `cand`, load 5–8)

| check | rc |
|---|---|
| `go build ./...` | 0 |
| `go vet` buildrepo/install/cmd/curator/crossconformance | 0 |
| `gofmt -l` on the 8 files (empty) | 0 |
| `golangci-lint run ./internal/buildrepo/... ./cmd/curator/... ./internal/crossconformance/...` (0 issues) | 0 |
| `go test -count=1 -p 1 ./internal/buildrepo/` (full package, 94.6 s) | 0 |
| cmd/curator `-run` TestDraftDiagnostics\|TestDraftDocsPin\|TestDraftTransport\|TestReviewProbe\|TestWithDraftRemediationTable\|TestDraftAttemptTransportTable\|TestDraftAttemptClauseShape\|TestEndpointExhaustionCarriesRemediation\|TestDraftRemediationThroughCLI — 11 top-level tests incl. LegacyGolden, ResolvedMatrix, ProviderAdmission, IdentityInvalid CLI (install+status), ReviewProbe (95.9 s) | 0 |
| internal/install `-run 'TestDraftTransport\|TestUserConfigIgnored\|TestGlobalDryRun'` | 0 |
| crossconformance `-run 'TestDraftSourcesSemanticCases$/^v2-external-build-(port\|alias)-refused$'` | subtests 2/2 PASS; parent rc=1 by design (`executed(2) != total(94)`), ratio `2 driven, 0 known-gap, 0 bound, 0 skipped` |
| crossconformance `TestDraftExternalStatusFailsClosedWithoutSource` (status fail-closed half) | 0 |
| **base-tree probe**: `TestReviewProbeLegacyLaneIdentityInvalid` extracted into clone `base` (tree `05b7aee1`, base `6a6e2a14`) | 0 — the frozen lane prints the same `source_unavailable` line before and after; no identity leak, no guidance |

Hosted per-lane ledgers (`test-evidence-<os>`, run 35518913747): every candidate test passes on ubuntu and macOS
(`TestDraftTransportIdentityInvalidSurfacesThroughCLI` 7.6/9.4 s, `TestReviewProbeLegacyLaneIdentityInvalid` 8.0/8.9 s,
both xconf rows, `TestDraftTransportLegacyGolden`); on Windows the CLI rows, the legacy probe and the two xconf rows `skip`
("test transport wrapper is POSIX-only") while `TestPipelinePreservesIdentityInvalidFromAcquisition`,
`TestPipelineCollapsesNonTransportPlanIdentityInvalid`, `TestWithDraftRemediationTable`, `TestDraftDocsPinExamples` pass.
Ratio lines: macOS `93 driven / 0 known-gap / 1 bound / 0 skipped`, ubuntu `92/0/1/1`, windows `42/0/1/51` — unchanged
from the sibling baseline; the two rows were and remain driven. Windows evidence for the install-level class is the
platform-neutral unit rows (stated bound, same shape as the accepted siblings).

## Gate attack (clone `probe`; `git checkout --` between mutants; tree `158987e1…` before and after each)

| mutant | unit (Preserves/Collapses) | xconf rows | CLI row (install+status) | legacy probe | remediation table / docs pin |
|---|---|---|---|---|---|
| M1 re-collapse: the §7 branch returns `source_unavailable` (re-mask at the same site) | **rc=1** | **0/2** | **FAIL** | pass (correctly scoped) | pass |
| M3 class-only preservation (rev2 F2 shape, prefix scope dropped) | **rc=1** (Collapses row) | 2/2 | pass | **FAIL** — frozen lane leaks `…identity_invalid: SSH requires the exact manager wrapper` | pass |
| M6 narrow to `Operation == OperationInstall` | rc=0 (helper bound) | **0/2 killed** | **FAIL killed** | pass | pass |
| M8 prefix const drifted (`transport-plan endpoint`) | **rc=1** | **0/2** | **FAIL** | pass | **table FAIL** |
| M2 re-widen remediation key to the bare class (rev2 F1 shape) | – | – | pass | pass (frozen lane prints no identity text) | **table FAIL**: both restored pins decorated |
| M5 remediation row removed | – | – | **FAIL** (remediation missing) | pass | **table FAIL** |
| M7 docs section removed | – | – | – | – | **docs pin FAIL** |

The re-mask is killed at the production entries (`install.Project` rows, `cli.run` install and status); the rev2 F1 and
F2 shapes are now each killed by a committed row (M2 by the restored stbg4d pins, M3 by the legacy-lane probe through
`cli.run`).

## Finding-by-finding (rev2 → rev3)

- **F1 — closed.** Row keyed on the refusal's static diagnostic; `withDraftRemediation` unchanged; stbg4d pin restored
  verbatim plus a real lane message; the legacy probe is a committed negative row. M2 killed.
- **F2 — closed (scoped, the preferred option).** Preservation is scoped by the §7 prefix, which is strictly narrower than
  lane scoping: the frozen lane is byte-identical (base-tree probe = candidate probe, `TestDraftTransportLegacyGolden`
  green, M3 killed at `cli.run`), and no other draft-lane refusal changes class either. The ruling is recorded in
  results.md "Revision 3"; no CHANGELOG entry is consistent with the "released v1 behaviour retained byte-identical" rule
  (only the opt-in resolved lane changes, as siblings 2wyzde/3v7x6j/2eg8nv did without an entry).
- **F3 — closed.** Docs section + `TestDraftDocsPinExamples` entries; remedy text identical to the row. M7 killed.
- **F4 — closed.** `status` row with the exact class line committed; the shared `printFailures` path is pinned, not assumed.
- **R1 — bounded correctly.** `AdmissionError` is exported but its `err` field is not and there is no exported constructor
  (`admission.go:46-57`), so `drafttransport.go:274-281` cannot use it in one line; those messages lack the §7 prefix and
  still collapse exactly as on base.

## Residuals (record only, not blocking)

- R-a: `docs/troubleshooting.md:358` intro still says the listed classes "appear only on the draft Skillfile lane"; the
  bare class can reach `install`/`status` on the frozen lane through the post-acquire binding invariants
  (`pipeline.go:426-446`, internal-consistency guards). The new section itself is scoped correctly ("on the resolved lane
  only … other diagnostics keep their lane behavior"); a one-clause intro qualifier would remove the ambiguity. Docs
  precision only.
- R-b: the pipeline predicate couples on diagnostic text (via the shared const). A typed refusal marker on
  `AdmissionError` would remove the text coupling; out of scope for this minor leaf and guarded by M8.
- R-c: M6 survives the unit row (helper bound); it dies at both production entries, which is what the AC requires.
- R-d: Windows runs the unit rows only (POSIX test wrapper), same bound as every sibling leaf.
- R-e: LOGBOOK.md not edited (campaign rules); this resource is the record.

## Checklist mapping
- Implementation matches AC — yes (true class end to end at `install.Project`, `install`, `status`; exact-class xconf rows
  driven; closed sanitized remediation row; frozen lane byte-identical).
- Solution fits project architecture — yes (root-cause placement in `RunPipeline`, presentation-layer row scoped to the
  refusal it describes, one shared const).
- Tests green — yes (independent narrow reruns above; hosted gate green on tree `158987e1…`).
- Verdict — `accept_cr(TASK-260920-1sbj7o, revision=3, evidence=TASK-260920-1sbj7o_review-verdict-rev3.md)`.
