# TASK-260720-1nvomm implementation and rework results

## Inputs and scope

- Baseline: HEAD and origin/main at 57c1f56846d221ecc55786bd3c2467ec32f11730 on agent/protocol-v6-core.
- Accepted contract SHA-256 verified as 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681.
- Product changes remain limited to protocol/core.md, SECURITY.md, and decisions/0004-compile-only-build-drivers.md. Schemas, vectors, generator, manager profile, CLI guide, and release metadata were not edited.

## Accepted-decision mapping

1. Closed driver and new-revision rule: core 4.2 and 9; decision 0004 Decision and Future-driver rule.
2. Dedicated build roots and static exclusion on real, cache-hit, and dry-run paths: core 3.1 and 4.2; decision 0004 Decision.
3. Vendor-only, native, cgo-disabled v1 model: core 4.2; decision 0004 Decision.
4. Assembly, host-object, active SysoFiles, cgo-import-dynamic, external-link, and libgcc rejection: core 4.2; SECURITY Compile-only build boundary; decision 0004 Decision.
5. Separate immutable compiled-artifact cache and receipt rather than skill-plus-commit runtime identity: core 9; decision 0004 Context and Decision.
6. Marker v2 build roots, source, receipt, artifact, and marker-v1 compatibility: core 10; decision 0004 Decision and Compatibility impact.
7. Compiler-free dry-run and no cache or manager mutation: core 9; decision 0004 Decision.
8. Private telemetry-off bootstrap: core 4.2; decision 0004 fixed Go-only pipeline.
9. curator-go-toolchain-v1, curator-build-source-v1, CCJ-1 cache keys, and canonical receipt bytes: core 8.1, 8.2, and 9; decision 0004 Decision.
10. Marker-excluding installed content hash versus all-file raw build-source identity: core 8 and 8.1; decision 0004 Decision.
11. Concurrent private builds with serialized publication, commit, rollback, recovery, consumer, and GC state: decision 0004 Decision. Manager operational detail remains downstream-owned.
12. Meson and Node bundler recipe/plugin prohibition: SECURITY Compile-only build boundary; decision 0004 Rejected alternatives.
13. Protected-state TCB, forced miss/rebuild, and same-principal/admin boundary: core 9; SECURITY Protected-cache trust boundary; decision 0004 Security impact.
14. Portable logical key, receipt, artifact identity, and validation versus non-portable physical layout: core 9; decision 0004 Decision and rejected layout alternative.
15. Schema 6 rc.4 claim-v2 transition with frozen rc.3 claim v1: decision 0004 Compatibility impact. Release artifacts remain downstream-owned.

## Security-boundary mapping

- No hooks or package-controlled execution: core 3 and 4.2; SECURITY lines under Compile-only build boundary; decision 0004 Context, Decision, and Rejected alternatives.
- Compiler-input trust and resource controls: SECURITY Compiler-input trust boundary; core 4.2 read-only source, bounded output, fixed process graph, and never-execute rule.
- Protected-cache trust and non-provenance of hashes: core 9 and 10; SECURITY Protected-cache trust boundary; decision 0004 Decision and Security impact.

## Reviewer-requested rework

- Core 4.2 now requires all three bootstrap probes to use a manager-owned empty working directory and an environment with private roots, GOENV=off, GOTOOLCHAIN=local, normalized locale, and no inherited GOROOT or target.
- Core 4.2 now trusts a go-list result under GOROOT only when Standard == true and Goroot == true; every other result is confined to the build root.

## Exclusion proof

Core 3.1 requires build_roots exclusion before locale rendering, cache lookup, or compiler execution, explicitly making the selected agent context identical for real builds, exact cache hits, and dry-runs. It also forbids runtime-store copying and prompt-visible generated artifacts. Core 10 makes any prompt-visible or runtime-copied build-root file non-current. Behavioral vectors are downstream-owned and were intentionally not edited here.

## Verification

- python3 tools/validate.py via the pinned task-local requirements environment: passed, 30 schemas and 93 vector files.
- make validate via the same environment: passed, including 8 Python tests and go test ./tools/....
- git diff --check: passed.
