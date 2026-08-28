# TASK-260811-tkurtl rework outcome — posture pivot to reject-by-default

Run: `RUN-260824-f1733b`
Role: developer
Date: 2026-08-24

## What changed and why

Seven consecutive review rounds each found one more real member of one class: a
file the pinned Apple Clang reads that the static portable-mode scanner did not
see. Round 6 found the integrated assembler, round 7 its macro-substitution
layer, round 8 its delimited numeric escapes, and finding L of the round-8
verdict found that an incomplete translation phase 2 dissolved every
token-level keyword the previous three rounds had added.

The class did not shrink under any of those fixes, because portable mode had
made the static scanner the *entire* proof of header closure. That construction
requires byte-perfect reproduction of a compiler front end plus an assembler,
and it does not terminate.

This round implements the operator's decision to change the acceptance posture
instead. Portable mode now **positively admits a small allowlist of trivially
safe, file-read-relevant forms and fails closed on everything else it cannot
prove reads no file.** A new exotic channel discovered after this is written
cannot be admitted, because it is not in an allowlist nobody added it to — and
closing it requires no new emulation.

Nothing in the accepted positive path from rounds 1-8 is regressed. Normal
SwiftPM C/C++/Objective-C/Objective-C++ targets with plain includes, imports,
module maps, and typed interop boundaries still admit, with the same language
classification and the same boundary evidence.

All changes are inside `internal/swiftpminterop`. `closuregraph`,
`swiftpmsource`, and the canonical goldens are untouched this round.

## Move 1 — translation phases 1-3 are finished, and that is a closed set

Phases 1-3 are small and enumerable, and they must be exact: a reject-by-default
keyword or directive match is only closed if a splice, a trigraph, or a comment
cannot reconstitute the token past the scanner. This is the one place where
reproducing the compiler is both required and bounded.

`spliceTranslationLines` now removes a backslash followed by **any run of
horizontal white space** (space, tab, vertical tab, form feed) and a newline.
Carriage returns need no case of their own because the normalization above
already folds them. The splice is unconditional and mode-independent, so unlike
phase-1 trigraph replacement there is nothing to bind per file and no reason to
reject instead.

The previous divergence was an admission hole, not a conservative choice. Its
premise — "the residual then fails closed" — holds only for constructs
recognized by *line position*: a `#`-introduced directive or a string literal,
both of which still reject. `startsAsmStatement`, `startsPragmaOperator`, and
`startsModuleImport` match their keyword **by prefix at any column**, so a split
*inside* the keyword left ordinary content bytes and the construct was never
entered, while the compiler reconstituted it and performed the read.

### Phase-axis parity argument

| Stage | Disposition | Evidence / residual |
| --- | --- | --- |
| Line-ending normalization | Reproduced exactly | `\r\n` and `\r` fold to `\n` before anything else runs. |
| Phase 1 — trigraphs | **Fails closed** | Replacement is mode-dependent on the pinned compiler (replaced under `-std=c89/c99/c11/c17`, `c++14`; ignored under the GNU modes SwiftPM selects, `c++17`, ObjC++ default). A source containing any of the nine sequences rejects rather than being translated under an assumed mode. |
| Phase 1 — BOM, NUL, Unicode white space | Reproduced for the verified set; **residual fails closed** | `lineStartWhiteSpace` carries the code points verified to keep `#include` a directive. A byte sequence it cannot classify demotes the line to content *unless a directive follows it*, in which case `readLineStartPrefix` rejects. |
| Phase 2 — line splices | **Reproduced exactly** (this round) | `\` + `[ \t\v\f]*` + `\n` removed unconditionally. Six separator spellings and twelve `-std`/language modes verified in the finding-L evidence. A backslash that is not part of a splice is preserved byte for byte. |
| Phase 3 — comments | Reproduced exactly | Block and line comments are white space; a comment inside a directive does not end it; a comment spanning lines restores line-start. An unterminated comment consumes to EOF, which is inert because the translation unit does not compile. |
| Phase 3 — literals | Reproduced for plain literals; **over-approximates toward rejection otherwise** | `literalEnd` never crosses a newline, so a lone delimiter is one ordinary byte and the rest of the line is still scanned. C++ raw strings are not modeled: the scanner may treat raw-string bytes as content to scan, which can only add a rejection, never an admission. C++ line-splice reversal inside raw strings diverges in the same safe direction. |
| Escape decoding | **Removed** | Its only consumer was the deleted assembler-template classifier. No admitted decision now depends on reproducing a C escape sequence, which retires the entire round-5/round-8 lexical parity surface. |

## Move 2 — the channel axis is reject-by-default

`portable mode admits allowlist X and rejects all else; therefore an unknown
channel cannot be admitted.`

### Allowlist X — the complete admitted read surface

| Admitted form | Treatment |
| --- | --- |
| `#include` / `#import` / `#include_next` with an **exact literal** operand | Resolved against the target's declared search roots, confined to admitted source or exactly one selected binding root, and joined to the transitive include worklist. |
| `@import NAME;` at any column | Resolved as a module reference against admitted module maps and selected external modules. |
| `#pragma clang module import NAME`, and its `_Pragma`/`__pragma` operator forms | Same module confinement. Admitted because it is the only module-import spelling available in a plain C translation unit, where `@import` is a syntax error. |
| Pragmas whose head is in the closed allowlist | Content. `once`, `mark`, `pack`, `push_macro`, `pop_macro`, `unused`, `weak`, `message`, `warning`, `region`, `endregion`, `STDC`, `options`; `clang` + {`diagnostic`, `attribute`, `assume_nonnull`, `system_header`, `arc_cf_code_audited`, `deprecated`, `fp`, `loop`, `unroll`, `optimize`, `section`, `final`, `max_tokens_here`, `max_tokens_total`, `restrict_expansion`}; `GCC` + {`diagnostic`, `visibility`, `poison`, `system_header`, `warning`, `error`, `push_options`, `pop_options`, `optimize`, `target`, `unroll`, `ivdep`, `novector`}. An empty pragma body is content. |
| Directives in `classifiableDirectives` | Content, with the body re-scanned for the token-level `@import` / `_Pragma` / `__pragma` / assembly channels that survive macro expansion. |
| Normal module maps | Parsed, confined, transitively seeded into the include worklist; the out-of-root guard and the module-map-outside-the-public-header-root guard are unchanged. |
| Typed interop boundaries | Unchanged. |

`__has_include` and `__has_embed` operands are scanned but not evaluated. They
are existence oracles: they can report whether a path exists, but they cannot
introduce a single byte into the translation unit, so they are not a read
channel and stay admitted knowingly rather than by omission.

### Everything else — rejected

| Rejected form | Code | Note |
| --- | --- | --- |
| Any `asm` / `__asm` / `__asm__` at a token boundary | `swiftpm_target_platform_unsupported` | The construct is rejected, not the spelling. See below. |
| `#embed`, in every operand form | `swiftpm_header_input_undeclared` | C23 resource inclusion. It really does paste bytes on the accepted profile, its operand grammar carries parameters this stage does not model, and no admitted SwiftPM C-family shape uses it. |
| Any `#`-introduced line the grammar cannot classify | `swiftpm_header_input_undeclared` | Includes computed/macro operands and any residual after phases 1-3. |
| Any pragma, `_Pragma`, or `__pragma` body outside the allowlist | `swiftpm_header_input_undeclared` | Covers `comment(lib, …)`, `include_alias`, `GCC dependency`, every unenumerated vendor spelling, and every `clang module` spelling other than `import`. |
| A `_Pragma`/`__pragma` operand this grammar cannot read | `swiftpm_header_input_undeclared` | Encoding-prefixed, raw, or macro operands; unbalanced parentheses. |
| `.s` / `.S` sources in a C-family target | `swiftpm_target_platform_unsupported` | Unchanged accepted behavior. |
| Any trigraph sequence | `swiftpm_header_input_undeclared` | Unchanged accepted behavior. |
| A module-map reference the confinement cannot resolve | `swiftpm_modulemap_escape` | Unchanged accepted behavior. |

### The assembler classifier is deleted, not demoted

`asmQualifiers`, `assemblerChannelDirectives`, `asmVerdict`, `classifyAssembly`,
`asmTemplateText`, `decodeStringLiteral`, `decodeDelimitedEscape`, and
`hexDigit` are removed together with their 62-case unit suite.

Keeping them as a "belt-and-braces secondary check" would have added exactly
zero assurance: they were reachable only *after* `startsAsmStatement` matched,
and that match is now itself the rejection. Retaining unreachable code whose
whole purpose was a parity claim this round abandons would misrepresent where
the proof lives.

The rejection is the contract:

> Portable mode admits no assembler stage. The integrated assembler runs inside
> the same `clang -c` invocation, shares no token with the preprocessor grammar,
> and `-H` reports none of its reads — so an observed-read provider is no
> backstop for it either in portable mode. `.incbin` pastes a file's bytes,
> `.include` assembles a file's text, `.linker_option` makes the linker load an
> undeclared library, `.secure_log_unique` writes a file, and
> `.macro`/`.irp`/`.irpc` parameter substitution can build any of those names
> from fragments before the assembler looks a directive up. All verified on the
> accepted Darwin profile in rounds 6-8. Adversarially complete *acceptance* of
> inline assembly is deferred to the observed-read provider in
> `TASK-260811-2qfnai`.

The cost is that a C-family target using any inline assembly is now rejected,
including ordinary `__asm__("nop")` and `extern int x __asm__("_sym")` label
spellings. That is a deliberate narrowing in the safe direction, recorded in the
task scope. No admitted SwiftPM source-only shape needs it.

`swiftpm_target_platform_unsupported` was chosen over a new
`swiftpm_inline_assembly_unsupported` because the accepted architecture decision
fixes a closed diagnostic vocabulary, and this is the same cause the `.s`/`.S`
source rejection already carries: a source shape with no accepted profile.

## Vectors

| Suite | Cases | What it pins |
| --- | ---: | --- |
| `TestH18InlineAssemblyRejectsTheTarget` | 48 | Every assembler spelling rounds 5-8 had to name individually — `.incbin`, `.include`, `.dump`, `.load`, `.linker_option`, `.secure_log_unique`, the seven macro-expansion directives, the `\x`/`\056`/`\x{2e}`/`\o{56}`/`\u{2e}`/`\N{…}` escape forms, literal concatenation, wide and raw templates, macro operands, aliased and macro-built keywords, the bare-`asm` identifier — plus the four finding-L splice separators, the `.s`/`.S` source forms, and the four bodies that were the round-7/8 *positive* control and now reject. Retained verbatim so the pivot is proved to lose no coverage. |
| `TestH19LineSplicesAreReproducedBeforeRecognition` | 13 | The finding-L shapes: `__as\`+ws+nl+`m__`, `_Pra\`+ws+nl+`gma`, `__prag\`+ws+nl+`ma`, `@imp\`+ws+nl+`ort`, and a spliced `#include` naming an escaping path — all reject. Positive controls: an `#include` split mid-name, split before its operand, and split *inside its literal* all resolve to the declared header, and a bare splice in a macro body is content. |
| `TestH20PragmaChannelsAreRejectByDefault` | 9 | `comment(lib, …)`, an unenumerated vendor pragma, `omp`, an unknown `clang` body, an unknown `GCC` body, the `_Pragma` and `__pragma` forms of the first, and an unenumerated pragma hidden in a `#define`. |
| `TestTranslationPhaseTwoSplicing` | 18 | Phase 2 directly: six separator spellings, CRLF and CR, consecutive splices, a keyword split, and the four non-splice forms a backslash can take (before text, before a space, trailing, escaped pair). |
| `TestPragmaAllowlistIsClosed` | 24 | The allowlist as a closed set: nine content bodies, twelve rejections, two resolved module imports. |
| `TestH17CompilerFileReadingChannelsAreClosed` | 30 | Repointed: `#embed` now rejects in every form including a **wholly contained** operand, which was a positive control before the pivot. The pragma positive control drops `comment(lib, …)` and gains `pop_macro`, `unused`, `STDC`, `clang assume_nonnull`, `GCC diagnostic`, and an empty pragma. |

Preserved positive path, unchanged and green: `S02`, `S03` (one Clang target
carrying all four C-family languages), `S05`/`S06` (C++ interop opt-in,
including the conditional selection-neutral case), `S07`/`S08` (Objective-C and
Objective-C++), `S10` (case-sensitive `.C`/`.M`), `H01`-`H16`, `R*`, `P*`,
`CGP05` including its conditional branch, `CGN*`, the cross-package include
closure, and the `no os/exec in swiftpminterop` seam guard.

Honest note on totals: the package went from 362 to **341 passing assertions**.
The decrease is the deleted 62-case classifier/escape-parity suite, which pinned
production code that no longer exists; 64 new cases replace it and `H18` grew.
Counting a test of deleted code as coverage would be the dishonest option.

## Gates

Every command was run as a standalone process; the exit code reported is the
real one.

| Gate | Command | Exit | Result |
| --- | --- | ---: | --- |
| Build | `go build ./...` | 0 | clean |
| Vet | `go vet ./...` | 0 | clean |
| Format | `gofmt -l internal/ cmd/` | 0 | no output |
| Focused | `go test -count=1 ./internal/swiftpminterop/` | 0 | ok, 7.4s |
| Coverage | `go test -count=1 -cover ./internal/swiftpminterop/` | 0 | 87.6% of statements |
| Verbose matrix | `go test -count=1 -v ./internal/swiftpminterop/` | 0 | 341 PASS / 0 FAIL, 79 top-level |
| Race | `go test -race -count=1 ./internal/swiftpminterop/` | 0 | ok, 49.9s |
| Lint (package) | `golangci-lint run ./internal/swiftpminterop/...` (v2.12.2) | 0 | 0 issues |
| Lint (repository) | `golangci-lint run ./...` (v2.12.2) | 0 | 0 issues |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2`; `canonical_references=pass` |
| Suite minus `cmd/curator` | `go test -count=1 $(go list ./... \| grep -v '/cmd/curator$')` | 0 | 51 ok, 2 no test files, 0 FAIL |
| Whitespace | `git diff --check` | 0 | clean |
| Board | `task-board --no-update-check validate` | 0 | `Board is valid. No issues found.` |

`cmd/curator` was **not** run in this round. The headless single-call cap is
10 minutes and that package takes roughly 8-10; `go list -deps ./cmd/curator`
has zero `swiftpminterop` matches, so it cannot be affected by this delta. It
remains the Orchestrator's gate.

Nothing was staged or committed.

## Supersessions

- Round-5 finding G1 ("`#embed` is an inclusion directive — literal operand
  resolved and confined … embedded file joins the transitive worklist") is
  superseded: `#embed` now rejects outright. Its escape, macro-operand, and
  parameter vectors are retained and still reject; its `contained-positive`
  vector is inverted to a rejection.
- Round-6 finding H, round-7 finding J, and round-8 finding K are superseded as
  *mechanisms*: their vectors are retained verbatim in `H18` and all still
  reject, but through the single construct-level rule rather than through
  per-directive and per-escape classification.
- Round-5's proofs that `#pragma comment(lib, …)` and `clang include_instead`
  are inert stand as evidence; under reject-by-default they are no longer
  load-bearing, and `comment` is rejected anyway because it names a library.

## Residual boundary, stated plainly

Portable mode's acceptance is now **narrower than the language**. A legitimate
C-family source that uses inline assembly, `#embed`, an OpenMP or vendor pragma,
or a computed include is rejected with a clear diagnostic before any process
starts. That is the intended trade: the channel surface is a closed allowlist
rather than an open emulation target, so the class of "a file the compiler reads
that the scanner does not see" is closed structurally instead of one spelling at
a time. Widening acceptance back out belongs to the observed-read provider in
`TASK-260811-2qfnai`, where the compiler itself reports what it opened.
