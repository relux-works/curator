# TASK-260729-3dr6hw: diagnose-install-race-timeouts

## Description
Read-only diagnosis of the exact uncached go test -count=1 -race ./... timeout in internal/install and internal/install/atomicity after cmd/curator passed. Produce a minimal test-only rework plan preserving behavior, unchanged package timeout, and candidate integrity.

## Scope
Inspect verifier3 race evidence and candidate test structure only. Attribute cumulative runtime by test/scenario and fixture setup, identify safe fixture reuse or scenario partitioning already compatible with existing patterns, quantify expected savings, and give a literal producer file/function allowlist. Do not edit candidate, change test timeout, skip cases, weaken assertions, run the full/race suite, install tools, or touch protocol/product behavior.

## Acceptance Criteria
Outcome records exact failing tests and timings from verifier evidence; maps every proposed optimization to preserved assertions and isolation invariants; recommends the smallest test-only patch with expected margin below 10 minutes for both packages under race; includes focused validation commands that do not overlap the active verifier; reviewer independently checks the plan.
