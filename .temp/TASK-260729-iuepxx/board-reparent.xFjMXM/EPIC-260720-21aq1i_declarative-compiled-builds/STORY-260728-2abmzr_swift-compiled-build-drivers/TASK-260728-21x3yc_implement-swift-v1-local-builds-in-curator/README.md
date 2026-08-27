# TASK-260728-21x3yc: implement-swift-v1-local-builds-in-curator

## Description
Implement local context-excluded swift-v1 builds in Go Curator using the accepted Swift policy and common toolchain preflight.

## Scope
Local build-root planning, source/audit identity, fixed offline worker, toolchain/SDK probes, private staging, protected cache, receipts/markers, atomic install, rollback, status/repair/GC and native tests.

## Acceptance Criteria
Vendored Swift source inside a skill builds a deterministic manager-named executable without entering runtime/context; forbidden SwiftPM features fail before compiler or mutation; cache, rollback, toolchain diagnostics and supported macOS/Windows gates pass.
