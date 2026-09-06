# TASK-260906-1xitqi revision 3 refresh results

## Outcome

The accepted revision-2 release candidate was refreshed uncommitted onto authoritative main b056e5dae73be4dc92f2a948992f0941d7283f89. No commit, push, tag, release, workflow suppression, or signing bypass was performed. Stable v0.14.0 remains the recommendation; exact completed-main/public-channel inventory, green baseline main CI 34015435043, six-target snapshot results, and parent-owned publication verification remain documented in TASK-260906-1xitqi_release-evidence.md.

## Preservation proof

The preserved candidate patch sha256 is c1f3167580769f23bc12997c8e2e2420a3d2e907a232628cd00399fb48bff5e9 and its accepted revision-2 tree is d3ce5197978965820d18c8b8d86636c7ac099144. All 13 non-LOGBOOK candidate paths byte-match the preserved accepted files after application. The two accepted release-fix LOGBOOK sections byte-match the preserved candidate; removing those inserted sections makes the refreshed LOGBOOK byte-identical to b056e5d. Thus both upstream stage-b/policy history and accepted release-fix history are retained. The 86-file 7320bc2..b056e5d upstream delta remains based at HEAD, including cmd/curator env/profile and internal envfragment, envmarker, envprofile, and envregistry sources.

## Focused validation rerun on combined base

- git apply/check and git diff --check: exit 0.
- Non-LOGBOOK byte comparison against preserved accepted candidate: 13/13 match, exit 0.
- Accepted LOGBOOK section and upstream remainder comparisons: exit 0.
- bash .github/ci/release-source-gate.sh: exit 0.
- bash .github/ci/gate-selftest.sh: exit 0, 94 passed and 0 failed; missing normal-CI gate, missing release ancestry gate, and earlier extra publisher mutants are rejected at the production workflow checker.
- Focused closureexec test run for TestCaptureTreePublishesImmutableReusesAndCleansFailedStaging, TestCaptureTreePreservesInvalidExistingTargets, and TestImmutableAdmittedTreeReplayAndTimeOfUseRechecks: exit 0.

Revision 2 full-suite and independent review evidence remain applicable to the byte-identical release logic. Per refresh instructions, the configured full build/vet/test suite is deferred to and required from the managed revision-3 handoff on the combined base. Hosted CI on the eventual exact integrated SHA and all public installation checks remain mandatory parent steps after signed integration/publication.