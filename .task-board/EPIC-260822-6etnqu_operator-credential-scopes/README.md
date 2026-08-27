# EPIC-260822-6etnqu: operator-credential-scopes

## Description
Scoped operator SSH credential selection for external build repositories, plus install-shape hardening and diagnostics improvements. Naming policy for every artifact of this epic: code, commits, docs, and task notes reference the Curator Protocol spec (curator-spec) and this repository only; never name other manager implementations.

## Scope
internal/config, internal/install, internal/buildrepo, cmd/curator, docs

## Acceptance Criteria
Operators can persist per-scope SSH credential selections in the global config and get candidate-driven onboarding; install shape hardening items verified or fixed with tests; diagnostics carry actionable provenance and remedies. CI green, interop conformance gate green.
