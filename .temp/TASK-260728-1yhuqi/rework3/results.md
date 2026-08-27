# TASK-260728-1yhuqi — Swift driver-pair security contract

Handoff. Ready for review, **cycle 4**.

> **Rework cycle 3 applied.** Reviewer cycle 3 independently replayed every
> cycle-2 repair green and returned CHANGES REQUESTED on two security-contract
> defects. Both are closed; the SwiftPM rejection, the direct-`swiftc`
> architecture, every cycle-1 closure and every verified cycle-2 repair are
> unchanged. The full record is `TASK-260728-1yhuqi_rework-cycle-3.md`. This
> document is updated in place so its numbers match the current artifacts.
>
> **F1 — the contract admitted the compiler macros decision 0008 requires it to
> reject.** Measured first, because the reviewer's proposed closure had to be
> shown achievable: no `-resource-dir` override, no toolchain presentation, no
> in-process-server override and no compiler flag suppresses a source-selected
> toolchain macro load while `swiftc` runs its own jobs. The decisive positive
> measurement is that the channel is entirely flag-driven — the same frontend job
> with no plugin flag rejects `@Observable` and loads nothing. So the manager now
> **executes the plan it verified**, job by job, with the closed plugin-channel
> token set deleted and five assertions holding, and `compile_argv` is no longer
> executed. `@Observable` fails closed with 0 load remarks; ordinary Swift,
> Foundation, `Codable` and regex literals are unaffected; determinism holds.
> Control `C13` restores the retired policy and reports the admission. Decision
> 0008 is conformed to, not reopened.
>
> **F2 — unknown flag channels were accepted.** Totality is now over every token
> rather than over path-shaped ones, under a closed per-job flag and operand
> grammar with an opaque-value rule and measured `-Xllvm`/`-Xcc` allow-sets.
> Measured: this toolchain really does define the joined
> `-load-pass-plugin=<lib>`; 16 of 16 unknown-channel negatives reject, the
> retired verifier admits 14 of them, and the live 101-token plan still verifies
> with 0 rejections. Control `C14` reports the gap.

> **Rework cycle 2 applied.** Reviewer cycle 2 confirmed every cycle-1 closure and
> independently replayed the probe, then returned CHANGES REQUESTED on five
> implementation-readiness blockers. All five are closed; the direct-`swiftc`
> architecture, the SwiftPM rejection and every cycle-1 closure are unchanged.
> The full rework record is `TASK-260728-1yhuqi_rework-cycle-2.md`. This document
> is updated in place so its numbers match the current artifacts.
>
> In short: **R1** one byte-exact whole-value compiler-banner grammar, stated
> once and consumed entirely, with 32 vectors and control `C10`; **R2** probe P2
> is now actually executed for the exact compile triple and gates green, under a
> closed three-class runtime-library rule whose base-installation class exists
> because the measured host returns `/usr/lib/swift`, with 10 negatives and
> control `C11`; **R3** `curator-swift-relpath-v1` gives one portable
> `(role, relpath)` serialization for the P3-resolved linker, plan-derived
> members and one-or-more SDK roots, so the Windows identity is implementable
> without a Windows claim; **R4** the `-###` insertion is stated as a
> construction with seven asserted properties, and the plan's physical-line
> grammar is total — LF-only, mandatory terminal LF, bare CR rejected — with
> control `C12`; **R5** `TASK-260729-rhjxtx` is linked `blocked_by` and its four
> inherited measurements are traced individually in both documents.
>
> Cycle-1 record: `TASK-260728-1yhuqi_rework-cycle-1.md`. In that cycle the
> classifier became line-1-only, the `swiftc -###` verifier became a fail-closed
> five-bucket grammar with resolved containment and permit-time re-binding,
> `curator-swift-module-v1` was minted, and the matrix and closure were stated
> explicitly.

Designs the closed local `swift-v1` and external `swift-repository-v1` pair from
the two independently accepted inputs — decision 0007 (`TASK-260728-1g0z69`,
toolchain requirement and preflight) and decision 0008 (`TASK-260728-2spy93`,
additional-language driver, version and artifact boundary) — plus the accepted
probe evidence of `TASK-260729-rhjxtx`, and selects an enforceable SwiftPM/swiftc
pipeline under the portable manager-worker policy by measuring it on a qualified
host.

Nothing here is a platform claim. Both identifiers remain reserved.

## Deliverables

| Artifact | What it is |
|---|---|
| `_decision-0011-swift-driver-pair.md` | The decision: the SwiftPM rejection and its measurement, the two-root closure, the manager-owned SDK presentation and plugin-closure verification, source ownership, the fixed process graph, the exhaustive rejection matrix, the `swift-tools-version` classifier and its security partition, identity, platform matrix, residual exposures, 29 rejected alternatives |
| `_swift-build-drivers-reference.md` | Implementation-ready reference: registry entry, probe vectors, normalization, target admission, fingerprint algorithm, SDK presentation and its nesting rule, source path grammar, `curator-swift-module-v1`, both argument vectors, the fail-closed `-###` grammar with its closed per-job token tables, five path buckets and permit-time re-binding, the plan-execution rule that deletes the macro/plugin channel, the operation-private environment, the rejection matrix with per-surface diagnostics and the reject/bound/inert partition, the ordered classifier, the closed recognised-outcome set, closure and control requirements, artifact rules, platform qualification, 219 conformance vectors |
| `_probe.tar.gz` | Standalone Go module `swiftboundaryprobe`: 19 non-test sources, 8 test files, 55 test functions, 7,841 lines |
| `_fixture-macos.json` | `swift-boundary-fixture-v1`, macOS 26.5 arm64 / Apple Swift 6.3.2: 23 cases, 23 matched, 0 divergences, 32 closure checks, 14 controls, 56 structural checks, the executed native-target admission holding, `green: true`, exit 0 |
| `_fixture-absent.json` | The same probe with no resolvable toolchain: 23 cases `not_run` with the reason, exit 0, nothing installed |
| `_command-evidence.log` | Every argv, environment, real exit code and bounded output excerpt the probe ran, rendered in order |
| `_gate-log.txt` | Gate transcript with real exit codes, the fourteen standalone expected-red controls, and attribution for the expected-red repository and spec gates |
| `_rework-cycle-1.md` | Cycle-1 rework record: the four reviewer findings, what closed each, the new measurements behind them, the two corrected counts, before/after tables |
| `_rework-cycle-2.md` | Cycle-2 rework record: the five reviewer blockers, what closed each, the measurement that changed the runtime-library rule the reviewer proposed, and the before/after probe and gate tables |
| `_rework-cycle-3.md` | Cycle-3 rework record: the two security-contract defects, the four negative and one decisive positive measurement behind the plan-execution design, the rejected detection alternatives, the closed token grammar, and the gate table |

Decision and reference are also written into the task worktree at
`.temp/TASK-260728-1yhuqi/curator-spec-worktree/{decisions/0011-swift-driver-pair.md,docs/swift-build-drivers.md}`.
Nothing was staged, committed, pinned or published.

## The question this task had to answer, and the answer

### Can the pre-compile rejection matrix be computed for a SwiftPM package?

No, and the measurement is unambiguous.

Decision 0008 section 7 requires every package-selected code-execution surface to
be rejected deterministically *before* the compile phase, and disqualifies a
driver whose surfaces cannot be. Rust's answer was yes: `cargo metadata` reports
`custom-build` and `proc-macro` targets **without running them**. Swift's answer
is no, because the thing that reports what a package declares *is* the package.

**Measured** on a `Package.swift` whose body writes to stderr and attempts a
filesystem write outside the package:

| Command | Manifest executed | Escape write landed |
|---|---|---|
| `swift package dump-package` | **yes** | no |
| `swift package --disable-sandbox dump-package` | **yes** | **yes** |

The manifest runs either way. The only thing that stopped the write was
SwiftPM's macOS `sandbox-exec` policy, which one documented flag removes — the
exact class of platform-specific, tool-supplied containment decision 0006 says
the portable policy does not provide.

There is no bounded alternative. **Measured**: `swift package tools-version`
returns exit 0 and `6.0.0` for a manifest whose body is `this is @@ not swift (((`,
while `dump-package` on the same file exits 1. The only datum SwiftPM will give
you without running the package is the comment on line 1.

**Decision: SwiftPM is not used at all**, in any stage, for any purpose. The
compiler is driven directly, the manager enumerates the source set itself, and
`Package.swift`'s first line is the bound project metadata — read as text, never
as a program.

That trade buys a graph phase stronger than Rust's rather than weaker.
**Measured**: `swiftc -###` over the driver's exact compile vector exits **0**
and prints its job plan both for a source containing `this is not swift @@@ (((`
and for a source path that does not exist, and writes nothing into the source
directory. Swift's graph phase reads no package byte as a program at all.

The cost is stated rather than softened: this driver builds single-module,
dependency-free Swift programs. A package with SwiftPM dependencies, plugins,
macro targets or binary targets is not buildable by `swift-v1`.

### Can the process graph stay inside a fingerprinted closure on macOS?

Yes, and more cleanly than Rust's.

**Measured**, `PATH` set to a directory of 32 logging shims:

| Run | PATH resolutions | Exit |
|---|---|---|
| the fixed compile vector, `-target` and `-sdk` pinned | **0** | 0 |
| the same vector plus `-use-ld=lld` (control) | **2** (`ld64.lld` twice) | 1 |

Swift differs from Rust in kind: `rustc` unpinned resolves `xcrun` and `cc`
through `PATH`, while `swiftc` unpinned resolves **nothing** and fails closed
with `unable to load standard library`. That makes the firing control
load-bearing — without it the zero is unearned, and the probe carries control
`C5` which reports exactly that failure mode.

The required closure is four executables, two of which are one file: `swiftc` is
a symlink to the single `swift-frontend` binary that dispatches on `argv[0]`; the
plan holds two `swift-frontend` jobs and one `clang` job, all absolute inside the
root; and `clang -print-prog-name=ld` reports `<root>/usr/bin/ld`. The SDK is a
data-only second root that starts nothing.

`swift` is **not** a member. It is the SwiftPM launcher, it is forbidden from
every stage, and the driver never invokes it, so requiring it to resolve would
add a portability constraint with no property behind it; its bytes are covered by
the root fingerprint regardless, and this contract uses it only as the upstream
oracle in the conformance probe. The rule is structural rather than a list —
every executable the verified plan names, plus the linker `clang` resolves, must
lie inside the root — which is what lets Linux add a fifth member
(`swift-autolink-extract`, measured) without contradicting anything.

Measured cost: 5,109 files / 2.57 GiB / 5.89 s for the toolchain root, 32,345
files / 0.71 GiB / 5.60 s for the SDK, per operation, stated rather than
memoised away.

## The finding that changed the design most

**The macro plugin surface is an executable outside the toolchain, and it is
reachable from ordinary package source.**

**Measured**: the driver injects plugin search paths into every frontend job
unasked, and the *external* ones are derived from the `-sdk` argument, three
ancestor levels up and then into `Developer/usr`. With the declared Xcode SDK path
passed directly, two distinct derived paths exist outside every fingerprinted root and
`#Predicate` loads `FoundationMacros` **through a `swift-plugin-server`
process** in that tree — and so does the toolchain's own `ObservationMacros`.

The manager therefore never passes the declared SDK path. It presents the SDK
through a directory it owns entirely:

```
<staging>/sdk/present/SDKs/MacOSX.sdk  ->  <declared platform-sdk root>
```

The nesting depth is fixed and measured: the derivation walks three ancestor
levels up and then into `Developer/usr`, so presenting here lands every derived
tree inside `<staging>/sdk`, which the manager keeps empty apart from the chain.

**Measured** with that presentation: 14 plugin components in the plan, **6
distinct** — 2 existing and both inside the toolchain root, 4 absent — and 0
existing outside a fingerprinted root; `#Predicate` fails
to compile with `external macro implementation type … could not be found`; and
`@Observable` still compiled with exit 0, loading
`<root>/usr/lib/swift/host/plugins/libObservationMacros.dylib` **in process**
with no server at all.

**That last sentence was the defect cycle 3 closed.** Presenting the SDK removes
an executable plugin *server* from the closure; it does not stop package source
selecting a macro implementation the compiler then runs, and decision 0008
section 7 forbids exactly that surface. Four measurements establish that nothing
in the argument vector, the environment or a presentation can suppress the load
while `swiftc` runs its own jobs: `-resource-dir` moves 0 of 10 plugin
components, a manager-owned toolchain symlink moves 0 of 10, an absent
in-process server path does not stop the load, and no compiler flag disables
plugin-path injection. The decisive positive measurement is that the channel is
entirely flag-driven — the same frontend job with **no** plugin flag rejects
`@Observable` with `plugin for module 'ObservationMacros' not found` and loads
nothing.

So the manager now **executes the plan it verified**, job by job, with the closed
plugin-channel token set deleted, and `compile_argv` is no longer executed. Five
assertions hold it: same job count and order; each executed argv reconstructed
independently from the verified argv and the deletion record; no surviving
channel token; no `-###` in an executed job; and every deleted flag a member of
the closed set. Measured: 10 deletions across 3 jobs, 0 survivors; Foundation,
`Codable`, a regex literal and string interpolation still build and the artifact
runs; `@Observable` fails closed with **0** load remarks; and two passes to the
same output path give one digest across two distinct intermediate directories.
Control `C13` restores the retired policy and reports the admission.

It also makes the load-bearing property literal. Before, the manager verified one
plan and let `swiftc` re-plan and execute a second one, so "the plan verified is
the plan executed" rested on the planner being deterministic. Now it is enforced.

The cost is stated rather than hidden: this driver compiles no Swift that uses
any macro, including the toolchain's own.

The presentation is not trusted on its own. The manager also **verifies** it
from the graph phase, with a fail-closed grammar that is total over **every
token** of every job line. Path values must be claimed by one of five buckets —
executable, plugin, search, source, output — and boundary-checked; containment is
computed on symlink-resolved paths; every verified path is bound and re-checked
immediately before the compile child starts; opaque values may not be
path-shaped, may not embed a separator, and must equal the constant the manager
chose; the two pass-through flags `-Xllvm` and `-Xcc` admit only values in a
per-platform measured allow-set; and a line, flag, value or operand no table
claims rejects the operation. Measured: 20 malformed plans and 16
unknown-channel vectors all reject, while the live 101-token plan verifies with
zero rejections. Cycle 2's totality stopped at path-shaped tokens and the gap was
live — this toolchain defines the joined `-load-pass-plugin=<lib>`, whose value
is a dynamic library the compiler loads, and the retired verifier admits 14 of
the 16. Controls `C6` and `C14` disable the two checks and report what they
admit.

## Other measured findings that shaped the contract

| # | Finding |
|---|---|
| 1 | `swiftc -print-target-info -target x86_64-unknown-linux-gnu` exits **0** and names `<root>/usr/lib/swift/linux`, which does not exist. Admission is a manager-side stat of `runtimeLibraryPaths`, not the print. |
| 2 | `target.unversionedTriple` (`arm64-apple-macosx`) is the identity form and is **not** a valid `-target` argument: it fails with `Swift requires a minimum deployment target of macOS 10.9.0`. The versioned `target.triple` is what the compiler gets. |
| 3 | Two compiles of the same sources to the **same** output path are byte-identical; changing only the output path changes the bytes, because the path reaches the Mach-O `LC_UUID`. Stated as a staging obligation, not as a reproducible-build claim. |
| 4 | `.swift-version` carrying `5.9.9-nonexistent` changed nothing — exit 0. That is what admits it as `compared` rather than `forbidden`. |
| 5 | The measured unsupported-tools-version floor is exactly `4.0.0`: `3.1` is refused by `swift build`, `4.0` is accepted. |
| 6 | Upstream silently reinterprets seven header forms — no space, uppercase keyword, extra whitespace, leading zeros, `-beta`, `+build`, and the specification below arbitrary manifest code. Each compares a version the author did not write. |
| 7 | A `TMPDIR` that does not exist aborts the driver with `couldNotFindTmpDir(…)` and produces nothing, so staging order is a hard requirement rather than hygiene. |

## One refinement of an accepted input

`TASK-260729-rhjxtx` finding 6 states that Swift rejects an unserved-but-known
target only **after** a frontend job starts. That reproduces here exactly — for a
**compile-only** vector: `-c -v` for `x86_64-unknown-linux-gnu` spawns 1 frontend
job and then fails on the standard library.

Under **this driver's** vector, which links, the same target fails at job
planning with `error: unableToFind(tool: "swift-autolink-extract")` and spawns
**0** frontend jobs. Both facts are true; the vector is what separates them. The
probe carries both runs in one structural check (`S12`) so the refinement cannot
be read as a contradiction, and the consequence is recorded: because the driver
always links, its `-###` graph phase refuses the target before any compiler
child, and the Stage A stat gate refuses it earlier still.

It also names the Linux consequence rather than burying it:
`swift-autolink-extract` is a **fourth** executable that a Linux qualification
must show resolving inside the fingerprinted root.

## The acceptance layers and the security partition

Curator reads `Package.swift`'s first line itself and never invokes SwiftPM to
do so, so the classifier is Curator-owned. Its alignment with upstream is
measured, using `swift package tools-version` as the isolated representability
oracle and `swift build` as the corroborating command.

Swift is better placed than Go on isolation, for the same structural reason Rust
is: **measured**, `swift package tools-version` reports `99.0.0` with exit **0**
on a 6.3.2 host while `swift build` exits 1 with the host-gate line, so the
isolated command structurally cannot be applying the host gate.

| Layer | Rejects | Host input |
|---|---|---|
| document | absent header, BOM before it | none |
| canonical form | the seven `F` members below | none |
| grammar | `6`, `6.0.0.0`, `notaversion`, empty | none |
| floor — **measured at 4.0.0** | `3.1`, `1.0` | none |
| host gate — **excluded from the layer measurement** | `99.0` on a 6.3.2 host | the resolved toolchain |

Unlike Rust's, this classifier has a **non-empty security partition** `F`, and
that is the honest outcome rather than an omission. Upstream accepts and
silently reinterprets `//swift-tools-version:6.0`, `// SWIFT-TOOLS-VERSION:6.0`,
`//   swift-tools-version:  6.0`, `06.0`, `6.0-beta`, `6.0+build`, and a
specification sitting below arbitrary manifest code. Curator refuses all seven
as `rejected-non-canonical-header` — never as a grammar rejection, which would
falsely assert upstream refuses them too. P1 (no widening) is asserted over all
cases; P2 (no narrowing) over cases outside `F`. **Measured**, both hold.

## Probe results

`swift-boundary-fixture-v1`, macOS 26.5 arm64, Apple Swift 6.3.2, real exit **0**:

```
cases: 23, matched 23, divergences 0, not run 0
alignment: P1 no-widening=true P2 no-narrowing=true (security partition empty=false)
closure: 32 checks, 0 yielded a verdict (must be 0)
controls: 14 of 14 failing as required
structural: 56 checks, 0 divergences
native target admission (P2 arm64-apple-macosx26.0): ok=true in-closure=1 base-installation=1 p1==p2=true
green: true
```

Recognition is whole-line exact against forms predicted **before** each command
runs, from the value under test plus constants the probe fixes from the resolved
toolchain — the full compiler version `6.3.2`, its major-minor form `6.3`, and
the package directory name upstream renders as an infix in every `swift build`
diagnostic. The isolated and corroborating sets cannot share one predictor for
exactly that reason. Two expected lines of different classes matching inside one
output is `unknown`, not first-wins.

Closure is measured, not asserted, in both laundering directions: 4 real
unrelated command failures, 20 value-bearing outcomes cross-fed over 338 pairs,
and 27 constructed cases. The constructed ones are disclosed as constructed. **20
outcomes are excluded from the cross-feed and the exclusion is printed with its
reason** — an exit-0 acceptance, a missing-specifier diagnostic and an
absent-header diagnostic name no value under test, so feeding them under another
value would test the classifier against text that was never about a value. The
exclusion is by measured property (the recognised line does not contain the
value), not by an allowlist of case names.

The fourteen controls are runnable from the same binary and each **must** fail:

| Control | Guards | Findings |
|---|---|---|
| C1 lead-only recognition | an outcome is a whole line, not a lead | 3 |
| C2 substring recognition | a diagnostic embedded in a longer line is not that diagnostic | 1 |
| C3 exit status as semantics | the floor, the host gate and a grammar rejection are not one exit status | 7 |
| C4 `swift build` as the isolated command | representability is measured by a command that cannot apply the host gate | 3 |
| C5 unearned `PATH`-closure zero | a zero without a firing control is not evidence | 1 |
| C6 plugin closure unchecked | the macro plugin surface is inside the fingerprinted closure | 2 |
| C7 lenient plan parsing | a plan shape the grammar does not cover is a rejection, not a skip | 16 |
| C8 lexical containment | containment is resolved, not a string prefix | 1 |
| C9 collapsing module derivation | two distinct command keys never share one module name | 3 |
| C10 prefix-only banner parser | the compiler banner is matched as a whole value, not by its prefix | 17 |
| C11 lenient runtime-library admission | the standard-library closure is total, not "whatever is already in the root" | 6 |
| C12 CR-normalizing plan splitter | LF is the only terminator and a bare CR is a rejection, never a stripped byte | 4 |
| C13 retired macro admission | a compiler macro selected by package source is a load the driver rejects, not a surface it admits | 1 |
| C14 path-shape-only totality | totality is over every token, not over the tokens already recognised as paths | 14 |

C5 through C14 are Swift-specific and each answers a place this contract could
have been quietly wrong: a meaningless zero, a trusted derivation rule, a parser
that skips what it cannot read, a containment test a symlink defeats, a name
mapping that merges two identities the protocol keeps apart, a version parser
that never reads the suffix it claims to match, a standard-library closure that
is a filter rather than a closure, a line splitter that quietly normalizes the
byte a Windows implementation would disagree about, a macro policy that called a
source-selected load an admitted surface, and a totality claim that stopped at
the tokens it already recognised.

Every control was also replayed **individually**; all fourteen exit 1.

A degraded run with no resolvable toolchain exits 0 with 23 cases `not_run` and
the reason recorded; nothing was installed, downloaded, updated or switched at
any point.

## Residual exposures, stated rather than closed

- **Macro expansion is no longer an exposure, because it cannot happen.** Cycle
  2 listed source-selected toolchain macro execution here as admitted. Section 4
  of the decision now deletes every plugin channel from the executed job set and
  asserts none survives. The entry is kept, inverted, so the change is recorded
  rather than quietly dropped.
- **Compile-time filesystem reads are bounded, not proven absent.** The admitted
  language has no `include_str!` analogue, and the surfaces checked and excluded
  are `-import-objc-header`, `-Xcc -include`, module maps, bridging headers and
  `.swiftinterface` inputs. With macro expansion removed, the read surface Swift
  contributes is the compiler front end itself. `STORY-260728-327soo` still
  receives the Rust compile-time read surface; it no longer receives a Swift
  macro-expansion one.
- **Foreign symbol declarations** (`@_silgen_name`, `@_cdecl`) can only name
  symbols the pinned link environment already resolves, which are
  base-installation libraries the artifact class already admits.

## Platform position

`platforms` holds exactly `(macos, arm64)`. That is the consequence of the
evidence, not a scoping choice.

- **macOS amd64** is a qualification obligation with a stated acceptance test.
- **Windows** gets a five-point implementation contract and **no** claim, and the
  contract now declines to assert a closure member count because the plan there is
  unmeasured, because
  `TASK-260729-rhjxtx` measured that no Swift toolchain exists on the reachable
  Windows host. The contract names the required shim set
  (`link.exe`, `lld-link.exe`, `cl.exe`, `clang.exe`, `ld.exe`,
  `swift-plugin-server.exe`, `where.exe`, `vswhere.exe`), requires the firing
  control, requires the plan verification to pass, and forbids resolving
  `link.exe`, `cl.exe`, `vswhere.exe` or a Visual Studio activation script from
  `PATH`, the registry or an environment variable.
- **Linux** stays excluded until `TASK-260728-1y8u4m`, with two named
  Linux-specific questions this host cannot answer: the fourth executable
  `swift-autolink-extract`, and the open-source compiler banner, which the
  normalization rule deliberately does **not** admit because no host in this task
  carried one.

## Decision number

This record takes `0011`, the lowest unclaimed number. Three in-flight records
claim lower ones and none is landed — `TASK-260728-12pnm1`
(`0009-rust-driver-pair.md`), `TASK-260728-1jafds`
(`0009-hardened-build-execution-profile.md`) and `TASK-260728-168smo`
(`0010-kotlin-native-driver-pair.md`), so `0009` is itself contested. If review
lands them in a different order this record renumbers rather than contests; the
renumber touches the filename, the title and two references in the reference
document.

## Gates

Real exit codes, each command run standalone; full transcript in `_gate-log.txt`.

| Gate | Exit |
|---|---|
| probe `gofmt -l .` | 0 |
| probe `go vet ./...` | 0 |
| probe `go build ./...` | 0 |
| probe `go test -count=1 ./...` | 0 (55 tests) |
| probe run, macOS, SDK presented | 0 (`green: true`) |
| probe run, toolchain absent | 0 (23 `not_run`) |
| curator `go build ./...` | 0 |
| curator `go vet ./...` | 0 |
| curator `go test ./...` | 0 |
| spec `tools/validate.py`, clean 57c1f56 | 0 (30 schemas, 93 vector files) |
| scoped link check over the two authored documents | 0 (6 local links, 0 broken) |
| curator `gofmt -l .` | **2** |
| curator `make check` | **2** |
| spec `tools/validate.py`, task worktree | **1** |

| probe tarball round-trip: extract, `gofmt`, `vet`, `test` | 0 |
| 14 expected-red controls, replayed individually | each **1**, as required |

`golangci-lint` was **not run**: the binary is not installed on this host
(`exit 127`, command not found).

The three non-zero gates are expected-red and attributed:

- `gofmt -l .` in curator lists **1141** files. **Zero** are from this task and
  **zero** are outside `.temp/`: every one lives under another task's scratch
  tree — an extracted Go distribution under `.temp/TASK-260728-2jaw7h/go1.25.1/`,
  `.temp/TASK-260729-1t1z2l/review-tmp/`, `.temp/TASK-260729-3jmqgl/worktree/`.
  This task's own module is clean (`gofmt -l .` exit 0 inside `probe-module`).
- `make check` fails at that same `gofmt` stage, for the same reason.
  `git status --porcelain --untracked-files=no` in curator is **empty**: no
  tracked project file was modified.
- `validate.py` in the task worktree passes its schema, manifest, review-evidence
  and vector-semantics checks and fails only its link check, with **3** broken
  links — two in `docs/portable-go-execution-policy.md` pointing at
  `conformance/v1/vectors/go-host-execution-policy.json` and one in
  `docs/external-build-repositories.md` pointing at `release/1.0.0-rc.5.json`.
  Both files were copied unmodified from another task's tree to close the link
  graph reachable from decision 0007; neither target exists at base commit
  `57c1f56`. The clean-tree baseline at the same commit exits **0**, and the
  scoped check over the two documents this task authored reports 6 local links
  and 0 broken. This is the same expected-red as `TASK-260728-12pnm1`.

## Scope kept

No normative curator-spec file was modified. `decisions/0011-swift-driver-pair.md`
and `docs/swift-build-drivers.md` are new files; decisions 0005 through 0008 and
three `docs/` documents were **copied** into the task worktree unmodified, with
byte identity verified against the accepted `TASK-260728-1g0z69` and
`TASK-260728-2spy93` worktrees before any edit. `git diff HEAD` in the task
worktree is empty — no tracked file was touched.

No schema, vector, release pin, dependency, generated corpus or release metadata
was altered. Nothing was staged, committed, pinned or published. The probe is a
standalone module under `.temp/` with its own `go.mod`. Nothing was installed,
downloaded, updated, activated or switched on any host.

## Open items for the reviewer

### New in cycle 3

- **The manager now executes the plan's jobs instead of the compile vector, and
  that is the one architectural judgement in this cycle.** It was not a
  preference: four measurements show the macro load cannot be suppressed while
  `swiftc` runs its own jobs, and decision 0008 forbids answering the surface
  with a runtime allowance. The reviewer should decide whether executing the
  verified jobs — with the deletion recorded, reconstructed and asserted — is the
  right closure, or whether the contract should instead reject the driver
  outright per decision 0008's disqualification clause, or whether decision 0008
  should be amended through its own reviewed change to permit fingerprinted
  toolchain macros. This contract does not reopen 0008.
- **Sequential, plan-order job execution is stated, and its portability is
  asserted rather than measured off macOS.** On this host the plan is
  topologically ordered — frontend jobs, then the link job — and sequential
  execution produced a working, byte-identical artifact. Windows and Linux carry
  the obligation to measure their own job sets. The reviewer should check whether
  anything else about `swiftc`'s own job scheduling is being assumed.
- **`compile_argv` is constructed and asserted over but never executed.** V1–V7
  keep their meaning as properties of the constructed vectors, and E1–E5 cover
  what actually runs. The reviewer should confirm that split reads correctly and
  that no property became vacuous.
- **The per-kind nullary sets and the `-Xllvm` / `-Xcc` allow-sets are
  per-platform measured tables.** On an unmeasured host they fail closed, which
  is intended, but it does mean macOS amd64 cannot qualify until its plan is
  measured. The reviewer should confirm that is the right cost for closing the
  unknown-token channel.
- **Macro-free is now a stated narrowing of the admitted package set.** No Swift
  using any macro — including `@Observable` — builds under this driver. The
  reviewer should confirm the narrowing is acceptable rather than a reason to
  revisit the driver's scope.

### New in cycle 2

- **The base-installation class is a deliberate widening, and it is the one
  judgement call in this cycle.** The reviewer proposed that every
  `runtimeLibraryPaths` entry must resolve inside a fingerprinted root. Measured,
  that rejects the host this contract is qualified on: macOS returns
  `/usr/lib/swift`, the OS-level Swift runtime. The reviewer's own text allowed
  "or explicitly define and justify a different closed set", and the closed set
  here is fingerprinted roots plus a registry-declared, per-platform, empty-by-
  default prefix list. Two things bound it: the class-B entry appears as **0
  tokens** in the verified job plan, so it is never a compiler input; and class-B
  membership enters identity under its own role token, so a runtime moving
  between the OS and the closure moves the cache key. The reviewer should decide
  whether that boundary is the right one, or whether the contract should instead
  reject the macOS host and require a toolchain that carries its own runtime.
- **`curator-swift-process-closure-v1` is in the receipt, not in the cache key.**
  The stated reason is that the plan-derived executable set cannot vary
  independently of inputs the key already binds. That is an argument, not a
  measurement, and it is the kind of argument worth attacking. If the reviewer
  can construct a case where the plan's executables change while both root
  digests, the compiler version and tag, the native target, the fixed vectors and
  the source set are all fixed, the object belongs in the key.
- **Windows is serializable but still unmeasured, and two gaps are named rather
  than filled**: the plan-derived member count, and the argument template for any
  SDK root beyond ordinal 0. The reviewer should confirm those are the only two
  left, and that nothing else in the identity algorithm silently assumes macOS.
- **The banner grammar is deliberately narrower than either cycle-1 form.** It
  rejects a parenthesised suffix, any non-ASCII byte and a leading zero. Each is
  a fail-closed choice with a stated widening path (measure the host that emits
  it). The reviewer should check that none of those narrowings would reject a
  plausible Apple release.
- **The `-###` insertion and the physical-line grammar are now stated verbatim
  in both documents and in the implementation.** The reviewer should confirm the
  three statements are word-for-word the same rule, since that identity is
  exactly what cycle 2 found missing.


Cycle-1 items that the cycle-1 verdict already resolved are dropped. What remains
open, plus what this cycle newly puts to the reviewer:

- Is `rejected-absent-header` the right disposition for a specification below
  line 1? The alternative — a declared whole-file scan — was rejected because
  upstream's scan was **measured** to reach a value inside a multi-line string
  literal, so a scan would make the manager's verdict a function of manifest body
  bytes. The cost is a less helpful diagnostic. The reviewer should decide
  whether that trade is right.
- ~~Is the plan verifier's totality claim scoped correctly?~~ **Closed in cycle
  3.** Totality is now over every token, under a closed per-job flag and operand
  grammar; the stated limit is gone and control `C14` reports what the old
  boundary admitted.
- Is the permit-time re-binding the right answer to the mutation race, given that
  the contract says plainly it narrows rather than closes the window and that the
  ownership requirement on the declaration channels is what actually closes it?
  The reviewer should decide whether both layers are warranted or whether the
  ownership requirement alone should carry it.
- Is `curator-swift-module-v1` the right shape? It is readable on the short
  branch and opaque on the overflow branch. An alternative — hash every key, so
  the mapping is uniform — was not taken because it makes every diagnostic
  unreadable. The reviewer should check the escape table and the 61-byte
  threshold.
- Is rejecting native inputs **anywhere** in the build-root subtree right, given
  that outside `Sources` they would be inert? The argument is the
  `Package.resolved` one — their presence declares an intent the driver cannot
  honour — but it is a narrowing of what packages are admitted.
- Is the ASCII-control-byte rule on source paths acceptable as the only
  source-name restriction? It is measured (a newline splits the plan line) and
  deliberately does not restrict quotes, spaces, `#` or a leading `@`, each of
  which was measured to round-trip through the plan grammar.
- Two cycle-1 counts were corrected downward here (plugin components existing
  outside the roots: three → two distinct; presented-SDK plugin paths: 6 inside /
  8 absent → 6 distinct, 2 existing inside / 4 absent). The security properties
  are unchanged. The reviewer should confirm the corrections rather than the
  originals.
