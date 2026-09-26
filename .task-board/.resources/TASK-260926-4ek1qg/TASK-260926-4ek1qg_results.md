# TASK-260926-4ek1qg — qualification results

## Changes

- Pinned the Implementations Go checkout to curator `0a62862130fd6cbd9db8b246da63fdb5aecdfaaf`. The workflow comment explains the lockstep with PR #88's `executable_identity_cases` and `hard_link_substitution_definition`.
- Configured non-cone sparse checkout with `/*` and `!/.task-board/`. The Go suite retains its implementation sources and declared submodule while omitting the deep board paths that caused the Windows checkout failure.
- Updated the Go coverage ledger and added a regression test for the new production consumer mapping row and replacement preflight case.
- Added the unreleased changelog entry.

## Go coverage declarations: old → curator 0a628621

| Previous Go declaration | Declaration at 0a628621 |
| --- | --- |
| `internal/skillspec.TestReleasedSchemaCases` | same |
| `internal/marker.TestReadAuthoritativeMarkerV4SchemaCases` | same |
| `internal/moduleroots.TestModuleRootVectors` | same |
| `internal/scriptpolicy.TestScriptHostExecutionPolicySectionsAreAllClassified` | same |
| `internal/scriptpolicy.TestScriptExecutionPolicyIdentityMatchesTheSuite` | same |
| `internal/scriptpolicy.TestScriptExecutionOptInCases` | same |
| `internal/scriptpolicy.TestARefusalPrecedesEveryWorkerSurface` | `internal/scriptpolicy.TestPreflightRefusalCases` |
| added for the PR #88 fields | `internal/scriptpolicy.TestScriptHostExecutionPolicyProductionConsumersCoverAllCases` |

The replacement preflight test covers the two mandatory-control-unavailable refusal shapes before worker startup and verifies that the three non-refusal shapes succeed. The new consumer-mapping test covers the published behavioral families, including executable identity cases, and requires the hard-link definition to be present.

## Evidence

The disposable curator clone was checked out at `0a62862130fd6cbd9db8b246da63fdb5aecdfaaf`. Each declared test was present in `go test -list .` output:

- `go test -list . ./internal/skillspec` — exit 0
- `go test -list . ./internal/marker` — exit 0
- `go test -list . ./internal/moduleroots` — exit 0
- `go test -list . ./internal/scriptpolicy` — exit 0

The disposable specification candidate was based on main `5746367`, cherry-picked PR #88 commit `b202b5d` (resulting candidate commit `9a00187`), and overlaid with the task changes. Against that candidate and the curator clone at 0a628621, the Implementations Go package step was split into bounded standalone runs:

- `go test -count=1 -json ./internal/interop` — exit 0
- `go test -count=1 -json ./internal/closure` — exit 0
- `go test -count=1 -json ./internal/skillspec` — exit 0
- `go test -count=1 -json ./internal/marker` — exit 0
- `go test -count=1 -json ./internal/moduleroots` — exit 0
- `go test -count=1 -json ./internal/scriptpolicy` — exit 0

The six JSON streams were combined, then checked against the candidate ledger:

- `python3 tools/implementation_coverage.py go --stream <combined-go-test.json>` — exit 0; 8 declared Go claims upheld.
- `python3 tools/implementation_coverage.py families --root conformance/v1` — exit 0; 19 declared claims served.
- `python -B -m unittest discover -s tools -p 'test_implementation_coverage.py'` — exit 0; 37 tests passed.
- `python tools/validate.py` — exit 0; 64 schemas and 1,169 vector files validated. The first attempt with the system interpreter exited 1 because `jsonschema` was unavailable; rerun passed in a temporary venv populated from `requirements-dev.txt`.
- `go test ./tools/...` — exit 0.
- `gofmt -l tools` — exit 0 with no files listed; `git diff --check` — exit 0; Ruby YAML parse of `.github/workflows/implementations.yml` — exit 0.
- The exact sparse-checkout patterns were applied to the disposable curator clone with Git's non-cone mode; the sparse setup and `test ! -e .task-board` each exited 0.

## Limits recorded

- The hosted Windows lane was unavailable in this run. The Git sparse-checkout behavior and exclusion were verified locally on macOS; Windows remains unverified pending hosted CI.
- Broad `python -B -m unittest discover -s tools -p 'test_*.py'` was stopped after about nine minutes to stay within the ten-minute command bound (exit 130) while `test_release_gate.py` setup recursively copied the full tree. No assertion failure had been reported before interruption. The focused 37-test coverage suite and schema/vector validator passed independently.
- No Python linter is configured in this repository; the repository formatting checks, changed-file whitespace check, workflow YAML parse, focused tests, validator, and Go tool tests passed.

## Revision 2 — partial-client lane

Applied the operator's option 1. Curator `0a62862130fd6cbd9db8b246da63fdb5aecdfaaf` remains the Go pin and consumes the full candidate core. Cocoaskills is pinned to current main `4a88aa0e47ab2b34ded16162cec7f97d0c7cf6f1`; its declared core support is `v1.0.0-rc.10`, plus the candidate `skillfile-sources-v1` corpus and schemas. Its workflow comment and the implementation-coverage ledger now state those scopes.

The Python core checkout is `relux-works/curator-spec@v1.0.0-rc.10` at `implementations/manager-core` (tag object `68a2ce966577c6e53f41c277909a6ddee6874723`). The manager schema-8 step overrides `CURATOR_CONFORMANCE_ROOT` to that checkout's `conformance/v1` and sets `CSK_REQUIRE_FULL_CANDIDATE_ROOT=1`. The draft step sets `CSK_DRAFT_SOURCES_SUITE_ROOT` to the candidate's `conformance/skillfile-sources-v1`; cocoaskills derives `schemas/v1` and `schemas/skillfile-sources-v1` from two parents above that path, so the suite and both schema roots are candidate bytes. The manager's draft suite is POSIX-only in its own CI, so this step is conditioned off on Windows. The job-level `CURATOR_CONFORMANCE_ROOT` remains the candidate `conformance/v1` for Go and the registry service.

`implementation_coverage.py families` now accepts `--implementation`; the workflow checks curator's rows against the candidate core and cocoaskills' rows against rc.10. The Python ledger comment declares the same boundary. The 11 existing manager core rows are all served by rc.10; the draft corpus runs in its dedicated conformance step.

### Refusal claim clarification

At curator `0a628621`, `TestPreflightRefusalCases` covers the two mandatory-control-unavailable install/invocation refusals before worker startup and verifies that three other preflight shapes succeed and start a worker. This is deliberately narrower than the removed `TestARefusalPrecedesEveryWorkerSurface` claim: the new suite defines those three shapes as successes, so the old “every enforced shape refuses” statement is no longer true. `TestScriptHostExecutionPolicyProductionConsumersCoverAllCases` adds source-derived mapping for all published behavioral cases, including executable identity and the hard-link definition, but it is a consumer-mapping assertion rather than a broader refusal assertion.

### Old-to-new Go declarations

| Declaration at curator `a3abcf34` | Declaration at curator `0a628621` |
| --- | --- |
| `internal/skillspec.TestReleasedSchemaCases` | unchanged |
| `internal/marker.TestReadAuthoritativeMarkerV4SchemaCases` | unchanged |
| `internal/moduleroots.TestModuleRootVectors` | unchanged |
| `internal/scriptpolicy.TestScriptHostExecutionPolicySectionsAreAllClassified` | unchanged; now classifies the PR #88 sections |
| `internal/scriptpolicy.TestScriptExecutionPolicyIdentityMatchesTheSuite` | unchanged |
| `internal/scriptpolicy.TestScriptExecutionOptInCases` | unchanged |
| `internal/scriptpolicy.TestARefusalPrecedesEveryWorkerSurface` | retired; replaced by the narrower `internal/scriptpolicy.TestPreflightRefusalCases` |
| new PR #88 behavioral-consumer coverage | `internal/scriptpolicy.TestScriptHostExecutionPolicyProductionConsumersCoverAllCases` |

### Validation on the disposable PR #88 candidate

The workspace under `/tmp/TASK-260926-4ek1qg-validation/workspace` was assembled from curator-spec content at `b202b5d` plus the final task files, curator at `0a628621`, cocoaskills at `4a88aa0e`, and registry at `d690bea6fab1c8e6392e05d3a3cdfcf1168bc914`. Each gate command below ran directly; JSON/JUnit evidence stayed under `/tmp` or the disposable checkout.

- `go test -list . ./internal/skillspec` — exit 0; `TestReleasedSchemaCases` listed.
- `go test -list . ./internal/marker` — exit 0; `TestReadAuthoritativeMarkerV4SchemaCases` listed.
- `go test -list . ./internal/moduleroots` — exit 0; `TestModuleRootVectors` listed.
- `go test -list . ./internal/scriptpolicy` — exit 0; all five declared script-policy cases at 0a listed, including `TestPreflightRefusalCases` and `TestScriptHostExecutionPolicyProductionConsumersCoverAllCases`; the removed broad test was absent.
- Bounded `go test -count=1 -json` runs for `./internal/interop`, `./internal/closure`, `./internal/skillspec`, `./internal/marker`, `./internal/moduleroots`, and `./internal/scriptpolicy` — exit 0 for each.
- `python tools/implementation_coverage.py go --stream <combined-go-test.json>` — exit 0; all 8 declared Go claims observed passing.
- `python tools/implementation_coverage.py families --implementation go --root conformance/v1` — exit 0; all 8 Go artefact claims served by the #88 candidate.
- `python tools/implementation_coverage.py families --implementation manager --root implementations/manager-core/conformance/v1` — exit 0; all 11 manager core artefact claims served by rc.10.
- `python implementations/manager/.github/scripts/candidate_consumption.py require --root implementations/manager-core/conformance/v1` — exit 0.
- With `CURATOR_CONFORMANCE_ROOT=.../implementations/manager-core/conformance/v1` and `CSK_REQUIRE_FULL_CANDIDATE_ROOT=1`, `pytest -q tests/test_schema8_candidate_conformance.py --junitxml=manager-schema8-results.xml` — exit 0; 183 passed. Its `implementation_coverage.py pytest` gate — exit 0; all 11 manager rows observed passing.
- With `CSK_DRAFT_SOURCES_SUITE_ROOT=.../conformance/skillfile-sources-v1`, `pytest -q tests/test_draft_sources_conformance.py --junitxml=manager-skillfile-sources-results.xml` — exit 0; 328 passed and 1 intentional skip in `test_draft_sources_setup_phase_skip_probe` (that probe verifies setup-time skips are recorded in JUnit). Pytest emitted one `record_property`/`xunit2` warning.
- The workflow's Python manager lifecycle selection — exit 0; 11 passed.
- `python tools/run_pytest_no_skips.py -q implementations/registry/tests/test_protocol_conformance.py` — exit 0; 80 passed.
- Final candidate `python tools/test_implementation_coverage.py` — exit 0; 39 tests passed. `python3.12 -m compileall -q tools/implementation_coverage.py tools/test_implementation_coverage.py` — exit 0.
- No repository lint configuration/tool was found. Ruff 0.16.9's isolated `E4,E7,E9,F` check on the two changed Python files — exit 0. A broader, unconfigured default `ruff check` probe — exit 1 on pre-existing file-mode/import-style findings in these files (`EXE001`, `I001`, `PLR0402`, `UP035`); it was not used as the project gate.
- Ruby YAML parse of `.github/workflows/implementations.yml` and `git diff --check` — exit 0.

### Regression and narrowing evidence

The named test `LedgerShapeTests.test_cocoaskills_partial_client_lane_uses_rc10_and_candidate_draft_sources` passes on the final candidate (exit 0). A narrowing mutant changed `CSK_DRAFT_SOURCES_SUITE_ROOT` from candidate `conformance/skillfile-sources-v1` to candidate `conformance/v1`; the named regression failed (exit 1). The workflow was restored from a saved copy, `cmp -s` confirmed byte equality (exit 0), and the test passed again (exit 0).

As a direct boundary reproduction, running cocoaskills' `test_script_policy_sections_are_all_classified` against the full #88 candidate core instead of rc.10 exited 1 and named the unclassified `executable_identity_cases` and `hard_link_substitution_definition`. This is the failure the partial-client scope prevents; the rc.10 lane passed all 183 schema-8 tests.

### Limits and rerun boundary

The hosted Windows checkout was not available here. The sparse pattern excluding `.task-board/` remains unchanged from revision 1, whose local Git sparse-checkout setup and board-path absence checks passed; the actual `windows-latest` checkout still requires hosted CI verification. Manager core/draft and the Go package runs were rerun locally against the disposable candidate as described above. No hosted all-OS workflow result is claimed.
