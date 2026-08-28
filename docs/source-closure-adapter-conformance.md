# Cross-adapter source-closure conformance

This document is the delivery-level view of Curator's language-aware
source-closure adapters: which ecosystems are supported and at what boundary,
what is explicitly unsupported, which stable diagnostics a caller may branch
on, how the cross-adapter conformance suite proves the shared contract, and
what an existing command has to do to move onto one of the profiles.

It describes tested behaviour only. Where a claim is proved somewhere other
than the cross-adapter suite, this document names the package that proves it.

## The one predicate

Every supported adapter proves the same thing before it builds, installs, or
publishes anything:

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

If any conjunct is false, unknown, unsupported, partially inspected, or not
reproducible from the captured evidence, the adapter fails before the affected
step. A package label, checksum, lock entry, manager cache, installed tree,
file suffix, or a successful unmanaged build never stands in for this proof.

## Supported profiles

| Path | Profile | Supported boundary | Package |
| --- | --- | --- | --- |
| Go | `go-v1` | Unchanged compatibility and security baseline for compiled commands. Not one of the six adapter paths the cross-adapter suite drives. | `internal/godriver`, `internal/buildsource` |
| Rust | `rust-source-v1` | One named Cargo package and one named binary on one native target under the pinned Cargo descriptor. | `internal/rustsource` |
| npm | `npm-source-v1` | Supported pinned `package-lock.json` / `npm-shrinkwrap.json` schemas, source-only packages. | `internal/npmsource` |
| pnpm | `pnpm-source-v1` | Supported pinned `pnpm-lock.yaml` schema, source-only packages. | `internal/pnpmsource` |
| Yarn Classic | `yarn-classic-source-v1` | Root `yarn.lock` v1 with the exact supported Yarn version and config. | `internal/yarnclassicsource` |
| Modern Yarn | `yarn-modern-source-v1` | Supported modern lock, `.yarnrc.yml`, cache key/compression, linker, and built-in plugins only. | `internal/yarnmodernsource` |
| Swift and the SwiftPM C family | `swiftpm-source-v1` | One executable product and one exact destination; source-only Swift targets plus separate Clang targets for C, C++, Objective-C, and Objective-C++ within the accepted toolchain/destination fixtures. | `internal/swiftpmsource`, `internal/swiftpminterop`, `internal/swiftpmbuild` |
| Python | protocol compatibility only | Curator ships no Python adapter. It exports canonical records, diagnostics, and conformance goldens for an independent implementation to consume. | `internal/nodesource` goldens, `internal/crossconformance` export |

All four Node package managers share one common Node contract — capture graph,
lifecycle suppression, runtime/manager binding, declared TypeScript and
generator lineage — and keep four separate lock parsers and materializers,
because their lock and installation semantics are not interchangeable.

## Explicit unsupported cases

These fail closed with a stable diagnostic. None of them is a warning, a
best-effort path, or an adapter flag.

**Dependency bytes**

- any native executable, object, static or dynamic library, framework or
  XCFramework payload, Node addon, Python extension module, JVM or Python
  bytecode, V8 code cache, WebAssembly module, or serialized compiler module,
  embedded directly or inside an archive;
- unknown, ambiguous, opaque, encrypted, malformed, unsupported, over-limit,
  partially inspected, unsafe-path, linked, or special-node content;
- a vendored precompiled binary with a valid signature: verified binary
  admission is a separate future capability and is currently unavailable.

**Identity and origin**

- a missing, stale, or unsupported lock; a mutable locator; a missing or
  mismatched integrity record; an ambient package-manager cache or an installed
  tree offered as closure authority;
- a workspace, path dependency, patch, backend path, or local package that
  resolves outside its declared immutable capture root;
- Git, hosted-Git, arbitrary HTTP archive, and bundled-dependency origins in
  the Node profiles; Swift package registries in `swiftpm-source-v1`; Git
  submodules, LFS pointers, checkout filters, and required hooks.

**Execution**

- any undeclared hook, plugin, generator, build backend or script, process,
  filesystem or environment read, write, network operation, target, toolchain,
  SDK, FFI boundary, dynamic load, or subprocess boundary;
- npm's implicit `binding.gyp` / `node-gyp` path, pnpm `.pnpmfile.*` hooks and
  side-effects cache, non-approved modern Yarn plugins and custom
  fetchers/resolvers/linkers, Rust build scripts, build dependencies, proc
  macros, `links` native inputs, package Cargo config, cross and custom
  targets, and active SwiftPM plugins and macros;
- a cycle in the derived execution projection.

**Outputs**

- a pre-existing or drifted output, and any cache entry without an
  independently derived expected input plus an exact protected receipt.

**Ecosystems**

- Kotlin/Gradle/Maven, Dart, and .NET; non-SwiftPM native build systems
  (CMake, Meson, Autotools, Make, Bazel, custom scripts); a new Python adapter;
  and external system-command admission. All are deferred and need a separate
  scope decision.

A recognized compiled dependency stays the primary failure even when the same
package would also require an unsupported native build or hook.

## Diagnostics

Stable codes are lowercase snake case and global within their policy or schema
major version. Human text is never a machine interface: branch on the code and
the structured `ecosystem`, `manager`, `phase`, and `reason` fields.

| Boundary | Codes |
| --- | --- |
| Artifact identity and class | `artifact_origin_unverified`, `artifact_compiled_dependency_forbidden`, `artifact_binary_admission_unavailable`, `artifact_type_ambiguous`, `artifact_opaque_dependency_forbidden`, `artifact_archive_invalid`, `artifact_archive_unsupported`, `artifact_archive_encrypted`, `artifact_archive_unsafe_path`, `artifact_archive_unsafe_entry`, `artifact_inspection_limit_exceeded`, `artifact_inspection_unavailable`, `artifact_generated_input_undeclared` |
| Toolchain and output trust | `artifact_toolchain_untrusted`, `artifact_toolchain_identity_changed`, `artifact_local_output_unreceipted`, `artifact_local_output_drift`, `artifact_policy_internal_error` |
| Common closure | `closure_lock_missing`, `closure_lock_format_unsupported`, `closure_lock_stale`, `closure_integrity_missing`, `closure_integrity_mismatch`, `closure_origin_unpinned`, `closure_graph_incomplete`, `closure_local_path_escape`, `closure_bundled_dependency_unsupported`, `closure_manager_plugin_undeclared`, `closure_hook_undeclared`, `closure_build_dependency_unlocked`, `closure_native_build_unsupported`, `closure_offline_input_missing`, `closure_network_attempted`, `closure_generated_output_drift`, `closure_runtime_identity_changed`, `closure_metadata_mismatch` |
| Graph, checkpoint, execution | `closure_graph_schema_unsupported`, `closure_graph_reference_invalid`, `closure_derivation_unauthorized`, `closure_derivation_drift`, `closure_build_cycle`, `closure_interop_undeclared`, `closure_checkpoint_invalid`, `closure_target_identity_changed`, `closure_process_undeclared`, `closure_input_undeclared`, `closure_write_undeclared` |
| Rust | `rust_lock_required`, `rust_lock_mismatch`, `rust_registry_identity_invalid`, `rust_git_identity_invalid`, `rust_path_dependency_escape`, `rust_vendor_transform_unsupported`, `rust_vendor_incomplete`, `rust_graph_incomplete`, `rust_feature_profile_mismatch`, `rust_target_unsupported`, `rust_build_script_unsupported`, `rust_proc_macro_unsupported`, `rust_native_link_unsupported`, `rust_config_untrusted`, `rust_undeclared_input`, `rust_offline_rebuild_failed` |
| SwiftPM | `swiftpm_resolution_unfrozen`, `swiftpm_resolved_file_out_of_date`, `swiftpm_dependency_pin_mismatch`, `swiftpm_dependency_origin_unsupported`, `swiftpm_dependency_mirror_missing`, `swiftpm_local_dependency_outside_closure`, `swiftpm_manifest_replay_drift`, `swiftpm_source_inventory_drift`, `swiftpm_mixed_language_target_unsupported`, `swiftpm_unsafe_build_setting_forbidden`, `swiftpm_modulemap_escape`, `swiftpm_header_input_undeclared`, `swiftpm_plugin_execution_unsupported`, `swiftpm_macro_execution_unsupported`, `swiftpm_target_platform_unsupported`, `swiftpm_offline_rebuild_failed`, `swiftpm_build_graph_drift` |

Precedence is deterministic: immutable origin, lock, or integrity failure, then
recursive artifact safety and classification, then toolchain/runtime/target
identity at the affected gate, then unauthorized or drifted pre-C5 evidence
derivation, then graph schema/reference/completeness/metadata/interop/cycle
failure, then unsupported capability or undeclared hook, then offline network,
process, read, or write violation, then output, receipt, or protected-cache
drift.

## The cross-adapter conformance suite

`internal/crossconformance` is the integration proof over the accepted adapter
contracts. It implements no adapter, starts no process, and re-derives no
security decision. Its production files import only the standard library, which
a guard enforces, so the oracle cannot end up checking itself.

Run it with:

```bash
go test -count=1 ./internal/crossconformance
```

It proves four things.

**1. The accepted canonical corpus.** All 53 published CGP05 and CGP10 records
are decoded, canonicalized, and hashed by an independent CCJ-1 scanner and
emitter written for this package. Every typed reference is resolved, the two
CGP05 branches are shown to reuse one exact capture while their platform,
selection, targets edge, binding, active graph, plan, C4, and C5 identities all
differ, and the two CGP10 branches are shown to keep every stable record while
only observation, execution, and publication identities move. A tampered corpus
is rejected with the exact reason.

**2. One normative semantic suite across every delivered path.** Rust, npm,
pnpm, Yarn Classic, modern Yarn, and SwiftPM/C-family each project one fixed
set of source inputs onto two exact targets through their own production APIs,
and each must satisfy:

| Obligation | Requirement |
| --- | --- |
| `capture.selection_neutral` | The capture holds no exact target platform, no toolchain component, and no `targets` or `uses_tool` edge. |
| `capture.stable_across_targets` | Two targets over the same inputs produce one exact capture identity. |
| `binding.target_authority` | The exact target and every concrete tool identity enter only through the selection overlay, bound by an explicit `targets` edge where the path emits edges. |
| `binding.diverges_per_target` | The selection-bound identity, active projection, and plan all differ between the two targets. |
| `records.deterministic` | Repeating the projection from freshly captured inputs reproduces every identity. |
| `evidence.causal_chain` | Every emitted checkpoint names its exact predecessor, C5 adds no graph record, and every pre-C5 evidence derivation answers with a causal receipt. |
| `artifact.shared_admission` | One deny-class payload produces one shared class, decision, primary diagnostic, and leaf digest through every adapter, and each profile admits exactly its own source grammars. |

A path that stops proving an obligation fails the suite: the coverage matrix is
checked for completeness at the end, so a silently dropped case is an error
rather than a gap.

Positive source admission is deliberately *not* required to be identical across
paths. The accepted policy lets an adapter narrow its allowed source grammars,
so a Go or Python source fixture is legitimately opaque under the Rust profile.
What no adapter may do is admit a class the shared policy denies.

**3. The rejection matrix.** Sixteen vectors are driven through the delivered
adapters' own seams. Each requires the same three things: a stable diagnostic
from a closed set, no affected process, and no publication.

| Vector | Family | Paths proved here |
| --- | --- | --- |
| `binding-duplicate-record` | graph | npm, pnpm, Yarn Classic, modern Yarn, SwiftPM |
| `binding-dangling-reference` | graph | npm, pnpm, Yarn Classic, modern Yarn, SwiftPM |
| `binding-wrong-kind` | graph | npm, pnpm, Yarn Classic, modern Yarn, SwiftPM |
| `binding-replaces-capture` | graph | npm, pnpm, Yarn Classic, modern Yarn, SwiftPM |
| `binding-missing-target` | graph | npm, pnpm, Yarn Classic, modern Yarn, SwiftPM |
| `build-cycle` | graph | npm, pnpm, Yarn Classic, modern Yarn |
| `compiled-dependency-bytes` | artifact | all six |
| `opaque-dependency-bytes` | artifact | all six |
| `verified-binary-unavailable` | artifact | all six |
| `integrity-mismatch` | artifact | npm, pnpm, Yarn Classic |
| `offline-input-missing` | execution | npm, modern Yarn, SwiftPM |
| `target-identity-drift` | identity | Rust, SwiftPM |
| `toolchain-identity-drift` | identity | SwiftPM |
| `undeclared-process` | execution | SwiftPM |
| `undeclared-input` | execution | npm, pnpm, Yarn Classic, modern Yarn, SwiftPM |
| `unreceipted-output` | output | all six |

Three published vectors are delegated rather than driven here, because their
failing condition can only be constructed behind a live verified execution
provider or a sealed in-package authority. An integration package could only
reach them by forging evidence, which would prove nothing. Each names its
owning packages, and a compile-time reference to each owner's own diagnostic
constant keeps the delegation honest:

| Delegated vector | Owning accepted suites |
| --- | --- |
| `network-attempted` | `internal/closureexec`, `internal/npmsource`, `internal/rustsource`, `internal/swiftpmbuild` |
| `undeclared-write` | `internal/closureexec`, `internal/rustsource`, `internal/swiftpmbuild` |
| `output-drift` | `internal/artifactpolicy`, `internal/swiftpmbuild`, `internal/nodesource` |

**4. A protocol export for independent implementations.** The suite emits the
whole contract — the accepted corpus with independently derived identities, the
counters this package derived from it, the obligation set, the delivered paths,
and the rejection matrix — as exact CCJ-1 at
[`internal/crossconformance/testdata/cross-adapter-protocol-export.json`](../internal/crossconformance/testdata/cross-adapter-protocol-export.json).
The file is committed, so a change to the corpus, the obligations, or the
matrix is a reviewable diff rather than a silent protocol change. No Python
code is added to this repository; the export is the interface.

### Environment

The suite is portable except for the Rust path, which drives the real
`rust-source-v1` manager so that the cross-adapter proof consumes the same C0
Cargo registration, pinned vendor transform, and metadata receipts the accepted
Rust suite does. `rust-source-v1` admits exactly one approved Cargo descriptor
per native target, so the Rust cases need that toolchain present — the same
requirement `internal/rustsource` already carries. Nothing else in the suite
needs Node, npm, pnpm, Yarn, or Swift installed.

## Migrating an existing command

An existing skill command does not move onto a profile by being renamed. It has
to be able to produce the closure the profile requires.

1. **Pick the profile.** One command product, one entry point, one ecosystem.
   A command whose sources span two ecosystems needs a declared subprocess or
   FFI boundary, not one adapter.
2. **Produce an authoritative lock in the repository.** Rust needs a committed
   `Cargo.lock`; the Node managers need their own lock at the supported schema;
   SwiftPM needs a top-level `Package.resolved`, or Curator's brokered resolver
   generates one and freezes it as closure input. A dependency package's own
   lockfile is never authority.
3. **Remove every compiled dependency payload.** Prebuilt `.node` addons,
   native libraries, vendored frameworks, wheels with extension modules, and
   JVM or Python bytecode all reject. There is no allow flag: the dependency
   has to ship source, or the command has to drop it.
4. **Declare every generator.** A TypeScript build, a code generator, or any
   other step that turns admitted source into shipped bytes must name its tool,
   config, inputs, outputs, and target. Undeclared generation rejects with
   `artifact_generated_input_undeclared` or `closure_build_dependency_unlocked`.
5. **Remove dependency lifecycle scripts and manager extensions.** Materializa-
   tion always runs with lifecycle execution disabled. A dependency that cannot
   function without one is unsupported in the pure-source profile.
6. **Check the target matrix.** A closure is target-specific. Each supported
   destination gets its own binding, active graph, plan, and receipt; the
   capture is shared.

A command that cannot satisfy these stays on its current path and is recorded
as unsupported migration evidence. The profiles are not weakened to admit it.

## References

- Delivery input: [`.spec/skill-facing-cli-source-closure.md`](../.spec/skill-facing-cli-source-closure.md)
- Accepted architecture decision and its reviewed research inputs, on the task
  board under `TASK-260810-1dgdos`
- Per-profile behaviour in [`README.md`](../README.md) under "Execution assurance"
