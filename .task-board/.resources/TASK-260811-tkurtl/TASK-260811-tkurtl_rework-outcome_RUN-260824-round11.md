# TASK-260811-tkurtl — rework outcome, round 11 (phase-4 residuals N1/N2/N3)

Role: developer. Model: Claude `claude-opus-5`.
Reviewed verdict answered: `TASK-260811-tkurtl_review-verdict_RUN-260824-d10094.md`.
Compiler for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
`arm64-apple-darwin25.5.0` — the accepted Darwin profile.

Scope: N1, N2, N3, and the re-derived step 4. No other change. Product code
touched: `internal/swiftpminterop/headers.go` and `internal/swiftpminterop/interop.go`;
tests: `internal/swiftpminterop/modulemap_test.go`. `closuregraph`,
`swiftpmsource`, `2qfnai`, and `x611eq` are untouched. Nothing staged or
committed.

`IncludeGrammarID` bumped `c-family-include-scanner-v9` -> `v10`.

---

## Compiler evidence first

Every disposition below was pinned before it was encoded. Full log:
`TASK-260811-tkurtl_round11-clang-evidence.log`. `SecretKit` is a module whose
only header is `#error SECRET_MODULE_WAS_READ`; `payload.bin` is a 12-byte
marker file.

| Probe | Source | Result |
| --- | --- | --- |
| N1a | `#define protocol import` + `@ protocol SecretKit;` | module built, `secret.h` read |
| N1b | `#define class import` + `@ class SecretKit;` | module built, `secret.h` read |
| N1f | `#define class im##port` + `@ class SecretKit;` | module built, `secret.h` read |
| N2 | `#define NoSuchKitXYZ SecretKit` + `@import NoSuchKitXYZ;` | **SecretKit** built and read |
| N2 control | `@import NoSuchKitXYZ;` alone | `module 'NoSuchKitXYZ' not found` |
| pragma control | `#define DeclaredKit SecretKit` + `#pragma clang module import DeclaredKit` | `module 'DeclaredKit' not found` — **not expanded** |
| `__pragma` | `#define AliasKit SecretKit` + `__pragma(clang module import AliasKit)`, `-fms-extensions` | **SecretKit** built and read — **expanded** |
| `#define import` | `#define import protocol` + `@import SecretKit;` | exit 0, **no import at all** |
| asm control | `__asm__(".incbin \"payload.bin\"");` | `.o` 520 B, payload bytes present |
| N3a | `#define A __as%:%:m__` + `A(…)` | `.o` **520 B, byte-size identical to control**, payload present |
| N3b | `#define J(a,b) a%:%:b` + `J(a,sm)(…)` | same |
| N3c | `#define A _Prag%:%:ma` + `A("clang module import SecretKit")` | module built and read |
| pragma head | `#define CL clang` + `#pragma CL module import SecretKit` | inert — `#pragma` lines are not expanded |
| pragma body | `#define once clang module import SecretKit` + `#pragma once` | inert — same |

Two dispositions here were **not** in the brief and change the fix:

- `__pragma(clang module import NAME)` **is** macro-expanded. The brief's
  asymmetry statement holds for `#pragma` and `_Pragma`; the Microsoft operator
  sits on `@import`'s side of it. It gets the same rejection.
- `#define import <anything>` makes `@import NAME;` import **nothing** while the
  scanner records a module import. Same evidence-integrity class as N2, reached
  from the opposite end, so `import` and `__experimental_modules_import` are in
  the N1 rejection set.

---

## N1 — the `@`-keyword allowlist is itself macro-definable

**Design choice: reject the `#define`, not the `@`.** Of the two options the
brief offered, "reject `@ IDENT` whenever IDENT is macro-defined at all" cannot
be decided where it is written: the realistic vector binds the macro in a header
and uses it in a `.c` file, and answering "is this identifier macro-defined?" at
the `@` needs the whole translation unit, which this scanner does not model
(it does not evaluate conditional inclusion). "Does this definition bind a name
the compiler expands after `@`?" is decidable from the definition alone, and it
is decided wherever the definition sits, because every admitted file of the
target's closure is scanned — including headers reached transitively.

`atPositionIdentifiers` (`headers.go`) is `objcAtKeywords` plus `import` and
`__experimental_modules_import`. `readMacroDefinition` rejects a `#define` that
binds any member, before any body analysis, so N1f (`#define class im##port`)
is rejected at the name and never depends on the paste layer.

`objcAtKeywords` keeps its job — `@ foo` for an unknown `foo` still rejects —
but it is now documented as a recognizer, not as the closure: the rule that
makes the position safe is the definition rejection.

**Narrowing, deliberate and recorded:** `#define interface struct` (the Windows
COM idiom) and a package-local `#define true 1` C89 shim are rejected. Neither
is needed by any admitted SwiftPM C-family shape, and both fail in the safe
direction. Every Objective-C construct `@` really introduces — `@interface`,
`@property`, `@end`, `@implementation`, `@synthesize`, `@selector`, `@protocol`,
`@class`, the literals and collections — is untouched and covered by positive
controls.

## N2 — the `@import` module name is macro-expanded

This is an evidence-integrity defect, not a keyword one: the `moduleDeclared`
gate is satisfied by a name the compiler never resolves, and the retained
closure evidence names the wrong module.

`IncludeReference` gains `ExpandedName`, set true for `@import NAME` and for
`__pragma(clang module import NAME)`, false for `#pragma clang module import
NAME` and for the destringized `_Pragma` operand. The asymmetry is encoded as a
comment at each of the four sites, with its probe, because the symmetry
assumption is the wrong one.

`scanIncludes` now returns a `scanResult` carrying the file's references **and**
the macro names it binds. `scanAndConfineIncludes` unions those across the
target's whole scanned closure and, after the worklist drains,
`rejectMacroDefinedModuleNames` rejects any `ExpandedName` module import whose
spelling — or any dot-component of it — is macro-defined. Portable mode does not
expand the name; it refuses the construct. A literal, non-macro `@import` is
every legitimate use and still admits and is still recorded.

The macro set is unioned per target rather than per translation unit, which
over-approximates in the safe direction: at worst a literal import is rejected
because an unrelated file of the same target binds a macro of that name.

## N3 — `%:%:` is the digraph spelling of `##`

`macroPasteWidth` reads both spellings and `collapseMacroPastes` short-circuits
on the absence of both, so a digraph-spelled paste is collapsed, and its
parameter form rejected, identically to `##`. `readDirective` already read `%:`
as the digraph for `#`, which is why the omission was a hole rather than an
unmodeled feature. Digraphs, unlike trigraphs, are unconditional in every mode
this profile admits, so nothing has to be bound per file.

---

## Step 4, re-derived over preprocessing tokens

Stated in `scanIncludes`' doc comment. Over tokens **including alternative
spellings**, because a spelling-level statement is not a statement about tokens:
`%:%:` and `##` are one operator and a rule naming only one leaves the other
open.

The scanner reads source tokens, so expansion reaches the compiler in exactly
two ways the source does not spell.

1. **A new token built from fragments.** Adjacent tokens never merge (verified:
   `#define A __as` + `#define B m__` + `A B(…)` is an unknown type name and
   yields an object with no payload), so the only builder is the paste operator
   in either spelling. `collapseMacroPastes` performs a fixed-fragment paste so
   the result is scanned like any other token stream, and rejects a paste taking
   a fragment from the call site. Macro output is not re-scanned for directives
   (verified: `#define INC #include "…"` invoked as `INC` is a syntax error and
   reads nothing), so a built token can only reach a token-level channel, and
   every one of those rejects.
2. **An existing identifier the compiler expands in a position this scanner
   reads literally.** No keyword rule reaches this class. It is settled by
   enumerating every identifier position the scanner records or gates on:

| position | compiler expands it? | scanner disposition |
| --- | --- | --- |
| `#include`/`#import`/`#include_next` operand | no | literal `"…"`/`<…>` header name used as written; any other operand is the computed-include form and rejects (H09) |
| directive name after `#` or `%:` | no | classified literally against a closed set |
| `@`-follower identifier | **YES** | rejected at the `#define` (`atPositionIdentifiers`) |
| `@import` module name | **YES** | rejected when macro-defined anywhere in the scanned closure (`rejectMacroDefinedModuleNames`) |
| `#pragma` head and operands | no | recorded literally; verified for head and module name |
| `_Pragma` string operand and its destringized tokens | no | recorded literally; a non-literal operand rejects |
| `__pragma` tokens | **YES** | same rejection as `@import` |
| module-map module, header, and extern names | n/a | parsed from a module map, which no preprocessing phase touches |

Every remaining channel — inline assembly, `#embed`, a pragma outside the
allowlist, an unclassifiable `#`-line, a C++ raw string — rejects outright, so it
exposes no identifier position to expand into.

The premises the reviewer verified true are kept and unchanged: no directive
re-scan of macro output, no adjacent-token merge without a paste, `_Pragma`'s
string operand not expanded, pragma-import not expanded.

---

## Vectors — H23, 22 subtests

New family `TestH23MacroExpandedIdentifierPositionsReject` in `modulemap_test.go`.
Rejections: N1a `protocol`, N1b `class`, N1c `selector`, N1d `end`, N1e `YES`,
N1f paste-built keyword macro, N1g `#define import`, N1h
`#define __experimental_modules_import`, N1i keyword macro bound in an included
header (the cross-file vector the design choice exists for), N2 aliased
`@import CLib` (the fixture's own admitted module, so the `moduleDeclared` gate
really is satisfied), N2b the same through `__pragma`, N2c the qualified
`CLib.Sub` form, N3a-N3e the five digraph pastes.

Controls: the pragma-import spelling is not expanded and still admits (asserting
`ExpandedName == false` on both the `#pragma` and `_Pragma` forms); a literal
`@import CLib; @import Foundation;` still admits and records both (asserting
`ExpandedName == true`); an ordinary `#define`; an ordinary digraph paste
(`foo%:%:bar`); literal `@protocol`/`@class` after `@`.

**Negative control run.** With the three fixes disabled in place (the
`atPositionIdentifiers` check short-circuited, the `##`-only paste short-circuit
restored, `ExpandedName` forced false), 17 of the 22 subtests FAIL — including
every N1, every N2, and N3a/N3b/N3c/N3e. The code was restored and re-verified
green. N3d passes either way, incidentally via the `@` rule, exactly as the
reviewer noted. The vectors bite.

---

## Evidence

Run directly as standalone processes; real exit codes.

| Gate | Command | Exit |
| --- | --- | ---: |
| Focused, verbose | `go test -count=1 -v ./internal/swiftpminterop/` | 0 — **392 PASS / 0 FAIL** (up from 369) |
| Focused, race | `go test -count=1 -race ./internal/swiftpminterop/` | 0 |
| Lint (pinned v2.12.2) | `golangci-lint run ./...` | 0 — `0 issues.` |
| Vet | `go vet ./...` | 0 |
| gofmt | `gofmt -l .` | no files listed |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 — `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2`, `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Suite minus `cmd/curator` (53 packages, one bounded call) | `go test -count=1 $(go list ./... \| grep -v '/cmd/curator$')` | 0 — 51 `ok`, 0 FAIL |
| Whitespace | `git diff --check` | 0 |
| Board | `task-board --no-update-check validate` | `Board is valid. No issues found.` |

`cmd/curator` was not run in this round: it takes ~10 min per run against a
10-minute per-call cap, and this change touches no package it exercises. The
monolithic full suite is the Orchestrator's gate.

Positive path preserved: all of `S02`, `S03`, `S05`-`S08`, `S10`, `H01`-`H22`,
`CGP05` including its conditional branch, `CGN*`, `R*`/`P*`, the transitive
worklist, the module-map out-of-root guard, and the cross-package include
closure are green and unchanged. Normal SwiftPM C/C++/ObjC/ObjC++ targets with
plain includes, imports, module maps, and typed boundaries still admit.
