# TASK-260728-1koh5v: implement-local-kotlin-driver-in-curator

## Description
Implement the selected local Kotlin driver in Go Curator for context-excluded vendored source inside skills.

## Scope
Local build-root planning, compiler/runtime preflight, source and dependency validation, fixed offline worker, private staging, artifact or runtime-bundle cache, receipt/marker and launcher integration, atomic install, rollback/currentness and native tests.

## Acceptance Criteria
Valid local Kotlin fixtures produce deterministic manager-named launchable artifacts without source entering runtime/context; forbidden scripts/plugins/network fail before compiler or mutation; toolchain diagnostics, cache, rollback and supported platform gates pass.
