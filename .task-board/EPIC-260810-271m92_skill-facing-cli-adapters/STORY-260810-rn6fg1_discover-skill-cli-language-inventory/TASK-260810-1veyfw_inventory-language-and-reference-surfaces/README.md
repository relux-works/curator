# TASK-260810-1veyfw: inventory-language-and-reference-surfaces

## Description
Inventory the real skill-facing CLI estate and the independent protocol implementations before adapter design.

## Scope
Inspect authoritative repositories, skill manifests, command declarations, build metadata, launchers, package-manager files, runtime requirements, mixed-language packages, and existing Python behavior. Classify Go as the implemented baseline; Swift/C-family, Rust, and Node/TypeScript as confirmed current targets; Python as a reference implementation; and Kotlin, Dart, and .NET as deferred.

## Acceptance Criteria
A cited evidence matrix records each current implementation or CLI surface, repository and path, language, package manager, build and launch entry points, lock and integrity metadata, transitive dependency shape, runtime requirements, mixed-language edges, and any precompiled payloads. It explicitly confirms or corrects the assumed relationship between Node/TypeScript and Python and keeps Kotlin outside the active investigation.
