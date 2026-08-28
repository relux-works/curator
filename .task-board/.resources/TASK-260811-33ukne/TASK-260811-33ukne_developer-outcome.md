# TASK-260811-33ukne developer outcome

Status target: `to-review`

## Delivered

- Added `internal/swiftpmsource` implementing the restricted
  `swiftpm-source-v1` resolution and C4 source-closure boundary.
- Binds exact Swift, SwiftPM, and PackageDescription identities in C0 and
  immediately rechecks them before every permitted manifest, resolution, and
  offline replay process seam.
- Captures and recursively admits the complete root tree before root manifest
  evaluation, and every source-control or contained local package tree before
  its own manifest evaluation.
- Selects version-specific `Package@swift-*` manifests deterministically and
  commits canonical manifest permits and receipts.
- Parses and freezes supported `Package.resolved` v2/v3 bytes, verifies a
  supplied lock against the admitted root copy, and supports a separately
  permitted broker-only generated-lock path.
- Reconciles every direct/transitive source-control declaration to one exact
  pin, broker receipt, immutable snapshot, Git revision/tree, and same-kind
  local mirror. Duplicate, dangling, wrong-origin/kind/revision, absent mirror,
  submodule, LFS, checkout-filter, hook, and escaping local-path shapes reject.
- Emits a destination-neutral package/product/target/source/condition
  `CaptureGraph`; concrete destination and Swift toolchain nodes/edges occur
  only in `SelectionBinding`; derives the exact active graph through shared C4
  projection.
- Replays all manifests and SwiftPM metadata from admitted trees and mirrors
  with fresh private roots, `--force-resolved-versions`, native build-system
  selection, experimental prebuilts disabled, and `network=none`.
- Rejects every dormant or active binary target, recursively rejected compiled
  bytes, selected plugins/macros, selected unsafe settings, source inventory
  drift, manifest drift, graph drift, and toolchain drift before the affected
  process.
- Added semantic conformance tests for R01-R13, P01-P08, selection-neutral
  CGP05 behavior, pre-process CGP11/CGN16-CGN18 behavior, local packages,
  transitive mirrors, and a real SwiftPM `dump-package` smoke.
- Documented the profile and exact test entry point in `README.md`.

The downstream SwiftPM C-family interop and build tasks continue to own module
map/header/compiler-read validation and C5-C7 compilation/publication. This
task does not widen that boundary.

## Validation evidence

| Gate | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/swiftpmsource` | 0 | Semantic and real SwiftPM vectors pass. |
| `go test -count=1 -cover ./internal/swiftpmsource` | 0 | 80.4% statement coverage. |
| `go test -count=1 -race ./internal/swiftpmsource` | 0 | Race detector clean. |
| `go vet ./internal/swiftpmsource` | 0 | Clean. |
| `golangci-lint run ./internal/swiftpmsource` | 0 | `0 issues.` |
| `make no-broad-suppression` | 0 | `no-broad-suppression: ok`. |
| shared artifact/closure/SwiftPM package test set | 0 | artifactpolicy, closureexec, closuregraph, and swiftpmsource pass. |
| `go build ./...` | 0 | Repository builds. |
| `go test -count=1 ./...` | 0 | Full repository suite passes; raw log attached separately. |
| canonical golden verifier | 0 | 53 records pass; CGP05 capture reused and references resolve. |
| `git diff --check` | 0 | Clean. |

Full-suite log SHA-256:
`ad8b0d56ff878d4e587969a2d869537215c7bd83f8deb1208fa7b93924404078`.

No files were staged or committed. Existing unrelated dirty-worktree changes
were preserved.
