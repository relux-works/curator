# TASK-260908-2kqa77 republish rev5 (tree-bound revalidation, no code change)

revision 5 = revision 4 unchanged; republished under the new board binary so the validation evidence is tree-bound (validation_not_bound_to_tree).

- Candidate tree OID (temp-index write-tree over HEAD + worktree path set): 4c24e01de711d512ce1acd2d71390e723a60089f — matches CR-4 candidate_tree_oid and the rev4 verdict re-derivation.
- Worktree path set vs rev4 patch: identical (M .github/ci/gate-selftest.sh, M .github/workflows/ci.yml, M CHANGELOG.md; new tools/goreleaserconfig/{gate,gate_test,wiring,wiring_test}.go + 3 testdata fixtures). No other paths touched.
- .github/ci/goreleaser-config-gate.sh stays deleted (verified absent); git diff --quiet .goreleaser.yml .github/workflows/release.yml clean (both untouched).
- Narrow re-verification this run (bash, macOS, set -o pipefail): go test -count=1 ./tools/goreleaserconfig/ -> EXIT=0 (ok); bash .github/ci/gate-selftest.sh -> EXIT=0 (gate-selftest: 198 passed, 0 failed).
- Content verdict stands per TASK-260908-2kqa77_review-verdict-rev4.md: ACCEPT-worthy, no findings; reviewer may accept rev5 on tree identity == 4c24e01d plus that verdict.
