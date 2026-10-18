# Review sweep — SNAP-PUBLISH revision 1

Candidate tree: 2e20b225f2794c9e524ee69332ed156424c612d8.
Base: 1724b4a52b4c2d163c0a3e6ab7c2ad2bbe54ac0c; fresh origin/main equals base.
The supplied brief has no numbered catalog table or time budgets. Use its four paired rows in supplied order. AC coverage was inspected: 5 of 5 library rows have named driving tests; CLI integration is explicitly separate (0 of 5 CLI rows). Windows qualification is refusal only.

1. Admission / protected identity — held. Executed TestDestinationReadOnlyAdmission, Boundary, OwnershipAndReadFailure, KernelACL, PriorKernelACL, ProtectedIdentity and PrivateParent through OpenDestination. Real macOS no-follow/ACL/permission attacks pass alongside valid controls. Narrowing M2, owner, writable mode, protected equality, empty target, private parent and ACL plants each cause named child failure (exit 1).
2. Replacement / competition — held. Executed TestDestinationReplacementRaces, TestPublicationNoReplaceRace and TestPublicationRetainedParentRace. Kernel-created empty directory, snapshot, file and link after final validation are preserved. Ancestor relocation returns refusal without redirected writes. No-replace weakening causes named failure.
3. Prior identity / idempotence — held. Executed TestDestinationPriorIdentity, PriorMissingMetadata, PriorKernelACL and TestPublicationPrivateAndPrior. Bytes/mode/extras/unreadable/version/generation changes refuse and exact prior succeeds. M6 and prior-member ACL plants cause named failures.
4. Failure / cleanup / complete publication — held. Executed TestPublicationStagingOnlyCleanup, KernelStagingFailure, CleanupSubstitutedRoot, CapturedIdentity and PostVerification. Cancellation and kernel failure leave no successful partial output; foreign sibling links survive; post-rename mismatch remains for operator. Three cleanup/post-verify plants cause named failures.

Commands: go test -count=1 -race -v ./internal/envsnapshot (exit 0); go test -tags snapshotcontrols -count=1 -run '^TestDestinationNarrowingControls$' -v ./internal/envsnapshot (exit 0, 16 named child failures). Full remote gate pending. These results are bounded macOS execution, not proof of absence. Supplemental free-hunt probes pending.

Free hunt: TestReviewerPublicationRetry (cancel, competitor, post-mismatch) and TestReviewerStageRootBeforeRename passed with -race through Publish, in an archive of the exact candidate plus only the supplemental test. The initial test-file creation used a wrong scratch-relative path and the ensuing no-tests invocation is NOT evidence. After correcting the scratch path, both named tests and all three retry subtests executed and passed (free-hunt.log). No product source was modified.

Fresh curator-spec main f9970ddf01ff6b52faaed3721ef49ce9dd9c5ca7 was read via GitHub API: snapshot sections 2/4/5, CIP-0014 destination/idempotence rows and environments section 5. The current absent-only/private-stage/post-verification boundary matches the owner decision. Caller search confirms CLI integration is absent, consistent with the explicit leaf bound.
