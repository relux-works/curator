# Reviewer verdict for TASK-260810-3urqbl

Verdict: **changes requested -> analysis**

## Goal and scope evidence

- Reviewer run: `RUN-260811-e96528`
- Authoritative goal immediately before verdict: `GOAL-260811-38ee21`
  revision 1
- Resolved scope: `TASK-260810-3urqbl`
- Review policy: `required`
- Reviewed outcome:
  `TASK-260810-3urqbl_rust-cargo-source-closure.md`
- Reviewed SHA-256:
  `c85fa21cabdc53ffcd3dbae4149b537eb575616367fa711e4451f516039e7d39`
- The orchestrator's focused R1/R2 progress directive was observed at the
  review checkpoint.

The goal requires exactly one evidence-backed reviewer branch. This artifact
records only the `changes_requested` branch; it records neither acceptance nor
a Stop-The-Line boundary.

## Focused assessment

R1 is resolved. The revised procedure now separates immutable acquisition,
shared-policy pre-vendor admission, confined offline vendoring from an admitted
private Cargo home, and post-vendor verification before metadata or build.
Procedure steps 3-7 and `PV01` make the zero-Cargo-spawn rejection boundary
explicit.

R2 is not fully resolved. The new `cargo-vendor-transform-v1` is materially
better than literal leaf equality, and its retained happy-path registry/Git
fixtures replay. However, two rules contradict the exact Cargo 0.92.0 source
that the transform claims to pin. Both contradictions cause deterministic
false transform mismatches for inputs the profile currently presents as
covered. An exact, implementation-ready origin-to-vendor transform therefore
has not yet been established.

## R2a — Registry `.cargo-ok` omission is basename-wide, not root-only

The outcome says that only exact package-root `.cargo-ok` is omittable and that
a nested same-named leaf is not (`.research/260811_rust-cargo-source-closure.md:264`).

Cargo 0.92.0 does something different:

1. The registry branch of `cargo vendor` calls `unpack_package_in` with the
   ordinary vendor filter
   ([`vendor.rs` lines 244-283](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/ops/vendor.rs#L244-L283)).
2. The shared registry unpacker then tests `entry_path.file_name()` against
   `PACKAGE_SOURCE_LOCK` (`.cargo-ok`) and skips the entry regardless of its
   parent path
   ([`registry/mod.rs` lines 1075-1081](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/registry/mod.rs#L1075-L1081)).

Consequently, `nested/.cargo-ok` is scanned at pre-admission but is not emitted
by registry vendoring. The submitted transform instead predicts
`copy_identical`, so post-vendor comparison rejects the missing leaf rather
than recording Cargo's actual deterministic omission. The existing `RV02`
fixture covers a nested `.gitignore`, not the distinct basename-wide
`.cargo-ok` behavior.

Required rework:

1. Make registry `.cargo-ok` handling source-specific and exact: either assign
   every basename-matching origin leaf a dedicated omission disposition, or
   explicitly reject that narrower input at pre-admission. Do not apply the
   registry unpack rule to Git sources, where `vendor_this` filters only the
   exact package-root relative path.
2. Bind the selected rule and every omitted leaf in the transform ledger and
   checkpoint.
3. Extend `RV02` with root and nested `.cargo-ok` archive members and assert
   the expected vendor leaf set, generated checksum map, and diagnostic for
   any deliberately unsupported branch.

## R2b — Tracked Git files under root `target/` are re-added

The outcome says the Git package projection excludes the package-root
`target/` directory (`.research/260811_rust-cargo-source-closure.md:287-288`).

For a Git package without an explicit `include` list, Cargo 0.92.0 first keeps
`target/` out of the ordinary working-tree walk, but then explicitly appends
unconflicted index entries under that prefix
([`path.rs` lines 720-785](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/path.rs#L720-L785)).
Cargo's own `include_files_called_target_git` regression test confirms that an
untracked root `target/foo.txt` is absent and the same file is present after it
is committed
([`package.rs` lines 4132-4193](https://github.com/rust-lang/cargo/blob/0.92.0/tests/testsuite/package.rs#L4132-L4193)).

The submitted transform predicts `omit_unselected` for that tracked leaf while
Cargo vendors it. The post-vendor tree therefore contains an allegedly
unexplained addition and fails `rust_vendor_incomplete`. This remains safely
fail-closed, but it is not the exact supported transform or implementation-ready
coverage promised by R2. Any compiled artifact in such a path must still be
rejected by the shared pre-vendor classifier; correcting selection must not
create a `target/` bypass.

Required rework:

1. Describe the pinned `PathSource` branches exactly, including the tracked
   root-`target/` exception when no `include` list is present and the distinct
   filesystem-walk behavior when `include` is present.
2. Extend `GV03` (or add a focused case) with tracked, untracked, and ignored
   root-`target/` files plus an `include`-list variant. Assert every origin
   disposition, the normalized manifest input file set, vendor output, and
   generated checksum bytes.
3. Add a compiled tracked-`target/` member variant proving that shared artifact
   denial occurs during pre-vendor admission with zero Cargo spawns.

## Independent verification

- Research source and board outcome are byte-identical: 986 lines, 74,973
  bytes, SHA-256
  `c85fa21cabdc53ffcd3dbae4149b537eb575616367fa711e4451f516039e7d39`.
- Document gates passed: nonempty, valid UTF-8, no trailing whitespace,
  balanced Markdown fences, required R1/R2/checkpoint terms present, all local
  links resolve, and source/resource parity holds.
- The reviewed Cargo checkout is tag `0.92.0`, peeled commit
  `ea2d97820c16195b0ca3fadb4319fe512c199a43`, matching physical Cargo 1.91.0.
- `verify_mapping.rb` exited 0 for the retained registry/Git happy paths,
  generated metadata, omissions, and forged-metadata divergence.
- Fresh empty-destination `cargo vendor --locked --offline --versioned-dirs`
  under macOS network denial and a read denial on the original Git repository
  exited 0. Its tree matched the retained regenerated tree; the key generated
  hashes were `35abe1...c81d`, `e0cf59...df3`, and `152f27...0ddb5`.
- A fresh private-`CARGO_HOME`, fresh-target `cargo build --frozen --release`
  under the same network/origin-read denial exited 0, and the binary printed
  `17`.
- `go test ./internal/buildmeta ./internal/buildcache ./internal/godriver`
  passed. A cached `go test ./...` reproduced the already-recorded historical
  test-input scan anomaly and reached about 17 GB RSS without package output;
  only that reviewer-launched coordinator was terminated. It is not used as
  verdict evidence.
- Official Cargo documentation independently confirms lock all-feature/all-
  target breadth, metadata target filtering, remote-source vendoring,
  `--frozen = --locked + --offline`, non-security directory checksums,
  build-script execution, proc-macro compiler-level access, hierarchical
  configuration, and recursive Git submodule fetching.
- No product or research code was modified during review.

## Routing decision

This is ordinary research rework, not a Stop-The-Line boundary. Route
`TASK-260810-3urqbl` to `analysis`, correct the two pinned-transform rules,
extend the conformance evidence, rerun the focused gates, and return through a
new reviewer cycle.
