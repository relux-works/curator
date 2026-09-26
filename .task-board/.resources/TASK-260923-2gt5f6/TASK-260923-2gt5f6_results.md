# TASK-260923-2gt5f6 results

## Implementation

Managed-home markers now read schema 1 and schema 2. New provision and repair writes publish schema-2 passthrough credential records with isolation, strategy, source role, backend, verified backend release, and provenance. Linkless Codex keyring/ambient and isolated macOS Claude per-home Keychain records omit `path`; linkable credentials carry their home-relative path. Marker writes go through the manager-home operation lock and transaction journal, using sibling staging and atomic installation. A schema-1 marker stays byte-for-byte unchanged when credential metadata is the only potential marker change; an unlink that already requires marker replacement upgrades it with the complete schema-2 record set.

Migration plan hashes include credential records. Migration publishes records with `migrated` provenance. Status rendering and link inventory skip pathless entries. Documentation and CI conformance/skip ledgers cover schema-v2 and the platform-limited Keychain case. `CHANGELOG.md` and `LOGBOOK.md` were not edited.

## Production-entry coverage

- `TestEnvResolveCredentialRecordProvisionAndRepair`: CLI resolve provisions a record, then repairs it with `repaired` provenance.
- `TestEnvResolveCredentialRecordPathlessKeyring`: CLI resolve emits an ambient keyring record without a serialized `path`.
- `TestEnvResolveCredentialRecordIsolatedKeychain`: CLI resolve emits an isolated Keychain record without `path` on macOS.
- `TestEnvMigratePlanApplyPi`: CLI migration emits a Pi record with `migrated` provenance.
- `TestEnvResolveKeepsSchema1BytesForMetadataOnly`: exact schema-1 marker bytes survive repair when no marker-state change is needed.
- `TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker`: an unlink-driven schema-1 replacement upgrades with the complete schema-2 record set.
- `TestResolvePublishesCredentialMarkerThroughLockedJournal`: marker publication uses a same-directory stage while a second manager-home lock is refused.
- `TestResolveRecoversCredentialMarkerAfterRenameCrash`: a simulated interruption before install leaves a journal/stage and the next resolve recovers the complete marker.

## Pinned conformance

The local curator-spec main is `a21905db733e2b9a85b2781d3e027f4031d78662`. `git diff --quiet dcc7f015e2d97edf2d52928afb6fd79ec8129e8b main -- conformance/v1/schema-cases/agent-environment-marker-v2` exited 0, confirming the v2 case family matches the pinned `dcc7f015` family. `TestParseAuthoritativeEnvMarkerV2SchemaCases` drove 26/26 cases through the production parser: 0 known-gap, 0 bound, 0 skipped. The suite has no separate credential-record vector family; path presence/absence and record enums are covered by these v2 schema cases. CLI-emitted markers round-trip through the strict reader and are checked field-for-field by the production-entry tests.

## Validation commands and real exit codes

- `CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envmarker -run '^TestParseAuthoritativeEnvMarkerV2SchemaCases$' -count=1 -v` — exit 0; 26/26 driven.
- `go test ./internal/envmarker -count=1` — exit 0.
- `go test ./internal/envprofile -run '^(TestResolvePublishesCredentialMarkerThroughLockedJournal|TestResolveRecoversCredentialMarkerAfterRenameCrash|TestResolveProvisionRepair|TestCodexKeyringAmbient|TestMigratePiWrongTargetToAgentRoot|TestMigrateNoSecretCopies|TestMigrateSchema1UnlinkPublishesCompleteSchema2Marker|TestMigrateInterruptedApplyRecovers|TestMigratePublishFailureRevertsLinks|TestCredentialLinkUnrecordedSymlinkRefuses)$' -count=1 -v` — exit 0.
- `go test ./cmd/curator -run '^(TestEnvResolveCredentialRecordProvisionAndRepair|TestEnvResolveCredentialRecordPathlessKeyring|TestEnvResolveCredentialRecordIsolatedKeychain|TestEnvResolveKeepsSchema1BytesForMetadataOnly)$' -count=1 -v` — exit 0.
- `go test ./cmd/curator -run '^TestEnvMigratePlanApplyPi$' -count=1 -v` — exit 0.
- `go build ./internal/envmarker ./internal/envprofile ./cmd/curator` — exit 0.
- `go vet ./internal/envmarker ./internal/envprofile ./cmd/curator` — exit 0.
- `golangci-lint run ./internal/envmarker ./internal/envprofile ./cmd/curator` — exit 0, 0 issues.
- `gofmt -l` over all modified Go files — exit 0, no files listed.
- `git diff --check` — exit 0.
- `CURATOR_CONFORMANCE_ROOT=... CI_LEDGER_GOOS='linux darwin windows' bash .github/ci/ledger-consistency.sh /tmp/TASK-260923-2gt5f6-ledger-consistency` — exit 0, 418 rows checked.
- `platform-case-gate.sh` with a synthetic skip event for the Keychain case — exit 0 for Linux and exit 0 for Windows. The actual test passed on macOS in the CLI test command above.

The remote Change Request revision 1 validation exposed `TestMigrateNoSecretCopies` still expecting only the relinked link and Codex marker to change. Migration now also writes the required Pi migration record, so the test was updated to expect both migrated-home markers and assert Pi provenance. The test failed before that expectation change (exit 1) and passed afterward (exit 0); the focused migration suite also passed afterward.

A broad `go test ./internal/envprofile -count=1` run in this recovery session was interrupted with Ctrl-C after about 4m30s (exit 1); it did not report a test assertion. The focused task-specific package selection above passed uncached. The configured full landing suite is left to the handoff validator and was not run manually.

## Mutant checks

The existing producer run applied these mutants temporarily, ran each candidate killer standalone, then restored the source. These mutation sites are unchanged by the migration-test expectation and CI-ledger updates in this revision:

| Mutant | Candidate killer and exit | Pre-feature selector and exit |
|---|---|---|
| Replace journal publication with direct marker `WriteFile` | `go test ./internal/envprofile -run '^TestResolvePublishesCredentialMarkerThroughLockedJournal$'` — exit 1 | `go test -count=1 ./internal/envprofile -run '^TestResolveProvisionRepair$'` — exit 0 |
| Disable schema-1 byte-preservation branch | `go test ./cmd/curator -run '^TestEnvResolveKeepsSchema1BytesForMetadataOnly$'` — exit 1 | `go test -count=1 ./cmd/curator -run '^TestEnvResolveRepairEmitsFragment$'` — exit 0 |
| Add `auth.json` path to the ambient keyring record | `go test ./cmd/curator -run '^TestEnvResolveCredentialRecordPathlessKeyring$'` — exit 1 | `go test -count=1 ./internal/envprofile -run '^TestResolveProvisionRepair$'` — exit 0 |

The third pre-feature selector uses Codex file storage; the old source had no v2 pathless-record writer. It is a representative old-behavior survival bound, not proof that the old writer had the new pathless branch.

## CHANGELOG entry (for release prep)

Managed-home credential records now carry schema-2 backend metadata and provenance, with pathless records for linkless strategies.

## Revision 3 (carry-forward republish)

Revision 2 was ACCEPTED on content (reviewer claude-opus-5-5 low; hosted gate run 36235825429 green, rose-air skipped).
Trunk advanced to `e8620502`; `worktree converge STORY-260923-1v3no2` carried the accepted delta uncommitted.
No source file was changed in this revision — verify-only republish plus this results note.

Per-path verification against `TASK-260923-2gt5f6_change-request_rev2.patch` (18 paths):

- 17 non-intersecting paths byte-identical to the rev2 post-image (reconstructed from the patch and compared byte-for-byte):
  `.github/ci/platform-cases.tsv`, `.github/ci/root-artifacts.tsv`, `.github/ci/skip-classes.tsv`,
  `cmd/curator/env_credential_marker_test.go` (new file), `cmd/curator/env_migrate_test.go`,
  `docs/environment-config.md`, `internal/envmarker/envmarker.go`, `internal/envmarker/envmarker_test.go`,
  `internal/envmarker/marker_env_schema_test.go`, `internal/envprofile/credential_link_test.go`,
  `internal/envprofile/credential_record_test.go` (new file), `internal/envprofile/lock.go`,
  `internal/envprofile/managed.go`, `internal/envprofile/managed_test.go`, `internal/envprofile/migrate.go`,
  `internal/envprofile/migrate_test.go`, `internal/envprofile/status.go`.
- 1 intersecting path, both sides present, no conflict markers: `.github/ci/conformance-case-counts.tsv`
  carries the trunk side (skillfile-sources-v1 rows replacing draft-sources-v1, new header from
  commit `5328d488`) AND this task's row `agent-environment-marker-v2/schema-cases 26`.
- `CHANGELOG.md` / `LOGBOOK.md` equal trunk (no hunk in rev2, nothing to revert); no stray root
  `TASK-*`/`BUG-*` files; `git diff --name-only HEAD -- . ':!.task-board'` lists only the 16 tracked
  rev2 paths, plus the 2 rev2 new files as untracked — no trunk revert, worktree is not a stale snapshot.

Focused bounded runs in this revision (real exit codes):

- `go test ./internal/envprofile -run 'Credential|Record|Marker' -count=1` — exit 0 (ok, 28.8s).
- `go test ./cmd/curator -run 'Env|Credential' -count=1` — exit 0 (ok, 253.5s).

All 14 DoD checklist items were already checked from revision 2; none left unchecked.
