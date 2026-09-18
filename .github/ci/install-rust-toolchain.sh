#!/usr/bin/env bash
# Install the pinned Rust toolchain for the CI lane.
#
# Reads the channel from rust-toolchain.toml -- the file rust-pin-guard.sh
# verifies against internal/rustsource.SupportedRustToolchainVersion in the
# preceding lane step -- installs exactly that toolchain with rustup, and
# prepends the rustup shim directory to PATH for the rest of the lane, so
# `rustc -vV` in the rustsource production cases resolves the pinned release
# instead of whatever the runner happens to ship (rose-air run 35121791685:
# no rustc on PATH failed the lane after the pnpm pin had gone green).
#
# rustup itself is a runner prerequisite, not something this script
# bootstraps: the hosted images ship it on PATH, and the self-hosted runner
# gets it once per docs/self-hosted-runner-setup.md -- either under
# ~/.cargo/bin (or $CARGO_HOME/bin) via rustup.rs, or as the Homebrew
# formula, which keeps the rustup binary under the Homebrew prefix
# (/opt/homebrew/bin on Apple silicon) and only the toolchain proxies under
# ~/.cargo/bin. The runner service starts from launchd with a minimal PATH
# (no /opt/homebrew/bin) and the lane shell reads no profiles, so this
# script resolves rustup itself, in order: PATH, $CARGO_HOME/bin,
# $HOMEBREW_PREFIX/bin (if set), /opt/homebrew/bin, /usr/local/bin. The
# directory that holds rustup is prepended to PATH and GITHUB_PATH before
# any rustup call; $CARGO_HOME/bin stays on both too because the proxies
# live there. Resolution never uses RUSTUP_HOME. A runner without rustup in
# any of those places fails here with that note named (rose-air run
# 35306933411), not with a generic rustc-not-found later.
#
# Usage (as a CI lane step; GITHUB_PATH must be set):
#   bash .github/ci/install-rust-toolchain.sh
#
# Self-test overrides:
#   CI_RUST_TOOLCHAIN_FILE toolchain file to read the channel from
#                          (default: rust-toolchain.toml)

set -euo pipefail

fail() { echo "rust-pin: $*" >&2; exit 1; }

file="${CI_RUST_TOOLCHAIN_FILE:-rust-toolchain.toml}"
[ -r "$file" ] || fail "cannot read $file"

channel="$(awk '/^[[:space:]]*\[toolchain[[:space:]]*\]/{in_toolchain=1; next} /^[[:space:]]*\[/{in_toolchain=0; next} in_toolchain && /^[[:space:]]*channel[[:space:]]*=[[:space:]]*"/{line=$0; sub(/^[^=]*=[[:space:]]*"/, "", line); sub(/".*$/, "", line); print line; exit}' "$file")"
[ -n "$channel" ] || fail "no channel in the [toolchain] section of $file"

cargo_home="${CARGO_HOME:-$HOME/.cargo}"
[ -n "${GITHUB_PATH:-}" ] || fail "GITHUB_PATH is not set; run this script as a CI lane step"
printf '%s\n' "$cargo_home/bin" >>"$GITHUB_PATH"
export PATH="$cargo_home/bin:$PATH"

# Resolve rustup: PATH (now led by $cargo_home/bin), then the Homebrew
# prefixes. The first executable wins and its directory is prepended so the
# rustup call below and every later step in the lane resolve the same binary.
rustup_bin=""
if command -v rustup >/dev/null 2>&1; then
	rustup_bin="$(command -v rustup)"
else
	for candidate in \
		"$cargo_home/bin/rustup" \
		${HOMEBREW_PREFIX:+"$HOMEBREW_PREFIX/bin/rustup"} \
		/opt/homebrew/bin/rustup \
		/usr/local/bin/rustup; do
		if [ -x "$candidate" ]; then
			rustup_bin="$candidate"
			break
		fi
	done
fi
[ -n "$rustup_bin" ] || fail "rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane"

rustup_dir="$(dirname "$rustup_bin")"
if [ "$rustup_dir" != "$cargo_home/bin" ]; then
	printf '%s\n' "$rustup_dir" >>"$GITHUB_PATH"
	export PATH="$rustup_dir:$PATH"
fi
echo "rust-pin: using rustup at $rustup_bin"

rustup toolchain install "$channel" --profile minimal

command -v rustc >/dev/null 2>&1 || fail "rustc is not on PATH after installing Rust $channel"
command -v cargo >/dev/null 2>&1 || fail "cargo is not on PATH after installing Rust $channel"
rustc -vV
cargo --version
rustup show active-toolchain
