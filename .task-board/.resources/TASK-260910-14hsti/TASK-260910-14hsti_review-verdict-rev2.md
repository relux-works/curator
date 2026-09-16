# TASK-260910-14hsti revision 2 — CHANGES_REQUESTED

Candidate: 0c42421a4b4eb5ef103947989fc954e67f3083e9; base: 12f1287ee0fb538f9ca004dd53b870e824e5baf2.
All 9 changed files independently compared byte-for-byte with the candidate Git tree: match, including after attacks. No repository code edited; attacks used Go overlays outside the repository. Scope is the four assigned packages, with no frozen schemas changed and draft/revision-1 comments present. Run is not goal-bound (spawn goal queried); no pending directives.

## Required rework

1. HIGH — Guards are not connected to acquisition or publication. Repository-wide non-test caller searches find only definitions for ValidateLocalPackage, EnumerateInputs, ValidateDestinations and TempStaging; Plan.Snapshot/Recheck are not called by the transaction publisher. New tests invoke these helpers directly. Thus the claim that all acceptance rows reach production entry points is unsupported. The existing StageProject unmanaged-conflict test is a real planning entry-point test, but manually calling ValidateDestinations afterwards does not prove StageProject enforces admitted-input protection. Wire the boundary state into actual local capture/planning and serialized publication, with negative tests driving those paths. If wiring belongs to another delivery leaf, coordinate that dependency explicitly rather than mark this enforcement AC complete on helper coverage. Do not broaden unrelated scope.

2. HIGH — Physical identity changes at the same pathname pass Recheck. internal/staging/boundaries.go:170 stores only canonical strings; line 248 compares only spelling. Independent overlay test created parent/, planned parent/out, called Snapshot, renamed parent to parent-old, created a new parent directory, then called Recheck(snapshot,nil). It returned nil. TestReviewerSamePathIdentityReplacement failed with "changed physical parent identity admitted at same spelling" (go test -count=1 -overlay <overlay.json> ./internal/staging -run TestReviewerSamePathIdentityReplacement, exit 1). The contract section 2 requires physical identity and changed-parent rechecks. Preserve and compare actual existing ancestor/object identities and cover same-spelling replacement and rollback through the real publication entry point.

## Independent checks

Host shell zsh; commands were not piped.
- go test -count=1 ./internal/snapshot ./internal/staging ./internal/privatedir ./internal/adapters: exit 0, all four packages pass (1.597s, 1.291s, 2.523s, 1.807s).
- go vet ./internal/snapshot ./internal/staging ./internal/privatedir ./internal/adapters: exit 0.
- gofmt -l on the nine changed Go files: exit 0, empty output.
- Existing candidate compiles through these test builds; no full local suite run.
- Read producer results and rev2 validation resource. Accepted as existing hosted evidence only: run 35078793023, exit 0, Ubuntu/macOS/Windows Test lanes and Ubuntu/macOS Race lanes successful. rose-air and Candidate suite skipped; their coverage is unverified. The attached log states test_case_coverage=unknown. Remote suite not rerun.

## Narrowing mutants

Both executed independently with -count=1 using temporary Go overlays; candidate bytes remained unchanged.
1. First ValidateLocalPackage output refusal narrowed from `else if pruned` to `else if pruned && directory == "."`. TestValidateManagedSourceRejected: exit 0, SURVIVED. The later post-resolution output guard still refuses this fixture; this survivor alone does not prove a behavior hole, but the test does not isolate pre-traversal refusal.
2. checkAdmittedOverlap reverse containment refusal narrowed to exact equality (`if reverse && current == input`), retaining forward protection. TestPlanRecheckRefusesAdmittedOverwriteBothDirections: exit 1, KILLED at boundaries_test.go:163, "admitted inside destination err = <nil>".
Measured result: 1/2 killed, 1/2 survived with the above bound. These are helper gate measurements, not production publication coverage.

## Acceptance trace and limits

- Physical paths/symlinks/case: TestValidateSymlinkManagedAndEscape, TestValidateCaseAliasUsesFilesystemIdentity, TestCanonicalizeResolvesSymlinksAndMissingSuffix, TestWithinCatchesSymlinkAliases — helper coverage; identity counterexample above.
- Authored versus managed, safe path dot: TestValidateBroadRootAllowed, TestValidateManagedSourceRejected, TestProjectOutputRootsCoverManagedTrees — helper coverage.
- root_inputs: TestValidateRootRequiresInputs, TestValidateRootInputsRefusals, TestEnumerateRootEntriesSelectFileOrTree — helper coverage. Producer explicitly bounds required context coverage to SKILL.md/effective manifest/runtime/build roots; do not claim all effective context inputs are independently verified.
- Reject before traversal/prune: TestEnumerateInputsPrunesAndRejects — helper coverage only.
- Publication recheck: TestPlanRecheckRejectsRetargetSincePlanning, TestPlanRecheckRefusesNewlyIntroducedLinks, TestPlanRecheckRefusesEntryLinkIntoAdmitted — helper coverage only; no transaction-publication caller.
- Unmanaged files: TestUnmanagedTakeoverRefusedBesideAdmittedCheck reaches adapters.StageProject and preserves foreign mine.md; TestCasingAliasHelperFlagsCaseVariants is helper-only.
- Production caller coverage for the four new integration surfaces checked (ValidateLocalPackage, EnumerateInputs, ValidateDestinations, Plan.Recheck): 0/4. This ratio describes caller inspection, not whole-contract test coverage.

No claim is made for runtime capture, publication rollback, absent-path Unicode aliases, or skipped platform lanes. No runtime-home, board-file, LOGBOOK.md or repository-code edits. Findings persist here and in board notes as required by the campaign prohibition on LOGBOOK.md edits.

Verdict: changes_requested. Route to to-dev for implementation and another independent review; not an external blocker.
