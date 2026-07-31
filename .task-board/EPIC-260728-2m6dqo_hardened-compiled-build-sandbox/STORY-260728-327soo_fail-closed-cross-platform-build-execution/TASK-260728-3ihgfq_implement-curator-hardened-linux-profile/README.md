# TASK-260728-3ihgfq: implement-curator-hardened-linux-profile

## Description
Implement the Curator hardened Linux worker using reviewed kernel primitives and capability detection.

## Scope
Identity-verified manager worker, namespaces, Landlock, seccomp/no-new-privileges, cgroup v2, private mounts or equivalent, process and artifact supervision, and Linux adversarial tests.

## Acceptance Criteria
On every claimed Linux kernel/configuration the six hard guarantees are enforced and adversarially tested; missing Landlock, namespace, seccomp or delegated cgroup capability fails closed before Go starts; no portable cache or receipt is reused as hardened.
