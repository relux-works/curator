# STORY-260728-2fsqtv: compiled-build-toolchain-preflight

## Description
Add manifest-declared toolchain requirements, trusted local preflight and manager-owned installation guidance for Go, Rust, Swift and the selected Kotlin driver across local and external builds.

## Scope
Version-constraint and toolchain-identity wire contract, fail-fast resolution order, source-file cross-checks, cache identity, typed diagnostics, versioned official-guidance catalog, Curator and csk implementations, macOS/Windows/Linux behavior and shared conformance. No automatic toolchain installation.

## Acceptance Criteria
Every compiled command performs trusted host-toolchain discovery and version preflight before source acquisition or persistent mutation; after local source validation or exact external acquisition and audit, a second metadata-compatibility check runs before compiler work. Missing, incompatible, untrusted or metadata-mismatched toolchains fail with stable diagnostics and platform-appropriate manager-owned guidance. Manifests cannot select executable paths, download URLs, channels, install commands or trust roots; resolved toolchain fingerprints remain cache and receipt identity; Curator and csk match shared vectors.
