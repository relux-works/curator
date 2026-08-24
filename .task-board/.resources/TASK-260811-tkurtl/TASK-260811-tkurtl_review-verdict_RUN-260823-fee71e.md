# Reviewer verdict for TASK-260811-tkurtl

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260823-fee71e` (Claude claude-opus-5), not goal-bound
- Reviewed delivery: producer `RUN-260823-5c7b26`, uncommitted `internal/swiftpminterop`
  plus three additive `internal/swiftpmsource` accessors, on clean `f8b7cc7`/`40142af`
- Reviewed outcome: `TASK-260811-tkurtl_developer-outcome.md`
- No code, staging, commit, reset, or clean was performed. The working tree is
  byte-identical to the state received (three temporary probe tests were added
  under the package, executed, and removed; `git status --short` is unchanged).

## Summary

The architecture, seam discipline, module-map grammar, portable-honesty
contract, and validation evidence all hold up and reproduce exactly. Three
confirmed defects block acceptance, each against an explicitly named mandatory
scope item. All three are demonstrated by executable probes, not inference.

## Per-item verdict

| # | Scope item | Verdict |
| ---: | --- | --- |
| 1 | Interop boundary correctness | **accepted** |
| 2 | Module map handling | **changes requested** (parser accepted; include scanner fails open — finding 2) |
| 3 | Header read closure | **changes requested** (findings 2 and 3) |
| 4 | Seam and guard discipline | **accepted** |
| 5 | Canonical graph integrity | **changes requested** (finding 1) |
| 6 | Portable honesty | **accepted** |
| 7 | Evidence | **accepted** |
| 8 | Scope hygiene | **accepted** |

## Findings

### Finding 1 — CONFIRMED — interop capture is not selection-neutral (scope item 5, CGP05)

`internal/swiftpminterop/interop.go:167` (`classifyTargets`) walks only targets in
`state.selected`, and `internal/swiftpminterop/interop.go:398`
(`directTargetDependencies`) filters declared target and product edges through
`conditionSelected(dependency.Condition, state.markers)`. Both the per-target
interop capture nodes (`swiftpm.interop.compile.*`, `.sources.*`, `.headers.*`,
`.objects.*`) and the `interop_boundary` capture nodes are therefore emitted
only for the destination-selected subset. A pruned or selected conditional
declaration is a selection/active fact, not a capture fact.

Failure scenario, executed: take the fixture package and put
`.when(platforms: [.macOS])` on `App`'s dependency on `CLib`, then close the
same admitted closure against the accepted Darwin and Linux destinations.

```
UPSTREAM capture digests: darwin=sha256:3560a162... linux=sha256:3560a162... equal=true
INTEROP  capture digests: darwin=sha256:b1316468... linux=sha256:55c2a21a... equal=false
darwin boundaries=1 targets=2 | linux boundaries=0 targets=1
```

The upstream `swiftpmsource` capture is correctly selection-neutral across the
two destinations; this stage's republished capture is not. That directly
contradicts design decision 5 in the producer outcome ("Darwin and Linux produce
byte-identical interop capture and different bindings") and the accepted
architecture's rule that `CaptureGraph` "excludes ... selected/pruned results".

`TestCGP05InteropCaptureIsSelectionNeutralAcrossDestinations` cannot observe this
because its fixture has no conditional edge, so the selected set is identical on
both branches. `TestPrunedConditionalDependencyDeclaresNoBoundary` and
`TestConditionalProductDependencyIsProjectedExactlyOnce` both assert the current,
selection-scoped behavior and would need to move their assertions to the active
projection.

Required: emit the interop capture records for every declared C-family target and
boundary regardless of condition, carry the condition on the capture edge, and
let `ProjectActive` record the selected/pruned verdict. Extend the CGP05 vector
with a conditional-dependency branch so the invariant is actually exercised.

Impact is conservative in direction (a capture identity becomes over-specific
rather than aliasing two destinations), so this is a stated-invariant and
cache-reuse defect rather than an admission hole — but item 5 is a named
acceptance gate and the delivery claims it is satisfied.

### Finding 2 — CONFIRMED — the include scanner silently drops non-literal `#include` (scope items 2, 3)

`internal/swiftpminterop/headers.go:24`:

```go
var includePattern = regexp.MustCompile(`(?m)^[ \t]*#[ \t]*(include|import|include_next)[ \t]*(?:"([^"\n]*)"|<([^>\n]*)>)`)
```

A directive whose operand is not a literal — the macro-computed
`#include MACRO` form — matches nothing and produces no reference and no
diagnostic. The module-map parser is explicitly fail-closed ("anything it cannot
resolve exactly is a rejection, never a silent skip"); the include scanner, which
is the *only* header-closure proof in the mode that exists today, is not.

Failure scenario, executed — `Sources/CLib/lib.c`:

```c
#include "CLib.h"
#define CURATOR_SECRET </etc/passwd>
#include CURATOR_SECRET
int value(void) { return 1; }
```

The closure succeeds. The recorded include set for `root:CLib` is exactly
`[{Spelling: CLib.h}]`; the computed include is absent, `Reads.Mode` is
`not-observed`, and no `swiftpm_header_input_undeclared` is raised.

This matters now specifically because no `ReadSetProvider` exists yet
(`TASK-260811-2qfnai` supplies it), so `not-observed` is the only reachable mode
and the declared static closure is the entire proof. `AssuranceVerified` does
fail closed, which is correct and is why this is a gap rather than a false
assurance claim — but the delivered contract states the declared closure "remains
the proof" in portable mode, and it silently is not.

Required: reject any `#`-directive line whose `include`/`import`/`include_next`
operand this grammar cannot resolve to an exact literal, with
`swiftpm_header_input_undeclared` (or a scanner-grammar escape), matching the
module-map parser's discipline. Add an `H`-family vector for the computed-include
form.

### Finding 3 — CONFIRMED — a non-default public-header layout is silently defaulted, not rejected (scope item 3)

`internal/swiftpminterop/interop.go:236` calls
`publicHeaderRoot(target.Path, "")`, and `headers.go:60` substitutes `include`
whenever the declared value is empty. The second argument is hardcoded empty at
the only call site because `swiftpmsource.Target` (`internal/swiftpmsource/types.go:69`)
carries no `publicHeadersPath` field at all — SwiftPM's declaration is dropped
before it reaches this stage.

The consequence is that the module map and headers actually governing the target
are never located. `moduleMapEvidence` looks for a custom map only inside the
assumed `include` directory, so a package whose real public-header directory is
elsewhere gets a *generated* map for the wrong directory instead.

Failure scenario, executed — add `Sources/CLib/pub/module.modulemap` containing
`module CLib { header "/etc/passwd" }` alongside `Sources/CLib/pub/CLib.h`:

```
PROBE3 parsed module map: Sources/CLib/include/module.modulemap generated=true
```

The escaping map is never parsed, never confined, and never digested. This is the
H03 vector the task exists to stop, defeated by a layout the model cannot
represent. `TestH03AbsoluteModuleMapHeaderIsRejected` and
`TestH03EscapingModuleMapHeaderIsRejected` both place the map under the assumed
`include` directory, so neither covers it.

Required: carry `publicHeadersPath` through the additive `swiftpmsource.Target`
contract and honor it here; and, independently, fail closed
(`swiftpm_target_platform_unsupported` or `closure_graph_incomplete`) on any
target whose public-header layout this profile cannot represent exactly, rather
than substituting a default. A conservative interim guard is to reject a C-family
target that contains a `module.modulemap` anywhere in its admitted tree outside
the resolved public-header root.

## Accepted items — evidence

**Item 1, interop boundary correctness.** `boundaries.go:56-102` derives typed
`c_abi`/`cxx_interop`/`objc_runtime` boundaries from declared evidence only.
Direct Swift/C++ requires both `Profile.CxxInterop` and an explicit
`interoperabilityMode(.Cxx)` on the consuming Swift target; a C++ public
interface without the opt-in is `closure_interop_undeclared` rather than an
implicit C edge, which is the correct reading of SwiftPM's non-propagation.
Restricted languages gate on `PlatformProfile` before any admitted byte is
classified (`platform.go:110`, `interop.go:175`). `classifyTarget`
(`language.go:80`) rejects a mixed Swift + C-family target with
`swiftpm_mixed_language_target_unsupported` before any header, module, or
boundary analysis. `S02`-`S09` all present and green.

**Item 2, module map handling (parser half).** `clang-modulemap-lexer-v1` covers
`module`/`explicit`/`framework`/`extern`, attribute lists, `header`/`textual`/
`private`/`exclude`, umbrella headers and directories, header attribute blocks,
`link`, `link framework`, `requires` with negation, `export`, `export_as`, `use`,
and `config_macros`. Every unrecognized character, qualifier, member, or
top-level form reaches an explicit `modulemapSyntax` rejection —
I traced each `default:` arm and each `skip*` early return; the two permissive
returns in `skipIdentifierList` hand the token back to `parseModuleBody`, whose
`default:` rejects it. `resolveReference` (`containment.go:118`) rejects absolute
and control-character spellings before touching the filesystem, so a missing file
cannot be mistaken for a contained one. `GenerateModuleMap` reproduces SwiftPM's
umbrella-header / single-header / umbrella-directory rules and is parsed and
confined on the same path as a custom map.

**Item 3, header read closure (the parts that hold).** `roots.resolve` classifies
into exactly `admitted_source`, `selected_binding`, or `undeclared`;
`realPathWithin` walks every component and rejects symlinks and special nodes;
`presentNode` requires a selected-external entry to actually exist, so a path
spelling that merely prefixes an SDK root grants nothing. System libraries admit
only from a Curator-selected `ExternalComponent` with its module map inside that
component's roots (`artifact_toolchain_untrusted` otherwise), and `confineLinks`
checks every `link`/`link framework` name against declared component libraries
and frameworks. `H01`, `H02`, `H04`-`H08` green.

**Item 4, seam and guard discipline.** Verified independently:
`grep -l exec.Command` over `internal/closureexec`, `internal/swiftpmsource`,
`internal/swiftpminterop` production files returns exactly
`closureexec/acquisition.go` and `closureexec/portable_runner.go`; no production
file in `internal/swiftpminterop` imports `os/exec`; the guard allowlist is
unchanged and matches `internal/rustsource/build_test.go:53` verbatim. I checked
the basename-keyed allowlist for collisions — no `acquisition.go` or
`portable_runner.go` exists in either newly globbed directory, so the allowlist
resolves to exactly the two intended shared seams. `internal/swiftpmsource`
changes are three additive accessors (`ProtectedRoot`, `Destination`,
`SelectionToolchain`) with no behavior, record, or digest change; I read the full
diff.

**Item 5, the parts that hold.** Platform, toolchain, SDK, sysroot, and system
component nodes and every `targets` / `uses_tool` / toolchain-scoped `requires` /
`provides_interop`-from-system edge are emitted into the binding overlay only
(`graph.go:29-45`, `graph.go:212-233`). C4 chains from the accepted source-closure
C4 and binds all four exact identities. `CGN03`, `CGN09`, `CGN15` re-project the
published records after adversarial mutation and reject canonically. Finding 1 is
scoped to *which* records are emitted, not to where selection facts live.

**Item 6, portable honesty.** `verifyReads` (`boundaries.go:151`) records
`not-observed` when there is no provider or any target reports `Observed: false`,
verifies every observed read that is supplied regardless, rejects observed
evidence without its issued receipt as `closure_derivation_unauthorized`, and
fails closed under `AssuranceVerified` without a full observed read set. Honest.

**Item 7, evidence.** Every gate reran independently and reproduced:

| Gate | Result |
| --- | --- |
| `go test -count=1 -cover ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/` | exit 0; interop coverage **86.1%**, matching the claim exactly |
| Test matrix | **58** top-level tests, **96** including subtests — both claims exact |
| `golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/...` (v2.12.2) | `0 issues.` |
| `go test -timeout 9m -count=1 $(go list ./... \| grep -v cmd/curator)` | exit 0, **51 ok**, 0 FAIL |
| `go build ./...`, `go vet ./...`, `gofmt -l` | clean |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2`; `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Orchestrator `TASK-260811-tkurtl_full-go-01.log` | SHA-256 **matches** `6c7a9c26...127e4`; `EXIT:0`, **52 ok**, 0 FAIL, `cmd/curator 379.870s` |

Accepted rather than rerun: the monolithic full suite, on the hash-verified
orchestrator log above. My own 51-package run plus that log's `cmd/curator` entry
account for all 52 packages. Not rerun: the producer's `-race` set and the
bounded `cmd/curator` subset, both subsumed by the hash-bound full run and the
suite I did rerun.

**Item 8, scope hygiene.** No Kotlin or Gradle reference anywhere in the package;
no non-`.go` file in `internal/swiftpminterop`, so no vendored compiled payload;
no reference to `TASK-260811-2qfnai` or `TASK-260811-x611eq` work in production
code. `P01`-`P09` all present, with binary, plugin, macro, renamed-compiled-payload
and generated-lineage vectors asserting the shared codes.

## Routing

`TASK-260811-tkurtl` -> `to-dev`. Findings 1-3 each have a concrete reproduction
and a concrete required change. As a reviewer-archetype run this supplies no
`commit_ack`.
