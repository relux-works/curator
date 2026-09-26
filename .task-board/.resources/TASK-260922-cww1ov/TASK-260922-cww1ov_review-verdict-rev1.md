# TASK-260922-cww1ov revision 1 review

Verdict: CHANGES_REQUESTED; route to to-dev. One coverage defect, no production defect demonstrated and no human decision needed.

## Required rework

**R1 (P2): the no-copy production row cannot detect credential copies into manager state.** `internal/envprofile/migrate_test.go:114` limits credentialHolders to environments, profiles and native roots. The separate snapshot helper at :55 excludes state too. Consequently TestMigrateNoSecretCopies (:278) cannot prove the required no-secret-copy property for the journal/state area used by migration. This was already a stated bound in F-C2 reviews; this final coverage leaf leaves it open.

Executable reproduction: attached TASK-260922-cww1ov_review-mutants.py, third mutation (`review-state-secret-copy`). In a disposable exact-candidate extraction, executeMigrationOps's relink arm reads op.To when it exists and writes those bytes to filepath.Join(filepath.Dir(migrationJournalPath(req.Home)), "credential-backup") with mode 0600, propagating write errors. `go test ./internal/envprofile -run '^TestMigrateNoSecretCopies$' -count=1 -timeout=90s -v` still exits 0. The fixture supplies a real native target, and migration has created the state directory/journal before this arm. Both native byte-identity assertions remain true while a new secret copy escapes both scans.

Fix the test, not production: include the entire temporary manager home (especially state, journal, backup/temp paths) in credential-content scanning, while retaining native roots and permitting legitimate non-secret lock/journal metadata changes. Add a narrowing mutant copying only into manager state and require this named production-entry row to kill it. If journal transient contents are claimed, inspect them during the existing interrupted-apply hook as well; a final post-success scan cannot see a deleted journal. Keep scope and residual bounds explicit. Re-run focused tests/mutants and publish fresh green hosted evidence for the revised candidate.

Also correct the evidence bookkeeping during rework: results.md says lane shapes are 39/4/3; actual ledger additions are 37 all-platform rows with Windows host-capability tolerance, 5 all-platform/no-tolerance rows, and 4 Unix-only rows (46 total). `.github/ci/platform-cases.tsv:475` says three ENOTDIR rows although there are four. The attached producer results file ends with a literal truncation marker in Bounds; replace it with a complete resource using file attachment. These documentation issues are secondary to R1.

## Exact identity and hosted evidence

Reviewed candidate 1541ba6082ed3479186703d29dc2c5afa7140ebd, base 09b25ef6629b41455d91dcb252ab4e4034e12750. All 22 changed paths byte-match the assigned workspace, including the untracked new credential_production_test.go. Relative to the F-C2 checkpoint, this leaf changes only tests, ledger, CHANGELOG and troubleshooting; zero production changes. Existing reviewer assertions remain; the unrecorded-link assertion is strengthened.

Independently queried GitHub run 35716709852: success, head 10ad9a1bc5b5bebb7b50bce75b6be36b5ab4b04c resolves to exactly the candidate tree. Hosted Ubuntu/macOS/Windows test jobs, Ubuntu/macOS race, lint, vet/ledger steps and gate self-tests succeeded. Rose-air and candidate-suite skipped, not passing. Full landing-suite evidence reused without replay. Gate: https://github.com/relux-works/curator/actions/runs/35716709852

## Independent execution (zsh; real exit codes)

- `go test ./cmd/curator -run 'TestEnv(Migrate|ResolveRepair)' -count=1 -timeout=180s`: exit 0, 138.447s.
- `go test ./internal/envprofile -run 'Test(CredentialLink|SharedToIsolated|StaleCredential|Dangling|Codex|PiProvision|Reviewer|Migrate)' -count=1 -timeout=120s -v`: exit 0, 73.759s; complete log attached. Covers both hazards, both dangling states, Codex admission, repair_failed, migration lock, plan/drift, rollback/recovery and no-silent-repair.
- Initial full package `go test ./internal/envprofile/... -count=1 -timeout=180s`: exit 1, timeout at 180.823s while TestMachineUseSkipsScopedAdapter was running. Not claimed passing; replaced for independent review scope by the bounded focused run above. Full package hosted success is separately established by exact-tree CI.
- `git diff --check`: exit 0. All 46 added ledger rows resolve to named committed test functions; no skip vocabulary change. Windows privilege probes follow the established host-capability pattern, with Unix skips gate-fatal.

## Independent mutation results

Mutations ran only in ignored disposable candidate code; production files in the assigned workspace were not edited. Disposable managed.go and migrate.go restored and byte-compared to candidate objects. Script and all logs attached.

| Mutant | Named result | Control |
| --- | --- | --- |
| Refuse regular files only when nonempty; replace empty files with the link | TestCredentialLinkRegularFileRefuses/empty FAIL, exit 1, admitted nil error | non-empty PASS |
| Recovery inventory validates Done operations only, omitting pending links | TestMigrateRecoveryRefusesUnexpectedTarget FAIL, exit 1, wrong late refusal instead of pre-write inventory refusal | TestMigrateInterruptedApplyRecovers PASS |
| Relink copies existing native bytes only into manager state/credential-backup | TestMigrateNoSecretCopies PASS, exit 0: SURVIVOR | native byte identity unchanged |

2/3 selected mutations killed by behavioral assertions; 1/3 survives and establishes R1. No build-failure kills. Producer's 47-mutation campaign was inspected but not independently replayed wholesale; its reported kills do not cover the newly demonstrated state-copy shape.

## Coverage and bounds

Committed API/CLI rows genuinely call Resolve, StatusOf, PlanMigration, ApplyMigration and CLI run(). Temporal print-before-write uses a writer hook, not final-output containment. Recovery unknown-target and marker drift, owned temp cleanup and foreign regular-temp preservation are driven at ApplyMigration. Both former 0017 hazards and Codex literal/unknown/malformed admission are exercised. The no-copy row proves native byte preservation and no copies only in its scanned roots, not the required broader claim.

TestMigrateFailedRelinkKeepsOldLink remains helper-direct and excluded from the 46 ledger production rows. Crash tests use InjectFault, not actual SIGKILL; owned temp setup is crafted. No cross-platform result inferred from local Darwin beyond verified hosted evidence. No product changes requested absent a real defect. Read F-C1 results and rev1/rev4 verdicts, F-C2 results and rev1/rev3/rev4 verdicts. Run goal queried: not goal-bound; no directives. No control-root LOGBOOK edit. Findings also attached as task-scoped review logbook; fresh producer revision and reviewer cycle required.
