# TASK-260819-3tjqon: implement-windows-enforcement-and-observation-provider

## Description
Implement the Windows service and driver provider selected by qualification.

## Scope
Create the workload security boundary; enforce and observe process, filesystem, registry, environment, network, and resource policy; bind process and token identity to the common session; use minifilter and WFP or other selected callbacks where required; detect event loss, driver unload, service death, reboot, tamper, and policy drift.

## Acceptance Criteria
On the supported matrix, undeclared attempts cannot escape or disappear from evidence; service, driver, queue, identity, reparse, race, network, and resource failures reject or invalidate the session; common and Windows adversarial tests pass.
