# TASK-260728-ns5yk7: implement-csk-hardened-linux-profile

## Description
Implement the independent csk hardened Linux execution path against the accepted normative profile.

## Scope
Python manager TCB, identity-verified worker boundary, Linux kernel controls, receipts/cache separation and shared plus manager-specific adversarial tests.

## Acceptance Criteria
csk independently enforces and proves all six hard guarantees on every claimed Linux profile, rejects missing capabilities before Go starts, and does not trust Curator private implementation outputs.
