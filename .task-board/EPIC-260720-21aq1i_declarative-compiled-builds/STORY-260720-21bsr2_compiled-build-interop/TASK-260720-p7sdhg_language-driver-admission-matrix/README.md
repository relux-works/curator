# Define the language-driver matrix and admission gate

## Description
Turn the accepted build-system research into a maintained language and toolchain matrix plus a normative-quality checklist for proposing future compile-only driver revisions.

## Scope
Work in curator-spec after the compile-only decision and TASK-260720-3lo9jc. Own docs/build-driver-admission.md plus its README and decision-record navigation links. Classify Go, Rust and Cargo, Zig, Swift and SwiftPM, direct C and C++ compilers, Make, CMake, Meson, Java and Kotlin build paths, .NET and MSBuild, Node and TypeScript compilers and bundlers, Deno, and Python packaging. Do not reserve speculative driver identifiers, imply that offline equals safe, or weaken the closed go-v1 contract.

## Acceptance Criteria
The matrix uses explicit supported-now, deferred-for-separate-design, and prohibited-without-a-separate-sandboxed-driver classes with a concise security rationale for every listed ecosystem. Make, CMake, Meson, Cargo build scripts and procedural macros, SwiftPM manifests and plugins, Maven and Gradle plugins, MSBuild tasks and generators, npm lifecycle scripts, bundler configs and plugins, Python build backends, and any generic shell or package-provided argv are named non-goals. The future-driver process requires a new versioned identifier, threat-model review, closed schema, manager-owned executable and argv graph, source and dependency identity, toolchain and target fingerprint, context isolation, cache and receipt model, dry-run semantics, resource limits, positive and negative vectors, authoring docs, two independent consumers, cross-platform parity evidence, compatibility and security review, and an immutable release. Links resolve, the accepted research remains traceable, and no future support or release evidence is fabricated.
