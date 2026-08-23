# TASK-260811-2fwvml: implement-rust-source-closure-adapter

## Description
Implement the Rust skill-facing CLI adapter using the Cargo closure strategy accepted by the research decision.

## Scope
Resolve the complete transitive Cargo source graph, capture immutable lock and source identities, prepare an offline rebuildable closure, account for features, build scripts, proc macros, git and registry dependencies, bind Rust toolchain and target identity, and reject prohibited compiled payloads or undeclared inputs.

## Acceptance Criteria
Rust CLI packages supported by the accepted profile build and execute from the captured closure with network disabled; transitive registry and git sources, feature resolution, build scripts, proc macros, target, and toolchain inputs are checkpointed; unsupported or binary-bearing cases fail closed with stable diagnostics; positive and negative fixtures and adapter tests pass.
