# TASK-260918-ryh3kw results — revision 6 candidate

## Requirements and implementation

Normative source: curator-spec `protocol/environments.md` §§8.4.1 (absence versus read failure), 8.5 (diagnostics), 1.3 (lock trust), 7.4 (passthrough and seed), and 12 (currency). The conformance root is curator-spec commit `17d887950623db8cacaeae9d985841398c6dfa02`; `23be89e` is its ancestor (exit 0).

| Requirement | Candidate evidence |
| --- | --- |
| Shared absence/read-failure classification | `internal/stateread` is the shared seam. The rev5 Windows fix classifies a missing leaf as absent only when existing ancestors are directories; non-directory ancestors, unreadable paths, malformed files, and other I/O errors remain unreadable. The bounded ancestor walk uses `Lstat` and only follows a symlink to confirm it resolves to a directory. |
| Read-site inventory | `TestManagerOwnedAbsenceReadsAreGuarded` emits a sorted file:line/function verdict for every site. Attached `TASK-260918-ryh3kw_read-site-inventory.tsv` lists 372/372 sites: 266 through `stateread`, 106 reviewed exceptions; scanner bound: 410 production Go files. The merged trunk added three sites relative to rev5; all three are covered. |
| Absence-shaped branch guard | The deny-by-default scanner rejects unreviewed readers and detects aliased `os.IsNotExist` plus aliased `errors.Is(..., fs.ErrNotExist)` branches. `TestManagerReadScannerFindsAliasedNotExistCollapse` is the negative proof. |
| New diagnostics and lock behavior | `environment_passthrough_unreadable` makes status unknown/non-current, resolve refuse, and repair leave the entry untouched. `environment_backup_record_unreadable` is exercised by status and restore. Unreadable or malformed locks yield `environment_store_untrusted`; status, resolve, repair, and update tests confirm they do not rebuild or write from an untrusted lock. |
| Pinned vectors | `TestEnvironmentsReadFailureVectors` drives 37 status/seed/lock/passthrough cases. `TestEnvUnmanageBackupRecordVectors` drives both restore cases through `env unmanage --restore-backups`. Total: 39/39 driven, zero skipped/bound/gap cases. `.github/ci/conformance-gaps.tsv` contains no read-failure or backup-record rows. The earlier temporary attribution to TASK-260927-1wc76r is superseded because its restore entry landed on trunk. |
| Trunk carry | Re-applied accepted rev5 (base `eca2bf27`, tree `b82d529f`) to trunk `6bd98d49`, retaining E3 Codex seed revision-B behavior in the four conflicted files. `git diff --name-only HEAD -- . ':!.task-board'` lists exactly the 17 task paths. No `CHANGELOG.md` or `LOGBOOK.md` edit. |
| Docs and release text | Updated `docs/cli.md` and `docs/troubleshooting.md`. Release-prep text is below. |

## Local validation

Commands were run as standalone processes; each exit is the observed process exit.

- `go test ./internal/stateread -count=1` — exit 0.
- `env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run 'ReadFailure|Guarded|Restore|Unmanage|Seed|Codex|Status' -count=1` — exit 0.
- `env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./internal/envprofile -run '^TestEnvironmentsReadFailureVectors$' -count=1 -v` — exit 0; 37/37 cases.
- `env CURATOR_CONFORMANCE_ROOT=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 go test ./cmd/curator -run '^TestEnvUnmanageBackupRecordVectors$' -count=1 -v` — exit 0; 2/2 restore cases.
- `go test ./cmd/curator -run '^TestEnv(Status|Resolve)' -count=1` — exit 0 (475.680s).
- `go test ./internal/envprofile -run '^(TestManagerOwnedAbsenceReadsAreGuarded|TestManagerReadScannerFindsAliasedNotExistCollapse)$' -count=1` — exit 0.
- `go test ./internal/envprofile -run '^(TestUnreadableLockIsUntrustedAcrossStatusResolveRepairAndUpdate|TestUnreadableLockRefusesUpdateWithoutRewriting|TestPassthroughReadFailuresAreUnknownAndNeverRepaired)$' -count=1` — exit 0.
- `CI_REQUIRE_FULL_ROOT=1 bash .github/ci/suite-plan.sh /Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1 "$TMPDIR/TASK-260918-ryh3kw-suite-plan"` — exit 0; 80 served, 0 deferred, 0 excluded.
- `go build -o "$TMPDIR/TASK-260918-ryh3kw-curator" ./cmd/curator` — exit 0.
- `golangci-lint run` — exit 0, 0 issues. `gofmt -l` on changed Go files — exit 0, no files listed. `git diff --check HEAD` — exit 0.

## Narrowing mutants

Each mutant test exited 1 as expected. Mutated sources were restored from `$TMPDIR` copies and verified with `cmp` (exit 0); the corresponding unmutated tests then passed.

| Rule | Mutant and regression | Command exit |
| --- | --- | ---: |
| Shared classifier | Treat every symlink ancestor as traversable; `TestReadsDistinguishAbsentFromBlockedParent`. | 1 |
| Guard | Stop recognizing aliased `os.IsNotExist`; `TestManagerReadScannerFindsAliasedNotExistCollapse`. | 1 |
| Lock classification | Wrap only absent locks as untrusted; `TestUnreadableLockIsUntrustedAcrossStatusResolveRepairAndUpdate`. | 1 |
| Passthrough | Apply unreadable refusal only when repair is off; `TestPassthroughReadFailuresAreUnknownAndNeverRepaired`. | 1 |
| Unreadable restore record | Return empty inventory on backup-directory read error; restore vector through `TestEnvUnmanageBackupRecordVectors`. | 1 |
| Absent restore record | Treat absence as unreadable; restore vector through `TestEnvUnmanageBackupRecordVectors`. | 1 |

## CHANGELOG entry (for release prep)

- Distinguish absent environment state from unreadable or malformed state across profile locks, provisioning seeds, passthrough links, and backup inventories; refuse operations that would otherwise fall back to absence or rebuild from untrusted lock data.
- Report unreadable credential passthrough entries as `environment_passthrough_unreadable`, with unknown currency and no repair relink.

## Hosted validation

Local checks do not establish hosted status. The published Change Request validation log is the arbiter; no hosted result for this revision is claimed here.
