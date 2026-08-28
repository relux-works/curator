# TASK-260811-1zwhx2: implement-swiftpm-c-family-source-closure-adapter

## Description
Implement the SwiftPM adapter for Swift and the accepted mixed C-family target profile.

## Scope
Resolve the complete SwiftPM source dependency graph, capture Package.resolved and immutable source identities, rebuild offline for supported Swift, C, C++, Objective-C, and Objective-C++ targets, represent target and FFI edges, bind Swift and native toolchain and platform identity, and reject binary targets, prebuilt libraries, undeclared plugins, or unsupported system dependencies.

## Acceptance Criteria
Approved pure-source SwiftPM packages and mixed-language targets rebuild and execute from the captured closure with network disabled; package, target, FFI, platform, and toolchain inputs are checkpointed; binary targets and other prohibited or undeclared inputs fail closed with stable diagnostics; portability limits are explicit; positive and negative fixtures and adapter tests pass.
