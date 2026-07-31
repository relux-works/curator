# STORY-260728-2abmzr: swift-external-build-driver

## Description
Define and implement closed local swift-v1 and external swift-repository-v1 CLI build drivers across curator-spec, Curator, csk and shared interoperability.

## Scope
Shared Swift threat model and versioned driver pair, vendored local build roots, external repository targets, toolchain preflight, locked dependencies, plugin and macro policy, manager implementations, macOS-first and Windows qualification, Linux follow-up, conformance and guidance.

## Acceptance Criteria
Both explicit Swift drivers use one closed manager-owned pipeline; package plugins, macros, arbitrary scripts, network, binary targets and unmanaged native inputs cannot escape it; local and external fixtures match across managers with verified artifacts, cache, rollback and diagnostics; required native and shared gates pass.
