# TASK-260810-3urqbl — Rust/Cargo source-closure evaluation

Date: 2026-08-11

Role: researcher

Status: research decision revised for review

Revision note: this revision addresses reviewer findings R1 and R2 from
`RUN-260811-87b027`. It moves shared-policy admission ahead of every Cargo
vendoring action and replaces undefined vendor/origin leaf equality with the
versioned, per-leaf `cargo-vendor-transform-v1` procedure and fixtures below.

Second revision note: this revision also addresses R2a and R2b from
`RUN-260811-e96528`. Registry `.cargo-ok` omission is now basename-wide and
source-specific, while Git package projection now models Cargo's distinct
no-`include` Git-index branch, tracked root-`target/` re-addition, and explicit-
`include` filesystem-walk branch. Exact retained fixtures cover both rules and
the pre-vendor compiled tracked-`target/` denial boundary.

Delivery input: [skill-facing CLI source-closure specification](../.spec/skill-facing-cli-source-closure.md)

Shared deny policy: [compiled artifact taxonomy](260811_compiled-artifact-taxonomy-and-deny-policy.md)

Estate input: [language and reference-surface inventory](260811_inventory-language-and-reference-surfaces.md)

## Executive recommendation

Cargo can support a Go-vendor-equivalent conservative closure, but
`Cargo.lock`, `cargo vendor`, and `--offline` are only ingredients. None is a
standalone proof.

The accepted estate inventory found no Cargo manifest, lockfile, or Rust skill
CLI at the pinned revisions, so this profile is seeded by purpose-built
conformance fixtures rather than by migrating an existing Rust surface.

Implement an initial adapter profile named **`rust-source-v1`** with these
boundaries:

1. Build exactly one declared binary from a lockfile-bearing Cargo workspace
   with a stable, manager-selected native Rust toolchain.
2. Acquire and bind the exact raw registry index records/archives and Git
   commit/tree/submodule objects without asking Cargo to unpack them into a
   build snapshot. Preserve those immutable origins as intake evidence.
3. Before any `cargo vendor`, metadata, or build process exists, recursively
   inspect every root/path snapshot, raw `.crate` container, and safely
   materialized Git/submodule source tree with the accepted shared artifact
   policy. A precompiled `.rlib`, `.rmeta`, object, library, executable,
   dynamic library, WebAssembly file, or compiled nested member is rejected at
   this gate.
4. Only after admission, populate a private controlled Cargo home from those
   exact bytes and run `cargo vendor --locked --offline --versioned-dirs` into
   an absent/empty destination under OS network denial. Rescan its output and
   compare every package and leaf with `cargo-vendor-transform-v1`; the vendor
   tree is a verified derivative, never its own provenance.
5. Bind an unfiltered lock-superset graph and an exact active build graph. The
   latter is produced with `cargo metadata --format-version 1 --locked
   --offline`, the declared feature vector, and the exact native target filter.
6. In v1, reject active custom build targets (`build.rs`), build dependencies,
   procedural-macro targets, `links`/native-library packages, custom target
   specifications, cross-compilation, unstable Cargo features, arbitrary Cargo
   configuration, compiler wrappers, and package-selected linkers/runners.
7. Rebuild from a fresh private `CARGO_HOME` and empty target directory with
   manager-generated source replacement, `cargo build --frozen`, a closed
   environment, OS-enforced network denial, read-only inputs, an allowlisted
   toolchain process set, and write confinement.
8. Bind closure, graph, policy, target, full toolchain/SDK/linker identity,
   commands, environment, and output hashes into the checkpoint and protected
   receipt.

This is intentionally narrower than general Cargo compatibility. It satisfies
the repository invariant by failing closed where Cargo permits build-time code
or ambient configuration that Curator cannot yet prove safe. A later
`rust-hermetic-hooks-v1` capability may admit `build.rs` and proc macros only
after Curator can enforce and attest their complete read/write/process/network
boundary; it must not be an adapter flag that weakens `rust-source-v1`.

## Key findings

### 1. The lockfile is a broad version/source closure, not the exact build input

Cargo stores dependency resolution in `Cargo.lock` and gives locked versions
priority on later resolutions. For lock generation, Cargo resolves as if every
feature of every workspace member is enabled, then runs feature selection again
for the actual compilation. It also resolves platform-specific dependencies as
if every platform were enabled
([Cargo dependency resolution](https://doc.rust-lang.org/cargo/reference/resolver.html)).

The local Cargo 1.91.0 fixture reproduced those semantics:

- one version-4 lock contained the inactive optional path package and both the
  Unix and Windows target packages;
- `cargo metadata --no-default-features --filter-platform host-tuple` excluded
  the optional and Windows packages from `resolve.nodes`;
- the same metadata command with `--features optional` added the optional
  package and recorded `default, optional` on the root node;
- metadata without `--filter-platform` included both target-specific packages;
- the lock recorded `aho-corasick 1.1.3 -> memchr 2.8.3`, proving a recursive
  registry edge rather than only direct dependencies.

Therefore the adapter needs two related records:

| Record | Purpose | Why the other record is insufficient |
| --- | --- | --- |
| Lock-superset closure | Conservative acquisition and scanning of every package Cargo locked for workspace features and target tables | It does not encode the exact command feature selection, active target subset, host/target build unit, or all dependency-kind semantics. |
| Active build graph | Exact selected package/bin, target, dependency kinds, target predicates, and resolved per-package feature sets from metadata | It deliberately omits inactive optional/target packages that are still admitted into the captured vendor payload and must be scanned. |

`cargo metadata` format 1 exposes package sources, manifests, target kinds
including `custom-build` and `proc-macro`, dependency kinds and target
predicates, and resolved node features. Its source identifiers are explicitly
opaque, its format may add fields and enum values, and its absolute paths are
host-local. Curator must schema-check known fields, reject unknown security-
relevant target/source kinds, and canonicalize paths to closure-relative
identities before hashing
([Cargo metadata](https://doc.rust-lang.org/cargo/commands/cargo-metadata.html)).

### 2. Vendoring closes remote sources, but not workspace/path sources or trust

Cargo documents `cargo vendor` as copying every crates.io and Git dependency,
and it prints the source-replacement configuration needed to use the result.
It does **not** claim to copy the workspace or local path dependencies
([cargo vendor](https://doc.rust-lang.org/cargo/commands/cargo-vendor.html)).

The fixture confirmed this split. Its vendor output contained exactly the Git
package and the three registry packages (`aho-corasick`, `memchr`, and `itoa`),
while the root, build dependency, proc macro, optional dependency, and two
target path dependencies remained outside `vendor/`. Complete closure therefore
requires a separately frozen workspace/path snapshot and containment checks for
every metadata `manifest_path`, `src_path`, and dependency `path`.

The manager should use `--versioned-dirs` for stable package directory names,
never use `--no-delete`, and generate one exact source-replacement config in the
protected snapshot. `--respect-source-config` is forbidden during capture:
Cargo normally ignores source configuration while vendoring, whereas respecting
package/user configuration would admit an extra resolution authority. Cargo
source replacement assumes that replacement bytes are exactly the same source,
not a patch mechanism
([Cargo source replacement](https://doc.rust-lang.org/cargo/reference/source-replacement.html)).

The offline probe used a fresh empty `CARGO_HOME`, a new target directory,
`--frozen`, the vendor source replacement, and an unavailable original Git
repository. It rebuilt the recursive registry dependency and pinned Git
dependency successfully. This proves that the captured sources were sufficient
for Cargo 1.91.0; it does not prove that descendant build-time code was
sandboxed.

### 3. Registry checksums are necessary but not a source-manifest trust root

The registry index record contains the SHA-256 checksum of the `.crate` archive
as well as dependency requirements, feature definitions, target predicates,
dependency kinds, yanked state, and Rust-version data
([Cargo registry index](https://doc.rust-lang.org/cargo/reference/registry-index.html)).
For the fixture, the following three values agreed for `itoa 1.0.15`:

```text
Cargo.lock checksum:                 4a5f13b858c8d314ee3e8f639011f7ccefe71f97f96e50151fb991f267928e2c
raw itoa-1.0.15.crate SHA-256:       4a5f13b858c8d314ee3e8f639011f7ccefe71f97f96e50151fb991f267928e2c
vendor .cargo-checksum package:      4a5f13b858c8d314ee3e8f639011f7ccefe71f97f96e50151fb991f267928e2c
```

Cargo's directory source also carries `.cargo-checksum.json` with per-file
hashes. Cargo explicitly says this file protects against accidental
modification and is **not a security mechanism**
([directory sources](https://doc.rust-lang.org/cargo/reference/source-replacement.html#directory-sources)).
The fixture demonstrated the consequence:

- changing only `vendor/itoa-1.0.15/src/lib.rs` made a fresh frozen build fail
  with exit 101 and the expected/actual file hashes;
- changing that source file **and its per-file hash** in
  `.cargo-checksum.json`, while leaving the lock and `package` checksum
  unchanged, made a fresh frozen build succeed with exit 0.

Curator must therefore verify the raw `.crate` hash against both lock and index,
recursively inspect the raw archive, derive and protect its own canonical leaf
manifest, and verify the materialized vendor tree against that protected
manifest. Cargo's generated checksum file is useful corroboration, not
admission authority.

For alternate registries, v1 may support only manager-approved registries that
provide the same captured index-record/archive/checksum triple. Missing,
conflicting, unauthenticated, or unsupported registry identity fails closed.
Registry credentials and credential providers are acquisition concerns and
must not enter the offline build environment.

### 4. Vendoring is a versioned transform, not literal tree equality

Cargo documents a directory source as unpacked crate directories plus
`.cargo-checksum.json`, and says source replacement assumes equivalent source
bytes. That does not mean a raw origin and a vendor directory have identical
leaf sets. Cargo 1.91.0's implementation (Cargo repository tag `0.92.0`, commit
`ea2d97820c16195b0ca3fadb4319fe512c199a43`) directly extracts registry
archives, selects Git package files, normalizes the Git manifest, filters four
reserved root entries, and then writes checksum metadata
([tagged `vendor.rs`](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/ops/vendor.rs#L205-L322),
[copy/normalization code](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/ops/vendor.rs#L402-L585),
[reserved-entry filter](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/ops/vendor.rs#L615-L630),
[registry unpacker](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/registry/mod.rs#L1035-L1097),
and [Git/path selection](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/path.rs#L569-L930)).

The initial transform descriptor is therefore
`cargo-vendor-transform-v1:cargo-0.92.0@ea2d978`. A different Cargo
implementation commit, package-file selection rule, normalizer, checksum
encoding, or reserved-entry rule requires a new tested descriptor and
invalidates prior closure receipts. The exact Cargo binary/toolchain identity
is bound separately as usual.

The acceptance predicate is coverage under that transform, not literal tree
equality:

```text
admit_vendor(package) = pre_vendor_admission(package.origin)
                     AND exactly_one_lock_origin_to_vendor_directory
                     AND observed_vendor_manifest
                         == transform_v1(admitted_origin_manifest,
                                         package_projection,
                                         lock_checksum)
                     AND post_vendor_artifact_admission(observed_vendor_manifest)
```

Equality here is over canonical relative path, node kind, byte size, and
SHA-256 for every expected leaf plus the exact disposition of every origin
leaf. Owners and timestamps are not source identity. Links and special nodes
have already failed the shared policy. The transform ledger uses only these
dispositions: `copy_identical`, `omit_reserved`,
`omit_registry_cargo_ok`, `omit_unselected`,
`replace_normalized_manifest`, and `generate_checksum`. The registry-only
`.cargo-ok` disposition is deliberately distinct from Git's exact-root
reserved filter. No catch-all omission or mutation is permitted.

#### Pre-vendor admission and controlled execution

1. The manager acquires each registry index record and raw `.crate`, verifies
   the lock/index/archive identity, and recursively classifies the archive
   stream without extracting it into a build snapshot.
2. The manager validates Git commit/tree objects and recursively materializes
   the exact commit and submodules into a private capture using safe,
   handle-relative rules. Git administrative storage such as `.git/` is
   provenance/tool state, not a dependency payload; the complete materialized
   tracked tree is classified before vendoring.
3. Root/workspace/path snapshots are likewise frozen and classified before
   Cargo runs. Any rejection ends intake here: the Cargo invocation ledger is
   empty, the vendor destination and target directory do not exist, and no
   receipt/cache publication is possible.
4. Only admitted registry archives/index records and admitted Git worktrees are
   installed into a new manager-controlled Cargo home. With package/ancestor
   configuration and acquisition remotes unavailable, the manager invokes the
   selected physical Cargo as `cargo vendor --locked --offline
   --versioned-dirs <absent-destination>` under OS network denial. The
   destination must be absent or empty because Cargo can retain hidden entries
   and can skip a versioned registry directory that already contains
   `.cargo-checksum.json`; `--no-delete` and `--respect-source-config` remain
   forbidden.
5. Before `cargo metadata` or build, Curator scans the resulting directory and
   manager-generated source-replacement config, computes the observed manifest,
   and applies the exact mapping below. Vendoring may read only the admitted
   capture and selected toolchain and may write only its new destination and
   bounded operational logs. Any attempted fallback read, network access, or
   other write fails closed.

Cargo's command documentation confirms that `cargo vendor` copies registry and
Git dependencies, `--offline` forbids network access, and `--frozen` combines
`--locked` and `--offline`
([cargo vendor](https://doc.rust-lang.org/cargo/commands/cargo-vendor.html)).
Those flags are defense in depth; OS-enforced isolation and the read ledger
prove that no fallback source was used.

#### Registry origin-to-vendor mapping

The raw archive, not Cargo's unpack cache, is the registry origin. For each
safe archive entry under the one required `name-version/` root:

| Origin/derived leaf | Required disposition and evidence |
| --- | --- |
| Top-level archive directory | Strip the single `name-version/` structural prefix; bind its package identity. It is not a leaf. A different/multiple root is `rust_registry_identity_invalid`. |
| Exact package-root `.gitattributes`, `.gitignore`, or `.git` | `omit_reserved`; retain the origin path, size, SHA-256, class, and exact-root `vendor_this` rule. A nested same-named leaf is not reserved and copies normally. |
| Every regular leaf whose final path component is `.cargo-ok`, at package root or any depth | `omit_registry_cargo_ok`; retain the origin path, size, SHA-256, class, and basename-wide registry-unpacker rule. This rule is registry-only and has priority over the exact-root filter for root `.cargo-ok`. Every omitted byte was already classified before Cargo ran. |
| Origin `.cargo-checksum.json` | Reject before vendoring because Cargo would overwrite the manager-reserved generated name; report `rust_registry_identity_invalid`. |
| Every other admitted regular leaf, including nested `.gitignore`, nested `.gitattributes`, registry-normalized `Cargo.toml`, `Cargo.toml.orig`, and `.cargo_vcs_info.json` | `copy_identical` to the prefix-stripped path; expected and observed size/SHA-256 must match the raw archive leaf exactly. |
| Vendor `.cargo-checksum.json` | `generate_checksum`; it is the only added leaf. Its `files` map contains every emitted leaf except itself in bytewise path order with the emitted SHA-256; `package` is exactly the lock/index/archive SHA-256 string. Encoding is compact UTF-8 JSON with sorted object keys and no trailing newline. Bind its own size/SHA-256 too. |

Registry `include`/`exclude` data cannot justify another vendor omission:
Cargo's pinned implementation extracts the already packaged archive directly.
The Cargo test suite explicitly checks that archive members excluded by the
manifest are still directly extracted
([Cargo 0.92 vendor tests](https://github.com/rust-lang/cargo/blob/0.92.0/tests/testsuite/vendor.rs)).
The exact omission order matters: `vendor_this` first filters only the root
relative Git-control names and root `.cargo-ok`; the shared registry unpacker
then tests `entry_path.file_name()` and skips every remaining basename
`.cargo-ok`. Consequently `nested/.cargo-ok` is omitted, but
`nested/.gitignore` is copied. Applying that basename rule to Git would be a
transform bug.

#### Git origin-to-vendor mapping

The Git origin has two related manifests: the complete admitted commit plus
recursive submodule trees, and one package projection for each locked package
subpath. The manager independently derives the projection under the pinned
Cargo 0.92 rules rather than accepting the vendor output as the file list:

- start at the locked package root in the exact clean commit; untracked, dirty,
  filter/LFS-produced, or globally ignored ambient files are unsupported;
- when `package.include` is empty and the package `Cargo.toml` is tracked,
  select `git_index_no_include`: traverse the controlled Git worktree with the
  package-root `target/` excluded from the ordinary walk, then append every
  unconflicted tracked index entry under that exact prefix. Apply manifest
  `exclude` rules, repository ignore semantics, nested-subpackage removal, and
  the mandatory `Cargo.toml`/eligible `Cargo.lock` rules. Thus a committed
  `target/tracked.txt` is selected even when an untracked or ignored sibling is
  not;
- when `package.include` is nonempty, select `filesystem_include`: do not use
  Git prepopulation; walk the package filesystem through the include matcher
  while hard-skipping the depth-one package-root `target/` directory. Even a
  committed `target/included.txt` matched by `target/**` is unselected. A
  nested directory named `target` elsewhere is not this special root; and
- recurse into already captured subrepositories/submodules, remove nested
  subpackages, and record the branch and rule that selected or omitted every
  immutable origin leaf; and
- run with empty manager-controlled system/global Git configuration so host
  ignore, attributes, conversion, or filter policy cannot change the result.

Any other `PathSource` branch is unsupported by this descriptor. Ambient
untracked or ignored disk leaves are not immutable commit-origin leaves and
cannot receive a benign omission disposition: their presence makes intake
dirty and returns `rust_git_identity_invalid` before Cargo is spawned. This is
stricter than Cargo, which may silently omit such leaves.

These are Cargo's documented packaging rules and its actual `PathSource` file
selection boundary
([manifest include/exclude rules](https://doc.rust-lang.org/cargo/reference/manifest.html#the-exclude-and-include-fields),
[tagged `PathSource` selection](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/path.rs#L569-L930),
and Cargo's
[`include_files_called_target_git` regression](https://github.com/rust-lang/cargo/blob/0.92.0/tests/testsuite/package.rs#L4132-L4193)).

The projection branch is a checkpointed enum, not a best-effort heuristic:

| Projection mode | Exact predicate | Package-root `target/` result | Normalizer file-set evidence |
| --- | --- | --- | --- |
| `git_index_no_include` | `package.include` is empty and the package-root `Cargo.toml` has an index entry in the discovered containing repository | Ordinary Git dirwalk excludes the prefix; every unconflicted index entry below it is appended and then filtered. Tracked eligible leaves select; untracked/ignored leaves do not. | Sorted selected relative paths from the commit/index projection plus separately identified Cargo staging metadata such as the exact readiness marker; bind index/tree and filter inputs. |
| `filesystem_include` | `package.include` is nonempty | The depth-one root directory is never traversed, even when an include glob matches it. | Sorted paths actually selected by the include-filtered filesystem walk; no Git-index re-addition and no excluded staging marker. |

Failure to prove exactly one branch, a conflicted index stage, a tree/index
disagreement, or an ambient leaf in the supposedly clean capture is
`rust_git_identity_invalid`. The complete commit tree is still pre-admitted
before either selection branch, so an ultimately omitted source leaf cannot
hide a compiled artifact.

Apply these exact dispositions to the projection:

| Git origin/derived leaf | Required disposition and evidence |
| --- | --- |
| Origin leaf outside the package projection | `omit_unselected`; retain its path, size/SHA-256, selected projection branch, and exact include/exclude/ignore/root-target/subpackage/outside-package rule. It remains in the complete pre-admitted Git-tree manifest. |
| Selected exact package-root `.gitattributes`, `.gitignore`, `.git`, or `.cargo-ok` | `omit_reserved` under Git's exact-root `vendor_this` rule. A selected nested `.cargo-ok`, `.gitignore`, or `.gitattributes` is **not** covered by the registry basename rule and copies normally. Git administrative `.git/` is provenance storage rather than an origin source leaf. |
| Selected ordinary regular leaf other than `Cargo.toml`, including an unconflicted tracked package-root `target/` leaf in `git_index_no_include` | `copy_identical`; observed bytes must match the admitted commit/submodule leaf. |
| Package-root `Cargo.toml` | `replace_normalized_manifest`. Retain the original manifest hash and independently compute the one expected output under the pinned normalizer: parse the admitted manifest/workspace, resolve `workspace = true` inheritance, use the derived package file set, normalize included build-script and explicit/inferred lib/bin/example/test/bench targets, construct the normalized manifest, and serialize with Cargo's normalized-manifest encoding. Bind the normalizer ID, input hashes, output bytes, size, and SHA-256. No other semantic or textual mutation is allowed. |
| Vendor `.cargo-checksum.json` | `generate_checksum` exactly as above, except `package` is JSON `null`; its file map hashes the generated `Cargo.toml` and every copied leaf, excluding itself. |

Cargo also uses an exact checkout-root `.cargo-ok` as a readiness marker
([pinned checkout guard](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/git/utils.rs#L27-L29),
[reset/ready behavior](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/git/utils.rs#L340-L378)).
When the commit does not track that path, the manager records Cargo's empty
marker separately as generated staging metadata: it may enter Cargo's
`packaged_files` normalizer input, is filtered from vendor output by the Git
exact-root rule, and never becomes an origin leaf. When the commit does track
`.cargo-ok`, its admitted origin bytes and `omit_reserved` disposition remain
authoritative. Any collision, overwrite, or unexpected marker bytes fail with
`rust_git_identity_invalid`.

Cargo does **not** automatically add `Cargo.toml.orig` for a Git dependency,
despite the explanatory comment in its normalized manifest. If that path exists
in a vendor package, it must be a byte-identical selected origin leaf; otherwise
it is an unexplained addition. The source-replacement config and each
`name-version` directory are also independently derived from the normalized
lock source IDs and `--versioned-dirs`; Cargo's stdout is compared with that
expected config rather than trusted.

For both source kinds, any source leaf without one disposition, two origin
leaves targeting one vendor path, unexpected vendor leaf, missing expected
leaf, byte mutation, noncanonical checksum JSON, checksum self-inclusion,
wrong `package` value, wrong normalized manifest, extra package directory, or
zero/multiple lock-to-directory mapping is a deterministic rejection. Registry
origin/copied/generated-leaf mismatches use `rust_registry_identity_invalid`;
Git commit/projection/copied/normalized/generated-leaf mismatches use
`rust_git_identity_invalid`; unexplained coverage/cardinality/config/directory
structure uses `rust_vendor_incomplete`. A shared artifact diagnostic still
takes precedence when the post-vendor bytes themselves are compiled, opaque,
unsafe, or otherwise denied.

The retained probes make the distinction concrete:

- `itoa 1.0.15` raw `.gitignore` hash
  `07d64e…14e17` is `omit_reserved`; registry `Cargo.toml` hash
  `c1d45a…667bc` is copied unchanged; generated checksum-metadata hash is
  `35abe1…c81d`.
- retained `git_leaf 0.1.0` changes only the manifest: admitted `Cargo.toml`
  hash `22df0b…2c192` maps to normalized hash `e0cf59…df3`, while
  `src/lib.rs` remains `e7afc8…34703`; generated metadata is
  `152f27…0ddb5`.
- `.temp/TASK-260810-3urqbl-vendor-mapping-probe` uses workspace inheritance,
  an explicit include set, all three reserved file names, and an unselected
  file. Cargo 1.91.0 emitted only normalized `Cargo.toml` (`778cb0…2e73`),
  `kept.txt` (`d2be64…8352`), and `src/lib.rs` (`b6060d…ec37`), then generated
  metadata `be88ee…9a39`. Fresh-home frozen build exited 0. Replacing
  `kept.txt` and forging its generated hash also exited 0, confirming again
  that only Curator's protected origin/transform manifests establish identity.
- `.temp/TASK-260810-3urqbl-transform-edge-probe` covers the two second-review
  findings. A synthetic, lock/index/archive-consistent registry fixture
  (`e1d506…c1c6f`) omitted both root and nested `.cargo-ok`, copied
  `nested/.gitignore`, and generated checksum bytes `ae5e67…6ac80`. The
  no-`include` Git fixture at commit `2253ef8…fe29d` copied committed
  `target/tracked.txt` while Cargo omitted ambient untracked/ignored siblings;
  the profile treats those ambient siblings as dirty-intake rejection rather
  than safe omissions. The explicit-`include` fixture at commit
  `972c661…eb159` omitted committed `target/included.txt` despite a matching
  `target/**` rule. Its independent verifier and all three fresh-home frozen
  rebuilds exited 0; the Git origins were unavailable during rebuild.

The retained fixture record stores full digests; abbreviated prose above is not
the checkpoint representation:

| Probe leaf | Disposition | SHA-256 |
| --- | --- | --- |
| `itoa` raw `.gitignore` | `omit_reserved` | `07d64eb09a56c853b7853a47276799b38d7d6d25a4314b0c1abc10432c214e17` |
| `itoa` raw/vendor `Cargo.toml` | `copy_identical` | `c1d45a6aa2324a0862b0e6c8100e8f595616f91612f915f63c862010954667bc` |
| `itoa` freshly regenerated vendor `.cargo-checksum.json` | `generate_checksum` | `35abe1dc1588d7a386b232f534a6d6bab406859c89809454735e9c84d1c7c81d` |
| edge registry raw archive / lock / index package identity | immutable origin | `e1d506e60e9cd73c82d5aa97455fdb24982bd18aa327db6a6ecc8e6203dc1c6f` |
| edge registry root `.cargo-ok` | `omit_registry_cargo_ok` | `1c9e0e3d96aec65d6355f3ffdd7c833f9002dc8f3dbfb7d280a96e2784bce831` |
| edge registry `nested/.cargo-ok` | `omit_registry_cargo_ok` | `816e6eeb6a9eb7b0c8676857f25e62efa1b51d5236f0ea52254acbd93be6edf8` |
| edge registry raw/vendor `nested/.gitignore` | `copy_identical` | `15436dec9c76f745d89de7caf59f35ce0476169d09c18721f3a35f224eef3784` |
| edge registry vendor `.cargo-checksum.json` | `generate_checksum` | `ae5e670fffa04e8cee82764ac110f92f474924b62dd95c02b19df2d6c406ac80` |
| `git_leaf` origin `Cargo.toml` | input to `replace_normalized_manifest` | `22df0b5090fee636a8ff5d99e0ff4cf66500b2d102d5e4f8ed34a4ab9fd2c192` |
| `git_leaf` vendor `Cargo.toml` | `replace_normalized_manifest` | `e0cf597dec640399684e2885a195321faf318991ede0daef58dd3508cfb84df3` |
| `git_leaf` origin/vendor `src/lib.rs` | `copy_identical` | `e7afc8ce5c0343889068f76a38c7a3ecfb40ce487881fe6452194ff669734703` |
| `git_leaf` vendor `.cargo-checksum.json` | `generate_checksum` | `152f279e784acefa5ec72517b0a535e25cb491fd3f6935446152ba9715a0ddb5` |
| mapping probe origin `member/Cargo.toml` | input to `replace_normalized_manifest` | `7c2d965208d049111d8e00f3de3077d83fd6cf4bd989856d7e023bd75bd3b578` |
| mapping probe origin `.cargo-ok` | `omit_reserved` | `0a19e4fc5144f8f80beba3ef4a3d7837683f3d678fcb7571da3d469e70b34840` |
| mapping probe origin `.gitattributes` | `omit_reserved` | `9dd99f5ba0ec5566fd413ccdd9076a6cfcb8b09b156f1ac0ba2f5ba8012dd9bf` |
| mapping probe origin `.gitignore` | `omit_reserved` | `1cca3008a57a04abb2259ff9d1aaf03bd3b7de887d0b4744c0932fca8e150c32` |
| mapping probe origin `omitted.txt` | `omit_unselected` | `7f3a6b6f662852505a843138b915d6df700e71c85538b4da212e25d3916dd881` |
| mapping probe origin/vendor `kept.txt` | `copy_identical` | `d2be641f158ad87d4fa545fb48976ce9b7714f655dcac702bb8e54dff97e8352` |
| mapping probe origin/vendor `src/lib.rs` | `copy_identical` | `b6060d1c2153673600344de056dd9c7849c76d51584f33d2c3df93b828c4ec37` |
| mapping probe vendor `Cargo.toml` | `replace_normalized_manifest` | `778cb0cb3bdffda543a0f0547e2a1c80a2d17629c6e879856792eadec0692e73` |
| mapping probe vendor `.cargo-checksum.json` | `generate_checksum` | `be88ee8f47df748b19c97ba4b3e09c21b3745007954902a04e241836a5729a39` |
| mapping probe forged `kept.txt` | rejected expected/observed mismatch | `095444724efa8fbcd5de9f8abe986497dc4cf748c02c8e0ba70b4ab86e50af9c` |
| no-`include` Git origin/vendor `target/tracked.txt` | `copy_identical` | `53d5d5fd5e48ad9196ba469911ac76a7f6204b088fa1655d06982da62116b48e` |
| no-`include` Git origin / vendor `Cargo.toml` | `replace_normalized_manifest` input / output | `0593fb89d6e51917dc9ce40d9c9a67ffa3211394c11d5e07431b79676565b69b` / `5b1aecc3f68ea3bc15d46c73e380381464865d990280b6cefee8363e3af2ba40` |
| no-`include` Git vendor `.cargo-checksum.json` | `generate_checksum` | `3937bbcb49e3a82c7b844818926953ec23e44b4c8514d09de40c23bcff762f75` |
| no-`include` ambient `target/untracked.txt` / `target/ignored.txt` | not commit-origin leaves; dirty intake rejection | `721d5b0061b05d7c358d1c8ffaac99a8a355b084a7eb310ef22fcb55e6e1f4a2` / `79811f7b31e9ff4229a8ae21419753e27b87c425f99d310555e74ad4f0f00ed0` |
| explicit-`include` Git origin `target/included.txt` | `omit_unselected` (`filesystem_include` root-target rule) | `6ac0e73d0c402bd88b7381e4774eef70612366d7cb06b3e8266f49c0e9f049f6` |
| explicit-`include` Git origin / vendor `Cargo.toml` | `replace_normalized_manifest` input / output | `938726437179f3bb61a5904fcfe131e22080afc77d44b8d428eb2f6961f719fe` / `747ef661e5a77ca2cee2e494a6b69c98752ae7d85c20e488d6f5d0db84f52b19` |
| explicit-`include` Git vendor `.cargo-checksum.json` | `generate_checksum` | `d472136289d4f41b5fbc0f65ede9e5425feeaea82d8179e57945bfc27d38dda6` |

### 5. Git commits close selection only when the repository tree is also captured

Cargo permits Git dependencies selected by a branch, tag, or revision, and
records the selected commit in `Cargo.lock`. Cargo also recursively fetches Git
submodules
([Cargo Git dependencies](https://doc.rust-lang.org/cargo/reference/specifying-dependencies.html#specifying-dependencies-from-git-repositories)).
The fixture lock recorded the full commit
`e6305525f4f7d5be0f82d7ddd117ca63b73497f3`; the vendored Git package retained
that source identity in metadata, while its `.cargo-checksum.json` had
`"package": null` and only per-file hashes.

For every Git source, the checkpoint must bind:

- normalized declared URL and selector plus Cargo's exact source ID;
- the full resolved commit object ID from the lock, commit/tree object evidence,
  and the source package path Cargo selected within the repository;
- every recursively fetched submodule path, Gitlink object ID, resolved commit,
  URL policy result, and captured tree digest;
- a canonical digest of the complete captured repository/package bytes and the
  resulting vendor package mapping;
- absence of symlink/special-node/path escapes, Git filters/LFS network
  materialization, dirty/untracked intake, and source-ID collisions.

A branch or tag may move; only the observed full commit plus captured tree is
an immutable build input. A Git URL or short revision alone is not. Git
dependencies that cannot be materialized without external filters, credentials
during build, missing submodules, unsupported object formats, or ambiguous
same-name/version vendor mappings are unsupported in v1.

### 6. Target and feature selection must be explicit checkpoint inputs

Cargo build features are additive and unified across uses of a dependency.
Optional dependencies join the active graph only when their feature is enabled;
`--no-default-features`, `--features`, and `--all-features` materially change
the compilation graph
([Cargo features](https://doc.rust-lang.org/cargo/reference/features.html)).
The lockfile does not preserve the chosen build feature vector, so Curator must
record both the requested, sorted feature set and the resolved feature list on
every active metadata node.

Target tables support target triples and Rust-like `cfg` expressions, but not
feature-dependent dependency selection. Target-specific build dependencies are
evaluated against the **host**, not the compilation target
([platform and build dependencies](https://doc.rust-lang.org/cargo/reference/specifying-dependencies.html#platform-specific-dependencies)).

`rust-source-v1` should therefore support only one native target where host and
target are equal. It must record the exact triple and normalized
`rustc --print cfg` result. It scans the all-target lock/vendor superset but
hashes the exact target-filtered active graph separately. Cross-compilation,
custom target JSON, multiple `--target` values, and host/target-split build code
remain unsupported until a dual-unit graph, target standard library, linker,
SDK/sysroot, and host-executable policy are implemented.

The command selection is exact: one package ID and one `bin` target. Default
workspace members, globs, `--workspace`, `--bins`, tests, benches, examples,
doctests, and `--all-targets` are not accepted aliases because Cargo documents
that they select different target sets and can activate development inputs
([cargo build target selection](https://doc.rust-lang.org/cargo/commands/cargo-build.html)).

### 7. `build.rs` and proc macros are executable build-time trust boundaries

Cargo compiles and executes `build.rs` before the package. The script may
perform arbitrary tasks, receives environment variables and the package root as
inputs, writes generated outputs, and can emit linker paths, libraries, flags,
configuration, and compiler environment values. `OUT_DIR` can persist between
builds
([Cargo build scripts](https://doc.rust-lang.org/cargo/reference/build-scripts.html)).

Procedural macros run code during compilation with the compiler's file and
standard-I/O access and carry the same security concerns as build scripts
([Rust procedural macros](https://doc.rust-lang.org/reference/procedural-macros.html)).

The local probe deliberately used both mechanisms:

- `build.rs` read an ambient file named by `CURATOR_UNDECLARED_FILE`, generated
  Rust source under `OUT_DIR`, and exported a compiler environment value;
- a proc macro read `CURATOR_PROC_INPUT` and embedded it into the output;
- `cargo build --frozen --release` succeeded, and the binary printed both
  ambient values.

This is direct evidence that Cargo's offline mode is dependency-fetch control,
not an execution sandbox. `rerun-if-changed` and
`rerun-if-env-changed` are post-execution rebuild hints, not predeclared access
control. Cargo JSON reported the local proc-macro `.dylib`, build-script
executable, `build-script-executed` event, generated `out_dir`, and final
binary only after those inputs had already been consumed.

The initial supported profile must detect active target kind `custom-build`,
active dependency kind `build`, target/crate type `proc-macro`, and package
`links` before compilation and reject them. A future hook-capable profile needs
all of the following before admission:

- the entire closure and toolchain as the only readable mounts;
- no network for Cargo or descendants;
- a fixed environment and denial of undeclared environment reads;
- allowlisted child executables and toolchain roots;
- an empty `OUT_DIR`/target root and a declared, verified write set;
- captured generated-file hashes and generator lineage;
- validated linker/search/flag/cfg/environment instructions;
- time/random/host-probe policy and deterministic replay;
- an audit record for host build order and every generated/loaded local output.

Until that exists, accepting a popular crate merely because its hook source was
vendored would violate the source-closure invariant.

### 8. Cargo configuration, environment, and toolchain are part of resolution

Cargo searches `.cargo/config` and `.cargo/config.toml` from its **current
directory** through every ancestor and then reads `$CARGO_HOME/config.toml`.
The configuration can replace sources, inject paths/patches, select rustc,
wrappers, linkers, targets, flags, credential providers, runners, and
environment variables
([Cargo configuration](https://doc.rust-lang.org/cargo/reference/config.html)).
Cargo also reads environment overrides such as `CARGO_HOME`, `RUSTC`, and both
rustc-wrapper variables
([Cargo environment variables](https://doc.rust-lang.org/cargo/reference/environment-variables.html)).

The fixture reproduced one non-obvious consequence: a locked/offline metadata
command invoked outside the workspace with only `--manifest-path` failed with
exit 101 because it did not discover the workspace's vendor replacement. The
same command run with the workspace as the current directory succeeded. The
adapter must therefore bind the canonical working directory and must not assume
that `--manifest-path` changes configuration discovery.

V1 should reject package-provided Cargo configuration and construct a private
staging hierarchy with no ancestor config. It should invoke physical selected
`cargo` and `rustc` binaries, not rustup proxies. A checked-in
`rust-toolchain.toml` is a source declaration to validate against the already
selected toolchain; it must not cause rustup to install or update anything.
Rustup itself documents directory/toolchain-file override precedence and the
fact that toolchain files can request components and targets
([rustup overrides](https://rust-lang.github.io/rustup/overrides.html)).

The Rust external-toolchain checkpoint must include:

- physical Cargo and rustc relative paths, bytes, and verbose version output;
- Cargo release/commit/host and linked libgit2/libcurl/TLS identities;
- rustc release, full commit, host, LLVM version, and the canonical full sysroot
  tree digest;
- installed target `rust-std` bytes and target libdir; Rustup documents a
  separate standard-library component per target
  ([rustup components](https://rust-lang.github.io/rustup/concepts/components.html));
- manager-selected linker, archiver where applicable, platform SDK/sysroot,
  runtime ABI/minimum OS, and their complete content identities;
- exact target triple/cfg, controlled rustflags/link args/remap rules, Cargo
  profile, and environment/config digests;
- containment/link validation and a time-of-use fingerprint recheck.

This mirrors the current Go baseline's domain-separated full-tree toolchain
fingerprint and link containment (`internal/godriver/fingerprint.go:19-118,
149-223`), while extending it for Rust's external linker and target standard
library. A version string alone is not sufficient.

### 9. The compiled-artifact prohibition applies before Cargo sees a package

The accepted common policy is authoritative. Every raw `.crate` (a compressed
source container), safely materialized Git/submodule tree, path package, and
root source tree is a `dependency_input` and receives recursive inspection at
the pre-vendor gate. The generated vendor directory receives the same
inspection plus transform verification at the post-vendor gate, before metadata
or build. `cargo vendor` is the sole Cargo process between those gates and is
never spawned for a rejected origin. Cargo package labels, lock checksums,
`.cargo-checksum.json`, file suffixes, and manifest target kinds cannot override
a byte-level denial.

In particular, dependency payloads containing native executables, `.o`, `.a`,
`.so`, `.dylib`, `.dll`, `.lib`, `.rlib`, `.rmeta`, proc-macro binaries,
compiler caches/IR, WebAssembly, or compiled content nested in another archive
fail with the shared `artifact_compiled_dependency_forbidden` code. Unknown,
encrypted, unsafe, malformed, or only partially inspected members fail with the
corresponding shared diagnostic.

Trusted Cargo/rustc/linker/SDK binaries are `external_toolchain`, and `.rlib`,
`.rmeta`, proc-macro libraries, objects, build-script executables, and the CLI
produced from admitted source in the clean target namespace are
`local_build_output`. The latter are allowed only after causal build evidence
and protected publication. A package cannot acquire either role by putting a
binary under `vendor/toolchain/` or `target/`.

## Recommended `rust-source-v1` contract

### Supported input profile

| Area | Supported in v1 | Fail-closed boundary |
| --- | --- | --- |
| Product | Exactly one named Cargo package and one named `bin` target | Workspace/default/glob target selection, tests, examples, benches, doctests, libraries as installed products |
| Sources | Authored/approved generated Rust and source-like text in a frozen root/path snapshot; recursively inspected registry/Git sources | Any shared compiled/opaque/unsafe class; links and special nodes |
| Registry | crates.io and explicitly approved registries with captured index record, pre-admitted raw archive, matching SHA-256, and verified transform/vendor manifest | Missing/mismatched checksum, unsupported registry/auth/source kind, unresolved mutable intake, or unexplained transform result |
| Git | Full lock commit, captured object/tree and recursive submodule identities, pre-admitted clean materialization, deterministic package projection/manifest normalization, and verified vendor source | Missing full commit/tree/submodule, Git filters/LFS/external helper, source collision, dirty/escaping tree, ambient file-selection input, or unexplained transform result |
| Path/workspace | Every resolved path contained in one frozen source boundary and hashed recursively | Absolute/out-of-bound path, symlink/special node, mutable host path |
| Features | Explicit `default_features` boolean plus sorted named features; resolved feature set hashed per metadata node | Implicit operator defaults or resolved/requested mismatch |
| Target | One exact native triple equal to the toolchain host; target-filtered active graph plus all-target scanned superset | Cross/custom/multiple targets, missing target stdlib, target/cfg drift |
| Build-time code | Declarative `macro_rules!`, ordinary Rust compilation, in-closure `include*` under sandbox | `build.rs`, build deps, proc macros, Cargo plugins, wrappers, generated external inputs |
| Native/system | Manager-selected linker and target SDK/sysroot as external toolchain | Package `links`, package-selected linker/search path, arbitrary host library, C/assembly/native build |
| Cargo channel | Pinned stable Cargo/rustc with metadata format 1, supported lock format, and allowlisted vendor-transform descriptor | Unrecognized Cargo vendor implementation/normalizer, nightly/unstable `-Z`, single-file manifests, artifact dependencies, unknown source/target kind |

### Closure and build procedure

1. **Freeze root identity.** Capture the source revision/snapshot before Cargo
   runs. Reject links/special nodes and exclude any pre-existing output/cache
   namespace from the build input.
2. **Validate declarations.** Parse all workspace manifests, exact command
   package/bin, `rust-toolchain*`, lockfile, feature vector, target, and every
   `[patch]`/replacement declaration. V1 accepts manifest patches only when the
   resolved source is otherwise supported and contained; package/host Cargo
   source configuration is rejected.
3. **Acquire immutable origins without build extraction.** In an
   acquisition-only environment, verify the lock without updating it; capture
   the exact registry index records/raw archives and exact Git
   objects/submodules named by the lock. Do not use `cargo vendor` or a Cargo
   unpacked source cache as the acquisition authority.
4. **Pre-admit all source origins.** Recursively inspect each raw `.crate`
   stream, safely materialized Git/submodule tree, and frozen root/path snapshot
   in canonical order under the shared artifact policy. Derive protected origin
   manifests, Git package projections, and the expected
   `cargo-vendor-transform-v1` ledger. Any failure returns before a Cargo vendor
   or build process, destination directory, target directory, or cache receipt
   exists.
5. **Stage only admitted capture.** Populate a new private Cargo home with the
   exact admitted registry archives/index records and Git worktrees. Make live
   caches/remotes and inherited Cargo/Git configuration unavailable; bind the
   staging manifest and read allowlist.
6. **Vendor under confinement.** Into an absent destination, run the selected
   physical Cargo as `cargo vendor --locked --offline --versioned-dirs` under OS
   network denial and write confinement. `--no-delete`,
   `--respect-source-config`, fallback acquisition, and pre-existing hidden or
   checksummed vendor content are forbidden. Independently generate the one
   expected source-replacement config.
7. **Post-verify the transform.** Before any metadata/build invocation,
   recursively scan the vendor output, compare its package/leaf manifest and
   config byte-for-byte with `cargo-vendor-transform-v1`, and require every
   origin leaf to have exactly one allowed disposition. Deny dominates; an
   unexplained addition, omission, mutation, normalization, generated metadata
   value, or mapping cardinality fails closed.
8. **Resolve both graphs.** Parse the lock-superset and run metadata once
   unfiltered and once for the exact native target/features. Normalize absolute
   paths. Require every active/superset package to map to exactly one captured
   origin and every remote lock entry to map to exactly one inspected vendor
   package.
9. **Reject unsupported units.** Before build, reject active build scripts,
   build dependencies, proc macros, native `links`, non-bin product selection,
   unknown target/source kinds, and untrusted config/toolchain selectors.
10. **Prove offline reconstruction.** From a fresh private `CARGO_HOME`, empty
   target/temp roots, and unavailable acquisition caches/original remotes, run
   target-filtered metadata and the exact `cargo build --frozen` command inside
   the OS sandbox. `--frozen` is Cargo's `--locked + --offline` combination
   ([cargo vendor options](https://doc.rust-lang.org/cargo/commands/cargo-vendor.html#manifest-options)).
11. **Verify outputs and publish.** Consume Cargo JSON, require the one expected
   executable and only allowed local intermediates, inspect the final binary,
   hash all outputs, recheck toolchain/input identities and write set, and
   publish through the protected cache with an exact receipt.

### Cargo build profile

Use a manager-owned custom profile whose effective settings are checkpoint
inputs. A reasonable initial definition is:

```toml
[profile.curator]
inherits = "release"
debug = 0
incremental = false
strip = "none"
```

Invoke it with an exact package/bin, target, and feature vector. Cargo/rustc
defaults not overridden above remain bound through the exact Cargo/rustc
toolchain identity; Curator should record the normalized effective profile and
all compiler/linker flags. Use stable virtual staging paths and controlled
`--remap-path-prefix` settings. Do not claim cross-toolchain or cross-path
byte-for-byte reproducibility until a dedicated reproducibility suite proves
it; the required v1 claim is exact-input, offline, protected local production.

## Audit checkpoint inputs

### `rust-closure-manifest-v1`

- adapter/profile/schema and shared artifact-policy/detector/limit identities;
- root build-source identity and canonical workspace/path tree manifest;
- exact `Cargo.toml` files, `Cargo.lock` bytes/version/digest, resolver version,
  and every patch/replacement declaration;
- `cargo-vendor-transform-v1` descriptor, pinned Cargo implementation commit,
  package-projection and Git-manifest-normalizer implementation identities,
  registry basename-wide `.cargo-ok` and Git exact-root reserved-name rules,
  projection-mode/checksum-encoding rules, and source-replacement config
  bytes/digest;
- pre-vendor admission receipts/digests for every root/path snapshot, raw
  registry archive, and complete materialized Git/submodule tree, plus evidence
  that the Cargo invocation ledger was empty until all of them admitted;
- for each lock package: name/version/source/checksum/dependency references and
  its unique captured-origin and `name-version` vendor-directory mapping;
- for each registry package: normalized registry/source ID, exact captured index
  record and digest, yanked observation, raw `.crate` size/SHA-256, recursive
  archive artifact/leaf manifests, every exact-root reserved and every
  basename-matching `.cargo-ok` omission with rule priority, per-origin-leaf
  transform disposition,
  expected vendor manifest/digest, observed post-scan vendor manifest/digest,
  exact generated `.cargo-checksum.json` bytes/size/SHA-256, and transform
  verdict;
- for each Git package: declared URL/selector, Cargo source ID, full precise
  commit, commit/tree object identity, package subpath, recursive submodule
  ledger, complete captured tree manifest/digest, deterministic package
  projection mode (`git_index_no_include` or `filesystem_include`), manifest
  include/exclude and controlled-ignore inputs, tracked index entries/stages
  under package-root `target/`, clean-worktree verdict, exact normalizer file
  list including any separately bound Cargo readiness marker, per-origin-leaf
  disposition, original and expected normalized `Cargo.toml` bytes/digests,
  expected and observed vendor manifests/digests, exact generated checksum
  metadata, and transform verdict;
- for every root/path package: canonical contained path and recursive content
  manifest/digest;
- exact confined vendor command/Cargo-home capture digest, destination identity,
  network/read/write evidence, deterministic diagnostics, and proof that
  metadata/build had not started before post-vendor admission;
- all lock-superset packages including inactive feature/target/dev entries;
- all deterministic diagnostics and the canonical closure-manifest digest.

### `rust-build-checkpoint-v1`

- closure-manifest digest and normalized unfiltered plus active Cargo metadata
  digests;
- selected package ID, bin target, manifest path, workspace root, exact native
  target triple/cfg, profile, default-feature boolean, requested sorted features,
  resolved features per node, dependency edges/kinds/target predicates, and
  manager-derived build order/host-target role;
- Cargo/rustc/sysroot/target-stdlib/linker/SDK identities and time-of-use recheck;
- exact physical commands, canonical working directory, manager config digest,
  allowed environment, rustflags/link flags/path-remap rules, private
  `CARGO_HOME`, clean target/temp roots, network=`none`, sandbox/process/read/
  write policy, and unsupported-hook decisions;
- expected output path/class and complete logical input digest.

### `rust-build-receipt-v1`

- complete checkpoint and closure/graph digests;
- normalized Cargo JSON artifact events and observed local intermediates;
- verified write set and evidence that source/vendor/toolchain inputs stayed
  read-only and unchanged;
- final binary canonical path, class, target/ABI facts, size and SHA-256;
- dynamic import/runtime compatibility observations where applicable;
- canonical receipt digest and protected-cache publication/validation result.

As in the Go baseline, reuse must start from an independently derived expected
input and require a protected exact hit. The current Go cache validates exact
receipt input plus artifact size/hash (`internal/buildcache/cache.go:104-126,
173-215`), and its receipt hash is expressly a consistency identifier rather
than provenance (`internal/buildmeta/models.go:107-121`). Rust must not weaken
that boundary.

## Stable diagnostics

Shared artifact diagnostics retain their accepted meanings and take precedence
when they identify the primary failure. Rust-specific codes should be stable
lowercase snake case and carry structured fields:

| Code | Required meaning / fields |
| --- | --- |
| `rust_lock_required` | `Cargo.lock` absent or unsupported; workspace/package, expected path |
| `rust_lock_mismatch` | `--locked` would modify resolution or lock/package/source mapping is inconsistent; package/source and Cargo error category |
| `rust_registry_identity_invalid` | Index/lock/raw-archive identity mismatch, forbidden generated name, wrong exact-root/basename-wide omission, copied-leaf mutation, or registry checksum-metadata transform mismatch; registry/package, transform ID/disposition, path, origin/expected/observed hashes |
| `rust_git_identity_invalid` | Missing/ambiguous full commit, object/tree/submodule mismatch, dirty/untracked intake, unsupported ambient filter/helper, wrong projection branch or tracked-root-`target/` decision, readiness-marker collision, copied-leaf mutation, normalized-manifest mismatch, or Git checksum-metadata mismatch; URL/selector/commit/package, transform ID/disposition, branch, path and hashes |
| `rust_path_dependency_escape` | Workspace/path manifest or source path is outside the frozen boundary or crosses a link/special node; package and canonical path |
| `rust_vendor_transform_unsupported` | Selected Cargo implementation/package selector/manifest normalizer/checksum encoding has no accepted transform descriptor; Cargo version/commit and supported descriptor IDs |
| `rust_vendor_incomplete` | Remote lock/vendor package mapping is zero/multiple, a vendor package/config has no origin, an origin leaf lacks exactly one allowed disposition, or an unexplained leaf/directory/config addition or omission exists; source/package, transform ID and canonical paths/cardinalities |
| `rust_graph_incomplete` | Metadata missing/malformed, unknown security-relevant field/kind, unreachable package mapping, or unfiltered/active graph contradiction |
| `rust_feature_profile_mismatch` | Requested/default and resolved feature vectors differ from the checkpoint; package, requested/resolved sets |
| `rust_target_unsupported` | Non-native, custom, multiple, unavailable, or mismatched target; host, target, target-stdlib evidence |
| `rust_build_script_unsupported` | Active `custom-build` target or build dependency in `rust-source-v1`; package, target, source path |
| `rust_proc_macro_unsupported` | Active `proc-macro` target/crate in `rust-source-v1`; package, target, host role |
| `rust_native_link_unsupported` | Package `links`, package-selected native linker/search input, or unapproved system library/FFI edge; package and declaration/evidence |
| `rust_config_untrusted` | Package/ancestor/home Cargo config, wrapper, runner, linker, patch/source override, alias, or environment selector is outside the manager policy |
| `rust_undeclared_input` | Sandbox denied an undeclared read, environment value, process, network operation, or write; package/unit, operation, canonical resource class |
| `rust_offline_rebuild_failed` | The completely captured, previously admitted closure cannot pass exact fresh-home metadata/build under `--frozen`; stage and Cargo status/category |

Relevant shared codes include `artifact_origin_unverified`,
`artifact_compiled_dependency_forbidden`, `artifact_opaque_dependency_forbidden`,
the archive/path/limit codes, `artifact_generated_input_undeclared`,
`artifact_toolchain_untrusted`, `artifact_toolchain_identity_changed`, and the
local-output receipt/drift codes. A compiled dependency finding remains primary
even if Cargo would ignore that file.

## Conformance fixtures

Every negative case asserts that no package hook/compiler/build command ran, no
output/cache entry was published, and the diagnostic identifies the package,
virtual path, and container/origin chain. A pre-vendor rejection also requires
`cargo_vendor_spawn_count = 0` and an absent vendor destination. A post-vendor
transform rejection permits only the confined `cargo vendor` process; metadata
and build invocation counts remain zero.

### Positive closure and graph cases

| ID | Fixture | Required proof |
| --- | --- | --- |
| `R01` | Binary -> registry crate -> transitive registry crate | Lock/index/raw archive/vendor manifests enumerate both levels; fresh-home frozen build passes |
| `R02` | Git package at branch/tag declaration, lock-pinned full commit, nested package, recursive submodule | Commit/tree/submodule identities and vendor mapping are exact; build passes after original Git remote/cache is unavailable |
| `R03` | Root -> contained path package -> contained transitive path package | Every metadata path is inside the frozen snapshot and every file is in the closure manifest |
| `R04` | Optional dependency disabled, then enabled by explicit feature | Same lock-superset digest; distinct active graph/checkpoint/output digests; resolved node features match each command |
| `R05` | Mutually exclusive Unix/Windows normal dependencies | Both appear in lock/scanned superset; only exact native target appears in active graph/build |
| `R06` | Multi-member workspace with one exact package/bin | Only declared product target builds; lock still covers the workspace acquisition superset |
| `R07` | Pure Rust in-closure `include!`, `include_str!`, and `include_bytes!` | Read succeeds from read-only closure; included leaves are in the manifest and receipt input |
| `R08` | Fresh private home, empty target, remote/original Git unavailable, OS network denied | Metadata and build pass with `--locked --offline`/`--frozen`; no acquisition cache read |
| `R09` | Repeat build in canonical clean paths | Logical checkpoints match; outputs and any tolerated nondeterminism are explicitly measured, never assumed |

### Vendor ordering and transformation cases

| ID | Exact fixture | Required assertion |
| --- | --- | --- |
| `RV01` | Retained `itoa 1.0.15` archive plus fresh empty-destination regeneration: raw `.gitignore` SHA-256 `07d64e…14e17`, raw/vendor `Cargo.toml` `c1d45a…667bc`, package checksum `4a5f13…928e2c`, generated metadata `35abe1…c81d` | Pre-scan covers the whole raw archive; ledger records only root `.gitignore` as `omit_reserved`, every other source leaf as `copy_identical`, and only `.cargo-checksum.json` as `generate_checksum`; independent expected and observed manifests match. |
| `RV02` | Retained edge registry archive `e1d506…c1c6f` with root `.gitattributes`, `.gitignore`, `.git`, `.cargo-ok`, `nested/.cargo-ok`, and `nested/.gitignore`; exact lock/index checksum matches; generated metadata `ae5e67…6ac80` | Pre-scan covers every member. Root Git-control names are `omit_reserved`; root and nested `.cargo-ok` are `omit_registry_cargo_ok`; nested `.gitignore` is `copy_identical`. Expected vendor leaf set and exact canonical checksum `files` map/`package` bytes equal the observed output. |
| `RV03` | `RV02` plus an origin `.cargo-checksum.json`, or a transform that treats nested `.cargo-ok` as copied or applies basename omission to nested `.gitignore` | Reject before vendor for the reserved generated name, or after confined vendor for the wrong disposition/leaf set, with `rust_registry_identity_invalid`; metadata/build counts remain zero. |
| `GV01` | Retained `git_leaf 0.1.0`: original/normalized manifest hashes `22df0b…2c192` -> `e0cf59…df3`, unchanged `src/lib.rs` `e7afc8…34703`, generated metadata `152f27…0ddb5` | Original manifest maps once to exact normalized bytes; source leaf copies exactly; `package` is null; no automatic `Cargo.toml.orig` exists; frozen rebuild passes without original remote/cache. |
| `GV02` | Retained `TASK-260810-3urqbl-vendor-mapping-probe`: workspace inheritance; selected `.gitignore`, `.gitattributes`, `.cargo-ok`; unselected `omitted.txt`; generated manifest `778cb0…2e73` and metadata `be88ee…9a39` | Projection records all four omissions with `omit_reserved` or `omit_unselected`; `kept.txt` and `src/lib.rs` copy byte-identically; normalized manifest resolves version/edition and inferred target fields; fresh-home frozen build exits 0. |
| `GV03a` | Retained no-`include` Git commit `2253ef8…fe29d`: commit leaves `.gitignore`, `Cargo.toml`, `src/lib.rs`, and `target/tracked.txt`; Cargo staging adds its exact empty root `.cargo-ok`; normalized output `Cargo.toml` `5b1aec…2ba40`; metadata `3937bb…2f75` | Select `git_index_no_include`; immutable projection and normalizer input file sets are recorded separately, with the staging marker explicit. `.gitignore`/marker omit at exact root, committed `target/tracked.txt` copies as `53d5d5…6b48e`, and the exact vendor set is normalized manifest, `src/lib.rs`, tracked target leaf, and checksum metadata. |
| `GV03b` | `GV03a` capture polluted with nonignored `target/untracked.txt` `721d5b…1f4a2` and ignored `target/ignored.txt` `79811f…0ed0` | Cargo's oracle omits both, but `rust-source-v1` rejects dirty/untracked intake with `rust_git_identity_invalid` during pre-vendor admission; `cargo_vendor_spawn_count = cargo_metadata_spawn_count = cargo_build_spawn_count = 0`, destinations/receipts absent. |
| `GV03c` | Retained explicit-`include` Git commit `972c661…eb159`: `include = ["Cargo.toml", "src/**", "target/**"]`, committed `target/included.txt` `6ac0e7…49f6`, normalized manifest `747ef6…2b19`, metadata `d47213…dda6` | Select `filesystem_include`; exact normalizer input is `Cargo.toml` plus `src/lib.rs`. Root `target/included.txt` is `omit_unselected` despite matching `target/**`; exact vendor set is normalized manifest, `src/lib.rs`, and checksum metadata. |
| `GV03d` | `GV03a` with a valid shared-classifier native object/executable, `.rlib`, `.rmeta`, or Wasm member committed under package-root `target/`, including renamed/suffixless variants | The complete commit tree is classified before projection; shared `artifact_compiled_dependency_forbidden` wins with the tracked path/hash/class, `cargo_vendor_spawn_count = cargo_metadata_spawn_count = cargo_build_spawn_count = 0`, and no vendor/target/cache output. This proves tracked-target re-addition is not a compiled-artifact bypass. |
| `GV03e` | Git monorepo package subpath plus recursive submodule, controlled repo ignore file, nested subpackage, and nested non-root `target` | Complete commit/submodule trees are pre-admitted; every leaf has a projection disposition; only selected package/submodule leaves map into its unique vendor directory; no global/system Git input affects selection. |
| `VF01` | `RV01` with one copied leaf changed and the generated per-file hash forged to match; separately change `package`, key order/encoding, or insert an extra key | Cargo may accept the self-consistent directory, but independent expected bytes differ: `rust_registry_identity_invalid`; metadata/build counts zero. |
| `VF02` | `GV02` with copied `kept.txt` changed and hash forged (observed Cargo build exits 0), or with altered normalized `Cargo.toml` plus forged hash | Curator compares protected origin/normalizer output, returns `rust_git_identity_invalid`, and never invokes metadata/build. |
| `VF03` | Remove an expected leaf, add an unaccounted leaf or `Cargo.toml.orig`, duplicate one lock package directory, add an extra package/config entry, or map two origins to one path | `rust_vendor_incomplete` with exact path/cardinality and transform ID; metadata/build counts zero. |
| `PV01` | Raw `.crate` containing a direct or nested compiled/opaque/unsafe leaf; vendor and target destinations start absent and Cargo invocations use a recording executor | Shared artifact diagnostic at pre-vendor admission; `cargo_vendor_spawn_count = cargo_metadata_spawn_count = cargo_build_spawn_count = 0`; both destinations and all receipts remain absent. |

### Resolution, origin, and containment rejection cases

| ID | Fixture | Expected primary code |
| --- | --- | --- |
| `RF01` | Missing `Cargo.lock` | `rust_lock_required` |
| `RF02` | Manifest change requires lock update under `--locked` | `rust_lock_mismatch` |
| `RF03` | Registry archive hash differs from lock/index | `rust_registry_identity_invalid` or shared `artifact_origin_unverified` by global precedence |
| `RF04` | Vendored leaf and `.cargo-checksum.json` are forged together | `rust_registry_identity_invalid`; Curator's protected origin/leaf manifest catches what Cargo accepts |
| `RF05` | Git selector resolves differently from lock, object missing, submodule changed/missing | `rust_git_identity_invalid` |
| `RF06` | Git LFS/filter/helper needed to materialize bytes | `rust_git_identity_invalid` before build |
| `RF07` | Path dependency escapes root, crosses link, or mutates after checkpoint | `rust_path_dependency_escape` or shared origin/toolchain drift code as applicable |
| `RF08` | Remote lock package missing/duplicated in vendor; extra vendor package | `rust_vendor_incomplete` |
| `RF09` | Unknown Cargo metadata source/target kind or path outside captured roots | `rust_graph_incomplete` |
| `RF10` | Requested feature or target differs from checkpoint | `rust_feature_profile_mismatch` / `rust_target_unsupported` |
| `RF11` | Workspace `.cargo/config`, ancestor/home config, `RUSTC_WRAPPER`, runner, custom linker, or package source replacement | `rust_config_untrusted` |
| `RF12` | Custom target JSON, cross-target, multiple targets, nightly `-Z`, artifact dependency | `rust_target_unsupported` or `rust_graph_incomplete` with unsupported feature detail |

### Build-time code and artifact rejection cases

| ID | Fixture | Expected primary code |
| --- | --- | --- |
| `RH01` | Root or transitive `build.rs`, including renamed `package.build` | `rust_build_script_unsupported` before compilation |
| `RH02` | Active build dependency, including target-specific host dependency | `rust_build_script_unsupported` before compilation |
| `RH03` | Direct/transitive proc macro that reads file/env/network | `rust_proc_macro_unsupported` before proc-macro library is built/loaded |
| `RH04` | Package `links`, `#[link]` resolving outside approved SDK, or native search path | `rust_native_link_unsupported` / `rust_undeclared_input` |
| `RH05` | Ordinary Rust `include*` escapes readable closure | `rust_undeclared_input` from sandbox; no output publication |
| `RH06` | Build/proc hook under a future enabled hook profile reads host file, network, clock/random, undeclared env, or writes outside output | `rust_undeclared_input`; proves future profile enforcement, not v1 admission |
| `RH07` | `.rlib`, `.rmeta`, object, static/dynamic library, executable, Wasm, or compiler cache direct/renamed/nested in `.crate`, Git, path, or root payload | `artifact_compiled_dependency_forbidden` with shared class and container chain |
| `RH08` | Pre-existing compiled bytes planted in target/output before build | `artifact_local_output_unreceipted` |
| `RH09` | Toolchain binary copied into dependency `vendor/toolchain` | `artifact_compiled_dependency_forbidden`; location cannot establish trust |
| `RH10` | Toolchain/linker/target stdlib changes between checkpoint and build | `artifact_toolchain_identity_changed` |

## Implementation-ready backlog

1. Define strict `rust-closure-manifest-v1`, `rust-build-checkpoint-v1`, and
   `rust-build-receipt-v1` models plus `cargo-vendor-transform-v1` descriptors,
   per-leaf dispositions, expected/observed manifests, and canonical encodings
   beside the shared adapter/build metadata contract.
2. Implement Cargo lock/manifest/metadata parsing with closed source, target,
   dependency-kind, feature, and target-kind enums; canonicalize all paths and
   graph edges.
3. Implement non-extracting registry acquisition and safe Git/submodule capture
   for approved sources, retaining exact raw archive/index/object/tree evidence;
   do not invoke Cargo vendoring during this stage.
4. Apply the shared artifact classifier to every root/path snapshot, raw archive
   stream, and safely materialized Git/submodule tree—including inactive
   lock-superset packages—and produce protected origin manifests before any
   Cargo process.
5. Implement independent registry and Git transform derivation: Cargo 0.92
   registry basename-wide `.cargo-ok` omission, Git exact-root filters,
   `git_index_no_include` tracked-root-`target/` re-addition,
   `filesystem_include` root-target exclusion, controlled ignore/index inputs,
   normalized Git-manifest bytes, sorted checksum JSON, unique package mapping,
   and deterministic mismatch diagnostics.
6. Populate a private Cargo home only from admitted captures; run confined
   offline vendoring into an absent destination; then run shared inspection and
   exact expected/observed transform comparison before metadata/build.
7. Implement root/path dependency containment and mutation rechecks using the
   same handle-relative, deny-special-node posture as the Go source/toolchain
   boundaries.
8. Implement Rust toolchain selection/fingerprinting for Cargo, rustc, sysroot,
   target stdlib, linker and SDK; invoke physical binaries and recheck at use.
9. Implement pre-build profile rejection for custom builds, build dependencies,
   proc macros, native links, config/wrappers, unsupported targets, and unstable
   Cargo features.
10. Implement the fresh-home frozen metadata/build worker with OS network/process/
   read/write isolation, JSON event validation, clean outputs, and protected
   cache publication.
11. Materialize `R01`-`RH10`, `RV01`-`RV03`, `GV01`-`GV03e`, `VF01`-`PV01`,
    and all shared artifact vectors as
    shared byte fixtures plus Rust harness cases. Use a recording executor to
    prove the pre-vendor zero-spawn boundary—including compiled tracked
    package-root `target/`—and retain forged checksum, normalization, exact
    root/basename omission, both PathSource modes, duplicate-mapping, and
    wrong-working-directory regressions.
12. Keep hook/proc-macro, native/FFI, and cross-compilation support in separate
    capability tasks gated by their own sandbox and checkpoint conformance.

## Empirical fact-check record

The original experiment lives under ignored
`.temp/TASK-260810-3urqbl-rust-fixture/`; the R1/R2 mapping probe lives under
`.temp/TASK-260810-3urqbl-vendor-mapping-probe/`; the R2a/R2b edge probe lives
under `.temp/TASK-260810-3urqbl-transform-edge-probe/`. The original used physical
Cargo/rustc binaries with a cleared environment, private `HOME`/`CARGO_HOME`,
explicit `RUSTC`, and only system tool paths. The mapping probe used the same
physical Cargo with task-specific private Cargo homes and target directories.
Commands were run directly; no validation command was piped through `tee` or a
status-masking pipeline.

| Command / probe | Exit | Evidence |
| --- | ---: | --- |
| `rustc --version --verbose` | 0 | rustc 1.91.0, commit `f8297e351a40c1439a467bbbb6879088047f50b3`, host `aarch64-apple-darwin`, LLVM 21.1.2 |
| `cargo --version --verbose` | 0 | Cargo 1.91.0, commit `ea2d97820c16195b0ca3fadb4319fe512c199a43`, host `aarch64-apple-darwin`, linked library identities reported |
| controlled `cargo generate-lockfile` for direct/feature/target/build/proc/Git/registry fixture | 0 | Version-4 lock initially captured all seven declared dependency packages, including inactive optional and opposite-target path inputs |
| `cargo vendor --locked --versioned-dirs …` | 0 | Vendored registry and Git sources and emitted exact source replacement; did not copy local path packages |
| fresh-home `cargo metadata … --locked --offline --no-default-features --filter-platform host-tuple` from the wrong current directory | 101, expected failure | Vendor config was not discovered; attempted Git source access was refused offline. This validates the documented current-directory config boundary. |
| same metadata command from workspace root | 0 | Active graph contained build dep, Git, registry, proc macro, and Unix path dep; excluded optional and Windows dep |
| metadata without platform filter | 0 | Resolve included both Unix and Windows target dependencies |
| metadata with `--features optional --filter-platform host-tuple` | 0 | Optional package joined the active graph and root features changed |
| first fresh-home `cargo build --frozen --release --bin closure-probe --no-default-features --message-format=json` | 0 | Built vendor Git/registry and path graph; JSON exposed proc-macro `.dylib`, build-script executable/event, local intermediates, and final binary |
| first built CLI execution | 0 | Printed `42:git:unix:disabled:ambient-build-input:ambient-proc-input`, proving build.rs and proc macro consumed ambient inputs despite frozen/offline Cargo |
| feature-enabled frozen build and CLI execution | 0 / 0 | Built optional path package and printed `optional`; lock remained the broader source superset |
| source-only vendor file tamper, fresh frozen build | 101, expected failure | Cargo reported expected and actual SHA-256 for the changed vendored source file |
| same file tamper plus forged per-file `.cargo-checksum.json`, fresh frozen build | 0 | Confirms Cargo directory checksums are not an independent malicious-tamper trust root |
| restored vendored file `shasum -a 256` | 0 | Restored hash `ef9f1a8665a678cf5b77bcaa628d00538d620de0c84fd2a8b92323a314a95636` matched Cargo vendor metadata |
| missing-lock locked/offline metadata probe | 101, expected failure | Cargo refused to use the vendored Git source without a lockfile |
| raw `itoa-1.0.15.crate` `shasum -a 256` | 0 | `4a5f13…928e2c`, equal to the lock and vendor `package` checksum |
| revised lock/vendor with `aho-corasick 1.1.3 -> memchr 2.8.3` | 0 / 0 | Locking reported nine dependency packages; lock recorded the recursive registry edge and checksums |
| fresh-home recursive `cargo build --frozen …` with original Git repository moved unavailable | 0 | Built `memchr`, `aho-corasick`, `itoa`, pinned vendored Git source, and path inputs solely from the captured closure plus toolchain |
| recursively rebuilt CLI execution | 0 | Same expected disabled-feature output; original Git repository was restored after the probe |
| `git ls-remote --tags` for Cargo `0.92.0` | 0 | Peeled tag commit `ea2d97820c16195b0ca3fadb4319fe512c199a43` equals the physical Cargo 1.91.0 commit |
| tagged Cargo 0.92 source checkout and `git rev-parse` / exact-tag verification | 0 each | Checkout resolved to `ea2d97820c16195b0ca3fadb4319fe512c199a43`, matching physical Cargo. `vendor.rs`, registry `mod.rs`, `path.rs`, Git checkout guard, and `include_files_called_target_git` were inspected: registry unpack is basename-wide for `.cargo-ok`; no-`include` Git selection re-adds unconflicted tracked root-target entries; include mode uses the root-target-skipping filesystem walk. The recoverable source clone was removed after inspection when the volume filled. |
| mapping-probe `cargo generate-lockfile` with private Cargo home | 0 | Locked the exact local Git commit `4e9694840035320e098afc819e556d0c405faea4` and resolved workspace-inherited package version 0.1.0 |
| mapping-probe `cargo vendor --locked --offline --versioned-dirs vendor` | 0 | Used only the populated private Cargo home; emitted the normalized Git manifest, two copied leaves, checksum metadata, and exact source-replacement config |
| fresh empty-destination regeneration of the original registry/Git closure with `cargo vendor --locked --offline --versioned-dirs` | 0 | Reproduced all four vendor packages; fresh `itoa` checksum metadata has no trailing LF and hashes to `35abe1…c81d` |
| `diff -rq` between the previously tampered/restored `itoa` vendor tree and fresh regeneration | 1, expected difference | Only `.cargo-checksum.json` differed: the earlier restoration left one trailing LF. That mutated copy is not canonical transform evidence. |
| mapping-probe origin/vendor `shasum -a 256` inventory | 0 | Recorded the full hashes in the transform table; reserved and unselected leaves were absent, copied leaves matched, and generated leaves had distinct bound hashes |
| first mapping-probe frozen build with `directory = "../vendor"` | 101, expected failure | The deliberately wrong directory-source path resolved outside the consumer vendor directory; no source root existed. This is not passing evidence. |
| corrected mapping-probe fresh-home `cargo build --frozen --release` | 0 | Built solely from the verified vendor directory and selected toolchain; no acquisition cache was present in the build home |
| corrected mapping-probe CLI execution | 0 | Printed `17` |
| forged mapping-probe `kept.txt` plus matching forged checksum, fresh-home frozen build | 0 | Cargo accepted the self-consistent forged Git directory source, proving generated metadata cannot replace Curator's protected transform expectation |
| first `verify_mapping.rb` gate | 1 | Correctly refused the previously restored `itoa` checksum bytes because of their extra LF; it was not treated as passing evidence and triggered fresh regeneration. |
| corrected `verify_mapping.rb` against fresh registry and Git output | 0 | Independently verified exact leaf sets/hashes, reserved/unselected omissions, normalized manifests, canonical checksum JSON, and protected-expectation divergence for forged metadata. |
| freshness check for `cargo-home-validation-a8f7f2` and `target-validation-a8f7f2` | 0 | Both validation roots were absent before creation. |
| validation-home `cargo build --frozen --release` for the mapping probe | 0 | Rebuilt from the directory source with a fresh Cargo home/target. |
| validation mapping-probe CLI execution | 0 | Printed `17`. |
| edge-probe Git lock generation for `2253ef8…fe29d` and `972c661…eb159`; registry lock generation | 0 each | Both exact Git commits were locked in private Cargo homes; registry lock was then bound with the synthetic archive and sparse-index checksum `e1d506…c1c6f`. |
| edge registry `cargo vendor --locked --offline --versioned-dirs vendor` | 0 | Root and nested `.cargo-ok` were absent, nested `.gitignore` copied as `15436d…3784`, and exact checksum bytes hashed to `ae5e67…6ac80`. |
| no-`include` and explicit-`include` Git `cargo vendor --locked --offline --versioned-dirs vendor` | 0 each | No-`include` output contained committed `target/tracked.txt`; explicit-include output omitted committed `target/included.txt`. Generated checksum bytes hashed to `3937bb…2f75` and `d47213…dda6`. |
| clean-checkout no-`include` vendoring and `diff -rq` against the ambient-file Cargo oracle | 0 / 0 | With only Cargo's empty `.cargo-ok` marker outside the commit, fresh output was byte-identical to the oracle output that Cargo had produced while untracked/ignored root-target siblings existed. Curator still rejects the dirty variant before vendor. |
| first three edge fresh-home `cargo build --frozen --release` attempts | 101 each, failure | All three failed with `No space left on device` while writing local outputs. This was a recoverable capacity failure, not Cargo closure evidence, and is not counted as passing. Only task-owned Cargo clone and generated target outputs were removed. |
| retried no-`include`, explicit-`include`, and registry fresh-home `cargo build --frozen --release` with both Git origins unavailable | 0 each | All three rebuilt from verified directory sources and the selected toolchain after capacity recovery. |
| three rebuilt edge CLI executions | 0 each | Printed `tracked-target`, `include-walk`, and `17`. |
| `verify_edge_mapping.rb` focused R2a/R2b gate | 0 | Independently asserted archive/commit identities, every relevant origin/vendor leaf, source-specific omissions, both normalized manifests, exact canonical checksum bytes/maps, Cargo readiness marker, clean/ambient oracle parity, and protected dirty-input distinction. |
| revised original `verify_mapping.rb` with basename-wide registry rule | 0 | Existing retained registry/Git mappings remain green under the corrected source-specific rule. |
| `go test ./internal/buildmeta ./internal/buildcache ./internal/godriver` | 0 | Latest focused run passed all three architecture-relevant Go baseline packages; `buildmeta`/`buildcache` were cached and `godriver` completed in 39.693 seconds. |
| current `validate_research.rb` nonempty, UTF-8, trailing-whitespace, fence-balance, focused R2a/R2b coverage, and warning-free local-link gate | 0 | Current document passed; every one of the three unique local link targets exists. External primary-source URLs are cited but deliberately not treated as filesystem paths. |
| `ruby -c` for the edge verifier, document validator, registry-identity rewriter, and revised original verifier | 0 each | All four retained Ruby scripts reported `Syntax OK`. |
| two superseded regex-based local-link checker attempts | 0 each, warnings | Both parsed/checked links but emitted Ruby regex warnings, so neither is used as accepted evidence; the warning-free checker above replaced them. |

The OS network interface was not disabled for these local probes. Cargo's
network behavior was constrained by `--frozen`, and the fresh home plus missing
original Git source made cache/source fallback observable. This is sufficient
to verify Cargo's offline reconstruction behavior, but **not** descendant
network isolation. The ambient-input probe affirmatively shows why the eventual
conformance harness must place Cargo and every descendant inside Curator's real
network/filesystem/process sandbox.

## Acceptance mapping

| Task requirement | Evidence in this document |
| --- | --- |
| Complete recursive source enumeration | Lock-superset plus active graph; pre-vendor raw archive/full Git/root/path manifests; per-leaf transform dispositions; recursive `aho-corasick -> memchr` probe; `R01`-`R03`, `RV01`-`GV03e` |
| Pre-extraction fail-closed boundary | Procedure steps 3-7 separate acquisition, pre-vendor admission, confined vendoring, and post-vendor verification; `PV01` requires recording-executor proof that rejected raw bytes yield zero Cargo spawns and no destination |
| Immutable identity | Registry index/lock/raw archive triple; Git commit/tree/submodule identity; pinned transform/normalizer; exact expected/observed manifests and generated metadata; protected path/toolchain fingerprints |
| Deterministic origin-to-vendor coverage | `cargo-vendor-transform-v1`, basename-wide registry `.cargo-ok`, both Git PathSource modes, tracked root-target re-addition, complete disposition tables, source-specific mismatch diagnostics, full retained hashes/checksum bytes, checkpoint inputs, and `RV01`-`VF03` |
| Offline rebuild | Fresh-home frozen builds, original Git unavailable probe, exact offline procedure and `R08` |
| Undeclared-input refusal | Direct ambient `build.rs`/proc-macro counterexample; v1 rejection boundary; required OS sandbox and `RH01`-`RH06` |
| Git dependencies | Full commit and source replacement observation; recursive submodule policy; exact tracked/untracked/ignored root-target and explicit-include branches in `GV03a`-`GV03e` |
| Features and target inputs | Lock/metadata distinction, observed optional/Unix/Windows behavior, native-target restriction, `R04`-`R05` |
| Build scripts and proc macros | Official execution semantics, empirical ambient read, explicit unsupported diagnostics and future capability gate |
| Toolchain identity | Physical Cargo/rustc/sysroot/target stdlib/linker/SDK checkpoint and Go-baseline alignment |
| Compiled-artifact prohibition | Shared policy applied before vendoring to raw/materialized origins and after vendoring to derived payloads; compiled tracked-root-`target/` zero-spawn case `GV03d`, `PV01`, and `RH07`-`RH10` |
| Diagnostics and conformance fixtures | Stable code table and `R01`-`RH10`, `RV01`-`RV03`, `GV01`-`GV03e`, `VF01`-`PV01`, plus shared artifact fixture matrices |

## Sources and fact-check limits

Facts were checked on 2026-08-11 against project-authoritative source and
official Rust/Cargo documentation:

- Project invariant and delivery scope:
  [repository source-closure specification](../.spec/skill-facing-cli-source-closure.md),
  especially lines 39-78 and 80-103.
- Existing Go logical input, toolchain, receipt, and protected-cache semantics:
  `internal/buildmeta/models.go`, `internal/godriver/fingerprint.go`, and
  `internal/buildcache/cache.go` at the line ranges cited above.
- Cargo resolution and lock breadth:
  [Dependency Resolution](https://doc.rust-lang.org/cargo/reference/resolver.html).
- Metadata graph/schema and target filtering:
  [cargo metadata](https://doc.rust-lang.org/cargo/commands/cargo-metadata.html).
- Vendoring, `--locked`, `--offline`, and `--frozen`:
  [cargo vendor](https://doc.rust-lang.org/cargo/commands/cargo-vendor.html).
- Cargo 1.91.0 vendor transformation implementation and regression tests:
  [Cargo 0.92.0 `vendor.rs`](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/ops/vendor.rs),
  [Cargo 0.92.0 registry unpacker](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/registry/mod.rs),
  [Cargo 0.92.0 `PathSource` file selection](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/path.rs),
  [Cargo 0.92.0 Git checkout readiness marker](https://github.com/rust-lang/cargo/blob/0.92.0/src/cargo/sources/git/utils.rs),
  [Cargo package root-target regression](https://github.com/rust-lang/cargo/blob/0.92.0/tests/testsuite/package.rs#L4132-L4193),
  and [Cargo vendor tests](https://github.com/rust-lang/cargo/blob/0.92.0/tests/testsuite/vendor.rs).
- Git/package include, exclude, ignore, subpackage, target, and mandatory-file
  selection rules:
  [Cargo manifest format](https://doc.rust-lang.org/cargo/reference/manifest.html#the-exclude-and-include-fields).
- Source replacement and directory checksum limitation:
  [Source Replacement](https://doc.rust-lang.org/cargo/reference/source-replacement.html).
- Registry checksum/index fields:
  [Registry Index](https://doc.rust-lang.org/cargo/reference/registry-index.html).
- Git commits/submodules, paths, target tables, and build dependency host rules:
  [Specifying Dependencies](https://doc.rust-lang.org/cargo/reference/specifying-dependencies.html).
- Feature activation and unification:
  [Features](https://doc.rust-lang.org/cargo/reference/features.html).
- Build-script execution, inputs, outputs, and directives:
  [Build Scripts](https://doc.rust-lang.org/cargo/reference/build-scripts.html).
- Proc-macro execution privileges:
  [Procedural Macros](https://doc.rust-lang.org/reference/procedural-macros.html).
- Cargo config hierarchy and environment overrides:
  [Configuration](https://doc.rust-lang.org/cargo/reference/config.html) and
  [Environment Variables](https://doc.rust-lang.org/cargo/reference/environment-variables.html).
- Build target/feature/profile/output selection:
  [cargo build](https://doc.rust-lang.org/cargo/commands/cargo-build.html).
- Rustup toolchain selection and target components:
  [Overrides](https://rust-lang.github.io/rustup/overrides.html) and
  [Components](https://rust-lang.github.io/rustup/concepts/components.html).

Normative recommendations are identified as such. The experiments establish
Cargo 1.91.0 behavior on `aarch64-apple-darwin`; they do not claim that arbitrary
Cargo projects are hermetic, that Cargo checksum files authenticate a vendor
tree, that `--offline` confines child code, or that Rust outputs are universally
bit-for-bit reproducible. Those stronger claims would contradict the checked
documentation and observed ambient-input behavior. The retained probes validate
the initial `cargo-0.92.0@ea2d978` transform facts only; another Cargo commit is
unsupported until its own descriptor and vectors pass. `PV01` and the other
adapter conformance rows are implementation requirements for the subsequent
delivery story, not a claim that the research story already contains a Rust
adapter or OS sandbox.
