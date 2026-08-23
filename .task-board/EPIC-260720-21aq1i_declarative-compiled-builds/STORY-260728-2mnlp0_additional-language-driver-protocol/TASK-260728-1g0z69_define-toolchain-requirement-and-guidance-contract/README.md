# TASK-260728-1g0z69: define-toolchain-requirement-and-guidance-contract

## Description
Define the closed manifest toolchain requirement, trusted resolution, version comparison and installation-guidance contract shared by all compiled drivers.

## Scope
Driver-to-toolchain mapping; exact versus bounded version constraints and prerelease policy; OS and architecture applicability; trusted executable resolution and fingerprinting; two-stage preflight with host availability/version before source acquisition and go.mod, rust-toolchain.toml, swift-tools-version or selected Kotlin metadata cross-check after validation/audit but before compiler work; typed missing, incompatible, untrusted and mismatch diagnostics; manager-owned guidance IDs and catalog lifecycle; no auto-install.

## Acceptance Criteria
The reviewed decision is canonical and implementation-ready, has deterministic cross-language version semantics, prevents package-selected paths, URLs, channels, commands and trust roots, defines cache and receipt identity effects, defines both fail-fast stages without bypassing source audit, and maps official platform guidance without embedding stale package-controlled instructions.
