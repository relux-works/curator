# Decision 0009: the Rust driver pair, `rust-v1` and `rust-repository-v1`

## Context

Decision 0008 reserved six driver identifiers and bound each to the portable
`manager-worker-v2` execution policy, receipt schema 3 for the local source mode
and 4 for the external one, and the single artifact class
`native-executable-v1`. Reservation is not admission: `rust-v1` and
`rust-repository-v1` leave the reserved namespace only when this contract is
accepted and `TASK-260728-251p01` moves them in the same change that mints the
schema admitting them.

Decision 0007 fixed one shared toolchain contract and left the `rust` registry
entry reserved with an explicit obligation list — probe vectors, normalization,
prerelease markers, root layout, primary-executable relpath, fingerprint
algorithm, companions, platforms, baseline, `compatibility` granularity and
initial set, and the metadata disposition table — to be completed on a qualified
host rather than asserted.

Rust is the hardest of the three reserved languages for one specific reason, and
decision 0004 already recorded it as the ground for deferring Cargo: `cargo
build` compiles and executes `build.rs`, and `rustc` loads and executes
procedural macros inside the compiler. Both are package-selected code running
during manager activity, which is the boundary the protocol exists to hold.
Decision 0008 section 7 therefore does not ask how to accommodate them; it
requires that every package-selected code-execution surface be rejected
deterministically **before** the compile phase, and that a surface which cannot
be is grounds to reject the driver.

Rust adds a second problem the Go driver never had. `go build` links with the Go
toolchain's own internal linker and starts nothing outside `GOROOT`. On macOS
`rustc` resolves the linker as the name `cc` through `PATH` and locates the
platform SDK by running `xcrun`. Both are measured below. Under decision 0008
section 6 item 3 an executable outside the fingerprinted closure is not
admissible, so a Rust driver either binds those components or rejects the
platform — and macOS is the platform this release primarily targets.

This decision resolves both problems with measurement rather than assertion, and
fixes the complete `rust-v1` and `rust-repository-v1` contract: the trusted
closure, the two argument vectors, the operation-private environment and
manager-owned Cargo configuration, the exhaustive pre-compile rejection matrix,
the Stage B disposition table and its acceptance layers, the artifact and
identity model, and the platform matrix. The implementation-ready reference is
[`docs/rust-build-drivers.md`](../docs/rust-build-drivers.md).

### Evidence basis

Every measured claim in this decision was produced on one host and is labelled
as such. Nothing here is a platform claim; decision 0008 section 9 keeps both
identifiers at an empty qualified-platform set until `TASK-260728-2bu2q6`.

| | Value |
|---|---|
| Host | macOS 26.5, arm64 |
| Rust | 1.91.0, `rustc 1.91.0 (f8297e351 2025-10-28)`, `cargo 1.91.0 (ea2d97820 2025-10-10)` |
| Root | a directly resolved `rustup`-produced toolchain directory, never the `rustup` shim |
| SDK | `MacOSX.sdk` from Xcode 26.5 |
| Windows | no Rust toolchain present on the reachable Windows host (`TASK-260729-rhjxtx`, 19 cases `not_run`) |
| Linux | not probed |

`TASK-260729-rhjxtx` supplied the version, target, metadata-readability and
selector measurements and is still review-pending; this decision re-measured
each fact it relies on and records the re-measurement, so no unreviewed
observation is load-bearing on its own.

## Decision

### 1. What the pair admits

`rust-v1` compiles one Rust program from a vendored, dependency-closed build
root inside the consuming skill snapshot. `rust-repository-v1` compiles one Rust
program from a build root inside a locked external Git repository. The two share
everything except source acquisition, audit subject, receipt schema, and marker
state, exactly as `go-v1` and `go-repository-v1` do.

| | `rust-v1` | `rust-repository-v1` |
|---|---|---|
| Source mode | local snapshot | external locked repository |
| Receipt schema | 3 | 4 |
| Execution policy | `manager-worker-v2` | `manager-worker-v2` |
| Source identity | `curator-build-source-v1` over the consuming snapshot | `curator-build-source-v1` over the external snapshot |
| Command shape | `buildCommandV8` | `repositoryBuildCommandV2` + `skillBuildTargetV2` |
| Project metadata | `Cargo.toml` at the build root | `Cargo.toml` at the selected build root |
| Artifact class | `native-executable-v1` | `native-executable-v1` |

`rust-repository-v1` reuses this decision's trusted closure, native target,
argument vectors, environment, Cargo configuration, rejection matrix,
disposition table, link policy, staging rules and no-execution rule without
reinterpretation or widening. Where a rule below says "the driver", it means
both identifiers.

### 2. The trusted closure is two roots, and only one of them holds executables

**Measured, macOS 26.5 arm64, Rust 1.91.0.** With `PATH` set to a directory
containing twenty logging shims (`cc`, `clang`, `ld`, `xcrun`, `ar`, `dsymutil`,
`sh`, `lld`, `ld64.lld`, `codesign` and ten more), each of which appends its own
name and argv to a log file and exits 127:

- a `cargo build` with no linker pinned and no SDK pinned resolved **`xcrun
  --sdk macosx --show-sdk-path`** and then **`cc`** from `PATH`, and failed with
  `error: linking with 'cc' failed: exit status: 127`;
- the same build with `SDKROOT` set to an absolute SDK path and
  `CARGO_ENCODED_RUSTFLAGS` carrying
  `-Clinker=<root>/lib/rustlib/<target>/bin/rust-lld` and
  `-Clinker-flavor=ld64.lld` produced **zero** entries in the log, exited 0, and
  produced a `Mach-O 64-bit executable arm64` that runs and whose only dynamic
  dependency is `/usr/lib/libSystem.B.dylib`;
- the graph command `cargo metadata --format-version 1 --locked --offline`
  produced **zero** entries in the log.

The control is what makes the zero meaningful: the shims demonstrably fire when
the pins are removed, in the same working directory, in the same run sequence.

Two consequences are fixed here.

**The process closure is exactly three executables, all inside the Rust
distribution root.** `<root>/bin/cargo` is the driver's trusted launcher;
`<root>/bin/rustc` and `<root>/lib/rustlib/<target>/bin/rust-lld` are the
fingerprinted regular executables below it. No Apple `clang`, `ld`, `xcrun`,
`ar` or `dsymutil` process participates. The `manager-worker-v2` graph is
therefore satisfied by the primary root alone:

```text
manager parent
  -> identity-verified manager-owned worker
       -> fingerprinted <rust-root>/bin/cargo
            -> fingerprinted <rust-root>/bin/rustc
            -> fingerprinted <rust-root>/lib/rustlib/<target>/bin/rust-lld
```

**The platform SDK is bound as a second fingerprinted root that contributes no
process.** It is data — headers, `.tbd` stubs and library stubs the linker
reads. Decision 0008 section 6 item 3 names an SDK explicitly among the
components a driver must bind or reject a platform over, so binding it is
required even though it starts nothing.

The `rust` registry entry therefore declares, per operating system, an ordered
closed list of **link-support roles** the Rust distribution does not itself
provide. On macOS that list is exactly one role, `platform-sdk`, and its value
is data-only. A link-support root is resolved through the same two declaration
channels decision 0007 section 3 fixes for a toolchain root — a root bundled
with the manager distribution, or trusted operator configuration in
manager-owned owner-protected state — and through nothing else. `PATH`, the
inherited environment, `xcrun`, `xcode-select`, `DEVELOPER_DIR`, a package byte,
a descriptor byte and a version-manager shim are all forbidden origins for it,
with the same force and the same diagnostics as for the primary root. A missing
or unusable link-support declaration is a Stage A failure before any source is
acquired.

This does **not** coin a new toolchain identifier. Decision 0007's closed set
`{go, rust, swift, kotlin, jdk}` is untouched, the `rust` entry declares **no
companion toolchain**, and `toolchain_identities` carries exactly one element
for both drivers. What changes is the internal shape of the algorithm this
decision owns: `curator-rust-toolchain-v1` fingerprints an ordered closure of
roots rather than a single tree.

`curator-rust-toolchain-v1` produces exactly:

```json
{
  "algorithm": "curator-rust-toolchain-v1",
  "rust_version": "release: 1.91.0",
  "cargo_version": "cargo 1.91.0 (ea2d97820 2025-10-10)",
  "launcher_relpath": "bin/cargo",
  "rustc_relpath": "bin/rustc",
  "linker_relpath": "lib/rustlib/aarch64-apple-darwin/bin/rust-lld",
  "roots": [
    {"role": "rust-distribution", "tree_sha256": "sha256:..."},
    {"role": "platform-sdk", "tree_sha256": "sha256:..."}
  ],
  "closure_sha256": "sha256:..."
}
```

`roles` is a closed manager-owned token set. No root *path* appears: toolchain
location is not portable identity, exactly as decision 0007 section 3.2
requires. Each per-root digest uses the same walk, ordering, record framing and
link rules as `curator-go-toolchain-v1`, with its own domain prefix;
`closure_sha256` is domain-separated over the ordered role and per-root digest
pairs, so a two-root closure can never collide with a one-root closure that
happens to hash the same bytes. `curator-go-toolchain-v1` is untouched, frozen
and Go-only, and neither Rust driver reuses, extends or aliases it.

**Measured fingerprint cost, same host.** The Rust distribution root is 657 MB
across 167 regular files and hashes in 1.73 s wall clock; the SDK root is 261 MB
across 32,345 regular files and 7,448 symlinks and hashes in 9.01 s. The cost is
per operation and per root, it is dominated by the SDK's file count rather than
its size, and it is stated rather than optimised away: memoising it across
operations would weaken exactly the property decision 0007 says fingerprinting
proves.

### 3. Native target: representability is not admission

**Measured, same host.** `rustc --print target-libdir --target
x86_64-unknown-linux-gnu` exits **0** and prints a path for a target whose
standard library is absent from the tree; `cargo build --target` for that
triple then fails only **after** `Compiling probe-positive v0.1.0`, that is
after the compile phase has begun. An unknown triple is different:
`--print target-libdir --target not-a-real-target` exits 1 before any
compilation.

The driver is native-only, so this matters as a gate rather than as a feature.
The native target is `rustc --print host-tuple`, which on this host reports
`aarch64-apple-darwin` and carries no platform-version component, so it is
directly usable as a target identity without normalization. Admission is a
manager-side check inside the fingerprinted tree: the directory
`<rust-root>/lib/rustlib/<native-target>/lib` MUST exist and MUST contain a
regular file matching `libstd-*.rlib`. A target whose standard library is absent
fails Stage A with `build_toolchain_platform_unsupported`, before source
acquisition and before any compiler child. `rustc --print target-list` and
`--print target-libdir` are representability surfaces and MUST NOT be used as
the admission test.

Cross-compilation is forbidden. The compile command always passes `--target`
with the resolved native tuple, so the target is manager-selected rather than
inferred, and the artifact path is fully determined.

### 4. Local source ownership: one build root, one package, one program

`rust-v1` reuses the schema-6 and schema-7 `build_roots` model without change,
and adds exactly the bindings decision 0008 section 4 requires of a local
driver.

- The build root MUST contain `Cargo.toml` directly, and that file MUST be the
  nearest ancestor `Cargo.toml` of `source_dir`.
- `source_dir` MUST equal its build root. A Rust package has no compilable
  sub-unit that a directory path can name — a binary target is selected by name,
  not by directory — so any other value would either be inert or would require a
  package-controlled selector member the boundary forbids.
- The build root MUST contain `Cargo.lock` directly and a `vendor` directory
  directly. `vendor` MAY be empty; **measured** that an empty vendored directory
  source resolves and reports metadata with exit 0.
- The resolved graph MUST have exactly one workspace member, and it MUST be the
  build-root package. **Measured** that a two-member workspace reports
  `workspace_members` of length 2 and `resolve.root` of `null`, so both shapes
  are decidable from the graph phase alone.
- The build-root package MUST have exactly one target whose `kind` is exactly
  `["bin"]`, and that target's name MUST match
  `^[A-Za-z0-9_][A-Za-z0-9_-]{0,63}$`.

That chain is the deterministic non-discovering mapping decision 0008 section 4
demands: build root → its single package → its single `bin` target → one
executable. The consuming manifest command key remains the sole naming
authority; the target name is used only as the manager-derived
`--bin` argument and to locate the produced file, and its grammar is checked
before it reaches an argument vector.

Two programs require two build roots. That duplicates a vendored tree, and it is
the accepted cost of refusing a package-controlled target selector.

### 5. External source ownership

A `rust-repository-v1` command requires a `skill-build.json` **schema 2**
descriptor. Against a schema-1 descriptor it fails
`build_descriptor_driver_unsupported`, and against an unsupported descriptor
version `build_descriptor_schema_unsupported`, with no fallback to another
target, another driver, `go-repository-v1`, a script, a system command or a
generic build facility.

The descriptor target's `build_root` and `source_dir` carry their schema-1
meaning. For this driver `source_dir` MUST equal `build_root`, by the same
argument as section 4; `build_root` MAY be `.`, which the descriptor already
admits and which does not affect the schema-6 prohibition on a local
`build_root` of `.`. Every rule of section 4 about `Cargo.toml`, `Cargo.lock`,
`vendor`, workspace membership and the single `bin` target applies unchanged to
the selected external build root.

The whole external snapshot remains the validation, identity and audit subject;
only the selected build root is compiler-visible; no external repository byte is
agent-facing or runtime-copied. Input MUST NOT come from the consuming skill,
another external repository, a sibling or parent directory outside the selected
build root, a host Cargo registry cache, a host `CARGO_HOME`, or the network.

### 6. The fixed process graph and exactly two commands

One `manager-worker-v2` session performs exactly one graph phase and exactly one
compile phase. The driver uses the graph phase, so its session shape is the
same "one graph command, one compile command" shape `manager-worker-v1` fixes
for Go, under the v2 identity and the v2 process graph.

With the canonical `source_dir` as the working directory, the manager MUST use
exactly these two argument vectors and MUST NOT alter, extend, reorder or repeat
them:

```text
cargo metadata --format-version 1 --locked --offline --color never --quiet
cargo build --locked --offline --color never --quiet --release --target <native-tuple> --bin <bin-target-name>
```

`<native-tuple>` is the Stage A resolved host tuple and `<bin-target-name>` is
the validated single `bin` target name from the graph phase. Both are
manager-derived from validated manager-owned data; neither is copied from a
manifest, a descriptor, or an unvalidated package string.

Three package-independent probe vectors run once per operation from the manager
parent during Stage A, from a manager-owned empty working directory:

```text
rustc -vV
rustc --print host-tuple
cargo --version
```

`--locked` is load-bearing and not a convenience. **Measured** that `cargo
metadata` without it *writes* `Cargo.lock` into the source tree — a write to the
frozen snapshot — and that with `--locked` and no lock file present it instead
exits **101** with `error: the lock file ... needs to be updated but --locked
was passed to prevent this` and writes nothing.

The produced file is `<CARGO_TARGET_DIR>/<native-tuple>/release/<bin-target-name>`
on Unix and `...\release\<bin-target-name>.exe` on Windows. `CARGO_TARGET_DIR`
is an operation-private manager staging root, so the artifact is produced inside
manager staging, is hashed there, and is published as `bin/<command>` or
`bin/<command>.exe` derived solely from the consuming manifest command key.

### 7. The operation-private environment and the manager-owned Cargo configuration

The environment starts empty except for indispensable operating-system process
variables. It sets exactly the variables the reference document enumerates:
private `HOME`, `TMPDIR`, `XDG_CONFIG_HOME`, `XDG_CACHE_HOME`, `CARGO_HOME` and
`CARGO_TARGET_DIR` (and on Windows private `APPDATA`, `LOCALAPPDATA`,
`USERPROFILE`, `TEMP` and `TMP`); `PATH` naming a manager-owned empty directory;
`LC_ALL=C` and `LANG=C`; absolute `RUSTC` and `RUSTDOC` inside the fingerprinted
root; empty `RUSTC_WRAPPER` and `RUSTC_WORKSPACE_WRAPPER`;
`CARGO_ENCODED_RUSTFLAGS` carrying exactly the fixed linker pins;
`CARGO_INCREMENTAL=0`; `CARGO_NET_OFFLINE=true`, `CARGO_NET_RETRY=0`,
`CARGO_NET_GIT_FETCH_WITH_CLI=false`; and on macOS `SDKROOT` naming the resolved
`platform-sdk` root. No Rust, Cargo, `rustup`, compiler, linker, SDK or
executable-search variable is inherited, and `RUSTC_BOOTSTRAP` MUST be absent
because it turns a release toolchain into one that accepts `-Z` flags.

Two of those settings are answers to measured attacks rather than hygiene.

**`RUSTC` must be absolute.** `TASK-260729-rhjxtx` measured that cargo resolves
`rustc` by name from `PATH` and fails with `could not execute process 'rustc
-vV' (never executed)` under a minimal `PATH`; without an explicit `RUSTC` the
second node of the process graph would be chosen by `PATH` order rather than by
the fingerprinted closure. This decision re-measured the working direction: with
absolute `RUSTC` and an empty `PATH`, no `PATH` lookup occurs at all.

**A package `.cargo/config.toml` is a live process-injection surface.**
**Measured**: a `.cargo/config.toml` at the build root carrying `[build]
rustc-wrapper = "/tmp/probe-wrapper.sh"` caused that script to be executed three
times during `cargo build`, while the manager's `CARGO_ENCODED_RUSTFLAGS` still
took precedence over the same file's `rustflags`. **Measured**: a
`.cargo/config.toml` in an *ancestor* directory above the build root is
discovered and applied the same way. **Measured**: setting `RUSTC_WRAPPER` to
the empty string in the manager environment neutralises both, and a config
`[env] RUSTC_WRAPPER = { value = "...", force = true }` does **not** override it.

Both defences are kept, and neither is trusted alone. The environment
neutralises the settings that have an environment counterpart; the rejection
matrix of section 8 refuses the file outright, because `[source]` replacement,
`[registries]`, `[patch]` and `[http]` have no environment counterpart and a
package config file takes precedence over the manager's own `$CARGO_HOME`
config. In addition, the manager MUST guarantee that no `.cargo` directory
exists in any ancestor of the compile working directory; that is a manager
staging obligation, not a package check, and it fails
`build_execution_control_unavailable` when it cannot be met.

The manager writes `$CARGO_HOME/config.toml` itself, with a fixed grammar and
exactly four tables — `[source.crates-io]` replacing with `curator-vendor`,
`[source.curator-vendor]` naming the absolute path of `<build_root>/vendor`,
`[net] offline = true`, and `[term]` fixing quiet and colour. The only
package-derived component is the driver-fixed relative path `vendor`; nothing a
package writes reaches that file.

### 8. The exhaustive pre-compile rejection matrix

The matrix is computed from the validated snapshot and the graph phase, before
the compile phase, and it is total: every surface below has exactly one verdict,
and a surface that cannot be decided there is grounds to reject the driver
rather than to add a runtime allowance.

Three properties hold across the whole matrix.

**The graph phase executes no package code.** **Measured**: a build root whose
path dependency declares `build = "build.rs"` and whose `build.rs` writes a
marker file, together with a second path dependency whose library is
`proc-macro = true` and whose macro writes a second marker file, produced
**neither** marker under `cargo metadata --format-version 1 --offline`, which
exited 0. This is the Rust analogue of the Swift `dump-package` finding, with
the opposite result: Rust's metadata query is bounded and does not compile or
run the package.

**The surfaces that matter are visible in that output.** **Measured**, the same
run reported `kind: ["custom-build"]` for the build script target, `kind:
["proc-macro"]` and `crate_types: ["proc-macro"]` for the macro crate,
`links: "probelib"` for the package declaring a native library, and `source`
per package.

**The matrix is host-independent.** The graph command is run without
`--filter-platform`, so `packages[]` is the union over every platform and every
feature rather than the resolution for this host. That is a deliberate
over-approximation: it is a superset of what the compile phase builds, so it can
never miss an admitted-and-then-executed surface, and the same snapshot produces
the same verdict on every host. Its cost is that a dependency that would only be
built on another platform, or only under a feature this build does not enable,
still rejects the command. That cost is accepted and stated rather than traded
for a host-dependent verdict.

| Surface | Verdict | Decided by |
|---|---|---|
| build script (`build.rs`, any `build` key, any `custom-build` target in any package) | reject | graph phase |
| procedural macro (`proc-macro` target kind or crate type in any package) | reject | graph phase |
| compiler plugin, codegen backend, `-Z` flag, `cargo-features` manifest key, `RUSTC_BOOTSTRAP` | reject | snapshot bytes and fixed environment |
| `rustc-wrapper`, `rustc-workspace-wrapper`, any `.cargo` directory in the build root subtree | reject | snapshot bytes, plus empty-valued environment |
| `links` key on any package | reject | graph phase |
| dependency source other than a build-root path package or the vendored crates.io registry — git, alternate registry, local-registry, unknown | reject | graph phase |
| `[patch]`, `[replace]`, `[workspace]` table in the build-root manifest | reject | snapshot bytes |
| more than one workspace member, or a virtual manifest | reject | graph phase |
| zero or more than one `bin` target in the build-root package | reject | graph phase |
| `crate_types` outside `{bin, lib, rlib}` — `dylib`, `cdylib`, `staticlib`, `proc-macro` | reject | graph phase |
| a package file or target source outside the build root, or a registry package outside `<build_root>/vendor` | reject | graph phase |
| native object, archive, shared library or foreign source file in the compiler-visible tree, by closed extension list | reject | snapshot bytes |
| network and registry access, index update, git fetch | reject | `--offline`, `--locked`, manager-written config, empty `PATH` |
| package-selected linker, linker flavour, link argument, library search path | reject | fixed `CARGO_ENCODED_RUSTFLAGS`, no config file, no build script |
| package-selected toolchain path, root, channel, mirror, installer, version manager | reject | decision 0007 resolution, and `rust-toolchain*` classified rather than honoured |
| cross-compilation, non-native `--target` | reject | fixed compile vector |
| Cargo features and `[profile]` tables | admit | see below |
| `#[link]` attributes and `extern` blocks in package source | admit, bounded | see section 13 |
| `include!`, `include_str!`, `include_bytes!` | admit, bounded | see section 13 |

Cargo features are admitted because they select which package source compiles,
which is the same class of choice as a Go build constraint, and because the
manifest command object cannot express one — the build always uses the root
package's default feature resolution. `[profile]` tables are admitted for the
same reason and one measurement: on a release toolchain, per-profile `rustflags`
are gated behind `-Z profile-rustflags`, and the driver rejects both `-Z` flags
and the `cargo-features` opt-in that would unlock them, so a profile can select
codegen tuning and cannot inject a flag or a process. A `[profile]` table that
enables link-time optimisation stays inside `rustc` and its bundled LLVM; it
starts no additional process.

The shared semantic class for a rejection in this matrix is
`build_package_code_execution_forbidden`; the reference document names the
per-surface diagnostic beneath it for each row.

### 9. Stage B metadata dispositions and Rust's acceptance layers

Decision 0007's disposition framework, precedence rule and channel
classification are fixed there and are not reopened. This section confirms the
expected `rust` rows on a qualified host, corrects one of them, and supplies the
acceptance-layer analysis decision 0007 requires.

**Confirmed by measurement.** A build root carrying a `rust-toolchain.toml` with
*both* `path = "/nonexistent"` and `channel = "nightly"` built successfully
through a directly resolved `cargo`, with zero `PATH` resolutions. The same file
redirects the `rustup` shim — `TASK-260729-rhjxtx` measured `error: invalid
toolchain: the path '/nonexistent/trusted/root' has no bin/ directory` and, for
`channel = "nightly"`, `info: syncing channel updates for
'nightly-aarch64-apple-darwin'` followed by a download attempt against a
deliberately unreachable dist server. The file is a live selector *through the
shim* and completely inert against the direct resolution decision 0007 mandates.
That is exactly what admits `channel` as `compared` rather than `forbidden`, and
it is why `path` stays `forbidden`: it names where a toolchain comes from.

**Corrected.** Decision 0007's expected row names `Cargo.toml` `rust-version`
with the rule "above resolved → mismatch". That is right for the host relation
and incomplete as a classifier, because the field has three host-independent
acceptance layers before the host relation is reachable at all.

**Measured**, `cargo metadata --no-deps --format-version 1 --offline`, one value
per fixture, exit code and first diagnostic line:

| Value | Exit | Layer that rejected | First line |
|---|---|---|---|
| `"1.85"` | 0 | — | accepted, `rust_version` `1.85` |
| `"1.85.0"` | 0 | — | accepted, `rust_version` `1.85.0` |
| `1.85` (TOML float) | 101 | document | `error: invalid type: floating point '1.85', expected a semver or workspace` |
| `"1.85.0-beta"` | 101 | grammar | `error: unexpected prerelease field, expected a version like "1.32"` |
| `"not-a-version"` | 101 | grammar | `error: unexpected prerelease field, expected a version like "1.32"` |
| `"1.85.0+build"` | 101 | grammar | `error: unexpected build field, expected a version like "1.32"` |
| `"1.85.0.1"` | 101 | grammar | `error: expected a version like "1.32"` |
| `"stable"` | 101 | grammar | `error: expected a version like "1.32"` |
| `""` | 101 | grammar | `error: expected a version like "1.32"` |
| `"1"` | 101 | edition floor | `error: failed to parse manifest at '<path>'`, caused by `rust-version 1 is older than first version (1.56.0) required by the specified edition (2021)` |

So Rust's structure is **three host-independent layers plus one host gate**,
where Go has two plus one:

| Layer | What it decides | Host input |
|---|---|---|
| document | the TOML document parses and the value has an admissible type | none |
| grammar | cargo's `rust-version` grammar can represent the value | none |
| edition floor | the value is not below the floor the manifest's own `edition` requires | none |
| host gate | the resolved toolchain satisfies the value | the resolved toolchain |

The three layers are pairwise independent in the same sense decision 0007
established for Go, and the independence is measured rather than argued: the
document layer accepts `"not-a-version"`, which the grammar layer rejects; the
grammar layer accepts `"1"`, which the edition floor rejects; and the edition
floor would accept `1.85`, which the document layer rejects as a float. No layer
contains another, so a classifier pinned to any one of them admits values the
others refuse.

The host gate is separate and is **excluded from the layer measurement**, for
exactly the reason decision 0007 gives: a gate that depends on the runner cannot
be part of a value's grammar. **Measured**, `cargo build --offline` with
`rust-version = "1.99"` on a 1.91.0 host exits 101 with `error: rustc 1.91.0 is
not supported by the following package:` and `probe-future@0.1.0 requires rustc
1.99`, and `build_started_compilation` is false — cargo applies the gate before
compiling. Curator does not wait for it: the value is compared against the
resolved toolchain identity in Stage B, so the rejection happens before cache
lookup and before any compiler child.

Rust is in one respect better placed than Go. The isolated representability
measurement needs no harness surgery: `cargo metadata --no-deps` reports
`rust_version` for a value above the host (**measured**, `"1.99"` on a 1.91.0
host, exit 0) and so structurally cannot be applying the host gate. The
corroborating command is `cargo build --offline`, whose outcome is classified
into `accepted`, `rejected-document`, `rejected-grammar`, `rejected-edition`,
`host-gate` and `unknown`, never into pass and fail, and which is required to
agree with the isolated measurement on the three layers.

The recognised outcome set is closed and whole-line exact. Every rejection above
renders a caret block whose third line is `5 | rust-version = <literal>` — the
literal under test at a line number the probe fixture fixes — so a recognised
outcome is the pair of one exactly predicted first line and one exactly
predicted source line, and the upstream constants it embeds (`"1.32"`, the
`1.56.0` edition floor, the edition token) are probe fixed constants. Anything
else is **unknown**, yields no verdict and fails the probe. A lead with an
unconstrained tail, and a substring found anywhere in the output, are families
rather than outcomes and MUST NOT be recognised. Closure is measured, not
asserted: the probe feeds the classifier unrelated real command failures, every
measured outcome cross-fed under the wrong value, and measured diagnostics
extended the way a later release would extend them, and requires each to produce
no verdict, reporting which laundering direction each fabrication belongs to.

The confirmed and corrected `rust` disposition table is:

| Source | Field | Disposition |
|---|---|---|
| `Cargo.toml` | `package.rust-version` | `classified` — three layers plus the host relation, per the reference document |
| `Cargo.toml` | `workspace.package.rust-version` | not reachable: a workspace is rejected by section 8, so this row is unreachable rather than admitted |
| `rust-toolchain.toml` | `toolchain.path` | `forbidden` |
| `rust-toolchain.toml` | `toolchain.channel` | `compared`, by decision 0007's channel table |
| `rust-toolchain.toml` | `toolchain.components`, `targets`, `profile` | `ignored` |
| `rust-toolchain` (legacy one-line form) | the bare channel string | `compared`, same classifier as `toolchain.channel` |

`rust-version` moves from `compared` to `classified` for the same structural
reason decision 0007 classified the two `go.mod` directives: one field whose
value space spans more than one disposition needs an ordered exhaustive value
classifier, not a single comparison rule. It gains no `forbidden` class, because
its value space is a version and nothing else. This is a narrowing of what
Curator accepts silently, not a widening, and it introduces no selector.

Both `rust-toolchain` files are evaluated in Unicode-scalar lexical order of
relative source path, so `rust-toolchain` precedes `rust-toolchain.toml`, and
within each file `forbidden` precedes `compared`, so a file carrying both `path`
and `nightly` is deterministically a package-influence rejection.

### 10. The `rust` registry entry

| Field | Value |
|---|---|
| `toolchain_id` | `rust` |
| `primary_relpath` | `bin/cargo`; `bin/cargo.exe` on Windows |
| `probe` | `rustc -vV`, `rustc --print host-tuple`, `cargo --version`, from a manager-owned empty working directory under the operation-private environment |
| `normalization` | the single line of `rustc -vV` **stdout** matching `^release: (0\|[1-9][0-9]*)\.(0\|[1-9][0-9]*)\.(0\|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$`; a non-empty fourth group is a prerelease; `cargo --version` stdout must normalize to the same triple or the version is undetermined |
| `fingerprint_algorithm` | `curator-rust-toolchain-v1`, section 2 |
| `baseline` | `{"kind":"at_least","min":"1.91.0"}` |
| `compatibility` | families `{(1, 91)}`; family granularity `(major, minor)` |
| `platforms` | `(macos, arm64)` only |
| `companions` | none |
| `link_support_roles` | `macos`: exactly `[platform-sdk]`, data-only |
| `metadata_sources` | `Cargo.toml` `package.rust-version`; `rust-toolchain.toml`; `rust-toolchain` |

`rustc --version` is deliberately not the version probe: it embeds a commit hash
on the same line, so the verbose form is the narrower surface. The `host:` line
of `rustc -vV` carries the same value as `--print host-tuple` on this host and is
deliberately not the target probe, for the same reason — the dedicated print is
narrower.

The baseline and the compatibility set are both `1.91` because 1.91.0 is the
only release this contract was measured against. Lowering the baseline requires
measuring the older release; adding a family requires testing it against the
driver's conformance vectors. Neither may be derived from version ordering, and
no package byte can reach either.

`platforms` holds one pair, and that is the honest consequence of the evidence
rather than a scoping choice. `(macos, amd64)` is a qualification obligation
with a stated acceptance test, not a claim: the Rust root ships an
`x86_64-apple-darwin` standard library, but no x86_64 macOS host was available,
so the pair is absent until a probe run on one exists. On a host whose pair is
not in `platforms`, Stage A fails `build_toolchain_platform_unsupported` from
the pre-resolution half of the check, on registry data alone.

### 11. Artifact, cache, receipt, marker and claim identity

The artifact class is `native-executable-v1` and nothing else. **Measured** on
this host that the pipeline produces exactly one Mach-O arm64 executable whose
sole dynamic dependency is `/usr/lib/libSystem.B.dylib` — a base-installation
library — and that it is `adhoc, linker-signed`. That signature is applied by
`rust-lld` as part of linking, which is exactly the case decision 0008 section 3
names as compiler output rather than a manager signing step: it is produced by
the driver's fixed argument vector, selects no signing identity, credential or
notarization, and reaches no network. The driver performs no manager
post-build signing, and a platform policy requiring a locally signed binary must
wait for the separately reviewed signer profile.

Every other file cargo leaves in the target directory — rlibs, `.rmeta`,
dependency info, incremental state, split debug information — is a compiler
by-product. It stays in operation-private staging, is discarded with it, and
never enters cache identity, the receipt, the marker, the shim relationship, or
publication. The manager MUST NOT execute the artifact for validation, version
discovery, smoke testing, post-processing, receipt generation, rollback or any
other reason.

The canonical build input binds, in addition to the members decision 0008
section 8 requires of every new driver, the complete
`curator-rust-toolchain-v1` identity of section 2 — including both root digests
and the closure digest — the native tuple, the validated `bin` target name, and
this closed policy object:

```json
{
  "dependency_mode": "vendor-directory",
  "network": "none",
  "workspace": false,
  "build_scripts": false,
  "proc_macros": false,
  "plugins": false,
  "features": "manifest-default",
  "target_mode": "native",
  "profile": "release",
  "linker": "toolchain-rust-lld",
  "link_mode": "internal",
  "native_inputs": false,
  "package_config": "rejected",
  "compiler_directives": "reject-nonstandard-native-inputs-v1",
  "incremental": false,
  "execution_policy": "manager-worker-v2"
}
```

`execution_policy` is the `const` `manager-worker-v2`, so a Rust entry can never
alias a Go entry, and `driver` differs besides. `network: "none"` denotes the
fixed offline Cargo configuration, `--offline`, `--locked`, the manager-written
`$CARGO_HOME` config and the empty `PATH`; it is not a claim of kernel-enforced
network denial, which remains the deferred `total-network-denial` guarantee.

Receipt schema 3 carries the local mode and schema 4 the external mode, each a
strict `oneOf` discriminated by the `driver` `const`, each carrying this policy
object, this toolchain identity and this native target. Marker schema 4 records
the driver, `receipt_schema_version` and `execution_policy` per build entry, and
a reader rejects a `rust-v1` entry claiming `manager-worker-v1` rather than
inferring the policy from the driver name. Conformance claim schema 4 asserts
`rust-v1` and `rust-repository-v1` with `execution_policy` `manager-worker-v2`
bound by the assertion's own `driver` `const`, and only if this contract is
accepted and `TASK-260728-251p01` moves both identifiers in the same change that
mints the schema.

The effective toolchain requirement and the `compatibility` set stay gates
rather than build inputs, exactly as decision 0007 fixes them.

### 12. Platform matrix

| Platform | Status |
|---|---|
| macOS arm64 | the complete pipeline is measured on one host; it enters a claim only through `TASK-260728-2bu2q6` with immutable native evidence for that exact tuple |
| macOS amd64 | qualification obligation; the standard library ships in the same root, and the acceptance test is a probe run on an x86_64 macOS host |
| Windows | implementation contract only, stated below; **no** platform claim, because no reachable Windows host had a Rust toolchain |
| Linux | excluded until `TASK-260728-1skseh`, then qualified by the same acceptance test |

**Windows implementation contract.** The Windows obligation is to reproduce
section 2's property — a process closure with no `PATH`-resolved executable —
and it has two candidate paths, neither of which this decision claims:

1. `x86_64-pc-windows-msvc` with `-C linker=<root>/lib/rustlib/<target>/bin/rust-lld`
   and `-C linker-flavor=lld-link`, with the MSVC and Windows SDK import
   libraries bound as one or more data-only `platform-sdk` link-support roots.
   `lld-link` is present in the same `gcc-ld` directory as the macOS flavours on
   the measured root, which makes this the path to test first.
2. `x86_64-pc-windows-gnu` with the target's bundled self-contained linking
   artifacts, which if sufficient would need no link-support root at all.

The acceptance test for either is exactly the one run on macOS: the twenty-shim
poisoned-`PATH` run must produce zero entries for both the graph and the compile
phase, with a control run that produces entries when the pins are removed, and
the produced executable must depend only on base-installation libraries. Until
that evidence exists, `platforms` excludes Windows and both drivers fail
`build_toolchain_platform_unsupported` there. An implementation MUST NOT ship a
Windows path that resolves `link.exe`, `cl.exe`, `gcc`, `ld`, `vswhere` or a
Visual Studio activation script from `PATH`, the registry, or an environment
variable, and MUST NOT answer the gap with a host-resolved tool or a downgraded
control.

**Linux qualification rules.** Linux enters `platforms` only when, on the
qualifying host: `rustc --print host-tuple` reports a tuple whose standard
library is present in the fingerprinted root by the section 3 stat; the
poisoned-`PATH` run is clean for both phases with a firing control; the produced
ELF executable's dynamic dependencies are all base-installation libraries of the
declared distribution baseline; and the `platform-sdk` role is either absent or
resolved from a declaration channel. `x86_64-unknown-linux-gnu` will require a
`platform-sdk` link-support root holding the C runtime startup objects and
`libc` stubs, and `x86_64-unknown-linux-musl` may not, which is the qualification
question rather than a claim.

### 13. Two residual exposures, stated rather than closed

Decision 0008's security section requires each driver contract to state its own
compiler-input exposure honestly and to refrain from relying on containment this
protocol does not provide. Two Rust surfaces are admitted with bounds rather
than rejected, and both are named here so that no reader, receipt, marker or
claim can imply otherwise.

**Compile-time file inclusion.** `include!`, `include_str!` and `include_bytes!`
resolve a path relative to the including file and can name a path outside the
build root. They are **reads**, not code execution, so decision 0008 section 7
does not require their rejection; but a read that lands in the produced binary
is an exfiltration surface, and the portable policy does not contain the
compiler's filesystem access. No deterministic pre-compile rejection exists that
is sound in both directions: the macro name is a token, not a byte pattern —
`include ! ( "x" )` and `#[cfg_attr(unix, ...)]` forms defeat a byte scan, while
a scan for the substring `include` rejects ordinary comments and identifiers.
The surface is therefore admitted, bounded only by the operation-private
environment and the frozen snapshot, and recorded as an input to
`STORY-260728-327soo`: none of the six deferred hardened guarantees currently
covers compile-time filesystem reads, so closing it needs a seventh.

**Foreign function declarations against base-installation libraries.** Package
source may declare `extern` blocks and `#[link]` attributes. The package cannot
supply a library to link — native files in the tree are rejected, `links` is
rejected, build scripts are rejected, and there is no admitted path that adds a
library search path — so `#[link]` can only name a library the pinned link
environment already resolves, which is a base-installation library. Decision
0008 section 3 already requires the artifact to depend on exactly those, so this
is inside the artifact class rather than an escape from it. What it is not is a
guarantee that the produced program is safe; the artifact remains untrusted
package output that the manager never executes.

Both statements are about the produced program and about compile-time reads.
Neither weakens the process-graph, network, write or resource properties of
section 2 and section 7.

### 14. Downstream obligations

- `TASK-260728-2jaw7h` lands the shared `toolchain` object; this decision adds
  no wire field and MUST NOT be read as adding one.
- `TASK-260728-251p01` integrates `rust-v1` and `rust-repository-v1` into
  manifest schema 8, descriptor schema 2, receipt schemas 3 and 4, marker schema
  4, claim schema 4 and the generated corpus, moving both identifiers from the
  reserved namespace to the admitted wire driver set in the same change that
  mints the schema admitting them, and extends decision 0008's boundary gate
  member-set table accordingly.
- `TASK-260728-q283m8` and `TASK-260728-13ioo0` implement the pair in Curator,
  `TASK-260728-2yxdo7` and `TASK-260728-gjxj1v` in csk, against the reference
  document rather than against this decision's prose.
- `TASK-260728-2bu2q6` qualifies the candidate and emits only evidence-backed
  driver and platform claims; `(macos, arm64)` is the only pair this contract
  supplies evidence for.
- `TASK-260728-16kefa` verifies cross-manager and native interop;
  `TASK-260728-26e3n2` runs the Linux qualification of section 12.
- `STORY-260728-327soo` receives the compile-time filesystem read surface of
  section 13 as a new input; it is not covered by any of the six existing
  deferred guarantees.

### 15. Enforcement

The boundary gate of decision 0008 section 11 needs no new mechanism for this
contract: `rust-v1` and `rust-repository-v1` are already two of the six reserved
identifiers it holds out of every surface file, and the exact member-set table
already covers the shapes they will occupy. Three additions belong to
`TASK-260728-251p01` at admission time and are named here so the gate is
extended rather than weakened:

1. the two identifiers move from the reserved namespace to the admitted set in
   the same change that mints receipt schemas 3 and 4, and the gate's driver
   `const` sets move with them;
2. the policy object of section 11 joins the exact member-set table, closed and
   `additionalProperties: false`, with `execution_policy` pinned to the
   `manager-worker-v2` `const`; and
3. the `curator-rust-toolchain-v1` identity joins the table as an object schema
   with its closed member set, its `roles` token set closed by `const`, and no
   member naming a filesystem path.

## Stable failure classes

These are architecture-level semantic classes and MUST remain distinguishable
from each other, from a cache hit, from an audit success, from source
unavailability and from a generic fallback. The reference document maps each to
its exact trigger and stage.

- `build_package_code_execution_forbidden`, the shared class of section 8;
- `build_descriptor_driver_unsupported` and
  `build_descriptor_schema_unsupported`, unchanged from decision 0008;
- `build_artifact_class_unsupported`, for a platform that cannot produce a
  single self-contained executable; and
- the twelve `build_toolchain_*` codes of decision 0007, unchanged, with
  `build_toolchain_platform_unsupported` carrying both the unsupported host pair
  and the absent-standard-library case of section 3, and
  `build_toolchain_metadata_mismatch` carrying the `rust-version` classifier of
  section 9.

## Rejected alternatives

- **Reject Cargo entirely and drive `rustc` directly, as decision 0004
  contemplated.** Rejected: for anything with a dependency it means
  re-implementing resolution and ordering inside the manager and issuing one
  `rustc` invocation per crate, which the `manager-worker-v2` session shape —
  exactly one compile command — does not admit. The narrower fix was taken
  instead: keep `cargo` as the single launcher and remove every surface through
  which a package can make it start something else.
- **Admit build scripts under a sandbox, a warning, or an allowlist of
  "well-known" crates.** Rejected: decision 0008 section 7 forbids answering a
  package-selected execution surface with a runtime allowance, and an allowlist
  of crate names is a trust store the protocol has no way to verify. The
  measured fact that `cargo metadata` reports `custom-build` targets without
  running them is what makes outright rejection implementable, so nothing is
  gained by weakening it.
- **Admit procedural macros because so much of the ecosystem depends on them.**
  Rejected for the same reason, and with the cost acknowledged plainly: a
  derive-heavy crate cannot be built by this driver. The honest consequence is a
  narrower admitted dependency set, not a weaker execution boundary.
- **Let the package's `.cargo/config.toml` configure the build, since it is the
  ecosystem's normal configuration surface.** Rejected: it was **measured** to
  execute a package-named script as a `rustc` wrapper, and its `[source]`,
  `[registries]`, `[patch]` and `[http]` tables have no environment counterpart
  and outrank the manager's own `$CARGO_HOME` config. It is refused outright,
  and the environment neutralisation is kept as a second layer rather than as
  the answer.
- **Neutralise the config file with `--config` command-line arguments instead of
  rejecting it.** Rejected: it would require the manager to enumerate every
  dangerous key forever, and a key added by a future cargo would be admitted by
  default. Rejecting the file is closed; enumerating overrides is open.
- **Resolve the linker as `cc` like every other Rust build does, and fingerprint
  the Apple toolchain that provides it.** Rejected on measurement: `cc` is a
  `PATH` lookup and `xcrun` is a second one, both outside any fingerprinted
  closure, and fingerprinting an Xcode toolchain to bring them inside would add
  gigabytes of tree and an executable graph the manager does not control. Using
  the `rust-lld` the distribution already ships removes both lookups and leaves
  the SDK as data.
- **Bind the platform SDK as a companion toolchain with its own identifier.**
  Rejected: decision 0007's identifier set is closed at five, an SDK is not a
  toolchain in that sense, and it contributes no process. Extending the
  algorithm this decision owns to fingerprint an ordered closure of roots keeps
  the closed set intact and keeps `toolchain_identities` at one element.
- **Fingerprint only the SDK subtree the linker actually reads.** Rejected: the
  linker is given `-syslibroot` and resolves within the SDK layout, so a subtree
  is not a closure, and choosing one would make the fingerprint a guess about
  linker behaviour that a toolchain update could invalidate silently. The
  measured 9.01 s cost is stated instead.
- **Memoise the toolchain fingerprint across operations to recover that cost.**
  Rejected: decision 0007 says fingerprinting proves the tree is stable across an
  operation and identical across operations, and a memo keyed on anything
  cheaper than the content proves neither.
- **Let `source_dir` name a package inside a Cargo workspace.** Rejected: the
  member's own `Cargo.toml` would be a nearer ancestor than the build root's,
  which contradicts decision 0008 section 4's nearest-ancestor rule, and reading
  the rule loosely for Rust would make the one structural guarantee about
  project metadata language-dependent. One package per build root is the cost.
- **Add a `bin`, `package`, `features` or `profile` member to the build command
  so a workspace could be addressed.** Rejected by decision 0008 section 4
  directly: the consuming command key is the sole naming authority and any
  further selector hands pipeline control back to untrusted data.
- **Run the graph command with `--filter-platform` so the rejection matrix
  matches what will actually be built.** Rejected: it makes the verdict
  host-dependent, so the same snapshot could be admitted on one machine and
  rejected on another, and the direction of the error is the wrong one — a
  surface excluded on this host is still in the snapshot. The union graph
  over-rejects deterministically, and that trade is stated in section 8 rather
  than hidden.
- **Reject Cargo features and `[profile]` tables along with everything else.**
  Rejected: features select which package source compiles, which is the same
  class of choice as a build constraint, and profiles cannot inject a flag on a
  release toolchain once `-Z` and `cargo-features` are rejected. Rejecting them
  would refuse ordinary packages while proving nothing.
- **Treat `rust-toolchain.toml` as forbidden in its entirety because it was
  measured to redirect the shim and to attempt a download.** Rejected: the same
  file was measured to be completely inert against the direct resolution
  decision 0007 mandates, so forbidding it would reject ordinary packages for a
  property that only exists on a resolution path Curator never takes. `path`
  stays `forbidden` because it names an origin; the rest is compared and
  discarded.
- **Keep `rust-version` as a single `compared` rule, as decision 0007's expected
  table has it.** Rejected on measurement: the field has three host-independent
  acceptance layers with distinct diagnostics, and a single comparison rule has
  no outcome for a value the ecosystem itself cannot represent. Those values
  would pass Stage B and fail later as a cargo error, after cache lookup — the
  exact deferral a pre-compiler cross-check exists to remove.
- **Measure Rust's representability layer with `cargo build`'s exit status.**
  Rejected for the reason decision 0007 already established for Go: that status
  is representability conjoined with the host gate and the edition floor, and it
  scores every representable future release as unrepresentable. The isolated
  measurement is `cargo metadata --no-deps`, which was **measured** to accept a
  `rust-version` above the host with exit 0 and therefore structurally cannot be
  applying the host gate.
- **Recognise cargo's diagnostics by their leading `error:` line alone.**
  Rejected: three distinct grammar rejections share the lead `error: expected a
  version like "1.32"` and none of them names the value under test, so a
  lead-only classifier answers for a family rather than an outcome — the same
  defect decision 0007 retired on the Go side. The caret block's source line is
  required as part of the recognised form because it is the only part that names
  the value.
- **Skip the `bin` target name check because it comes from the manager's own
  graph phase.** Rejected: it comes from package bytes by way of the graph
  phase, and it reaches an argument vector. It is validated against a closed
  grammar before it is used, in the same spirit as every other manager-derived
  value.
- **Claim macOS amd64 because the standard library ships in the same root.**
  Rejected: shipping a standard library is not evidence that the pipeline runs,
  and decision 0008 section 9 requires immutable native evidence for the exact
  tuple. The pair is a qualification obligation with a stated acceptance test.
- **Ship a Windows path that resolves `link.exe` through a Visual Studio
  activation script.** Rejected: it reintroduces exactly the `PATH`-resolved
  executable that section 2 removes on macOS, and it would make the Windows
  process graph depend on host state the manager neither selects nor
  fingerprints. Windows waits for `lld-link` or self-contained evidence.
- **Byte-scan package source for `include!` and reject on a match.** Rejected:
  it is unsound in both directions — `include ! ( )` and `cfg_attr` forms escape
  it, and the substring rejects ordinary comments — and an unsound check
  presented as a guarantee is worse than a stated exposure. Section 13 states
  it.

## Compatibility impact

This decision changes no bytes. It adds no schema, no vector, no generated case
and no release metadata; it does not alter the rc.5 conformance manifest digest
or any pin. Manifest schemas 1 through 7, `skill-build.json` schema 1, receipt
schemas 1 and 2, marker schemas 1 through 3, claim schemas 1 through 3,
`Skillfile.dev.json` schema 2, `manager-worker-v1`, `capability-evidence-v1`,
`rc5-native-control-inventory-v1`, `curator-go-toolchain-v1`,
`curator-build-source-v1` and every rc.4 and rc.5 conformance byte keep their
exact contents and meanings, and every `go-v1` and `go-repository-v1` identity is
unchanged.

`rust-v1` and `rust-repository-v1` remain reserved. Until `TASK-260728-251p01`
moves them, every schema including manifest schema 8 as first minted MUST reject
both, and a manager MUST treat each as an unknown driver.

`curator-rust-toolchain-v1` is a new algorithm identifier that no existing
artifact names, so introducing it moves no frozen byte. It does not reuse,
extend or alias `curator-go-toolchain-v1`.

The `rust-version` disposition changes from decision 0007's expected `compared`
row to `classified`. That row was explicitly reserved for confirmation or
correction by this decision, no schema or vector currently depends on it, and
the change narrows what Curator accepts silently rather than widening it.

This decision reserves `0009` as its decision number. Swift
(`TASK-260728-1yhuqi`) and Kotlin (`TASK-260728-168smo`) are unstarted at the
time of writing; if either lands `0009` first, this record renumbers rather than
contests, exactly as `TASK-260728-2spy93` did when `0007` was claimed.

## Security impact

The central claim of section 2 is narrow, checkable and measured: on the probed
host, neither the graph phase nor the compile phase resolves a single executable
through `PATH`, and the control run without the linker and SDK pins resolves two.
The whole process closure is three executables inside one fingerprinted
distribution, which is a stronger property than the contract requires and
strictly stronger than resolving a platform compiler driver would give.

The exposure this driver adds relative to `go-v1` is real and is not minimised.
A Rust build front end is a larger untrusted parser and code generator than the
Go one, and Rust's mainstream build path executes package-selected code by
design. Section 8 removes that path rather than containing it, and every removal
is decided before the compile phase from data the graph phase was **measured**
not to execute. What remains is the ordinary compiler-input exposure the
portable policy already admits — denial of service through resource consumption,
and compiler vulnerabilities reached by adversarial source — bounded by the
parent-enforced deadline, output and artifact limits and by whichever
native-control inventory entries the host provides, and by nothing stronger. The
six deferred hardened guarantees are not claimed, named as controls, or implied.

Two surfaces are admitted with bounds rather than rejected, and section 13 states
both: compile-time file inclusion, which is a read the portable policy does not
contain and which no existing deferred guarantee covers, and foreign function
declarations against base-installation libraries, which the artifact class
already admits. Neither is presented as closed.

Fingerprinting remains honest about what it proves. It proves that both roots
are stable across an operation and identical across operations. It does not
prove that the Rust distribution is genuinely upstream's, and it does not prove
that the SDK is genuinely Apple's; verifying either remains the operator's
responsibility at configuration time, and this contract performs no signature
verification of a toolchain or an SDK. A `tree_sha256` in a receipt must not be
read as provenance.

Refusing auto-install, prerelease hosts, package-selected paths, channels,
mirrors, registries and version managers is inherited unchanged from decision
0007 and is not reopened here. The one thing this decision adds to that surface
is the link-support root, and it is resolved through the same two declaration
channels, with the same forbidden origins and the same diagnostics, precisely so
that admitting it introduces no new way for a package to influence what runs.
