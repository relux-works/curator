# Reviewer verdict for TASK-260811-tkurtl — round 4

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260824-a61b7f` (Claude claude-opus-5). `task-board spawn goal
  "$TASK_BOARD_RUN_ID"` reports no active goal, so this run is not goal-bound.
- Reviewed delivery: rework `RUN-260823-dc87cf` against findings D, E, F of
  `TASK-260811-tkurtl_review-verdict_RUN-260823-1c4c22.md`, plus the producer's
  own class-closure self-check.
- Reviewed outcome: `TASK-260811-tkurtl_rework-outcome_RUN-260823-1c4c22.md`.
- No code, staging, commit, reset, or clean was performed. One temporary probe
  file was created inside `internal/swiftpminterop/`, executed, and removed;
  `git status --short` is byte-identical to the state received (the only
  additions are this run's own board resources). Compiler probes live under
  `.temp/TASK-260811-tkurtl/probe/`.

## Summary

Findings D, E, and F are all **fixed** and reproduce as closed against their
exact round-3 probes. The fixes are the right ones: rejecting a trigraph rather
than translating it under an assumed mode is correct given the producer's
compiler re-probe (which corrects my round-3 table — replacement is
mode-dependent, and "always translate" would have opened the `??/` splice hole
in the other direction); the line-start prefix reader preserves line start
across the verified white-space set and rejects rather than demotes an
unclassifiable prefix; and `moduleMapSeeds` plus the recursive
`confineModuleMapClosure` close the seed hole at both levels. The producer's
own self-check found and fixed three further real holes (token-level `@import`,
the unterminated-literal state) that I re-verified.

Acceptance is blocked on scope item 4, the decisive one. The class is **not
closed**. Enumerating the file-reading channels the pinned Apple Clang has —
rather than the translation phases the scanner already models — surfaces three
more members of exactly the round-2/3/4 class: a read the compiler performs that
the declared closure never sees. All three are `#`-introduced lines that reach
`classifiableDirectives` (`headers.go:33-38`) and are dropped at
`headers.go:546` as benign, or are not `#`-introduced at all.

The most severe, `#embed`, needs no `-fmodules`, no ISO `-std`, and no exotic
spelling: it reads arbitrary file bytes in the **default** GNU C mode SwiftPM
selects when a target declares no `cLanguageStandard`, with only a
`-Wc23-extensions` warning.

## Per-item verdict

| # | Scope item (round-4 brief) | Verdict |
| ---: | --- | --- |
| 1 | Finding D closed — trigraphs | **accepted** |
| 2 | Finding E closed — BOM/NBSP | **accepted** |
| 3 | Finding F closed — module-map seeding | **accepted** |
| 4 | Class closure | **changes requested** (findings **G1**, **G2**, **G3**) |
| 5 | No regression of previously accepted behavior | **accepted** |
| 6 | Evidence | **accepted** |

## Findings

### Finding G1 — CONFIRMED — `#embed` reads an arbitrary file and is dropped as a benign directive

`"embed": true` sits in `classifiableDirectives` (`headers.go:35`), so
`readDirective` returns nil at `headers.go:546` with no reference recorded. But
`#embed` is not a classification directive — it is the C23 resource-inclusion
directive, and it pastes the named file's **bytes** into the translation unit.

Compiler evidence on this host (Apple clang 21.0.0, `clang-2100.1.1.101`), each
run in a plain `.c` file:

| Probe | Result |
| --- | --- |
| `#embed </etc/passwd>`, **default** mode | reads; `-Wc23-extensions` warning only, exit 0 |
| same under `-std=gnu17` | reads |
| `#embed "../../../../../../../../etc/passwd"` | reads |
| `-H` on the same file | lists `/etc/passwd` |

The read is proved by content, not by absence of an error:

```c
static const unsigned char d[] = {
#embed "../../../../../../../../etc/passwd"
};
_Static_assert(sizeof(d) == 1, "deliberately wrong: proves the real byte count");
```

```
error: static assertion failed due to requirement 'sizeof (d) == 1'
```

`/etc/passwd` is 9344 bytes on this host, and the assertion that its size is 1
is what fails — the compiler really pasted the file. The complementary
`_Static_assert(sizeof(d) > 100, ...)` passes.

Curator, executed probe (`Sources/CLib/lib.c`):

```
PROBE "embed absolute"   ACCEPTED includes=[CLib.h]
PROBE "embed quoted esc" ACCEPTED includes=[CLib.h]
```

The closure succeeds, `/etc/passwd` appears in no include set, and no
`swiftpm_header_input_undeclared` is raised. Under `-std=c++17` and
Objective-C++ the directive errors, but C and Objective-C targets are precisely
what a SwiftPM Clang target compiles, so the unsupported C++ modes narrow
nothing.

Required: treat `embed` as an inclusion directive — an exact literal operand
resolved and confined like `#include`, with a non-literal operand rejected by
`literalIncludeOperand` as it already is for `#include MACRO`. It must also be
added to the transitive worklist: an embedded file is a byte-for-byte input to
the translation unit. Add H-family vectors for the absolute and quoted-escaping
forms and for a contained positive.

### Finding G2 — CONFIRMED — `#pragma clang module import` is the C spelling of `@import` and is invisible

The scanner recognizes `@import` at the token level and treats an unresolvable
module as `swiftpm_header_input_undeclared` — correctly, because importing a
module reads its module map and headers. Clang's `#pragma` spelling of the same
operation is dropped: `pragma` is in `classifiableDirectives`
(`headers.go:36`), and `_Pragma("clang module import …")` is not a
`#`-introduced line at all, so no backstop engages.

Compiler evidence, against a module whose only header is an `#error` marker:

| Source | Result |
| --- | --- |
| `#pragma clang module import SecretKit` in `.c` | **reads** — `error: SECRET_MODULE_HEADER_READ` |
| `_Pragma("clang module import SecretKit")` in `.c` | **reads** — same error |
| `@import SecretKit;` in the **same `.c`** | does **not** import (`error: expected identifier or '('`) |

The last row is the point: the pragma form is strictly more reachable than the
form the scanner already covers. `@import` is Objective-C syntax, so in a plain
C translation unit the pragma is the *only* module-import spelling — and it is
the one that is invisible. Both forms need `-fmodules`, exactly the same
precondition as the `@import` the delivered design already treats as in scope,
so this is not a weaker reachability argument.

Curator, executed probes:

```
PROBE "pragma module imp"  ACCEPTED includes=[CLib.h]
PROBE "_Pragma module imp" ACCEPTED includes=[CLib.h]
```

Required: recognize `#pragma clang module import <dotted-name>` and the
`_Pragma` string-literal form as module imports, routed through the same
`confineInclude` module path as `@import`. A `clang module` pragma spelling this
grammar cannot resolve must reject, consistent with the closed-grammar rule the
scanner already applies to `@import`.

### Finding G3 — CONFIRMED — an inline `#pragma clang module build` declares a module map the confinement stage never parses

`confineModuleMapClosure` parses and confines on-disk module maps. Clang also
accepts a module map written **inline inside a C source**, between
`#pragma clang module build` and `#pragma clang module endbuild`. That inline
map can name an absolute header outside the package, and Clang reads it — the
same escape H03 rejects for an on-disk map, through a channel the module-map
stage never sees.

Compiler evidence:

```c
#pragma clang module build SecretKit
module SecretKit { header "/abs/path/outside/secret.h" }
#pragma clang module endbuild
#pragma clang module import SecretKit
```

```
While building module 'SecretKit' imported from mbuild2.c:1:
In file included from <module-includes>:1:
/…/probe/mod/secret.h:1:2: error: SECRET_MODULE_HEADER_READ
```

With the header spelled relatively and not on the search path, the same probe
reports `SecretKit.map:1:27: error: header 'secret.h' not found` — the inline
map was parsed and resolution was attempted either way.

Curator: the `build`/`endbuild` lines classify as `pragma` and drop, and the
`module SecretKit { … }` line between them is ordinary content. Nothing rejects.

Required: reject a `clang module build`/`endbuild` pragma outright. The
alternative — parsing the inline map with the module-map grammar and confining
it — is more machinery than the profile needs, and no admitted SwiftPM shape
requires an inline module map. A rejection is the proportional fail-closed
answer and keeps the module-map authority in one place.

### Not a finding, recorded

- `#pragma GCC dependency "/etc/passwd"` — executed, `exit 0`, and `-H` reports
  no read. This clang does not implement it. No action.
- `__has_embed(</etc/passwd>)` — evaluates true for a file outside the closure,
  so it is an existence oracle, not a content read. It cannot introduce bytes,
  and `#if` operands are already scanned-but-not-evaluated by design. Worth a
  note, not a change.
- The module-map grammar's reference set (`header`/`textual`/`private`/
  `exclude`, `umbrella header`, `umbrella directory`, `extern module`, plus
  `link`/`config_macros`/`requires`) is the complete set of file-naming
  constructs in Clang's module-map language. I found no missing on-disk
  reference kind — G3 is an inline channel, not a grammar gap.

## Accepted items — evidence

**Item 1 — finding D closed.** Reproduced with the round-3 probe verbatim:

```
PROBE-D code="swiftpm_header_input_undeclared"
        "source contains a trigraph sequence whose phase-1 replacement depends
         on a language mode this stage cannot bind per file"
```

The producer's correction of my round-3 table is right and I withdraw that
table: replacement is mode-dependent on this clang, and the `??/`-splice
counter-case means unconditional translation would have traded one hole for
another. Rejecting any of the nine trigraph sequences is the correct closed
answer, and `findTrigraph`'s leftmost scan reports the pair the compiler would
replace for `???=`. `H14`'s six rejections and the `??x` control are green.

**Item 2 — finding E closed.** All three round-3 probes now reject:

```
PROBE-E utf8 bom prefix   code="swiftpm_header_input_undeclared"
PROBE-E utf8 bom mid file code="swiftpm_header_input_undeclared"
PROBE-E nbsp prefix       code="swiftpm_header_input_undeclared"
```

`readLineStartPrefix` (`headers.go:385`) is the right shape: it preserves
`atLineStart` across `lineStartWhiteSpace`, and rejects only when an
unclassifiable run is followed by `#` or `%:` — so an ordinary non-ASCII
identifier opening a line is still content, not a false rejection. Treating a
mid-file BOM as white space where this clang does not is an over-approximation
toward rejection, which is the safe direction. `H15`'s 16 rejections, the three
artifact-policy-dominated byte classes, and the three positive controls are
green.

**Item 3 — finding F closed.** Reproduced with the round-3 probe verbatim:

```
PROBE-F code="swiftpm_header_input_undeclared"
```

`moduleMapSeeds` (`interop.go:628`) seeds from every `ResolvedAdmitted`
resolution, expands an umbrella directory to its members, and fails closed on a
resolution that no longer exists; `confineModuleMapClosure` (`interop.go:484`)
parses, link-confines, and recurses into an admitted `extern module` map under a
`visited` cycle guard, folding its resolutions into the same `ResolvedRefs`
list — so the seed set and the confinement set cannot drift apart. `H12`'s new
out-of-root, two-hop, extern, and umbrella-directory subtests, the preserved
in-root control, and both positive recorded-closure cases are green.

**Item 5 — no regression.** All six round-2 controls still reject, executed:

| Spelling | Code |
| --- | --- |
| spliced keyword `#inc\`⏎`lude` | `swiftpm_header_input_undeclared` |
| comment prefix `/* */ #include` | `swiftpm_header_input_undeclared` |
| form-feed prefix | `swiftpm_header_input_undeclared` |
| digraph `%:include` | `swiftpm_header_input_undeclared` |
| mid-line `@import SecretKit;` | `swiftpm_header_input_undeclared` |
| `#curator_secret` | `swiftpm_header_input_undeclared` |

Every named vector family is present and green in my own rerun: `S02`-`S09`
including both conditional-`Cxx` S05 vectors and the preserved
`swiftpm_target_platform_unsupported` control, `H01`-`H16`, `P01`-`P09`,
`CGN03`, `CGN09`, `CGN15`, and all three `CGP05` cases including the conditional
selection-neutrality vector. `internal/closuregraph/testdata` is unmodified and
the Ruby oracle reports `canonical_goldens=pass labeled_records=53
cgp05_target_branches=2 cgp10_observation_branches=2` and
`canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2
cgp10_all_refs_resolve=true`. Seam and guard discipline holds:
`grep -l exec.Command` over the three production trees returns exactly
`closureexec/acquisition.go` and `closureexec/portable_runner.go`, and the only
`os/exec` reference in `internal/swiftpminterop` is `guard_test.go`'s allowlist.
Scope hygiene holds: every file in the package is `.go`, and no Kotlin/Gradle
reference or creep into `TASK-260811-2qfnai`/`TASK-260811-x611eq` appears.

**Item 6 — evidence.** Every gate I reran reproduced:

| Gate | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go test -count=1 -cover` over `swiftpminterop`, `swiftpmsource`, `closuregraph`, `closureexec` | exit 0; interop coverage **86.4%** — matches the claim |
| `go test -race -count=1` over the three changed-or-adjacent packages | exit 0 |
| `golangci-lint run` over the three package trees | `0 issues.` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | both lines pass, 53 labeled records |
| `task-board validate` | `Board is valid. No issues found.` |
| Orchestrator `TASK-260811-tkurtl_full-go-04.log` | SHA-256 **5945be6629887b53d52abd701bd74120adccbfd0450d52135b1d781647aa4d18** matches the brief exactly; `EXIT:0`, **52 ok**, 0 `FAIL` |

Accepted rather than rerun: the monolithic full suite, on the hash-verified
orchestrator log above. It exceeds the 10-minute single-call cap, and the brief
directs reliance on the hash-bound log. My four focused packages plus that log's
remaining entries account for all 52 packages. The producer's claim that
`go list -deps ./cmd/curator` does not reach `swiftpminterop` is consistent with
the round-4 delta being confined to that package.

## Routing

`TASK-260811-tkurtl` -> `to-dev`. Findings D, E, and F are closed and should not
be reopened; keep `findTrigraph` and the trigraph rejection, `readLineStartPrefix`
and `lineStartWhiteSpace`, `moduleMapSeeds`, `confineModuleMapClosure`, the
token-level `@import` recognition, the `literalEnd` single-byte fallback, and the
`H12`/`H14`/`H15`/`H16` vectors.

G1, G2, and G3 each have an executed Curator reproduction, compiler evidence
from this host, and a concrete required change, all inside
`internal/swiftpminterop/headers.go` (plus vectors). They share one root cause
that the previous three rounds did not address: `classifiableDirectives` was
built as "directives that are not `#include`", but two of its members —
`embed` and `pragma` — name directives that read files. The next round should
re-derive that set from "does this directive cause the compiler to open a
file", not from directive-name familiarity, and should treat `_Pragma` as a
token-level channel the same way `@import` now is.

The multi-package interop coverage gap noted in round 3 remains open and is
still worth closing in the same round.

As a reviewer-archetype run this supplies no `commit_ack`.
