# TASK-260924-1fnikm: promote-skillfile-sources-to-released-protocol

## Description
PARKED 2026-09-24: per the operator handoff, the curator-spec promotion of skillfile-sources-v1 and repository-transport rev1-2 (remove unreleased, move draft schemas and conformance/draft-sources-v1 vectors to the released namespace, add the global-scope rule) is authored from the other manager side. This leaf only reviews that external PR against the gap matrix TASK-260924-3re9jo when it opens, and does not author a competing change.

## Scope
(define task scope)

## Acceptance Criteria
no 'unreleased'/'opt-in draft' wording left for these extensions; schemas and cases in the released corpus with regenerate-check clean; validate green; schema-1 cases untouched
