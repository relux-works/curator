# Reviewer verdict for TASK-260811-tkurtl — round 5

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260824-dd03f8` (Claude claude-opus-5). `task-board spawn
  goal "$TASK_BOARD_RUN_ID"` reports no active goal, so this run is not
  goal-bound.
- Reviewed delivery: rework `RUN-260824-a2ed30` against findings G1, G2, G3 of
  `TASK-260811-tkurtl_review-verdict_RUN-260824-a61b7f.md`, plus that verdict's
  decisive channel-axis closure requirement.
- Reviewed outcome: `TASK-260811-tkurtl_rework-outcome_RUN-260824-a61b7f.md`.
- No code was modified, staged, committed, reset, or cleaned. Two temporary
  probe test files were created inside `internal/swiftpminterop/`, executed, and
  removed; `go build ./...` is green and `git status --short` is byte-identical
  to the state received apart from this run's own board resources. Compiler
  probes live under `.temp/TASK-260811-tkurtl/probe-r5/`.

## Summary

G1, G2, and G3 are all **fixed** and reproduce as closed against the round-4
probes verbatim. The fixes are the right shape: `#embed` moved into
`inclusionDirectives` and joins the transitive worklist, the three pragma
module-import spellings route through the same `confineInclude` module path as
`@import`, `clang module build`/`endbuild` rejects outright with no inline-map
machinery, and `scanDirectiveChannels` correctly classifies a macro-hidden
`_Pragma` at its definition. The producer's own audit found four further real
holes (encoding-prefixed literals, `_Pragma(M)`, `_Pragma` inside a define, the
stringizing `DO(#x)` form) that I re-verified as closed.

Acceptance is blocked on scope item 4, again the decisive one. The class is
**not closed**. The producer's enumeration is truthful *within the axis it
chose* — the C preprocessor's directive and pragma space. Its dispositions
spot-check correct. But the axis itself is incomplete: it enumerates channels
of **one** compiler stage. Clang's **integrated assembler** is a second
file-reading stage in the same `clang -c` invocation, it has two directives that
open arbitrary files, and it is reachable from an ordinary `.c` source in the
default mode with no flags at all.

Four executed Curator counterexamples follow, plus a second, independent
confirmed finding on language classification.

## Per-item verdict

| # | Scope item (round-5 brief) | Verdict |
| ---: | --- | --- |
| 1 | G1 closed — `#embed` | **accepted** |
| 2 | G2 closed — pragma module import | **accepted** |
| 3 | G3 closed — inline module map | **accepted** |
| 4 | Channel-axis closure (decisive) | **changes requested** (finding **H**) |
| 5 | No regression of rounds 1–4 | **accepted** |
| 6 | Evidence | **accepted** |

A second finding, **I**, is outside the round-5 brief's scope items but is a
confirmed defect against the task's own acceptance criteria ("classify C, C++,
Objective-C, and Objective-C++ sources") and should be fixed in the same round.

## Findings

### Finding H — CONFIRMED — the integrated assembler reads arbitrary files and no stage sees it

`.incbin` and `.include` are GNU-assembler directives that open a named file.
Clang's integrated assembler implements both, and both are reachable from source
Curator admits, through two distinct doors:

1. **Inline assembly in any C-family source.** `__asm__(".incbin \"…\"")` at
   file scope is ordinary C. The scanner's `run()` sees `__asm__` as an
   identifier, the parenthesised string as a literal, and consumes both as
   content. No `#`, no pragma, no `_Pragma`, no `@import` — the entire round-5
   grammar is bypassed by construction.
2. **`.s` / `.S` sources.** `swiftpmsource.swiftPMSourceExtension`
   (`internal/swiftpmsource/executor_runtime.go:513`) admits `.s` and `.S`, and
   `internal/swiftpmsource/swiftpmsource_test.go:556` asserts that it does — so
   assembly is a declared, expected shape, not an exotic one.
   `classifyTarget` (`internal/swiftpminterop/language.go:92-98`) skips an
   unclassified extension with `continue`, so a target of `.c` + `.S` classifies
   as `KindClang`/`[c]` and passes. `scanAndConfineIncludes`
   (`internal/swiftpminterop/interop.go:571-580`) then queues the `.S` file and
   scans it with the **C preprocessor grammar**, which does not model a single
   assembler directive.

`grep -rn 'asm\|incbin\|\.include' internal/swiftpminterop/*.go` returns nothing
outside an unrelated `includeSearchRoots` call: there is no defence anywhere in
the package.

**Compiler evidence** (Apple clang 21.0.0, `clang-2100.1.1.101`,
`arm64-apple-darwin25.5.0`), default mode, no flags:

| Probe | Result |
| --- | --- |
| `__asm__(".incbin \"/tmp/…probe.bin\"");` in a plain `.c` | exit 0 |
| same, bytes in the object | present at offset `0x188`: `THIS_IS_SECRET_…` |
| same with a **missing** file | `error: Could not find incbin file '…'`, exit 1 |
| `asm(".include \"/tmp/…asm.inc\"");` in a plain `.c` | exit 0 |
| same with a **missing** file | `error: Could not find include file '…'`, exit 1 |
| `clang -fsyntax-only -H` on the `.incbin` source | **reports no read at all** |
| `.incbin "/tmp/…probe.bin"` in a `.S` source | exit 0, bytes in the object |
| same with a **missing** file | `error: Could not find incbin file '…'`, exit 1 |

The read is proved by content and by the missing-file error, not by absence of a
diagnostic — the same standard round 4 applied to `#embed`. The `-H` row matters
independently: this channel is invisible to header-read verification too, so the
observed-read path is not a backstop for it either.

**Curator, executed** (temporary probe test, since removed):

```
PROBE "inline asm incbin absolute" ACCEPTED includes=[CLib.h]
PROBE "inline asm incbin escaping" ACCEPTED includes=[CLib.h]
PROBE "inline asm include"         ACCEPTED includes=[CLib.h]
PROBE assembly-source              ACCEPTED includes=[CLib.h stdio.h]
```

All four closures succeed, `/etc/passwd` appears in no include set, and no
`swiftpm_header_input_undeclared` is raised. The fourth vector is a `CLib` target
declaring `Sources/CLib/lib.c` plus `Sources/CLib/boot.S`, where `boot.S`
contains `.incbin "/etc/passwd"`.

This is exactly the round-2/3/4/5 class: a read the compiler performs that the
declared closure never sees. In portable mode the scanner is the entire header
proof, so this is an admission hole, not a diagnostic gap.

**Required.** The channel question has to be asked of *every* stage `clang -c`
runs, not only the preprocessor. Concretely:

- Recognize `asm`, `__asm`, and `__asm__` at a token boundary — the same
  treatment `_Pragma`/`__pragma` already receive — and classify the assembler
  string operands. A `.include` or `.incbin` operand is a file read: resolve the
  exact literal, confine it, and join it to the transitive worklist, or reject.
  An operand this grammar cannot read must reject, per the closed-grammar rule
  already applied to `_Pragma`.
- Decide `.s`/`.S` explicitly. Either give assembly sources a real assembler
  grammar, or reject a C-family target that declares one
  (`swiftpm_target_platform_unsupported` is the natural code) — scanning them
  with the C preprocessor grammar is neither. Note that the current behaviour is
  also unsound in the false-positive direction: a lowercase `.s` file is *not*
  preprocessed by clang, so an ordinary `# comment` line in it would be rejected
  as an unclassifiable directive.
- Close the enumeration on the stage axis and record it: preprocessor (done),
  integrated assembler (this finding), driver/linker (`#pragma comment(lib|
  linker, …)` — I re-verified both are inert here, 0 `LC_LINKER_OPTION` load
  commands), and any other stage the pinned `clang -c` runs.
- Add an H-family vector set for the inline-asm absolute, escaping, and
  `.include` forms, for the `.S`-source form, for a non-literal asm operand, and
  a positive control that an ordinary inline-asm body with no file-reading
  directive stays content.

### Finding I — CONFIRMED — `.C` and `.M` sources are classified as C and Objective-C, so the C++ interop gate never fires

`sourceLanguage` (`internal/swiftpminterop/language.go:45`) lowercases the
extension before matching, so `.C` maps to `LanguageC` and `.M` maps to
`LanguageObjC`. Clang's driver maps them the other way:

```
$ clang -### -c up.C   →  "-x" "c++"
$ clang -### -c up.M   →  "-x" "objective-c++"
```

`swiftPMSourceExtension` also lowercases, so both files are admitted as target
sources. No case-sensitive filesystem is needed — a target simply containing
`impl.C` is enough.

**Curator, executed** (temporary probe test, since removed):

```
PROBE impl.C languages=[c]              (clang compiles it as C++)
PROBE impl.M languages=[objective-c]    (clang compiles it as Objective-C++)
```

The consequence is a real gate bypass, not a cosmetic label. `implementationCxx`
(`internal/swiftpminterop/boundaries.go:69`) is computed from exactly this
language set, and it drives two rejections: the `Profile.CxxInterop` check at
`boundaries.go:79`, and the `closure_interop_undeclared` at `boundaries.go:92`
that fires when a provider exposes C++ and the Swift consumer declares no
`.interoperabilityMode(.Cxx)`. A provider whose implementation is `impl.C`
reports `[c]`, so neither fires — the S06 class of vector passes when it should
reject. `graph.go:141` is skipped for the same reason, and the recorded
`languages` evidence at `graph.go:386` is wrong.

**Required.** Match the extension case-sensitively where the compiler does:
`.C` → `LanguageCXX`, `.M` → `LanguageObjCXX`, lowercase `.c`/`.m` unchanged.
Add an S-family vector for each, asserting the recorded language and that a
Swift consumer without `.interoperabilityMode(.Cxx)` is rejected against a
`.C` provider.

### What I enumerated beyond the producer's list, and what closed

Audited adversarially, spot-verified against the pinned clang:

- **The producer's preprocessing-directive table.** Dispositions check out. I
  re-verified the two that carry real risk: `comment(lib, …)` and
  `comment(linker, …)` both emit **0** `LC_LINKER_OPTION` load commands here, so
  "provably inert" is truthful; `include_alias` and `GCC dependency` are
  rejections, which is the safe answer regardless.
- **The pragma and `_Pragma`/`__pragma` token space.** All 24 rejected spellings
  and all 6 positive/control subtests of `H17` are green in my rerun. The
  macro-hidden `_Pragma` classification at the definition is the correct call.
- **C++20 module imports** (a channel absent from the producer's list). Not a
  channel on this compiler in any mode SwiftPM can select: `import "/tmp/x.h";`
  fails with `error: unknown type name 'import'` under both `-std=c++20` and
  `-std=c++20 -fmodules`. Header units need `-fmodule-header`, which no admitted
  SwiftPM shape passes. Closed.
- **Diagnostics-driven reads.** Clang re-opens a file only to print a snippet of
  a file it already opened. Not an independent channel.
- **The integrated assembler.** Open — finding H. `.include` and `.incbin` are
  the two file-reading directives in the assembler's grammar; both are
  implemented here and both are reachable from admitted source.

The class is closed for the preprocessor stage on the pinned compiler. It is not
closed across stages, and the correct closure argument for the next round is a
stage enumeration, not a longer directive list.

## Accepted items — evidence

**Item 1 — G1 closed.** The round-4 probes reject verbatim:
`embed_of_an_absolute_path`, `embed_escaping_the_package`,
`embed_with_a_macro_operand`, `embed_with_a_limit_parameter`,
`embed_in_a_header`, and `embed_reaches_an_escaping_include` all PASS, and the
positive `a_contained_embed_resolves_and_is_recorded` proves `data.h` is
recorded once with `Embed` set and `Angled` false. `inclusionDirectives`
(`headers.go:34`) is the right home for it and `confineInclude`
(`interop.go:668`) returns the resolved unit so the embedded file joins the
worklist — the `embed reaches an escaping include` vector proves that join,
because its rejection originates in a directive *inside* the embedded file.

**Item 2 — G2 closed.** `pragma_module_import`, `_Pragma_module_import`,
`pragma_module_import_no_name`, `pragma_module_import_trailing`,
`_Pragma_with_a_wide_literal`, `_Pragma_with_a_u8_literal`,
`_Pragma_with_a_macro_operand`, `_Pragma_without_parentheses`,
`_Pragma_inside_a_define`, `stringizing__Pragma_define`,
`__pragma_module_import`, `__pragma_include_alias`, and
`unbalanced___pragma_operand` all PASS; the positive
`a_pragma_module_import_of_a_declared_module_resolves` records exactly 2 module
imports through both spellings. `applyPragma` (`headers.go:778`) routes through
the same `confineInclude` module path as `@import`, and the closed-grammar
rejection in `classifyPragmaBody` (`headers.go:703`) covers every unresolvable
`clang module` spelling.

**Item 3 — G3 closed.** `pragma_module_build`, `pragma_module_endbuild_alone`,
`pragma_module_load`, and `pragma_module_begin` all PASS. No inline-map parsing
machinery was added; module-map authority stays in `confineModuleMapClosure`
alone, which is the proportional answer the round-4 verdict asked for.

**Item 5 — no regression.** `internal/swiftpminterop` runs 219 PASS / **0
FAIL**. Every named family is present and green: `S02`–`S09` including both
conditional-`Cxx` `S05` vectors and the preserved
`TestS05CxxInteropRequiresAcceptedDestinationProfile` control, `H01`–`H17`,
`P01`–`P09`, `CGN03` (three cases), `CGN09`, `CGN15`, and all three `CGP05`
cases including `TestCGP05ConditionalEdgeKeepsInteropCaptureSelectionNeutral`.
`internal/closuregraph/testdata` is unmodified (`git status --short` on it is
empty). Seam and guard discipline holds: `grep -rl exec.Command` over the three
production trees returns exactly `closureexec/portable_runner.go` and
`closureexec/acquisition.go`, and the only `os/exec` reference in
`internal/swiftpminterop` is `guard_test.go`. Scope hygiene holds: every file in
the package is `.go`, and no Kotlin/Gradle reference or creep into
`TASK-260811-2qfnai`/`TASK-260811-x611eq` appears.

**Item 6 — evidence.** Every gate I reran reproduced:

| Gate | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go test -count=1 -cover` over the four packages | exit 0; interop coverage **87.0%** — matches the claim |
| `go test -race -count=1` over the same four | exit 0 |
| `golangci-lint run` over the three package trees | `0 issues.` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | both lines pass, `labeled_records=53`, `cgp05_target_branches=2`, `cgp10_observation_branches=2` |
| `git diff --check` | exit 0 |
| `task-board validate` | `Board is valid. No issues found.` |
| Orchestrator `TASK-260811-tkurtl_full-go-05.log` | SHA-256 **f3cd65d0a9be852d8fc52cdd0536178c5e382b157af0f5dfc837060a9e606f28** matches the brief exactly; `EXIT:0`, **52 ok**, 0 `FAIL` |

Accepted rather than rerun: the monolithic full suite, on the hash-verified
orchestrator log above, per the brief. My four focused packages plus that log's
remaining entries account for all 52 packages.

## Routing

`TASK-260811-tkurtl` -> `to-dev`.

Do not reopen G1, G2, or G3. Keep `inclusionDirectives` with `embed`, the
`IncludeReference.Embed` field and its worklist join, `classifyPragmaBody`,
`applyPragma`, `readPragmaOperator`, `balancedParenthesis`, `destringizePragma`,
`scanDirectiveChannels`, the `clang module build`/`endbuild` rejection, and the
`H17` vectors — along with everything rounds 1–4 accepted.

Finding H is the decisive one and has four executed Curator counterexamples plus
compiler evidence from this host. Finding I has two. Both are narrow fixes in
`internal/swiftpminterop` (`headers.go` for H, `language.go` for I) plus vectors.

The next round's closure argument should be made on the **stage** axis:
enumerate every stage the pinned `clang -c` runs (preprocessor, integrated
assembler, driver/linker directives), state each stage's file-reading channels
and their disposition, and record which stages provably read nothing. The
directive axis is now well covered; it was the wrong axis to be complete on
alone.

The multi-package interop coverage gap noted in round 3 and repeated in rounds 4
and 5 remains open and is still worth closing in the same round.

As a reviewer-archetype run this supplies no `commit_ack`.
