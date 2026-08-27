# Curator Go build driver

## Description
Implement the accepted protocol contract in the Go reference manager from current origin/main. Build active compiled artifacts only after validation and trust gates, using manager-generated Go commands and manager-local caches, then atomically install verified outputs and launch them through managed shims.

## Scope
Curator skillspec parsing and validation, skill check, closure activation, install lifecycle, build driver package, runtime store and receipts, marker/currentness behavior where required, CLI diagnostics, README/authoring docs, fixtures, unit/integration tests, and cross-platform CI behavior.

## Acceptance Criteria
Curator accepts valid schema v6 manifests and rejects unsafe declarations; missing/incompatible Go toolchains fail clearly; builds use fixed isolated environment and readonly modules; cache identity includes source/build/toolchain/target inputs; corrupt or stale artifacts rebuild; dry-run performs no build/cache mutation; failed builds preserve the previous install; built commands launch with forwarded args and exit status; existing tests plus new race/vet/lint-relevant tests pass.
