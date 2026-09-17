# TASK-260916-vygif0: rose-air-lane-rust-toolchain

## Description
After the pnpm pin landed (7c3ce2f) the rose-air lane on main (run 35121791685) executes the three real-pnpm cases green but still fails: internal/rustsource TestProductionManagerCapturesRegistryFromRawPaths and TestProductionManagerCapturesGitWithoutCallerProjection call rustc -vV (authority_external_test.go:229) and t.Fatal when rustc is absent — the self-hosted macbook-iv runner has no Rust toolchain on PATH while hosted macOS images ship rustup. Same shape as the pnpm problem: the lane must provide the toolchain deterministically. Options for the operator: (a) pin a Rust toolchain in the CI lane (a rust-toolchain action or rustup-init with a pinned version, identical in every lane that runs internal/rustsource), or (b) install rustc on macbook-iv by hand. Evidence: artifact test-evidence-rose-air of run 35121791685.

## Scope
(define task scope)

## Acceptance Criteria
rose-air lane on a main push is green; internal/rustsource production cases run (not skipped) on rose-air; hosted lanes unchanged or pinned identically.
