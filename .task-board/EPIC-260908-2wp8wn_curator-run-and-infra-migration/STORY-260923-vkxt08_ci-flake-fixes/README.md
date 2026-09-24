# STORY-260923-vkxt08: ci-flake-fixes

## Description
Hosted-gate flakes that cost the campaign republish cycles, fixed one at a time so they can run beside the CI-readiness Story.

## Scope
Test determinism and CI budget only; no product behaviour change.

## Acceptance Criteria
Each flake is fixed with the failing row made deterministic, proven by repeated hosted runs, without weakening the semantics it guards.
