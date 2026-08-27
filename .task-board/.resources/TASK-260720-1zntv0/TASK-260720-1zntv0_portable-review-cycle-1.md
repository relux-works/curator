# TASK-260720-1zntv0 independent review cycle 1

## Verdict

CHANGES REQUESTED. Route to to-dev. The successful-path implementation and native evidence are substantial, but the exact rc.5 mandatory-control failure boundary and capability-evidence truthfulness are not yet enforced for native application failures.

## R1 — native control applicability is not proved before worker start

The accepted contract requires a mandatory portable control that cannot be applied to reject with build_execution_control_unavailable before the worker starts and publish nothing. It also requires an available capability entry to report applied only when the control is actually applied.

Windows preflight does not exercise the operation that can fail at application time. controls_windows.go lines 72-92 only creates/configures an empty Job Object and allocates then deletes a PROC_THREAD_ATTRIBUTE_HANDLE_LIST. It does not assign a process to the Job Object and does not install the actual inherited-handle list used by a compiler launch. After the worker is already running, lines 130-138 create another Job Object and AssignProcessToJobObject can still fail with build_execution_control_unavailable. The inherited-handle control is appended to Applied at lines 121-128 even though its actual os/exec installation occurs later in workerserver.go lines 403-414. workerserver.go lines 94-110 emits the applied evidence before that launch. Thus a supported host can cross the required pre-worker boundary, or receive an applied record for a mechanism that later fails to install.

The same boundary is structurally open on macOS: the parent probe checks Getpgid/Kill(0) and Getrlimit, while Setsid and Setrlimit are first attempted inside the already-running worker at controls_darwin.go lines 74-102 and can return build_execution_control_unavailable there.

Required rework: make per-operation preflight prove that every platform-available control is applicable to the exact operation before worker execution begins, and arrange worker launch/application so no build_execution_control_unavailable path first appears after the worker starts. On Windows this likely requires a suspended or otherwise atomic manager-owned launch that configures the Job Object and explicit handle list before resuming the worker. Emit status applied only after the actual mechanism is installed. Add failure-injection tests for each probe/apply/install seam that assert worker_started=false, compiler_started=false, the exact stable diagnostic, no evidence claiming applied, and no output publication. The implementation must retain the accepted portable-versus-hardened boundary; do not turn normatively unavailable inventory controls or any deferred hardened guarantee into a rejection.

## R2 — Go child start failure can panic the worker

workerserver.go lines 414-424 calls command.ProcessState.ExitCode whenever command.Run returns an error. ProcessState is nil when process creation fails, including an actual Windows inherited-handle-list installation failure or a launcher replacement race after verification, so the worker can panic instead of returning a stable bounded result.

Required rework: handle a nil ProcessState without dereference, preserve the correct stable diagnostic and started-count semantics, and add an injected child-start failure test proving no panic, no build continuation, complete domain teardown, and no artifact publication.

## Independent evidence

Provenance: accepted conformance manifest sha256:58f8d2299c8f4a5ed78546913e567f637f1cc905dc5212b460f4097be7ff2af9 and host-execution vector sha256:c3d42f763afdcfa229430e7de5bb9f1e9f44607a7790aef6f4e0bf6d1bc644de. Assigned worktree remains at 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8 with zero later commits, nothing staged, and git diff --check clean. Comparison with TASK-260720-6i3cya confirms the task delta is limited to internal/godriver, internal/buildmeta, and cmd/curator. No code, candidate, release, pin, ref, or provenance file was modified during review.

Green local gates: go build ./..., go vet ./..., go test ./... -count=1, go test -race ./... -count=1, rc.5 focused conformance plus real vendored Go build, focused identity/protocol/package-influence/domain-teardown/RLIMIT/evidence adversarial tests, gofmt check, and scoped golangci-lint with 0 issues.

Green native gates on byte-matched task-owned source: macOS amd64 task packages, real vendored build, and full vet; Windows amd64 task packages, real vendored build, and full vet. Local macOS arm64 full and race gates are green.

Honest external reds: Windows go test ./... reproduces failures only in unchanged predecessor packages buildcache, buildsource, globalbins, runtimestore, and shell; task-owned packages pass. Whole-repository golangci-lint reports 45 findings outside the task delta, while the task-scoped lint command reports 0 issues. These reds are not the reason for this verdict. Linux remains compile/mock-only as allowed by rc5-native-control-inventory-v1.
