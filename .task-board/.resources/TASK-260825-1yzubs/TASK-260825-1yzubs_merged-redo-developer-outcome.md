# TASK-260825-1yzubs merged-base developer outcome

## Result

The redo is implemented against the adapter-landed Story workspace. `run` now
accepts an explicit configuration source, stdout, and stderr. The production
boundary resolves `config.UserPath()` once and passes the process streams; the
command core and command helpers use an invocation-scoped `cli` value. Tests
therefore no longer need `CURATOR_CONFIG` or process-global stdout/stderr
capture.

The injectable seam is documented at the production call site and is driven by
`TestRunUsesInjectedConfigSourceAndWriters`, which runs two concurrent CLI
invocations with distinct config/project/writer sets and rejects cross-output.
The merged HTTPS config and verified-provider paths were re-verified and adapted
rather than taking the blueprint hunks verbatim.

The former five-subtest compiled repair/rollback/recovery case is split into
three independent parallel top-level fixtures. The Darwin host-GOROOT test lock
is held once by the `cmd/curator` test process, not once per parallel fixture.
This retains cross-package/process isolation while avoiding internal lock
serialization; fixed worker and CLI-helper subprocesses bypass the parent lock
to avoid self-deadlock.

All review-named main-only cases were inventoried. Five top-level tests remain
serial for unrelated, honest process-global reasons: two mutate the verified
provider resolver, one swaps stdin, and two select a private Git credential
HOME. No config/stdout/stderr-bound serial case remains.

## Timing evidence

All commands were direct standalone processes with real exit codes; no gate was
piped through `tee`.

| Run | State | Exit | Wall clock | Verdict |
| --- | --- | ---: | ---: | --- |
| Prior producer baseline | pre-refactor, pre-merge evidence | 0 | 408.01s | baseline |
| `full-package-01` | merged refactor with per-test GOROOT locks | 0 | 323s | green suite, performance AC red |
| Final uncached 1 | final tree | 0 | 209s | under 4m |
| Final uncached 2 | final tree | 0 | 212s | under 4m |
| Final uncached 3 | final tree | 0 | 222s | under 4m |

Final maximum is 222s (3m42s), 18s below the four-minute bound. The three final
runs were consecutive and used `go test ./cmd/curator -count=1 -timeout=9m`.

## Coverage

The baseline was produced from the saved merged-workspace pre-change source
snapshot in `.temp/`, not reused from the earlier pre-merge run.

| Metric | Pre-change | Final | Delta |
| --- | ---: | ---: | ---: |
| `go tool cover` statements | 62.2% | 63.3% | +1.1pp |
| Covered statements | 929 | 951 | +22 |
| Total statements | 1492 | 1502 | +10 |

Both coverage commands exited 0. The additional production statements are the
injection plumbing; coverage increased rather than regressed.

## Validation

| Command | Exit | Evidence |
| --- | ---: | --- |
| Focused concurrent seam test | 0 | `focused-seam-01.log`, final full runs |
| Merged HTTPS CLI focused tests | 0 | `focused-https-01.log` |
| Representative concurrent `-race` run | 0 | `focused-race-01.log` |
| `go test ./cmd/curator -count=1 -timeout=9m` ×3 | 0 / 0 / 0 | `final-full-01..03.log` and `.time` |
| Baseline and final coverage | 0 / 0 | `coverage-baseline.*`, `coverage-final.*` |
| `go vet ./cmd/curator ./internal/testtoolchain` | 0 | `go-vet-01.log` |
| `go build ./cmd/curator` | 0 | `go-build-01.log` |
| `golangci-lint run` (v2.12.2) | 0 | `golangci-lint-01.log`, `tool-readiness-02.log` |
| `gofmt -l cmd/curator internal/testtoolchain` cleanliness gate | 0 | `gofmt-check-01.log` |
| `git diff --check` | 0 | `git-diff-check-01.log` |
| `task-board validate` | 0 | `task-board-validate-01.log` |

No files were staged or committed. The generated root `curator` build binary
was removed after the successful build. Unrelated existing Story-worktree
changes were preserved.
