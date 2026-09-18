# BUG-260918-3u6qqq: rose-air-rustup-not-found-on-runner-path

## Description
rose-air lane step Install pinned Rust toolchain via rustup fails with rust-pin: rustup is not installed on this runner (main run 35290929324, rerun after the operator installed rustup on macbook-iv on 2026-09-18). .github/ci/install-rust-toolchain.sh checks command -v rustup BEFORE adding CARGO_HOME/bin to PATH; the self-hosted runner service starts from launchd with a minimal PATH and the lane shell is bash --noprofile --norc, so a per-user rustup in ~/.cargo/bin (installed with --no-modify-path per docs/self-hosted-runner-setup.md) is invisible. Hosted images pass only because they ship rustup on PATH.

## Scope
.github/ci/install-rust-toolchain.sh, its self-test under .github/ci (if present), docs/self-hosted-runner-setup.md; no workflow topology change

## Acceptance Criteria
install-rust-toolchain.sh resolves rustup from PATH OR from ${CARGO_HOME:-$HOME/.cargo}/bin (and RUSTUP_HOME-independent), prepends that bin dir to PATH and GITHUB_PATH before the rustup check, and still fails closed with the named docs note when rustup is absent in both places; hosted lanes unchanged; a script self-test covers: rustup only under CARGO_HOME/bin (pass), rustup absent everywhere (named failure), and the existing channel-parsing cases; docs/self-hosted-runner-setup.md states that the runner PATH is not read from shell profiles and that ~/.cargo/bin is found by the lane itself.
