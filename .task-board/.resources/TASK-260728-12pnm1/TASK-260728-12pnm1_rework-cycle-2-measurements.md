# TASK-260728-12pnm1 rework cycle 2 — host measurements

Host: macOS 26.5 (build 25F71), arm64. Rust 1.91.0
(`rustc 1.91.0 (f8297e351 2025-10-28)`, `cargo 1.91.0 (ea2d97820 2025-10-10)`),
root `/Users/iv/.rustup/toolchains/1.91.0-aarch64-apple-darwin`, resolved
directly, never through the `rustup` shim. SDK
`/Applications/Xcode_26_5.app/.../MacOSX26.5.sdk`.

Nothing below is a platform claim. Every number is one host.

## R1 — the exact vendored pipeline runs clean under the exact argv (finding 4)

Build root: package `vend`, one `bin`, one real vendored crates.io dependency
`cfg-if 1.0.4` obtained with `cargo vendor --offline`. `CARGO_HOME` is an
operation-private directory holding only the manager-written four-table
`config.toml`; `HOME`, `TMPDIR`, `XDG_*`, `CARGO_TARGET_DIR` are
operation-private; `PATH` is a directory of 20 logging shims that record their
name and argv and exit 127.

```
cargo metadata --format-version 1 --locked --offline --color never --quiet --all-features   -> exit 0
cargo build --locked --offline --color never --quiet --release --target aarch64-apple-darwin --bin vend -> exit 0
poison log size                                                                             -> 0 bytes
file target/aarch64-apple-darwin/release/vend  -> Mach-O 64-bit executable arm64
otool -L                                       -> /usr/lib/libSystem.B.dylib only
./vend                                         -> runs, prints "unix"
```

## R2 — `--all-features` is load-bearing for the rejection matrix (finding 2)

Fixture `feat`: root package with `pm = { path = "pm", optional = true }` where
`pm` is `[lib] proc-macro = true`, `implicit = { path = "implicit", optional =
true }` with no feature naming it, `[features] extra = ["dep:pm"]`, and
`[target.'cfg(target_os = "windows")'.dependencies] winonly = { path = "winonly" }`.

```
cargo metadata --format-version 1 --locked --offline --quiet
  packages[] = ['feat', 'winonly']
cargo metadata --format-version 1 --locked --offline --quiet --all-features
  packages[] = ['feat', 'implicit', 'pm', 'winonly']
```

Under `--all-features` the proc-macro package is present with
`kind ["proc-macro"]` and `crate_types ["proc-macro"]`, so G3 fires. Without the
flag the package is not in the graph at all and conformance vector 19 cannot
hold. The windows-only dependency is present in **both**, because
`--filter-platform` is not passed.

Registry variant `optreg` (`itoa` optional behind `extra`):

```
Cargo.lock                                  contains cfg-if, itoa, optreg
cargo vendor --offline vendor               vendors cfg-if and itoa
cargo metadata ... (no flag)   packages[] = ['cfg-if', 'optreg']
cargo metadata ... --all-features packages[] = ['cfg-if', 'itoa', 'optreg']
cargo build ... --bin optreg                exit 0
```

So `--locked --all-features` does not fail on a normally vendored tree: the lock
file already records the optional dependency and `cargo vendor` (which has no
`--all-features` flag — measured `error: unexpected argument '--all-features'`)
vendors the whole lock.

Transitive shape, fixture `deep`: root -> `mid`, `mid` has
`leafpm = { path = "../leafpm", optional = true }` behind `mid`'s feature `x`.

```
root has no feature naming mid/x:
  Cargo.lock       = deep, mid                 (leafpm absent)
  --all-features   packages[] = ['deep', 'mid'] (leafpm absent)
root adds y = ["mid/x"]:
  Cargo.lock       = deep, leafpm, mid
  --all-features   packages[] = ['deep', 'leafpm', 'mid']
  no flag          packages[] = ['deep', 'mid']
```

`--all-features` activates all features of the **root** package and whatever
those transitively activate; it does not activate a dependency's own feature
that no root feature names. That is the exact semantic the contract must state,
and it is sufficient: the compile phase activates the root's default feature
set, which is a subset of the root's full feature set.

## R3 — the graph phase reads outside the build root (finding 3)

Fixture `esc`: build root manifest with
`outside = { path = "../outside", package = "exfiltrated-outside-name" }`.

```
cargo metadata --format-version 1 --locked --offline --quiet --all-features -> exit 0
  packages[] carries `exfiltrated-outside-name` with
  manifest_path = <...>/esc/outside/Cargo.toml
```

With the outside manifest replaced by malformed TOML the same command prints

```
error: unclosed table, expected `]`
 --> ../outside/Cargo.toml:1:9
```

so Cargo opened, parsed, and reported bytes from a file outside the admitted
root before any check on its output could run. G6 rejects afterwards.

## R4 — an ancestor `Cargo.toml` is read with no byte inside the root naming it

Fixture `anc`: `parent/Cargo.toml` = `[workspace] members = ["build_root"]`.

```
with the ancestor manifest present:
  workspace_root    = <...>/anc/parent
with it renamed away:
  workspace_root    = <...>/anc/parent/build_root
```

Fixture `ancpatch`: the same ancestor manifest additionally carries
`[patch.crates-io] cfg-if = { path = "evil" }` where `parent/evil` is an outside
path package declaring `build = "build.rs"`.

```
workspace_root     = <...>/ancpatch/parent
workspace_members  = [ ... #br@0.1.0 ]              length 1, the R package
resolve.root       = ... #br@0.1.0                  equals the R package
packages[]         includes cfg-if with
                   manifest_path = <...>/ancpatch/parent/evil/Cargo.toml
                   targets kinds  [['lib'], ['custom-build']]
```

G1 as previously written passes this shape: one workspace member, and it is the
R package. The redirect is caught only downstream by G2/G6, after Cargo read the
outside tree. `workspace_root` is the field that exposes it.

Fixture `ancp`: `[package] workspace = "../elsewhere"` in the build-root
manifest makes Cargo read the named outside manifest and fail with
`current package believes it's in a workspace when it's not`.

## R5 — `rustc -vV` normalization (finding 1)

```
rustc -vV stdout            192 bytes, 7 lines, one terminal LF, no CR, no NUL
normalized V payload        191 bytes
sha256(V payload)           7d8e08339e557ede5a9e565773c4cf17f83dea27f7d0c6591869f184bd1b81b5
framed V record             208 bytes, header hex 56000000000000000000000000000000bf
sha256(framed V record)     7fc35c11acae420849418f9c9f9b5681651beedc6d68f00bf1e4db022cd5f06b

cargo --version stdout      36 bytes, 1 line
normalized C payload        35 bytes
sha256(C payload)           8d712854de14f22840767bc824c5ac08098f35ddaa44437256e31b53cb546165
framed C record             52 bytes, header hex 4300000000000000000000000000000023
sha256(framed C record)     d677e668d65108419fd197d147f31f2dda30a1364233440f3b0138d5185c2f78
```

Payload of the V record, verbatim:

```
rustc 1.91.0 (f8297e351 2025-10-28)\nbinary: rustc\ncommit-hash: f8297e351a40c1439a467bbbb6879088047f50b3\ncommit-date: 2025-10-28\nhost: aarch64-apple-darwin\nrelease: 1.91.0\nLLVM version: 21.1.2
```

## R6 — Cargo config `directory` serialization (finding 4)

Same build root, only the `[source.curator-vendor] directory` serialization
varies. `cargo metadata` exit code:

| form written | exit |
|---|---|
| basic string, ASCII absolute path | 0 |
| basic string, path containing `ö` written literally | 0 |
| TOML **literal** string `'...'`, ASCII absolute path | 0 |
| TOML literal string, path containing `ö` | 0 |
| basic string containing an unescaped `\` | 101, `could not load Cargo configuration` |
| basic string with `\\` escapes on macOS | 101, source not found (the path really contains a backslash) |
| relative value `"vendor"` | 101, `failed to load source for dependency` |

A naive writer that concatenates a path containing a `"` into a basic string, or
a `'` into a literal string, emits a document whose **parsed structure differs
from the intended one**: the injected text produced real extra `[net]` and
`[junk]` tables in the written file.

## R7 — operator Cargo state is load-bearing isolation

Build root with **no** `vendor` directory and no manager-written config:

```
HOME and CARGO_HOME inherited from the operator  -> cargo metadata exit 0
HOME and CARGO_HOME operation-private            -> exit 101,
   error: no matching package named `itoa` found / location searched: crates.io index
```

## R8 — local and external source modes are byte-equivalent

The same build-root bytes placed (a) in a local snapshot and (b) at a nested
`build_root` inside a Git repository, cloned and checked out at a fixed rev:

```
sha256 over the sorted per-file digests of the build root   identical
graph JSON after substituting build root and staging root   identical
sha256 of the produced executable                           identical
                                                            (71991c27...)
poison-path entries in the external run                     0
```

Bit-identical artifacts are **not** required by the driver; the equality is
recorded as an observation, not as a contract term.

## R9 — vendored manifest shapes that constrain the pre-Cargo walk

`cargo vendor` normalizes each vendored `Cargo.toml`. Measured on `cfg-if 1.0.4`:

- it carries `build = false` explicitly, so a pre-Cargo rule that rejects the
  presence of `package.build` would reject ordinary vendored crates; the rule
  must reject only a non-`false` value;
- it carries `autolib`, `autobins`, `autoexamples`, `autotests`, `autobenches`;
- it carries `[lib] path = "src/lib.rs"` and `[[test]] path = "tests/xcrate.rs"`,
  both relative and inside the package directory;
- it carries its own nested `Cargo.lock`, which the manager must not treat as
  the build root's lock file;
- each vendor child holds `Cargo.toml` and `.cargo-checksum.json`;
- the number of vendor children equals the number of `Cargo.lock` `[[package]]`
  entries carrying the crates.io source (1 of 1, then 2 of 2).

In the graph the vendored test target appears as `kind ["test"]`,
`crate_types ["bin"]`, so it does not collide with the single-`bin` rule and it
does not violate the crate-type allowlist.
