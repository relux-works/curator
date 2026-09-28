# TASK-260924-291k0q: gate-skillfile-sources-independence-from-rc11-rc12-core

## Description
Operator decision: cocoaskills is a partial client (core v1.0.0-rc.10 + skillfile-sources-v1). curator-spec v1.0.0-rc.13 already pins conformance/skillfile-sources-v1 by its own manifest against core rc.10, and the Implementations CI checks cocoaskills against rc.10 + skillfile-sources-v1. Add a CI gate that keeps this independence a checked invariant: every $ref from schemas/skillfile-sources-v1 into schemas/v1 must resolve to a definition byte-identical at tag v1.0.0-rc.10 (fail on absent/different); protocol/skillfile-sources.md and repository-transport.md must not cite a core/registry/manager/environments clause introduced after rc.10 (an explicit informative-only marker allowed, see #96 review note N1); the gate runs in the existing Specification CI.

## Scope
(define task scope)

## Acceptance Criteria
Gate job green on the candidate; a mutant that repoints one draft $ref to an rc.12-only v1 definition fails it; a mutant prose citation to an rc.12-only clause fails it; README of draft-sources-v1 states the rc.10 baseline for partial clients; CHANGELOG entry under unreleased.
