# TASK-261003-1uzji7 — design-v1-v2-hash-migration-and-writer-flip: revision 5 review

Verdict: **accepted**. No acceptance-blocking findings. Acceptance authorizes the producer integration lifecycle; this reviewer does not commit, integrate, supply commit_ack, or mark delivery done.

## Exact scope and provenance

- CR-TASK-261003-1uzji7-5, revision 5; base `77fabd45b88b7fd839a85f017d863e03c926953e`; candidate tree `9aed2f6612261f1861497da2bf6eb0eee879a4f4`.
- Reviewed the entire 39-path delta, migration design and producer evidence. A read-only comparison confirmed 39/39 changed worktree files byte-equal to the candidate blobs.
- The CR validation resource names hosted run [37315678860](https://github.com/relux-works/curator/actions/runs/37315678860), reports successful completion and `[exit 0]`. Independently queried its GitHub jobs and downloaded its artifacts. GitHub commit `b8c879a18da7e324cfc6c5d0d559e6958a2e0e60` names exactly the candidate tree above; the green run is therefore evidence for this revision.
- Fresh remote HEAD advertisement and exact main fetch agree at `40bc4c6dcf79f2eb7d26f3a703e8f224c7bcab9b`. Main changes since the CR base are board-only, with no candidate overlap or changed validation inputs. This does not claim a combined-tree validation run.
- The remote advertises rc.3 tag object `5fae9ada1d807a584f80f45a1fadd7894d9eb4db`, peeling to `ca1b776fb580ec0cee0173bf150daf063023aeaa`. `git merge-base --is-ancestor` against the CR base exited 0. The post-rc.3 prerequisite is satisfied.
- `task-board spawn goal` reported no active goal: this run is not goal-bound.

## Hosted execution evidence

Read `go-test.json` and the platform-case outputs from the five downloaded test/race artifacts. Counts below are terminal test records including subtests, not counts of independent acceptance criteria.

| Hosted lane | Passed records | Skipped records | Failed records | New migration roots |
| --- | ---: | ---: | ---: | --- |
| Test Linux | 10055 | 62 | 0 | 9/9 passed |
| Test macOS | 10678 | 36 | 0 | 9/9 passed |
| Test Windows | 10239 | 319 | 0 | 9/9 passed |
| Race Linux | 10055 | 62 | 0 | 9/9 passed |
| Race macOS | 10678 | 36 | 0 | 9/9 passed |

All 45/45 required migration-root executions passed, with no migration skip, including the schema-1 resolve/migrate and version/pin mismatch subcases. The nine roots are:

- TestRC14MigrationRehashesLegacyIdentities
- TestRC14ResolveRepairMigratesSiblingHomes
- TestRC14IdentityMigrationRollsBackEveryEntry
- TestRC14IdentityMigrationRefusesSiblingPlanDrift
- TestRC14IdentityMigrationRefusesVersionAndPinMismatch
- TestRC14IdentityMigrationRecoversInterruptedCommit
- TestRC14Schema1IdentityMigrationWithoutCredentialOperations
- TestRC14IdentityMigrationPreservesFallbackCopies
- TestRC14UseMigratesLegacyProfileBeforeNativePublication

Hosted Lint, all three platform gates/self-tests, interop, naming gate, and all nine Go-driver build jobs succeeded. The CR command exited 0. Self-hosted rose-air and optional candidate-suite jobs were skipped; Windows race is not configured. No coverage for those lanes is claimed. Existing unrelated skips remain recorded and checked by the platform ledger.

Per the binding hosted-evidence instruction, I ran **no local Go build, test, vet, lint, red-baseline execution, or mutant execution**. Earlier producer local results are historical evidence and were not rerun here. Reviewer read-only checks, GitHub download/query commands, and exact candidate `git diff --check` returned exit 0.

## Swept surfaces and acceptance checks

| Surface | Code inspected and conclusion | Executed evidence / bound |
| --- | --- | --- |
| Rehash and immutable store | identity_migration.go:159 verifies old pins using lock framing; :194 copies the old snapshot with EnsureState, :198 independently recomputes v2 and compares the new key before assigning StateHash. Old entries are retained; git commit pins retain their independent meaning. | Genuine v1 Resolve then real ApplyMigration; unchanged old pin is rejected, new lock pin equals fresh v2 computation; 5/5 lanes pass. |
| Atomic publication | identity_migration.go:54 prepares all lock/document/surface/marker/credential entry replacements with exact preimages and commits through the existing home-locked transaction engine. Managed siblings outside --env join the same transaction. | Rollback, interrupted recovery, sibling conversion and fallback-copy regressions all pass. Rollback injects after target index 2 and compares every entry in profiles/environments/native scope; this is not an exhaustive injection at every possible fault point. Newly cached immutable store entries may survive without activating a profile. |
| Recovery and transient readers | beginOperation recovers durable journals before mutation. Prepare stages durable sidecars before scratch cleanup; commit errors select rollback. Lock-free readers refuse mismatches rather than accepting mixed identities. | The panic seam enters real Commit after target index 2; subsequent Resolve repairs/replays and verifies both sibling markers and credential unlink. This is an in-process interruption simulation, not a newly run OS-kill experiment. |
| Input refusal and drift | identityMigrationInputs binds lock, source, stores, sibling markers/surfaces and rendered targets. ApplyMigration checks the expected plan before staging and rechecks bound input digests before publishing. No failed read is treated as absence. | Sibling plan drift refuses without changing profile scope. Marker-version, lock-version and store-pin negatives pass, alongside plan/print/conflict refusal tests. |
| Route boundaries | identity_migration.go:423 uses the hardened environments root for managed-home routes; profile lock/doc targets get containment and the existing managedPath component-link walk. Input validation still checks marker, lock and immutable-store boundaries/pins. | Windows regressions now reach Commit without requiring the ordinary manager-home root to have an owner-only DACL. The profile-document route deliberately follows existing profile publisher semantics; it does not add a new owner-only profiles-root policy. |
| Legacy readers/currentness | store_boundary.go:202 and managed.go:254 use the lock's version. Core status paths use recorded.ContentHashVersion rather than the new writer default. Parsed locks retain explicit schema/version. | Genuine v1 bare Resolve under v2 writers passes. Legacy marker rewrites/currentness, frozen-lock rejection and own-version reader tests pass on all three test platforms. |
| Marker schemas and fixtures | Core §10 requires core marker v5 for every manifest band, retaining v4 build-record rules. Environments §8.2 requires marker v3 with hash_version 2. Frozen 1/2 environment fixtures now provision genuine v1 bytes rather than relabelling v2 digests. | Build-bearing schema-band/status/GC cases, CLI end-to-end status, schema-1 byte-preserving legacy-lane repair, pre-rule seed preservation and real v1-to-v2 schema-1 migration all pass. |
| NUL / cross-version negatives | Frozen v1 collision fixtures explicitly select v1; v2 framing separation and unknown-version refusals remain active. Version mismatches are not fallback identities. | Both colliding skill trees, deep-NUL Install/Update, v2 NUL/record-boundary vectors, identity version inequality and v1-artifact/v2-registry refusal pass in all five lanes. |
| Exact conformance gap | Only snapshot-acquisition/cases/byte-exact-snapshot is removed from the rc.14 gap ledger. Published count remains 1; the coverage regression also refuses a wrong-framing outcome after removal. | Actual production git extraction case passes for autocrlf=true and false in all five lanes: **1 driven / 1 published, 0 known-gap, 0 bound, 0 skipped**. |
| Protected scope | Full delta contains no LOGBOOK.md, CHANGELOG.md, release-pin or remote-gate.sh change. EnableV2Writers=true is tested. | Exact candidate path list and hosted lint/naming gates support the review; no product code changed during this review. |

## Writer call-site coverage

All eight design families were traced to production and executed evidence: **8/8 reviewed and covered**, with the bounds stated below.

| Design family | Production route | Hosted test evidence |
| --- | --- | --- |
| Install target/publication | install/targets.go, install.go -> Project install and registry publication | TestEndToEndInstall; TestV2RegistryAttestationUsesVersionedInstallPath; CLI end-to-end install/status |
| Core markers | marker.Write / Current; schema-zero expected objects select the new writer, parsed carriers select themselves | TestWriteEmitsCoreMarkerV5WithHashVersion2; authoritative v2 writer; production install/status. Frozen draft-package carriers deliberately keep v1 pending the separate spec follow-up. |
| Context resolution | contextresolve.Resolve selects state-pin framing | TestMinimalResolution; legacy fixture Install and real migration/Use |
| State store | contextstore.EnsureState -> ContentHash | TestStateEntryIsContentKeyed; TestContentHashUsesVersion2Framing; real migration verifies new store pin |
| Context locks | contextlock default construction plus canonical publication | Production fixture Install; migration lock-v2 assertions; explicit default-switch, frozen-lock and schema conformance negatives |
| Materialized surfaces | assembleHome and native materializeOne use lock-declared framing | Genuine migration fresh-v2 surface computation, sibling Resolve verification, fallback-copy and native Use regression |
| Environment marker publication | Resolve --repair and finalizeMarker | TestResolvePublishesCredentialMarkerThroughLockedJournalV2WriterMode; schema-1 migration through Resolve |
| Environment migration publication | PlanMigration / ApplyMigration; changedMarkers follows the carrier's own framing | Genuine v1 regression, schema-1 unlink, drift/rollback/recovery; CLI TestEnvMigratePlanApplyPi and negative apply cases |

## Red-first and mutant analysis (read-only)

Against the base's changedMarkers, an unchanged v1 marker is copied and only Version/HashVersion plus credential records change. Neither its StateSHA256 nor its lock/store identity is converted. The restored TestRC14MigrationRehashesLegacyIdentities explicitly turns the writer on after genuine v1 production provisioning: its oldPin equality assertion, fresh v2 surface-hash assertion and lock.HashVersion assertion therefore fail on that base behavior. This establishes defect sensitivity by reading; it is not a claimed executed red run in this review.

A mutant assigning the old member.StateHash instead of EnsureState's new key is killed by oldPin equality and the fresh-v2 store assertion. A mutant merely relabelling marker surface hashes is killed by hashing the actual published files with v2. Narrowing conversion to only the requested environment is killed by sibling marker assertions and subsequent Resolve verification. Narrowing the aggregate identity-validation gate to a marker-only comparison, bypassing lock-framed store verification, is killed by the store-pin case: its marker version remains valid, but PlanMigration must still refuse the changed store bytes. The lock-version case additionally rejects a v2-labelled lock over genuine v1 bytes. Removing only one check can survive if another independent check still refuses the same input; these tests do not establish that every individual comparison is independently necessary. Bypassing the joint transaction is killed because rollback requires the fault to return an error and restore all entries, and interruption recovery requires the transaction hook to fire and later complete both siblings plus credential unlink. No mutant was executed.

## Normative reference and disposition

Read the protocol files at the committed CI spec pin `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`: [core §§8,10](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/core.md) and [environments §§1.3,5.6,8.2](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md). The marker pin changes follow those requirements; local v1 fixture selection preserves frozen meanings rather than relaxing validation.

Findings and decisions are persisted here instead of LOGBOOK.md, as the binding brief prohibits editing that file. Accept revision 5 with this task-scoped evidence and route to integrating. Landing remains the authorized producer's integration transaction; review acceptance is not a landing.
