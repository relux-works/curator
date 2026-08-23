# Mandatory Rust rework after RUN-260819-a16a5c

Preserve the accepted architecture and all currently green behavior:

- portable assurance is the functional default and must execute Cargo metadata/build through committed `closureexec` permits and executor-issued receipts;
- portable receipts must honestly retain `network=not-observed` and must not claim lossless process/read/write observation;
- verified assurance remains optional, provider-backed, and fail-closed before session or process creation when unavailable;
- Cargo identity remains bound to exact operator-approved executable bytes;
- vendored compiled binaries remain forbidden.

Implement every required change from reviewer verdict `RUN-260819-a16a5c` before handoff:

1. Materialize named conformance coverage for CGP05, R01-R09, RF09-RF12, RH01-RH10, and protected-cache regressions. Add explicit named or table-driven fixtures/assertions for transitive registry, unavailable-original Git, transitive contained paths, feature-selected identities, target pruning, exact multi-member selection, include confinement, second clean rebuild/nondeterminism measurement, custom targets, unstable `-Z`, artifact dependencies, include escape, planted outputs, and toolchain drift. Negative vectors must prove zero later process starts and zero publication.
2. Replace synthetic post-success audit mutations as the sole proof of forbidden behavior. Through an injected verified provider/enforcement seam, model or observe actual attempted network, child-process, undeclared-read, and undeclared-write operations; assert stable diagnostics and no receipt/publication.
3. Implement deterministic Rust-facing undeclared-input diagnostics for RH05/RH06 using `rust_undeclared_input`, or document and test an explicitly superseding accepted shared-code branch. The accepted Rust contract currently requires `rust_undeclared_input`.
4. Add true pre-C0 process-start instrumentation covering every production launch seam across `NewManager`, Cargo registration, Rust build-tool registration, discovery, assurance preflight, C0, and permit creation. Assert zero starts until execution crosses a committed permit. Static source scans may supplement but not replace instrumentation.

Validation before handoff:

- targeted named conformance tests and stable diagnostic assertions;
- package tests for `internal/rustsource`, `internal/closureexec`, and affected packages;
- race tests for Rust/closure execution;
- `make lint`, `go vet ./...`, `go build ./...`, fresh `go test -count=1 ./...`;
- `git diff --check` and `task-board validate`;
- outcome evidence maps each named family/vector to its test and result.

Do not commit, stage, reset, clean, or overwrite unrelated user changes. Handoff only when every reviewer requirement is evidenced.
