# Specify the normative go-v1 manager lifecycle

## Description
Extend the manager profile with the complete closed go-v1 execution and installation lifecycle. The manager must validate and audit immutable source before any compiler work, derive all process arguments and output paths, isolate toolchain state, keep private builds outside shared commit locks, and publish only verified immutable artifacts through a manager-home transaction.

## Scope
Work in curator-spec origin/main commit 57c1f56846d221ecc55786bd3c2467ec32f11730. Own profiles/manager.md only. Specify trusted Go 1.23 plus allowlisted families, curator-go-toolchain-v1, private telemetry-off initialization, the clean environment, native vendor-only target, fixed direct argv for telemetry, version, env, list, and build, dependency and directive rejection, cache trust and receipts, currentness, status, repair, rollback, recovery, lock ordering, consumer-last commit, and locked GC. CLI wording and schemas are separate tasks.

## Acceptance Criteria
The profile states source validation, context exclusion, build-source digest, closure and collisions, source audit and registry gates before cache lookup or compilation; all five Go argv forms and the fixed environment are normative and no shell or package-selected executable or argument can enter the process graph; go list checks every active dependency and embed input, rejects native files, syso, non-standard assembly, cgo, go:cgo_import_dynamic, path escapes, unknown toolchains, workspaces, downloads, PGO, and external-link fallback; dry-run never runs go list or go build and leaves all listed persistent paths unchanged; real installs build misses before mutation, then use deterministic project and manager-home lock ordering, protected cache publication, one journal, reverse rollback, consumer last, recovery, repair, and GC; build failure preserves prior installation byte-for-byte; python3 tools/validate.py passes.
