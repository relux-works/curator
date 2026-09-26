# TASK-260922-cww1ov republish — revision 4 = revision 3 unchanged

revision 4 = revision 3 unchanged; republished under the new board binary so the validation evidence is tree-bound (validation_not_bound_to_tree).

## Tree identity (verified in worktree, no file changed)
- HEAD: 607770e0 (F-C2 replay) -> d3476cb6 (F-C1 replay) -> 48da2690 (trunk, rc.12 pin)
- Worktree tree via temp index (`git read-tree HEAD` + `git add -A` to GIT_INDEX_FILE): `768bacfa2c52a0906b80e233e8af14a275137ede`
- Matches reviewer-recorded CR candidate tree in `TASK-260922-cww1ov_review-verdict-rev3.md` (768bacfa).
- `git status --short` (uncommitted set, 7 modified + 1 untracked = 8 paths):
  - ` M .github/ci/platform-cases.tsv`
  - ` M CHANGELOG.md`
  - ` M cmd/curator/env_migrate_test.go`
  - ` M docs/troubleshooting.md`
  - ` M internal/envprofile/credential_link_test.go`
  - ` M internal/envprofile/migrate_test.go`
  - ` M internal/envprofile/reviewer_recovery_test.go`
  - `?? internal/envprofile/credential_production_test.go`
- `git diff HEAD --stat`: 8 files changed, 494 insertions(+), 4 deletions(-). No file changed by this run.

## Verdict status
- `TASK-260922-cww1ov_review-verdict-rev3.md` is ACCEPT (refresh-only, content acceptable); `accept_cr` was refused by the runtime with `validation_not_bound_to_tree`, routed to-dev for republish of the SAME tree. There is no changes-requested finding to answer, so no new regression test or mutant is due in this revision per `republish-tree-bound-evidence.md` (change NO file). The Review Round Brief's "rejection" premise does not match the read verdict file.
- Directives for RUN-260923-ff9a40: none recorded (`task-board spawn directives` returned no directives).
- Change NO file in this revision; handoff republishes the same tree so the landing suite runs once and publishes revision 4 with tree-bound evidence.
