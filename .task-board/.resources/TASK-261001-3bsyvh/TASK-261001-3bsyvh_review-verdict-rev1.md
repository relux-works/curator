# TASK-261001-3bsyvh — carrier identity review, rev1: ACCEPTED

Accepted delta: TASK-260728-rjxrgs rev1 (base 5432c85f, tree ac008990, snapshot refs/campaign/rjxrgs-rev1-20261001 = fe2b5f61).
Carrier: CR rev1 (base bab2433b, tree 1e1c5d83). The substantive review is TASK-260728-rjxrgs_review-verdict (rev1); this review covers identity only.

Evidence (all non-.task-board paths):
1. Path lists: 17 vs 17, byte-identical (cmp).
2. +/- line multisets: 1021 vs 1021 lines, byte-identical after sort (two-way, cmp). `git diff --numstat` per-path add/del counts are identical.
3. Blobs: for every one of the 17 paths, blob at the snapshot == blob at 1e1c5d83 (17/17 SAME).
4. Base movement 5432c85f -> bab2433b touches only internal/envprofile/named_absence_boundary_test.go and internal/pathboundary/named_absence_test.go. Neither is among the 17 (grep -Fxf overlap: none). external_lifecycle_conformance_test.go is absent from both bases and is a pure add.
5. `git merge-tree --write-tree --merge-base=5432c85f bab2433b <snapshot>` yields tree 1e1c5d83 = the candidate tree exactly (clean merge, no conflicts).
6. Worktree vs candidate tree: `git diff 1e1c5d83` is empty (0 lines).
7. Stray files: the diff against bab2433b has no results file, CHANGELOG, LOGBOOK or .task-board path. TASK-260728-rjxrgs_results.md is not in the tree. The added lines are byte-identical to the accepted ones (item 2), so nothing the accepted delta lacked was introduced, including any employer name.

Not re-run: the test gate. It is attached to the CR as green. Identity of the bytes is what this verdict covers.