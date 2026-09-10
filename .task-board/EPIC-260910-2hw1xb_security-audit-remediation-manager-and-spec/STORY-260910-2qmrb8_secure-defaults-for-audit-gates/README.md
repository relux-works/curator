# STORY-260910-2qmrb8: secure-defaults-for-audit-gates

## Description
Findings S1+S3 (Medium): audit mode, registry policy and both allowlists default permissive (advisory/advisory/empty/empty), and an unreachable registry during install yields only a routine warning, so revocation hiding via network DoS is possible under defaults. Recommend/ship hardened defaults and make the residual explicit.

## Scope
curator-spec registry.md/core.md + internal/config defaults + install-time warnings

## Acceptance Criteria
A hardened-defaults profile is specified; install-time unreachable-registry is surfaced prominently; SECURITY.md names the residual that revocation is network-dependent under advisory policy
