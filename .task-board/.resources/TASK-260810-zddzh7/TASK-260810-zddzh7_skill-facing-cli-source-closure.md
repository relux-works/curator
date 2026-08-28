# Skill-facing CLI source closure

Status: delivery input for `EPIC-260810-271m92` and its research and
implementation stories.

## Context

Curator needs language-aware build adapters for CLI commands shipped by skills.
The existing Go path is the security baseline: resolve the complete recursive
source dependency closure, build from a controlled snapshot with a trusted
toolchain, and bind audit evidence and cache receipts to the exact inputs.

The Python reference implementation is independent from this repository. Its
behavior is an input to the inventory and must be verified rather than assumed
to be the implementation location for new adapters.

## Current delivery scope

- Go remains the implemented baseline.
- Rust and Node/TypeScript are confirmed adapter targets for this cycle.
- Swift and the SwiftPM-supported C family are confirmed adapter targets for
  this cycle. The C family includes C, C++, Objective-C, and Objective-C++ only
  where the accepted SwiftPM strategy can represent and rebuild the source
  boundary conservatively.
- Python remains in scope as an existing protocol implementation and ecosystem
  baseline. Research must determine which Python behavior should be shared with
  or mirrored by the Node/TypeScript path; a new Python implementation is not a
  default deliverable.
- Kotlin is explicitly deferred and must not block research or implementation
  in this cycle. Preserve its source-only Gradle/Maven investigation as future
  backlog, but do not launch it without renewed user approval.
- Dart and .NET remain deferred candidates and are not delivery commitments in
  this cycle.
- Swift Package Manager is the preferred hypothesis for a shared package
  boundary across Swift, C, C++, Objective-C, and Objective-C++. Research must
  verify its exact coverage and conservative failure modes before it becomes a
  normative adapter choice.

## Source closure invariant

Every supported adapter must either produce a complete, recursively resolved,
auditable source closure or fail closed before build or installation. A literal
checked-in vendor directory is not required, but the resulting closure must
provide the same security properties as Go vendoring:

1. Every transitive dependency is enumerated and bound to immutable identity.
2. Lock data, checksums, source snapshots, build metadata, and toolchain
   identity are available for audit and checkpoint generation.
3. A build can run offline from the captured closure without resolving new
   network inputs.
4. Dependency hooks, plugins, generators, and build scripts cannot silently
   add undeclared inputs or escape the trusted build boundary.
5. Mixed-language edges remain explicit in one dependency graph, including
   build order, toolchain identity, target platform, and FFI boundaries.
6. An ecosystem that cannot satisfy these properties is unsupported or limited
   to a narrower source-only profile until a conservative strategy exists.

## Vendored compiled artifact prohibition

Curator must reject vendored precompiled executable code in every language
adapter until Curator has an explicit binary-admission capability that verifies
signatures, provenance, artifact identity, target platform, and audit policy.

The default-deny class includes native executables, object files, static and
dynamic libraries, framework and XCFramework payloads, native Node addons,
Python extension modules, JVM bytecode, WebAssembly, and equivalent compiled
or intermediate executable artifacts embedded directly or inside archives.
Pure source archives and inspectable source-like text remain eligible only when
their full contents and immutable origin can be audited.

This prohibition applies to dependency and package payloads admitted into the
source closure. It does not prohibit trusted toolchain executables selected
outside that closure or binaries built locally from the admitted source inside
Curator's protected build pipeline. Locally built outputs still require
content-addressed receipts and protected-cache validation.

No ecosystem-specific adapter may weaken this rule. Future verified-binary
support must be a separate Curator capability and policy decision.

## Required research

The discovery cycle must produce evidence for:

- the existing skill-facing CLI and protocol-implementation inventory;
- a precise artifact taxonomy and common rejection diagnostics;
- Rust recursive source closure and offline build behavior;
- SwiftPM source closure and mixed Swift/C-family package behavior;
- Node/TypeScript and Python source-package policies, including native addon,
  wheel, lifecycle-script, and generated-code handling;
- mixed-language graph, FFI, build-order, and toolchain checkpoint semantics;
- conformance vectors that prove complete closure, offline reproducibility,
  binary rejection, and fail-closed handling of undeclared inputs.

Kotlin source-only feasibility under Gradle/Maven conventions remains a
documented future question, including the impact of rejecting JAR and other
bytecode dependencies, but is outside the active research graph.

## Discovery deliverable

The final research decision must provide an ecosystem matrix, a shared adapter
contract, language-specific closure strategies, explicit unsupported cases,
diagnostic codes, audit checkpoint inputs, conformance-test requirements, and
an implementation-ready backlog. The discovery Story remains research-only;
accepted synthesis unlocks the implementation stories in the same Epic.

## Delivery completion

This cycle is complete only when the accepted research decision has been
translated into atomic implementation tasks and the Rust, Node/TypeScript, and
SwiftPM/C-family paths have been implemented, tested, reviewed, and integrated
with shared conformance coverage. Any ecosystem that cannot meet the source
closure invariant must fail closed with an explicit unsupported result rather
than weaken the policy.
