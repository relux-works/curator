# SwiftPM mixed Swift/C-family source closure

Status: research decision prepared for review under `TASK-260810-zddzh7`.

Fact-check date: 2026-08-11.

## Context

Curator needs a source-closure adapter for skill-facing Swift command-line
products and for C, C++, Objective-C, and Objective-C++ code that Swift Package
Manager can represent. The governing invariant is stronger than “SwiftPM can
build it”: every transitive input must be immutable and auditable, the captured
graph must rebuild without network access, executable build-time extensions
must not add undeclared inputs, and precompiled dependency payloads must be
rejected.

This decision uses the repository's source-closure specification, accepted
artifact policy, and accepted inventory:

- [Skill-facing CLI source closure](../.spec/skill-facing-cli-source-closure.md)
- [Compiled artifact taxonomy and deny policy](260811_compiled-artifact-taxonomy-and-deny-policy.md)
- [Revision-pinned language and CLI inventory](260811_inventory-language-and-reference-surfaces.md)

The current inventory supplies real Swift CLI packages and real transitive C
targets (Yams, swift-atomics, swift-system, and swift-nio). It does not supply a
current skill CLI with a C++, Objective-C, or Objective-C++ package edge, so
those languages were checked with focused fixtures instead of inferred from the
C result.

## Executive decision

**Adopt SwiftPM as the preferred graph and build boundary for a restricted
`swiftpm-source-v1` profile. Do not treat unwrapped SwiftPM as the source-closure
security boundary.**

The profile supports source-only SwiftPM packages when Curator:

1. selects one explicit executable product and one exact target destination;
2. evaluates manifests in a controlled environment using a fingerprinted
   Swift/SwiftPM toolchain;
3. freezes the root `Package.resolved`, every source-control revision, every
   package snapshot, and the resolved package/target/source graph;
4. rewrites every source-control location to a captured local mirror and
   rebuilds from fresh cache and scratch paths with the original origins
   unavailable and external network denied;
5. keeps Swift and C-family code in separate SwiftPM targets, while allowing C,
   C++, Objective-C, and Objective-C++ files to coexist in a Clang target when
   the selected toolchain/destination fixture passes;
6. parses and confines headers and module maps, then verifies compiler-observed
   header/module reads against the captured source roots or the selected
   toolchain/SDK roots;
7. rejects every binary target and every compiled artifact in dependency
   payloads;
8. rejects active build-tool plugins and active macros in v1, never invokes
   command plugins, and disables SwiftPM's experimental prebuilt SwiftSyntax
   path; and
9. accepts a system-library edge only when it resolves wholly inside an
   independently selected and fingerprinted SDK/sysroot/toolchain boundary.

SwiftPM therefore covers the required languages conservatively **inside this
profile**. A separate C-family strategy is still required for CMake/Autotools or
other non-SwiftPM graphs, arbitrary host libraries, generated headers or custom
build scripts, per-file behavior SwiftPM cannot represent, and unvalidated
Objective-C runtimes/toolchains.

## Key findings

1. **Package-level language coverage is real; target-level mixing has a hard
   boundary.** Swift's PackageDescription documentation names Swift,
   Objective-C, Objective-C++, C, and C++ packages. On Apple Swift 6.3.2, one
   SwiftPM Clang target compiled nested `.c`, `.cpp`, `.m`, and `.mm` files and
   linked them into a Swift executable. A target containing both `.swift` and
   `.c` failed with SwiftPM's explicit “mixed language source files” error.
   Separate target edges are therefore normative.
2. **SwiftPM recursively discovers valid source files, not the entire security
   closure.** `swift package describe --type json` found deeply nested source
   files, but did not enumerate the custom module map, public headers, or a
   header reached through that module map. Curator must inventory the full
   package snapshot independently.
3. **A custom module map can escape the package.** SwiftPM 6.3.2 successfully
   built and ran a regular Clang target whose `module.modulemap` named an
   absolute header outside the package. This is a concrete undeclared-input
   path, not a theoretical limitation. Module-map parsing and compiler read-set
   verification are mandatory.
4. **The top-level lock is necessary but insufficient.** SwiftPM resolves the
   dependency graph recursively and a version-3 `Package.resolved` recorded
   both direct and transitive fixture pins with exact revisions. SwiftPM's own
   documentation says a dependency package's `Package.resolved` is ignored;
   only the top-level/leaf lock governs a build. The lock contains resolution
   state, not Curator's per-file SHA-256 source manifest, so both are required.
5. **`--force-resolved-versions` freezes resolution but is not offline mode.**
   It rejected a stale lock as required, yet a fresh build still fetched missing
   checkouts. SwiftPM has no security-grade `--offline` switch. `--skip-update`
   exists in the installed CLI but emitted a deprecation warning. Curator must
   enforce network=`none` outside SwiftPM and provide every pin through local
   mirrors.
6. **Recursive capture is observable and enforceable.** With both transitive
   repositories mirrored, a fresh-cache/fresh-scratch build succeeded after
   the original repositories were renamed out of reach. Removing only the
   transitive `B` mirror made the same build fail. This is the required
   fail-closed behavior.
7. **Mirror identity has kind-specific normalization.** A local
   `file://.../A` dependency was pinned as `localSourceControl` at an absolute
   filesystem location. A mirror keyed only by the raw `file://` spelling was
   not applied; a `file://` mirror value for the normalized local kind was then
   rejected as an invalid absolute path. The working mapping used the resolved
   absolute path as both original-kind key and local mirror value. Audit data
   must retain raw manifest location, canonical SwiftPM location, kind, and the
   generated mirror entry.
8. **Direct Swift/C++ interoperability is available but version-sensitive.** A
   Swift target with `.interoperabilityMode(.Cxx)` imported a C++ struct and
   called its method successfully. Swift documents the feature as evolving and
   notes that Swift targets in the dependency chain must also enable C++
   interoperability. Pin the compatibility/toolchain profile; do not infer
   support for every C++ API shape from one build.
9. **Objective-C and Objective-C++ work on the tested Darwin boundary.**
   Separate `.m` and `.mm` targets exposed Objective-C classes to Swift and
   linked through the selected macOS Foundation SDK. This proves the Apple
   Swift 6.3.2/macOS 26.5 profile only. Other destination/toolchain/runtime
   combinations remain unsupported until their fixture passes.
10. **Plugins and macros are host executables, not passive metadata.** SwiftPM
    runs plugins as separate processes; build plugins can generate sources and
    command plugins can request network or package-directory writes. Macro
    implementations run while compiling clients. Active instances need a
    future declared-read/write execution contract and are rejected in v1.
11. **SwiftPM now has a prebuilt macro fast path.** The installed 6.3.2 help
    reports experimental prebuilts enabled by default. SwiftPM's merged change
    describes downloading and extracting prebuilt SwiftSyntax libraries for
    macros. Every Curator resolve, describe, build, test, and run invocation
    must pass `--disable-experimental-prebuilts`, even though active macros are
    initially rejected.
12. **System-library targets are declarations of external trust, not source
    capture.** They can consult `pkg-config`, module maps, provider hints, and
    host libraries. Provider metadata must never auto-install or confer trust.
    Only paths already inside a selected SDK/sysroot/toolchain may be used.

## Supported profile

Status terms:

- **Supported**: admitted when all listed gates pass.
- **Restricted**: admitted only for the named toolchain, destination, or graph
  shape and corresponding conformance evidence.
- **Unsupported**: fail before build or extension execution.

| Package or target shape | Status in `swiftpm-source-v1` | Rule |
| --- | --- | --- |
| Swift-only executable/library targets | Supported | Exact source snapshot, explicit product, lock/graph replay, trusted Swift toolchain. |
| Swift target depending on C target | Supported | Separate targets; headers/modules confined and audited. This matches the real inventory. |
| One Clang target containing C, C++, Objective-C, and Objective-C++ | Restricted | Allowed on a destination/toolchain where the four-language fixture passes; never mix Swift source into it. |
| C-only target | Supported | Clang, headers, C standard, include graph, and target triple are checkpointed. |
| C++ target behind a C-compatible public shim | Supported | Clang++/standard library/toolchain identity and C++ standard are checkpointed. |
| Direct Swift import of C++ API | Restricted | Tools version supports it, every relevant Swift target enables `.interoperabilityMode(.Cxx)`, exact compatibility/toolchain is pinned, API fixture passes. |
| Objective-C target imported by Swift | Restricted | Initial baseline is Apple/Darwin SDK and runtime only; framework/system edges must remain inside the selected SDK. |
| Objective-C++ implementation imported through C/Objective-C headers | Restricted | Initial baseline is Apple/Darwin only; direct C++ exposure additionally follows C++ interoperability restrictions. |
| Swift and any C-family source in the same target | Unsupported | SwiftPM rejects the shape; require target separation or fail `swiftpm_mixed_language_target_unsupported`. |
| Recursive target source discovery | Supported as build metadata | Never used as the full closure inventory; reconcile it with the independently hashed package tree. |
| Conventional public headers and SwiftPM-generated module map | Supported | Hash every header; reproduce the documented umbrella-header/directory rules. |
| Custom `module.modulemap` | Restricted | Parse the full module-map grammar; every header, umbrella, extern module, framework, link, and config reference must resolve to closure or trusted SDK roots. |
| In-root `.package(path:)` dependency | Restricted | Canonical real path must remain inside an explicitly captured workspace root; hash it as a separate package node because it has no lock pin. |
| Absolute/out-of-root local package dependency | Unsupported | Mutable undeclared input; `swiftpm_local_dependency_outside_closure`. |
| Remote source-control dependency using version/range/branch/revision | Supported after acquisition | Root lock must record an exact revision; capture source snapshot and local mirror. Branch/tag names are metadata, not offline identity. |
| Git submodule, LFS pointer, checkout filter, or required Git hook | Unsupported in v1 | These introduce another fetch/execution plane. Add explicit graph support later or fail closed. |
| Swift package registry dependency | Unsupported in source-control v1 | Do not transform SCM to registry. A later registry profile can capture signed/checksummed source archives under the shared archive policy. |
| Version-specific `Package@swift-*` manifests/tags | Restricted | Capture every manifest, selected manifest/tag, tools-version selection, and exact toolchain; replay selection offline. |
| Source-like resources | Restricted | Must pass the central artifact/source grammar profile. Opaque binary resources remain rejected while the v1 inert-data allowlist is empty. |
| System-library target inside selected SDK/sysroot | Restricted | Parse module map and sanitized `pkg-config` output; fingerprint every external file/tool. |
| Homebrew/apt/provider hint or arbitrary host `pkg-config` result | Unsupported | A provider is a hint, not provenance; `artifact_toolchain_untrusted`. |
| Safe `cSettings`/`cxxSettings`/`swiftSettings`/`linkerSettings` | Restricted | Normalize conditions, traits, paths, defines, linked SDK libraries/frameworks, and target configuration. |
| Any `.unsafeFlags` setting | Unsupported in v1 | Arbitrary flags can add inputs, tools, plugins, response files, or output paths. |
| Dormant plugin/macro declarations unreachable from selected product | Restricted | Source remains scanned and hashed; Curator never invokes command plugins and proves no selected target reaches the extension. |
| Active build-tool plugin | Unsupported in v1 | Reject before compiling/running the plugin or its tool dependencies. |
| Active macro target | Unsupported in v1 | Reject before macro compilation/execution or prebuilt retrieval. |
| Any local or remote `binaryTarget`, referenced or dormant | Unsupported | `artifact_binary_admission_unavailable`; checksum does not create verified-binary admission. |
| `.xcframework`, `.framework`, `.a`, `.dylib`, `.so`, object, `.swiftmodule`, or other compiled dependency payload | Unsupported | Shared recursive artifact policy returns `artifact_compiled_dependency_forbidden`. |
| Native SwiftPM build system | Supported | Pin `--build-system native`; preview `swiftbuild` and discouraged `xcode` modes are separate profiles. |
| Unvalidated target triple/SDK/toolset | Unsupported | No host-default fallback; `swiftpm_target_platform_unsupported`. |

## Why SwiftPM metadata is not enough

### `Package.resolved`

SwiftPM recursively walks package requirements and writes resolution state to
the top-level `Package.resolved`. Official documentation also makes two limits
explicit:

- dependency packages' own lockfiles do not pin the graph when they are used by
  another package; and
- ordinary commands can implicitly resolve unless forced to use the lock.

Curator's rules are therefore:

1. A checked-in lock may seed acquisition, but is not trusted until every pin
   is fetched, scanned, and matched.
2. If the root CLI package has no lock, acquisition may generate one in a
   disposable namespace. The generated bytes become required closure input;
   build may not continue in the same unfrozen resolver state.
3. The accepted root lock must have a supported schema, deterministic pin set,
   and a SHA-256 digest. Every source-control pin must carry exact revision,
   resolved kind/location, and any version/branch metadata.
4. Each pin maps one-to-one to a captured source snapshot with a canonical
   per-path SHA-256 manifest and a read-only local mirror containing the pinned
   commit. A Git revision alone is not Curator's artifact manifest.
5. `swift package show-dependencies --format json`, all selected manifests, and
   the lock must agree on identity, origin, version, revision, and recursive
   edges.
6. `--force-resolved-versions` is mandatory. An out-of-date lock is rejection,
   never permission to update during build.

The fixture lock had schema version 3 and SHA-256
`41ea90d68498020de5580bf6d79cc9059ea8e822130428bfacfa90b651cf5453`.
It pinned:

| Identity | Version | Commit | Git tree | Deterministic fixture source-archive SHA-256 |
| --- | --- | --- | --- | --- |
| `a` | `1.0.0` | `c55f94ba8a12de63ae9ad3eb57d9b9d482a89f37` | `86b24c104ad2dfe24d11790ab154a680cf84e981` | `6cb61191ca0e95c033c0439808989e6321042aba7f3e4ed4dde1cb7e4bb1d857` |
| `b` | `1.0.0` | `7d3a8c1fc0032519af8c4faabdda21b185e6c552` | `73dcba1fe8c5b25c56aaf16368eccf0fbe466253` | `13448d3f8e864ae6910ce13308363582a969c741eaf04b59f10020fd0ba3eecf` |

Production identity uses the shared canonical artifact manifest, not a tar
encoding; the archive hashes above make this focused fixture independently
repeatable.

### Manifest execution and selection

`Package.swift` is executable Swift, not declarative JSON. SwiftPM also selects
version-specific manifests such as `Package@swift-6.3.swift` according to the
active toolchain. Curator must:

- capture the unversioned and every version-specific manifest before
  evaluation;
- use a selected, fully fingerprinted PackageDescription runtime;
- run manifest evaluation with external network denied, a sanitized
  environment, a read-only package snapshot, and no host home/config access;
- expose only the package snapshot and selected toolchain/SDK roots, making any
  other filesystem read fail;
- record the selected manifest, tools version, arguments, environment, stdout/
  diagnostics, and normalized `dump-package` result; and
- repeat evaluation during offline replay and reject any normalized graph
  difference.

For acquisition, source-control I/O belongs to a manager-owned broker. The
broker fetches allowed canonical origins into quarantine without Git hooks or
checkout filters, recursively adds mirrors, and returns immutable snapshots to
the resolver. If the first implementation temporarily uses an online resolver
namespace, that result is only a candidate: it has no admission value until the
same lock and graph replay from local mirrors under network=`none`.

### Source, header, and module closure

SwiftPM documents recursive search for valid target sources and automatic
module-map generation for conventional C-family header layouts. That is useful
build metadata but omits security-relevant files.

For each captured package, Curator must independently enumerate and hash:

- every selected and alternate manifest;
- all target sources, headers, module maps, resources, configuration, license,
  and source-like generated inputs in the admitted Git tree;
- target path, `exclude`, explicit `sources`, `publicHeadersPath`, resources,
  language standards, settings, conditions, traits, products, and target
  edges; and
- every include/module import observed by the actual compiler.

The module-map parser must reject absolute or escaping package references,
symlinks, `extern module` paths outside approved roots, undeclared framework or
library edges, and ambiguous/malformed maps. C-family dependency files or a
trusted compiler dependency scan then prove that every read is under one of:

1. the package's captured source root;
2. another declared captured package root; or
3. the selected fingerprinted toolchain/SDK/sysroot.

An include present only because of `CPATH`, `C_INCLUDE_PATH`,
`CPLUS_INCLUDE_PATH`, `OBJC_INCLUDE_PATH`, a response file, or arbitrary
`-I`/`-F`/`-L` flag is undeclared. Curator clears those variables and rejects
unsafe flags before compilation.

## Plugins, macros, and generation

SwiftPM distinguishes build-tool and command plugins. Official documentation
states that plugins are separate processes; sandboxing is platform-dependent,
all plugins can write temporary data, and command plugins can request network
or package-directory writes. Build-tool dependencies may themselves be source
executables or binary targets. Macros are also executable compiler plugins that
run while their clients compile.

The v1 decision is deliberately narrow:

- allow source declarations for unreachable/dormant plugins and macros so that
  packages such as swift-argument-parser are not rejected merely for vending
  optional command plugins;
- select the CLI product explicitly and prove no selected target transitively
  uses a plugin or macro target;
- never invoke `swift package <plugin-command>`;
- reject any active build-plugin usage before plugin compilation;
- reject any active macro dependency before macro compilation;
- reject every binary target even if it is only a plugin tool; and
- pass `--disable-experimental-prebuilts` to every SwiftPM command.

A future restricted extension profile needs a separate host-execution node for
each plugin/macro: captured source and tool dependencies, host toolchain/triple,
network=`none`, immutable read roots, empty declared output root, complete
command/environment, observed reads/writes, generated-file hashes, and causal
receipts. SwiftPM's own sandbox alone is not that proof.

## System libraries and SDK edges

A `.systemLibrary` target supplies a module map and may invoke `pkg-config` to
obtain include and link paths. Package provider entries (for example Homebrew)
are installation hints, not immutable package inputs or proof of trust.

The only v1 admission is an `external_toolchain` edge already selected by
Curator:

1. run a fingerprinted `pkg-config` only if required;
2. give it an isolated configuration path containing manager-approved `.pc`
   files;
3. normalize its output and reject shell fragments or paths outside the
   selected SDK/sysroot/toolchain;
4. parse the system module map and apply the same containment rules; and
5. record every header, library, framework, version, target ABI, and content
   fingerprint in the toolchain checkpoint.

Curator never runs package-manager provider installation. An arbitrary
`/usr/local`, `/opt/homebrew`, Homebrew Cellar, apt, or user-provided
`PKG_CONFIG_PATH` result is `artifact_toolchain_untrusted` until a separate
external-dependency admission capability selects and fingerprints it.

## Toolchain and target identity

The exact toolchain affects manifest selection, source compatibility, C++
interop, plugin/macro host builds, SDK modules, link results, and default
deployment targets. A version string alone is too weak.

The observed fact-check host was:

| Component | Observed identity |
| --- | --- |
| Swift / SwiftPM | Apple Swift 6.3.2; SwiftPM 6.3.2; `swiftlang-6.3.2.1.108` |
| Swift target | `arm64-apple-macosx26.0`; runtime compatibility 6.2 |
| Clang | Apple Clang 21.0.0, `clang-2100.1.1.101` |
| Xcode | 26.5, build `17F42` |
| macOS SDK | 26.5, build `25F70` |
| SDK path | `/Applications/Xcode_26_5.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX26.5.sdk` |
| Build tools | `swiftc`, `clang`, `clang++`, and `ld` under the selected Xcode default toolchain |
| Host | Darwin 25.5.0, arm64 |

The production checkpoint records at least:

- SwiftPM and adapter version/content fingerprints;
- `swiftc`, PackageDescription runtime, `clang`, `clang++`, assembler if used,
  linker, archiver, libc/C++ runtime, and their selected root tree digest;
- Xcode/Swift SDK/toolset identifier, SDK/sysroot tree digest, platform,
  architecture, triple, ABI, deployment minimum, and runtime compatibility;
- host triple separately from target triple (required for any future
  plugin/macro support);
- tools version, selected versioned manifest, Swift language mode, C and C++
  standards, C++ interoperability mode, configuration, traits, sanitizer/index
  choices, build-system kind, jobs only if output-sensitive, and all explicit
  build flags;
- generated mirror configuration and isolated cache/config/security/scratch
  roots;
- sanitized `PATH`, `DEVELOPER_DIR`, `SDKROOT`, deployment variables,
  compiler overrides, include/library/framework variables, locale/timezone, and
  an explicit allowlist of remaining environment; and
- a time-of-use recheck before manifest evaluation, planning, compilation, and
  protected output publication.

Target conditions and minimum platforms are graph inputs. SwiftPM validates
dependency deployment compatibility, but Curator must also bind the actual
destination that selects `.when(platforms:)`, build configuration, and traits.
No implicit host destination or user SwiftPM configuration is accepted.

## Recursive capture and offline rebuild procedure

### Checkpoint 0 — policy and root

1. Select `swiftpm-source-v1`, shared artifact policy version, explicit CLI
   product, build configuration, traits, target triple/SDK, and native build
   system.
2. Freeze the root package bytes and run the shared recursive artifact scanner.
3. Reject unsafe nodes, compiled payloads, opaque disallowed resources, Git
   submodules/LFS/filters/hooks, and source paths outside the root.

### Checkpoint 1 — controlled acquisition

1. Evaluate the selected root manifest in the controlled manifest sandbox.
2. Broker each source-control fetch into quarantine; never give manifests or
   package hooks general network access.
3. Before evaluating a fetched package manifest, scan and freeze its full
   pinned source tree.
4. Continue until the resolver produces the root lock and recursive graph.
5. Reject registries, unconfined local paths, source-control kind changes,
   undeclared authentication helpers, binary targets, and unsupported
   executable extensions.

### Checkpoint 2 — immutable source closure

Commit one canonical `swiftpm-source-manifest-v1` containing:

- root lock bytes/digest/schema/origin hash and normalized pin records;
- raw and canonical origin, source-control kind, version/branch requirement,
  exact revision, mirror identity, and captured snapshot digest for every
  package;
- every package's complete artifact manifest digest;
- every selected/alternate Package.swift digest and selected tools version;
- normalized `dump-package`, `show-dependencies`, product/target graph, actual
  recursive sources, all headers/module maps/resources, and their digests;
- language per source/target, FFI edges, build order, settings/conditions,
  traits, plugin/macro reachability, system/SDK edges, and policy verdicts; and
- a canonical digest over the entire record.

### Checkpoint 3 — offline replay and planning

1. Create fresh empty scratch, cache, config, security, and output roots.
2. Generate mirrors from every canonical origin to a read-only captured local
   repository of the same SwiftPM source-control kind.
3. Make original origins unavailable and enforce external network=`none` at the
   process boundary.
4. Disable netrc, keychain, SCM-to-registry transformation, experimental
   prebuilts, user configuration, and automatic resolution. Use the exact root
   lock with `--force-resolved-versions`.
5. Re-run manifest selection, `dump-package`, recursive graph resolution, and
   source inventory. Every canonical digest must match Checkpoint 2.
6. Resolve the selected product's target graph and verify no active plugin,
   macro, binary target, unsafe setting, or untrusted system edge.
7. Commit the build-plan digest before compiler execution.

`--skip-update` is not part of the normative command because SwiftPM 6.3.2
warns that it is deprecated. Local mirrors plus an externally enforced network
deny are the durable controls.

### Checkpoint 4 — build and receipt

1. Build only the selected product from the read-only source closure into an
   empty output/scratch namespace.
2. Capture normalized compiler/linker commands, environment, target and host
   toolchains, compiler dependency files/read set, and write set.
3. Reject any read outside captured source or trusted toolchain roots and any
   undeclared generated input/output.
4. Record every locally built intermediate/output class, path, size, and
   SHA-256; bind them to the source, graph, policy, toolchain, target, and build
   plan digests.
5. Publish only through the manager's protected cache and validate exact input,
   size, digest, class, and receipt on reuse.

## Stable diagnostics

Use the accepted shared `artifact_*` vocabulary whenever it already describes
the failure. The SwiftPM adapter adds only graph-specific codes.

| Code | Required condition |
| --- | --- |
| `swiftpm_resolution_unfrozen` | Build was requested before a root lock and captured recursive graph were committed. |
| `swiftpm_resolved_file_out_of_date` | Manifest/traits/config requirements disagree with the frozen lock; automatic update is forbidden. |
| `swiftpm_dependency_pin_mismatch` | Lock, resolved graph, mirror commit, or snapshot digest disagree for a package identity. |
| `swiftpm_dependency_origin_unsupported` | Registry, source-control transformation, unsupported transport/kind, submodule/LFS/filter, or other unmodeled origin. |
| `swiftpm_dependency_mirror_missing` | Any direct or transitive lock pin lacks a local captured mirror. |
| `swiftpm_local_dependency_outside_closure` | A path dependency resolves outside an admitted workspace root or is mutable. |
| `swiftpm_manifest_replay_drift` | Selected manifest, normalized dump, diagnostics, or package graph differs under offline replay. |
| `swiftpm_source_inventory_drift` | SwiftPM's actual target/source graph differs from the captured source manifest. |
| `swiftpm_mixed_language_target_unsupported` | One target contains Swift and C-family sources. |
| `swiftpm_unsafe_build_setting_forbidden` | Any selected target uses `unsafeFlags` or an equivalent unmodeled flag path. |
| `swiftpm_modulemap_escape` | Module map references an absolute, escaping, untrusted, or otherwise undeclared input. |
| `swiftpm_header_input_undeclared` | Compiler-observed include/module read is outside closure and trusted toolchain roots. |
| `swiftpm_plugin_execution_unsupported` | Selected graph would compile/run a build or command plugin in v1. |
| `swiftpm_macro_execution_unsupported` | Selected graph would compile/run a macro implementation in v1. |
| `swiftpm_target_platform_unsupported` | Destination/toolchain/language/SDK combination has no accepted profile/fixture. |
| `swiftpm_offline_rebuild_failed` | Frozen local-mirror replay or build tried an unavailable input or otherwise failed. |
| `swiftpm_build_graph_drift` | Planned commands, target order, host/target edges, system edges, or FFI graph differs from checkpoint. |
| `artifact_binary_admission_unavailable` | Any binary target requests absent verified-binary capability. |
| `artifact_compiled_dependency_forbidden` | Recursive payload scan finds compiled native/VM/IR/module bytes. |
| `artifact_generated_input_undeclared` | Generated source/header/resource has no accepted generator lineage. |
| `artifact_toolchain_untrusted` | System library, SDK, `pkg-config`, compiler, linker, or host path lacks external-toolchain evidence. |
| `artifact_toolchain_identity_changed` | Selected toolchain/SDK bytes or identity change between checkpoint and use. |

Every rejection includes package identity, target/product path, manifest and
lock digests, target destination, relevant virtual path, and the rule-specific
observations. No rejected case executes a plugin/macro/build or publishes an
output/cache record.

## Required conformance fixtures

### Language and target shapes

| ID | Fixture | Expected result |
| --- | --- | --- |
| `S01` | Swift executable only | Supported, source-only build and receipt. |
| `S02` | Swift target depends on nested C target with generated module map | Supported; all sources/headers and edge inventoried. |
| `S03` | One Clang target contains nested `.c`, `.cpp`, `.m`, `.mm`; Swift client uses C ABI | Supported on accepted Darwin profile; all four compiler invocations observed. |
| `S04` | Swift and C source in one target | Expected rejection: `swiftpm_mixed_language_target_unsupported`; SwiftPM itself exits non-zero. |
| `S05` | Swift imports C++ struct with `.interoperabilityMode(.Cxx)` | Restricted pass with exact compatibility/toolchain identity. |
| `S06` | One dependent Swift target omits required C++ mode | Expected fail closed; no implicit flag propagation. |
| `S07` | Swift imports Objective-C class from `.m` target | Restricted Darwin pass. |
| `S08` | Swift imports Objective-C class implemented in `.mm` using C++ | Restricted Darwin pass with Objective-C++ and C++ toolchain edges. |
| `S09` | Same Objective-C/Objective-C++ package on unvalidated triple | `swiftpm_target_platform_unsupported`. |
| `S10` | Deeply nested valid sources plus explicit `sources`/`exclude` variants | SwiftPM inventory exactly reconciles with canonical package tree. |

### Headers, module maps, and system inputs

| ID | Fixture | Expected result |
| --- | --- | --- |
| `H01` | Conventional umbrella header layout | Generated module map reproduced and all headers hashed. |
| `H02` | Valid custom module map wholly inside package | Supported with parsed reference manifest. |
| `H03` | Custom module map names absolute out-of-package header | `swiftpm_modulemap_escape` before build; empirical raw SwiftPM would otherwise succeed. |
| `H04` | Source includes absolute or `../` header outside package | `swiftpm_header_input_undeclared`. |
| `H05` | Header appears only through environment include path/unsafe flag | Environment is cleared or `swiftpm_unsafe_build_setting_forbidden`. |
| `H06` | System target points into selected SDK and sanitized `.pc` data | Restricted pass with external-toolchain checkpoint. |
| `H07` | System target resolves to Homebrew/user/arbitrary host path | `artifact_toolchain_untrusted`; provider hint is not executed. |
| `H08` | Module map declares untrusted linked library/framework | `swiftpm_modulemap_escape` or `artifact_toolchain_untrusted`. |

### Resolution and offline closure

| ID | Fixture | Expected result |
| --- | --- | --- |
| `R01` | Root → A → B source-control graph | Root lock and graph contain both exact revisions; snapshot manifest has every package. |
| `R02` | No checked-in root lock | Acquisition generates and freezes lock; build cannot start before checkpoint. |
| `R03` | Dependency package contains a conflicting own lock | Its lock is captured as text but ignored for graph identity; root lock controls. |
| `R04` | Frozen lock no longer satisfies root manifest | `swiftpm_resolved_file_out_of_date`; expected non-zero force-resolved gate. |
| `R05` | All pins mirrored, original origins unavailable, fresh cache/scratch, network none | Offline build and CLI execution pass with identical graph. |
| `R06` | Only direct A mirror captured; transitive B mirror absent | `swiftpm_dependency_mirror_missing`/offline build non-zero before output. |
| `R07` | Raw `file://` location normalizes to local absolute-path source-control kind | Audit retains both forms; generated mirror key/value preserve resolved kind. |
| `R08` | Mirror commit differs from lock revision | `swiftpm_dependency_pin_mismatch`. |
| `R09` | Branch/tag moves after capture | Offline replay still uses exact captured revision; remote state is irrelevant. |
| `R10` | In-root versus absolute out-of-root path dependencies | In-root captured node passes; out-of-root path rejects. |
| `R11` | Registry or SCM-to-registry replacement | `swiftpm_dependency_origin_unsupported` in v1. |
| `R12` | Version-specific manifest/toolchain variants | Selected manifest and graph match exact toolchain checkpoint; drift rejects. |
| `R13` | Git submodule, LFS pointer, or checkout filter | Fail closed before materialization/execution. |

### Executable extensions and compiled artifacts

| ID | Fixture | Expected result |
| --- | --- | --- |
| `P01` | Dependency declares dormant command/build plugins; selected product does not reach them | Source captured; no plugin compile/run; product build passes. |
| `P02` | Selected target uses build-tool plugin | `swiftpm_plugin_execution_unsupported` before plugin/tool build; marker output absent. |
| `P03` | Attempt to invoke command plugin, including requested network/write permission | Same rejection; command never runs. |
| `P04` | Selected target depends on macro | `swiftpm_macro_execution_unsupported` before macro/prebuilt activity. |
| `P05` | SwiftPM invocation omits `--disable-experimental-prebuilts` | Adapter command-policy test fails before launch. |
| `P06` | Local/remote, referenced/dormant `binaryTarget` | `artifact_binary_admission_unavailable` before artifact fetch/use. |
| `P07` | Renamed/nested `.xcframework`, framework, object, library, Swift module | Shared recursive detector returns compiled-dependency rejection. |
| `P08` | Plugin tool is a binary target | Binary rejection dominates plugin handling. |
| `P09` | Generated header/source appears without approved lineage | `artifact_generated_input_undeclared`. |

Every negative vector also asserts: original package/build extension did not
execute, external network was unavailable, output namespace stayed empty, and
no protected cache publication occurred.

## Empirical fact-check record

Fixtures live under `.temp/swiftpm-research-fixtures/`; they are supporting
evidence, not the durable product. Commands were run directly as standalone
processes unless the two archive-hash pipelines explicitly used zsh
`pipefail`.

| Command/gate | Exit | Evidence |
| --- | ---: | --- |
| `swift --version` | 0 | Apple Swift 6.3.2, arm64 macOS target. |
| `clang --version` | 0 | Apple Clang 21.0.0 from Xcode 26.5. |
| `swift package --help` and `swift build --help` | 0 each | Force-resolved, isolated path, target/toolchain, native build-system, and experimental-prebuilt flags inspected. |
| `swift build` on `mixed-family` | 0 | Compiled nested C, C++, Objective-C, and Objective-C++ sources plus Swift client. |
| `mixed-family` executable | 0 | Printed `39`. |
| `swift package describe --type json` on `mixed-family` | 0 | Reported both Clang targets and all five nested source paths; did not report headers/module map. |
| `swift build` on `mixed-swift-c-rejected` | **1, expected failure** | SwiftPM reported that a target with mixed Swift and C-family sources is unsupported. |
| `swift build` and executable on `swift-cxx-interop` | 0 / 0 | Direct C++ struct import returned `42`. A sibling scratch-path spelling produced a non-fatal missing-PCM warning on one run; a package-contained scratch path rebuilt cleanly. |
| `swift build` and executable on `swift-objc-objcpp` | 0 / 0 | `.m` and `.mm` targets exposed Objective-C APIs to Swift and printed `36`. |
| `swift package dump-package` on `policy-rejections` | 0 | Deterministically exposed target types `binary`, `plugin`, `macro`, and `system` before build. |
| `swift build` and executable on `modulemap-escape` | 0 / 0 | Raw SwiftPM consumed an absolute header outside package and printed `73`, proving the boundary gap. |
| `swift package describe --type json` on `modulemap-escape` | 0 | Listed only `dummy.c`; omitted the escaping module-map header. |
| Initial recursive `swift package resolve` | 0 | Root → A → B resolved and wrote both exact pins. |
| Forced `show-dependencies --format json` | 0 | Reported A and transitive B; `--skip-update` warned that it is deprecated. |
| Fresh-cache offline-mirror build with both A and B, originals unavailable | 0 | Fetched only captured local mirrors, built Root. |
| Offline Root executable | 0 | Printed `41`. |
| Offline build after removing only B mirror | **1, expected failure** | B original was unavailable; transitive omission failed closed. |
| Forced build after changing root requirement to exact 2.0.0 | **1, expected failure** | Reported out-of-date resolved file with automatic resolution disabled. Requirement was restored afterward. |
| First mirror replay keyed only by raw `file://` spellings | **1, investigative failure** | SwiftPM normalized pins to absolute local locations and bypassed those keys. |
| Second replay with local original key but `file://` mirror value | **1, investigative failure** | SwiftPM rejected a kind-mismatched mirror value as invalid absolute path. |
| Third replay with canonical local-path key/value | 0 | Proved the required kind-preserving mirror mapping. |
| `git fsck --full` on captured A and B mirrors | 0 each | Captured repository object graphs valid. |
| `git archive ... \| shasum -a 256` under zsh `pipefail` | 0 each | Fixture snapshot hashes recorded above. |

The missing-PCM warning is not used as success evidence; the independent clean
rebuild and executable run are the passing C++ gates. The two mirror failures
are recorded as failures and directly inform the canonical-origin checkpoint.

## Primary-source fact check

Claims about SwiftPM and language behavior were checked against these project
or vendor primary sources:

- SwiftPM [PackageDescription overview](https://docs.swift.org/swiftpm/documentation/packagedescription/)
  for the five package languages and manifest model.
- SwiftPM [Target API](https://docs.swift.org/swiftpm/documentation/packagedescription/target/)
  for target kinds, headers, binary targets, system targets, plugins, and build
  settings.
- SwiftPM [C language target guide](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/creatingclanguagetargets/)
  for public-header layout and automatic/custom module maps.
- SwiftPM [`sources` API](https://docs.swift.org/swiftpm/documentation/packagedescription/target/sources/)
  for recursive valid-source discovery.
- SwiftPM [dependency resolution guide](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/resolvingpackageversions/)
  for top-level lock behavior, dependency lockfiles being ignored, implicit
  resolution, and forced resolved versions.
- SwiftPM [dependency failure guide](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/resolvingdependencyfailures/)
  for recursive graph traversal and `show-dependencies`.
- SwiftPM [adding dependencies](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/addingdependencies/)
  for remote/local/system/binary dependency forms.
- SwiftPM [version-specific packaging](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/swiftversionspecificpackaging/)
  for manifest and tag selection by toolchain version.
- SwiftPM [system-library guide](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/addingsystemlibrarydependency/)
  for module maps and `pkg-config` behavior.
- SwiftPM [plugin overview](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/plugins/),
  [build-tool plugin guide](https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/writingbuildtoolplugin/),
  and [plugin permissions](https://docs.swift.org/swiftpm/documentation/packagedescription/pluginpermission/)
  for executable extension behavior, tool dependencies, sandbox portability,
  and network/write permissions.
- The Swift book [Macros](https://docs.swift.org/swift-book/documentation/the-swift-programming-language/macros/)
  for macro implementation execution during client compilation.
- Swift.org [C++ interoperability guide](https://www.swift.org/documentation/cxx-interop/),
  [SwiftPM setup](https://www.swift.org/documentation/cxx-interop/project-build-setup/),
  and [current constraints](https://www.swift.org/documentation/cxx-interop/status/)
  for Clang modules, opt-in mode, evolving compatibility, and dependency-chain
  constraints.
- SwiftPM's merged [prebuilt SwiftSyntax change](https://github.com/swiftlang/swift-package-manager/pull/8142)
  and [Swift 6.2 release entry](https://github.com/swiftlang/swift-package-manager/releases/tag/swift-6.2-RELEASE)
  for the macro prebuilt download/extraction path.
- SwiftPM [CHANGELOG](https://github.com/swiftlang/swift-package-manager/blob/main/CHANGELOG.md)
  for tools-version features, offline cache behavior, plugins, macros, C++
  interoperability, and Swift SDK evolution.

Policy recommendations—especially the narrower v1 support surface—are Curator
decisions derived from those facts and the accepted repository security
invariants. No source says SwiftPM itself provides Curator's immutable snapshot,
external network denial, module-map containment, or protected output receipt.

## Gaps requiring a separate C-family strategy

SwiftPM should not be stretched into a generic native-build adapter. Route the
following to a separately approved C-family strategy or reject them:

1. C/C++/Objective-C/Objective-C++ projects without a SwiftPM manifest and
   target graph.
2. CMake, Meson, Autotools, Make, Bazel, custom shell generators, or package
   manager recipes not modeled as SwiftPM targets.
3. Arbitrary host libraries, package-provider installation, non-SDK
   `pkg-config` roots, and libraries whose source cannot be captured and built.
4. Required submodules, Git LFS, checkout filters, hooks, or opaque generated
   vendor trees until those are explicit closure nodes.
5. Generated headers/sources/resources, code generators, active plugins, or
   macros without declared read/write execution receipts.
6. Per-file compiler/link behavior, nonstandard source extensions, assembler
   or other languages, custom link ordering, linker scripts, or response files
   outside the accepted SwiftPM setting model.
7. Direct C++ API shapes outside the pinned interop compatibility surface.
8. Objective-C/Objective-C++ destinations without an accepted compiler,
   runtime, SDK, and conformance fixture.
9. Any package requiring prebuilt libraries/frameworks/modules. Those remain
   rejected until the separate verified-binary capability exists; a C-family
   adapter may not weaken this rule.

The separate strategy must reuse the same artifact policy, source/toolchain
checkpoints, network denial, and protected receipts. It is not an escape hatch
for binaries or undeclared inputs.

## Implementation-ready recommendations

1. Implement `swiftpm-source-v1` as a wrapper around SwiftPM's manifest/graph/
   native-build behavior, not as a call to unrestricted `swift build`.
2. Add a source-control acquisition broker and kind-preserving local mirror
   generator; capture raw and canonical origins plus exact revisions.
3. Add per-package manifest evaluation and canonical `dump-package`/
   `show-dependencies`/source-graph records under a controlled environment.
4. Add a Clang module-map parser and post-plan compiler read-set verifier.
5. Add selected-product reachability analysis for plugin/macro/dormant targets,
   binary targets, unsafe settings, and system-library edges.
6. Extend shared build metadata with Swift/Clang host and target toolchains,
   SDK/sysroot, language/interoperability modes, traits, graph/build-plan
   digests, and sorted multi-output receipts.
7. Land the conformance fixture families above before admitting a real package.
   Use the inventory's iOS-testing Swift→C graph as the first realistic case,
   the currency scraper as the missing-lock plus dormant-plugin case, and the
   focused C++/Objective-C/Objective-C++ fixtures for conditional languages.
8. Create separate backlog for non-SwiftPM C-family graphs instead of adding
   unsafe flags or plugin exceptions to this adapter.

## Acceptance mapping

| Assigned requirement | Evidence in this document |
| --- | --- |
| Determine conservative coverage for Swift, C, C++, Objective-C, Objective-C++ | Executive decision, supported-profile matrix, language fixtures `S01`–`S10`, and empirical builds. |
| Define complete recursive source capture | `Package.resolved` rules, manifest/source closure, Checkpoints 0–4, `swiftpm-source-manifest-v1`. |
| Assess Package.resolved and source-control dependencies | Lock section, recursive A→B fixture, local-mirror normalization finding, resolution fixtures `R01`–`R13`. |
| Assess headers and module maps | Source/header section, proven absolute module-map escape, fixtures `H01`–`H08`. |
| Assess system-library targets | System-library section and trusted SDK-only matrix decision. |
| Assess plugins and macros | Extension section, dormant-versus-active reachability rule, host-execution gap, `P01`–`P05`. |
| Reject binary targets and compiled payloads | Matrix, shared diagnostics, `P06`–`P08`. |
| Bind toolchain, SDK, platform, and target resolution | Toolchain checkpoint, observed identity table, target-condition rules. |
| Prove network isolation and offline rebuild | Offline procedure; successful fresh mirror build; expected failure without transitive B. |
| List supported, restricted, and unsupported cases | Supported-profile matrix. |
| Recommend diagnostics and conformance fixtures | Stable-diagnostic table and fixture tables. |
| Identify gaps needing separate C-family strategy | Dedicated gaps section and implementation recommendation 8. |

The evidence supports SwiftPM as the preferred boundary only with the wrapper
profile and fail-closed restrictions above. It does not support direct admission
of arbitrary SwiftPM packages or weakening the shared compiled-artifact policy.
