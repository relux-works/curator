# EPIC-260825-2p2kl3: https-build-repository-credentials

## Description
Authenticate private HTTPS build repositories through the manager credential broker the manager profile reserves, reaching feature parity with the SSH credential surface this repository already ships. Naming policy for every artifact of this epic: code, commits, docs, and task notes reference the Curator Protocol spec and this repository only; never name other manager implementations. The spec was amended for this work: profiles/manager.md now states what a broker may resolve, and core.md 12.2 now requires a manager offering an identity-unbound credential selection to let the operator bind it to a host and to document the exposure.

## Scope
internal/config, internal/install, internal/buildrepo, cmd/curator, docs

## Acceptance Criteria
A private HTTPS build repository authenticates end to end; no secret reaches config, flags, logs, receipts, markers or diagnostics; the broker fails closed on a foreign host, prompt, state or absent material; anonymous HTTPS is unchanged; the run-wide override is host-bindable per core 12.2; full gate set and interop conformance gate green; landed through a pull request.
