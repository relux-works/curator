# TASK-260819-13htlk: qualify-linux-verified-provider

## Description
Run independent conformance and fault injection across the supported Linux matrix.

## Scope
Cover supported kernels and distributions, bare metal and explicitly supported virtualization or containers, filesystem variants, IPv4 and IPv6, reboot, crash, overload, event queue pressure, upgrade, rollback, revocation, and unsupported configurations. Capture immutable machine and kernel configuration evidence.

## Acceptance Criteria
Every supported cell passes the common and Linux suites with provider-bound receipts; every unsupported or partially configured cell fails before workload start with stable diagnostics; evidence is independently reviewed and release-gate consumable.
