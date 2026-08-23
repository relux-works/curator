# TASK-260810-3urqbl reviewer verdict

- Run: `RUN-260811-05742e`
- Role: reviewer
- Authoritative board goal at verdict checkpoint: `GOAL-260811-d4fba8` revision 1
- Resolved scope: `TASK-260810-3urqbl`
- Review policy: `required`
- Final directive checkpoint: `request_progress:cce41a`, issued under the same
  goal revision, requested the focused R2a/R2b, RV02/RV03, GV03a-GV03e,
  outcome-parity, and focused-gate review recorded below; it is satisfied
- Reviewed outcome: `TASK-260810-3urqbl_rust-cargo-source-closure.md`
- Reviewed SHA-256: `620c789545273a1c4fc9c9baf25b9db8e4220c79f3fb1ad299ac1bdcd7e51423`
- Verdict: **accepted**

This artifact records only the accepted reviewer branch.

## Assessment

The revised Rust/Cargo research satisfies the task acceptance criteria and the
shared source-closure invariant. It recommends a deliberately narrow
`rust-source-v1` profile that captures and admits the complete lock-superset
source closure, verifies a pinned Cargo origin-to-vendor transform, derives and
binds the active target/feature graph, rejects executable build-time extension
points that cannot yet be sandboxed conservatively, and requires a fresh-home
frozen rebuild with protected checkpoint and receipt evidence.

No acceptance-blocking correctness, security-boundary, architecture-fit, or
evidence gap remains in the reviewed revision.

## Prior rework closure

1. **R1 resolved — pre-vendor admission is now normative.** Immutable registry,
   Git/submodule, root, and path origins are recursively classified before any
   Cargo process or vendor destination exists. Only admitted captures enter a
   private Cargo home; confined offline vendoring is followed by a separate
   post-vendor scan and transform comparison before metadata or build. `PV01`
   requires a recording executor to prove zero Cargo spawns and absent outputs
   for rejected raw input.
2. **R2 resolved — equality is a versioned transform, not literal leaf-set
   equality.** `cargo-vendor-transform-v1:cargo-0.92.0@ea2d978` defines the
   source-specific per-leaf dispositions, normalized Git manifest, canonical
   generated checksum bytes, expected/observed manifests, unique package
   cardinality, and deterministic mismatch diagnostics bound into
   `rust-closure-manifest-v1`.
3. **R2a resolved — registry `.cargo-ok` is basename-wide.** Every registry
   archive regular leaf whose final component is `.cargo-ok`, at root or any
   depth, receives `omit_registry_cargo_ok`; the rule is not applied to Git.
   `RV02` asserts root and nested omissions, a copied nested `.gitignore`, the
   exact vendor leaf set, and generated checksum bytes. `RV03` covers wrong
   dispositions and the reserved generated-name collision.
4. **R2b resolved — both Git root-`target/` branches are exact.** The
   `git_index_no_include` branch re-adds eligible unconflicted tracked index
   entries under root `target/`; `filesystem_include` uses the filesystem walk
   and hard-skips that root directory even when an include glob matches it.
   Dirty untracked/ignored intake is rejected before Cargo. `GV03a`-`GV03e`
   cover tracked, untracked, ignored, explicit-include, compiled tracked-target,
   monorepo, submodule, and nested-target cases.

## Acceptance mapping

| Requirement | Accepted evidence |
| --- | --- |
| Complete recursive source enumeration | Lock-superset plus target/feature active graph; raw registry archive, full Git/submodule, workspace, and path manifests; unique lock/origin/vendor mapping; recursive `aho-corasick -> memchr` fixture. |
| Immutable identity | Registry lock/index/archive checksum triple; Git commit/tree/submodule identity; pinned transform and normalizer; protected expected/observed manifests; full toolchain/SDK/linker identity. |
| Offline reconstruction | Private admitted Cargo home, exact source replacement, absent vendor destination, `--locked --offline` vendoring, and fresh private-home `cargo build --frozen` with network and origin fallback denied. |
| Undeclared-input refusal | Active `build.rs`, build dependencies, proc macros, `links`, wrappers, untrusted config, and unsupported targets fail before compilation; future hook support is separately capability-gated behind OS-enforced read/write/process/network controls. |
| Common compiled-artifact prohibition | Shared recursive classifier runs before vendoring and again after vendoring; compiled tracked-root-`target/` input is explicitly a zero-Cargo-spawn rejection case. |
| Audit checkpoints and receipts | `rust-closure-manifest-v1`, `rust-build-checkpoint-v1`, and `rust-build-receipt-v1` bind closure, transform, graphs, target/features, toolchain, commands/environment, sandbox policy, outputs, and protected-cache validation. |
| Diagnostics and unsupported cases | Stable Rust-specific codes preserve shared-code precedence; the supported-profile table and `R01`-`RH10`, `RV01`-`RV03`, `GV01`-`GV03e`, `VF01`-`PV01` fixtures fail closed. |
| Architecture fit | The proposal preserves Curator's canonical logical-input model, full-tree toolchain fingerprint, stable diagnostic vocabulary, independently derived cache expectation, exact receipt validation, and protected publication semantics. |

## Independent verification

- The repository research file and board outcome are byte-identical: 1,099
  lines, 90,084 bytes, SHA-256
  `620c789545273a1c4fc9c9baf25b9db8e4220c79f3fb1ad299ac1bdcd7e51423`.
- `verify_edge_mapping.rb` exited 0 and independently checked the synthetic
  registry archive/lock/index identity, root and nested `.cargo-ok`, copied
  nested `.gitignore`, exact checksum bytes, both Git commits, tracked
  root-target inclusion, explicit-include root-target omission, Cargo's staging
  marker, and clean-versus-dirty oracle distinction.
- The revised original `verify_mapping.rb` exited 0 for the earlier registry and
  Git origin/vendor mappings, generated metadata, omissions, and forged-metadata
  divergence.
- `validate_research.rb` exited 0 for nonempty UTF-8, trailing whitespace,
  balanced fences, focused R2 coverage, and all three local links.
- Three new fresh-home builds with physical Cargo/rustc 1.91.0 ran under
  `sandbox-exec` network denial and explicit read denial of both Git origin
  repositories. All `cargo build --frozen --release` commands exited 0, and the
  binaries printed `tracked-target`, `include-walk`, and `17`. Negative `ls`
  probes against both denied origins exited 1 with `Operation not permitted`,
  proving the read-denial rule was active rather than decorative.
- `go test ./internal/buildmeta ./internal/buildcache ./internal/godriver`
  exited 0; `godriver` completed in 26.144 seconds and the other two packages
  were valid cached passes. A repository-wide `go test ./...` was not used
  because prior reviewer runs established a separate high-memory test-input
  scan anomaly; the architecture-relevant packages are the scoped test gate.
- No product or research code was modified during review. The only new local
  file is this task-scoped review outcome; board lifecycle/resource mutations
  are the assigned reviewer deliverable.

## Primary-source fact check

- Cargo's pinned registry unpacker applies the `.cargo-ok` check to
  `entry_path.file_name()`, after the vendor include filter:
  [Cargo 0.92.0 registry unpacker](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/registry/mod.rs#L1035-L1097).
- Cargo's pinned no-`include` Git path appends unconflicted tracked index entries
  under root `target/`, while its filesystem walk hard-skips the depth-one root
  directory:
  [Cargo 0.92.0 PathSource](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/path.rs#L569-L930) and
  [Cargo's root-target regression](https://github.com/rust-lang/cargo/blob/0.92.0/tests/testsuite/package.rs#L4132-L4193).
- The vendor implementation directly extracts registry packages, uses
  `PathSource::list_files` for Git, normalizes Git manifests, writes sorted
  checksum metadata, and filters only exact package-root reserved names:
  [Cargo 0.92.0 vendor implementation](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/ops/vendor.rs#L205-L630).
- Official Cargo documentation confirms remote-source vendoring,
  `--frozen = --locked + --offline`, and that directory checksums protect only
  against accidental modification:
  [cargo vendor](https://doc.rust-lang.org/cargo/commands/cargo-vendor.html) and
  [directory sources](https://doc.rust-lang.org/cargo/reference/source-replacement.html#directory-sources).
- Official Cargo metadata and resolver documentation confirms whole-workspace
  resolution, target filtering, lockfile all-feature/all-target breadth, and
  second-pass feature selection:
  [cargo metadata](https://doc.rust-lang.org/cargo/commands/cargo-metadata.html) and
  [dependency resolution](https://doc.rust-lang.org/cargo/reference/resolver.html).
- Official Rust/Cargo documentation confirms that build scripts and procedural
  macros execute build-time code with ambient access, supporting the v1 reject
  boundary:
  [build scripts](https://doc.rust-lang.org/cargo/reference/build-scripts.html) and
  [procedural macros](https://doc.rust-lang.org/reference/procedural-macros.html).

## Verdict branch

Accepted. Attach this outcome to `TASK-260810-3urqbl` and transition the task
from `reviewing` to `done`. No `commit_ack` is supplied by this reviewer; the
task transition does not complete its parent Story because two sibling research
tasks remain in `backlog`.
