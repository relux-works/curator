# Review note for TASK-260924-4mzun5 (reviewer)
Check the decision against the landed spec (curator-spec 7eaeb73f: core §4.4, draft-sources-v2): the legacy lane must either record the directory where the spec says or refuse with the spec diagnostic; production-entry rows (positive + negative) and the `?`/`[` glob vectors on the dependency path must exist and must fail when the behaviour is removed (name the mutant you ran, in a disposable clone, compile-only locally per R223; the hosted gate is the suite arbiter). accept_cr with the live checklist if it holds.

## Revision 2
Verify F1–F5 of the revision-1 verdict are fixed at the production entries (your own probes, in a disposable clone, compile-only locally per R223; hosted gate for suites).

## Revision 3
Verify F4 (explicit tag change accepted, same-tag move refused) and F6 (status not current when the declared ref changed at the same commit) at project and global production entries.

Severity rule (tb-R226): only P0/P1 findings block acceptance; file P2 findings as separate BUG notes in your verdict (they do not hold the landing); assign severity yourself with a one-line reason.


## Addendum 2026-10-09
The revision under review is the rework-3 candidate carried onto trunk 97822069 by `worktree converge` (trunk had advanced on CHANGELOG.md through the N6 and N7 landings). The orchestrator verified that every changed file except CHANGELOG.md is byte-identical to revision 3 and that CHANGELOG.md is trunk plus the 4mzun5 entry. Review the candidate as a whole against current trunk; severity rule tb-R226 applies (only P0/P1 block).


## Addendum 2026-10-09 (revision 5, tb-R226 loop breaker)
Revision 4 was rejected with one P1 (F4: an unrelated live tag on the prior commit suppresses same-tag movement refusal under StrictTags). Revision 5 comes from one gpt-6-astra max producer pass. Check F4 against the rev-4 verdict's required rework (validated previous declared ref, no live-tag inference, alias probes permanent on project and global entries, narrowing mutant killed). The P2 note is BUG-261009-2s6t3y (separate; not a reason to hold acceptance). tb-R226: only P0/P1 block.


Verdict format: in the `verdict-findings` block, `severity` must be one of `bypass`, `regression`, `robustness` or `note` (the board refuses anything else and the next producer cannot start); put the tb-R226 level (P0/P1/P2) in `severity_reason`.


Note 07:10Z: before the rework pass the workspace was converged onto trunk 22cf8b4f (N9 landed): CHANGELOG.md and internal/audit/audit.go merged cleanly with N9; review the combination against current trunk.
