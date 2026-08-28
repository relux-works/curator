# TASK-260825-1yzubs republish evidence

Run: `RUN-260828-fa363e`  
Role: developer  
Date: 2026-08-28  
Repository content changes made by this run: none

## Candidate and base verification

- Story branch tip: `73f17ef093dee4621197241f7105e423d03916af`.
- Managed workspace base: `main`.
- Candidate reconstructed from the worktree through a temporary alternate Git index: `867b50ae1a7cccc14cdd7cc2a070b11b2e3d4656`.
- The candidate hash matches the accepted candidate named in the republish brief.
- `git diff --name-only main 867b50ae1a7cccc14cdd7cc2a070b11b2e3d4656` reports exactly 11 paths:
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
- Deleted paths: none.
- The final post-validation candidate reconstruction still produced the same tree hash and path set.

## Three consecutive uncached package runs

Exact command for every run: `go test -count=1 ./cmd/curator`, wrapped only by `/usr/bin/time -p` with stdout/stderr redirected to its task-scoped log. Each command ran as a standalone foreground process.

| Run | Exit | Package time | Wall time | AC margin (<240s) |
| --- | ---: | ---: | ---: | ---: |
| 1 | 0 | 222.967s | 223.62s | 16.38s |
| 2 | 0 | 216.418s | 216.97s | 23.03s |
| 3 | 0 | 215.798s | 216.35s | 23.65s |
| Mean | — | 218.394s | 218.98s | 21.02s |

The prior accepted before measurement recorded on the task was 408.01s. Against the refreshed three-run mean, wall time decreased by 189.03s (46.3%). All three refreshed runs satisfy the under-four-minute target.

## Coverage comparison

Coverage commands used `go test -count=1 -coverprofile=... ./cmd/curator` as standalone foreground processes. The `main` baseline ran from a `git archive main` snapshot under `.temp/`; it did not switch or modify the managed Story branch.

| Metric | Current main | Candidate | Delta |
| --- | ---: | ---: | ---: |
| `go test` statement coverage | 62.3% | 63.3% | +1.0pp |
| Instrumented wall time | 338.22s | 218.76s | -119.46s (-35.3%) |
| Exit code | 0 | 0 | — |

Coverage did not regress.

## Other validation gates

| Gate | Command | Exit | Result |
| --- | --- | ---: | --- |
| Focused race | `go test -race -count=1 -run '^(TestRunUsesInjectedConfigSourceAndWriters|TestGCRunsSerializedAcrossConcurrentInvocations|TestGlobalStatusWithoutASkillfileStaysSilentAndCurrent|TestGlobalStatusRejectsPositionalArguments)$' ./cmd/curator` | 0 | Green |
| Build | `go build -o .temp/TASK-260825-1yzubs/curator-build ./cmd/curator` | 0 | Green |
| Formatting | `gofmt -l` over all 10 changed Go files, with an empty-output assertion | 0 | Clean |
| Vet | `go vet ./cmd/curator` | 0 | Clean |
| Lint | `golangci-lint run ./cmd/curator/...` | 0 | Clean; installed version is exactly 2.12.2 |
| Diff whitespace | `git diff --check` | 0 | Clean |
| Board validation | `task-board validate` from the control root against the authoritative board | 0 | Reports `598 issue(s) found.` |

The republish brief predicted that `task-board validate` would report `1741 issue(s) found.` The authoritative board's refreshed result is instead exit 0 with `598 issue(s) found.` This evidence records the observed result rather than substituting the stale predicted count.

## Evidence-honesty audit notes

- An initial gofmt wrapper was discarded because zsh passed a newline-delimited path list as one argument; the wrapper's final shell status was 0 even though the inner gofmt invocation emitted a path error. It is not counted above. The explicit-path rerun exited 0 and produced an empty output file.
- An initial board-validation wrapper exited 1 before launching `task-board validate` because its relative redirection directory did not exist in the control root. It is not a validator result. The standalone rerun with an absolute task-scoped log path exited 0 and produced the issue count above.
- A baseline scratch-preparation command containing a removal operation was rejected by policy before execution. A new `mktemp -d` snapshot was used instead.

## Outcome

The accepted candidate remains byte-identical, its diff against `main` is restricted to the expected 11 task-owned paths, the three refreshed uncached package runs are below four minutes, focused race/build/format/vet/lint/diff gates are green, and coverage improves by 1.0 percentage point. No repository content was changed during this republish run.
