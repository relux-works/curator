# TASK-260811-tkurtl rework outcome — phase-4 closure

Run: `RUN-260824-7b8a44`
Role: developer
Model: Claude `claude-opus-5`
Date: 2026-08-24
Compiler for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
`arm64-apple-darwin25.5.0` — the accepted Darwin profile.

## Scope

Exactly the two items reviewer `RUN-260824-ed3a24` routed back: blocking
finding M (macro-reconstituted channel keywords, translation phase 4) and the
secondary C++ raw-string observation. Nothing else. The reject-by-default
posture accepted last round is applied unchanged — the fix rejects the
constructs that can deliver a channel keyword past the scanner; it does not
build a macro expander.

Files touched this round: `internal/swiftpminterop/headers.go` and
`internal/swiftpminterop/modulemap_test.go`. `IncludeGrammarID` is bumped
`c-family-include-scanner-v8` -> `v9`.

## Finding M — the two narrowings

The closed channel-keyword set is `asm`, `__asm`, `__asm__`, `_Pragma`,
`__pragma`, `import`. Phases 1-3 close every *lexical* reconstitution of those
spellings. Phase 4 — macro expansion — runs after the scanner has read the
file, and it was still open. Reproduced independently on the pinned compiler
before any code changed:

| Probe | Source | Compiler result |
| --- | --- | --- |
| control | `__asm__(".incbin \"payload.bin\"");` | `.o` 544 B, payload bytes present |
| M1 | `#define J(a,b) a##b` + `J(a,sm)(…)` | `.o` byte-size identical to control, payload present |
| M2 | `#define J(a,b) a##b` + `J(__as,m__)(…)` | same |
| M3 | `#define J(a,b) a##b` + `J(_Prag,ma)("clang module import SecretKit")` | module built, `secret.h` read (`#error SECRET_MODULE_WAS_READ` fires) |
| M4 | `#define I import` + `@ I SecretKit;` | module built, header read |
| M5 (found here, not in the verdict) | `#define AT @` + `AT import SecretKit;` | module built, header read |
| fixed object-like paste | `#define A __as##m__` + `A(…)` | payload present |
| fixed paste chain | `#define A a##s##m` + `A(…)` | payload present |
| parameter + fixed fragment | `#define S(a) __as##a` + `S(m__)(…)` | payload present |
| variadic parameter paste | `#define V(...) as##__VA_ARGS__` + `V(m)(…)` | payload present |
| fixed paste to `_Pragma` | `#define A _Prag##ma` + `A("clang module import SecretKit")` | module built, header read |
| fixed paste to `@import` | `#define A @im##port` + `A SecretKit;` | module built, header read |

M5 is a hole the verdict did not name and this round found while enumerating
the `@` axis: the `@` itself, not only the identifier after it, can arrive from
a macro. It is closed by the same rule.

### Narrowing 1 — `##` paste

`readMacroDefinition` classifies every `#define` before the body is scanned:

- **A paste that joins a parameter to another identifier-shaped token
  rejects** (`swiftpm_header_input_undeclared`). At least one fragment comes
  from the call site, so no body-local analysis can bound the result. This is
  a real narrowing of accepted portable input — parameter pasting is an
  ordinary C idiom.
- **A paste of fixed fragments is performed, then scanned.** The `##` and the
  white space around it are deleted and the resulting token stream is scanned
  like any other, so `#define A __as##m__` rejects on the `__asm__` it builds
  (`swiftpm_target_platform_unsupported`, the inline-assembly rule) and
  `#define A _Prag##ma` rejects on the `_Pragma` it builds.

The GNU `, ## __VA_ARGS__` comma-deletion idiom is deliberately preserved: its
left operand is a punctuator, which contributes no identifier characters, so
the paste cannot build an identifier keyword and the argument itself remains a
whole token visible at the call site. `#define LOG(fmt, ...) printf(fmt,
##__VA_ARGS__)` compiles on the pinned compiler and still admits here.

**Recorded narrowing — inverted positive vector.** `modulemap_test.go` pinned
`#define JOIN(a, b) a##b` as an admitted positive inside
`TestH17…/an ordinary define is not a pragma channel`. That spelling is now a
rejection. The vector was removed from the positive control (the rest of it —
`#define WIDTH 8`, `#define TEXT "_Pragma is only a word here"` — is unchanged
and still admits) and the parameter-paste form now appears in `H21` as a
rejection. This is the same deliberate inversion `#embed` received in round 5.

### Narrowing 2 — `@` followed by a non-`import` identifier

`readAtToken` classifies every `@` that `startsModuleImport` did not already
consume. Objective-C's `@`-keywords are a closed set (`objcAtKeywords`), and
everything else `@` may introduce is a literal or a collection: `@"…"`, `@'c'`,
`@[…]`, `@{…}`, `@(…)`, `@42`. An identifier outside that set is not valid
Objective-C on its own, so it is a macro, and a macro there expands to
`import`; it rejects. `@` at end of input, or before any other byte, rejects
for the same reason (that closes M5). `@__experimental_modules_import` —
Clang's older spelling of `@import` — is deliberately absent from the keyword
set and therefore rejects.

Cost: none for legitimate code. `@interface`/`@property`/`@end`/
`@implementation`/`@synthesize`/`@selector`, the boxed literals, and the
collection syntax are all in the positive controls and stay green; literal
`@import NAME;` is unchanged.

## Secondary — C++ raw strings are modeled as a rejection, and the parity row is corrected

The previous outcome's phase-3 parity row claimed the raw-string divergence
"can only add a rejection, never an admission". That was wrong and the reviewer
was right. `R"x(" /* )x"` hands the scanner an unmatched `"` followed by `/*`;
`skipBlockComment` then swallows the rest of the file while the compiler sees
no comment at all. Verified here: that prologue followed by
`__asm__(".incbin \"payload.bin\"");` compiles (`raw.o`, 824 B) with the named
file's bytes in the object.

Of the two options the brief offered, this round takes the stronger one: the
construct is **rejected in this grammar** rather than deferred to a defense in
another component. `rawStringPrefixAt` recognizes the five encoding prefixes
(`R`, `LR`, `uR`, `UR`, `u8R`) at a token boundary, in content, inside a
directive body, and inside a macro definition, and `rejectRawString` fails the
target with `swiftpm_header_input_undeclared`. The rationale is the trigraph
rationale: whether `R"` opens a raw string at all is language-mode dependent,
and this stage cannot bind a language mode per file.

`artifactpolicy` independently rejects several raw-string spellings as opaque,
but that ordering is now irrelevant to the proof: `H22` asserts the scanner's
own verdict directly through `scanIncludes` for five spellings, and asserts
end-to-end that an *ordinary* raw string — which `artifactpolicy` admits —
still rejects at this boundary. A future `artifactpolicy` relaxation cannot
silently open this.

**Corrected parity row** (replaces the phase-3 literals row in
`RUN-260824-f1733b`):

| Stage | Disposition | Evidence / residual |
| --- | --- | --- |
| Phase 3 — literals | Reproduced for plain literals; C++ raw strings **fail closed** | `literalEnd` never crosses a newline, so a lone delimiter is one ordinary byte and the rest of the line is still scanned. C++ raw strings are *not* a safe over-approximation: `R"x(" /* )x"` produced an unmatched `"` then a `/*` that swallowed the file tail while the compiler saw no comment, which was load-bearing, not conservative. The construct is therefore rejected outright, in content, in directive bodies, and in macro definitions. |

## Cross-layer closure argument

The claim to check is: *portable mode admits no construct that can deliver one
of the six channel keywords, or an inclusion operand, into the compiler without
this stage seeing it.*

1. **Line-ending normalization and phase 1.** `\r\n`/`\r` fold first. Trigraphs
   reject (mode-dependent). BOM/NUL/Unicode white space is reproduced for the
   verified code-point set and any unclassifiable sequence before a directive
   rejects.
2. **Phase 2.** `\` + `[ \t\v\f]*` + `\n` is spliced unconditionally and
   mode-independently, so no keyword can be split across a line. Accepted last
   round; unchanged.
3. **Phase 3.** Comments are white space and restore line-start correctly;
   plain literals are reproduced; raw strings now reject (above). No construct
   at this layer can hide or build a keyword.
4. **Phase 4 — macro expansion.** Expansion cannot *create* an identifier
   token: a macro's replacement list and its arguments are both token
   sequences that already exist in the source, and adjacent tokens never merge.
   The single exception is `##`. Therefore a channel keyword reaching the
   compiler must either (a) appear literally as a token somewhere in the
   source — in content, where `run()` matches it at a token boundary, or in a
   `#define` body, where `scanDirectiveChannels` matches it, both already
   pinned by `H18`'s `#define K asm` / `#define STMT(x) __asm__(x)` vectors —
   or (b) be built by `##`, which this round rejects when a fragment comes from
   a call site and performs-then-scans when the fragments are fixed. `@` is not
   an identifier and cannot be pasted into one, so the `@import` channel is
   covered by the separate `@`-rule, including the case where the `@` itself
   comes from a macro.
5. **The channel allowlist (accepted last round).** Every `#`-introduced line
   is either an inclusion directive with an exact literal operand, a pragma
   inside the closed allowlist, a directive in `classifiableDirectives` that
   names no file, or a rejection. `#embed`, inline assembly in any spelling,
   `.s`/`.S` sources, and every pragma or `_Pragma`/`__pragma` spelling outside
   the allowlist reject.
6. **Later compiler phases.** Phases 5-8 (character-set conversion, adjacent
   string-literal concatenation, translation of the token stream, linking)
   consume tokens that already exist; none of them opens a file named by source
   text. The two file-reading stages that do — the integrated assembler and the
   module builder — are reached only through the constructs enumerated above,
   and the assembler is rejected wholesale rather than modeled.

The class is therefore closed at every layer whose input is source text. What
remains outside the proof is unchanged and stated in the task scope: the
adversarially complete *acceptance* of these constructs is deferred to the
observed-read provider in `TASK-260811-2qfnai`. Portable mode's answer for all
of them is a rejection, not a guess.

## New vectors

`H21` — `TestH21MacroReconstitutedChannelKeywordsReject`, 14 rejections and 5
positive controls:

- M1/M2/M3 pasted `asm`, `__asm__`, `_Pragma`; M4 macro-`@`-import; M5
  macro-`@` before a literal `import`.
- fixed pastes that build `__asm__`, `_Pragma`, `@import`, and a fixed paste
  *chain* (`a##s##m`).
- a parameter pasted to a fixed fragment, and a variadic parameter pasted to an
  identifier.
- an unreadable function-like parameter list, asserted at both layers (the
  shared classifier rejects it as opaque first; `scanIncludes` rejects it on
  its own).
- `@` before an unknown keyword and before `@__experimental_modules_import`.
- positives: a fixed paste that builds an ordinary token, the GNU
  comma-deletion idiom, Objective-C literals/collections, Objective-C
  `@`-keywords, and all three literal module-import spellings resolving
  together.

`H22` — `TestH22RawStringLiteralsRejectTheTarget`, 6 rejections and 1 positive:
the comment-swallowing prologue, ordinary/wide/UTF-8 raw strings and one inside
a `#define` asserted directly against `scanIncludes`, an ordinary raw string
asserted end-to-end through `Close()`, and a token-boundary positive proving an
identifier that merely ends in `R` plus ordinary literal concatenation stay
content.

## Preserved

Every accepted positive path from rounds 1-8 is green: literal
include/import/module-import resolution, the transitive worklist, module-map
handling and the out-of-root guard, the condition-neutral `Cxx` gate,
selection neutrality and the `CGP05` conditional vector, `closuregraph`
`Condition` canonicalization (53 golden records), `publicHeadersPath`,
case-sensitive `.C`/`.M`, seam/guard discipline, portable honesty, and scope
hygiene. `S02`, `S03`, `S05`-`S08`, `S10`, `H01`-`H20`, `H10`/`H11`, `CGN*`,
`R*`/`P*`, and the cross-package include closure all still pass. Normal
legitimate SwiftPM C-family targets still admit.

Focused suite: **369 PASS / 0 FAIL**, up from 341. The 28 net-new entries are
`H21` (14 rejections + 5 positives) and `H22` (6 rejections + 1 positive) plus
their two parent tests; the inverted `JOIN` spelling lived inside an existing
subtest, so no subtest was removed.

## Evidence

Every command was run as a standalone process; the exit code reported is the
command's own.

| Gate | Command | Exit |
| --- | --- | ---: |
| Focused suite | `go test -count=1 ./internal/swiftpminterop/` | 0 |
| Focused verbose | `go test -count=1 -v ./internal/swiftpminterop/` | 0 (369 PASS, 0 FAIL) |
| Race | `go test -count=1 -race ./internal/swiftpminterop/` | 0 |
| Suite minus `cmd/curator` | `go test -count=1 $(go list ./... \| grep -v 'curator/cmd/curator$')` | 0 |
| Lint | `golangci-lint run ./...` (pinned v2.12.2) | 0 (`0 issues.`) |
| Format | `gofmt -l ./internal ./cmd` | 0, no output |
| Vet | `go vet ./...` | 0 |
| Build | `go build ./...` | 0 |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 (`labeled_records=53`, `cgp05_target_branches=2`, `cgp10_observation_branches=2`, `canonical_references=pass`) |
| Whitespace | `git diff --check` | 0 |
| Board | `task-board --no-update-check validate` | `Board is valid. No issues found.` |

`cmd/curator` was not run in this round's bounded calls; it is unaffected by
this change (no file outside `internal/swiftpminterop/` was touched) and the
monolithic full-suite run is the Orchestrator's gate.

Attached logs: `round9-focused`, `round9-verbose`, `round9-race`,
`round9-nocmd`, `round9-lint`, `round9-vet`, `round9-golden`, and
`round9-clang-evidence` (the compiler probe transcript behind every table row
above).

Nothing was staged or committed.
