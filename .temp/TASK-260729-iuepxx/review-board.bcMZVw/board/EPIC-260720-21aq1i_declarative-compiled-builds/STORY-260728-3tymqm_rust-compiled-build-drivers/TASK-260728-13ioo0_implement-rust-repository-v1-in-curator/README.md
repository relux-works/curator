# TASK-260728-13ioo0: implement-rust-repository-v1-in-curator

## Description
Implement the accepted rust-repository-v1 driver in Go Curator on top of external snapshot audit, private staging, protected cache and atomic publication.

## Scope
Schema and target models, trusted Rust toolchain resolution, fixed offline build graph, source and toolchain currentness, receipt/marker integration, cache hit/miss, dry-run, project/global/hybrid install, rollback, status, repair, GC and macOS/Windows tests.

## Acceptance Criteria
Curator accepts only the closed Rust contract; package-controlled build commands and forbidden Cargo behaviors fail before compiler or mutation; valid vendored projects build deterministic manager-named artifacts; cache, receipt, marker, rollback, dry-run and currentness semantics match the shared vectors; scoped lint, Go tests and native macOS/Windows gates pass.
