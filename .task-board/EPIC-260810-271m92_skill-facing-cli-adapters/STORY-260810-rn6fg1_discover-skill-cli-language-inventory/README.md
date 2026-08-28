# STORY-260810-rn6fg1: discover-skill-cli-language-inventory

## Description
Inventory the current skill-facing CLI and protocol-implementation estate, then produce a conservative source-closure architecture for the confirmed Swift/C-family, Rust, and Node/TypeScript targets under a no-vendored-compiled-artifacts policy.

## Scope
Treat Go as the implemented security baseline and the independent Python implementation as a reference input. Evaluate recursive source closure and offline rebuilds for Rust, Node/TypeScript, and SwiftPM mixed Swift/C/C++/Objective-C/Objective-C++ packages; mixed-language dependency graphs; FFI boundaries; toolchain identity; audit checkpoints; and binary-artifact rejection. Kotlin, Dart, and .NET are deferred and outside the active research graph. This Story is research-only and its accepted synthesis unlocks implementation Stories.

## Acceptance Criteria
Reviewed research and synthesis define the supported language matrix, complete recursive source-closure invariants, language-specific conservative strategies, mixed-language graph semantics, compiled-artifact rejection rules, diagnostic codes, audit checkpoint inputs, conformance vectors, explicit unsupported cases, and an implementation-ready backlog for Rust, Node/TypeScript, and SwiftPM/C-family delivery.
