# TASK-260926-4hd81z: accept-rc13-suite-protocol-version

## Description
internal/scriptpolicy TestScriptExecutionPolicyIdentityMatchesTheSuite (conformance_test.go ~507-510) requires the pinned suite protocol_version == 1.0.0-rc.9 (the rc.12-pinned script-worker-v1 protocol). The rc.13 release candidate (curator-spec PR #97, head f6bd748c) labels every vector 1.0.0-rc.13 with no change to script-worker-v1 or the execution-policy identity. Make the check accept exactly the closed set of protocol versions whose script-worker-v1 identity is unchanged ({1.0.0-rc.9, 1.0.0-rc.13}) while still binding schema_version, execution_policy and the interpreter set; keep it failing for any other version (e.g. rc.14 or a renamed policy). Verify against BOTH curator's own pinned suite (dcc7f015, rc.9 labels) and the rc.13 candidate root (conformance/v1 from PR #97 head).

## Scope
(define task scope)

## Acceptance Criteria
test passes against the rc.9-labelled pinned suite and the rc.13 candidate; a mutant accepting any version is killed (e.g. rc.14 row); go test ./internal/scriptpolicy green on both roots with real exit codes; no CHANGELOG edit
