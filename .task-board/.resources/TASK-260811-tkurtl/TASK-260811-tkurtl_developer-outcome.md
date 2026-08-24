# TASK-260811-tkurtl — SwiftPM C-family interop validation

Status: ready for review

Branch: `codex/legacy-board-repair` (uncommitted working tree; nothing staged)

Date: 2026-08-24

## Deliverable

New package `internal/swiftpminterop` implements the SwiftPM mixed Swift and
C-family boundary of `swiftpm-source-v1`. It consumes an accepted
`*swiftpmsource.Capture` (TASK-260811-33ukne, run RUN-260823-d5f73b, commits
`f8b7cc7`/`40142af`), proves the C-family target, module-map, header, system,
and interop contract, and republishes the closure as an extended capture graph
plus one exact selection binding, active projection, and chained C4 checkpoint.

The stage never compiles, links, or starts any process. Observed compiler read
sets arrive only through the manager-owned `ReadSetProvider` seam, and portable
assurance reports `not-observed` truthfully.

| File | Purpose |
| --- | --- |
| `errors.go` | Shared diagnostic vocabulary; `Code` is a type alias of `swiftpmsource.Code` so no cause is renamed at the phase boundary. |
| `types.go` | `Config`, `ExternalComponent`, `SystemLibrary`, `PlatformProfile`, `ReadSetProvider`, `TargetInterop`, `Boundary`, `Result`. |
| `language.go` | Closed five-language classification and separate Swift/Clang capture-target families. |
| `modulemap.go` | Clang module-map lexer/parser (`clang-modulemap-lexer-v1`). |
| `headers.go` | Public-header inventory with exact digests, SwiftPM generated-module-map reproduction, C-family include/`@import` scanner. |
| `containment.go` | Admitted-root and selected-external-root confinement with symlink and existence rules. |
| `platform.go` | Exact external identity validation, destination-profile gating, time-of-use recheck. |
| `boundaries.go` | Typed interop-boundary derivation and read-set verification. |
| `graph.go` | Capture/binding/active/checkpoint publication. |
| `manager.go` | Production entry point. |

`internal/swiftpmsource` gained three additive accessors only —
`PackageEvidence.ProtectedRoot`, `Capture.Destination`,
`Capture.SelectionToolchain`. No accepted behavior, record, or digest changed.

## Design decisions

1. **Separate capture targets, hard mixed-language rejection.** `classifyTarget`
   derives `KindSwift`, `KindClang`, or `KindSystem` from the admitted source
   extensions. A target holding both Swift and any C-family source rejects with
   `swiftpm_mixed_language_target_unsupported` before any header, module, or
   boundary analysis. Plugin, macro, and binary targets reject with their own
   stable codes if they ever reach this stage.

2. **Module maps are parsed, never trusted.** The accepted research proved that
   a custom module map can name an absolute external header while
   `swift package describe` omits both the module map and the header. The
   parser resolves the full grammar this profile admits — `module`,
   `explicit module`, `framework module`, `extern module`, attribute lists,
   `header`/`textual`/`private`/`exclude`, umbrella headers and directories,
   header attribute blocks, `link`, `link framework`, `requires`, `export`,
   `export_as`, `use`, and `config_macros`. Anything it cannot resolve exactly
   is `swiftpm_modulemap_escape`, never a silent skip.

3. **SwiftPM's generated module map is reproduced, not assumed.** When a target
   has no custom map, `GenerateModuleMap` applies SwiftPM's documented rules
   (umbrella header named after the module, single directly contained header,
   otherwise umbrella directory) and the reproduction is parsed and confined on
   the same path as a custom map.

4. **Every read resolves to admitted source or one selected binding node.**
   `roots.resolve` classifies each path as `admitted_source`,
   `selected_binding`, or `undeclared`. Admitted resolution walks every path
   component and rejects symlinks and special nodes; selected-external
   resolution requires the entry to actually exist, so a path spelling that
   merely starts with an SDK root grants nothing. Include scanning covers
   quoted, angled, and `@import` forms across every admitted C-family source
   and header.

5. **Selection neutrality is preserved.** Interop capture records carry only
   contract identities (`c-abi-v1`, `itanium-cxx-abi-v1`,
   `cxx-standard-library-v1`, `objc-runtime-v1`). The exact platform, Swift and
   Clang toolchains, SDK, sysroot, and system components — and every `targets`,
   `uses_tool`, toolchain-scoped `requires`, and `provides_interop` edge that
   binds them — exist only in the selection binding. Darwin and Linux produce
   byte-identical interop capture and different bindings.

6. **Restricted languages are gated by an accepted destination profile.**
   Objective-C and Objective-C++ need `PlatformProfile.ObjectiveCRuntime`;
   C++ needs `CxxStandardModes`; direct Swift/C++ import additionally needs
   `CxxInterop` and an explicit `interoperabilityMode(.Cxx)` on the consuming
   Swift target. An unlisted destination triple rejects before any admitted
   byte is classified.

7. **Interop mode is derived from declared evidence, not from a successful
   build.** A provider whose admitted public interface is a C++-only header
   (`.hh`/`.hpp`/`.hxx`/`.h++`) consumed by a Swift target without the opt-in
   is `closure_interop_undeclared`; SwiftPM never propagates the mode
   implicitly. A provider target containing Objective-C or Objective-C++ binds
   the Objective-C runtime boundary regardless of which symbols the consumer
   calls, because the target links that runtime. This is a deliberate modeling
   choice: it is strictly more conservative than inferring a C-only edge.

8. **System libraries are external trust, never captured source.** A
   `.systemLibrary` target is admitted only when a Curator-selected
   `ExternalComponent` exists, its module map resolves inside that component's
   contained roots, and every reference in it resolves to the same selected
   binding. The boundary provider is then the toolchain component itself, so an
   external declaration never masquerades as an admitted producing target. A
   Homebrew, `/usr/local`, or user path is `artifact_toolchain_untrusted`.

9. **Link edges are confined.** Every `link` and `link framework` name must
   already be declared by a selected component; an unvetted library or
   framework is `artifact_toolchain_untrusted`.

10. **Portable-honest, fail-closed read evidence.** With no provider, or with a
    provider that reports `Observed: false`, `ReadSetEvidence.Mode` is
    `not-observed` and the declared static closure remains the proof. Under
    `closureexec.AssuranceVerified` an unobserved read set fails closed with
    `swiftpm_header_input_undeclared`. Observed evidence without its issued
    derivation receipt is `closure_derivation_unauthorized`.

11. **Exact toolchain identity is rechecked immediately before use.** The Clang
    and Clang++ identities are rechecked before the first admitted byte is
    read; drift is `artifact_toolchain_identity_changed` and no classification,
    module-map parse, or graph publication follows.

12. **`ErrorCode` resolves one vocabulary.** Shared closure services raise their
    stable code as the leading token of a plain error rather than as a typed
    adapter failure. `swiftpminterop.ErrorCode` resolves both encodings against
    a closed set, so callers see one code space.

## Conformance coverage

58 top-level tests, 96 including subtests, 86.1% statement coverage.

| Family | Vectors | Where |
| --- | --- | --- |
| Language/target | `S02`–`S09` plus destination-triple and unsafe-flag rejection | `language_test.go` |
| Headers/modules/system | `H01`–`H08` plus extern-module escape, malformed grammar, admitted link edges | `modulemap_test.go` |
| Extensions/artifacts | `P01`–`P09` | `extension_test.go` |
| Cross-language graph | `CGP05`, `CGN03`, `CGN09`, `CGN15`, checkpoint chaining | `conformance_test.go` |
| Read sets and authority | Portable/verified assurance, undeclared observed read, receiptless evidence, sysroot binding, incomplete identities, manager contract | `readset_test.go` |
| Grammar/classification | Full module-map grammar, generated-map selection rules, closed language set, Clang module imports, conditional projection | `parser_test.go` |
| Process boundary | Cross-adapter `exec.Command` guard extended to this adapter surface | `guard_test.go` |

Notes on specific vectors:

- `P05` is enforced continuously: the fixture manifest evaluator rejects any
  permit whose argv omits `--disable-experimental-prebuilts`, so every positive
  test in the package is also a P05 assertion.
- `P02`, `P04`, `P06`, `P07`, and `P08` are rejected by the upstream source
  closure before this stage runs; each test asserts the exact shared code and
  additionally drives `classifyTarget` directly, so the interop stage would
  reject the same shape independently.
- `CGN15` and `CGN09` structural cases re-project the exact published records
  after an adversarial mutation (duplicate logical key, dangling endpoint,
  wrong-kind endpoint, slot bound twice, omitted/duplicated `targets` edge,
  capture replacement) and assert `closure_graph_reference_invalid`.
- The cross-adapter guard now covers `internal/swiftpminterop`,
  `internal/swiftpmsource`, and `internal/closureexec`; the allowlist is
  unchanged (`acquisition.go`, `portable_runner.go`) and this package does not
  import `os/exec` at all.

## Validation evidence

Every command below ran directly as a standalone process from the repository
root. Reported codes are real exit codes.

| Gate | Command | Exit | Artifact |
| --- | --- | ---: | --- |
| Focused package tests | `go test -timeout 20m -count=1 ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closureexec/ ./internal/closuregraph/ ./internal/artifactpolicy/...` | 0 | `focused-tests-01.log` |
| Race, focused set | `go test -race -timeout 20m -count=1 ./internal/swiftpminterop/ ./internal/swiftpmsource/` | 0 | `race-tests-01.log` |
| Lint, changed packages | `golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/...` (v2.12.2) | 0 | `lint-01.log` |
| Lint, whole repository | `golangci-lint run ./...` (v2.12.2) | 0 | `lint-full-01.log` |
| Repository suite minus `cmd/curator` | `go test -timeout 30m -count=1 $(go list ./... \| grep -v cmd/curator)` | 0 | `suite-no-cmd-01.log` |
| `cmd/curator` bounded subset | `go test -timeout 9m -count=1 -run 'TestStatus\|TestGlobalStatus\|TestGC\|TestToolchainHost\|TestLifecycle' ./cmd/curator/` | 0 | `cmd-curator-subset-01.log` |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | `canonical-verifier-01.log` |
| Build | `go build ./...` | 0 | inline |
| Vet | `go vet ./...` | 0 | inline |
| Format | `gofmt -l $(git ls-files '*.go')` — empty | 0 | inline |
| Whitespace | `git diff --check` | 0 | inline |
| Board | `task-board --no-update-check validate` | 0 | `Board is valid. No issues found.` |

Log digests (SHA-256):

| Artifact | SHA-256 |
| --- | --- |
| `focused-tests-01.log` | `f6892fcace053274ec50a51796aef0af61089a6cffd2712ae824694c121168cc` |
| `race-tests-01.log` | `bf420dcfda4185e2031bb97a73bde23735005e473c6d65650986f75b61c21803` |
| `lint-01.log` | `e92606b0bf483111dff0a120c315ea165821348f31365020e2468a0059095c47` |
| `lint-full-01.log` | `e92606b0bf483111dff0a120c315ea165821348f31365020e2468a0059095c47` |
| `suite-no-cmd-01.log` | `04f94f7a0983c9083831f4b299876031ea5261bbc7476eb33e28fff619137626` |
| `cmd-curator-subset-01.log` | `140fbe08da256e9824312df9eb0035a4c1b6c8ef6f5f51ac119a3d69dccb509d` |
| `canonical-verifier-01.log` | `1847364de63c9d8e706d54739f97d05e8406cca5f9cd7aec4bb12c028998b75a` |

Slowest packages in the repository run: `internal/rustsource` 137.248s,
`internal/artifactpolicy` 127.147s, `internal/install/atomicity` 110.464s,
`internal/install` 110.360s, `internal/godriver` 87.954s,
`internal/npmsource` 82.478s.

## Not run

The monolithic `go test ./...` including the full `cmd/curator` package was not
run in this session: a single bounded call cannot hold it (`cmd/curator` alone
is roughly ten minutes) and this run may not background long commands. Instead
the complete repository suite minus `cmd/curator` ran green in one call, plus a
bounded `cmd/curator` subset. `go list -test -deps ./cmd/curator` contains no
`swiftpm*` package, and no file in its dependency closure changed, so the delta
cannot reach it. The Orchestrator should still run the monolithic suite before
acceptance if the review contract requires it.

No commit and no staging were performed.

## Follow-on

`TASK-260811-2qfnai` (SwiftPM offline build) consumes `Result.Graph`,
`Result.Binding`, `Result.Active`, and `Result.C4` and supplies the verified
`ReadSetProvider` that turns `ReadSetEvidence.Mode` from `not-observed` into
`observed`. `TASK-260811-x611eq` consumes `Result.EvidenceDigest` as the
adapter-level conformance anchor. Neither was started.
