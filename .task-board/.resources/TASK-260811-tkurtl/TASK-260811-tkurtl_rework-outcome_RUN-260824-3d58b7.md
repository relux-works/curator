# Rework outcome for TASK-260811-tkurtl — round 8 (finding K)

Run: `RUN-260824-070fd7` (Claude claude-opus-5). `task-board spawn goal
"$TASK_BOARD_RUN_ID"` reports `none (run is not goal-bound)`; no directives
recorded.

Reviewed input: `TASK-260811-tkurtl_review-verdict_RUN-260824-3d58b7.md`
(round 7 — finding **J** accepted, no regressions, finding **K** blocking).

Scope: exactly finding **K**. Nothing rounds 1–7 accepted was reopened.

## Finding K — the fix

`decodeStringLiteral` (`internal/swiftpminterop/headers.go`) modelled the C11 /
C++11 escape set and fell through on everything else ("an escape the standard
does not define yields its own character"). The pinned Apple Clang also accepts
the **delimited numeric escapes** `\x{…}` and `\o{…}` as an extension in every
language mode with no diagnostic, so `\x{2e}` is one `.` to the compiler and the
four characters `x{2e}` to the scanner. Both round-7 rules were bypassed at
once: the decoded template carried no `.` for the directive-name scan and no `\`
for the residual-substitution-marker rejection.

Three changes, all in `internal/swiftpminterop/headers.go`:

1. **`decodeStringLiteral` decodes `\x{…}` and `\o{…}`.** A new helper
   `decodeDelimitedEscape(value, start, shift)` reads the delimited body at
   shift 4 (hex) or shift 3 (octal), and reproduces the compiler's four
   rejections exactly: empty body, unterminated body, a digit outside the base,
   and a value past one byte. The accumulator is range-checked after **every**
   digit, so a long body cannot wrap into an in-range byte.
   Decoding rather than blanket-rejecting is what the reviewer prescribed and is
   the stronger move: it makes the composite probe reject *through the existing
   round-7 rule*, because `\x{5c}` now becomes the residual `\`.
2. **`decodeStringLiteral` rejects every universal-character-name form** —
   `\uXXXX`, `\UXXXXXXXX`, `\u{X+}`, `\N{NAME}` — and rejects `\o` not followed
   by `{`. The function now returns `(string, bool)`.
3. **`asmTemplateText` propagates the rejection** and returns
   `(text, reason, ok)`, so an unreadable literal spelling and an
   unreproducible escape form stay distinguishable in the diagnostic. The new
   reason is *"assembly template carries an escape sequence whose decoded value
   this scanner grammar declines to reproduce"*; the code is
   `swiftpm_header_input_undeclared` as before.

The existing simple/hex/octal decoding and the `\\` → residual-`\` rejection are
untouched. `IncludeGrammarID` is bumped `c-family-include-scanner-v6` →
`c-family-include-scanner-v7`.

## The lexical-axis parity enumeration

This is the row beneath the round-7 stage-axis table the reviewer asked for.
Every escape form the pinned compiler accepts in a string literal, and the
decoder's disposition for each. Verified form by form against Apple clang
21.0.0 `clang-2100.1.1.101`, `arm64-apple-darwin25.5.0`, by compiling
`const char probe[] = "<form>";` with `-Wall -Wextra` and printing the resulting
bytes — never by absence of a diagnostic. Full log:
`TASK-260811-tkurtl_round8-parity-evidence.log`.

| Escape form | Compiler disposition | Decoder disposition | Parity |
| --- | --- | --- | --- |
| `\'` `\"` `\?` `\\` | accepted → `27` / `22` / `3f` / `5c` | decoded to the same byte | **decodes** |
| `\a` `\b` `\f` `\n` `\r` `\t` `\v` | accepted → `07 08 0c 0a 0d 09 0b` | decoded to the same byte | **decodes** |
| `\e` | accepted → `1b` (GNU extension, no warning) | decoded to `1b` | **decodes** |
| `\q` and every other undefined escape | accepted with 2 warnings → `71` (its own character) | decoded to its own character | **decodes** |
| `\xH+` in range | accepted → that byte (`\x2e` → `2e`, `\x2E` → `2e`) | decoded to that byte | **decodes** |
| `\O{1,3}` in range | accepted → that byte (`\056` → `2e`, `\0` → `00`, `\1\12\123` → `01 0a 53`) | decoded to that byte | **decodes** |
| `\x{H+}` in range | accepted, **0 warnings**, → that byte (`\x{2e}` `\x{2E}` `\x{0000002e}` → `2e`; `\x{0}` → `00`; `\x{ff}` → `ff`; `\x{5c}` → `5c`) | **decoded to that byte (new)** | **decodes** |
| `\o{O+}` in range | accepted, 0 warnings → that byte (`\o{56}` `\o{056}` → `2e`; `\o{0}` → `00`) | **decoded to that byte (new)** | **decodes** |
| `\x{}` / `\o{}` | `error: delimited escape sequence cannot be empty` | **rejected (new)** | **rejects** |
| `\x{2e` (unterminated) | `error: expected '}'` | **rejected (new)** | **rejects** |
| `\x{g}` / `\o{8}` | `error: invalid digit … in escape sequence` | **rejected (new)** | **rejects** |
| `\x{12345}` / `\o{777}` | `error: hex/octal escape sequence out of range` | **rejected (new)** | **rejects** |
| `\o` / `\o56` (no brace) | `error: expected '{' after '\o' escape sequence` | **rejected (new)** | **rejects** |
| `\N` (no brace) | `error: expected '{' after '\N' escape sequence` | **rejected (new)** | **rejects** |
| `\u{2e}` `\U0000002e` `\N{FULL STOP}` `\u{5c}` | `error: character '.' / '\' cannot be specified by a universal character name` | **rejected (new)** | **rejects** — over-approximate on a form the compiler already refuses |
| `\u{e9}` `\N{LATIN SMALL LETTER E WITH ACUTE}` | accepted → `c3 a9` | **rejected (new)** | **rejects** — over-approximate; a UCN can never denote `.` or `\`, so no admitted asm template needs one |
| `\x` with no hex digits (`\x`, `\xz`) | `error: \x used with no following hex digits` | passes `x` through | **provably inert**: the translation unit does not compile, so the assembler never sees the template |
| `\xH+` past one byte (`\x2e69`) | `error: hex escape sequence out of range` | truncates to the low byte | **provably inert**, same reason |
| `\O+` past `\377` (`\400`) | `error: octal escape sequence out of range` | wraps | **provably inert**, same reason |
| trailing lone `\` | cannot occur — it would escape the closing quote | yields a residual `\` | **rejects** via the round-7 residual-marker rule |

Closure argument for the axis: the set above is the complete escape grammar of a
C string literal — a `\` is followed by a simple-escape character, an octal
digit, `x`, `o`, `u`, `U`, `N`, or an undefined character, and every one of those
eight cases now has an explicit row. There is no ninth branch left to fall
through. Combined with the rows the earlier rounds filled in — trigraphs (round
4), splices and comments (round 3), BOM/NUL/Unicode white space (round 4) — the
scanner's reproduction of translation phases 1–5 is complete for the input
`classifyAssembly` consumes.

**Mode independence.** The delimited forms are not a C23-only spelling. Body
`\x{2e}\o{56}` decodes to `2e 2e` with zero warnings at `-std=c99`, `c11`,
`c17`, `gnu17`, `c23`, `c++98`, `gnu++17`, `c++20`, `c++23`, `-x objective-c`,
and `-x objective-c++` — and at `-std=c89` too (2 pedantic warnings, same
bytes). Unlike trigraphs, this needs no per-mode branch.

## Compiler evidence for the channel probes

`__asm__(TEMPLATE);` at file scope in a plain `.c`, `clang -c`. Reads are proved
by a unique marker (`secret-probe-payload-r8` in `payload.bin`), by `nm`, by the
`LC_LINKER_OPTION` count, or by the missing-file error. Baseline `nop` object is
520 B, 0 marker hits.

| Template | Compiler result |
| --- | --- |
| `.incbin "payload.bin"` (control) | exit 0, 544 B, marker **1** |
| `\x{2e}incbin "payload.bin"` | exit 0, 544 B, marker **1** |
| `\o{56}incbin "payload.bin"` | exit 0, 544 B, marker **1** |
| `\x{2e}\x{69}ncbin "payload.bin"` (name split across two escapes) | exit 0, 544 B, marker **1** |
| `\x{2e}include "inc_src.s"` | exit 0, 560 B, `nm` shows `_probe_included_symbol_r8` |
| `\x{2e}linker_option "-lSecretProbeLib"` | exit 0, exactly **1** `LC_LINKER_OPTION` |
| composite `\x{2e}macro D a` / `\x{2e}\x{5c}a "payload.bin"` / `\x{2e}endm` / `D incbin` | exit 0, 544 B, marker **1** — the finding-J macro layer re-entered with no literal `.macro` and no literal `\` |
| the same composite with a missing file | exit 1, `error: Could not find incbin file 'nosuchfile_zzz_r8.bin'` — the read really is attempted |
| `\u{2e}incbin …` and `\N{FULL STOP}incbin …` (controls) | exit 1, `character '.' cannot be specified by a universal character name` |

All six reviewer counterexamples reproduced independently in this run before the
fix was written.

## New vectors

**`TestH18IntegratedAssemblerChannelsAreClosed`** — 9 new rejected vectors, all
end-to-end through the closure:

| Vector | Expected code |
| --- | --- |
| `asm delimited hex escaped directive` | `swiftpm_header_input_undeclared` |
| `asm delimited octal escaped directive` | `swiftpm_header_input_undeclared` |
| `asm delimited split directive name` | `swiftpm_header_input_undeclared` |
| `asm delimited escaped include` | `swiftpm_header_input_undeclared` |
| `asm delimited escaped linker option` | `artifact_toolchain_untrusted` — it decodes to `.linker_option`, so the round-6 non-read channel classifies it, which is the right code rather than the generic one |
| `asm delimited escape builds a macro` | `swiftpm_header_input_undeclared` |
| `asm braced universal character name` | `swiftpm_header_input_undeclared` |
| `asm numeric universal character name` | `swiftpm_header_input_undeclared` |
| `asm named universal character name` | `swiftpm_header_input_undeclared` |

**`TestAssemblyTemplateGrammarIsClosed`** — 10 new decoded cases pinning the
delimited decoding (`\x{2e}incbin` → `.incbin`, the `\o{…}` and split-name and
leading-zero and `\x{5c}` forms) and 18 new rejected-escape cases covering every
malformed delimited body, `\o` without a brace, and all four UCN spellings. The
two positive controls — `H18/ordinary_inline_assembly_is_content` and
`H18/an_identifier_that_contains_an_assembly_keyword_is_content` — are unchanged
and green.

## Over-rejection cost

The residue is an assembler template that spells a byte with a UCN
(`\u{e9}`, `\N{…}`) or with a delimited escape the compiler itself refuses. No
admitted SwiftPM shape uses either; ordinary inline assembly carries no
delimited escape and no UCN at all, which the positive control pins.

## Gates

Every command run directly as a standalone process; no `tee`, no pipe chain.

| Gate | Exit | Result |
| --- | ---: | --- |
| `go build ./...` | 0 | clean |
| `go test -count=1 ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/` | 0 | 3 `ok` |
| `go test -count=1 -v ./internal/swiftpminterop/` | 0 | **362 PASS / 0 FAIL** (was 325/0; +37 = 9 H18 + 10 decoded + 18 rejected) |
| `go test -count=1 -race` over the same three | 0 | 3 `ok` |
| `go vet` over the same three | 0 | clean |
| `gofmt -l` over the same three | 0 | no output |
| `golangci-lint run` (v2.12.2, go1.25.5) over the same three trees | 0 | `0 issues.` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | both lines pass, `labeled_records=53`, `cgp05_target_branches=2`, `cgp10_observation_branches=2` |
| `go test -count=1 <53 packages, suite minus cmd/curator>` | 0 | **51 ok, 2 no test files, 0 FAIL** |
| `git diff --check` | 0 | clean |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

`cmd/curator` was not run in this session: it is ~10 min on its own and the
monolithic full suite is the Orchestrator's job per the brief. No code outside
`internal/swiftpminterop/` changed, so it is unaffected.

Nothing was staged, committed, reset, or cleaned. Files changed:
`internal/swiftpminterop/headers.go`, `internal/swiftpminterop/parser_test.go`,
`internal/swiftpminterop/modulemap_test.go`, plus one appended `LOGBOOK.md`
entry. Compiler probes live under `.temp/TASK-260811-tkurtl/probe-r8/`.
