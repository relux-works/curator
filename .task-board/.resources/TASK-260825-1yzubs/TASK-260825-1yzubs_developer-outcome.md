# TASK-260825-1yzubs developer outcome

## Result

The `cmd/curator` command core is invocation-scoped and its full package now runs below the four-minute acceptance bound across three consecutive uncached runs.

## Injectable seams

- `main()` is the only command entry point that resolves `config.UserPath()` and wires `os.Stdout` / `os.Stderr`.
- The documented `run(args, source, stdout, stderr)` seam accepts an injectable `configSource` and invocation-local writers.
- The internal `cli` receiver carries config, stdout, stderr, and the user-home seam through the command call chain.
- Flag diagnostics, Bubble Tea output, build repair notices, and operator SSH resolution use the invocation-local writers.
- Tests use explicit config paths and private buffers. Environment-sensitive `CURATOR_GO` cases use a helper subprocess with an explicit `Cmd.Env`, so parallel tests do not mutate process-global environment.
- Search review found no remaining `t.Setenv`, `os.Stdout` assignment, or `os.Stderr` assignment in `cmd/curator` tests.

## Test restructuring

- Restored the predecessor's already-reviewed parallel cases and parallelized the remaining hermetic top-level and subtests.
- Added `TestRunUsesInjectedConfigSourceAndWriters`, including a negative assertion that ambient config is not consulted by the testable core and positive assertions for isolated stdout/stderr.
- Replaced global output capture with per-invocation buffers, including the prior real-stdout cases.
- Split `TestCompiledProjectStatusRepairRollbackRecovery` into independent status/untrusted, corrupt-state repair, and commit-rollback groups. Install and upgrade fixtures run in parallel, retaining all repair, refusal, and no-mutation assertions without one shared serial fixture dominating the package.

## Wall-clock evidence

Command for every after run: `go test -timeout 30m -count=1 ./cmd/curator`.

| Run | Package time | Wall clock | Exit | Change from baseline wall clock |
| --- | ---: | ---: | ---: | ---: |
| Baseline | `405.441s` | `408.01s` | 0 | — |
| After 1 | `230.372s` | `230.91s` | 0 | `-177.10s` (`-43.4%`) |
| After 2 | `206.581s` | `207.11s` | 0 | `-200.90s` (`-49.2%`) |
| After 3 | `231.517s` | `232.05s` | 0 | `-175.96s` (`-43.1%`) |
| After average | — | `223.36s` | 0 | `-184.65s` (`-45.3%`) |

All three acceptance runs are below 240 seconds and produced no flaky failures.

## Coverage

`go test -timeout 30m -count=1 -coverprofile=... ./cmd/curator` exited 0.

| Metric | Baseline | After | Delta |
| --- | ---: | ---: | ---: |
| Statement coverage | `60.1%` | `61.3%` | `+1.2pp` |
| Covered statements | `780 / 1297` | `800 / 1305` | `+20 covered` |

Coverage did not regress; the added injectable-seam behavior is covered.

## Validation

| Gate | Exit | Result |
| --- | ---: | --- |
| Three uncached full `cmd/curator` runs | 0 / 0 / 0 | Pass; all below four minutes |
| Focused `go test -race` | 0 | Pass |
| Coverage run | 0 | Pass; `61.3%` |
| `gofmt -l cmd/curator` | 0 | Pass; no output |
| `go vet ./cmd/curator` | 0 | Pass |
| `golangci-lint v2.12.2 run ./cmd/curator/...` | 0 | Pass |
| `git diff --check` | 0 | Pass |
| `task-board --no-update-check validate` | 0 | Pass |
| Cross-package test excluding `cmd/curator` | 1 | Expected repository setup failure described below |

The requested cross-package command was run directly and truthfully failed because `internal/ui/ui_test.go` imports a module replaced by the absent local directory `./agents/skills/skill-go-testing-tools/tuitestkit`. Every other package reached by the command passed, including `internal/config` and `internal/godriver`. The missing replacement predates this change, is outside the allowed `cmd/curator/**` file scope, and is not masked as a passing gate.

## Scope and anomalies

- Modified only `cmd/curator/**`; no documentation or board checkout files were edited.
- The story worktree did not contain the predecessor's parallel-test commit even though its review resource described it. The applicable predecessor test changes were restored in this worktree and then extended.
- The repository has a root `LOGBOOK.md`, but the task's explicit file scope excludes it and this board CLI exposes no logbook mutation. The two anomalies above are therefore persisted here and in task notes instead of modifying the concurrently owned document.
- No files were staged or committed.

