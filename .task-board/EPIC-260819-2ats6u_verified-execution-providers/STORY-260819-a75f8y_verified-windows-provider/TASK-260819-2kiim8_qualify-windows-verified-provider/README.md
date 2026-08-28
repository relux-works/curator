# TASK-260819-2kiim8: qualify-windows-verified-provider

## Description
Run independent conformance and fault injection across the supported Windows matrix.

## Scope
Cover supported Windows builds and security configurations, NTFS and relevant filesystem behavior, IPv4 and IPv6, reboot, crash, overload, queue pressure, service and driver restart, upgrade, rollback, revocation, and unsupported configurations. Capture immutable host build and policy evidence.

## Acceptance Criteria
Every supported cell passes the common and Windows suites with provider-bound receipts; every unsupported or partially configured cell fails before workload start with stable diagnostics; evidence is independently reviewed and release-gate consumable.
