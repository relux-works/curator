# TASK-260728-2spy93: define-additional-driver-version-and-artifact-boundary

## Description
Decide the next protocol version, local build-root and neutral external descriptor evolution, structured toolchain requirements and artifact/launcher model shared by Rust, Swift and Kotlin driver pairs.

## Scope
Specify manifest schema and skill-build.json schema versioning; six explicit local/repository drivers; context-excluded local roots; external target ownership; native executable versus runtime-bundle classes; manager-derived naming; toolchain constraints; cache/receipt/marker/claim versions; mixed commands and platform claims. Preserve schemas 1-7 and Go semantics.

## Acceptance Criteria
One reviewed decision fixes the version, source and wire ownership boundaries, prevents generic language and arbitrary build/install fields, defines or rejects non-native runtime artifacts, integrates trusted toolchain preflight, and gives exact downstream obligations without fabricated claims.
