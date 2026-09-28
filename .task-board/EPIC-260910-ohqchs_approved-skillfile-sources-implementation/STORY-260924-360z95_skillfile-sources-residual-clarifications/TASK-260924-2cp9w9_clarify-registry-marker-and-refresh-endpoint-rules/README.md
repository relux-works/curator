# TASK-260924-2cp9w9: clarify-registry-marker-and-refresh-endpoint-rules

## Description
curator-spec: settle the three group-A interpretations from gap matrix TASK-260924-3re9jo with normative text + distinguishing vectors: (1) registry evidence: which dimensions authorize a positive attestation and whether a non-exact revoked record still denies under advisory policy; raw package tree hash vs lock-projected context hash; (2) install-marker v5 local go-v1 arm: raw JSON closure (null external-only members invalid) and declared_tag ref-name grammar (empty/invalid invalid); (3) explicit refresh of an existing checkout uses the current resolved source-policy endpoint (not a stale stored origin) and scp spelling with an alias port (render or fail closed). Keep each change minimal; coordinate with the other manager's pending promotion PR by editing only these clauses.

## Scope
(define task scope)

## Acceptance Criteria
three clauses stated normatively with distinguishing vectors; validate recipe lines + regenerate-check green; no other normative change
