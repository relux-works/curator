# Implement portable source-closure CLI adapters

## Description
Implement the accepted portable source-closure architecture for Rust, Node/TypeScript, and SwiftPM/C-family adapters. Portable mode is CLI-only, verifies admitted inputs and declared outputs, records exact enforced and observed capabilities, and never claims verified host observation.

## Scope
Build the portable assurance foundation and ecosystem adapters for Rust, Node/TypeScript, and SwiftPM mixed Swift/C/C++/Objective-C/Objective-C++ targets. Keep Go as baseline and Python as compatibility reference. Exclude Kotlin, Dart, .NET, verified providers, and verified binary admission. Reject vendored compiled artifacts globally.

## Acceptance Criteria
Supported adapters recursively capture source closure, rebuild offline from immutable inputs, bind dependency and toolchain identity, reject vendored compiled artifacts and undeclared hooks, emit portable receipts that cannot satisfy verified policies, expose stable diagnostics, and pass independent language-specific and cross-language conformance review.
