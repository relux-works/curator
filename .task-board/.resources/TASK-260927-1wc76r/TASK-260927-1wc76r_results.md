# TASK-260927-1wc76r results

## Implementation

Added `curator env unmanage [--restore-backups] [--env <env-id>]` through the CLI and `internal/envprofile.Unmanage`. The implementation reads native-home markers, plans all selected removals/restores before the first native-home mutation, removes only marker-recorded non-credential surfaces, restores the newest numeric backup generation before deleting the marker, and clears the current-profile records for the selected scope. Without `--restore-backups`, backup generations remain and the CLI prints their path.

The restore behavior follows curator-spec v1.0.0-rc.13 at commit `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`: `protocol/environments.md` §8.3 defines versioned backup sets and newest-generation restore; §8.4.1 distinguishes absent inventory from unreadable state; §9.2 defines unmanage, marker cleanup, current-record clearing, and the credential/managed-home boundaries. An unreadable or inconsistent inventory returns `environment_backup_record_unreadable` with the typed `stateread` failure and stops before native-home writes. A proven absent inventory restores no files.

The CLI accepts the specified `--target` spelling, but secondary fixed-home target writes remain explicitly deferred with the existing runtime boundary: target home resolution and surface writes are not implemented elsewhere in this tree. Such a request is rejected before mutation. The rc.13 unmanage/restore vector inventory has no target-specific case.

## Vector and gap evidence

The rc.13 environment-vector scan found exactly two `operation: restore` cases: `backup-record-unreadable-restore-stops` and `backup-record-absent-restore-nothing`. No environment vector uses `operation: unmanage`; the other name containing “unmanaged” is a takeover conflict case and is outside this task.

`TestEnvUnmanageBackupRecordVectors` drives both published cases through `cli.run` and records **2 driven, 0 known-gap, 0 bound, 0 skipped, 2 total**. The baseline gap-ledger count for these two IDs was **0**; after the change it is **0**. No owned rows were present to delete (the earlier task had treated these two vectors as bounds).

## Mutant checks

Both required mutants were applied temporarily and killed; the source was restored and the final green run was performed afterward.

- Unreadable-as-absent mutant: the unreadable inventory vector unexpectedly returned exit 0, and this focused vector command exited **1**:

  ```sh
  CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^TestEnvUnmanageBackupRecordVectors$' -count=1 -v
  ```

- Restore-without-record mutant: a forced physical re-probe after the injected absent observation restored `operator-owned context`; this test command exited **1**:

  ```sh
  go test ./cmd/curator -run '^TestEnvUnmanageDoesNotRestoreWithoutAnObservedRecord$' -count=1 -v
  ```

## Validation

Focused CLI test command, exit **0** (58.880s):

```sh
CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^Test(EnvUnmanageBackupRecordVectors|EnvUnmanageDoesNotRestoreWithoutAnObservedRecord|EnvUnmanageRestoresNewestBackupGeneration|EnvUnmanageRestoresRecordedTakeoverBackup|EnvUnmanageLeavesBackupsWithoutRestoreFlag)$' -count=1 -v
```

Both published vectors plus absent-record refusal, newest-generation restore, actual takeover-backup restore, and no-flag retention passed.

| Command | Exit | Result |
| --- | ---: | --- |
| `go build -o /tmp/curator-260927-1wc76r ./cmd/curator` | 0 | CLI build succeeded. |
| `golangci-lint run ./cmd/curator ./internal/envprofile` | 0 | 0 issues (both runs). |
| `gofmt -l internal/envprofile/unmanage.go internal/envprofile/switch.go cmd/curator/env.go cmd/curator/main.go cmd/curator/env_unmanage_test.go` | 0 | No files listed. |
| `git diff --check` | 0 | Clean. |
| Gap-ledger count at baseline / current | 0 / 0 | No rows for the two restore IDs. |

Mutation-check commands intentionally exited **1** as described above. One broad `go test ./internal/envprofile -count=1` run was interrupted after roughly six minutes while other Story runs were active on the shared host; it returned **1** and is not passing evidence. Four earlier focused-test attempts also returned **1** while correcting the fixture (first-install activation, use of `env resolve` rather than native takeover, a compile-time redeclaration, and an absence seam that did not cover the new `lstat` observation); those issues were corrected, and the final focused command above exited 0.

Local validation ran on `darwin` (`go env GOOS`). Linux and Windows execution remains unverified here; the platform-case ledger requires those CI lanes to execute the vector test.

## Change and logbook policy

No repository `CHANGELOG.md` or `LOGBOOK.md` was edited. Findings and validation are recorded in this task-scoped result resource, per campaign rules.

## CHANGELOG entry (for release prep)

Added `curator env unmanage --restore-backups` to restore the newest takeover backup generation while returning marker-recorded native surfaces to operator ownership. Unreadable backup inventories fail closed; absent inventories restore nothing.
