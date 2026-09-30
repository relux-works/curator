# STORY-260930-jzq0dx: accept-content-hash-v2-candidate-suite

## Description
curator-spec PR #116 (TASK-260917-2vapkz content-hash v2) adds frozen-shape negative cases to released families (install-marker-v4, context-lock-v1, agent-environment-marker-v2) and new families; the spec Implementations job (pinned curator 80fd617f) fails: marker/install-marker-v4/schema-cases publishes 28, pinned count 27. Lockstep: curator main must pass against both rc.13 and the candidate suite, then the spec PR pins that curator commit.

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
