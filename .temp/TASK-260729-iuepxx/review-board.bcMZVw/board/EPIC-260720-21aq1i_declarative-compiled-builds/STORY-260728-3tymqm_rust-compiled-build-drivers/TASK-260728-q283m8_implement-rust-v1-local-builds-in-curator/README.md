# TASK-260728-q283m8: implement-rust-v1-local-builds-in-curator

## Description
Implement local context-excluded rust-v1 builds in Go Curator using the accepted Rust policy, common toolchain preflight, private staging, protected cache and atomic publication.

## Scope
Schema planning for local build_roots/source_dir, source fingerprint and audit ordering, fixed offline Cargo/rustc worker, cache/receipt/marker identity, dry-run, project/global/hybrid install, rollback, status/repair/GC and macOS/Windows tests.

## Acceptance Criteria
Vendored Rust source inside a skill builds a deterministic manager-named executable without entering runtime or agent context; forbidden Cargo behavior fails before compiler or mutation; cache, rollback, currentness, toolchain diagnostics and native gates pass without regressing Go/script commands.
