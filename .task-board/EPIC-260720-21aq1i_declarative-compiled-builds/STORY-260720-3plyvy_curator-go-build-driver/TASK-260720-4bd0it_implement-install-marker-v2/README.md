# Implement install marker v2 and build currentness

## Description
Extend the marker package to read historical marker v1, strictly read and write marker v2, and make compiled-command currentness depend on static context exclusion, raw build-source identity, logical cache receipts, and artifact bytes.

## Scope
Own internal/marker models, validation, writing, Current APIs, and tests. Preserve marker v1 reading and currentness for schema 1 through 5 installs; every new installation mutation writes marker v2. Marker v2 carries sorted build_roots, conditional build_source, and a sorted build map with driver, key, receipt hash, artifact hash, and relative artifact path. Currentness accepts callbacks for raw snapshot and protected cache inspection so marker does not import installer orchestration. Keep installed-tree ContentSHA256 marker-excluding behavior unchanged and do not implement CLI rendering or GC.

## Acceptance Criteria
Valid marker v1 remains readable and current for eligible legacy installs but is rewritten as v2 on mutation; marker v2 enforces skill schema through 6, non-nil sorted sets, build_source exactly when builds are non-empty, and exact per-command build fields with no unknown values; currentness is false or unknown as prescribed for missing raw snapshots, changed build roots, build-source mismatch, context-visible build files, untrusted cache boundary, missing or corrupt receipt, wrong input, key, target or toolchain, receipt hash drift, artifact drift, or path mismatch; package-provided root marker bytes are represented by build_source while installed content hashing still excludes the manager marker; authoritative marker schema cases and legacy marker tests pass.
