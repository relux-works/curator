# BUG-260729-r0fe02: stabilize-godriver-cancellation-race-contract

## Description
The accepted Curator composite intermittently fails under the mandatory race gate: internal/godriver TestFingerprintCancellationStaysFailClosed/cancelled_between_the_walk_and_the_digest can observe fail-closed toolchain_mutated while the assertion permits only nil or toolchain_timeout. Determine the intended cancellation-vs-mutation precedence and make the contract deterministic without weakening fail-closed behavior.

## Scope
Own only internal/godriver fingerprint cancellation behavior and its focused tests, plus task-scoped evidence. Start from the exact accepted TASK-260720-jrrgw9 candidate. Do not change CI, protocol vectors, timeouts, product manifests outside godriver, or broad error acceptance.

## Acceptance Criteria
A focused -race repetition reliably proves the cancellation boundary contract; legitimate concurrent mutation remains toolchain_mutated, cancellation at the defined phase boundary returns the specified stable fail-closed result, and the test no longer flakes. Non-race focused godriver tests pass, the accepted product behavior is not weakened, and the patch is independently reviewed before TASK-260720-1pvfj5 resumes its final race gate.
