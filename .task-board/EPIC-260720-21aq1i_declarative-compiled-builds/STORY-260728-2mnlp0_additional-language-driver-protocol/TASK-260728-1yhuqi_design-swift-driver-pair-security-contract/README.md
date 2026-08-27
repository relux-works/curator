# TASK-260728-1yhuqi: design-swift-driver-pair-security-contract

## Description
Design closed local swift-v1 and external swift-repository-v1 drivers and select an enforceable SwiftPM/swiftc pipeline under the portable manager-worker policy.

## Scope
Trusted Swift toolchain and manifest requirement; local build_roots versus external skill-build.json targets; Package.swift and Package.resolved grammar; dependency vendoring; products/targets/configurations; plugins, macros, scripts, binary/system-library targets, C/C++ interop, unsafeFlags, environment, network, linker and SDK selection; deterministic executable output and platform qualification.

## Acceptance Criteria
The reviewed contract defines both source modes, exact toolchain and SDK identities, one exhaustive allow/reject matrix and fixed manager-owned process graph; package code cannot select scripts/plugins/macros/network/native inputs outside the contract; macOS and Windows requirements plus Linux follow-up are implementation-ready.
