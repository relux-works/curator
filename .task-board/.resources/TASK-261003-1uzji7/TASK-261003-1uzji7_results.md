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


## Converged base e5489b6b

Republish verification on 2026-10-04, following uzji7-republish.md. The actual
Story HEAD is 54bed271b7609bf206a04369202473c430d0d96a, a descendant of
requested base e5489b6ba22c9e9cf7a925e03f3653d115188999; the later upstream
commit narrows audit trust-pin handling. No branch commit, switch, rebase,
merge, or product-code edit was performed in this run.

Carried-delta verification: 22 modified tracked paths and 6 untracked paths,
28 total. The tracked diff reports 256 insertions and 99 deletions; untracked
files are excluded from that stat. Direct byte comparison against
refs/campaign/uzji7-snapshot-20261004 (1f8e0da6810e304fb48b4a46a425fe45409ecd35)
found 27 identical paths. The remaining platform ledger differs only by the
three converged N3 rows; all nine migration rows are intact. The design,
implementation, regressions, writer flip, and exact conformance gap deletion
were retained. LOGBOOK.md and CHANGELOG.md remain untouched.

Commands rerun here directly as standalone processes, without pipes:

- `GOFLAGS=-work go test ./internal/envprofile -run
  '^(TestRC14|TestMigrateSchema1UnlinkPublishesCompleteSchema|TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember)'
  -count=1 -timeout=4m -v`: exit **0**, package time **98.326s**. All nine
  rc.14 top-level cases passed, including the restored genuine-v1 regression,
  fresh-v2 pin/surface comparisons, sibling migration, exact rollback,
  plan drift, marker/lock/store-pin mismatch refusal, interrupted-commit
  recovery, schema-1 no-credential operations through both production entries,
  fallback preservation, and Use. Both legacy schema-1 credential cases and
  three top-level NUL/collision negatives passed. **14/14 top-level cases
  passed; 0 skipped**.
- `GOFLAGS=-work go test ./internal/hashing -run
  '^(TestManagedWriterVersion|TestContentSHA256V2|TestIdentityEqualityRequiresHashVersion|TestByteLayout|TestMarkerExcluded|TestEmptyTree|TestOrderIndependence|TestNormalize)'
  -count=1 -timeout=4m -v`: exit **0**, package time **0.490s**.
  **12/12 top-level cases passed; 0 skipped**, covering default-v2 and explicit
  version selection, legacy byte framing, v2 vectors including embedded NUL,
  v1 collision separation, adjacent-record boundaries, version-aware equality,
  and unknown-version refusal.
- `git diff --check`: exit **0**.
- `git merge-base --is-ancestor e5489b6b HEAD`: exit **0**.
- `git merge-base --is-ancestor v0.15.0-rc.3 HEAD`: exit **0**. The locally
  present annotated rc.3 tag peels to ca1b776fb580ec0cee0173bf150daf063023aeaa.
  This proves local ancestry only; release integration remains the
  orchestrator's responsibility.

The original red-first, install/CLI/context entry points, exact production
snapshot acquisition (**1 driven / 1 published, 0 known-gap/bound/skipped**),
scoped lint and build evidence above and in
TASK-261003-1uzji7_handoff-validation.md is accepted as **prior-run evidence**,
not claimed as rerun on the converged base. This run compiled and executed
only the targeted envprofile and hashing tests. No standalone CLI build,
lint, vet, full suite, race/platform matrix, snapshot acquisition, or hosted
validation was launched here, per the latest limited republish instruction.
The hosted CR gate result is **unknown in this run** and remains the arbiter.
No command hung beyond five minutes; no retry or background validation remains.

The named results outcome was absent from the live outcome inventory, so this
republish creates it with the existing validation history and this appended
section rather than pretending an existing results resource was updated.
Findings are recorded here instead of the prohibited LOGBOOK.md edit.
The candidate is ready for developer handoff and review; acceptance is pending.


### Republish handoff checklist reconciliation

The first `task-board handoff TASK-261003-1uzji7 --role developer` attempt
returned exit **1**, refusing unchecked items 9, 10, 11 and 12. It did not
publish a successful handoff. The CLI also warned that it could not compare
run board content; no command-progress credit is inferred from that warning.

Items 9 and 10 are supported by the carried implementation, the attached
ownership/rollback design and prior production-entry evidence, with byte-exact
retention verified here. The migration uses the existing transaction engine;
legacy reader framing is independent from the default writer, and verified
snapshots are rehashed before activation rather than relabelled. Item 11 is
supported by the two freshly green targeted commands and the prior scoped
checks; it does not assert a full-suite or hosted-gate result.

Item 12 is conditionally **not applicable at this developer handoff**. The live
board notes record that one prior reviewer run was cancelled for host load and
the next failed preparation because the pre-converge platform ledger overlapped
the incoming base. Neither issued a review verdict. No changes-requested
verdict exists to attach or route in this run. Checking this conditional item
records that absence, not reviewer acceptance; the orchestrator still owns
review and any subsequent rejection/rework branch. The latest convergence
addresses the documented preparation conflict. No product-code repair was
required by the handoff refusal.


## Converged base e5489b6b — republish follow-up 2026-10-05

The latest republish instruction was followed without product-code changes.
Actual HEAD remains 54bed271b7609bf206a04369202473c430d0d96a;
`git merge-base --is-ancestor e5489b6b HEAD` exited **0**.
`git merge-base --is-ancestor v0.15.0-rc.3 HEAD` also exited **0**.
These establish local ancestry; remote tag verification and release integration
remain owned by the orchestrator.

`git status --short` and `git diff --stat` exited **0** and retain the same
28 carried paths: 22 modified tracked paths and 6 untracked paths, with
256 tracked insertions and 99 deletions. A direct byte comparison against
refs/campaign/uzji7-snapshot-20261004 exited **0**, confirming 27 identical
paths. The platform ledger retains every snapshot row and adds only the three
converged upstream restore-mode rows. No carried implementation was lost.
LOGBOOK.md and CHANGELOG.md have no delta and were not edited.

### Validation executed in this follow-up

The following command ran directly as a standalone process, without a pipe,
with GOFLAGS=-work and fresh execution (-count=1):

```sh
GOFLAGS=-work go test ./internal/envprofile ./internal/hashing -run '^(TestRC14|TestManagedWriterVersion|TestContentSHA256V2SeparatesTheV1Collision|TestContentSHA256V2LengthFramingSeparatesAdjacentRecordBoundary|TestIdentityEqualityRequiresHashVersion|TestContentSHA256V2RejectsUnknownVersion|TestMigrateSchema1UnlinkPublishesCompleteSchema|TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember)' -count=1 -timeout=4m -v
```

Actual command exit: **0**. Environment profile: **14/14 top-level cases
passed**, package time **142.942s**. Hashing: **6/6 top-level cases passed**,
package time **0.959s**. **0 skipped** in either package.
All nine RC14 regressions ran, including TestRC14MigrationRehashesLegacyIdentities,
fresh v2 state/surface computation, sibling migration, rollback, drift refusal,
marker/lock/store-pin mismatch negatives, interrupted-commit recovery,
schema-1 no-credential-operation migration, fallback copies, and native Use.
The legacy reader check executes through Resolve before migration, followed by
v2 Resolve checks. Both credential schema-1 cases and all three NUL/collision
negative entry tests passed. Hashing covered the default v2 writer, explicit
writer selection, v1 collision separation, adjacent-record framing,
version-aware equality, and unsupported-version refusal.

`git diff --check` ran directly and exited **0**. No command was retried,
terminated, or left running. The tracked-run directive query returned no
directives (exit **0**).

The earlier attached red-first regression, install/CLI/context production
entry tests, exact snapshot acquisition (1 driven / 1 published, 0 known-gap,
0 bound, 0 skipped), scoped lint and build checks are accepted as **prior-run
evidence**, not commands rerun here. No standalone build, vet, lint, full
local suite, race/platform matrix or snapshot test was launched in this
follow-up, following the latest targeted-only republish instruction. The
hosted CR validation gate was not queried or launched here; its result is
**unknown in this run**, and it remains the acceptance arbiter.

Findings and decisions are recorded in this updated task outcome instead of
LOGBOOK.md, as required. The live checklist was already fully checked; this
follow-up does not infer hosted acceptance from that checklist. The existing
implementation remains uncommitted for the developer handoff snapshot.


## Converged base e5489b6b — current republish verification 2026-10-05

Followed the latest developer republish instruction. No product code or tests
were edited in this run. HEAD is 54bed271b7609bf206a04369202473c430d0d96a;
requested base e5489b6b and the local v0.15.0-rc.3 tag are ancestors (each
ancestry check exited **0**). Release integration and remote tag authority
remain the orchestrator's responsibility.

`git status --short` and `git diff --stat` exited **0**. The carried delta
still contains 28 paths: 22 modified tracked and 6 untracked. The tracked
stat is 256 insertions and 99 deletions. Direct comparison against
refs/campaign/uzji7-snapshot-20261004 exited **0**: 27 files are byte-identical;
the platform ledger differs and retains every snapshot row. No implementation,
regression, design note or conformance gap deletion was lost. LOGBOOK.md and
CHANGELOG.md have no delta and were not edited. No commit, switch, rebase,
merge or tag creation was performed.

### Fresh commands and actual exits

Each test ran directly as a separate standalone process, without pipes,
with GOFLAGS=-work, -count=1 and a bounded package timeout.

```sh
GOFLAGS=-work go test ./internal/hashing -run '^(TestManagedWriterVersion|TestContentSHA256V2|TestIdentityEqualityRequiresHashVersion)' -count=1 -timeout=4m -v
```

Exit **0**, package time **0.328s**; **7/7 top-level cases passed**, **0 skipped**.
This proves the default v2 writer, explicit version selection, v2 conformance
vectors (including embedded NUL), v1 collision separation, adjacent-record
framing, version-aware equality and unsupported-version rejection.

```sh
GOFLAGS=-work go test ./internal/envprofile -run '^(TestRC14|TestResolveRejectsEnvironmentMarkerHashVersionMismatch|TestInstallBlocksBothV1CollidingSkillTrees|TestInstallBlocksDeepNULFileInContextPathSnapshot|TestUpdateBlocksDeepNULFileInNewContextMember)' -count=1 -timeout=4m -v
```

Exit **0**, package time **95.446s**; **13/13 top-level cases passed**,
**0 skipped**. All nine RC14 regressions executed, including the restored
TestRC14MigrationRehashesLegacyIdentities, real legacy-to-v2 state and surface
rehashing, sibling migration, rollback of every transaction entry, plan drift
refusal, marker/lock/store-pin mismatch refusal, interrupted commit recovery,
schema-1 migration through Resolve and ApplyMigration, fallback preservation
and native Use. The additional four cases prove environment marker version
mismatch rejection and NUL/collision rejection through Install and Update,
with a clean-skill positive control.

`git diff --check` exited **0**. The tracked-run directive query exited **0**
and returned no directives. No command hung beyond five minutes, no command
was retried or terminated, and no verification process remains running.

### Evidence boundaries and decisions

This run freshly compiled and executed only the authorized targeted envprofile
and hashing selections: **20/20 top-level cases passed**, **0 skipped**.
Original red-first evidence, other install/CLI/context production entries,
legacy carrier readers, exact snapshot acquisition (**1 driven / 1 published,
0 known-gap, 0 bound, 0 skipped**), scoped repository lint and CLI build are
accepted from the already-attached outcomes as **prior-run evidence**; they
were not rerun here. The carried conformance delta deletes only the owned
rc.14 byte-exact-snapshot gap, supported by that prior exact-case evidence.

No standalone build, vet, lint, full local suite, race/platform matrix,
snapshot acquisition test or hosted validation command ran in this follow-up,
as directed by the targeted-only republish instruction. Hosted CR validation
was neither launched nor queried; its result is **unknown in this run** and
remains the acceptance arbiter. The already-checked developer checklist does
not assert hosted acceptance. Findings and decisions are recorded in this
updated task outcome instead of the prohibited LOGBOOK.md edit.

The candidate remains uncommitted and ready for developer handoff to review.
