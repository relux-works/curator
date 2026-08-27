# Audit the csk released-suite pin and CI gate

## Description
Verify that the csk implementation story advances its protocol-suite pin only to the qualified released schema v6 revision and that the compiled-build consumer is a required cross-platform gate. Apply only narrowly scoped CI corrections if the handoff diverges.

## Scope
Work in csk after TASK-260720-25d05o and the csk integrated verification task. Audit .github/workflows/ci.yml, every curator-spec checkout ref, the compiled-build conformance invocation, and any directly related no-skip guard. Candidate-phase testing may use an explicitly supplied CURATOR_CONFORMANCE_ROOT but must leave the committed release pin at the previous released suite. The eventual pin must equal the exact immutable protocol release commit recorded by TASK-260720-25d05o. Do not use a branch, mutable tag, feature-worktree SHA, guessed future release, or Curator implementation detail.

## Acceptance Criteria
Every csk CI job that supplies CURATOR_CONFORMANCE_ROOT resolves the same qualified full curator-spec release commit and pin history shows that it moved only after TASK-260720-25d05o evidence existed. Linux, macOS, and Windows across the supported Python matrix install the supported Go toolchain and execute the full Python suite consumer with no compiled-build case skipped, xfailed, duplicated, or silently filtered. The focused protocol test, full pytest suite, and strict type check pass against the released suite, existing schema 1 through 5 coverage remains active, and a negative CI guard detects an old suite that lacks compiled-build cases. Any correction is limited to CI or test-gate wiring and the outcome records exact refs and CI links without fabricating a claim.
