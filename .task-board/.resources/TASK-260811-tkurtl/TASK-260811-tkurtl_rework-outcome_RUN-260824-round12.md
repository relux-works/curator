# TASK-260811-tkurtl — rework outcome, round 12 (finding P: the build-setting macro oracle)

Role: developer. Model: Claude `claude-opus-5`.
Reviewed verdict answered: `TASK-260811-tkurtl_review-verdict_RUN-260824-74cbb4.md`.
Compiler for every probe: Apple clang version 21.0.0 (`clang-2100.1.1.101`),
`arm64-apple-darwin25.5.0`. SwiftPM/Swift for every manifest probe: Swift 6.3.2
(`swiftlang-6.3.2.1.108`), Swift Package Manager 6.3.2 — the accepted Darwin
profile.

Scope: finding P only. N1, N2, N3, the step-4 position table, and every
accepted item from rounds 1-11 are preserved unchanged.

Product code touched: `internal/swiftpminterop/buildsettings.go` (new),
`internal/swiftpminterop/interop.go`, `internal/swiftpminterop/headers.go`
(comment only), `internal/swiftpminterop/platform.go`. Tests:
`internal/swiftpminterop/buildsettings_test.go` (new). `closuregraph`,
`swiftpmsource`, `2qfnai`, and `x611eq` are untouched. Nothing staged or
committed.

`IncludeGrammarID` stays at `c-family-include-scanner-v10`: the scanner grammar
over a file is byte-unchanged this round. What changed is the closure-level
macro oracle and a build-setting axis outside the include grammar, so bumping
the grammar identity would misattribute the change.

---

## Compiler evidence first

Full log: `TASK-260811-tkurtl_round12-clang-evidence.log`. `SecretKit` is a
module whose only header is `#error SECRET_MODULE_WAS_READ`, so "the header was
read" is an observable, not an inference.

| Probe | Invocation | Exit | Result |
| --- | --- | ---: | --- |
| D1 | `-Dprotocol=import` + `@ protocol SecretKit;` | 1 | SecretKit built, `secret.h` read |
| D1 control | same source, no `-D` | 0 | clean; nothing imported |
| D2 control | `@import NoSuchKitXYZ;` alone | 1 | `module 'NoSuchKitXYZ' not found` |
| D2 | `-DNoSuchKitXYZ=SecretKit` + `@import NoSuchKitXYZ;` | 1 | SecretKit built, `secret.h` read; clang prints `note: expanded from macro 'NoSuchKitXYZ'` at `<command line>:1` |

The D2 note is the finding in one line: the macro that redirects the module
import is bound on the command line, not in any admitted file, so an oracle
built from source `#define`s cannot see it. Both closing rules from round 11
therefore reproduced one level down, exactly as the verdict states.

The build-setting kind axis was enumerated the same way rather than from
memory. Full log: `TASK-260811-tkurtl_round12-swiftpm-setting-kinds.log`. A
manifest declaring every kind `PackageDescription` vends was run through
`swift package dump-package`; the emitted encoding is
`{"kind":{"<name>":{"_0":…,"_1":…}},"tool":"c|cxx|swift|linker"}` with an
optional `"condition"` member, `_0` a string for every kind except
`unsafeFlags` (an array) and absent for the nullary `strictMemorySafety`.
Verified members, fourteen with a distinct encoding:

```
define  headerSearchPath  unsafeFlags  linkedLibrary  linkedFramework
interoperabilityMode  enableUpcomingFeature  enableExperimentalFeature
swiftLanguageMode  treatAllWarnings  treatWarning  enableWarning
disableWarning  strictMemorySafety  defaultIsolation
```

`swiftLanguageVersion` is the deprecated PackageDescription spelling and SwiftPM
6.3.2 already serializes it as `swiftLanguageMode` (verified under tools version
6.0); it is retained in the table only against an older serializer.

---

## The fix — the stronger reject-by-default close (the verdict's option 2)

The `define` member is not patched in isolation. The whole kind axis is
enumerated and closed reject-by-default, the way the pragma axis was.

`swiftpmsource.decodeBuildSetting` is unchanged, so no canonical capture record
moves: it already retains the raw setting JSON verbatim in `BuildSetting.Value`,
and `decodeSettingKind` (`internal/swiftpminterop/buildsettings.go`) recovers
the real kind and its operands from there. A `Value` that is not a JSON object
is a directly constructed record; its declared `Kind`/`Value` are then the kind
and its single operand, which is the identity the fixture path had before.

### Per-kind disposition

A kind is admitted only when it is provably BOTH **macro-inert** (it cannot bind
a preprocessor macro) and **resolution-inert** (it cannot change where a file is
found or read). The axis is small enough to prove rather than enumerate:
exactly one non-`unsafeFlags` kind reaches the compiler as `-D` and exactly one
reaches it as `-I`.

| Kind | macro-inert | resolution-inert | Disposition |
| --- | :---: | :---: | --- |
| `define` | **no** | yes | **Routed.** Every define NAME is fed into both oracles: a name in `atPositionIdentifiers` rejects (`swiftpm_header_input_undeclared`, the N1 route), and every name is unioned into the macro set `rejectMacroDefinedModuleNames` consumes (the N2 route). Otherwise admitted — an ordinary define is legitimate and stays legitimate. |
| `headerSearchPath` | yes | **no** | **Reject** (`swiftpm_unsafe_build_setting_forbidden`). It is the only non-`unsafeFlags` kind that reaches the compiler as `-I`, and portable mode's include closure resolves a reference against the target's own roots, so it cannot follow the resolution the flag creates. This narrows accepted input in the safe direction, consistent with the accepted scope statement. |
| `unsafeFlags` | **no** | **no** | **Reject.** Unbounded on both axes. It now rejects on its own kind name, so the record's `Unsafe` flag is corroboration rather than the only gate. |
| `linkedLibrary` | yes | yes | **Admitted iff declared.** Reaches the linker, not the compiler, so both axes are closed; still gated through the same `linkDeclared` rule `confineLinks` applies to a module-map link edge — an undeclared external library is `artifact_toolchain_untrusted` under the accepted SwiftPM outcome. |
| `linkedFramework` | yes | yes | Same, on the framework list. |
| `interoperabilityMode` | yes | yes | **Admitted-inert**, and now actually consumed: see below. |
| `enableUpcomingFeature` | yes | yes | Admitted-inert. Maps to `-enable-upcoming-feature <name>`; carries no `-D` and no `-I` for any operand value, which is why an open-ended feature name does not reopen either axis. |
| `enableExperimentalFeature` | yes | yes | Admitted-inert, same argument. |
| `swiftLanguageMode` (and the deprecated `swiftLanguageVersion` spelling) | yes | yes | Admitted-inert. Closed language-version enum. |
| `treatAllWarnings`, `treatWarning`, `enableWarning`, `disableWarning` | yes | yes | Admitted-inert. Diagnostic severity only. |
| `strictMemorySafety` | yes | yes | Admitted-inert. Nullary. |
| `defaultIsolation` | yes | yes | Admitted-inert. Isolation default only. |
| any kind the table does not name | unprovable | unprovable | **Reject.** `settingReject` is the zero value of `settingDisposition` AND an explicit membership check runs, so a kind a later SwiftPM release adds fails closed twice over rather than reading as inert. |
| a payload whose encoding this stage cannot read | unknown | unknown | **Reject.** No `kind` member, two kinds, a non-object `kind`, or a non-string/non-array operand: an unknown shape is an unknown effect on both axes. |

Conditions are the one axis where the oracle is deliberately narrower than the
declaration: a setting the destination prunes never reaches a compiler
invocation, so it binds nothing. This mirrors the pre-existing unsafe gate. Note
that `decodeBuildSetting` does not populate `BuildSetting.Condition` at all from
real dump JSON, so on the real path every setting reads as selected — maximum
rejection, the safe direction.

### Where it is enforced

- `disposeBuildSettings` (`buildsettings.go`) runs inside `classifyTargets`, for
  a selected target only, and replaces the old inline unsafe loop (same
  diagnostic, same fields, same condition semantics). It returns the target's
  define names.
- `closeState.settingDefines` carries them, keyed by `package:target`.
- `scanAndConfineIncludes` seeds the per-target macro set from that map before
  the first file is opened, so `rejectMacroDefinedModuleNames` sees the union of
  source `#define`s and selected `.define` settings.
- The `atPositionIdentifiers` rejection for a build-setting define lands in
  `classifyTargets`, i.e. before any file is scanned and before any process
  could start.

### The step-4 comment now states the oracle's provenance

`scanIncludes`'s phase-4 argument previously enumerated identifier *positions*
exhaustively and left the *oracle* behind two of those rows implicit. It now
says, at the table:

- the oracle is the union of two inputs — every source `#define` the scanned
  closure binds (via `noteMacro`), and every `.define` build setting the
  destination selects (via `disposeBuildSettings`);
- why input 2 exists, with the D1/D2 compiler results inline; and
- that the build-setting axis behind input 2 is itself enumerated
  reject-by-default, so the oracle cannot be bypassed by a channel that binds a
  macro without spelling `#define`.

The two affected table rows were reworded to name the binding site ("a `#define`
or a `.define` build setting") rather than only `#define`.
`rejectMacroDefinedModuleNames`'s own comment carries the same statement.

### One correctness repair the axis exposed

`cxxInteropSetting` read `BuildSetting.Kind`, which `decodeBuildSetting` folds
to `swiftpm-setting` for every real manifest. The C++ interop gate could
therefore see an `.interoperabilityMode(.Cxx)` declared by a directly
constructed fixture record but never one SwiftPM actually emitted. It now reads
the decoded kind. No existing vector changes behavior (no fixture used the real
encoding); `TestH24InteroperabilityModeIsReadFromTheDecodedKind` pins both
encodings and the two negative cases.

---

## Vectors

New family `H24`, 39 subtests, `internal/swiftpminterop/buildsettings_test.go`.
Every setting payload is attached in the exact shape `decodeBuildSetting`
produces (folded `swiftpm-setting` kind, raw JSON in `Value`, `Unsafe` by the
same substring rule) and every payload is a literal copy of real
`dump-package` output.

**Reproduced through `Close()` on the fixture's own admitted `root:CLib` target:**

| Vector | Body + setting | Verdict |
| --- | --- | --- |
| D1 | `@ protocol SecretKit;` + `{"kind":{"define":{"_0":"protocol=import"}},"tool":"c"}` | `swiftpm_header_input_undeclared` — *"build setting binds an identifier the compiler expands in the `@`-keyword position…"*; no `Result` published |
| D2 | `@import CLib;` + `{"kind":{"define":{"_0":"CLib=SecretKit"}},"tool":"c"}` | `swiftpm_header_input_undeclared` — *"Clang module import names a macro-defined identifier…"*; no `Result` published, so no wrong module is recorded |
| D2 control | the identical source with no setting | admits; records exactly one module import, spelling `CLib` |
| pruned condition | D1's setting under `platform=linux` | admits — a pruned setting binds nothing |

Both rejections were confirmed to travel their intended route by printing the
message, not just the code: D1 through the N1 route at the setting, D2 through
the N2 route with the build-setting-seeded oracle. Neither is an incidental
rejection from some other gate.

**Per-kind disposition vectors** (one subtest per row of the table above, 27 in
total), including: `define` plain / keyed / on the `swift` tool / binding an
`@`-position identifier / binding the `import` spelling / naming no identifier;
`headerSearchPath`; `unsafeFlags` (asserted through the kind name);
`linkedLibrary` and `linkedFramework` both declared by the SDK component and
undeclared (`artifact_toolchain_untrusted`); each of the nine inert kinds; an
unnamed future kind; and four unreadable payload shapes.

**Positive path:** a normal C target carrying a benign `.define("CLIB_FEATURE",
to: "1")`, a `cxx`-tool `.define("NDEBUG")`, a declared `.linkedLibrary("c")`, a
declared `.linkedFramework("Foundation")`, and `.treatAllWarnings(as: .error)`
admits with both of its plain includes (`CLib.h`, `stdio.h`) intact.

Every accepted positive path from rounds 1-11 is preserved: `S02`, `S03`,
`S05`-`S08`, `S10` (case-sensitive `.C`/`.M`), `H01`-`H23` including all of
N1a-N1i, N2/N2b/N2c, N3a-N3e and their positive controls, `CGP05` with its
conditional branch, `CGN*`, `R*`/`P*`, the transitive worklist, the module-map
out-of-root guard, `publicHeadersPath`, and the cross-package include closure.

---

## Evidence

Run directly as standalone processes; real exit codes; no `tee`, no pipe in any
gate command.

| Gate | Command | Exit |
| --- | --- | ---: |
| Focused, verbose | `go test -count=1 -v ./internal/swiftpminterop/` | 0 — **431 PASS / 0 FAIL** (up from 392) |
| Focused, race | `go test -count=1 -race ./internal/swiftpminterop/` | 0 |
| Lint (pinned v2.12.2) | `golangci-lint run ./...` | 0 — `0 issues.` |
| Vet | `go vet ./...` | 0 |
| gofmt | `gofmt -l ./cmd ./internal` | 0 — no files listed |
| Canonical goldens | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 — `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2`, `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Suite minus `cmd/curator` (one bounded call) | `go test -count=1 $(go list ./... \| grep -v '/cmd/curator$')` | 0 — 51 `ok`, 0 FAIL |
| Whitespace | `git diff --check` | 0 |
| Board | `task-board --no-update-check validate` | `Board is valid. No issues found.` |

`gofmt -l .` also lists 79 files, every one of them under `.temp/` — gitignored
scratch and vendored trees belonging to other tasks, none of them tracked. The
gate above is therefore scoped to the tracked source trees; the unscoped
invocation and its full output are recorded in
`TASK-260811-tkurtl_round12-gofmt-unscoped.log` rather than presented as clean.

`cmd/curator` was not run in this round: it takes ~10 minutes per run against a
10-minute per-call cap, and this change touches no package it exercises. The
monolithic full suite is the Orchestrator's gate. This is an explicit
not-run, not a pass.

## Scope hygiene

`git status --short -- internal/` shows only the pre-existing `closuregraph`
and `swiftpmsource` deltas from earlier rounds; neither was touched this round.
`internal/swiftpminterop/` is this task's untracked delivery; within it only
`buildsettings.go`, `buildsettings_test.go`, `interop.go`, `headers.go`, and
`platform.go` carry this round's changes. Nothing is staged or committed.
