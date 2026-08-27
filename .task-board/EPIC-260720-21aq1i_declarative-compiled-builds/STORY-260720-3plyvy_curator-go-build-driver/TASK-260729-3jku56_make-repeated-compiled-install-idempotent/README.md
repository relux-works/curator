# TASK-260729-3jku56: make-repeated-compiled-install-idempotent

## Description
Wire marker.BuildCurrentness into install staging so a repeated install of an unchanged compiled skill is reported as up-to-date instead of re-staging its context directory.

## Scope
Own internal/install staging (stageNode and its marker currentness call) and its tests. marker.Current is called without a marker.BuildCurrentness value, and that call fails closed for any marker carrying build state, so the context directory is re-staged on every run. Measured behavior: a second `curator install app` over an unchanged compiled project reports installed rather than up-to-date, while the plan line still reports outcome=cache-hit, so nothing recompiles. The behavior is safe but wasteful and misreports an unchanged installation. Supply the independently derived raw snapshot, cache inspection, planned inputs, and complete context and runtime file sets that BuildCurrentness requires; do not weaken the fail-closed default.

## Acceptance Criteria
A second install of an unchanged compiled project reports up-to-date and re-stages nothing; a changed build source, target, toolchain, cache entry, or context boundary still re-stages; marker.Current keeps failing closed when the independently derived build state is unavailable; the existing install, staging, and atomicity tests stay green.
