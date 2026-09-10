# STORY-260910-6bo7ej: tofu-and-equivocation-mitigations

## Description
Finding S2 (Medium): first-use snapshot pinning is trust-on-first-use with no authenticated bootstrap checkpoint, and divergent monotonic views across clients (equivocation) have no client-side detection. Provide checkpoint tooling and an optional cross-registry check.

## Scope
curator-spec registry.md + registry client

## Acceptance Criteria
A signed bootstrap checkpoint interchange is specified; TOFU residual is documented; an optional cross-registry Merkle-root comparison warning exists
