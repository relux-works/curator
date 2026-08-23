# TASK-260811-3kbf3l: implement-rust-offline-build-adapter

## Description
Implement the executable rust-source-v1 build profile over an admitted Cargo closure. Spec trace: .spec/skill-facing-cli-source-closure.md Current delivery scope Rust; Source closure invariant items 2-6; Delivery completion. Accepted input: TASK-260810-3urqbl Recommended rust-source-v1 contract.

## Scope
Select one package and bin on one native target; consume the C4 selection binding, require its target platform and physical Cargo, rustc, sysroot, target stdlib, linker, and SDK nodes and typed targets or uses-tool edges to resolve uniquely, and reject binding or time-of-use drift. Reject active build scripts, build dependencies, proc macros, native links, config wrappers, cross targets, and unstable features; run fresh-home frozen metadata and build under the protected boundary; validate Cargo JSON and publish receipted outputs without changing capture or planned identities.

## Acceptance Criteria
Supported pure-Rust fixtures build and execute offline from the verified closure; capture remains selection-neutral while exact target, feature, binding, active graph, toolchain, command, sandbox, and output identities are checkpointed. Missing, duplicate, dangling, wrong-kind, or drifted platform or tool bindings and every unsupported unit or undeclared input fail before compilation or publication with stable diagnostics. CGP05 target-binding cases, R01-R09, RF09-RF12, RH01-RH10, and protected-cache regressions pass.
