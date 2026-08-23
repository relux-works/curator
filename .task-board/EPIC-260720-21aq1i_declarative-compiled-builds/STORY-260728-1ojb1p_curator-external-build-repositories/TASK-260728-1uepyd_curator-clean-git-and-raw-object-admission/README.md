# TASK-260728-1uepyd: curator-clean-git-and-raw-object-admission

## Description
Implement Curator source acquisition and raw-object admission for network and operator-substituted external Git repositories using the accepted clean process and object boundary. Freeze a complete regular-file snapshot at the exact effective commit without checkout, archive, attributes, filters, or repository-selected behavior.

## Scope
Trusted Git discovery/version pin, private init/fetch, HTTPS and exact SSH wrapper, optional exact-tag-only acquisition, narrow local repository admission, pack/index inventory, full-OID cat-file stream, commit/tag/tree/blob parsing and recomputation, LFS/submodule/link/special-file rejection, race and resource bounds, and adversarial tests.

## Acceptance Criteria
Tagged declarations use exact refs/tags/<tag> as their sole acquisition path and verify terminal commit equality; untagged declarations use only the full lock; all ambient config, helpers, hooks, filters, alternates, grafts, replacement refs, promisor/partial/lazy state, submodules, LFS pointers, links, special modes, malformed objects, unexpected children/writes, and source races fail with specified typed errors before audit or cache; SHA-1 and SHA-256 network/local vectors produce the canonical raw snapshot bytes and digest; no compiler child can start; tests pass.
