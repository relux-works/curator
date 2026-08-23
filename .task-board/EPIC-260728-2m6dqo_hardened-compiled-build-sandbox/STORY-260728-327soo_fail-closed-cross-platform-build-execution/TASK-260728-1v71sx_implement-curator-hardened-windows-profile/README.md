# TASK-260728-1v71sx: implement-curator-hardened-windows-profile

## Description
Implement or explicitly capability-gate a Curator hardened Windows profile using a reviewed AppContainer, Job Object, Hyper-V boundary or equivalent supported mechanism.

## Scope
Suspended process creation, restricted identity/AppContainer, Job Object descendant and resource accounting, filesystem/network policy, exact executable policy, capability detection, and adversarial tests on ssh win.

## Acceptance Criteria
Every supported Windows configuration proves all six hard guarantees before it may advertise the hardened profile; otherwise it rejects before Go starts with a stable diagnostic. Administrator setup is not treated as evidence unless runtime enforcement is verified.
