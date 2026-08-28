# TASK-260825-1yzubs revalidation evidence

Run: `RUN-260828-5374d6`

No repository content was changed during this run.

## Candidate identity and scope

- Expected candidate tree: `867b50ae1a7cccc14cdd7cc2a070b11b2e3d4656`
- Independently reconstructed candidate tree before validation: `867b50ae1a7cccc14cdd7cc2a070b11b2e3d4656`
- Independently reconstructed candidate tree after validation: `867b50ae1a7cccc14cdd7cc2a070b11b2e3d4656`
- Diff against `main` (`de31754e854e385fca04de9cafeae06667a96123`) contains exactly the 11 paths recorded by CR revision 5:
  - `LOGBOOK.md`
  - `cmd/curator/builds.go`
  - `cmd/curator/builds_test.go`
  - `cmd/curator/gc_test.go`
  - `cmd/curator/global_status_test.go`
  - `cmd/curator/lifecycle_conformance_test.go`
  - `cmd/curator/main.go`
  - `cmd/curator/main_test.go`
  - `cmd/curator/status_test.go`
  - `cmd/curator/toolchain_remedy_test.go`
  - `internal/testtoolchain/lock.go`
- Path comparison against CR revision 5: no differences, exit code 0.

## Configured validation suite

Each command ran directly as a standalone foreground process, without `tee` or a pipe chain.

| Command | Exit code | Observed wall-clock | Log |
| --- | ---: | ---: | --- |
| `git submodule update --init --recursive` | 0 | 0.007s | `submodule-update-01.log` |
| `go build ./...` | 0 | 0.914s | `go-build-01.log` |
| `go vet ./...` | 0 | 0.629s | `go-vet-01.log` |
| `go test -count=1 -timeout 30m ./...` | 0 | 264.62s | `go-test-all-01.log` |

Post-validation checks:

- `git diff --check`: exit code 0.
- `task-board validate`: exit code 0. Its output still reports 598 pre-existing board diagnostics, chiefly missing historical resource payloads; the raw output is attached and this was not part of the configured repository validation suite.

## Result

The earlier `go vet ./...` environmental failure does not reproduce after the managed worktree's dependency checkout is present. The exact accepted candidate tree passes the complete configured suite and is suitable for republishing with a passing validation attestation.
