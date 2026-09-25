# STORY-260923-laeycm: ci-flake-windows-snapshot-concurrency

## Description
Own Story so the Windows snapshot concurrency fix runs in parallel with the other flake fixes.

## Scope
Test/runner determinism for one flake; no weakening of the guarded contract.

## Acceptance Criteria
The flake is fixed with evidence and repeated green hosted runs.
