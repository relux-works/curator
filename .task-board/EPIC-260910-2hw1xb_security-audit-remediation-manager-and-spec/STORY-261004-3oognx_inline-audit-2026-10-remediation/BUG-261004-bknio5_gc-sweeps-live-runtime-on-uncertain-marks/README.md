# BUG-261004-bknio5: gc-sweeps-live-runtime-on-uncertain-marks

## Description
Report: docs/security-audit-2026-10-inline.md, finding N1 (priority: first). internal/scopes/gc.go Collect calls sweepRuntime before checking marked.uncertain, so a truncated consumers.json or an invalid/unreadable install marker makes `curator gc` (exit 0, warning only) delete runtime still used by installed commands; the shim then fails with 'no such file or directory'. The uncertainty guard protects only the build cache.

## Scope
internal/scopes/gc.go, cmd/curator gc path, regressions in cmd/curator and internal/scopes

## Acceptance Criteria
1. When the live reference set cannot be proven complete (any marked.uncertain), no runtime object is removed; the existing conservative behaviour for the consumer registry and build cache is kept.
2. Red-first CLI regressions through run([gc]): truncated consumers.json; invalid install marker; unreadable install marker; a repeated GC after each. Each proves a working shim runs before and after GC.
3. Normal GC with a complete reference set still removes genuinely unreferenced runtime (control row).
4. Warnings state that the runtime sweep was skipped and why.
