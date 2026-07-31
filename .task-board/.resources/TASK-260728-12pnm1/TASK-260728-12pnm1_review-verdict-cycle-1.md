# TASK-260728-12pnm1 review verdict, cycle 1

## Verdict

CHANGES REQUESTED. Route to `analysis` for contract and evidence rework. This is ordinary recoverable design work, not a stop-the-line blocker.

## What passed

- The submitted decision, reference, results, fixtures, command evidence, and probe archive are task-scoped and readable.
- The board copies of decision 0009 and the Rust reference are byte-identical to the task worktree copies.
- Independent replay on macOS arm64 with Rust/Cargo 1.91.0: probe `go test ./...`, `go vet ./...`, and `go build ./...` exit 0. The full probe and `-controls` replay both exit 0 with 19 of 19 cases matched, P1 and P2 true, 24 closure checks yielding no verdict, all 6 controls failing as required, and all 8 structural checks matching.
- Curator `go test ./...` and `go vet ./...` exit 0.
- The `rust-version` change from `compared` to the seven-class `classified` boundary is a compatible narrowing owned by this reserved driver contract. The measured TOML, grammar, edition-floor, and separate host-gate layers plus the mandatory catch-all satisfy the accepted decision 0007 P1/P2 model.
- The measured macOS `cargo -> rustc -> rust-lld` closure, data-only SDK role, honest macOS-arm64-only registry scope, unqualified Windows/Linux boundaries, and deferred signing boundary are coherent.

## Blocking findings

### 1. `curator-rust-toolchain-v1` rejects its own normative `rustc -vV` input

Reference section 2.2 says the `V` record carries the normalized `rustc -vV` stdout, then requires exactly one terminal LF and rejects every other CR or LF. Reference section 1.1 and the fixture show that valid Rust 1.91.0 `rustc -vV` stdout is multiline. Therefore no conforming implementation can compute the shown identity from the admitted toolchain.

Required rework: define the exact `V` payload unambiguously. Either bind only the already extracted whole `release:` line, or define a multiline normalization and framing rule. Update the identity example, algorithm prose, probe, and deterministic positive and mutation vectors so the byte stream is independently reproducible.

### 2. The graph phase is not the claimed union over every feature

The fixed graph argv in reference section 4 is `cargo metadata --format-version 1 --locked --offline --color never --quiet`; it does not pass `--all-features`. Cargo 1.91.0 exposes `--all-features` as the operation that activates all available features, and Cargo documents that optional dependencies are not activated by default. Reference section 6.1 nevertheless says `packages[]` is the union over every platform and every feature, and conformance vector 19 requires a proc macro reachable only through a non-default feature to be rejected by that union graph. That vector cannot follow from the normative argv.

Required rework: choose and state one closed semantic. Either add `--all-features` to the graph argv and retain the deliberate over-rejection, or scope the matrix to the exact default-feature compile graph and remove the false all-feature claim/vector. In either branch, prove that every package unit the compile command can execute is present in the graph checked before the permit.

Primary references: https://doc.rust-lang.org/cargo/commands/cargo-metadata.html and https://doc.rust-lang.org/cargo/reference/features.html.

### 3. G6/G7 reject outside-root dependencies only after Cargo has resolved them

The contract says only the selected build root is compiler-visible and input must not come from a sibling, parent, or host path. Yet G6/G7 decide an outside-root path dependency from `cargo metadata` output `manifest_path` and `src_path`. Cargo path dependencies may use `../...` and must point to the external package directory containing its `Cargo.toml`; Cargo must read and resolve that manifest before it can emit the path that G6 rejects. An absolute package path has the same problem. The post-resolution check prevents compilation but does not prove that the graph phase cannot read or resolve outside the admitted root.

Required rework: add a pre-graph, snapshot-derived path/source closure gate that rejects every Cargo path origin escaping `build_root` before Cargo starts, or define another portable manager-owned mechanism that makes such reads impossible. Do not defer this to a hardened filesystem guarantee. Add adversarial relative and absolute outside-root cases proving no external manifest is read or resolved.

Primary reference: https://doc.rust-lang.org/cargo/reference/specifying-dependencies.html#specifying-path-dependencies.

### 4. The executable evidence does not exercise the normative vendored/configured pipeline

The replayed closure fixture is dependency-free. Probe `baseEnv` creates an empty `CARGO_HOME` but never writes the normative four-table config. The structural graph/build commands omit parts of the exact vectors such as `--locked` and `--bin`, and no probe case builds one vendored crates.io dependency in either local or external source mode. Therefore the green probe does not substantiate the handoff claim that the complete vendored pipeline and local/external equivalence were measured.

The config contract is also not serialization-complete: it places an absolute `<build_root>/vendor` path inside a TOML basic string while claiming no package-derived byte reaches the file, although local and descriptor build-root components influence that absolute path. It does not define escaping/canonicalization for quotes, backslashes, control characters, or Windows paths.

Required rework: define exact TOML serialization for the directory value on macOS and Windows, including path normalization and rejection rules. Extend executable evidence to run the exact two argv vectors with the manager-written config, a real vendored crates.io dependency, isolated host Cargo state, and both source-mode mappings. Add adversarial path-serialization vectors and prove no operator Cargo cache/config is read.

## Re-review gate

A fresh reviewer should receive revised task-scoped decision/reference/results artifacts and probe evidence that closes all four findings. The probe and repository Go gates must remain green, the six classifier controls must still fail as required, accepted decisions 0007/0008 and frozen release artifacts must remain unchanged, and Windows/Linux must remain qualification obligations rather than inferred claims.