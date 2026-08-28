# TASK-260819-267jf8: qualify-linux-kernel-primitives-and-support-matrix

## Description
Prove a conservative Linux mechanism composition and minimum kernel or distribution matrix.

## Scope
Evaluate namespaces, mount and immutable views, cgroups v2, seccomp notify, Landlock, fanotify, LSM or eBPF hooks, netfilter or network namespaces, pidfds, io_uring and alternate syscall paths, overlay filesystems, container nesting, event loss, lockdown, Secure Boot, and privilege requirements against every common verified capability.

## Acceptance Criteria
A reviewed matrix selects exact supported kernel configurations and mechanisms, provides runnable probes and adversarial evidence, detects missing hooks and partial enforcement before execution, and rejects unsupported kernels or configurations instead of weakening guarantees.
