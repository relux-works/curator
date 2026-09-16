# TASK-260910-1a75qd revision 3 — ACCEPTED

Reviewed candidate tree 2b4b02fbd92a35524f561c877dc73efae11cc61c against base 12f1287ee0fb538f9ca004dd53b870e824e5baf2. All five working files match candidate blobs before and after verification; no source bytes modified. Diff is confined to internal/sourcelock. Frozen schemas, parser, collections and legacy identity semantics unchanged. Package documentation labels unreleased extension and separates context locks explicitly.

Contracts inspected: skillfile-sources sections 3/5, repository-transport revision 1 section 1, draft skillfile-lock-v1 and source-types-v1 schemas, producer results and revision 3 hosted validation log. No blocking findings remain. Prior P1 is fixed in validRepository, reached through Package.Validate from New, Write and Read/Parse. Correctly recomputed persisted digest ensures normalization, not integrity, causes refusal.

## Independent evidence

Shell zsh. go test -count=1 ./internal/sourcelock: exit 0 (0.579s). go vet ./internal/sourcelock: separately rerun exit 0. gofmt -l internal/sourcelock: exit 0, no output. No full local suite run.

Two narrowing mutants executed with Go overlays and -count=1, leaving candidate files untouched:
- M1 restrict suffix rejection to /.git: exit 1; TestRepositoryValidation and TestNoncanonicalRepositoryRejected fail.
- M2 restrict suffix rejection to single-level repository paths: exit 1; TestRepositoryValidation fails on github.com/acme/kit.git.
Measured kill rate 2/2; survivors 0/2. This is targeted coverage, not exhaustive clause mutation.

Additional test-only overlays substitute Example.org/kit, https://example.org/kit and git@example.org/kit into TestNoncanonicalRepositoryRejected. All three runs exit 0, driving constructor, New, Write, Parse and persisted Read with recomputed digest. Together with committed .git regression, 4/4 requested noncanonical classes are refused through those entry points. Canonical example.org/kit remains accepted. Extra spelling probes are reviewer evidence, not added repository tests.

Hosted evidence accepted from TASK-260910-1a75qd_change-request_rev3-validation.log: run https://github.com/relux-works/curator/actions/runs/35083634775 exit 0. Its commit 2174d7a636c6eb2bdb84c9aff5366a6eee79721f resolves locally to exact reviewed tree 2b4b02fbd92a35524f561c877dc73efae11cc61c. Ubuntu/macOS/Windows tests, lint, interop, naming, race and gate self-tests green. Candidate suite and rose-air skipped; no rose-air or exhaustive platform-case-equivalence claim.

## Acceptance traceability

7/7 task acceptance categories exercised at scoped package API entry points:
1. Persistence and frozen selection: TestWriteReadRoundTrip (Write/Read), TestNewSortsMembersAndSealsDigest (New), TestSelectionSemantics (Parse/CheckMembership).
2. Portable identity vs machine binding: TestMachineSeparation (Parse/Object), TestBindingsRoundTrip (WriteBindings/ReadBindings), TestRebindingPreservesIdentity (Digest/CheckFresh).
3. Staleness: TestCheckStale (CheckStale), TestBindingsCheckFresh (CheckFresh).
4. Malformed locks: TestMalformedLocks and TestSchemaFixtures (Parse), new persisted Read regression.
5. Membership: TestCheckNames and TestCheckMembership, duplicate/ordering negative tests.
6. Three identity arms: TestThreeArmsRoundTrip and TestCrossArmConfusionRejected (Parse), TestPackageDigestsGolden.
7. Context/package lock separation: TestContextLockConfusion drives both readers.

Bounds: this leaf is the dedicated lock model and validation API, with no manager/CLI consumer yet. End-to-end install, launch, refresh orchestration, snapshot availability/capture, audit, marker/receipt publication and physical binding resolution remain sibling work; package tests do not establish those behaviors. Bindings validate lexical native absolute paths, not filesystem canonicalization or refresh comparison of actual locations. Existing context-lock implementation remains untouched and its boundary regression passed. Schema fixtures with placeholder digests/directory mismatch prove structural bounds only, correctly documented by producer. No live credentials or runtime-home access.

Run goal queried: not goal-bound. No directives. Acceptance is for revision 3 only and routes to integrating, not done. This task-scoped outcome records findings in lieu of forbidden LOGBOOK.md changes.
