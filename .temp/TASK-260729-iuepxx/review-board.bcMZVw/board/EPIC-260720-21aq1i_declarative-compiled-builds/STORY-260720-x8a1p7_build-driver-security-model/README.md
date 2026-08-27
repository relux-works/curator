# Build-driver security model

## Description
Research and recommend the protocol boundary for manager-owned compilation of untrusted skill source. Compare Go and other candidate languages by whether a fixed manager-generated build command can compile without executing package-provided hooks, plugins, macros, manifests, or arbitrary argv.

## Scope
Inspect curator-spec rc.3, Curator main, and csk main. Cover Go, Rust/Cargo, Zig, Swift/SwiftPM, C/C++ toolchains and CMake/Meson, Java/Kotlin with javac/Gradle/Maven, .NET/MSBuild, Node/TypeScript bundlers, Deno, Python packagers, and any clearly relevant compile-only toolchain. Define cache-key inputs, environment isolation, dependency fetching, platform targeting, dry-run behavior, rollback, and audit ordering.

## Acceptance Criteria
An outcome resource gives a recommended manifest model, threat model, normative Go invocation constraints, cache/receipt model, compatibility and migration impact, and a language matrix classifying safe-now, restricted-future, and prohibited-without-sandbox drivers with concrete rationale.
