# STORY-260819-14jbv5: verified-linux-provider

## Description
Deliver a signed Linux verified provider using a reviewed composition of kernel enforcement and observation mechanisms.

## Scope
Select and implement the minimum supported kernel and distribution matrix using appropriate namespaces, cgroups, seccomp, Landlock, LSM or eBPF, fanotify, and network controls. Include authenticated IPC, privilege separation, packaging, updates, and platform conformance.

## Acceptance Criteria
On the declared Linux support matrix, verified mode establishes every required capability, rejects unsupported kernels and partial enforcement before workload start, produces provider-bound receipts, passes adversarial and lifecycle conformance, and ships as a signed separately installed package.
