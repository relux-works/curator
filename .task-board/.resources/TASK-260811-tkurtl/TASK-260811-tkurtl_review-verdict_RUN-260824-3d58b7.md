# Reviewer verdict for TASK-260811-tkurtl — round 7

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260824-3d58b7` (Claude claude-opus-5). `task-board spawn
  goal "$TASK_BOARD_RUN_ID"` reports `none (run is not goal-bound)`.
- Reviewed delivery: rework `RUN-260824-d853b2` against finding **J**
  (the integrated assembler's macro/parameter-substitution layer) of
  `TASK-260811-tkurtl_review-verdict_RUN-260824-7c80fa.md`.
- Reviewed outcome: `TASK-260811-tkurtl_rework-outcome_RUN-260824-7c80fa.md`.
- Nothing was staged, committed, reset, or cleaned. No product or test source
  was modified. One temporary probe test file was created inside
  `internal/swiftpminterop/`, executed, and removed; `go build ./...` is green
  and `git status --short` is byte-identical to the state received apart from
  this run's own board resources and one appended `LOGBOOK.md` entry. Compiler
  probes live under `.temp/TASK-260811-tkurtl/probe-r7rev/`.

## Summary

Finding **J** itself is closed, correctly and on the mechanism rather than the
spellings. Both prescribed moves are present and independent, all five round-6
probes reproduce as rejected, and both positive controls stay green.

Acceptance is blocked on scope item 2 — the class-closure check the brief made
the acceptance-deciding statement. The brief named the exact place to look
("escape-decoding corner cases"), and there is one: `decodeStringLiteral`
models the C escape set of C11/C++11, but the pinned Apple Clang also accepts
the **delimited numeric escape sequences** `\x{…}` and `\o{…}`, in every
language mode, with no diagnostic even under `-Wall -Wextra`. `\x{2e}` is a
`.` to the compiler and the two characters `x{2e}` to the scanner, so the
decoded template the scanner classifies is not the assembler text the compiler
emits. Both round-7 moves are bypassed at once: the reconstructed name carries
no `.` for the name scan and no `\` for the residual-backslash rejection.

Six executed Curator counterexamples follow, five of them compiler-verified
with the payload in the object, including one that re-enters the very macro
layer finding J closed.

## Per-item verdict

| # | Scope item (round-7 brief) | Verdict |
| ---: | --- | --- |
| 1 | Finding J closed — mechanism, not spellings | **accepted** |
| 2 | Class closure — final adversarial check (decisive) | **changes requested** (finding **K**) |
| 3 | No regression of rounds 1–6 | **accepted** |
| 4 | Evidence | **accepted** |

## Finding K — CONFIRMED — delimited numeric escapes desynchronize the decoded template from the assembler text

`decodeStringLiteral` (`internal/swiftpminterop/headers.go:932`) handles the
simple escapes, `\xHH…`, and `\NNN`, and falls through to "an escape the
standard does not define yields its own character". That fall-through is what
this compiler does for `\q`. It is **not** what it does for `\x{…}` and
`\o{…}`: Clang implements delimited escape sequences as an extension available
in every mode, so `\x{2e}` is one character, `.`, and `\o{56}` is the same
character. The scanner's `\x` branch reads the byte after `x`, finds `{` is not
a hex digit, emits a literal `x`, and copies `{2e}` through; the `\o` branch
does not exist at all, so `o{56}` is copied through. Either way the decoded
text contains no `.` before the directive name and no `\` anywhere, so
`classifyAssembly` returns content on both of its rules.

This is not the `\\`-survives-decode case the brief hypothesised — that one is
already safe, because `"\\"` decodes to one `\` and is rejected. It is the
opposite direction: an escape the compiler decodes into the marker byte that
the scanner does not decode at all.

**Compiler evidence** (Apple clang 21.0.0, `clang-2100.1.1.101`,
`arm64-apple-darwin25.5.0`), plain `.c`, `clang -c`, default mode, no flags.
Reads are proved by a unique marker in the object (`payload.bin` containing
`secret-probe-payload-r7rev`), by `/etc/passwd` content, by `nm`, or by the
missing-file error — never by absence of a diagnostic. Baseline object without
a directive is 520 B; the direct `.incbin "payload.bin"` control is 560 B with
1 marker hit.

| Probe inside `__asm__(…)` in a plain `.c` | Result |
| --- | --- |
| `\x{2e}incbin "payload.bin"` | exit 0, 552 B, marker hits **1** |
| `\o{56}incbin "payload.bin"` | exit 0, 552 B, marker hits **1** |
| `\x{2e}incbin "/etc/passwd"` | exit 0, 9872 B, 3 hits of `root:` |
| `\x{2e}\x{69}ncbin "payload.bin"` (name split across two escapes) | exit 0, marker hits **1** |
| `\x{2e}include "inc_src.s"` | exit 0; `nm` shows `_probe_included_symbol_r7rev` — the file was assembled |
| `\x{2e}linker_option "-lSecretProbeLib"` | exit 0; exactly **1** `LC_LINKER_OPTION`, `string #1 -lSecretProbeLib` |
| **composite:** `\x{2e}macro D a` / `\x{2e}\x{5c}a "payload.bin"` / `\x{2e}endm` / `D incbin` | exit 0, 560 B, marker hits **1** — the macro layer of finding J, re-entered with no literal `.macro` and no literal `\` |
| same shape, **missing** file | exit 1, `error: Could not find incbin file 'nosuchfile_zzz_r7rev.bin'`, and the rendered line is `.incbin "…"` |
| `-Wall -Wextra` on the `\x{2e}` form | exit 0, **0 warnings** |
| `-std=gnu17 / c17 / gnu++17 / c++20` | exit 0 and marker hits 1 in **all four**; not mode-dependent |
| `\u{2e}incbin …` and `\N{FULL STOP}incbin …` (controls) | exit 1, `character '.' cannot be specified by a universal character name` — the UCN forms are blocked by the basic-character restriction, so `\x{}`/`\o{}` are the mechanism |
| `\x2e69incbin …` (control) | exit 1, `hex escape sequence out of range` — the undelimited form cannot reach past one byte |

**Curator, executed** (temporary probe test, since removed):

```
PROBE "delimited hex incbin"    ACCEPTED (no rejection)
PROBE "delimited octal incbin"  ACCEPTED (no rejection)
PROBE "delimited hex include"   ACCEPTED (no rejection)
PROBE "delimited whole name"    ACCEPTED (no rejection)
PROBE "delimited macro layer"   ACCEPTED (no rejection)
PROBE "delimited linker option" ACCEPTED (no rejection)
```

All six closures succeed and no diagnostic is raised. In portable mode the
scanner is the entire header proof and `-H` reports nothing for this channel,
so this is an admission hole, not a diagnostic gap — the same class as rounds
2–7.

**Required.** Close it in `decodeStringLiteral`, which is the single point
where the scanner claims to reproduce the compiler's escape decoding:

- **Decode `\x{…}` and `\o{…}`** to the value they denote, or **reject a
  template containing either**. Decoding is the smaller change and keeps the
  decoder's contract ("what the compiler passes on is what gets classified")
  literally true; rejecting is acceptable too, since no admitted SwiftPM shape
  spells an assembler directive with a delimited escape. Either way the
  composite probe must then reject, because `\x{5c}` becomes the residual `\`
  the round-7 rule already rejects.
- **Reject `\u{…}` and `\N{…}`** rather than passing `u{…}`/`N{…}` through.
  Both are compile errors *today* for a basic character, so no current
  counterexample exists — but the fall-through is unsound for the same reason
  as `\x{}`, and the decoder is the one place the grammar asserts parity with
  the compiler. This is the cheap direction.
- **State the parity argument explicitly** in the next outcome: enumerate the
  escape forms the pinned compiler accepts in a string literal, and for each
  one say whether the decoder decodes it, rejects it, or is provably identical
  to the compiler in passing it through. The round-7 outcome's stage-axis table
  is the right shape one level up; this is the missing row beneath it.

Add `H18` vectors for the hex, octal, split-name, `.include`,
`.linker_option`, and composite-macro-layer spellings, keep both positive
controls green, and bump `IncludeGrammarID` again.

## Why this is the same class, one level down

Rounds 2–7 each closed a layer and left the layer beneath it. Round 5 closed
directive spellings and missed a stage; round 6 closed stages and missed the
assembler's expansion layer; round 7 closed that layer and missed the **lexical
decoding** beneath every one of them. `classifyAssembly` and
`assemblerChannelDirectives` are both correct — they are simply being handed
the wrong bytes.

That also means the round-7 outcome's closure statement is true of the layer it
audited and not of the input to it. The next round's closure argument should be
on the **lexical** axis: before any classification runs, the scanner's
reproduction of translation phases 1–5 must be complete, and each escape form
the compiler accepts must be decoded, rejected, or proven inert. Trigraphs
(round 4), splices and comments (round 3), BOM/NUL/Unicode white space (round
4), and simple/hex/octal escapes (round 5) are rows already filled in; the
delimited forms are the missing ones. I found no further mismatch beyond them:
`\u{}`/`\N{}` are compile errors for every character that could build a
directive, undelimited `\xHH…` beyond one byte is a compile error, and `\\`
correctly yields a rejected residual `\`.

## Item 1 — finding J closed — evidence

Both prescribed moves are present and genuinely independent.

- **Residual backslash.** `classifyAssembly` (`headers.go:1023`) returns
  `CodeHeaderInputUndeclared` before reading any name when the decoded text
  contains `\`. Placement is right: after `asmTemplateText` has concatenated
  adjacent literals and run `decodeStringLiteral`, so a surviving backslash
  really is the assembler's own marker.
- **Expansion directives.** `assemblerChannelDirectives` gains `macro`, `irp`,
  `irpc`, `rept`, `altmacro`, `purgem`, `macros_on`, all
  `CodeHeaderInputUndeclared`, additively — the six round-6 members are
  unchanged.
- **All five round-6 probes reproduce as rejected**, as `H18` subtests
  `asm_macro_argument_builds_a_directive_name`, `asm_irp_builds_a_directive_name`,
  `asm_empty_separator_splices_a_directive`,
  `asm_two_arguments_splice_a_directive`, and
  `asm_macro_argument_builds_an_include`, plus the seven expansion-directive
  subtests and `asm_altmacro_ampersand_concatenation`.
- **Independence is proved, not asserted.** `TestAssemblyTemplateGrammarIsClosed`
  carries three marker-only bodies with no expansion directive
  (`a_bare_substitution_marker`, `a_trailing_substitution_marker`,
  `a_spliced_name_with_no_directive`) and seven directive-only bodies with no
  backslash (`macro_with_no_marker` … `macros_on_with_no_marker`). Each move
  alone closes the reviewer's five spellings.
- **Both positive controls green:** `H18/ordinary_inline_assembly_is_content`
  (including `"nop"`, extended asm with constraints and clobbers, and
  `extern int named __asm__("_named_symbol")`) and
  `H18/an_identifier_that_contains_an_assembly_keyword_is_content`.
- **Over-rejection is bounded and correctly characterised.** The residue is an
  assembler string literal containing an escaped backslash (`.ascii "a\\b"`),
  which no admitted SwiftPM shape uses.

`TestH18IntegratedAssemblerChannelsAreClosed` and
`TestAssemblyTemplateGrammarIsClosed` run **93 subtests, 0 failures**.

## Item 3 — no regression — evidence

- `internal/swiftpminterop` runs **325 PASS / 0 FAIL** (claim: 325/0, up from
  302 by the 13 `H18` and 10 grammar cases).
- Every named family present and green: `H01`–`H18`, `S02`–`S10`, `P01`–`P09`
  (P02/P03 and P06/P08 as combined tests), `CGN03`, `CGN09`, `CGN15`, all three
  `CGP05` cases, and the three cross-package vectors from round 6.
- `IncludeGrammarID` is `c-family-include-scanner-v6`, bumped as required.
- `internal/closuregraph/testdata` is unmodified.
- Seam discipline holds: the only `os/exec` reference in
  `internal/swiftpminterop` is the guard test's own literal.
- Scope hygiene holds: every file in the package is `.go`; no creep into
  `TASK-260811-2qfnai`/`TASK-260811-x611eq`.
- No product code outside `internal/swiftpminterop/` changed this round.

## Item 4 — evidence — gates rerun

| Gate | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go test -count=1` over interop, source, closuregraph | exit 0, 3 `ok` |
| `go test -count=1 -v ./internal/swiftpminterop/` | 325 PASS / 0 FAIL |
| `golangci-lint run` over the three package trees | `0 issues.` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | both lines pass, `labeled_records=53`, `cgp05_target_branches=2`, `cgp10_observation_branches=2` |
| `git diff --check` | exit 0 |
| `task-board validate` | `Board is valid. No issues found.` |
| Orchestrator `TASK-260811-tkurtl_full-go-07.log` | SHA-256 **f2419ad1ce6de9b985f16d1e311c24c1eddcdbfc17b61e6222c59f2e8db5c63f** matches the brief exactly; `EXIT:0`, **52 ok**, 0 `FAIL` |

Accepted rather than rerun: the monolithic full suite, on the hash-verified
orchestrator log above, and the `-race` runs, per the brief. My three focused
packages plus that log's remaining entries account for all 52 packages.

The stage-axis enumeration in the round-7 outcome does record the macro layer
as rejected wholesale, as required — that row is correct. What it does not
carry is the lexical row beneath it, which is finding **K**.

## Routing

`TASK-260811-tkurtl` -> `to-dev`.

Do not reopen anything rounds 1–7 accepted. Keep the residual-backslash
rejection, the seven expansion directives, `assemblerChannelDirectives`,
`asmTemplateText`, `classifyAssembly`, `startsAsmStatement`/`asmKeyword`,
`readAsmStatement`, the `.s`/`.S` target rejection, finding I and `S10`, the
per-identity `fakeEvaluator`, `IncludeReference.SourcePackage`, `H18`, and
`TestAssemblyTemplateGrammarIsClosed`.

Finding **K** is the only blocker. It is a narrow change in
`decodeStringLiteral` (`internal/swiftpminterop/headers.go:932`) with six
`H18` vectors and the escape-form parity statement. Verify each new
disposition against the pinned compiler the way rounds 4–7 did.

As a reviewer-archetype run this supplies no `commit_ack`.
