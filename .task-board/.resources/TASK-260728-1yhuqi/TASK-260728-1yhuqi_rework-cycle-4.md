# TASK-260728-1yhuqi — rework cycle 4

Closes the two blocking findings of
`TASK-260728-1yhuqi_review-verdict-cycle-4.md` (reviewer run
`RUN-260729-906f3c`). Every independently accepted closure from cycles 1–3 is
preserved: the SwiftPM rejection, the direct-`swiftc` graph/compile model, the
line-1-only manifest classifier, the toolchain/SDK/native admission rules, the
byte-exact compiler-banner grammar, `curator-swift-module-v1`,
`curator-swift-relpath-v1`, the closed per-job plan token grammar, resolved
containment and permit-time re-binding.

One thing is **retired**, deliberately: cycle 3's manager-executed plan with the
plugin channel deleted. It worked. It is not `manager-worker-v2`.

---

## Finding 1 — the compile phase was multiple commands under a policy that
## admits one

### What the reviewer found

Accepted decision 0008 section 7 closes a `manager-worker-v2` session to at most
one graph command and **exactly one** compile command of exactly one
driver-defined command, with the tool executables started by the driver's own
trusted launcher, and states that a driver which cannot map onto that shape MUST
NOT be admitted and that widening the shape requires another execution-policy
identity, a new claim schema version and its own review. Cycle 3 executed the
plan's jobs directly from the manager while still binding
`execution_policy = "manager-worker-v2"`.

### What closed it

`compile_argv` is executed again, once, after the permit. The manager starts
**exactly two** processes and nothing else; `swiftc` starts its own
`swift-frontend`, `clang` and `ld` children, which is the v2 process graph.

| | manager-started commands | compile-phase parentage |
|---|---|---|
| this contract | **2** — `swiftc -###`, then `swiftc` | `swiftc` → frontend / clang / ld |
| retired cycle-3 design | **4** measured, 1 graph + 3 plan jobs | manager → `swift-frontend`, `swift-frontend`, `clang` |

Both numbers are measured, not asserted: `S62` records them side by side from
one run, and expected-red control `C15` restores the retired design and reports
the count and the parentage it produces.

### What this costs, stated rather than hidden

Cycle 3 could say "the plan verified **is** the plan executed" as a fact about
which processes start. That is now back to an equality of **inputs**: both
commands come from one builder, over one source set, in one environment, and
differ by one inserted token, and every bound path is re-resolved and re-checked
at the permit — but the compile command re-derives its own plan. Reference
section 11 and decision section 14 both record this residual explicitly, and both
name what would close it: an execution-policy identity that admits a
manager-driven job set, which is a decision 0008 change and not a driver
contract's to make.

### What was NOT done

No new execution-policy identity, claim schema version or capability-evidence
version is minted. Decision 0008 is conformed to, not reopened.

---

## Finding 2 — macro use reached a compile child instead of being rejected
## before the permit

### What the reviewer found

Decision 0008 section 7 requires the pre-compile matrix to reject every
package-selected compiler macro **before** the compile phase, under
`build_package_code_execution_forbidden`, and forbids answering such a surface
with a runtime allowance. Cycle 3 removed the load capability but let
`@Observable` pass verification, receive a compile permit, and fail inside the
first frontend job. The matrix claimed a pre-compile rejection it did not
implement.

### What closed it: `curator-swift-source-admission-v1`

A Stage-B rule over the raw bytes of the compiled source set, applied after
enumeration and **before the graph command runs**:

| Rule | Check | Diagnostic |
|---|---|---|
| A1 | readable regular file | `swift_source_unreadable` |
| A2 | well-formed UTF-8 per RFC 3629, no NUL | `swift_source_encoding_forbidden` |
| A3 | no byte `0x40` (`@`) or `0x23` (`#`) at any offset, in any context | `swift_source_macro_selector_forbidden` |

All three report `build_package_code_execution_forbidden`. A rejected source set
starts **no** command: no graph phase, no plan, no permit, no compile child, no
artifact.

The rule does not parse Swift, does not skip comments or string literals, does
not normalize, and reads nothing a package supplies except the source bytes. It
is a byte scan, so it is total over every possible file.

**A2 precedes A3 and the order is normative.** A3's claim is "no `0x40` byte
implies no U+0040 code point", which holds only on well-formed UTF-8.

### Why those two bytes are the whole surface — measured, not argued

The reviewer asked for a rule proven from the admitted Swift grammar. The
stronger available evidence is that **the compiler states both requirements
itself**. All on Apple Swift 6.3.2 / macOS 26.5 arm64:

| Vector | Measured |
|---|---|
| `externalMacro(module:type:)` with no sigil | `error: expansion of macro 'externalMacro(module:type:)' requires leading '#'` |
| a `macro` declaration with no role attribute | `error: macro 'stringify' must declare its applicable roles via '@freestanding' or '@attached'` |
| `Observable final class Box {}` | not an attribute; parse error |
| `\u{40}Observable` | escape outside a literal is not syntax; parse error |
| `＠Observable` (U+FF20) | identifier character, not an attribute marker; parse error |
| overlong UTF-8 for U+0040 (`0xC1 0x80`) | carries no `0x40` byte; rejected by A2, and independently by the compiler with `invalid UTF-8 found in source file` |
| `import Observation` with no macro use | compiles, **0** macro-load remarks |
| the admitted rich-Swift source set | compiles under one `swiftc` command, plan carries **5** plugin components, **0** load remarks, artifact runs |

Swift has exactly two macro-use spellings — attached (a custom attribute, needs
`0x40`) and freestanding (an expansion, needs `0x23`). A source set carrying
neither byte cannot name a macro.

### The collateral is inventoried rather than discovered

The rule over-rejects on purpose. Reference 3.3 lists what goes with it:
`@main`, `@available`, `@escaping`, `@inlinable`, the `@Sendable` attribute
spelling, `@objc`, `@_cdecl`, `@_silgen_name`, every property-wrapper use;
`#if`/`#elseif`/`#endif`, `#available`, `#file`, `#line`, `#function`,
`#selector`, `#keyPath`, raw string literals `#"…"#`, extended regex literals
`#/…/#`; and any sigil inside a comment or a string literal.

Two rows moved from **admit, bounded** to **reject**: `#if` conditional
compilation in source, and `@_cdecl` / `@_silgen_name`. Section 14 of the
decision and section 11 of the reference record the change rather than dropping
the old entries.

What stays admitted is measured, not promised: the standard library,
`Foundation`, `Codable` and `Sendable` conformances, `actor`, `async`/`await`,
generics and `where` clauses, bare regex literals, string interpolation,
multi-line strings, custom operators, protocols, extensions and Unicode
identifiers. Vectors `SA01`–`SA20` assert the admitted half.

### The plugin channel is now stated as inert rather than absent

With one compile command the manager cannot edit the frontend jobs, so the
injected plugin search paths stay in the plan. The contract does not pretend
otherwise. Two independent conditions are checked and neither is asked to carry
the other:

- **closure** — every plugin path in the plan resolves inside a fingerprinted
  root or is absent (4.1.3), re-checked at the permit (4.1.4);
- **selection** — no admitted source set can name a macro (3.3).

Measured inertness: 5 plugin components in the plan, 0 load remarks, artifact
runs. The residual — a toolchain that begins loading a plugin with no source-side
selector — is named in both documents and bounded by the closure check rather
than eliminated.

---

## Rejected alternatives added

- Keep cycle 3's design. Rejected: measured 4 manager commands and changed
  parentage; keeping it means mislabelling the session or reopening decision
  0008 from inside a driver contract.
- Mint a third execution-policy identity. Rejected **here**, not on the merits;
  it is a decision 0008 change with its own review, and section 14 keeps the
  option visible.
- Detect macro use by asking the compiler. Rejected on a cycle-3 measurement:
  `-scan-dependencies` reports `macroDependencies` as the module's dependency
  *closure*, not its use, so a macro-free file still lists `SwiftMacros`.
- Reject only `@` at token start, or only outside comments and literals.
  Rejected: both need a Swift lexer in the manager, and Swift's raw-string
  delimiters make "inside a literal" itself a `#`-counting problem.
- Rely on the compiler rejecting a macro it cannot load. Rejected: that is the
  cycle-3 state the review named — a runtime allowance after a permit.

---

## Probe changes

| Item | Before (cycle 3) | After (cycle 4) |
|---|---|---|
| structural checks | 56 | **62** |
| expected-red controls | 14 | **15** |
| cases / closure checks | 23 / 32 | 23 / 32, unchanged |
| non-test Go files / test files | 19 / 8 | **21 / 9** |
| test functions / lines | 55 / 7841 | **63 / 8662** |

New: `sourceadmission.go` (the rule, the retired variant for `C13`, the
hand-written RFC 3629 validator), `sourceadmission_test.go` (`SA01`–`SA20`,
`SR01`–`SR22`, `SE01`–`SE10`, a standard-library agreement test, a totality test
over every single-byte and `let <b>` body, an unreadable-source test and a
path-order determinism test), `structural_cycle4.go` (`S54`–`S64` and the
`BuildOutcome` session helper that records exactly which processes the manager
started).

Retired: `S46`–`S50`, which measured the manager-executed plan. `S42`–`S44` are
kept — they are the negative measurements that show the channel cannot be
suppressed by argv, environment or presentation, which is now the reason the
surface is closed at the source end. `S45` is kept and retitled: it measures what
a macro-selecting source does once the contract's own compile command has
started, which is exactly what `S55` proves never happens.
`StripPluginChannel` / `ExecutePlan` survive only so `C15` can restore the
retired design from the same binary.

`C13` keeps its identifier and changes its premise: it now restores the retired
**Stage B** (no source admission) rather than the retired deletion, and reports
that the macro-selecting source is admitted, receives a permit, and loads
`libObservationMacros.dylib`. The guard text records the change.

### New structural checks

| ID | Asserts |
|---|---|
| `S54` | the rule scans every byte of every compiled source and admits ordinary Swift |
| `S55` | `@Observable` rejected at Stage B, `build_package_code_execution_forbidden`, offset 19, byte `0x40`, **0** commands, no artifact |
| `S56` | `#Predicate` rejected on byte `0x23`, **0** commands |
| `S57` | the compiler itself requires `#` to expand and `@` to declare a macro |
| `S58` | homoglyph, Unicode escape and bare attribute name are all not attributes; the homoglyph file is deliberately *admitted* by the rule and still cannot select a macro |
| `S59` | overlong UTF-8 for U+0040 carries no `0x40`, is rejected by A2, and is independently rejected by the compiler |
| `S60` | the admitted set builds under **2** manager commands with **5** plugin components in the plan and **0** load remarks; artifact runs |
| `S61` | `import Observation` with no macro use loads nothing |
| `S62` | this design starts 2 commands; the retired one starts 4 |
| `S63` | both commands are one program differing by `-###` inserted at index 1, exactly once, never in the compile vector |
| `S64` | two full sessions to one output path give one digest |

`-Rmacro-loading` is **not** a member of either contract vector — the closed
token grammar rejects the `-Xfrontend` pass-through that carries it — so every
remark count above comes from a separate evidence-only compile of the same source
to a throwaway path. That is stated in the probe, in reference 4.2.3 and here,
so a reviewer does not have to infer it.

---

## Results

| Gate | Result |
|---|---|
| probe `gofmt -l` / `go vet` / `go test -count=1` / `go build` | exit 0, 0, 0, 0 |
| native run | 23 cases / 23 matched / 0 divergences; 32 closure checks / 0 verdicts; **15 of 15** controls red; **62** structural checks / 0 divergences; executed P2 admission ok; `green: true`; exit 0 |
| degraded run (no resolvable toolchain) | 23 `not_run` with the reason recorded, exit 0, nothing installed |
| each control replayed individually | `C1`–`C15`, every one exit 1 |
| tarball round-trip | 21 non-test / 9 test files, 63 test functions, 8662 lines, `go test` exit 0; sha256 `3446bfa42c24bad7102982b0b5e117f9a7f31cf9b3956c45eed37bb1510ba16b` |

### Expected-red gates, attributed

- **curator repo `gofmt -l .`** — exit 2, 754 paths. **0** under
  `.temp/TASK-260728-1yhuqi`, **0** outside `.temp/`, **0** modified tracked
  files. The paths are other tasks' scratch trees.
- **spec `validate.py` in the task worktree** — exit 1, failing only its link
  check on `docs/external-build-repositories.md`. The clean baseline at
  `57c1f56` exits 0 (30 schemas, 93 vector files). Scoped over the two documents
  this task authored: `docs/swift-build-drivers.md` 4 links / 0 broken,
  `decisions/0011-swift-driver-pair.md` 2 links / 0 broken. The broken links are
  in two documents this task did not author (3 links total), whose targets do not
  exist at `57c1f56`.

### Hygiene

Spec worktree: 0 staged, 0 tracked modifications. Curator repo: 0 tracked
modifications. Nothing staged, committed, pinned, published or installed on any
host. No platform claim widened: macOS arm64 remains the only measured tuple,
Windows remains an implementation contract with no claim, Linux remains deferred.

---

## Open items carried forward

Unchanged from cycle 2, and neither is affected by this rework:

1. the Windows plan-derived closure member **count** is unmeasured — an
   implementation takes it from the plan it verifies;
2. the beyond-ordinal-0 SDK argument template for a multi-root platform is
   unmeasured — minting it is part of the Windows obligation.

New, and stated rather than closed:

3. the compile phase re-derives its own plan, so "the plan verified is the plan
   executed" is an equality of inputs rather than of processes. Closing it needs
   an execution-policy identity that admits a manager-driven job set, which is a
   decision 0008 change.
