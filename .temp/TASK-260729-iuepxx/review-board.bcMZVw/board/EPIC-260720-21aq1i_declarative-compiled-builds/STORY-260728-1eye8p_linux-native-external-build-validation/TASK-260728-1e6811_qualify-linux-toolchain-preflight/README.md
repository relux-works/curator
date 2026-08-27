# TASK-260728-1e6811: qualify-linux-toolchain-preflight

## Description
Run the deferred clean-host Linux qualification of the shared two-stage toolchain preflight after the compiled language drivers are otherwise interoperable.

## Scope
Go, Rust, Swift and selected Kotlin local and repository cases on a recorded Linux distro and architecture; trusted discovery; exact and bounded versions; absent, incompatible, untrusted and source-metadata-mismatch failures; manager-owned Linux guidance; fingerprint/cache/receipt identity; no-mutation evidence. This task is non-gating for macOS and Windows delivery.

## Acceptance Criteria
Curator and csk pass the shared Linux preflight corpus for every available compiled driver on a clean host; failure cases occur at the specified stage without persistent mutation; exact distro, architecture, toolchain provenance and evidence are recorded; unavailable platform capabilities are reported honestly and do not retroactively block prior platform acceptance.
