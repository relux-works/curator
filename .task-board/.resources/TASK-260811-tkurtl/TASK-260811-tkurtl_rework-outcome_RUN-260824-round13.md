# TASK-260811-tkurtl rework outcome — round 13, finding Q

Run: `RUN-260824` (developer, Claude `claude-opus-5`)
Answering verdict: `TASK-260811-tkurtl_review-verdict_RUN-260824-92eab0.md`
Compiler for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
`arm64-apple-darwin25.5.0` — the accepted Darwin profile.

Scope: finding Q only. Nothing else from round 12 was touched — the kind axis
enumeration, the `headerSearchPath`/`unsafeFlags`/link/unknown-kind
dispositions, the D1/D2 name routes into both oracles, the `cxxInteropSetting`
repair, and the 39 H24 vectors are all as accepted.

## Finding Q — what was wrong and what changed

`disposeBuildSettings`' `define` disposition checked only the setting's NAME
(`splitLeadingIdentifier`) and discarded everything after `=`. The source route
does not discard it: `readMacroDefinition` runs `collapseMacroPastes` then
`scanDirectiveChannels` over the replacement list — round 9 finding M. So a
build-setting define whose value was a channel-carrying macro body was admitted
while the source spelling of the identical body rejected.

### One analyzer, many inputs

The fix routes the second input into the first input's analyzer instead of
growing a second one. Two functions were factored out of `readMacroDefinition`
in `internal/swiftpminterop/headers.go` and are now called by both routes:

- `analyzeMacroBody(pkg, target, sourcePkg, source, body, parameters)` — the
  single phase-4 replacement-list analyzer. It is exactly the pair
  `readMacroDefinition` already ran: `collapseMacroPastes` (which already covers
  `##` and the `%:%:` digraph from N3) followed by `scanDirectiveChannels` over
  the collapsed stream. An unresolvable paste comes back as `errParameterPaste`
  so each caller renders it in its own input's terms; every other verdict is
  already a formed rejection carrying this stage's codes, which is why both
  routes produce the SAME diagnostic code for the same body.
- `readMacroParameters(rest)` — the function-like parameter reader, shared for
  the same reason: the parameter set is what lets `collapseMacroPastes` tell a
  definition-resolvable paste from one that takes a fragment from the call site,
  and that distinction must not differ between `#define J(a,b) a##b` and
  `.define("J(a,b)", to: "a##b")`.

`readMacroDefinition` is now a thin caller of both. It has no body logic left of
its own, so the two routes cannot drift.

In `internal/swiftpminterop/buildsettings.go`, `analyzeSettingDefineBody` parses
the value out of the retained setting payload the same way finding P reads the
name — `{"kind":{"define":{"_0":"NAME=VALUE"}},"tool":"c"}` — and calls
`analyzeMacroBody`. `splitSettingDefine` separates the `-D` operand into name,
optional parameter list, and replacement list: `NAME`, `NAME=VALUE`, or
`NAME(a,b)=VALUE`. A define with no `=` binds `1`, carries no channel, and stays
admitted exactly as today. An operand this stage cannot separate rejects
(`swiftpm_unsafe_build_setting_forbidden`) rather than being guessed at. The
analysis runs alongside the existing name check, before the first file is
opened.

One deliberate divergence, documented at the call site: a body that resolves to
a Clang module import is REJECTED rather than confined. A build setting is not
an admitted source file, so the reference belongs to no scanned unit, has no
directory to resolve against, and cannot join the include worklist. Under the
accepted reject-by-default posture the construct is refused. The ordinary
`-DFEATURE=1` body produces no reference at all, so the positive path is
untouched.

`decodeBuildSetting` and every canonical capture record are unchanged.
`IncludeGrammarID` holds at `c-family-include-scanner-v10`: the scanner grammar
over a file did not change, only which inputs are fed to it.

## Terminating statement — the macro-INPUT surface is closed

The macro-input surface for an admitted C-family target is now:

    source `#define`  (name closed by M/N1/N2, body closed by M)
  UNION
    build-setting `define`  (name closed by P, body closed by Q)

These are the only two macro-binding inputs the pinned `clang -c` honors for an
admitted C-family target. A macro reaches that compiler from exactly two places:
a preprocessing directive in a translation unit it reads, and a `-D` argument.
`-D` is reachable only through the `define` build-setting kind — the kind axis
accepted last round enumerates every kind PackageDescription vends and proves
`define` is the only one that spells `-D`, `headerSearchPath` the only
non-`unsafeFlags` one that spells `-I` (and it rejects), `unsafeFlags` rejects,
and every remaining kind is a Swift/Clang language mode, feature name,
diagnostic severity, or isolation default that spells neither. There is no
third macro source: portable mode passes no response file, no `@file`, no
`-include`, and no environment include/define variable, and an unknown kind
rejects twice over.

With the NAME and the BODY of BOTH inputs routed through the same reject logic,
the macro layer is closed across all input surfaces. Combined with the accepted
spelling axis (round 9 M / round 11 N3), position axis (round 11 N1/N2), and
kind axis (round 12 P), the read-invisibility class is closed.

## Compiler evidence — independently re-derived, not accepted on report

`payload.bin` holds the marker `CURATOR_Q_MARKER_PAYLOAD_BYTES`; `SecretKit`'s
only header is `#error SECRET_MODULE_WAS_READ`.

| Probe | Invocation | Exit | Result |
| --- | --- | ---: | --- |
| Q1 | `clang -c -D'A=__asm__' d.c` where `d.c` is `A(".incbin \"payload.bin\"");` | 0 | `d.o` sha256 `70777455b1ceb0c4d743e64f2c51bbcc22fe8b4941ab2802cb580c08e85e5f47` |
| Q1 direct control | the same statement spelled `__asm__(…)` directly, no `-D` | 0 | `direct.o` sha256 **identical**: `70777455b1ceb0c4d743e64f2c51bbcc22fe8b4941ab2802cb580c08e85e5f47` |
| Q1 payload check | `grep -a -c CURATOR_Q_MARKER_PAYLOAD_BYTES d.o` | 0 | 1 occurrence — the named file's bytes are in the object |
| Q1 negative control | the same source, no `-D` | **1, expected failure** | `error: expected parameter declarator` — `A` is undeclared, so the setting is the entire vector |
| Q3 | `clang -fsyntax-only -fmodules -I secret -D'A=_Pragma' p.c` where `p.c` is `A("clang module import SecretKit")` | **1, expected failure** | `While building module 'SecretKit' … error: SECRET_MODULE_WAS_READ` — the undeclared module was read |
| Q3 control | the same source, no `-D` | **1, expected failure** | `expected parameter declarator` — nothing imported |

The two expected-red module probes are red because the marker header is an
`#error`; the read itself is the finding, and the control proves the read does
not happen without the setting.

Probe artifacts under `.temp/TASK-260811-tkurtl/q-probe/` (gitignored).

## Tests — `TestH25BuildSettingDefineBodyIsAnalyzedLikeASourceDefine`

New file section in `internal/swiftpminterop/buildsettings_test.go`. Every
vector is asserted TWICE — once bound by the build setting, once with the
identical body bound by a source `#define` and no setting at all — so the
one-analyzer claim is proven by equal codes rather than asserted.

| Vector | Setting operand | Source control | Use site | Code (both) |
| --- | --- | --- | --- | --- |
| Q1 | `A=__asm__` | `#define A __asm__` | `A(".incbin \"payload.bin\"");` | `swiftpm_target_platform_unsupported` |
| Q1b | `A=asm` | `#define A asm` | same | `swiftpm_target_platform_unsupported` |
| Q3 | `A=_Pragma` | `#define A _Pragma` | `A("clang module import SecretKit")` | `swiftpm_header_input_undeclared` |
| Q4 | `A=__pragma` | `#define A __pragma` | `A(clang module import SecretKit)` | `swiftpm_header_input_undeclared` |
| Q5 | `A=__as##m__` | `#define A __as##m__` | `A(".incbin \"payload.bin\"");` | `swiftpm_target_platform_unsupported` |
| Q6 | `J(a,b)=a##b` | `#define J(a,b) a##b` | `J(__as,m__)(".incbin \"payload.bin\"");` | `swiftpm_header_input_undeclared` |

Plus four more:

- a define body that performs a module import
  (`IMP=_Pragma("clang module import SecretKit")`) rejects
  `swiftpm_header_input_undeclared`;
- an operand this stage cannot separate (`A B=__asm__`) rejects
  `swiftpm_unsafe_build_setting_forbidden`;
- a PRUNED `A=__asm__` setting binds nothing and still admits — the condition
  axis stays ahead of the body axis exactly as it is ahead of the name axis;
- **positive**: a normal target carrying `FEATURE=1`, `MAX=256`,
  `BANNER="portable mode"`, the function-like `SQUARE(x)=((x) * (x))`, and the
  bodiless `NDEBUG` still ADMITS with both plain includes intact
  (`CLib.h`, `stdio.h`).

Every rejected case is asserted to publish no `Result`.

### The vectors are real, not vacuous

Verified by stubbing `analyzeSettingDefineBody` out of the call site and
rerunning: all six `bound by the build setting` subtests plus the module-import
and operand-separation cases FAIL (they admit), while all six source controls,
the pruned case, and the positive still pass. The stub was reverted immediately;
`buildsettings.go` is byte-restored from the pre-stub copy.

## Preservation

All rounds 1-13 acceptance preserved. Focused package: **454 PASS / 0 FAIL**, up
23 from the 431 the reviewer independently reproduced last round (12 Q subtests
+ 4 extras, counted with their parent group lines). Every M, N1/N2/N3, P, H17,
H18, H19, H20, H23, H24, S*, H*, R*, P*, CGP*, CGN* family is green. Normal
legitimate SwiftPM C/C++/ObjC/ObjC++ targets with plain includes, imports,
module maps, and typed boundaries still admit.

Scope hygiene: only `internal/swiftpminterop/headers.go`,
`internal/swiftpminterop/buildsettings.go`, and
`internal/swiftpminterop/buildsettings_test.go` changed this round.
`closuregraph`, `swiftpmsource`, `2qfnai`, and `x611eq` are untouched; the
pre-existing worktree delta is unchanged.

## Evidence

| Gate | Command | Exit | Result |
| --- | --- | ---: | --- |
| Focused (verbose) | `go test -count=1 -v ./internal/swiftpminterop/` | 0 | 454 PASS / 0 FAIL, `ok` 16.452s |
| Race | `go test -count=1 -race ./internal/swiftpminterop/` | 0 | `ok` 69.243s |
| Lint | `golangci-lint run ./...` (v2.12.2, go1.25.5) | 0 | `0 issues.` |
| gofmt | `gofmt -l ./cmd ./internal` | 0 | empty output |
| vet | `go vet ./internal/swiftpminterop/` | 0 | clean |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2`; `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Suite minus `cmd/curator` | `go test -count=1 -timeout 25m $(go list ./... \| grep -v cmd/curator)` | 0 | 51 `ok`, 0 `FAIL` |
| Working tree | `git diff --check` | 0 | clean |
| Board | `task-board --no-update-check validate` | 0 | `Board is valid. No issues found.` |

`gofmt -l ./cmd ./internal` is the scoped form; an unscoped `gofmt -l .` still
reports the pre-existing `.temp/` fixture files, which are gitignored scratch
and not product code.

Not run this round, stated plainly: the monolithic full `go test ./...`
including `cmd/curator` (that package alone needs ~10 minutes and a `-timeout
30m`; it is the Orchestrator's gate, and it is untouched by this round's three
files). Nothing was staged or committed.
