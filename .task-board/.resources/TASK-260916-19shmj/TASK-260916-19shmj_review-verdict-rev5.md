# TASK-260916-19shmj — review verdict rev5: ACCEPTED

Reviewer: claude-opus-5-5 (low). Candidate tree 9c4311a5 (the worktree write-tree reproduced 9c4311a5 byte-for-byte before and after the mutants). Reviewed the delta eca2bf27..9c4311a5 (9 paths), as the review note says.

## Findings
- Rule (rc.13 §8.3.1: managed writes never follow a link; §9.5 backups):
  - `managedPath` and `managedDirectory` check every component below the root with Lstat and refuse a link with `environment_write_would_follow_link` (nofollow.go:552-652).
  - `atomicManagedFile` and `atomicManagedLink` stage a same-directory private entry and rename it over the target entry, so the final component is replaced and never opened.
  - `readManagedRegular` uses O_NOFOLLOW, FILE_FLAG_OPEN_REPARSE_POINT and a SameFile check.
- Recorded links keep their behaviour. `ensureCredentialLink` still runs its Lstat classification and the `environment_credential_conflict` refusals before any replacement (managed.go:1299-1390). A recorded link is not backed up. An unrecorded link is backed up as a link only under authorized takeover (managed.go:~1010).
- Windows:
  - `renameManagedEntry` uses FileRenameInfo with ReplaceIfExists.
  - `replaceManagedEntry` handles a refused rename with Lstat-only removal of the ENTRY (symlink, junction/irregular, or empty directory) and then retries the rename. A non-empty directory is not removed.
  - Platform bound: on Windows there is a non-atomic window between the remove and the retried rename.
- Copy fallback is gated on `errSymlinkUnavailable`, which is returned only when os.Symlink of the private entry fails (managed.go:1063).
- The 11 vectors are 10 driven and 1 bound. The bound is `backup-symlinked-target-refused`. I judge the bound acceptable:
  - applyPlan only adds a symlink to the backup set under authorized takeover when the link is not recorded.
  - Unauthorized takeover is refused by the foreign-manager inventory; the vector row `takeover-symlinked-target-unauthorized-stopped` drives this.
  - openBackup copies link text with Readlink and never reads through the link.
  - Residual: the helper itself does not refuse. If a future caller backs up an unauthorized link, that caller needs the refusal.
- Ledger: `environments-write-nofollow/cases 11` was added to conformance-case-counts.tsv, and root-artifacts.tsv now lists the vector. `claudeSeed` was removed from the read-guard allow-list; it now goes through stateread/readManagedRegular. No CHANGELOG, LOGBOOK or stray files.

## Independent runs (zsh, pipefail, real exit codes, CURATOR_CONFORMANCE_ROOT = curator-spec 23435129 conformance/v1 archived to $TMPDIR)
- `go test ./internal/envprofile -run 'Nofollow|Symlink|Credential|Link|Drift|Passthrough|Seed|Repair|ManagerOwnedAbsence'` → exit 0 (154.6 s). Coverage line: "environments-write-nofollow/cases: 10 driven, 0 known-gap, 1 bound, 0 skipped, 11 total".
- `GOOS=windows go vet ./internal/envprofile` → exit 0.
- Hosted gate for rev5 is green on every lane, including Windows (per the orchestrator note and the validation log). Revs 2-4 were red on Windows.

## Mutants (run with -run 'WriteNofollow|StoreSymlink|Symlink')
| mutant | result |
|---|---|
| M1: Lstat→Stat on the managedPath parent walk | exit 1, 3 FAIL, killed |
| M2: atomicManagedFile opens through (os.WriteFile(full)) | exit 1, 5 FAIL, killed |
| M3: copy fallback on any link error (errors.As gate disabled) | exit 0 on darwin with the broad -run too: **survives on unix** |

M3 bound: on unix, rename-over never fails for the vector shapes, so the fallback branch is never reached. The killing evidence is the Windows hosted lane: revs 2-4 failed 22 tests on exactly this path, and rev5 is green. Residual for a follow-up: add a unix row with a planted non-empty real directory at a link path. That row would kill M3 on unix, because M3 copies store bytes into the planted directory where the gated code refuses.

## CHANGELOG entry (from producer, for release prep)
Managed environment writes (materialize, takeover, repair, backups, seeds) no longer follow symlinks: planted links at targets are replaced, and parent-component links refuse with environment_write_would_follow_link (environments rc.13 §8.3.1).
