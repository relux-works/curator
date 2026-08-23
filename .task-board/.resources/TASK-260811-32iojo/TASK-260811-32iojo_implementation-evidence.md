# TASK-260811-32iojo implementation evidence

Status at handoff: ready for review.

## Delivered scope

- Added `internal/yarnmodernsource` with a closed modern Yarn 4.9.2 / lock v8 parser and `yarn-modern-source-v1` profile.
- Binds the exact Yarn release, runtime/tool identity, manager-owned built-in plugin set, `.yarnrc.yml`, cache key, compression, linker, supported architectures, package conditions, patch locators and bytes, and lock checksums.
- Accepts exact normalized Yarn ZIPs and deterministic raw-tgz normalization, recursively admits archive bytes, reconciles embedded package metadata, and verifies `cacheKey/SHA-512` against `yarn.lock`.
- Rejects local/downloaded plugins, unsupported resolver/fetcher protocols including Git, undeclared patches, mutable npm locators, lifecycle scripts, implicit `binding.gyp` native builds, compiled/opaque payloads through the shared artifact policy, and manager configuration outside the closed profile.
- Discards ambient `node_modules`, `.yarn/cache`, `.yarn/install-state.gz`, `.pnp.cjs`, `.pnp.loader.mjs`, unplugged state, and SDK state before admission; preserves and independently verifies declared `.yarn/patches`.
- Derives one immutable private replay tree from the admitted project plus exact `.yarn/cache`, eliminating overlapping work-copy ordering. C6 runs only `install --immutable --immutable-cache --mode=skip-build` with `YARN_ENABLE_NETWORK=0`, scripts disabled, empty private HOME, and the shared protected executor. PnP/install state is regenerated as output; PnP and node-modules linker results are reconciled.
- Added a Modern Yarn 4.9.2 harness row to the existing README tooling table.

## Tests and vectors

`internal/yarnmodernsource/conformance_test.go` covers shared S01/S03/S06/S08 and modern Yarn N01/N02/N04/N05/N10/N11/N12/N13 semantics, including:

- canonical lock-order identity, exact runtime/config/plugin identity, conditions and checksums;
- plugin, Git, patch, lock/runtime/target/checksum negative vectors with zero manager execution;
- raw archive normalization, normalized ZIP inspection, lifecycle/native rejection;
- poisoned/preseeded state removal and declared patch retention;
- protected private-cache replay through exact C0/C5 authority for both PnP and node-modules linkers;
- immutable-cache tamper rejection, PnP regeneration, installed payload reconciliation, and protected Node invocation.

Focused statement coverage is 68.8%. The uncovered statements are predominantly defensive filesystem/archive/parser error branches; both positive linker paths and the security-critical extension, lifecycle, native, ambient-state, checksum, offline, C0/C5, and cache-drift paths execute in tests.

## Validation evidence

Every command below ran as a standalone process; exit codes are literal.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/yarnmodernsource` | 0 | `.temp/TASK-260811-32iojo/go-test-focused-final.log` |
| `go test -count=1 -race ./internal/yarnmodernsource` | 0 | `.temp/TASK-260811-32iojo/go-test-race-final.log` |
| focused coverage command | 0 | `.temp/TASK-260811-32iojo/go-test-coverage-03.log`, `coverage-03.out`; 68.8% |
| `golangci-lint run` | 0 | `.temp/TASK-260811-32iojo/golangci-lint-full-final.log` |
| `go vet ./...` | 0 | `.temp/TASK-260811-32iojo/go-vet-full-final.log` |
| `go build ./...` | 0 | `.temp/TASK-260811-32iojo/go-build-full-final.log` |
| `go test -count=1 ./...` | 0 | `.temp/TASK-260811-32iojo/go-test-full-02.log`; modern Yarn package 4.681s |
| `git diff --check` and package `gofmt -l` emptiness gate | 0 | terminal gate after full suite |
| install task-local `@yarnpkg/cli-dist@4.9.2` with scripts disabled | 0 | `.temp/TASK-260811-32iojo/toolchain/`; validation-only tool |
| real Yarn online capture fixture | 0 | `.temp/TASK-260811-32iojo/real-yarn-capture-02.log` |
| real Yarn offline immutable replay without preseeded PnP/install state | 0 | `.temp/TASK-260811-32iojo/real-yarn-offline-02.log`; `Done in 0s 22ms` |

Expected-red exploratory gate: real Yarn with `--check-cache` plus `enableNetwork: false` exited 1 (`real-yarn-offline-01.log`, YN0080). Yarn's check-cache mode deliberately refetches the registry package, so it contradicts offline replay. The implementation removed that flag and independently verifies normalized ZIP SHA-512 against the lock before the immutable network-disabled install.

## Source identity

| File | SHA-256 |
| --- | --- |
| `capture.go` | `48aa1ae4515ad66dee084d8ff63301f13ece9ff83338cbb6981d165feac6928d` |
| `conformance_test.go` | `22f73fff84adf966d1a4d3cb946a1e1562c8b72cf1ecf20e5340010d63d75752` |
| `errors.go` | `61fc044d9291133c31146cd550512775024bd14ef2cbe3123b66262fc0137302` |
| `lock.go` | `5f6a5160ccf162a658039bf963f36c40f110e7f2bddcb08b998bad2408323fbb` |
| `materialize.go` | `441ba8ee62cc103bba46a0273a6779b6dce07564fa48398a6f56f721cd78c842` |

No files were staged or committed. Existing unrelated dirty-worktree changes were preserved.

## Logbook finding

Pinned Yarn 4.9.2 proves that `--check-cache` is incompatible with `enableNetwork: false`: it attempts a registry fetch and returns YN0080 even when the exact cache archive is present. The profile therefore owns checksum reconciliation before C6 and uses immutable/immutable-cache/skip-build during the network-disabled replay. A second implementation issue found during testing was overlapping project/cache work copies; this was replaced with one causally receipted immutable replay tree so process inputs are independent of receipt sort order.
