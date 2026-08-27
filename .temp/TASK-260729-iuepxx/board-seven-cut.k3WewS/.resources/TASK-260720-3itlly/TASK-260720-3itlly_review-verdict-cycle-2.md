# TASK-260720-3itlly reviewer verdict — changes requested

## Verdict

Route to `to-dev`. The cycle-2 pre-handoff toolchain verification and source-token rechecks close the two previously reported races, and all independent build/test/lint gates pass. One toolchain failure path still occurs after live mutation, so the pre-mutation guarantee is not yet satisfied.

## R1 — deferred release still reports toolchain drift after live mutation (high)

`BuildSession` now documents `Close` as a pure private-root release, but the real adapter does not implement that contract. `goSession.Close` delegates to `godriver.Session.Close`, and the driver close operation calls `VerifyToolchain` again before removing its operation root.

Both `Project` and `Global` defer `releasePlan` immediately after planning. The deferred close therefore runs only after `OnStaged`, consumer recording (project), context/runtime installation, cleanup, shims/global bins, env files, adapters, and runtime GC. If the pre-handoff verification succeeds but the toolchain changes before deferred close, the run returns `failed` only after live state has changed. This is exactly the late-failure boundary forbidden by the acceptance criteria.

Evidence:

- `internal/install/builddeps.go:21-28` says verification is separate so close is cleanup-only.
- `internal/install/builddeps.go:176-177` calls `session.Session.Close()`.
- `internal/godriver/session.go:301-320` shows that close re-runs `VerifyToolchain`.
- `internal/install/install.go:378,413-518` and `internal/install/global.go:172,204-261` show release is deferred across all live mutations.
- `internal/install/plan.go:226-234` intentionally converts the late close error into a failed result.
- `TestSessionReleaseFailureIsReported` asserts only the failed status/error and does not assert preservation; the project/global preservation tests inject failure into the new pre-handoff `VerifyToolchain`, not the real late close path.

Required rework:

1. Make the post-handoff release path genuinely cleanup-only, or otherwise ensure every failure that is reported as toolchain trust failure is resolved before `OnStaged` and before the first shared mutation.
2. Add project and global regression tests where pre-handoff verification succeeds but the release path returns a toolchain-drift error, asserting no consumer, install, runtime, shim/adapter, or live-cache state changed.
3. Keep the staged artifacts readable for the handoff while preserving the operation-private cleanup guarantee.

## Independent verification

- `git diff --check` — pass
- `gofmt -l internal/install` — pass, no output
- `go build ./...` — pass
- `go vet ./...` — pass
- `go test -count=1 ./internal/install/` — pass
- `go test -count=1 ./...` — pass, all 36 packages
- `golangci-lint v2.1.6 run ./internal/install/...` — pass, 0 issues
- focused cycle-2 trust/release tests — pass

No product code was modified by the reviewer.
