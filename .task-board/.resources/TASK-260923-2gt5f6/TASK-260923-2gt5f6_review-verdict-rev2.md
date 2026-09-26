# TASK-260923-2gt5f6 review verdict — rev2 ACCEPTED (reviewer claude-opus-5-5 low, full profile, role body present)

Candidate tree: the worktree's temp-index write-tree is 28d207d0c129e3e9e7b568377ea675176365c267, which matches the CR. rev2.patch sha256 is 2266d400… and matches.

## rev1 failure and fix
The rev1 hosted gate failed on internal/envprofile::TestMigrateNoSecretCopies. Provisioning now writes a schema-2 Pi marker, so migration rewrites that marker with provenance=migrated (changedMarkers, migrate.go). The test still expected only 2 changed paths.
The rev2 fix changes the test to expect 3 paths (the link, the Pi marker and the Codex marker) and asserts that the Pi record has provenance "migrated".
rev2 also adds a darwin-only ledger row for TestEnvResolveCredentialRecordIsolatedKeychain, plus a skip-class. The fix is correct: the new expectation follows the spec (a v2 home records migration provenance), and schema-1 relinks still keep their bytes (migrate.go changedMarkers `VersionV1 && len(dropped)==0 → continue`).

## Findings
- Record written via op.publish (transaction engine inside beginOperation's manager lock) at managed.go applyPlan. Provenance is provisioned or repaired. Migrate publishes records with provenance=migrated.
- Schema-1 marker kept when legacyMarkerProjectionMatches holds (managed.go applyPlan). Otherwise the whole marker is replaced with v2.
- Linkless records have no path: ambient / keyring-backed Codex / isolated per-home-keychain (credentialRecords, validateCredentialRecord).
- Pinned suite vectors: agent-environment-marker-v2 schema-cases (26) are driven by marker_env_schema_test.go with a ledger count row.
- Crash between temp and rename: TestResolveRecoversCredentialMarkerAfterRenameCrash uses the transactionOptions seam.
- Hosted gate for rev2: run 36235825429 succeeded on every lane. rose-air was skipped, so ARM64 is unverified.
- No CHANGELOG/LOGBOOK edit and no stray files (18 paths: product/tests/docs/.github/ci).

## Reruns (zsh, pipefail, -count=1)
- go test ./cmd/curator -run 'TestEnvResolveCredentialRecord|TestEnvResolveKeepsSchema1|TestEnvMigrate' → ok, exit 0
- go test ./internal/envmarker ./internal/envprofile -run 'Credential|Migrate|Marker|Schema' → ok, exit 0

## Mutants (disposable rsync copy)
- M1: schema-1 rewrite (`if false && …legacyMarkerProjectionMatches`). TestEnvResolveKeepsSchema1BytesForMetadataOnly fails, exit 1. KILLED.
- M2: path written for the keyring (ambient-backend) Codex record. TestEnvResolveCredentialRecordPathlessKeyring fails, exit 1. KILLED. It is caught first by the internal effective-link guard, not the schema validator.
- Lock-less mutant: not run by me (bound). Covered by the producer's TestResolvePublishesCredentialMarkerThroughLockedJournal.

## Residuals
- Pre-feature survival is not measurable because the base has no v2 writer (producer note).
- The macOS Keychain row is darwin-only by design.
