# BUG-260920-2wyzde: registry-evidence-exact-match-not-implemented

## Description
From TASK-260910-1xya7x review rev3 (corpus cases attestation-evidence-wrong-name, attestation-evidence-wrong-context): draft §4 requires network-Git members to use existing registry evidence only with exact name, canonical repository, commit and context hash matching; registry.Matches (internal/registry/registry.go:297) is content==content || (identity && commit) — the record name is never compared and a wrong context hash is accepted when identity+commit match. No exact-match layer exists on the draft lane.

## Scope
internal/registry matching for draft-lane evidence, its callers in internal/install/audit

## Acceptance Criteria
On the draft lane, registry evidence is admitted only with exact name + canonical repository + commit + context hash; both corpus cases flip to driven-pass; wrong-name and wrong-context records refused fail-closed at install.Project; legacy v1 matching unchanged; narrowing mutants (drop name compare, drop context compare) killed.
