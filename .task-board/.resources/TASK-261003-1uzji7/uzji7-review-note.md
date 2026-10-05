# Review note — TASK-261003-1uzji7 rc.4 atomic v1→v2 migration + writer flip (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). Release-critical for rc.4. Read the brief (uzji7-brief.md) and 1foyf3-blocker-evidence.md: flipping the writer alone made ApplyMigration relabel v1 identities as v2.
Verify, with real exit codes and GOFLAGS=-work (targeted packages only; the mini is under tb-R136 load limits):
1. The migration REHASHES; it never relabels. Construct a v1-identified profile or home, migrate it, and check that the new identity equals a fresh v2 computation of the same bytes and differs from the v1 digest string. Then mutate the code to relabel and confirm a test kills that mutant.
2. Atomicity and rollback: interrupt mid-migration (a fault-injection seam, if one exists) and check that the state is either fully v1 or fully v2; rollback restores v1.
3. Readers accept v1 and v2 during the transition. Legacy-reader, NUL and mismatched-version negatives exist and pass. TestRC14MigrationRehashesLegacyIdentities is present and passes.
4. EnableV2Writers=true. Every WriteVersion() call site the design lists is covered by a real-entry test, or bounded with a reason.
5. Conformance: the rc.14 byte-exact-snapshot gap row was removed only if that case now passes; exact counts are consistent.
6. No CHANGELOG.md or LOGBOOK.md edits. Cite the hosted CR validation gate result, or say it has not run.
accept_cr if all hold; otherwise changes requested with concrete findings.

## HOSTED-EVIDENCE MODE (tb-keeper desk #52 decision, 2026-10-04; binding, overrides the "rerun" steps above)
The Mac mini has host-wide exec stalls. Do NOT run go build, go test, go vet or lint locally.
- Verify by reading code and the diff.
- Use the HOSTED evidence: the CR validation gate run on GitHub. Get its run id from the CR validation resource or `gh run list`. Download its artifacts with `gh run download <id> -R relux-works/curator` run from the worktree; go-test JSON or log artifacts show which tests ran and passed.
- Red-first: show from the diff that each new regression asserts the defect, i.e. it would fail on the base logic. Explain the assertion against the base code; do not execute it.
- Mutants: argue kill or survive by reading. Do not execute.
- If the hosted gate did not run the new tests, or is red, that is a finding (changes requested), not something to fix locally.
Keep board commands to the minimum needed: one outcome resource and the verdict.
