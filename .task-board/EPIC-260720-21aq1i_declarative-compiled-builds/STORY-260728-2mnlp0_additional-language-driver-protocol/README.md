# STORY-260728-2mnlp0: additional-language-driver-protocol

## Description
Establish the shared protocol and version boundary for closed local and external Rust, Swift and Kotlin build-driver pairs using context-excluded build roots and the manager-neutral skill-build.json external descriptor.

## Scope
Next-version manifest, build-root and descriptor ownership; six explicit driver identifiers; artifact and launcher classes; toolchain requirements; cache/receipt/marker/claim identity; execution-policy interaction; generated schemas/cases and candidate qualification. No manager implementation.

## Acceptance Criteria
Schemas 1-7 and existing Go drivers remain frozen after the descriptor rename; the next version admits only the six explicit Rust, Swift and selected Kotlin local/repository drivers plus structured toolchain requirements; no generic language, build command, argv, environment, output or signing surface appears; accepted threat models generate deterministic schemas, vectors and honest release evidence.
