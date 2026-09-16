# TASK-260910-1a75qd — review verdict revision 2

Verdict: CHANGES_REQUESTED. Route to to-dev.

Reviewed base 12f1287ee0fb538f9ca004dd53b870e824e5baf2 and exact candidate tree b31b6991b8deaf4e80577edff93aa5619fcf8bc9. All five worktree Go files were byte-compared with candidate blobs and matched. No repository code was edited; attacks used temporary Go overlays.

## Required rework
P1: internal/sourcelock/sourcelock.go:177 validRepository delegates to identity.ValidCanonical, which checks path grammar but does not enforce repository-transport revision 1 section 1 normalization (one terminal .git removed; identity MUST already be normalized). A network package with repository example.org/kit.git is accepted by New, Write and Read. This permits different portable package identities/store digests for an endpoint and its canonical identity.

Independent reproduction: append a temporary overlay test to sourcelock_test.go; copy goldenMembers(), replace each network-git member's Repository with example.org/kit.git, call New(goldenManifestSHA256, members), Write(PathIn(t.TempDir()), lock), then Read(path). Require rejection. Command go test -count=1 -overlay <temporary-overlay.json> ./internal/sourcelock -run TestReviewerRejectNoncanonicalRepository exited 1: "Write and Read accepted noncanonical repository example.org/kit.git". This drives persisted lock validation, not just a helper.

Fix normalization validation in the scoped lock module without widening legacy identity semantics. Add a persisted-reader regression with a correctly recomputed lock digest so rejection cannot be credited to unrelated integrity failure. Keep canonical example.org/kit positive. Rerun narrow gates and hand off a new revision.

## Independent verification
- zsh: go test -count=1 ./internal/sourcelock: exit 0, 0.458s.
- go vet ./internal/sourcelock: exit 0.
- gofmt -l internal/sourcelock: exit 0, no output.
- Narrowing mutant M1: enforce member/package directory agreement only for network-git, not configured-git. TestDirectoryAgreement exit 1; killed.
- Narrowing mutant M2: duplicate destination detection exact-case only. TestDuplicateNamesConflict exit 1; killed.
Measured mutant result: 2/2 killed, 0/2 survived. Temporary overlays left original bytes unchanged.
- Read producer results and revision-2 validation log. Hosted run https://github.com/relux-works/curator/actions/runs/35077873374 succeeded, including Ubuntu/macOS/Windows tests, lint, conformance, Ubuntu/macOS race checks. Gate commit d6a2800843d2c4a92a30f20281fe378183ff2afa resolves to the exact candidate tree above. This is accepted attached platform evidence, not a locally rerun full suite. rose-air and Candidate suite were skipped; no passing claim for them.

## Acceptance mapping and bounds
- Persistence/exact identities: TestWriteReadRoundTrip drives Write -> Read; normalization hole above remains.
- Frozen selection: TestSelectionSemantics drives New -> Parse and CheckMembership, distinguishing zero, null and collection siblings.
- Machine separation: TestMachineSeparation, TestBindingsRoundTrip, TestRebindingPreservesIdentity cover portable object shape and separate bindings persistence/digests.
- Stale/malformed: TestCheckStale and TestMalformedLocks drive CheckStale and Parse.
- Membership: TestCheckNames, TestCheckMembership, TestDuplicateNamesConflict, TestUnsortedMembersRejected cover exact set/records/order and conflicts.
- Three arms: TestThreeArmsRoundTrip and TestCrossArmConfusionRejected; TestDirectoryAgreement covers both Git arms.
- Context lock isolation: TestContextLockConfusion cross-rejects the two readers.
These are package API tests; repository search found no production imports of sourcelock outside this package. They do not establish resolve/install/launch orchestration or a complete local installation. Snapshot capture, refresh, source-policy interpretation, physical-path canonicalization and publication integration remain sibling work; Bindings.Validate checks lexical absolute/clean shape only.

Scope is limited to the five new sourcelock files, uses existing protocoljson, and leaves frozen wire schemas untouched. Draft labeling and dedicated-package separation fit the project. The Windows fixture/mode repairs are supported by the hosted run.

Contract inspected: skillfile-sources section 3/5, repository-transport revision 1, source-types-v1 and skillfile-lock-v1 schemas. Schema fixture structural success is not normative normalization proof.

Finding recorded here as the durable review log; LOGBOOK.md edits are prohibited by the campaign instructions. Run goal queried before verdict: not goal-bound. No human decision or external blocker is required.
