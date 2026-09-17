#!/usr/bin/env bash
# rust pin guard: rust-toolchain.toml must name exactly the release
# internal/rustsource admits, or the lane stops before installing anything.
#
# The Go constant internal/rustsource.SupportedRustToolchainVersion is the
# single source of truth; rust-toolchain.toml is the file the lanes install
# from (and the file local cargo/rustup reads), kept as a committed file --
# rather than extracted at install time -- so the installed toolchain is
# visible in the repository and reproducible outside CI. This script is what
# keeps the file honest: a Go-side bump that forgets the file fails the lane
# here, instead of installing a toolchain the closed Cargo registry would
# refuse downstream; a file bump without the Go constant fails the same way.
#
# A read failure is never a match: an unreadable file, a file with no
# [toolchain] channel, a non-minimal profile, a floating channel, an
# unreadable source file, or a source file with no declaration all fail
# closed.
#
# Usage:
#   bash .github/ci/rust-pin-guard.sh
#
# Self-test overrides:
#   CI_RUST_GO_SOURCE      Go source to read the constant from
#                          (default: internal/rustsource/toolchain_registry.go)
#   CI_RUST_TOOLCHAIN_FILE toolchain file to read the channel from
#                          (default: rust-toolchain.toml)

set -u

fail() { echo "rust-pin: $*" >&2; exit 1; }

file="${CI_RUST_TOOLCHAIN_FILE:-rust-toolchain.toml}"
[ -r "$file" ] || fail "cannot read $file"

channel="$(awk '/^[[:space:]]*\[toolchain[[:space:]]*\]/{in_toolchain=1; next} /^[[:space:]]*\[/{in_toolchain=0; next} in_toolchain && /^[[:space:]]*channel[[:space:]]*=[[:space:]]*"/{line=$0; sub(/^[^=]*=[[:space:]]*"/, "", line); sub(/".*$/, "", line); print line; exit}' "$file")"
[ -n "$channel" ] || fail "no channel in the [toolchain] section of $file"
printf '%s' "$channel" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$' || fail "$file pins channel \"$channel\", which is not an exact x.y.z release"

profile="$(awk '/^[[:space:]]*\[toolchain[[:space:]]*\]/{in_toolchain=1; next} /^[[:space:]]*\[/{in_toolchain=0; next} in_toolchain && /^[[:space:]]*profile[[:space:]]*=[[:space:]]*"/{line=$0; sub(/^[^=]*=[[:space:]]*"/, "", line); sub(/".*$/, "", line); print line; exit}' "$file")"
[ "$profile" = "minimal" ] || fail "$file sets profile \"$profile\", want \"minimal\""

src="${CI_RUST_GO_SOURCE:-internal/rustsource/toolchain_registry.go}"
[ -r "$src" ] || fail "cannot read $src"

line="$(grep -E 'SupportedRustToolchainVersion[[:space:]]*=[[:space:]]*"' "$src" | head -n 1)"
[ -n "$line" ] || fail "no SupportedRustToolchainVersion declaration in $src"
go_pin="$(printf '%s' "$line" | sed -n 's/.*SupportedRustToolchainVersion[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p')"
[ -n "$go_pin" ] || fail "cannot parse SupportedRustToolchainVersion from $src"

[ "$channel" = "$go_pin" ] || fail "rust-toolchain.toml channel=$channel disagrees with internal/rustsource SupportedRustToolchainVersion=$go_pin"

printf 'rust-pin: toolchain file and Go agree on Rust %s\n' "$channel"
