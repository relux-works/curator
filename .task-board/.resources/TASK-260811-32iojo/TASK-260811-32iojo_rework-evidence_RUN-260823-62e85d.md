# TASK-260811-32iojo cycle rework evidence

Run: `RUN-260823-62e85d`

Resolved reviewer verdict: `TASK-260811-32iojo_review-verdict_RUN-260823-9899c8.md`.

## Implementation

- Changed modern Yarn peer-virtualization traversal so an edge returning to the exact in-progress package instance closes an ordinary non-ordering runtime SCC instead of returning `closure_graph_incomplete`.
- Retained fail-closed behavior when a dependency path returns to the same source locator through a different derived peer context. This is reported as a non-well-founded context cycle and cannot enter canonical graph evidence.
- Preserved provider precedence, closed semver checks, optional and implicit `@types/*` peers, unique PnP alias bijection, strict lock/rc parsing, immutable artifact authority, lifecycle suppression, and compiled-payload denial.
- Added permanent positive vectors for the exact Yarn 4.9.2 workspace `a -> b -> a` shape, a remote `a@npm:1.0.0 <-> b@npm:1.0.0` cycle, and a workspace cycle whose `b` instance is virtualized by an exact `host` peer context.
- Each positive vector covers Parse, canonical selected edges, capture/admission, immutable private cache, materialization, regenerated PnP reconciliation, and Node invocation. The real-tool vector executes all three through the pinned Yarn 4.9.2 verified OS-denied provider.
- Added a negative vector where the same source is recursively reached through two different host provider contexts; it fails with `closure_graph_incomplete` before graph output.
- Updated the modern Yarn README boundary from blanket cyclic-context rejection to ordinary SCC preservation plus different-context recursion rejection.

## Test-first evidence

The initial workspace-cycle regression ran before the product change and exited 1:

`go test -count=1 ./internal/yarnmodernsource -run '^TestModernWorkspaceRuntimeCycleIsRetained$'`

Expected failure: `closure_graph_incomplete: peer virtualization contains a recursive context`. This was the exact reviewer defect, not a passing gate.

## Validation

Every gate below ran directly as a standalone process. No result was piped through `tee`.

| Command | Exit | Evidence |
| --- | ---: | --- |
| Focused workspace cycle regression after fix | 0 | Ordinary cycle retained as selected runtime edges. |
| Three synthetic full-pipeline cycle vectors | 0 | Workspace, remote, and peer-adjacent cycles captured, materialized, PnP-reconciled, and invoked. |
| `CURATOR_TEST_YARN_MODERN_JS=... go test -count=1 ./internal/yarnmodernsource -run '^TestModernRuntimeCyclesThroughRealPinnedYarn$' -v` | 0 | All three physical Yarn 4.9.2 verified-provider subtests passed in 5.294s. |
| Focused cycle positive and negative boundary tests | 0 | Exact-instance SCCs admit; different-context recursion rejects. |
| `CURATOR_TEST_YARN_MODERN_JS=... go test -count=1 -race ./internal/yarnmodernsource` | 0 | Package passed in 22.758s. |
| `CURATOR_TEST_YARN_MODERN_JS=... go test -count=1 -cover ./internal/yarnmodernsource` | 0 | 81.4% statement coverage. |
| `golangci-lint run ./internal/yarnmodernsource` | 0 | `0 issues.` |
| `go vet ./internal/yarnmodernsource` | 0 | No diagnostics. |
| `go build ./internal/yarnmodernsource` | 0 | Build succeeded. |
| `gofmt -l internal/yarnmodernsource` | 0 | Empty output. |
| `git diff --check -- README.md internal/yarnmodernsource` | 0 | No whitespace errors. |
| Binary-deny vector `TestS05N06CompiledPayloadDirectRenamedAndNestedFailsBeforeExecution` | 0 | Direct, renamed, and nested compiled payload denial remains green. |
| Active Go package Kotlin-exclusion check | 0 | `kotlin_exclusion=pass active_go_packages_only`. |
| `CURATOR_TEST_YARN_MODERN_JS=... go test -count=1 ./...` | 0 | Full uncached repository suite passed; `cmd/curator` 519.222s and `internal/yarnmodernsource` 54.990s. |

## Tool readiness

- `task-board` 0.24.3-17-g7ac2be8, Go 1.25.5 darwin/arm64, Git 2.50.1, ripgrep 15.2.0, and golangci-lint 2.12.2 resolved and executed successfully.
- Pinned Yarn entry point: `.temp/TASK-260811-32iojo/toolchain/node_modules/@yarnpkg/cli-dist/bin/yarn.js`; the concrete runner independently verifies package version 4.9.2 and fingerprints its bytes before use.

## Source identity

| File | SHA-256 |
| --- | --- |
| `internal/yarnmodernsource/lock.go` | `dd919507e3a65c0dd2e73ab95de1bdb67e251eaf943df1734256f2f845a6d6e4` |
| `internal/yarnmodernsource/conformance_test.go` | `67d5d84d62f66f0ad7102cd57cfb7e32198fbc621736d2ba28eebf4600f9db6c` |
| `README.md` | `7707a3d7b3485cfa6995a260fa287db12241cd2d7a6dd2ea90b5921b051c934f` |

The worktree was already broadly dirty. This rework touched only the modern Yarn package and its existing README section. Nothing was staged, committed, reset, cleaned, or deleted.
