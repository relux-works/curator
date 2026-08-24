# Reviewer verdict for TASK-260811-tkurtl — N1/N2/N3 and the step-4 enumeration, round 12

Verdict: **changes requested -> `to-dev`**

- Reviewer run: `RUN-260824-74cbb4`
- Model: Claude `claude-opus-5`
- Reviewed delivery: `TASK-260811-tkurtl_rework-outcome_RUN-260824-round11.md`
  (the brief cites the outcome for RUN-260824-36e509; the artifact on the board
  carries the `_round11` suffix — same run, no other candidate exists)
- Answering verdict: `TASK-260811-tkurtl_review-verdict_RUN-260824-d10094.md`
- Compiler for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
  `arm64-apple-darwin25.5.0` — the accepted Darwin profile
- No product code, board file, or git state was modified. Two probe files added
  under `internal/swiftpminterop/` were removed; `git status` is unchanged apart
  from the pre-existing worktree delta.

N1, N2, and N3 are closed for every shape the previous verdict named, and closed
the right way. The step-4 enumeration is the acceptance-deciding item and it is
**not complete**: it enumerates identifier *positions* exhaustively but leaves
the *macro oracle* that decides those positions incomplete. A macro delivered
through a SwiftPM build setting instead of a source `#define` reproduces both N1
and N2 end-to-end, with a wrong module recorded.

---

## Items 1-3 — N1, N2, N3: accepted

Reproduced independently through the real `Close()` path on the admitted
`root:CLib` C target. All ten reject, each with the intended diagnostic:

| Probe | Body | Verdict |
| --- | --- | --- |
| N1a-N1e | `#define protocol\|class\|selector\|end\|YES import` + `@ KW SecretKit;` | `swiftpm_header_input_undeclared` — macro binds an `@`-keyword-position identifier |
| N1f | `#define class im##port` + `@ class SecretKit;` | same, at the name, before any body analysis |
| N2 | `#define CLib SecretKit` + `@import CLib;` | `swiftpm_header_input_undeclared` — module import names a macro-defined identifier |
| N3a | `#define A __as%:%:m__` | `swiftpm_target_platform_unsupported` — inline assembly |
| N3b | `#define J(a,b) a%:%:b` | `swiftpm_header_input_undeclared` — call-site fragment paste |
| N3c | `#define A _Prag%:%:ma` | `swiftpm_header_input_undeclared` — `_Pragma` operand |

**N1 — the design call is right.** Rejecting the `#define` rather than `@ IDENT`
is the only one of the two offered rules that is decidable where it is written,
and the outcome states exactly why: the realistic vector binds in a header and
uses in a `.c` file, and this scanner does not model conditional inclusion, so
"is IDENT macro-defined here?" is not answerable at the `@`. `N1i` proves the
cross-file case bites. The `interface`/`true` narrowing is real, named, and in
the safe direction. `objcAtKeywords` is now correctly documented as a
recognizer rather than as the closure.

**N2 — closed, and the asymmetry is honored.** `ExpandedName` is true for
`@import` and for `__pragma`, false for `#pragma clang module import` and for the
destringized `_Pragma` operand, each with its probe recorded as a comment at the
site. The `__pragma` disposition was **not** in my brief and the outcome found it
by probing rather than assuming symmetry — that is the right instinct and it is
what a symmetry assumption would have missed. The `#define import protocol`
inverse (compiles, imports nothing, scanner would record an import) is the same
integrity class found from the opposite end and is also closed. The pragma-form
control admits and asserts `ExpandedName == false`; the literal
`@import CLib; @import Foundation;` positive admits and records both.

Confirmed for the record: with the fix in place the recorded evidence can no
longer name a module the compiler doesn't import **when the binding is a source
`#define`**. See finding P for the case where it still can.

**N3 — closed without over-rejecting.** `macroPasteWidth` reads both spellings.
I checked the normalization does not over-reach: a lone `%:` stringize operator
in a macro body, an ordinary `foo%:%:bar` paste of benign fragments, and a body
containing `%` and `:` but no paste all still admit.

## Item 4 — step-4 enumeration: **BLOCKING finding P**

The per-position table is right about every position it lists, and I could not
break it on the spelling axis. What it does not state is where the answer to
"is this identifier macro-defined?" comes from. Both closing rules read the same
oracle, and that oracle is source `#define`s only:

- `atPositionIdentifiers` is checked in `readMacroDefinition`, reached only from
  `readDirective` on a `#define` line;
- `rejectMacroDefinedModuleNames` consumes `scanResult.macros`, populated only by
  `noteMacro` from that same site.

A SwiftPM `.define` build setting is a macro the compiler binds and neither rule
sees. `decodeBuildSetting` (`internal/swiftpmsource/executor_runtime.go:556`)
folds every setting to `Kind: "swiftpm-setting"` with the raw JSON as `Value`
and sets `Unsafe` only when the payload contains `"unsafeFlags"`; both
`swiftpmsource` (`manifest.go:246`, `manifest.go:306`) and `swiftpminterop`
(`interop.go:227`) then inspect only `Unsafe`. A `define` setting is admitted and
invisible.

Compiler evidence — both shapes read the module:

| Probe | Invocation | Result |
| --- | --- | --- |
| D1 | `-Dprotocol=import` + `@ protocol SecretKit;` | SecretKit built, `secret.h` read (`#error SECRET_MODULE_WAS_READ` fires) |
| D2 control | `@import NoSuchKitXYZ;` alone | `fatal error: module 'NoSuchKitXYZ' not found` |
| D2 | `-DNoSuchKitXYZ=SecretKit` + `@import NoSuchKitXYZ;` | SecretKit built, `secret.h` read |

Scanner evidence — both admitted end-to-end through `Close()` on the fixture's
own `root:CLib` target, with the setting attached to the target the same shape
`decodeBuildSetting` produces:

```
--- FAIL: TestR12BuildSettingDefineRoute/D1_@-position_identifier_rebound_by_a_cSettings_define
    recorded reference: {Spelling:"CLib.h" ModuleImport:false}
    ADMISSION HOLE: closure admitted the target with no rejection
      (setting {"define":[{"name":"protocol","value":"import"}]})

--- FAIL: TestR12BuildSettingDefineRoute/D2_@import_module_name_rebound_by_a_cSettings_define
    recorded reference: {Spelling:"CLib" ModuleImport:true ExpandedName:true}
    ADMISSION HOLE: closure admitted the target with no rejection
      (setting {"define":[{"name":"CLib","value":"SecretKit"}]})
```

D2 is exactly the wrong-module read the brief names as decisive: the closure
retains `ModuleImport: true, Spelling: "CLib", ExpandedName: true` — its own
admitted module, satisfying `moduleDeclared` — while the compiler resolves the
aliased name. D1 records no module reference at all for an import that happens.

Proportionality, stated honestly: in the admitted profile clang's module search
is confined to admitted package roots plus the selected SDK roots, so an aliased
name lands on an SDK/toolchain module or another package's module, not on
arbitrary bytes. The blast radius is smaller than `.incbin`. It is nonetheless
precisely the defect N2 exists to close — an undeclared module edge plus
evidence naming a module the target does not read — and `swiftpm_header_input_undeclared`
is the code the accepted SwiftPM outcome assigns it.

This is one finding, not a redesign: `BuildSetting.Value` already retains the
raw setting JSON, so the definitions are recoverable where they are needed. The
shape of the fix is yours; the two moves that mirror what already works are
rejecting a `define` setting that binds an `atPositionIdentifiers` member, and
unioning every `define` setting name into the macro set
`rejectMacroDefinedModuleNames` consumes. Under the accepted reject-by-default
posture the stronger option — enumerate the build-setting kind axis the way the
pragma axis was enumerated and reject any kind portable mode cannot prove is
macro-inert and resolution-inert — is also defensible and would close the axis
rather than this one member of it.

Please also state the oracle's provenance in the step-4 comment, not only the
positions. The table as written reads as complete because every row's
disposition is correct; what makes it incomplete is one level down.

### Spelling axis — probed, no further hole

Everything below was executed; each rejects or is inert on the pinned compiler.

| Probe | Result |
| --- | --- |
| `%:%:` split across a line splice (`__as%:\`+nl+`%:m__`) | rejects — phase 2 runs before the paste layer |
| the same split mid-operator (`__as%\`+nl+`:%:m__`) | rejects |
| `_Pragma("clang " "module import SecretKit")` | compiler: `_Pragma takes a parenthesized string literal`; scanner rejects the operand anyway |
| `#define H "CLib.h"` + `#include H` | rejects — computed include |
| trigraph spelling of `##` (`??=??=`) | whole-file rejection, unchanged |

The reviewer-verified premises the outcome preserved — no directive re-scan of
macro output, no adjacent-token merge without a paste, `_Pragma`'s operand not
expanded, `#pragma` import not expanded — are carried unchanged and correctly
attributed.

## Item 5 — positive path and no regression: accepted

392 PASS / 0 FAIL, reproduced independently (`go test -count=1 -v
./internal/swiftpminterop/`), matching the reported count and up 23 from 369.
Spot-checked positives all green: literal `@import CLib;`, an `@encode`
construct after `@`, plain quoted and angled includes, the three
digraph-over-rejection cases above. `S02`, `S03`, `S05`-`S08`, `S10`
(case-sensitive `.C`/`.M`), `H01`-`H23`, `CGP05` including its conditional
branch, `CGN*`, `R*`/`P*`, the transitive worklist, the module-map out-of-root
guard, `publicHeadersPath`, and the cross-package include closure are present
and green. `gofmt -l` clean, `go vet ./internal/swiftpminterop/` exit 0.

Scope hygiene intact: only `headers.go`, `interop.go`, and `modulemap_test.go`
carry this round's timestamps; `closuregraph`, `swiftpmsource`, `2qfnai`, and
`x611eq` are untouched. `IncludeGrammarID` correctly bumped to
`c-family-include-scanner-v10`.

## Item 6 — evidence: accepted

`TASK-260811-tkurtl_full-go-11.log` ends `EXIT:0`, 52 `ok` packages, 0 failures;
sha256 `cd8854030f82a6620b1dc292d4775c8461d3a45f55f0099e1b2ee8785f39353d` —
matches the brief exactly. `task-board --no-update-check validate` ->
`Board is valid. No issues found.`

Rerun by me: the focused package suite (verbose, 392/0), `gofmt`, `go vet`, the
compiler probe matrix, and the scanner probe matrix. Accepted without rerun:
race, lint, `cmd/curator`, canonical goldens, and the monolithic full-suite log.

---

## Routing

`to-dev` for finding P — the macro oracle behind the step-4 enumeration, and the
step-4 comment that should state its provenance.

Everything else in this round is accepted and needs no further work. N1, N2, and
N3 are closed; the spelling axis held under adversarial probing; the positive
path is intact.

No stop-the-line condition: this is ordinary, recoverable rework inside the
accepted reject-by-default posture, against one concrete and reproduced shape,
with the retained setting payload already sufficient to close it.

As a reviewer-archetype run this supplies no `commit_ack`.
