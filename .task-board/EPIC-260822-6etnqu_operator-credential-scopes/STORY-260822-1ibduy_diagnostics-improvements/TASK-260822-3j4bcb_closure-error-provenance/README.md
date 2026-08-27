# TASK-260822-3j4bcb: closure-error-provenance

## Description
When a dependency-closure node's manifest fails to parse, the error names the node and its resolved ref (e.g. Invalid skill manifest for <name> tag <ref> (via <chain>)); today the error surfaces under the root declaration and reads as if the declaring skill were broken. Wrap the parse call sites in the closure/install planning paths.

## Scope
(define task scope)

## Acceptance Criteria
Test with a broken transitive manifest asserts name+ref+chain in the error; protocol error codes unchanged; go test green.
