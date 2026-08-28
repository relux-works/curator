# TASK-260811-3kbf3l implementation and verification

Implemented the executable rust-source-v1 build profile over admitted Cargo closures.

## Delivered
- Exact native target/toolchain binding for Cargo, rustc, sysroot, target stdlib, linker, and SDK, with duplicate, dangling, wrong-kind, cross-target, and time-of-use drift rejection.
- Fail-closed rejection of build scripts, build dependencies, proc macros, native links, ambient Cargo configuration/wrappers, unstable features, and undeclared inputs.
- Fresh-home frozen Cargo metadata/build execution, normalized metadata reconciliation, strict Cargo JSON event validation, and selection-neutral capture with exact target/feature/binding/plan identities.
- Protected executable publication, causal receipts, exact cache reuse, and fail-closed cache corruption handling.
- Registry-backed offline end-to-end, path-only, event, hook/config, graph-binding, toolchain-drift, undeclared-input, and protected-cache regressions.

## Security correction
Toolchain registration starts no child process before C0. Darwin SDK selection is filesystem-bound through /var/db/xcode_select_link, and displayed version evidence is derived from content identities rather than hard-coded version claims. A static zero-process regression covers this constraint.

## Findings
Predecessor metadata retained replay-absolute paths, remote vendor manifests were not recognized by graph containment, path-only vendoring could not issue nonempty evidence, and protected blobs discarded executable mode. Each issue was corrected with regression coverage.

## Verification
- make lint: exit 0, golangci-lint v2.12.2, 0 issues.
- go test -count=1 ./internal/rustsource ./internal/closureexec: exit 0 after the lint-only corrections.
- go vet ./...: exit 0.
- go build ./...: exit 0 after the lint-only corrections.
- go test -count=1 ./...: exit 0; exact-current behavioral implementation passed repository-wide.
- task-board validate: exit 0 before handoff reconciliation.
