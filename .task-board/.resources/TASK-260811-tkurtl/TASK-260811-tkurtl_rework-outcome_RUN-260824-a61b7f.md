# Rework outcome for TASK-260811-tkurtl — round 5

Scope: findings G1, G2, G3 of
`TASK-260811-tkurtl_review-verdict_RUN-260824-a61b7f.md`, plus the decisive
channel-axis closure that verdict made a precondition for acceptance. Findings
D, E, F and every earlier accepted behavior are preserved unchanged.

All product changes are inside `internal/swiftpminterop/headers.go` (grammar)
and one documentation block in `internal/swiftpminterop/interop.go`. No other
package changed; `internal/closuregraph/testdata` is unmodified. Nothing was
staged or committed.

## Finding G1 — `#embed` is a content read, not a benign directive — fixed

`embed` moved out of `classifiableDirectives` into `inclusionDirectives`. A
`#embed` operand is now resolved with `literalIncludeOperand`, confined by
`confineInclude` exactly like an `#include`, recorded on the reference as
`IncludeReference.Embed`, and returned to the caller so the embedded file joins
the transitive worklist — its bytes are translation-unit input.

Non-literal operands (`#embed SECRET`) reject through the same
`literalIncludeOperand` path that already rejects `#include MACRO`. C23
`#embed` parameters (`limit(1)`, `if_empty(…)`, …) are trailing tokens
`literalIncludeOperand` cannot prove, so they reject too.

Scanning the embedded file with the preprocessor grammar over-approximates
toward rejection — Clang does not preprocess embedded bytes — which is the safe
direction and costs nothing, because no admitted SwiftPM shape uses `#embed` at
all. That tradeoff is documented at `confineInclude`.

Compiler evidence re-derived on this host (Apple clang 21.0.0,
`clang-2100.1.1.101`, arm64-apple-darwin25.5.0):

| Probe | Result |
| --- | --- |
| `#embed "data.txt"`, default mode | exit 0, `-Wc23-extensions` warning only |
| `-H` on the same file | no extra read reported for a contained operand |

The reviewer's byte-count `_Static_assert` proof against `/etc/passwd` is
accepted as-is and was not re-run against a host file.

New vectors (`H17`): `embed of an absolute path`, `embed escaping the
package`, `embed with a macro operand`, `embed with a limit parameter`,
`embed in a header`, `embed reaches an escaping include` (proves the worklist
join — the rejection comes from a directive *inside* the embedded file), and
the positive `a contained embed resolves and is recorded`.

## Finding G2 — `#pragma clang module import` and the `_Pragma` form — fixed

Three recognition paths were added and all route through the same
`confineInclude` module path as `@import`:

1. `#pragma clang module import <dotted-name>` — classified in `readDirective`
   before `classifiableDirectives` is consulted.
2. `_Pragma("clang module import <dotted-name>")` — a new token-level channel
   recognized at a token boundary, with the operand destringized per the
   standard (`\"` → `"`, `\\` → `\`).
3. The Microsoft `__pragma(clang module import <dotted-name>)` spelling, whose
   operand is raw tokens inside a balanced parenthesis run.

Per the closed-grammar rule, a `clang module` spelling this grammar cannot
resolve rejects, and a `_Pragma`/`__pragma` operand it cannot read rejects.

Compiler evidence on this host, against a module whose only header is an
`#error` marker:

| Source | Result |
| --- | --- |
| `#pragma clang module import SecretKit` in `.c` | reads — `error: SECRET_MODULE_HEADER_READ` |
| `_Pragma("clang module import SecretKit")` in `.c` | reads — same error |
| `_Pragma(u8"clang module import SecretKit")` | reads |
| `#define M "clang module import SecretKit"` + `_Pragma(M)` | reads |
| `#define IMP _Pragma("clang module import SecretKit")` + `IMP` | reads |
| `#define DO(x) _Pragma(#x)` + `DO(clang module import SecretKit)` | reads |
| `__pragma(clang module import SecretKit)`, default mode | syntax error |
| same under `-fms-extensions` | reads |
| `#pragma clang module import SecretKit` without `-fmodules` | `module 'SecretKit' not found` |

The last four rows are findings of this round's own channel audit, not of the
verdict. Two consequences:

- An encoding-prefixed literal (`L"…"`, `u8"…"`) really does work, so the
  grammar's refusal to read it is necessary rather than cosmetic. It rejects.
- A `_Pragma` hidden inside a macro definition expands at a site no grammar
  short of a preprocessor can recognize, so the *definition* is where it has to
  be classified. `scanDirectiveChannels` re-scans every classified directive
  body for the `@import`, `_Pragma`, and `__pragma` token channels; the
  stringizing form `_Pragma(#x)` has no literal operand and rejects.

New vectors (`H17`): `pragma module import`, `_Pragma module import`,
`pragma module import no name`, `pragma module import trailing`,
`_Pragma with a wide literal`, `_Pragma with a u8 literal`,
`_Pragma with a macro operand`, `_Pragma without parentheses`,
`_Pragma inside a define`, `stringizing _Pragma define`,
`__pragma module import`, `__pragma include_alias`, the artifact-policy-
dominated `unbalanced __pragma operand` (which also asserts the scanner's own
verdict directly, because the shared classifier shadows it in a fixture), and
the positive `a pragma module import of a declared module resolves`.

## Finding G3 — inline module map via `#pragma clang module build` — fixed

`clang module build` / `endbuild` reject outright, as the reviewer directed. No
inline-map parsing machinery was built; the module-map authority stays in
`confineModuleMapClosure` alone. The same rejection covers `load`, `begin`,
`end`, and every other `clang module` spelling.

Compiler evidence on this host:

| Source | Result |
| --- | --- |
| `#pragma clang module build` + inline map naming `/tmp/...` + `endbuild` | inline map parsed; `error: header '/tmp/...' not found` at `SecretKit.map:1:27` |
| `#pragma clang module load "nonexistent.pcm"` | `fatal error: module 'nonexistent.pcm' not found` |
| `#pragma clang module begin SecretKit` | reaches module machinery: `must specify '-fmodule-name='SecretKit''` |

New vectors (`H17`): `pragma module build`, `pragma module endbuild alone`,
`pragma module load`, `pragma module begin`.

## Channel-axis enumeration

`classifiableDirectives` was re-derived from the question "can this directive
or pragma cause the compiler to read a file, or change where files are read
from?" Every member has one of three dispositions: **routed** through
resolution and confinement, **rejected**, or **provably cannot** with the
evidence recorded.

### Preprocessing directives

| Directive | Reads or redirects? | Disposition |
| --- | --- | --- |
| `#include`, `#include_next`, `#import` | yes | routed: exact literal operand resolved, confined, and enqueued; non-literal rejects |
| `#embed` | **yes** — verified, default GNU C mode | routed identically (G1) |
| `#pragma` | depends on the body | classified before the set is consulted; see the pragma table |
| `#define`, `#undef` | no file named | benign; body re-scanned for the `@import`/`_Pragma`/`__pragma` token channels |
| `#assert`, `#unassert` | no — obsolete GCC assertions take a predicate, not a path | benign; body re-scanned |
| `#if`, `#ifdef`, `#ifndef`, `#elif`, `#elifdef`, `#elifndef`, `#else`, `#endif` | no — `__has_include`/`__has_embed` are existence oracles that cannot introduce bytes | benign; operands scanned-not-evaluated by design; body re-scanned |
| `#error`, `#warning` | no — diagnostics only | benign; body re-scanned |
| `#ident`, `#sccs` | no — emit a string into the object | benign; body re-scanned |
| `#line` | no — changes only `__FILE__` and diagnostic position, never a search path | benign; body re-scanned |
| any other `#`-introduced line | unknown | rejects (pre-existing backstop, preserved) |

### Pragma space

| Pragma | Reads or redirects? | Disposition | Evidence |
| --- | --- | --- | --- |
| `clang module import NAME` | yes | routed as a module import | reads the module's headers; the **only** module-import spelling in plain C |
| `clang module build` / `endbuild` | yes — inline module map | rejects | inline map parsed and header resolution attempted |
| `clang module load` | yes — module file | rejects | `module '…' not found` |
| `clang module begin` / `end` | yes — module context | rejects | reaches module machinery |
| any other `clang module …` | unknown | rejects | closed-grammar rule |
| `include_alias("a","b")` | **yes** — substitutes the aliased file | rejects | aliasing a missing header onto an `#error` marker fires it under `-fms-extensions`; silently inert and unwarned without the flag |
| `GCC dependency "f"` | names a file a conforming implementation opens | rejects | this clang does not implement it (exit 0, `-H` reports no read), so no admitted shape depends on its effect |
| `comment(lib, "SecretLib")` | **no** — provably inert here | benign | compiling it emits **0** `LC_LINKER_OPTION` load commands and the name appears nowhere in the object |
| `clang include_instead("…")` | no — diagnostic suggestion only | benign | `-H` lists only the including header |
| `once`, `pack`, `mark`, `push_macro`/`pop_macro`, `clang diagnostic`, `clang attribute`, `clang system_header`, `GCC visibility`, all others | no file named | benign | macro state, diagnostics, or attributes only |

### Token-level channels (not `#`-introduced)

| Channel | Reads? | Disposition |
| --- | --- | --- |
| `@import NAME;` | yes | routed as a module import (round 4, preserved) |
| `_Pragma("…")` | yes | destringized and classified through the pragma table |
| `_Pragma` with a macro, encoding-prefixed, or raw operand, or without parentheses | yes (the macro form was proven to import) | rejects |
| `__pragma(tokens)` | yes under `-fms-extensions` | balanced-paren operand classified through the pragma table |
| `__pragma` with unbalanced parentheses | unknown | rejects |
| `_Pragma`/`__pragma`/`@import` inside any directive body | yes — macro-hidden expansion | classified at the definition |

Not a channel and unchanged, per the verdict's "Not a finding, recorded"
section: `__has_embed` is an existence oracle, and the on-disk module-map
reference grammar is complete.

## Vectors and validation

`IncludeGrammarID` is now `c-family-include-scanner-v4`.
`IncludeReference` gains an `Embed` field; it is not part of the interop
evidence digest, so no canonical record or golden changed.

`H17` adds 24 rejected spellings plus 6 further subtests (the artifact-policy-
dominated unbalanced `__pragma`, a contained embed, a declared module import
through both pragma spellings, a benign-pragma line covering ten spellings, an
ordinary `#define`, and an identifier-boundary case) — 30 subtests in all.
Package matrix: 70 top-level tests, 149 with subtests, 0 failures.

| Gate | Command | Exit |
| --- | --- | ---: |
| Format | `gofmt -l` over the four packages | 0, no output |
| Build | `go build ./...` | 0 |
| Vet | `go vet` over the four packages | 0 |
| Focused + cover | `go test -count=1 -cover` over `swiftpminterop`, `swiftpmsource`, `closuregraph`, `closureexec` | 0; interop coverage **87.0%** |
| Race | `go test -race -count=1` over the same four | 0 |
| Lint | `golangci-lint run` (v2.12.2) over the three package trees | 0, `0 issues.` |
| Canonical oracle | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0; `labeled_records=53`, both lines pass |
| Suite minus `cmd/curator` | `go test -count=1` over 53 packages | 0; 51 `ok`, 2 `[no test files]`, 0 `FAIL` |
| Whitespace | `git diff --check` | 0 |
| Board | `task-board validate` | `Board is valid. No issues found.` |

`cmd/curator` was not run in this round: the headless single-call cap is ten
minutes and that package takes roughly eight, and `go list -deps ./cmd/curator`
returns zero matches for `swiftpminterop`, so the round-5 delta cannot reach it.
It remains the Orchestrator's gate.

## Preserved

Findings D, E, F stay closed: `findTrigraph` and the trigraph rejection,
`readLineStartPrefix` and `lineStartWhiteSpace`, `moduleMapSeeds`,
`confineModuleMapClosure`, token-level `@import`, the `literalEnd` single-byte
fallback, and the `H12`/`H14`/`H15`/`H16` vectors are unchanged. Every named
family is green: `S02`–`S09` including both conditional-`Cxx` `S05` vectors and
the `swiftpm_target_platform_unsupported` control, `H01`–`H17`, `P01`–`P09`,
`CGN03`, `CGN09`, `CGN15`, and all three `CGP05` cases.

## Open, out of this round's scope

The multi-package interop coverage gap noted in round 3 and repeated in the
round-4 verdict is still open. The round-4 brief scoped this run to G1–G3 plus
the channel axis and said "no other scope", so it was not addressed here.
