# TASK-260827-18tswm developer outcome

## Outcome

The adapter delivery's known x86_64 Linux/macOS CI failures are either fixed at
their actual source or converted to narrowly classified host-capability/opt-in
skips. No assertion was weakened, no toolchain descriptor or digest was
invented, and the release-pin promotion and workflow `SPEC_PIN` are untouched.

The working-tree patch is unstaged and uncommitted for orchestrator review and
integration.

## Failure disposition

| Area | Root cause | Disposition | Skip reason / class |
| --- | --- | --- | --- |
| Rust source tests | The closed Cargo registry approves only `aarch64-apple-darwin`; x86_64 runners cannot establish an approved native Cargo identity. | Added a read-only native-descriptor availability query and required it in all 13 real-manager/registration tests. A present descriptor with a byte-digest mismatch remains fatal. | `no operator-approved Cargo descriptor for native target <target>` / `host-capability` |
| Cross-conformance Rust path | Rust registration failure prevented every-path projection and then tripped completeness. | Other adapters and Rust shared-artifact admission still run. The suite records Rust unavailability and permits exactly the six enumerated Rust manager-obligation gaps; any additional gap fails. | Same native-descriptor reason / `host-capability` |
| pnpm unclassified skips | Captured runners lack the pinned executable/profile. | Classified the verbatim observed shapes. | `pinned pnpm executable unavailable` or `pnpm ... is outside pinned profile ...` / `host-capability` |
| Yarn Classic unclassified skip | Captured runners lack the exact classic tool, or expose it outside the previously assumed Homebrew layout. | Classified absent exact Yarn; added support for Homebrew `libexec` and global package-root `bin/yarn.js`, with positive and negative layout tests. | `Yarn Classic is not installed`, `exact Yarn ... is unavailable`, or `exact Yarn Classic ... is installed without a supported pinned tool root` / `host-capability` |
| Yarn Modern unclassified skips | The real integration requires an explicit path to the pinned `@yarnpkg/cli-dist` entry point. | Classified the exact env-var instruction instead of relying on the generic `=1` opt-in regex. | `set CURATOR_TEST_YARN_MODERN_JS to the @yarnpkg/cli-dist 4.9.2 bin/yarn.js integration tool` / `opt-in` |
| Verified-provider CLI cases on Linux | The native-control inventory has no Linux record. | Both named tests now call the existing `requireNativeControlInventoryPlatform` guard. | Existing inventory-named reason / `platform-control` |
| Ubuntu Swift | Swift is installed but cannot link a manifest: the captured error combines `Invalid manifest` with clang `posix_spawn failed`. | Added one exact compound-error classifier shared by the affected SwiftPM integration tests. Merely having Swift, or any other Swift failure, remains fatal. | `SwiftPM manifest linker unavailable: clang posix_spawn failed while linking a manifest` / `host-capability` |
| npm extra package | Real tests used a Darwin/arm64 target fixture on Linux, where npm correctly installed a Linux-only optional dependency outside that target graph. | The two native real-tool cases now derive npm OS/CPU/libc from the host. Parser and cross-target tests retain the frozen Darwin fixture. Linux expects the optional package in its native materialization. | Fixed; no skip |
| pnpm ambient symlink | The fake runner attempted `os.Symlink("", link)` for the intentionally omitted dependency, causing raw Linux ENOENT before production validation ran. | The fake runner now omits that link, allowing the real validator to return the asserted `closure_graph_incomplete`. | Fixed; no skip |
| macOS Yarn Classic `~/.yarn/libexec` | The GitHub runner's exact Yarn uses the global package-root layout rather than Homebrew `libexec`. | Both supported pinned layouts are discovered and tested. Unsupported exact layouts refuse precisely. | Fixed for the captured layout; classified fallback as above |

The captured Ubuntu skip stream also contained
`OS-level network-denial harness is unavailable`; its verbatim reason is now
classified as `host-capability`.

## Local validation

All commands below were run directly as standalone processes. Logs are under
`.temp/TASK-260827-18tswm/`.

| Validation | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/rustsource` | 0 | `go-test-rustsource.log` |
| `go test -count=1 ./internal/crossconformance` (final rerun) | 0 | `go-test-crossconformance-03.log` |
| `go test -count=1 ./internal/npmsource` | 0 | `go-test-npmsource.log` |
| `go test -count=1 ./internal/pnpmsource` | 0 | `go-test-pnpmsource.log` |
| `go test -count=1 ./internal/yarnclassicsource` | 0 | `go-test-yarnclassicsource.log` |
| `go test -count=1 ./internal/yarnmodernsource` | 0 | `go-test-yarnmodernsource.log` |
| `go test -count=1 ./internal/swiftpmsource` | 0 | `go-test-swiftpmsource.log` |
| `go test -count=1 ./internal/swiftpmbuild` | 0 | `go-test-swiftpmbuild.log` |
| Targeted two-case `cmd/curator` run | 0 | `go-test-cmd-curator-targeted.log` |
| `go build ./...` | 0 | `go-build-all-02.log` |
| `go vet ./...` | 0 | `go-vet-all-03.log` |
| Pinned `golangci-lint` 2.12.2 | 0 | `golangci-lint-2.12.2-02.log` |
| `test -z "$(gofmt -l cmd internal)"` | 0 | `gofmt-check.log` |
| `.github/ci/gate-selftest.sh` | 0 | `gate-selftest.log` |
| Ledger consistency | 0 | `ledger-consistency.log` |
| No-broad-suppression gate | 0 | `no-broad-suppression.log` |
| Captured macOS platform-case replay | 0, zero `UNCLASSIFIED` | `platform-gate-macos.log` |
| Captured Ubuntu Tier-2 skip classification | 0, zero `UNCLASSIFIED` | `skip-class-gate-ubuntu.log` |
| Synthetic new Rust/Swift/Yarn reasons through platform gate | 0 | `new-host-skip-gate.log` |
| `git diff --check` | 0 | `git-diff-check.log` |
| Release-pin/workflow scope check | 0 | `pin-scope-check.log` |
| `task-board validate` | 0 | `task-board-validate.log` |

## Evidence boundaries and non-green attempts

- The first `go vet ./...` exited 1 because the pinned conformance submodule was
  not initialized in this worktree. `git submodule update --init --recursive`
  exited 0, checked out the pinned revision, and the final vet rerun exited 0.
- An intermediate cross-conformance rerun exited 1 after the Rust availability
  query was temporarily placed in an internal `_test.go` file and was therefore
  unavailable to the external package. The query was moved to the closed
  production registry as a read-only lookup; the final package rerun exited 0.
- The full Ubuntu replay of historical run `33033124914` exits 1 even though it
  reports zero `UNCLASSIFIED` skips: the current ledger requires
  `internal/godriver::TestModuleRootVectorsDriveTheWholeBuild`, which was not in
  that older test stream. The isolated Tier-2 classification replay exits 0.
- The exact full naming workflow scan was interrupted at exit 130 because this
  task worktree contains a very large ignored `.temp` evidence tree. A
  tracked-files-only diagnostic exits 1 on pre-existing `.research` and board
  prose unrelated to this patch. No naming claim is made from those attempts.
- The authoritative served-spec interop gate was not run because
  `CURATOR_CONFORMANCE_ROOT` is unset and no served conformance manifest exists
  in this worktree. The affected `internal/crossconformance` package did run.
- `task-board validate` exits 0 but reports 1,742 pre-existing board-wide
  consistency/resource findings across unrelated items. It reports no
  task-specific failure for this outcome, but its output is not represented as
  a clean board-wide validation.
- This developer run cannot push the branch and, per the assignment, does not
  own CI watching. Therefore the full remote Linux/macOS/Windows matrix, passing
  run URL, and fresh test-evidence artifacts remain unverified and must be
  supplied by the landing orchestrator after integration. They are not inferred
  from local results.
