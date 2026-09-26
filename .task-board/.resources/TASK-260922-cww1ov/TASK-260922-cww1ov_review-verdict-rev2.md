# TASK-260922-cww1ov revision 2 review

Verdict: ACCEPTED for producer integration. R1 and the required bookkeeping are resolved. No delivery-blocking finding in the revision-2 delta.

## Identity and scope

Candidate tree c099b8d28984f6d57487c465b79e1199f4a843e5; base 09b25ef6629b41455d91dcb252ab4e4034e12750. All 22 candidate changed paths byte-match the assigned workspace, including the untracked credential_production_test.go. Tests and attacks ran in an ignored git-archive extraction of that exact tree. The delta from revision 1 (1541ba6082ed3479186703d29dc2c5afa7140ebd) changes only migrate_test.go and the ledger comment. No production change, no accepted assertion deleted or weakened. Prior accepted coverage remains accepted under the binding rev2 note; it was not re-opened or exhaustively rerun.

## R1 verification

credentialHolders now walks the entire temporary manager home plus all native roots. It has no path allowlist for state, journals, backups or temporary files. Regular file content containing the exact fixture sentinel is counted; only the known native credential file may hold each sentinel. Non-secret lock/journal metadata is allowed by content, not by path. Existing native byte-identity and credential-scope snapshot assertions remain.

TestMigrateNoSecretCopies drives PlanMigration and ApplyMigration, then Resolve. Its new subtest drives an interrupted ApplyMigration, requires a standing parseable journal, scans the full home and reads journal bytes before retry, then checks recovery and scans again. The scan is demonstrably non-vacuous:

| Independent attack | Result |
| --- | --- |
| Original review-state-secret-copy: relink copies op.To bytes only into state/credential-backup | KILLED, exit 1 at migrate_test.go:369; extra Pi credential holder named |
| review-interrupted-state-copy: only the InjectFault path copies op.From bytes to state/transient-backup before the crash callback | KILLED, exit 1 at migrate_test.go:490; interrupted subtest names extra codex holder while the journal stands |

2/2 selected attacks killed by content assertions, not compilation or journal-parse failure. The second attack passes the parent success-path checks before failing the new interrupted subtest. Script and both logs attached. The disposable production file was restored and byte-compared with the candidate object; assigned source was never edited.

## Independent execution and hosted evidence

Shell zsh. `go test ./internal/envprofile -run '^TestMigrate(NoSecretCopies|SyscallFailureRollsBack|InterruptedApplyRecovers)$' -count=1 -timeout=120s -v`: exit 0, 4.579s. Includes the new interrupted subtest and the other caller of the expanded helper. Mutants use Python subprocess argv, Go -count=1, timeout 90s, and record real exit codes. Revision delta diff --check: exit 0.

Independently queried GitHub run 35722220821: success. Head e12b6d59792e49daea04ecedca829b43c93b9b9b resolves to the exact candidate tree. Ubuntu/macOS/Windows test lanes, Ubuntu/macOS race, lint, naming, interop and three gate self-test jobs succeeded. Candidate-suite and rose-air jobs skipped, not passing. Exact-tree hosted full validation is reused; no full landing-suite replay. URL: https://github.com/relux-works/curator/actions/runs/35722220821 . Raw job evidence attached.

## Bookkeeping and bounds

Independently counted the ledger additions: 46 total, 37 linux/darwin/windows with Windows host-capability tolerance, 5 all-platform without tolerance, 4 Unix-only with Windows host-capability tolerance. Revision 2 changes no row or skip vocabulary and correctly says four ENOTDIR rows. TASK-260922-cww1ov_results-rev2.md is complete, contains coverage/mutant tables and corrected counts, and ends with the handoff section rather than a truncation marker. The old results resource is historical and superseded by this named rev2 resource.

Bounds: exact plaintext fixture bytes in regular files at the observed success/interruption/recovery states; not arbitrary encoded/fragmented secrets or every intermediate instruction. Symlinks are excluded from holder counting, with legitimate native targets separately scanned. Crash injection is not a real SIGKILL. The unchanged credential-scope snapshot still excludes manager state; the broader content scanner supplies the state no-copy proof. No journal-field content mutant is claimed; the independent interrupted-state attack directly proves the new crash-state scan executes. Full package/build/lint/platform evidence comes from the exact-tree hosted gate rather than a new local full run. Earlier accepted API/CLI/refusal rows and campaign results are reused, not claimed independently rerun here.

Run goal queried: not goal-bound; no directives. Review logbook: the prior state-copy survivor is now killed, and the interrupted-state probe also fails as intended; no new defect or host-state change. No control-root LOGBOOK edit. Attach evidence before accept_cr(revision=2); checkpoint/integration and delivery closure remain with the bound producer.
