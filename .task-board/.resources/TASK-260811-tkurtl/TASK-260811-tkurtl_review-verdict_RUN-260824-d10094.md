# Reviewer verdict for TASK-260811-tkurtl — phase-4 closure, round 10

Verdict: **changes requested -> `to-dev`**

- Reviewer run: `RUN-260824-d10094`
- Model: Claude `claude-opus-5`
- Reviewed delivery: `TASK-260811-tkurtl_rework-outcome_RUN-260824-7b8a44.md`
  (the brief cites this artifact as `..._RUN-260824-ed3a24.md`; no such outcome
  exists — `ed3a24` is the prior verdict and `7b8a44` is the phase-4 outcome
  answering it)
- Compiler for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
  `arm64-apple-darwin25.5.0` — the accepted Darwin profile
- No product code, board file, or git state was modified. The one probe file
  added under `internal/swiftpminterop/` was removed; `git status` is unchanged
  apart from the pre-existing worktree delta.

The four probes finding M named are closed, and closed correctly. The class they
belong to is not. Three residual routes deliver the same channel keyword — or a
different module than the one recorded — with **no rejection at all**, and all
three are reachable end-to-end through `Close()` in the standard fixture.

---

## Blocking finding N1 — `objcAtKeywords` is an allowlist of macro-definable identifiers

`readAtToken` rejects `@` followed by an identifier outside the 29-member
`objcAtKeywords` set. Every identifier **inside** that set is an ordinary
identifier to the preprocessor, so it can be `#define`d as a macro expanding to
`import`. Finding M's shape, routed through the allowlist rather than around it.

Compiler evidence (`review-clang-evidence` log):

| Probe | Source | Compiler result |
| --- | --- | --- |
| N1a | `#define protocol import` + `@ protocol SecretKit;` | module built, `secret.h` read (`#error SECRET_MODULE_WAS_READ` fires) |
| N1b | `#define class import` + `@ class SecretKit;` | same |
| N1c | `#define selector import` + `@ selector SecretKit;` | same |
| N1d | `#define end import` + `@ end SecretKit;` | same |

Scanner evidence (`review-scanner-probe` log), all through the real `Close()`
path on the admitted `root:CLib` C target:

```
--- FAIL: .../N1a_@-keyword_macro_protocol        ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../N1b_@-keyword_macro_class           ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../N1c_@-keyword_macro_selector        ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../N1d_@-keyword_macro_end             ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../N1e_@-keyword_macro_YES             ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../N1f_@-keyword_macro_built_by_paste  ADMISSION HOLE: closure admitted the target with no rejection
```

N1f is `#define class im##port` — the fixed-fragment paste is collapsed
correctly and then admitted, because the collapsed body `define class import`
contains no `@`, and the `@`-side rule then waves `@ class` through. The two
halves of the phase-4 fix each behave as designed and the composition still
leaks.

`#define <objcAtKeyword>` is not a legitimate C or Objective-C idiom, so
rejecting a `#define` whose macro **name** is in `objcAtKeywords` looks like the
cheap closed move — but the design call is yours. What is not defensible is the
current state, where the allowlist that exists to make `@` fail closed is itself
the delivery vehicle.

---

## Blocking finding N2 — the `@import` module name is macro-expanded; the scanner records and gates on the pre-expansion spelling

`readModuleImport` records `IncludeReference{Spelling: name, ModuleImport: true}`
and `confineInclude` gates that spelling through `moduleDeclared`. The pinned
compiler expands the identifier after `@import` before resolving the module.

| Probe | Source | Compiler result |
| --- | --- | --- |
| control | `@import NoSuchKitXYZ;` | `fatal error: module 'NoSuchKitXYZ' not found` |
| N2 | `#define NoSuchKitXYZ SecretKit` + `@import NoSuchKitXYZ;` | `SecretKit` built, `secret.h` read |
| control | `#define DeclaredKit SecretKit` + `#pragma clang module import DeclaredKit` | imports **`DeclaredKit`** — the pragma spelling is *not* expanded |

End-to-end, with the fixture's own admitted module `CLib`:

```
--- FAIL: .../N2_aliased_module_import
    ADMISSION HOLE: admitted, recorded module import "CLib" while the compiler imports SecretKit
```

Two separate defects in one shape: the `moduleDeclared` gate is satisfied by a
name the compiler never resolves, and the recorded closure evidence names the
wrong module. The `@import` / `#pragma clang module import` asymmetry is real
and verified — worth encoding as a comment wherever the fix lands, because the
obvious symmetry assumption is the wrong one.

---

## Blocking finding N3 — `%:%:` is the digraph spelling of `##` and `collapseMacroPastes` does not see it

`collapseMacroPastes` short-circuits on `!strings.Contains(body, "##")`.
`readDirective` already understands `%:` as a directive introducer, so digraph
awareness exists in this grammar — just not in the paste layer.

| Probe | Source | Compiler result |
| --- | --- | --- |
| control | `__asm__(".incbin \"payload.bin\"");` | `.o` 528 B, payload bytes present |
| N3a | `#define A __as%:%:m__` + `A(…)` | `.o` **byte-size identical to the control**, payload bytes present |
| N3b | `#define J(a,b) a%:%:b` + `J(a,sm)(…)` | same |
| N3c | `#define A _Prag%:%:ma` + `A("clang module import SecretKit")` | module built, `secret.h` read |

```
--- FAIL: .../N3a_object-like_digraph_paste    ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../N3b_function-like_digraph_paste  ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../N3c_digraph_paste_to_pragma      ADMISSION HOLE: closure admitted the target with no rejection
```

N3a/N3b are arbitrary absolute-path file reads with no confinement — the same
`.incbin` channel `H18` exists to close. N3b in particular is the *exact* M1
body with `##` respelled, and the M1 rejection is right next to it in `H21`.
(`#define A @im%:%:port` happens to reject, but only incidentally, via the `@`
rule — not via the paste layer.)

---

## Cross-layer closure argument — audited, structure sound, two premises false

This was the acceptance-deciding item, so I checked the argument's own claims
rather than its conclusion.

**Premises that hold, verified on the pinned compiler:**

- *"the preprocessor does not re-scan macro output for directives"* — confirmed.
  `#define INC #include "sk/secret.h"` invoked as `INC` produces
  `error: expected identifier or '('` and reads nothing. The brief asked for
  this specifically; it is correct.
- *"adjacent tokens never merge"* — confirmed. `#define A __as` + `#define B m__`
  + `A B(".incbin …");` produces `error: unknown type name '__as'` and an object
  with no payload bytes. Object-like macro chains therefore cannot assemble a
  keyword without `##`, and a chain that *does* (`a##s##m`) is collapsed and
  rejected correctly.
- Steps 1-3 and 5 are accurate and were accepted last round.
- `_Pragma` produced by expansion is handled: `#define P _Pragma` rejects at the
  definition, and `_Pragma("…")`'s string operand is not macro-expanded
  (verified — `#define X SecretKit` + `_Pragma("clang module import X")` yields
  `module 'X' not found`).

**Premises that are false as written:**

- Step 4's *"the single exception is `##`"* is a spelling-level claim, and
  `%:%:` is the same operator (N3).
- Step 4's `@` sub-claim — *"`@` is not an identifier and cannot be pasted into
  one, so the `@import` channel is covered by the separate `@`-rule"* — is true
  about the `@` and false about what follows it (N1), and says nothing about the
  module name (N2).

So the argument's shape is right and its method is right; it is not yet the
statement that closes the class. Please re-derive step 4 over **preprocessing
tokens including their alternative spellings**, and state explicitly, per
identifier position the scanner records or gates on, whether the compiler
expands it. That last enumeration is what would have caught N2, which is not a
channel-keyword problem at all — it is an evidence-integrity problem that the
keyword framing does not reach.

---

## Item 3 — raw strings: accepted, and done the stronger way

The outcome took the harder of the two offered options. `rawStringPrefixAt`
recognizes the five encoding prefixes at a token boundary in content, in
directive bodies (via `readDirectiveBody`'s `s.rawString` latch) and in macro
definitions, and `rejectRawString` fails the target. `H22` asserts `scanIncludes`'
own verdict for five spellings — so the proof no longer depends on
`artifactpolicy`'s ordering — plus an end-to-end ordinary raw string that
`artifactpolicy` **admits** and this stage rejects, plus a token-boundary
positive (`myR"tail"`, `int R = 1;`). The phase-3 parity row is corrected and now
calls the divergence load-bearing rather than conservative. Nothing further
needed here.

## Item 4 — positive path: accepted, no regression

369 PASS / 0 FAIL, reproduced independently (`go test -count=1 -v
./internal/swiftpminterop/`), matching the reported count and up 28 from 341.
`S02`, `S03`, `S05`-`S08`, `S10` (case-sensitive `.C`/`.M`), `H01`-`H22`,
`H10`/`H11` (`publicHeadersPath` and the non-representable layout), `CGP05`
including its conditional branch, `CGN*`, `R*`/`P*`, the transitive worklist,
the module-map out-of-root guard, and the cross-package include closure are all
present and green. The narrowings — the inverted `JOIN` positive, parameter
pasting, `@` + macro — are deliberate, recorded, and not held against the work.
Seam/guard discipline and scope hygiene are intact: this round touched only
`internal/swiftpminterop/headers.go` and `modulemap_test.go`, and
`closuregraph`, `swiftpmsource`, `2qfnai`, and `x611eq` are untouched.

## Item 5 — evidence: accepted

`TASK-260811-tkurtl_full-go-10.log` ends `EXIT:0`, 52 `ok` packages, 0 failures;
sha256 `d1bd6ba34f8802c2b071d16aaff8475927c1b0790746ace04b9ad0f727eb2bdb` —
matches the brief exactly. `task-board --no-update-check validate` ->
`Board is valid. No issues found.`

Rerun by me: focused package suite (verbose, 369/0), board validate, the
compiler probe matrix, and the scanner probe matrix. Accepted without rerun:
race, lint, vet, gofmt, `cmd/curator`, canonical goldens, and the monolithic
full-suite log.

---

## Routing

`to-dev` for implementation rework on N1, N2, and N3, and for a step-4 closure
argument re-derived over token spellings and over every identifier position the
scanner records or gates on.

No stop-the-line condition: this is ordinary, recoverable rework inside the
accepted reject-by-default posture, against a concrete and now-enumerated set of
shapes.

As a reviewer-archetype run this supplies no `commit_ack`.
