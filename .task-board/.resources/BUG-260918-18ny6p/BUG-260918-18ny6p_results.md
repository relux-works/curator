# BUG-260918-18ny6p results — Homebrew rustup resolution in install-rust-toolchain.sh

Worktree: .temp/STORY-260915-3w11un/worktree, branch task-board/story/STORY-260915-3w11un, base ee3a564. Shell for all commands: bash via zsh Bash tool, set -o pipefail where pipes were used; gate commands run standalone with output redirected to a file (no tee).

## Change
- .github/ci/install-rust-toolchain.sh: after prepending $CARGO_HOME/bin, rustup is resolved in order PATH → $CARGO_HOME/bin/rustup → $HOMEBREW_PREFIX/bin/rustup (if set) → /opt/homebrew/bin/rustup → /usr/local/bin/rustup. The directory of the first executable found is prepended to PATH and appended to GITHUB_PATH (CARGO_HOME/bin stays on both for the proxies) before the rustup call; the script logs "rust-pin: using rustup at <path>". Absence everywhere fails closed with the unchanged docs-named message. Hosted lanes: rustup is on PATH, so the first branch fires and behaviour is unchanged.
- .github/ci/gate-selftest.sh: new row "the installer finds rustup under the Homebrew prefix without PATH help" (HOMEBREW_PREFIX=$WORK/fake-brew, PATH filtered of rustup, CARGO_HOME/bin holds only fake rustc/cargo proxies) plus 3 assert_contains rows (install invocation, fake-brew/bin in GITHUB_PATH, CARGO_HOME/bin still in GITHUB_PATH). Absent-everywhere row now also sets HOMEBREW_PREFIX to an empty prefix; it is reported through the harness `skip` helper (not ok) on a host that ships rustup under /opt/homebrew/bin or /usr/local/bin, since the script probes those unconditionally. On this host and on the hosted images that row runs (named failure).
- docs/self-hosted-runner-setup.md: rustup may be the rustup.rs or the Homebrew install; launchd service PATH lacks /opt/homebrew/bin; the lane probes PATH, ~/.cargo/bin, then the Homebrew prefixes itself; the runner .path file may carry /opt/homebrew/bin as an alternative, not required.

## Evidence (real exit codes)
| Command | Exit |
|---|---|
| bash -n .github/ci/install-rust-toolchain.sh | 0 |
| bash -n .github/ci/gate-selftest.sh | 0 |
| bash .github/ci/gate-selftest.sh (full, log /tmp/gate-selftest-18ny6p-2.log) — 187 passed, 0 failed | 0 |
| install-rust-toolchain section only (lines 942–1056 + harness prelude, /tmp/irs-section.sh) — 14 passed, 0 failed, 0 skipped | 0 |
| direct: fake rustup only under HOMEBREW_PREFIX/bin, proxies under CARGO_HOME/bin, PATH filtered → GITHUB_PATH = cargo/bin, brew/bin; log = toolchain install 1.92.0 --profile minimal | 0 |
| direct: absent everywhere (empty CARGO_HOME and HOMEBREW_PREFIX) → "rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md" | 1 (expected) |
| shellcheck | not run: shellcheck is not installed on this host |

First full self-test run (log /tmp/gate-selftest-18ny6p.log) exited 1: the new "Homebrew bin dir is recorded" row failed because the script canonicalised the directory with cd+pwd while $WORK contained a doubled slash (TMPDIR ends in /). Fixed by using plain dirname; second full run exited 0.

## Mutants (narrowing)
| Mutant | Result |
|---|---|
| Homebrew probes removed from the candidate list (CARGO_HOME probe kept) | section run: 3 FAIL (Homebrew rows), exit 1 — killed |
| rustup dir found but not prepended to PATH/GITHUB_PATH | direct run: "rustup: command not found", exit 127; brew/bin absent from GITHUB_PATH — killed by the contains row |

## Bounds
- Rose-air behaviour with the real Homebrew rustup is verified only on the next main push (per brief); this host has no /opt/homebrew/bin/rustup, so the fixed-prefix probes are exercised only via HOMEBREW_PREFIX pointing at a fake prefix. The /opt/homebrew/bin and /usr/local/bin literals are the same loop as the HOMEBREW_PREFIX candidate, so the mutant kill covers the loop, not each literal.
- No workflow topology change; rust-toolchain.toml, rust-pin-guard.sh, ci.yml, internal/rustsource untouched.

| golangci-lint run (repo lint; no Go files changed) — 0 issues | 0 |

## Revision 2 (republish, 2026-09-18)

revision 2 = revision 1 unchanged; gate rerun after the naming-gate resource fix on main

Rev2 verification (developer rerun, worktree converged onto main 6401d3c; `git status --short` lists exactly the three rev1 paths, `git diff --stat` 3 files, 91 insertions, 21 deletions; no file changed for rev2): bash -n on both scripts exit 0; full `bash .github/ci/gate-selftest.sh` exit 0 — 187 passed, 0 failed, including "the installer finds rustup under the Homebrew prefix without PATH help" (pass), "the installer finds rustup under CARGO_HOME/bin without PATH help" (pass), and "a runner without rustup fails" (named failure, exit 1 expected).
