# TASK-260811-2h4m0s implementation blocker

Date: 2026-08-19

## Constraint

The accepted operator-owned Rust authority requires `NewManager` to select
Cargo 1.91.0 / Cargo 0.92.0 commit
`ea2d97820c16195b0ca3fadb4319fe512c199a43` through a closed, centrally owned
selector. No request may supply a root or executable, and selection may not
depend on ambient `PATH`, `HOME`, `CARGO_HOME`, or `RUSTUP_HOME`.

The repository currently has no such Cargo selector or trusted Cargo
installation record. The only closed selector in `internal/artifactpolicy` is
`ToolchainSelectorRuntimeGoV1`, whose root is intrinsically available from the
running Go program. Searches of all Go composition and configuration surfaces
found no Cargo selector, installation registry, or operator-owned Rust
toolchain preflight.

## Evidence and stopped attempt

The Rust manager prototype now has the raw-data-only public API, private
manager/capture ownership, portable/verified assurance ordering, and an issued
`closureexec.Executor` receipt for the hidden Git projection worker. Focused
tests pass. However, its provisional Cargo selection resolves `cargo` from
ambient `PATH` and derives rustup homes from ambient `HOME`. Its provisional
vendor bridge also materializes the independently expected output directly
instead of starting the selected Cargo process. Both contradict the accepted
decision and the run directive; they cannot be handed to review as security
implementation.

The shared derivation executor can represent the vendor output without a model
change: the independently derived transform supplies the exact sorted leaf set,
so every vendor file can be an `ExpectedEvidence` output. The unresolved
authority is how the composition root obtains the immutable Cargo installation
without adding a caller-controlled seam.

## Viable options

1. **Add a trusted operator-owned Cargo installation registry (recommended).**
   The application startup/install layer records and verifies the exact Cargo
   toolchain root, then exposes only an opaque sealed selector to
   `artifactpolicy`/`rustsource`. `ManagerConfig` remains data/policy-only. This
   is reusable and matches the existing Go selector pattern, but changes a
   shared ownership boundary and needs an approved source/lifecycle for the
   installed Cargo root.
2. **Bundle the pinned Cargo/Rust toolchain with Curator.** The root becomes an
   intrinsic release asset and can use a closed selector. This is simpler at
   runtime but materially changes release size, licensing/update operations,
   and the toolchain distribution model.
3. **Narrow Rust v1 to registry/path capture without Cargo execution.** This
   avoids the missing selector but violates the accepted Rust delivery scope
   and the task acceptance criteria, so it is not recommended.

An environment variable, public constructor field, option, callback, global
test hook, `exec.LookPath`, rustup proxy lookup, or ambient home discovery is
not a viable option because each makes the authority substitutable.

## Decision needed

Choose option 1 or 2. For option 1, specify which trusted Curator
startup/install component owns the Cargo root and its persisted immutable
identity. After that decision, implementation can replace the provisional
selector, execute the exact Cargo vendor permit through the manager-owned
executor, and finish the retained Cargo conformance gates.

## Validation evidence

- `gofmt -w internal/rustsource/*.go`: exit 0.
- `go test -count=1 ./internal/rustsource`: exit 0.
- `go test -count=1 ./internal/closureexec`: exit 0.
- `go test -count=1 ./internal/artifactpolicy`: the command runner returned no
  exit status, so it is not reported as passing.

No acceptance checklist item was checked from these focused prototype tests.
