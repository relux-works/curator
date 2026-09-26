# TASK-260922-cww1ov results — F-C3 0017 production-entry tests

Producer: developer. Worktree: `.temp/STORY-260922-1cenbr/worktree`
(uncommitted). Base: F-C1 + F-C2 checkpointed on the Story branch. This
leaf makes the coverage COMPLETE and CI-registered; it does not
re-implement. Zero production lines changed.

## Outcome

46 production-entry rows on temporary stores drive `Resolve`,
`StatusOf`, `PlanMigration`, `ApplyMigration`, and `curator env
resolve|migrate` across both 0017 hazards, the mis-targeted vs
dangling-to-declared split, the migration end-to-end, every repair
refusal class with the codex admission table, and no silent repair.
Three rows are new (the `environment_repair_failed` class at both
entries, and apply-under-lock); one existing row gained a negative
assertion pinning the recorded/foreign refusal split (named below, no
weakening); 36 existing rows gained the Windows symlink privilege
probe. Every row is registered in
`.github/ci/platform-cases.tsv` (46 new rows, one section, no new skip
class). Every row carries a narrowing mutant, all killed with green
witnesses. `go test ./internal/envprofile/` passes; ledger-consistency
and the platform-case gate (synthetic Windows stream + negative
control) pass.

## Coverage table

Entry key: R=`Resolve`, S=`StatusOf`, P=`PlanMigration`,
A=`ApplyMigration`, C=`curator env …` via `run()`. Lanes: ALL =
required linux,darwin,windows; probe rows tolerate the
host-capability symlink skip on windows only; ENOTDIR rows are
unix-only with the windows skip tolerated. Every refusal row asserts
the exact diagnostic: conflict admitted (or re-classed) means the
test fails.

### 1. The two 0017 hazards

| Hazard | Row | Entry | Mutant | Lanes |
|---|---|---|---|---|
| (a) stale link surviving shared→isolated | TestSharedToIsolatedRemovesStaleLink | R+P+A | C1 repair silently unlinks declared-target stale links | ALL+probe |
| (b) regular file at a link path | TestCredentialLinkRegularFileRefuses (non-empty+empty) | R | C2 refuse only non-empty files | ALL+probe |
| (b) stale-side file/link/dir | TestStaleCredentialLinkRefusals | R | C3 basename-match foreign stale links unlinked | ALL+probe |
| (b) migration side | TestMigrateConflictsRefuse | P+A | C19 empty-file conflict admitted as absent | ALL+probe |
| (b) migration side, CLI | TestEnvMigrateConflictRefuses | C | C41 regular-file conflicts exit 0 | ALL+probe |

### 2. Dangling native target, both states

| State | Row | Entry | Mutant | Lanes |
|---|---|---|---|---|
| mis-targeted → conflict, migration relinks | TestDanglingPiLinkReportedDetached | S+R | C4 basename-only wanted-target compare | ALL+probe |
| mis-targeted → conflict, CLI | TestEnvResolveRepairNeedsMigration | C | C42 migration-needed refusals exit 0 | ALL+probe |
| dangling-to-declared → pending, provision succeeds | TestPiProvisionTargetsAgentRoot | R | C7 pending collapses into stale | ALL+probe |
| dangling-to-declared → pending repair succeeds | TestDanglingExpectedTargetRepairSucceeds | R | C12 pending collapses into stale (repair half) | ALL+probe |
| dangling-to-declared → never silent | TestReviewerDanglingExpectedTarget | S+R | C10 status drops the pending finding | ALL+probe |
| inspection failure, stale not absence | TestCredentialLinkTargetInspectionFailure | R+S | C13 inspection failure downgraded to warning | unix-only |
| directory rule §7.4 | TestCredentialLinkDirectory | R | C8 single-entry dirs treated as empty | ALL+probe |
| recorded-only bound | TestCredentialLinkUnrecordedSymlinkRefuses | R | C9 unrecorded links get migration-needed refusal | ALL+probe |

### 3. Migration end-to-end

| Aspect | Row | Entry | Mutant | Lanes |
|---|---|---|---|---|
| inspect→plan→apply→converge | TestMigratePiWrongTargetToAgentRoot | P+A+R | C17 dangling-From mis-targeted plans no relink | ALL+probe |
| inspect→plan→apply, CLI | TestEnvMigratePlanApplyPi | C | C40 apply prints the plan only under a scope filter | ALL+probe |
| plan required (lib + CLI) | TestMigrateApplyRequiresPlan, TestEnvMigrateApplyRequiresPlan | A, C | C27 plan required only under a scope filter; C44 hash required only under a filter | ALL+probe |
| plan printed before mutation | TestMigratePrintFailureRefusesBeforeMutation | A | C28 print failure swallowed on single-op plans; C45 prints after mutation | ALL+probe |
| print before mutation, temporal | TestEnvMigratePrintBeforeWrite | C | C45b prints after mutation (hook kills; containment cannot) | ALL+probe |
| apply under the lock (NEW) | TestMigrateApplyLockContention | A | C39 deadline lock errors lose the lock diagnostic | ALL+probe |
| drift: links | TestMigratePlanDriftRefuses | P+A | C21 prefix hash accepted as the plan | ALL+probe |
| drift: marker digest | TestMigrateMarkerDriftRefuses | P+A | C29 pi marker identity dropped from the hash | ALL+probe |
| journal: syscall rollback | TestMigrateSyscallFailureRollsBack | A | C30 rename failures swallowed | ALL+probe |
| journal: rollback both phases | TestMigrateRollbackRestoresPriorState | A | C22 marker restore skipped when links moved | ALL+probe |
| journal: publish failure | TestMigratePublishFailureRevertsLinks | A | C32 rollback restores relinks but not unlinks | ALL+probe |
| recovery: interrupted recovers | TestMigrateInterruptedApplyRecovers | P+A | C31 only sealed journals recover | ALL+probe |
| recovery: sealed deletes | TestMigrateSealedJournalDeletesWithoutTouching | P+A | C33 sealed all-done journals roll back | ALL+probe |
| recovery: broken refuses | TestMigrateBrokenJournalRefuses | P+A | C34 apply ignores a broken journal | ALL+probe |
| recovery: unknown states refuse | TestMigrateRecoveryRefusesUnexpectedTarget | A | C35 recovery validates Done ops only | ALL+probe |
| recovery: marker drift | TestReviewerRecoveryMarkerDrift | A | C37 recovery republishes markers without validating | ALL+probe |
| recovery: owned temp cleaned | TestMigrateRecoveryCleansOwnedTemp | A | C36 owned temps cleaned only when target absent | ALL+probe |
| recovery: foreign temp preserved | TestReviewerRecoveryPreservesRegularTemp | A | C38 glob deletes regular temp-shaped files | ALL+probe |
| no secret bytes copied | TestMigrateNoSecretCopies | P+A+R | C18 relink leaves a bytes backup when To exists | ALL+probe |
| old-root bytes warn, proceed | TestMigrateOldRootBytesWarnAndProceed | P+A+R | C23 old-root bytes block when new root absent | ALL+probe |
| pending plans nothing | TestMigratePendingNeedsNothing | P+A | C24 recorded pending entries plan a relink | ALL+probe |
| inspect/plan read-only | TestMigrateInspectIsReadOnly | P | C26 planning ensures a default profile | ALL+probe |
| Pi roots uninspectable | TestMigratePiRootInspectionFailure | P+A | C20 uninspectable Pi root reads as absent | unix-only |
| scope gates | TestMigrateUnknownOperands | P+A | C25 unknown env admitted on initial c | ALL |
| CLI usage gates | TestEnvMigrateUsage | C | C43 --expect rejected only alongside --inspect | ALL |

### 4. Every repair refusal class + codex admission table

| Class | Row | Entry | Mutant | Lanes |
|---|---|---|---|---|
| conflict | (rows above: 1b, 2-mis-targeted, 3-conflicts) | R/A/C | C1–C4, C9, C19, C41 | — |
| isolated_unsupported | TestCodexIsolatedAdmission | R | C5 admit auto, keep keyring refusal | ALL |
| isolated_unsupported, literals | TestReviewerLiteralCodexSelector | R | C11 single-quoted keyring/auto read as absent | ALL |
| admission spellings | TestCodexCredentialStoreTOMLSpellings | R | C14 nested-table key counts as selector | ALL |
| credential_unsupported, shared | TestCodexUnknownStoreSharedRefuses | R | C6 unknown passes through (isolated gate still refuses) | ALL+probe |
| credential_unsupported, malformed | TestCodexMalformedStoreSharedRefuses | R | C15 invalid TOML defaults to file | ALL+probe |
| repair_failed (NEW) | TestCredentialLinkInspectionRepairFails | R | C16 inspection failures return bare conflict | unix-only |
| repair_failed, CLI (NEW) | TestEnvResolveRepairFailedOnUninspectable | C | C46 repair_failed exits 0 | unix-only |

### 5. No silent repair

Pinned by the refusal halves of TestSharedToIsolatedRemovesStaleLink
(C1), TestDanglingPiLinkReportedDetached (C4), and
TestEnvResolveRepairNeedsMigration (C42): every mutant that migrates
inside repair (unlink, re-point, or swallowed refusal) fails the row.

## Mutant table

Driver: `/tmp/cww1ov_mutants.py` (scratch, not committed). One mutant
at a time, tree restored after each (`git status` shows only the
intended files; `go build ./...` green). No build-failure kills: the
driver reports those INVALID, never kills. 47/47 killed, 47/47
witnesses green.

| ID | Production site | Type | Killer (FAIL) | Witness (pass) |
|---|---|---|---|---|
| C1 | removeStaleCredentialLinks (managed.go) | conditional admit | SharedToIsolated | StaleRefusals/regular_file |
| C2 | ensureCredentialLink (managed.go) | narrowing | RegularFile/empty | RegularFile/non-empty |
| C3 | removeStaleCredentialLinks (managed.go) | conditional admit | StaleRefusals/foreign_symlink | StaleRefusals/file+dir |
| C4 | checkPassthrough (managed.go) | narrowing | DanglingPi | DanglingRepairSucceeds |
| C5 | checkIsolatedCredentialStore (managed.go) | arm drop | IsolatedAdmission/auto | IsolatedAdmission/keyring |
| C6 | codexCredentialStore (managed.go) | partition | UnknownShared | LiteralSelector/ephemeral |
| C7 | checkPassthrough (managed.go) | collapse | PiProvision (+RepairSucceeds, ReviewerDangling co-kill) | DanglingPi |
| C8 | ensureCredentialLink (managed.go) | narrowing | Directory/non-empty | Directory/empty |
| C9 | ensureCredentialLink (managed.go) | collapse | UnrecordedSymlink | DanglingPi |
| C10 | StatusOf (status.go) | narrowing | ReviewerDangling | DanglingRepairSucceeds |
| C11 | codexCredentialStore (managed.go) | conditional | LiteralSelector/keyring+auto (+TOML multi-line auto co-kill) | LiteralSelector/ephemeral |
| C12 | checkPassthrough (managed.go) | collapse | DanglingRepairSucceeds | DanglingPi |
| C13 | checkPassthrough (managed.go) | narrowing | InspectionFailure | DanglingRepairSucceeds |
| C14 | codexCredentialStore (managed.go) | narrowing | TOMLSpellings/nested+dotted | TOMLSpellings/single-quoted |
| C15 | codexCredentialStore (managed.go) | narrowing | MalformedShared/invalid | MalformedShared/integer |
| C16 | repairUnderLock (managed.go) | carve-out | InspectionRepairFails | RegularFile/non-empty |
| C17 | classifyLink (migrate.go) | conditional | MigratePi | OldRootBytes |
| C18 | executeMigrationOps (migrate.go) | conditional | NoSecretCopies | OldRootBytes |
| C19 | classifyPath (migrate.go) | conditional admit | Conflicts/empty | Conflicts/non-empty |
| C20 | statRoot (migrate.go) | collapse | PiRootInspection | OldRootBytes |
| C21 | ApplyMigration drift check (migrate.go) | narrowing | PlanDrift | MarkerDrift |
| C22 | rollbackMigrationMarkersAndLinks (migrate.go) | conditional | Rollback/post_publish | Rollback/link_phase |
| C23 | inventoryPiRoots (migrate.go) | conditional flip | OldRootBytes | MigratePi |
| C24 | classifyLink (migrate.go) | conditional | PendingNeedsNothing | MigratePi |
| C25 | migrationAdapters (migrate.go) | conditional admit | UnknownOperands | InspectIsReadOnly |
| C26 | PlanMigration (migrate.go) | bound injection | InspectIsReadOnly | UnknownOperands |
| C27 | ApplyMigration plan gate (migrate.go) | conditional | ApplyRequiresPlan | MigratePi |
| C28 | ApplyMigration print gate (migrate.go) | conditional | PrintFailure | MigratePi |
| C29 | migrationHash (migrate.go) | narrowing | MarkerDrift | ReviewerMarkerDrift |
| C30 | atomicRelinkJournaled (migrate.go) | conditional swallow | Syscall/rename | Syscall/symlink |
| C31 | recoverMigrationJournal (migrate.go) | conditional | InterruptedRecovers | SealedJournal |
| C32 | reconcileJournalLinks (migrate.go) | conditional | PublishFailure | Syscall/symlink |
| C33 | recoverMigrationJournal sealed path (migrate.go) | conditional | SealedJournal | InterruptedRecovers |
| C34 | recoverMigrationJournal (migrate.go) | entry-conditional | BrokenJournal | SealedJournal |
| C35 | validateRecoveryInventory (migrate.go) | narrowing | RecoveryUnexpectedTarget | InterruptedRecovers |
| C36 | reconcileJournalLinks (migrate.go) | conditional | RecoveryCleansOwnedTemp | ReviewerPreservesTemp |
| C37 | validateRecoveryInventory + re-check (migrate.go) | stage-conditional | ReviewerMarkerDrift | RecoveryUnexpectedTarget |
| C38 | reconcileJournalLinks (migrate.go) | conditional | ReviewerPreservesTemp | RecoveryCleansOwnedTemp |
| C39 | beginOperation (lock.go) | narrowing | ApplyLockContention (+ResolveLockContention co-kill) | MigratePi |
| C40 | cmdEnvMigrate (envmigrate.go) | conditional | PlanApplyPi (+PrintBeforeWrite co-kill) | MigrateUsage |
| C41 | cmdEnvMigrate (envmigrate.go) | conditional | MigrateConflictRefuses | RepairNeedsMigration |
| C42 | cmdEnvResolve (env.go) | conditional | RepairNeedsMigration | MigrateConflictRefuses |
| C43 | cmdEnvMigrate (envmigrate.go) | narrowing | MigrateUsage | PlanApplyPi |
| C44 | cmdEnvMigrate (envmigrate.go) | conditional | ApplyRequiresPlan (+Usage co-kill) | PlanApplyPi |
| C45 | ApplyMigration print order (migrate.go) | conditional reorder | PrintFailure | InterruptedRecovers |
| C45b | same patch, CLI entry | conditional reorder | PrintBeforeWrite | MigrateUsage |
| C46 | cmdEnvResolve (env.go) | conditional | RepairFailedOnUninspectable | RepairNeedsMigration |

Disclosed iterations (mutant bugs, not test bugs): C14 first shape
shadowed `ok` (no-op, survived once, fixed); C30 first shape
(reconcile skip) was unobservable behind the best-effort temp cleanup,
replaced by the rename-swallow; C37 first shapes skipped the marker
republish along with validation (drift survived to the drift check),
corrected to skip validation but keep republish; C43 first shape fell
through to the next usage gate with the same exit code, retargeted to
the --plan --expect case.

## Ledger delta

`.github/ci/platform-cases.tsv`: +46 rows in one new section
("Decision 0017 … (F-C1/F-C2/F-C3)"). 39× internal/envprofile, 7×
cmd/curator. Shapes: 39 must_all+skip_windows/host-capability
(link-capability probe), 4 must_all with no tolerance
(double-quoted/single-quoted/TOML admission, unknown operands, CLI
usage), 3 unix-only must_linux,darwin+skip_windows/host-capability
(ENOTDIR inspection). No `.tsv` vocabulary change, no new skip class,
no `Parent/*` rows (all skips are parent-level probes).

## Test strengthening (one, named)

TestCredentialLinkUnrecordedSymlinkRefuses gained a negative assertion:
an unrecorded (foreign) refusal must NOT say "migration needed" (only
the recorded mis-targeted refusal points at migration). No other
existing assertion was touched; the probe insertions (31 lib + 5 CLI
first-line calls) are no-ops on capable hosts and gate-fatal skips on
unix if they ever fire.

## Bounds (not owned here)

- No product change: the suite found no defect; production is
  byte-identical.
- TestMigrateFailedRelinkKeepsOldLink stays helper-direct by
  construction (real-OS oversized-target rejection, unplannable
...[truncated 1899 chars]
## Revision 8 (carry-forward republish)

Trunk moved to `1511b345`; the orchestrator ran `worktree converge
STORY-260922-1cenbr`. Content accepted at revision 5 is unchanged; this
revision only re-verifies the converged tree and republishes evidence.
Base of revision 5: `48da2690`. Revision 5 patch: 22 paths in
`TASK-260922-cww1ov_change-request_rev5.patch` (applies cleanly onto
`48da2690`, dry-run exit 0).

Trunk `48da2690..1511b345` (non-board paths) touched:
`.github/ci/gate-selftest.sh`, `.github/ci/platform-cases.tsv`,
`.github/ci/platform-exclusions.tsv`, `.github/workflows/ci.yml`,
`CHANGELOG.md`, `internal/gitops/*`, `tools/goreleaserconfig/*`.
Intersecting with the 22 revision-5 paths: exactly
`CHANGELOG.md` and `.github/ci/platform-cases.tsv`.

Per-path verification (worktree file vs revision-5 reconstruction
`48da2690` + rev5 patch, sha256-12):

- Byte-identical (20/20, all non-trunk-touched paths):
  `cmd/curator/env.go`, `cmd/curator/env_migrate_test.go`,
  `cmd/curator/env_test.go`, `cmd/curator/envalias_test.go`,
  `cmd/curator/envmigrate.go`, `cmd/curator/hook_posture_test.go`,
  `cmd/curator/hook_test.go`, `cmd/curator/profile_test.go`,
  `docs/cli.md`, `docs/troubleshooting.md`,
  `internal/envprofile/credential_link_test.go`,
  `internal/envprofile/credential_production_test.go`,
  `internal/envprofile/managed.go`, `internal/envprofile/managed_test.go`,
  `internal/envprofile/migrate.go`, `internal/envprofile/migrate_test.go`,
  `internal/envprofile/reviewer_recovery_test.go`,
  `internal/envprofile/status.go`, `internal/envprofile/status_test.go`,
  `internal/envregistry/envregistry.go`.
- Intersecting, both sides present, no markers, no duplication:
  - `.github/ci/platform-cases.tsv`: all 46 revision-5 F-C3 rows
    present (e.g. `TestSharedToIsolatedRemovesStaleLink`,
    `TestMigrateInterruptedApplyRecovers` ×1 each, F-C3 section
    header ×1) AND all 4 trunk gitops fold rows present ×1 each
    (`TestExtractRefusesDirectoryComponentFold`,
    `TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive`,
    `TestExtractRefusesFileDirectoryFold`,
    `TestExtractRefusesNestedDirectoryComponentFold`); zero duplicated
    data rows; zero conflict markers.
  - `CHANGELOG.md`: all three revision-5 F-C3 blocks present
    ("Production-entry test suite", "Explicit credential migration",
    codex/pi repair blocks) AND the trunk dirfold block present
    ("folds directory components" ×1). One converge drop found and
    repaired: the trunk GoReleaser rc-channel block
    ("CI guard for the GoReleaser rc channel values",
    TASK-260908-2kqa77) was absent from the converged file while
    present at `1511b345:CHANGELOG.md:72`; restored byte-identical in
    trunk position (after "(Manager profile §8).", before
    "- E2: direct-only"), verified by 10-line sequence match against
    trunk. Zero conflict markers. Nothing else changed: `git status`
    still shows the same 7 modified + 1 untracked set
    (`platform-cases.tsv`, `CHANGELOG.md`,
    `cmd/curator/env_migrate_test.go`, `docs/troubleshooting.md`,
    `credential_link_test.go`, `migrate_test.go`,
    `reviewer_recovery_test.go`, + untracked
    `credential_production_test.go`).

Focused bounded run (direct process, no pipe; `set -o pipefail`
equivalent — real exit code preserved):

- `go test -count=1 -timeout 9m -run
  'Credential|Isolat|Passthrough|Migrat' ./internal/envprofile/` →
  `ok ... 32.731s`, exit 0.

Checklist: all 13 DoD items were already `done` from the accepted
revision; no unchecked item remains, so revision 8 cites the list as-is
with no new checks needed.


## Revision 8 re-verification (carry-8 rerun, trunk still `1511b345`)

The task returned to `development` after the prior revision-8 handoff
(a successor run reported `stale-anchor` on CR construction — an
orchestrator-level provenance issue, not a content issue). This run
re-verified the converged tree without changing any file
(`git status` unchanged: same 7 modified + 1 untracked;
`internal/envprofile/credential_production_test.go` still untracked).

- Per-path re-verification against
  `TASK-260922-cww1ov_change-request_rev5.patch` (22 paths) and the
  accepted candidate tree `768bacfa2c52a0906b80e233e8af14a275137ede`,
  via `git rev-parse`/`git hash-object` per path:
  - 20/20 trunk-untouched paths byte-identical to revision 5
    (worktree blob == candidate-tree blob == rev5 post-image):
    `cmd/curator/env.go`, `env_migrate_test.go` (new),
    `env_test.go`, `envalias_test.go`, `envmigrate.go` (new),
    `hook_posture_test.go`, `hook_test.go`, `profile_test.go`,
    `docs/cli.md`, `docs/troubleshooting.md`,
    `internal/envprofile/credential_link_test.go` (new),
    `credential_production_test.go` (new, untracked),
    `managed.go`, `managed_test.go`, `migrate.go` (new),
    `migrate_test.go` (new), `reviewer_recovery_test.go` (new),
    `status.go`, `status_test.go`,
    `internal/envregistry/envregistry.go`.
    Trunk-untouched proven by rev5 pre-image blob == `1511b345` blob
    (or path absent from trunk for the 7 new files).
  - 5 of those 20 differ from story HEAD (sibling F-C1/F-C2 content);
    the HEAD→worktree diff is purely additive (insertions only, plus
    4 comment/roots lines of cww1ov's own scan-scope strengthening in
    `migrate_test.go`), so no sibling content was dropped.
  - 2/2 intersecting paths carry both sides, verified line-by-line
    (every non-blank added line from each side present, zero missing):
    `.github/ci/platform-cases.tsv` (61/61 rev5 lines incl. the F-C3
    section; 4/4 trunk gitops-fold rows) and `CHANGELOG.md` (75/75
    rev5 lines incl. all F-C3 blocks; 20/20 trunk lines incl. both
    the dirfold block and the restored GoReleaser rc-channel block,
    each ×1). Zero conflict markers in all 8 worktree files; zero
    duplicated data rows in the ledger.
- Focused bounded run, fresh in this turn (direct process, real exit
  code): `go test -count=1 -timeout 9m -run
  'Credential|Isolat|Passthrough|Migrat' ./internal/envprofile/` →
  `ok ... 22.662s`, exit 0. (An earlier same-turn run without
  `-count=1` also passed, `ok ... 23.141s`, exit 0.)
- Checklist: all 13 DoD items remain `done`; no unchecked item, so no
  new checks — revision 8 cites the list as-is.

## Revision 8 (carry-forward republish) — rerun re-verification, trunk `1511b345`

Republish run after the return to `development` (successor CR
construction reported `stale-anchor`, an orchestrator-level provenance
issue, not a content issue). No worktree file changed in this run
(`git status` before/after: same 7 modified + 1 untracked;
`internal/envprofile/credential_production_test.go` still untracked).
Per-path verification against
`TASK-260922-cww1ov_change-request_rev5.patch` (22 paths):

- 20/20 trunk-untouched paths byte-identical to revision 5
  (`git hash-object` worktree == rev5 post-image blob; trunk-untouched
  proven by rev5 pre-image blob == `1511b345` blob, or path absent from
  trunk for the 7 new files): every `cmd/curator` path, `docs/cli.md`,
  `docs/troubleshooting.md`, all `internal/envprofile` paths except the
  two intersecting ones below, `internal/envregistry/envregistry.go`.
- 2/2 intersecting paths carry both sides, verified line-by-line with
  zero missing lines, zero conflict markers (`^<<<<<<< `/`^>>>>>>> `
  absent), zero duplicated data rows:
  - `.github/ci/platform-cases.tsv`: 62/62 rev5-added lines and 4/4
    trunk gitops-fold rows present. Independent
    `git merge-file trunk base rev5` exits 0 (clean); worktree equals
    that clean merge plus one blank separator line at the rev5/trunk
    block junction (line 530) — cosmetic whitespace only.
  - `CHANGELOG.md`: 75/75 rev5-added lines and 20/20 trunk-added lines
    present (incl. the GoReleaser rc-channel block, exactly ×1).
    Independent `git merge-file` exits 0 (clean); worktree equals that
    clean merge plus one blank separator line (line 22).
- Focused bounded run, fresh this turn as a direct standalone process
  (no pipe; real exit code): `go test -count=1 -timeout 9m -run
  'Credential|Isolat|Passthrough|Migrat' ./internal/envprofile/` →
  `ok ... 35.260s`, exit 0, no `--- FAIL` lines.
- Checklist: all 13 DoD items remain `done`; no unchecked item, so
  revision 8 cites the list as-is with no new checks.

## Revision 8 — refresh onto 1511b345 (2026-09-24)

The requested trunk delta was combined into the preserved, uncommitted candidate, excluding `.task-board/**`. `git diff fad88136 origin/main -- . ':!.task-board' | git apply --3way` returned exit 1 at the already-edited `CHANGELOG.md`; the follow-up `git diff fad88136 origin/main -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` returned exit 0. The final changelog contains both the F-C3 0017 entry and trunk's GoReleaser guard entry. The incoming non-changelog files are `.github/ci/gate-selftest.sh`, `.github/workflows/ci.yml`, and all nine files under `tools/goreleaserconfig/` (including tests and fixtures). The incoming diff changed no `internal/envprofile` reader, so there was no additional reader migration to make. `git restore --staged :/` returned exit 0; the candidate remains uncommitted and unstaged.

Refresh did not proceed. `task-board worktree refresh-candidate TASK-260922-cww1ov` returned exit 1: `candidate refresh requires a rework revision; TASK-260922-cww1ov is ready`. `task-board worktree status STORY-260922-1cenbr` identified revision 7 as `ready`. Retiring it from this producer run is explicitly prohibited: `task-board m 'withdraw_cr(TASK-260922-cww1ov, revision=7, reason=...)'` returned exit 1 with `change_request_withdraw_unauthorized`; only the orchestrator may retire an unintegrated revision governing its own route, or the revision must be answered through the review cycle. The exclusive lease remains held by this live producer. No branch/checkpoint replay or post-refresh verification was possible.

Commands run directly against the combined but not refreshed candidate:

- `go test -count=1 -timeout 9m ./internal/envprofile/...` — exit 1 after 540.821s. It timed out in pre-existing `TestSCPOverlayResolvesAsGit`, subtest `git@example.com:personal`, while `git clone` waited on the SSH remote. This is a suite failure; it is not reported as passing.
- `go test -count=1 -timeout 9m -run 'Credential|Isolat|Passthrough|Migrat' ./internal/envprofile/` — exit 0 (`ok`, 51.923s). This reran the production credential repair/refusal, isolation, migration/recovery, and no-copy rows.
- `go test -count=1 -timeout 9m -run 'TestEnvMigrate|TestEnvResolve' ./cmd/curator` — exit 0 (`ok`, 82.018s), covering the CLI entry points.
- `go vet ./internal/envprofile/... ./cmd/curator ./tools/goreleaserconfig` — exit 0.
- `go build -o /tmp/curator-task-cww1ov ./cmd/curator` — exit 0.
- `gofmt -l` on all changed Go files, including `tools/goreleaserconfig/` — exit 0, no filenames printed.
- `go test -count=1 ./tools/goreleaserconfig` — exit 0 (`ok`, 0.708s).
- `bash .github/ci/gate-selftest.sh` — exit 0, 198 passed and 0 failed.
- `git diff --check` — exit 0.

These are local checks on the old checkpoint with the incoming trunk delta applied. Hosted lanes were not rerun after refresh because the board did not permit the refresh. The existing attached hosted-lane evidence predates this attempt.

**External route needed:** the orchestrator must route stale ready revision 7 through its authorized withdrawal or the review cycle and return the task to a refresh-eligible state. Then a live producer can retry `worktree refresh-candidate`, follow the replay-resolution template if needed, rerun post-refresh verification, and publish the refreshed results/handoff.

Board state after the above verification: `task-board m 'set_status(TASK-260922-cww1ov, status=blocked)'` exited 0. The board moved this task from `development` to `blocked` and demoted `STORY-260922-1cenbr` from `development` to `integrating`. No refreshed CR was published and the developer handoff command was not run because the authorized refresh/review route is still outstanding.


## Revision 9 — refresh onto `ab34556e` (2026-09-26)

### Refresh and candidate

- Required `git stash create` ran before the refresh. Recorded stash object: `b23a8a324d0abd4e0f6d235e8af1987a2abc506e`; it was not applied or dropped.
- The prescribed `git diff 1511b345 ab34556e -- . ':!.task-board' ':!CHANGELOG.md' | git apply --3way` exited 1 at overlapping `.github/ci/platform-cases.tsv` and `docs/troubleshooting.md`. The retry excluded those overlapping files as well as `CHANGELOG.md` and exited 0; the index was reset with exit 0 so the candidate remained unstaged. The ledger was merged against both sides. The troubleshooting conflict retained the environment credential guidance and trunk's script-worker diagnostics.
- First `task-board worktree refresh-candidate TASK-260922-cww1ov` exited 1 because replay of checkpoint `584627497b24f9c50b341b5c20e67ba72ea94f87` required an explicit `docs/troubleshooting.md` resolution. I filled the generated `resolution-template.json` and retried with `--replay-resolutions`; exit 0, outcome `refresh_advanced`, trunk `ab34556e`, new branch tip `18d9c6aa70716c40532c88a0f8566f6e84fc04cc`. The resolution SHA-256 was `bd18e33ba946aa152511bfadcffa90f91a217d0352ee4ae24b7d88811ba71bd6`; no replay worktree was hand-edited or committed.
- `CHANGELOG.md` was restored byte-for-byte from `ab34556e`; `cmp` exited 0. The `.task-board` checkout artifact and `.github/ci/skip-classes.tsv` have no candidate diff. No root `TASK-*`/`BUG-*`, `test/`, or `ledger/` paths remain. No files are staged.

### Trunk state-reader migrations and named regression

The trunk refresh introduced state readers that treated failed reads as absence. A new `internal/stateread` seam preserves typed present/absent/unreadable results and classifies a missing leaf as absent only while its ancestor chain is traversable. It also handles the Windows path-not-found result for a child beneath a regular file. Migrated the trunk readers at these production call sites:

- `scriptworker.LoadShimSidecar` and `scriptworker.ShimSidecarFor`, exercised from `RunShim` and curator executable dispatch;
- `runtimestore.ManagedEnforcedShimsIn`, so a failed inventory listing cannot become an empty inventory;
- `install.lockedNetworkRepository`, so an unreadable locked checkout cannot fall through to a later candidate;
- `cmd/curator.dispatchEnforcedShim`, so an unreadable sidecar cannot select ordinary CLI dispatch.

Named regression: `TestReadsDistinguishAbsentFromBlockedParent` exercises all four filesystem read operations and feeds the Windows path-not-found shape through the production classifier on every lane. `TestLoadShimSidecarRefusesUnreadablePath`, `TestEnforcedShimDispatchRefusesUnreadableSidecar`, `TestManagedEnforcedShimsInRefusesUnreadableInventory`, and `TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts` drive the remaining production boundaries. Each is registered for Linux, macOS, and Windows; no skip class was added.

The refreshed ledger exposed a stale trunk row for deleted `TestARefusalPrecedesEveryWorkerSurface`. Replaced it with the live `TestEnforcedShapesProceedPastAdmission` row and current admission description. This records R5's actual behavior and fixes the ledger-consistency failure without changing script-policy behavior.

### Narrowing mutants for refreshed readers

| Production test | Narrowed gate | Result |
|---|---|---|
| `TestReadsDistinguishAbsentFromBlockedParent` | Return absent for every `os.IsNotExist` result without checking ancestors | Killed; `go test -count=1 ./internal/stateread -run '^TestReadsDistinguishAbsentFromBlockedParent$'` exited 1 as expected |
| `TestEnforcedShimDispatchRefusesUnreadableSidecar` | Convert a `ShimSidecarFor` error into ordinary CLI dispatch | Killed; `go test -count=1 ./cmd/curator -run '^TestEnforcedShimDispatchRefusesUnreadableSidecar$'` exited 1 as expected |
| `TestLoadShimSidecarRefusesUnreadablePath` | Convert an unreadable sidecar to the no-sidecar result | Killed; `go test -count=1 ./internal/scriptworker -run '^TestLoadShimSidecarRefusesUnreadablePath$'` exited 1 as expected |
| `TestManagedEnforcedShimsInRefusesUnreadableInventory` | Return an empty inventory with no error after a failed `ReadDir` | Killed; `go test -count=1 ./internal/runtimestore -run '^TestManagedEnforcedShimsInRefusesUnreadableInventory$'` exited 1 as expected |
| `TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts` | Continue to the next locked repository after failed `Lstat` | Killed; `go test -count=1 ./internal/install -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$'` exited 1 as expected |

All mutants were removed and the production files restored from saved copies or exact inverse patches. `rg 'MUTANT:'` finds no mutation residue. The existing F-C3 coverage and narrowing-mutant tables remain in this resource; this revision reran the 0017 behavior rows and dedicated no-copy scan.

### Validation record

| Command | Exit | Result |
|---|---:|---|
| Initial focused reader command over `internal/stateread`, `internal/scriptworker`, `internal/runtimestore`, `internal/install`, and `cmd/curator` | 1 | Initial typed-error assertions and two Git fixture directories were incorrect; fixed before rerun. |
| `go test -count=1 -timeout 9m -run 'Credential|Isolat|Passthrough|Migrat' ./internal/envprofile` | 0 | 26.510s; F-C3 production credential repair/refusal, isolation, migration, and recovery selection. |
| `go test -count=1 -timeout 7m -run '^TestMigrateNoSecretCopies$' ./internal/envprofile` | 0 | 3.303s; dedicated no-copy scan. |
| `go test -count=1 -timeout 9m ./internal/envprofile` | 0 | 524.389s. `go list ./internal/envprofile/...` lists only this package, so the `...` acceptance scope is the same package. |
| Focused combined reader, env CLI, and dispatcher selection over the five packages | 0 | 112.506s; all new reader rows and env migration/repair CLI rows passed. |
| `go test -count=1 -timeout 9m ./cmd/curator` | 1 | Timed out at 540.916s in `TestDraftTransportProviderAdmission/unknown_ssh_provider_refuses_without_fetching`, while the Go toolchain fingerprint path was reading toolchain files. This is reported as a failure; task-related env CLI and dispatch tests passed in the focused run. |
| `go vet ./internal/envprofile ./cmd/curator ./internal/install ./internal/runtimestore ./internal/scriptworker ./internal/stateread` | 0 | Touched packages clean. |
| `env GOOS=windows go vet ./internal/envprofile` | 0 | Windows compile/vet path clean. |
| `golangci-lint run ./internal/envprofile ./cmd/curator ./internal/install ./internal/runtimestore ./internal/scriptworker ./internal/stateread` | 1, then 0 | First run found two missing exported-constant comments in `stateread`; comments added, rerun reported `0 issues`. |
| `bash .github/ci/ledger-consistency.sh /tmp/TASK-260922-cww1ov-ledger-consistency-r3` | 0 | Final ledger: 296 rows checked across Linux, macOS, Windows. An earlier run found the stale script-policy row (exit 1); after correction, both 297-row and final 296-row checks passed. |
| `git diff --check` | 0 | Clean. |
| `gofmt -l` over touched Go files | 0 | No output. |
| `cmp -s CHANGELOG.md <(git show ab34556e:CHANGELOG.md)` | 0 | Exact trunk bytes. |
| `bash .github/ci/ledger-consistency.sh` | 2 | Usage error because no evidence directory was supplied; the correctly formed command above was run. |

All five refreshed-reader mutant probes each exited 1 because their named regression failed under the narrowed mutant. The first, pre-fix focused test command also exited 1 and is retained above rather than presented as green. Hosted-lane results are supplied by the producer handoff gate.

The rev5 artifact `TASK-260922-cww1ov_review-verdict-rev5.md` says **ACCEPT**, with no findings; it is an identity-only review carrying rev3's content acceptance. That conflicts with the attached brief's claim that rev5 rejected the work. No unspecified rev5 rejection was invented; the named regressions and narrowing mutants from the earlier coverage table remain the evidence for the actual 0017 refusal and admission bounds.


### Final post-edit confirmation

After consolidating the simulated Windows result into `TestReadsDistinguishAbsentFromBlockedParent` and adding the exported-constant comments:

- Focused state-reader package selection across `internal/stateread`, `internal/scriptworker`, `internal/runtimestore`, and `internal/install`: exit 0.
- `go test -count=1 ./cmd/curator -run '^TestEnforcedShimDispatchRefusesUnreadableSidecar$'`: exit 0.
- `go vet ./internal/envprofile ./cmd/curator ./internal/install ./internal/runtimestore ./internal/scriptworker ./internal/stateread`: exit 0.
- Targeted `golangci-lint run` over the six touched Go packages: exit 0 (`0 issues`).
- Final `git diff --check`, `gofmt -l` over touched Go files, changelog byte comparison, conflict-marker search, mutation-residue search, staged-index check, and `.task-board`/skip-class diff check: each exit 0.


## CHANGELOG entry (for release prep)

The following six Story entries are copied verbatim from the pre-restore changelog bytes. The repository `CHANGELOG.md` remains byte-identical to trunk.

- Production-entry test suite for the Decision 0017 credential-link
  repairs and the explicit credential migration (tests only, no behavior
  change). Forty-six rows on temporary stores drive `Resolve`,
  `StatusOf`, `PlanMigration`, `ApplyMigration`, and `curator env
  resolve|migrate`: both 0017 hazards, the mis-targeted vs
  dangling-to-declared split, the inspect → plan → apply migration with
  journal/recovery and no secret copies, every repair refusal class
  with the codex admission table, and no silent repair — each refusal
  proven by a narrowing mutant (conflict admitted means the test
  fails). Every row is registered in
  `.github/ci/platform-cases.tsv`; symlink-dependent rows probe the
  host and skip under the existing `host-capability` vocabulary where
  the Windows host forbids unprivileged creation.


- Explicit credential migration: `curator env migrate
  --inspect|--plan|--apply` (Spec environments §7.4/§10.1, manager
  §12.4/§12.5, Decision 0017). Inspect inventories the old marker, every
  recorded link target, and both Pi roots (`~/.pi/auth.json` and
  `~/.pi/agent/auth.json`) per profile and environment, read-only and
  without reading credential bytes; plan prints the exact operations
  (relink a recorded symlink at the declared native path, unlink a
  stale recorded link) with a plan hash covering the marker identities;
  apply requires the `--expect` hash of a prior plan, prints the locked,
  revalidated plan before the first mutation, and executes exactly the
  printed plan under the manager-home mutation lock through a durable
  migration journal (temp-link + atomic rename per relink), refusing on
  plan drift and on conflicts. A syscall failure mid-op or a failed
  marker publication reverts the links; an interrupted apply leaves the
  journal for the next `--apply --expect <plan-hash>` to recover to the
  prior state before executing. Conflicts that need the operator — an
  isolated→shared account choice, a regular file at a link path, two
  live Pi credentials — refuse with `environment_credential_conflict`
  naming the exact out-of-band decision. No step copies, moves, or
  rewrites credential bytes; the effective mode is preserved and the v1
  marker still records path and strategy only.


- `codex_cli` `isolated` is now admitted under effective `file` credential
  storage only (Spec environments §7.4, Decision 0017). The native
  `config.toml` is parsed as TOML and only the top-level
  `cli_auth_credentials_store` key counts, in any valid spelling; an
  absent file or key resolves to the platform default `file`, so
  `isolated` stays admitted; `keyring` or `auto` is refused with
  `environment_isolated_unsupported`; a selector outside the verified
  `file`/`keyring`/`auto` set fails closed with
  `environment_credential_unsupported` at provisioning and repair, and a
  `config.toml` that does not parse fails closed naming the file instead
  of reading as absent.

- New `pi` managed homes link `auth.json` at the corrected native root
  `~/.pi/agent` (Spec environments §7.4, Decision 0017 choice 3). Existing
  homes linked at the old `~/.pi/auth.json` target are reported detached
  with `environment_credential_conflict`-class wording; the move to the
  agent root runs as the explicit `curator env migrate` step, which
  repair points at instead of re-pointing silently.

- `env resolve --repair` no longer migrates credential ownership (Spec
  environments §10.1, manager §12.4). A recorded link aimed at the wrong
  native target (for example a `pi` home still linked at the pre-0017
  root) and a stale recorded link left behind by a mode change now
  refuse with `environment_credential_conflict` pointing at `curator env
  migrate --plan` / `--apply --expect <plan-hash>` instead of being
  re-pointed or unlinked silently;
  previously repair moved them itself. Repair still re-links an absent
  path and replaces an empty directory, and still refuses a regular
  file, a foreign link, or a non-empty directory without touching bytes.


- Credential-link repairs are fix-first and never destroy bytes (Spec
  environments §7.4/§10.1, manager §12.4, Decision 0017). A stale recorded
  link (`shared`→`isolated`, or a store gone ambient) and a mis-targeted
  recorded link (for example a `pi` home still aimed at the pre-0017
  native root) refuse with `environment_credential_conflict` naming the
  path and pointing at the explicit `curator env migrate` step, removing
  nothing — previously the path was unconditionally removed and
  re-linked. A regular file, a foreign link, or a directory at a link
  path refuses the same way. A correctly targeted link whose native
  target does not exist yet is the distinct detached-pending state — the
  normal pre-login shape — reported as a warning by provisioning, repair,
  and bare resolve, which all succeed loudly; it is never stale and never
  a conflict. A native target that cannot be inspected stays a stale
  conflict/inspection diagnostic.

## Revision 9 follow-up — refreshed ledger and local re-verification (2026-09-26)

This recheck supersedes the earlier 296/297-row counts above. The candidate ledger then had 296 rows, while current trunk had 355; 110 trunk rows were missing and the candidate contained 51 task rows. I rebuilt it from the refreshed trunk bytes and retained the 46 F-C3 rows plus all five state-reader rows. The final file has 406 rows, exactly all 355 refreshed-trunk rows plus the 51 task rows, with no same-key row changes. The ledger gate verifies all 406 rows on Linux, macOS, and Windows.

### Refresh record

- Safety snapshot for this run: git stash create returned 3a5367f3d9e95679f3e4d96b7a5aa6029d1de20d. It was not applied or dropped.
- The prescribed trunk-delta command, with pipefail enabled, exited 1 on overlapping dirty candidate paths. git merge-base --is-ancestor 1511b345 ab34556e and git merge-base --is-ancestor ab34556e HEAD both exited 0: the requested trunk delta was already in the Story history. I retained the existing state-reader migrations and explicitly rebuilt the ledger from refreshed trunk plus task-only rows.
- task-board worktree refresh-candidate TASK-260922-cww1ov returned exit 0 with refresh_already_current: protected main had advanced to faf509ae692670f02987d6924c3c5cfcad614473 and Story branch tip is e3440c47dd9b428b0d5f5003cfb6991ccff177d3. The requested ab34556e is an ancestor of this refresh base. No replay conflict needed a resolution template on the observed retry.
- Restored the checkout-local .task-board artifact and out-of-scope internal/snapshot files to refreshed HEAD. They are not part of the candidate delta. CHANGELOG.md compares byte-for-byte with current main; all six Story entry texts remain verbatim under “CHANGELOG entry (for release prep)”.
- The current task change-request still answers the revision 5 brief: the attached verdict says ACCEPTED and records no finding. The named coverage and mutants below that earlier section are retained; no rejection was invented.

### Local verification record

All commands were run as standalone processes; exit codes are actual process results.

| Command | Exit | Result |
|---|---:|---|
| bash .github/ci/ledger-consistency.sh /tmp/TASK-260922-cww1ov-ledger-consistency-r9 | 0 | 406 rows checked across linux, darwin, windows |
| go test -count=1 -timeout 9m -run 'Test(CredentialLink|SharedToIsolated|StaleCredential|Dangling|Codex|PiProvision|Reviewer|Migrate)' ./internal/envprofile/ | 0 | 15.242s; hazards, dangling target, codex admission/refusals, migration and recovery selection |
| go test -count=1 -timeout 7m -run '^TestMigrateNoSecretCopies$' ./internal/envprofile/ | 0 | 2.679s; dedicated no-copy scan |
| go test -count=1 -timeout 9m ./internal/envprofile/... | 1 | Timed out at 540.454s in TestWeightConflictErrorAndWarning/warning, blocked in the Git-backed overlay clone for git@example.com:personal; no 0017 failure was reported |
| go test -count=1 -timeout 8m -run 'Test(EnvMigrate|EnvResolveRepair|EnforcedShimDispatchRefusesUnreadableSidecar)' ./cmd/curator | 0 | 73.093s; env migration/repair CLI and production dispatcher selection |
| go test -count=1 -timeout 7m -run 'TestReadsDistinguishAbsentFromBlockedParent|TestLoadShimSidecarRefusesUnreadablePath|TestEnforcedShimDispatchRefusesUnreadableSidecar|TestManagedEnforcedShimsInRefusesUnreadableInventory|TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts' ./internal/stateread ./internal/scriptworker ./cmd/curator ./internal/runtimestore ./internal/install | 0 | All five packages passed: 0.889s, 1.515s, 2.150s, 3.005s, and 1.140s |
| go vet ./internal/envprofile ./cmd/curator ./internal/install ./internal/runtimestore ./internal/scriptworker ./internal/stateread | 0 | Touched packages clean |
| GOOS=windows go vet ./internal/envprofile | 0 | Windows compile/vet path clean |
| golangci-lint run ./internal/envprofile ./cmd/curator ./internal/install ./internal/runtimestore ./internal/scriptworker ./internal/stateread | 0 | 0 issues |
| gofmt -l over all changed Go source and test files | 0 | No filenames printed |
| git diff --check | 0 | Clean |
| git diff --cached --quiet | 0 | Nothing staged |
| git diff --quiet main -- .github/ci/skip-classes.tsv | 0 | No skip-class changes |
| cmp -s CHANGELOG.md <(git show main:CHANGELOG.md) | 0 | Exact current-trunk bytes |

### Evidence carried forward

The narrowing-mutant runs were not repeated in this refresh recheck. I accept the attached prior evidence: 47/47 F-C3 mutants killed with passing witnesses, plus all five refreshed-reader narrowing mutants killed by their named production-entry tests. No build-failure kill was counted. These rows and mutant descriptions remain in the coverage and mutant tables above. The handoff gate still supplies hosted-lane results for the refreshed candidate.

## Notes (moved from LOGBOOK)

### 2026-09-26 — revision 9 refresh reconciliation (TASK-260922-cww1ov)

- FINDING: a prior 297-row ledger count came from an incomplete candidate file. The refreshed trunk ledger had 355 rows; the candidate was missing 110 of them. Rebuilt the ledger from current trunk and retained all 51 task rows. `ledger-consistency.sh` now checks 406 rows across Linux, macOS, and Windows with exit 0; no trunk row is dropped and there are no same-key row changes.
- REFRESH: trunk `ab34556e` was already an ancestor of the candidate. `task-board worktree refresh-candidate` confirmed `refresh_already_current` on current trunk `faf509ae`, with branch tip `e3440c47` (exit 0). `CHANGELOG.md` matches current trunk byte-for-byte. Restored the local `.task-board` checkout artifact and out-of-scope snapshot files to refreshed `HEAD`.
- VALIDATION: focused 0017, no-copy, env CLI, and refreshed state-reader tests passed. The package-wide envprofile run exited 1 at its 9-minute timeout in the unrelated Git-backed `TestWeightConflictErrorAndWarning/warning` fixture while cloning `git@example.com:personal`; no credential-row failure appeared. The failure remains recorded in the task results.

### 2026-09-26 — refreshed credential-link candidate onto `ab34556e`

TASK-260922-cww1ov refreshed the uncommitted envprofile candidate onto the
current trunk. The refresh exposed one stale hosted-ledger row: the old
`TestARefusalPrecedesEveryWorkerSurface` test had been removed when R5 made the
enforced shapes reachable. Replaced that row with the live
`TestEnforcedShapesProceedPastAdmission` assertion; ledger consistency now
checks 297 rows across Linux, macOS, and Windows. The trunk refresh also added
sidecar, shim-inventory, and locked-checkout readers; they now preserve
unreadable state separately from proven absence, including Windows
path-not-found below a regular-file ancestor.

## Revision 11 — LOGBOOK reverted, refresh onto 0a628621

Recovery snapshot captured before rework: `git stash create` returned
`3dc9791d6d092d27e1ce36fc00650017add565f5`.

The required trunk combine command returned exit 1 because the three-way patch
could not match the platform ledger index. I applied the remaining trunk paths
with that ledger excluded (exit 0), preserved all candidate rows, and merged
the one trunk executable-identity row into `.github/ci/platform-cases.tsv`.
The refreshed tree contains both the trunk row and the 46 F-C3 rows plus the
state-reader seam rows. No conflict markers remain and nothing is staged.

`CHANGELOG.md` and `LOGBOOK.md` match trunk `0a628621` byte-for-byte. The
working-tree diff from `0a628621` contains only Story paths and the required
trunk-reader seam migration; it contains no `.task-board` checkout artifacts,
`LOGBOOK.md`, or `CHANGELOG.md`.

`task-board worktree refresh-candidate TASK-260922-cww1ov` exited 0 with
`refresh_advanced`, trunk and reviewed trunk `0a62862130fd6cbd9db8b246da63fdb5aecdfaaf`,
and branch `b120247dd9e7a75fb85f7bf3afc2ece5cf504540`.

The only newly added production reader in that trunk delta was
`internal/conformancecoverage/coverage.go`. Its repository-marker inspection
and both ledger reads now use `internal/stateread`; repository-root discovery
continues only on proven absence and fails closed on unreadable metadata. The
sidecar, shim-inventory, and locked-checkout readers already use the same seam.
Added `TestCoverageStateReadsDistinguishAbsentFromUnreadable` and
`TestRepositoryRootDoesNotFallbackAfterBlockedMarkerRead`.

### Fresh bounded validations

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout 9m -run 'Test(CredentialLink|SharedToIsolated|StaleCredential|Dangling|Codex|PiProvision|Reviewer|Migrate)' ./internal/envprofile/` | 0 | 81.965s; selected F-C3 production-entry rows, including no-copy. |
| `go test -count=1 -timeout 8m -run 'Test(EnvMigrate|EnvResolveRepair|EnforcedShimDispatchRefusesUnreadableSidecar)' ./cmd/curator` | 0 | 111.265s; migration/refusal CLI and dispatch boundary. |
| `go test -count=1 ./internal/stateread` | 0 | 0.985s. |
| `go test -count=1 ./internal/conformancecoverage` | 0 | 1.167s; new blocked-read and no-fallback regressions. |
| Refreshed-reader selection across `internal/stateread`, `internal/scriptworker`, `cmd/curator`, `internal/runtimestore`, and `internal/install` | 0 | All five packages passed, 1.101s, 1.985s, 3.057s, 1.039s, and 2.033s. |
| `go vet ./internal/envprofile ./cmd/curator ./internal/install ./internal/runtimestore ./internal/scriptworker ./internal/stateread ./internal/conformancecoverage` | 0 | Clean. |
| `GOOS=windows go vet ./internal/envprofile ./internal/install` | 0 | Clean. |
| `bash .github/ci/ledger-consistency.sh /tmp/TASK-260922-cww1ov-ledger` | 0 | 407 rows checked across Linux, macOS, and Windows. |
| `go test -count=1 -timeout 9m ./internal/envprofile/...` | 1 | Timed out after 541.704s during unrelated `TestImportSkillForeignPinnedByRevision` temporary-tree cleanup in `os.RemoveAll`. The targeted 0017 selection passed. |

The 47/47 F-C3 narrowing-mutant results and prior three-hosted-lane evidence
remain attached in the earlier results sections; they were not rerun in this
refresh-only pass. The earlier hosted-lane evidence predates this refresh and
is not claimed as a current-tree hosted pass.

### Bounded envprofile inventory and mutant evidence

`go test -list '^Test' ./internal/envprofile/` exited 0 and listed 213 top-level
tests. The complete list was then run in five non-overlapping name ranges:

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout 7m -run '^Test[A-H]' ./internal/envprofile/` | 0 | 121.668s; 34 top-level tests. |
| `go test -count=1 -timeout 7m -run '^Test[I-L]' ./internal/envprofile/` | 0 | 139.248s; 38 top-level tests, including the import case. |
| `go test -count=1 -timeout 7m -run '^Test[M-P]' ./internal/envprofile/` | 0 | 212.147s; 61 top-level tests. |
| `go test -count=1 -timeout 7m -run '^Test[Q-S]' ./internal/envprofile/` | 0 | 198.222s; 57 top-level tests, including resolve drift repair. |
| `go test -count=1 -timeout 7m -run '^Test[T-Z]' ./internal/envprofile/` | 0 | 221.728s; 23 top-level tests. |

Together those partitions cover 213/213 listed top-level envprofile tests with
all nested subtests enabled. The two unpartitioned acceptance attempts remain
reported as timeouts above; the partition results are not relabeled as a green
unpartitioned invocation.

A narrowing mutant for the new trunk-reader regression changed
`readStateBytes` to treat every `stateread.ReadFile` error as absence. The
expected-red `go test -count=1 -run
'^TestCoverageStateReadsDistinguishAbsentFromUnreadable$'
./internal/conformancecoverage` exited 1 because the blocked-parent case was
misclassified as absent. The saved source was restored byte-for-byte (`cmp`
exit 0), then `go test -count=1 ./internal/conformancecoverage` passed (exit 0).
This is a narrowed-classification kill with a passing witness, not a
build-failure kill.

### Final compile and scope notes

- `go build ./internal/conformancecoverage ./internal/envprofile ./cmd/curator`: exit 0.
- `git diff --check`: exit 0.
- `gofmt -l internal/conformancecoverage/coverage.go internal/conformancecoverage/coverage_test.go`: exit 0, no filenames printed.
- `golangci-lint run --timeout=8m ./internal/conformancecoverage ./internal/envprofile ./cmd/curator ./internal/install ./internal/runtimestore ./internal/scriptworker ./internal/stateread`: exit 0, 0 issues.
- `cmd/curator` scope-specific env migrate/resolve and enforced-dispatch tests
  were rerun in the commands above and passed. The complete `cmd/curator`
  package was not rerun in this refresh; the earlier attached full-package run
  exited 1 at its 9-minute timeout in unrelated
  `TestDraftTransportProviderAdmission/unknown_ssh_provider_refuses_without_fetching`.
- Hosted lanes were not rerun in this session. The prior attached three-lane
  result predates this refreshed tree and is not represented here as current
  hosted evidence; the handoff/review route must establish it for this tree.

## Revision 12 — refresh onto trunk 60498052

### Refresh and scope

- Revision 11 was accepted. `git diff 0a628621 60498052 -- . ':!.task-board' ':!CHANGELOG.md' ':!LOGBOOK.md' | git apply --3way` was attempted as instructed and exited 1 because `.github/ci/platform-cases.tsv` and `internal/install/draftsources.go` no longer matched the current checkpoint index. The partial apply was not treated as a green merge.
- `git merge-tree --write-tree HEAD 60498052` exited 0. I applied the resulting incoming source delta, then merged `internal/install/draftsources.go` against the current candidate and trunk versions. Both sides remain present on the shared ledger/profile test paths and in draft replay.
- `task-board worktree refresh-candidate TASK-260922-cww1ov` exited 0 (`refresh_advanced`), with trunk and reviewed trunk `60498052a1833f7511bd16086d953c8099fe7eed`, branch `762509dbdbfe4b9a6259d55eb85b8cd4b7ed0b02`.
- `git diff --name-only 60498052 -- . ':!.task-board'` contains only Story paths. `CHANGELOG.md` and `LOGBOOK.md` have no diff from refreshed trunk. No merge markers remain and the index is empty.

### Production changes and reader coverage

This refresh adds no 0017 production behavior. It closes the absent-versus-unreadable fallback gap in trunk-added state readers and documents the diagnostic:

| Reader / production call site | Production-path row | Narrowing mutant |
| --- | --- | --- |
| `loadMachinePolicy` used by bare envprofile entries | `TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault/blocked_parent` | unreadable config path takes the permissive default; blocked-parent row fails |
| `snapshot.OpenLocal` used by `draftFrozenNodes` | `TestDraftInstallRefusesUnreadableLocalSnapshotCache`; `TestFrozenSnapshotReadersRefuseBlockedParents/local_snapshot` | unreadable store parent reports snapshot absence; both rows fail |
| `snapshot.AuthenticateGit` used by `replayMissingDraftSnapshot` and `lockedMemberSkillTree` | `TestDraftReplayRefusesUnreadableGitSnapshotCache`; `TestFrozenSnapshotReadersRefuseBlockedParents/git_snapshot` | unreadable cache parent reports snapshot absence and enters replay; both rows fail |
| `lockedNetworkRepository` used by `draftFrozenInput` | `TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts` | unreadable checkout continues as if absent; its blocked-parent row fails |

The first three production refusal rows are registered for Linux, macOS, and Windows in `.github/ci/platform-cases.tsv`; the existing checkout row remains registered on all three lanes. No skip class was added. The envprofile and snapshot diagnostics keep proven absence eligible for their documented fallback and stop when an ancestor blocks inspection.

`docs/troubleshooting.md` now explains `manager_state_unreadable`, including when a higher-level diagnostic wraps it. `CHANGELOG.md` remains byte-identical to refreshed trunk per the refresh instruction.

### F-C3 regression coverage

The accepted revision 11 coverage table and its 47/47 narrowing mutants remain unchanged; no F-C3 test was weakened and no 0017 product code changed. The focused 0017 API and CLI rows, including migration journal/recovery, no-copy, all refusal classes, the Codex admission table, and no-silent-repair rows, were rerun green below. The three-lane ledger gate reports 417 rows checked across Linux, macOS, and Windows.

### Revision 12 validation

Every command below ran as its own process. Times are package-reported durations.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -timeout 7m -run 'Test(CredentialLink|SharedToIsolated|StaleCredential|Dangling|Codex|PiProvision|Reviewer|Migrate)' ./internal/envprofile/` | 0 | 19.766s; F-C3 production-entry rows, including migration/recovery and no-copy. |
| `go test -count=1 -timeout 8m -run 'TestEnv(Migrate|ResolveRepair)' ./cmd/curator/` | 0 | 44.604s; CLI migration/refusal rows. |
| `go test -count=1 ./internal/stateread ./internal/snapshot` | 0 | 2.600s before mutants; 1.854s after restoration. |
| `go test -count=1 -timeout 7m -run '^(TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts|TestDraftReplayRefusesUnreadableGitSnapshotCache|TestDraftInstallRefusesUnreadableLocalSnapshotCache)$' ./internal/install` | 0 | 1.811s before the discarded fixture experiment. |
| `go test -count=1 -timeout 9m -run '^TestDraftSourcesPlaybookCollectionAcceptanceThroughProductionCLI$' ./internal/crossconformance` | 0 | 53.725s before final reader restoration; 46.878s after restoration. |
| `go test -count=1 -timeout 7m -run '^Test[A-H]' ./internal/envprofile/` | 0 | 43.093s. |
| `go test -count=1 -timeout 7m -run '^Test[I-L]' ./internal/envprofile/` | 0 | 101.004s. |
| `go test -count=1 -timeout 7m -run '^Test[M-P]' ./internal/envprofile/` | 0 | 114.338s. |
| `go test -count=1 -timeout 7m -run '^Test[Q-S]' ./internal/envprofile/` | 0 | 126.093s. |
| `go test -count=1 -timeout 7m -run '^Test[T-Z]' ./internal/envprofile/` | 0 | 89.207s. |
| Final combined machine-policy and F-C3 selection in `internal/envprofile` | 0 | 14.258s. |
| Final `go test -count=1 ./internal/stateread ./internal/snapshot` | 0 | 1.854s. |
| Final install replay/refusal and declared-source replay selection | 0 | 6.596s. |
| Final env CLI, profile-path and sidecar-dispatch selection in `cmd/curator` | 0 | 42.640s. |
| Final playbook production-CLI acceptance row | 0 | 46.878s. |
| `bash .github/ci/ledger-consistency.sh /tmp/TASK-260922-cww1ov-ledger-rev12` | 0 | 417 rows checked across linux, darwin, windows. |
| `go vet ./internal/envprofile ./internal/stateread ./cmd/curator ./internal/install ./internal/snapshot ./internal/conformancecoverage ./internal/scriptworker ./internal/runtimestore` | 0 | Clean. |
| `GOOS=windows go vet ./internal/envprofile ./internal/install` | 0 | Clean. |
| `go build ./internal/envprofile ./internal/stateread ./internal/snapshot ./internal/install ./cmd/curator ./internal/conformancecoverage ./internal/scriptworker ./internal/runtimestore` | 0 | Built. |
| `golangci-lint run --timeout=8m ./internal/envprofile ./internal/stateread ./cmd/curator ./internal/install ./internal/snapshot ./internal/conformancecoverage ./internal/scriptworker ./internal/runtimestore` | 0 | 0 issues. |
| `git diff --check` | 0 | Clean. |
| `gofmt -l` over changed and new Go files | 0 | No files printed. |
| Merge-marker scan excluding `.task-board` | 0 | No merge markers. |

### New expected-red narrowing mutants

Each mutant was applied temporarily, the named test command exited 1, and the production source was restored before the corresponding green witness above. These are gate kills, not build-failure kills.

| Site | Mutant | Expected-red command | Exit / observation |
| --- | --- | --- | --- |
| `loadMachinePolicy` | unreadable stateread errors return the empty permissive policy | `go test -count=1 -run '^TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault$/^blocked_parent$' ./internal/envprofile` | 1; blocked-parent test observed `err=nil` and permissive policy. |
| `localSnapshotMetadata` | unreadable state maps to `source_snapshot_unavailable` | `go test -count=1 -run '^Test(DraftInstallRefusesUnreadableLocalSnapshotCache|FrozenSnapshotReadersRefuseBlockedParents)$' ./internal/snapshot ./internal/install` | 1; direct reader lost the typed unreadable class and production install returned the platform's raw not-a-directory error. |
| `snapshot.AuthenticateGit` | unreadable cache errors map to snapshot absence | `go test -count=1 -run '^Test(DraftReplayRefusesUnreadableGitSnapshotCache|FrozenSnapshotReadersRefuseBlockedParents)$' ./internal/snapshot ./internal/install` | 1; direct reader and production install failed their typed-state assertions after entering replay. |
| `lockedNetworkRepository` | unreadable checkout inspection continues like absence | `go test -count=1 -run '^TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts$/^failed-lstat-stops-fallback$' ./internal/install` | 1; the blocked checkout collapsed to `no checkout for locked repository`. |

### Discarded fixture attempt and hosted gate

An exploratory `TestDraftInstallRefusesBlockedRootCheckout` command exited 1 because its fixture built an alias-selected root (`Selection != nil`) and therefore did not enter the legacy transitive-member branch it was intended to exercise. I removed that test and its row; it is not product evidence. The registered `lockedNetworkRepository` production-reader row covers the tested transitive checkout helper.

The revision 11 hosted result predates trunk `60498052` and is not represented as revision 12 evidence. This developer run established local Linux checks, Windows vet, and cross-lane ledger coverage only; the current hosted test/race/gate lanes remain to be established by the review handoff.
