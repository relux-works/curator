# STORY-260906-2rjb4q: stage-b-managed-homes-and-fragment

## Description
Implementation stage (b) (curator, Go): managed homes with their surfaces and marker records; provisioning seeds and passthrough strategies per adapter with the isolation matrix and liveness rows; read-only env resolve with lock-free verification, environment_home_stale and --repair under the mutation lock; the closed launch-env-fragment-v1 with lock_sha256, precedence, system_prompt, mcp and path_prepend; MCP launch-channel materialization per adapter with the package allowlist; the curator run umbrella dispatch; env status matrix. Conformance subset: the referenced-*, system-prompt-composed and mcp-* expected sets that stage (a) left stage-deferred. Spec: curator-spec main f39f4a9.

## Scope
(define story scope)

## Acceptance Criteria
Every listed surface implemented behind the cli/curator.md rows; the referenced, system-prompt and MCP vector sets pass byte for byte through CURATOR_CONFORMANCE_ROOT and their stage-deferred skips are removed; go build/vet/test and the platform-case gate green; PR reviewed and landed on curator main by fast-forward of the reviewed head.
