# Implement pure compiled-build planning

## Description
Refactor project and global planning so all validation and trust gates precede toolchain and cache planning, and dry-run reports deterministic build outcomes without any persistent mutation or source-aware Go command.

## Scope
Own a new src/csk/builds/planner.py, the planning portions of src/csk/installer.py and src/csk/global_install.py, read-only paths in src/csk/audit_registry.py, and dry-run lock routing in src/csk/cli.py. Produce provider-first and command-lexical build plans from validated closures, source identities, trusted toolchains, and protected cache lookups. Defer compilation, publication, markers, shims, and target swaps to downstream tasks.

## Acceptance Criteria
Resolution and frozen snapshot validation, manifest and build-root checks, closure and collision planning, source audit, registry and attestation gates, moved-tag policy, and MCP and system requirements complete before any toolchain probe or cache lookup. The build plan records exact input, key, target, and cache-hit, would-preflight-and-build, would-rebuild-untrusted-cache, corrupt, or unsupported result in provider-first and lexical command order. Dry-run acquires no mutation or project lock, performs no recovery, runs only package-independent Go probes, never runs go list or go build, and creates no audit record, registry cache or state, fingerprint memo, Go cache, build staging, cache entry, journal, runtime, context, marker, shim, adapter, consumer, or GC mutation. Generation recheck retries or reports concurrent_state_change. Before-and-after filesystem and registry-state tests, focused pytest, and strict mypy pass.
