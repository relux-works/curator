# TASK-260906-2b3nar — revision 5 republish note (tree-bound revalidation, no content change)

revision 5 = revision 4 unchanged; republished under the new board binary so the validation evidence is tree-bound (validation_not_bound_to_tree).

## Candidate identity (verified in worktree, HEAD 48da2690)

- `git status --short` shows exactly the accepted rev4 path set:
  `M .github/ci/platform-cases.tsv`, `M CHANGELOG.md`, `M internal/gitops/gitops.go`,
  `?? internal/gitops/dirfold_test.go` (tracked diff 91 insertions / 9 deletions across 3 files; untracked test 134 lines).
- Candidate tree (temp index over HEAD + `add -A`) = `0e5edef1d4c9eaac4525a9e88b3c765f520e4f5e`, identical to the rev4 reviewer's recorded tree.
- No file changed in this run (republish only).

## Narrow revalidation on this tree (shell: bash, real exit codes)

- `go test ./internal/gitops ./internal/snapshot -count=1` — exit 0 (gitops ok 42.8s; snapshot ok 4.4s).
- Four fold rows `go test ./internal/gitops -count=1 -v -run 'TestExtractRefusesDirectoryComponentFold|TestExtractAdmitsDirectoryComponentFoldWhenCaseSensitive|TestExtractRefusesFileDirectoryFold|TestExtractRefusesNestedDirectoryComponentFold'` — exit 0: RefusesDirectoryComponentFold PASS, AdmitsDirectoryComponentFoldWhenCaseSensitive SKIP (named reason: case-folding host temp dir, same bound as rev1/rev4), RefusesFileDirectoryFold PASS, RefusesNestedDirectoryComponentFold PASS.
- Prior verdict evidence reused: rev4 ACCEPT (`TASK-260906-2b3nar_review-verdict-rev4.md`); mutants m0–m3 stand killed on byte-identical code (rev1 table in `TASK-260906-2b3nar_results.md`).

No LOGBOOK.md edit. Ready for review.
