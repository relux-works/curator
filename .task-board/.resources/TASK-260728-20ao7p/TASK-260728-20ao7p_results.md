# TASK-260728-20ao7p — curator-rc5-e2e-docs-macos-windows

Developer rework for Windows gate fix 1, run RUN-261001-32a6ee.

## Root cause and changes

The Go compiler already derives `bin/<command>.exe` for Windows. The external
adapter reads that executable into bytes; `DiskProtectedStore.StoreArtifact`
then wrote those bytes to the extensionless physical cache file `artifact`.
Both `stageExternalBuilds` and `stageRuntimeAndShims` reconstructed that same
extensionless path, and the Windows `.cmd` shim invoked it. This explains the
reported hosted run 36796244453 failure: cmd.exe could not execute the cache
file despite a valid compiled binary and logical receipt path.

`buildrepo.CacheArtifactName` now derives `artifact.exe` for Windows and
`artifact` for other targets. The protected writer and reader derive that name
from receipt input (including receipt-3's nested build input). Install staging
uses the compiler's target GOOS for both hits and misses; shim publication
carries the staged basename into the final protected cache target. Existing
adoption, Windows security, and crossconformance fixtures now address the
actual executable file, preserving their tampering checks.

The blackbox HOME is created with `privatedir.Make` and proved with
`privatedir.Validate`, yielding the inheritance-protected owner-only DACL on
Windows. The lifecycle now fails if successful installs report that cache
sweep was skipped. It preserves shim execution, cache-hit/no-rewrite checks,
and removal assertions. No Windows skip was introduced.

## Receipt, marker, and spec consistency

The pinned rc.13 [Protocol Core §4.2.2](https://github.com/relux-works/curator-spec/blob/23435129ebc4c29e5b7f75ec72a0aa0cd3f16065/protocol/core.md#422-schema-7-external-repositories-and-go-repository-v1)
derives the logical artifact path solely from the consuming command key:
`bin/<command>` on Unix and `bin/<command>.exe` on Windows. Private staging
and shim paths are independently manager-derived. The physical protected
cache filename is implementation-private.

Receipt metadata continues to use `buildmeta.ArtifactPath`; marker records
continue to use the same function in `externalMarkerBuild`. No receipt input,
schema, marker field, execution policy, cache-key derivation, or namespace
changes. The focused production-store test checks writes, exact cache hits,
logical receipt paths, and derived cache keys for 3 target OSes × 2 receipt
versions (6 positive cases); the 2 Windows cases additionally reject intact
bytes renamed to an extensionless legacy cache filename. Such an older
Windows entry fails ordinary artifact verification and follows existing
untrusted-cache recovery rather than being silently accepted.

## Native fixture and author guide

The existing native blackbox builds the real CLI once with `testcli.Binary`
and uses a committed local checkout selected through schema-2
`Skillfile.dev.json`. It exercises admission refusal before pruning the
fixture's extra Git administration children, then install/build/cache/shim,
shim output, a cache-hit reinstall, and remove/reconcile. It makes no remote
repository fetch. Schema-7 published repositories accept HTTPS/SSH, so the
local development substitution is the valid no-network fixture surface.

The author guide is present at `docs/external-build-repositories.md`, linked
from README. It covers schema 7, locked commits and optional tags, local
substitutions and their narrow admission layout, admission → audit → cache →
compiler, and the specification's full authoring rules.

The pre-existing platform ledger row permits the Linux exclusion because
rc5-native-control-inventory-v1 has no Linux execution record. Its exact
runtime skip reason is: “rc5-native-control-inventory-v1 defines no record for
host linux; the portable execution policy is specified for macOS and Windows
only”. The Linux refusal has existing positive inventory tests. No additional
CI/workflow changes were made during this rework; the existing blackbox ledger
row is retained. No other PR #17 changes were revived.

## Validation run directly by this developer

Host: darwin/amd64; Go 1.26.0. No gate command was piped through tee.

- `go test ./internal/buildrepo -run 'TestProtectedArtifactNameMatchesTargetPlatform|TestReceiptArtifactPathMatchesTargetPlatform|TestExternalReceiptV2CacheKeyVector|TestAdoptArtifact' -count=1` — exit 0.
- `go test ./cmd/curator -run NativeBlackbox -count=1 -v` — exit 0, lifecycle passed in 281.41s (package 283.425s).
- `go test ./internal/buildrepo -count=1` — exit 1, failed (601.433s). `TestResolvedTransportTotalDeadlineBoundsSlowFetch` observed one fetch instead of two; the process subsequently hit the default 10-minute package timeout in `TestResolvedTransportClosedGrammarTable/https-connect-unreachable/.../positive`, while validating the fixture Git tool. This is not a passing suite and is not an expected-red gate. Transport implementation and transport tests were not changed. The cause of the broad-run transport failure is unconfirmed.
- `go test ./internal/buildrepo -run 'Protected|Adopt|Receipt|Pipeline|Collect|PrepareNamespaces|Signing|Substitution' -count=1 -timeout=4m` — exit 0 (22.610s), bounded rerun covering the changed cache/receipt/adoption scope.
- `go test ./internal/buildrepo -run '^TestResolvedTransportTotalDeadlineBoundsSlowFetch$' -count=1 -timeout=2m` — exit 0 (8.775s). The broad-run assertion did not reproduce in this isolated rerun; no transport code or timing limit was changed.
- `go test ./internal/buildrepo -run '^TestResolvedTransportClosedGrammarTable$/^https-connect-unreachable$' -count=1 -timeout=2m` — exit 0 (9.433s), isolated rerun of the case active when the broad process timed out.
- `go test ./internal/install -run 'External|Repository' -count=1` — exit 0.
- `go test ./internal/crossconformance -run 'External|Guard' -count=1` — exit 0.
- `env GOOS=windows go vet ./...` — exit 0.
- `golangci-lint run ./internal/buildrepo/... ./internal/install/... ./internal/crossconformance/... ./cmd/curator/...` — exit 0, 0 issues.
- `go build -o .temp/TASK-260728-20ao7p/curator-gatefix ./cmd/curator` — exit 0.
- `gofmt -l` over all touched Go files — exit 0, no output.
- `git diff --check` — exit 0.

Observed native output: ordinary checkout install exit 1 with
`build_repository_source_unavailable`; admitted install exit 0 with
`outcome=would-preflight-and-build`; shim exit 0 with `blackbox-tool-ok`;
reinstall exit 0 with `outcome=cache-hit` and the same cache key
`sha256:b8b5926b82db0783c6c3ecbc7e536142c2919206fe12217c30d8e7b750592863`;
remove exit 0 with `removed skill-a`; reconcile exit 0. The filesystem
assertions verified no publication after refusal, untouched artifact mtime and
receipt on the cache hit, and removal of the shim, installed skill and build
root. No cache-sweep-skipped warning was emitted.

Windows-native execution and the hosted runner gate were not run here: this
host is macOS, and the runner produces that evidence after handoff. No prior
attached test outcome is used to claim these local checks passed. Full
repository tests and race tests were not run; the commands above cover the
changed production seam and external install lifecycle.

## Entry text (no CHANGELOG or LOGBOOK edit)

The native lifecycle blackbox exposed a Windows external-build production
defect: the protected cache stripped the executable suffix before shim
publication. Retain `.exe` on the physical Windows artifact and use the same
name for store lookup, install staging, and shim targets. Give the test HOME
the product's private-directory security shape so cache sweep is exercised.
Preserve logical receipt/marker paths and cache input/key compatibility with
rc.13. Windows qualification remains with the hosted gate.

Work remains uncommitted in the supplied Story worktree for review handoff.
