# TASK-260728-1yhuqi — rework cycle 3

Closes the two security-contract defects of
`TASK-260728-1yhuqi_review-verdict-cycle-3.md`. The independently supported
SwiftPM rejection and direct-`swiftc` architecture, every cycle-1 closure and
every verified cycle-2 repair are preserved.

Host: macOS 26.5 arm64, Apple Swift 6.3.2 (`swiftlang-6.3.2.1.108`), Go 1.25.5.
Probe: `swift-boundary-fixture-v1`, 23 cases / 23 matched / 0 divergences, 32
closure checks with 0 verdicts, **14 of 14** expected-red controls failing as
required and each replayed individually exiting 1, **56** structural checks with
0 divergences, executed P2 admission holding, `green: true`, exit 0. Degraded
run: 23 `not_run`, exit 0, nothing installed.

---

## Finding 1 — the contract admitted the compiler macros decision 0008 rejects

### What the reviewer required, and what had to be measured first

The verdict's recommendation was to "reject actual macro/plugin loads" because
"the plan already exposes the load flags". That is half true and the other half
decides the design, so it was measured before anything was written.

The plan does expose the *channel* — `-plugin-path`, `-external-plugin-path`,
`-in-process-plugin-server-path`. It does **not** expose the *load*: the driver
plans jobs without reading source, so the plan for `@Observable` source and for
macro-free source is the same plan. A load is selected inside the frontend, by
package source, after the permit. So "detect the load in the plan and reject"
has nothing to detect.

That leaves two candidate closures: suppress the channel, or detect the use.

### Four negative measurements: the channel cannot be suppressed by argument

| Attempt | Measured | Fixture |
|---|---|---|
| `-resource-dir <manager-owned>` | **0 of 10** plugin components move; they derive from the driver's own executable location, not the resource directory | `S42` |
| present the toolchain the way section 2.2 presents the SDK — invoke `swiftc` through a manager-owned symlink | **0 of 10** move; the driver resolves its own executable before deriving anything | `S43` |
| point `-in-process-plugin-server-path` at an absent file | the compile still exits **0** and the macro **still loads** — the implementation is `dlopen`ed directly, not fetched through the named server | `S44` |
| look for a disabling flag | neither `swiftc` nor `swift-frontend` defines one; there is no `-no-plugins`, `-disable-plugin-search` or equivalent in either option list | option-list read |

So while `swiftc` runs its own jobs, a source-selected toolchain macro load
cannot be prevented by argv, environment or presentation. `S45` keeps the
retired behaviour on the record: the `swiftc`-driven compile of `@Observable`
exits 0 and loads
`<root>/usr/lib/swift/host/plugins/libObservationMacros.dylib`.

### The decisive positive measurement

The channel is entirely flag-driven. The same frontend job with **no** plugin
flag rejects `@Observable` with

```
error: external macro implementation type 'ObservationMacros.ObservableMacro'
could not be found for macro 'Observable()'; plugin for module
'ObservationMacros' not found
```

and loads nothing; with `-plugin-path` present it loads the implementation.

### Why detection was rejected instead

`swift-frontend -scan-dependencies` was measured as a candidate oracle. It
reports `macroDependencies` for the dependency **closure**, not for use:
measured, a source file with no macro at all still reports `SwiftMacros` from
the standard library while its compile emits **zero** load remarks. A
non-empty-means-reject rule would therefore reject every build; a narrower
reading would admit a package that genuinely uses a stdlib macro. A parse-only
pass has the opposite problem — telling a macro attribute from an ordinary one
needs name resolution. Both are recorded as rejected alternatives in the
decision.

### The closure: `curator-swift-plan-execution-v1`

The manager executes **the plan it verified**, job by job, with the closed
plugin-channel token set deleted, and does not execute `compile_argv`.

```text
channel-flags  := -plugin-path / -external-plugin-path /
                  -in-process-plugin-server-path / -load-plugin-library /
                  -load-plugin-executable / -load-resolved-plugin /
                  -cas-plugin-path / -cas-plugin-option
channel-joined := -load-pass-plugin=

executed[j] := verified[j] with every channel-flag token and its value removed,
               and every channel-joined token removed
```

Jobs run in plan order, sequentially, stopping at the first non-zero exit. The
manager creates each job's `-o` parent, because the plan names a temporary
directory `swiftc` would otherwise have created; that directory is already
required to be operation-private by the output bucket.

Five assertions, mechanical rather than argued:

| # | Property | Fixture |
|---|---|---|
| E1 | same job count, same order | `S46` |
| E2 | each executed argv equals its verified argv with exactly the **recorded** deletions removed, reconstructed independently from the deletion record | `S46` |
| E3 | no executed token opens a plugin channel, in either spelling | `S46` |
| E4 | no executed job argv contains `-###` | `S46` |
| E5 | every deleted flag is a member of the closed channel set | `S50` |

### Measured behaviour of the replacement

| Quantity | Measured | Fixture |
|---|---|---|
| deletions in the default two-source plan | **10** across 3 jobs — 4 `-external-plugin-path`, 2 `-in-process-plugin-server-path`, 4 `-plugin-path` | `S46` |
| surviving channel tokens | **0**, E1–E4 all holding | `S46` |
| ordinary Swift | Foundation, `Codable`, a regex literal, string interpolation and `Sendable` build; the artifact runs and prints its expected output | `S47` |
| `@Observable` | first frontend job exits **1** with the missing-implementation diagnostic, **0** load remarks | `S48` |
| determinism | **1** distinct artifact digest across 2 passes that used **2** distinct intermediate directories | `S49` |
| deleted flags | all three deleted flag names are members of the closed set | `S50` |

Control `C13` restores the retired policy and reports the admission (1 finding).
`S48` is the positive assertion the reviewer required alongside it.

### What this strengthens, and what it costs

**Strengthens.** Under the retired policy the manager verified one plan and then
let `swiftc` re-plan and execute a second one; "the plan verified is the plan
executed" rested on the planner being deterministic. It is now enforced. The
`invoked` projection of `curator-swift-process-closure-v1` now records paths the
manager itself passed to `exec`, and the 4/3 invoked/resolved counts are
unchanged.

**Costs, stated rather than hidden.** This driver compiles no Swift that uses
any macro, including the toolchain's own. Sections 2 and 4 of the decision state
it, and `TASK-260728-1egim2` carries it as the second thing an author meets.

### Points reopened for the reviewer to check

Decision 0008 is **not** reopened. Its macro prohibition is now conformed to
rather than argued with, and the alternative of amending 0008 to permit
fingerprinted toolchain macros is recorded as a rejected alternative that
belongs to that decision's own reviewed change.

---

## Finding 2 — unknown flag channels were accepted

### The gap was live, not theoretical

Cycle 2's totality was over tokens the verifier already recognised as
path-shaped. Three families passed with no verdict at all:

- an unknown flag beginning with `-` fell off the end of the dispatch chain;
- a joined `-flag=value` carrier was never split, so its value never became
  path-shaped. **Measured** (`S53`): this toolchain defines
  `-load-pass-plugin=<path>`, whose value is a dynamic library the compiler
  loads;
- `-Xllvm` and `-Xcc` hand their value to a **second option parser**, so
  `-Xllvm -load-pass-plugin=<lib>` and `-Xcc -isystem<dir>` carry a path through
  a token that is not itself path-shaped.

### The closure: totality over every token

A token is admitted only by being named in a table. Job kinds are read off the
plan — a job is a **frontend** job exactly when its first argument is
`-frontend`, a **link** job otherwise — and each kind has its own admitted set.

- **common valued flags** (either kind), each landing in a boundary-checked
  bucket: `-sdk`, `-isysroot`, `--sysroot`, `-resource-dir`, `-I`, `-L`, `-F`
  (search); `-o` (output); `-new-driver-path` (executable); `-primary-file`
  (source); the eight plugin flags (plugin);
- **per-kind nullary flags**, measured: ten for the frontend kind, `-O3` for the
  link kind;
- **opaque-value flags** with the rule that closes the value carrier — an opaque
  value MUST NOT be path-shaped, MUST NOT contain `/`, `\` or `#`, MUST match
  its class charset, and where the manager chose the value MUST equal it byte
  for byte: `-target` / `--target=`, `-swift-version`, `-module-name`,
  `-target-sdk-version`, `-target-sdk-name`, and the two pass-throughs
  `-Xllvm` / `-Xcc`, whose values must be members of the platform's **measured**
  allow-set because their value is parsed by another parser;
- **joined spellings** `--target=`, `-I`, `-L`, `-F`, longest prefix wins, empty
  value rejects;
- **positional operands**: a source-set member in a frontend job, an
  operation-private path in a link job;
- an `@`-leading token is a **response file** and rejects ahead of every other
  rule, so no rule can be reached by argument expansion; an empty token rejects.

The five path buckets, resolved containment, per-path binding and permit-time
re-verification are all unchanged.

### Measured

| Quantity | Measured | Fixture |
|---|---|---|
| live plan under the closed grammar | 3 jobs, **101 tokens**, every one claimed, **0 rejections** | `S51` |
| unknown-channel negatives | **16 of 16 rejected**, 0 still admitted | `S52` |
| retired path-shape-only verifier on the same 16 | admits **14** | `S52`, `C14` |
| `-load-pass-plugin=` is a live flag | present in the driver's own option list | `S53` |

The retired verifier catches only two of the sixteen, and both by accident:
`-Xcc -I<dir>` matches its joined-search scan and `-module-name <abs path>`
matches its path-shape scan. Every other channel — including the joined LLVM
pass plugin and the `-isystem` pass-through — passed it silently.

Extension is a measured contract change: a future Swift patch emitting a new
token fails closed with `build_execution_control_unavailable`, and the token
enters a table only after the emitted plan is measured. Windows and Linux
qualification now carry the obligation to measure their own per-kind nullary set
and pass-through allow-sets before qualifying; an unmeasured token fails closed
until it is added.

---

## Document changes

`decisions/0011-swift-driver-pair.md`

- section 2: the admitted set is now single-module, dependency-free **and
  macro-free**, with the consequence for source that merely *uses* a macro;
- section 4: rewritten conclusion — the four negative measurements, the decisive
  positive one, the deletion construction, the five assertions, the measured
  results, the stated cost; the verification paragraph now states totality over
  every token and cites the `-load-pass-plugin=` measurement;
- section 8: `graph_argv` is executed and `compile_argv` is not; V1–V7 keep their
  meaning as properties of the constructed vectors;
- section 9 matrix: the `admit` row for toolchain macros is replaced by a
  reject-the-load row, plus rows for every other plugin spelling, for unknown
  plan tokens, and for opaque value carriers;
- section 12 policy object: `"macros": false`, `"plugin_closure_check":
  "job-plan-verified-v2"`, and two new members `"plugin_channel":
  "plan-deleted-v1"` and `"job_execution": "manager-executed-plan-v1"`;
- section 13: Windows and Linux obligations extended with the per-platform flag
  tables and the plan-execution requirement, with an explicit prohibition on
  falling back to the retired policy;
- section 14: the "toolchain macros run inside the compiler" exposure is
  **inverted and recorded as closed**, not silently dropped;
- section 15: `STORY-260728-327soo` no longer receives a Swift macro-expansion
  read surface; `TASK-260728-1egim2` gains the macro-free authoring consequence;
- rejected alternatives: three added — keep admitting fingerprinted toolchain
  macros; detect the load by dependency scan or parse-only pass; keep
  path-shape-only totality — each with its measurement.

`docs/swift-build-drivers.md`

- 2.3: who starts what under plan execution, with the 4/3 counts unchanged;
- 4: `graph_argv` executed, `compile_argv` not;
- 4.1.2: rewritten as the closed per-job token grammar;
- 4.1.5: the 101-token positive result and both negative families;
- **4.2 new**: plan execution and the deleted macro/plugin channel — the four
  negative measurements, the construction, E1–E5, the measured results;
- 6.3: new reject rows; 6.5: no macro is admitted;
- 7.6: `C13` and `C14`; 9: the deletion failure class;
- 11: the macro exposure inverted; 12.1: a ninth acceptance-test item and the
  per-platform flag-table obligation; 12.2: Windows measurement obligations;
- 13: `PV01`–`PV36` and the new `XV01`–`XV06` execution group.

## Probe changes

New: `planclosed.go` (closed grammar), `planexec.go` (deletion, E1–E5,
execution), `structural_cycle3.go` (`S42`–`S53`), `planclosed_test.go`,
`planexec_test.go`. `plan.go`'s verifier is retained verbatim as
`VerifyPlanPathShapeOnly` for control `C14`, with its per-bucket rule extracted
into one `checkBucket` both verifiers call so they cannot disagree. `controls.go`
gains `C13` and `C14`.

19 non-test Go files, 8 test files, 55 test functions, 7841 lines; the tarball
extracts and `gofmt`/`vet`/`test` clean.

## Gates

| Gate | Result |
|---|---|
| probe `gofmt -l` / `go vet` / `go test` / `go build` | all exit 0 |
| native probe | green, exit 0 |
| degraded probe | 23 `not_run`, exit 0 |
| 14 expected-red controls, replayed individually | each exit 1 |
| probe tarball round-trip | extracts, vet and test clean, exit 0 |
| curator `gofmt -l .` | expected-red: 1141 paths, **0** under this task, **0** outside `.temp/`, **0** modified tracked files |
| spec `validate.py`, task worktree | expected-red: link check only, 3 links in 2 documents this task did not author |
| spec `validate.py`, clean baseline at `57c1f56` | exit 0, 30 schemas, 93 vector files |
| scoped link check over the two authored documents | 6 links, 0 broken |

Nothing staged, committed, pinned, published or installed on any host. No
platform widening: `(macos, arm64)` remains the only measured pair.
