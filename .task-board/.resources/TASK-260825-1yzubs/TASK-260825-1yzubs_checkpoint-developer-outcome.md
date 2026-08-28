# TASK-260825-1yzubs checkpoint developer outcome

## Result

The task-owned patch was applied with `git apply --3way` onto checkpoint
`defbc368d8f973e121c3ae53a635aec7be97d3f6`. Conflicts were resolved so both
deliveries survive:

- `run` receives an explicit config source plus stdout/stderr writers, and the
  command implementation carries those dependencies through an invocation-local
  `cli` value;
- tests use private config/writer seams and parallelize the formerly global-state
  constrained cases;
- the heavy compiled status repair/rollback/recovery case remains split into
  independent parallel fixtures;
- the landed Windows production-binary `.exe` suffix and host-GOROOT
  cross-package isolation remain present.

The first focused compiled-fixture probe exposed a merge-only self-deadlock:
package `TestMain` held the Darwin host-GOROOT lock while three landed helpers
attempted to reacquire it. The probe exited 1 after its two-minute timeout. The
redundant in-package acquisitions were removed; package-level ownership remains,
and the exact probe rerun exited 0 in 1.259 seconds.

## Timing

All gates ran as direct standalone processes and their real exit codes were
recorded. The three final package runs were consecutive and used
`go test ./cmd/curator -count=1 -timeout=9m`.

| Run | Exit | Wall clock | Four-minute margin |
| --- | ---: | ---: | ---: |
| Pre-refactor baseline | 0 | 408.01s | -168.01s |
| Checkpoint run 1 | 0 | 220.48s | +19.52s |
| Checkpoint run 2 | 0 | 220.41s | +19.59s |
| Checkpoint run 3 | 0 | 214.14s | +25.86s |

The worst fresh run is 187.53s (46.0%) faster than the baseline and remains
under four minutes.

## Coverage

| Metric | Pre-change | Checkpoint | Delta |
| --- | ---: | ---: | ---: |
| Statement coverage | 62.2% | 63.3% | +1.1pp |
| Covered statements | 929 | 951 | +22 |
| Total statements | 1492 | 1502 | +10 |

The fresh coverage command exited 0. Coverage increased rather than regressed.

## Validation

| Command | Exit | Evidence |
| --- | ---: | --- |
| Injectable seam focused test | 0 | `focused-seam-01.log` |
| Initial compiled lock diagnostic | 1 | `focused-compiled-lock-01.log` (merge self-deadlock, two-minute timeout) |
| Exact compiled lock rerun | 0 | `focused-compiled-lock-02.log` |
| Focused injectable-seam `-race` | 0 | `fresh-focused-race-01.log`, `.time` |
| Full `cmd/curator` ×3 | 0 / 0 / 0 | `fresh-final-full-01..03.log`, `.time` |
| Full coverage | 0 | `fresh-coverage-final.log`, `.out`, `.time` |
| `go build ./cmd/curator` | 0 | `fresh-go-build-01.log` |
| `go vet ./cmd/curator ./internal/testtoolchain` | 0 | `fresh-go-vet-01.log` |
| golangci-lint 2.12.2 | 0 | `fresh-golangci-lint-01.log` (`0 issues`) |
| gofmt cleanliness gate | 0 | `fresh-gofmt-check-01.log` |
| `git diff --check` | 0 | `fresh-git-diff-check-01.log` |
| `task-board validate` | 0 | `fresh-task-board-validate-01.log` |

`task-board validate` exited 0 while its output reported 1741 issue(s) found;
the output is not semantically clean.

No files were staged or committed.
