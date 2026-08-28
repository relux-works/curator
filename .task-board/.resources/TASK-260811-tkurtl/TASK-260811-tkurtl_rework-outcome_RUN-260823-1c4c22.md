# Rework outcome — TASK-260811-tkurtl round 4 (findings D, E, F)

Run: `RUN-260823-1c4c22` verdict (round 3) -> this rework.
Model: Claude claude-opus-5. Scope: `internal/swiftpminterop/` only.
No staging, no commit, no reset, no clean.

Every round-1/2/3 accepted behavior is preserved: the transitive include
worklist, the stateful directive scanner and its closed classification,
`literalIncludeOperand`, the condition-neutral `Cxx` gate and selection
neutrality, closuregraph `Condition` canonicalization (53 golden records
untouched), `publicHeadersPath` handling, seam/guard discipline, portable
read-set honesty, and scope hygiene.

## Finding D — trigraph replacement (translation phase 1)

**Reviewer claim, corrected.** The verdict's table says `??=include` reads the
header under the *default* and `-std=gnu17` as well as `-std=c17`. Re-probed on
this host with an exact `-H` match, that is not what the pinned Apple Clang
21.0.0 does — trigraph replacement is **mode-dependent**:

| Mode | `??=include "secret.h"` |
| --- | --- |
| default (gnu) | ignored (`-Wtrigraphs: trigraph ignored`) |
| `-std=gnu17` | ignored |
| `-std=c89` / `c99` / `c11` / `c17` | **replaced — reads the header** |
| `-std=c++14` | **replaced — reads the header** |
| `-std=c++17`, Objective-C++ default | ignored |

So finding D is real (a target declaring `cLanguageStandard: .c17` performs an
inclusion that was invisible), but "always translate" would have opened a new
hole in the other direction, also confirmed against the compiler:

```
int a;??/
#include "secret.h"      -> READS under the GNU default (??/ is not a splice)
                         -> would be spliced away under an ISO mode
```

**Fix.** `scanIncludes` (`headers.go`) rejects any source containing one of the
nine trigraph sequences with `swiftpm_header_input_undeclared` and the exact
sequence in the diagnostic. The replacement cannot be bound per file — one
target may compile C and Objective-C sources under one declared standard and
share headers with C++ translation units under another — so a trigraph is
unsupported rather than classified under an assumed mode. `findTrigraph` scans
leftmost, so `???=` reports the pair the compiler would replace.

The reviewer's correct-by-accident case `#inc??/`⏎`lude` still rejects, now on
the trigraph rule rather than the unknown-directive backstop.

## Finding E — BOM and non-ASCII white space before `#`

Confirmed per code point with `clang -fsyntax-only -H` (exact match on the `-H`
output line, not the echoed source):

| Prefix before `#include "secret.h"` | Reads |
| --- | --- |
| UTF-8 BOM at file start | yes |
| U+0085, U+00A0, U+1680, U+2000, U+200A, U+2028, U+2029, U+202F, U+205F, U+3000 | yes |
| embedded NUL | yes |
| `/* */` then U+00A0 | yes |
| UTF-8 BOM **mid-file** | **no** — clang lexes it as an identifier character (the verdict recorded "yes"; re-probed in `.c` and `.mm`, both error out) |
| U+200B, a non-ASCII identifier byte | no |
| `#` U+00A0 `include` | yes (rejected here as an unclassifiable directive) |

**Fix.** `directiveScanner.run` gains a line-start case for NUL and any byte
`>= 0x80`; `readLineStartPrefix` consumes the leading run and:

- keeps `atLineStart` across the BOM, NUL, and Clang's Unicode white-space set
  (`lineStartWhiteSpace`);
- rejects the line when the run contains a byte it cannot classify as white
  space **and** a `#` or `%:` directive follows it — the closed-grammar rule,
  rather than silently demoting the line;
- demotes the line normally when no directive follows, so an ordinary non-ASCII
  identifier at the start of a line is not a false rejection.

Mid-file BOM is treated as white space here even though this clang does not:
that over-approximates toward rejection, never toward admission.

## Finding F — module-map headers outside the public root

Confirmed: `clang -fsyntax-only -fmodules -fmodule-map-file=… use.c` reads
`include/../hidden.h` and its `#include </etc/passwd>` while building the
module; a two-hop chain (`hidden.h` -> `deeper.h`) was also read.

**Fix, two parts.**

1. `moduleMapSeeds` (`interop.go`) seeds the include worklist from every
   `ModuleMapEvidence.ResolvedRefs` entry with `Class == ResolvedAdmitted`, so
   module members join the same fixpoint scan. A reference resolving to a
   directory (`umbrella "."`, `umbrella "../shared"`) seeds every header below
   it; an extern module map is not scanned with the C grammar; a resolution
   that no longer exists fails closed.
2. `confineModuleMapClosure` replaces the flat `confineModuleMapReferences`:
   an admitted `extern module` map is now parsed, link-confined, and confined
   recursively (cycle-guarded), and its resolutions join the same
   `ResolvedRefs` list. Without this, an extern map was admitted by path and
   never opened — the same hole one level deeper.

## Class-closure self-check — three more holes found and fixed

Enumerating the compiler's phases and every permissive default left in the
scanner, then probing each against the real compiler with `-fmodules
-fimplicit-module-maps` against a module whose only header is an `#error`
marker (control: plain `@import Secret;` **reads**):

| Probe | Compiler | Scanner before | Action |
| --- | --- | --- | --- |
| `@ import Secret;` | READS | missed | fixed |
| `@/*c*/import Secret;` | READS | missed | fixed |
| `@`⏎`import Secret;` | READS | missed | fixed |
| `@`U+00A0`import Secret;` | READS | missed | fixed |
| `@import /*c*/ Secret;` | READS | rejected | now scanned exactly |
| `@import Secret`⏎`;` | READS | rejected | now scanned exactly |
| `int x = 1'0; @import Secret;` (C++14 digit separator) | READS | missed | fixed |
| `x@import Secret;` | READS | found | unchanged |
| `// @import Secret;` under `-std=c89` | no | skipped | correct |
| `@import` / `#include` (UCN) | no | content | correct |
| raw string `R"(`⏎`#include …)"` | no | over-scans | fail-closed direction |
| `#include <a//b.h>` | reads `a//b.h` | rejects | fail-closed |

Fixes: `startsModuleImport`/`readModuleImport` now recognize `@import` at the
token level through `skipTrivia` (white space, line breaks, comments,
non-ASCII white space between `@`, `import`, the dotted name, and `;`), and
`literalEnd` treats a quote with no partner on its logical line as one ordinary
byte instead of consuming to end of line, so a digit separator can no longer
swallow the rest of the line.

The two unbalanced-quote spellings never reach the scanner in practice — the
shared recursive artifact classifier rejects the source as
`artifact_opaque_dependency_forbidden` first. Both layers now reject
independently; the vectors assert the layer that actually fires.

Remaining permissive defaults audited and left as-is because each errs closed:
`#if 0` regions are scanned anyway; an unresolvable include operand, an
unknown `#`-introduced directive, an unclassifiable `@import` name, a header
name containing `//`, and an unopenable queued unit all reject.

## New conformance vectors

- `H12` +4 subtests: module-map header outside the public root; the same escape
  two hops away; an extern module map's header; an umbrella-directory member;
  plus a positive "contained module-map closure is recorded" case. The in-root
  control is preserved.
- `H14` (new, 6+1): `??=include`, `#inc??/`⏎`lude`, the `??/` splice that would
  hide a directive, a trigraph naming an admitted header, a trigraph inside a
  string literal, a trigraph in a transitively included header, and a control
  proving `??x` is not a trigraph.
- `H15` (new, 16+3): leading BOM, mid-file BOM, and each verified Unicode
  white-space code point before an escaping directive; a BOM after ASCII space;
  a BOM in a transitively included header; U+00A0 inside the directive; U+200B
  and a non-ASCII identifier byte as unclassifiable-before-a-directive
  rejections; NUL/U+0085/invalid-UTF-8 asserted against the artifact policy
  that dominates them; and three positive controls (BOM before an admitted
  include, a non-ASCII identifier introducing no directive, non-ASCII white
  space mid-line).
- `H16` (new, 7+2+3): the module-import separator spellings above, the two
  policy-dominated quote spellings, and three positive controls (`@"…"` is not
  an import, balanced digit separators, `@interface`/`@end`).

## Evidence

| Gate | Result |
| --- | --- |
| `gofmt -l ./internal ./cmd` | empty, exit 0 |
| `go vet` (4 packages) | exit 0 |
| `go test -count=1 -cover` (4 packages) | exit 0; `swiftpminterop` 86.4% |
| `go test -race -count=1` (3 packages) | exit 0 |
| `golangci-lint run` v2.12.2 (3 package trees) | `0 issues.`, exit 0 |
| `ruby …canonical-golden-verifier.rb` | both lines pass, 53 labeled records |
| `go test -count=1` over 53 packages (suite minus `cmd/curator`) | exit 0, 51 ok, 2 without tests |
| `git diff --check` | exit 0 |
| `task-board validate` | `Board is valid. No issues found.` |

Test matrix for `internal/swiftpminterop`: **69** top-level, **188** including
subtests (was 66/138). Coverage moved 86.6% -> 86.4%: the new NUL and
invalid-UTF-8 branches are unreachable through `close()` because the artifact
policy rejects those bytes first.

`cmd/curator` was not run: `go list -deps ./cmd/curator` does not reach
`swiftpminterop`, and the monolithic full suite is the Orchestrator's gate.
