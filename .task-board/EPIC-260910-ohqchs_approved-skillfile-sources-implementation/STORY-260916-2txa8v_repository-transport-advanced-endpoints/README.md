# STORY-260916-2txa8v: repository-transport-advanced-endpoints

## Description
Operator decision 2026-09-16 (option C): specify and then implement the advanced repository endpoint mappings that transport revision 1 explicitly left undecided — non-default ports, operator-declared mirrors, and host aliases — as a normative repository-transport revision 2 amendment in curator-spec, keeping portable identity, lock invariants and fail-closed security of revision 1. Implementation in Curator follows the accepted spec as separate leaves.

## Scope
(define story scope)

## Acceptance Criteria
Spec: protocol/repository-transport.md gains revision 2 (or a companion section) covering custom ports, mirrors and host aliases with exact identity rules, machine-policy schema changes (source-policy schema 2 or additive fields), failure classes, portable lock invariants, secret handling and conformance vectors; UNRESOLVED_QUESTIONS.md updated; make validate green; independent review accepted; landed on curator-spec main through a signed PR. Implementation: Curator resolves declared logical identities through the extended machine policy with tests and conformance vectors.
