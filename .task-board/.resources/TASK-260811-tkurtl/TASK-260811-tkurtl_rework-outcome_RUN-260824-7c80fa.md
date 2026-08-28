# Rework outcome for TASK-260811-tkurtl — round 7

Addresses: finding **J** of
`TASK-260811-tkurtl_review-verdict_RUN-260824-7c80fa.md` (Claude
claude-opus-5). That verdict's finding **J** was its only blocker; every other
item across rounds 1–6 is preserved unchanged.

Nothing was staged, committed, reset, or cleaned. Compiler probes live under
`.temp/TASK-260811-tkurtl/probe-r7/`; gate logs under
`.temp/TASK-260811-tkurtl/round7-*.log`.

## Finding J — the assembler's macro layer synthesizes directive names

`.macro`/`.irp`/`.irpc` bodies undergo `\`-parameter substitution **before**
the integrated assembler looks a directive name up. `classifyAssembly` saw a
`\` after the `.`, got an empty identifier from `splitLeadingIdentifier`, and
continued as content — so the rejected name never had to appear in the
template at all.

### Compiler evidence, this run

Apple clang 21.0.0 (`clang-2100.1.1.101`), `arm64-apple-darwin25.5.0`, plain
`.c`, `clang -c`, default mode, no flags. Reads are proved by a unique marker
string in the object (`payload.bin` containing `secret-probe-payload-r7`) and,
for the negative form, by the missing-file error — never by absence of a
diagnostic. Using a marker payload rather than `/etc/passwd` makes the read
proof exact rather than size-inferred; the reviewer's `/etc/passwd` runs are
not contradicted, only re-proved on a distinguishable payload.

| Probe inside `__asm__(…)` in a plain `.c` | Result |
| --- | --- |
| baseline, no directive | exit 0, object 512 B, marker hits 0 |
| `.incbin "payload.bin"` (direct control) | exit 0, object 536 B, marker hits **1** |
| `.macro D a` / `.\a "payload.bin"` / `.endm` / `D incbin` | exit 0, 536 B, marker hits **1** |
| `.irp x,incbin` / `.\x "payload.bin"` / `.endr` | exit 0, 536 B, marker hits **1** |
| `.macro D a` / `.inc\a\()bin "payload.bin"` / `.endm` / `D ""` | exit 0, 536 B, marker hits **1** |
| `.macro D a b` / `.\a\b "payload.bin"` / `.endm` / `D inc bin` | exit 0, 536 B, marker hits **1** |
| `.macro D a` / `.\a "inc_src.s"` / `.endm` / `D include` | exit 0; `nm` shows `_probe_included_symbol_r7` — the file was assembled |
| same shape, **missing** file | exit 1, `error: Could not find incbin file 'nosuchfile_zzz_r7.bin'` at `<instantiation>:1:9`, `note: while in macro instantiation` |

Directive-existence enumeration against the shipped assembler — an
unrecognized name really is diagnosed, so a clean exit is a positive
recognition result:

| Directive probe | exit | `unknown directive` |
| --- | ---: | ---: |
| `.macro` / `.endm` | 0 | 0 |
| `.irp` / `.endr` | 0 | 0 |
| `.irpc` / `.endr` | 0 | 0 |
| `.rept` / `.endr` | 0 | 0 |
| `.altmacro` | 0 | 0 |
| `.purgem` | 0 | 0 |
| `.macros_on` | 0 | 0 |
| `.zzz_not_a_directive_r7` (control) | 1 | **1** |

`.altmacro` `&`-concatenation re-verified **not** a channel here:
`.altmacro` / `.macro D a` / `.inc&a&bin "payload.bin"` / `.endm` / `D bin`
gives `error: unknown directive` at `<instantiation>:1:1` and 0 marker hits.
It is therefore not modeled as a mechanism; the `.altmacro` directive is still
rejected because it selects a macro dialect this grammar does not model.

### Fix — the mechanism, not the five spellings

Two independent, composable moves in `internal/swiftpminterop/headers.go`.
Either one alone closes all five reviewer spellings; both are present so the
closure does not depend on a single rule.

1. **Residual backslash in the decoded template rejects** — `classifyAssembly`
   now returns `CodeHeaderInputUndeclared` before reading any name when the
   decoded text contains a `\`. C escape decoding has already run at that
   point, so a surviving backslash *is* the assembler's substitution marker
   and nothing else. This is the same closed-grammar rejection already applied
   to a non-literal macro operand and a raw-string template.
2. **The macro-expansion directives reject** — `assemblerChannelDirectives`
   gains `macro`, `irp`, `irpc`, `rept`, `altmacro`, `purgem`, `macros_on`,
   each `CodeHeaderInputUndeclared`. Each opens an expansion layer this stage
   does not evaluate.

`IncludeGrammarID` bumped `c-family-include-scanner-v5` →
`c-family-include-scanner-v6`.

Cost on admitted shapes is nil: ordinary inline assembly decodes to no
backslash at all. `"nop\n\t"` decodes to a newline and a tab; extended-asm
constraints, clobbers, and `extern int named __asm__("_named_symbol")` carry
none. The over-rejected residue is an assembler string literal that itself
contains an escaped backslash (`.ascii "a\\b"`), which no admitted SwiftPM
shape uses; that is the conservative direction.

### New vectors

`H18` (`modulemap_test.go`), 13 new subtests — the five reviewer spellings
plus the seven expansion directives plus the `&`-concatenation form:

```
asm_macro_argument_builds_a_directive_name   asm_macro_directive
asm_irp_builds_a_directive_name              asm_irp_directive
asm_empty_separator_splices_a_directive      asm_irpc_directive
asm_two_arguments_splice_a_directive         asm_rept_directive
asm_macro_argument_builds_an_include         asm_altmacro_directive
asm_altmacro_ampersand_concatenation         asm_purgem_directive
                                             asm_macros_on_directive
```

`TestAssemblyTemplateGrammarIsClosed` (`parser_test.go`), 10 new cases proving
the two moves are independent: three bodies carrying a substitution marker with
**no** expansion directive anywhere in the text, and seven expansion-directive
bodies with **no** backslash anywhere in the text.

Both positive controls stay green:
`H18/ordinary_inline_assembly_is_content` and
`H18/an_identifier_that_contains_an_assembly_keyword_is_content`.

## Stage-axis enumeration — updated

Only the assembler's expansion row changes; every other row is carried forward
from the round-6 verdict's audit unchanged.

| Stage | Expansion/substitution layer running before its lookup | Disposition |
| --- | --- | --- |
| Preprocessor | macro expansion, `_Pragma`/`__pragma`, keyword aliases, splice, trigraphs, digraphs | Evaluated or rejected at the definition — rounds 2–5, preserved |
| Integrated assembler — file channels | none; literal directive names | 6 rejected names: `incbin`, `include`, `dump`, `load`, `linker_option`, `secure_log_unique` — round 6, preserved |
| **Integrated assembler — macro/parameter substitution** | `\`-parameter substitution in `.macro`/`.irp`/`.irpc` bodies; `.altmacro` `&`-concatenation proven inert | **Rejected wholesale** — residual `\` rejects, and all 7 expansion directives reject. The layer is never entered, so no synthesized name can reach directive lookup |
| Driver/linker | none reachable; Curator owns argv and environment; `#pragma comment(lib\|linker,…)` inert (0 `LC_LINKER_OPTION`) | Provably inert — round 6, preserved |
| Parse/Sema/CodeGen | C++20 module import not selectable by SwiftPM; diagnostics reopen only opened files | Provably inert — round 6, preserved |
| `__has_include` / `__has_embed` | existence predicates; injection still travels a classified directive | Not an independent channel — round 6, preserved |

The macro layer is the last open member of the class per the round-6 verdict's
own completed stage-axis audit. It is now closed by rejection rather than by
evaluation, which is the same answer the grammar already gives every other
layer it declines to model.

## Preserved from rounds 1–6

Not reopened, not modified: finding **I** and `S10`; the multi-package
vectors, per-identity `fakeEvaluator`, and `IncludeReference.SourcePackage`;
the `.s`/`.S` target rejection; the assembler directive-name table (extended
additively only); `asmTemplateText`'s template grammar; `decodeStringLiteral`;
token-boundary `asm` recognition via `startsAsmStatement`/`asmKeyword`; the
bare-`asm` decision; `readAsmStatement`; and everything rounds 1–5 accepted.
`internal/closuregraph/testdata` is unmodified.

## Evidence

Every gate run as a standalone process; exit codes are the process's own.

| Gate | Exit | Result |
| --- | ---: | --- |
| `go build ./...` | 0 | clean |
| `go test -count=1` over interop, source, closuregraph | 0 | 3 `ok` |
| `go test -count=1 -race` over the same three | 0 | 3 `ok` (interop 46.4s, source 13.5s, closuregraph 108.6s) |
| `go test -count=1 -v ./internal/swiftpminterop/` | 0 | **325 PASS / 0 FAIL** (was 302/0; +13 `H18`, +10 grammar) |
| `go vet` over the same three | 0 | clean |
| `gofmt -l` over the same three | 0 | 0 files |
| `golangci-lint run` (v2.12.2, go1.25.5) over the same three | 0 | `0 issues.` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | both lines pass, `labeled_records=53`, `cgp05_target_branches=2`, `cgp10_observation_branches=2` |
| `go test -count=1 -timeout 9m` over all 53 packages except `cmd/curator` | 0 | 51 `ok`, 2 `[no test files]`, 0 `FAIL` |
| `git diff --check` | 0 | clean |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

`cmd/curator` is excluded from the bounded call by the brief; it is the
Orchestrator's monolithic-suite responsibility. No product code outside
`internal/swiftpminterop/` changed this round.

Logs: `round7-focused.log`, `round7-race.log`, `round7-verbose.log`,
`round7-vet.log`, `round7-gofmt.log`, `round7-lint.log`, `round7-golden.log`,
`round7-nocmd.log`.
