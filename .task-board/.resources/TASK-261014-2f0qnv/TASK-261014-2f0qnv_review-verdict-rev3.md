# TASK-261014-2f0qnv review verdict, CR rev3 (reviewer B, deciding): ACCEPTED

Candidate tree 265d93f4093227dfa6e1a05ea49c72dbb5a3551e (recomputed from the worktree: identical). Patch sha256 68073e44... matches. Spec 6648e18 vectors file byte-identical to the checked-in testdata copy. Hosted gate for rev3 (run 38570739007) exited 0, all required jobs green; rev3 validation log read to its last line.

## findings
[]  (no blocking findings)

## Surface table (each row attacked via public entry Resolve / cmdEnvResolve / StatusOf, in a scratch copy; worktree untouched)
| Row | Result | Attacks |
| --- | --- | --- |
| Refusal before repair lock / write / rebuild | held | M5 refusal moved after lock acquisition -> TestPreserveLockMigrationRefusalPrecedesBusyLock + 3 vectors fail; M4 refusal moved after store rebuild -> TestPreserveLockRefusesBeforeStoreRebuild + vectors fail; M2 refusal dropped -> fails |
| Diagnostic / operator command text | held | M3 command gets `--preserve-lock` appended -> vectors fail (suffix check) |
| Lock-changed race -> environment_repair_failed | held | M1 re-check dropped -> RaceRefusesBeforeJournalRecovery, RevalidationHoldsRepairLock, vectors fail; M6 in-lock check removed before journal recovery -> RaceRefusesBeforeJournalRecovery fails |
| `--preserve-lock` without `--repair` usage | held | M7 CLI guard removed -> TestEnvResolvePreserveLockRequiresRepair fails; API guard also present in Resolve |
| Current-home lock-free path / plain --repair migration / status read-only | held | driven by vectors v2-current-lock-free, plain-repair-migrates, v1-current-status-read-only; CLI test covers refuse -> status -> plain migrate on one home |
| Live-home / private identifiers | held | all tests use temp homes; grep for local paths in README/tests clean |

AC coverage: 10 of 10 vector rows driven through Resolve/StatusOf/cmdEnvResolve (map in gaps file).

## Notes (non-blocking)
- `--repair --dry-run --preserve-lock` on a v1 lock refuses with the migration diagnostic (pre-check precedes the dry-run branch). Spec silent on dry-run; defensible, not a finding.
- Race tests inject a same-framing replacement before acquisition; writers after acquisition are outside the spec'd window (declared bound).
- Local `cmd/curator` TestEnvStatus* failures reproduce identically on the unmodified base (host claude_code tool-version drift); environmental, not caused by this CR. Hosted CI is green.
- Side effect to disclose: I ran `git add -N` on the worktree once and reverted it with a path-scoped `git reset`; no content changed.
- Free hunt: no further defects found.
