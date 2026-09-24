# TASK-260916-2ok97n review verdict — CR revision 6: CHANGES REQUESTED

Reviewer run, 2026-09-24. Candidate tree 13ea6286 on base 1511b345, checked out read-only in a throwaway clone at /tmp/rv6.

## Blocking finding F1: CI artifacts and task documents are committed into the product tree
The candidate adds 24 paths at the repository root. They total about 90.5k of the 107k inserted lines, and none of them is product code, ledger, or documentation:
- `test/` (go-test.json and go-test-served.json at 35,896 lines each, observed-cases.tsv, skips-observed.tsv, plan-*.txt, suite-plan.txt, platform-cases.txt). This is a downloaded Windows gate artifact: `test/suite-plan.txt:2` reads `root=D:\a\curator\curator/protocol-spec/...`.
- `ledger/` (case-inventory.tsv at 8,374 lines, go-list-{darwin,linux,windows}.{txt,err}, ledger-consistency.txt, exclusions.tsv). This is also a gate artifact directory. It is not `.github/ci/platform-cases.tsv`.
- `TASK-260916-2ok97n_results.md` and `TASK-260916-2ok97n_handoff-blocker.md`. These are task outcome documents and belong on the board as resources, not at the trunk root.
These look like a `gh run download` into the worktree followed by a broad add. If this lands, trunk carries roughly 90k lines of one-run CI output, including a Windows runner path, plus board-only documents.

Required rework: remove every path under `test/` and `ledger/`, and both root `TASK-*.md` files, from the candidate. Attach anything worth keeping as task-scoped outcome resources. Then re-handoff and confirm that the diff stat contains only product, test, docs, CHANGELOG, and `.github/ci` paths. Optionally, gitignore the artifact download location.

## Scope of this review
- Checked: the candidate path set; the conformance classification (`internal/scriptpolicy/conformance_test.go` no longer carries unreachable or not-implemented classifications; the one remaining "not implemented" string at :634 is a completeness fatal); and the diagnostic listing 14 board paths restored from base, which is not a failure.
- Not done this cycle: the full content review of note items 3–6 (the Windows cmd.exe hard-link bound, Job Object duplicated-handle binding, stream model R-e, and the mutants I was to re-apply myself). F1 blocks landing on its own. The next review cycle will do these checks on the cleaned revision. No acceptance is implied for them.
