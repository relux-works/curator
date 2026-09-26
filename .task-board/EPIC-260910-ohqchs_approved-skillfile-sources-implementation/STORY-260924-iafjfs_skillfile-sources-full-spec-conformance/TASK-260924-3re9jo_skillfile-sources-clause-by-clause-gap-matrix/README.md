# TASK-260924-3re9jo: skillfile-sources-clause-by-clause-gap-matrix

## Description
Research: clause-by-clause matrix of curator-spec main protocol/skillfile-sources.md + repository-transport.md (rev 1 and 2) + the draft-sources-v1 schemas and their conformance cases against Curator main: implemented+driven / implemented-not-driven / known-gap (with the board element or review residual that recorded it) / missing / draft-gated (how Curator gates schema 2 today and what the spec requires of a reader that supports the extension). Output: the matrix, the exact list of implementation leaves to create (scope, AC, dependencies, parallelisable groups), and whether the spec itself needs a promotion change (unreleased extension -> released) with the precise spec edits.

## Scope
(define task scope)

## Acceptance Criteria
matrix covers every normative MUST/MUST NOT and every draft-sources conformance case; each row cites spec line and Curator file:line or test; leaf list is directly creatable; no product change
