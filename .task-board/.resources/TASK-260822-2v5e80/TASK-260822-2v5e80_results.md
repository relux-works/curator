# TASK-260822-2v5e80 — toolchain-shim-remedy: implementation + verification

**Run:** RUN-260822-f42998 (fourth run; RUN-260822-88df86 and RUN-260822-3b7262 died at exit 124, RUN-260822-ded722 at exit 1)
**Branch:** `task/TASK-260822-2v5e80-toolchain-shim-remedy`, worktree `.temp/TASK-260822-2v5e80/worktree`, base `origin/main` = `6a9b201`
**Full change:** attached as `TASK-260822-2v5e80_remedy.patch` (verified byte-equal to the worktree diff; `git apply --check` exit 0 on pristine `6a9b201`)

## What changed

| File | Change |
| --- | --- |
| `internal/godriver/errors.go` | `Diagnostic` gains a `Remedy` field; `Error()` renders `go-v1 <code>: <detail>; <remedy>`; new `diagnosticRemedy` / `diagnosticErrRemedy` constructors |
| `internal/godriver/session.go` | `toolchainSelectionRemedy` const; both `toolchain_executable_mismatch` sites (`selectToolchain`, `validateProbe`) carry it |
| `internal/godriver/toolchain_remedy_test.go` (new) | Both mismatch sites asserted: code, protocol detail byte for byte, remedy, rendered line; plus no-remedy rendering unchanged |
| `cmd/curator/toolchain_remedy_test.go` (new) | End-to-end: `install` with a version-manager selection prints detail + remedy on one line and still prints the closed selection rule |
| `internal/install/diagnostics_test.go` | Remedy survives `RedactDiagnostic` (bounded, path-redacting) on the operator failure surface |

Remedy text, exactly as the task specifies:

```
put the real GOROOT/bin first on PATH, e.g. PATH="$(go env GOROOT)/bin:$PATH"
```

## Protocol strings — unchanged

Both details are byte for byte what they were before, and the remedy is a separate field, never folded into `Detail`:

- `selected Go executable is not the regular executable under the derived GOROOT`
- `go env GOROOT does not match the selected toolchain`

`git diff 6a9b201 -- internal/godriver/session.go` shows no edit inside either string literal. No `*.md` or `*.json` in the repo references `toolchain_executable_mismatch`, so no doc or vector text moves with this.

## Gates — each a standalone process, real exit codes

Toolchain: `GOROOT=/Users/iv/.goenv/versions/1.25.5`, `GOTOOLCHAIN=local`, `GOENV=off`, `PATH` fronted by that GOROOT/bin — the bare host `go` is a goenv shim, i.e. the exact selection this diagnostic now remedies.

| Gate | Exit | Notes |
| --- | ---: | --- |
| `gofmt -l cmd internal` | 0 | no output |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `golangci-lint run` | 0 | 0 issues |
| `go test -count=1 -run 'TestToolchainExecutableMismatchCarriesTheOperatorRemedy\|TestDiagnosticRenderingIsUnchangedWithoutARemedy' ./internal/godriver/` | 0 | 0.515s |
| `go test -count=1 -run TestGoToolchainRemedyReachesTheOperatorIntact ./internal/install/` | 0 | 0.410s |
| `go test -count=1 -run TestInstallPrintsTheRemedyAVersionManagerSelectionEarns ./cmd/curator/` | 0 | 0.733s |
| `go test -count=1 ./...` | 0 | 41 packages ok, 0 FAIL; `cmd/curator` 527.6s, `internal/install` 118.8s, `internal/godriver` 82.1s. Log: `TASK-260822-2v5e80_go-test-all.log` |

Not run: `make ci-test` / conformance vectors — the go-v1 vector suite skips at SPEC_PIN `00b1688a` (it publishes no `vectors/build-drivers.json`), and the harness compares diagnostic codes only, never detail text (`internal/godriver/builddriver_rejection_conformance_test.go:513`). Nothing in this change touches a code.

## For the reviewer

- **Two branches exist for this task.** `task/TASK-260822-2v5e80-toolchain-shim-remedy-r88df86` (worktree `.temp/TASK-260822-2v5e80/worktree-r88df86`) is a duplicate produced by the concurrent second spawn: same design, same operator text, fewer tests. Discard it; the canonical branch is `task/TASK-260822-2v5e80-toolchain-shim-remedy`. Left in place rather than deleted, because picking is the reviewer's call.
- Nothing is committed and nothing is pushed, per repo policy.
- Known display limit, already logged: the remedy survives the operator failure line (`Result.Errors`, 193 runes, under the 240-rune `RedactDiagnostic` bound) but is what the `...` eats in a *build report row*, which prefixes 101 more runes and truncates at 294 (`cmd/curator/builds.go:424`). Making it fully visible there needs the bound or the prefix changed — out of this task's scope.
