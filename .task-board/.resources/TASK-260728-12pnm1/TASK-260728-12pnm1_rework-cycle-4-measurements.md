# TASK-260728-12pnm1 — rework cycle 4 measurements

Every measurement below was produced on macOS 26.5 arm64 with
`rustc 1.91.0 (f8297e351 2025-10-28)` and `cargo 1.91.0 (ea2d97820 2025-10-10)`,
against a directly resolved toolchain root, never a `rustup` shim. Nothing here
is a platform claim. Raw output:
`TASK-260728-12pnm1_cycle4-raw-measurements.log`.

Unless stated otherwise, each run used the contract's exact two argument
vectors, an `env -i` environment, absolute `RUSTC`/`RUSTDOC`, the pinned
`rust-lld` with `-Clinker-flavor=ld64.lld`, a pinned `SDKROOT`, the
manager-written `$CARGO_HOME/config.toml`, private `HOME`/`TMPDIR`/XDG roots,
and a `PATH` containing only 30 logging shims that record their name and argv
and exit 127. The probe's own replays use the 25-name acceptance-test shim list
of reference section 13.1; the standalone sweep uses a 30-name superset that
adds `objdump`, `dwp`, `llvm-dwp`, `llvm-strip`, `xcode-select`, `swiftc`,
`dtrace` and `rust-lld` — the last so that even the pinned linker would be
caught if it were ever resolved by name rather than by absolute path. It never
was.

---

## M1. Which stable `[profile]` keys can start a process

Eighty-four build roots, one profile shape each. The raw log's `M1` section
lists all 84 row by row, plus the three vendored fixtures of M6 — 87 rows in
total. The complete stable key set was swept: `opt-level`, `debug`, `split-debuginfo`, `strip`, `debug-assertions`,
`overflow-checks`, `lto`, `panic`, `incremental`, `codegen-units`, `rpath`,
`inherits`, `trim-paths`, plus `[profile.<p>.package.<spec>]`,
`[profile.<p>.build-override]`, custom profiles, and the non-selected `dev`,
`test` and `bench` profiles.

**Exactly one key reaches an executable outside the closure.**

| Shape | `dsymutil` | Note |
|---|---|---|
| no profile table | no | baseline |
| `opt-level` ∈ {0,1,2,3,"s","z"} | no | 6 runs |
| `debug` ∈ {false,true,0,1,2,"none","line-directives-only","line-tables-only","limited","full"} | no | 10 runs, default `split-debuginfo` |
| `strip` ∈ {false,true,"none","debuginfo","symbols"} | no | 10 runs, with and without `debug = 2` |
| `lto` ∈ {false,true,"off","thin","fat"} | no | 5 runs |
| `panic`, `incremental`, `codegen-units`, `rpath`, `debug-assertions`, `overflow-checks` | no | 9 runs |
| `split-debuginfo = "off"` / `"unpacked"`, any `debug` | no | 4 runs |
| **`split-debuginfo = "packed"` with any non-zero `debug`** | **yes** | 7 runs; also with `strip = "symbols"` |
| `split-debuginfo = "packed"` with `debug` ∈ {false,0,"none"} | no | no debug info to package |
| **`[profile.release.package.<root package>] split-debuginfo = "packed"`** | **yes** | a package override reaches the root |
| `[profile.release.package."*"]` | no | measured: `"*"` covers dependencies, not the root |
| `[profile.release.build-override]` | no | build scripts are already rejected |
| `[profile.dev]`, `[profile.test]`, `[profile.bench]`, a custom profile | no | not selected by `--release` |
| `trim-paths` (6 spellings), `rustflags`, `codegen-backend` | n/a | exit **101**: require an unstable `cargo-features` opt-in |
| `split-debuginfo` as an integer or boolean | n/a | exit **101**: `invalid type … expected a string` |
| an unknown profile key, `totally-unknown-key = 42` | no | **exit 0, no diagnostic** — why the gate is a positive allowlist |

In every `dsymutil` row the shim really ran, returned **127**, and
`cargo build` still exited **0** with only
`warning: processing debug info with 'dsymutil' failed: exit status: 127`.

## M2. The dangerous case is not `"packed"` — it is *any unrecognized value*

A logging `RUSTC` wrapper captured the flags cargo forwards:

| Profile value | cargo forwards | `dsymutil` |
|---|---|---|
| key absent | `-Csplit-debuginfo=unpacked` | no |
| `"unpacked"` | `-Csplit-debuginfo=unpacked` | no |
| `"off"` | `-Csplit-debuginfo=off` | no |
| `"packed"` | `-Csplit-debuginfo=packed` | **yes** |
| `"wat"` | **nothing** | **yes** |
| `""` | **nothing** | **yes** |
| `"PACKED"` | **nothing** | **yes** |

`rustc --print split-debuginfo --target aarch64-apple-darwin` prints `off`,
`packed`, `unpacked`. Running `rustc -g` **directly**, with no
`-Csplit-debuginfo` at all, under the same poisoned `PATH`, invoked `dsymutil`:
rustc's own default for this target is `packed`.

Two consequences fix the shape of the fix.

1. A **deny-list** naming `"packed"` is bypassed by any garbage string, silently
   — no warning, no error, exit 0.
2. The previous revision's pipeline was safe only because cargo happens to emit
   `unpacked` when the key is absent. That is a **toolchain** default reachable
   with no package byte selecting it, so a snapshot-byte rule alone cannot own
   the property.

Hence the two mechanisms: a positive allowlist in Stage P, and a manager-owned
`-Csplit-debuginfo` pin in `CARGO_ENCODED_RUSTFLAGS`.

## M3. The pin

With `-Csplit-debuginfo=off` appended to `CARGO_ENCODED_RUSTFLAGS`:

| Fixture | pinned | unpinned |
|---|---|---|
| `split-debuginfo = "packed"` | 0 resolutions, build 0, artifact runs | 1 resolution, `dsymutil` |
| `split-debuginfo = "wat"` | 0 resolutions, build 0, artifact runs | 1 resolution, `dsymutil` |
| `split-debuginfo = ""` | 0 resolutions, build 0, artifact runs | 1 resolution, `dsymutil` |
| `[profile.release.package.<root>] split-debuginfo = "packed"` | 0 resolutions, build 0, artifact runs | 1 resolution, `dsymutil` |
| no profile table | 0 resolutions, build 0, artifact runs | 0 resolutions |

Every pinned run printed the program's own stdout, so the pin does not merely
suppress the helper — the artifact is produced and runs.

## M4. The pin changes artifact bytes, deterministically

Same build root, same absolute path, target directory cleared between runs:

| `CARGO_ENCODED_RUSTFLAGS` tail | sha256 | size |
|---|---|---|
| — | `33c634e4…` | 407 088 |
| — (repeat) | `33c634e4…` | 407 088 |
| `-Csplit-debuginfo=unpacked` | `738d8268…` | 407 088 |
| `-Csplit-debuginfo=off` | `db383465…` | 407 088 |

Repeated builds under one pin are byte-identical; two pins are not, because
`CARGO_ENCODED_RUSTFLAGS` participates in cargo's fingerprint. The pin therefore
belongs to the driver policy object and to cache identity, which is why
`debuginfo_packaging` joins the policy object.

## M5. `strip` is effective and starts no process

Produced-binary sizes: `strip = false` 427 168, `"none"` 427 168,
`"debuginfo"` 407 088, `"symbols"` 336 656, baseline 407 088. The setting is
acted on, and no `strip` shim was ever invoked, so admitting the key is an
absence of process rather than an absence of behaviour.

## M6. A dependency's `[profile]` table is inert — the over-rejection is real and bounded

| Fixture | result |
|---|---|
| path dependency inside `R` carrying `[profile.release] debug = 2, split-debuginfo = "packed"` | 0 resolutions, exit 0, no warning |
| vendored `cfg-if` `Cargo.toml` edited, `.cargo-checksum.json` **not** repaired | `cargo metadata` exit **0**; `cargo build` exit 101, `the listed checksum … has changed` |
| the same edit with the checksum repaired | 0 resolutions, exit 0 |

Only the root manifest's profiles are honoured. Applying the gate to every
manifest under `R` is therefore an over-rejection — the same kind the `[patch]`
row already carries, and taken for the same reason: the rule stays closed and
needs no reasoning about which manifest cargo treats as the workspace root. The
middle row also shows the vendor checksum is **not** a pre-compile gate: it
fires only at `cargo build`, after the graph phase has already accepted.

## M7. The over-rejection is bounded by real-world frequency

Across the **506** published crate manifests in the host's registry cache, **37**
carry a `[profile*]` table. Between them they use **seven** keys:

| key | manifests | distinct values seen |
|---|---|---|
| `debug` | 34 | `2`, `true` |
| `lto` | 25 | `"fat"`, `"thin"`, `true` |
| `opt-level` | 17 | `1`, `2`, `3` |
| `codegen-units` | 15 | `1` |
| `panic` | 13 | `"abort"` |
| `incremental` | 3 | `false` |
| `inherits` | 1 | `"release"` |

`split-debuginfo` occurrences: **0**. All 37 pass the new gate.

## M8. Cargo target auto-discovery, and why the admitted set needs none of it

Every reported `src_path`, with existence checked:

| Fixture | reported targets | all exist |
|---|---|---|
| `src/main.rs` only | `bin discprobe → src/main.rs` | yes |
| plus `src/lib.rs` | `lib`, `bin` | yes |
| plus `src/bin/extra.rs` | two `bin` targets | yes |
| plus `src/bin/sub/main.rs` | two `bin` targets | yes |
| plus `examples/e.rs`, `tests/t.rs`, `benches/b.rs` | `example`, `test`, `bench` | yes |
| plus `examples/e/main.rs` | `example e → examples/e/main.rs` | yes |
| plus `build.rs` | `custom-build → build.rs` | yes |
| `[[bin]] name = <package name>` with no `path` | resolves to `src/main.rs` | yes |
| vendored `cfg-if` | `vendor/cfg-if/Cargo.toml`, `…/src/lib.rs`, `…/tests/xcrate.rs` | yes |
| **`[[bin]] path = "src/ghost.rs"`, file absent** | reported with that `src_path` | **no** — declared, so section 4.3 already decided it |
| **`[lib] path = "src/nolib.rs"`, file absent** | reported | **no** — same |
| `[[bin]] name = "ghostonly"` with no file | — | graph exits **101**, `can't find 'ghostonly' bin at …` |
| `[lib] name = "onlylib"` with no file | — | graph exits **101** |
| no target at all | — | graph exits **101**, `no targets specified in the manifest` |

So a reported path is always either an existing file under `R` — which the
link-free walk enumerates — or a declared path the grammar already decided.
There is no third case, and no phantom path is ever reported. That is what makes
`A = walked ∪ declared-absent` a total superset without reimplementing Cargo's
discovery rules.

A second measurement supports the same conclusion from the other side: `cargo
vendor` normalization writes `autolib = false`, `autobins = false`,
`autoexamples = false`, `autotests = false`, `autobenches = false` and explicit
`[lib] path` / `[[test]] path` keys into every vendored manifest, so a vendored
crate relies on no auto-discovery at all. Only the root manifest does.

---

## Probe results

| run | result |
|---|---|
| full, macOS arm64 | **exit 0, green** — 19/19 cases, P1 and P2 hold, 24 closure checks with 0 verdicts, **15 of 15** controls failing as required, **124** host-independent vectors with 0 divergences, **16** structural checks with 0 divergences |
| controls-only replay | exit 0, 15 of 15 failing as required |
| no resolvable toolchain (`env -i`, `HOME=/nonexistent`) | exit 0, 19 cases `not_run` with the reason recorded, the 124 vectors still run with 0 divergences, nothing installed |

New in this revision:

- structural `profile-tables-cannot-start-a-process` — the six-run matrix of M1
  and M3 inside the probe;
- structural `stage-p-admitted-set-covers-auto-discovered-targets` — the graph
  over a default-layout package plus a declared absent leaf, with every reported
  path required to be in `A`;
- controls **C13** (profiles admitted wholesale), **C14** (a `"packed"`
  deny-list) and **C15** (the superseded G11 input set), each required to fail
  and each additionally re-checking its replacement;
- vectors **F1–F27** (profile grammar) and **A1–A16** (admitted path set), all
  host-independent.
