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


---

# TASK-261002-2ipeqa — accept-rc14-candidate-suite

## Revision 2

**Ready for review under the binding ~10:40Z decision: the hosted CR validation gate is the arbiter. Local stalls are recorded below and are not grounds to block this handoff.**

Base: `64345d71aa539ef71b9f0b4d3fb9b28d1c9d3fb7`. Frozen changed-file identity: `2bdd5a194aa9ada2eede06b7b180b4319aaa3acdab229b9f026f26953ccb8cf1`. Uncommitted in the assigned Story worktree.

Candidate rc.14 commit `e3a88cedba7a844c594348ee89654471db086f98`, core manifest `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`, unchanged source manifest `061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`.

The independent audit authenticates both corpora and recounts **103/103 family pins / 1,870 case entries**, including all schema-index counts, vector groups, fixture groups and five CRC32 semantic batches. Every rc.14 pin matches the measured count; earlier digest counts remain unchanged. The core suite still differs from historical Muse only in protocol-version metadata.

rc.14 retains **nine posture/provisioning gaps**. Removed the five marker cross-field cases in each of marker v3 and v4 (ten ledger rows), and both Windows hard-link rows. This trunk still had four stale Windows rows under the earlier hash-v2/Muse digests; those were removed as well to preserve their acceptance. The new regression refuses these stale marker-v3/v4 and Windows rows across all accepted digests. Historical Muse hash-v2 rows remain historical. The CHANGELOG retains both trunk fix entries and one rc.14 entry.

Writer stays **OFF** (`EnableV2Writers = false`); SPEC_PIN remains rc.13 commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`; both protected files and LOGBOOK.md are byte-identical to this trunk. Findings live in board notes/outcomes. The remaining generic logbook checklist item was removed through the board CLI because the binding host rules prohibit logbook edits; the board/outcome checklist remains.

Every validation below ran directly as a standalone child process, without tee or a pipe. Go builds/tests use **GOFLAGS=-work**, retaining WORK directories and holding their own ~/.mini-build-lock ownership. Presence/consumption checks refer to the candidate spec's exact Go workflow shape. No full-module, race, cross-platform, Python manager or registry pass is claimed. Revision 1 evidence below is historical only and is not reused to validate this changed source tree.


### Revision 2 hosted-gate handoff update

The source tree remains exactly the previously validated Revision 2 tree: all 16 changed-file SHA256 hashes and base match `source-identity.json` (identity verification exit **0**). No product code was changed in this handoff run. This run independently re-authenticated/recounted the candidate corpus: **103/103 configured families, 1,870 entries**, count audit exit **0** (0.187s). Presence gate rerun exit **0** for all eight declared Go consumption claims. Whitespace, protected-file comparison and formatting reruns each exit **0** with no findings. Writer remains OFF; SPEC_PIN and LOGBOOK.md are unchanged.

Accepted from already-attached Revision 2 evidence, rather than rerun here: candidate coverage command exit **0** (22.436s) and default-pin command exit **0** (16.202s), both with GOFLAGS=-work, shared-lock ownership, and syspolicyd running / crashes **366 → 366**. Previously recorded marker/scriptworker attempts remain failing/interrupted: wrapper exits **130**, underlying Go numeric exits and exact durations **unknown**, first raw log contains a ten-minute timeout. Recovery crash counts were **366 → 367 → 368**. This session observed syspolicyd running at **368**, then performed only the non-Go checks above. No additional long build/test was started, so no new build/test crash-count pair exists.

Unrun on Revision 2: the exact six-package Implementations Go command and consumption gate, fresh lint/build, full candidate matrix, fresh historical policy/production reruns and narrowing probes. Revision 1 evidence stays historical. Local lint/build remain unverified and are routed to hosted validation as explained in the checklist alignment below. The required hosted CR gate is responsible for validation before acceptance. It is initiated by the board validation/review flow, not by this developer run.

Exact pending spec Go command: `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-261002-2ipeqa-spec-rc14/conformance/v1 GOFLAGS=-work go test -count=1 -json ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy`, followed by `python3 <candidate-spec>/tools/implementation_coverage.py go --stream <captured-json>`. The presence half passed independently. No Python-manager or registry conformance is claimed by this Curator task.

Findings are saved in board notes and task-scoped outcomes because the binding host rules prohibit LOGBOOK.md edits. The later hosted-gate decision supersedes the earlier recovery recommendation; no human decision is required to route this revision to review.

## Revision 2 anomalies and recovery

After backoff and shared-lock acquisition, the exact marker/scriptworker rerun again stalled host process launches. It was interrupted during the Go wait with **wrapper exit 130**; Go numeric exit remains **unknown**. The second diagnostic shell was also interrupted with exit 130. The assumptions that preserving WORK directories and waiting five minutes would enable this test did not hold on this host. No compensating product-code or test workaround was introduced.

**Missing validation:** a green marker/scriptworker run, the exact six-package spec Implementations rerun and its consumption gate, fresh lint, and fresh build. Earlier-candidate policy/production reruns and fresh narrowing probes were not run. Candidate/default required checks and the independent 103/103 audit are green as recorded, but they cannot substitute for the missing commands. Local lint/build remain unverified and are routed to hosted validation as explained in the checklist alignment below.

**Binding decision (~10:40Z):** hand off Revision 2 for the hosted CR gate (`scripts/remote-gate.sh`). Do not retry a stalled local command more than once, and do not block for the host crash-loop. The marker/scriptworker command already stalled on both the initial attempt and its one retry, so no further Go attempts were made in this handoff run. Hosted validation remains pending; no hosted pass is claimed.

The first marker/scriptworker attempt produced marker PASS and a scriptworker **ten-minute timeout** in TestRunShimToleratesUnusableDiagnosticsDir. Its wrapper then blocked in the post-test launchctl observation and was interrupted with **wrapper exit 130**. The Go numeric exit was not persisted and remains **unknown**; this is failing/interrupted evidence, never a pass. The count-audit shell also stalled before its preflight and was interrupted with exit 130. Two diagnostic ps probes exited 130 on interruption; two interactive shell probes exited 1 on interruption. An eventual bash echo probe returned exit 0 when host execution recovered.

Syspolicyd recovered at about 09:48 UTC, running with successive crashes **367**, previously **366**. More than five minutes of backoff elapsed before the next attempted Go check at 09:53:21. Shared-lock refusals returned **99 without starting Go**; the lock was held by another live validation process. The result recorder now saves the child's exit code before attempting a bounded post-run host observation, preventing a future host stall from losing it.

Two exploratory independent count-audit attempts exited **1** because the audit script initially missed build-drivers' argv field and treated schema_version as a content-hash vector. Correcting only the audit field mapping yielded exit 0 and exact counts for 103/103 families. No pin, product implementation or corpus bytes were changed to compensate.

## Revision 2 recorded command results

| Attempt | Exact argv | Exit | Seconds | syspolicyd before → after | Crashes before → after |
| --- | --- | ---: | ---: | --- | --- |
| corpus-audit | `python3 /tmp/TASK-261002-2ipeqa-revision2/audit.py` | 0 | 0.47 | running → running | 366 → 366 |
| candidate-targeted | `go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance\|Coverage' -count=1` | 0 | 22.436 | running → running | 366 → 366 |
| default-pin-targeted | `go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance\|Coverage' -count=1` | 0 | 16.202 | running → running | 366 → 366 |
| count-audit-recovered | `python3 /tmp/TASK-261002-2ipeqa-revision2/count-audit.py` | 1 | 0.445 | running → running | 367 → 367 |
| count-audit-corrected | `python3 /tmp/TASK-261002-2ipeqa-revision2/count-audit.py` | 1 | 0.333 | running → running | 367 → 367 |
| count-audit-green | `python3 /tmp/TASK-261002-2ipeqa-revision2/count-audit.py` | 0 | 0.152 | running → running | 367 → 367 |
| spec-presence | `python3 /tmp/TASK-261002-2ipeqa-spec-rc14/tools/implementation_coverage.py families --implementation go --root /tmp/TASK-261002-2ipeqa-spec-rc14/conformance/v1` | 0 | 0.102 | running → running | 367 → 367 |
| marker-scriptworker-host-interrupted | `go test ./internal/marker ./internal/scriptworker -count=1` | 130 (wrapper; Go unknown) | None | running → unknown | 366 → None |
| count-audit-host-interrupted | `'not started / interrupted shell'` | 130 | None | unknown → unknown | unknown → unknown |
| whitespace | `git diff --check` | 0 | 0.067 | running → running | 367 → 367 |
| protected-files | `git diff --exit-code -- LOGBOOK.md internal/hashing/hashing.go .github/workflows/ci.yml` | 0 | 0.033 | running → running | 367 → 367 |
| formatting | `gofmt -l internal/conformancecoverage/rc14_test.go internal/conformancecoverage/coverage.go internal/conformancecoverage/content_hash_v2_gaps_test.go internal/config/environments_conformance_test.go internal/registry/schema_conformance_test.go internal/envmarker/marker_env_schema_test.go internal/marker/schema_coverage_test.go internal/contextlock/schema_conformance_test.go internal/envprofile/muse_test.go internal/envprofile/read_failure_conformance_test.go cmd/curator/muse_test.go internal/buildrepo/acquisition_conformance_test.go internal/scriptpolicy/conformance_test.go` | 0 | 0.025 | running → running | 367 → 367 |
| marker-scriptworker-second-host-interruption | `go test ./internal/marker ./internal/scriptworker -count=1` | 130 (wrapper; Go unknown) | None | running → unknown | 367 → None |
| second-diagnostic-shell-interruption | `'not started / interrupted shell'` | 130 | None | unknown → unknown | unknown → unknown |

## Revision 1 (historical)

Ready for review. Repository changes are uncommitted in the assigned Story worktree.

Base: `20368da65ffe8516e99825b4bff6604a71a4f462`. Frozen changed-file identity: `343ca1555656213807ed621119d49e13950d66db38d13977dc5fef07b3384d9b`; see `source-identity.json` for all 16 file digests.

Candidate commit: `e3a88cedba7a844c594348ee89654471db086f98` (candidate-rc14).
Core manifest: `6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5`.
Source manifest: `061ec05ddb1746b72168d157f81d2930047cd62371ffa46cf585844da61ee6be`.
Candidate tree: `244e26aed2d1281fa148660f94448b65efa94d2c9ce1ca8bc64c25725136f4a8`.

All 1,294 published manifest entries authenticate. The 1,295-file tree also includes manifest.json. Its manifest path set equals the historical Muse corpus; every changed published document authenticates against the historical digest and differs only in protocol_version metadata. The unchanged executable corpus supports the Muse count template: **103 families / 1,870 pinned case entries**. Source corpus unchanged.

Added the rc.14 digest without changing any earlier digest's count or gap rows. Implemented hash-v2 consumers and Muse consumers now dispatch to rc.14; historical Muse hash-v2 gap accounting remains historical. Acquisition and read-failure checks enforce the selected suite's exact release label; script-worker label acceptance includes rc.14 and still rejects rc.15 and empty labels. Added coverage-policy and unknown/near-match digest regressions, plus the existing production-entry consumers now execute against rc.14. One CHANGELOG line.

**Writer stays OFF**: internal/hashing/hashing.go is byte-identical to base and EnableV2Writers remains false. **SPEC_PIN unchanged**: `.github/workflows/ci.yml` is byte-identical to base and remains `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` (rc.13). LOGBOOK.md unchanged.

Gap ledger derived from execution, rather than assumed from the template: **21 rows** = 16 current implementation gaps plus five existing rc.13 marker-v3 gaps owned by BUG-260923-2afgyq. The actual marker-v3 consumer measured **23 driven + 5 known-gap / 28**. Copying only the 16-row hash-v2 template initially failed the exact Implementations run; copying all five existing marker-v3 rows then passed and verified none is stale. No historical hash-v2 gaps were copied into rc.14.

## Accepted validation

All commands below ran directly as standalone child processes, without tee or a pipe. All Go builds/tests used GOFLAGS=-work and retained their WORK directories. Commands and real exit codes for every recorded attempt are in the table and runs.jsonl; JSON streams and stderr streams are attached.

- Required candidate command: `CURATOR_CONFORMANCE_ROOT=/tmp/TASK-261002-2ipeqa-spec-rc14/conformance/v1 go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance|Coverage' -count=1` — **exit 0**, accepted locked rerun `candidate-targeted-locked`.
- Required default-pin command: same command with `/tmp/TASK-261002-2ipeqa-spec-rc13/conformance/v1` — **exit 0**, accepted `default-pin-targeted-locked`. The checkout resolves to the committed SPEC_PIN.
- Exact Go manager command from candidate curator-spec `.github/workflows/implementations.yml`: `go test -count=1 -json ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy`, with rc.14 root — **exit 0**, `spec-implementations-green`.
- Spec workflow presence gate (`implementation_coverage.py families --implementation go --root ...`) and real-stream consumption gate (`implementation_coverage.py go --stream ...`) — **exit 0 each, 8/8 claims upheld**.
- Scoped reader/acquisition suite — **exit 0**, accepted `candidate-readers-locked`.
- Full conformancecoverage + hashing package tests — **exit 0**, `candidate-hash-vectors`; includes writer-OFF regression and all five executable vectors.
- Muse link states and registry/layout — **exit 0**, `candidate-muse-profile`; fragment schema/production cases — **exit 0**, `candidate-muse-fragments`; profile read failures — **exit 0**, `candidate-read-failures`.
- Historical hash-v2 and historical Muse policy selection/dispatch tests — **exit 0 each**. This is policy compatibility evidence, not a full historical-suite replay. All earlier count/gap rows were also compared byte-for-byte to base and are unchanged.
- `go build -o /tmp/TASK-261002-2ipeqa-curator-bin ./cmd/curator` — **exit 0**.
- `golangci-lint run --timeout 8m` with GOLANGCI_LINT_CACHE=/tmp/TASK-261002-2ipeqa-lint-cache and GOCACHE=/tmp/TASK-261002-2ipeqa-go-cache — **exit 0, 0 issues**.
- Formatting and whitespace validation — **exit 0**, no gofmt output and no diff errors.

## Negative evidence and bounds

Both narrowing probes failed with **exit 1**, as required. Narrowing implemented hash-v2 admission back to the earlier digest failed TestCoverageCandidateDispatch/rc14; narrowing rc.14's marker-v3 count from 29 to 28 failed TestRC14CoveragePolicyRetainsExactCountsAndCurrentGaps. Both files were restored from byte copies and all 16 changed-file digests were verified against the frozen identity. The candidate and default-pin tests subsequently passed under the shared build lock. Unknown, near-match, and self-labelled rc.14 manifests are rejected by the coverage tests.

Scoped JSON streams measured **39/103 configured family pins**. Specifically: implemented hash-v2 families **81/81 cases**, Muse link states **16/16**, fragments **36/36**, profile read failures **37/37**, spec implementation claims **8/8**. Fragment evidence comprises **4/36 production Resolve + CLI emissions** and **32/36 static schema/reader oracles**. The profile vector contains two additional restore rows whose count is checked by this test but whose restore entry points are outside this profile test. Do not interpret the published “driven” label as a production-emission claim for those static rows. See coverage-summary.json for measured tallies and stream references.

No claim of a full-module, race, or cross-platform candidate matrix pass. The 83-package candidate plan was validated with CI_REQUIRE_FULL_ROOT=1 (83 served, 0 deferred, 0 excluded), but the full planned lane was not executed green. Extra broad CLI and profile attempts were deliberately stopped to keep checks bounded: real **exit 1**, 346.527s and 101.604s respectively, not passes. Their partial streams are retained for diagnosis; scoped reruns are the accepted evidence. The unexecuted remaining broad batches are not claimed. Python implementation/registry workflow lanes are outside this Curator leaf and were not run.

## Anomalies and host discipline

- Initial exact Implementations attempts failed **exit 1** on the old script-worker label check and absent marker-v3 gap rows. The accepted exact command is the third invocation, after correcting both digest/metadata dispatch and the measured five-row ledger omission.
- Exploratory corpus auditing failed on modified bytes in an older cached fixture checkout (Python exit 1; the surrounding exploratory shell later returned 0 after read-only searches), then on misclassifying the env-passthrough vector family as a schema-index family (standalone Python exit 1). A recorded audit rerun failed exit 1 while the five marker-v3 gaps were being reconciled. None is accepted green evidence. The corrected audit authenticates rc.14 directly and the changed historical documents, avoiding the modified older fixture; accepted corpus-audit-v3-measured exited 0.
- Initial lint exited **1** on cached findings referencing deleted worktrees. No suppressions or repository changes were added. Task-specific lint and Go caches resolved the cache contamination; fresh lint exited 0.
- The shared ~/.mini-build-lock appeared during this session. Two scoped invocations ran before the launcher became lock-aware and are superseded by candidate-targeted-locked and candidate-readers-locked. Subsequent Go/lint commands atomically acquired the directory lock and released only their own ownership. Earlier accepted spec commands ran while the lock did not exist.
- Syspolicyd was **running** before and after every recorded command; successive crashes stayed **359 → 359** throughout. No crash/backoff event occurred. The per-command counts and durations are in the table below.
- Binding host-rules.md says “Never edit LOGBOOK.md.” Findings were recorded in board notes and task outcomes. The conditional general logbook checklist item was explicitly replaced with the equivalent board/outcome recording item, retaining the LOGBOOK prohibition.

## Recorded command results

GOFLAGS=-work for every row. Candidate-prefixed tests used the rc.14 root; default-pin tests used the fresh rc.13 checkout; historical policy tests used the named historical root. Narrowing probes used temporary mutations restored before accepted reruns. Other rows use no conformance override unless their command names a root.

| Attempt | Exact process argv | Exit | Seconds | syspolicyd before → after | Crashes before → after |
| --- | --- | ---: | ---: | --- | --- |
| candidate-targeted | `go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance\|Coverage' -count=1` | 0 | 24.981 | running → running | 359 → 359 |
| corpus-audit | `python3 /tmp/TASK-261002-2ipeqa-evidence/audit.py` | 0 | 0.235 | running → running | 359 → 359 |
| default-pin-targeted | `go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance\|Coverage' -count=1` | 0 | 19.515 | running → running | 359 → 359 |
| spec-families | `python3 /tmp/TASK-261002-2ipeqa-spec-rc14/tools/implementation_coverage.py families --implementation go --root /tmp/TASK-261002-2ipeqa-spec-rc14/conformance/v1` | 0 | 0.154 | running → running | 359 → 359 |
| candidate-plan | `bash .github/ci/suite-plan.sh /tmp/TASK-261002-2ipeqa-spec-rc14/conformance/v1 /tmp/TASK-261002-2ipeqa-evidence/candidate-plan` | 0 | 8.73 | running → running | 359 → 359 |
| spec-implementations | `go test -count=1 -json ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy` | 1 | 31.742 | running → running | 359 → 359 |
| corpus-audit-rerun | `python3 /tmp/TASK-261002-2ipeqa-evidence/audit.py` | 1 | 0.301 | running → running | 359 → 359 |
| spec-implementations-rerun | `go test -count=1 -json ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy` | 1 | 26.058 | running → running | 359 → 359 |
| marker-v3-measure | `go test ./internal/marker -run '^TestReadAuthoritativeMarkerV3SchemaCases$' -count=1 -v` | 0 | 0.812 | running → running | 359 → 359 |
| corpus-audit-v3-measured | `python3 /tmp/TASK-261002-2ipeqa-evidence/audit.py` | 0 | 0.344 | running → running | 359 → 359 |
| spec-implementations-green | `go test -count=1 -json ./internal/interop ./internal/closure ./internal/skillspec ./internal/marker ./internal/moduleroots ./internal/scriptpolicy` | 0 | 27.967 | running → running | 359 → 359 |
| spec-consumption | `python3 /tmp/TASK-261002-2ipeqa-spec-rc14/tools/implementation_coverage.py go --stream /tmp/TASK-261002-2ipeqa-evidence/spec-implementations-green.log` | 0 | 0.11 | running → running | 359 → 359 |
| whitespace | `git diff --check` | 0 | 0.112 | running → running | 359 → 359 |
| formatting | `gofmt -l internal/conformancecoverage/rc14_test.go internal/conformancecoverage/coverage.go internal/conformancecoverage/content_hash_v2_gaps_test.go internal/config/environments_conformance_test.go internal/registry/schema_conformance_test.go internal/envmarker/marker_env_schema_test.go internal/marker/schema_coverage_test.go internal/contextlock/schema_conformance_test.go internal/envprofile/muse_test.go internal/envprofile/read_failure_conformance_test.go cmd/curator/muse_test.go internal/buildrepo/acquisition_conformance_test.go internal/scriptpolicy/conformance_test.go` | 0 | 0.075 | running → running | 359 → 359 |
| candidate-batch-0 | `go test -count=1 -json -timeout 8m github.com/relux-works/curator/cmd/curator` | 1 | 346.527 | running → running | 359 → 359 |
| candidate-batch-1 | `go test -count=1 -json -timeout 8m github.com/relux-works/curator/internal/envprofile` | 1 | 101.604 | running → running | 359 → 359 |
| candidate-readers | `go test -json -count=1 -timeout 8m ./internal/conformancecoverage ./internal/contextlock ./internal/envmarker ./internal/config ./internal/registry ./internal/buildrepo -run 'Coverage\|SchemaCases\|CarrierSchema\|ExternalRepositoryAcquisitionConformance'` | 0 | 22.388 | running → running | 359 → 359 |
| candidate-targeted-current | `go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance\|Coverage' -count=1` | 0 | 26.046 | running → running | 359 → 359 |
| narrowing-dispatch | `go test ./internal/conformancecoverage -run '^TestCoverageCandidateDispatch$' -count=1` | 1 | 1.705 | running → running | 359 → 359 |
| candidate-record | `bash .github/ci/candidate-suite.sh record /tmp/TASK-261002-2ipeqa-spec-rc14/conformance/v1 /tmp/TASK-261002-2ipeqa-evidence/candidate-identity` | 0 | 50.618 | running → running | 359 → 359 |
| narrowing-count | `go test ./internal/conformancecoverage -run '^TestRC14CoveragePolicyRetainsExactCountsAndCurrentGaps$' -count=1` | 1 | 2.651 | running → running | 359 → 359 |
| candidate-targeted-locked | `go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance\|Coverage' -count=1` | 0 | 20.032 | running → running | 359 → 359 |
| default-pin-targeted-locked | `go test ./internal/conformancecoverage ./cmd/curator -run 'Conformance\|Coverage' -count=1` | 0 | 19.39 | running → running | 359 → 359 |
| candidate-hash-vectors | `go test -json ./internal/conformancecoverage ./internal/hashing -count=1` | 0 | 2.021 | running → running | 359 → 359 |
| candidate-muse-profile | `go test -json ./internal/envprofile -run '^TestMusePublished' -count=1 -timeout 8m` | 0 | 28.753 | running → running | 359 → 359 |
| candidate-muse-fragments | `go test -json ./cmd/curator -run '^TestMuseFragmentV3PublishedCases$' -count=1 -timeout 8m` | 0 | 5.991 | running → running | 359 → 359 |
| candidate-read-failures | `go test -json ./internal/envprofile -run '^TestEnvironmentsReadFailureVectors$' -count=1 -timeout 8m` | 0 | 45.283 | running → running | 359 → 359 |
| candidate-readers-locked | `go test -json -count=1 -timeout 8m ./internal/conformancecoverage ./internal/contextlock ./internal/envmarker ./internal/config ./internal/registry ./internal/buildrepo -run 'Coverage\|SchemaCases\|CarrierSchema\|ExternalRepositoryAcquisitionConformance'` | 0 | 10.211 | running → running | 359 → 359 |
| historical-hash-policy | `go test ./internal/conformancecoverage -run 'TestCommittedCoveragePolicyParses\|TestCoverageCandidateDispatch' -count=1` | 0 | 1.742 | running → running | 359 → 359 |
| historical-muse-policy | `go test ./internal/conformancecoverage -run 'TestCommittedCoveragePolicyParses\|TestCoverageCandidateDispatch' -count=1` | 0 | 0.892 | running → running | 359 → 359 |
| build | `go build -o /tmp/TASK-261002-2ipeqa-curator-bin ./cmd/curator` | 0 | 3.812 | running → running | 359 → 359 |
| lint | `golangci-lint run --timeout 8m` | 1 | 31.099 | running → running | 359 → 359 |
| lint-fresh-cache | `golangci-lint run --timeout 8m` | 0 | 78.139 | running → running | 359 → 359 |
| protected-logbook | `git diff --exit-code -- LOGBOOK.md` | 0 | 0.032 | running → running | 359 → 359 |

Second recovery observed at 2026-10-02 10:08 UTC: syspolicyd running, successive crashes 368 (367 before the exact retry). No later Go validation started. Both interruptions reproduced the host failure despite -work, exclusive lock ownership and required backoff. This earlier host diagnosis is retained as evidence; the later binding decision authorizes review handoff with hosted validation pending.

### Revision 2 checklist alignment for the authorized handoff

The first `task-board handoff TASK-261002-2ipeqa --role developer` exited **1**, refusing unchecked local lint/build and generic logbook items. Under the later binding hosted-gate decision, those local execution requirements were replaced with a checked evidence requirement: record unrun local lint/build and stalled tests explicitly in outcomes and route validation to the hosted CR gate. This checkbox attests only that reporting/routing; it does not attest that lint/build passed. The generic logbook item was removed because host-rules prohibit LOGBOOK.md edits; the checked board-notes/outcome/LOGBOOK-unchanged item remains. Required hosted validation is still pending before acceptance. No additional Go retry was attempted.
