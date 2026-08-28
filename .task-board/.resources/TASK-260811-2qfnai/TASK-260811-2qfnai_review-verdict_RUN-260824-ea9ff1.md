# Independent re-review verdict — TASK-260811-2qfnai F1/F2 rework

Reviewer run: `RUN-260824-ea9ff1` (Claude claude-opus-5, independent re-review)
Rework run under review: `RUN-260824-24b2e4`
Prior verdict re-reviewed: `TASK-260811-2qfnai_review-verdict_RUN-260824-145133.md`
(changes requested → `to-dev`; 8 of 10 scope items accepted)
Run goal binding: `task-board spawn goal "$TASK_BOARD_RUN_ID"` reported
`Active Goal: none (run is not goal-bound)`.

**Verdict: accepted → `done`.**

F1 is closed and closed transitively, F2 is closed, and the two tests that
would have caught each defect genuinely bite — I proved that by mutation, not
by reading the producer's claim. The R01/R05 vacuity is repaired, all five nits
are cleared, and nothing in the eight previously accepted scope items
regressed.

## Per-item verdicts

| # | Mandatory scope item | Verdict |
| ---: | --- | --- |
| 1 | F1 closed — the keystone | accepted |
| 2 | F1 test gap closed — the test bites | accepted |
| 3 | F2 closed, `TestUndeclaredGeneratedObjectFailsClosed` repointed | accepted |
| 4 | R01/R05 vacuity repaired | accepted |
| 5 | Nits addressed | accepted |
| 6 | No regression of the 8 accepted items | accepted |
| 7 | Evidence | accepted |

---

## 1. F1 — closed, and closed for every package

`readset.go:336 mapObservedRead` now takes the root identity **and the full**
`admittedPackageRoots` map (the map the previous reviewer flagged as computed
and thrown away). The branch structure is exactly the required one:

- a read outside the work root is returned verbatim, `inBuildTree=false` — H03
  behaviour is unchanged and an escaping absolute header still fails closed
  downstream;
- a read inside the work root but not under `.curator/` is rebound to the root
  package's admitted protected tree;
- a read under `.curator/scratch/checkouts/<identity>/…` is rebound to **that
  dependency's** admitted protected root via `joinAdmitted`;
- a `checkouts/` read whose first component matches no admitted identity, and a
  bare `checkouts` read, **fail closed** with `swiftpm_header_input_undeclared`
  (`joinAdmitted` returns the failure; nothing is dropped);
- only a genuinely derived read below `.curator/` — module cache, the module
  maps SwiftPM generates below the scratch tree — is reported as build state
  and dropped, and the comment at `readset.go:289` now says exactly that.

`harvestDependencyFiles` also fails closed when the capture admitted no package
at all, instead of indexing `Packages[0]` blind.

I verified the keystone claim holds *transitively*, not just that the rewrite
happens:

- `swiftpmsource/capture.go:134` builds `packages := []PackageEvidence{rootPackage}`
  and never sorts it, so `Packages[0]` really is the root package — the
  root-identity assumption is sound, not accidental (with packages `dep` and
  `root`, a sorted slice would have silently mis-bound them).
- `swiftpminterop/interop.go:137` calls `roots.addAdmitted` for **every**
  capture package, so a rewritten dependency path resolves as
  `ResolvedAdmitted` for that dependency in `containment.go:84 resolve`.
- `containment.go:172 realPathWithin` walks every component with `Lstat` and
  rejects a missing node or a symlink, so a `checkouts/` read of a file that is
  *not* in the admitted tree — a smuggled file materialized into the checkout —
  is rewritten and then fails closed as undeclared. The fix is rewrite *plus*
  existence proof, which is the right shape.

Reviewer's reproduction scenario, in unit form: a `file://`-shaped source-control
dependency whose `.d` names `checkouts/<id>/Sources/CDep/dep.c` and
`checkouts/<id>/Sources/CDep/include/CDep.h` now produces an observed read set
containing both, rebound to the dependency's admitted root — see item 2.

## 2. F1 test gap — closed, and the test bites

This was the decisive check. I copied the repository to a scratch tree
(`/tmp`, since removed), reverted **F1's checkouts branch** (forcing every
`.curator/` read back to `inBuildTree`) and **F2's narrowing** (back to the
dead `strings.HasSuffix(candidate, slot.Source+".o")` form), and ran the
package:

```
--- FAIL: TestS03SameBaseNameSourcesInDifferentDirectoriesResolve
--- FAIL: TestRealSwiftPMBuildMatchesThePlannedLayout
--- FAIL: TestObservedReadMappingSeparatesBuildTreeFromAdmittedSource
--- FAIL: TestHarvestedReadSetCoversEveryAdmittedPackage
--- FAIL: TestHarvestedReadSetFailsClosedOnUnadmittedCheckout
```

All four named tests fail under the exact original defect, plus the real-
toolchain vector. The gap that hid F1 is genuinely closed:

- `TestHarvestedReadSetCoversEveryAdmittedPackage` drives the real
  `harvestDependencyFiles` over a **multi-package** capture with a pinned
  source-control dependency and asserts the dependency target's read set is
  exactly `{dep.c, CDep.h}` rebound to the dependency's admitted root **plus**
  the SDK header verbatim, while the generated module map below the scratch
  tree drops and the root target's read rebinds to the root's admitted tree. If
  `dep`'s reads were still discarded the expected set would not match.
- `TestHarvestedReadSetFailsClosedOnUnadmittedCheckout` asserts the
  no-matching-identity `checkouts/` read is `swiftpm_header_input_undeclared`
  rather than dropped.
- `TestObservedReadMappingSeparatesBuildTreeFromAdmittedSource` covers all six
  branches including bare `checkouts` and the external verbatim read.

## 3. F2 — closed, and the undeclared-object test no longer passes for the wrong reason

`resolveProducedObject` now narrows on `targetRelativeSource(slot)` compared for
**equality**, where the target-relative path comes from a single shared helper:
`swiftpmsource.Target.SourceRoot()` (`types.go:87`, convention default applied
exactly once and reused by `enumerateTargetSources`) →
`swiftpminterop.TargetInterop.SourceRoot` (`interop.go:244`) →
`swiftpmbuild.ObjectSlot.SourceRoot` (`plan.go:313`). Manifest normalization and
downstream reconciliation cannot disagree.

- `TestS03SameBaseNameSourcesInDifferentDirectoriesResolve` builds one Clang
  target with `Sources/CLib/a/x.c` + `Sources/CLib/b/x.c`, asserts four declared
  slots, both objects resolving to **distinct published bytes** under their own
  logical paths, five observations and a five-entry write set. It fails under
  the reverted narrowing.
- `TestRealSwiftPMBuildMatchesThePlannedLayout` now carries
  `Sources/CLib/a/shared.c` + `Sources/CLib/b/shared.c` against the real Apple
  Swift 6.3.2 toolchain and asserts the resolution is bijective **and exhausts**
  the produced set, for the Clang target and the multi-source Swift target. It
  also fails under the reverted narrowing, so the real-toolchain layout claim is
  load-bearing rather than decorative.
- `TestUndeclaredGeneratedObjectFailsClosed` is repointed to a genuinely
  undeclared `CLib.build/generated/smuggled.c.o` and, crucially, asserts inline
  that the declared slot *still resolves* to `lib.c.o` — so it can no longer
  pass because both candidates were eliminated. I confirmed it now fails for the
  reason it claims: neutering `requireNoUndeclaredObject` in the scratch tree
  produced `error code = "", want "artifact_local_output_unreceipted"`.

The new `requireNoUndeclaredObject` exhaustion check is the right complement:
now that resolution is exact rather than accidentally ambiguous, an object no
declared slot claims is the thing that must fail closed, and duplicate claims
are rejected too.

## 4. R01/R05 vacuity — repaired

`fixture.addSourceControlDependency` (`fixture_test.go:98`) now binds a real
admitted dependency tree, a captured local mirror, a one-pin frozen
`Package.resolved`, and the manifest edge the selected product reaches. The
fixture is no longer `"pins":[]`.

I confirmed the repaired test is non-vacuous by mutation: dropping the mirror
`InputMount` from `build.go:168` produces

```
read set = []string{"inputs/build-root"}, want the admitted build root plus "inputs/mirrors/dep"
```

`TestGeneratedMirrorConfigurationPreservesSourceControlKind` decodes the real
generated `.curator/config/mirrors.json` and asserts
`https://example.invalid/dep` → `file://<exec-root>/inputs/mirrors/dep`, i.e.
the `remoteSourceControl` kind mapping end to end.
`TestDuplicateMirrorReceiptFailsClosed` asserts `swiftpm_dependency_mirror_missing`
with zero process starts and no publication.

## 5. Nits — all five addressed, one of them answered rather than papered over

- `types.go:68` — `ObjectSlot` doc now says one slot per **source**; the new
  `SourceRoot` field is documented.
- `plan.go` — the `AssurancePortable` keepalive and the `closureexec` import are
  gone (`grep closureexec plan.go` is empty).
- `plan.go:335` — the link action `DomainID` key is now `target_triple`.
- `binding.go:65` — a `SlotLinker` entry in `Config.Slots` is now **rejected**
  with `closure_graph_reference_invalid`.
- `binding.go:165` — the edge-activation question is answered with a stated
  invariant, and I verified the invariant against the shared contract rather
  than taking the comment at face value: `closuregraph/projection.go:139` emits
  activations for `conditionalIDs` only, `validation.go:395` rejects an
  activation record on an unconditional edge, and `validation.go:1136`
  `selectedEdges` uses the identical `condition() == nil || selected` rule. The
  build stage now matches shared semantics exactly.

## 6. No regression

- Binding overlay/identity, single-resolution, C4→C4′→C5 chaining, offline
  execution flags and isolation, the fail-closed matrix, publication zero-start
  reuse, seam/guard discipline and scope hygiene are all still proved by their
  own tests, and the whole repository suite is green.
- Portable mode is untouched: `TestPortableObserverReportsNotObserved` still
  asserts not-observed with zero process starts, and tkurtl's reject-by-default
  portable verdict is preserved.
- Verified mode still fails closed without a compatible provider:
  `TestVerifiedBuildRequiresObservedReadSet` asserts
  `swiftpm_header_input_undeclared`, and `closureexec.NewOSBoundary` still
  returns `closure_derivation_unauthorized` on Darwin and every other platform.
- Scope hygiene holds: no `os/exec` import in any `swiftpmbuild`/`swiftpminterop`
  production file (only the real-toolchain integration test, which the guard
  allowlist already covers), and zero non-Go files under either package.

## 7. Evidence — what I reran versus what I accepted

Rerun by me, directly, exit codes are real:

| Gate | Command | Exit |
| --- | --- | ---: |
| Focused suite | `go test -count=1 ./internal/swiftpmbuild/ ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closureexec/ ./internal/closuregraph/` | 0 |
| Repository suite excluding `cmd/curator`, part 1/2 (40 packages) | `go test -timeout 9m -count=1 $(head -40 <pkg list>)` | 0, 39 `ok`, 0 `FAIL` |
| Repository suite excluding `cmd/curator`, part 2/2 (13 packages) | `go test -timeout 9m -count=1 $(tail -13 <pkg list>)` | 0, 13 `ok`, 0 `FAIL` |
| Race | `go test -race -count=1 ./internal/swiftpmbuild/ ./internal/swiftpminterop/ ./internal/swiftpmsource/` | 0 |
| Real toolchain vector | `go test -count=1 -v -run TestRealSwiftPMBuildMatchesThePlannedLayout ./internal/swiftpmbuild/` | 0, `PASS` |
| Lint | `golangci-lint run ./...` | 0, `0 issues.` |
| Format | `gofmt -l ./cmd ./internal` | 0, empty |
| Vet | `go vet ./...` | 0 |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0, `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2` / `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Board | `task-board --no-update-check validate` | 0, `Board is valid. No issues found.` |

The suite was split into two bounded halves because the single-call cap is ten
minutes; 39 + 13 `ok` reproduces the producer's 52-package claim exactly (the
53rd listed package has no test files).

Accepted from the orchestrator-attached monolithic log rather than rerun, because
`cmd/curator` alone takes ~420 s and the monolith exceeds the cap:

- `TASK-260811-2qfnai_full-go-02.log`, SHA-256 verified by me as
  `d5e2343dd7e5251c3678c9d515b2a19ccb028f71c467f1812437698712dbd472`
  (exactly the hash in the review brief), first line
  `ok github.com/relux-works/curator/cmd/curator 419.129s`, 53 `ok` packages,
  zero `FAIL`, trailing `EXIT:0`.

Mutation probes were run in a scratch copy of the repository under `/tmp`
(since removed). The working tree was never modified, staged, committed, reset,
or cleaned; `git status` outside `.task-board` shows only the same nine modified
tracked files and the two untracked adapter packages that were there when this
review started.

## Non-blocking observations for the integration leaf (`TASK-260811-x611eq`)

None of these blocks acceptance; all are hardening, and none is reachable today
because verified assurance has no provider.

1. **Coverage is still asserted nowhere.** `swiftpminterop/boundaries.go:154
   verifyReads` asserts *containment* — every observed read resolves to
   something declared — and never asserts that a target's declared sources
   actually appear in its observed read set. A `.build` directory with zero
   `.d` files yields `observed[target] = []`, which is `present`, so
   `ObserveReads` returns `Observed: true` with an empty read set and
   `Reads.Mode` is still `observed`. F1's discard was the mechanism that made
   this reachable and it is gone, and inside `swiftpm-source-v1` a package
   cannot suppress `-emit-dependencies` (`unsafeFlags` is rejected), so this is
   defence in depth rather than a live hole. Worth an assertion that each
   selected target's declared sources appear in its observed reads.
2. **The checkout directory name is assumed equal to the package identity.**
   That holds today because Curator mirrors to `inputs/mirrors/<identity>` and
   SwiftPM names the checkout from the mirror URL basename. SwiftPM's
   collision-suffixed naming would fail closed rather than mis-admit, which is
   the safe direction, but a fixture pinning this assumption is cheap.
3. `resolveProducedObject` only narrows when more than one base-name candidate
   matches; a single match is accepted without checking the target-relative
   path. Safe, because `requireNoUndeclaredObject` plus duplicate-claim
   rejection make the mapping bijective across the target — but comparing the
   target-relative path first would be simpler than base name plus fallback.
4. `requireNoUndeclaredObject` only inspects targets that have at least one
   declared slot, since `produced` is populated lazily by `targetObjects`. A
   selected compile target always has at least one source and therefore at
   least one slot, so this is currently unreachable.

No product code was modified by this review. As a reviewer-archetype run it
supplies no `commit_ack`.
