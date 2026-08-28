# TASK-260819-2tr2rh release-target guard developer handoff

## Outcome

- Recovery branch: `fix/rc8-release-merge-policy`
- Pull request: https://github.com/relux-works/curator-spec/pull/21
- Exact head: `8d2e61f6c5b46a802e9aca690275e126bcfb9d9a`
- Signed commits: `3e3bf2fc6c3611bcbf33233325970cc4cc5b129a`, `8d2e61f6c5b46a802e9aca690275e126bcfb9d9a`
- GitHub verification for the exact head: `verified=true`, `reason=valid`
- Repository merge policy now permits squash only: squash enabled, rebase disabled, merge commits disabled.
- PR #21 adds a tested maintainer merge-policy preflight and a main-branch post-integration release-provenance job. The latter rejects unsigned release targets using `tools/verify_release_commit.py` before tagging.
- `RELEASE.md` now requires the reviewed GitHub squash path and explicitly prohibits rebase-and-merge for release targets.
- rc.7 history/tag were not changed. No rc.8 tag or release was created. No verified implementation or platform claim was added.

## Green gates (exit 0)

- Live merge-policy verifier after repository setting change.
- Clean-worktree schema/vector validation: 49 schemas and 471 vector files.
- Clean-worktree Python suite: 91 tests.
- `go test ./tools/...`.
- `go run ./tools/generate-vectors -root .`.
- Authoritative regeneration diff, `git diff --check`, Go formatting, and clean-tree checks.
- rc.8 release gate at exact PR head `8d2e61f6c5b46a802e9aca690275e126bcfb9d9a`.
- Local `git verify-commit HEAD` and GitHub commit verification for the exact PR head.
- Specification CI run 32201136998: success across Ubuntu, macOS, Windows, Formatting, and Links; the PR-only Release target provenance job is intentionally skipped.
- Implementation conformance run 32201136935: success on Ubuntu, macOS, and Windows.

## Truthful red / environment evidence

- Initial `python` test command exited 127 because this shell exposes `python3` and the project `.venv/bin/python`, not `python`; rerun through the project virtual environment passed.
- Live policy verifier exited 1 before the setting change because rebase merging was enabled; rerun after disabling rebase passed.
- Local-worktree release gate exited 1 because the mandated untracked `task-board.config.json` makes that checkout non-clean; the exact-commit detached clean-worktree gate passed.
- First PR Formatting run 32200983827 exited nonzero because the Actions token's read permissions omit REST merge-setting fields. The workflow was corrected without adding a broad maintainer token: policy verification remains a maintainer preflight, while ordinary-token post-merge commit provenance is enforced on main. Replacement runs are green.
- Optional PyYAML parse command exited 1 because PyYAML is not installed in the project virtual environment; Ruby YAML parsing exited 0, and GitHub parsed and ran the corrected workflow successfully.

## Required reviewer / publication continuation

Independently review PR #21 and this permission-boundary adjustment. If accepted, squash-merge through GitHub (rebase is disabled), verify the resulting GitHub-signed main commit and the new Release target provenance job, rerun the clean rc.8 release gate, then create and push the signed annotated `v1.0.0-rc.8` tag and verify prerelease assets and immutable evidence. Do not tag the unsigned `792c53c1887ce02b4b9c1d3954312c919ffb62ef` commit.
