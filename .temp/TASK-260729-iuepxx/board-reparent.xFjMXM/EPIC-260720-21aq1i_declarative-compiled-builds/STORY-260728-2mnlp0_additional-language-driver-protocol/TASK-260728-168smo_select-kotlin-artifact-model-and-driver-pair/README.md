# TASK-260728-168smo: select-kotlin-artifact-model-and-driver-pair

## Description
Select the Kotlin CLI artifact/runtime model and design closed local and external driver contracts before any implementation.

## Scope
Compare Kotlin/JVM executable JAR plus trusted runtime/bundle options with Kotlin/Native executables; choose driver identifiers; define trusted compiler/JDK or native toolchain requirements; local build_roots versus external skill-build.json targets; source/dependency layout; Gradle, Maven, KSP, compiler plugins, scripts, annotations, native interop, network and launcher policy; cache and platform identity.

## Acceptance Criteria
A reviewed decision selects one implementable artifact model and paired driver identifiers, defines deterministic launch and distribution semantics, exhaustively allows/rejects package-selected build behavior without a generic Gradle escape hatch, integrates toolchain preflight, and fixes macOS/Windows/Linux qualification requirements.
