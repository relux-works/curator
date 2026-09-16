# TASK-260916-1x0ogh: spec-provider-resolution-trust-roots

## Description
curator-spec: replace ambient-PATH umbrella discovery (§11) with resolution from the manager install directory plus an explicit machine-config provider directory list, or at minimum refuse providers in directories writable by anyone other than the operator; name the S6-injected PATH case; vectors.

## Scope
protocol/environments.md §11/§12.1, vectors

## Acceptance Criteria
Environments revision merged with the trust-root rule and vectors
