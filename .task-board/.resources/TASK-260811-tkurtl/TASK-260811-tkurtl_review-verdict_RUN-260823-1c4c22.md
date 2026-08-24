# Reviewer verdict for TASK-260811-tkurtl — round 3

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260823-1c4c22` (Claude claude-opus-5), not goal-bound
  (`task-board spawn goal` reports `Active Goal: none`).
- Reviewed delivery: rework `RUN-260823-6849bf` against the three CONFIRMED
  findings (A, B, C) of `TASK-260811-tkurtl_review-verdict_RUN-260823-b80fbf.md`.
- Reviewed outcome: `TASK-260811-tkurtl_rework-outcome_RUN-260823-b80fbf.md`.
- No code, staging, commit, reset, or clean was performed. Four temporary probe
  test files were created inside `internal/swiftpminterop/`, executed, and
  removed; `git status --short` is 25 lines, byte-identical to the state
  received.

## Summary

All three round-2 findings reproduce as **fixed** against their original
probes. Finding C is fully closed and the fix is the right one — condition-
neutral rather than `boundary.Selected`, so capture stays byte-identical.
Finding A's transitive worklist is correct and reaches the private-header
escape, the three-hop chain, and the cycle. Finding B's stateful scanner
rejects all four compiler-verified round-2 spellings, mid-line `@import`, and
both preserved controls.

Acceptance is blocked by three CONFIRMED defects found by the adversarial
probing the round-3 brief required. All three are the same class as round-2
finding B — a header the pinned Apple Clang actually reads is invisible to the
declared closure, which in portable mode (`Reads.Mode = not-observed`) is the
entire header proof:

- **D** — translation phase 1 (trigraph replacement) is not performed, so
  `??=include </etc/passwd>` is not a directive to this scanner at all;
- **E** — a UTF-8 BOM or a non-ASCII space before `#` clears `atLineStart`, so
  the directive on that line is silently dropped;
- **F** — a module-map `header` reference that resolves outside the public-
  header root is admitted but never seeded into the include scan, so every
  directive it declares is invisible.

D, E, and F were each verified against the real compiler on this host, not
inferred.

## Per-item verdict

| # | Scope item (round-3 brief) | Verdict |
| ---: | --- | --- |
| 1 | Finding A closed — transitive include scan | **accepted** for its stated requirement; finding **F** is a newly found hole of the same class in the *seed* set |
| 2 | Finding B closed — compiler-faithful directive recognition | **changes requested** (findings **D** and **E**) |
| 3 | Finding C closed — conditional `Cxx` opt-in | **accepted** |
| 4 | No regression of previously accepted behavior | **accepted** |
| 5 | Evidence | **accepted** |

## Findings

### Finding D — CONFIRMED — trigraph `??=` makes an inclusion directive invisible

`spliceTranslationLines` (`internal/swiftpminterop/headers.go:256`) performs
line-ending normalization and phase-2 splicing but not phase-1 trigraph
replacement. `??=` is the trigraph for `#`, and trigraphs are still part of C —
they were removed from C++17 only. The pinned Apple Clang honours them in every
C mode SwiftPM can select, including the default with no `cLanguageStandard`.

Compiler evidence, `clang <std> -fsyntax-only -H` on
`??=include "secret.h"`:

| `-std` | reads `secret.h` |
| --- | --- |
| *(default)* | yes |
| `-std=gnu17` | yes |
| `-std=gnu11` | yes |
| `-std=c17` | yes |

Curator, executed probe (`Sources/CLib/lib.c` = `??=include </etc/passwd>`):

```
PROBE-B "trigraph hash" ACCEPTED, includes=[]
```

The closure succeeds, `/etc/passwd` appears in no include set and raises no
`swiftpm_header_input_undeclared`. This is not an unclassifiable `#`-line that
the new closed grammar rejects — there is no `#` byte in the file, so the
"anything else `#`-introduced is a rejection" backstop never engages.

A near miss worth recording as correct-by-accident: `#inc??/`⏎`lude <...>` —
`??/` is the trigraph for backslash, so the compiler splices it and reads the
header — **rejects** in Curator, because `#inc??/` reaches the closed
classifier as the unknown directive `inc`. The backstop covers that spelling;
it cannot cover `??=`.

Required: perform trigraph replacement in `spliceTranslationLines` before
splicing (at minimum `??=` and `??/`, and ideally the full set, since `??/`
composes with phase 2), or reject any source containing a `??` sequence this
grammar does not translate. Add H-family vectors for `??=include` and the
`??/` splice.

### Finding E — CONFIRMED — a BOM or non-ASCII space before `#` drops the directive

`directiveScanner.run` (`headers.go:273`) only keeps `atLineStart` across ASCII
horizontal space (`horizontalSpace`, `headers.go:500`). Any other byte falls to
`default:`, which sets `atLineStart = false`. A UTF-8 BOM (`EF BB BF`) and a
U+00A0 no-break space (`C2 A0`) are both bytes the compiler skips as
whitespace before a directive, and both clear the flag here.

Compiler evidence, `clang -fsyntax-only -H` (default std):

| Input | reads `secret.h` |
| --- | --- |
| `<BOM>#include "secret.h"` at file start | yes |
| `int a;`⏎`<BOM>#include "secret.h"` mid-file | yes |
| `<U+00A0>#include "secret.h"` | yes |
| the BOM and NBSP forms as `.mm` via `clang++ -x objective-c++` | yes |

Curator, executed probes:

```
PROBE-B "utf8 bom prefix"   ACCEPTED, includes=[]
PROBE-B "utf8 bom mid file" ACCEPTED, includes=[]
PROBE-B "nbsp prefix"       ACCEPTED, includes=[]
```

A BOM is the ordinary shape here, not an exotic one: any editor that writes
UTF-8-with-BOM produces the first form, and the escape is then admitted with no
diagnostic at all.

Required: treat the UTF-8 BOM and the non-ASCII white-space the target compiler
skips as line-start-preserving white space, and — consistent with the closed
grammar this round adopted — reject any byte sequence before a `#` that this
scanner cannot classify as white space, rather than silently demoting the line
to content. Add H-family vectors for the BOM (leading and mid-file) and NBSP.

### Finding F — CONFIRMED — a module-map header outside the public-header root is admitted but never scanned

`scanAndConfineIncludes` (`interop.go:537`) seeds the worklist from
`interop.Sources` (non-Swift) and `headerPaths(interop.Headers)`, and
`interop.Headers` is `inventoryHeaders(root, interop.PublicHeaderRoot)` —
the public-header root only. A custom module map may name a header **outside**
that root but inside the package, and `confineModuleMapReferences`
(`interop.go:476`) admits exactly that: it rejects only references that leave
the package or land in another package's tree. Such a header is a resolved
admitted reference that this stage never opens.

Failure scenario, executed:

```go
"Sources/CLib/include/module.modulemap": "module CLib {\n    header \"CLib.h\"\n    header \"../hidden.h\"\n    export *\n}\n",
"Sources/CLib/hidden.h":                 "#include </etc/passwd>\nint hidden(void);\n",
```

```
PROBE-MM ACCEPTED; scanned sources=map[Sources/CLib/lib.c:true]
```

Control, correctly rejecting: the same escape in `Sources/CLib/include/extra.h`
— inside the root, therefore inventoried and seeded — returns
`swiftpm_header_input_undeclared`.

The compiler really does read it. Building the module from that map on this
host:

```
clang -fsyntax-only -fmodules -fimplicit-module-maps \
      -fmodule-map-file=include/module.modulemap -I include -H use.c
In file included from include/../hidden.h:1:
/etc/passwd:1:1: error: expected identifier or '('
```

Required: seed the include worklist from every admitted resolution the module
map produced (`ModuleMapEvidence.ResolvedRefs` with
`Class == ResolvedAdmitted`), not only from the public-header inventory, so the
module's own header set is covered by the same fixpoint. Add an H12 subtest for
the out-of-root module-map header.

## Accepted items — evidence

**Item 1 — finding A closed, transitive include scan.** Reproduced with the
round-2 probe verbatim:

```
PROBE-A1 err=swiftpm_header_input_undeclared: source includes an absolute header path
```

The worklist is correctly shaped. `scanIncludes` opens each unit against
`state.packageRoots[current.pkg]` (`interop.go:564`) and a quoted include
resolves against `path.Dir(current.relative)` of the *including* file
(`interop.go:607`), not the consumer's directory — executed and observed:

```
PROBE-A-rec source=Sources/CLib/lib.c          spelling=detail/a.h
PROBE-A-rec source=Sources/CLib/detail/a.h     spelling=b.h
PROBE-A-rec source=Sources/CLib/detail/b.h     spelling=a.h
```

`detail/a.h` including `"b.h"` resolves only because the base is `a.h`'s own
directory, and the `a.h`↔`b.h` cycle terminates on `visited`. The delivered
`H12` covers the two-hop private-header escape, a three-hop escape, an escape
reached only through the umbrella header, an unopenable queued unit
(`swiftpm_source_inventory_drift` for a directory), and the recorded contained
closure — all five green in my rerun. A symlinked unit cannot enter the queue:
`realPathWithin` (`containment.go:172`) rejects it at resolution.

Cross-package hop: **read-verified, not executed.** `confineInclude` returns
`&sourceUnit{pkg: resolution.Package, ...}` and the worklist opens it under
that package's root, while `includeSearchRoots` keeps the *consumer's* search
path — which is the right semantics, since `-I` is fixed per translation unit,
not per included file. I could not execute it: every fixture in
`internal/swiftpminterop` is single-package (`fakeBroker` rejects pins), so the
package has no multi-package interop vector at all. That is a coverage gap
worth closing alongside the fixes above, not a defect I can demonstrate.

Not a defect, checked and dismissed: a `.c` file present in the target tree but
absent from the manifest's `Sources` is not scanned. In real operation
`enumerateTargetSources` (`swiftpmsource/executor_runtime.go:461`) walks the
tree and only honours an explicit `sources:` declaration — which SwiftPM itself
also honours — so the fixture's fake evaluator, not the adapter, produced that
shape.

Over-approximation, correct: an escape inside `#if 0` still rejects. The
scanner does not evaluate conditionals, which errs closed.

**Item 3 — finding C closed, conditional `Cxx` opt-in.** Reproduced with the
round-2 probe verbatim (both the `CxxLib` dependency and the
`.interoperabilityMode(.Cxx)` opt-in under `.when(platforms: [.macOS])`):

```
PROBE-C darwin err=<nil> | linux err=<nil>
PROBE-C capture darwin=sha256:3730e4dd… linux=sha256:3730e4dd… equal=true
PROBE-C evidence darwin=sha256:91252f6c… linux=sha256:433b6536… equal=false
PROBE-C mode d="cxx_interop" l="cxx_interop" selected d=true l=false
PROBE-C declared d=true l=true | mode d=true l=false
```

Both destinations close, the capture is byte-identical, the boundary mode is
the same on both, and only the selection verdict and the evidence digest
differ. The gate is genuinely condition-neutral: `cxxInteropDeclared`
(`platform.go:150`) ignores `setting.Condition` entirely, and `boundaries.go:78`
reads `consumer.CxxInteropDeclared` while `CxxInteropMode` keeps the exact
destination verdict for the evidence record. The producer's reason for choosing
the neutral gate over `boundary.Selected` is correct and worth keeping:
`Mode`, `ABI`, `Runtime`, `InterfaceContract`, and `CallingConvention` all live
in `InteropBoundaryPayload`, i.e. capture, so a `Selected`-gated classification
would have published `cxx_interop` on Darwin and `c_abi` on Linux for one
admitted closure — trading a false rejection for a CGP05 violation.

Controls, all reproduced:

```
PROBE-C-control          code="swiftpm_target_platform_unsupported"   (unconditional dependency + conditional opt-in on Linux)
PROBE-C-uncond           mode="cxx_interop" selected=true             (S05 unchanged)
PROBE-C-uncond-missing   code="closure_interop_undeclared"            (S06 unchanged)
```

**Item 4 — no regression.** Selection neutrality with the CGP05 conditional
vector reproduces at the same digest round 2 recorded:

```
PROBE-CGP05 capture d=sha256:b1c1e85c… l=sha256:b1c1e85c… equal=true
PROBE-CGP05 evidence equal=false | boundaries d=1 l=1 | selected d=true l=false
```

Round-2 finding B's four spellings and both preserved controls all still
reject, executed:

| Spelling | Code |
| --- | --- |
| spliced keyword `#inc\`⏎`lude` | `swiftpm_header_input_undeclared` |
| comment prefix `/* */ #include` | `swiftpm_header_input_undeclared` |
| form-feed prefix | `swiftpm_header_input_undeclared` |
| digraph `%:include` | `swiftpm_header_input_undeclared` |
| mid-line `@import Secret;` | `swiftpm_header_input_undeclared` |
| control — `#include \`⏎`</etc/passwd>` | `swiftpm_header_input_undeclared` |
| control — CRLF literal | `swiftpm_header_input_undeclared` |
| control — `#curator_secret` | `swiftpm_header_input_undeclared` |

Additional adversarial spellings that the scanner handles correctly (each
verified to read the header under the real compiler first): a comment between
`#` and the keyword (`# /*c*/ include`), a splice inside the operand
(`</etc/pas\`⏎`swd>`), and a splice inside the digraph (`%\`⏎`:include`) — all
three reject.

`closuregraph` canonicalization is untouched: `internal/closuregraph/testdata`
is unmodified and the Ruby oracle reports `canonical_goldens=pass
labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2` and
`canonical_references=pass cgp05_capture_reused=true
explicit_target_bindings=2 cgp10_all_refs_resolve=true`. `publicHeadersPath`
handling and the out-of-root module-map guard (`H10`, `H11`) are green.
Seam and guard discipline is intact: `grep -l exec.Command` over
`internal/closureexec`, `internal/swiftpmsource`, `internal/swiftpminterop`
production files returns exactly `closureexec/acquisition.go` and
`closureexec/portable_runner.go`, the `guard_test.go` allowlist is unchanged,
and no interop production file imports `os/exec`. Portable read-set honesty,
typed boundary gates, and `S02`-`S09`/`H01`-`H08`/`P01`-`P09`/`CGN03`/`CGN09`/
`CGN15` are all present and green. Scope hygiene holds: no Kotlin or Gradle
reference, no non-`.go` file in the package, and no creep into
`TASK-260811-2qfnai` or `TASK-260811-x611eq`.

Round-3 confinement independently confirmed: `git diff --stat` shows no new
modification outside `internal/swiftpminterop/`, and every tracked changed file
(`closuregraph/edge.go` 02:24, `swiftpmsource/manifest.go` and
`executor_runtime.go` 02:11, `manager.go` 01:20) predates the round-3 edits in
`internal/swiftpminterop/` (03:10-03:13).

**Item 5 — evidence.** Every gate I reran reproduced:

| Gate | Result |
| --- | --- |
| `go test -count=1 -cover ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/` | exit 0; interop coverage **86.6%** — matches the claim |
| Test matrix, `internal/swiftpminterop` | **66** top-level, **138** including subtests — matches the claim |
| `go test -race -count=1` over the three changed-or-adjacent packages | exit 0 |
| `golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/... ./internal/closuregraph/...` | `0 issues.` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | both lines pass, 53 labeled records |
| `task-board validate` | `Board is valid. No issues found.` |
| Orchestrator `TASK-260811-tkurtl_full-go-03.log` | SHA-256 **6bb1b70aa4cd92511d6e47a9982f3de0c4d6e7b2a6064a1c37af8a6e2c60420d** matches the recorded description; `EXIT:0`, **52 ok**, 0 FAIL, `cmd/curator 491.334s` |

Accepted rather than rerun: the monolithic full suite and the 51-package
non-`cmd/curator` slice, on the hash-verified orchestrator log above — my four
focused packages plus that log's remaining entries account for all 52 packages.
The producer's own round-3 focused/race/lint/golden logs were independently
reproduced rather than accepted. The producer's statement that
`go list -deps ./cmd/curator` does not reach `swiftpminterop` is consistent
with the round-3 delta being confined to that package.

## Routing

`TASK-260811-tkurtl` -> `to-dev`. Round-2 findings A, B, and C are closed and
should not be reopened; keep the transitive worklist, the stateful directive
scanner and its closed classification, `literalIncludeOperand`, `H09`, `H12`,
`H13`, the `cxxInteropDeclared`/`cxxInteropSelected` split, and both new S05
vectors. Findings D, E, and F each have an executed reproduction, compiler
evidence, and a concrete required change, all inside
`internal/swiftpminterop/`. The missing multi-package interop fixture is a
coverage gap to close in the same round.

As a reviewer-archetype run this supplies no `commit_ack`.
