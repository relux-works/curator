# TASK-260728-12pnm1 — Rust driver-pair security contract

Developer handoff. Ready for review.

Designs the closed local `rust-v1` and external `rust-repository-v1` pair from
the two independently accepted inputs — decision 0007 (`TASK-260728-1g0z69`,
toolchain requirement and preflight) and decision 0008 (`TASK-260728-2spy93`,
additional-language driver, version and artifact boundary) — and proves on a
qualified host which Cargo and rustc behaviours can be admitted under the
portable manager-worker policy.

Nothing here is a platform claim. Both identifiers remain reserved.

## Deliverables

| Artifact | What it is |
|---|---|
| `_decision-0009-rust-driver-pair.md` | The decision: closure, source ownership, process graph, rejection matrix, acceptance layers, identity, platform matrix, residual exposures, 26 rejected alternatives |
| `_rust-build-drivers-reference.md` | Implementation-ready reference: registry entry, fingerprint algorithm, both argument vectors, the operation-private environment and manager-written Cargo config, the rejection matrix with per-surface diagnostics, the ordered Stage B classifier, the closed recognised-outcome set, 87 conformance vectors |
| `_probe.tar.gz` | Standalone Go module `rustboundaryprobe`: 9 sources, 1 test file, 14 test functions, plus `render.py` |
| `_fixture-macos.json` | `rust-boundary-fixture-v1`, macOS 26.5 arm64 / Rust 1.91.0: 19 cases, 19 matched, 0 divergences, 24 closure checks, 6 controls, 8 structural checks, `green: true`, exit 0 |
| `_fixture-absent.json` | The same probe with no resolvable toolchain: 18 cases `not_run` with the reason, exit 0, nothing installed |
| `_command-evidence.log` | Every argv, environment, real exit code and bounded output excerpt the probe ran, rendered in order |
| `_gate-log.txt` | Gate transcript with real exit codes and attribution for the two expected-red gates |

Decision and reference also written into the task worktree at
`.temp/TASK-260728-12pnm1/curator-spec-worktree/{decisions/0009-rust-driver-pair.md,docs/rust-build-drivers.md}`.
Nothing was staged, committed, pinned or published.

## The two questions this task had to answer

### 1. Can the pre-compile rejection matrix actually be computed?

Decision 0008 section 7 requires every package-selected code-execution surface
to be rejected deterministically *before* the compile phase, and disqualifies a
driver whose surfaces cannot be. Rust's two surfaces are `build.rs` and
procedural macros, and decision 0004 named exactly those when it deferred Cargo.

**Measured**: `cargo metadata --format-version 1 --offline` over a build root
whose path dependency declares `build = "build.rs"` with a marker-writing build
script, and whose second path dependency declares `[lib] proc-macro = true` with
a marker-writing macro, wrote **neither** marker and exited 0 — and the same
output reported `kind: ["custom-build"]`, `kind: ["proc-macro"]` with
`crate_types: ["proc-macro"]`, `links: "probelib"` and a per-package `source`.

The graph phase is therefore bounded and code-execution-free, and rows G2, G3,
G4 and G5 of the matrix are decidable from its output alone. This is the Rust
analogue of the Swift `dump-package` finding with the opposite result.

The matrix is run **without** `--filter-platform`, so it is host-independent and
over-approximates. That trade is stated in the decision rather than hidden.

### 2. Can the process graph stay inside a fingerprinted closure on macOS?

On macOS `rustc` resolves its linker as the name `cc` and locates the SDK by
running `xcrun`. Both are `PATH` lookups outside any fingerprinted tree, which
decision 0008 section 6 item 3 does not admit.

**Measured**, with `PATH` set to a directory of twenty logging shims that record
their name and argv and exit 127:

| Run | PATH resolutions |
|---|---|
| control: no linker pin, no SDK pin | `xcrun --sdk macosx --show-sdk-path`, then `cc <full link line>`; build failed `error: linking with 'cc' failed: exit status: 127` |
| graph phase, pinned | **0** |
| compile phase, pinned | **0** |

The pins are `SDKROOT=<declared SDK root>` and
`-Clinker=<root>/lib/rustlib/<target>/bin/rust-lld -Clinker-flavor=ld64.lld`.
The produced artifact is a `Mach-O 64-bit executable arm64` that runs, is
`adhoc, linker-signed` by the linker itself, and whose only dynamic dependency
is `/usr/lib/libSystem.B.dylib`.

So the whole process closure is `cargo → rustc → rust-lld`, three executables
inside one fingerprinted distribution, and the platform SDK is a **data-only**
second fingerprinted root that contributes no process. That is stronger than the
contract requires and strictly stronger than resolving a platform compiler
driver would give.

Consequence for the identity: `curator-rust-toolchain-v1` fingerprints an
ordered closure of roots rather than one tree, with a domain-separated
`closure_sha256` over `(role, tree digest)` pairs. Decision 0007's closed
toolchain-identifier set `{go, rust, swift, kotlin, jdk}` is untouched, the
`rust` entry declares no companion, and `toolchain_identities` stays at one
element. Measured cost: 657 MB / 167 files / 1.73 s for the Rust root, 261 MB /
32,345 files / 9.01 s for the SDK, per operation, stated rather than memoised
away.

## Other measured findings that changed the design

| # | Finding |
|---|---|
| 1 | `cargo metadata` without `--locked` **writes `Cargo.lock` into the source tree** — a write to the frozen snapshot. With `--locked` and no lock file it exits 101 and writes nothing. `--locked` is on both vectors for that reason. |
| 2 | A package `.cargo/config.toml` carrying `[build] rustc-wrapper` **executed the named script three times** during `cargo build`. An **ancestor** `.cargo/config.toml` above the build root is also discovered and applied. |
| 3 | `RUSTC_WRAPPER=` (set and empty) in the manager environment neutralises both, and a config `[env] RUSTC_WRAPPER = { value = "…", force = true }` does **not** override the neutralisation. The file is still rejected outright, because `[source]`, `[registries]`, `[patch]` and `[http]` have no environment counterpart and a package config outranks the manager's own `$CARGO_HOME` config. |
| 4 | A `rust-toolchain.toml` carrying **both** `path = "/nonexistent"` and `channel = "nightly"` is completely inert against a directly resolved `cargo`: the build succeeded with 0 `PATH` resolutions. It is a live selector only through the `rustup` shim. That is what admits `channel` as `compared` and keeps `path` `forbidden`. |
| 5 | A two-member workspace reports `workspace_members` of length 2 and `resolve.root` of `null`; both shapes are decidable in the graph phase. |
| 6 | A git dependency under `--offline` fails inside the graph phase, before any compile. |
| 7 | `rustc --print target-libdir --target <t>` exits **0** for a target whose standard library is absent, and `cargo build` then fails only after `Compiling …`. Admission is a manager-side stat of `<root>/lib/rustlib/<target>/lib` for `libstd-*.rlib`. |
| 8 | An empty `vendor` directory resolves and reports metadata with exit 0, so requiring it is an authoring rule rather than a dependency rule. |

## The acceptance layers, and one correction to decision 0007

Decision 0007's expected table names `Cargo.toml` `rust-version` as `compared`
with the rule "above resolved → mismatch". Measured, that is right for the host
relation and incomplete as a classifier: the field has **three host-independent
acceptance layers** before the host relation is reachable, where Go has two.

| Layer | Rejects | Host input |
|---|---|---|
| document | `1.85` as a TOML float → `error: invalid type: floating point '1.85', expected a semver or workspace` | none |
| grammar | `"1.85.0-beta"`, `"1.85.0+build"`, `"1.85.0.1"`, `"stable"`, `""`, `"not-a-version"` | none |
| edition floor | `"1"` at edition 2018, `"1.31"` at 2021, `"1.56"` at 2024 — measured floors 1.31.0 / 1.56.0 / 1.85.0, and edition 2015 admits `"1"` | none |
| host gate — **excluded from the layer measurement** | `"1.99"` on a 1.91.0 host → `error: rustc 1.91.0 is not supported by the following package:` | the resolved toolchain |

The three layers are pairwise independent, **measured** rather than argued: the
document layer accepts `"not-a-version"` which grammar rejects; grammar accepts
`"1"` which the edition floor rejects at 2018 and above; the edition floor would
accept `1.85` which the document layer rejects as a float.

The disposition therefore moves from `compared` to `classified`, with a
seven-class ordered exhaustive classifier and a mandatory catch-all. That is a
narrowing of what Curator accepts silently, introduces no selector, and no
schema or vector currently depends on the old row.

Rust is better placed than Go on isolation: `cargo metadata --no-deps` reports
`rust_version` `1.99` on a 1.91.0 host with exit 0, so it structurally cannot be
applying the host gate — no harness surgery is needed to isolate
representability. `cargo build --offline` is the corroborating measurement,
classified into six outcomes rather than into pass and fail.

Because the field's value space is a version and nothing else, its security
partition `F` is empty, so P1 and P2 collapse to the satisfiable equality
`C = Upstream`. The probe checks it as such; both hold.

## Probe results

`rust-boundary-fixture-v1`, macOS 26.5 arm64, Rust 1.91.0, real exit **0**:

```
cases: 19, matched 19, divergences 0, not run 0
alignment: P1 no-widening=true P2 no-narrowing=true (security partition empty=true)
closure: 24 checks, 0 yielded a verdict (must be 0)
controls: 6 of 6 failing as required
structural: 8 checks, 0 divergences
green: true
```

Recognition is whole-line exact against forms predicted **before** each command
runs, from the value under test and the probe's fixed constants — and it is a
**pair**, one first line plus one caret source line, because three distinct
grammar rejections share the first line `error: expected a version like "1.32"`
and none of the first lines names the value. Anything outside the closed set is
`unknown`, yields no verdict and fails the probe.

Closure is measured, not asserted, in both laundering directions: 4 real
unrelated command failures, 12 measured outcomes cross-fed under a wrong value,
and 8 constructed cases — 24 checks, all yielding no verdict, 5 in the
acceptance direction and 19 in the rejection direction. The constructed ones are
disclosed as constructed: an outcome upstream has not yet written cannot be
measured on any host.

The six controls are runnable from the same binary and each **must** fail:

| Control | Guards | Findings |
|---|---|---|
| C1 open classifier | closure in both directions | 4 unrelated failures would become verdicts |
| C2 lead-only recognition | three grammar classes collapsing into one | 6 diagnostics recognised under a different value |
| C3 exit status as semantics | host gate folded into the grammar layer | 1 case |
| C4 `cargo build` as the isolated command | the same folding by command choice | isolated=accepted vs corroborating=host-gate |
| C5 edition floor folded into grammar | one host-independent layer swallowing another | 3 values |
| C6 substring matching | recognising a family rather than an outcome | 6 outcomes |

A degraded run with no resolvable toolchain exits 0 with 18 cases `not_run` and
the reason recorded; nothing was installed, downloaded, updated or switched at
any point.

## Two residual exposures, stated rather than closed

- **Compile-time file inclusion.** `include!`, `include_str!` and
  `include_bytes!` can name a path outside the build root. They are reads, not
  code execution, so decision 0008 section 7 does not require their rejection —
  but no sound pre-compile rejection exists (the macro name is a token, so
  `include ! ( "x" )` and `cfg_attr` forms defeat a byte scan, while the
  substring `include` rejects ordinary comments), and **none of the six deferred
  hardened guarantees covers compile-time filesystem reads**. Recorded as a new
  input to `STORY-260728-327soo`.
- **Foreign function declarations against base-installation libraries.** The
  package cannot supply a library to link and cannot add a search path, so
  `#[link]` can only name a library the pinned link environment already
  resolves — which decision 0008 section 3 already requires the artifact to
  depend on. Not a claim that the produced program is safe.

## Platform position

`platforms` holds exactly `(macos, arm64)`. That is the honest consequence of
the evidence, not a scoping choice.

- **macOS amd64** is a qualification obligation: the Rust root ships an
  `x86_64-apple-darwin` standard library, but no x86_64 macOS host was
  available, so shipping a standard library is not evidence that the pipeline
  runs.
- **Windows** gets an implementation contract with two candidate paths —
  `lld-link` with the MSVC/Windows SDK import libraries as data-only roots
  (`lld-link` is present in the same `gcc-ld` directory as the macOS flavours on
  the measured root), or `windows-gnu` self-contained — and **no** platform
  claim, because `TASK-260729-rhjxtx` measured that no Rust toolchain exists on
  the reachable Windows host. An implementation MUST NOT ship a Windows path
  that resolves `link.exe`, `cl.exe`, `gcc`, `ld`, `vswhere` or a Visual Studio
  activation script.
- **Linux** stays excluded until `TASK-260728-1skseh`, with the same acceptance
  test plus a base-installation dependency check.

The acceptance test is identical on every candidate platform and is stated in
reference section 12.1, including the requirement that the control run must fire
— without it the zero proves nothing.

## Gates

Real exit codes, each command run standalone; full transcript in
`_gate-log.txt`.

| Gate | Exit |
|---|---|
| probe `gofmt -l .` | 0 |
| probe `go vet ./...` | 0 |
| probe `go build ./...` | 0 |
| probe `go test -count=1 ./...` | 0 |
| probe run, macOS, SDK pinned | 0 (`green: true`) |
| probe run, toolchain absent | 0 (18 `not_run`) |
| curator `go build ./...` | 0 |
| curator `go vet ./...` | 0 |
| curator `go test ./...` | 0 |
| curator `gofmt -l .` | 0 |
| curator `make check` | **2** |
| spec `tools/validate.py`, clean 57c1f56 | 0 |
| spec `tools/validate.py`, task worktree | **1** |
| scoped link check over the two authored documents | 0 |

`golangci-lint` was **not run**: the binary is not installed on this host.

Both non-zero gates are expected-red and attributed:

- `make check` fails at its `gofmt` stage on four vendored third-party files
  under `.temp/TASK-260720-1zntv0/cycle2/curator/vendor/`, another task's
  scratch tree. Zero are from this task, and no tracked project file was
  modified.
- `validate.py` passes its schema, manifest, review-evidence and
  vector-semantics checks; its link check reports 3 broken links, all inside
  `docs/external-build-repositories.md` and `docs/portable-go-execution-policy.md`,
  which were copied into the task worktree from another task's in-flight tree to
  close the link graph reachable from decision 0007. They point at
  `release/1.0.0-rc.5.json` and
  `conformance/v1/vectors/go-host-execution-policy.json`, neither of which
  exists at base commit 57c1f56. The clean-tree baseline at the same commit
  exits 0, and the scoped check over the two documents this task authored
  reports 4 local links and 0 broken.

## Scope kept

No normative curator-spec file was modified: decision 0009 and
`docs/rust-build-drivers.md` are new files, and the accepted decisions 0005
through 0008 plus two `docs/` documents were **copied** into the task worktree
unmodified so the link graph reachable from decision 0007 could be evaluated.
No schema, vector, release pin, dependency, generated corpus or release metadata
was touched. No frozen artifact was altered, nothing was staged, committed,
pinned or published. The probe is a standalone module under `.temp/` with its
own `go.mod`. Nothing was installed on any host.

## Open item for the reviewer

Decision number `0009` is reserved by this record. Swift (`TASK-260728-1yhuqi`)
and Kotlin (`TASK-260728-168smo`) were both in `backlog` at the time of writing;
if either lands `0009` first, this record renumbers rather than contests, as
`TASK-260728-2spy93` did when `0007` was claimed.
