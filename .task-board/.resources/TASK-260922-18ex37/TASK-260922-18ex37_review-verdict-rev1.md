# TASK-260922-18ex37 review verdict — CR rev1 (tree 26ccbcf2): CHANGES REQUESTED

## F1 (BLOCKING, per review note item 0)
The candidate tree has the stray root file `TASK-260922-3bbvrs_results.md` (checked with `git ls-tree -r --name-only 26ccbcf2`). Task documents are board resources and must not be repository files. Fix: delete it from the worktree and keep the content only as the board resource. No other stray root `TASK-*`/`BUG-*`, `test/` or `ledger/` paths were found.

## Checks that passed (bounded)
- Pin: `.github/workflows/ci.yml` SPEC_PIN is now dcc7f015e2d97edf2d52928afb6fd79ec8129e8b. `internal/crossconformance/testdata/draft-sources-v1/DRAFT_SOURCES_PIN` is now the same commit. The old pin dced9b83 no longer appears anywhere outside `.task-board`.
- `internal/scriptpolicy/conformance_test.go:63` classifies `executable_identity_cases` as refusedBeforeReached.

## Not done in this round
I stopped at the blocking finding. I did not independently re-run the tallies, gap-row attribution or hosted gate. The next reviewer cycle must do the full review of items 2–4 against the rev2 tree.
