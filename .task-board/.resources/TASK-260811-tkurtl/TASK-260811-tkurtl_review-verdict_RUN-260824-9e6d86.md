# Reviewer verdict for TASK-260811-tkurtl — finding Q (build-setting define body), intended terminal review

Verdict: **accepted -> `done`**

- Reviewer run: `RUN-260824-9e6d86`
- Model: Claude `claude-opus-5`
- Reviewed delivery: `TASK-260811-tkurtl_rework-outcome_RUN-260824-round13.md` (RUN-260824-37dfc9)
- Answering verdict: `TASK-260811-tkurtl_review-verdict_RUN-260824-92eab0.md`
- Compiler for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
  `arm64-apple-darwin25.5.0` — the accepted Darwin profile.
- No product code, board file, or git state was modified. The only file I wrote
  and reverted was a byte-restored `buildsettings.go` after a vacuity probe;
  `diff` confirms it is identical to the delivered file and `git status` is
  unchanged apart from the pre-existing worktree delta.

Finding Q — the last accepted verdict's single blocker, the `.define`
build-setting **body** discarded while the source `#define` body is analysed —
is closed, and closed the way the verdict asked for: the body is routed through
the SAME analyzer the source route uses, not a forked second one. With it, the
read-invisibility class is closed across all four of its axes. **Accept.**

---

## Item 1 — finding Q closed, one analyzer, not a fork: accepted

The fix factors `analyzeMacroBody` (`headers.go:1317`) and `readMacroParameters`
(`headers.go:1278`) out of `readMacroDefinition` and calls both from
`analyzeSettingDefineBody` (`buildsettings.go:199`). `analyzeMacroBody` is
exactly `collapseMacroPastes` (which already covers `##` and the `%:%:` digraph
from N3) followed by `scanDirectiveChannels` over the collapsed stream — the
finding-M pair. `readMacroDefinition` now has no body logic of its own, so the
two routes cannot drift. `splitSettingDefine` parses the `-D` operand into name,
optional parameter list, and replacement list; a define with no `=` binds `1`
and stays admitted; an operand it cannot separate rejects
`swiftpm_unsafe_build_setting_forbidden` rather than being guessed at.

Reproduced through the real `Close()` path on the fixture's own admitted
`root:CLib` target (`TestH25…`, 12 subtests + 4 extras). Every Q vector is
asserted TWICE — once bound by the setting, once with the identical body bound
by a source `#define` and no setting — and both assert the SAME code:

| Vector | Setting operand | Source control | Code (both) |
| --- | --- | --- | --- |
| Q1 | `A=__asm__` | `#define A __asm__` | `swiftpm_target_platform_unsupported` |
| Q1b | `A=asm` | `#define A asm` | `swiftpm_target_platform_unsupported` |
| Q3 | `A=_Pragma` | `#define A _Pragma` | `swiftpm_header_input_undeclared` |
| Q4 | `A=__pragma` | `#define A __pragma` | `swiftpm_header_input_undeclared` |
| Q5 | `A=__as##m__` | `#define A __as##m__` | `swiftpm_target_platform_unsupported` |
| Q6 | `J(a,b)=a##b` | `#define J(a,b) a##b` | `swiftpm_header_input_undeclared` |

Q1 and Q6 are round-9 finding-M's own vectors (`__as##m__`, `J(__as,m__)`)
reproduced one level down, exactly as N1/N2 reproduced their round to finding P.
A body that resolves to a module import (`IMP=_Pragma("clang module import
SecretKit")`) rejects `swiftpm_header_input_undeclared` — correctly, because a
build setting belongs to no scanned unit, has no directory to resolve against,
and cannot join the include worklist; under reject-by-default it is refused, not
confined. Every rejected case asserts a nil `Result`.

**Non-vacuity verified independently.** I stubbed the `analyzeSettingDefineBody`
call site out of `disposeBuildSettings` and reran `TestH25`: all six
`bound by the build setting` subtests plus the module-import and
operand-separation cases FAILED (they admitted), while all six source controls,
the pruned case, and the positive stayed green. Reverted immediately; the file
is byte-identical to the delivery. The vectors are real, not decorative.

**Compiler evidence re-derived.** `-D'A=__asm__'` with `A(".incbin
\"payload.bin\"");` produces an object byte-identical to the direct `__asm__(…)`
control (both sha256 `70777455…`) with the named file's bytes inside it;
`-D'A=_Pragma'` with `A("clang module import SecretKit")` builds an undeclared
module and reads its `#error`-marker header. Both source spellings already
reject. This is a route gap the fix closes, not a grammar gap.

## Item 2 — macro-input surface exhausted, no third input: accepted (adversarially audited)

The outcome's terminal claim is that source `#define` and build-setting
`define` are the ONLY two macro-binding inputs the pinned `clang -c` honours for
an admitted C-family target. I attacked that claim on every candidate the brief
named and could not find a third:

- **prefix/force-include header (`-include`/`-imacros`)** — no PackageDescription
  build-setting kind spells either; the round-12 kind axis enumerates every kind
  and only `define`→`-D` and `headerSearchPath`→`-I` (rejected) reach the
  compiler as flags, with `unsafeFlags` (rejected) the sole escape. A prefix
  header is therefore reachable only through `unsafeFlags`, which rejects. Not a
  third input.
- **response file / `@file`** — portable mode passes exact argv and generates no
  response file; not attacker-controllable, and the executor is 2qfnai's surface
  regardless.
- **environment-driven define / include vars** — cleared per the accepted
  outcome (H05 `swiftpm_unsafe_build_setting_forbidden` / env-clearing).
- **module-map `config_macros`** — parsed by `skipIdentifierList`
  (`modulemap.go:331`) as a bare macro-NAME list. It declares which macros a
  module's build depends on; it binds no value, so it cannot carry a channel
  body. Not a macro-binding input.
- **compiler builtins / SDK-header macros** — builtins are compiler-fixed and not
  attacker-controllable. SDK macros such as `__DARWIN_ALIAS(sym)→__asm(…)` and the
  `API_AVAILABLE_BEGIN→_Pragma(…)` family are real, but they live in the selected,
  fingerprinted SDK/toolchain root that portable mode admits-by-selection without
  scanning (`systemHeaderDeclared`/`moduleDeclared` return without joining the
  worklist). That is the accepted H06/H07 trusted-root architecture, and the asm
  strings the attacker does not control are symbol labels, not `.incbin` file
  reads. Not an attacker-controllable input reaching an admitted target.

The macro-input surface is exhausted: two inputs, each with its NAME (M/N1/N2
for source, P for setting) and its BODY (M for source, Q for setting) routed
through the same reject logic.

## Item 3 — whole-class closure: confirmed (acceptance-deciding statement)

With SPELLING (round 9 M / round 11 N3), POSITION (round 11 N1/N2/N3), KIND
(round 12 P / build-setting kind axis), and ORACLE-INPUT — the name and body of
both macro-binding inputs — all closed, and phases 1-3 (trigraphs H14, line
splices H19, comments, digraphs) reproduced so a keyword cannot be reconstituted
past the scanner, the read-invisibility class is closed for the pinned compiler
under the accepted reject-by-default posture. No admission hole survives on any
axis I could execute.

## Item 4 — positive path and no regression: accepted

Reproduced independently: `go test -count=1 -v ./internal/swiftpminterop/` →
**454 PASS / 0 FAIL**, matching the reported count (up 23 from the 431 last
round). Every accepted family present and green: S02/S03/S04/S05/S06/S07/S08/
S09/S10, H12–H25, CGP05 (all three, selection-neutral + conditional-edge), CGN*
(03/09/15…), M/N1/N2/N3, H17/H18/H19/H20/H21/H22/H23/H24. The benign positive
(`FEATURE=1`, `MAX=256`, `BANNER="portable mode"`, function-like
`SQUARE(x)=((x) * (x))`, bodiless `NDEBUG`) ADMITS with both plain includes
(`CLib.h`, `stdio.h`) intact; `linkedLibrary`/`linkedFramework` still admit
through the SDK while undeclared links reject. Normal C/C++/ObjC/ObjC++ targets
with plain includes, imports, module maps, and typed boundaries (incl.
case-sensitive `.C`/`.M`) still admit.

Scope hygiene intact: only `headers.go`, `buildsettings.go`, and
`buildsettings_test.go` carry the round-13 delivery timestamps; `closuregraph`
(02:24) and `swiftpmsource` (05:17) are the pre-existing worktree delta,
untouched by this round; `2qfnai`/`x611eq` are untouched. No package outside
`swiftpminterop` imports it, so the shared-analyzer refactor has zero external
blast radius. `IncludeGrammarID` correctly holds at
`c-family-include-scanner-v10`: the scanner grammar over a file did not change,
only which inputs are fed to it.

## Item 5 — evidence: accepted

Reran by me: focused verbose (454/0), the H25 vacuity probe, `gofmt -l ./cmd
./internal` (clean), `go vet ./internal/swiftpminterop/` (0),
`golangci-lint run ./internal/swiftpminterop/` (`0 issues.`),
`go test -race ./internal/swiftpminterop/` (`ok` 68.523s), the canonical golden
verifier (`canonical_goldens=pass labeled_records=53`;
`canonical_references=pass`), the compiler probe matrix above, and
`task-board validate` (`Board is valid. No issues found.`).

Accepted from the orchestrator-attached hash-bound monolith:
`TASK-260811-tkurtl_full-go-13.log` — I verified sha256
`c7ae60feb005b9a235348d31cd0fa95c861ba7dbb2b1f7b8482d94655c2bcd86` matches the
brief exactly, `EXIT:0`, 52 `ok` packages, 0 `FAIL`. `cmd/curator` (its ~10-min
package) is covered there and is untouched by this round's three files.

---

## Routing

All Definition-of-Done rework items through round-13 finding Q are satisfied and
independently verified. Accept -> `done`.

As a reviewer-archetype run this supplies no `commit_ack`: acceptance evidence
is recorded here for the commit-owning mover, which commits this round's three
`swiftpminterop` files and makes the final `done` transition with
`commit_ack=scope_committed`. No stop-the-line condition.
