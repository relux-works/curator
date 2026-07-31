# TASK-260728-1yhuqi — rework cycle 1

Answers the four acceptance gaps in
`TASK-260728-1yhuqi_review-verdict-cycle-1.md`. The independently supported
architecture is preserved unchanged: SwiftPM is still rejected in its entirety on
measured manifest execution, and the driver still invokes `swiftc` directly over
a manager-enumerated source set.

Everything below was measured on the same host as cycle 1 — macOS 26.5 arm64,
Apple Swift 6.3.2 (`swiftlang-6.3.2.1.108`), `XcodeDefault.xctoolchain` and
`MacOSX.sdk` from Xcode 26.5. No platform claim changed.

---

## Finding 1 — the first-line rule contradicted the classifier

**Resolved by choosing line 1 and removing the scan.**

`CuratorClassify` no longer calls `scanForHeader`. Its entire input is the bytes
up to the first LF with one trailing CR removed, and no byte after that LF can
change a verdict. The class definitions were rewritten to say what was actually
examined:

| Class | Old rule | New rule |
|---|---|---|
| `rejected-absent-header` | line 1 carries none **and no other line does** | **line 1** carries none; later lines are not consulted |
| `rejected-non-canonical-header` | a specification exists, not as the canonical line 1 | **line 1** carries one, not canonically |

A specification below line 1 therefore classifies as `rejected-absent-header`.
It stays in the security partition `F`, which now has two shapes — six line-1
reinterpretations, and two below-line-1 forms — for eight members total. No
member of `F` may classify as `rejected-grammar`, and a test enforces that.

**Why line 1 rather than a declared whole-file scan.** New measurement, S10b:

```
$ swift package tools-version      # manifest whose only specification is
                                   # inside a multi-line string literal
9.9.0
exit=0
```

Upstream's scan reaches bytes inside a string constant. A whole-file scan would
have let arbitrary manifest body bytes set the version the manager compares,
which is precisely the input the driver exists to exclude. The better diagnostic
was not worth making the boundary depend on the bytes it exists to refuse.

Two agreement cases were added so the narrowing is bounded rather than
open-ended, and both were measured:

- a canonical line 1 followed by `// swift-tools-version:99.0` on line 2 yields
  `6.0.0` from upstream — line 1 decides for both, so this is **not** in `F`;
- `let s = "// swift-tools-version:9.9"` is found by neither, because the line
  does not begin with the comment marker — also **not** in `F`.

`scanForHeader` survives, renamed `upstreamScanForHeader` and documented as an
upstream *model* used only to predict what the oracle will print. `CuratorClassify`
never calls it.

Touched: decision §6, §10, rejected alternatives; reference §7.1, §7.2, §7.3;
`classifier.go`, `run.go`, `cases.go`, `classifier_test.go`.

---

## Finding 2 — the `-###` verifier was not fail-closed

**Resolved by replacing the scan with a closed grammar, resolved containment,
and permit-time re-binding.** New file `plan.go`, new reference §4.1.1–4.1.5.

### The grammar, measured rather than assumed

New measurements fixed the quoting rule before the parser was written:

| Source name | Rendered in the plan |
|---|---|
| `has space.swift` | single-quoted |
| `has'quote.swift` | single-quoted, inner quote as `'\''` |
| `has#hash.swift` | single-quoted |
| `back\slash.swift` | single-quoted, backslash literal |
| `@at.swift` | bare absolute path |
| `new\nline.swift` | **splits the job across physical lines** |

So the grammar is POSIX single-quote quoting, and the last row is a real defect
in a line-oriented parser. `TokenizeJobLine` implements exactly that and returns
an **error** — never a shorter token list — for an unterminated quoted run, a
dangling backslash, or any ASCII control byte.

### Totality over the path surface

The verifier no longer checks "executables and plugin paths". Every path-shaped
token must be claimed by exactly one of five buckets, and an unclaimed one
rejects:

| Bucket | Rule |
|---|---|
| executable | resolves inside the `swift-toolchain` root, regular, executable |
| plugin | resolves inside a fingerprinted root, **or** does not exist |
| search | exists and resolves inside a fingerprinted root |
| source | byte-equal to a member of the manager's own ordered source set |
| output | resolves, or has a parent that resolves, inside operation-private state |

`-new-driver-path` moved into the executable bucket after the measurement showed
the frontend job names `swift-driver` by absolute path; `-sdk`, `--sysroot`,
`-resource-dir`, `-I`, `-F`, `-L` and their joined spellings moved into the
search bucket. Flag grammar is explicit: a flag with no following token or an
empty value rejects; `-external-plugin-path` must be exactly `<dir>#<server>`;
a `#` in any other plugin flag rejects. A non-flag token naming an existing
entry relative to the working directory rejects, closing the relative-path
channel.

Measured on the emitted plan: 3 jobs, buckets `executable=5 plugin=14 search=5
source=4 output=5`, **0 rejections**.

### Containment and TOCTOU

`insideAny` (lexical prefix) was replaced by `contained`, which symlink-resolves
both sides and compares byte-exactly on components. The root itself is inside
itself, because `--sysroot` names it directly. A case variant on a
case-insensitive volume fails closed. A dangling symlink — an entry that exists
and resolves nowhere — rejects.

`VerifyPlan` now returns bindings carrying each resolved path and its file
identity, plus each plugin path verified absent. `Reverify` runs immediately
before the compile child and rejects on a changed resolution, a changed identity,
or an absent plugin path that has appeared. The contract states plainly that this
narrows the window and that what closes it is the ownership requirement the
declaration channels already impose.

### Negative controls, which cycle 1 had none of

Twenty runnable negative vectors, one per failure family: relative executable;
unknown wrapper line; executable outside the root; executable that does not
exist; unmatched quote; dangling backslash; plugin flag with no value; with an
empty value; `-external-plugin-path` with one component; with three; `#` in a
single-path plugin flag; plugin path existing outside every root; search path
outside every root; joined search path outside every root; search path that does
not exist; source not in the manager's set; output outside operation-private
state; unclaimed absolute positional; blank line; empty plan; ASCII control byte;
`-new-driver-path` outside the root.

**20 of 20 reject**, while the plan the toolchain actually emits verifies clean —
a grammar that refused everything would prove nothing.

Three of these are exercised against real filesystem state rather than synthetic
text: S19 builds a symlink escape, S20 creates a plugin path that was verified
absent, S21 rewrites a bound executable between graph and permit.

### Vector equality

S17 asserts token-for-token that the graph vector is the compile vector with
`-###` inserted at index 0, from a single builder. The decision now also states
that the planned **job** argv is *not* reproducible — it carries a per-run
`TemporaryDirectory.XXXXXX` under the operation-private `TMPDIR` — so the
verification is a bucket-and-boundary check, never a comparison against a fixed
expected plan.

### New expected-red controls

- `C7` restores the lenient scan: it admits **16 of the 20** malformed plans.
- `C8` restores lexical containment: it admits the symlink escape.

Touched: decision §4, §8, §9, security impact, rejected alternatives; reference
§4, §4.1; new `plan.go`, new `plan_test.go`, `structural.go`, `controls.go`.

---

## Finding 3 — module-name derivation was not total

**Resolved by `curator-swift-module-v1`.** New file `modulename.go`, new
reference §3.2.

Escape into `[A-Za-z0-9]` with a prefix-free code (`z`→`zz`, `.`→`zd`, `-`→`zh`,
`_`→`zu`), then branch on length: `Sk_` + escaped when the escape fits in 61
bytes, otherwise `Tk_` + 40 hex digits of a domain-separated SHA-256 of the
**whole** key + `_` + a 20-byte readable prefix. Both branches land on ≤ 64
bytes; the long branch is exactly 64.

| Property | How it is established |
|---|---|
| total | every protocol-valid key has a result, over a 229-key corpus covering punctuation, leading digits, both sides of the length boundary and the overflow branch |
| deterministic | no host, clock or filesystem input |
| injective (short) | `DecodeModuleName` inverts it; the test asserts round-trip for every short result |
| collision-resistant (long) | 160-bit digest of the whole key, not of a truncation |
| branch-separated | `Sk_` and `Tk_` are disjoint prefixes |
| cannot shadow | the prefix forces an uppercase first letter, so no keyword and not `Swift` — which a bare escape would produce for the key `wift`. Measured: 0 of 341 inventoried toolchain and SDK module names carry either prefix |

The reviewer's collision is closed and tested: `my-tool`→`Sk_myzhtool`,
`my.tool`→`Sk_myzdtool`, `my_tool`→`Sk_myzutool`. A key outside the protocol
grammar is rejected with `build_source_layout_invalid`, never coerced.

**Identity no longer depends on injectivity at all.** The canonical build input
now binds `command_key` alongside `module_name`, so two commands cannot share a
cache identity even under a hypothetical collision. Receipt negatives grew from
10 to 12 to cover a `module_name` that is not the derivation of its
`command_key`, and an absent `command_key`.

`C9` restores the replacement rule and reports three collisions.

Touched: decision §6, §12, §16, rejected alternatives; reference §3, §3.2, §8,
§13; new `modulename.go`, new `modulename_test.go`.

---

## Finding 4 — matrix and Windows closure gaps

### Matrix

Three verdicts are now named and kept apart — **reject**, **bound**, **inert** —
and the build-root subtree is partitioned totally between them. New explicit
rows:

| Surface | Verdict | Basis |
|---|---|---|
| response file (`@`-leading argument) | reject as unreachable | **measured** that `swiftc` honours `@file`, and **measured** that a source named `@resp.swift` reaches the compiler as an absolute path and compiles |
| `unsafeFlags`, `swiftSettings`, `linkerSettings`, `cSettings`, prebuild/postbuild commands, manifest `#if` | reject as an input | the read stops at the first LF; no flag member exists |
| build configuration selector — debug/release, `-Onone`, `.xcconfig`, `.xcodeproj`, a scheme | reject as unreachable | the vector fixes `-O`; none of those files is compiler-visible |
| scripts — `.sh`, `Makefile`, hooks, executable-bit files | inert outside `Sources`, reject inside | fixed process graph, empty `PATH`, and §4.1 rejects any executable it does not account for |
| plugin / macro / binary / system-library **declarations** | reject as an input | named individually rather than folded into "the body" |
| non-compiler-visible files outside `Sources` | inert | new §6.6 states what inert means: in the audit subject and source identity, never opened, never compiled, never executed |
| control byte in a source relative path | reject | **measured** that it breaks the plan grammar |
| the §4.1 plan and permit rows | reject | ten rows, replacing three |

Native inputs are now rejected anywhere in the build-root subtree rather than
only in "the compiler-visible tree", with the reason stated: outside `Sources`
they would be inert, and they are rejected anyway because their presence
declares an intent the driver cannot honour.

### Windows closure

The four-versus-five inconsistency is resolved in favour of **four**, and the
rule is made structural rather than a list:

> Every executable the verified job plan names, plus the linker
> `clang -print-prog-name` resolves, MUST resolve inside the `swift-toolchain`
> root.

macOS instantiates that to four: `swiftc`, `swift-frontend`, `clang`, `ld`.
`swift` is **probe-only** — the SwiftPM launcher, forbidden from the pipeline,
never invoked by the driver, still covered by the root fingerprint, used in this
contract solely as the upstream oracle. Reference §2.1, §12.1 step 3 and §12.2
now agree, and S2b checks four members while separately recording that `swift`
exists and is never invoked.

Windows names an *expected* shape and explicitly declines to assert a count,
because the plan there is unmeasured and Linux was measured to add a fifth
member (`swift-autolink-extract`). An implementation must read the member set off
the plan it verifies. No Windows qualification is claimed.

### SDK presentation

While closing the derivation question the presentation moved to a fixed nesting
depth, because the derivation was measured to walk **three** ancestor levels up
and then into `Developer/usr`:

```
<staging>/sdk/present/SDKs/<name> -> <declared platform-sdk root>
```

so every derived tree lands inside `<staging>/sdk`, which the manager keeps empty
apart from the chain. Five explicit guarantees replace the previous one.
S26 measures it: 8 SDK-derived components, all inside the presentation base, none
existing.

Touched: decision §3, §4, §9, §13; reference §2.1, §2.2, §6.1–6.6, §12.1, §12.2,
§12.3.

---

## Corrections to cycle-1 measured numbers

Two counts in the cycle-1 documents were re-measured and corrected. Both were
overstatements of component counts, not of the property:

| Claim | Cycle 1 | Corrected |
|---|---|---|
| plugin paths existing outside every root under the declared SDK path | "three of those derived paths exist" | **2 distinct** — the plugin directory and the plugin server. The third derived path (`usr/local/...`) does not exist |
| plugin paths under the presentation | "14 plugin paths, 6 inside the toolchain root, 8 absent" | **14 components, 6 distinct**: 2 existing and both inside the toolchain root, 4 absent. The earlier figures counted per-job repeats |

The security property is unchanged in both cases: 0 existing outside a
fingerprinted root under the presentation, and a live external plugin tree under
the declared path.

Also corrected: the closure sentence "four executables, three of them one file"
became "four required executables, two of them one file", because `swift` left
the required set.

---

## Probe

`swift-boundary-fixture-v1`, same fixture version, extended.

| | Cycle 1 | Rework |
|---|---|---|
| cases | 20 | **23** |
| matched | 20 | **23** |
| divergences | 0 | **0** |
| closure checks | 32, 0 verdicts | **32, 0 verdicts** |
| expected-red controls | 6 of 6 | **9 of 9** |
| structural checks | 18 | **30** |
| structural divergences | 0 | **0** |
| green | true | **true** |
| module | 9 sources, 1 test file, 2579 lines | **11 sources, 3 test files, 4368 lines** |

New structural checks: S10b (upstream scans into string literals), S17 (vector
equality), S18 (20 malformed plans rejected), S19 (symlink escape), S20 (absent
plugin path that appears), S21 (bound executable that changes), S22 (the real
plan verifies under the strict grammar), S23 (module derivation total,
collision-free, decodable), S24 (reserved prefix in the measured closure), S25
(response files live and unreachable), S26 (SDK-derived tree stays in
operation-private state), S27 (control byte breaks the plan grammar).

New cases: `later-specification-ignored`,
`specification-inside-a-single-line-string`,
`F-inside-a-multi-line-string-literal`. `F-not-first-line` was renamed
`F-below-line-one` and its Curator expectation moved to
`rejected-absent-header`.

Degraded run with no resolvable toolchain: 23 cases `not_run`, exit 0, nothing
installed or activated.

---

## Gates

| Gate | Exit | Note |
|---|---|---|
| probe `gofmt -l .` | 0 | |
| probe `go vet ./...` | 0 | |
| probe `go build ./...` | 0 | |
| probe `go test -count=1 ./...` | 0 | |
| tarball extract + `go test -count=1 ./...` | 0 | the attached archive is self-contained |
| probe run, macOS | 0 | green |
| probe run, toolchain absent | 0 | 23 `not_run` |
| expected-red `C1`–`C9` | 1 each | all nine fire as required |
| curator `go build ./...` | 0 | |
| curator `go vet ./...` | 0 | |
| curator `go test ./...` | 0 | |
| spec `validate.py`, clean tree at `57c1f56` | 0 | 30 schemas, 93 vector files |
| spec `validate.py`, task worktree | **1** | expected red, attributed below |
| curator `gofmt -l .` | **2** | expected red, attributed below |
| curator `make check` | **2** | same cause as `gofmt` |

**`validate.py` attribution.** The failure is its link check on three links
inside `docs/portable-go-execution-policy.md` and
`docs/external-build-repositories.md`, copied unmodified from another task tree
to close the link graph reachable from decision 0007. Their targets
(`release/1.0.0-rc.5.json`, `conformance/v1/vectors/go-host-execution-policy.json`)
do not exist at base `57c1f56`. This task authored neither document. The clean
baseline at the same commit exits 0, and a scoped link check over the two
documents this task did author reports **6 local links, 0 broken**.

**`gofmt` / `make check` attribution.** 754 listed paths, **0 outside `.temp/`**,
**0 under `.temp/TASK-260728-1yhuqi`**. Every one belongs to another task's
scratch tree: `TASK-260720-1zntv0`, `TASK-260728-2jaw7h` (a vendored go1.25.1
source drop), `TASK-260729-1t1z2l`, `TASK-260729-3jmqgl`. Modified tracked files
in the curator repo: **0**.

---

## What did not change

- SwiftPM is still rejected in its entirety, on the same measurement.
- The driver still drives `swiftc` directly over a total enumeration of
  `Sources/**/*.swift`.
- The two argument vectors are unchanged apart from the module-name value.
- `platforms` is still `[(macos, arm64)]`. No Windows or Linux claim is made.
- Nothing was staged, committed, pinned or published; nothing was installed,
  downloaded or activated on any host.
- Decision number `0011` is unchanged; `0009` remains contested and `0010`
  claimed, so the renumber note stands.
