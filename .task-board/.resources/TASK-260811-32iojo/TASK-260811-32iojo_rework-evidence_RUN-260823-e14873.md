# TASK-260811-32iojo condition rework evidence

Run: `RUN-260823-e14873`

## Reviewer finding resolved

The modern Yarn profile no longer converts condition parser errors into the
admissible `condition_unsupported` prune reason.

- Every condition on every admitted lock entry is validated during lock parse,
  including optional, optional-peer, and unreachable entries.
- Malformed or unsupported conditions return
  `closure_lock_format_unsupported` with an empty graph, lock digest, raw-lock
  digest, and configuration digest, so no capture, manager, build, cache, or
  publication authority is emitted.
- Selection defensively propagates evaluator errors if an already-parsed graph
  is mutated before reuse.
- The evaluator implements the pinned Yarn 4.9.2 `tinylogic` grammar for exact
  `os`, `cpu`, and `libc` selectors, unary `!`, grouping, `&`, `|`, and `^`,
  with Yarn's left-to-right binary semantics. Repeated negation matches Yarn:
  `!!os=linux` is positive and `!!!os=darwin` is negative.
- `ConditionGrammarID` is bound into the canonical configuration identity and
  exposed in `Layout`; raw accepted condition expressions remain bound in the
  canonical lock records.

## Pinned source audit

The exact `@yarnpkg/cli/4.9.2` source tag resolves to commit
`a9edb7777f04ba16f51503ef6775325b353b67cc`.

- `packages/yarnpkg-core/sources/Manifest.ts` normalizes leading `!` counts:
  even counts emit a positive selector and odd counts emit one negation.
- `packages/yarnpkg-core/sources/structUtils.ts` uses `tinylogic` with selector
  regex `(os|cpu|libc)=([a-z0-9_-]+)`.
- The pinned `tinylogic` 2.0.0 grammar accepts unary `!`, parentheses, and
  left-to-right `|`, `&`, and `^`; it rejects `os=linux && cpu=x64`.
- A direct pinned-source probe exited 0 and confirmed rejection of `&&`, true
  evaluation for `!!os=linux` and `!!!os=darwin`, and the expected `|`, `^`,
  grouping, and trailing-whitespace behavior.

## Files changed

- `internal/yarnmodernsource/lock.go`
- `internal/yarnmodernsource/conformance_test.go`

## Validation

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go run ./.temp/TASK-260811-32iojo-review-5/condition_probe.go` before rework | 0 | Reproduced fail-open `&&`, `!!`, and `!!!` behavior with nonempty digests. |
| Same reviewer probe after rework | 0 | `&&` fails with `closure_lock_format_unsupported` and empty digest; `!!`/`!!!` select correctly. |
| Focused N10 regression selection | 0 | Condition grammar, malformed optional/optional-peer/unreachable, and repeated-negation tests pass. |
| `go test -count=1 ./internal/yarnmodernsource` | 0 | Focused package suite passes. |
| `go test -count=1 -race ./internal/yarnmodernsource` | 0 | Race gate passes. |
| `go test -count=1 -coverprofile=.temp/TASK-260811-32iojo/coverage-rework-conditions-final.out ./internal/yarnmodernsource` | 0 | 76.4% statement coverage. |
| First `golangci-lint run ./internal/yarnmodernsource` | 1 | Truthfully failed on two ST1005 capitalized error strings; both were corrected. |
| Repeated standalone `golangci-lint run ./internal/yarnmodernsource` | 0 | `0 issues.` |
| `go vet ./internal/yarnmodernsource` | 0 | Vet passes. |
| `go build ./internal/yarnmodernsource` | 0 | Focused package build passes. |
| `gofmt -l internal/yarnmodernsource` | 0 | No output. |
| `go build ./...` | 0 | Repository build passes. |
| `CURATOR_TEST_YARN_MODERN_JS=.../4.9.2/.../yarn.js go test -count=1 -run '^TestN01RealPinnedYarnPnPInvokeThroughVerifiedExecutor$' ./internal/yarnmodernsource` | 0 | Real Yarn PnP materialization and dependency invocation pass through `sandbox-exec` with OS-level network denial. |
| `go test -count=1 ./...` | 0 | Full uncached repository suite passes; `cmd/curator` 377.401s, `internal/yarnmodernsource` 7.397s. |
| `git diff --check` | 0 | No whitespace errors in tracked changes. |
| `task-board validate` | 0 | Board valid. |

No files were staged, committed, reset, cleaned, or destructively modified.
The pre-existing dirty worktree was preserved.
