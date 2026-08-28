# Reviewer verdict for TASK-260811-2h4m0s

Verdict: **changes requested -> to-dev**

Run: `RUN-260819-d5560f`

The implementation is not accepted. Focused tests pass, but the production
authority path required by the accepted design is absent and the Cargo selector
executes outside the permitted lifecycle.

## Findings

1. **The Git oracle is not production-reachable from the shipped Curator executable.**
   `internal/rustsource/oracle_worker.go` installs the hidden mode only through
   package initialization, but `cmd/curator` does not import
   `internal/rustsource`. `go list -deps ./cmd/curator` contains no rustsource
   package. A freshly built `cmd/curator` invoked as
   `__curator_rust_git_oracle_v1` exits 2 and prints
   `curator: unknown command "__curator_rust_git_oracle_v1"`. The external
   positive tests pass because the Go test binary itself imports rustsource;
   they do not exercise the shipped executable copied by `NewManager` in a
   production composition root. Consequently a real manager cannot obtain the
   required Git projection receipt and current-scope Git capture is not
   delivered.

2. **Cargo selection starts an undeclared process before C0 and also precedes
   the oracle dispatch.** `internal/rustsource/toolchain_registry.go:28`
   initializes `registeredCargo` at package load. Lines 31-37 resolve ambient
   `rustup` through `PATH` and execute `rustup which cargo --toolchain 1.91.0`.
   Go initializes package variables before running the oracle `init()` at
   `internal/rustsource/oracle_worker.go:31`. Thus every binary that does import
   rustsource can spawn rustup before `NewManager`, before its C0/tool admission,
   and before any derivation permit; an oracle child can also spawn this
   undeclared child process before entering its fixed worker mode. This violates
   the accepted no-process-before-C0 and single fixed oracle process boundary,
   while portable tests cannot observe the child-process escape.

## Required rework

- Integrate the sealed Rust manager and hidden oracle into the trusted Curator
  composition root so the actual shipped executable recognizes the exact
  internal mode. Add a production-binary integration test, rather than relying
  only on a package test binary.
- Remove process execution from package initialization. Perform closed,
  operator-owned Cargo selection/admission within the authorized manager/C0
  lifecycle, and ensure the oracle mode dispatch occurs without Cargo/rustup,
  ambient PATH, or any undeclared child process.
- Add regressions proving zero process starts before C0, no oracle child
  process, and successful Git projection through the built production binary.
  Retain the existing raw-API, transform, drift, and zero-spawn tests.

## Verification evidence

- `go test -count=1 ./internal/rustsource ./internal/closureexec`: pass
  (`rustsource` 17.355s, `closureexec` 2.355s).
- `go list -deps ./cmd/curator`: rustsource dependency absent.
- Fresh `go build -o <temp>/curator ./cmd/curator`, followed by an empty-env
  invocation of the internal oracle mode: exit 2, unknown command.

No product code was modified by this reviewer. This is recoverable code rework,
so `blocked` is not appropriate. No `commit_ack` is supplied.
