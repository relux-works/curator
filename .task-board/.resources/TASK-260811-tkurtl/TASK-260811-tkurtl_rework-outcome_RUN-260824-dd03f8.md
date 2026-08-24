# Rework outcome for TASK-260811-tkurtl — round 6

Scope: findings **H** and **I** of
`TASK-260811-tkurtl_review-verdict_RUN-260824-dd03f8.md`, plus the
multi-package interop coverage gap the reviewer has now flagged in three
consecutive verdicts. Nothing accepted in rounds 1–5 was reopened.

Producer run: `RUN-260824-d648de` (Claude claude-opus-5). Compiler probes live
under `.temp/TASK-260811-tkurtl/probe-r6/`; refreshed logs under
`.temp/TASK-260811-tkurtl/logs/`. Nothing was staged or committed.

Pinned compiler for every probe below: Apple clang 21.0.0
(`clang-2100.1.1.101`), `arm64-apple-darwin25.5.0`, macOS 26.5, default mode,
no flags unless a row says otherwise. "Reads" is proved by the payload
appearing in the object *and* by the missing-file error — never by the absence
of a diagnostic.

## Finding H — the integrated assembler is a second file-reading stage

### What was wrong

`clang -c` runs one `-cc1` that contains a preprocessor **and** an integrated
assembler. Rounds 2–5 closed the preprocessor's directive and pragma space
exhaustively; the assembler shares no token with it. Two of its directives open
arbitrary files, and both were reachable from source Curator admits:

- `__asm__(".incbin \"/etc/passwd\"");` at file scope in a plain `.c` — the
  scanner consumed `__asm__` as an identifier and the parenthesised string as a
  literal, so no grammar engaged at all.
- a `.s`/`.S` source, admitted by `swiftPMSourceExtension`, skipped by
  `classifyTarget`, then scanned with the C preprocessor grammar, which models
  no assembler directive.

### Fix 1 — assembly statements are recognized and their templates classified

`headers.go` now recognizes `asm`, `__asm`, and `__asm__` at a token boundary,
exactly as `_Pragma`/`__pragma` already are (`startsAsmStatement`,
`asmKeyword`, `readAsmStatement`), in `run()` and in `scanDirectiveChannels`
(so a keyword hidden in a `#define` is classified at its definition). The
grammar is closed at three levels:

1. **The statement.** Only `asm` + optional C qualifier tokens + a balanced
   parenthesis run is readable. Anything else rejects — including bare `asm`
   with no template (see the self-check below).
2. **The template** (`asmTemplateText`). The template is the operand before the
   first top-level `:`; the output/input/clobber/label operands that follow are
   C expressions and register constraints that never reach the assembler as
   text. The template must be a run of plain adjacent `"…"` literals.
   Adjacent-literal concatenation is honoured; a macro operand, an
   encoding-prefixed literal (`L`, `u8`), a C++ raw string, an unterminated
   literal or comment, or any other token rejects.
3. **The escapes** (`decodeStringLiteral`). Full C escape decoding including
   `\xNN` and octal, because the assembler sees the *decoded* bytes.

`classifyAssembly` then answers the channel question on the decoded text. Every
`.`-introduced name is examined wherever it appears, not only at a statement
start: the assembler's statement separator and comment character are
target-dependent, and a `.macro` body can hide a spelling that only runs at
expansion. Over-approximating on position is free, because no name in the
rejected set is used by an instruction stream or a label reference.

### The assembler channel set, and why the operands reject rather than resolve

| Directive | Disposition | Diagnostic | Compiler evidence |
| --- | --- | --- | --- |
| `.incbin "F"` | reject | `swiftpm_header_input_undeclared` | exit 0; payload in the object at `0x188`; missing F → `error: Could not find incbin file`; `-H` reports **no read** |
| `.include "F"` | reject | `swiftpm_header_input_undeclared` | exit 0; missing F → `error: Could not find include file` |
| `.linker_option "-lNAME"` | reject | `artifact_toolchain_untrusted` | emits exactly **1** `LC_LINKER_OPTION`, where `#pragma comment(lib\|linker, …)` emits **0** |
| `.secure_log_unique "T"` | reject | `closure_write_undeclared` | with `AS_SECURE_LOG_FILE` set, appends the marker to that path |
| `.dump "F"`, `.load "F"` | reject | `swiftpm_header_input_undeclared` | `warning: ignoring directive … for now` — inert here, so no admitted shape depends on the effect; rejected because the spelling is a file channel on any assembler that implements it (same reasoning as `#pragma GCC dependency`) |
| `.file`, `.cv_file`, `.ident` | content | — | `.cv_file 1 "/etc/passwd"` exits 0 and **none** of that file's content appears in the object |
| `.align`, `.arch`, `.set`, `.section`, `.globl`, `.macro`, mnemonics | content | — | layout, symbols, and emitted bytes only |

**Decision, with reasoning (the brief asked for it explicitly): the
`.include`/`.incbin` operands reject rather than resolve-and-confine.** A
relative operand resolves against the *assembler process working directory*,
not the including file's directory — verified: the same source with the same
operand reads when compiled from the file's directory and reports `Could not
find incbin file` when compiled from its parent. The process working directory
is not a declared closure root, no header search path applies, and no
dependency output records the read, so there is no base this stage could
confine such an operand against. No admitted SwiftPM shape uses either
directive. Rejection is therefore the closed answer, and there is
correspondingly no "contained `.incbin` resolves" positive vector — unlike
`#embed`, where a package-relative operand does have a declared resolution
base.

### Fix 2 — `.s`/`.S` sources are unsupported, not partially inspected

`assemblySource` (`language.go`) plus a check in `classifyTarget`: a target
declaring an assembly source rejects with
`swiftpm_target_platform_unsupported`, condition-neutrally, before any header
or module analysis. **Choice and reasoning:** rejecting, not writing an
assembler grammar. No admitted SwiftPM shape in the accepted profile needs an
assembly source, and the current behaviour was unsound in *both* directions —
`.incbin "/etc/passwd"` in a `.S` file reads (verified, payload in the object)
while a lowercase `.s` file is not preprocessed at all, so an ordinary
`# comment` line in it would have been rejected as an unclassifiable directive.
A real assembler grammar would be a large closed-grammar surface bought for a
shape nothing needs.

Admission is deliberately unchanged: `swiftPMSourceExtension` still admits
`.s`/`.S`, so the bytes are still hashed and inventoried. The rejection is a
target-level verdict, not a reason to drop source from the closure.

### Stage-axis enumeration (closed)

`clang -### -c` shows one `-cc1` invocation with `-emit-obj`; there is no
separate `as` process and no linker under `-c`.

| Stage | File-reading channels | Disposition |
| --- | --- | --- |
| Driver | argv, response files, config files | Curator owns argv and the environment; no admitted source byte reaches it. `#pragma comment(lib\|linker, …)` re-verified inert this round: **0** `LC_LINKER_OPTION`. |
| Preprocessor | `#include`/`include_next`/`import`/`embed`, `@import`, `#pragma clang module import`, `_Pragma`, `__pragma`, macro-hidden forms, trigraphs, phase-2/3 translation, line-start byte sequences | Closed rounds 2–5; every disposition preserved and green (`H09`–`H17`). |
| Parse / Sema / CodeGen | C++20 module imports | Reviewer-verified not a channel on this compiler in any mode SwiftPM can select (`import "…";` → `error: unknown type name 'import'` under `-std=c++20` and `-std=c++20 -fmodules`; header units need `-fmodule-header`, which no admitted shape passes). Diagnostics re-open only a file already opened. |
| **Integrated assembler** | `.incbin`, `.include`, `.linker_option`, `.secure_log_unique`, `.dump`, `.load` | **Closed this round** — table above, `H18`. |
| Linker | — | Not run under `-c`. The one directive that reaches it, `.linker_option`, is rejected at the assembler stage. |

### Self-check: four further holes found and closed after the first fix

Probing my own grammar the way rounds 4 and 5 did surfaced four more real
channels; each was proven against the compiler before being closed.

| Spelling | Compiler | Grammar after the first fix | Now |
| --- | --- | --- | --- |
| `#define K asm` + `K(".incbin \"/etc/passwd\"");` | **reads** | bare `asm` with no `(` was left as content → the expansion site was invisible | **rejects** |
| `#define K2 __asm__` + `K2(…)` | reads | already rejected at the definition | rejects |
| `#define STMT(x) __asm__(x)` + `STMT(…)` | reads | already rejected (identifier operand) | rejects |
| `__asm__(\n#include "tpl.inc"\n);` | reads | already rejected (no `(` after the keyword) | rejects |
| `__asm__(R"(.incbin "/etc/passwd")")` in `.cpp` | reads | already rejected (raw literal) | rejects |
| `"\x2eincbin"` / `"\056incbin"` | reads | needed escape decoding | rejects |
| `".incbin " "\"/etc/passwd\""` | reads | needed literal concatenation | rejects |
| `".macro emb\n.incbin …\n.endm\nemb"` | reads | needed position-agnostic recognition | rejects |

The bare-`asm` decision changed as a result. It is now rejected under the same
rule the trigraph decision uses: the disposition depends on a language mode
this stage cannot bind per file (`asm` is a keyword in the GNU modes SwiftPM
selects by default and an ordinary identifier in strict ISO C), and the
permissive reading is the unsound one. The cost is a strict-ISO-C source using
`asm` as an identifier, which no admitted SwiftPM shape does and which does not
compile in the default mode.

### `H18` vectors

`TestH18IntegratedAssemblerChannelsAreClosed` — 28 subtests, all green.
Rejections: inline-asm `.incbin` absolute / escaping / relative, `.include`,
`.dump`, `.load`, `.incbin` in a header, macro operand, wide-literal template,
raw-string template in a C++ source, `__asm { … }` block, bare `__asm__ nop;`,
literal concatenation, `\x2e`- and `\056`-escaped directives, asm inside a
`#define`, `.incbin` inside an asm `.macro` body, `.linker_option`
(`artifact_toolchain_untrusted`), `.secure_log_unique`
(`closure_write_undeclared`), the four self-check alias/build forms, bare `asm`
used as an identifier, and both `.S` and `.s` assembly-source targets
(`swiftpm_target_platform_unsupported`).

Positive controls: an ordinary inline-assembly body (`.align`/`.arch`/`nop`,
`__asm__ __volatile__`, an `asm` symbol label, and extended asm with
constraints and a clobber list) stays content with exactly one recorded
include; an identifier that merely contains an assembly keyword stays content.

`TestAssemblyTemplateGrammarIsClosed` (`parser_test.go`) covers the template
grammar and escape decoder directly: 12 unreadable operand forms, 10 readable
ones, 12 escape decodings, and the classifier's boundary rule (a dotted label
reference, a suffixed identifier, and ordinary directives stay content;
leading, tab-indented, upper-case, and `;`-separated channel directives
reject).

**Decisiveness, measured.** With `startsAsmStatement` and `assemblySource`
stubbed to `false`, 20 of the 20 rejection subtests present at that point
failed and both positive controls still passed.

## Finding I — `.C` and `.M` classify as C and Objective-C

### What was wrong

`sourceLanguage` and `swiftpmsource.targetLanguages` lowercased the extension.
The driver is case-sensitive on exactly two suffixes — re-verified this round:

```
clang -### -c up.C  →  "-x" "c++"          clang -### -c lo.c  →  "-x" "c"
clang -### -c up.M  →  "-x" "objective-c++" clang -### -c lo.m  →  "-x" "objective-c"
```

So a provider implemented as `impl.C` reported `[c]`, `implementationCxx` stayed
false, the restricted-profile gate never saw a C++ target, the direct-C++
boundary was never bound, and the recorded `languages` evidence was wrong.

### Fix

Case-sensitive mapping where the compiler is case-sensitive, in both surfaces:
`.C` → `LanguageCXX` / `"c++"`, `.M` → `LanguageObjCXX` / `"objective-c++"`,
every other extension unchanged and still case-insensitive. Admission is
deliberately unchanged and now documented: `swiftPMSourceExtension` still
admits both, because the driver compiles them and their bytes are target source
either way. (The dead `".S"` case after `strings.ToLower` was removed; `.S` is
still admitted through `".s"`, and the test now asserts `.C`, `.M`, and `.S`
admission explicitly.)

### `S10` vectors

`TestS10UpperCaseExtensionsSelectTheCompilersLanguage` — 7 subtests: `.C`
records `c++`; `.M` records `objective-c++`; `.C` is gated by the C++ standard
profile; `.M` is gated by the Objective-C runtime profile; a `.C` provider with
the Swift opt-in binds `InteropCXX` with `c++` provider languages; the same
without an accepted profile is `swiftpm_target_platform_unsupported`; and the
S06 case — a Swift consumer declaring no `.interoperabilityMode(.Cxx)` against
a `.C` provider — is `closure_interop_undeclared`.

`swiftpmsource_test.go` asserts `targetLanguages({a.C, b.M}) == [c++,
objective-c++]` and the three admission suffixes.

**Decisiveness, measured.** With the case-sensitive branch removed, 4 of the 7
subtests failed. The other 3 pass either way, because for those shapes an
existing gate (the `.hpp` interface for the two C++ ones, the Objective-C
family for `.M`) already fired; they are kept because they are the vectors the
brief named, and the language assertion inside the first two is what makes the
set decisive.

## Multi-package interop coverage gap — closed

Flagged in rounds 3, 4, and 5 as read-verified but never executed, because
every fixture was single-package (`fakeBroker` rejects pins).

`fakeEvaluator` now answers per package identity, so a fixture can declare an
in-root `.package(path:)` dependency — the shape the accepted profile already
supports. `newMultiPackageFixture` builds a root package with a C target that
depends on a C target in a vendored package through a product edge. Three
vectors, all green:

- `TestCrossPackageIncludeClosureIsScanned` — the worklist really opens the
  vendor header under the vendor root, scans its directives, records the
  transitive reference, and binds two C-ABI boundaries.
- `TestCrossPackageIncludeEscapeFailsClosed` — an escaping include *inside* the
  vendor header fails closed, proving the cross-package file is genuinely
  scanned rather than merely queued.
- `TestCrossPackageIncludeKeepsThePerTargetSearchPath` — a vendor header that
  resolves only against the consumer's search roots fails closed, because the
  vendor package's own target compiles that header with only its own roots.
  Nothing widens a dependency's search path to its consumer's.

Executing it surfaced an evidence ambiguity: `IncludeReference.Package` names
the *consuming target's* package, while `Source` is relative to the package
that owns the scanned file. Once the worklist crosses a package boundary those
differ, and two packages may hold the same relative path. `IncludeReference`
gains an additive `SourcePackage` field naming the owning package, threaded
through `scanIncludes` and used in the record sort key. No field was removed
and no existing consumer changed — `Includes` is not read outside this package.

## Preserved

Everything accepted across all five verdicts. `inclusionDirectives` with
`embed`, `IncludeReference.Embed` and its worklist join, `classifyPragmaBody`,
`applyPragma`, `readPragmaOperator`, `balancedParenthesis`,
`destringizePragma`, `scanDirectiveChannels`, the `clang module`
`build`/`endbuild` rejection, the 24 rejected pragma spellings and the `H17`
controls, `H01`–`H17`, `S02`–`S09` including both conditional-`Cxx` `S05`
vectors and the `TestS05CxxInteropRequiresAcceptedDestinationProfile` control,
`P01`–`P09`, `CGN03`/`CGN09`/`CGN15`, and all three `CGP05` cases.

`IncludeGrammarID` moves `c-family-include-scanner-v4` →
`c-family-include-scanner-v5`. `closure_write_undeclared` is added to the
package's stable-code set from the accepted shared vocabulary.

`internal/swiftpminterop` now runs **302 PASS / 0 FAIL** (was 219/0). Coverage
**88.0%** (was 87.0%).

## Refreshed evidence

Every command run directly as a standalone process; real exit codes.

| Gate | Exit | Result |
| --- | ---: | --- |
| `gofmt -l internal/ cmd/` | 0 | no output |
| `go build ./...` | 0 | — |
| `go vet ./...` | 0 | — |
| `golangci-lint run` (three package trees, v2.12.2) | 0 | `0 issues.` |
| `go test -count=1 -cover` (4 packages) | 0 | interop **88.0%**, source 80.1%, graph 80.7%, exec 58.0% |
| `go test -race -count=1` (4 packages) | 0 | — |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | `labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2`; references pass |
| `go test -count=1 $(go list ./... \| grep -v cmd/curator)` | 0 | 51 `ok`, 2 `[no test files]`, 0 `FAIL` |
| `git diff --check` | 0 | — |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

The monolithic full suite including `cmd/curator` is the Orchestrator's gate
per the brief and was **not** run here.

Attached logs: `..._round6-focused.log`, `..._round6-race.log`,
`..._round6-lint.log`, `..._round6-golden.log`, `..._round6-nocmd.log`
(SHA-256 `d8dab245559f574200d0cf7edbe83dbda0e3cb2882aaed024fced9a17e50f354`),
`..._round6-verbose.log`.

## Files changed

| File | Change |
| --- | --- |
| `internal/swiftpminterop/headers.go` | assembly statement recognition, template grammar, escape decoding, channel classification; `SourcePackage`; grammar ID v5 |
| `internal/swiftpminterop/language.go` | case-sensitive `.C`/`.M`; `assemblySource` and the `classifyTarget` rejection |
| `internal/swiftpminterop/interop.go` | owning package threaded into `scanIncludes`; record sort key |
| `internal/swiftpminterop/errors.go` | `CodeWriteUndeclared` |
| `internal/swiftpmsource/graph.go` | case-sensitive `targetLanguages` |
| `internal/swiftpmsource/executor_runtime.go` | documented admission; dead `".S"` case removed |
| `internal/swiftpminterop/modulemap_test.go` | `H18` |
| `internal/swiftpminterop/language_test.go` | `S10` |
| `internal/swiftpminterop/parser_test.go` | template-grammar and escape vectors |
| `internal/swiftpminterop/conformance_test.go` | multi-package fixture and three cross-package vectors |
| `internal/swiftpminterop/fixture_test.go` | per-identity `fakeEvaluator` |
| `internal/swiftpmsource/swiftpmsource_test.go` | case-sensitivity and admission assertions |

Nothing was staged, committed, reset, or cleaned.
