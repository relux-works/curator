#!/bin/sh
# Curator installer: detects os/arch, downloads the latest release from
# GitHub, verifies the checksum, and installs the binary.
#
#   curl -fsSL https://raw.githubusercontent.com/relux-works/curator/main/install.sh | sh
#
# Environment:
#   CURATOR_VERSION   install a specific version (default: latest)
#   CURATOR_BIN_DIR   install directory (default: /usr/local/bin, falls back to ~/.local/bin)
set -eu

REPO="relux-works/curator"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin|linux) ;;
  *) echo "curator installer: unsupported OS: $os (use Scoop on Windows)" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "curator installer: unsupported architecture: $arch" >&2; exit 1 ;;
esac

version="${CURATOR_VERSION:-}"
if [ -z "$version" ]; then
  version=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" |
    grep '"tag_name"' | head -1 | cut -d '"' -f 4)
fi
[ -n "$version" ] || { echo "curator installer: could not determine the latest version" >&2; exit 1; }
bare=${version#v}

archive="curator_${bare}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$version"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "downloading curator $version ($os/$arch)..."
curl -fsSL -o "$tmp/$archive" "$base/$archive"
curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt"

echo "verifying checksum..."
(
  cd "$tmp"
  grep " $archive\$" checksums.txt > wanted.txt
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c wanted.txt >/dev/null
  else
    shasum -a 256 -c wanted.txt >/dev/null
  fi
)

tar -xzf "$tmp/$archive" -C "$tmp" curator

bin_dir="${CURATOR_BIN_DIR:-/usr/local/bin}"
if [ ! -w "$bin_dir" ]; then
  bin_dir="$HOME/.local/bin"
  mkdir -p "$bin_dir"
fi
install -m 0755 "$tmp/curator" "$bin_dir/curator"

echo "installed $("$bin_dir/curator" --version) to $bin_dir/curator"
case ":$PATH:" in
  *":$bin_dir:"*) ;;
  *) echo "note: add $bin_dir to your PATH" ;;
esac
