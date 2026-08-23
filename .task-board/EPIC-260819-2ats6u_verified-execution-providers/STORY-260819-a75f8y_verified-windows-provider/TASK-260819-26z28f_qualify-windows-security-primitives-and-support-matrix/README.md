# TASK-260819-26z28f: qualify-windows-security-primitives-and-support-matrix

## Description
Prove a conservative Windows mechanism composition and support matrix.

## Scope
Evaluate service and driver trust, Job Objects, restricted tokens and AppContainer, process creation callbacks, filesystem minifilter, registry callbacks, WFP, ETW limitations, WDAC, object namespaces, reparse points, alternate data streams, network bypasses, event loss, driver unload, VBS and HVCI, signing tiers, and supported Windows editions against every common capability.

## Acceptance Criteria
A reviewed matrix selects exact user-mode and kernel-mode mechanisms and minimum Windows builds, includes runnable probes and adversarial evidence, detects missing policy or partial enforcement before execution, and rejects unsupported systems rather than substituting telemetry for enforcement.
