# Adapter source-closure architecture decision

Status: prepared for review

Board task: `TASK-260810-1dgdos`

Run: `RUN-260811-021641`

Authoritative run goal at the drafting checkpoint: `GOAL-260811-001110`
revision 1, resolved scope `TASK-260810-1dgdos`

Date: 2026-08-11

## Decision

Curator will implement one shared, fail-closed source-closure security and
evidence contract with separate ecosystem adapters for Rust/Cargo,
Node/TypeScript package managers, and SwiftPM with its conservatively supported
C-family targets. Go remains the implemented behavior baseline. Python remains
an independent protocol implementation and ecosystem reference; this cycle
exports shared schemas and conformance goldens to it but adds no Python adapter
or shared Python/Node repository, runtime, cache, or package graph. Kotlin,
Dart, .NET, and verified dependency binaries remain deferred.

Every supported adapter must prove the same predicate:

```text
supported_build(selection) =
    immutable_recursive_capture(selection)
    AND complete_recursive_artifact_admission
    AND closed_selection_and_build_graph
    AND trusted_toolchain_and_target_binding
    AND authorized_declared_execution_only
    AND offline_replay_from_empty_ambient_state
    AND causally_receipted_protected_outputs
```

If any conjunct is false, unknown, unsupported, partially inspected, or
non-reproducible from the captured evidence, Curator fails before the affected
build, install, or publication step. No package label, checksum, lock entry,
manager cache, installed tree, file suffix, or successful unmanaged build may
stand in for this proof.

No human-only architecture decision remains. All apparent conflicts among the
accepted outcomes resolve as ecosystem-specific profile data beneath one
security model:

- the artifact classifier owns byte admission for every adapter;
- the language-neutral graph owns dependency, selection, target, toolchain,
  action, output, FFI, and subprocess semantics;
- the checkpoint chain owns causal evidence and cache identity;
- the protected executor owns process, read, environment, write, and network
  confinement; and
- each ecosystem adapter owns only its lock/origin parser, capture transform,
  materialization rules, supported build profile, and typed evidence fields.

## Governing requirements

The following labels are shorthand for the authoritative headings in
`.spec/skill-facing-cli-source-closure.md`.

| Label | Requirement |
| --- | --- |
| `SCOPE` | Current delivery scope: Go baseline; Rust, Node/TypeScript, Swift, and SwiftPM-supported C/C++/Objective-C/Objective-C++ targets; independent Python reference; Kotlin, Dart, and .NET deferred. |
| `SCI-1` | Enumerate every transitive dependency and bind immutable identity. |
| `SCI-2` | Retain lock, checksum, source snapshot, build metadata, target, and toolchain evidence for audit/checkpoints. |
| `SCI-3` | Rebuild offline from the captured closure without new network resolution. |
| `SCI-4` | Hooks, plugins, generators, and build scripts cannot add undeclared inputs or escape the trusted boundary. |
| `SCI-5` | Mixed-language edges explicitly record build order, toolchain, target platform, and FFI/process boundaries. |
| `SCI-6` | Unsupported ecosystems or shapes fail closed or use a narrower source-only profile. |
| `VCAP` | Vendored compiled executable/intermediate dependency payloads are rejected until a separate verified-binary capability exists. |
| `RESEARCH` | Produce the inventory, taxonomy, ecosystem strategies, mixed-language/checkpoint contract, diagnostics, and conformance evidence. |
| `DISCOVERY` | Publish the ecosystem matrix, common contract, unsupported cases, diagnostics, checkpoints, conformance requirements, and implementation-ready backlog. |
| `DELIVERY` | Translate the accepted decision into atomic implementation work and integrate reviewed Rust, Node/TypeScript, and SwiftPM/C-family paths with shared conformance. |

## Reviewed outcomes consumed

Every active discovery outcome is accepted on the board (`status=done`) and is
normative input to this synthesis.

| Outcome | Accepted artifact and SHA-256 | Decision consumed |
| --- | --- | --- |
| `TASK-260810-1veyfw` — inventory language and reference surfaces | `260811_inventory-language-and-reference-surfaces.md`, `59e8337ef489cbbfd961a7640db1ee01c2a85421057c580654f83cba106ee89c` | Go is the only implemented adapter; Rust and Node need purpose-built fixtures; Swift and real Swift-to-C targets exist; currency exchange supplies a real Go-to-Swift subprocess case; Python is independent; revision state, generated outputs, external system commands, and source payloads are separate authorities. Accepted verdict: `TASK-260810-1veyfw_reviewer-verdict_RUN-260811-46059b.md`. |
| `TASK-260810-29vk09` — compiled-artifact taxonomy | `260811_compiled-artifact-taxonomy-and-deny-policy.md`, `c5334433d6eddf37109e612a97866024a17d38c15cff7d7e5e36dac69fe0df15` | One deny-dominant recursive classifier, causal trust roles, canonical `artifact-manifest-v1`, stable artifact diagnostics, closed limits, exact ELF `ET_DYN` handling, and unavailable `verified-binary-v1`. Accepted verdict: `TASK-260810-29vk09_review-verdict_RUN-260811-e91a91.md`. |
| `TASK-260810-2n3sbi` — Node/TypeScript and Python closure | `260811_node-typescript-and-python-source-closure.md`, `68ecaad383fc3fd7b2704065f0d1e7d78446c5c7f535b4fcbfdd669e7003fe4f` | Shared semantics but separate implementations; distinct npm, pnpm, Yarn Classic, and modern Yarn profiles; raw artifacts as authority; caches as derived state; lifecycle suppression; declared TypeScript generation; independent Python protocol fixtures. Accepted verdict: `TASK-260810-2n3sbi_review-verdict_RUN-260811-76cafd.md`. |
| `TASK-260810-3urqbl` — Rust/Cargo closure | `260811_rust-cargo-source-closure.md`, `620c789545273a1c4fc9c9baf25b9db8e4220c79f3fb1ad299ac1bdcd7e51423` | `rust-source-v1`; pre-Cargo origin admission; lock-superset plus active graph; pinned `cargo-vendor-transform-v1`; native target only; build scripts, proc macros, native links, ambient config, and cross/custom targets rejected. Accepted verdict: `TASK-260810-3urqbl_reviewer-verdict_RUN-260811-05742e.md`. |
| `TASK-260810-zddzh7` — SwiftPM mixed C-family closure | `260811_swiftpm-mixed-c-family-source-closure.md`, `361b389e54809d0bce44ea9698860e04de26a0f5ab96219481d17aca47135b3a` | Restricted `swiftpm-source-v1`; intake before executable manifest evaluation; kind-preserving mirrors; separate Swift and Clang targets; module-map/header read verification; pinned toolchain/SDK/platform profiles; active plugins/macros and binaries rejected. Accepted verdict: `TASK-260810-zddzh7_review-verdict_RUN-260811-89e84f.md`. |
| `TASK-260810-1uu9lk` — cross-language graph and checkpoints | `260811_cross-language-closure-graph-and-checkpoints.md`, `874c3a40b9ff9fcf130f37f22e9f5aa2bdd6d9f246403e059c98e84736e27bbc` | Selection-neutral capture, selection binding, active graph, deterministic execution DAG, pre-C5 derivation permits/receipts, C0-C7, exact CGP05/CGP10 canonical records, and the reconciled 14-leaf implementation DAG. Accepted verdict: `TASK-260810-1uu9lk_review-verdict_RUN-260811-e60eda.md`. |

The independent canonical verifier
`260811_cross-language-closure-canonical-golden-verifier.rb` has SHA-256
`2254776d4780e4c32ee37ecbf1b22ad092f029ae3ca3be1749ef373c8162d075`.

## Supported language and ecosystem matrix

| Ecosystem/profile | Cycle decision | Supported boundary | Required closure strategy | Explicit fail-closed boundary |
| --- | --- | --- | --- | --- |
| Go baseline | Preserve | Existing `go-v1`/`go-repository-v1` behavior is the compatibility and security baseline. | Fixed vendored recursive graph, network-off build, selected full GOROOT fingerprint, exact receipt/protected-cache checks. | Existing cgo/native/host-object restrictions remain; no shared contract may weaken the Go path. |
| Rust `rust-source-v1` | Implement | Exactly one named Cargo package and one named binary on one native target with stable pinned Cargo/rustc. | Capture root/path trees, registry index plus raw `.crate`, exact Git commit/tree/submodules; admit all origins before Cargo; run confined pinned vendor transform; reconcile lock superset and active target/feature graph; frozen fresh-home build. | Build scripts, build dependencies, proc macros, `links`/native inputs, package Cargo config/wrappers/runners, custom/cross/multiple targets, unstable features, artifact dependencies, compiled or opaque payloads. |
| Node/TypeScript pure-source common profile | Implement | Source-only runtime packages and separately declared root TypeScript/generator actions. | Common canonical package/peer/workspace/condition graph and runtime/tool binding; raw packages as authority; lifecycle suppressed; fresh private derived manager state; OS-denied offline replay. | Dependency lifecycle-required packages, implicit `node-gyp`, native addons or native source builds, Wasm/V8 caches, undeclared generators/plugins, mutable Git/HTTP sources, ambient caches. |
| npm | Implement separately | Supported pinned `package-lock.json`/`npm-shrinkwrap.json` schemas. | Bind root/workspaces, resolved locators, SRI plus Curator digest, raw tarballs; private cache; `npm ci --ignore-scripts --offline`; metadata reconciliation. | Missing/stale lock, mutable locator, bundled dependencies, implicit `binding.gyp`, integrity/metadata drift, native/opaque payloads. |
| pnpm | Implement separately | Supported pinned `pnpm-lock.yaml` schema. | Bind importers, snapshots, peer contexts, overrides/patches/target settings; raw packages and local roots; private frozen store; offline scripts-disabled materialization. | `.pnpmfile.*`, custom resolvers/fetchers, side-effects cache, undeclared patches, uncaptured `file:` roots, native/opaque payloads. |
| Yarn Classic 1.x | Implement separately | Supported root `yarn.lock`, manifests, exact Yarn version/config. | Capture source tarballs; task-private offline mirror; empty ordinary cache; frozen/offline/ignore-scripts materialization. | Dependency-subtree lock authority, checksum update, lifecycle execution, missing/mutable artifacts, ambient cache, native/opaque payloads. |
| Modern Yarn | Implement separately | Supported modern lock/config/linker/cache profile with approved built-in plugins only. | Bind exact Yarn release, `.yarnrc.yml`, cache key/compression, linker, conditions, patches and checksums; immutable private cache; network disabled, immutable lock/cache, skip-build. | Local/downloaded plugins, custom fetchers/resolvers/linkers, Git pack hooks, undeclared patches, preseeded PnP/install state, lifecycle execution, native/opaque payloads. |
| Python reference | Protocol compatibility only | Independent implementation may consume/export identical schemas, diagnostics, and semantic fixtures. Existing pure-source packaging research remains reference evidence. | Compare canonical records and expected fixture outcomes at the protocol boundary; keep interpreter/frontend/backend/tag evidence in the Python implementation. | No new Curator Python adapter, shared Node/Python code, venv, cache, repository, or runtime task. Current online `pip`, bytecode, wheel-native, and venv payload behavior is not admitted. |
| Swift `swiftpm-source-v1` | Implement | Source-only Swift targets and exact product/destination under controlled SwiftPM. | Admit complete root/dependency trees before manifest interpretation; freeze top lock and exact revisions; kind-preserving local mirrors; controlled graph replay; isolated native build. | Active plugins/macros, binary targets, unsafe flags, untrusted system libraries, registries, Git submodules/LFS/filters, unvalidated destinations. |
| C under SwiftPM | Implement | Separate Clang target with admitted C sources/headers and contained/generated module map. | Independent tree inventory, module-map parser, observed compiler header/module reads, Clang/SDK/platform binding. | Swift and C-family source in one target, escaping header/module reads, arbitrary host/system library, compiled dependency. |
| C++ under SwiftPM | Restricted implementation | C++ target behind a C-compatible shim; direct Swift C++ import only under an accepted toolchain/platform/API fixture and transitive `.interoperabilityMode(.Cxx)`. | Bind C++ standard/library, Clang ABI, target/SDK, header/module graph, and compatibility profile. | Unsupported API shapes, missing interop opt-in, unvalidated platform/toolchain, prebuilt library/module. |
| Objective-C / Objective-C++ under SwiftPM | Restricted implementation | Initial Apple/Darwin SDK/runtime fixtures; Objective-C++ normally exposed through C/Objective-C headers, with C++ rules applied if directly exposed. | Bind runtime, SDK/framework/system edges, headers/modules, target and Clang/Swift identities. | Non-Darwin or otherwise unvalidated runtime/toolchain/destination, arbitrary frameworks/libraries, compiled dependency. |
| Non-SwiftPM C family | Not this cycle | None. | Requires a separately approved graph/build-system profile. | CMake, Meson, Autotools, Make, Bazel, custom scripts, and other native graphs remain unsupported. |
| Kotlin/JVM, Dart, .NET | Deferred | None. | Preserve evidence only; renewed approval is required for research or implementation. | No current task or dependency. JAR/JVM bytecode remains a compiled deny class. |

## Shared adapter contract

### Trust roles and artifact admission

The shared classifier assigns trust before class and exposes four causal roles:

| Role | Current handling |
| --- | --- |
| `dependency_input` | Recursively inspect every reachable byte/node/container member before manager execution. Any compiled, opaque, ambiguous, unsafe, encrypted, unsupported, malformed, over-limit, or partially inspected content rejects the package and ancestors. |
| `external_toolchain` | Allowed only when Curator selected and fingerprinted the complete contained toolchain/SDK/sysroot outside the dependency closure and rechecks it immediately before use. Package paths and metadata cannot claim this role. |
| `local_build_output` | Allowed only when causally produced after admission in a clean protected namespace by a declared action, observed and hashed, and published with an exact protected receipt. Copying an existing binary does not establish production. |
| `verified_binary_candidate` | Rejected in this cycle with `artifact_binary_admission_unavailable`; no verifier or allow flag exists. |

`artifact-manifest-v1` is canonical, policy/version/limit/detector bound, and
records immutable origin, trust role, every canonical node/member, class,
decision, diagnostic, container chain, size, hash, and final digest. Deny
dominates. Suffixes and package metadata are evidence only; compiled bytes with
wrong or absent names still reject, and deny-indicating names with benign bytes
are ambiguous and reject.

The v1 container registry covers ZIP/ZIP64, tar variants, gzip envelopes, and
native archive formats needed to identify libraries. Safe-path, link/special
node, collision, size/count/depth/ratio, and incomplete-inspection rules are
global. Adapters may narrow allowed source grammars; they may not admit a class
the shared policy denies.

### Canonical graph and selection

The graph is a typed multigraph with four distinct identities:

```text
CaptureGraph
    + SelectionContext(selection)
    + SelectionBinding(selection)
    -> ActiveGraph(selection)
    -> deterministic BuildPlan(selection)
```

- `CaptureGraph` is the conservative selection-neutral lock/resolution
  superset. It includes inactive optional, feature, platform, peer, dormant
  extension, and development declarations, but excludes requested target facts,
  concrete platform/toolchain nodes, concrete `targets`/tool edges, and
  selected/pruned results.
- `SelectionContext` carries requested products, features, markers/cfg values,
  peer context, evaluator identities, and intrinsic target-platform IDs.
- `SelectionBinding` is the only authority that adds exact target-platform and
  Curator-selected external-toolchain nodes plus concrete `targets`,
  `uses_tool`, toolchain-scoped `requires`, and SDK/interop binding edges.
- `ActiveGraph` combines selected capture records with the exact binding and
  records every selected or pruned conditional declaration.

The ten closed node kinds are `command_product`, `package_instance`,
`source_set`, `target_unit`, `action`, `generated_artifact`,
`interop_boundary`, `toolchain_component`, `target_platform`, and
`output_artifact`. The eleven closed edge kinds are `declares`, `resolves_to`,
`requires`, `reads`, `uses_tool`, `targets`, `produces`, `provides_interop`,
`consumes_interop`, `invokes`, and `publishes`.

Nodes contain intrinsic immutable declarations only. All package/source,
provider/consumer, action/output, platform/tool, interop, and publication
relationships live in edges. Observed C6 output bytes are separate produced
artifact observations and never mutate an expected output node. Duplicate,
dangling, wrong-kind, relationally contaminated, capture-replacing, or
zero/multiply bound slot records reject canonically.

Non-ordering runtime/peer cycles are retained as canonical SCC evidence. Any
cycle in the execution projection rejects with `closure_build_cycle`. Acyclic
actions are ordered by stable Kahn waves sorted by canonical action ID.

### Mixed-language and process policy

Every language/process boundary is an `interop_boundary` with an explicit mode:
`c_abi`, `cxx_interop`, `objc_runtime`, `native_link`, `dynamic_load`,
`host_extension`, or `subprocess_protocol`.

Provider and consumer attach through separate typed edges; platform,
toolchain, SDK, ABI/runtime, symbol/header/module/interface, and link/load or
protocol evidence is explicit. Compile/link boundaries create ordering arcs.
Runtime-only subprocess relations do not invent build order; co-publication is
expressed separately. An interop declaration never changes artifact trust.

Initial decisions are:

- Swift to C is supported with separate targets and full header/module/read
  evidence.
- Swift to C++ is restricted to accepted exact toolchain/platform/API profiles.
- Swift to Objective-C/Objective-C++ is restricted initially to accepted
  Darwin SDK/runtime profiles.
- Go `exchange` to the Swift scraper is a supported graph shape through a
  versioned JSON `subprocess_protocol` only after both closures build and
  publish independently.
- Rust native/build-script/proc-macro edges and Node native addon/Wasm edges are
  representable but rejected by the current source-only profiles.

### Checkpoint and execution contract

Every checkpoint contains its schema/name, exact predecessor, canonical
payload, decision, deterministic diagnostics, and domain-separated identity.

| Checkpoint | Gate |
| --- | --- |
| `C0.profile` | Bind policy, detector/limit/source profiles, requested selection, intrinsic platform records/roles, supported manager schemas, configuration policy, capabilities, and complete external toolchains for any executable that can derive evidence before C5. No process runs before C0. |
| `C1.resolve` | Bind frozen root/workspace declarations, authoritative lock candidate, unevaluated conditions, parser/evaluator identities, candidate graph, and ordered intake/derivation journal. Executable manifests run only after their complete tree is admitted. |
| `C2.capture` | Aggregate every immutable root/dependency origin, protected raw/tree handle, digest, containment record, source-control object/submodule evidence, and broker receipt. Network is confined to a manager acquisition broker that runs no package code. |
| `C3.admit` | Aggregate recursive artifact manifests for every source input. Cargo uniquely uses `C3a` origin admission, a permitted confined vendor derivation, and `C3b` exact transform/post-vendor admission. |
| `C4.close` | Reconcile capture, selection context, binding, active graph, package metadata, selected/pruned conditions, platform/toolchain/SDK bindings, interop/system declarations, SCCs, and graph identities. Mirror/metadata derivations require permits. |
| `C5.plan` | Bind the acyclic deterministic action DAG, commands, tools, cwd, host/target, environment, process/read/write/network policies, generated lineage, expected immutable outputs, bundle/runtime edges, and build plan. C5 adds no graph record and cannot retroactively authorize earlier execution. |
| `C6.offline` | Run from read-only admitted closure plus trusted toolchain with empty task-specific home/config/cache/output/temp roots, OS network denial, frozen manager modes, complete action/I/O/process audit, toolchain rechecks, and separate produced-output observations. |
| `C7.publish` | Reinspect outputs, reconcile the exact write set and runtime entry points, generate canonical execution/publication receipts, atomically publish to protected storage, and validate reuse from an independently derived expected input. |

Pre-C5 executable evidence derivation is governed by three records:
`intake_admission_receipt`, `derivation_permit`, and `derivation_receipt`.
The permit is committed before launch and binds admitted inputs, the C0
toolchain, executable, argv, cwd, environment, host/target, process/read/write/
network policy, expected evidence, limits, and immediate time-of-use recheck.
Observed drift or undeclared behavior cannot enter C1-C4.

The portable identity rule is strict CCJ plus domain-separated SHA-256:

```text
ID(label, payload) = SHA256(UTF8(label) || 0x00 || CCJ(payload))
```

Temporary paths, timestamps, discovery order, and display logs are excluded
from portable identity. Locks, origins, manifests, conditions, bindings,
targets, tools, action slots, environment policy, and expected outputs are not.

## Ecosystem strategies

### Rust/Cargo

`rust-source-v1` keeps two graph views: an all-feature/all-platform lock
superset for conservative acquisition and admission, and the exact active
native-target/feature graph for build. It captures registry index records and
raw archives, exact Git commit/tree/submodule bytes, and contained root/path
trees without treating Cargo caches as provenance.

Every origin admits before any Cargo process. The selected physical Cargo is
C0-bound, then may run `cargo vendor --locked --offline --versioned-dirs` only
under an exact permit into an absent destination. The derivative is accepted
only when it equals the pinned `cargo-vendor-transform-v1` expectation,
including source-specific per-leaf dispositions, normalized Git manifests,
canonical checksum metadata, unique lock-to-directory mapping, and post-vendor
artifact admission. The initial descriptor is bound to Cargo 0.92.0 / Cargo
1.91.0 commit `ea2d97820c16195b0ca3fadb4319fe512c199a43`; another implementation
needs a new descriptor and fixtures.

The build task consumes the closed graph and binding, rejects unsupported host
code/native/config/cross shapes before compilation, and runs exact fresh-home
metadata/build with `--frozen` under the shared executor. Cargo checksum files
remain corroboration only: the retained research proved a forged source plus
forged `.cargo-checksum.json` can still build.

### Node/TypeScript and Python compatibility

The common Node task owns canonical runtime/manager/target bindings, lifecycle
suppression, package/peer/workspace/condition graph mapping, shipped generated
JavaScript classification, and declared TypeScript/generator lineage. It does
not flatten manager behavior into one parser or materializer.

npm, pnpm, Yarn Classic, and modern Yarn each own a separate exact lock,
origin, workspace/peer/condition, private cache/store/mirror, frozen offline
materialization, and extension-rejection profile. Raw admitted package bytes
remain authority. `node_modules`, a pnpm store, Yarn cache/PnP/install state,
or an ambient manager cache is derived output only.

Python compatibility is narrowly limited to shared protocol schemas,
diagnostic semantics, and conformance goldens. The independent Python
implementation retains its own interpreter, lock/frontend/backend, marker/tag,
PEP 517, wheel/sdist, and runtime evidence. No Python product code is required
to unblock Curator's Node adapter, and no new Python implementation task is
created.

### SwiftPM and C family

`swiftpm-source-v1` wraps SwiftPM; unrestricted `swift build` is not the
security boundary. The adapter selects one executable product and exact
destination, fingerprints Swift/SwiftPM/PackageDescription before use, admits
the entire root before root manifest evaluation and every dependency before
its manifest evaluation, freezes the top-level `Package.resolved`, captures
exact source snapshots, and replays through kind-preserving local mirrors with
original origins unavailable and network denied.

The source-closure task owns manifest/lock/origin/mirror and selection-neutral
package/target/source/condition evidence. The interop task owns separate Swift
and Clang target rules, C/C++/Objective-C/Objective-C++ classification, module
map parsing, header/module/read confinement, FFI declarations, SDK/system
edges, C++ mode, and Objective-C runtime restrictions. The build task consumes
the exact binding and performs only native SwiftPM planning/build with
experimental prebuilts disabled, fresh roots, forced pins, observed compiler
reads/writes, and protected output publication.

SwiftPM's own source enumeration is not closure evidence: the accepted research
proved a custom module map can reference an absolute external header while
`swift package describe` omits both the module map and header. Module-map
parsing and compiler read-set verification are therefore mandatory.

## Diagnostics and precedence

Stable codes are global within their policy/schema major version. Human text is
not a machine interface. The implementation must preserve the accepted
`artifact_*`, `closure_*`, `rust_*`, and `swiftpm_*` vocabularies and add no
manager-specific code when structured `ecosystem`, `manager`, `phase`, and
`reason` fields suffice.

Primary precedence is:

1. immutable origin, lock, or integrity failure;
2. recursive artifact safety/classification failure;
3. toolchain, runtime, or target identity failure at the affected pre-use gate;
4. unauthorized or drifted pre-C5 evidence derivation;
5. graph schema/reference/completeness, metadata, interop, or cycle failure;
6. unsupported capability or undeclared hook/build dependency;
7. offline network, process, read/environment, or write violation; and
8. output, receipt, or protected-cache drift.

Representative global codes are:

| Boundary | Required stable codes |
| --- | --- |
| Artifact identity/class | `artifact_origin_unverified`, `artifact_compiled_dependency_forbidden`, `artifact_binary_admission_unavailable`, `artifact_type_ambiguous`, `artifact_opaque_dependency_forbidden`, archive/path/entry/limit/inspection codes. |
| Toolchain/output trust | `artifact_toolchain_untrusted`, `artifact_toolchain_identity_changed`, `artifact_local_output_unreceipted`, `artifact_local_output_drift`. |
| Common closure | lock/integrity/origin/graph/metadata/hook/build-dependency/offline/network/runtime/generated-output codes from the accepted Node/Python outcome. |
| Graph/checkpoint | `closure_graph_schema_unsupported`, `closure_graph_incomplete`, `closure_graph_reference_invalid`, `closure_derivation_unauthorized`, `closure_derivation_drift`, `closure_build_cycle`, `closure_interop_undeclared`, `closure_checkpoint_invalid`, target/process/input/write codes. |
| Rust | lock, registry/Git identity, path escape, vendor transform/incomplete, graph/feature/target, build-script/proc-macro/native-link/config/undeclared/offline codes from `rust-source-v1`. |
| SwiftPM | resolution/lock/pin/origin/mirror/path/manifest/source/mixed-target/unsafe-setting/module/header/plugin/macro/platform/offline/build-graph codes from `swiftpm-source-v1`. |

The normative active code sets are explicit:

- artifact policy: `artifact_origin_unverified`,
  `artifact_compiled_dependency_forbidden`,
  `artifact_binary_admission_unavailable`, `artifact_type_ambiguous`,
  `artifact_opaque_dependency_forbidden`, `artifact_archive_invalid`,
  `artifact_archive_unsupported`, `artifact_archive_encrypted`,
  `artifact_archive_unsafe_path`, `artifact_archive_unsafe_entry`,
  `artifact_inspection_limit_exceeded`, `artifact_inspection_unavailable`,
  `artifact_generated_input_undeclared`, `artifact_toolchain_untrusted`,
  `artifact_toolchain_identity_changed`,
  `artifact_local_output_unreceipted`, `artifact_local_output_drift`, and
  `artifact_policy_internal_error`;
- common closure: `closure_lock_missing`,
  `closure_lock_format_unsupported`, `closure_lock_stale`,
  `closure_integrity_missing`, `closure_integrity_mismatch`,
  `closure_origin_unpinned`, `closure_graph_incomplete`,
  `closure_local_path_escape`, `closure_bundled_dependency_unsupported`,
  `closure_manager_plugin_undeclared`, `closure_hook_undeclared`,
  `closure_build_dependency_unlocked`, `closure_native_build_unsupported`,
  `closure_offline_input_missing`, `closure_network_attempted`,
  `closure_generated_output_drift`, `closure_runtime_identity_changed`, and
  `closure_metadata_mismatch`;
- graph/checkpoint execution: `closure_graph_schema_unsupported`,
  `closure_graph_reference_invalid`, `closure_derivation_unauthorized`,
  `closure_derivation_drift`, `closure_build_cycle`,
  `closure_interop_undeclared`, `closure_checkpoint_invalid`,
  `closure_target_identity_changed`, `closure_process_undeclared`,
  `closure_input_undeclared`, and `closure_write_undeclared`;
- Rust: `rust_lock_required`, `rust_lock_mismatch`,
  `rust_registry_identity_invalid`, `rust_git_identity_invalid`,
  `rust_path_dependency_escape`, `rust_vendor_transform_unsupported`,
  `rust_vendor_incomplete`, `rust_graph_incomplete`,
  `rust_feature_profile_mismatch`, `rust_target_unsupported`,
  `rust_build_script_unsupported`, `rust_proc_macro_unsupported`,
  `rust_native_link_unsupported`, `rust_config_untrusted`,
  `rust_undeclared_input`, and `rust_offline_rebuild_failed`; and
- SwiftPM: `swiftpm_resolution_unfrozen`,
  `swiftpm_resolved_file_out_of_date`, `swiftpm_dependency_pin_mismatch`,
  `swiftpm_dependency_origin_unsupported`,
  `swiftpm_dependency_mirror_missing`,
  `swiftpm_local_dependency_outside_closure`,
  `swiftpm_manifest_replay_drift`, `swiftpm_source_inventory_drift`,
  `swiftpm_mixed_language_target_unsupported`,
  `swiftpm_unsafe_build_setting_forbidden`, `swiftpm_modulemap_escape`,
  `swiftpm_header_input_undeclared`, `swiftpm_plugin_execution_unsupported`,
  `swiftpm_macro_execution_unsupported`,
  `swiftpm_target_platform_unsupported`, `swiftpm_offline_rebuild_failed`, and
  `swiftpm_build_graph_drift`.

Their structured field contracts and human rendering remain exactly those in
the cited accepted outcomes; adapters cannot rename a shared cause or replace
it with a manager-specific string.

A recognized compiled dependency remains primary even when the same package
also requires an unsupported native build or hook.

## Conformance requirements

The integration task must publish reusable input bytes, declarations, exact
canonical records/digests, expected diagnostic and checkpoint, selected/pruned
sets, action start counters, network audit, write set, and receipts. Adapter
wrappers may add typed evidence but cannot weaken a shared expected outcome.

Required families are:

- artifact taxonomy `A01-A08`, `C01-C12`, `F01-F14`, `T01-T05`, and current
  unavailable-capability `V01`;
- cross-language `CGP01-CGP11` and `CGN01-CGN18`;
- Node/Python semantic `S01-S08`, manager `N01-N13`, and independent Python
  `P01-P13` protocol exports;
- Rust `R*`, `RV*`, `GV*`, `VF*`, `PV*`, `RF*`, and `RH*`, including zero Cargo
  spawns for pre-vendor rejection and the two exact Git PathSource branches;
- SwiftPM language `S*`, header/module/system `H*`, resolution/mirror `R*`, and
  extension/artifact `P*`; and
- inventory-derived extensionless Mach-O, revision-state split, Swift-to-C,
  Go-to-Swift subprocess, missing-lock, Python bytecode/venv, and transitive
  drift cases as fixtures or explicit unsupported migration evidence.

The exact CGP05 and CGP10 corpus contains 53 labeled records. Implementations
must independently canonicalize and hash every label/payload, resolve every
reference, prove one exact selection-neutral capture across Darwin and Linux
CGP05 branches with platform nodes and `targets` edges only in bindings, and
prove that two CGP10 output observations change only observation/execution/
publication identities. The accepted Ruby verifier is an executable oracle,
not a substitute for implementation tests.

Every negative fixture also proves no forbidden later checkpoint, affected
process, output, or protected publication occurred. Every positive offline
fixture uses OS-level network denial and empty ambient state; a manager
`--offline` switch alone is insufficient.

## Implementation backlog reconciliation

`STORY-260811-2epsp4` already contains the accepted minimal decomposition. This
synthesis verifies and adopts it without adding tasks.

| Active task | Atomic deliverable | Direct blockers | Spec trace |
| --- | --- | --- | --- |
| `TASK-260811-2gazym` — shared artifact admission | Classifier, recursive containers, path/limit policy, trust roles, `artifact-manifest-v1`. | `TASK-260810-1dgdos` | `SCI-1/2/4/6`, `VCAP`, artifact research. |
| `TASK-260811-i3154q` — canonical graph/checkpoints | Capture/context/binding/active schemas, typed nodes/edges, cycles/order, immutable outputs/observations, C0-C7 codecs, exact goldens. | `TASK-260810-1dgdos` | `SCI-1/2/5`, `RESEARCH`, cross-language outcome. |
| `TASK-260811-27xisf` — protected execution/receipts | Immutable intake, pre-C5 permits/receipts, offline sandbox, observations, multi-output receipts, protected cache. | `TASK-260811-2gazym`, `TASK-260811-i3154q` | `SCI-2/3/4`, `VCAP`, `DELIVERY`. |
| `TASK-260811-2h4m0s` — Cargo capture/vendor transform | Cargo declarations, origin capture/admission, target/tool binding, pinned vendor transform, graph closure. | `TASK-260811-27xisf`, `TASK-260811-2gazym`, `TASK-260811-i3154q` | `SCOPE`, `SCI-1/2/3/4/6`, `VCAP`, Rust outcome. |
| `TASK-260811-3kbf3l` — Rust offline build | Native-target/toolchain profile rejection, frozen build, event validation, protected output. | `TASK-260811-2h4m0s` | `SCOPE`, `SCI-2/3/4/5/6`, `DELIVERY`. |
| `TASK-260811-3twayo` — Node runtime/build contract | Common Node graph, target/runtime/manager binding, lifecycle suppression, TS/generator lineage, Python goldens. | `TASK-260811-27xisf`, `TASK-260811-2gazym`, `TASK-260811-i3154q` | `SCOPE`, `SCI-2/3/4/5/6`, Node/Python outcome. |
| `TASK-260811-1u42b9` — npm profile | npm lock/workspace/tarball/private-cache/offline materializer. | `TASK-260811-3twayo` | `SCOPE`, `SCI-1/2/3/4/6`, npm outcome. |
| `TASK-260811-3ksxig` — pnpm profile | pnpm lock/importer/peer/store/hook restrictions. | `TASK-260811-3twayo` | Same requirements, pnpm outcome. |
| `TASK-260811-twq9ad` — Yarn Classic profile | Yarn 1 lock/workspace/offline mirror. | `TASK-260811-3twayo` | Same requirements, Yarn Classic outcome. |
| `TASK-260811-32iojo` — modern Yarn profile | Modern lock/config/plugin/cache/linker profile. | `TASK-260811-3twayo` | Same requirements, modern Yarn outcome. |
| `TASK-260811-33ukne` — SwiftPM resolution/closure | Intake-before-manifest, exact capture/lock/mirrors, selection-neutral graph and bindings, offline replay. | `TASK-260811-27xisf`, `TASK-260811-2gazym`, `TASK-260811-i3154q` | `SCOPE`, `SCI-1/2/3/4/6`, `VCAP`, SwiftPM outcome. |
| `TASK-260811-tkurtl` — SwiftPM C-family interop | Target-language, module/header/system/FFI and platform/toolchain validation. | `TASK-260811-33ukne` | `SCOPE`, `SCI-4/5/6`, `VCAP`, SwiftPM outcome. |
| `TASK-260811-2qfnai` — SwiftPM offline build | Native plan/build, exact binding consumption, observed I/O, output receipts. | `TASK-260811-tkurtl` | `SCOPE`, `SCI-2/3/4/5/6`, `DELIVERY`. |
| `TASK-260811-x611eq` — cross-adapter conformance | Shared and ecosystem fixture execution, exact goldens, E2E offline/deny proof, task-local support documentation. | Rust build; four Node manager profiles; SwiftPM build | `SCI-1..6`, `VCAP`, `RESEARCH`, `DELIVERY`. |

The decomposition is proportional because each shared leaf owns a distinct
causal trust boundary; Cargo separates immutable intake/transform from build;
Node's four managers have non-interchangeable lock/materialization semantics;
SwiftPM separates resolution, interop, and execution; and integration owns only
cross-adapter proof. Combining adjacent leaves would cross trust or manager
boundaries. Splitting them into schema, docs, per-fixture, or generic quality
tasks would add ceremony without a new deliverable.

Four broad placeholders were closed—not deleted—before implementation:
`TASK-260811-13xlp0`, `TASK-260811-2fwvml`, `TASK-260811-ojduq3`, and
`TASK-260811-1zwhx2`. Their notes identify the atomic replacements and they
retain no active blocker links.

After this synthesis is accepted, the implementation waves are:

| Wave | Tasks |
| ---: | --- |
| 1 | `TASK-260811-2gazym` and `TASK-260811-i3154q` in parallel. |
| 2 | `TASK-260811-27xisf`. |
| 3 | `TASK-260811-2h4m0s`, `TASK-260811-3twayo`, and `TASK-260811-33ukne` in parallel. |
| 4 | `TASK-260811-3kbf3l`, all four Node manager profiles, and `TASK-260811-tkurtl` in parallel. |
| 5 | `TASK-260811-2qfnai`. |
| 6 | `TASK-260811-x611eq`. |

The delivery critical path is:

```text
TASK-260810-1dgdos
-> TASK-260811-2gazym
-> TASK-260811-27xisf
-> TASK-260811-33ukne
-> TASK-260811-tkurtl
-> TASK-260811-2qfnai
-> TASK-260811-x611eq
```

## Scope and justified-gap audit

No beyond-literal-spec element is required or created. Before accepting the
backlog, this synthesis checked `SCOPE`, `SCI-1` through `SCI-6`, `VCAP`,
`RESEARCH`, `DISCOVERY`, `DELIVERY`, and the entire explicit deferred/
unsupported set from the source specification and all accepted outcomes.

| Considered addition | Check result | Disposition |
| --- | --- | --- |
| New Python adapter or shared Node/Python implementation state | The spec makes Python an independent reference and not a default deliverable; accepted research resolves compatibility at the schema/fixture boundary. | Rejected; export goldens only. |
| Kotlin/Gradle/Maven work | Explicitly deferred and requires renewed approval. | Rejected; no task or dependency. |
| Dart or .NET adapter | Explicitly deferred; no current surface or commitment. | Rejected. |
| Verified-binary admission | Explicitly a separate future capability; current policy must reject. | Rejected; unavailable seam only. |
| Non-SwiftPM native build systems | SwiftPM research identifies them as a separate strategy, while this cycle names SwiftPM-supported C family. | Rejected for this cycle. |
| Rust hooks/proc macros/native/cross capability | The source-only v1 profile can satisfy current requirements by explicit rejection; a permissive capability would expand scope and weaken the accepted boundary. | Rejected; negative fixtures retained. |
| Node native-addon build profile | The accepted pure-source profile explicitly rejects it and no requirement demands it. | Rejected. |
| Active Swift plugin/macro execution | Accepted v1 rejects it; a host-execution capability would be new scope. | Rejected. |
| External-system command admission | Inventory distinguishes `glab`/`sentry-cli` from package closure; the spec does not require a system-binary provenance capability. | Rejected for this cycle. |
| Separate documentation, quality-gate, or research tasks | Documentation and verification belong in the relevant implementation/integration AC; all architecture questions are resolved. | Rejected as ceremonial scope. |

Because no addition survives this audit, there is no `Justified gap` record to
attach to a new element. No research task is created: the spec leaves no open
question that blocks an implementation decision. Future capabilities require a
new explicit scope decision rather than a placeholder in this delivery graph.

## Explicit unsupported cases

- any dependency payload containing a native executable, object, static or
  dynamic library, framework/XCFramework, Node addon, Python extension, JVM
  bytecode, Python bytecode, V8 cache, WebAssembly, serialized compiler module,
  or equivalent compiled/intermediate code;
- unknown, ambiguous, opaque, encrypted, malformed, unsupported, over-limit,
  partially inspected, unsafe-path, linked, or special-node dependency content;
- mutable or incomplete origin/lock/checksum/graph identity and any ambient
  package-manager cache or installed tree offered as closure authority;
- undeclared hook, plugin, generator, build backend/script, process, filesystem
  or environment read, write, network operation, target, toolchain, SDK, FFI,
  dynamic-load, or subprocess boundary;
- a build-order cycle in the derived action graph;
- pre-existing or drifted outputs and any cache entry lacking an independently
  derived expected input plus exact protected receipt;
- Rust source shapes outside `rust-source-v1`, Node source shapes outside the
  pure-source manager profiles, and Swift/C-family shapes outside
  `swiftpm-source-v1` and accepted target fixtures;
- a new Python adapter, Kotlin/Gradle/Maven, Dart, .NET, verified binaries,
  non-SwiftPM native systems, and external system-command admission in this
  cycle; and
- universal cross-toolchain/host bit reproducibility as an unstated claim.
  The contract guarantees exact logical identity, offline reconstruction, and
  causal receipts; a stronger reproducibility profile requires its own proof.

## Risks and mitigations

| Risk | Consequence | Mitigation and owning task |
| --- | --- | --- |
| Shared detector/parser gaps or resource exhaustion | False admission or incomplete evidence. | Closed registry, deny-on-error/unknown, versioned limits, streaming manifests, shared byte fixtures in `TASK-260811-2gazym`. |
| Graph identity contamination or recursive references | Target-specific cache aliasing or non-canonical records. | Edge-only relations, selection binding, immutable expected outputs, exact CGP05/CGP10 verifier in `TASK-260811-i3154q`. |
| Pre-C5 package-manager execution reads unadmitted inputs | Resolution evidence could be attacker-influenced before the build plan. | Intake-before-interpretation, C0 toolchains, committed derivation permits, causal receipts, zero-start vectors in `TASK-260811-27xisf` and ecosystem capture tasks. |
| OS sandbox differences | Manager flags alone may not block descendant network/process/file access. | Outer manager-owned isolation, complete observed I/O/process/network audit, platform-specific conformance in `TASK-260811-27xisf`. |
| Cargo vendor behavior changes with Cargo release | Derived vendor bytes no longer match the accepted mapping. | Version/commit-bound transform descriptor and fail-closed unsupported transform in `TASK-260811-2h4m0s`. |
| SwiftPM executable manifests or module maps escape the package | Hidden host input reaches resolution/compilation. | Admit each full tree before manifest execution, parse module maps, verify compiler reads in `TASK-260811-33ukne` and `TASK-260811-tkurtl`. |
| Node manager semantics are flattened | Peers, workspaces, hooks, cache state, or conditions become hidden. | One common contract plus four separate manager leaves and manager-specific negative fixtures. |
| Toolchain/SDK mutation between checkpoint and use | Same logical source could consume different trusted binaries. | Full fingerprints, contained paths, immediate time-of-use rechecks, target/tool edges, and no affected process on drift. |
| Locally generated outputs are mistaken for source or trusted cache | Dependency binaries could bypass admission or stale outputs could publish. | Disjoint empty output roots, causal observations/receipts, output reinspection and exact protected publication. |
| Existing estate migrations lack manifests/locks or contain drift | A real CLI cannot immediately use the new adapters. | Treat estate cases as fixtures/unsupported migration evidence; do not weaken profiles. Migration work requires separately scoped command metadata after the adapters exist. |
| Test corpus and full repository gates are resource intensive | Validation may be slow or expose ambient-cache memory pressure. | Keep deterministic focused validators, run uncached authoritative gates, report resource anomalies truthfully, and never count an interrupted/cached traversal as green. |

## Requirement-to-backlog evidence

| Requirement | Decision evidence | Delivery proof |
| --- | --- | --- |
| `SCOPE` | Language matrix and ecosystem strategies. | Rust, common Node plus four managers, and three SwiftPM leaves; Python goldens only; no Kotlin task. |
| `SCI-1` | Recursive artifact manifest plus selection-neutral capture graph. | Artifact, graph, Cargo capture, Node manager, and SwiftPM resolution tasks; `CGP01`, `CGP05`, artifact and ecosystem transitive vectors. |
| `SCI-2` | Canonical identities, C0-C7, first-binding rules, tool/target bindings, expected outputs/observations. | Every shared/ecosystem task and exact CGP05/CGP10 records. |
| `SCI-3` | C6 fresh-state offline replay and protected execution boundary. | Executor plus Rust build, Node materializers, SwiftPM build, `CGP08`/`CGN13`, ecosystem offline vectors. |
| `SCI-4` | Pre-C5 derivation plane, declared action slots, observed I/O/process/network reconciliation. | Executor, Node common, Rust/Swift profile rejection, `CGP11`, `CGN04/05/10/16-18`. |
| `SCI-5` | Typed graph, selection binding, stable action waves, first-class interop boundaries. | Graph, Swift interop/build, Rust/Node target evidence, integration; `CGP02/03/05/10`, `CGN02/03/09/15`. |
| `SCI-6` | Closed profiles and unsupported table. | Every adapter negative suite; no best-effort flag. |
| `VCAP` | Shared causal trust roles and deny-dominant classifier. | Artifact service plus every adapter wrapper; shared `C*`, `F*`, `T*`, `CGN06/08/11`. |
| `RESEARCH`/`DISCOVERY` | All six reviewed inputs, this synthesis, diagnostics, checkpoints, risks, conformance, and backlog sections. | Reviewer verdict on `TASK-260810-1dgdos`. |
| `DELIVERY` | Fourteen atomic leaves and dependency waves. | Producer/reviewer lifecycle through accepted implementations and final integration; research status is not mistaken for delivery completion. |

## Acceptance and completeness audit

| `TASK-260810-1dgdos` gate | Evidence in this outcome/current board |
| --- | --- |
| Consume and cite all reviewed discovery outcomes | Reviewed-outcomes table names all six accepted outcomes, artifacts, digests, verdicts, and consumed decisions. |
| Resolve language matrix, shared contract, strategies, unsupported cases | Dedicated matrix and contract/strategy/unsupported sections; no human-only decision remains. |
| Publish architecture, backlog, dependencies, risks | This task-scoped outcome, 14-leaf table, waves/critical path, and risk table. |
| Reconcile `STORY-260811-2epsp4` while keeping Kotlin deferred | Live story matches the adopted 14 active leaves and four closed placeholders; Kotlin has no active task/dependency. |
| Proportional smallest board | Proportionality rationale and rejected ceremonial splits. |
| Per-element spec traceability | Every active leaf already contains `Spec trace:` and is mapped above to concrete spec labels and accepted inputs. |
| Justified-gap records | No beyond-literal element was created; the pre-creation gap/out-of-scope audit rejects every candidate. |
| Research only for open questions | No open question remains and no research task is created. |
| Dependencies linked | Direct blockers, waves, and critical path are explicit and will be verified from the live plan. |
| Atomic tasks | One deliverable per active leaf; superseded cross-boundary placeholders remain closed. |
| Completeness | Requirement-to-backlog table covers every specification heading and named delivery target. |
| Planning artifacts linked; diagrams optional | This architecture outcome is task-scoped and will be attached as a new board outcome. No separate `.planning` artifact or diagram was produced because the typed tables and live plan are the authoritative relationship model. |
| Important findings recorded | The decision log below and task board notes serve as the project logbook; the current CLI exposes no separate logbook mutation. |

## Decision log and anomalies

1. The inventory's tracked extensionless Mach-O proves classification cannot be
   suffix-based and supplies a high-value cross-adapter fixture.
2. Live/configured/installed command surfaces are revision-specific authorities;
   repository-name deduplication loses real commands and cannot define closure.
3. The existing Swift estate has real Swift-to-C edges but missing tracked lock
   evidence; current success on one machine is not an admissible closure.
4. SwiftPM manifest evaluation and Cargo vendoring/metadata are executable
   evidence derivations. They require authority before C5 rather than being
   retroactively covered by a build plan.
5. Cargo vendoring is a versioned transform, not provenance or literal origin
   equality. Registry and Git transforms have intentionally different omission
   and normalization rules.
6. npm, pnpm, Yarn Classic, and modern Yarn share security outcomes but not lock
   or materialization implementations.
7. Python and Node share protocol semantics only. Python packaging state is not
   a shortcut to a Node adapter and is not a current Curator deliverable.
8. FFI and subprocess relationships are typed boundary nodes, not informal
   dependency strings. Currency exchange is a subprocess protocol, not FFI.
9. Capture must remain selection-neutral; exact target and tool identities live
   in the binding overlay. Otherwise one captured input could alias multiple
   target-specific builds.
10. Expected outputs are immutable declarations. Produced bytes are C6
    observations; changing output bytes cannot rewrite C4/C5 closure identity.
11. No diagram is warranted: the closed graph tables, checkpoint table, and
    implementation waves communicate the architecture more precisely.
12. No product code, test code, or configuration is changed by this synthesis.

## Verification

The decision and live board were checked after drafting:

- UTF-8, terminal newline, trailing-whitespace, required-outcome/task citation,
  balanced-fence, and `git diff --check` gates passed for this artifact;
- the accepted canonical verifier reported
  `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2
  cgp10_observation_branches=2` and
  `canonical_references=pass cgp05_capture_reused=true
  explicit_target_bindings=2 cgp10_all_refs_resolve=true`;
- the implementation backlog audit reported
  `implementation_backlog=pass active=14 closed_superseded=4
  traced_atomic_leaves=14`;
- the exact blocker-map audit reported
  `dependency_contract=pass active_edges_exact=true
  closed_blockers_empty=true`;
- the task-related planning query resolved 21 relevant elements, the completed
  six discovery prerequisites plus synthesis and six implementation waves, and
  the critical path
  `TASK-260810-1dgdos -> TASK-260811-2gazym -> TASK-260811-27xisf ->
  TASK-260811-33ukne -> TASK-260811-tkurtl -> TASK-260811-2qfnai ->
  TASK-260811-x611eq`;
- the story-local children-only plan correctly refused to hide the two
  out-of-scope blockers from `TASK-260810-1dgdos`; the related-scope plan above
  is the authoritative cross-story DAG until this synthesis is accepted;
- deferred Kotlin story `STORY-260811-1tybyr` is `closed`, explicitly requires
  renewed approval, and has no active delivery dependency; and
- `task-board validate` reported `Board is valid. No issues found.`

No product code changed. Even so, the operator's synthesis checkpoint required
a fresh repository-wide gate: `go test -count=1 ./...` exited 0 on 2026-08-11.
The slowest packages were `cmd/curator` (357.919s),
`internal/install/atomicity` (107.731s), `internal/install` (104.548s), and
`internal/godriver` (70.427s). This task therefore verifies the current product
suite as well as the new document, exact canonical fixtures, and board
contract.

## References

- `.spec/skill-facing-cli-source-closure.md`
- `.research/260811_inventory-language-and-reference-surfaces.md`
- `.research/260811_compiled-artifact-taxonomy-and-deny-policy.md`
- `.research/260811_node-typescript-and-python-source-closure.md`
- `.research/260811_rust-cargo-source-closure.md`
- `.research/260811_swiftpm-mixed-c-family-source-closure.md`
- `.research/260811_cross-language-closure-graph-and-checkpoints.md`
- `.research/260811_cross-language-closure-canonical-golden-verifier.rb`
- Existing Go baseline: `internal/buildsource`, `internal/buildmeta`,
  `internal/buildcache`, `internal/godriver`, `internal/buildrepo`, and
  `internal/install`.
