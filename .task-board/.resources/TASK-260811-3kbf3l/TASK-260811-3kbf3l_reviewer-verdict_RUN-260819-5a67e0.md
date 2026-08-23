# Reviewer verdict for TASK-260811-3kbf3l

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260819-5a67e0`
Goal check immediately before verdict: run is not goal-bound.
Directives: none.
Reviewed producer outcome: `TASK-260811-3kbf3l_implementation-and-verification.md`.

## Required changes

1. **Cargo build execution bypasses the protected execution/receipt boundary.** `internal/rustsource/build.go:73-90` implements `osCargoBuildRunner` with direct `exec.CommandContext`. `Build` preflights assurance and derives a cache input at lines 133-141, but metadata and build are launched directly through that runner at lines 203 and 224; no build-scoped `DerivationPermit` is committed and no `Executor.Execute`/enforcement provider receipt covers either process. The code then constructs an `ExecutionReceipt` with `Network: "none"` at line 265 without observed process/read/write/network evidence. This violates the task scope, C6 protected-boundary contract, R08/RH undeclared-input guarantees, and the explicit rule that Cargo `--offline`/`--frozen` alone is insufficient. Route both Cargo invocations through the manager-owned committed permit and enforced/observed boundary, reconcile the complete audit, and publish only from issued receipt evidence. Add regressions proving attempted network, child process, undeclared read/write, and input mutation reject before publication.

2. **The mandatory zero-process pre-C0 regression does not observe process starts.** `internal/rustsource/build_test.go:15-24` only scans `build_toolchain.go` for three strings. The actual Cargo registration implementation is in `toolchain_registry.go`, so the test neither covers every equivalent discovery path nor observes `NewManager`/registration process starts as required. Replace or supplement this with an instrumentation boundary that counts every process-start attempt while manager construction and all build-tool registration run; assert zero through assurance preflight, committed C0, and the explicit permit boundary.

3. **Cargo version/implementation identity is asserted without evidence proving the claims.** `internal/rustsource/toolchain_registry.go:39-50` selects a user-writable directory by the name `1.91.0-<target>`; lines 111-114 then assign version `1.91.0` and commit `ea2d97820c16195b0ca3fadb4319fe512c199a43` to arbitrary fingerprinted bytes. A directory name plus a newly computed content digest proves drift stability, not that those bytes implement the claimed Cargo release/commit. Bind the accepted descriptor to exact approved byte identities or admitted filesystem metadata that proves those facts, and reject all other content before use.

## Positive evidence

- The `/usr/bin/xcrun` registration subprocess is absent and current registration source performs filesystem-only discovery.
- The focused registration/unit/binding/path-metadata reviewer test command exited 0.
- `TestRustSourceV1BuildsExecutesAndReusesProtectedOfflineOutput` passed.
- `go vet ./internal/rustsource` exited 0.
- The passing tests do not exercise an enforced/observed C6 boundary and therefore cannot satisfy the rejected requirements above.

No product code was modified. As a reviewer-archetype run, this verdict supplies no `commit_ack`.