# TASK-260728-12pnm1 — rework cycle 3 host measurements

Every figure below was produced on this host and is reproducible from the
attached probe and the raw transcript
`TASK-260728-12pnm1_cycle3-raw-measurements.log`.

- host: macOS 26.5 (25F71), arm64
- toolchain: `rustc 1.91.0 (f8297e351 2025-10-28)`, `cargo 1.91.0 (ea2d97820 2025-10-10)`
- resolved distribution root: `~/.rustup/toolchains/1.91.0-aarch64-apple-darwin`

macOS arm64 remains the only measured platform. Nothing here widens a platform
claim, and no Windows or Linux toolchain was present.

---

## Blocker 1 — the size boundary is decided after CRLF folding

### M3.1 The reviewer's constructed pair, replayed

The superseded order applied `len(stdout) <= 4096` to the **raw** stream and
folded CRLF afterwards. Replaying that order against the exact pair the reviewer
constructed:

| stream | raw bytes | superseded verdict | replacement verdict |
|---|---|---|---|
| 4093 `x`, LF, `A`, LF | 4096 | **admitted** | admitted |
| the same content with CRLF endings | 4098 | **rejected: over 4096 raw bytes** | admitted |

Folding the second stream produces the first byte for byte, so two encodings of
one distribution's `rustc -vV` received different verdicts. Under the
replacement both are admitted with payload 4095 bytes and the identical framed
digest `f168ab077a348ef9…`.

### M3.2 A realistic multiline shape is worse

A 410-line stream whose folded form is also exactly 4096 bytes expands to
**4506** raw bytes under CRLF — 410 bytes of expansion, not 2. The superseded
order rejected it while admitting the LF form; the replacement admits both with
the identical digest `f8bf84b67b4517c4…`.

### M3.3 The 4096/4097 boundary, in pairs

| folded stream | LF form | CRLF form | replacement verdict |
|---|---|---|---|
| exactly 4096 bytes, 2 lines | 4096 raw | 4098 raw | both admitted, identical payload and digest |
| exactly 4097 bytes, 2 lines | 4097 raw | 4099 raw | both rejected, `folded 4097 > 4096` |
| exactly 4096 bytes, 410 lines | 4096 raw | 4506 raw | both admitted, identical payload and digest |
| exactly 4097 bytes, 410 lines | 4097 raw | 4507 raw | both rejected, `folded 4097 > 4096` |

These are vectors N14 through N17.

### M3.4 The raw capture bound of 8192 is tight and non-lossy

Folding replaces a two-byte CRLF with one byte and rewrites nothing else, so
`len(folded) >= ceil(len(raw)/2)`.

| raw stream | raw bytes | folded bytes | replacement verdict |
|---|---|---|---|
| 4095 CRLF pairs | 8190 | 4095 | rejected by the **terminal-LF rule** |
| 4096 CRLF pairs | 8192 | 4096 | rejected by the **terminal-LF rule** |
| 4097 CRLF pairs | 8194 | 4097 | rejected by the raw capture limit |

The maximal expansion of an in-bound folded stream reaches 8192 raw bytes
exactly, passes the capture limit and is decided by a **semantic** rule, so
8192 is the smallest bound that never pre-empts an admission decision. Checked
over every raw length from 0 to 32768: **0** lengths above the capture bound
could fold within the semantic bound. These are vectors N18 and N19.

### M3.5 The published digests are unchanged

Both orders produce, for the measured 192-byte stream:

```text
payload 191 bytes  sha256 7d8e08339e557ede5a9e565773c4cf17f83dea27f7d0c6591869f184bd1b81b5
record  208 bytes  sha256 7fc35c11acae420849418f9c9f9b5681651beedc6d68f00bf1e4db022cd5f06b
```

and for `cargo --version`, payload `8d712854…`, record `d677e668…`. The
reordering moves no identity.

---

## Blocker 2a — Stage P must reject an auto-discovered build script

### M1 What cargo actually discovers

`cargo metadata --format-version 1 --locked --offline --quiet --all-features`,
run with `PATH=/nonexistent`, a private `CARGO_HOME` and a private `HOME`:

| # | manifest | files present | `custom-build` target reported |
|---|---|---|---|
| 1 | **no `build` key** | `build.rs` | **YES** |
| 2 | `build = false` | `build.rs` | no |
| 3 | `build = false` | — the `cargo vendor` shape | no |
| 4 | `build = "build.rs"` | `build.rs` | YES |
| 5 | no `build` key | `custom.rs` | no |
| 6 | no `build` key | `sub/build.rs`, `src/build.rs` | no |
| 7 | no `build` key, **edition 2024** | `build.rs` | YES |
| 8 | no `build` key | a **directory** named `build.rs` | no |

Row 1 is the gap. The manifest declares nothing, so a forbidden-key gate has
nothing to reject; row 4 was already rejected by the `package.build` row, and
rows 2, 3, 5, 6 and 8 are inert. Discovery keys on exactly one thing: a regular
file named `build.rs` in the manifest's own directory. Edition 2024 does not
change it.

### M2 The consequence is real code execution

Same fixture as row 1, with a `build.rs` that writes a marker file:

```text
cargo metadata … exit=0   marker present: NO
cargo build --release --locked --offline exit=0
                          marker present: YES
                          content: "auto-discovered build script executed"
```

The graph phase still executes nothing — the decision 0008 section 7 premise
holds — but the compile phase runs the script. Before the file rule this shape
was decided only by G2, that is, after Cargo had read and resolved the tree.

### The rule taken

One `lstat` per manifest directory, on the name alone: any entry named
`build.rs` directly in a manifest's directory is
`build_rust_build_script_forbidden`. It over-rejects rows 2 and 8, both measured
inert, neither produced by `cargo vendor`. Stating it on the name rather than on
`package.build` means it needs no reasoning about Cargo's discovery semantics and
stays correct if they change.

---

## Blocker 2b — the physical path check must be total

### M4 An absent leaf has no canonical form

Fixture: `R/pkg/src/lib.rs` exists, `R/pkg/afile` is a regular file,
`R/pkg/linkdir` is a symbolic link to `R/elsewhere`.

| path | canonicalization |
|---|---|
| `pkg/src/lib.rs` — exists | resolves |
| `pkg/src/absent.rs` — the permitted absent leaf | **ERROR: no such file or directory** |
| `pkg/linkdir/absent.rs` | **ERROR**, and the error names `R/elsewhere/absent.rs` — a path outside the package directory |
| `pkg/afile/absent.rs` | **ERROR: not a directory** |
| `pkg/src` — the deepest existing ancestor | resolves |

`realpath(3)` agrees directly: `NULL` with `errno 2` for the absent leaf, a
resolved path for the existing directory.

Row 2 is the impossibility the review named: section 4.3 permitted the leaf to
be absent and required its canonical form in the same paragraph. Row 3 is worse
than impossible — a naive canonicalization **reports a path outside the package
directory** before failing, so it is not a containment check at all. Row 5 is
the anchor the replacement uses.

### The algorithm taken

Containment is decided by the lexical grammar plus a non-following component
walk that rejects any symbolic link or reparse point at any depth including the
leaf, and stops at the first component that is absent or is a non-directory.
Canonicalization is applied only to the deepest existing ancestor, which always
exists — in the worst case the build root itself — and is a cross-check for case
folding and Windows short names rather than the primary gate. Vectors P31
through P37 pin every branch, and a totality test drives nine shapes through both
leaf policies asserting that none escapes without a verdict.

### The G2-through-G8 counterpart claim

Section 7.2 previously asserted collectively that every row G2 through G8 has a
snapshot-byte counterpart. That was false for exactly the row-1 shape above. It
is now a per-row table naming each counterpart, with G2 citing both the
`package.build` key row and the new file rule.
