# TASK-260819-2fn4ku: implement-linux-enforcement-and-observation-provider

## Description
Implement the privileged Linux provider selected by the qualification task.

## Scope
Create the isolated workload boundary; enforce process, filesystem, environment, network, and resource policy; collect lossless attempt evidence through the selected kernel hooks; bind pidfds, namespaces, cgroups, mount and executable identities to the common session; detect queue loss, provider death, reboot, and policy drift.

## Acceptance Criteria
On the supported matrix, undeclared attempts cannot escape or disappear from evidence; provider, queue, hook, namespace, identity, race, and resource failures reject or invalidate the session; common and Linux adversarial tests pass.
