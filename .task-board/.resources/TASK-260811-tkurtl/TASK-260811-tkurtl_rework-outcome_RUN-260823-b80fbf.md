# Rework outcome — TASK-260811-tkurtl, round 3

Producer run: implementer/developer (Claude claude-opus-5), not goal-bound.

Input: `TASK-260811-tkurtl_review-verdict_RUN-260823-b80fbf.md` — three CONFIRMED
findings (A, B, C). Scope was exactly those three; nothing else was touched.
Every change in this round is inside `internal/swiftpminterop/`. `git diff
--stat` shows no new modification to `internal/closuregraph`,
`internal/swiftpmsource`, or any other package: the round-1/round-2 changes
there are the ones both verdicts already accepted, unchanged.

## Finding A — the include scan is now transitive

**Fix.** `inspectTargets` no longer scans `interop.Sources ++
headerPaths(interop.Headers)` in its first pass. The second pass is now
`scanAndConfineIncludes` (`interop.go`), a worklist over `sourceUnit{pkg,
relative}`:

- it seeds from the target's declared non-Swift sources and its public-header
  inventory;
- `confineInclude` returns the `Resolution` of every reference it admits, and
  the worklist enqueues that resolved file, so the closure of admitted
  references is scanned to fixpoint (`visited` terminates include-guard cycles);
- a reference that resolves into another admitted package is scanned against
  **that** package's root, and a quoted include is resolved relative to the
  directory of the file currently being scanned, not the consumer's;
- a queued unit that cannot be opened and scanned fails closed —
  `scanIncludes` returns `swiftpm_source_inventory_drift` for an unreadable
  target (a directory, a special node, a vanished file);
- `resolveIncludeSpelling` now returns `(Resolution, bool)` instead of a bare
  bool so the admitted resolution is available to requeue.

Swift sources are deliberately excluded from the seed set: Swift performs no
textual inclusion, so applying the C grammar there would reject ordinary Swift
`#` syntax (`#if os(macOS)`, `#Preview`) while proving nothing. Every
non-Swift unit — headers, `.c`/`.cpp`/`.m`/`.mm`, and any extensionless or
oddly-suffixed file a directive reaches — is scanned, so
`#include "payload.txt"` cannot hide directives behind a suffix.

**New vectors** — `TestH12TransitiveIncludeClosureIsScanned` (5 subtests):

| Subtest | Assertion |
| --- | --- |
| private header escape is reached | the reviewer's probe verbatim: `lib.c` includes `private.h`, which includes `</etc/passwd>` → `swiftpm_header_input_undeclared` |
| escape three levels deep is reached | `lib.c` → `detail/a.h` → `detail/b.h` → `../../../../outside.h` → rejected |
| public header of a dependency is reached | an escape declared in a second public header reached only through the umbrella header → rejected |
| admitted reference that cannot be opened fails closed | `#include "detail"` resolving to a directory → `swiftpm_source_inventory_drift` |
| contained transitive closure is admitted and recorded | a legal 3-file chain with a cycle closes, and `lib.c`, `private.h`, `deeper.h` all appear as scanned sources |

## Finding B — directive recognition now fails closed

**Fix.** `directivePattern`/`moduleImportPattern` are gone. `scanIncludes` now
runs the compiler's own pre-directive translation and then a stateful scanner
(`headers.go`):

- `spliceTranslationLines` normalizes `\r\n` and lone `\r` to `\n` and removes
  every backslash-newline splice (translation phase 2). A backslash separated
  from its newline by white space is deliberately **not** spliced: the residual
  operand then fails closed rather than being resolved on a guess.
- `directiveScanner.run` tracks `atLineStart` across white space, comments,
  string/character literals, and `@import`, so a directive introduced after a
  comment is still recognized. A block comment that spans lines sets
  `atLineStart`, matching Clang, which marks the following token as
  start-of-line.
- `readDirective` accepts `#` and the `%:` digraph, and `splitDirectiveName`
  skips the full horizontal white-space set including the form feed and
  vertical tab.
- `readDirectiveBody` replaces each comment inside the directive with one space
  and does **not** end the directive at a block comment's newline — verified
  against the real compiler, which continues the directive.
- Classification is closed: `inclusionDirectives` (`include`, `include_next`,
  `import`) must resolve to an exact literal via the retained
  `literalIncludeOperand`; `classifiableDirectives` is the explicit set of the
  19 non-inclusion directives; the null directive is accepted; **anything else
  `#`-introduced is a rejection**, so an unrecognized spelling is a diagnostic
  and never a hole.
- `@import` is recognized at any column (`startsModuleImport`), and an
  `@import` this grammar cannot resolve to `identifier(.identifier)* ;` is
  rejected rather than dropped.
- `IncludeGrammarID` bumped to `c-family-include-scanner-v2`.

**Compiler evidence.** All four spellings were re-confirmed against the pinned
Apple Clang on this host before writing the fix, each reading the named header
under `clang -std=c17 -fsyntax-only -H`:

| Spelling | `-H` reports the header |
| --- | --- |
| `#inc\`⏎`lude "secret.h"` | yes |
| `/* */ #include "secret.h"` | yes |
| `\f#include "secret.h"` | yes |
| `%:include "secret.h"` | yes |

Two further behaviors were probed directly rather than inferred:
`#include /*`⏎`*/ "secret.h"` reads the header (a block comment does not end a
directive) and `int a;`⏎`/*`⏎`*/ #include "secret.h"` also reads it (a
line-spanning comment restores start-of-line). The scanner reproduces both.

**New vectors** — `TestH13DirectiveRecognitionMatchesTheCompilerTranslation`
(15 subtests). Eleven reject with `swiftpm_header_input_undeclared`: the four
compiler-verified spellings, mid-line `@import`, an unclassifiable
`#curator_secret` directive, a `# 1 "/etc/passwd"` line marker, an unresolvable
`@import 3Secret;`, the line-spanning-comment restart, and **both controls the
verdict required to stay unchanged** — `#include \`⏎`</etc/passwd>` and the
CRLF literal. Four positive controls prove the grammar did not become a blunt
instrument: a comment inside a directive does not end it; `#if`/`#warning`/
`#elif`/`#else`/`#pragma`/`#endif`/`#undef`/`#`/`#line` all pass; character
literals and a quoted `"// not a comment"` do not desynchronize; a Swift source
with `#if os(macOS)` and `#Preview` is not scanned with the C grammar.

## Finding C — the declaration-level C++ gate is condition-neutral

**Fix.** `cxxInteropSelected` was the only input to the declaration-level gate
and is destination-evaluated. `platform.go` now splits it:

- `cxxInteropDeclared(target)` — condition-neutral: does the consumer declare
  `interoperabilityMode(.Cxx)` at all;
- `cxxInteropSelected(target, markers)` — the exact destination verdict,
  unchanged in meaning, now sharing `cxxInteropSetting`.

`TargetInterop` carries both: `CxxInteropMode` (destination verdict, the only
consumer of which is now the destination-specific evidence digest) and
`CxxInteropDeclared` (the declaration). `boundaries.go:78` reads
`CxxInteropDeclared`.

This is the neutral-gate branch rather than the `boundary.Selected` branch on
purpose. `boundary.Mode`, `ABI`, `Runtime`, `InterfaceContract`, and
`CallingConvention` are all inside `InteropBoundaryPayload`, i.e. **capture**.
Gating on `Selected` would have classified the same declared edge as
`cxx_interop` on Darwin and `c_abi` on Linux, publishing two different capture
records for one admitted closure — trading a false rejection for a CGP05
neutrality violation. The neutral declaration keeps the capture byte-identical
and leaves both destination gates (`Profile.CxxInterop`,
`Profile.ObjectiveCRuntime`) exactly where they were, on `boundary.Selected`.

**New vectors** in the S05/S06 family:

- `TestS05ConditionalCxxInteropOptInIsSelectionNeutral` — the reviewer's probe:
  `App` declares both the `CxxLib` dependency and the `.interoperabilityMode`
  opt-in under `.when(platforms: [.macOS])`. Both destinations now close; the
  capture `GraphDigest` is byte-identical; the boundary mode is `cxx_interop`
  on both; `Selected` is true on Darwin and false on Linux;
  `CxxInteropDeclared` is true on both while `CxxInteropMode` is true only on
  Darwin; the evidence digests differ.
- `TestS05ConditionalCxxInteropOptInStillRequiresAnAcceptedProfile` — the
  control the verdict named: unconditional dependency plus conditional opt-in
  on Linux still rejects with `swiftpm_target_platform_unsupported` from
  `requireLanguageProfile`.

## Preserved

Nothing the two verdicts accepted was reopened. Round-1 finding 1 (selection
neutrality, `CGP05` conditional branch), the `closuregraph`
`ConsumesInteropPayload.Condition` contract and its absent-by-default
canonicalization, round-1 finding 3 (`publicHeadersPath`, `H10`/`H11`), the
`literalIncludeOperand` operand gate and `H09`, seam/guard discipline, portable
read-set honesty, and scope hygiene are all untouched and green. The
`closuregraph` canonical goldens file is unmodified and the Ruby oracle still
reports 53 labeled records.

Test matrix, `internal/swiftpminterop`: **66** top-level, **138** including
subtests (was 62/114). Coverage **86.6%** (was 86.0%).

## Evidence

| Gate | Real exit | Result |
| --- | ---: | --- |
| `gofmt -l internal/ cmd/` | 0 | no output |
| `go build ./...` | 0 | clean |
| `go vet ./...` | 0 | clean |
| `go test -count=1 -cover ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/` | 0 | 4 ok; interop coverage 86.6% |
| `go test -race -count=1` over the three changed-or-adjacent packages | 0 | 3 ok |
| `golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/... ./internal/closuregraph/...` (v2.12.2) | 0 | `0 issues.` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | `canonical_goldens=pass labeled_records=53 …`; `canonical_references=pass …` |
| `go test -timeout 9m -count=1 $(go list ./... \| grep -v cmd/curator)` | 0 | **51 ok**, 0 FAIL |
| `go test -count=1 -v ./internal/swiftpminterop/` | 0 | 66 top-level PASS, 138 PASS incl. subtests, 0 FAIL |
| `git diff --check` | 0 | no whitespace damage |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

**Not run, stated plainly:** the monolithic `cmd/curator` package. The headless
single-call cap is 10 minutes and that package alone takes ~8 minutes, so it is
the Orchestrator's gate. It is not a coverage gap for this delta:
`go list -deps ./cmd/curator | grep -c swiftpminterop` is **0**, so no change
in this round is reachable from that package, and this round modified no file
outside `internal/swiftpminterop/`.

Nothing was staged or committed.
