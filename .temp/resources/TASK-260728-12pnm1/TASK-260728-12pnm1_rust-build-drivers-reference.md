# Rust build drivers: `rust-v1` and `rust-repository-v1`

Implementation-ready reference for decision
[0009](../decisions/0009-rust-driver-pair.md), under the boundary of decision
[0008](../decisions/0008-additional-language-driver-boundary.md) and the shared
toolchain contract of decision
[0007](../decisions/0007-compiled-build-toolchain-preflight.md).

Both identifiers are **reserved**, not admitted. Until `TASK-260728-251p01`
moves them into the admitted wire driver set in the same change that mints
receipt schemas 3 and 4, every schema MUST reject them and a manager MUST treat
each as an unknown driver.

Every measured value in this document was produced on macOS 26.5 arm64 with
Rust 1.91.0 (`rustc 1.91.0 (f8297e351 2025-10-28)`, `cargo 1.91.0 (ea2d97820
2025-10-10)`) against a directly resolved toolchain root, never a `rustup` shim.
Nothing here is a platform claim.

## 1. `toolchain-registry-v1`: the `rust` entry

| Field | Value |
|---|---|
| `toolchain_id` | `rust` |
| `primary_relpath` | `macos`: `bin/cargo` |
| `probe` | `macos`: three vectors, section 1.1 |
| `normalization` | `rust.rustc.vV.release`, section 1.2 |
| `fingerprint_algorithm` | `curator-rust-toolchain-v1`, section 2 |
| `baseline` | `{"kind":"at_least","min":"1.91.0"}` |
| `compatibility` | families `{(1, 91)}`; family granularity `(major, minor)` |
| `platforms` | `(macos, arm64)` |
| `companions` | none |
| `link_support_roles` | `macos`: `["platform-sdk"]` |
| `metadata_sources` | `Cargo.toml` `package.rust-version`; `rust-toolchain.toml`; `rust-toolchain` |

Driver mapping: `rust-v1` and `rust-repository-v1` both map to primary `rust`
with no companion. A driver with no registry entry is unsupported; there is no
generic mapping and no fallback.

`link_support_roles` is a per-operating-system ordered closed list of roles the
Rust distribution does not itself provide. It is a manager-owned registry field
in the same sense as `platforms`; a package MUST NOT name a role, and no
manifest, descriptor, repository byte, environment value or `PATH` entry may
supply or influence one. An entry that declares a role for an operating system
outside its `platforms` set is unreachable data and fails the same release gate
that checks guidance reachability.

`platforms` holds exactly one pair. `(macos, amd64)`, Windows and Linux are
qualification obligations with the acceptance tests of section 12, not claims.
On a host whose pair is not in `platforms`, Stage A fails
`build_toolchain_platform_unsupported` with `check` = `host_pair`, before
resolution, on registry data alone.

### 1.1 Probe vectors

Run once per operation from the manager parent, from a manager-owned empty
working directory, under the operation-private environment of section 5, with
`<root>` the resolved Rust distribution root:

```text
<root>/bin/rustc -vV
<root>/bin/rustc --print host-tuple
<root>/bin/cargo --version
```

Measured output, verbatim:

```text
$ rustc -vV
rustc 1.91.0 (f8297e351 2025-10-28)
binary: rustc
commit-hash: f8297e351a40c1439a467bbbb6879088047f50b3
commit-date: 2025-10-28
host: aarch64-apple-darwin
release: 1.91.0
LLVM version: 21.1.2

$ rustc --print host-tuple
aarch64-apple-darwin

$ cargo --version
cargo 1.91.0 (ea2d97820 2025-10-10)
```

`rustc --version` is deliberately not used: it embeds the commit hash on the same
line, so `-vV` is the narrower surface. The `host:` line of `-vV` carries the
same value as `--print host-tuple` on this host and is deliberately not used as
the target probe, for the same reason.

### 1.2 Normalization — `rust.rustc.vV.release`

Read `rustc -vV` **stdout**. Require at most 4096 bytes, valid UTF-8, no NUL.
Split on LF. Exactly one line MUST match, anchored over the whole line:

```text
^release: (0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$
```

Groups 1 through 3 are the canonical triple. A non-empty group 4 is a
prerelease: `build_toolchain_prerelease_unsupported`. Zero matches or more than
one match is `build_toolchain_version_undetermined`, never a default.

`cargo --version` **stdout** MUST match, anchored over the whole line:

```text
^cargo (0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)? \([0-9a-f]+ [0-9]{4}-[0-9]{2}-[0-9]{2}\)$
```

Its triple MUST equal `rustc`'s and its group 4 MUST be empty. A mismatch, or a
prerelease marker on either, is `build_toolchain_version_undetermined` — a root
whose launcher and compiler disagree is not a usable identity, and the driver
starts both.

Architecture mapping for `platforms`: the trailing components of the host tuple
map `aarch64` to `arm64` and `x86_64` to `amd64`; the operating-system component
maps `apple-darwin` to `macos`, `pc-windows-*` to `windows`, `unknown-linux-*`
to `linux`. Any other tuple is `build_toolchain_platform_unsupported` with
`check` = `reported_target`.

### 1.3 Native target admission

Representability and admission are different questions and the ecosystem answers
only the first cheaply.

**Measured**: `rustc --print target-libdir --target x86_64-unknown-linux-gnu`
exits **0** and prints
`<root>/lib/rustlib/x86_64-unknown-linux-gnu/lib` on a host where that directory
does not exist; `cargo build --target x86_64-unknown-linux-gnu` then fails only
after emitting `Compiling probe-positive v0.1.0`, that is after the compile
phase has begun. An unknown triple is refused earlier: `--print target-libdir
--target not-a-real-target` exits **1** with `error: error loading target
specification: could not find specification for target "not-a-real-target"`.

Admission is therefore a manager-side check inside the fingerprinted tree, run
in Stage A after normalization:

1. `<root>/lib/rustlib/<native-tuple>/lib` MUST be an existing directory inside
   the fingerprinted Rust distribution root;
2. it MUST contain at least one regular file whose name matches
   `^libstd-[0-9a-f]+\.rlib$`.

A failure is `build_toolchain_platform_unsupported` with `check` =
`reported_target`, before source acquisition, before cache lookup and before any
compiler child. `rustc --print target-list` and `--print target-libdir` MUST NOT
be used as the admission test.

## 2. `curator-rust-toolchain-v1`

### 2.1 Resolution

The Rust distribution root and every `link_support_roles` root are resolved
through exactly the two declaration channels of decision 0007 section 3 —
`operator_config` then `bundled` — and through nothing else. `PATH`, the
inherited environment (including `RUSTUP_HOME`, `CARGO_HOME`, `RUSTC`,
`RUSTUP_TOOLCHAIN`, `DEVELOPER_DIR`, `SDKROOT`, `TOOLCHAINS`), a runtime root,
project `.agents/bin`, any shim, a manifest or descriptor value, `xcrun`,
`xcode-select` and any version-manager wrapper (`rustup`, `asdf`, `mise`) are
forbidden origins with the diagnostics decision 0007 section 5 fixes. An
operator MAY configure the concrete root that `rustup` produced; the manager
resolves that root directly and never through the shim.

Inside the resolved Rust distribution root the manager requires three regular
executables, and starts no other program below the worker:

| Relpath | Role |
|---|---|
| `bin/cargo` | trusted launcher, the `primary_relpath` |
| `bin/rustc` | compiler |
| `lib/rustlib/<native-tuple>/bin/rust-lld` | linker |

Each MUST be a regular executable inside the tree being fingerprinted, never a
wrapper and never outside it. A missing or non-regular member is
`build_toolchain_untrusted` with `substep` = `shape`.

### 2.2 Identity

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
    {"role": "rust-distribution", "tree_sha256": "sha256:<hex>"},
    {"role": "platform-sdk",      "tree_sha256": "sha256:<hex>"}
  ],
  "closure_sha256": "sha256:<hex>"
}
```

`roles` is the closed token set `{rust-distribution, platform-sdk}`. No root
path appears anywhere in the identity: toolchain location is not portable
identity. The `roots` array is ordered `rust-distribution` first, then each
`link_support_roles` entry in its registry order.

Per-root `tree_sha256` uses the walk, ordering, record framing, link rules and
`kind` alphabet of `curator-go-toolchain-v1` — walk without following links, the
root itself is not a record, relative components must be valid Unicode scalar
values, `/`-joined paths encoded as UTF-8 without case folding or normalization,
duplicate encoded paths and special files rejected, symlinks relative and
non-dangling and resolving within the root with independent tree records for
their referents, hard links as independent regular-file records, path bytes
sorted in unsigned bytewise order — with SHA-256 initialized by the exact ASCII
`curator-rust-toolchain-v1/<role>` followed by `0x00` and each entry appended as:

```text
kind || uint64be(path_byte_length) || path_utf8 ||
uint64be(payload_byte_length) || payload
```

The `rust-distribution` root appends two final records with empty paths: `V`
carrying the normalized `rustc -vV` stdout and `C` carrying the normalized
`cargo --version` stdout, each normalized by requiring at most 4096 bytes and
exactly one terminal LF optionally preceded by CR, removing that terminator, and
rejecting any other CR, LF, NUL, empty or invalid UTF-8 content. A
`platform-sdk` root appends no version record.

`closure_sha256` initializes SHA-256 with the exact ASCII
`curator-rust-toolchain-v1/closure` followed by `0x00` and appends, for each
element of `roots` in order:

```text
ASCII("R") || uint64be(role_byte_length) || role_ascii ||
uint64be(32) || raw_tree_digest_bytes
```

so a one-root closure and a two-root closure can never collide even if their
tree digests coincide. Prefix every emitted digest with `sha256:`.

Permissions, ownership, timestamps, ACLs and extended attributes are not hash
inputs, but the three executables of section 2.1 MUST be regular and executable
at use time. Every root MUST remain unchanged through the last child exit, and
every identity MUST be re-verified after the last child exits and before
publication; a change rejects the operation before publication.

**Measured cost**, same host: the Rust distribution root is 657 MB across 167
regular files and hashes in 1.73 s wall clock; the macOS SDK root is 261 MB
across 32,345 regular files and 7,448 symlinks and hashes in 9.01 s. The cost is
per operation and per root and MUST NOT be memoised across operations.

## 3. Source layout

Identical for both drivers. For `rust-v1` the root is the command's declared
local build root inside the consuming skill snapshot; for
`rust-repository-v1` it is the descriptor target's `build_root` inside the
locked external repository snapshot.

| # | Requirement | Diagnostic on failure |
|---|---|---|
| L1 | `source_dir` equals `build_root` | `build_rust_source_dir_invalid` |
| L2 | `Cargo.toml` exists directly in `build_root` and is the nearest ancestor `Cargo.toml` of `source_dir` | `build_rust_manifest_missing` |
| L3 | `Cargo.lock` exists directly in `build_root` | `build_rust_lockfile_missing` |
| L4 | `vendor` exists directly in `build_root` and is a directory; it MAY be empty | `build_rust_vendor_missing` |
| L5 | no `.cargo` directory exists anywhere in the `build_root` subtree | `build_rust_package_config_forbidden` |
| L6 | no file in the `build_root` subtree has a native extension from the closed list of section 6.3 | `build_rust_native_input_forbidden` |
| L7 | the `build_root` `Cargo.toml` contains no `cargo-features`, `[patch]`, `[replace]` or `[workspace]` table | `build_rust_manifest_key_forbidden` |

**Measured** that an empty `vendor` directory resolves and reports metadata with
exit 0, so L4 is an authoring requirement rather than a dependency requirement.

Snapshot validation, link-free directory rules, build-root disjointness and the
`build_roots` model are unchanged from manifest schema 6 and are not restated.
For the external mode the whole repository snapshot remains the validation,
identity and audit subject; only the selected build root is compiler-visible;
and input MUST NOT come from the consuming skill, another external repository, a
sibling or parent directory outside the selected build root, a host Cargo
registry cache, a host `CARGO_HOME`, or the network.

## 4. The two argument vectors

The working directory is the canonical `source_dir` for both. Neither the parent
nor the worker MAY alter, extend, reorder or repeat them.

```text
cargo metadata --format-version 1 --locked --offline --color never --quiet
cargo build --locked --offline --color never --quiet --release --target <native-tuple> --bin <bin-target-name>
```

`<native-tuple>` is the Stage A resolved host tuple. `<bin-target-name>` is the
single `bin` target name from the graph phase, validated against
`^[A-Za-z0-9_][A-Za-z0-9_-]{0,63}$` **before** it is placed in an argument
vector; a name outside that grammar is `build_rust_bin_target_invalid` and no
compile phase starts.

`--locked` is load-bearing. **Measured**: without it, `cargo metadata` writes
`Cargo.lock` into the source tree — a write to the frozen snapshot. With it and
no lock file present, `cargo metadata` exits **101** with
`error: the lock file <path> needs to be updated but --locked was passed to
prevent this` and writes nothing.

The produced file is:

```text
<CARGO_TARGET_DIR>/<native-tuple>/release/<bin-target-name>          (Unix)
<CARGO_TARGET_DIR>\<native-tuple>\release\<bin-target-name>.exe      (Windows)
```

`CARGO_TARGET_DIR` is an operation-private manager staging root. The manager
hashes the file there, sets manager-defined executable permissions, publishes it
as `bin/<command>` or `bin/<command>.exe` derived solely from the consuming
manifest command key, and MUST NOT execute it for validation, version discovery,
smoke testing, post-processing, receipt generation, rollback or any other
reason. Every other file in the target directory is a compiler by-product,
stays in staging, is discarded with it, and never enters cache identity, the
receipt, the marker, the shim relationship or publication.

## 5. Operation-private environment

The environment starts empty except for indispensable operating-system process
variables, and is identical for the probe vectors, the graph phase and the
compile phase.

| Variable | Value |
|---|---|
| `PATH` | a manager-owned empty directory |
| `HOME`, `TMPDIR`, `XDG_CONFIG_HOME`, `XDG_CACHE_HOME` | operation-private roots |
| `APPDATA`, `LOCALAPPDATA`, `USERPROFILE`, `TEMP`, `TMP` | operation-private roots, Windows only |
| `LC_ALL`, `LANG` | `C` |
| `CARGO_HOME` | operation-private root holding only the manager-written config of section 5.1 |
| `CARGO_TARGET_DIR` | operation-private staging root |
| `RUSTC` | absolute `<root>/bin/rustc` |
| `RUSTDOC` | absolute `<root>/bin/rustdoc` |
| `RUSTC_WRAPPER`, `RUSTC_WORKSPACE_WRAPPER` | set and empty |
| `CARGO_ENCODED_RUSTFLAGS` | exactly `-Clinker=<root>/lib/rustlib/<native-tuple>/bin/rust-lld` `0x1F` `-Clinker-flavor=<flavor>` |
| `CARGO_ENCODED_RUSTDOCFLAGS` | set and empty |
| `CARGO_INCREMENTAL` | `0` |
| `CARGO_NET_OFFLINE` | `true` |
| `CARGO_NET_RETRY` | `0` |
| `CARGO_NET_GIT_FETCH_WITH_CLI` | `false` |
| `CARGO_TERM_COLOR` | `never` |
| `SDKROOT` | the resolved `platform-sdk` root, macOS only |

`<flavor>` is `ld64.lld` on macOS. Windows and Linux flavours are qualification
obligations, section 12.

Every other Rust, Cargo, `rustup`, compiler, linker, SDK and executable-search
variable MUST be absent, and none may be inherited. In particular
`RUSTC_BOOTSTRAP` MUST be absent, because it makes a release toolchain accept
`-Z` flags; and `RUSTFLAGS`, `RUSTDOCFLAGS`, `CARGO_BUILD_*` other than the
target directory, `CARGO_TARGET_*_LINKER`, `CARGO_TARGET_*_RUSTFLAGS`,
`CARGO_REGISTRIES_*`, `CARGO_REGISTRY_*`, `CARGO_HTTP_*`, `CARGO_UNSTABLE_*`,
`RUSTUP_*`, `RUSTC_LOG`, `DEVELOPER_DIR`, `MACOSX_DEPLOYMENT_TARGET`,
`LD_LIBRARY_PATH`, `DYLD_*`, `LIBRARY_PATH`, `CPATH`, `CC`, `CXX` and `AR` MUST
be absent.

Three of these are answers to measured behaviour rather than hygiene:

- **`RUSTC` absolute.** Cargo otherwise resolves `rustc` by name from `PATH`;
  under a minimal `PATH` `TASK-260729-rhjxtx` measured
  `could not execute process 'rustc -vV' (never executed)`, and under a
  populated one the second node of the process graph would be chosen by `PATH`
  order.
- **`SDKROOT` set.** **Measured**: without it, `rustc` runs
  `xcrun --sdk macosx --show-sdk-path` resolved from `PATH`; with it, no `xcrun`
  lookup occurs.
- **`RUSTC_WRAPPER` and `RUSTC_WORKSPACE_WRAPPER` set and empty.** **Measured**:
  a package `.cargo/config.toml` carrying `[build] rustc-wrapper` executed the
  named script three times during `cargo build`; setting `RUSTC_WRAPPER` to the
  empty string in the manager environment neutralised it, and a config
  `[env] RUSTC_WRAPPER = { value = "...", force = true }` did not override the
  neutralisation.

Environment neutralisation is a second layer, never the answer. `[source]`,
`[registries]`, `[patch]` and `[http]` config tables have no environment
counterpart and a package config file outranks the manager's own `$CARGO_HOME`
config, so L5 rejects the file outright.

**Measured**: a `.cargo/config.toml` in an *ancestor* directory above the build
root is discovered and applied. The manager MUST therefore guarantee that no
`.cargo` directory exists in any ancestor of the compile working directory, up
to the filesystem root. That is a manager staging obligation; when it cannot be
met the operation fails `build_execution_control_unavailable` before the worker
starts.

### 5.1 The manager-written `$CARGO_HOME/config.toml`

Written by the manager before the graph phase, with exactly these four tables
and nothing else:

```toml
[source.crates-io]
replace-with = "curator-vendor"

[source.curator-vendor]
directory = "<absolute path of <build_root>/vendor>"

[net]
offline = true

[term]
quiet = true
color = "never"
```

The only package-derived component is the driver-fixed relative path `vendor`
under the build root. No package or descriptor byte reaches this file, and the
manager MUST NOT write any other key into it.

## 6. Pre-compile rejection matrix

Computed from the validated snapshot and the graph phase, before the compile
phase. Total by construction: every row has one verdict, and the list of rows is
closed. Rows are evaluated in the order given; the first failure is the reported
diagnostic.

### 6.1 Graph-phase properties

**Measured**: `cargo metadata --format-version 1 --offline` over a build root
whose path dependency declares `build = "build.rs"` with a `build.rs` that
writes a marker file, and whose second path dependency declares
`[lib] proc-macro = true` with a macro that writes a second marker file,
produced **neither** marker and exited 0. The same run reported
`kind: ["custom-build"]` for the build-script target, `kind: ["proc-macro"]` and
`crate_types: ["proc-macro"]` for the macro crate, `links: "probelib"` for the
package declaring a native library, and a `source` value per package.

The graph command is run **without** `--filter-platform`, so `packages[]` is the
union over every platform and every feature rather than this host's resolution.
The matrix is therefore host-independent, and it over-approximates: a dependency
that would only be built on another platform, or only under a feature this build
does not enable, still rejects the command. That is the accepted trade.

### 6.2 Rows decided by the graph phase

Let `R` be the build root and `V` be `<R>/vendor`.

| # | Rejected when | Diagnostic |
|---|---|---|
| G1 | `workspace_members` has other than exactly one element, or that element is not the `R` package, or `resolve.root` is null or differs | `build_rust_workspace_forbidden` |
| G2 | any package has a target whose `kind` contains `custom-build` | `build_rust_build_script_forbidden` |
| G3 | any package has a target whose `kind` or `crate_types` contains `proc-macro` | `build_rust_proc_macro_forbidden` |
| G4 | any package has a non-null `links` | `build_rust_native_link_declaration_forbidden` |
| G5 | any package `source` is neither null nor exactly `registry+https://github.com/rust-lang/crates.io-index` | `build_rust_dependency_source_forbidden` |
| G6 | a package with null `source` has a `manifest_path` outside `R`, or any target `src_path` outside `R` | `build_rust_input_outside_build_root` |
| G7 | a package with the crates.io `source` has a `manifest_path` outside `V` | `build_rust_input_outside_build_root` |
| G8 | any target has a `crate_types` member outside `{bin, lib, rlib}` | `build_rust_crate_type_forbidden` |
| G9 | the `R` package has other than exactly one target whose `kind` is exactly `["bin"]` | `build_rust_bin_target_ambiguous` |
| G10 | that target's `name` is outside `^[A-Za-z0-9_][A-Za-z0-9_-]{0,63}$` | `build_rust_bin_target_invalid` |

G2, G3, G4, G5 and G8 are the `build_package_code_execution_forbidden` semantic
class; the rest are driver-specific structural rejections. G5 covers git
dependencies, alternate registries, `local-registry` and `sparse` sources and any
future source kind, because it is an exact-match allowlist of two values rather
than a deny-list. **Measured**: a git dependency under `--offline` fails inside
the graph phase with `error: failed to get 'anyhow' as a dependency of package
'gd v0.1.0 (...)'` / `Caused by: failed to load source for dependency 'anyhow'`,
before any compile phase, so G5 has a fail-closed backstop in cargo itself.

### 6.3 Rows decided by snapshot bytes

L5 and L7 of section 3, plus the closed native-input extension list of L6, case
folded for comparison:

```text
.o .obj .a .lib .so .dylib .dll .tbd .rlib .rmeta .bc .ll .pdb .exp .res
.s .S .asm .c .cc .cpp .cxx .h .hh .hpp .m .mm .def .rc
```

and any path component ending in `.framework` or `.dSYM`. A match is
`build_rust_native_input_forbidden`. This is defence in depth over an already
closed path — without build scripts there is no admitted way to compile or link
such a file — and it is the direct analogue of `go-v1`'s `SysoFiles`, `CFiles`
and `SFiles` rejection.

### 6.4 Rows decided by the fixed environment and argument vectors

| Surface | Closed by |
|---|---|
| network, registry index, git fetch, crate download | `--offline`, `--locked`, `CARGO_NET_OFFLINE`, the manager-written config, and an empty `PATH` |
| package-selected linker, linker flavour, link argument, library search path | `CARGO_ENCODED_RUSTFLAGS` fixed by the manager, no package config file, no build script |
| package-selected `rustc`, `rustdoc`, wrapper | absolute `RUSTC` and `RUSTDOC`, empty `RUSTC_WRAPPER` and `RUSTC_WORKSPACE_WRAPPER` |
| `-Z` flags, unstable manifest keys, nightly features | release-channel toolchain, `RUSTC_BOOTSTRAP` absent, `cargo-features` rejected by L7 |
| cross-compilation | `--target` fixed to the resolved native tuple |
| incremental state | `CARGO_INCREMENTAL=0` |
| package-selected toolchain path, root, channel, mirror, installer, version manager | decision 0007 resolution; `rust-toolchain*` classified rather than honoured, section 7 |

### 6.5 Admitted surfaces

Cargo **features** are admitted: they select which package source compiles,
which is the same class of choice as a Go build constraint, and the build always
uses the root package's default feature resolution because the command object
cannot express one.

`[profile]` tables are admitted: on a release toolchain per-profile `rustflags`
are gated behind `-Z profile-rustflags`, and both `-Z` flags and the
`cargo-features` opt-in that would unlock them are rejected, so a profile selects
codegen tuning and cannot inject a flag or start a process. Link-time
optimisation stays inside `rustc` and its bundled LLVM.

Two further surfaces are admitted with bounds and are stated as residual
exposures rather than closed; see section 11.

## 7. Stage B — metadata dispositions

Decision 0007's disposition framework, precedence rule, file-shape gate and
channel classification are fixed there and are not reopened. Files are evaluated
in Unicode-scalar lexical order of relative source path, so `Cargo.toml`
precedes `rust-toolchain` precedes `rust-toolchain.toml`; within each file
`forbidden` classes precede `compared` classes.

| Source | Field | Disposition |
|---|---|---|
| `Cargo.toml` | `package.rust-version` | `classified`, section 7.2 |
| `Cargo.toml` | `workspace.package.rust-version` | unreachable: a workspace is rejected by G1 |
| `rust-toolchain.toml` | `toolchain.path` | `forbidden` |
| `rust-toolchain.toml` | `toolchain.channel` | `compared`, section 7.3 |
| `rust-toolchain.toml` | `toolchain.components`, `toolchain.targets`, `toolchain.profile` | `ignored` |
| `rust-toolchain` | the bare channel string | `compared`, section 7.3 |

**Measured**: a build root carrying a `rust-toolchain.toml` with *both*
`path = "/nonexistent"` and `channel = "nightly"` built successfully through a
directly resolved `cargo`, with zero `PATH` resolutions. The same file redirects
the `rustup` shim — `TASK-260729-rhjxtx` measured `error: invalid toolchain: the
path '/nonexistent/trusted/root' has no bin/ directory` for `path`, and
`info: syncing channel updates for 'nightly-aarch64-apple-darwin'` followed by a
download attempt for `channel = "nightly"`. The file is a live selector through
the shim and completely inert against direct resolution, which is exactly what
admits `channel` as `compared` and keeps `path` `forbidden`.

### 7.1 File-shape gate

For each `metadata_sources` file present in the validated tree, the gate covers
file **syntax** only: a file the ecosystem's own grammar rejects, including a
key the ecosystem permits at most once appearing more than once, is
`build_toolchain_metadata_mismatch` with `assertion` = `unclassifiable` and a
`source_ref` naming the file or the field path. It asserts nothing about the
semantics of fields the entry does not read.

For `Cargo.toml` the gate is "the TOML document does not parse". A `rust-version`
whose TOML *type* is wrong is deliberately **not** a gate case: the document
parses and the field extracts as a TOML value, so it is classifier class 1 of
section 7.2, for the same reason decision 0007 refused to route a shape-valid
but unrepresentable Go value to its gate.

### 7.2 Classifier — `Cargo.toml` `package.rust-version`

Rust has **three host-independent acceptance layers** plus one host gate, where
Go has two plus one. All four are measured, one value per fixture, with
`cargo metadata --no-deps --format-version 1 --offline`:

| Value | Exit | Layer | First diagnostic line |
|---|---|---|---|
| `"1.85"` | 0 | — | accepted, `rust_version` `1.85` |
| `"1.85.0"` | 0 | — | accepted, `rust_version` `1.85.0` |
| `1.85` | 101 | document | `error: invalid type: floating point '1.85', expected a semver or workspace` |
| `"1.85.0-beta"` | 101 | grammar | `error: unexpected prerelease field, expected a version like "1.32"` |
| `"not-a-version"` | 101 | grammar | `error: unexpected prerelease field, expected a version like "1.32"` |
| `"1.85.0+build"` | 101 | grammar | `error: unexpected build field, expected a version like "1.32"` |
| `"1.85.0.1"` | 101 | grammar | `error: expected a version like "1.32"` |
| `"stable"` | 101 | grammar | `error: expected a version like "1.32"` |
| `""` | 101 | grammar | `error: expected a version like "1.32"` |
| `"1"` @ edition 2015 | 0 | — | accepted, `rust_version` `1` |
| `"1"` @ edition 2018 | 101 | edition floor | `rust-version 1 is older than first version (1.31.0) required by the specified edition (2018)` |
| `"1.31"` @ edition 2018 | 0 | — | accepted |
| `"1.31"` @ edition 2021 | 101 | edition floor | `rust-version 1.31 is older than first version (1.56.0) required by the specified edition (2021)` |
| `"1.56"` @ edition 2024 | 101 | edition floor | `rust-version 1.56 is older than first version (1.85.0) required by the specified edition (2024)` |
| `"1.85"` @ edition 2024 | 0 | — | accepted |

Measured edition floors: 2015 admits `"1"`; 2018 requires 1.31.0; 2021 requires
1.56.0; 2024 requires 1.85.0. The floor depends on the manifest's own `edition`
field and on no host input.

The three layers are pairwise independent, and the independence is measured
rather than argued: the document layer accepts `"not-a-version"` which the
grammar layer rejects; the grammar layer accepts `"1"` which the edition floor
rejects at edition 2018 and above; and the edition floor would accept `1.85`
which the document layer rejects as a float. No layer contains another.

**The host gate is excluded from the layer measurement.** **Measured**:
`cargo metadata --no-deps` reports `rust_version` `1.99` on a 1.91.0 host with
exit 0, so it structurally cannot be applying the host gate; and `cargo build
--offline` with the same manifest exits 101 with
`error: rustc 1.91.0 is not supported by the following package:` and
`probe-future@0.1.0 requires rustc 1.99`, with compilation not started.

Canonicalization: a value of `MAJOR` canonicalizes to `(MAJOR, 0, 0)`,
`MAJOR.MINOR` to `(MAJOR, MINOR, 0)`, `MAJOR.MINOR.PATCH` to itself.

The ordered exhaustive classifier. There is no `forbidden` class, because the
field's value space is a version and nothing else. The catch-all is mandatory
and last.

| # | Class | Disposition | Outcome |
|---|---|---|---|
| 1 | the TOML value is not a string, or is the table form `{ workspace = true }` | — | `build_toolchain_metadata_mismatch`, `assertion` `unclassifiable` |
| 2 | a string the grammar cannot represent: a prerelease field, a build field, more than three dot-separated components, an empty string, or any non-numeric component | — | `build_toolchain_metadata_mismatch`, `assertion` `unclassifiable` |
| 3 | a grammar-representable string strictly below the floor of the manifest's `edition` | — | `build_toolchain_metadata_mismatch`, `assertion` `unclassifiable` |
| 4 | a grammar-representable string at or above the edition floor whose canonical triple is **above** the resolved toolchain triple | `compared` | `build_toolchain_metadata_mismatch`, `assertion` the derived canonical `at_least` |
| 5 | a grammar-representable string at or above the edition floor whose canonical triple is at or below the resolved toolchain triple | `compared` | permitted, and never honoured |
| 6 | the field is absent | — | contributes no assertion |
| 7 | anything else | — | `build_toolchain_metadata_mismatch`, `assertion` `unclassifiable` |

Classes 1 through 3 are host-independent, so their vectors need no Rust
toolchain on the runner. Classes 4 and 5 take the resolved version as fixture
input, exactly as decision 0007 fixes for every Stage B classification case.

Class 1 includes the workspace-inheritance table because a workspace is rejected
by G1 and the value would otherwise have no resolvable meaning. It is
deliberately not routed to G1: a Stage B classification is a statement about a
value, and G1 is a statement about the graph.

### 7.3 Classifier — `rust-toolchain.toml` `toolchain.channel` and the legacy `rust-toolchain` file

Decision 0007's channel table applies unchanged and is not restated: a
canonicalizable version literal becomes an `at_least` assertion; `stable` is
permitted and never honoured; `beta`, `nightly` and dated channels are a
mismatch because they assert a prerelease host that is never resolved; anything
else is `build_toolchain_metadata_mismatch`, never a default and never a
selector.

The legacy one-line `rust-toolchain` file carries a bare channel string and uses
the identical classifier. Its file-shape gate is "the file is not exactly one
line of printable non-empty content after trimming a single trailing newline".

`toolchain.path` is `forbidden` and is evaluated before every `compared` class,
so a file carrying both `path` and `nightly` is deterministically
`build_toolchain_package_influence_forbidden` with the `toolchain-root`
`origin_class`.

### 7.4 Recognised command outcomes are a closed set

The corroborating command classifier is closed. A recognised outcome is one
whole diagnostic line, matched **exactly** against a form predicted before the
command ran from the value under test and the probe's own fixed constants. An
outcome outside the set is **unknown**, yields no verdict and fails the probe. A
lead with an unconstrained tail, and a substring found anywhere in the output,
are families rather than outcomes and MUST NOT be recognised.

Every grammar and document rejection renders a caret block whose third line is
exactly `<N> | rust-version = <literal>`, where `<N>` is the fixture's fixed
line number and `<literal>` is the value under test as written. The recognised
form is therefore the **pair** of one predicted first line and one predicted
source line, because three distinct grammar rejections share the first line
`error: expected a version like "1.32"` and none of the first lines names the
value.

| Class | First line, exact | Second element, exact |
|---|---|---|
| document | `error: invalid type: <toml-type-phrase>, expected a semver or workspace` | `<N> \| rust-version = <literal>` |
| grammar/prerelease | `error: unexpected prerelease field, expected a version like "1.32"` | `<N> \| rust-version = <literal>` |
| grammar/build | `error: unexpected build field, expected a version like "1.32"` | `<N> \| rust-version = <literal>` |
| grammar/other | `error: expected a version like "1.32"` | `<N> \| rust-version = <literal>` |
| edition floor | `error: failed to parse manifest at \`<manifest-path>\`` | `  rust-version <literal-body> is older than first version (<floor>) required by the specified edition (<edition>)` |
| host gate | `error: rustc <host-version> is not supported by the following package:` | `  <package>@<version> requires rustc <literal-body>` |
| accepted | exit 0 | `packages[0].rust_version` equals `<literal-body>` |

`"1.32"`, the edition floors `1.31.0`, `1.56.0` and `1.85.0`, and the manifest
line number are probe fixed constants. If a later cargo release changes any of
them the probe turns red rather than quietly changing what it measures, which is
the correct direction for a check whose purpose is to notice upstream moving.

The command forms are the narrowest that exercise the layer under test:
`cargo metadata --no-deps --format-version 1 --offline` for the three layers,
because it was measured not to apply the host gate; and `cargo build --offline`
as the corroborating measurement, whose outcome is classified into `accepted`,
`rejected-document`, `rejected-grammar`, `rejected-edition`, `host-gate` and
`unknown`, never into pass and fail.

### 7.5 Closure is measured, not asserted

The probe carries a closure section that feeds the classifier outcomes
deliberately outside the recognised set and requires each to yield no verdict,
reporting for every fabrication which of the two laundering directions it
belongs to:

- **direction A, acceptance laundering**: real unrelated command failures — a
  missing manifest, an unreadable working directory, an unknown subcommand, a
  `--locked` lock-file failure — must not be scored as upstream acceptance;
- **direction B, rejection laundering**: every measured outcome cross-fed under
  a different value, and every measured diagnostic extended the way a later
  release would extend it, must not be scored as a rejection verdict for the
  value actually under test.

The extended-diagnostic checks are constructed and are disclosed as such: an
outcome upstream has not yet written cannot be measured on any host. Taking text
upstream did emit and changing it the way a later release would is the honest
form of that check.

### 7.6 Controls required to fail

Each is runnable from the probe binary and each MUST fail; a control that stops
failing is a regression.

| # | Control | What it guards |
|---|---|---|
| C1 | an open classifier with a fall-through verdict | closure in both directions |
| C2 | lead-only recognition, dropping the caret source line | three grammar classes collapsing into one |
| C3 | exit status as the semantic measurement (`cargo build` exit 0 means representable) | folding the host gate and the edition floor into the grammar layer |
| C4 | `cargo build` as the isolated representability command | the same folding, arrived at by choosing a wider command |
| C5 | the edition floor folded into the grammar classifier | one host-independent layer swallowing another |
| C6 | substring matching anywhere in combined output | recognising a family rather than an outcome |

C3 and C4 are not redundant: C3 changes how the outcome is read, C4 changes
which command produces it, and either alone leaves the other's defect
reachable.

## 8. Identity, cache, receipt, marker, claim

The canonical build input binds, in addition to the members decision 0008
section 8 requires of every new driver:

- the complete `curator-rust-toolchain-v1` identity of section 2, including both
  per-root digests and `closure_sha256`, as the single element of
  `toolchain_identities`;
- the resolved native tuple;
- the validated `bin` target name; and
- this closed policy object, whose `execution_policy` is the `const`
  `manager-worker-v2`:

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

`network: "none"` denotes the fixed offline Cargo configuration, `--offline`,
`--locked`, the manager-written `$CARGO_HOME` config and the empty `PATH`. It is
not a claim of kernel-enforced network denial; that guarantee is
`total-network-denial`, deferred to `STORY-260728-327soo`.

The logical cache key is the SHA-256 of `CCJ-1` over the complete input, exactly
as for the Go drivers. Receipt schema 3 carries the local mode and schema 4 the
external mode, each a strict `oneOf` discriminated by the `driver` `const`.
Marker schema 4 records `driver`, `receipt_schema_version` and
`execution_policy` per build entry; a reader rejects a `rust-v1` entry claiming
`manager-worker-v1` rather than inferring the policy from the driver name.
Conformance claim schema 4 asserts each identifier with `execution_policy`
selected by the assertion's own `driver` `const`.

The effective toolchain requirement and the `compatibility` set are gates, not
build inputs.

`rust-repository-v1`'s input additionally carries the members decision 0008
section 5 and `protocol/core.md` section 9.2 already fix for an external
command — repository identifier, declared and effective source state,
substitution, external `curator-build-source-v1`, descriptor path and selected
target — plus `"source_kind":"locked-external-git-v1"`.

## 9. Diagnostics

Driver-specific codes, all beneath the `build_package_code_execution_forbidden`
semantic class where marked, all fired before the compile phase:

| Code | Stage | Trigger | Class |
|---|---|---|---|
| `build_rust_source_dir_invalid` | validation | L1 | structural |
| `build_rust_manifest_missing` | validation | L2 | structural |
| `build_rust_lockfile_missing` | validation | L3 | structural |
| `build_rust_vendor_missing` | validation | L4 | structural |
| `build_rust_package_config_forbidden` | validation | L5 | code execution |
| `build_rust_native_input_forbidden` | validation | L6 | code execution |
| `build_rust_manifest_key_forbidden` | validation | L7 | code execution |
| `build_rust_workspace_forbidden` | graph | G1 | structural |
| `build_rust_build_script_forbidden` | graph | G2 | code execution |
| `build_rust_proc_macro_forbidden` | graph | G3 | code execution |
| `build_rust_native_link_declaration_forbidden` | graph | G4 | code execution |
| `build_rust_dependency_source_forbidden` | graph | G5 | code execution |
| `build_rust_input_outside_build_root` | graph | G6, G7 | structural |
| `build_rust_crate_type_forbidden` | graph | G8 | code execution |
| `build_rust_bin_target_ambiguous` | graph | G9 | structural |
| `build_rust_bin_target_invalid` | graph | G10 | structural |

The twelve `build_toolchain_*` codes of decision 0007 apply unchanged.
`build_toolchain_platform_unsupported` carries `check` = `host_pair` for a pair
outside `platforms` and `check` = `reported_target` for a reported tuple that
does not map or whose standard library is absent by section 1.3.
`build_descriptor_driver_unsupported` and `build_descriptor_schema_unsupported`
apply to the external mode unchanged. `build_artifact_class_unsupported` applies
to a platform that cannot produce a single self-contained executable.

No diagnostic reproduces an unvalidated package byte.

## 10. Artifact

Exactly one bounded regular file, class `native-executable-v1`.

**Measured** on this host: the pipeline produces a `Mach-O 64-bit executable
arm64` whose only dynamic dependency reported by `otool -L` is
`/usr/lib/libSystem.B.dylib`, a base-installation library, and whose code
signature is `adhoc, linker-signed`. That signature is applied by `rust-lld` as
part of linking: it is produced by the driver's fixed argument vector, selects no
signing identity, credential, entitlement or notarization, and reaches no
network. It is compiler output, not a manager signing step. The driver performs
no manager post-build signing, timestamping or notarization, and a platform
policy requiring a locally signed binary must wait for the separately versioned
and reviewed signer profile.

The driver does not require bit-reproducible artifacts.

## 11. Residual exposures

Two surfaces are admitted with bounds rather than rejected. Both are stated so
that no reader, receipt, marker or claim can imply otherwise.

**Compile-time file inclusion.** `include!`, `include_str!` and `include_bytes!`
resolve a path relative to the including file and can name a path outside the
build root. They are reads, not code execution, so decision 0008 section 7 does
not require their rejection; the portable policy does not contain the compiler's
filesystem access, and none of the six deferred hardened guarantees covers
compile-time reads. No sound deterministic pre-compile rejection exists: the
macro name is a token rather than a byte pattern, so `include ! ( "x" )` and
`cfg_attr` forms defeat a byte scan, while a scan for the substring `include`
rejects ordinary comments and identifiers. Recorded as a new input to
`STORY-260728-327soo`.

**Foreign function declarations against base-installation libraries.** Package
source may declare `extern` blocks and `#[link]` attributes. The package cannot
supply a library to link — native files are rejected by L6, `links` by G4, build
scripts by G2, and no admitted path adds a library search path — so `#[link]`
can only name a library the pinned link environment already resolves, which is a
base-installation library. Decision 0008 section 3 already requires the artifact
to depend on exactly those. This is not a claim that the produced program is
safe; it remains untrusted package output the manager never executes.

## 12. Platform matrix and qualification

| Platform | Status |
|---|---|
| macOS arm64 | complete pipeline measured on one host; enters a claim only through `TASK-260728-2bu2q6` |
| macOS amd64 | qualification obligation |
| Windows | implementation contract only; no platform claim |
| Linux | excluded until `TASK-260728-1skseh` |

### 12.1 The acceptance test

Identical on every candidate platform, and it is what `platforms` membership
means:

1. `PATH` is set to a directory containing logging shims for at least `cc`,
   `c++`, `clang`, `clang++`, `ld`, `ld64`, `xcrun`, `ar`, `ranlib`, `dsymutil`,
   `strip`, `sh`, `bash`, `env`, `lld`, `ld.lld`, `ld64.lld`, `gcc`,
   `install_name_tool`, `codesign` and, on Windows, `link`, `cl`, `lib`,
   `vswhere`, `cmd`; each records its name and argv and exits 127.
2. The graph phase and the compile phase both run to completion with **zero**
   recorded entries.
3. A control run with the linker pin and the SDK pin removed records at least
   one entry. Without the control, the zero proves nothing.
4. The produced executable runs and its dynamic dependencies are all
   base-installation libraries of the declared platform baseline.
5. `<root>/lib/rustlib/<tuple>/lib` contains `libstd-*.rlib` inside the
   fingerprinted tree.

**Measured on macOS arm64**: steps 1, 2 and 4 pass with twenty shims and zero
entries for both phases; step 3 recorded `xcrun --sdk macosx --show-sdk-path`
and `cc <full link line>` and the build failed with
`error: linking with 'cc' failed: exit status: 127`; step 5 passes.

### 12.2 Windows implementation contract

Two candidate paths, neither claimed:

1. `x86_64-pc-windows-msvc` with
   `-C linker=<root>/lib/rustlib/<tuple>/bin/rust-lld` and
   `-C linker-flavor=lld-link`, with the MSVC and Windows SDK import libraries
   bound as one or more data-only `platform-sdk` link-support roots. `lld-link`
   is present alongside the macOS flavours in `lib/rustlib/<tuple>/bin/gcc-ld/`
   on the measured root, which makes this the path to test first.
2. `x86_64-pc-windows-gnu` with the target's bundled self-contained linking
   artifacts, which if sufficient would need no link-support root at all.

Until section 12.1 passes with a firing control, `platforms` excludes Windows
and both drivers fail `build_toolchain_platform_unsupported` there. An
implementation MUST NOT ship a Windows path that resolves `link.exe`, `cl.exe`,
`gcc`, `ld`, `vswhere` or a Visual Studio activation script from `PATH`, the
registry or an environment variable, and MUST NOT answer the gap with a
host-resolved tool or a downgraded control.

### 12.3 Linux qualification rules

Linux enters `platforms` only when section 12.1 passes on the qualifying host
and, in addition: the produced ELF executable's dynamic dependencies are all
base-installation libraries of the declared distribution baseline, and the
`platform-sdk` role is either absent or resolved from a declaration channel.
`x86_64-unknown-linux-gnu` is expected to need a `platform-sdk` root holding the
C runtime startup objects and `libc` stubs; `x86_64-unknown-linux-musl` may need
none. Which of those holds is the qualification question, not a claim.

## 13. Conformance vector inventory

Positive:

1. local vendored build root, one package, one `bin`, no dependencies, builds
   and publishes `bin/<command>`;
2. the same with one vendored crates.io dependency;
3. the same with an empty `vendor` directory;
4. external repository target with `build_root` `.`;
5. external repository target with a nested `build_root`;
6. `rust-version` at the resolved toolchain triple, permitted;
7. `rust-version` below the resolved toolchain triple, permitted;
8. `rust-toolchain.toml` `channel = "stable"`, permitted and never honoured;
9. `rust-toolchain.toml` `channel` a version literal at or below the resolved
   triple, permitted and never honoured;
10. a `[profile.release]` table setting `opt-level` and `lto`, admitted;
11. a package with an enabled non-default feature reached through the default
    feature set, admitted;
12. cache hit on an unchanged input, with both preflight stages still run;
13. cache miss on a changed `closure_sha256` with an unchanged source identity.

Negative, rejection matrix:

14. build script in the root package; 15. build script in a path dependency;
16. build script in a vendored dependency; 17. proc-macro crate type in a path
dependency; 18. proc-macro crate type in a vendored dependency; 19. proc-macro
reachable only through a non-default feature, rejected by the union graph;
20. non-null `links`; 21. git dependency; 22. alternate-registry dependency;
23. two workspace members; 24. virtual manifest; 25. zero `bin` targets;
26. two `bin` targets; 27. `cdylib` crate type; 28. `staticlib` crate type;
29. path dependency outside the build root; 30. vendored package outside
`vendor`; 31. `.cargo/config.toml` at the build root; 32. `.cargo/config.toml`
in a subdirectory; 33. `cargo-features` key; 34. `[patch]` table;
35. `[replace]` table; 36. `[workspace]` table; 37. a `.a` file in the tree;
38. a `.c` file in the tree; 39. a `.dylib` file in the tree; 40. missing
`Cargo.lock`; 41. missing `vendor`; 42. `source_dir` below rather than equal to
`build_root`; 43. `Cargo.toml` absent from the build root; 44. an intervening
`Cargo.toml` between build root and `source_dir`; 45. a `bin` target name
outside the closed grammar.

Negative, Stage B classifier — all host-independent except where noted:

46. `rust-version` as a TOML float; 47. as a TOML integer; 48. as the workspace
inheritance table; 49. with a prerelease field; 50. with a build field;
51. with four components; 52. `"stable"`; 53. the empty string;
54. `"1"` at edition 2018; 55. `"1.31"` at edition 2021; 56. `"1.56"` at
edition 2024; 57. `"1"` at edition 2015, permitted; 58. `rust-version` above the
resolved triple, resolved version as fixture input; 59. `Cargo.toml` that does
not parse as TOML, file-shape gate; 60. `rust-toolchain.toml` `path`;
61. `rust-toolchain.toml` `channel = "nightly"`; 62. a dated channel;
63. an unknown channel token; 64. `rust-toolchain.toml` carrying both `path` and
`nightly`, disposition precedence; 65. legacy `rust-toolchain` with a bare
channel; 66. legacy `rust-toolchain` with more than one line, file-shape gate;
67. both `rust-toolchain` and `rust-toolchain.toml` present, lexical ordering.

Negative, toolchain and platform:

68. host pair outside `platforms`, before resolution; 69. reported tuple that
does not map; 70. reported tuple whose `libstd-*.rlib` is absent;
71. prerelease `rustc` release line; 72. `rustc` and `cargo` triples disagree;
73. `rustc -vV` with no `release:` line; 74. resolved version outside
`compatibility`; 75. resolved version outside the effective requirement at
Stage A; 76. descriptor requirement narrowing past the resolved version at
Stage B; 77. `rust-v1` against a schema-1 descriptor; 78. an unsupported
descriptor schema version; 79. command and descriptor drivers disagree;
80. a missing `platform-sdk` declaration on macOS.

Negative, boundary and identity:

81. every frozen schema rejects `rust-v1`; 82. every frozen schema rejects
`rust-repository-v1`; 83. a marker-v4 entry pairing `rust-v1` with
`manager-worker-v1`; 84. a claim asserting `rust-v1` with `manager-worker-v1`;
85. a build input whose policy object carries an extra member; 86. a
`curator-rust-toolchain-v1` identity carrying a root path; 87. a one-root and a
two-root closure with identical tree digests, required not to collide.
