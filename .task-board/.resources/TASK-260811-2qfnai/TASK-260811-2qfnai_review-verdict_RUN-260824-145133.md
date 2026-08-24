# Independent review verdict — TASK-260811-2qfnai SwiftPM offline build adapter

Reviewer run: `RUN-260824-145133` (Claude claude-opus-5, independent acceptance review)
Producer run under review: `RUN-260824-7abada`
Run goal binding: this run is not goal-bound (`task-board spawn goal` reported
`Active Goal: none`).

**Verdict: changes requested → `to-dev`.**

Two confirmed defects, both reproduced empirically against the real Apple Swift
6.3.2 toolchain on this host, sit inside the two things this task exists to
deliver: the observed-read authority (scope item 5) and the reconciliation of
declared outputs against real produced bytes (scope items 4/7). Everything else
in the mandatory review scope is accepted. The delivery is close: both findings
are local, precisely located, and each has a mechanical fix.

## Per-item verdicts

| # | Scope item | Verdict |
| ---: | --- | --- |
| 1 | Binding overlay / identity | accepted |
| 2 | Single-resolution rule | accepted |
| 3 | Identity preservation / chaining | accepted |
| 4 | Offline build execution | accepted with finding F2 |
| 5 | **Observed-read provider (keystone)** | **changes requested — F1** |
| 6 | Fail-closed matrix | accepted |
| 7 | Publication | accepted with finding F2 |
| 8 | Seam / guard discipline | accepted |
| 9 | Scope hygiene | accepted |
| 10 | Evidence | accepted |

---

## F1 — the observed read set silently discards every dependency package's reads

**Scope item 5. Confirmed, with a real-toolchain reproduction.**

`readset.go:321 mapObservedRead` classifies any observed read whose path is
below `<work-copy>/.curator/` as `inBuildTree`, and
`readset.go:286 harvestDependencyFiles` then **drops it** with the comment
"Reads of locally produced build outputs are covered by the same permitted,
network-denied derivation; they are not part of the admitted source closure."

That comment is true for the module cache and generated module maps. It is
false for dependency package source: SwiftPM materializes every source-control
dependency into `<scratch>/checkouts/`, which with the planned
`--scratch-path .curator/scratch` is `<work>/.curator/scratch/checkouts/...`.
Those are dependency **source and headers**, not build outputs.

Reproduction (real `swift` 6.3.2, `file://` source-control dependency, exact
planned argv shape):

```
$ cat .curator/scratch/arm64-apple-macosx/debug/CDep.build/*.d
dependencies: \
  /.../MacOSX26.5.sdk/usr/include/Darwin.modulemap \
  /private/tmp/2qfnai-mirror/root/.curator/scratch/checkouts/origin/Sources/CDep/dep.c \
  /private/tmp/2qfnai-mirror/root/.curator/scratch/checkouts/origin/Sources/CDep/include/CDep.h
```

Both `checkouts/...` reads start with `.curator/` relative to the work root, so
`mapObservedRead` returns `inBuildTree=true` and `harvestDependencyFiles` skips
them. The observed read set handed to `swiftpminterop` for target `CDep` is
reduced to the single absolute SDK path. `swiftpminterop.verifyReads`
(`boundaries.go:194`) only asserts *"every observed read resolves to something
declared"* — it asserts no coverage — so an empty-of-package read set passes
silently and `Reads.Mode` is still reported as `observed`.

Net effect: the keystone claim — that the verified header proof rests on what
the selected compiler actually read — holds only for the **root** package.
Every transitive dependency's compiler read set is discarded before it is ever
checked against the admitted roots. The producer's own outcome states the
opposite ("a read of the private work copy is rewritten to the admitted
protected tree it was copied from"); that is only implemented for the root.

Corroborating signal that this was an oversight rather than a decision:
`readset.go:246 admittedPackageRoots` builds a per-package root map for the
whole capture, and `readset.go:282` then passes only
`packageRoots[rootIdentity]`. The per-package map is computed and thrown away.

Not currently exploitable in production, and I want to be precise about that:
`ObserveReads` returns early unless assurance is `AssuranceVerified`, and
`closureexec.NewOSBoundary` fails closed on every platform
(`boundary_darwin.go`, `boundary_unsupported.go`), so no verified run can
execute today. This is a latent under-verification that becomes real the moment
an enforce-and-observe provider lands, in exactly the path that is the reason
this task exists. It also does not admit an escaping absolute header: an
out-of-package read (H03-shaped) is outside the work root, is returned
verbatim, and still fails closed — and `swiftpminterop`'s static module-map
containment remains as defence in depth.

Required change: map a read under `<work>/.curator/scratch/checkouts/<identity>/…`
back to that dependency's admitted protected root (the map is already built),
and reserve the `inBuildTree` drop for genuinely derived build state. A read
under `.curator/scratch/checkouts/` that matches no admitted package identity
must fail closed rather than be dropped.

## F2 — a legal C-family target with same-named sources in different directories is falsely rejected

**Scope items 4 and 7. Confirmed, with a real-toolchain reproduction.**

`build.go:320 readProducedObject` resolves a declared object slot by matching
the produced `.o` files under `<Target>.build` on **base name**, then, when more
than one matches, narrows with
`strings.HasSuffix(candidate, slot.Source+".o")`.

`candidate` is relative to `<Target>.build` and therefore **target**-relative
(`a/x.c.o`). `slot.Source` is **package**-relative
(`Sources/CLib/a/x.c`, from `swiftpmsource.Target.Sources` via
`swiftpminterop.interop.go:244`). The suffix can never match, so the
disambiguation branch is dead: whenever it fires it reduces the match set to
zero and the slot fails with `artifact_local_output_unreceipted`.

Real SwiftPM layout, verified on this host:

```
$ find .../CLib.build -name '*.o'   # sources Sources/CLib/a/x.c, Sources/CLib/b/x.c
a/x.c.o
b/x.c.o
```

Running the exact resolution logic against that real tree:

```
source=Sources/CLib/a/x.c  -> err=declared object slot did not resolve to exactly
                                 one produced object (matches=0, candidates=[a/x.c.o b/x.c.o])
source=Sources/CLib/b/x.c  -> err=... (same)
```

So a perfectly legal `swiftpm-source-v1` package — one Clang target with
`a/x.c` and `b/x.c` — cannot be built or published. It fails closed, so this is
not a security regression, but it violates the AC's "supported SwiftPM products
rebuild and execute offline" for a supported shape.

`TestUndeclaredGeneratedObjectFailsClosed` does not catch this: it plants
`CLib.build/nested/lib.c.o` alongside `CLib.build/lib.c.o`, and passes because
the narrowing eliminates *both* candidates — i.e. it passes for the wrong
reason and masks the defect it looks like it covers.

Required change: compare against the source path relative to the target's
declared path (`Target.Path`), which is what SwiftPM mirrors below
`<Target>.build`. Add a conformance case with two same-base-name sources in
different subdirectories of one Clang target, asserting both objects resolve
and publish.

## Findings that do not block acceptance

- `readset.go:246` / `readset.go:282` — `admittedPackageRoots` computes a
  per-package map and only the root entry is used. Fix with F1.
- `types.go:66` — `ObjectSlot`'s doc still says "exactly one slot per selected
  compile target"; the contract is now one slot per **source**.
- `plan.go:584` — `var _ = closureexec.AssurancePortable` exists only to keep an
  otherwise unused import alive. Drop the import instead.
- `plan.go:335` — the link action's `DomainID` payload stores the target triple
  under the key `"configuration"`. Identity is still correct (the output node's
  `LogicalPath` carries the real configuration), but the key is misleading.
- `conformance_test.go:71` `TestR01R05OfflineBuildMountsAdmittedInputsWithNetworkDenied`
  asserts `ReadSet == [inputs/build-root]`. The fixture has zero pins and zero
  mirrors (`"pins":[]`, `fakeBroker` errors), so the test proves nothing about
  mounting admitted mirrors read-only despite its name. The mirror mount, the
  duplicate-receipt rejection, and the generated `mirrors.json` kind mapping are
  untested end to end in this package.
- `ReadSetObserver.observe` and `harvestDependencyFiles` — the verified
  observation pass — have **no test at all**. Only the portable early-return,
  the authority contract, the grammar parser, and a three-case
  `mapObservedRead` unit are covered. This is the gap that let F1 through.
- `binding.go:99` — a `SlotLinker` entry in `Config.Slots` is silently ignored
  rather than rejected, since the linker is bound from `Config.Linker`.
- `binding.go:180` — edges carrying no activation record at all are treated as
  selected (`pruned` only holds explicitly non-selected activations). Benign if
  the shared projection guarantees an activation per edge; worth an assertion.

## What was accepted, and on what evidence

- **Binding overlay (1).** `Config.Slots` carries roles only; every physical
  identity is read back from the accepted C4 binding node via
  `resolveSlots`/`toolIdentity`. Missing role, unknown role, two slots on one
  node, wrong-kind node, dangling node, unreferenced node, duplicate component
  role, and unused slot all fail in `indexBinding`/`resolveSlots` before any
  process starts; `recheckSlots` runs again inside the executor's time-of-use
  callback (`build.go:216`). The linker is the only self-selected component and
  is published with its own `targets`/`requires` edges and `ToolchainSelector`
  (`plan.go:linkRecords`, `plan.go:buildAuthority`). Incomplete linker identity
  is `artifact_toolchain_untrusted`. `SlotClangCXX` is required exactly when a
  selected target carries C++/Objective-C++ source.
- **Single-resolution rule (2).** `validateActionBindings` proves it over the
  whole republished active graph, including that each `uses_tool` edge names the
  same executable as its bound component.
- **Identity preservation (3).** `preservesAcceptedIdentities` proves every
  accepted node and edge identity survives; `validateAcceptedChain` proves build
  C4′ chains from interop C4 which chains from source C4, rejects a foreign
  selection context, and requires a frozen lock. Capture stays
  selection-neutral (verified by the canonical verifier's CGP05 branches).
- **Offline execution (4).** Exact argv verified against the real toolchain:
  `--disable-experimental-prebuilts --force-resolved-versions --disable-netrc
  --build-system native` with isolated `--cache-path/--config-path/
  --security-path/--scratch-path`, one `--configuration`, one `--triple`, one
  `--product`. `HOME` and every `SWIFTPM_*` path resolve inside the private work
  copy; `PATH` is confined to `<execution-root>/bin`. Mirrors are mounted
  read-only from admitted intake receipts, network is `none`, and the output
  root must be empty before the permitted action. The `{execution-root}`
  placeholder keeps `CommandID` portable across two independent execution roots.
  I independently verified that `unversionedTriple` is right: passing
  `--triple arm64-apple-macosx26.0` really does produce
  `.build/arm64-apple-macosx/`, and that `description.json` really is emitted at
  `<scratch>/<unversioned-triple>/<configuration>/`, so the observation pass's
  declared evidence path is correct.
- **Fail-closed matrix (6).** Graph drift, missing mirror pin, unfrozen lock,
  missing declared object, missing product, pre-existing output root, ambiguous
  produced object, and output drift each fail with `requireNoPublication`
  asserted. Network/undeclared read/write/process are rejected by
  `reconcileCommand` and by `mapExecutionError`, which renames no shared cause.
- **Publication (7).** Every declared output is reinspected from the retained
  private work copy, hashed, validated against the immutable graph records with
  `ValidateAgainst`, sorted by exact observation identity, staged, and published
  atomically. `TestOfflineBuildPublishesAndReusesExactly` proves exact reuse from
  an independently derived expected input with **zero** process starts
  (`replan.starts == 0`) and identical artifact path and cache-input identity.
  `TestPublicationDoesNotMutateExpectedGraphRecords` covers non-mutation.
- **Seam/guard (8).** No `os/exec` import anywhere in `swiftpmbuild` production
  code; `guard_test.go` extends the cross-adapter guard over `swiftpmbuild`,
  `swiftpminterop`, `swiftpmsource`, and `closureexec` with the allowlist
  unchanged (`acquisition.go`, `portable_runner.go` only). Notably the offline
  build test is not a mock: it drives the real `ManagerProcessRunner`, which
  really forks a real process, and asserts the launched argv equals the
  committed argv.
- **Scope hygiene (9).** Zero non-Go files added under `swiftpmbuild`/
  `swiftpminterop`; no Kotlin or Gradle reference; no vendored binaries; no
  `TASK-260811-x611eq` work started. The `swiftpmsource` additions
  (`RootInput`, `OfflineMirrors`, `ProtectedRoot`, `Destination`,
  `SelectionToolchain`) are additive, and the generated `mirrors.json` mapping is
  byte-for-byte the same shape as the already-accepted
  `executor_runtime.go` replay, including the `file://` treatment of
  `remoteSourceControl` kinds.
- **The upstream interop fix is correct and correctly located.** SwiftPM's
  native build system does emit one object per source file — I verified this
  directly for flat and nested Clang sources and for multi-source Swift targets.
  The previous per-target `generated_artifact` declaration was both factually
  wrong and, because `PublicationEvidence.ValidateForPublication` and
  `ProtectedStore.Publish` require observations to cover exactly the declared
  outputs and to equal the write set, made the whole SwiftPM closure
  unpublishable. Fixing it in `swiftpminterop` rather than compensating in
  `swiftpmbuild` is the right call: a write slot must be bound exactly once, so
  production cannot be attached to an accepted action from a later stage.
- **Conformance labels (10).** Full `S01`–`S10`, `R01`–`R13`, `H01`–`H08`,
  `P01`–`P09`, and `CGP05` coverage exists across `swiftpmsource`,
  `swiftpminterop`, and `swiftpmbuild`; the split across the three leaves is the
  accepted decomposition.

## Gates — what I reran versus what I accepted from the producer's evidence

Rerun by me, directly, exit codes are real:

| Gate | Command | Exit |
| --- | --- | ---: |
| Repository suite excluding `cmd/curator` | `go test -timeout 25m -count=1 $(go list ./... \| grep -v cmd/curator)` | 0 (52 packages `ok`) |
| Focused suite | `go test -count=1 ./internal/swiftpmbuild/ ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closureexec/ ./internal/closuregraph/ ./internal/artifactpolicy/` | 0 |
| Real toolchain vector | `go test -count=1 -run TestRealSwiftPMBuildMatchesThePlannedLayout ./internal/swiftpmbuild/` | 0 |
| Lint | `golangci-lint run ./...` | 0, `0 issues.` |
| Format | `gofmt -l ./cmd ./internal` | 0, empty |
| Vet | `go vet ./...` | 0 |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0, `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2` / `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Board | `task-board --no-update-check validate` | 0, `Board is valid. No issues found.` |

Accepted from the orchestrator-attached monolithic log rather than rerun (the
`cmd/curator` package alone takes ~400 s and the whole monolith exceeds the
10-minute single-call cap):

- `TASK-260811-2qfnai_full-go-01.log`, SHA-256
  `d8c4366f0c9f9bc8336c47828d2670e47a225c717102c0f9d2aa0a68c838ca76`,
  trailing `EXIT:0`, first line
  `ok github.com/relux-works/curator/cmd/curator 402.374s`, covering every
  package including `cmd/curator`.

Empirical probes I ran against the real Apple Swift 6.3.2 toolchain, outside the
Go suite, to verify the two findings and three accepted claims: dependency-file
contents for a path dependency and for a `file://` source-control dependency;
`checkouts/` location under an explicit `--scratch-path`; `description.json`
location; scratch-directory naming under a versioned `--triple`; and the
duplicate-base-name Clang object layout plus a verbatim replay of
`readProducedObject`'s resolution logic against it. Probe trees live under
`/tmp/2qfnai-*` and are disposable.

No product code was modified by this review. As a reviewer-archetype run it
supplies no `commit_ack`.

## What the next producer must do

1. Fix F1 in `readset.go`: map dependency checkout reads back to their admitted
   package roots; fail closed on a `checkouts/` path with no matching admitted
   identity; keep the `inBuildTree` drop for genuinely derived build state. Add
   a test that a dependency target's observed read set contains its admitted
   source and headers.
2. Fix F2 in `build.go`: disambiguate produced objects using the
   target-path-relative source path. Add a conformance case with two
   same-base-name sources in different subdirectories of one Clang target, and
   fix `TestUndeclaredGeneratedObjectFailsClosed` so it fails for the reason it
   claims.
3. Add coverage for `ReadSetObserver.observe`/`harvestDependencyFiles`, and give
   `TestR01R05...` a fixture that actually has a mirrored dependency so the
   mirror mount, duplicate-receipt rejection, and `mirrors.json` kind mapping are
   exercised.
4. Clear the minor items above.
