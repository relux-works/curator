# BUG-260918-3u6qqq revision 1 — independent review run 2

Verdict: ACCEPT. Reviewer run RUN-260918-c1777c. Handoff explicitly binds CR-BUG-260918-3u6qqq-1 revision 1. `task-board spawn goal "$TASK_BOARD_RUN_ID"` reports no active goal (not goal-bound).

Reviewed base a4a3bcee268b39edffc2f84de37bc9b869c1dd57 → candidate tree 62e0d5a6e93922631e52d68915e997e8f45b5fbe. Only the three scoped files differ. No product code changed by reviewer.

Installer lines 38–45 compute CARGO_HOME independently of RUSTUP_HOME, require and append GITHUB_PATH, export PATH, then resolve and invoke rustup. Existing channel parsing, install arguments, post-install checks and diagnostics unchanged. Missing rustup fails with the docs-named note. Docs accurately describe launchd and no shell-profile sourcing; workflow topology unchanged.

Independent verification: extracted unchanged harness helpers and the contiguous rust-pin-guard/install-rust-toolchain sections from candidate gate-selftest.sh, with HERE set to the verified worktree. Bash ran actual production scripts with fake rustup, without network. 19/19 assertions pass, exit 0. An initial extraction selected an earlier end marker and executed zero assertions; discarded and corrected using end-marker search after section start. Only the corrected 19-assertion run counts.

Narrowing attack: separate scratch copy of installer moves command -v rustup ahead of cargo_home/PATH setup, preserving other behavior. Same rows produce 16 pass/3 fail, exit 1: CARGO_HOME-only success, invocation and GITHUB_PATH assertions fail. Required narrowing mutant killed 1/1. Production files were never mutated.

Hosted evidence independently read using gh: https://github.com/relux-works/curator/actions/runs/35300299336 succeeded; head 1eeb7433fd9bbbdff443497b90a3bf273d160dda resolves via GitHub git/commits API to exact candidate tree 62e0d5a6e93922631e52d68915e997e8f45b5fbe. All 11 executed jobs succeeded, including Linux/macOS/Windows tests and gate self-tests, lint and race jobs. Full landing suite was not rerun locally. Rose-air and candidate-suite jobs skipped; real rose-air verification remains after landing, not claimed passing here.

Shellcheck unavailable; bash syntax and diff whitespace checks pass. No architecture mismatch or review findings. Local assertions cover fake-rustup behavior, not a real install on this host.

## Local check exit codes
['git', 'diff', '--exit-code', '62e0d5a6e93922631e52d68915e997e8f45b5fbe', '--', '.', ':!.task-board']: exit 0

['bash', '-n', '.github/ci/gate-selftest.sh', '.github/ci/install-rust-toolchain.sh']: exit 0

['git', 'diff', '--check']: exit 0

## Narrow test output
```text
=== rust-pin-guard.sh: the committed toolchain file cannot drift ===
ok    the committed rust-toolchain.toml agrees with SupportedRustToolchainVersion
ok    a bumped toolchain file is rejected
ok    a changed supported value is rejected
ok    a missing toolchain file fails closed
ok    a toolchain file with no channel fails closed
ok    a floating channel is rejected
ok    a non-minimal profile is rejected
ok    an unreadable Go source fails closed
ok    a Go source with no declaration fails closed

=== install-rust-toolchain.sh: rustup installs the filed channel, shims lead PATH ===
ok    the installer installs exactly the filed channel
ok    the install invocation names the filed channel
ok    the shim directory is prepended for the rest of the lane
ok    the installer finds rustup under CARGO_HOME/bin without PATH help
ok    the CARGO_HOME-only install names the filed channel
ok    the CARGO_HOME bin dir is recorded for the rest of the lane
ok    a runner without rustup fails
ok    the failure names the runner-setup note
ok    the installer names no release; it reads the channel from the file
ok    the installer reads its channel from rust-toolchain.toml

PASS=19 FAIL=0 SKIPPED=0
EXIT=0
```
## Mutant output
```text
=== rust-pin-guard.sh: the committed toolchain file cannot drift ===
ok    the committed rust-toolchain.toml agrees with SupportedRustToolchainVersion
ok    a bumped toolchain file is rejected
ok    a changed supported value is rejected
ok    a missing toolchain file fails closed
ok    a toolchain file with no channel fails closed
ok    a floating channel is rejected
ok    a non-minimal profile is rejected
ok    an unreadable Go source fails closed
ok    a Go source with no declaration fails closed

=== install-rust-toolchain.sh: rustup installs the filed channel, shims lead PATH ===
ok    the installer installs exactly the filed channel
ok    the install invocation names the filed channel
ok    the shim directory is prepended for the rest of the lane
FAIL  the installer finds rustup under CARGO_HOME/bin without PATH help
      exit 1, want 0; output: rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane 
FAIL  the CARGO_HOME-only install names the filed channel
      expected to find: toolchain install 1.92.0 --profile minimal
FAIL  the CARGO_HOME bin dir is recorded for the rest of the lane
      expected to find: /var/folders/xk/2m1x7tqd61z26cmdwvdz7_v40000gn/T//gate-selftest.ua52Zd/fake-cargo-only/bin
ok    a runner without rustup fails
ok    the failure names the runner-setup note
ok    the installer names no release; it reads the channel from the file
ok    the installer reads its channel from rust-toolchain.toml

PASS=16 FAIL=3 SKIPPED=0
EXIT=1
```
