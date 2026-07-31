# TASK-260729-rhjxtx: probe-rust-swift-kotlin-toolchain-boundaries

## Description
Build host-verified probe fixtures for Rust, Swift, and Kotlin Native toolchain detection, version normalization, target identity, metadata compatibility, and compiler-boundary behavior needed by the three language-driver contracts. Capture macOS-primary and available Windows evidence without auto-installing toolchains.

## Scope
Task-owned probe/evidence only. Use the current accepted language artifact boundary and treat the pending shared toolchain contract as provisional input; do not change normative curator-spec files, implementation repos, release pins, or dependencies. No package-controlled paths, environment overrides, URLs, channels, install commands, or trust roots. Kotlin means Kotlin Native, not a JVM runtime bundle. Linux remains later non-gating validation.

## Acceptance Criteria
For each language, identify authoritative version/target commands and source metadata surfaces; provide parsable positive fixtures and adversarial malformed/too-new/incompatible controls; distinguish representability from host-version/target gates; record exact macOS and reachable Windows commands, versions, outputs, and exit codes; absence of a required toolchain is reported as evidence with manager-owned primary-source installation guidance inputs, never auto-installed; outputs are reusable by TASK-260728-12pnm1, TASK-260728-1yhuqi, and TASK-260728-168smo.
