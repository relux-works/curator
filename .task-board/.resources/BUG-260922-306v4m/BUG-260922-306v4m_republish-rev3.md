# BUG-260922-306v4m republish evidence — rev3 (tree-bound revalidation)

Republish of the accepted rev1/rev2 tree under the new board binary
(`validation_not_bound_to_tree` refused the rev1 `accept_cr` on content
already ACCEPTED; no code change requested).

revision 3 = revision 2 unchanged; republished under the new board binary so the validation evidence is tree-bound (validation_not_bound_to_tree).

## Tree identity (this run, RUN-260923-b4c8b5)

- `git status --short` path set: `M .github/ci/gate-selftest.sh`,
  `M .github/ci/install-rust-toolchain.sh`, `M CHANGELOG.md`,
  `M docs/self-hosted-runner-setup.md` — identical to both
  `BUG-260922-306v4m_change-request_rev1.patch` and
  `BUG-260922-306v4m_change-request_rev2.patch` path sets.
- `diff rev1.patch rev2.patch`: zero differences (exit 0).
- `cmp /tmp/worktree.diff (git diff) rev2.patch`: IDENTICAL.
- No repository file changed in this run; work left uncommitted on
  `task-board/story/STORY-260915-3w11un` as required. Scratch only under /tmp.

## Fresh validation on this tree (this run, real exit codes, bash 3.2)

- `bash -n .github/ci/install-rust-toolchain.sh`: exit 0.
- `bash -n .github/ci/gate-selftest.sh`: exit 0.
- `bash .github/ci/gate-selftest.sh`: 224 passed, 0 failed, exit 0.
- Fast fixtures (independent, /tmp/rustdiag scratch, channel 1.92.0):
  - absent-everywhere (set CARGO_HOME + non-exec prefix stub):
    exit 1, block marker present, 4 candidate lines
    (cargo absent, prefix exists-not-executable, /opt/homebrew absent,
    /usr/local absent), remedy sentence byte-equal and last line.
  - defaulted CARGO_HOME / unset HOMEBREW_PREFIX/RUNNER_*:
    exit 1, `CARGO_HOME=<home>/.cargo (defaulted)`, `RUNNER_NAME=unset`,
    `HOMEBREW_PREFIX=unset`, exactly 3 candidate lines, remedy last.
  - rustup only under CARGO_HOME/bin (fake executable rustup + stub
    rustc/cargo): exit 0, `rust-pin: using rustup at <fake>/bin/rustup`,
    GITHUB_PATH exactly one line (the shim dir), zero
    `runner diagnostics:` lines.
- No secret patterns: only named variables printed (RUNNER_NAME,
  RUNNER_OS, hostname, whoami, HOME, CARGO_HOME, HOMEBREW_PREFIX, PATH,
  per-candidate lines, bounded listings, command -v / type -a).

## Note on rev2 landing validation

Rev2 `*-validation.log` (run 35855672743) shows Gate self-test
ubuntu/macos/windows all success; the failure is Test (windows-latest)
go test + platform-case gate (go test exit 1, gate exit 0) — outside this
leaf's files. This rev3 republishes the identical tree so the new board
binary can bind validation evidence to it.

## Board writes in this run

- `set_status(BUG-260922-306v4m, status=development)` — ok (turn start).
- This outcome resource only; checklist untouched (already all done);
  no code, docs, or CHANGELOG edits.
