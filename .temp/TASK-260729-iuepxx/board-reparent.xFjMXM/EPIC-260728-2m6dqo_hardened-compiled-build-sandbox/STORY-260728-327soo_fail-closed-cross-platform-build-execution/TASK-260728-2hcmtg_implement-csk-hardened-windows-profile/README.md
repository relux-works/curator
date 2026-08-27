# TASK-260728-2hcmtg: implement-csk-hardened-windows-profile

## Description
Implement or capability-gate the independent csk hardened Windows execution path against the accepted profile.

## Scope
Python manager and native worker TCB, AppContainer/Job Object or Hyper-V boundary, receipts/cache separation and adversarial tests on ssh win.

## Acceptance Criteria
csk independently proves all six guarantees on every claimed Windows configuration or rejects before Go starts; administrator privileges alone never satisfy capability preflight.
