# TASK-260728-12pnm1 — rework cycle 4 handoff

Both blocking findings of `TASK-260728-12pnm1_review-verdict-cycle-3.md` are
closed, each by a host measurement, a contract change and an executable check.
Every accepted cycle-1, cycle-2 and cycle-3 finding is retained. Decisions 0007
and 0008 and every frozen artifact are untouched. macOS arm64 remains the only
measured platform. Nothing was staged, committed, published or pinned.

---

## Blocker 1 — an admitted `[profile]` table launched a package-selected process

**The finding was correct and reproduced exactly.** `[profile.release]` with
`debug = 2` and `split-debuginfo = "packed"`, run through the contract's exact
two argument vectors with private `HOME`/`CARGO_HOME`/target/tmp roots, absolute
`RUSTC`/`RUSTDOC`, pinned `rust-lld`, `ld64.lld` flavour, pinned SDK, the
manager-written offline config and a poisoned `PATH`, resolved `dsymutil`
through `PATH`. The shim ran, returned 127, and `cargo build` still exited 0.

### What the audit found beyond the finding

The full stable profile key set was swept — 84 build roots, one profile shape
each. **Exactly one key** reaches an executable outside the closure, and two
measurements make the obvious fix wrong:

1. **An unrecognized value is worse than `"packed"`.** With
   `split-debuginfo = "wat"`, `""` or `"PACKED"`, cargo forwards **no**
   `-Csplit-debuginfo` flag at all — captured from a logging `RUSTC` wrapper —
   and `rustc`'s own default for `aarch64-apple-darwin` is `packed`, so
   `dsymutil` runs anyway with no warning and exit 0. A deny-list naming
   `"packed"` is bypassed by any garbage string.
2. **An unknown profile key is accepted silently.**
   `totally-unknown-key = 42` exits 0 with no diagnostic, so a key deny-list
   cannot stay closed against a future stabilization.

A third measurement matters for ownership: because the fallback is a *toolchain*
default, the previous revision's pipeline was safe only because cargo happens to
emit `unpacked` when the key is absent. That is not a property this contract
controls.

### The fix — two mechanisms, in the contract's own order

**Primary, from snapshot bytes (new reference section 4.7).** A closed
**positive allowlist** over every manifest in the build root: three admitted
table shapes (`[profile.<name>]`, `[profile.<name>.package.<spec>]`,
`[profile.<name>.build-override]`), eleven admitted keys each with its own closed
value set, and `split-debuginfo` rejected **outright at every value**, because
the driver publishes one executable and discards every by-product, so no
packaging of debug information has an admitted purpose. Every other key —
including a future stabilization — is rejected. Diagnostic:
`build_rust_manifest_key_forbidden`; nothing new is minted.

**Second layer, in the environment (reference section 6).**
`CARGO_ENCODED_RUSTFLAGS` gains `-Csplit-debuginfo=<pin>`, a manager constant
resolved per operating system, `off` on macOS. Measured: with the pin,
`"packed"`, an unrecognized value, the empty string and a
`[profile.release.package.<root>]` override all produce **0** `PATH`
resolutions, build with exit 0 and yield a running artifact; without it each
resolves `dsymutil`. The Windows and Linux values are qualification obligations
and are explicitly **not** assumed to be `off`, because the packaging step
differs there — a `PDB` written by the pinned linker on MSVC, an external `dwp`
on Linux.

**The over-rejection is bounded by measurement, not by hope.** Of the 506
published crate manifests in the host's registry cache, 37 carry a profile
table, they use seven keys between them (`debug`, `lto`, `opt-level`,
`codegen-units`, `panic`, `incremental`, `inherits`), **none** uses
`split-debuginfo`, and all 37 pass the gate. A profile table in a path or
vendored dependency was measured **inert** and is rejected anyway — the same
stated over-rejection the `[patch]` row already carries, taken so the rule needs
no reasoning about which manifest cargo treats as the workspace root.

**The matrix row that did not exist.** `cargo metadata` reports no profile
information at all, so no row G0 through G11 can see this surface. That is now
stated in section 7.2 as the reason the verdict lives in snapshot bytes and the
environment rather than in the graph phase, and section 7.5 carries a per-shape
allow/reject table.

**Acceptance test.** Section 13.1 gains a new **step 4**: the profile replay, with
the unpinned run as its own control and required to fire. Section 2.1 now states
that "three executables" is a conclusion of the gates rather than a property of
Rust, and that any future admitted surface must re-run step 4 before the table
may be restated.

**Evidence.** Structural check `profile-tables-cannot-start-a-process` runs the
six-run matrix inside the probe. Controls **C13** (profiles admitted wholesale)
and **C14** (a `"packed"` deny-list) are required to fail; C14 additionally
re-runs its fixture with the pin and reports itself as *not* failing if any
resolution survives. Vectors **F1 through F27** are host-independent.

---

## Blocker 2 — G11 depended on an undefined Stage P admitted set

**The finding was correct.** Section 4.1 defined `M` as manifests, section 4.3
admitted declared paths, and sections 4.7 and G11 then required every reported
`manifest_path` and `src_path` to be "in the Stage P admitted set" — a set the
contract never constructed. An ordinary package uses `src/main.rs` with no
`[[bin]].path` key, so an implementation had to either reject an ordinary
positive build at G11 or reimplement Cargo's target auto-discovery inside the
security boundary.

### The fix — define the set by the walk, not by discovery

New reference section 4.8:

> `A` is every path the link-free walk of section 4.1 enumerated under `R` —
> every file and every directory, at any depth, including under `V` — together
> with every leaf that section 4.3 admitted and permitted to be absent.

The walk never follows a link and never descends below one, so nothing reachable
only through a link is a member. `A` is finite, computed before any cargo
process, and needs no knowledge of Cargo's discovery rules.

**Sufficiency is measured, not argued.** Every path `cargo metadata` reports is
one of exactly two things, and there is no third case:

1. an **existing** file under `R` — auto-discovery finds targets by looking at
   the filesystem. Measured across `src/main.rs`, `src/lib.rs`,
   `src/bin/<n>.rs`, `src/bin/<n>/main.rs`, `examples/<n>.rs`,
   `examples/<n>/main.rs`, `tests/<n>.rs`, `benches/<n>.rs`, `build.rs`, an
   explicit `[[bin]] name = <package name>` with no `path`, and every vendored
   package's manifest and sources: all present, all under `R`;
2. a path a **path-bearing key declared**, which section 4.3 already decided.
   Measured: `[[bin]] path = "src/ghost.rs"` and `[lib] path = "src/nolib.rs"`
   with no such file are reported with `exists = false` — exactly the absent-leaf
   case 4.3 admits.

Measured that no third case exists: a name-only `[[bin]]` or `[lib]` with no
discoverable file makes the graph phase exit **101**
(``can't find `ghostonly` bin at `src/bin/ghostonly.rs` … ``) rather than report
a phantom `src_path`, and a package with no target at all exits 101 with
`no targets specified in the manifest`.

The three cases the verdict asked to be stated explicitly are answered in 4.8:
**unreachable manifests** under `R` are in `A` and cost nothing because
membership is a superset property; **missing default source leaves** are in `A`
when declared and cannot be reported when undeclared; and the **root's
single-bin rule** stays G9's job — `A` decides containment, G9 decides shape, and
an auto-discovered `src/bin/<n>.rs` is rejected by G9 exactly as before.

**Evidence.** Structural check
`stage-p-admitted-set-covers-auto-discovered-targets` runs the exact graph
vector over a default-layout package using every auto-discovered shape plus one
declared absent leaf, and requires every reported path to be in `A` — measured
**0 outside**. Control **C15** keeps the superseded input set runnable and is
required to fail by rejecting an ordinary `src/main.rs`; it also re-checks the
replacement. Vectors **A1 through A16** are host-independent and cover the six
auto-discovered shapes, an unreachable vendored manifest and its source, the
declared absent leaf, an undeclared absent path, an outside path, a file below a
symbolic link, the link component itself, the superseded set's failure, and G11
itself over a reported list carrying one escaping manifest.

---

## Contract deltas

| | before | after |
|---|---|---|
| conformance vectors | 150 | **195** |
| controls required to fail | 12 | **15** |
| probe host-independent vectors | 81 | **124** |
| probe structural checks | 14 | **16** |
| new diagnostics | — | **none** |
| policy-object members | 18 | **20** (`profile_policy`, `debuginfo_packaging`) |

The profile grammar rejects strictly more manifests, with the over-rejection
bounded at zero across 506 measured real manifests. The admitted-set definition
is not a rule change but a definition where there was none: it gives a verdict
where the superseded text gave none, and the verdict for an ordinary
auto-discovered `src/main.rs` is *admitted*, so it also removes a false
rejection. The `-Csplit-debuginfo` pin is the one change that alters produced
bytes; it is recorded in the policy object and enters cache identity, and it
invalidates no published artifact because none exists.

---

## Verification

Probe, from the extracted tarball:

```text
gofmt -l .      0 files
go vet ./...    exit 0
go build ./...  exit 0
go test ./...   exit 0
```

| probe run | result |
|---|---|
| full, macOS arm64 | **exit 0, green** — 19/19 cases, P1 and P2 hold, 24 closure checks with 0 verdicts, **15 of 15** controls failing as required, **124** host-independent vectors with 0 divergences, **16** structural checks with 0 divergences |
| controls-only replay | exit 0, 15 of 15 failing as required |
| no resolvable toolchain (`env -i`, `HOME=/nonexistent`) | exit 0, 19 cases `not_run` with the reason recorded, the 124 vectors still run, nothing installed |

Repository: `go build ./...`, `go vet ./...`, `go test ./...` all exit 0;
`gofmt -l` reports 0 files outside `.temp/` and `.task-board/`.

Two expected-red gates, both fully attributed to **other tasks'** scratch trees:

- `make check` exit 2 — the `gofmt` half lists 1521 files under
  `.temp/TASK-260728-2jaw7h` (a vendored go1.25.1 source tree), 4 under
  `.temp/TASK-260720-1zntv0`, 2 under `.temp/TASK-260729-1t1z2l` and 1 under
  `.temp/TASK-260729-3jmqgl`. **Zero** occurrences of `TASK-260728-12pnm1`. Its
  `go vet` and `go test` halves are green.
- `tools/validate.py` exit 1 — a broken local link `../release/1.0.0-rc.5.json`
  in `docs/external-build-repositories.md`, an untracked scratch file belonging
  to another task. Moving the six untracked scratch paths aside gives a
  clean-tree baseline of **exit 0**, "validated 30 schemas and 93 vector files";
  all six were restored afterwards.
- Scoped link check over the two authored documents: **0 broken of 4** local
  links.

`golangci-lint` was **NOT RUN** — it is not installed on this host.

---

## Board impact

**No new board element is created.** Both fixes add implementation obligations
to the already-linked `TASK-260728-q283m8`, `TASK-260728-13ioo0`,
`TASK-260728-2yxdo7` and `TASK-260728-gjxj1v` — the profile gate, the
`-Csplit-debuginfo` pin and the admitted-set construction are manager code, not
schema work. `TASK-260728-251p01` already carries the policy-object obligation
and gains two further `const` members (`profile_policy`,
`debuginfo_packaging`) inside the member set it was already going to mint; both
identifiers remain reserved and receipt schemas 3 and 4 remain unminted, so the
addition costs no frozen byte. Neither fix mints a diagnostic, a schema member
outside that object, or a pin. Creating elements for them would be decomposition
for symmetry rather than for a requirement.

## Standing limits, restated rather than quietly dropped

- `platforms` holds only `(macos, arm64)`. macOS amd64, Windows and Linux remain
  qualification obligations with a stated acceptance test, never claims. The
  Windows TOML serialization obligations of section 6.1 remain unmeasured, and
  the Windows and Linux `-Csplit-debuginfo` pin values are new unmeasured
  qualification obligations of the same kind.
- The two residual exposures of section 12 are unchanged and still open:
  compile-time file inclusion through `include_str!`, which needs a seventh
  deferred guarantee in `STORY-260728-327soo`, and FFI against base-installation
  libraries.
- Decision number `0009` stays reserved; it renumbers if Swift
  (`TASK-260728-1yhuqi`) or Kotlin (`TASK-260728-168smo`) lands it first.
