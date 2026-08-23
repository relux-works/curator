# Reviewer verdict for TASK-260810-zddzh7

Verdict: **accepted -> done**

## Goal and scope evidence

- Reviewer run: `RUN-260811-89e84f`
- Authoritative goal immediately before verdict: `GOAL-260811-233766`
  revision 1
- Resolved scope: `TASK-260810-zddzh7`
- Review policy: `required`
- The goal was neither cancelled nor superseded at the verdict checkpoint.
- The orchestrator's progress and convergence directives were observed and
  acknowledged. This artifact supplies the requested final verdict.

The goal requires exactly one evidence-backed reviewer branch. This artifact
records the `accepted` branch; it records no `changes_requested` or
`stop_the_line` branch.

## Acceptance result

The submitted research decision satisfies the task description, scope,
acceptance criteria, and reviewer checklist. No acceptance-blocking finding
remains.

1. **Conservative language and package-shape coverage.** The supported-profile
   matrix distinguishes supported, toolchain/destination-restricted, and
   unsupported cases for Swift, C, C++, Objective-C, and Objective-C++.
   Swift and C-family source must occupy separate targets. A single Clang
   target may contain `.c`, `.cpp`, `.m`, and `.mm` only where the exact
   toolchain/destination fixture passes. Direct Swift/C++ interoperability and
   Objective-C-family imports remain explicitly restricted.
2. **Complete recursive capture.** The decision binds the root
   `Package.resolved`, every direct and transitive source-control pin, exact
   revisions, canonical origins and source-control kinds, immutable package
   snapshots, all manifests, normalized package/target/source graphs, headers,
   module maps, settings, target conditions, SDK/system edges, and a canonical
   closure digest. Dependency lockfiles are correctly treated as non-authority
   beneath the top-level package.
3. **Offline and fail-closed replay.** The profile requires kind-preserving
   local mirrors for every pin, fresh cache/config/security/scratch roots,
   unavailable original origins, externally enforced `network=none`, disabled
   credentials and SCM-to-registry transformation, disabled SwiftSyntax
   prebuilts, and `--force-resolved-versions`. Missing transitive inputs,
   stale locks, graph drift, and source inventory drift reject before output
   publication.
4. **Undeclared-input controls.** The module-map escape fixture proves that
   SwiftPM metadata alone is not the security closure. The recommended wrapper
   independently inventories package bytes, parses and confines module maps,
   sanitizes include/library/framework inputs, and reconciles compiler-observed
   reads against captured package or trusted toolchain/SDK roots.
5. **Plugins, macros, system libraries, and binaries.** Active plugins and
   macros reject before compilation/execution in v1; command plugins are never
   invoked; prebuilt macro retrieval is disabled. System-library/pkg-config
   edges are admitted only inside an independently selected and fingerprinted
   SDK/sysroot/toolchain boundary. Every local or remote `binaryTarget`, even
   dormant, and every compiled dependency payload reject under the accepted
   shared artifact policy.
6. **Checkpoints, diagnostics, and fixtures.** Checkpoints 0 through 4 define
   policy/root admission, controlled acquisition, immutable closure commit,
   offline graph replay/build planning, and protected build/output receipts.
   Stable `swiftpm_*` codes supplement, without weakening, the accepted
   `artifact_*` vocabulary. Fixture families `S01`-`S10`, `H01`-`H08`,
   `R01`-`R13`, and `P01`-`P09` cover positive and fail-closed cases.
7. **Separate C-family strategy.** The document explicitly routes non-SwiftPM
   graphs, custom build systems/generators, arbitrary host libraries,
   unsupported source-control features, unmodeled per-file/link behavior,
   unvalidated Objective-C runtimes, and prebuilt-required packages to a
   separate strategy or rejection. It does not stretch SwiftPM into a generic
   native-build escape hatch.

## Primary-source fact check

The relevant vendor facts were independently checked against current primary
sources on 2026-08-11:

- Swift PackageDescription identifies Swift, Objective-C, Objective-C++, C,
  and C++ as package languages, and SwiftPM documents recursive valid-source
  discovery:
  https://docs.swift.org/swiftpm/documentation/packagedescription/ and
  https://docs.swift.org/swiftpm/documentation/packagedescription/target/sources/
- SwiftPM documents the top-level `Package.resolved` behavior, ignored
  dependency lockfiles, implicit resolution, recursive traversal, and
  force-resolved mode:
  https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/resolvingpackageversions/
  and
  https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/resolvingdependencyfailures/
- The C-target and system-library documentation confirms public-header/module-
  map behavior and `pkg-config`/provider/system edges:
  https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/creatingclanguagetargets/
  and
  https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/addingsystemlibrarydependency/
- SwiftPM documents plugins as separate sandboxed processes with temporary
  writes and optionally approved command-plugin network/package writes; the
  Swift language guide confirms macro implementations execute during client
  compilation:
  https://docs.swift.org/swiftpm/documentation/packagemanagerdocs/plugins/ and
  https://docs.swift.org/swift-book/documentation/the-swift-programming-language/macros/
- Swift's C++ interoperability guide confirms per-target opt-in and the
  dependency-chain constraint:
  https://www.swift.org/documentation/cxx-interop/project-build-setup/ and
  https://www.swift.org/documentation/cxx-interop/status/
- SwiftPM's merged prebuilt SwiftSyntax work describes downloading and
  extracting macro support libraries:
  https://github.com/swiftlang/swift-package-manager/pull/8142

These sources support the submitted factual claims. The narrower
`swiftpm-source-v1` policy is a Curator security decision, not a claim that
unwrapped SwiftPM supplies immutable closure, network denial, or protected
receipts.

## Independent fixture and validation evidence

- Research source and board outcome are byte-identical: 703 lines, 48,074
  bytes, SHA-256
  `361b389e54809d0bce44ea9698860e04de26a0f5ab96219481d17aca47135b3a`.
- Document validation passed: nonempty, no trailing whitespace, balanced
  Markdown fences, required sections present, and local specification/research
  links resolve.
- Observed toolchain identity matches the document: Apple Swift/SwiftPM 6.3.2,
  `swiftlang-6.3.2.1.108`, Apple Clang 21.0.0, Xcode 26.5 build `17F42`, macOS
  SDK 26.5 build `25F70`, arm64 Darwin host.
- `swift build` from clean isolated roots compiled one Clang target's C, C++,
  Objective-C, and Objective-C++ files and ran the Swift client with output
  `39`.
- Clean direct Swift/C++ and Darwin Objective-C/Objective-C++ builds passed and
  ran with outputs `42` and `36`.
- The mixed Swift+C single-target fixture failed with SwiftPM's expected
  `mixed language source files` diagnostic.
- Raw SwiftPM built the absolute-header custom module map and ran with output
  `73`; `swift package describe --type json` exposed only `dummy.c`, proving
  the documented enumeration gap. The mixed-family description recursively
  exposed all six source paths but no headers/module maps.
- `dump-package` exposed target types `binary`, `macro`, `plugin`, and `system`
  without building them. Installed help confirms experimental SwiftSyntax
  prebuilts default on and exposes the required disable flag.
- With original A and B repositories absent, an outer macOS network-deny
  sandbox, fresh ambient state, captured local mirrors, and force-resolved
  mode rebuilt the root and ran it with output `41`. A copied root configured
  with only the direct A mirror failed on missing transitive B as expected.
- Both mirror object graphs passed `git fsck --full`; lock digest, commits,
  trees, and deterministic archive hashes match the submitted evidence.
- `git diff --check` exited 0.
- `go test -count=1 ./...` exited 0 across every repository package. Notable
  durations were `cmd/curator` 398.253s, `internal/godriver` 81.449s,
  `internal/install` 121.372s, and `internal/install/atomicity` 120.651s.

## Architecture fit and non-blocking anomalies

The recommendation fits the existing architecture: adapters remain callers of
manager-owned closure/artifact policy, toolchain fingerprinting, build-input
metadata, protected cache, and receipt services. The profile extends the
existing Go baseline's trust boundaries instead of creating a SwiftPM-specific
bypass.

Two reviewer harness observations are preserved for implementation:

1. Wrapping SwiftPM in an outer macOS `sandbox-exec` network deny while leaving
   SwiftPM's own subprocess sandbox enabled failed with nested
   `sandbox_apply: Operation not permitted`. The same clean offline replay
   passed when the manager-owned outer sandbox was authoritative and SwiftPM
   received `--disable-sandbox`. The adapter must choose and checkpoint one
   authoritative sandbox composition; it must not silently drop the outer
   network/read/write policy. This is non-blocking because the research already
   assigns security enforcement to the Curator wrapper rather than SwiftPM's
   platform-dependent sandbox.
2. A default cached `go test ./...` entered the known historical `.temp`
   test-cache input scan and grew to approximately 14 GB RSS. The reviewer
   terminated only that reviewer-launched coordinator and reran the identical
   package set with `-count=1`; the full uncached suite passed. This is an
   unrelated workspace/cache-accounting anomaly, not a SwiftPM research or
   product regression.

## Routing decision

Accept the research deliverable and route `TASK-260810-zddzh7` to `done`.
This reviewer run supplies no `commit_ack`. The transition cannot complete the
parent Story because multiple sibling discovery tasks remain active.
