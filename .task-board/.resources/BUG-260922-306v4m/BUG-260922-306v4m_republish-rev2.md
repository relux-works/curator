# BUG-260922-306v4m republish evidence — rev2 (tree-bound revalidation)

Republish of the accepted rev1 tree under the new board binary
(`validation_not_bound_to_tree` refused the rev1 `accept_cr` on content
already ACCEPTED; no code change requested).

## Tree identity

- `git status --short` path set: `M .github/ci/gate-selftest.sh`,
  `M .github/ci/install-rust-toolchain.sh`, `M CHANGELOG.md`,
  `M docs/self-hosted-runner-setup.md` — identical to the
  `BUG-260922-306v4m_change-request_rev1.patch` path set.
- `diff /tmp/rev1.patch` (downloaded board resource) vs `git diff`
  of this worktree: IDENTICAL (zero differences).
- No repository file changed in this run; work left uncommitted on
  `task-board/story/STORY-260915-3w11un` as required.

## Fresh validation on this tree (this run, real exit codes)

- `bash .github/ci/gate-selftest.sh`: 224 passed, 0 failed, exit 0.
- `bash -n` on both edited scripts: pass.
- No secret patterns in the install-script diff (only named variables printed).

## Board writes in this run

- `set_status(BUG-260922-306v4m, status=development)` — ok.
- `BUG-260922-306v4m_results.md` updated with the line:
  "revision 2 = revision 1 unchanged; republished under the new board
  binary so the validation evidence is tree-bound
  (validation_not_bound_to_tree)".
