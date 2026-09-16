# TASK-260916-3dmjbc: implement-script-manager-worker-invocation

## Description
R1 of TASK-260916-3gcc00_reconciliation.md: implement the script manager/worker invocation path in Curator (Go): consume the schema-8 script execution fields, resolve and bind the node-v1 / python3-v1 interpreter identities, create the fixed authenticated manager re-execution path, route installed enforced shims through it, bind streams and private runtime paths, tear down descendants. Replace the unconditional script_execution_policy_unsupported admission with real admission for supported policies while keeping unsupported policies fail-closed.

## Scope
(define task scope)

## Acceptance Criteria
Enforced script commands with node-v1/python3-v1 launch through the manager-owned worker; unsupported policies still refuse; the script-worker-v1 conformance vectors for launch are driven at the production entry (cmd/curator + install/shim path).
