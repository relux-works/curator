# TASK-260916-2ok97n — R5 runtime conformance qualification

## Candidate and scope

- Worktree base/checkpoint: `d239c342fcad506559eaa4ee613d6f8606d2df71`; R1–R4 checkpoints are already in this Story branch. R5 edits remain uncommitted.
- Conformance pin: curator-spec `dced9b8317e0e8af79edf2d0539b32bd22b6c85b` (`dced9b8`), release qualification rc.12. The vector's protocol version is `1.0.0-rc.9`.
- The stream model remains the bounded request/result model. Pass-through streams would require protocol and cross-platform handle-inheritance work beyond R5. Docs now state the bounds: piped/file stdin is capped at 64 MiB; terminal stdin is null-bound; stdout and stderr share a 16 MiB budget; overflow refuses without forwarding partial capture. Interactive/live output and larger combined output are unsupported.

## Conformance classification

All twelve vector top-level keys now classify as consumed and point to one or more production-entry rows. `TestScriptHostExecutionPolicySectionsAreAllClassified` checks the exact JSON key set and ledger registration. `TestScriptHostExecutionPolicyProductionConsumersCoverAllCases` checks all named cases and controls against the pinned vector and ledger: 33/33 behavioural cases and 11/11 mandatory controls. The eight-control inventory and both evidence record field sets are also checked against the pin. No `unreachable` or `not implemented` classification remains.

| Vector key | Registered production-entry consumers |
| --- | --- |
| `schema_version` | `internal/install::TestScriptOptInCasesAtInstallEntry` |
| `protocol_version` | `internal/scriptworker::TestBuiltCuratorWorkerHandshake` |
| `execution_policy` | `internal/install::TestEnforcedInstallAndLaunchAtCLIEntry` |
| `interpreters` | `internal/scriptworker::TestProductionBinaryLaunchesWhenHostProvides`, `TestWindowsRealInterpretersRunDeclaredExec` |
| `opt_in_cases` | `internal/install::TestScriptOptInCasesAtInstallEntry` |
| `capability_derivation_cases` | `internal/scriptworker::TestCapabilityDerivationAllFieldsAbsentDenyByDefault`, `TestCapabilityDerivationNetworkHostsReportingOnly`, `TestCapabilityDerivationExecIsManagerResolved`, `TestCapabilityDerivationSecretsRemainIdentifiers` |
| `mandatory_controls` | `internal/scriptworker::TestHostProbeReportsClosedInventory`, `internal/install::TestEnforcedScriptCommandIsRefusedAtInstall`, `internal/scriptworker::TestPreflightRefusesUnavailableControlAtInvocation` |
| `native_control_inventory` | `internal/scriptworker::TestHostProbeReportsClosedInventory`, `TestLinuxHostProbeAndEvidenceAreConsistent`, `TestMacOSNativeControlsAppliedAtInvocation`, `TestWindowsJobLimitsAppliedAndConfirmed` |
| `capability_evidence_record` | `internal/scriptworker::TestScriptEvidenceValidRecordSucceeds`, `TestRunShimWritesDiagnosticsRecord` |
| `capability_evidence_cases` | `internal/scriptworker::TestLinuxHostConditionalUnavailableSucceeds`, `TestScriptEvidenceMutationsRefuseBeforePermit`, `TestScriptEvidenceSecondRecordRefuses` |
| `preflight_cases` | `internal/install::TestEnforcedScriptCommandIsRefusedAtInstall`; `internal/scriptworker::TestPreflightRefusesUnavailableControlAtInvocation`, `TestLinuxPidsMaxProbeAvailableApplies`, `TestLinuxPidsMaxProbeUnavailableSucceeds`, `TestFixedUnavailableControlDoesNotReject` |
| `audit_label_cases` | `internal/install::TestScriptAuditLabelsAtInstallEntry`, `TestScriptAuditLabelsAtGlobalInstallEntry`; `cmd/curator::TestCLIAuditEmitsScriptLabels` |

The 33 named cases remain tied to their individual test rows in `scriptVectorCaseConsumers`: 6 opt-in, 4 capability derivation, 14 evidence, 5 preflight, and 4 audit cases. In the evidence family, the valid Linux host-conditional case maps to `TestLinuxHostConditionalUnavailableSucceeds`, the second-record case maps to `TestScriptEvidenceSecondRecordRefuses`, and the other 12 forgery/shape/status cases map to `TestScriptEvidenceMutationsRefuseBeforePermit`. The static map is checked for both missing and stale vector names.

The eleven mandatory controls map to production tests for process graph, worker identity, interpreter identity, environment, PATH/exec resolution, offline configuration, private runtime area, stream binding, native controls, evidence closure, and descendant teardown. `inventory-controls-applied` maps to the closed inventory plus the Linux Landlock and Windows Job Object rows. `native_control_inventory` also now names the macOS invocation row.

## Ledger changes

Added 27 R5 rows covering the vector-to-row ratio, interpreter replacement, stdin and exit status, macOS native controls, real Windows interpreters, audit Gate/persistence/formatting, skillcheck, project/global install, and CLI audit text/JSON paths. Three existing platform rows were corrected so Linux cgroup and network-namespace cases require Linux and tolerate Darwin/Windows skips with the existing `platform-control` class. No skip class was added. Linux cgroup/Landlock/netns, macOS native controls, and Windows Job Object/interpreter rows require their native lane.

The Darwin subset platform-case gate passed. Its only tolerated skip was `TestWindowsRealInterpretersRunDeclaredExec`, correctly classified as Windows-only `platform-control`. The ledger consistency gate checked all 344 rows across Linux, Darwin, and Windows and passed.

## Platform matrix and host evidence

The operator guide and troubleshooting page now document the control matrix and every worker refusal class. The matrix is:

| Platform | Available and applied | Host-conditional | Fixed unavailable |
| --- | --- | --- | --- |
| Linux | descendant termination, per-file size, inherited-handle restriction | process count and aggregate memory (delegated cgroup v2); descendant exec and filesystem writes (Landlock); network isolation (network namespace) | none |
| macOS | descendant termination, per-file size, inherited-handle restriction | none | process count, aggregate memory, descendant exec, filesystem writes, network isolation |
| Windows | descendant termination, process count and aggregate memory (Job Object), inherited-handle restriction | none | per-file size, descendant exec, filesystem writes, network isolation |

| Lane | Evidence in this developer run |
| --- | --- |
| macOS x86_64 | Focused production-entry platform-case gate passed; native macOS control test is registered. Darwin package and audit/install/CLI subsets passed. |
| Ubuntu hosted | Not observed on this uncommitted candidate. The Linux test binary cross-compiled successfully; this is compile evidence only. Hosted execution remains for the integration/landing run. |
| Windows hosted | Not observed on this uncommitted candidate. The Windows test binary cross-compiled successfully; this is compile evidence only. `TestWindowsRealInterpretersRunDeclaredExec` must run on `windows-latest` with real `python.exe`, `node.exe`, and the declared `cmd.exe` grant to establish R-d. |
| rose-air ARM64 | Unverified. As specified in the addendum, the self-hosted lane runs on main pushes and main is red for the unrelated runner issue BUG-260922-k6eypp, being diagnosed by BUG-260922-306v4m. On the landing run rose-air should add the Darwin-required R5 rows, especially `TestMacOSNativeControlsAppliedAtInvocation`, the production vector mapping, launch/stream rows, and the common audit/validation/install/CLI rows listed in the ledger. |

Known unrelated Windows bounds from the carried review notes: `internal/snapshot::TestConcurrentGetAcceptsOneImmutablePublisher` has a pre-existing Windows file-locking flake; `internal/managerlock` has the tiny-deadline flake BUG-260922-6chzf9. No candidate Windows lane was observed, so neither is reported as a candidate failure or pass.

## Narrowing mutants

| Mutant | Result |
| --- | --- |
| Bypass interpreter-file re-verification only for `python3-v1` in `serveRun` | Killed by `TestScriptWorkerRechecksInterpreterBeforeExec` (mutant run exit 1: worker returned success where the test required refusal). The restored-source focused test passed, exit 0. |
| Accept exactly the sidecar JSON field `extra` by adding it to `ShimSidecar`, while leaving `DisallowUnknownFields` enabled | Killed by `TestLoadShimSidecarRejectsMalformed/unknown-field` through production `RunShim` (mutant run exit 1 in 2.65 s: RunShim launched the stub and returned 0 instead of refusing). The temporary field was removed. An earlier unbounded attempt was interrupted with exit 130 after about 9m20s; the bounded 30-second retry produced the decisive result. |
| Remove `decoder.DisallowUnknownFields()` | Also killed by the production-entry malformed-sidecar test (exit 1; the unknown-field case reached and ran the stub). This is broader evidence than the exact-field mutant above. |

Evidence/attestation mutants from the R3 implementation remain in the existing `TestScriptEvidenceMutationsRefuseBeforePermit` cases; R5 maps all fourteen pinned evidence shapes to that production invocation row, with the valid, host-conditional, and second-record cases mapped to their corresponding invocation tests.

## Verification and exit codes

All commands below ran as standalone processes. The latest successful runs are the final candidate evidence; earlier failures are retained where they reveal a timing or executor bound.

| Command/check | Exit | Result |
| --- | ---: | --- |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.temp/resources/r5-conformance/conformance/v1" go test -count=1 -timeout=3m ./internal/scriptpolicy` | 0 | Pinned rc.12 vector consumer passed in 0.388 s. |
| `go test -count=1 -timeout=5m ./internal/scriptworker` (latest rerun) | 0 | Full worker package passed in 17.974 s. |
| Same full worker package command, two earlier attempts | 1 each | Timed out after 330.514 s and 300.439 s, waiting for child-worker frames in two different tests. Even trivial shell commands then stalled; a fresh rerun after shell recovery passed without a code change. The timeouts remain recorded as an intermittent local-executor/runtime risk. |
| `go test -count=1 -timeout=7m ./internal/install -run '^(TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced|TestActiveEnforcedScriptCommandsAdmits|TestActiveEnforcedScriptCommandsRefusesUnavailableHost|TestEnforcedScriptCommandInstallsNativeLauncher|TestScriptOptInCasesAtInstallEntry|TestGlobalEnforcedInstallExcludesForwardingMirror|TestEnforcedInstallAndLaunchAtCLIEntry|TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher|TestScriptAuditLabelsAtInstallEntry|TestScriptAuditLabelsAtGlobalInstallEntry)$'` | 0 | Production install/launcher and project/global audit subset passed in 366.601 s. |
| `go test -count=1 -timeout=3m ./internal/skillcheck` | 0 | Validation audit rows passed in 1.041 s. |
| `go test -count=1 -timeout=3m ./cmd/curator -run '^(TestCLIAuditEmitsScriptLabels|TestCLIAuditJSONCarriesScriptLabels)$'` | 0 | CLI audit text and JSON entry tests passed in 2.304 s. |
| Fresh Darwin R5 JSON stream, serial package execution (`go test -p 1 -json -count=1 -timeout=7m` over the six listed R5 packages, with the pinned conformance root and tests from `r5-platform-cases-subset.tsv`) | 0 | The current registered test subset completed and produced `.temp/resources/r5-platform-test.json`. Serial package execution avoided a prior multi-package run that timed out in install and CLI subprocess-output waits. |
| `CI_PLATFORM_CASES=.temp/resources/r5-platform-cases-subset.tsv bash .github/ci/platform-case-gate.sh .temp/resources/r5-platform-test.json .temp/resources/r5-platform-gate-evidence` | 0 | Darwin gate passed: all 27 observed production-entry rows passed; the one recorded skip was the Windows-only real-interpreter row, tolerated by its `platform-control` ledger entry. |
| First fresh multi-package JSON run with default package parallelism | 1 | `internal/install` and `cmd/curator` hit 7-minute test timeouts while reading child output. The serial `-p 1` run above passed. This failed attempt is not counted as platform evidence. |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.temp/resources/r5-conformance/conformance/v1" bash .github/ci/ledger-consistency.sh .temp/resources/r5-ledger-final` | 0 | All 344 platform ledger rows checked across Linux, Darwin, and Windows. |
| `go build ./...`; `go vet ./...`; `golangci-lint run`; `bash .github/ci/no-broad-suppression.sh`; `git diff --check`; `gofmt -d` over all changed Go files | 0 each | Build, vet, lint, suppression, whitespace, and formatting checks passed. Lint reported 0 issues and a non-fatal warning about an unrelated missing temporary worktree `STORY-260919-37szes`. |
| `GOOS=linux GOARCH=amd64 go test -c ./internal/scriptworker -o .temp/resources/scriptworker-linux.test`; `GOOS=windows GOARCH=amd64 go test -c ./internal/scriptworker -o .temp/resources/scriptworker-windows.test.exe` | 0 each | Linux and Windows test binaries cross-compiled; these are compile checks, not runtime evidence. |
| Pre-exec interpreter-check narrowing mutant: `go test -count=1 -timeout=60s ./internal/scriptworker -run '^TestScriptWorkerRechecksInterpreterBeforeExec$'` | 1 (expected) | Removing only the worker-side `verifyInterpreterFile` check in `serveRun` while retaining accept-time verification made the production-entry test receive `result` instead of the required failure. The test was rerun green after restoration. `cmp -s` and `git diff --exit-code -- internal/scriptworker/worker.go` each exited 0 after restoration. |
| Sidecar unknown-field mutants | 1 each (expected) | Adding only `extra` to the sidecar struct, then separately removing `DisallowUnknownFields`, was killed by `TestLoadShimSidecarRejectsMalformed/unknown-field` through `RunShim`; each mutant allowed the stub to launch. Mutations were restored. |
| Focused earlier `scriptpolicy` assertion while adding evidence-field checks | 1 | A field-order assertion failed; it was corrected to compare field sets and the pinned-vector suite passed. |
| Earlier ledger invocation without the required evidence-directory argument | 2 | Usage error; corrected command above passed. |
| Earlier lint attempts | 1 each | Found capitalized `Landlock` errors; wording was fixed and the final lint command above passed. |

Evidence/attestation narrowing cases from R3 remain production-invocation tests in `TestScriptEvidenceMutationsRefuseBeforePermit`; R5 maps all fourteen pinned evidence shapes to registered invocation rows. The fresh Darwin platform gate does not replace the Linux and Windows host lanes.

The full landing suite was not run manually; the handoff runtime owns its single configured run. Ubuntu hosted, Windows hosted (including real `python.exe`, `node.exe`, and declared `cmd.exe`), and rose-air runtime evidence remain unobserved in this developer run. They must be recorded by the integration/landing owner. Rose-air remains “to be observed on the landing run” under the task brief; the known unrelated main runner issue BUG-260922-k6eypp remains a qualification bound, not evidence of candidate success or failure.

## Logbook boundary

No `logbook` executable is available on this host, and campaign rules prohibit editing `LOGBOOK.md`. The runtime stalls, serial retry outcome, and cross-platform evidence bounds are recorded in this task-scoped result and board notes instead.

## Revision 2 — Windows real-interpreter rework

Work continued from the revision-1 candidate at checkpoint `d239c342fcad506559eaa4ee613d6f8606d2df71`; all changes remain uncommitted in the Story worktree.

### Windows failure evidence and diagnosis

I downloaded the `test-evidence-windows-latest` artifact from hosted run [35861781844](https://github.com/relux-works/curator/actions/runs/35861781844), candidate SHA `e0d0fdaa4813dd8f9cee55c5df0de85ded65e1af`. The JSON stream confirms:

- `TestScriptHostExecutionPolicySectionsAreAllClassified` and `TestScriptHostExecutionPolicyProductionConsumersCoverAllCases` saw the CRLF blank line as `"\r"` and failed with `line 43 has 1 columns`.
- Both real interpreters were found. Python ran but `Report.ResolvedExec["cmd.exe"]` was empty; Node failed with `spawnSync cmd.exe ENOENT`. The declared executable was absent from the manager-built PATH farm.

The production path is manager-side: `runSession` calls `DeriveProfile`, which calls `resolveDeclaredExec` and `ResolveExec` before starting the worker. The worker environment builder already case-folds Windows environment keys and sets `SYSTEMROOT` from the captured manager `HostEnvironment` (falling back to the manager process environment). The old default resolver independently read `os.LookupEnv("SystemRoot")`, so the search list was not derived from the same captured input. The failed run did not log its `SYSTEMROOT` value, so it does not establish whether that specific process value was absent; the code-level source mismatch and the missing farm entry are established.

`DeriveProfile` now builds the default search list from the same manager environment used for worker environment construction. Windows lookup uses its case-insensitive `SYSTEMROOT` entry and searches `%SystemRoot%\System32` before `%SystemRoot%`. The existing Windows farm copies the manager-resolved executable into its private directory, and the worker receives that directory as its sole `PATH`. No environment passthrough was added. The Windows real-interpreter row now logs the manager root, search directories, resolved target, and farm entries; it requires the exact absolute `%SystemRoot%\System32\cmd.exe` result and proves both Python and Node can start it.

The platform-case reader now removes one trailing carriage return and ignores whitespace-only and indented comment lines like the shell ledger gate. `TestPlatformCaseLedgerRowsAcceptCRLFBlankLinesAndComments` feeds a CRLF ledger with both blank and comment lines. `TestDefaultExecSearchDirsUsesManagerSystemRoot` proves the Windows search order uses the captured manager value, case-insensitively. Both new tests are registered for Linux, Darwin, and Windows. The ledger itself was not altered to hide its blank lines; no skip class was added.

### Narrowing mutants

| Mutant | Result |
| --- | --- |
| Drop the Windows `System32` search directory | `TestDefaultExecSearchDirsUsesManagerSystemRoot` failed as expected, exit **1**, because the returned list omitted `System32`. After restoration, the combined search-list and production derivation tests passed, exit **0**. |
| Omit `resolved[name] = identity.Path` from `resolveDeclaredExec` | Production-entry `TestCapabilityDerivationExecIsManagerResolved` failed as expected, exit **1**, with an empty resolved path. After restoration, that production test passed, exit **0**. |

The real Windows interpreter row adds the same System32 and resolution-record assertions at the actual launch boundary. It was not runnable on this Darwin host; hosted Windows runtime evidence is still required. The cross-compiled Windows test binary is compile evidence only.

### Revision-2 verification

Commands ran as standalone processes. Expected mutant failures are reported as failures; each was followed by a green run after restoration.

| Command/check | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout=2m ./internal/scriptpolicy -run '^TestPlatformCaseLedgerRowsAcceptCRLFBlankLinesAndComments$'` | 0 | CRLF, blank, and comment fixture passed. |
| `go test -count=1 -timeout=3m ./internal/scriptworker -run '^(TestDefaultExecSearchDirsUsesManagerSystemRoot|TestCapabilityDerivationExecIsManagerResolved)$'` | 0 | Search-list and manager-resolved production derivation rows passed after mutation restoration. |
| `go test -count=1 -timeout=5m ./internal/scriptworker` | 0 | Full scriptworker package passed in 39.843 s. |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.temp/resources/r5-conformance/conformance/v1" go test -count=1 -timeout=3m ./internal/scriptpolicy` | 0 | Full pinned scriptpolicy package passed in 1.220 s. |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.temp/resources/r5-conformance/conformance/v1" bash .github/ci/ledger-consistency.sh .temp/resources/r5-ledger-rework2` | 0 | All 346 ledger rows matched Linux, Darwin, and Windows test inventories. |
| `go build ./...` | 0 | Repository build passed. |
| `go vet ./...` | 0 | Vet passed. |
| `golangci-lint run` | 0 | 0 issues. It printed the existing non-fatal warning for a missing unrelated temporary worktree `STORY-260919-37szes`. |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/scriptworker -o .temp/resources/scriptworker-windows-rework2.test.exe` | 0 | Windows worker tests, including the real-interpreter row, cross-compiled. |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/scriptpolicy -o .temp/resources/scriptpolicy-windows-rework2.test.exe` | 0 | Windows ledger-parser and conformance tests cross-compiled. |
| `git diff --check` and `gofmt -d` over the changed Go files | 0 each | No whitespace or formatting differences. |

Hosted Ubuntu/macOS/Windows execution and rose-air are not claimed from this developer run. Per the rewritten handoff checklist, hosted evidence is to be produced by the handoff gate and judged by the reviewer; rose-air remains to be observed on the landing run. In particular, the previous Windows run is a real failure, not passing evidence for this revision.

### Revision-2 logbook boundary

No `logbook` executable is available and campaign rules prohibit editing `LOGBOOK.md`. The prior Windows failures, resolver source mismatch, both mutant results, and the unverified hosted-runtime bound are recorded here as the task-scoped logbook-equivalent outcome.

## Revision 3 — Windows System32 identity and unresolved grants

Work continued from the R2/R4 checkpoint `d239c342fcad506559eaa4ee613d6f8606d2df71`. All changes remain uncommitted in the Story worktree.

### Root cause and production change

The previous revision already made the default Windows search list use the captured manager environment. In the remaining failure, `SystemRoot`, the ordered `System32`/root search list, and the exact `cmd.exe` declaration were present in the Windows test output, but `ResolveExec` still returned no identity. The declaration is parsed from `CapabilitiesRaw` in `runSession` (`internal/scriptworker/client.go:79-93`), then resolved before the worker start (`client.go:94-96`); `buildPathFarm` can only copy entries that resolved successfully.

The remaining candidate filter was the unconditional multiple-link refusal in `readExecIdentityAt` at the R2 checkpoint (`internal/scriptworker/exec.go:179-184`). Windows component-store files can be hard-linked into System32. When that applies to the runner's `cmd.exe`, the identity gate returned `found=false`, leaving both `Report.ResolvedExec` empty and the farm interpreter-only. The earlier hosted artifact did not include cmd.exe's link count, so that host-specific link state was not directly observed; this diagnosis is an inference from the remaining code path and the empty resolution report. Microsoft documents that Windows component-store files are commonly hard-linked into directories outside WinSxS ([Manage the Component Store](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/manage-the-component-store?view=windows-11)). The real Windows test now logs cmd.exe's link/reparse check so the hosted rerun can confirm the runner's exact state.

`resolveExecForPlatform` and `readExecIdentityAt` now allow a multiply-linked source only when all of these hold: the manager is using the default Windows search list, the manager's captured `SYSTEMROOT` supplied the canonical `System32` root, and the resolved candidate is physically below that root (`internal/scriptworker/exec.go:57-90, 159-168, 220-281`). Explicit search directories remain under the strict single-link rule. The source is still hashed and rechecked at launch; Windows copies those bytes into the private PATH farm. Linux and macOS retain the same fixed directory lists and single-link behavior.

`DeriveProfile` now resolves nil `ExecSearchDirs` against the same captured manager environment and reports an unresolved grant as the typed `script_execution_declared_exec_unresolved` error before PATH-farm construction (`internal/scriptworker/capabilities.go:154-185`). The production caller returns at `client.go:94-96`, before nonce creation and before `domain.launch`. This also closes the second required gap: Launch cannot start a worker whose declared exec is absent from the manager farm.

### Rows and mutant results

- `TestDeriveProfileUsesDefaultExecSearchDirsAndBuildsDeclaredExecFarm` injects the Windows platform and mixed-case manager `SYSTEMROOT` on all hosted OSes, leaves `ExecSearchDirs` nil, models a component-store hard link, and checks both `ResolvedExec["cmd.exe"]` and the farm entries. Its explicit-directory subcase confirms the hard-link allowance does not extend to a caller-supplied directory.
- `TestLaunchRefusesUnresolvedDeclaredExecBeforeWorker` drives `Launch` with an unresolved declaration and a same-name executable on caller `PATH`; it requires the typed refusal and no interpreter marker.
- `TestCapabilityDerivationExecIsManagerResolved` now matches the pinned vector exactly: its only declared exec grant is `git`. Missing-name behavior is covered by the separate refusal row.
- The Windows real-interpreter test still runs real `python.exe` and `node.exe` through `Launch` with the declared `cmd.exe` grant. It also logs the link/reparse check from the actual System32 target.
- Both new test functions are registered cross-platform in `.github/ci/platform-cases.tsv`; ledger consistency reports 348 rows for linux/darwin/windows. No skip class was added.

All mutants were run in the disposable copy `.temp/resources/r5-mutant`, not in the candidate worktree:

| Mutant | Test command in the disposable copy | Real result |
| --- | --- | --- |
| Drop declared exec entries while building the PATH farm | `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestDeriveProfileUsesDefaultExecSearchDirsAndBuildsDeclaredExecFarm$'` | Killed, exit **1**: farm contained only `stubinterp`, missing `cmd.exe`. |
| Treat nil `ExecSearchDirs` as an empty list | `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestDeriveProfileUsesDefaultExecSearchDirsAndBuildsDeclaredExecFarm$'` | Killed, exit **1**: derivation returned `script_execution_declared_exec_unresolved` for `cmd.exe`. |
| Narrow the unresolved-grant refusal to two or more missing names | `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestLaunchRefusesUnresolvedDeclaredExecBeforeWorker$'` | Killed, exit **1**: Launch returned nil for the single missing grant. |

All three commands ran with working directory `.temp/resources/r5-mutant`.

### Revision-3 validation

Every validation command below was run as a standalone process. Exit codes are the actual process results; the three mutant failures above are expected evidence.

| Command/check | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout=5m ./internal/scriptworker/...` | 0 | Full worker package passed in 35.956 s. |
| `go test -race -count=1 -timeout=5m ./internal/scriptworker/...` | 0 | Race-enabled worker package passed in 78.618 s. |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.temp/resources/r5-conformance/conformance/v1" go test -count=1 -timeout=3m ./internal/scriptpolicy -run '^(TestScriptHostExecutionPolicySectionsAreAllClassified|TestScriptHostExecutionPolicyProductionConsumersCoverAllCases)$'` | 0 | Pinned vector classification and 33/33 case / 11/11 control coverage passed. |
| `bash .github/ci/ledger-consistency.sh .temp/resources/r5-ledger-rework3` | 0 | All 348 rows matched linux, darwin, and windows test inventories. |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/scriptworker -o .temp/resources/scriptworker-windows-r5-rework3.test.exe` | 0 | Windows test binary, including the real-interpreter row, cross-compiled. This is not Windows runtime evidence. |
| `go build ./internal/scriptworker/...` | 0 | Worker package build passed. |
| `go vet ./internal/scriptworker/...` | 0 | Worker package vet passed. |
| `golangci-lint run ./internal/scriptworker/...` | 1 then 0 | Initial run failed SA1024 for a duplicate cutset character. After correction, rerun passed with 0 issues. |
| `gofmt -d` over changed Go files; `git diff --check` | 0 each | No formatting or whitespace errors. |

### Platform evidence bounds

This developer session ran on Darwin and did not run on Ubuntu or Windows. Windows runtime execution of the real Python/Node and declared `cmd.exe` row remains unverified here; the successful Windows cross-compile is compile evidence only. The hosted Ubuntu/macOS/Windows gate must judge the candidate. Rose-air is still “to be observed on the landing run”; it should add its Darwin-required rows, including `TestMacOSNativeControlsAppliedAtInvocation`, the vector mapping, common launch/stream rows, and audit/install/CLI rows. The full landing suite was not run manually; the handoff runtime owns its single configured run.

## Revision 3 — refresh onto fad88136

### Refresh and candidate identity

On 2026-09-23, the authorized `origin` SSH URL was queried fresh through the existing SSH agent; `refs/heads/main` advertised `fad881368b632f43a18b01681a8fc7def110cbae`. `task-board worktree refresh-candidate TASK-260916-2ok97n` then returned exit **0**, `Outcome=refresh_advanced`, `TrunkOID=ReviewedTrunkOID=fad881368b632f43a18b01681a8fc7def110cbae`, and `BranchOID=ac0e9774aea90ef689379ec138e9bd0132904d98`. The refreshed HEAD is `TASK-260916-1xib1x: emit-script-audit-labels`; `git merge-base --is-ancestor fad881368b632f43a18b01681a8fc7def110cbae HEAD` returned exit **0**. The repository remote and persistent Git configuration were not changed.

The first scoped `git apply --3way` attempt returned exit **1** because the already-edited ledger and changelog worktree files differed from the index. I temporarily staged only those two candidate files, retried the five-file scoped patch successfully (exit **0**), and unstaged the incoming paths. `git diff --cached --quiet` returned exit **0** before and after refresh. The carried handoff-blocker and two scriptworker test files have hashes identical to the R3 safety ref `refs/campaign/r5-rev3-delta-20260923` (`3a3be93a…`, `ce806441…`, `166bff36…` respectively); the result document is intentionally updated by this section.

### Per-file incoming delta

Compared with the saved R3 candidate, excluding the worktree's `.task-board` snapshot, the incoming change is exactly these five code/config files:

| File | Incoming content |
| --- | --- |
| `.github/ci/platform-cases.tsv` | Four gitops directory-fold cases added; the prior R5 cgroup/Landlock/netns rows already matched the refreshed platform shape. |
| `.github/ci/platform-exclusions.tsv` | Clarifies that the default exclusion applies to explicitly supplied pre-vector roots; the committed pin uses its vector. |
| `CHANGELOG.md` | Preserves the R5 script-worker entry and adds the git snapshot directory-component collision fix. |
| `internal/gitops/gitops.go` | Tracks folded ancestor directory components and file/directory collisions before streaming. |
| `internal/gitops/dirfold_test.go` | Adds case-fold and case-sensitive extraction tests for directory and file/directory collisions. |

The refresh leaves nine tracked `.task-board` snapshot paths modified in the Story worktree because the old branch snapshot differs from the refreshed trunk snapshot. I did not edit those board bytes; the authoritative board remains the control-root `.task-board` and all board mutations use `task-board`. The previous task Change Request patch likewise contained no `.task-board` paths. The carried task results, handoff note, and scriptworker tests remain available as untracked worktree deliverables; the scriptworker test file contents match the saved R3 candidate exactly.

### Rerun validation and exit codes

Commands rerun in this refresh turn (each a standalone process):

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/scriptworker/... ./internal/gitops/...` | 0 | Both packages passed (39.777s and 20.409s). |
| `sh .github/ci/ledger-consistency.sh` | 2 | Usage failure: this script requires an evidence-directory argument. |
| `sh .github/ci/ledger-consistency.sh .temp/resources/r5-refresh-fad88136-ledger` | 0 | 352 ledger rows matched compiled test inventories across Linux, Darwin, and Windows. |
| `sh .github/ci/gate-selftest.sh` | 0 | 185 passed, 0 failed. |
| `git merge-base --is-ancestor fad881368b632f43a18b01681a8fc7def110cbae HEAD` | 0 | Refreshed branch descends from the advertised trunk. |

No full landing suite was run manually; the campaign assigns its single run to the handoff runtime. This refresh turn produced no hosted Ubuntu/macOS/Windows or rose-air runtime result. Existing candidate evidence remains as described above: hosted execution is delegated to the handoff/landing gate; rose-air remains unverified and is expected only on landing (the known runner issue in the refresh addendum is external to this candidate).

### Refresh access attempts

The first refresh attempt returned exit **1** because the inherited SSH-agent socket could not authenticate. A temporary Git URL rewrite reached GitHub but made the effective remote disagree with the inherited repository binding, so that attempt also returned exit **1** and was discarded. The successful retry used `GIT_SSH_COMMAND='ssh -o BatchMode=yes'` with the unchanged authorized SSH remote and existing agent. No remote, key, or persistent configuration was modified.

## Revision 4 — unresolved declared execs are report-only

This revision follows the `protocol/core.md` derived-capabilities rule supplied
with the rework: a declared executable name the manager cannot resolve stays
out of PATH and is reported; it does not refuse launch or consult caller PATH.
The earlier Revision 3 statement that unresolved declarations return
`script_execution_declared_exec_unresolved` before worker start is superseded.

### Production change and row

`DeriveProfile` now carries unresolved names through profile construction.
Only manager-resolved identities enter `buildPathFarm`; `DerivationReport`
retains the names in its existing `UnresolvedExec` field (`unresolved_exec` in
the result-only JSON record). The launch refusal code and the obsolete
troubleshooting description were removed. The installer continues to report
unresolved names in its existing result message.

`TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied` drives
`Launch` with two unresolved grants and a same-name executable only on caller
PATH. The test requires a successful launch, checks the interpreter's own
`exec.LookPath` result is `missing`, verifies neither unresolved name is in
the resolved report or farm, and checks both names appear in
`Report.UnresolvedExec`. Two missing names also keep the previous narrowing
mutant (refuse only when at least two are missing) observable. The test is
registered for Linux, macOS, and Windows in the platform-case ledger.
`TestEnforcedScriptCommandInstallsNativeLauncher` was left unchanged and
passed with its single unresolved grant.

The Windows multiple-link allowance from Revision 3 remains bounded to the
default Windows search list only. The trusted root is the canonical
`%SystemRoot%\System32` derived from the manager's captured `SYSTEMROOT`; a
candidate must be physically below that root. Explicit search directories do
not receive the allowance. The selected path is still the manager-authorized
System32 path: an NTFS component-store hard link adds another name for that
same file object, rather than selecting an executable from caller PATH or an
untrusted search root. The source identity is hashed and rechecked, and the
verified bytes are copied into the private farm. The current protocol wording
still conflicts with this narrowly scoped platform case; the orchestrator will
file the spec erratum that makes the intended System32 exception explicit.

### Narrowing mutants

All mutant runs used the disposable copy
`/tmp/TASK-260916-2ok97n-r5-rework4-mutant`, never the candidate worktree.

| Mutant | Command (run from the disposable copy) | Result |
| --- | --- | --- |
| Drop resolved declared execs while building the farm (Revision 3) | `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestDeriveProfileUsesDefaultExecSearchDirsAndBuildsDeclaredExecFarm$'` | Killed, exit **1**: the derivation row found only `stubinterp`, not `cmd.exe`. |
| Treat nil `ExecSearchDirs` as an empty list (Revision 3) | `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestDeriveProfileUsesDefaultExecSearchDirsAndBuildsDeclaredExecFarm$'` | Killed, exit **1**: the derivation row had no manager-resolved `cmd.exe`. |
| Refuse only when at least two names are unresolved (Revision 3) | `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied$'` | Killed, exit **1**: the production `Launch` row refused its two missing names. |
| Drop unresolved names from `DerivationReport` | `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied$'` | Killed, exit **1**: the row observed an empty unresolved report. The initial field-removal encoding caused only an unused-variable compile error (exit **1**) and was not counted; the corrected mutant dropped the entries and reached this assertion. |
| Add caller PATH to exec resolution | `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied$'` | Killed, exit **1**: the interpreter resolved the caller decoy from the farm, contrary to the required `missing` result. |

After restoring each mutant, `cmp -s` verified the disposable `exec.go` and
`capabilities.go` matched their saved candidate copies (exit **0**). The clean
copy then passed `go test -count=1 -timeout=3m ./internal/scriptworker
-run '^TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied$'` (exit
**0**).

### Revision 4 validation and exit codes

Each command ran as a standalone process. Expected mutant failures above are
reported as failures. The first broad command timed out and is not counted as
passing evidence; the scoped install cases were rerun separately with test
parallelism limited to one.

| Command/check | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout=3m ./internal/scriptworker -run '^TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied$'` | 0 | New production `Launch` case passed on Darwin. |
| `go test -count=1 -timeout=8m ./internal/scriptworker/... ./internal/install -run 'Script|Enforced'` | 1 | Timed out at 8m in `internal/install` with five parallel filesystem-transaction subtests still running (two global audit cases, two project audit cases, and one opt-in case). The scriptworker package printed `ok`; the combined command did not pass. |
| `go test -count=1 -timeout=5m -parallel 1 ./internal/install -run '^TestEnforcedScriptCommandInstallsNativeLauncher$'` | 0 | The unchanged install-and-launch test passed in 104.276s. |
| `go test -count=1 -timeout=5m -parallel 1 ./internal/install -run '^TestScriptOptInCasesAtInstallEntry$'` | 0 | All six opt-in installation cases passed in 88.501s. |
| `go test -count=1 -timeout=5m ./internal/scriptworker/... -run 'Script|Enforced'` | 0 | Scriptworker subset passed in 12.355s. |
| `go test -race -count=1 -timeout=5m ./internal/scriptworker/...` | 0 | Full race-enabled scriptworker suite passed in 85.218s. |
| `env CURATOR_CONFORMANCE_ROOT="$PWD/.temp/resources/r5-conformance/conformance/v1" go test -count=1 -timeout=3m ./internal/scriptpolicy -run '^(TestScriptHostExecutionPolicySectionsAreAllClassified|TestScriptHostExecutionPolicyProductionConsumersCoverAllCases)$'` | 0 | Classification and production-consumer coverage passed: 33/33 named cases and 11/11 mandatory controls. |
| `bash .github/ci/gate-selftest.sh` | 0 | 185 passed, 0 failed. |
| `bash .github/ci/ledger-consistency.sh .temp/resources/r5-rework4-ledger` | 0 | All 352 ledger rows matched compiled test inventories for Linux, Darwin, and Windows. |
| `go test -json -count=1 -timeout=3m ./internal/scriptworker -run '^TestLaunchUnresolvedDeclaredExecReportedAndCallerPathDenied$' > .temp/resources/TASK-260916-2ok97n-platform-test.json` | 0 | Produced the Darwin JSON stream used for the registered-case gate. |
| `CI_PLATFORM_CASES=/tmp/TASK-260916-2ok97n-r5-rework4-mutant/platform-case.tsv bash .github/ci/platform-case-gate.sh .temp/resources/TASK-260916-2ok97n-platform-test.json .temp/resources/TASK-260916-2ok97n-platform-gate` | 0 | 1/1 registered case executed; 0 skips. Raw JSON and gate output are attached as `TASK-260916-2ok97n_rework4-platform-evidence.tar.gz`. |
| `go build ./...` | 0 | Full repository build passed. |
| `go vet ./...` | 0 | Full repository vet passed. |
| `golangci-lint run` | 0 | 0 issues, with the existing non-fatal warning about the missing temporary worktree `STORY-260919-37szes`. |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/scriptworker -o .temp/resources/TASK-260916-2ok97n_scriptworker_windows_rework4.test.exe` | 0 | Windows test binary cross-compiled; this is compile evidence, not hosted runtime evidence. |
| `GOOS=linux GOARCH=amd64 go test -c ./internal/scriptworker -o .temp/resources/TASK-260916-2ok97n_scriptworker_linux_rework4.test` | 0 | Linux test binary cross-compiled; this is compile evidence, not hosted runtime evidence. |
| `gofmt -d internal/scriptworker/capabilities.go internal/scriptworker/scriptworker.go internal/scriptworker/exec.go internal/scriptworker/derive_test.go internal/scriptworker/exec_test.go` | 0 | No formatting differences. |
| `git diff --check` | 0 | No whitespace errors. |

Hosted Ubuntu and Windows runtime results were not directly observed in this
developer run. Per the task contract, the handoff gate produces hosted-lane
evidence for reviewer judgment, and rose-air is observed on the landing run.
The cross-compiles above do not substitute for either runtime result.

## Revision 4 — refresh onto 1511b345

### Refresh and candidate identity

The rev-4 safety ref is `refs/campaign/r5-rev4-delta-20260924`
(`bdeaa1c824f80eec8533ed645a13007395988c33`, parent
`ac0e9774aea90ef689379ec138e9bd0132904d98`). The first requested
three-way apply returned exit **1** because the rev-4 `CHANGELOG.md` worktree
edit differed from the index. I temporarily staged only `CHANGELOG.md`, reran
`git diff fad88136 1511b345 -- . ':!.task-board' | git apply --3way`, and it
returned exit **0**. `git restore --staged -- .` returned **0**, and
`git diff --cached --quiet` returned **0**; the candidate was unstaged before
refresh.

`task-board worktree refresh-candidate TASK-260916-2ok97n` returned exit **0**:
`Outcome=refresh_advanced`, `TrunkOID=ReviewedTrunkOID=1511b345c143acfd78b5db0ab4f3176f5ce6ce94`,
and `BranchOID=601f8942c2e094fb7b8acfeb6f0e8ca164914c69`. No conflict
resolution template was needed. `git merge-base --is-ancestor 1511b345 HEAD`
returned **0**.

Candidate composition checks all returned **0**:

- Hash comparison confirmed each non-board, non-changelog path in the rev-4
  safety ref still has identical bytes in the candidate, including the R5
  implementation, tests, and results file.
- `.github/ci/gate-selftest.sh`, `.github/workflows/ci.yml`, and
  `tools/goreleaserconfig/` in the refreshed candidate match trunk `1511b345`.
- `git diff --stat refs/campaign/r5-rev4-delta-20260924 -- CHANGELOG.md`
  reports exactly 10 added lines and no removals: the trunk release-channel
  note is present while the rev-4 R5 changelog content remains intact.
- The incoming patch excluded `.task-board` as instructed. The board-managed
  refresh updated the worktree's checkout snapshot; I did not edit those
  snapshot files directly.

### Bounded refresh validation

Each command below ran as a standalone process. The combined install subset
timed out, so I split every matching install test into bounded package/test
masks and reran them successfully.

| Command/check | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/scriptworker/... -run 'Script|Enforced'` | 0 | Scriptworker subset passed in 13.748s. |
| `go test -count=1 -timeout=8m -parallel 1 ./internal/install -run 'Script|Enforced'` | 1 | Timed out at 8m while an audit-label subtest was active; not counted as passing. |
| `go test -count=1 -timeout=3m -parallel 4 ./internal/install -run '^(TestDraftAuditEnforcedScriptAdmitted|TestDraftLocalEnforcedCommandRefused|TestEnforcedScriptCommandIsRefusedAtInstall|TestDeclaredOnlySchema8ScriptCommandInstallsUnchanged|TestActiveScriptCommandWritersSplitDeclaredOnlyAndEnforced|TestActiveEnforcedScriptCommandsAdmits|TestActiveEnforcedScriptCommandsRefusesUnavailableHost|TestGlobalEnforcedInstallExcludesForwardingMirror|TestScriptOnlyInstallPerformsNoToolchainOrCacheWork)$'` | 0 | All nine passed in 129.850s. |
| `go test -count=1 -timeout=4m -parallel 4 ./internal/install -run '^TestScriptAuditLabelsAtInstallEntry$'` | 0 | All four project install audit-label shapes passed in 121.463s. |
| `go test -count=1 -timeout=4m -parallel 4 ./internal/install -run '^TestScriptAuditLabelsAtGlobalInstallEntry$'` | 0 | All four global install audit-label shapes passed in 115.614s. |
| `go test -count=1 -timeout=4m -parallel 1 ./internal/install -run '^TestEnforcedScriptCommandInstallsNativeLauncher$'` | 0 | Native launcher install passed in 80.013s. |
| `go test -count=1 -timeout=4m -parallel 4 ./internal/install -run '^TestScriptOptInCasesAtInstallEntry$'` | 0 | All six opt-in cases passed in 79.031s. |
| `go test -count=1 -timeout=4m -parallel 1 ./internal/install -run '^TestEnforcedInstallAndLaunchAtCLIEntry$'` | 0 | CLI install-and-launch passed in 130.744s. |
| `go test -count=1 -timeout=4m -parallel 1 ./internal/install -run '^TestEnforcedToDeclaredOnlyFlipRemovesNativeLauncher$'` | 0 | The enforcement-to-declared-only transition passed in 176.154s. |
| `sh .github/ci/ledger-consistency.sh` | 2 | Expected usage exit: this script requires an evidence-directory argument. |
| `sh .github/ci/ledger-consistency.sh .temp/resources/TASK-260916-2ok97n-refresh-1511-ledger` | 0 | 352 ledger rows matched compiled test inventories for Linux, Darwin, and Windows. |
| `sh .github/ci/gate-selftest.sh` | 0 | 198 passed, 0 failed, including the incoming GoReleaser gate wiring checks. |
| `go test -count=1 ./tools/goreleaserconfig/` | 0 | Incoming GoReleaser configuration gate tests passed. |
| `go build ./...` | 0 | Full repository build passed. |
| `golangci-lint run` | 0 | 0 issues. |
| `git diff --check -- . ':!.task-board'` | 0 | No whitespace errors in non-board candidate changes. |
| `gofmt -d` over changed R5 and incoming Go files | 0 | No formatting differences. |

The raw ledger inventories and report are attached as
`TASK-260916-2ok97n_refresh-1511-ledger.tar.gz`.

This refresh run did not execute hosted Ubuntu/macOS/Windows or rose-air
runtime lanes. Hosted evidence remains assigned to the handoff gate for
reviewer judgment; rose-air remains a landing-run observation.

## Revision 5 — bind Windows go-v1 evidence to the manager job

### Windows regression and fix

The downloaded Windows evidence for CI run `35936900617` contains the reported
failure in `cmd/curator`
`TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt`: the real
go-v1 seed build was rejected because the worker observed job limit flags
`0x0`, below the required `0x2308`. The issue was that
`QueryInformationJobObject(0, ...)` asks for the worker's current immediate
job. In a hosted runner with nested/ambient job context, that was not a bound
query of the manager's private Job Object.

The diff review traced the caller: the newly required CLI test goes through
the existing `godriver` build worker. The added exports in
`internal/godriver/identity.go` forward to unchanged identity helpers, while
the script worker owns a separate Windows control path; neither changes the
go-v1 Job Object lifecycle. The qualification exposed the build worker's
assumption that a null-handle query identifies its private job.

The Windows manager now confirms the suspended worker's membership and exact
limits against its private Job Object before resuming it. It duplicates a
query-only handle for that exact object into the worker; the worker checks its
own membership and reads the same object's limits before attesting evidence.
The worker closes its duplicate after confirmation, leaving the manager's
handle as the owner that preserves kill-on-close teardown. Missing and foreign
handles both refuse before the compiler starts. The foreign-handle test gives
the worker a different Job Object with matching limits, so it fails if the
membership binding is removed.

The change is confined to the Windows go-v1 control handoff. Existing script
worker behavior remains covered: Windows System32-first manager lookup and
report-only unresolved exec declarations remain unchanged and their existing
tests passed.

### Revision 5 validation and exit codes

Each local command ran as a standalone process. The ledger was first run with
the CLI build-dispatch case declared only on Darwin and Windows; it correctly
failed because that test is compiled on Linux too. After declaring all three
hosted platforms, the ledger passed.

| Command/check | Exit | Result |
| --- | ---: | --- |
| `gh run download 35936900617 -n test-evidence-windows-latest` | 0 | Downloaded the Windows `test/go-test.json`; it records the job-limit attestation regression above. |
| `go test -count=1 -timeout=5m ./internal/godriver/...` | 0 | Bounded godriver suite passed on Darwin (127.452s). |
| `go test -count=1 -timeout=5m ./internal/scriptworker/...` | 0 | Bounded scriptworker suite passed on Darwin (25.091s), including System32-first resolution and unresolved-exec reporting. |
| `go test -count=1 -timeout=5m -parallel 1 ./cmd/curator -run '^TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt$'` | 0 | Production CLI build-dispatch and receipt path passed on Darwin (36.849s). |
| `go test -count=1 -timeout=5m -parallel 1 ./cmd/curator -run '^TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache$'` | 0 | Adjacent verified build-dispatch refusal case passed on Darwin (29.617s). |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/godriver -o .temp/resources/TASK-260916-2ok97n_godriver_windows_rev5.test.exe` | 0 | Cross-compiled the Windows driver tests, including missing and foreign Job Object handle refusals. |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/scriptworker -o .temp/resources/TASK-260916-2ok97n_scriptworker_windows_rev5.test.exe` | 0 | Cross-compiled the Windows script-worker tests. |
| `GOOS=windows GOARCH=amd64 go test -c ./cmd/curator -o .temp/resources/TASK-260916-2ok97n_curator_windows_rev5.test.exe` | 0 | Cross-compiled the Windows CLI tests. |
| `bash .github/ci/ledger-consistency.sh .temp/resources/TASK-260916-2ok97n-rev5-ledger` (before ledger correction) | 1 | The new CLI dispatch row omitted Linux; the ledger rejected the compiled-but-undeclared Linux test. |
| `bash .github/ci/ledger-consistency.sh .temp/resources/TASK-260916-2ok97n-rev5-ledger` (after correction) | 0 | All 355 rows matched compiled inventories across Linux, Darwin, and Windows. |
| `go build ./...` | 0 | Full repository build passed. |
| `golangci-lint run` | 0 | 0 issues. |
| `gofmt -d` over changed Go files | 0 | No formatting differences. |
| `git diff --check -- . ':!.task-board'` | 0 | No whitespace errors in the candidate changes. |

Windows test binaries were cross-compiled but not executed on this Darwin
host. The handoff gate must supply hosted Ubuntu/macOS/Windows runtime evidence
for reviewer judgment; rose-air remains a landing-run observation. The local
cross-compile is not a substitute for Windows runtime evidence. `logbook` is
not installed in this environment; this revision's regression analysis and
decision are recorded here in the task-scoped outcome artifact.

## Revision 6 — tolerate the Linux inventory skip

The rework-5 hosted run `35944695862` failed only because
`TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt` was declared
required on Linux even though the test intentionally skips there: the
`rc5-native-control-inventory-v1` has no Linux record. Updated its ledger row
to require Darwin and Windows, tolerate Linux with the existing
`platform-control` class, and explain the host-specific inventory reason. The
ledger report now records `must=darwin,windows skip=linux` for this case. No
runtime code changed. Per rework-5, I did not run refresh-candidate because
trunk is frozen for R5.

### Revision 6 validation and exit codes

Commands ran as standalone processes. The literal ledger command requested
was run first; because the script requires an evidence-directory argument, I
then ran the same gate with a task-scoped evidence directory.

| Command/check | Exit | Result |
| --- | ---: | --- |
| `sh .github/ci/ledger-consistency.sh` | 2 | Usage failure: `ledger-consistency.sh <evidence-dir>` is required. |
| `sh .github/ci/ledger-consistency.sh .temp/resources/TASK-260916-2ok97n-rev6-ledger` | 0 | 355 rows checked across Linux, Darwin, and Windows; the target row requires Darwin/Windows and tolerates Linux. |
| `sh .github/ci/gate-selftest.sh` | 0 | 198 passed, 0 failed, including platform-control skip handling and shipped-ledger checks. |
| `go test -count=1 -timeout=5m -parallel 1 ./cmd/curator -run '^TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt$'` | 0 | Production CLI build-dispatch and receipt test passed on Darwin in 99.423s. |
| `uname -s` | 0 | `Darwin`; this was the local host for the requested test. |

No hosted lane or rose-air runtime was run in this revision. Those results
remain for the configured handoff/landing run and reviewer judgment.

## Revision 7 (artifact cleanup)

Per the binding rework-6 instruction (the only current instruction for this
revision): deleted exactly the stray untracked additions from the worktree —
the full `test/` and `ledger/` directories (a `gh run download` of a Windows
gate artifact) and the root files `TASK-260916-2ok97n_results.md` and
`TASK-260916-2ok97n_handoff-blocker.md`. No other file was touched; every
product/test/docs/CHANGELOG/`.github/ci` change of revision 6 is
byte-identical (per-file SHA-256 list below, captured before deletion and
re-verified after).

Board copies: `task-board resource get` of both documents into `$TMPDIR`
followed by `cmp -s` against the worktree files confirmed the board copies of
the results and blocker documents were already current
(`RESULTS-IN-SYNC`, `BLOCKER-IN-SYNC`), so no content change was needed for
the blocker. This Revision 7 section was appended to the results resource via
`task-board resource update ... --type outcome`. No evidence was downloaded
to the worktree in this revision.

Proof after deletion:

- `git status --short` (excluding the worktree `.task-board` checkout
  snapshot) lists only modified product/test/docs/CHANGELOG/`.github/ci`
  paths plus the three kept untracked test files
  (`internal/godriver/controls_windows_test.go`,
  `internal/scriptworker/exec_test.go`,
  `internal/scriptworker/windows_real_interpreter_test.go`); `test/`,
  `ledger/`, and the two root task documents are gone.
- `git diff --stat <story-base> -- . ':!.task-board'` contains only those
  product, test, docs, CHANGELOG, and `.github/ci` paths.
- `shasum -a 256` over every kept path matches the pre-deletion list below.

No code, test, docs, CHANGELOG, or `.github/ci` bytes changed in this
revision; no validation command was rerun (revision-6 evidence stands).
`refresh-candidate` was not run, per instruction.

### Kept-path SHA-256 (revision 6 bytes, re-verified after cleanup)
719a90f23608739c40715b6f9b20bb3e981cae93736a1cc2e49028db972403c0  .github/ci/platform-cases.tsv
b1e1e1270537d2503cc224646b75b0f539808377c5fe1c0958d5d049a33652b2  .github/ci/skip-classes.tsv
97c55451b85a14d13cf922c2d48bb1697897aa852e879540884b1b0bfdb3c064  CHANGELOG.md
415bb070434189ebdf87bfb98cd90fa2737651b24ad52eed39f515b93769b45c  cmd/curator/auditlabels_test.go
430581dd01725f5d50552861499264173ebe674a6041e97447352876323bc6fc  cmd/curator/main.go
fb61df5abf095724cdd16985153e61828a4a540e9a27546f383b747e1765f78d  cmd/curator/main_test.go
fd49643b3930fc86a9febcf75c6f51f95753ec9df5c42c7fef91c0316080722a  docs/authoring-cli-commands.md
a0c254d2c08ed55ca003afa0152c82e7cd9f44b9b5ed9d29b124c36276e23ba7  docs/cli.md
ac177b7fd0fb6aeb3c07e0b42e36063287e0588d8cbc61cb38fb7cb3a6788d0b  docs/script-interpreters.md
c93a6623ed87db1bb73f3df756144b20097cd7e65ed8630666d392cbe91e8aa0  docs/troubleshooting.md
14aee4b487b10d6232a4cc32ad146689797c9b25393c30d345c7402895f581e9  internal/audit/audit.go
44156ea038a62feabf90b31e4f451e0b0f4bf3b21004844131a823633c53c45b  internal/audit/auditlabels_test.go
4cd7d83dea706fbaea387a39421518ecfe6052e823697cb334f38676a3116512  internal/config/config.go
bac563d23077755d12c438443e64d4096fd9748e11e5832ace79c7ef4b8b9ab6  internal/config/scriptdiagnostics_test.go
21d5cac53edeaf9a778493b0d1e3a227b819108dcd82a5c8e6be44946c6ff94c  internal/config/scriptinterpreters_test.go
f29f46aa6a076c7fa3e9873c2db73d54012f1820398d8b683efa0c5235f5324d  internal/godriver/controls_darwin.go
04daaf6eb1a1d999873eeb9b228e5ff36c92560f4ed01c75c8d9f7d83ba614cd  internal/godriver/controls_other.go
a01c0b6046adbfd0726bdb7b072ac8981615450200065823dc02e3a07f8296da  internal/godriver/controls_windows.go
666cbfa54cfe2bbb62f5982042ae514c19962f8c1dea665318b66997d1506f2c  internal/godriver/identity.go
8693c725b846853763ab199994bdb853fb9c444aa6a47e6a8630d690c2efeced  internal/godriver/worker_test.go
e2bef79d327aff5f152787c9b2d1c8654666026387422428431d7fb791d93fb9  internal/godriver/workerclient.go
5090587a7a044359eb5369f8228563e0281fa053bcbeff36e91ae6382fa36b45  internal/godriver/workerproto.go
df34535071b68f26e4e21f46495391f39e1414c0e0000d146dd6e3cecfe3308e  internal/godriver/workerserver.go
89694095ecb61d513443b4352f9040ae0a44b8a8fcfc41d9ce101ef6663f70ea  internal/install/auditlabels_test.go
070abfa424a033af97df758e88faeb6a6802803dbb0de8d471fce063aad91e60  internal/install/draftaudit_test.go
fa7e790f96f7df2e2facbb4f4d9e5e7303cd4e14c975934cb1fc4cdd9b674fd4  internal/install/draftruntime_test.go
958913d3250946d449cbb0980bf4a2f380d7a784ead2870c48e7dcaa94583a40  internal/install/global.go
91f9f036977386b51cd01141bc38fd43d638d41fba7b5eb1f926a21a770f8f9d  internal/install/install.go
fd1045318b7a6da24658bc0dbb77e64516e24dd6ad5c0838fce5d0160f1da03b  internal/install/scriptpolicy_test.go
d64974eae45129741171c1a02a9b969d51e62202a7b4960d6250a1938142ecd3  internal/install/targets.go
d2339ec7ca249b0fe34c1080e96f5e656072033d4d4b9f2e4a33945f79618cc8  internal/runtimestore/enforced.go
e1c6dcd77ba7737e6c2a5a1e11842e299f4ea5a24d079ed31b778fa962e458ee  internal/runtimestore/enforced_test.go
7b2a91f7849611c0e0e79785e5f00beb2198d8fcf40528c73eb731e7320611d2  internal/scriptpolicy/conformance_test.go
fe84da76c21287bb7d486adc294a6475fda9dac6cff15858f3d5331714221bb4  internal/scriptpolicy/labels.go
0bb6e334d691ceee6b0a87f9092b3bd4a035eeb1deedd56488cac2ba41609504  internal/scriptpolicy/labels_test.go
07d8164c4e2e5a53b903935e5f3810c0ff3a3934d72fca3874bcb030d3051132  internal/scriptpolicy/scriptpolicy.go
ae67a6e9d1024d8cbb23de3f7d5f142b703d445704b4186b7d63f4fdaa14514f  internal/scriptpolicy/scriptpolicy_test.go
e115e38d959f97f4d22e774a060f9797ea994957a922a62478f194fd053b9665  internal/scriptworker/apply.go
10111758965d01d30adc5fd0feb2cc9e117fbfb2780ef56c82892d97c25ea4f9  internal/scriptworker/apply_other.go
0e7fea50a2c4300b907bd32be33387d982cb10c3f7ad2104f86a817b419f3276  internal/scriptworker/apply_unix.go
22fbf19232e6cb7ffd8a84505224f7568fc208eb26a079525cc0f4abb596c00b  internal/scriptworker/apply_windows.go
19b2ad80f17ab5f8e954329dd982eec58d989ed721ee3e26d1bc708a3a323d2c  internal/scriptworker/buffer.go
8d07ffa63dbea5f42c953dd40a8b25464d80c47b01d0d6c254099e6109d3cf39  internal/scriptworker/capabilities.go
fa2d4d583c22d90408923cbeb55e0b66eada64f78811a2adf0b2dd9a8d780924  internal/scriptworker/cgroup.go
5628e395e4093aa52f1fdbe5f1b50146c75b6dde969e02c8b3fd5740e77c75d5  internal/scriptworker/client.go
f45dccb78be8fa749f7c317a9cf06716de250d7c103be6c547a0235b3782db03  internal/scriptworker/client_test.go
cb7cbf2278a17fedcf0979713ab304a4e3bb9888ae54f084fb1da9fb52284c51  internal/scriptworker/derive_test.go
860a27c86521a74b60eebe2d3e889d8e9f89a51ab9cb3830a4c866c0c1f3577d  internal/scriptworker/evidence_test.go
b3940e3b20a7286b1f13504a889b37cbef81d3739ed572b95656da84cf807f83  internal/scriptworker/exec.go
a3aa8b9099cd5ac6db246578d8b790be2397c442c9fa812e2b723d08f27edbac  internal/scriptworker/exec_unix.go
b833739306c10f78c7138b233a63ece615d137d8ca16af83aaf5599c02e01403  internal/scriptworker/exec_windows.go
8cba9c0c729825329960dad8c49601ba3f64e8b749e73331bfea0b9e8eeb7031  internal/scriptworker/interpreter.go
14892d826eb13dd0bc29274eb9657e1376c8116ced3767046948d1f973aead86  internal/scriptworker/interpreter_test.go
25ed12c4cd7562e0b830158342e1d1f39421de099e7b860a8c826672d9ee9e35  internal/scriptworker/interpreter_unix.go
58fd47282318b89fa3e34b234486e221a4d4451c2fea3f4bf90cb7d3eb9ac128  internal/scriptworker/interpreter_windows.go
2744d7e86a580a98fadd8950636fc10ce917acd8288c050f1764388bdac553e3  internal/scriptworker/inventory.go
69a228aec7b706aa15ffdbd3e6b365d252a494ad218cc3dea69c545fa6948717  internal/scriptworker/inventory_other.go
6bc68dced8745c575a5d9f1a1908521756584af622b1681b3ad9976c008b0b6e  internal/scriptworker/inventory_unix.go
48c238fd019db5c26d6eaf5d0424d36fa3da36400a274a63a466ee4db839c306  internal/scriptworker/inventory_windows.go
ebb07d4e835cbf71e60e923f084bfad15803444edb86231940440e5f18dbd298  internal/scriptworker/landlock.go
aaeae3688caae99b0fb4df23715f8bb71ec7ae928f51856577957fc32f1c983b  internal/scriptworker/landlock_const_linux.go
33feaf31cb6aa1a73ca6074692cdc106165d87bfd30dc03aabb9149d6145328e  internal/scriptworker/landlock_const_other.go
d2ccd8fa1821c1ed29f891ac907b289b0e1ada53da81638299272d8a4f4293fd  internal/scriptworker/landlock_linux.go
793cba1471c31f116f374d1907decaeea5cc88c30fee8af9172d0cd80fa98689  internal/scriptworker/landlock_other.go
bb88d66a18ab3f9c9cf69ca90993ff30bc8e96d4a7bc120f9b129e2e2cd493a6  internal/scriptworker/landlock_test.go
06010d738d348d87643f42dc384020c40ab53e4c67acf8cc573220958a4cba41  internal/scriptworker/launcher.go
6b751414cfaaac459c8ba35199609b1e2904747ceee14b49a1296bba3f349284  internal/scriptworker/launcher_test.go
5268207d3b804cbeb9d9a27922d72fbf4516d2639e3d283286c07f292f5d1398  internal/scriptworker/main_test.go
6f91574096e7a0cc58c4dcd2ab2c0e79040fa9b5ae66261d155a177926f3c3ea  internal/scriptworker/netns_linux.go
6eb7dd61b1083ed4943ac50428672f4b8c4e46b3aabd16bcb96ab99159342a97  internal/scriptworker/netns_other.go
4ab7f55cc2f8cf670292c91163b7695a1d085ca6b87af9720fff10c870ee67b1  internal/scriptworker/preflight_test.go
5bdc7c4749d33baa295fe9aded5c8c3bb571f3a2f06582fc14948e7455410ac6  internal/scriptworker/process_alive_unix_test.go
005feedae470dd98a84ee10a240aff8f6649f7ceaf1031ac39554b86d8525640  internal/scriptworker/process_alive_windows_test.go
8c91e89b2ed1f83a06c81c07628169b9f99fa92e82b618accf5d920cf839fbed  internal/scriptworker/protocol.go
a0f376683259572d30c9c59f45b6992382e602ef38649a765fd1bf57b7d86144  internal/scriptworker/reviewer_abi1_test.go
040a6195b90bbafb35d7d67d732ead84d226f0b92aab81831761ce4509471f90  internal/scriptworker/scriptworker.go
0e32a0b3e3c36bdb8700d8c1f6316b870074d4013f565e14b52c8d855808a2e6  internal/scriptworker/teardown_unix.go
7287dc3be866abc82fe43990f1dc0b45ceb5302becd90f67d69c95b960346d75  internal/scriptworker/teardown_windows.go
5c8c0efc8d100105de08202ace1d38182753d227fb6f9e5e2301d6906d39dc4e  internal/scriptworker/testdata/forgeworker/main.go
64b477305f8110c170b3598c0cb48f6cc85962feb9bbf9a8efdf013bbd46a686  internal/scriptworker/testdata/stubinterp/main.go
a077076e8be0904de8d7773ac07ef39038c1a5c0ae2f14b2b5d9aaf0cc8049c4  internal/scriptworker/testdata/stubinterp/mkfifo_other.go
265ab42a6e22ef381cc31b687b4adbd4714e4123947fb9f4d4675addcfd3633d  internal/scriptworker/testdata/stubinterp/mkfifo_unix.go
25db73f98f6152b14c64cb15ba079f1e9ce09b4b7a83a8324381e3459a065ffb  internal/scriptworker/worker.go
11f5af6e90167d788ddb99dd0be9d61a8c17f1b3e93c6d5b2bbe2e6942950537  internal/scriptworker/worker_test.go
ab7fab60555e92e7dcc235e14ee6ec3d443bea919e850cb425ebdc17b5fc4d2d  internal/skillcheck/auditlabels_test.go
c78179d0aa643997a7421c399bae250f7bc71c39f346333de7e3a92ca7e71783  internal/skillcheck/skillcheck.go
bcc303afed19a6c8c3385794456871e02fe51691db53528fe344d261836c8c7d  internal/skillcheck/skillcheck_test.go
bbe16ec90e2b859c7d51203301259d9c22fbc96ec9c5fc066f4de6be5c9e83a7  internal/godriver/controls_windows_test.go
c937e1696dd5f563beb20591a0aca29a77b233ddc87510a4b4201ba5ad8c0b2d  internal/scriptworker/exec_test.go
0ae9dc6dc4070c14dcd953406ea05e608b22a5cd859cb2c9c2292c3a4950f1c7  internal/scriptworker/windows_real_interpreter_test.go
