# TASK-260922-cww1ov results — F-C3 0017 production-entry tests (revision 2)

Producer: developer. Worktree: `.temp/STORY-260922-1cenbr/worktree`
(uncommitted). Base: F-C1 + F-C2 checkpointed on the Story branch, plus
the accepted revision 1 of this leaf. Revision 2 answers the rev1
verdict (`TASK-260922-cww1ov_review-verdict-rev1.md`): one P2 coverage
defect (R1 — the no-copy row could not see manager state) plus evidence
bookkeeping. Everything else in revision 1 is accepted as reviewed and
is unchanged. Zero production lines changed (rev1 and rev2 alike).

## Rev2 changes (what moved since rev1)

R1 — the no-copy row now sees manager state (TEST fix, no production
change):

- `credentialHolders` (`internal/envprofile/migrate_test.go:118`) now
  scans the ENTIRE temporary manager home — environments, profiles,
  state, the journal, backups, temp paths, no path allowlist — plus
  every native root. Allowance: none, stated in the helper comment:
  lock/journal metadata holds only hashes, paths, and digests, never
  credential content, so any holder below the home is a secret copy
  and fails the row.
- `TestMigrateNoSecretCopies` (`migrate_test.go:291`) gains an
  interrupted-apply subtest (`:414`, "interrupted journal holds no
  credential bytes"): a fresh fixture crashes after the first
  operation through the F-C2 crash-injection point; while the journal
  stands the test asserts it parses, holds no credential bytes
  (explicit read), and the whole-home scan finds only the two native
  holders; the recovered apply then converges clean. A deleted journal
  cannot be inspected after success, so this covers transient journal
  content.
- New narrowing mutant C47 (relink copies bytes ONLY into manager
  state — the exact rev1 survivor shape,
  `<dir of migrationJournalPath>/credential-backup`): executed in
  rev2, KILLED by the named row with a green witness (table below).

Bookkeeping (required, secondary):

- Lane shapes recounted from `.github/ci/platform-cases.tsv`: 37
  all-platform rows with Windows host-capability tolerance, 5
  all-platform rows with no tolerance, 4 unix-only rows (46 total).
  Rev1 results.md said 39/4/3; the true numbers are 37/5/4.
- `.github/ci/platform-cases.tsv:474-476` said three ENOTDIR rows;
  there are four. Comment fixed.
- This file is the COMPLETE results resource (rev1's attachment ended
  in a truncation marker). Written to disk, then attached with
  `task-board resource add`.

## Outcome

46 production-entry rows on temporary stores drive `Resolve`,
`StatusOf`, `PlanMigration`, `ApplyMigration`, and `curator env
resolve|migrate` across both 0017 hazards, the mis-targeted vs
dangling-to-declared split, the migration end-to-end, every repair
refusal class with the codex admission table, and no silent repair.
Three rows are new (the `environment_repair_failed` class at both
entries, and apply-under-lock); one existing row gained a negative
assertion pinning the recorded/foreign refusal split (rev1, named
below, no weakening); the no-copy row is strengthened in rev2 as
above (no weakening — all prior assertions intact); 36 existing rows
gained the Windows symlink privilege probe. Every row is registered
in `.github/ci/platform-cases.tsv` (46 new rows, one section, no new
skip class). Every row carries a narrowing mutant: 47/47 killed in
rev1 (accepted), C18 re-executed in rev2 (killed), C47 new in rev2
(killed) — 48/48 with green witnesses. `go test
./internal/envprofile/` focused masks pass; ledger-consistency passes.

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
| no secret bytes copied | TestMigrateNoSecretCopies | P+A+R | C18 in-scope backup; C47 state-only copy (rev2) | ALL+probe |
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

Rev1 driver: `/tmp/cww1ov_mutants.py` (scratch, not committed). One
mutant at a time, tree restored after each (`git status` shows only
the intended files; `go build ./...` green). No build-failure kills:
the driver reports those INVALID, never kills. 47/47 killed, 47/47
witnesses green (accepted in rev1; not re-executed wholesale in rev2
per the rework brief).

Rev2 driver: `/tmp/cww1ov_rev2_mutants.py` (scratch, attached as
`TASK-260922-cww1ov_rev2-mutants.py` with its log for the reviewer).
C18 re-executed against the strengthened row: KILLED. C47 new:
KILLED. Both witnesses green. No build-failure kills. Total: 48/48
killed, 48/48 witnesses green (47 accepted + C47 new; C18 counted in
both campaigns).

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
| C18 | executeMigrationOps (migrate.go) | conditional | NoSecretCopies (rev1 + rev2 re-run) | OldRootBytes |
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
| C47 (rev2, NEW) | executeMigrationOps relink arm (migrate.go) | conditional (state-only copy) | NoSecretCopies FAIL: `no new file carries Pi credential content: map[…/state/credential-backup:true …]` | MigratePiWrongTargetToAgentRoot pass |

C47 shape (the rev1 survivor, now killed): in the `MigrateOpRelink`
arm, when `op.To` reads, write those bytes to
`filepath.Join(filepath.Dir(migrationJournalPath(req.Home)),
"credential-backup")` (0600), propagating write errors. Under rev1's
scan the row exited 0 (reviewer's
`TASK-260922-cww1ov_review-state-secret-copy.log`); under rev2's
whole-home scan it fails at the holders assertion naming the state
path, and the in-scope witness (snapshot + byte-identity, no state
scan) still passes — proving the bound is narrowed to the state scan,
not a blanket failure.

Disclosed iterations (mutant bugs, not test bugs, rev1): C14 first
shape shadowed `ok` (no-op, survived once, fixed); C30 first shape
(reconcile skip) was unobservable behind the best-effort temp cleanup,
replaced by the rename-swallow; C37 first shapes skipped the marker
republish along with validation (drift survived to the drift check),
corrected to skip validation but keep republish; C43 first shape fell
through to the next usage gate with the same exit code, retargeted to
the --plan --expect case. Rev2: no iterations — C47 killed on first
execution, C18 re-killed on first execution.

## Ledger delta (corrected recount)

`.github/ci/platform-cases.tsv`: +46 rows in one new section
("Decision 0017 … (F-C1/F-C2/F-C3)"). 39× internal/envprofile, 7×
cmd/curator. Shapes, recounted from the ledger in rev2 (rev1's
39/4/3 was a miscount): 37 must_all+skip_windows/host-capability
(link-capability probe), 5 must_all with no tolerance
(double-quoted/single-quoted/TOML admission, unknown operands, CLI
usage), 4 unix-only must_linux,darwin+skip_windows/host-capability
(ENOTDIR inspection). No `.tsv` vocabulary change, no new skip class,
no `Parent/*` rows (all skips are parent-level probes). The section
comment now says four ENOTDIR rows (was three).

## Test strengthening (two, named)

1. (rev1) TestCredentialLinkUnrecordedSymlinkRefuses gained a negative
   assertion: an unrecorded (foreign) refusal must NOT say "migration
   needed" (only the recorded mis-targeted refusal points at
   migration). No other existing assertion was touched; the probe
   insertions (31 lib + 5 CLI first-line calls) are no-ops on capable
   hosts and gate-fatal skips on unix if they ever fire.
2. (rev2) TestMigrateNoSecretCopies: the credential-content scan
   covers the entire manager home (was: environments/profiles only),
   and a new interrupted-apply subtest scans while the journal stands
   plus an explicit standing-journal read. All prior assertions in the
   row are intact; the widened `credentialHolders` also strengthens
   TestMigrateSyscallFailureRollsBack's no-copy assertion (same
   helper, still green).

## Bounds (not owned here — complete)

- No product change: the suite found no defect; production is
  byte-identical. `git status` shows only `*_test.go`,
  `.github/ci/platform-cases.tsv`, `CHANGELOG.md`, and
  `docs/troubleshooting.md`; no product file is touched.
- TestMigrateFailedRelinkKeepsOldLink stays helper-direct by
  construction (real-OS oversized-target rejection is unplannable
  through the production entries) and is excluded from the 46 ledger
  production rows.
- Crash tests use the `InjectFault` seam, not an actual SIGKILL; the
  owned-temp setup is crafted. The interrupted-apply subtest proves
  the standing journal parses and holds no credential bytes, but no
  journal-content mutant is claimed: the journal is JSON, so any
  mutant embedding credential bytes in it breaks journal validity
  and fails closed for the wrong reason. The transient scan reuses
  the C47-proven whole-home helper at the crash point instead.
- Full local package suite is not claimed: the reviewer observed a
  180s local timeout on the full `./internal/envprofile` package in
  rev1. The three hosted lanes are the arbiter; the bounded focused
  masks below are green locally.
- No cross-platform result is inferred from local Darwin beyond
  verified hosted evidence. The rev1 hosted gate (run 35716709852,
  exact rev1 tree) is accepted per the rev1 verdict; a fresh hosted
  gate for the revised candidate runs at publish.
- CLI rows are accepted from rev1 and the reviewer's independent
  rerun; rev2 touches no CLI file, so they were not re-run.

## Evidence (rev2; shell: bash; `-count=1`; bounded timeouts)

Reran in rev2 (this session, real exit codes):

- `go vet ./internal/envprofile/` → exit 0.
- `go test ./internal/envprofile -run
  '^TestMigrateNoSecretCopies$' -count=1 -timeout=120s -v` → exit 0
  (12.958s; parent + interrupted subtest pass).
- `go test ./internal/envprofile -run
  'TestMigrate(SyscallFailureRollsBack|InterruptedApplyRecovers|
  PiWrongTargetToAgentRoot|OldRootBytesWarnAndProceed|
  RollbackRestoresPriorState|PublishFailureRevertsLinks|
  RecoveryRefusesUnexpectedTarget|SealedJournalDeletesWithoutTouching|
  BrokenJournalRefuses)' -count=1 -timeout=300s` → exit 0 (30.843s).
- `python3 /tmp/cww1ov_rev2_mutants.py` → exit 0: C18 KILLED
  (killer FAIL naming `environments/acme/pi/auth.json.bak`, witness
  pass), C47 KILLED (killer FAIL naming `state/credential-backup`,
  witness pass). Tree restored after each (`git status` shows only
  the intended files; `credential-backup` absent from `migrate.go`).
- `go test ./internal/envprofile -run
  'Test(CredentialLink|SharedToIsolated|StaleCredential|Dangling|
  Codex|PiProvision|Reviewer|Migrate)' -count=1 -timeout=280s` (the
  reviewer's focused mask) → exit 0 (76.538s).
- `bash .github/ci/ledger-consistency.sh /tmp/cww1ov-rev2-ledger` →
  exit 0 (287 rows checked across linux darwin windows).
- `git diff --check` → exit 0. `go build ./...` → exit 0.

Accepted without re-run (per the rework brief; untouched by rev2):

- Rev1's C1–C46 campaign except C18 (accepted as reviewed).
- CLI rows (`cmd/curator`: TestEnvMigrate*, TestEnvResolveRepair*) —
  no CLI file changed in rev2.
- Rev1 hosted gate run 35716709852 (exact rev1 tree; reused, not
  replayed). Fresh hosted evidence for the revised candidate is
  produced by the publish gate.

## Handoff

Ready for review (revision 2). Role handoff via `task-board handoff
TASK-260922-cww1ov --role developer`. This is the FINAL leaf of
STORY-260922-1cenbr: on acceptance the orchestrator integrates the
Story.

---

# Revision 3 — base refresh onto trunk 48da2690 (refresh-only, no content change)

Producer: developer. Revision 2 was ACCEPTED
(`TASK-260922-cww1ov_review-verdict-rev2.md`) but the integration
refused: trunk advanced to `48da2690` (rc.12 pin promotion, PR #79)
with a `CHANGELOG.md` change revision 2 also makes, so the record
went stale and acceptance was released (`worktree
invalidate-acceptance`). This revision is the base refresh per
`cww1ov-refresh-1.md`: reparent the candidate onto current trunk,
prove byte-identity except the combined CHANGELOG, re-run the owned
narrow rows, handoff. No product change, no test weakened, no row
re-designed. F-C1/F-C2 reviewer rows stay as committed.

Naming note: `TASK-260922-cww1ov_results.md` (rev1) is historical and
is left untouched; this `TASK-260922-cww1ov_results-rev3.md` is the
rev2 record verbatim plus this section, following the rev2 precedent.

## Refresh execution

`task-board worktree refresh-candidate TASK-260922-cww1ov` (shell
bash, from the control root) first refused with a regular-file
replay conflict (exit 1): replaying the F-C1 checkpoint `637e5b43`
onto `48da2690` collides in `cmd/curator/env_test.go`,
`TestEnvStatusMatrix` — trunk inserts the §12 provider-stub block
(`bin` + `PATH`, 15 lines) and F-C1 inserts
`writeNativeCredentials(t)` (1 line) at the same anchor
(`source, _ := profileHome(t)`). Both insertions are independent
test setup (credential-seed files vs stub binaries on `PATH`).

Resolution (test-only combination, no production line): union — keep
trunk's 15-line block AND F-C1's line, F-C1 line first so each
insertion stays adjacent to its own anchor side. Verified before
submitting: no conflict markers remain, `gofmt` clean, diff vs
trunk is exactly the two F-C1 one-line insertions (the other hunk
had applied cleanly in the replay), sha256 of the resolved file
`754fd00ff0f55ddd53918bf212c341082aae6c04d68c215dfefb222bd00aa56b`.
Submitted via `--replay-resolutions` with the packet bound to the
checkpoint, path and content digest
(`checkpoint_oid`/`path`/`sha256`/`content`-base64,
`resolution refused: missing unused replay resolution` guided the
schema; resolved bytes verified, never a hand-committed replay).
Result: `refresh_advanced`, exit 0. New chain: `607770e0`
(replayed F-C2) on `d3476cb6` (replayed F-C1) on `48da2690`.

The tool preserves working bytes exactly, so after the branch move
the tree showed the trunk delta as worktree-vs-HEAD noise (41
paths). Every one of those 41 worktree files was verified
byte-equal to the old-F-C2 blob (or absent in both; 41 checked, 0
mismatches — old bytes survive in git objects and the rev2 patch),
then synced to the replayed HEAD. `cmd/curator/env_test.go` took
the HEAD version carrying the union resolution above.

## What stayed identical (byte-level)

- Uncommitted candidate: 6 of 7 tracked files byte-identical to
  rev2 (sha256 pre == post): `.github/ci/platform-cases.tsv`
  (`02911995…`), `cmd/curator/env_migrate_test.go` (`d3a04674…`),
  `docs/troubleshooting.md` (`457ccdb9…`),
  `internal/envprofile/credential_link_test.go` (`2c620931…`),
  `internal/envprofile/migrate_test.go` (`2b4c8855…`),
  `internal/envprofile/reviewer_recovery_test.go` (`66fb9b5a…`);
  untracked `internal/envprofile/credential_production_test.go`
  identical (`30b1c2d1…`).
- `CHANGELOG.md`: worktree file = replayed-HEAD blob + our 13
  F-C3 lines, inserted after line 8 (the same `8a9,21` hunk shape
  as rev2: F-C3 entry first, then trunk's E4/rc.12 block, then the
  replayed F-C1/F-C2 entries, each exactly once). Verified by
  deleting lines 9–21 and byte-comparing against the HEAD blob.
- F-C1 replay: same changed-file list as the original checkpoint,
  same added/removed lines (sorted `+`/`-` sha256 `48cd6432…`
  both sides — the resolution only moved hunk position).
- F-C2 replay: same file list, same added/removed lines (sha256
  `f4e17522…` both sides; raw patches differ only in index hashes
  and hunk line numbers).
- Story path set vs trunk: 22 paths before, 22 after, empty diff.
- `git status --short`: exactly the rev2 8 paths (7 modified +
  1 untracked). `git diff --stat/--numstat` vs HEAD: identical
  shape to rev2 (317 insertions, 4 deletions; per-file counts
  unchanged).

## Re-runs on the refreshed tree (shell bash, `-count=1`, real exits)

- `go build ./...` → exit 0 (6.1s).
- `go vet ./internal/envprofile/ ./cmd/curator/` → exit 0 (test
  binaries compile: no symbol collision with trunk-added test
  files).
- `go test ./internal/envprofile -run
  '^TestMigrateNoSecretCopies$' -count=1 -timeout=150s -v` →
  exit 0 (2.4s; parent + `interrupted_journal_holds_no_credential_bytes`
  pass).
- Reviewer's focused mask `Test(CredentialLink|SharedToIsolated|
  StaleCredential|Dangling|Codex|PiProvision|Reviewer|Migrate)` →
  exit 0 (16.7s): 43 PASS, 0 FAIL, 0 SKIP. Covers both 0017
  hazards, both dangling states, the codex admission table, and
  the migration/recovery rows.
- `go test ./cmd/curator -run 'TestEnv' -count=1
  -timeout=280s -v` → exit 0 (217.5s): all 7 owned CLI rows pass
  (`TestEnvMigratePlanApplyPi`, `TestEnvMigrateConflictRefuses`,
  `TestEnvResolveRepairNeedsMigration`, `TestEnvMigrateUsage`,
  `TestEnvMigrateApplyRequiresPlan`,
  `TestEnvMigratePrintBeforeWrite`,
  `TestEnvResolveRepairFailedOnUninspectable`), 0 SKIP, 0 FAIL.
  `TestEnvStatusMatrix` passes, validating the `env_test.go`
  union resolution end to end against trunk's §12 assertions.
  Trunk production files under `cmd/curator` changed, so the CLI
  re-run was required, not optional.
- Rev2 mutant driver (`TASK-260922-cww1ov_rev2-mutants.py`,
  anchors hold: trunk never touched `migrate.go`) → exit 0:
  C18 KILLED (in-scope backup named at `migrate_test.go:369`),
  C47 KILLED (state-only `state/credential-backup` copy named),
  both witnesses pass; tree restored to the 8 paths, no
  `credential-backup` residue in `migrate.go`.
- New-trunk coherence: 11 trunk-added `internal/envprofile`
  tests (surfacing/admission/S4) → exit 0 (29.8s), all pass —
  no interaction with the candidate test files.
- `TestEnvironmentsEnvPassthroughVectors` → SKIP, exit 0
  (`CURATOR_CONFORMANCE_ROOT is not set`).
- `bash .github/ci/ledger-consistency.sh` → exit 0 (287 rows
  across linux/darwin/windows). `git diff --check` → exit 0.
  `gofmt -l` over all touched Go files → clean.

## Skip/execute analysis (the rc.12 pin)

`SPEC_PIN` moved `87a0d006…` → `dced9b83…` (v1.0.0-rc.12,
`.github/workflows/ci.yml:53`). Vectors are consumed only
through `CURATOR_CONFORMANCE_ROOT`, which is unset locally —
exactly as in rev2. None of the 46 owned rows references
conformance vectors (grep over the five owned test files:
no match), so NO owned row changed skip/execute state on the new
base: previously-executing rows execute (43 lib + 7 CLI, zero
skips on darwin), and the only vector-gated test in scope is
trunk's `TestEnvironmentsEnvPassthroughVectors`, which skips as
before. No row changed behaviour; no fix was needed; no other
row exists that "previously SKIPPED" and "now executes".

## Handoff

Ready for review (revision 3, refresh only). Role handoff via
`task-board handoff TASK-260922-cww1ov --role developer`. This is
still the FINAL leaf of STORY-260922-1cenbr: on re-acceptance the
orchestrator integrates the Story. Publish only on a green gate.
