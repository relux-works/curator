# TASK-261003-1uzji7 rev5 — validation: rev4-gate remainder (developer)

Base: 77fabd45 (converged). Rev4 hosted gate run 37307472674 was RED with 2 failures on ubuntu/macOS (family C remainder) and 12 on Windows (11 envprofile leaves + the same 2). This run fixes all of them. Prior rework4 fixes and evidence (TASK-261003-1uzji7_rework4-validation.md) still stand; this resource covers only the rev4 remainder.

## Fix table

| Family | Root cause | Fix (file:line) | Failing tests it should turn green | Verified |
|---|---|---|---|---|
| C-remainder (all platforms) | Under v2 writers the downgraded fixture is a v1 marker over a v2 lock (v2 digests relabelled under schema 1). Repair keeps the bytes via the legacy projection, then the post-repair currency check (managed.go:2721 verifyHome: marker HashVersion 0 vs plan 2) refuses it: environment_repair_failed, exit 1. Keeping relabelled bytes current would violate rehash-never-relabel and environments §8.2 (current writers MUST use schema 3). | cmd/curator/env_credential_marker_test.go:151,213 — pin the frozen v1 lane for both schema-1 tests (install/provision/downgrade/repair), restoring exact base conditions: genuine v1 bytes, bytes-kept repair. Reuses the serial pinV1MarkerWriters helper; no production change. | TestEnvResolveKeepsSchema1BytesForMetadataOnly, TestEnvResolvePreservesPreRuleCodexSeedAndReportsUnstrippedHome (ubuntu/macOS/Windows) | Hosted-pending (cmd/curator forbidden locally). Argued green: identical flow was green at base; v1 lane untouched since. |
| Windows boundary (10 leaves) | prepareIdentityMigration validated every staged live route with ValidateRouteWithOwner(req.Home, ...) (identity_migration.go:426). That API proves the ROOT node itself owner-only; the manager home is never a hardened root (Install publishes profile records through plain MkdirAll parents, lock.go:105; real operator homes carry inherited access), so on Windows its inherited DACL fails the owner-only mutation check: permissions: <home> boundary check failed. Sibling publishers (op.publish, writeStoreDoc, contextlock.Write) never validate req.Home; the marker boundary uses EnvRoot (store_boundary.go:101). Same latent defect at :361 for docs against ProfilesDir. | internal/envprofile/identity_migration.go:423-451 — per-target hardened roots: managed-home targets validate against EnvRoot (same boundary as preflight + validateMarkerBoundary, owner seam preserved); lock/rendered-doc targets get containment + the managedPath component link-walk (same discipline as writeStoreDoc); anything outside both trees is refused fail-closed. identity_migration.go:360-367 — docs pre-validation removed, covered by the unified loop. No validation weakened: symlink/containment still refused on every route; ownership/DACL still proven on the credential-bearing managed-home routes. | TestMigrateSchema1UnlinkPublishesCompleteSchema3MarkerV2WriterMode, TestRC14MigrationRehashesLegacyIdentities, TestRC14ResolveRepairMigratesSiblingHomes, TestRC14IdentityMigrationRollsBackEveryEntry, TestRC14IdentityMigrationPreservesFallbackCopies, TestRC14Schema1IdentityMigrationWithoutCredentialOperations/migrate, .../resolve, TestRC14UseMigratesLegacyProfileBeforeNativePublication (Windows) | Hosted-pending (internal/envprofile forbidden locally). Unix behavior unchanged for legit trees (same checks minus the always-passing home root-self node). |
| Windows downstream (1 leaf) | TestRC14IdentityMigrationRecoversInterruptedCommit panics inside the transaction fault hook; the boundary failure above aborted prepare before the transaction started, so no panic fired: unexpected interruption: <nil>. | Same boundary fix; no test change. | TestRC14IdentityMigrationRecoversInterruptedCommit (Windows) | Hosted-pending, same gate. |

## Spec decisions (binding for the pins)

- Environments §8.2: current writers MUST use schema 3 with hash_version 2; schemas 1/2 carry no hash_version. A test demanding v2-writer repair keep schema-1 bytes contradicts the MUST — the bytes-kept repair is specified behavior of the frozen v1 lane only, hence the v1 pin, not a production weakening.
- Rehash-never-relabel holds: the v1-pinned downgrade converts genuine v1 digests (provisioned under v1) to a genuine v1 marker; no v2 digest is ever presented under a v1 schema. The synthetic v1-marker/v2-lock state remains refused fail-closed (exit 1), which is the design (.research identity-migration note: marker/lock mismatch is refused).

## Commands actually run (real exit codes)

1. gofmt -l internal/envprofile/identity_migration.go cmd/curator/env_credential_marker_test.go → exit 0, no output (clean).
2. git diff --check → exit 0, no output.

## NOT run (binding rework4/host rules R193/R194)

- internal/envprofile and cmd/curator tests: forbidden locally; the hosted CR gate is the arbiter. The changed lines were verified by reading only (targets enumerated against both roots; unix-green flows re-checked for behavior preservation).
- go build, go vet, lint, full suite, race/platform matrix: not run; hosted gate owns them.
- LOGBOOK.md, CHANGELOG.md, scripts/remote-gate.sh: untouched. Conformance counts/ledger: untouched.
