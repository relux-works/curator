# BUG-260923-krcm6m: windows-registry-future-bound-not-refused

## Description
Hosted run 35901867517 (windows-latest, candidate BUG-260923-11jgkt rev2): internal/registry TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/0s/one_second_past_the_bound failed: registry_test.go:734 skew 0s offset 1s: refused=false want true (subtest took 1.55s). The test passes an explicit now; find what in CheckSnapshotsWithPolicy (or the snapshot cache in the temp dir) is time- or platform-dependent at skew 0 and fix the cause (product or test), never widen the bound.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
root cause with evidence (hosted artefact); fix at the cause; the exact-edge assertion keeps its strength; the narrowing mutant (threshold +1s) still killed; row stable under repetition (-count=50) locally
