# TASK-260810-135n37: evaluate-kotlin-source-only-closure

## Description
Determine whether Kotlin CLI dependencies can satisfy Curator source-only closure under Gradle and Maven conventions.

## Scope
Assess Gradle dependency locking and verification, Maven coordinates and repositories, source dependencies and composite builds, plugins, Kotlin and JVM toolchain identity, offline operation, and the consequences of rejecting JAR, class, native, and other precompiled payloads.

## Acceptance Criteria
The outcome recommends a viable Kotlin source-only profile or a fail-closed limitation, enumerates recursive source and toolchain inputs, defines offline and checkpoint behavior, rejects compiled payloads, and documents diagnostics, risks, and conformance fixtures.
