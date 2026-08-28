# Reviewer verdict for TASK-260811-tkurtl — posture pivot, round 9

Verdict: **changes requested -> `to-dev`**

- Reviewer run: `RUN-260824-ed3a24`
- Model: Claude `claude-opus-5`
- Reviewed delivery: `TASK-260811-tkurtl_rework-outcome_RUN-260824-f1733b.md` (reject-by-default pivot)
- Compiler used for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
  `arm64-apple-darwin25.5.0` — the accepted Darwin profile
- No product code, board file, or git state was modified. Nothing staged or committed.
  The one probe file added under `internal/swiftpminterop/` was removed; the package
  is green and `git status` is unchanged apart from the pre-existing worktree delta.

The pivot is the right structural move and most of it landed correctly. It does
not yet terminate, because the reject is enforced on **source-text tokens** while
the pinned compiler reconstitutes the same channel keywords through **macro
expansion**, which this scanner does not perform and therefore never sees. Four
shapes reach a real file read with no rejection at all.

---

## Blocking finding M — macro-reconstituted channel keywords bypass every reject

`startsAsmStatement`, `startsPragmaOperator`, and `startsModuleImport` all match a
literal keyword spelling in the translated text. Translation phases 1-3 are now
correct, so no *lexical* trick reconstitutes those keywords — but **phase 4 macro
expansion still does**, and it runs after the scanner has already admitted the file.

`scanDirectiveChannels` correctly catches the case where a macro *body literally
contains* the keyword (`#define K asm`, `#define STMT(x) __asm__(x)` — both in
`H18`, both rejecting). It does not catch a keyword **assembled from fragments
that are not the keyword**, nor a keyword supplied by an object-like macro in a
position the scanner reads as ordinary content.

### Compiler evidence (`review-probe-clang-evidence` log)

| Probe | Source | Compiler result |
| --- | --- | --- |
| control | `__asm__(".incbin \"payload.bin\"");` | `.o` = 1592 B, payload bytes present |
| M1 | `#define J(a,b) a##b` + `J(a,sm)(".incbin \"payload.bin\"");` | `.o` **byte-size identical to the control**, payload bytes present |
| M2 | `#define J(a,b) a##b` + `J(__as,m__)(".incbin \"payload.bin\"");` | same — payload bytes present |
| M3 | `#define J(a,b) a##b` + `J(_Prag,ma)("clang module import SecretKit")` | module built and `secret.h` read (`#error SECRET_MODULE_WAS_READ` fires) |
| M4 | `#define I import` + `@ I SecretKit;` | module built and `secret.h` read (same marker fires) |

M1/M2 are arbitrary **absolute-path file reads** with no confinement whatsoever —
the same `.incbin` channel `H18` exists to close. M3/M4 evade `moduleDeclared`
entirely, so a module reference the target never declares is imported unchecked.

### Scanner evidence (`review-probe-macro-reconstitution` log)

All four bodies were run through the real `Close()` path in the standard fixture
(`Sources/CLib/lib.c`, admitted C target). Result:

```
--- FAIL: .../pasted_asm_keyword        ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../pasted___asm___keyword    ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../pasted__Pragma_keyword    ADMISSION HOLE: closure admitted the target with no rejection
--- FAIL: .../macro_import_after_at     ADMISSION HOLE: closure admitted the target with no rejection
```

Not a wrong diagnostic, not a narrowed one — **no error is returned at all**.
They also clear `artifactpolicy`, so nothing upstream catches them either.

### Why the current grammar misses them

- **M1/M2** — `run()` walks `J`, `(`, `a`, `,`, `sm`, `)`. `asmKeyword()` tests the
  prefixes `__asm__` / `__asm` / `asm` at each cursor; at `a` the text is `a,sm)`
  and at `s` it is `sm)`. Neither matches. The keyword exists only after `##`
  pastes two operands that arrive from the **call site**, which is ordinary content.
- **M3** — identical, with `_Pragma` as the pasted result. `classifyPragmaBody`
  never runs because `startsPragmaOperator` never fires.
- **M4** — `startsModuleImport` requires the identifier after `@` to be literally
  `import`. `I` is not, so the `@` falls to `default: s.index++` and is consumed as
  one content byte. Clang expands `I` to `import` and imports.

### What closing it requires (producer's design call, not mine)

The closed set here is the **channel keyword set** the scanner already owns
(`asm`, `__asm`, `__asm__`, `_Pragma`, `__pragma`, `import`). The reject-by-default
form of the rule is: portable mode must not admit a construct that can deliver one
of those keywords into a position the scanner cannot see.

Two boundaries need a decision, and both carry a cost worth stating plainly:

1. **`##` paste.** For a *function-like* macro the fragments come from the call
   site, so no body-local analysis is sound; the closed move is to reject a
   function-like macro that pastes a parameter. For an *object-like* macro the
   fragments are in the body and can be concatenated and tested directly. Note the
   cost: `modulemap_test.go:895` currently pins `#define JOIN(a, b) a##b` as an
   admitted positive, and parameter pasting is common in real C headers — so this
   narrows portable acceptance further. That is consistent with the accepted pivot
   contract, but it should be a deliberate, recorded narrowing with the positive
   vector inverted, exactly as `#embed` was handled this round.
2. **`@` followed by a non-`import` identifier.** Objective-C `@`-keywords are a
   closed set. An identifier after `@` that is not in it is a macro that may expand
   to `import`, so it must reject rather than fall through as content.

If either narrowing is judged too costly, the honest alternative is to state in the
scope that portable mode's channel proof holds only up to macro expansion and to
move that residual to the observed-read provider — but as written the scope claims
the channel axis "FAILS CLOSED on everything else it cannot prove reads no file",
and these four shapes contradict that claim.

---

## Secondary observation (non-blocking) — C++ raw strings are unmodeled and only defended in depth

The outcome states the raw-string divergence "can only add a rejection, never an
admission". That is not quite right: a raw string can hand the scanner an
**unmatched `"` followed by `/*`**, at which point `skipBlockComment` swallows the
rest of the file while the compiler sees no comment at all.

Verified on the pinned compiler — `const char* s = R"x(" /* )x";` followed by
`__asm__(".incbin \"payload.bin\"");` compiles cleanly and the payload bytes land
in the object (`raw.o`, 1872 B). The interop scanner is genuinely fooled by that
shape.

It is **not currently an admission hole**, because every spelling I tried (single
line, comment-closed-later, and multi-line raw string) is rejected upstream by
`artifactpolicy`'s source-text lexer with `artifact_opaque_dependency_forbidden`,
while an ordinary `R"x(hi)x"` is admitted. So the defense exists — but it lives in
another component, is incidental to this grammar, and is pinned by no vector here.
Please either pin it with an `H`-family vector that asserts the rejection (so a
future `artifactpolicy` relaxation cannot silently open it), or model raw strings
in `literalEnd`. Correct the phase-3 parity row either way: the divergence is
currently load-bearing, not merely conservative.

---

## Everything else in the brief — accepted

**1. Phases 1-3.** Correct and complete as far as I could probe.
`spliceTranslationLines` now removes `\` + `[ \t\v\f]*` + `\n` unconditionally and
mode-independently; the finding-L shapes (`__as\`+ws+nl+`m__`, `_Pra\`+ws+nl+`gma`,
`__prag\`+ws+nl+`ma`, `@imp\`+ws+nl+`ort`, spliced escaping `#include`) all reject
in `H19`, and all six separator variants plus CRLF/CR, consecutive splices and the
four non-splice backslash forms are pinned in `TestTranslationPhaseTwoSplicing`
(18 cases). Phase ordering is right: `findTrigraph` runs on the **pre-splice** text,
which matches the compiler's phase 1 -> phase 2 order, so `??/`-built splices and
`?\`+nl+`?=` pseudo-trigraphs both land correctly (reject / no-trigraph). Comments,
BOM/NUL/Unicode white space (`H15`), and the line-start prefix reject are sound.
The one gap is phase 4, above — not phases 1-3.

**2. Channel allowlist — closed and correctly enforced, at the lexical level.**
I enumerated the admitted surface against the code and it matches the outcome
exactly: literal `#include`/`#import`/`#include_next` only (`literalIncludeOperand`
rejects computed, empty, unterminated, and trailing-token forms), `@import NAME;`,
`#pragma clang module import NAME` plus its `_Pragma`/`__pragma` forms, the closed
`safePragmaHeads`/`safeClangPragmas`/`safeGCCPragmas` head lists, module maps, typed
boundaries. `readDirective`'s final `return reject(...)` makes any unclassified
`#`-line a rejection rather than a dropped line; `classifyPragmaBody`'s default is
`reject`; `#embed` rejects ahead of `inclusionDirectives`; `.s`/`.S`/`.asm` reject in
`classifyTarget`. `TestPragmaAllowlistIsClosed` (24) and `H20` (9) pin the closure.
`__has_include`/`__has_embed` staying admitted as existence oracles is correct —
they introduce no bytes.

**3. Assembler-classifier removal.** Safe as a *removal*: `rejectAsmStatement` is
reached from both `run()` and `scanDirectiveChannels`, and everything the deleted
classifier used to decide now rejects earlier and harder. The 48 `H18` cases are
retained verbatim and all still reject, including the four bodies that were the
round-7/8 positive control. The gap is not the removal — it is that the keyword
match feeding it is evadable (finding M).

**4. Positive path — no regression.** 341 PASS / 0 FAIL, independently reproduced
(`go test -count=1 -v ./internal/swiftpminterop/`), matching the reported count.
`S02`, `S03` (all four C-family languages in one Clang target), `S05`/`S06`
including the conditional selection-neutral opt-in, `S07`/`S08`, `S10`
(case-sensitive `.C`/`.M`), `H01`-`H20`, `H10`/`H11` (`publicHeadersPath` and the
non-representable layout), `CGP05` including its conditional branch, `CGN*`,
`R*`/`P*`, and the cross-package include closure are all present and green.
Canonical goldens reproduced independently:
`canonical_goldens=pass labeled_records=53 cgp05_target_branches=2
cgp10_observation_branches=2`, `canonical_references=pass`.

**5. Seam/guard discipline and scope hygiene.** Clean. No `os/exec` in
`swiftpminterop` (the only occurrence is the guard test asserting its absence, and
`TestInteropAdapterStartsNoProcessOutsideTheSharedSeams` passes). No Kotlin, no
vendored binaries. `git status` shows this round touched only
`internal/swiftpminterop/`; `closuregraph`, `swiftpmsource`, `2qfnai`, and `x611eq`
are untouched.

**6. Evidence.** `TASK-260811-tkurtl_full-go-09.log` is present, ends `EXIT:0`,
lists `internal/swiftpminterop 114.453s` and every other package `ok`;
sha256 `f9e89bc2b5054ff584a6bf7d788e05242b40efaa6c62328713f2e4965f29accf`.
`task-board --no-update-check validate` -> `Board is valid. No issues found.`
`go build ./...` exit 0. Reruns I executed myself: focused package suite (x3),
verbose matrix, canonical golden verifier, board validate, build. Accepted without
rerun: race, lint, `cmd/curator`, and the monolithic full-suite log.

---

## Routing

`to-dev` for implementation rework on finding M, plus the secondary raw-string
observation. No stop-the-line condition: this is ordinary, recoverable rework
inside the accepted posture, with a concrete closed keyword set to work against.

As a reviewer-archetype run this supplies no `commit_ack`.
