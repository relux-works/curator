# TASK-260728-3n67j6: implement-curator-hardened-macos-profile

## Description
Implement or explicitly capability-gate a Curator hardened macOS profile using a reviewed signed helper, sandboxed worker, VM boundary or equivalent supported mechanism.

## Scope
Native macOS execution boundary, code identity and notarization, filesystem and network confinement, descendant and resource controls, exact executable policy, capability detection, and adversarial tests on ssh relux.

## Acceptance Criteria
Every supported macOS configuration proves all six hard guarantees before it may advertise the hardened profile; otherwise it rejects before Go starts with a stable diagnostic. Signing, helper identity and TOCTOU checks are independently reviewed.
