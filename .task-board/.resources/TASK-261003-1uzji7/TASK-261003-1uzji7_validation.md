# TASK-261003-1uzji7 — atomic v1-to-v2 identity migration and writer flip

Handoff follow-up: TASK-261003-1uzji7_handoff-validation.md records scoped lint
exit 1 followed by exit 0, lint-only fixes, targeted regression exit 0, and a
fresh build exit 0. It supersedes the earlier pending scoped-lint/logbook
checklist statements below. Findings are recorded in outcomes under the
explicit prohibition on LOGBOOK.md edits. Hosted acceptance remains pending.

Ready-for-review evidence; hosted acceptance and integration are pending.
The candidate is uncommitted in the assigned Story worktree. No commit,
branch switch, merge, rebase, or release tag was created. LOGBOOK.md and
CHANGELOG.md were not edited, per the binding produce-mode instruction.
The remote query for v0.15.0-rc.3 returned no advertised tag (exit 0);
integration must satisfy the post-rc.3 landing condition.

## Implementation

- EnableV2Writers=true. Parsed locks select reader framing independently of
  the writer switch. Managed and in-place marker writers use lock-declared
  framing, preventing a carrier-only relabel.
- Resolve --repair, ApplyMigration and Use invoke the coordinated identity
  transaction. Genuine legacy state snapshots are verified with v1, copied
  under recomputed v2 store keys, and retained at their old keys for sharing.
  Lock-v2, regenerated headers/documents, surface entries and marker-v3 publish
  together. Explicit credential link/record changes join that transaction.
- All provisioned managed sibling environments participate, including those
  outside --env. Plans bind their markers, surfaces, rendered bytes and pinned
  store trees. Drift and invalid/mismatched identities fail before activation.
- Durable transaction rollback/recovery covers exact file and symlink
  preimages. Tool-owned seeds/credential payloads are neither copied nor read
  by identity migration. Symlink fallback copies and their metadata survive.
- Schema-1 credential metadata is expanded using observed declarations, with
  identity-only siblings retaining their recorded credential mode.
- Nine migration regressions are required in the platform-case ledger;
  only Windows may skip for unavailable symlink capability. No local skip was
  observed. Full platform-ledger verification belongs to the hosted gate.
- Only the rc.14 snapshot-acquisition/cases/byte-exact-snapshot gap row was
  removed; exact published count remains 1. Other gap/count rows are unchanged.

The detailed ownership and recovery design is in
TASK-261003-1uzji7_identity-migration.md. Native in-place surfaces follow ordinary
profile-update semantics: their old explicitly versioned markers become stale
until Use regenerates them after first invoking the same profile identity migration. The atomic conversion unit covers the profile lock
and its manager-owned homes, without adopting arbitrary native tool state.

## Commands and actual exits

All Go commands below ran directly with GOFLAGS=-work, without tee or a pipe
chain. Each test command used -count=1 -timeout=4m; the verbose checks also
used -v. Scratch WORK paths and local root paths are omitted from this public
artifact. No already-attached validation evidence substitutes for these runs.

### Red baseline and implementation diagnostics

The same command, `go test ./internal/envprofile -run
'^TestRC14MigrationRehashesLegacyIdentities$' -count=1 -timeout=4m`, ran six times:

1. Exit **1**: test compilation diagnostic, incorrect fixture request signature.
2. Exit **1**: fixture lacked Repair=true and refused an unprovisioned home.
3. Exit **1**: genuine production red baseline. ApplyMigration retained the v1
   state pin, published the v1 surface hash as v2, and left the lock at v1.
4. Exit **1**: implementation compilation diagnostic, unused import.
5. Exit **1**: strict lock reader rejected a newline appended to CCJ-1 bytes.
6. Exit **0**: genuine transition passed after canonical-byte publication.

These failures are failing checks, not passing expected-red gates.

### Scoped integration checks

- Exit **1**: envprofile selection of RC14, marker-version mismatch, legacy
  schema migration, NUL, credential publication/recovery, Pi provisioning and
  linked Use. Only the NUL legacy assertions failed: they implicitly used the
  new v2 writer. Their intended v1 writer is now explicit; rejection assertions
  remain intact.
- Exit **0**: `go test ./internal/envprofile -run
  '^(TestRC14|TestMigrateSchema1UnlinkPublishesCompleteSchema|TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember)'
  -count=1 -timeout=4m -v`. Includes both collision shapes, a clean-skill control,
  deep NUL install/update negatives, genuine legacy migrations, siblings,
  rollback, plan drift and version/store-pin mismatch negatives. Package time
  89.278s.
- Exit **0**: `go test ./internal/envprofile -run '^TestRC14' -count=1
  -timeout=4m -v`. Seven then-present top-level migration cases passed,
  including interrupted-commit recovery and schema-1 no-op credential migration
  through Resolve and ApplyMigration. Package time 133.170s.
- Exit **0**: `go test ./internal/envprofile -run
  '^TestRC14IdentityMigrationPreservesFallbackCopies$' -count=1 -timeout=4m -v`.
  The added eighth case preserves an actual directory fallback copy and checks
  the resulting home through Resolve. Package time 9.834s.
- Exit **0**: `go test ./internal/envprofile -run
  '^(TestRC14MigrationRehashesLegacyIdentities|TestRC14IdentityMigrationPreservesFallbackCopies)$'
  -count=1 -timeout=4m -v`. Repeated after adding the explicit platform-capability
  probe. Both executed without skips; package time 18.518s.
- The last credential-mode preservation recheck is recorded below.

### Hash/carrier and production install checks

- Exit **0**: `go test ./internal/hashing ./internal/contextlock ./internal/marker
  -run '^(TestManagedWriterVersion|TestContentSHA256V2|TestIdentityEqualityRequiresHashVersion|TestUnversionedLockObjectFollowsManagedWriterSwitch|TestResolvedDeltaIncludesStatePinWhenHashVersionChanges|TestWritePreservesContentHashInRC13Mode|TestReadLegacyV1AndRewrite|TestCurrentRejectsV1MarkerForV2Expectation|TestWriteAlwaysProducesCanonicalMarkerV2|TestMarkerV[34]|TestWriteUsesWireCompatibleEmptyValues)'
  -count=1 -timeout=4m`. Versioned identity comparison, NUL collision separation,
  unknown versions, legacy readers/writers and carrier mismatches executed.
- Exit **1**: `go test ./internal/install ./internal/contextresolve
  ./internal/contextstore ./internal/envprofile -run
  '^(TestEndToEndInstall|TestV2RegistryAttestationUsesVersionedInstallPath|TestV1InstallRejectsV2RegistryEvidence|TestStateEntryIsContentKeyed|TestContentHashUsesVersion2Framing|TestMinimalResolution|TestRC14IdentityMigrationRecoversInterruptedCommit)$'
  -count=1 -timeout=4m -v`. The v1-artifact/v2-registry negative fixture did not
  select v1 under the new default. The helper now selects the requested artifact
  framing explicitly. Context resolution, content-keyed state storage, v2
  framing, default Project install, v2 registry install, and recovery passed in
  this run; the overall command remains a failure.
- Exit **1**: `go test ./internal/install ./cmd/curator -run
  '^(TestEndToEndInstall|TestV2RegistryAttestationUsesVersionedInstallPath|TestV1InstallRejectsV2RegistryEvidence|TestEnvMigratePlanApplyPi|TestEnvMigrateConflictRefuses|TestEnvMigrateApplyRequiresPlan|TestEnvMigratePrintBeforeWrite)$'
  -count=1 -timeout=4m -v`. All three install cases passed after correcting
  version selection. The CLI positive case still expected marker schema 2;
  updated to require schema 3 with hash_version 2. All CLI negatives passed.
- Exit **0**: `go test ./cmd/curator -run
  '^(TestEnvMigratePlanApplyPi|TestEnvMigrateConflictRefuses|TestEnvMigrateApplyRequiresPlan|TestEnvMigratePrintBeforeWrite)$'
  -count=1 -timeout=4m -v`. All four actual run() entry-point checks passed;
  package time 57.628s.

### Exact published snapshot

The local root was the committed CI rc.14 pin
43bf0a2506d5c354a73bbc3ea4623d4653db10c7, with manifest SHA-256
6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5.
CURATOR_CONFORMANCE_ROOT was set to that private local checkout.

- Exit **1**: `go test ./internal/interop/environments -run
  '^TestConformanceSnapshotAcquisition$' -count=1 -timeout=4m -v`.
  Both autocrlf cases passed through gitops.Extract; the coverage gate truthfully
  failed because the now-passing case still had a stale known-gap row.
- Exit **0**, after removing only that row: `go test
  ./internal/interop/environments ./internal/conformancecoverage -run
  '^(TestConformanceSnapshotAcquisition|TestRC14SnapshotCaseIsDrivenAfterProfileHashMigration)$'
  -count=1 -timeout=4m -v`. Exact measured tally: **1 driven / 1 published**,
  **0 known-gap, 0 bound, 0 skipped**. autocrlf=true and false both passed.

### Build and scoped validation

- `GOFLAGS=-work go build -o .temp/TASK-261003-1uzji7/curator ./cmd/curator`:
  first and second invocations each exit **0**. The binary remains ignored and
  is not attached. The post-preservation recheck is recorded below.
- `GOFLAGS=-work go vet ./internal/envprofile ./internal/contextlock
  ./internal/hashing ./internal/marker ./internal/conformancecoverage`:
  first and second invocations each exit **0**. The post-preservation recheck
  is recorded below.
- Every executed `git diff --check` exited **0**.
- Scoped gofmt validation over tracked changed and untracked Go files exited
  **0**; the final recheck is recorded below.

No full local suite, full lint, cross-platform/race matrix, or hosted gate ran
in this produce-mode session. Hosted gate green and release/tag ordering are
**unknown/pending**, not inferred from these targeted checks. The logbook
checklist stays unchecked because the current task explicitly forbids edits.

## Last scoped rechecks

- Exit **0**: `GOFLAGS=-work go test ./internal/envprofile -run
  '^(TestRC14ResolveRepairMigratesSiblingHomes|TestRC14Schema1IdentityMigrationWithoutCredentialOperations|TestRC14MigrationRehashesLegacyIdentities)$'
  -count=1 -timeout=4m -v`. Repeated after preserving credential mode for
  siblings outside the explicit credential inventory; package time 55.560s.
- Exit **0**: `GOFLAGS=-work go test ./internal/install -run
  '^(TestEndToEndInstall|TestV2RegistryAttestationUsesVersionedInstallPath|TestV1InstallRejectsV2RegistryEvidence|TestRegistryAuditedInstall)$'
  -count=1 -timeout=4m -v`. The three existing named cases executed and passed;
  TestRegistryAuditedInstall matches no test and supplies no evidence.
  Package time 17.346s. Registry fixtures that mutate the writer selection
  run serially, so parallel install tests start after selection is restored.
- Exit **0**: `GOFLAGS=-work go test ./internal/envprofile -run
  '^(TestRC14UseMigratesLegacyProfileBeforeNativePublication|TestUseMaterializesLinkedHomes)$'
  -count=1 -timeout=4m -v`. Both production Use paths passed; package time
  8.884s. The native writer never relabels a legacy state pin.
- The third CLI build and scoped vet invocations exited **0** after the
  credential-mode preservation change.
- The fourth CLI build exited **0** after the Use entry-point change.
- The fourth scoped vet exited **0**, additionally including ./internal/install.
- Both additional scoped gofmt validations exited **0**, including the last
  check over every changed/untracked Go source. All additional whitespace
  validations exited **0**.

Every required local command has returned; no background verification is left
running. The hosted gate, full lint, full suite, race/platform execution and
post-rc.3 integration remain pending and are not claimed green.
