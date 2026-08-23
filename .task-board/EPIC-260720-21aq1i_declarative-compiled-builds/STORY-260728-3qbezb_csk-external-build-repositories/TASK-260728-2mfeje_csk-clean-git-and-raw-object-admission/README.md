# TASK-260728-2mfeje: csk-clean-git-and-raw-object-admission

## Description
Implement csk clean Git acquisition and raw-object admission independently in Python for network and local-substituted repositories. Do not rely on git archive, checkout, repository attributes, or source-controlled configuration unless byte-equivalence to the accepted raw-object contract is formally proven.

## Scope
Trusted Git probe, private init/fetch, HTTPS/SSH wrapper, exact-tag and full-lock paths, narrow local repository config/ref/pack admission, fixed full-OID cat-file protocol, object parsers and hash recomputation, complete graph proof, LFS/submodule/link/special-file rejection, bounds/race handling, and adversarial tests.

## Acceptance Criteria
Network and local sources produce canonical SHA-1/SHA-256 raw snapshot bytes identical to the spec vectors; tagged declarations always fetch only the exact tag and verify its terminal commit; no ambient Git behavior, helper, hook, filter, alternate, replace, graft, promisor, lazy fetch, submodule, LFS pointer, link, special mode, malformed object, unexpected process, or source race can pass admission; errors occur before audit/cache/compiler; Python and shared vector suites pass.
