# TASK-260916-y4sa6s: spec-signer-allowlist-and-update-confirmation

## Description
curator-spec: specify a lockable per-source signer allowlist (SSH/GPG tag or commit signatures verified before a candidate enters the lock) in Decision 0012 / environments §4 and §12.1, the profile update resolved-version delta with mandatory per-run confirmation when a class: system module or MCP declaration is introduced or changed, and the named residual for latest; conformance vectors for accepted, unsigned and wrong-signer candidates.

## Scope
curator-spec decisions/0012, protocol/environments.md §4/§8/§12, schemas + conformance vectors

## Acceptance Criteria
Environments revision merged with the signer allowlist, the update-delta confirmation rule and vectors
