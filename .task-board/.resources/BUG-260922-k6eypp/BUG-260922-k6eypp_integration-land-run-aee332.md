# BUG-260922-k6eypp — integration landing preconditions (RUN-260928-aee332)

Binding: accepted CR-BUG-260922-k6eypp-3 revision 3, integration run. Per the integration assignment this run executes no `worktree integrate`, no `handoff`, no `set_status`, and changes no repo file; the runner performs the bound landing synchronously after this turn.

## Board preconditions (read-only, exit 0)
- `task-board q 'get(BUG-260922-k6eypp)'` → status `integrating` (exit 0).
- `spawn status` → RUN-260928-aee332 running/executing; `spawn directives` → no directives (exit 0).
- Outcome resources present: `BUG-260922-k6eypp_change-request_rev3.patch`, `BUG-260922-k6eypp_change-request_rev3-validation.log` (remote gate run 36371182493, all lanes success incl. Gate self-test windows-latest, exit 0), `BUG-260922-k6eypp_review-verdict-rev3.md` (rev3 ACCEPTED), prior `BUG-260922-k6eypp_integration-land.md` (RUN-260928-21bac1).

## Worktree preconditions (read-only, exit 0)
- Branch: `task-board/story/STORY-260923-11vn9k`.
- `git status --short` → clean (no output); `git status` → "nothing to commit, working tree clean".
- HEAD `1dd314e3` is the landed rev3 commit ("Change-Request: CR-BUG-260922-k6eypp-3 revision 3"); `git diff 86552087..HEAD --stat` → 3 files, 196 insertions, 32 deletions (matches review base 86552087 and candidate tree).
- Spot-checked read-only: installer resolves `rustup_proxy_dir` from real path of `rustup_bin` plus `rustup which --toolchain "$channel" rustc` fallback, prepends to PATH/GITHUB_PATH, still fails closed, pinned channel unchanged; selftest Homebrew block guarded by `MINGW*|MSYS*|CYGWIN*` with explicit `skip ... POSIX-only` on Windows.

## Integration command
- NOT executed by this run per binding. Bound command for the runner/orchestrator (from k6eypp-integrate-land.md):
  `task-board worktree integrate STORY-260923-11vn9k --cr BUG-260922-k6eypp --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)"`
- No refusal encountered (no attempt made). No repo writes, no board status writes by this run.

## Handoff
- Fresh outcome evidence: this file.
- Board left at `integrating`; `done` is reserved for the integration transaction.
