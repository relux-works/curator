# TASK-260729-2kaopg tester evidence

Date: 2026-07-29
Worktree: `.temp/TASK-260729-2kaopg/worktree`
Role: tester

## Test change

Added `TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved` to
`cmd/curator/global_status_test.go`. It pins the fail-closed `--check` contract
when the read-only global plan cannot resolve the user home whose adapter and
shim state it must inspect, while preserving plain status's historical zero
exit and warning surface.

## Focused tests

- `go test -count=1 ./cmd/curator -run '^TestGlobalStatusReportsATransitivelyResolvedCompiledCommand$' -v`
  - exit 0; 26.074s.
- `go test -count=1 ./cmd/curator -run '^TestGlobalStatus' -v`
  - exit 0; 144.477s.
  - Covers current state; source, input, command, cache, and context drift;
    unusable toolchain; transitive compiled commands; no-compiled-command
    compatibility; empty global scope; unprovable closure; positional rejection.
- `go test -count=1 ./cmd/curator -run '^TestGlobalStatusFailsCheckWhenTheUserHomeCannotBeResolved$' -v`
  - exit 0; 0.542s.
- `go test -count=1 ./internal/install/atomicity -v`
  - exit 0; 371.708s.
  - Reproduced the package implicated in the prior disk-exhaustion run without
    the cacheable syscall-log growth.

## Coverage

- `go test -count=1 -coverprofile=.temp/TASK-260729-2kaopg/global-status.cover ./cmd/curator -run '^TestGlobalStatus'`
  - pre-change measurement: exit 0; 162.122s; broad package coverage 27.4%.
  - post-change measurement: exit 0; 137.736s; broad package coverage 27.7%.
- `go tool cover -func=.temp/TASK-260729-2kaopg/global-status.cover`
  - exit 0.
  - Affected functions: `cmdGlobalStatus` 96.9%, `statusReport` 86.2%,
    `globalStatusPlan` 100%, `globalStatusScope` 100%, `checkFailed` 100%,
    `factsOf` 100%, `Describe` 100%, `plannedRows` 100%, and `markerMoved`
    100%.

## Repository gates

- `git diff --check` (rerun after the test edit)
  - exit 0.
- `gofmt -l cmd/curator/global_status_test.go cmd/curator/builds_test.go cmd/curator/builds.go cmd/curator/main.go`
  (rerun after the test edit)
  - exit 0; no output.
- `go build ./...` (post-change)
  - exit 0.
- `go vet ./...` (post-change)
  - exit 0.
- `golangci-lint --version`
  - exit 127; executable unavailable. No installation attempted because the
    task directive forbids host-software installation. Preserved implementer
    artifact `../logs/lint-02.log` records `0 issues.`

## Default-timeout full-suite anomaly

`go test -count=1 ./...` was run twice as a standalone foreground process with
disk monitoring. Both runs truthfully failed with exit 1 at Go's fixed
10-minute `cmd/curator` package timeout:

1. First run: `cmd/curator` timed out at 602.465s while
   `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` had run
   12s. All other listed packages passed. A separate predecessor-worktree
   `go test -race` was active. The alleged victim passed in isolation:
   `go test -count=1 ./cmd/curator -run
   '^TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall$' -v`
   exited 0 in 39.955s.
2. Clean-load replay: a process probe confirmed no other `go test`; the package
   timed out at 601.057s while the later
   `TestDryRunNeverClaimsACompletedCompilerCheck` had run 6s. All other listed
   packages again passed. The alleged victim passed in isolation:
   `go test -count=1 ./cmd/curator -run
   '^TestDryRunNeverClaimsACompletedCompilerCheck$' -v` exited 0 in 9.334s.

This is aggregate package duration, not a semantic test failure. It is recorded
in `LOGBOOK.md` with the recommendation to set an explicit CI timeout or split
the integration package. The default-timeout command is not reported green.

Disk did not exhaust under `-count=1`:

- first full run: 15,680,576 KiB before; 15,079,720 KiB after;
- clean-load run: 14,785,080 KiB before; 14,575,380 KiB after.

## Extended-timeout semantic gate

- `go test -count=1 -timeout 30m ./...`
  - exit 0.
  - `cmd/curator`: 572.541s.
  - `internal/install`: 289.268s.
  - `internal/install/atomicity`: 422.822s.
  - Every listed repository package passed.
- Disk: 14,545,288 KiB before; 14,360,600 KiB after.
