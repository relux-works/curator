# TASK-260729-3jmqgl: prototype-macos-hardened-capability-probes

## Description
Build executable macOS capability probes that test the six hardened build guarantees before curator/csk integration: network denial, read-only source and toolchain, writes confined to a private build directory, process-tree/resource/disk/time limits, allowlisted executable launch, and fail-closed admission when any guarantee is unavailable. Produce reusable evidence and a platform capability matrix without claiming production enforcement.

## Scope
macOS-primary, task-owned worktree. Probe and test harness only; no normative curator-spec changes, no curator/csk production integration, no weakening or silent fallback, no staging/commit/publish. The separate hardened story remains non-gating for the main compiled-skill work.

## Acceptance Criteria
Each guarantee has a positive capability test and an adversarial escape/negative control; the harness reports a stable machine-readable result and exits nonzero when any required capability cannot be established; unsupported/private/deprecated platform mechanisms are identified explicitly; tests run on ssh relux or the local macOS host with exact commands and exit codes; outcome documents what can be reused by both curator and csk macOS implementations.
