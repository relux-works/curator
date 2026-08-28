# TASK-260810-3urqbl reviewer verdict

- Run: `RUN-260811-87b027`
- Role: reviewer
- Board goal: `GOAL-260811-33d6ad` revision 1
- Resolved scope: `TASK-260810-3urqbl`
- Reviewed outcome: `TASK-260810-3urqbl_rust-cargo-source-closure.md`
- Reviewed digest: `sha256:300f294974ae9d1ea29d3e63253517015889b04f4024c91be77c9175fc945202`
- Verdict: **changes requested → analysis**

## Acceptance assessment

The research is strong and substantially answers the Cargo.lock, metadata, registry, Git, target, feature, build-script, proc-macro, toolchain, offline-build, diagnostics, checkpoint, and conformance questions. The board resource is byte-identical to `.research/260811_rust-cargo-source-closure.md`, the core Cargo claims are supported by official documentation, and the retained Cargo 1.91.0 fixture replayed successfully.

Acceptance is withheld because the normative closure procedure does not yet preserve the accepted pre-extraction admission boundary, and its origin-to-vendor equality rule is not implementable as written.

## R1 — Raw dependencies are inspected after `cargo vendor` consumes them (high)

The accepted shared artifact policy requires inspection after immutable package capture/checksum verification and **before dependency extraction into a build snapshot** (`.research/260811_compiled-artifact-taxonomy-and-deny-policy.md:687-689`).

The proposed Rust procedure instead runs `cargo vendor --locked --versioned-dirs` at step 4 and only at step 5 recursively inspects the raw packages, Git trees, and generated vendor tree (`.research/260811_rust-cargo-source-closure.md:374-380`). This also contradicts the document statements that every raw payload is scanned before Cargo sees it (`lines 33-38, 322-324`).

Impact: a malformed, unsafe-path, opaque, encrypted, compiled, or otherwise rejected raw package can be parsed/unpacked/copied by Cargo before Curator reaches the deny decision. A post-vendor scan cannot prove the required fail-closed boundary.

Required rework:

1. Split admission into a pre-vendor stage and a post-vendor verification stage.
2. After immutable acquisition and checksum/origin binding, recursively inspect raw `.crate` containers and safely materialized Git/submodule trees with the shared classifier **before** `cargo vendor` or another build-snapshot extractor consumes them.
3. Run vendoring only from the already admitted, manager-controlled capture with offline/no-fallback resolution and a private controlled Cargo home; do not let vendoring reacquire live bytes.
4. Then inspect the generated vendor tree and compare it to the admitted origins before metadata/build.
5. Add a conformance case proving that a rejected raw archive stops before vendor output or any Cargo vendor/build action exists.

Official support: [cargo vendor](https://doc.rust-lang.org/cargo/commands/cargo-vendor.html) copies remote dependency sources into the vendor directory, while [Cargo source replacement](https://doc.rust-lang.org/cargo/reference/source-replacement.html) describes directory sources as unpacked crates. The accepted shared policy, not Cargo extraction behavior, owns admission.

## R2 — “Vendor-to-origin leaf equality” lacks a defined transformation (high)

The procedure requires `vendor-to-origin leaf equality` and checkpoint equality verdicts (`lines 377-380, 430-435`) but does not define which generated, omitted, or normalized files are compared.

The retained `itoa 1.0.15` fixture disproves literal leaf-set equality:

- raw `.crate` only: `.gitignore`
- generated vendor directory only: `.cargo-checksum.json`

Cargo officially documents `.cargo-checksum.json` as associated directory-source metadata and explicitly says it is not a malicious-tamper security mechanism. Therefore the protected Curator mapping must be independent and deterministic rather than an undefined literal comparison.

Required rework:

1. Define a versioned origin-to-vendor transformation/coverage algorithm for registry and Git sources.
2. Enumerate allowed Cargo-generated metadata, permitted omissions, any manifest normalization, and the exact hashes retained for each origin leaf and generated leaf.
3. Make any unexplained addition, omission, mutation, duplicate mapping, or normalization a deterministic `rust_registry_identity_invalid`, `rust_git_identity_invalid`, or `rust_vendor_incomplete` result.
4. Bind the transformation algorithm/version and both manifests in `rust-closure-manifest-v1`.
5. Add exact registry and Git fixtures for generated/omitted files and a forged generated-metadata case.

## Independent verification performed

- Board outcome and research file both hash to `300f294974ae9d1ea29d3e63253517015889b04f4024c91be77c9175fc945202`; byte comparison passed.
- Markdown local-link, balanced-fence, trailing-whitespace, and readability checks passed.
- Official Cargo documentation confirms all-feature lock resolution, all-target lock resolution, metadata target filtering, remote-only vendoring, `--frozen = --locked + --offline`, non-security directory checksums, build-script execution, proc-macro compiler-level access, hierarchical Cargo configuration, Git submodule recursion, target-specific host build dependencies, and custom profile semantics.
- Fresh private `CARGO_HOME` and empty-target replay with physical Cargo/rustc 1.91.0 succeeded under `--frozen`; the active graph included the recursive `aho-corasick -> memchr` edge and excluded the inactive optional/Windows packages.
- The replayed binary printed `42:git:unix:disabled:ambient-build-input:ambient-proc-input`, independently confirming that frozen/offline Cargo is not an execution sandbox.
- Feature-enabled metadata added the optional package; unfiltered metadata added the Windows package; wrong-current-directory offline metadata failed with exit 101 because the workspace vendor config was not discovered.
- Registry archive hashes and the full Git commit in the retained fixture match the research record.
- Relevant baseline tests passed: `go test ./internal/buildmeta ./internal/buildcache ./internal/godriver`.
- A repository-wide `go test ./...` was attempted but did not complete after 11 minutes amid concurrent repository-wide test runs and was stopped; it is not used as acceptance evidence.
- No product or research code was modified during review.

## Re-review gate

A revised outcome must correct both findings, preserve the existing supported/unsupported profile decisions, add the pre-vendor and deterministic mapping fixtures, update the checkpoint schema/equality wording, rerun document gates and relevant tests, and return through a new reviewer cycle.

This is ordinary research rework, not a Stop-The-Line boundary. The required board branch is `analysis`.