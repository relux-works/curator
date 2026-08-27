# TASK-260720-1zntv0 cycle 2 — closing review findings R1 and R2

Rework of the portable `manager-worker-v1` go-v1 preflight and build in the
existing producer worktree. This document is the producer handoff; it makes no
acceptance judgement about its own work.

## 1. Provenance and boundaries

- Assigned worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
  (retained; no new worktree was created and no state was reimported).
- `HEAD` is still exactly `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`.
  `git rev-list --count 17804cea..HEAD` = `0`. `git diff --cached --quiet` exits
  `0`. `git diff --check` exits `0`. Nothing was staged, committed, tagged,
  pushed, or published.
- Authoritative conformance input, read only, re-verified after the rework:
  `/Users/iv/Developer/ReluxWorks/.temp/TASK-260728-zb2s4z/curator-spec-worktree`
  - `conformance/v1/manifest.json` =
    `sha256:58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9`
  - `conformance/v1/vectors/go-host-execution-policy.json` =
    `sha256:c3d42f763afdcfa229430e7de5bb9f1e9f44607a7790aef6f4e0bf6d1bc644de`
- The pre-revision rc.4 candidate at
  `.temp/TASK-260720-3ag6pi/worktree/conformance/v1` was consumed only as
  additional test input.
- Handed-off `internal/godriver` tree digest (SHA-256 over the sorted per-file
  SHA-256 list): `55e23e323a84ff76be4768ed664f08a61cb2975cbc4c5bb50bb77a9fdeccda19`.
- No cache was published, no marker written, no live install mutated, no ref,
  tag, pin, release or platform claim created or edited, and no built artifact
  was ever executed.

## 2. R1 — every native control is installed by the parent before the worker runs

The reviewed implementation applied native controls **inside** the already
running worker, so a supported host could cross the required pre-worker boundary
and an `applied` status could be emitted for a mechanism that had not been
installed. Control ownership moved to the manager parent.

### 2.1 New boundary rule, enforced in code

> Every mandatory portable control, including every platform-available inventory
> control, is proved applicable and installed by the manager parent before the
> worker executes. Only the parent, before worker execution, emits
> `build_execution_control_unavailable`. Once the worker is running, a
> contradiction between what the parent installed and what the worker observes is
> `build_execution_capability_evidence_invalid`, and a refusal at the worker's
> single permitted process-creation site is a bounded child result — never a
> control-availability claim.

A source audit of `internal/godriver` shows every remaining
`CodeControlUnavailable` site is in the parent (`probeNativeControlsFor`,
`prepareControlDomain`, `controlDomain.launch`, `controlDomain.attach`,
`workerClient.destroyBeforeExecution`) or in the unreachable non-macOS,
non-Windows stub. `workerserver.go` and `observeNativeControls` contain none.
`rawWorker.expectFailure` fails any test whose worker-side rejection is that
code, so the invariant is checked on every worker rejection scenario.

### 2.2 Per-operation preflight now performs the real operation

`probeNativeControls` takes the operation's normalized `ResourceLimits` and each
probe performs the exact operation the control will perform:

| Control | macOS probe | Windows probe |
| --- | --- | --- |
| `descendant-domain-termination` | `Getpgid` plus `Kill(-group, 0)` permission check | real Job Object with `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` |
| `active-process-count-limit` | normatively unavailable | real Job Object with the **exact** `ActiveProcessLimit` of this operation |
| `aggregate-memory-limit` | normatively unavailable | real Job Object with the **exact** job and process memory limits |
| `per-file-size-limit` | `RLIMIT_FSIZE` **set to the exact byte bound, verified, restored** | normatively unavailable |
| `inherited-handle-restriction` | `F_GETFD`/`F_SETFD` `FD_CLOEXEC` round trip | `NewProcThreadAttributeList`, the primitive `os/exec` uses |

### 2.3 Installation is atomic with the launch

`controlDomain` is created before any process exists and owns the launch.

- **Windows.** The parent creates the private Job Object during
  `prepareControlDomain`, creates the worker with `CREATE_SUSPENDED`, opens it
  with `PROCESS_SET_QUOTA|PROCESS_TERMINATE`, calls `AssignProcessToJobObject`,
  and only then resumes every thread of the process
  (`CreateToolhelp32Snapshot`/`Thread32First`/`OpenThread`/`ResumeThread`). Any
  failure terminates the worker **while it is still suspended**, so it never
  executes an instruction. The explicit `PROC_THREAD_ATTRIBUTE_HANDLE_LIST` is
  installed by process creation itself, so it is real before the worker runs.
  The parent now holds the job handle for the domain's lifetime and closes it
  last during teardown, so kill-on-close terminates anything still inside.
- **macOS.** The private session is created by the kernel between fork and exec
  via `SysProcAttr.Setsid`, and `RLIMIT_FSIZE` is lowered in the manager across
  the fork under a mutex and restored as soon as the child exists, so the worker
  and every descendant inherit it. There is therefore no post-`Start` install
  step at all on macOS; the last seam is crossed before `fork`.

`controlDomain.installedControls()` returns nothing until the domain is
installed, so an `applied` status cannot exist before its mechanism does. The
manager derives the `capability-evidence-v1` record from that installed set and
validates it **before** the request is even sent.

### 2.4 The worker only confirms

`applyNativeControls` is gone. `observeNativeControls` confirms that what the
parent installed really governs this process:

- macOS: `Getpgid(0) == getpid()`; `RLIMIT_FSIZE.Cur` equals the installed
  bound; `FD_CLOEXEC` set and verified on the protocol descriptors.
- Windows: `QueryInformationJobObject` on the worker's own job requires the exact
  installed limit flags, `ActiveProcessLimit` and `JobMemoryLimit`; the attribute
  list is allocatable in this process.

The worker derives the same closed record independently and the parent requires
the two to be **identical** (`reflect.DeepEqual`) plus a matching applied-control
report. A difference is `build_execution_capability_evidence_invalid`.

The parent additionally proves at runtime that a worker destroyed before its
domain was installed produced **no session output**, and says so in the
diagnostic if it did.

### 2.5 Failure-injection tests for each seam

`controlSeamFault` is a test-only injection point at three named seams —
`availability-probe`, `domain-preparation`, `domain-installation`. It is `nil` in
production.

`TestEveryControlSeamFailureRejectsBeforeTheWorkerExecutes` drives a complete
`Build` with a real failure at each seam and requires, for every one:

- the seam was actually crossed (the fault fired);
- `DiagnosticCode(err) == build_execution_control_unavailable`;
- the diagnostic does **not** report worker session output, i.e. the worker never
  ran a session (`worker_started=false`);
- the stub launcher recorded no new invocation, i.e. `compiler_started=false`;
- the returned `Result.Evidence` is empty — no record, no `applied` entry;
- the returned `Result.Artifact` is empty and no staged output survives anywhere
  under the operation-private root (`published=false`).

`TestControlSeamFailureNeverDriftsIntoAHardenedClaim` proves the accepted
portable-versus-hardened boundary survived: after the injected rejection, the
same fixture builds successfully while at least one inventory control is
normatively unavailable on the platform, and no deferred hardened guarantee
appears in the record.

`TestControlDomainInstallsBeforeTheWorkerIsReleased` proves the ordering
directly: a prepared-but-uninstalled domain reports no installed control, the
record derived from it has no `applied` entry and is rejected, and a launched
domain reports exactly the probed-available set.

`TestParentRejectsAWorkerRecordThatDiffersFromWhatWasInstalled` proves the
manager owns the applied status.

## 3. R2 — a refused Go child no longer dereferences a nil process state

`runGo` read `command.ProcessState.ExitCode()` whenever `command.Run` returned an
error. `ProcessState` is `nil` when process creation fails, so the worker could
panic.

- `runGo` now returns `(*workerResult, error)`. A launcher that fails its
  identity check ends the session with `build_execution_worker_identity_invalid`
  and creates no process — previously that path returned a result with
  `Started == 0`, which the parent misreported as a process-graph violation.
- A refused creation sets `workerResult.StartFailed`, `ExitCode = -1` and the
  refusal detail, and never touches `ProcessState`.
- `Started` semantics are stated and preserved: it counts how many times the
  worker used its **single permitted process-creation site**, so a refused
  attempt still consumes one and an extra program below the worker is still
  detected. The parent's exact-count guard is unchanged.
- The parent turns a `StartFailed` result into an uncoded error, which
  `classifyListFailure`/`classifyBuildFailure` map to the existing stable
  `go_list_failed`/`go_build_failed`. No new diagnostic code was invented.
  The list phase returns before any permit is issued, so a refused list cannot
  continue into a build.

Tests:

- `TestWorkerReturnsABoundedResultWhenTheGoChildCannotStart` injects a **real**
  refusal against the shipped worker subprocess: the manager-owned working
  directory is inside the frozen snapshot but does not exist, so the operating
  system refuses `chdir` and leaves no process state. It requires
  `StartFailed`, `ExitCode == -1`, `Started == 1`, empty captured output, no
  artifact, a second identically bounded refusal for the build phase, and a
  clean worker exit — proving no panic. `StartFailed` can only be set when
  `ProcessState` was nil, so this is positive evidence the seam is the one the
  review named.
- `TestBuildFailsClosedWhenTheGoChildCannotStart` drives the refusal through a
  complete `Build` on Windows using the native `active-process-count-limit` set
  to admit the worker alone, so the kernel refuses the compiler. It requires
  `go_list_failed`, no Go child invocation, and no staged output. It skips on
  macOS, where no inventory control can refuse a child natively.
- `TestWorkerDomainTeardownSurvivesARefusedChild` requires the complete domain to
  be terminated and joined and no artifact to exist after a refused creation.
- `TestParentClassifiesAChildStartFailureWithoutBuildContinuation` covers the
  parent half on every platform from a crafted frame.

## 4. Other changes in this cycle

- `validateProbeSet` inside the worker now reports
  `build_execution_capability_evidence_invalid` instead of
  `build_execution_control_unavailable`. It is defence in depth over a tampered
  request; the normative availability decision already happened in the parent.
  This is consistent with the accepted
  `capability_evidence_record.consistency_rules` entry
  `availability-probed-per-operation-before-worker-launch`.
  `TestWorkerRejectsAProbeThatDropsAPlatformAvailableControl` and
  `TestProbeSetGuardMatchesTheInventoryExactly` were updated accordingly.
- `startRawWorker` launches through the real control domain, so every raw
  identity/protocol test now runs against a worker governed by the real
  controls rather than an unconstrained process.
- The dead write-only `workerClient.applied` field was removed.

Nothing in the inventory, the mandatory-control set, the evidence vocabulary,
the deferred-guarantee list, the failure boundary, the process graph, the session
states, the argument vectors, the closed environment or the cache identity
changed. The whole rc.5 conformance suite is unchanged and green.

## 5. Files

New: `internal/godriver/boundary_test.go`.

Rewritten: `internal/godriver/controls_darwin.go`,
`internal/godriver/controls_windows.go`, `internal/godriver/controls_other.go`.

Changed: `internal/godriver/controls.go` (control seams, limits-aware probe,
`installableControls`), `internal/godriver/workerclient.go` (domain ownership,
parent-derived evidence, cross-check, `destroyBeforeExecution`, `StartFailed`
classification, domain teardown), `internal/godriver/workerserver.go`
(observe instead of apply, `runGo` error return, nil-safe process state,
evidence-fault probe codes), `internal/godriver/workerproto.go`
(`StartFailed`, documented `Started` semantics), `internal/godriver/build.go`
(limits-aware probe), `internal/godriver/worker_test.go`,
`internal/godriver/guards_test.go`.

No file outside `internal/godriver` was changed in this cycle.

## 6. Gates and exact exit codes

Every command was run directly as a standalone process. No gate was piped
through `tee` or a pipe chain. All gates below ran on the exact handed-off bytes
(`internal/godriver` digest `55e23e32…`). Logs:
`.temp/TASK-260720-1zntv0/logs/cycle2/`.

### 6.1 macOS arm64 — local host (Darwin 25.5.0, Go 1.25.5 darwin/arm64)

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `gofmt -l internal cmd` (empty output) | 0 |
| `go test ./... -count=1` | 0 |
| `go test -race ./... -count=1` | 0 |
| `go test ./internal/godriver/ -count=1` | 0 |
| `CURATOR_CONFORMANCE_ROOT=<rc.5> CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver/ ./internal/buildmeta/ -count=1` | 0 |
| `CURATOR_CONFORMANCE_ROOT=<rc.4 candidate> go test ./internal/godriver/ ./internal/buildmeta/ -count=1` | 0 |
| `CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver -run TestRealGoV1VendoredBuildIsBoundedAndNotLaunched -count=1 -v` | 0 |
| `golangci-lint run ./internal/godriver/... ./internal/buildmeta/... ./cmd/curator/...` — `0 issues.` | 0 |
| `GOOS=windows GOARCH=amd64 go test -c ./internal/godriver/` | 0 |
| `GOOS=linux GOARCH=amd64 go test -c ./internal/godriver/` | 0 |
| `GOOS=linux GOARCH=amd64 go build ./...` | 0 |
| `GOOS=linux GOARCH=arm64 go build ./...` | 0 |
| `GOOS=linux GOARCH=amd64 go vet ./internal/godriver/` | 0 |
| `git diff --check` | 0 |
| `git diff --cached --quiet` | 0 |

### 6.2 macOS amd64 — `ssh relux` (macOS 15.7.4, Go 1.25.5 darwin/amd64)

Source shipped as a vendored tarball and proved byte-identical: the 143 sorted
`cmd`+`internal` `*.go` path/SHA-256 pairs `diff` clean against the local tree.

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `go test ./... -count=1` | 0 |
| `CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver -run TestRealGoV1VendoredBuildIsBoundedAndNotLaunched -count=1` | 0 |
| focused seam, domain, refused-child, teardown, `RLIMIT_FSIZE` and parent-classification tests, `-v` | 0 |

All three seam subtests, `TestPerFileSizeLimitIsReallyApplied`,
`TestBuildTerminatesTheCompleteWorkerDomain`,
`TestWorkerReturnsABoundedResultWhenTheGoChildCannotStart` and
`TestWorkerDomainTeardownSurvivesARefusedChild` pass natively.

### 6.3 Windows amd64 — `ssh win` (Go 1.25.5 windows/amd64, Git 2.51.0)

Source proved byte-identical the same way (143 path/SHA-256 pairs, `diff`
clean). Exit codes were captured at runtime with `&& echo EXIT_0 || echo
EXIT_NONZERO`; the earlier `%ERRORLEVEL%` form expands at parse time and was
discarded as meaningless rather than reported.

| Command | Exit |
| --- | --- |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `CURATOR_CONFORMANCE_ROOT=<rc.5> CURATOR_REAL_GO_BUILD_TEST=1 go test ./internal/godriver/ ./internal/buildmeta/ ./cmd/curator/ -count=1 -v` | 0 |
| `go test ./... -count=1` | **non-zero** — see 6.4 |

Windows is the platform the review's R1 and R2 both named, and both are proved
natively there:

- `TestEveryControlSeamFailureRejectsBeforeTheWorkerExecutes/domain-installation`
  exercises the real suspended-launch path: the worker is created suspended, the
  install fails, and it is terminated before it is ever resumed.
- `TestBuildFailsClosedWhenTheGoChildCannotStart` passes: the real Job Object
  active-process limit refuses the compiler creation, the worker returns a
  bounded result instead of panicking, and the build fails closed with nothing
  published.
- `TestRealGoV1VendoredBuildIsBoundedAndNotLaunched` passes in 44.84 s, so the
  suspended-launch and job-assignment path does not break a real vendored build.

### 6.4 Reported red: pre-existing Windows gaps outside this task

`go test ./...` on Windows exits non-zero. The failing packages are exactly the
five recorded in cycle 1 and none of them is created or modified by this task:
`internal/buildcache`, `internal/buildsource`, `internal/globalbins`,
`internal/runtimestore`, `internal/shell`. `internal/godriver`,
`internal/buildmeta` and `cmd/curator` pass. These are reported, not fixed,
because they are outside this task's ownership.

### 6.5 Expected-red gates during this cycle, reported truthfully as failures

1. `go test ./internal/godriver/ -count=1` after moving control installation into
   the parent: exit 1, one failure —
   `TestWorkerRejectsEveryIdentityAndRequestMutation/probe_claims_an_unavailable_control`
   with `no macOS mechanism for inventory control "active-process-count-limit"`.
   Root cause: the test scenario's request and the domain's probe set shared one
   backing array, so tampering with what the worker was *told* also changed what
   the manager *installed*. Closed by giving the scenario an independent copy,
   not by weakening the assertion.
2. First Windows `CURATOR_REAL_GO_BUILD_TEST=1` run: the real-build test skipped.
   Root cause: `set VAR=1 && …` in `cmd.exe` stores the trailing space, so the
   value was `"1 "`. Closed with the quoted `set "VAR=1"` form; the test then
   ran and passed.
3. First Windows exit-code capture: `cmd … & echo %ERRORLEVEL%` expands at parse
   time and always printed `0`. Those readings were discarded and every Windows
   gate was re-run with runtime `&&`/`||` capture. That is how 6.3 and 6.4 were
   produced, and it is why the honest non-zero full-suite result in 6.4 is
   reported rather than a false `0`.

### 6.6 Not run, and why

- **Linux runtime.** `rc5-native-control-inventory-v1` covers exactly macOS and
  Windows, so the portable policy is not defined for Linux. The implementation
  rejects there before the worker starts, which
  `TestProbeRejectsAnUncoveredPlatformBeforeTheWorker` proves. Linux coverage in
  this task is compile, vet, test-binary link and that rejection path only. No
  Linux host was used and none is claimed.
- **`make lint` as configured.** The Makefile target runs a repository-wide
  `golangci-lint run`, which still exits non-zero on findings in packages this
  task does not own. The scoped result (`0 issues.`, exit 0) is reported instead;
  no green is claimed for the repository-wide target.

## 7. Boundaries preserved

- No cache published, no marker written, no live install mutated, no ref, tag,
  pin, release or platform claim created or edited.
- The built artifact is never started.
- No package-selected program, argv, environment value, output path, flag, hook,
  plugin or generator is reachable; the package build-command surface is
  unchanged and still closed.
- The portable profile still claims none of the six deferred hardened
  guarantees, and their absence still rejects nothing.
- The conformance candidate worktrees and the rc.4 candidate were read only.
