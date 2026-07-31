# STORY-260728-3tymqm: rust-external-build-driver

## Description
Define and implement closed local rust-v1 and external rust-repository-v1 CLI build drivers across curator-spec, Curator, csk and shared interoperability without repository- or skill-supplied build commands.

## Scope
Shared Rust threat model and versioned driver pair, vendored local build roots, external repository targets, toolchain preflight, wire and receipt identities, Curator and csk implementations, macOS/Windows qualification, Linux follow-up, conformance and guidance.

## Acceptance Criteria
Both explicit Rust drivers share one closed offline toolchain policy while preserving distinct source identities; build scripts, proc macros, plugins, network and native-link inputs are rejected or manager-controlled; local skill and external repository fixtures build matching verified manager-named artifacts in both managers; no generic command execution exists.
