# repository-transport revision 2: ports, mirrors and host aliases (spec amendment)

## Description
Operator decision 2026-09-16 (option C): draft the normative curator-spec amendment for the advanced repository endpoint mappings that transport revision 1 left undecided — non-default ports, operator-declared mirrors and SSH/host aliases. Spec only; the Curator implementation is a separate leaf under STORY-260916-2txa8v.

## Scope
Repository identity, proposed Skillfile sources, permitted Git acquisition lanes, per-machine endpoint and authentication configuration, lock/audit identity, compatibility, bounded fallback and strict external-build transport integration.

## Acceptance Criteria
protocol/repository-transport.md (revision 2 section or companion file) specifies: (1) identity: canonical host/path stays the portable identity; ports, mirrors and aliases are machine-policy endpoint properties and never part of identity, lock, receipts or package manifests; (2) source-policy schema extension (additive, versioned) for endpoints with explicit non-default ports, mirror endpoints mapped to an exact canonical identity, and operator-declared host aliases resolved only through operator configuration, never from user ssh/git config; (3) bounded resolution and failure classes consistent with revision 1 (at most one attempt per listed endpoint, fail-closed table, no generated URLs); (4) secret handling and audit provenance; (5) conformance vectors (positive and refusal) under the existing conformance tree; (6) UNRESOLVED_QUESTIONS.md updated to record what is now decided; make validate passes.
