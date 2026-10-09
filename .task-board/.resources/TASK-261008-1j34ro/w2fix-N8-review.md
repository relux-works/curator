# Review note (TASK-261008-1j34ro, fix N8)
Check the fix against section N8 of .research/261004_inline-audit-wave-2.md: the invariant holds at the production boundary, the regression test fails without the fix (name the mutant you ran in a disposable clone; compile-only locally per R223, the hosted gate runs suites), the negative controls still pass, the CHANGELOG line is accurate. accept_cr with the live checklist if it holds.

Severity rule (tb-R226): only P0/P1 findings block acceptance; file P2 findings as separate BUG notes in your verdict (they do not hold the landing); assign severity yourself with a one-line reason.


## Addendum 2026-10-09 (revision 2)
Revision 1 was rejected with F1 (explicit empty --allow skips validation) and F2 (no proof that refusal precedes configuration access; WalkDir errors swallowed). Check both are fixed, with hosted green, hosted red (production fixes reverted) and the ordering mutant (configuration loading before the CLI parse) failing the no-Load assertion. The workspace was converged onto trunk d46f2b1f; CHANGELOG.md carries N7 and N8 entries. Severity rule tb-R226 applies (only P0/P1 block).


Verdict format: in the `verdict-findings` block, `severity` must be one of `bypass`, `regression`, `robustness` or `note` (the board refuses anything else and the next producer cannot start); put the tb-R226 level (P0/P1/P2) in `severity_reason`.


## Addendum 2026-10-09 08:10Z (revision 3)
Revision 3 republishes the rework-1 candidate (revision 2, never reviewed) after `worktree converge` onto trunk 3cb461b3 (N9 landed): CHANGELOG.md keeps both entries and internal/audit/audit.go is a clean three-way merge of N8 and N9 (equal to the orchestrator dry run); every other N8 file is byte-identical to revision 2. Review F1/F2 from the revision-1 verdict and the merged audit.go against current trunk.


## Addendum 2026-10-09 10:10Z (revision 4)
Revision 4 republishes the same rework after a second converge onto d7001974 (4mzun5 landed): CHANGELOG.md adds only the N8 entry; cmd/curator/main.go and internal/audit/audit.go are clean three-way merges of N8 with N9 and 4mzun5 (equal to the orchestrator dry runs); hashing.go and both test files are byte-identical. Review F1/F2 from the revision-1 verdict and the merged files against current trunk.
