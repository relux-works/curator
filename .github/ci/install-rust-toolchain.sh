#!/usr/bin/env bash
# Install the pinned Rust toolchain for the CI lane.
#
# Reads the channel from rust-toolchain.toml -- the file rust-pin-guard.sh
# verifies against internal/rustsource.SupportedRustToolchainVersion in the
# preceding lane step -- installs exactly that toolchain with rustup, and
# prepends the Rust proxy directory to PATH for the rest of the lane, so
# `rustc -vV` in the rustsource production cases resolves the pinned release
# instead of whatever the runner happens to ship (rose-air run 35121791685:
# no rustc on PATH failed the lane after the pnpm pin had gone green).
#
# rustup itself is a runner prerequisite, not something this script
# bootstraps: the hosted images ship it on PATH, and the self-hosted runner
# gets it once per docs/self-hosted-runner-setup.md -- either under
# ~/.cargo/bin (or $CARGO_HOME/bin) via rustup.rs, or as the Homebrew
# formula, which links rustup under the Homebrew prefix (/opt/homebrew/bin
# on Apple silicon) and can keep rustc/cargo proxies in the formula keg. The
# runner service starts from launchd with a minimal PATH
# (no /opt/homebrew/bin) and the lane shell reads no profiles, so this
# script resolves rustup itself, in order: PATH, $CARGO_HOME/bin,
# $HOMEBREW_PREFIX/bin (if set), /opt/homebrew/bin, /usr/local/bin, and
# always invokes it by that absolute path. Only directories holding Rust
# proxies reach PATH and GITHUB_PATH, never a shared prefix bin such as
# the Homebrew one (which also holds Homebrew go): $CARGO_HOME/bin, the
# symlink-resolved Homebrew keg of rustup, or a lane-private directory of
# links to exactly the Rust proxies. After installation, if rustc/cargo are
# still absent, the pinned toolchain bin is added. go and node must resolve
# to the same paths afterwards, or the script fails naming the directory.
# Resolution never uses RUSTUP_HOME. A runner without rustup in
# any of those places fails here with that note named (rose-air run
# 35306933411), not with a generic rustc-not-found later; the failure
# prints the runner diagnostics (runner name, searched PATH, per-candidate
# state) before the note so the next red run is self-diagnosing.
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
# Non-Rust tools whose resolution the Rust lane must never change.
guarded_tools="go node"
guarded_before_go="$(command -v go 2>/dev/null || true)"
guarded_before_node="$(command -v node 2>/dev/null || true)"
guarded_before() {
	case "$1" in
		go) printf '%s' "$guarded_before_go" ;;
		node) printf '%s' "$guarded_before_node" ;;
	esac
}

# refuse_shadowing_dir <dir>: fail if <dir> holds a guarded tool other than
# the one resolved before this script touched PATH.
refuse_shadowing_dir() {
	for _tool in $guarded_tools; do
		_before="$(guarded_before "$_tool")"
		if [ -x "$1/$_tool" ] && [ ! -d "$1/$_tool" ] && [ "$1/$_tool" != "$_before" ]; then
			fail "refusing to add $1 to PATH: it would shadow $_tool (${_before:-not on PATH} before the Rust lane, $1/$_tool in the added directory)"
		fi
	done
}

verify_lane_resolution() {
	for _tool in $guarded_tools; do
		_before="$(guarded_before "$_tool")"
		_after="$(command -v "$_tool" 2>/dev/null || true)"
		if [ "$_after" != "$_before" ]; then
			fail "the Rust lane changed $_tool resolution: ${_before:-not on PATH} before, ${_after:-not on PATH} after; shadowing directory: $(dirname "${_after:-${_before:-?}}")"
		fi
	done
}

refuse_shadowing_dir "$cargo_home/bin"
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

# Failure-path diagnostics (BUG-260922-306v4m): the remedy sentence alone
# cannot tell a different runner (the label set is self-hosted+macOS+ARM64
# at organisation level, so the machine the operator checked may not be the
# machine that ran), a different service user ($HOME/.cargo/bin resolving
# elsewhere), or a rustup living somewhere the probe does not look (an
# asdf/mise shim, ~/.local/bin, /usr/local/cargo/bin, /opt/rust/bin)
# apart -- rose-air runs 35663049586 and 35725359745 failed every main push
# with no evidence of which. The failure therefore names the runner and
# everything it searched before the remedy sentence. Failure path only: the
# success path below is untouched. Every probe here is guarded so the
# diagnostics cannot fail under `set -euo pipefail`; only the named
# variables are printed, never the whole environment.
print_rustup_diagnostics() {
	_diag_host="$(hostname 2>/dev/null || echo unknown)"
	_diag_user="$(whoami 2>/dev/null || echo unknown)"
	if [ -n "${CARGO_HOME:-}" ]; then
		_diag_cargo_note='set'
	else
		_diag_cargo_note='defaulted'
	fi
	echo "rust-pin: rustup not found; runner diagnostics:"
	echo "rust-pin:   RUNNER_NAME=${RUNNER_NAME:-unset}"
	echo "rust-pin:   RUNNER_OS=${RUNNER_OS:-unset}"
	echo "rust-pin:   hostname=$_diag_host"
	echo "rust-pin:   whoami=$_diag_user"
	echo "rust-pin:   HOME=$HOME"
	echo "rust-pin:   CARGO_HOME=$cargo_home ($_diag_cargo_note)"
	echo "rust-pin:   HOMEBREW_PREFIX=${HOMEBREW_PREFIX:-unset}"
	echo "rust-pin:   PATH=$PATH"
	for _diag_candidate in \
		"$cargo_home/bin/rustup" \
		${HOMEBREW_PREFIX:+"$HOMEBREW_PREFIX/bin/rustup"} \
		/opt/homebrew/bin/rustup \
		/usr/local/bin/rustup; do
		if [ -x "$_diag_candidate" ]; then
			_diag_state='executable'
		elif [ -e "$_diag_candidate" ] || [ -L "$_diag_candidate" ]; then
			_diag_state='exists-not-executable'
		else
			_diag_state='absent'
		fi
		echo "rust-pin:   candidate $_diag_candidate: $_diag_state"
	done
	for _diag_dir in "$cargo_home/bin" /opt/homebrew/bin /usr/local/bin; do
		if [ ! -d "$_diag_dir" ]; then
			echo "rust-pin:   listing $_diag_dir: absent"
			continue
		fi
		_diag_hits="$(ls -1 "$_diag_dir" 2>/dev/null | grep -i -E 'rust|cargo' | head -n 20 || true)"
		if [ -z "$_diag_hits" ]; then
			echo "rust-pin:   listing $_diag_dir: no names containing rust or cargo"
		else
			echo "rust-pin:   listing $_diag_dir:"
			printf '%s\n' "$_diag_hits" | sed 's/^/rust-pin:     /'
		fi
	done
	_diag_v="$(command -v rustup 2>/dev/null || echo '(no rustup on PATH)')"
	echo "rust-pin:   command -v rustup: $_diag_v"
	echo "rust-pin:   type -a rustup:"
	type -a rustup 2>&1 | sed 's/^/rust-pin:     /' || true
}

if [ -z "$rustup_bin" ]; then
	print_rustup_diagnostics >&2
	fail "rustup is not installed on this runner; install it once per docs/self-hosted-runner-setup.md, then re-run this lane"
fi

# Expose only directories that hold Rust proxies, never a shared prefix bin
# (TASK-260929-2wgjam): on a runner whose rustup is the Homebrew link
# under the Apple-silicon Homebrew prefix, prepending that prefix bin also
# put Homebrew `go` ahead of the setup-go toolchain and toolchain-identity.sh
# failed with a newer Go instead of the go.mod pin. Every lane PATH addition goes through
# prepend_lane_path, which refuses a directory holding a different go/node
# than the one resolved before this script ran; the final check re-verifies
# the resolution. rustup itself is always invoked by its absolute path.
real_binary_dir() {
	_real_path="$1"
	command -v readlink >/dev/null 2>&1 || return 1
	case "$_real_path" in
		/*) ;;
		*) _real_path="$PWD/$_real_path" ;;
	esac
	_link_hops=0
	while [ -L "$_real_path" ]; do
		_link_hops=$((_link_hops + 1))
		[ "$_link_hops" -le 40 ] || return 1
		_link_dir="$(cd -P "$(dirname "$_real_path")" 2>/dev/null && pwd)" || return 1
		_link_target="$(readlink "$_real_path")" || return 1
		case "$_link_target" in
			/*) _real_path="$_link_target" ;;
			*) _real_path="$_link_dir/$_link_target" ;;
		esac
	done
	cd -P "$(dirname "$_real_path")" 2>/dev/null && pwd
}

prepend_lane_path() {
	_lane_dir="$1"
	[ -n "$_lane_dir" ] || return 0
	refuse_shadowing_dir "$_lane_dir"
	printf '%s\n' "$_lane_dir" >>"$GITHUB_PATH"
	export PATH="$_lane_dir:$PATH"
}

rustup_dir="$(dirname "$rustup_bin")"
rustup_real_dir="$(real_binary_dir "$rustup_bin" 2>/dev/null || true)"
# Compare canonical forms: Git Bash spells one directory as /tmp/... and
# /c/Users/.../Temp/..., which must not read as a symlinked keg.
rustup_dir_canon="$(cd -P "$rustup_dir" 2>/dev/null && pwd || printf '%s' "$rustup_dir")"
if [ "$rustup_dir" = "$cargo_home/bin" ]; then
	: # already on PATH and GITHUB_PATH
elif [ -n "$rustup_real_dir" ] && [ "$rustup_real_dir" != "$rustup_dir_canon" ]; then
	# A symlinked rustup (Homebrew formula): its real directory is the keg,
	# which also holds the rustc/cargo proxies when the formula ships them.
	rustup_proxy_dir="$rustup_real_dir"
	echo "rust-pin: adding rustup proxy directory $rustup_proxy_dir"
	prepend_lane_path "$rustup_proxy_dir"
else
	# rustup is a plain file in a directory that may be shared with other
	# tools: link exactly the Rust proxies into a lane-private directory.
	lane_rust_bin="$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/rust-lane-bin.XXXXXX")" ||
		fail "cannot create a lane-private Rust bin directory"
	for _tool in rustup rustc cargo rustfmt cargo-fmt cargo-clippy clippy-driver; do
		if [ -x "$rustup_dir/$_tool" ]; then
			ln -s "$rustup_dir/$_tool" "$lane_rust_bin/$_tool"
		fi
	done
	echo "rust-pin: adding lane-private Rust bin directory $lane_rust_bin"
	prepend_lane_path "$lane_rust_bin"
fi
echo "rust-pin: using rustup at $rustup_bin"

"$rustup_bin" toolchain install "$channel" --profile minimal

# rustup.rs exposes its proxies in CARGO_HOME/bin, which was added above;
# a Homebrew keg exposes them beside the real rustup, added just above. If
# neither holds them, use the installed pinned compiler's directory.
lane_has_rust_tools() {
	command -v rustc >/dev/null 2>&1 && command -v cargo >/dev/null 2>&1
}

if ! lane_has_rust_tools; then
	if pinned_rustc="$("$rustup_bin" which --toolchain "$channel" rustc 2>/dev/null)"; then
		pinned_rustc_dir="$(dirname "$pinned_rustc")"
		if [ -n "$pinned_rustc_dir" ]; then
			echo "rust-pin: adding pinned toolchain directory $pinned_rustc_dir"
			prepend_lane_path "$pinned_rustc_dir"
		fi
	fi
fi

# Invariant: every non-Rust tool recorded before the PATH changes resolves
# to the same path after them.
verify_lane_resolution
command -v rustc >/dev/null 2>&1 || fail "rustc is not on PATH after installing Rust $channel"
command -v cargo >/dev/null 2>&1 || fail "cargo is not on PATH after installing Rust $channel"
rustc -vV
cargo --version
"$rustup_bin" show active-toolchain
