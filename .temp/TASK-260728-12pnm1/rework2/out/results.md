# TASK-260728-12pnm1 — rework cycle 2 results

Handoff for a fresh independent review. Four review blockers from cycle 1 were
the whole scope; each is closed by a measurement, a contract change and an
executable check. Every green finding from cycle 1 is retained, accepted
decisions 0007 and 0008 are untouched, no frozen artifact moved, and macOS arm64
remains the only measured platform.

Host: macOS 26.5 (25F71) arm64, Rust 1.91.0 (`rustc 1.91.0 (f8297e351
2025-10-28)`, `cargo 1.91.0 (ea2d97820 2025-10-10)`), root resolved directly,
never through the `rustup` shim. Nothing here is a platform claim.

## Blocker 1 — `curator-rust-toolchain-v1` rejected its own normative input

**What was wrong.** Reference section 2.2 required the `V` record to carry the
normalized `rustc -vV` stdout while forbidding every CR and LF other than one
terminator. `rustc -vV` stdout is multiline, so no conforming implementation
could compute the shown identity from the admitted toolchain.

**Measured.** `rustc -vV` stdout on this host is 192 bytes over 7 lines with one
terminal LF, no CR and no NUL. `cargo --version` stdout is 36 bytes over one
line.

**What changed.** A closed multiline normalization, reference section 2.3: at
most 4096 bytes, valid UTF-8, no NUL; every CRLF folded to LF so a Windows and a
Unix probe of the same distribution agree; any remaining bare CR rejected;
exactly one terminal LF not preceded by another; that terminator removed; the
payload non-empty with no U+007F and no sub-U+0020 scalar other than an interior
LF. The `V` payload may carry interior LFs; the `C` payload may not. Every
failure is `build_toolchain_version_undetermined`.

The whole verbose stream is bound rather than one extracted line, which is
strictly more binding: the identity changes when the commit hash, commit date,
host line or LLVM version changes even though the release triple does not. The
identity's `rust_version` member stays the single `release:` line and is
documented as a display member, not the hashed payload.

**Reproducible worked example**, recorded in the reference and asserted by the
probe:

| | `V` | `C` |
|---|---|---|
| payload bytes | 191 | 35 |
| `sha256(payload)` | `7d8e0833…bd1b81b5` | `8d712854…cb546165` |
| framed record bytes | 208 | 52 |
| `sha256(framed record)` | `7fc35c11…2cd5f06b` | `d677e668…185c2f78` |

**Evidence.** Thirteen conformance vectors N1–N13 in the reference; probe vector
group `normalization`, 15 checks, all matching, all host-independent except N1;
five unit tests including one that asserts the old single-line rule *would*
reject the real stream.

## Blocker 2 — the graph phase was not the union it claimed

**What was wrong.** The graph vector was
`cargo metadata --format-version 1 --locked --offline --color never --quiet`,
with no `--all-features`, while reference section 6.1 claimed a union over every
feature and vector 19 required a feature-gated proc macro to be rejected by it.

**Measured**, fixture with an optional `proc-macro` path dependency behind the
non-default feature `extra`, a second optional dependency behind its implicit
feature, and a windows-only target dependency:

```
without --all-features   packages[] = ['feat', 'winonly']
with    --all-features   packages[] = ['feat', 'implicit', 'pm', 'winonly']
```

Under the flag `pm` carries `kind ["proc-macro"]` and
`crate_types ["proc-macro"]`, so G3 fires. Without it the package is not in the
graph at all — the contract had a hole, not merely a false sentence.

**What changed.** `--all-features` is in the normative graph vector, and the
semantic is now stated exactly rather than as "every feature": the resolution is
over every platform, every dependency kind, and every feature **of the root
package** plus what those transitively activate. **Measured** that it is *not*
the union over every package's features: with root → `mid` and `mid` carrying an
optional proc macro behind `mid`'s own feature `x`, that macro is absent from
`Cargo.lock` and from the `--all-features` graph unless a root feature names
`mid/x`.

**The subset proof**, four independent directions, each measured:

1. the compile phase activates the root's default features, a subset of its full
   set, and Cargo activation is additive;
2. without `--filter-platform` a `cfg(target_os = "windows")` dependency is in
   `packages[]` on a macOS host;
3. `cargo metadata` reports dev and build dependency kinds that `cargo build
   --bin` never builds, and a build dependency implies a build script, already
   rejected;
4. `--locked` does not start failing under the flag, because `Cargo.lock`
   already records the optional dependencies reachable from root features and
   `cargo vendor` vendors the whole lock — measured on a registry fixture where
   `itoa` is optional behind `extra`, and `cargo vendor` has no `--all-features`
   flag at all (`error: unexpected argument '--all-features' found`).

Two costs are stated rather than hidden: `--all-features` can fail resolution on
mutually exclusive features, and it can pull a build-script dependency into the
graph the default build would never touch. Both fail closed.

**Evidence.** Structural check `all-features-exposes-optional-proc-macro`
(match); control C7 running the vector without the flag, failing as required;
one unit test pinning the exact argv string.

## Blocker 3 — escape was decided after Cargo had already read outside the root

**What was wrong.** G6/G7 decided an outside-root path dependency from
`cargo metadata` output. That prevents compilation; it does not prevent the
read.

**Measured.** A build root declaring
`outside = { path = "../outside", package = "exfiltrated-outside-name" }` makes
the graph command exit 0 and report a package whose `manifest_path` is outside
the build root. With the outside manifest replaced by malformed TOML the same
command prints `error: unclosed table, expected ]` and
`--> ../outside/Cargo.toml:1:9` — Cargo opened, parsed and quoted bytes from a
file the contract calls not compiler-visible.

**Measured, and worse.** An ancestor `Cargo.toml` reaches Cargo with **no byte
inside the build root naming it**. With `parent/Cargo.toml` carrying
`[workspace] members = ["build_root"]` and
`[patch.crates-io] cfg-if = { path = "evil" }`, where `parent/evil` declares
`build = "build.rs"`, the graph reported `workspace_root` = `parent`,
`workspace_members` of length **one** whose single element **is** the build-root
package, `resolve.root` equal to it, and a `cfg-if` whose `manifest_path` is
`parent/evil/Cargo.toml` with kinds `[["lib"], ["custom-build"]]`. The previous
G1 row passes that shape exactly. This was found while fixing blocker 3 and is
a second, independent escape.

**What changed.** A new **Stage P**, decision section 8 and reference section 4:
a manager-owned closure over snapshot bytes with the manager's own TOML parser,
before any cargo process.

- 4.1 inventories every `Cargo.toml` under the build root, vendored manifests
  included, and opens nothing outside it.
- 4.2 rejects `cargo-features`, `[workspace]`, `package.workspace`, `[patch]`,
  `[replace]`, workspace inheritance, `package.links`, proc-macro and plugin
  libraries, crate types outside `{bin, lib, rlib}`, and `git`/`rev`/`branch`/
  `tag`/`registry` dependency keys — in every manifest, not only the root's.
  `package.build` is admitted only when it is exactly `false`, because
  **measured**, the `cargo vendor` normalization of `cfg-if 1.0.4` writes
  `build = false` explicitly and the blunt rule would reject ordinary vendored
  crates.
- 4.3 is a closed seven-rule relative-path grammar over every path-bearing key,
  decided **on the declared string before any filesystem call is made with it**,
  plus a non-following physical check.
- 4.4 and 4.5 close dependency origins and the `Cargo.lock` / `vendor`
  correspondence.
- 4.6 makes both ancestor origins staging obligations: no `.cargo` directory and
  no `Cargo.toml` in any ancestor, failing `build_execution_control_unavailable`.
- 4.7 demotes the graph rows to a consistency cross-check and adds two: G0
  requires `workspace_root` to equal the build root, and G11 requires every
  reported `manifest_path` and `src_path` to be in the Stage P admitted set,
  reported as `build_rust_graph_inconsistent` — a manager fault, never a package
  rejection.

Stage P is explicitly **not** containment: it bounds what Cargo resolves, not
what `rustc` reads through `include_str!`, which stays the stated residual
exposure.

**Evidence.** Probe vector group `prewalk`, 30 checks covering relative,
absolute, `..`, drive, UNC, backslash, empty-component, dot-component, trailing
dot and space, Windows reserved device names, control characters, home prefix,
over-length, symlink traversal, missing dependency versus missing target path,
and both ancestor obligations in both directions; structural checks
`graph-phase-reads-outside-build-root` and
`ancestor-workspace-manifest-escapes-old-rows`, both matching; control C8, which
fails as required by showing Cargo quoting the outside manifest; four unit
tests, including one asserting `sub/../other` is rejected even though it
resolves inside.

## Blocker 4 — the exact vendored pipeline and the configuration were unproven

**What was wrong.** The probe's base environment created an empty `CARGO_HOME`
and never wrote the normative config; its structural commands omitted `--locked`
and `--bin`; no case built a vendored dependency in either source mode. The
config contract was also not serialization-complete and falsely claimed no
package-derived byte reached the file.

**What changed, part one — the claim.** The purity claim is **withdrawn**.
`build_root` is selected by the consuming manifest's `build_roots` entry or by
the external descriptor, so the absolute directory value carries
manager-validated but package-derived components. Containment is now a closed
representability rule plus a serialization that cannot escape: the value is
emitted verbatim inside a TOML **literal** string, which has no escape
sequences, and a path containing `'`, a control character, U+007F, a NUL, a
Windows verbatim or UNC prefix, or any non-absolute path, is rejected before the
file is written, as `build_execution_control_unavailable` — a manager fault,
never a package diagnostic. The whole file is a fixed byte template: UTF-8, no
BOM, LF only, one terminal LF, 89 constant bytes before the value and 61 after.
After writing, the manager MUST re-read and re-parse and require exactly four
tables and a byte-equal directory value.

**Measured**, varying only that value:

| serialization | exit |
|---|---|
| literal string, ASCII absolute path | 0 |
| literal string, path containing `ö` | 0 |
| basic string, ASCII absolute path | 0 |
| basic string with an unescaped `\` — what a Windows path is | 101, `could not load Cargo configuration` |
| relative value `vendor` | 101, `failed to load source for dependency` |

**Measured** that a naive writer concatenating a path containing `"` into a
basic string emits a file that really does carry extra `[net]` and `[junk]`
tables with `offline = false` among them.

**What changed, part two — the evidence.** The probe now runs the exact
pipeline: a build root with a vendored crates.io-source dependency, the
manager-written config produced by the rule-bound serializer and verified by
write-back, an operation-private `CARGO_HOME` and `HOME`, `CARGO_TARGET_DIR` in
staging, a `PATH` of 20 logging shims, and the exact two argument vectors.
Result: graph exit 0, build exit 0, **0** PATH resolutions, artifact runs and
prints its expected output.

Both source-mode mappings are exercised over byte-identical build roots: a local
build root, and a nested `crates/tool` build root inside a larger external
snapshot that also carries `docs/README.md` and `OUTSIDE.txt`. The build-root
digests are equal, the graph JSON is identical after substituting the build root
and the staging root, and no file outside the build root appears in the external
graph.

Separately, on the host and outside the probe, the same measurement was made
with a **real** crates.io crate: `cfg-if 1.0.4` obtained by `cargo vendor
--offline`, in a local build root and in a Git repository cloned and checked out
at a fixed revision. Both produced graph exit 0, build exit 0, zero PATH
resolutions, a running `Mach-O 64-bit executable arm64` depending only on
`/usr/lib/libSystem.B.dylib`, identical build-root digests, identical graph JSON
after substitution, and a byte-identical executable
(`71991c27…`). Bit-reproducibility is **not** a contract term; the equality is
recorded as an observation. The probe constructs its vendored dependency rather
than downloading one, so that it needs no network and no operator registry
cache; that choice is stated in the probe source rather than implied.

**Measured** that operator Cargo state is load-bearing: a build root with no
`vendor` directory resolves with exit 0 when the operator's `HOME` and
`CARGO_HOME` are visible and fails with exit 101 and `error: no matching package
named 'itoa' found` under operation-private ones.

**Evidence.** Structural checks `exact-pipeline-vendored-isolated` and
`local-and-external-source-mapping-equivalence`, both matching; probe vector
group `config`, 23 checks including every rejection form, the write-back check
per accepted value, the template length, the template digest and the
table-header count; controls C9 (naive writer injects; rule-bound writer
refuses) and C10 (inherited operator state resolves what isolation refuses),
both failing as required; five unit tests.

## What the probe reports now

```
cases 19/19 matched, 0 divergences
alignment P1 no-widening = true, P2 no-narrowing = true
closure 24 checks, 0 yielded a verdict
controls 10 of 10 failing as required
host-independent vectors 68, 0 divergences
structural 13 checks, 0 divergences
green: true, exit 0
```

The six cycle-1 controls are unchanged and still fail as required. Controls C7
through C10 are new. Host-independent vectors run even when no toolchain
resolves: the absent-toolchain run reports 19 cases `not_run` with the reason,
68 vectors with 0 divergences, exit 0, nothing installed.

## Contract deltas a reviewer should adjudicate

1. Graph argv gains `--all-features`. A narrowing: the matrix sees strictly more
   packages.
2. `curator-rust-toolchain-v1` version records bind whole normalized streams
   under a closed multiline rule. The algorithm identifier is new and
   unreferenced, so no existing digest moves.
3. Stage P, three new diagnostics (`build_rust_manifest_unparsable`,
   `build_rust_vendor_incomplete`, `build_rust_graph_inconsistent`), two new
   graph rows (G0, G11) and a second ancestor staging obligation. Every one
   rejects something previously admitted or admitted late.
4. The policy object gains the `const` members `feature_audit:
   "root-all-features"` and `source_closure: "manager-prewalk-v1"`. Both
   identifiers are still reserved and receipt schemas 3 and 4 are still unminted,
   so this costs no frozen byte.
5. The reference's conformance inventory grows from 87 to 134 vectors, entirely
   from the four fixes.

## Platform position, unchanged

`platforms` holds `(macos, arm64)` only. macOS amd64, Windows and Linux remain
qualification obligations with a stated acceptance test. Two Windows obligations
were **added** and not claimed: that a drive-letter absolute path with `\`
separators is accepted inside the TOML literal string, and that rejecting
verbatim and UNC prefixes does not exclude a staging location the manager needs.
The acceptance test itself gained two requirements — the clean run must use a
real vendored dependency with the manager-written config and isolated Cargo
state, and the directory value must serialize and survive write-back under the
platform's native path spelling.

## Residual exposures, unchanged in substance

Compile-time file inclusion and foreign function declarations against
base-installation libraries remain admitted with bounds. One sentence was added:
Stage P does **not** close compile-time inclusion, because it bounds what Cargo
resolves and an `include_str!` argument is a token inside a source file `rustc`
opens directly. It stays an input to `STORY-260728-327soo` needing a seventh
deferred guarantee.

## Gates

Probe `gofmt`/`go vet`/`go build`/`go test` all exit 0. Probe runs exit 0, green
true. Curator `go build`, `go vet`, `go test` exit 0; `gofmt -l .` lists 0 files
outside `.temp/`. Two expected-red gates are fully attributed in the gate log:
`make check` exit 2 fails at its gofmt stage on 1525 files, every one under
another task's scratch tree and none under this task's, with `go vet` and
`go test` passing inside it; spec `validate.py` exit 1 fails only on links inside
two documents copied in from another task's in-flight tree, with a clean-tree
baseline at exit 0 and a scoped link check over this task's two authored
documents at 0 broken of 4. `golangci-lint` was NOT RUN — it is not installed on
this host.

No file was staged, committed, published or pinned.
