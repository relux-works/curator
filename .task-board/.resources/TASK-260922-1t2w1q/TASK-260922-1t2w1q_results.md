# TASK-260922-1t2w1q results — F-C2 explicit credential migration, Revision 2

Producer: developer. Worktree: `.temp/STORY-260922-1cenbr/worktree` (uncommitted).
Normative source: curator-spec main 05053cd7 — environments §7.4/§10.1,
manager §12.4/§12.5, decision 0017 (choice 3 Pi roots, option O3
inspect → plan → apply). Base: revision-1 tree (F-C1 checkpointed) +
binding rework-1 (P1-a/b/c, P2). The four reviewer probes are committed
as rows (see row table).

## Revision 2 fixes

P1-a — every apply requires a prior complete plan identity.
`ApplyMigration` refuses with zero writes when `Expect` is empty
(conflict diag, naming `--plan`/`--expect`), before touching the
journal; the CLI rejects bare `--apply` as usage. The test that blessed
no-expect apply is replaced by a refusal row; repair hints now name
`--plan` then `--apply --expect <plan-hash>`.

P1-b — the locked, revalidated plan prints before the first mutation.
`ApplyMigration` takes `MigrateRequest.Print` (the CLI passes stdout)
and writes the plan under the held lock after re-inventory, before any
mutation; a write failure fails before anything changes. Recovery (when
present) announces itself on the same writer before its mutations. A
temporal CLI row (writer hook observing the link on first write)
replaces final-string containment.

P1-c — durable migration journal + atomic replacement + recovery.
Journal: `<home>/state/env-migration.json` (manager state, never the
credential scope; holds no credential bytes), schema
`{version:1, plan:<hash>, ops:[{profile,env,kind,path,from,to,done}],
markers:[{live,prior,new}], markers_done, complete}`,
written temp+sync+atomic-rename before the first mutation and after
each landed op. Relink sequence: `symlink(target, dir/.migrate-<rand>.tmp)`
then `rename(tmp, link)` — never remove-then-create, so any syscall
failure leaves the old link standing; a failed rename removes its own
temp. Recovery rules (deterministic, rollback-to-prior): the next apply
carrying a plan hash recovers under the lock before planning — links
converge to `From` in reverse order (non-symlinks fail closed naming
the operator choice), prior marker bytes republish, strays matching
`.migrate-*.tmp` sweep, journal deletes; a sealed (`complete`) journal
only deletes and never rolls back finished work; an unusable journal
refuses naming backup-first. Inspect/plan surface a leftover journal
read-only (banner + `recover:` footer with the original hash) and
recover nothing. A failed marker publication now reverts the links
(rev1 left them standing).

P2 — the plan hash covers the marker identities: one
`sha256(raw marker bytes)` line per provisioned home (no credential
reads) plus the recovery state; canonical domain bumped to
`migration-plan-v2`.

## Row table (Revision 2 deltas; all through production entries on temp stores)

| Fix | Test | Entry | Asserts |
|---|---|---|---|
| P1-a | `TestMigrateApplyRequiresPlan` (committed reviewer missing-plan) | `ApplyMigration` | refuses conflict + `needs a prior plan` + `--plan`/`--expect`; zero applied; link, journal, snapshot untouched |
| P1-a | `TestEnvMigrateApplyRequiresPlan` (CLI probe) | `run()` | `--apply` alone → usage naming `--expect`; link untouched |
| P1-a | updated `TestMigratePlanDriftRefuses`, `assertMigrateConflict`, no-copy/old-root/pending/rollback rows, `TestSharedToIsolatedRemovesStaleLink`, `TestEnvMigrateConflictRefuses` | plan-then-apply | every apply carries its fresh hash; converged empty plan applies clean |
| P1-b | `TestMigratePrintFailureRefusesBeforeMutation` | `ApplyMigration` + failing writer | `cannot print` refusal; zero writes, no journal |
| P1-b | `TestEnvMigratePrintBeforeWrite` (committed reviewer probe) | `run()` + writer hook | first stdout write precedes the link move; link migrated after |
| P1-c | `TestMigrateFailedRelinkKeepsOldLink` (committed reviewer probe, helper-level bound) | `executeMigrationOps` + `reconcileJournalLinks` | OS-rejected oversized target: old link stands, no strays |
| P1-c | `TestMigrateSyscallFailureRollsBack` (symlink + rename halves) | `ApplyMigration` + injected failure | old link stands; rolled back; holder set unchanged; journal deleted; snapshot identical |
| P1-c | `TestMigrateInterruptedApplyRecovers` | `ApplyMigration` crash fault → `PlanMigration` → `ApplyMigration` | kill leaves journal (1/2 done) + half-done links; inspect/plan banner read-only; same-hash retry recovers then migrates fully; announcement precedes plan print; journal gone |
| P1-c | `TestMigratePublishFailureRevertsLinks` | `ApplyMigration` + `publish-begin` fault | `marker publish failed` + rolled back; links/markers/natives restored; journal deleted |
| P1-c | `TestMigrateSealedJournalDeletesWithoutTouching` | crafted sealed journal + `ApplyMigration` | retry deletes without touching completed work; honest drift refusal; re-plan converges |
| P1-c | `TestMigrateBrokenJournalRefuses` | garbage journal + both entries | `unusable` + `blocked:` banner; apply refuses backup-first; zero writes; journal kept |
| P2 | `TestMigrateMarkerDriftRefuses` (committed reviewer probe) | `PlanMigration` → edit `lock_sha256` → `ApplyMigration` | stale hash drift-refuses; zero writes, no journal |
| repair hint | `TestDanglingPiLinkReportedDetached`, `TestEnvResolveRepairNeedsMigration` (unchanged assertions) | `Resolve --repair` | still refuse `migration needed` + `env migrate` + path + targets; new hint names `--plan` then `--apply --expect` |

Rev1 rows otherwise unchanged and green (Pi operator case, no-copy,
conflicts, old-root warning, drift, rollback link/post-publish
phases — now also asserting journal deletion and no strays — pending,
scope gates, read-only planning, CLI plan/apply/drift/usage).

## Mutant table (shell `bash`, `-count=1`; scratch driver `/tmp/fc2r2_mutants.py`)

Tree verified restored after every mutant (`go build ./...` exit 0).

| Mutant | Site | Killer (failed) | Witness (passed) |
|---|---|---|---|
| R1 drop the plan requirement (`if false &&`) | `ApplyMigration` | `TestMigrateApplyRequiresPlan` (1) | Pi relink row (0) |
| R2 print after mutation | `ApplyMigration` | lib print-failure row (1) + `TestEnvMigratePrintBeforeWrite` (1) | Pi relink row (0) |
| R3 remove-then-create relink | `atomicRelink` | `TestMigrateSyscallFailureRollsBack` (1) | pending row (0) |
| R4 hash without marker identity | `migrationHash` | `TestMigrateMarkerDriftRefuses` (1) | Pi relink row (0) |
| M1r skip drift check (rev1 re-run) | `ApplyMigration` | `TestMigratePlanDriftRefuses` (1) | Pi relink row (0) |
| M2r copy instead of relink (rev1 re-run) | `executeMigrationOps` | `TestMigrateNoSecretCopies` (1) | pending row (0) |

## Evidence (shell `bash`, `set -o pipefail` where piped)

- `go vet ./internal/envprofile/ ./cmd/curator/`: exit 0.
- `go test ./internal/envprofile/ -run 'TestMigrate|TestSharedToIsolated|TestCredentialLink|TestDangling' -count=1`: ok 8.759s, exit 0.
- `go test ./internal/envprofile/ -count=1`: ok 341.899s, exit 0 (full package, pre-lint-rename tree; renames re-covered by the focused run + mutant witnesses above).
- `go test ./cmd/curator/ -run 'TestEnvMigrate|TestEnvResolveRepairNeedsMigration' -count=1 -v`: 6/6 PASS, 28.400s, exit 0.
- `golangci-lint run ./internal/envprofile/... ./cmd/curator/...`: 0 issues, exit 0.
- `git diff --check`: clean. Changed paths (11): `cmd/curator/{env.go,envmigrate.go,env_migrate_test.go}`,
  `internal/envprofile/{migrate.go,migrate_test.go,managed.go,credential_link_test.go}`,
  `internal/envregistry/envregistry.go`, `docs/{cli.md,troubleshooting.md}`, `CHANGELOG.md`.
- No new diagnostics: refusals reuse `environment_credential_conflict`
  with distinct asserted phrases. No `envmarker` changes (frozen v1).
  No credential-byte reads added (`grep ReadFile migrate.go` hits the
  marker + journal paths only).

## Bounds

- Windows rename-over-symlink and the oversized-target OS rejection run
  on the hosted lanes (unverifiable on this macOS host); failures there
  fail closed with rollback, never partial state.
- Skip-the-lock mutant not run (manager-lock contention rows from F-C1
  cover the lock; apply shares `beginOperation`).
- True SIGKILL crash is simulated via the `crash` fault point (journal
  + half-done links through the real write path), not a killed process.
- F-C3 owns the follow-through classification of the operator's live Pi
  case; rev1's full row/mutant tables are superseded by this note.

revision 3 = revision 2 unchanged; gate rerun after a Windows managerlock timing flake (internal/managerlock TestSubprocessExpectedAcquiredWithTinyDeadlineReportsBlocked, run 35702368558; leaf patch untouched, verified via git status --short).

## Revision 4 — recovery validation-before-write + exact temp ownership (rework-2)

Base: revision-3 tree (revision 2 unchanged) + binding rework-2. The straight-line path is
unchanged and accepted; this revision fixes the two P1 recovery defects only.

P1-1 — recovery validated before any write. `recoverMigrationJournal` now calls
`validateRecoveryInventory` before announcing or mutating: every journaled marker must read
back as the recorded prior or intended bytes, every journaled link must sit at its recorded
prior or intended target (relink: From or To; unlink: absent or From; unlink with empty From:
absent only), and every recorded temp must be absent or the owned symlink. Anything else
refuses naming the exact operator action, preserving the unknown state and the journal; only
proven transaction states reconcile. The plan-hash check after recovery is the re-check, so a
refused recovery never reaches it. Mutation-time re-checks close the TOCTOU: each link move
re-validates got-is-To (relink) before the journaled rename, and markers re-validate
(prior-or-new) immediately before `op.publish`. Sealed journals still delete only, with no
touch and no validation. Production call sites: `recoverMigrationJournal`,
`validateRecoveryInventory`, `validateRecoveryMarkers`, `reconcileLinkToPriorJournaled` in
`internal/envprofile/migrate.go`.

P1-2 — exact temp ownership, no glob deletes. Journal ops gain `temp` + `temp_target`
(omitempty, v1-compatible): `atomicRelinkJournaled` journals the temp intent before the
symlink, then symlink + rename; the caller clears Temp with Done in one journal write. A kill
between symlink and rename leaves an owned temp the next recovery cleans; a failed rename
removes its own temp best effort (a leftover stays journaled). `reconcileJournalLinks` removes
only each op's recorded Temp and only via `removeOwnedTemp` (absent = no-op; regular file,
directory, or foreign-target symlink = mismatch, never deleted). Foreign `.migrate-*.tmp`
paths that no journal records are ignored (preserved, apply succeeds). `readMigrationJournal`
rejects Temp/Marker-Live paths escaping the manager home or missing the temp pattern/target.
`sweepMigrationTemps` (glob + blind remove) is deleted. Recovery and in-process rollback share
the journaled path; `reconcileJournalLinks` now takes the journal pointer.

Journal format (rev4): `{version:1, plan, ops:[{profile,env,kind,path,from,to,done,
temp,temp_target}], markers:[{live,prior,new}], markers_done, complete}`. Atomic sequence per
relink: pick absent `.migrate-<rand>.tmp` in the link dir (skip existing without journaling) →
journal Temp intent → symlink(target, tmp) → rename(tmp, link) → Done + clear Temp in one
journal write. Recovery sequence: validate-all → announce → clean owned temps + converge links
in reverse (journaled move-backs) → re-validate markers → publish priors → delete journal.

## Row table (Revision 4 deltas; all through `ApplyMigration` on temp stores)

| Fix | Test | Entry | Asserts |
|---|---|---|---|
| P1-1 | `TestReviewerRecoveryMarkerDrift` (committed reviewer probe) | `ApplyMigration` crash → edit codex `lock_sha256` → retry | refuses (non-nil); edited marker bytes preserved (not overwritten) |
| P1-1 | `TestMigrateRecoveryRefusesUnexpectedTarget` | `ApplyMigration` crash → re-point pending link out of band → retry | `recovery found` + unexpected target + `out of band` + `re-run`; zero applied; unexpected link, half-done absent link, journal, snapshot all preserved |
| P1-2 | `TestReviewerRecoveryPreservesRegularTemp` (committed reviewer probe) | `ApplyMigration` crash → regular file at `pi/.migrate-operator-backup.tmp` → retry | apply succeeds; foreign regular file preserved |
| P1-2 | `TestMigrateRecoveryCleansOwnedTemp` | crafted journal (1 relink, Done=false, Temp+TempTarget) + owned symlink → `ApplyMigration` | recovery reported; owned temp deleted; Pi link migrated; journal deleted |
| kept | `TestMigrateInterruptedApplyRecovers` | crash → `PlanMigration` → retry | still recovers then migrates fully; announce precedes plan print |

Rev2 rows otherwise unchanged and green (including `TestMigrateFailedRelinkKeepsOldLink`,
updated to the journal-pointer `reconcileJournalLinks` signature, and the sealed/broken
journal rows).

## Mutant table (Revision 4; shell `bash`, `-count=1`; manual edit + `cp` restore)

Tree verified restored after every mutant (`go build ./...` exit 0).

| Mutant | Site | Killer (failed) | Witness (passed) |
|---|---|---|---|
| Glob cleanup restored (glob `.migrate-*.tmp` + blind remove in `reconcileJournalLinks`) | `reconcileJournalLinks` | `TestReviewerRecoveryPreservesRegularTemp` (1, `recovery deleted unrelated regular file`) | `TestMigrateRecoveryCleansOwnedTemp` (0) |
| Skip pre-recovery validation (`if false` around `validateRecoveryInventory`) | `recoverMigrationJournal` | `TestMigrateRecoveryRefusesUnexpectedTarget` (1, wrong error `cannot restore prior links` after partial writes vs `recovery found` before any write) | `TestMigrateInterruptedApplyRecovers` (0) |
| Unconditional recovery (skip pre-validation + marker re-check; original bug shape) | `recoverMigrationJournal` | `TestReviewerRecoveryMarkerDrift` (1, `recovery overwrote drifted marker before refusing: apply error=<nil>` + `stale plan accepted`) | `TestMigrateInterruptedApplyRecovers` (0) |

## Evidence (Revision 4; shell `bash`, `set -o pipefail` where piped)

- `go build ./...`: exit 0.
- `go vet ./internal/envprofile/ ./cmd/curator/`: exit 0.
- `go test ./internal/envprofile -run 'TestReviewerRecovery|TestMigrateRecovery|TestMigrateInterruptedApplyRecovers|TestMigrateFailedRelinkKeepsOldLink' -count=1 -timeout=120s -v`: 6/6 PASS, exit 0.
- `go test ./internal/envprofile -run 'TestMigrate|TestSharedToIsolated|TestCredentialLink|TestDangling|TestReviewer' -count=1 -timeout=300s`: ok 12.513s, exit 0.
- `go test ./cmd/curator -run 'TestEnvMigrate|TestEnvResolveRepairNeedsMigration|TestReviewerMigrationCLI' -count=1 -timeout=300s`: ok 35.098s, exit 0.
- `golangci-lint run ./internal/envprofile/... ./cmd/curator/...`: 0 issues, exit 0.
- `git diff --check`: clean. Changed paths (12 = 11 rev3 + 1): rev3 paths unchanged
  (`cmd/curator/{env.go,envmigrate.go,env_migrate_test.go}`,
  `internal/envprofile/{migrate.go,migrate_test.go,managed.go,credential_link_test.go}`,
  `internal/envregistry/envregistry.go`, `docs/{cli.md,troubleshooting.md}`, `CHANGELOG.md`)
  plus `internal/envprofile/reviewer_recovery_test.go` (committed reviewer probes). Rev4 edits
  touch only `internal/envprofile/{migrate.go,migrate_test.go,reviewer_recovery_test.go}`;
  docs/CHANGELOG need no change (recovery still "recovers to the prior state", now fail-closed).
- No credential-byte reads added (`grep ReadFile internal/envprofile/migrate.go` hits the managed
  marker, the journaled manager record (validation + re-check), and the fixed journal path only).
  No `envmarker` changes (frozen v1). No new diagnostics; recovery refusals name the out-of-band
  fix with `re-run`.

## Bounds (Revision 4)

- True SIGKILL still simulated via the `crash` fault point (journal + half-done links through the
  real write path), not a killed process; the symlink-to-rename kill window for owned temps is
  covered by the crafted-journal positive row, not by a real kill between the two syscalls.
- Operator edits racing the held lock between a re-check and its syscall (microsecond window)
  are narrowed by per-op + marker re-checks but not atomic-CAS; the reviewed probes edit between
  crash and retry, which the pre-write validation closes fully.
- Windows rename-over-symlink runs on the hosted lanes (unverifiable on this macOS host).
- Skip-the-lock mutant not run (unchanged from rev2: manager-lock contention rows from F-C1
  cover the lock; apply shares `beginOperation`).
