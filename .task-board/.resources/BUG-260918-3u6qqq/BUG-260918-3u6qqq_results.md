# BUG-260918-3u6qqq results (developer)

## Fix
- `.github/ci/install-rust-toolchain.sh`: `cargo_home=${CARGO_HOME:-$HOME/.cargo}` computed first; `$cargo_home/bin` appended to `GITHUB_PATH` and prepended to `PATH` (lines 38-41) BEFORE the `command -v rustup` check (line 43) and the `rustup toolchain install` call (line 45). Fail-closed docs-named message unchanged. No `RUSTUP_HOME` use. Channel parsing, post-install rustc/cargo checks, diagnostics untouched.
- `.github/ci/gate-selftest.sh`: new rows — rustup ONLY under `CARGO_HOME/bin` with `RUSTUP_HOME` pointed elsewhere passes and records the filed-channel install + bin dir; absent-everywhere row now pins `CARGO_HOME` to an empty prefix (no `~/.cargo` leak) and still fails with the docs-named message.
- `docs/self-hosted-runner-setup.md`: `--no-modify-path` install form plus a paragraph stating the launchd service PATH / no-profile-sourcing behaviour and that the lane locates `~/.cargo/bin` (or `$CARGO_HOME/bin`) itself; no service-PATH edit instructed.

## Evidence (this host, bash 3.2, exit codes real)
- Narrow probe `/tmp/rustup-probe.sh`: row A (CARGO_HOME-only) exit=0, log shows `rustup toolchain install 1.92.0 --profile minimal`, GITHUB_PATH holds the bin dir; row B (absent) exit=1 with `rust-pin: rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane`.
- Mutant attack `/tmp/mutant-attack.sh` (PATH-only order restored): row A exits 1 — MUTANT KILLED by the new self-test row.
- Full suite `bash .github/ci/gate-selftest.sh`: exit=0, 183 passed, 0 failed (new rows green, channel-parsing and lane-wiring rows unchanged).
- `bash -n` on both edited scripts: clean. `shellcheck`: not installed on this host — unavailable, not run.
- `git status`: only the 3 in-scope files modified; no workflow topology change.

## Not verified here
- rose-air self-hosted lane runs on main pushes only; its verification happens after landing (hosted gate runs once at handoff).