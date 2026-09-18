# BUG-260918-18ny6p: rose-air-homebrew-rustup-not-on-runner-path

## Description
rose-air lane on main d00fe7a5 (run 35306933411, 04:55Z) fails again with rust-pin: rustup is not installed on this runner. The runner rustup is the Homebrew build (/opt/homebrew/bin/rustup; log of the green run 35302875022 at 03:41Z shows shell /opt/homebrew/bin/bash and rustup self-update disabled for this build). When the runner service starts with a minimal launchd PATH the job shell is /bin/bash, /opt/homebrew/bin is absent, and ~/.cargo/bin holds only toolchain proxies (no rustup binary), so the CARGO_HOME/bin lookup from BUG-260918-3u6qqq does not find it. install-rust-toolchain.sh must also probe the Homebrew prefixes (/opt/homebrew/bin, /usr/local/bin, $HOMEBREW_PREFIX/bin) before failing closed.

## Scope
.github/ci/install-rust-toolchain.sh, its rows in .github/ci/gate-selftest.sh, docs/self-hosted-runner-setup.md; no workflow topology change

## Acceptance Criteria
install-rust-toolchain.sh resolves rustup from PATH, then ${CARGO_HOME:-$HOME/.cargo}/bin, then the Homebrew prefixes ($HOMEBREW_PREFIX/bin if set, /opt/homebrew/bin, /usr/local/bin), prepending the directory that contains rustup (and CARGO_HOME/bin for the proxies) to PATH and GITHUB_PATH before any rustup call; absence everywhere still fails closed with the docs-named message; hosted lanes unchanged; gate-selftest rows cover rustup only under a fake Homebrew prefix (pass), rustup only under CARGO_HOME/bin (pass, existing), absent everywhere (named failure); docs state that the launchd runner PATH lacks /opt/homebrew/bin and that the lane locates a Homebrew rustup itself (optionally mention the runner .path file as an alternative).
