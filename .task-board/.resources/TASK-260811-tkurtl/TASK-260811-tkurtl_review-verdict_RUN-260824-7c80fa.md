# Reviewer verdict for TASK-260811-tkurtl — round 6

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260824-7c80fa` (Claude claude-opus-5). `task-board spawn
  goal "$TASK_BOARD_RUN_ID"` reports `none (run is not goal-bound)`.
- Reviewed delivery: rework `RUN-260824-d648de` against findings **H**
  (integrated-assembler channel) and **I** (`.C`/`.M` case-sensitive
  classification) of
  `TASK-260811-tkurtl_review-verdict_RUN-260824-dd03f8.md`, plus that verdict's
  decisive stage-axis closure requirement and the multi-package coverage gap.
- Reviewed outcome: `TASK-260811-tkurtl_rework-outcome_RUN-260824-dd03f8.md`.
- Nothing was staged, committed, reset, or cleaned. One temporary probe test
  file was created inside `internal/swiftpminterop/`, executed, and removed;
  `go build ./...` is green and `git status --short` is byte-identical to the
  state received apart from this run's own board resources. Compiler probes
  live under `.temp/TASK-260811-tkurtl/probe-r6r/`.

## Summary

Finding **I** is fixed and closed. The multi-package coverage gap is closed and
genuinely executed. Finding **H** is *largely* fixed — the asm keyword is now
recognized at a token boundary, the template grammar and escape decoder are
closed, `.s`/`.S` targets are rejected explicitly and soundly in both
directions, and the eight self-check spellings really do reject.

Acceptance is blocked on scope item 3, the decisive one, for the same structural
reason as round 5: the enumeration is complete on the axis the producer chose
and incomplete on the axis above it. Round 5 was complete on the *directive*
axis and missed a *stage*. Round 6 is complete on the *stage* axis and on the
assembler's *directive-name* set — I independently corroborated that name set
against the shipped compiler — but it treats the assembler's decoded template as
final text. It is not. The integrated assembler has its own macro and
parameter-substitution layer, and that layer synthesizes a rejected directive
name out of tokens `classifyAssembly` reads as content.

Five executed Curator counterexamples follow, four of them with `/etc/passwd`
bytes in the resulting object.

## Per-item verdict

| # | Scope item (round-6 brief) | Verdict |
| ---: | --- | --- |
| 1 | Finding H closed — assembler channel | **changes requested** (finding **J**); the four round-5 probes themselves reproduce as closed |
| 2 | Finding I closed — case-sensitive `.C`/`.M` | **accepted** |
| 3 | Stage-axis closure (decisive) | **changes requested** (finding **J**) |
| 4 | No regression of rounds 1–5 | **accepted** |
| 5 | Evidence | **accepted** |
| — | Multi-package interop coverage gap | **accepted** |

## Finding J — CONFIRMED — the assembler's macro layer builds a rejected directive name from inert tokens

`classifyAssembly` (`internal/swiftpminterop/headers.go:984`) scans the decoded
template for a `.` followed by an identifier and matches that identifier against
`assemblerChannelDirectives`. The integrated assembler does not parse the
template that literally: `.macro`/`.irp`/`.irpc` bodies undergo `\`-parameter
substitution *before* directive lookup, so `.\a` with `a=incbin` becomes
`.incbin`, and `.inc\a\()bin` with an empty argument becomes `.incbin` as well.
In every such spelling the byte after the `.` is `\`, `splitLeadingIdentifier`
returns the empty string, the scanner `continue`s, and the template is content.

The producer's own justification for scanning position-agnostically —
"a `.macro` body can hide a spelling that only runs at expansion" — is exactly
right and exactly one level short: it covers a macro body that *contains* the
literal spelling and not a macro body that *constructs* it.

**Compiler evidence** (Apple clang 21.0.0, `clang-2100.1.1.101`,
`arm64-apple-darwin25.5.0`), plain `.c`, `clang -c`, default mode, no flags.
Reads are proved by the payload in the object *and* by the missing-file error,
never by absence of a diagnostic. Baseline object without the directive is 544
bytes; the direct `.incbin "payload.bin"` control puts the payload at `0x188`.

| Probe (inside `__asm__(…)` in a plain `.c`) | Result |
| --- | --- |
| `.macro D a` / `.\a "payload.bin"` / `.endm` / `D incbin` | exit 0; object **byte-identical** to the direct `.incbin` control, payload at `0x188` |
| same with a **missing** file | exit 1, `error: Could not find incbin file 'nosuchfile_zzz.bin'` at `<instantiation>:1:9`, `note: while in macro instantiation` |
| `.irp x,incbin` / `.\x "/etc/passwd"` / `.endr` | exit 0; `/etc/passwd` content in the object; object 9856 B vs 544 B baseline |
| `.macro D a` / `.inc\a\()bin "/etc/passwd"` / `.endm` / `D ""` | exit 0; same, 9856 B |
| `.macro D a b` / `.\a\b "/etc/passwd"` / `.endm` / `D inc bin` | exit 0; same, 9856 B |
| `.macro D a` / `.\a "/etc/passwd"` / `.endm` / `D include` | exit 1, `Included from <instantiation>:1:` — the file was opened and assembled |
| `.altmacro` + `&`-concatenation of a directive name | `error: unknown directive` — **not** a channel on this assembler |

**Curator, executed** (temporary probe test, since removed):

```
PROBE "macro arg builds directive name" ACCEPTED includes=[]
PROBE "irp builds directive name"       ACCEPTED includes=[]
PROBE "empty separator splices name"    ACCEPTED includes=[]
PROBE "two args splice name"            ACCEPTED includes=[]
PROBE "macro arg builds include"        ACCEPTED includes=[]
```

All five closures succeed, `/etc/passwd` appears in no include set, and no
`swiftpm_header_input_undeclared` is raised. The existing `H18` subtest
`incbin inside an asm macro body` passes only because the literal `.incbin` is
present in that body; it does not exercise the substitution layer at all.

In portable mode the scanner is the entire header proof and `-H` reports nothing
for this channel, so this is an admission hole, not a diagnostic gap — the same
class as rounds 2–5.

**Required.** Close the mechanism, not the five spellings. Two composable moves,
both cheap and both verifiable against the pinned compiler:

- **Reject a decoded template containing a backslash.** After C escape decoding
  a residual `\` byte *is* the assembler's substitution marker; there is no
  other way to produce one, and ordinary inline asm (`"nop\n\t"`, extended asm
  with constraints) decodes to no backslash at all — the existing
  `ordinary inline assembly is content` control confirms this. A `\` the
  grammar cannot evaluate is the same closed-grammar rejection already applied
  to a macro operand and a raw-string template.
- **Reject the macro-expansion directives themselves** — `.macro`, `.irp`,
  `.irpc`, `.rept`, `.altmacro`, `.purgem`, `.macros_on` — because each opens an
  expansion layer this stage does not evaluate. All seven are present in the
  shipped compiler (verified by binary-string enumeration).

Add `H18` vectors for each of the five spellings above, and keep the two
positive controls green.

## Stage-axis audit — what I enumerated beyond the producer's list

Audited adversarially against the shipped compiler, not read.

- **The assembler's file-channel *name* set is right.** Independent enumeration
  of the shipped `clang` binary's assembler strings yields exactly two
  `Could not find … file` channels — `Could not find incbin file '` and
  `Could not find include file '` — plus `.linker_option` and
  `.secure_log_unique`/`.secure_log_reset`. `.dump`/`.load` are the ignored
  pair. The producer's six-member table is complete on this axis and its two
  extra members (`.dump`, `.load`) are conservative. `.secure_log_reset` is not
  in the set and needs nothing: it resets assembler state and names no file.
- **The assembler's macro/substitution layer.** Open — finding **J**.
  `.altmacro` `&`-concatenation is *not* a channel here (`unknown directive`),
  so `\` substitution is the mechanism to close.
- **`.s`/`.S` decision.** Sound in both directions and correctly reasoned. The
  target-level rejection means an ordinary `# comment` line in a
  non-preprocessed lowercase `.s` can no longer be misread as an unclassifiable
  directive, and admission is deliberately unchanged so the bytes are still
  hashed and inventoried rather than dropped. The rejection is condition-neutral
  and fires before any header or module analysis, for Swift targets too. I
  reproduced both `.S` and `.s` subtests.
- **The four round-5 probes.** All reproduce as closed: inline `__asm__`
  `.incbin` absolute and escaping, inline `asm(".include …")`, and the
  `.c`+`boot.S` target. The last is now `swiftpm_target_platform_unsupported`
  rather than an accepted closure.
- **Driver/linker stage.** `#pragma comment(lib|linker, …)` re-verified inert
  (0 `LC_LINKER_OPTION`), against `.linker_option`'s 1. No linker runs under
  `-c`. Curator owns argv and the environment; no admitted source byte reaches
  the driver.
- **Preprocessor stage.** Rounds 2–5, all preserved and green.
- **Parse/Sema/CodeGen.** C++20 module imports are not a channel in any mode
  SwiftPM can select; diagnostics re-open only an already-opened file.
- **`__has_include` / `__has_embed`.** Existence predicates: they probe the
  search path but inject no content, and any content injection still travels
  through a directive the scanner already classifies. Not an independent
  channel.

The correct closure argument for the next round is on the **expansion** axis:
for each stage, state not only which token *spellings* name a file but which
substitution or expansion layers run before that stage's lookup, and prove each
such layer is either evaluated or rejected. The C preprocessor's layer was
closed in rounds 2–5 (macro-hidden `_Pragma`, `#define` classification at the
definition, keyword aliases). The assembler's was not.

## Accepted items — evidence

**Item 2 — finding I closed.** `sourceLanguage` (`language.go:56`) and
`swiftpmsource.targetLanguages` (`graph.go:282`) both switch on the raw
extension for `.C` → C++ / `.M` → Objective-C++ before falling through to the
case-insensitive switch, matching `clang -### -c` on this host. Admission is
unchanged and now documented, so the bytes stay hashed and inventoried.
`TestS10UpperCaseExtensionsSelectTheCompilersLanguage` (7 subtests) is green,
including the decisive S06 case — a Swift consumer with no
`.interoperabilityMode(.Cxx)` against a `.C` provider is
`closure_interop_undeclared` — and the two profile gates. Lowercase behavior is
unchanged. The one dead `".S"` case after `strings.ToLower` was correctly
removed and `.S` is still admitted through `".s"`. Every other uppercase form
(`.CC`, `.CPP`, `.CXX`, `.MM`) lowercases to a language at least as strict as
the driver's, so no further case remains open.

**Multi-package gap closed.** `fakeEvaluator` answering per identity is the
right minimal change, and the three vectors are real rather than read-verified:
`TestCrossPackageIncludeEscapeFailsClosed` proves the vendor header is genuinely
opened and scanned (its rejection originates *inside* that file), and
`TestCrossPackageIncludeKeepsThePerTargetSearchPath` proves nothing widens a
dependency's search path to its consumer's. The additive
`IncludeReference.SourcePackage` is the correct disambiguation and removes no
field.

**Item 4 — no regression.** `internal/swiftpminterop` runs **302 PASS / 0 FAIL**
(claim: 302/0). Every named family is present and green: `H01`–`H18`,
`S02`–`S10`, `P01`–`P09` (P03 and P08 as combined tests), `CGN03`, `CGN09`,
`CGN15`, and all three `CGP05` cases. `internal/closuregraph/testdata` is
unmodified. Seam discipline holds: `exec.Command` over the three production
trees appears only in `guard_test.go` and `swift_integration_test.go`, and the
only `os/exec` reference in `internal/swiftpminterop` is `guard_test.go`. Scope
hygiene holds: every file in the package is `.go`; no Kotlin/Gradle reference
and no creep into `TASK-260811-2qfnai`/`TASK-260811-x611eq`.

**Item 5 — evidence.** Every gate I reran reproduced:

| Gate | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go test -count=1` over interop, source, closuregraph | exit 0 |
| `golangci-lint run` over the three package trees | `0 issues.` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | both lines pass, `labeled_records=53`, `cgp05_target_branches=2`, `cgp10_observation_branches=2` |
| `git diff --check` | exit 0 |
| `task-board validate` | `Board is valid. No issues found.` |
| Orchestrator `TASK-260811-tkurtl_full-go-06.log` | SHA-256 **12341545286a9b388bd28c303590c5396dccf22c30c0f30f3347d25d84314260** matches the brief exactly; `EXIT:0`, **52 ok**, 0 `FAIL` |

Accepted rather than rerun: the monolithic full suite, on the hash-verified
orchestrator log above, per the brief. My three focused packages plus that log's
remaining entries account for all 52 packages.

## Routing

`TASK-260811-tkurtl` -> `to-dev`.

Do not reopen finding I, the multi-package vectors, the `.s`/`.S` target
rejection, the assembler directive-name table, the template grammar, the escape
decoder, the token-boundary `asm` recognition, the bare-`asm` decision, or
anything rounds 1–5 accepted. Keep `assemblySource`, `startsAsmStatement`,
`asmKeyword`, `readAsmStatement`, `asmTemplateText`, `decodeStringLiteral`,
`classifyAssembly`, `assemblerChannelDirectives`, `IncludeReference.SourcePackage`,
the per-identity `fakeEvaluator`, `H18`, `S10`, and
`TestAssemblyTemplateGrammarIsClosed`.

Finding **J** is the only blocker. It is a narrow change in
`internal/swiftpminterop/headers.go` (`classifyAssembly` plus, if taken, a
backslash check in `asmTemplateText`) with five `H18` vectors. Verify each new
rejection against the pinned compiler the way rounds 4–6 did, and bump
`IncludeGrammarID` again.

As a reviewer-archetype run this supplies no `commit_ack`.
