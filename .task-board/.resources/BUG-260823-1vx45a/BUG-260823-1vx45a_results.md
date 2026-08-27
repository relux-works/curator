# BUG-260823-1vx45a developer outcome

## Root cause

`TestManagerLockHelper` used one 200ms context for both expected-blocked and expected-acquired subprocess probes. The context covered lock-state directory creation, file opening, retry work, and two sequential acquisitions in `try-key` mode. The helper serialized every `context.DeadlineExceeded` as `blocked`, so slow uncontended acquisition and real cross-process contention were indistinguishable. Native Windows reproduction with a 1ns deadline returned `blocked` on an uncontended lock. No evidence supported the rejected revision-1 lock-handle/t.TempDir cleanup theory.

## Implementation

- Parent helper call sites explicitly state the expected outcome.
- Expected-blocked probes retain a 200ms deadline; expected-acquired probes use 30s.
- `TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked` drives the real subprocess entry point against an uncontended lock with a 1ns deadline and requires `blocked`. It also requires the acquired deadline to stay strictly wider than the blocked deadline.
- Helper locks are released before result classification. Any release error fails the helper independently and cannot be joined with an acquisition timeout and mislabeled `blocked`.
- Windows-only helper call sites were updated. All tests remain under `t.TempDir`; none were skipped or moved.
- The unrelated 50ms timeout in `TestCancellationReleasesPartialAcquisition` remains correct: the second lock is deliberately held and cancellation is the expected negative outcome.

## Validation ledger

| Gate | Platform | Exit | Result |
| --- | --- | ---: | --- |
| `go test -count=20 ./internal/managerlock -run '^TestSubprocess.*$'` | macOS arm64, Go 1.25.5 | 0 | pass, 20 consecutive runs |
| `go test -count=5 ./internal/managerlock` | macOS arm64, Go 1.25.5 | 0 | pass, 5 consecutive package runs |
| `go test -race -count=3 ./internal/managerlock -run '^TestSubprocess.*$'` | macOS arm64, Go 1.25.5 | 0 | pass, 3 consecutive race runs |
| same subprocess command | Ubuntu Linux arm64, Go 1.25.5 | 0 | pass, 20 consecutive runs |
| same subprocess command | native Windows amd64, Go 1.25.5 | 0 | pass, 20 consecutive runs |
| `go vet ./...` | macOS | 0 | pass |
| `golangci-lint run` | macOS | 0 | pass, 0 issues |
| `go build ./...` | macOS | 0 | pass |
| narrowing mutant: acquired deadline reduced to 200ms | macOS | 1 | expected red; regression rejected equal acquired/blocked deadlines |
| restored regression rerun | macOS | 0 | pass |
| `go test ./...` | macOS | 143 | interrupted after approximately seven minutes with no package output, before the single-call ceiling; not reported as passing |

Linux used an isolated Go 1.25.5 toolchain under `/tmp/BUG-260823-1vx45a-go-20260827`; its official `go.dev` downloads JSON SHA256 was checked before extraction. Windows ran a task-scoped source snapshot under `C:\Users\admin\AppData\Local\Temp\BUG-260823-1vx45a-codex-20260827`. Windows CI publication remains owned by the orchestrator as stated in the revision brief.

The final macOS, Linux, and Windows copies matched byte-for-byte: `managerlock_test.go` SHA256 `ffd2bbffd30842f07960052efed39c52687bf301d9340a8f71b56b85152ae9b2`; `identity_windows_test.go` SHA256 `0f7a3a263e140d03032d243f80acba29923209be62ef423ed4f4b07242323b1f`.

Logs are under `.temp/BUG-260823-1vx45a/`.
