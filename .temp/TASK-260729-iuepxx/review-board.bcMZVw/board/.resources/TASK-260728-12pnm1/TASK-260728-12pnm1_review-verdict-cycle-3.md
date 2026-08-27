# TASK-260728-12pnm1 review cycle 3 verdict

## CHANGES REQUESTED

Route to `analysis`. This is contract and research rework, not implementation-code
rework. The reviewer supplied no `commit_ack` and changed no project code.

The two cycle-2 findings are closed: the 8192-byte raw capture bound followed by
CRLF folding and the 4096-byte semantic bound is total and non-lossy, and the
Stage P component walk now gives permitted absent leaves a total verdict while
the new `build.rs` file rule closes Cargo's keyless auto-discovery case.

Two independent blockers remain.

## Blocking finding 1: admitted profile settings launch a package-selected process

Reference section 7.5 admits every `[profile]` table and claims a profile
"cannot inject a flag or start a process." Decision section 9 repeats that
profiles can select codegen tuning without starting another process. The
allow/reject matrix has no row for stable `split-debuginfo`, and positive vector
10 covers only `opt-level` and `lto`.

That claim is false on the contract's only measured platform. With Rust 1.91.0
on macOS arm64, this otherwise admitted root manifest:

```toml
[package]
name = "profile-process-probe"
version = "0.1.0"
edition = "2021"

[profile.release]
debug = 2
split-debuginfo = "packed"
```

was run through the contract's exact graph and compile vectors, private
`HOME`/`CARGO_HOME`/target/tmp roots, absolute `RUSTC`/`RUSTDOC`, pinned
`rust-lld`, `ld64.lld` flavour, pinned SDK, offline Cargo config, and a poisoned
`PATH` containing only a logging `dsymutil` shim.

Observed:

```text
cargo metadata ... --all-features     exit 0
cargo build ... --release ...         exit 0
warning: processing debug info with `dsymutil` failed: exit status: 127
process log:
dsymutil invoked: .../target/aarch64-apple-darwin/release/deps/profile_process_probe-9d33143aee5cdb92
```

The shim really executed. Cargo still returned success after the shim returned
127. This is not a hypothetical future behaviour: rustc's documented macOS
`packed` split-debuginfo mode runs external `dsymutil`, and the measured replay
shows the exact admitted pipeline reaches it.

This falsifies all of:

- the claimed three-executable process closure;
- the poisoned-`PATH` acceptance test's required zero entries;
- the statement that admitted profiles cannot start a process;
- the acceptance criterion requiring every package-selected process/native
  input to have an explicit allow/reject/control verdict before compile.

Required rework:

1. Define a closed Stage P profile grammar. At minimum,
   `split-debuginfo = "packed"` must not remain generally admitted; either
   reject every process-capable profile form before Cargo or prove a
   manager-owned override that prevents the lookup under the exact pipeline.
2. Audit all stable profile keys and platform defaults for auxiliary tools,
   including the Windows and Linux qualification paths, and record the result
   row by row in the matrix.
3. Add a macOS poisoned-`PATH` vector using the fixture above, with the current
   contract as an expected-red control and the replacement required to produce
   zero entries.
4. Update the process graph, admitted-surfaces prose, conformance inventory and
   receipt-policy meaning consistently.

Primary references:

- https://doc.rust-lang.org/stable/rustc/codegen-options/index.html#split-debuginfo
- https://doc.rust-lang.org/cargo/reference/profiles.html#split-debuginfo
- https://doc.rust-lang.org/beta/nightly-rustc/src/rustc_codegen_ssa/back/link.rs.html

## Blocking finding 2: G11 depends on an undefined Stage P admitted source set

Reference section 4.1 defines `M` only as the set of `Cargo.toml` files. Section
4.3 admits paths only when a manifest contains one of the enumerated
path-bearing keys. Sections 4.7 and 7.2/G11 then require every graph
`manifest_path` and target `src_path` to be in "the Stage P admitted set," but
the contract never defines that set or adds Cargo's auto-discovered source
targets to it.

An ordinary positive package uses `src/main.rs` without a `[[bin]].path` key.
Libraries and auxiliary targets similarly have Cargo-defined default paths.
Under the written algorithm `Cargo.toml` is in `M`, but `src/main.rs` enters no
declared-path rule. An implementation therefore has two incompatible choices:
reject an ordinary positive build at G11, or invent an unreviewed Cargo
auto-discovery algorithm inside Stage P. The probe has no G11 positive replay;
positive vector 18 checks only the absence of `build.rs`, and structural checks
never construct or compare the admitted set.

Required rework:

1. Define the Stage P admitted-set construction completely, including
   auto-discovered `src/main.rs`, `src/lib.rs`, named/unnamed bins, examples,
   tests and benches, or replace G11 with a check whose input set is already
   defined.
2. State how that construction behaves for unreachable manifests inventoried
   under `R`, missing default source leaves, and the root's single-bin rule.
3. Add positive and negative G11 vectors that exercise declared and
   auto-discovered target paths, and make the executable probe compare the
   manager-side set with actual metadata.

## Independent verification

- Extracted probe: `gofmt -l .` reported no files; `go vet ./...`,
  `go build ./...`, and `go test ./...` exited 0.
- Full macOS arm64 replay with the declared Rust root and SDK exited 0:
  19/19 cases, P1 and P2 true, 24 closure checks with zero verdicts, all
  12 controls failing as required, 81 vectors with zero divergences, and
  14 structural checks with zero divergences.
- The new profile-process replay above is outside the submitted probe and
  exposes a missing matrix row; the submitted green result therefore does not
  contradict it.
- The cycle-2 CRLF and absent-path/build-script blockers were independently
  checked and are closed.

Re-review requires both blockers to be closed in the decision, normative
reference, executable probe and conformance inventory, with the poisoned-path
profile replay green in the rejecting/controlled direction and ordinary
auto-discovered `src/main.rs` receiving a defined G11 verdict.
