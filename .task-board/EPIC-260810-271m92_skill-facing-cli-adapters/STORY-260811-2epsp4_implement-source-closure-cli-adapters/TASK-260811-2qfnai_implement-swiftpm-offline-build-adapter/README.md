# TASK-260811-2qfnai: implement-swiftpm-offline-build-adapter

## Description
Implement offline planning, build, and publication for the swiftpm-source-v1 adapter. Spec trace: .spec/skill-facing-cli-source-closure.md Current delivery scope Swift and C family; Source closure invariant items 2-6; Delivery completion. Accepted input: TASK-260810-zddzh7 Checkpoints 3-4 and toolchain identity.

## Scope
Plan the selected product and native SwiftPM build system from the admitted mirror capture plus exact C4 platform, SwiftPM, swiftc, PackageDescription, Clang, linker, and SDK binding overlay; require every action target and tool slot to resolve once, preserve capture and binding identities through C5, and reject wrong-kind or drifted bindings. Disable experimental prebuilts; use fresh isolated scratch, cache, config, security, and output roots with network none and forced resolved versions; reconcile commands and compiler read or write sets; inspect and publish sorted observations and receipts without mutating expected graph records.

## Acceptance Criteria
Supported SwiftPM products rebuild and execute offline with original origins unavailable; capture remains selection-neutral while binding, active graph, toolchain, SDK, platform, command, FFI, read and write sets, and output identities are checkpointed. Missing, duplicate, dangling, wrong-kind, incompatible, or drifted target or tool bindings, graph drift, missing mirrors, untrusted inputs, undeclared generation, network use, or output drift fails without cache publication. CGP05 target-binding cases, S01-S10, R01-R13, and exact adapter output-receipt cases pass.
