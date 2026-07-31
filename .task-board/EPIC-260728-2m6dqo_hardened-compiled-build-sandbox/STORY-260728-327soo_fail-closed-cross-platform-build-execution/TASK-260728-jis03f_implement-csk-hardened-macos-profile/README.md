# TASK-260728-jis03f: implement-csk-hardened-macos-profile

## Description
Implement or capability-gate the independent csk hardened macOS execution path against the accepted profile.

## Scope
Python manager and signed/native worker TCB, macOS sandbox or VM boundary, signing/notarization, receipts/cache separation and adversarial tests on ssh relux.

## Acceptance Criteria
csk independently proves all six guarantees on every claimed macOS configuration or rejects before Go starts; helper identity, signing and package-influence exclusions are reviewed.
