# STORY-260905-1n0iy8: stage-a-context-packages-lock-and-monolithic

## Description
Implementation stage (a) core (curator, Go): agent-context.json and agent-mcp.json parsing/validation, versions and npm-shaped ranges, joint resolution and the semver lock (CCJ-1, lock_sha256), per-package store entries (git via object-database extraction, local), always-strict audit with the unpinnable context-secret-material detector and scoped waivers, monolithic materialization with the curator-root-context-v2 header under both precedence primitives, linked switching as one transaction with versioned backups and the environment marker, profile install/list/update/remove/use/sync CLI, global-scope migration into the default profile lock. Conformance subset: context-versions.json, context-detectors.json, environments.json monolithic and weights sets, snapshot-acquisition.json via CURATOR_CONFORMANCE_ROOT. Spec: curator-spec main fd237ba. Base: curator main after the acquisition PR #58 lands.

## Scope
(define story scope)

## Acceptance Criteria
Every listed surface implemented behind the cli/curator.md rows; the conformance subset passes byte for byte against curator-spec fd237ba conformance/v1 in the candidate lane; go build/vet/test green; hosted CI green apart from the known adapter-suite redness; PR reviewed and landed on curator main by fast-forward of the reviewed head.
