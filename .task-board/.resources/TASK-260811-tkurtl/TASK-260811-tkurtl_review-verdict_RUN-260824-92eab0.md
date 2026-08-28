# Reviewer verdict for TASK-260811-tkurtl — finding P and the build-setting axis, round 13

Verdict: **changes requested -> `to-dev`**

- Reviewer run: `RUN-260824-92eab0`
- Model: Claude `claude-opus-5`
- Reviewed delivery: `TASK-260811-tkurtl_rework-outcome_RUN-260824-round12.md` (RUN-260824-660e2f)
- Answering verdict: `TASK-260811-tkurtl_review-verdict_RUN-260824-74cbb4.md`
- Compiler for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
  `arm64-apple-darwin25.5.0` — the accepted Darwin profile
- No product code, board file, or git state was modified. Two probe files added
  under `internal/swiftpminterop/` were removed; `git status` is unchanged apart
  from the pre-existing worktree delta.

Finding P as the previous verdict framed it — the define **name** invisible to
both macro oracles — is closed, and closed the stronger way. The kind axis is
enumerated soundly and I could not break `headerSearchPath`, `unsafeFlags`, the
link gate, or any admitted-inert kind. The oracle-input axis is **not** closed:
the `define` route reads only the setting's *name*. The setting's *body* is
never analysed, and a body is a channel-delivery vector that round 9 finding M
already established and closed for source `#define`s. All six round-9-class
bodies reproduce end-to-end through a build setting, two of them confirmed as
real file reads on the pinned compiler.

---

## Item 1 — finding P (D1/D2, both oracles): accepted

Reproduced independently. `disposeBuildSettings`
(`internal/swiftpminterop/buildsettings.go:126`) rejects a define whose name is
an `atPositionIdentifiers` member before any file is scanned (the N1 route), and
`state.settingDefines` is seeded into the per-target macro set in
`scanAndConfineIncludes` (`interop.go:600-603`) before the first file is opened,
so `rejectMacroDefinedModuleNames` consumes the union (the N2 route). Both
oracles are fed; the wiring is correct.

D1, D2, the D2 control, and the pruned-condition case all behave as reported —
D2 publishes no `Result`, so no wrong module is recorded. The setting payloads in
the vectors are literal `dump-package` output: I re-derived the encoding from
`TASK-260811-tkurtl_round12-swiftpm-setting-kinds.log`, and
`.define(name, to: value)` really does serialize as
`{"kind":{"define":{"_0":"NAME=VALUE"}},"tool":"c"}`.

Nice call on `decodeBuildSetting` staying untouched — recovering the kind from
the retained `Value` moves no canonical capture record. The
`cxxInteropSetting` repair the axis exposed is real and correctly scoped.

## Item 2 — build-setting kind axis: **BLOCKING finding Q on the `define` row**

The per-kind table is right about every kind it names, and the two-axis
argument ("`define` is the only `-D`, `headerSearchPath` is the only `-I`") holds
against what I could probe:

| Probed | Result |
| --- | --- |
| `headerSearchPath` admitted? | no — rejects on the kind (`swiftpm_unsafe_build_setting_forbidden`). Correct: it is the only non-`unsafeFlags` `-I` and the include closure cannot follow it |
| `unsafeFlags` gated only by the `Unsafe` substring flag? | no — rejects on the kind name too |
| `linkedLibrary`/`linkedFramework` ungated? | no — `linkDeclared`, undeclared -> `artifact_toolchain_untrusted` |
| unnamed future kind / unreadable payload | reject, twice over (zero value + membership check) |
| any admitted-inert kind reaching `-D` or `-I` | none found: every one is a Swift language/feature flag, a diagnostic severity, or an isolation default |

**What is missing is one level below the row, exactly as last round.** The
`define` disposition does this and only this:

```go
name, _ := splitLeadingIdentifier(strings.TrimLeft(value, " \t"))
if atPositionIdentifiers[name] { … reject … }
defines[name] = true
```

The body after `=` is discarded. The source route does not discard it:
`readMacroDefinition` (`headers.go:1231`) runs `collapseMacroPastes` and then
`scanDirectiveChannels(collapsed)` over the body, which is precisely what round 9
finding M added — a macro body can deliver a channel keyword into a position the
token scanner reads as ordinary content, and the call site `A(…)` is invisible
because `A` is not a channel keyword.

### Compiler evidence — a build-setting body really reads files

`payload.bin` contains a marker string; `SecretKit`'s only header is
`#error SECRET_MODULE_WAS_READ`.

| Probe | Invocation | Result |
| --- | --- | --- |
| Q1 | `clang -c -D'A=__asm__' d.c` where `d.c` has `A(".incbin \"payload.bin\"");` | exit 0; `payload.bin`'s bytes present in `d.o` at `0x180` |
| Q1 direct control | the same source spelling `__asm__(…)` directly, no `-D` | exit 0; object **byte-identical**, sha256 `8e639ccfe3a3ebcab9c95b2aa6a0872f71c1e60ece92d1eb6cd41adfa40f7334` |
| Q1 negative control | the same source, no `-D` | exit 1 — `A` is undeclared; the setting is the entire vector |
| Q3 | `clang -fsyntax-only -fmodules -I secret -D'A=_Pragma' p.c` where `p.c` has `A("clang module import SecretKit")` | `While building module 'SecretKit' … error: SECRET_MODULE_WAS_READ` |
| Q3 control | the same source, no `-D` | `expected parameter declarator` — nothing imported |

Q1 is `.incbin` reading an arbitrary file through the assembler stage portable
mode declares it does not admit; Q3 is an undeclared module import.

### Scanner evidence — all six admit, and the source spelling of each rejects

Run through the real `Close()` path on the fixture's own admitted `root:CLib`
target, with the setting attached in the shape `decodeBuildSetting` produces:

| Vector | Setting | Source body | Closure |
| --- | --- | --- | --- |
| Q1 | `{"define":{"_0":"A=__asm__"}}` | `A(".incbin \"payload.bin\"");` | **ADMITTED**, includes = `["CLib.h"]` |
| Q1b | `{"define":{"_0":"A=asm"}}` | same | **ADMITTED** |
| Q3 | `{"define":{"_0":"A=_Pragma"}}` | `A("clang module import SecretKit")` | **ADMITTED** |
| Q4 | `{"define":{"_0":"A=__pragma"}}` | `A(clang module import SecretKit)` | **ADMITTED** |
| Q5 | `{"define":{"_0":"A=__as##m__"}}` | `A(".incbin \"payload.bin\"");` | **ADMITTED** |
| Q6 | `{"define":{"_0":"J(a,b)=a##b"}}` | `J(__as,m__)(".incbin \"payload.bin\"");` | **ADMITTED** |

The matching controls prove this is a route gap, not a grammar gap — the same
macro spelled in source rejects every time:

| Source control | Diagnostic |
| --- | --- |
| `#define A __asm__` | `swiftpm_target_platform_unsupported` — inline-assembly construct |
| `#define A _Pragma` | `swiftpm_header_input_undeclared` — unclassifiable `_Pragma` operand |
| `#define A __as##m__` | `swiftpm_target_platform_unsupported` — inline-assembly construct |
| `#define J(a,b) a##b` | `swiftpm_header_input_undeclared` — call-site fragment paste |

Q1 and Q6 are round 9 finding M's own vectors (`J(__as,m__)`, `__as##m__`),
reproduced verbatim one level down — the same relationship round 11's N1/N2 had
to round 12's finding P.

### What is NOT affected

I checked the neighbouring positions so the rework brief is bounded:

- **`@`-position is safe.** `.define("I", to: "import")` + `@ I SecretKit;`
  rejects, identically to the source `#define I import` control — the `@`
  allowlist closes at the `@`, not at the binding, so a body that reaches
  `import` cannot get through. No change needed there.
- **Computed includes are safe.** `#include A` rejects at the use site
  regardless of where `A` is bound.
- **`#embed` is unreachable by expansion** — a directive is not formed by macro
  expansion.

So the reachable channel set through a define body is exactly inline assembly
(`asm`/`__asm`/`__asm__`), `_Pragma`, `__pragma`, and `##`/`%:%:` paste into any
of them. All are already classified by `scanDirectiveChannels` and
`collapseMacroPastes`; they are simply not run on this input.

### Shape of the fix (yours to choose)

The two moves that mirror what already works:

1. run the build-setting define's body through the same
   `collapseMacroPastes` + `scanDirectiveChannels` pair `readMacroDefinition`
   uses, so `A=__asm__`, `A=_Pragma`, `A=__as##m__` and the parameter-paste form
   reject at the setting, before any file is scanned; and
2. parse the function-like form (`J(a,b)=a##b`) into its parameter set the same
   way, so the parameter-paste rule can fire.

Under the accepted reject-by-default posture the blunter close is also
defensible: admit a define body only when it is provably inert (empty, or a
numeric/string literal), and reject any body carrying an identifier or a
punctuator this stage does not classify. That narrows `.define("X", to: "Y")`
further, but the axis statement becomes "a build-setting define binds a name and
an inert body, both proven" with nothing left implicit. Either way, please make
the step-4 comment say that the oracle's second input is checked on **both**
its name and its body — that is the sentence whose absence is the finding.

## Item 3 — oracle provenance stated: accepted as far as it goes

`scanIncludes`'s phase-4 table now names the union and both binding sites,
`rejectMacroDefinedModuleNames` carries the same statement, and the two affected
rows say "a `#define` or a `.define` build setting". The three-axis closure
claim (spelling + position + oracle-input) is stated. It does not hold yet, for
the reason above: the oracle-input axis is closed for *names* and open for
*bodies*.

## Item 4 — positive path and no regression: accepted

431 PASS / 0 FAIL, reproduced independently (`go test -count=1 -v
./internal/swiftpminterop/`), matching the reported count and up 39 from 392.
Family spot-check in my own log: `S02` 1, `S03` 1, `S05` 4, `S10` 8, `H01` 2,
`H12` 12, `H17` 30, `H18` 48, `H23` 23, `H24` 39, `CGP05` 3, `CGN*` 13 — all
green. The benign-settings positive admits with both plain includes intact;
`linkedLibrary("c")` and `linkedFramework("Foundation")` admit through the SDK
component while `z`/`SecretKit` reject. `gofmt -l ./cmd ./internal` clean,
`go vet ./internal/swiftpminterop/` exit 0.

Scope hygiene intact: only `buildsettings.go`, `buildsettings_test.go`,
`interop.go`, `headers.go`, and `platform.go` carry this round's timestamps;
`closuregraph`, `swiftpmsource`, `2qfnai`, and `x611eq` are untouched. The
`IncludeGrammarID` hold at `c-family-include-scanner-v10` is correctly reasoned
— the scanner grammar over a file did not change this round. Note that if the
fix runs the channel classifier over a build-setting body, the grammar identity
still does not change; the *closure* input does.

## Item 5 — evidence: accepted

`TASK-260811-tkurtl_full-go-12.log` ends `EXIT:0`, 52 `ok` packages, 0 `FAIL`;
sha256 `6f2bc1d660844b63a5dcdfeb270f23f022ea943c27043a8817117a975e7b6118` —
matches the brief exactly. `task-board --no-update-check validate` ->
`Board is valid. No issues found.` The `gofmt -l .` honesty note about the 79
`.temp/` files is the right call and is corroborated by the attached unscoped log.

Rerun by me: the focused package suite (verbose, 431/0), family spot-checks,
`gofmt`, `go vet`, the compiler probe matrix above, and the scanner probe matrix
above. Accepted without rerun: race, lint, `cmd/curator`, canonical goldens, and
the monolithic full-suite log.

---

## Routing

`to-dev` for finding Q — the build-setting `define` body, which reproduces round
9 finding M's channel vectors through the input finding P added.

Everything else this round is accepted and needs no further work: the kind axis
enumeration, the two-axis inertness argument, the `headerSearchPath`/
`unsafeFlags`/link/unknown-kind dispositions, the D1/D2 name routes into both
oracles, the `cxxInteropSetting` repair, the 39 H24 vectors, the positive path,
and the evidence.

No stop-the-line condition: this is ordinary, recoverable rework inside the
accepted reject-by-default posture, against six concrete and executed shapes,
with the classifier the fix needs already present and already applied to the
identical source construct.

As a reviewer-archetype run this supplies no `commit_ack`.
