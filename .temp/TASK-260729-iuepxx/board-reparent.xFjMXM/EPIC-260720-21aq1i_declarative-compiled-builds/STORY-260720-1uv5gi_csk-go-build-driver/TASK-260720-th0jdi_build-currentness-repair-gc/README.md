# Implement build currentness, repair, and GC

## Description
Extend project and global status, install repair behavior, and locked garbage collection to understand marker v2, protected build entries, receipts, artifacts, raw source identity, current toolchain and target, and in-flight journals.

## Scope
Own src/csk/status.py, build-currentness helpers, src/csk/gc.py, global status integration, and focused CLI and maintenance tests. Keep marker v1 currentness for schema 1 through 5. For v2 builds, inspect the selected raw snapshot and static context boundary, recompute logical identities, verify protected-state provenance and receipt and artifact bytes, and confirm marker and shim references. A normal install is the csk repair operation; do not invent a separate package-controlled repair hook.

## Acceptance Criteria
Status reports up-to-date only when ref, installed content, build roots and context exclusion, build source, toolchain and target, cache key, protected boundary, canonical receipt, artifact path, hash and size, marker build fields, and managed shim agree. Missing, corrupt, wrong-target, wrong-toolchain, unsupported, context-leaking, or untrusted build state is non-current in text and JSON and makes --check fail. Reinstall rebuilds or repairs from a revalidated snapshot without adopting untrusted bytes. GC runs under the home lock, marks keys from valid marker v2 files and active journals across project, global, and hybrid stores, retains uncertain state conservatively, and sweeps only unreferenced protected entries after the grace rule while preserving runtime and snapshot GC compatibility. Focused status, global status, repair, GC, corruption, provenance, and strict-mypy gates pass.
