# BUG-260731-27h1yc: curator-windows-lane-masked-test-failures

## Description
Curator Test (windows-latest) fails in five packages beyond internal/runtimestore. These were fully masked on main: the go vet step aborted the job before any test ran (BUG-260731-11bpa4), so no Windows test result has ever been produced in Curator CI. They surfaced for the first time on BUG-260731-11bpa4 PR 10 run 30619686990 job 91121004339, where go vet now succeeds and the platform-case gate reaches real execution. Failing required cases: internal/buildsource TestFrozenTokenRejectsRootReplacement; internal/install TestEndToEndInstall; internal/install/atomicity TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder, TestStaleAdapterRemovalRollsBackToTheExactPriorEntry, TestAdapterMirrorLinksAreJournaledAndRestoredExactly. Pre-existing and not caused by PR 10: that PR touches only .github/ci/gate-selftest.sh, .github/ci/toolchain-identity.sh, internal/interop/golden_test.go and the two internal/runtimestore test files, none of which are imported by the failing packages. Out of scope for BUG-260731-11bpa4, whose AC is package-scoped to internal/runtimestore. Detailed per-test output is in the Upload gate evidence artifact of that job (.temp/ci-evidence/test/go-test.json), not in the job log, which prints only stage exit codes.

## Scope
Curator Windows lane for internal/buildsource, internal/install and internal/install/atomicity. Diagnose each failure on a native windows-latest runner and fix the real Windows behavior or the Windows expectation. Do not delete, skip or platform-exclude a required case to make the gate pass, and do not weaken the platform-case ledger.

## Acceptance Criteria
Curator Test (windows-latest) reports no required-case failures for internal/buildsource, internal/install or internal/install/atomicity, with the platform-case ledger unchanged or strengthened rather than relaxed.
