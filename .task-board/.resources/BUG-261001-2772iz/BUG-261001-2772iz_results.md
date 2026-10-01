# BUG-261001-2772iz — askpass-pipe-refusal-epipe

Ready for review. Changes remain uncommitted in the Story worktree. Recovery run: RUN-261001-8b3ce4.

## Cause and fix

Production `runCommandWithHTTPSSecret` in `internal/buildrepo/admission.go` starts the fetch child, closes the parent copy of the inherited endpoint, and runs `Serve` concurrently with `cmd.Wait`. The original Unix transport eagerly wrote the secret to `os.Pipe`; when an askpass refusal exited before that write, there were no readers and the write returned EPIPE.

The Unix transport now uses a private inherited AF_UNIX socket pair. `readHTTPSBrokerSecret` sends one request byte only after `RunHTTPSCredentialBroker` accepts the exact password prompt and validates the state. `Serve` waits for that request before writing secret bytes. EOF before any request means no secret was needed. The caller independently preserves the child's exit status, so silent refusal remains exit 1 and cannot become authentication success. Username prompts require no secret delivery.

Request-before-write avoids trying to infer a refusal from exit status 1, which can also mean a real delivery failure. Requested write failures are still returned, invalid requests deliver no secret, and cancellation suppresses only the closed-file error caused by interrupting the transport operation. Descriptors are close-on-exec under ForkLock, with only the child endpoint explicitly inherited. Nonblocking descriptors allow Close to interrupt pending I/O. Windows retains its existing named-pipe protocol.

No additional secret sink, state field, environment secret, or production logging was added. Existing tests cover secret-free state, redacted diagnostics, legacy environment refusal, compiled-binary dispatch, child environment secrecy, and the real Git HTTPS flow. New negative tests cover requested-write EPIPE with the reading endpoint open, an invalid request with zero delivered bytes, password refusal on a closed socket, and cancellation without a request. CHANGELOG.md has one line under Unreleased. LOGBOOK.md is untouched by explicit task instruction; findings are recorded on the board instead.

## Deterministic reproduction and mutant

`TestHTTPSBrokerSecretTransportChildExitsBeforeServe` launches the existing TestMain askpass basename dispatch into the production `RunHTTPSCredentialBroker`, closes the parent's inherited endpoint with `ChildStarted`, and waits for the real child to exit before invoking `Serve` with a live context. This forces child-exit-before-write without sleeps, race instrumentation, or scheduler luck. The context is only a hang bound; expiration fails the test. Child status and stdout/stderr are asserted before serving.

The prior attached developer outcome records the initial unchanged-main reproduction at baseline HEAD `54d4aface89e22bf4a4116505eba3ecfe668efa8` (main independently verified in that run): exit 1, EPIPE in all four incident refusal cases plus username. That initial baseline evidence is accepted from the preceding run, not claimed as newly executed here.

This recovery run independently replaced the entire Unix transport source with `git show HEAD:internal/buildrepo/httpsbroker_pipe_unix.go` and ran the deterministic regression. The mutant compiled and failed with EPIPE in 5/5 cases, exit 1. Restoring the saved fixed source and verifying byte equality with `cmp` yielded exit 0 for the same regression. No mutant remains. The compiled production binary is separately exercised by the required crossconformance race x50 command.

## Recovery of the hosted validation failure

The previous automated Change Request gate, `sh scripts/remote-gate.sh`, exited 1 in hosted run 36837352988. Its macOS Race and Test lanes passed; Linux Test and Race failed. Inspection of the downloaded Linux JSON artifacts identified the in-process password fixture, and the Race artifact also identified the empty-password fixture.

Those fixtures wrapped a nonblocking descriptor that an existing `os.File` already owned. Linux cannot register the same descriptor with epoll twice; the second wrapper falls back to direct reads and can return EAGAIN before the secret is available. A standalone Linux diagnostic that double-wraps a socket before any server write deterministically returned `resource temporarily unavailable`. The original Linux password fixture passed one single run, then failed 68/100 repetitions (command exit 1).

The broker prompt fixtures now launch the materialized askpass child with the real inherited transport, matching production descriptor ownership. The closed-socket negative test also uses the real child. The obsolete platform-specific helpers exposing an already-owned descriptor were removed. Refusal and password assertions remain, stderr must stay empty, and test failures redact credential output. The same Linux password mask then passed 100/100 repetitions (exit 0).

The first full Linux package race run exited 1 because the Docker mount exposed the worktree's `.git` file but omitted its absolute metadata target. The real-Git test reported `fatal: not a git repository` for that target. Mounting the actual Git metadata read-only fixed the environment: the exact affected test passed separately, then the full package race command passed. No source change was used to address this environment failure.

## Validation and limits

Host: Go 1.26.0 darwin/amd64. Linux runtime checks used the Go 1.26.0 Docker image in Colima on linux/amd64. All listed commands were executed directly as standalone processes and their actual exit codes were observed; no gate was piped through tee. Docker returned the enclosed go test process status. Raw evidence and exact commands are in the new task-scoped `BUG-261001-2772iz_recovery-validation.md` outcome.

| Validation | Real exit | Outcome |
| --- | --- | --- |
| Existing focused Darwin HTTPS tests before fixture changes | 0 | Passed. |
| Original Linux password fixture, count=1 | 0 | Passed once; not evidence of stability. |
| Original Linux password fixture, count=100 | 1 | Failed 68/100 repetitions. |
| Fixed Linux password fixture, count=100 | 0 | Passed 100/100 repetitions. |
| Focused Darwin HTTPS tests after fixture changes | 0 | Passed. |
| Original-transport mutant, deterministic child-exit regression | 1 | Expected red: EPIPE in 5/5 cases. |
| Restored fix, same deterministic regression | 0 | Passed. |
| `go test -race ./internal/crossconformance -run TestDraftSourcesBrokerAskpassDispatch -count=50` | 0 | Darwin x50 green, 92.877s. |
| `go test -race ./internal/buildrepo ./internal/testcli -count=1` | 0 | Darwin buildrepo green, 284.240s; testcli has no test files. |
| Same package race command in Linux Docker, without metadata mount | 1 | Environment failure in the real-Git test; retained as failing evidence. |
| Exact affected real-Git test in Linux Docker with read-only metadata mount | 0 | Passed, 5.877s. |
| Same full package race command in Linux Docker with read-only metadata mount | 0 | Linux buildrepo green, 45.437s; testcli has no test files. |
| `golangci-lint run` | 0 | 0 issues. |
| `go vet ./...` | 0 | Passed. |
| `go build -o /tmp/BUG-261001-2772iz-recovery/curator-darwin ./cmd/curator` | 0 | Native CLI build passed. |
| Windows `go test -c` for internal/buildrepo | 0 | Test compilation passed; no Windows runtime claim. |
| `git diff --check` | 0 | Passed. |
| Changed-file gofmt validation | 0 | Passed. |
| `git diff --exit-code -- LOGBOOK.md` | 0 | Untouched. |

No full repository suite or successful replay of the hosted matrix is claimed by this recovery run. The next automatic Change Request validation remains the configured hosted gate. Initial main reproduction and the previous hosted failure are explicitly accepted from already-attached evidence; all validation listed above was rerun in this recovery session.

## Current source identity

- `CHANGELOG.md` SHA-256 `6fc83450332200b8dc2ac718f32eca171a0099fdd7c6f8f7fd3710bd6e86786a`
- `internal/buildrepo/httpsbroker_pipe_unix.go` SHA-256 `5b922ad8362fdbeab1a03958ad76c6478a27c78d58aa74395d3fb4f368384cfd`
- `internal/buildrepo/httpsbroker_pipe_unix_test.go` SHA-256 `38f4bfc5c8dd332109e0ff1869d15e9e6b3a3afc947817b9f374bbac32d2d971`
- `internal/buildrepo/httpsbroker_test.go` SHA-256 `444ffd3cfd98dee7426cbe99746b2d9aab91e8365494f1701d127a0978d045ea`
- `internal/buildrepo/httpsbroker_test_pipe_unix_test.go` SHA-256 `c6f19bca01e6dec754823fb4f124afe81c91ab02f4b697f85bd43725370a1b2d`
- `internal/buildrepo/httpsbroker_test_pipe_windows_test.go` SHA-256 `626f9002e2a34b0a90701185e34685fb1d7faef747e651a45531c76578b63316`
