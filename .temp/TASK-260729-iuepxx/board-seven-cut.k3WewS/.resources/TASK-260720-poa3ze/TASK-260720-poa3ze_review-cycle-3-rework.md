# Review-cycle 3 rework evidence — compile-only build drivers

**Task:** `TASK-260720-poa3ze`  
**Date:** 2026-07-20  
**Status:** Ready for review  
**Primary outcome:** `.research/260720_compile-only-build-drivers.md`

## Reviewer finding closure map

| Returned finding | Resolution in the revised research outcome | Reviewer evidence |
|---|---|---|
| Build inputs could remain visible to agent context, especially on cache hits and dry-runs that skip `go list`. | Schema 6 adds explicit, link-free, pairwise-disjoint `build_roots`. V1 rejects root modules, requires the selected module and every non-standard package/source/embed input to remain below one build root, excludes every build-root subtree before locale/context copying, and never runtime-copies it. The exclusion is static, so real builds, cache hits, and dry-runs produce the same context boundary without compiler discovery. | Sections 3.1, 3.4 rules 3–7/15, 3.5 positive and negative JSON vectors, 5.5 currentness, 6.2 dry-run, and 8.2 conformance inventory. |
| Whole-install rollback lacked isolation from another project's successful shared-state mutation. | Real operations take sorted project locks, release any optional one-key build lock, and then hold one exclusive manager-home mutation lock across recovery, cache publication, all project/global/hybrid/adapter/consumer swaps, reverse rollback, and GC. Shared state is revalidated after lock acquisition; consumer state commits last; rollback never releases the lock. Compilation remains private and concurrent outside the home lock. | Sections 6.1–6.3, including the two-project success and success-versus-rollback JSON vectors; section 9 decisions 11/15. |
| `CGO_ENABLED=0` and empty native-file arrays do not block ordinary `//go:cgo_import_dynamic`. | Before `go build`, the driver reads every active non-standard `GoFiles` file and conservatively rejects the exact ASCII token. The cache input/receipt carries policy `reject-nonstandard-cgo-import-dynamic-v1`; cache hits require an exact input/policy match, and older receipts are misses. The negative vector has empty cgo/native arrays and asserts no build/publication. | Sections 3.4 rules 16/18, 4.4 rejection vector, 4.6, 5.2/5.4 cache semantics, 8.2 negative inventory, and section 11's compiler/linker source citations. |
| Meson, Node/TypeScript bundlers, and `cli/curator.md` disposition were absent. | Meson is classified with Make/CMake as a package-selected recipe surface, with `run_command()` and `custom_target(command:)` evidence. Webpack configuration/loaders/plugins and esbuild plugins are separately classified as executable package-controlled callbacks. The affected-artifact inventory now specifies install/upgrade, dry-run, status, and no-generic-hook changes for `cli/curator.md`. | Sections 7, 8.1, 9 decision 12, and the primary-source rows in section 11. |

## Verification evidence

- Immutable inspected refs still equal their local `origin/main` and the detached worktrees remained clean after testing:
  - curator-spec `57c1f56846d221ecc55786bd3c2467ec32f11730`
  - curator `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
  - cocoaskills `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`
- Revised report: 1,064 lines; SHA-256 `6618fae3312e0b212e2e1e7a6daa3c2c8aa1d09abdb8e165564129373b86cb61`.
- All 21 fenced JSON examples parse as strict JSON.
- Independent CCJ-1 recomputation matches the report fixtures:
  - cache key `sha256:3fcd714a40e8918eb67dbd35d435875dcce6c9047da811a1fa26626e5e57be48`
  - receipt hash `sha256:750f5f7557ffe1dac1ec06b7d47cb5976e390a7dc7d57ea616c9edda91013c11`
- All 15 newly added primary-source URLs returned HTTP 200 on 2026-07-20. The relevant source bodies were checked directly; no secondary article is used as authority.
- Exact-ref regression gates passed:
  - curator-spec: 30 schemas and 93 vectors validated; 8 Python unit tests passed; `go test ./tools/...` passed.
  - curator: `go test ./...` passed.
  - cocoaskills: 488 passed, 18 platform skips.
- The initial spec/cocoaskills commands used the system Python and stopped before collection because `jsonschema`/`pytest` were absent. Both were rerun through the existing cocoaskills project virtual environment without dependency installation or repository changes.
- Go 1.25.5 Darwin/arm64 directive smoke: the non-standard root package listed `GoFiles:["main.go"]` and null/empty `CgoFiles`, `CFiles`, `CXXFiles`, `SFiles`, and `SysoFiles`. The fixed internal-link build produced artifact SHA-256 `593b311eab95d760ad10855dfdf394220f773f29bcfb5e6972f1958ba0c1b64d`; `otool -L` showed package-selected `libReviewerProbe.dylib`. The artifact was never executed. This directly supports the mandatory pre-build token rejection.
- Only `.research/`, `.temp/`, and task-board resources were changed; no product or specification source file was modified.

## Handoff recommendation

Review the revised outcome as the implementable v1 contract: schema 6 with context-excluded `build_roots`, only closed `go-v1`, pre-cache raw-snapshot identity, fixed offline/native/internal-link semantics, directive-aware receipts, and manager-home-isolated commit/rollback/GC. Generic build hooks and every other surveyed toolchain remain deferred or prohibited pending a separately closed driver.

This rework is handed off to review.
