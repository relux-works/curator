# TASK-261002-2ipeqa — accept-rc14-candidate-suite

## Revision 3

Ready for review under the binding revision-3 rework brief. Full hosted rc.14 matrix on this revision remains pending for the orchestrator's re-dispatch; no green full candidate matrix is claimed.

Base: `64345d71aa539ef71b9f0b4d3fb9b28d1c9d3fb7`. Frozen uncommitted tree: `9706982aa19b64fa55b95123e68a34b3640818ca`; 17 changed paths. Candidate core: curator-spec `e3a88cedba7a844c594348ee89654471db086f98`, manifest `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`; source suite remains `061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`.

### Changes and findings

Added exactly one new known-gap row for rc.14: `snapshot-acquisition/cases / byte-exact-snapshot`, owner **TASK-261002-1foyf3 — rc14-pin-and-v2-writer-cutover**. Exact reason: "rc.14 expects curator-content-v2 writes; writer enabled at the rc.14 pin cut-over". Total rc.14 gap rows **10**: the nine existing posture/provisioning gaps plus this one. None of the seven fixed marker/Windows cases returns to the ledger. Snapshot published count stays **1**; the two autocrlf variants remain subchecks of that one case, not two published cases.

The existing production-entry test `TestConformanceSnapshotAcquisition` still drives `gitops.Extract` and validates exact paths, bytes, line endings and unexpanded placeholders under autocrlf true/false. Its hash comparison now returns the measured mismatch through the existing `RunOutcomes → Check` accounting path. Adding a ledger row alone would leave its former `t.Fatalf` subtests red. Fixture, extraction and byte-preservation errors still call `t.Fatal`/`t.Fatalf`, so a known-gap row does not suppress those failed tests. No product writer change.

Independent inspection of **all 3 OS candidate JSON streams** from hosted run [37005296654](https://github.com/relux-works/curator/actions/runs/37005296654) found the expected snapshot hash mismatch **and four further missing lifecycle pins**, contrary to the brief's "only this one" summary. Those four errors are count accounting, not other v2-write dependencies. The corpus publishes `mixed_build_cases=6`, `path_shim_cases=3`, `signing_cases=4`, `transaction_cases=4`; their production consumers are the four `TestAuthoritative*CasesUseProjectInstallEntry` tests in `internal/install/external_lifecycle_conformance_test.go` (ProjectInstall). Added the missing count rows for rc.14 and both older candidates, which publish the same cases. rc.13 count rows remain byte-for-byte unchanged. Existing older-candidate rows remain unchanged, with four added rows each; no other new gaps.

Authenticated/recounted **107/107 configured rc.14 family pins / 1,887 case entries**, correcting the previous **103 / 1,870** inventory that omitted those 17 lifecycle entries. This ratio is against the configured published-family pins, not a claim to enumerate every normative clause. `count-audit.json` contains exact family counts. The rc.14 and historical Muse count maps still match. Historical hash-v2 and Muse manifest/lifecycle identities were checked and their scoped policy regressions rerun green.

Writer **OFF** (`EnableV2Writers=false`), `.github/workflows/ci.yml` SPEC_PIN remains `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`, and LOGBOOK.md is unchanged against base. One CHANGELOG line retains the trunk fix entries. Findings are in board notes and outcomes under the binding prohibition on LOGBOOK edits.

### Fresh execution on revision 3

Each Go command ran directly as a standalone process without tee or a pipeline, with `GOFLAGS=-work` and owned shared `~/.mini-build-lock`; every WORK directory was retained. Before/after launchctl observations show syspolicyd **running**, successive crashes **369 → 369** on every command below. No crash/backoff event or local stall occurred in this revision.

| Run | Exact argv | Real exit | Seconds | Crashes before → after |
| --- | --- | ---: | ---: | --- |
| candidate-targeted | `go test ./internal/conformancecoverage ./cmd/curator -run Conformance|Coverage -count=1` | 0 | 14.751 | 369 → 369 |
| candidate-snapshot-and-install | `go test -count=1 -json ./internal/interop/environments ./internal/install -run TestConformanceSnapshotAcquisition|TestAuthoritative(MixedBuild|PathShim|Signing|Transaction)CasesUseProjectInstallEntry` | 0 | 65.858 | 369 → 369 |
| default-pin-targeted | `go test ./internal/conformancecoverage ./cmd/curator -run Conformance|Coverage -count=1` | 0 | 13.754 | 369 → 369 |
| spec-implementations | `go test -count=1 -json ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy` | 0 | 25.543 | 369 → 369 |
| narrow-snapshot-gap | `go test ./internal/conformancecoverage ./internal/interop/environments -run TestRC14CoverageSnapshotWriteGapIsExactAndSuiteScoped|TestConformanceSnapshotAcquisition -count=1` | 1 | 1.646 | 369 → 369 |
| narrow-lifecycle-count | `go test ./internal/conformancecoverage -run TestCoverageLifecycleCountPinsCoverPublishedInstallCases -count=1` | 1 | 0.757 | 369 → 369 |
| restored-regressions | `go test ./internal/conformancecoverage ./internal/interop/environments -run Coverage|TestConformanceSnapshotAcquisition -count=1` | 0 | 1.607 | 369 → 369 |
| build | `go build -o /tmp/TASK-261002-2ipeqa-revision3/curator-bin ./cmd/curator` | 0 | 1.772 | 369 → 369 |
| historical-hash-v2-policy | `go test ./internal/conformancecoverage -run CoverageCandidateDispatch|RC14Coverage|CoverageLifecycle -count=1` | 0 | 0.734 | 369 → 369 |
| historical-muse-policy | `go test ./internal/conformancecoverage -run CoverageCandidateDispatch|RC14Coverage|CoverageLifecycle -count=1` | 0 | 0.698 | 369 → 369 |

`candidate-targeted`, `candidate-snapshot-and-install`, `spec-implementations`, the mutants and restored regressions used `/tmp/TASK-261002-2ipeqa-spec-rc14/conformance/v1`. `default-pin-targeted` used the authenticated SPEC_PIN checkout `/tmp/TASK-261002-2ipeqa-spec-rc13/conformance/v1`. The two historical policy runs used their own manifest identities. Raw stdout/stderr and process exits/durations are attached in the revision-3 archive.

The exact candidate spec workflow command **ran here**, exit **0**: `go test -count=1 -json ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy`. Its real JSON stream fed `python3 /tmp/TASK-261002-2ipeqa-spec-rc14/tools/implementation_coverage.py go --stream /tmp/TASK-261002-2ipeqa-revision3/spec-implementations.log`, exit **0**, **8/8 claims**. The corresponding `families --implementation go --root .../conformance/v1` presence gate also exited **0**, **8/8 claims**.

Additional direct checks: independent count audit **exit 0** (107/107); corrected three-OS hosted failure audit **exit 0**; compatibility/protected-file checks **exit 0**; `git diff --check` **exit 0**; gofmt check **exit 0**, no output. An initial exploratory hosted audit exited **1** because its assertion expected only the snapshot package; that failure exposed the four missing install pins and was corrected without changing the downloaded evidence. It is not a passing audit.

Measured scoped production-entry coverage: snapshot **1/1** classified as known gap with both autocrlf settings executed; lifecycle **17/17** classified: 16 driven, 1 platform bound. The signing `platform-requires-local-signing` row remains the existing bound, never called a production pass. Raw stream tallies and failure audit enumerate every observed case/package. Historical policy checks do not claim full historical production-suite replay.

### Review rejection response and negative evidence

Read and answered `TASK-261002-2ipeqa_review-verdict-rev2.md`: a push-triggered default hosted green cannot establish candidate validation. The orchestrator's explicit candidate dispatch on rev2 ran and exposed the snapshot/count issues above. This revision fixes those issues and reruns the exact six-package Implementations command plus its real-stream consumption gate locally. The **new full hosted candidate dispatch remains required**, and is delegated to the orchestrator by the binding rework-3 brief; this handoff does not reinterpret default CI as candidate CI.

Named regression `TestRC14CoverageSnapshotWriteGapIsExactAndSuiteScoped` asserts the exact case, digest scope, owner/reason, known-gap tally, stale passing-gap refusal, and refusal of another failed snapshot. Narrowing its ledger case from `byte-exact-snapshot` to `byte-exact-snapshot/autocrlf=true` preserves a row but wrongly covers only a subcheck: the named regression and the **real `gitops.Extract` consumer** both fail with real **exit 1**. This expected-red result is a failing gate, proving the exact scope.

Named regression `TestCoverageLifecycleCountPinsCoverPublishedInstallCases` derives all four family sizes from the selected corpus and checks their pins. Narrowing rc.14 `signing_cases` from **4 to 3** fails it with real **exit 1**. Both mutations were restored from byte copies, all 17 frozen file hashes rechecked, and `restored-regressions` then passed **exit 0**. Neither mutant is presented as passing validation.

### Explicit limits and handoff routing

Fresh local lint and the full hosted candidate/default/platform/race matrices were **not run** on revision 3. Fresh local lint is routed to the hosted CR validation gate per the binding hosted-arbiter decision. Local build **did run**, exit 0 (table above); revision-2's missing build is historical, not this revision's result. The exact `go test ./internal/marker ./internal/scriptworker -count=1` was **not rerun**: its initial and one permitted retry already stalled on revision 2; the binding brief prohibits further retries. Those earlier wrappers exited **130**, underlying Go numeric exits/durations remain **unknown**, and their logs show a ten-minute timeout and crash counts 366→367→368. Revision 3 does independently rerun the marker package within the exact six-package Implementations command, exit 0; this does not establish the unrun full scriptworker package.

No prior revision's green is reused as validation of this changed tree. Hosted run 37005296654 is **failure** and is used only to identify the observed candidate defects and absence of other observed v2-write mismatches within its 3 OS streams. It is not a revision-3 pass. The two generic injected checklist entries for local "Lint clean" and logbook writing were removed through the CLI because the binding decision routes lint to hosted validation and prohibits LOGBOOK edits; the existing explicit routing/evidence items remain. Relevant fresh scoped tests and build are green; full hosted acceptance remains pending.
