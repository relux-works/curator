# Reviewer verdict for TASK-260811-tkurtl — round 8

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260824-b9ad8d` (Claude claude-opus-5). `task-board spawn
  goal "$TASK_BOARD_RUN_ID"` reports `none (run is not goal-bound)`; no
  directives recorded.
- Reviewed delivery: rework `RUN-260824-070fd7` against finding **K** (delimited
  numeric escapes `\x{…}` / `\o{…}`) of
  `TASK-260811-tkurtl_review-verdict_RUN-260824-3d58b7.md`.
- Reviewed outcome: `TASK-260811-tkurtl_rework-outcome_RUN-260824-3d58b7.md`.
- Nothing was staged, committed, reset, or cleaned. No product or test source
  was modified. One temporary probe test file was created inside
  `internal/swiftpminterop/`, executed, and removed; `go build ./...` is green
  afterwards and `git status --short` is byte-identical to the state received
  apart from this run's own board resources. Compiler probes live under
  `.temp/TASK-260811-tkurtl/probe-r9/`.

## Summary

Finding **K** is closed, correctly and at the right level. `decodeStringLiteral`
now decodes `\x{…}` and `\o{…}`, reproduces all four of the compiler's
delimited-content rejections, rejects every UCN spelling and brace-less `\o`,
and threads the rejection out through `asmTemplateText`. I re-derived the parity
table row by row against the pinned compiler and found no error in it.

Acceptance is blocked on scope item 2 — the class-closure check the brief made
the acceptance-deciding statement. The declared closure is stated on the
**escape** row of the lexical axis, and it is sound there. The row *above* it —
translation phase 2, line splicing — is not closed, and the outcome inherits an
explicit fail-closed claim from round 3 that is false for three of the channels.

`spliceTranslationLines` removes only the exact byte pair `\` `\n`. The pinned
Apple Clang also splices `\` + a run of horizontal white space + `\n`, in every
language mode, warning `-Wbackslash-newline-escape` but performing the splice.
That is a *token-reconstituting* splice: `__as\ ⏎m__` is `__asm__` to the
compiler and four unrelated fragments to the scanner. The `#`-directive channel
survives this (the residual `#inc` is an unclassified directive and rejects, as
the round-3 comment claims), and so does a split inside an asm string literal.
The three channels recognized by **token prefix at an arbitrary column** —
`__asm__`/`__asm`/`asm`, `_Pragma`/`__pragma`, and `@import` — do not: splitting
the keyword makes them vanish with no residual to reject.

Three executed Curator counterexamples follow, all three compiler-verified with
the bytes proven in the object or by a `#error` marker firing inside the module
header.

## Per-item verdict

| # | Scope item (round-8 brief) | Verdict |
| ---: | --- | --- |
| 1 | Finding K closed; H18 vectors; `IncludeGrammarID` bumped | **accepted** |
| 2 | Lexical-axis closure — decisive, intended terminal | **changes requested** (finding **L**) |
| 3 | No regression of rounds 1–7 | **accepted** |
| 4 | Evidence | **accepted** |

## Finding L — CONFIRMED — a backslash/white-space/newline splice dissolves the token-level channel keywords

`spliceTranslationLines` (`internal/swiftpminterop/headers.go:445`) performs
`strings.ReplaceAll(text, "\\\n", "")` and its doc comment states the divergence
deliberately: *"A backslash separated from its newline by white space is
deliberately left unspliced — the residual operand then fails closed instead of
being resolved on a guess."*

The premise is right and the conclusion does not follow. There is a residual to
fail closed on only when the split lands inside a construct the scanner
recognizes **by line position** — a `#`-introduced directive, or a string
literal, both of which reject. `startsAsmStatement`, `startsPragmaOperator`, and
`startsModuleImport` recognize their keyword at any column by prefix, so a split
*inside the keyword* leaves nothing behind: `__as`, `\`, ` `, `m__` are ordinary
content bytes and the statement is never entered.

**Compiler evidence** (Apple clang 21.0.0, `clang-2100.1.1.101`,
`arm64-apple-darwin25.5.0`, plain `clang -c`, no flags except where a module is
required). Reads are proved by a unique marker in the object
(`secret-probe-payload-r9` in `payload.bin`), by `nm`, by the missing-file
error, or by an `#error` marker firing inside the module header — never by
absence of a diagnostic. Baseline object with no directive is 504 B, 0 marker
hits; the direct `.incbin` control is 528 B, 1 marker hit.

| Probe | Result |
| --- | --- |
| `__as\ ⏎m__(".incbin \"payload.bin\"");` in a plain `.c` | exit 0, **528 B, marker hits 1** — byte-size identical to the direct control |
| the same with a missing file | exit 1, `error: Could not find incbin file 'nosuch_zzz_r9.bin'`; the rendered line is `.incbin "…"` |
| `__as\ ⏎m__(".include \"inc_src.s\"");` | exit 0; `nm` shows `_probe_included_symbol_r9` — the file was assembled |
| `_Pra\ ⏎gma("message(\"…\")")` | exit 0; the pragma fires — the operator really is reconstituted |
| `_Pra\ ⏎gma("clang module import SecretKit")` with `-fmodules` | exit 1, `error: SECRET_MODULE_HEADER_WAS_READ_r9` — the module header was read |
| `@imp\ ⏎ort SecretKit;` in a `.m` with `-fmodules` | exit 1, same `#error` — same read |
| separator `\ ` / `\t` / `\v` / `\f` / `"  \t "` / `\r` | **all six** splice: 528 B, marker hits 1 |
| `-std=c89 c99 c11 c17 gnu17 c23`, `c++98 gnu++17 c++20 c++23`, `-x objective-c`, `-x objective-c++` | marker hits 1 in **all twelve** — unlike trigraphs this is *not* mode-dependent |

**Curator, executed** (temporary probe test, since removed):

```
PROBE "asm keyword split by whitespace splice"     ACCEPTED (no rejection)
PROBE "pragma operator split by whitespace splice" ACCEPTED (no rejection)
PROBE "module import split by whitespace splice"   ACCEPTED (no rejection)
PROBE "include directive split mid-name"           REJECTED: swiftpm_header_input_undeclared
PROBE "asm template literal split by splice"       REJECTED: swiftpm_header_input_undeclared
```

The last two are the shapes the round-3 comment reasoned about, and they behave
as it claims. The first three are the ones it does not cover. In portable mode
the scanner is the entire header proof and `-H` reports nothing for the
assembler channel, so this is an admission hole, not a diagnostic gap — the same
class as rounds 2–8.

**Required.** Close it in `spliceTranslationLines`, which is the single point
where the scanner claims to reproduce translation phase 2:

- **Splice `\` + horizontal white space + `\n`** exactly as the compiler does.
  This is the smaller and stronger move: the behaviour is unconditional and
  mode-independent on the pinned compiler (twelve modes verified above), so
  unlike the trigraph case there is nothing to bind per file and no reason to
  reject instead. Use the same horizontal-white-space set the rest of the file
  already uses (`horizontalSpace`, i.e. space, tab, vertical tab, form feed);
  `\r` needs no special case because the CR normalisation above already folds
  it. Rejecting any source containing the sequence is acceptable too but costs
  a real, if unusual, source shape for no security gain.
- **State the phase-axis parity argument explicitly** in the next outcome, the
  way round 8 did for the escape row. Enumerate the phase-1/2 transformations
  the pinned compiler performs before tokenization — trigraph replacement
  (round 4, rejected because mode-dependent), end-of-line normalisation, and
  line splicing in *every* spelling the compiler accepts — and for each say
  whether the scanner reproduces it, rejects it, or is provably identical. The
  round-8 escape table is the right shape one row down; this is the missing row
  above it.

Add `H19` vectors (or extend `H13`) for the three token-level spellings —
`__asm__` split, `_Pragma` split, `@import` split — plus the two controls that
already fail closed, keep both `H18` positive controls green, and bump
`IncludeGrammarID` again.

## Why this is the same class, one row up

Round 5 closed directive spellings and missed a stage; round 6 closed stages and
missed the assembler's expansion layer; round 7 closed that layer and missed the
escape decoding beneath it; round 8 closed the escape decoding and missed the
**splice above it**. `decodeStringLiteral` is now correct, and so are
`classifyAssembly` and `assemblerChannelDirectives` — they are simply never
reached, because the keyword that would reach them was never assembled.

The round-8 closure statement is precise about its own scope — *"the scanner's
reproduction of translation phases 1–5 is complete for the input
`classifyAssembly` consumes"* — and it is true of that input. What it does not
cover is the step that decides whether `classifyAssembly` is called at all.

I looked for further mechanisms on all four axes and found none beyond this one.
Specifically checked and **clean**:

- **Escape row, re-derived against the compiler.** `\x{2e}` and `\o{56}` decode
  to `2e` with 0 warnings under `-Wall -Wextra`; `\x{}`, `\x{12345}`, `\o{8}`,
  `\o56`, `\u{2e}`, `\N{FULL STOP}` are all errors with the messages the table
  records; `\u{e9}` is accepted as `c3 a9` and is over-rejected, which is safe;
  `\e` is `1b`; `\q` is `71` with a warning. Every disposition in the round-8
  table holds. The only immaterial slip is "2 warnings" for `\q`, which is 1
  here.
- **`#line` does not move the quote-include search directory.**
  `#line 1 "../b/main.c"` followed by `#include "hdr.h"` still resolves against
  the real file's directory — the `classifiableDirectives` comment's claim is
  correct, verified with a `_Static_assert` discriminating the two candidate
  headers.
- **Link channels.** `#pragma comment(lib, …)`, `#pragma comment(linker, …)`,
  and `#pragma detect_mismatch(…)` all emit **0** `LC_LINKER_OPTION` on Mach-O
  where `__asm__(".linker_option …")` emits exactly **1**. Round 5's
  inertness proof extends to the two spellings it did not name.
- **C++20 named-module `import`.** `import "secret.h";` and `import <secret.h>;`
  under `-std=c++20` and `-std=c++23` are `error: unknown type name 'import'` /
  `use of undeclared identifier` without a modules flag, so the bare `import`
  keyword is provably inert and needs no recognition.
- **Assembler directive spellings.** I enumerated the MC directive table by
  probing for `unknown directive`. `.rep` is an accepted alias of `.rept` and is
  **not** in `assemblerChannelDirectives`; `.macros_off`, `.noaltmacro`,
  `.exitm`, `.endm`, `.endmacro`, `.endr`, `.bundle_lock`, `.bundle_unlock`,
  `.bundle_align_mode`, `.err`, `.warning`, `.print`, `.abort`, and `.reloc`
  are accepted too. **None is a hole**: `.rept`/`.rep` perform no `\`-parameter
  substitution, so a body of theirs must spell any file-reading directive
  literally, which the whole-template name scan sees, and any `\` marker is
  caught by the round-7 residual rule regardless. The rest are terminators,
  mode switches, or diagnostics that name no file. Worth adding `.rep` to the
  set for consistency, but it is **not** a required change and not part of this
  finding.

## Item 1 — finding K closed — evidence

- **Decoding is present and correct.** `decodeDelimitedEscape(value, start,
  shift)` reads the delimited body at shift 4 / 3 and range-checks the
  accumulator **after every digit**, so `\x{0000002e}` decodes while `\x{12345}`
  cannot wrap into an in-range byte. Empty body, unterminated body, and an
  out-of-base digit each reject. `\o` without `{` rejects; `u`, `U`, and `N`
  reject unconditionally. Cursor advance is right in both branches (the `x`
  branch's `continue` runs the loop post-statement, so `}` is consumed).
- **The composite probe rejects through the round-7 rule, as prescribed.**
  `\x{5c}` decodes to a residual `\`, and the grammar case `"\\x{5c}a"` →
  `"\\a"` pins it.
- **All nine `H18` vectors are present and green**, end-to-end through the
  closure: the hex, octal, split-name, `.include`, `.linker_option`, and
  composite-macro spellings, plus the three UCN forms. The `.linker_option`
  vector correctly lands on `artifact_toolchain_untrusted` rather than the
  generic code, because it decodes to a real round-6 channel member.
- **28 new grammar cases** — 10 decoded (`\x{2e}`, `\x{2E}`, `\o{56}`,
  `\o{056}`, `\x{0000002e}`, split name, `\x{5c}`, `\x{0}`, `\x{ff}`, `\o{0}`)
  and 18 rejected (every malformed delimited body, brace-less `\o`, all four UCN
  spellings). Simple/hex/octal decoding and `\\` → residual-`\` are preserved
  and still pinned.
- **Both positive controls green:** `H18/ordinary_inline_assembly_is_content`
  and `H18/an_identifier_that_contains_an_assembly_keyword_is_content`.
- `IncludeGrammarID` is `c-family-include-scanner-v7`, bumped as required.
- Over-rejection is bounded and honestly characterised: a template spelling a
  byte with a UCN, or with delimited content the compiler itself refuses.

## Item 3 — no regression — evidence

- `internal/swiftpminterop` runs **362 PASS / 0 FAIL** (claim: 362/0, up from
  325 by 9 `H18` + 28 grammar cases).
- Every named family present and green: `H01`–`H18`, `S02`–`S10`, `P01`–`P09`
  (P02/P03 and P06/P08 combined), `CGN03`, `CGN09`, `CGN15`, all three `CGP05`
  cases, and the round-6 cross-package vectors.
- `internal/closuregraph/testdata` is unmodified.
- Seam discipline holds: the only `os/exec` reference in
  `internal/swiftpminterop` is the guard test's own literal.
- Scope hygiene holds: every file in the package is `.go`; no creep into
  `TASK-260811-2qfnai` / `TASK-260811-x611eq`.
- No product code outside `internal/swiftpminterop/` changed this round.

## Item 4 — evidence — gates rerun

| Gate | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go test -count=1` over interop, source, closuregraph | 3 `ok`, 0 FAIL |
| `go test -count=1 -v ./internal/swiftpminterop/` | **362 PASS / 0 FAIL** |
| `golangci-lint run` over the three package trees | `0 issues.` |
| `gofmt -l` over the three trees | no output |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | both lines pass, `labeled_records=53`, `cgp05_target_branches=2`, `cgp10_observation_branches=2` |
| `git diff --check` | exit 0 |
| `task-board validate` | `Board is valid. No issues found.` |
| Orchestrator `TASK-260811-tkurtl_full-go-08.log` | SHA-256 **f91711019672bce20e5dd22e229907d6441c42ff4f9a2fe66bf31ffbd7aa3842** matches the brief exactly; `EXIT:0`, **52 ok**, **0 FAIL** |

Accepted rather than rerun: the monolithic full suite, on the hash-verified
orchestrator log above, and the `-race` runs, per the brief. My three focused
packages plus that log's remaining entries account for all 52 packages.

## Routing

`TASK-260811-tkurtl` -> `to-dev`.

Do not reopen anything rounds 1–8 accepted. Keep `decodeStringLiteral`,
`decodeDelimitedEscape`, the UCN and brace-less-`\o` rejections, the
`asmTemplateText` three-result signature and its distinct reason, the
residual-backslash rejection, `assemblerChannelDirectives`, `classifyAssembly`,
`startsAsmStatement`/`asmKeyword`, `readAsmStatement`, the `.s`/`.S` target
rejection, finding I and `S10`, the per-identity `fakeEvaluator`,
`IncludeReference.SourcePackage`, `H18`, and
`TestAssemblyTemplateGrammarIsClosed`.

Finding **L** is the only blocker. It is a one-line change in
`spliceTranslationLines` (`internal/swiftpminterop/headers.go:445`) plus the
three token-level vectors and the phase-axis parity statement. Verify each new
disposition against the pinned compiler the way rounds 4–8 did.

As a reviewer-archetype run this supplies no `commit_ack`.
