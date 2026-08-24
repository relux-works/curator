# TASK-260811-2qfnai rework outcome — F1/F2 and the test gaps that hid them

Producer run: focused rework after reviewer `RUN-260824-145133`
(verdict: changes requested → `to-dev`, 8 of 10 scope items accepted).

Scope of this run: reviewer findings F1 and F2, the three test gaps the
reviewer named as load-bearing, and the five cheap non-blocking nits. Nothing
else. Every accepted item was preserved.

## F1 — the observed read set no longer discards a dependency package's reads

`readset.go` classified any read below `<work>/.curator/` as `inBuildTree` and
`harvestDependencyFiles` dropped it. That is right for the module cache and the
module maps SwiftPM generates, and wrong for `<work>/.curator/scratch/checkouts/
<identity>/...`, which is a source-control dependency's **admitted source and
headers**. The verified header proof — the keystone of this task — therefore
covered only the root package.

Changes:

- `mapObservedRead` now takes the root identity plus the **full** per-package
  admitted-root map (`admittedPackageRoots` already built it; `readset.go:282`
  previously threw all but the root entry away — the reviewer's tell).
- A read under `.curator/scratch/checkouts/<identity>/…` is rewritten to that
  dependency's admitted protected root, exactly as a root-package read already
  was. `checkoutsSegments` names the prefix once.
- A `checkouts/` read whose first component matches **no** admitted package
  identity, and a bare read of `checkouts` itself, **fail closed** with
  `swiftpm_header_input_undeclared`. They are never dropped.
- The `inBuildTree` drop is now reserved for genuinely derived build state, and
  its comment says so.
- `harvestDependencyFiles` also fails closed when the capture admitted no
  package root at all, instead of indexing `Packages[0]` blind.

H03 behaviour is unchanged: an out-of-package absolute read is outside the work
root, is returned verbatim, and still fails closed downstream. `swiftpminterop`
static module-map containment remains as defence in depth.

## F2 — same-base-name sources in different directories now resolve

`readProducedObject` matched a produced `.o` on base name and then narrowed
with `strings.HasSuffix(candidate, slot.Source+".o")`. `candidate` is
**target**-relative (`a/x.c.o`); `slot.Source` is **package**-relative
(`Sources/CLib/a/x.c`), so the narrowing branch could only ever reduce the
match set to zero, and a legal one-Clang-target package with `a/x.c` + `b/x.c`
failed with `artifact_local_output_unreceipted`.

Changes:

- `swiftpmsource.Target.SourceRoot()` applies SwiftPM's documented convention
  default (`Sources/<Name>`, `Tests/<Name>` for a test target) exactly once;
  `enumerateTargetSources` now uses the same helper, so manifest normalization
  and downstream reconciliation cannot disagree.
- That root is carried through `swiftpminterop.TargetInterop.SourceRoot` into
  `swiftpmbuild.ObjectSlot.SourceRoot`.
- `resolveProducedObject` disambiguates on the **target-relative** source path
  (`targetRelativeSource`), compared for equality rather than by suffix, which
  is exactly the tree SwiftPM mirrors below `<Target>.build`.
- The produced object set is now **exhausted**: `requireNoUndeclaredObject`
  rejects any `.o` below a selected target build directory that no declared
  slot claims, and two slots resolving to one object are rejected too. This is
  what makes undeclared local generation fail closed now that resolution is
  exact rather than accidentally ambiguous.

## Test gaps closed

- `TestHarvestedReadSetCoversEveryAdmittedPackage` — new. Drives
  `harvestDependencyFiles` over a real multi-package capture with a pinned
  source-control dependency: the dependency target's `checkouts/` source and
  header reads must be rewritten to that dependency's admitted protected root,
  the generated module map below the scratch tree must still drop, and an SDK
  read must survive verbatim.
- `TestHarvestedReadSetFailsClosedOnUnadmittedCheckout` — new. A `checkouts/`
  read with no matching identity is `swiftpm_header_input_undeclared`.
- `TestObservedReadMappingSeparatesBuildTreeFromAdmittedSource` — extended with
  the dependency-checkout, unadmitted-checkout, and bare-`checkouts` branches.
- `TestS03SameBaseNameSourcesInDifferentDirectoriesResolve` — new conformance
  case: one Clang target, `Sources/CLib/a/x.c` and `Sources/CLib/b/x.c`, both
  objects resolve to **distinct** published bytes under their own declared
  logical paths, five observations and a five-entry write set.
- `TestUndeclaredGeneratedObjectFailsClosed` — repointed. It now plants a
  genuinely undeclared `CLib.build/generated/smuggled.c.o` and asserts both
  that the build fails closed with `artifact_local_output_unreceipted` and that
  the declared slot itself still resolves — so it can no longer pass for the
  wrong reason (previously both candidates were eliminated).
- `TestR01R05OfflineBuildMountsAdmittedInputsWithNetworkDenied` — repaired. The
  fixture now carries a real pinned remote source-control dependency with an
  admitted mirror, so the test asserts the read set is the admitted build root
  **plus** the mirror mount, that network is `none`, and that the dependency's
  own object publishes.
- `TestGeneratedMirrorConfigurationPreservesSourceControlKind` — new. Decodes
  the generated `.curator/config/mirrors.json` and asserts the
  `remoteSourceControl` origin maps onto `file://<execution-root>/inputs/
  mirrors/dep`.
- `TestDuplicateMirrorReceiptFailsClosed` — new. Two admitted mirrors sharing
  one intake receipt are rejected with `swiftpm_dependency_mirror_missing`,
  with zero process starts and no publication.
- `TestRealSwiftPMBuildMatchesThePlannedLayout` — extended with
  `Sources/CLib/a/shared.c` and `Sources/CLib/b/shared.c` against the real
  Apple Swift 6.3.2 toolchain, asserting every declared source resolves to a
  distinct produced object and that the resolved set exhausts the produced set.

Mutation check (recorded because it is the honest evidence that these tests are
load-bearing): with F1's `checkouts` branch disabled and F2's narrowing
reverted to the old suffix form, `TestS03SameBaseNameSourcesInDifferent
DirectoriesResolve`, `TestObservedReadMappingSeparatesBuildTreeFromAdmitted
Source`, `TestHarvestedReadSetCoversEveryAdmittedPackage`, and
`TestHarvestedReadSetFailsClosedOnUnadmittedCheckout` all fail. The product
files were restored immediately afterwards.

## Nits cleared

- `types.go` — `ObjectSlot` doc now says one slot per **source**, and the new
  `SourceRoot` field is documented.
- `plan.go` — the `var _ = closureexec.AssurancePortable` keepalive and the
  `closureexec` import are gone.
- `plan.go` — the link action's `DomainID` payload key `"configuration"` is now
  `"target_triple"`, which is what it actually holds.
- `binding.go` — a `SlotLinker` entry in `Config.Slots` is now **rejected**
  (`closure_graph_reference_invalid`) instead of silently ignored; the linker
  is bound from `Config.Linker`.
- `binding.go` — the edge-activation question is answered rather than changed:
  `closuregraph.projection` emits an activation for **conditional edges only**,
  and `closuregraph.validation` rejects an activation record on an
  unconditional edge, so treating an edge with no activation as selected is
  required by the shared contract. The invariant is now stated in a comment.

## Preserved

Binding overlay and identity, the single-resolution rule, C4→C4′→C5 chaining,
offline execution flags and isolation, the fail-closed matrix, publication
zero-start reuse, and seam/guard discipline are unchanged. No `os/exec` import
entered `swiftpmbuild`; the guard allowlist is untouched. No non-Go file was
added. No `TASK-260811-x611eq` work was started.

## Gates — all run directly as standalone processes, real exit codes

| Gate | Command | Exit |
| --- | --- | ---: |
| Focused suite | `go test -count=1 ./internal/swiftpmbuild/ ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closureexec/ ./internal/closuregraph/ ./internal/artifactpolicy/` | 0 |
| Real toolchain vector | `go test -count=1 -run TestRealSwiftPMBuildMatchesThePlannedLayout ./internal/swiftpmbuild/` | 0 |
| Race | `go test -race -count=1 ./internal/swiftpmbuild/ ./internal/swiftpminterop/ ./internal/swiftpmsource/` | 0 |
| Repository suite minus `cmd/curator` | `go test -timeout 9m -count=1 $(go list ./... \| grep -v cmd/curator)` | 0, 52 packages `ok` |
| Lint | `golangci-lint run ./...` (v2.12.2) | 0, `0 issues.` |
| Format | `gofmt -l ./cmd ./internal` | 0, empty |
| Vet | `go vet ./...` | 0 |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0, `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2` / `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Whitespace | `git diff --check` | 0 |
| Board | `task-board --no-update-check validate` | 0, `Board is valid. No issues found.` |

Not run in this session: the monolithic full suite including `cmd/curator`
(~400 s for that package alone, above the single-call cap). It belongs to the
Orchestrator. The 52-package run above covers everything else, and no file
outside `internal/swiftpmbuild`, `internal/swiftpminterop`, and
`internal/swiftpmsource` changed.

Nothing was staged or committed.
