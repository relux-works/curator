#!/bin/sh
# Curator installer: detects os/arch, downloads the latest release from
# GitHub, verifies the checksum, and installs the binary.
#
#   curl -fsSL https://raw.githubusercontent.com/relux-works/curator/main/install.sh | sh
#
# Environment:
#   CURATOR_VERSION   install a specific version (default: latest)
#   CURATOR_BIN_DIR   install directory (default: /usr/local/bin, falls back to ~/.local/bin)
#   CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1
#                     bypass release attestation, signature, and checksum verification
set -eu

REPO="relux-works/curator"
SIGNER_IDENTITY_REGEX='^https://github.com/relux-works/curator/\.github/workflows/release\.yml@refs/tags/v'
SIGNER_WORKFLOW="$REPO/.github/workflows/release.yml"
OIDC_ISSUER="https://token.actions.githubusercontent.com"

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
if ! curl -fsSL -o "$tmp/$archive" "$base/$archive"; then
  echo "curator installer: failed to download $archive; refusing to install" >&2
  exit 1
fi

if [ "${CURATOR_INSTALL_INSECURE_SKIP_VERIFY:-}" = "1" ]; then
  echo "curator installer: WARNING: CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1 bypasses all release attestation, signature, and archive checksum verification. The downloaded executable will be installed without integrity verification." >&2
else
  if ! curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt"; then
    echo "curator installer: failed to download checksums.txt; refusing to install" >&2
    exit 1
  fi

  gh_help=""
  use_gh=0
  if command -v gh >/dev/null 2>&1 && gh_help=$(gh attestation verify --help 2>&1); then
    use_gh=1
  fi

  if [ "$use_gh" -eq 1 ]; then
    echo "verifying GitHub build attestation..."
    if printf '%s\n' "$gh_help" | grep -q -- '--signer-workflow'; then
      if ! gh attestation verify "$tmp/checksums.txt" \
        --repo "$REPO" \
        --signer-workflow "$SIGNER_WORKFLOW" \
        --cert-oidc-issuer "$OIDC_ISSUER"; then
        echo "curator installer: GitHub attestation verification failed for checksums.txt; refusing to install" >&2
        exit 1
      fi
    else
      if ! gh attestation verify "$tmp/checksums.txt" \
        --repo "$REPO" \
        --cert-identity-regex "$SIGNER_IDENTITY_REGEX" \
        --cert-oidc-issuer "$OIDC_ISSUER"; then
        echo "curator installer: GitHub attestation verification failed for checksums.txt; refusing to install" >&2
        exit 1
      fi
    fi
  elif command -v cosign >/dev/null 2>&1; then
    echo "verifying checksums.txt signature with cosign..."
    if ! curl -fsSL -o "$tmp/checksums.txt.sig" "$base/checksums.txt.sig" ||
      ! curl -fsSL -o "$tmp/checksums.txt.pem" "$base/checksums.txt.pem"; then
      echo "curator installer: failed to download the checksums.txt signature or certificate; refusing to install" >&2
      exit 1
    fi
    if ! cosign verify-blob "$tmp/checksums.txt" \
      --signature "$tmp/checksums.txt.sig" \
      --certificate "$tmp/checksums.txt.pem" \
      --certificate-identity-regexp "$SIGNER_IDENTITY_REGEX" \
      --certificate-oidc-issuer "$OIDC_ISSUER"; then
      echo "curator installer: cosign signature verification failed for checksums.txt; refusing to install" >&2
      exit 1
    fi
  else
    echo "curator installer: release verification requires 'gh attestation verify' or cosign; install GitHub CLI (https://cli.github.com/) or cosign (https://docs.sigstore.dev/cosign/system_config/installation/), then retry; refusing to install without verification" >&2
    exit 1
  fi

  echo "verifying archive checksum..."
  if ! (
    cd "$tmp"
    awk -v archive="$archive" '
      $2 == archive { print; matches++ }
      END { if (matches != 1) exit 1 }
    ' checksums.txt > wanted.txt || exit 1
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum -c wanted.txt >/dev/null
    else
      shasum -a 256 -c wanted.txt >/dev/null
    fi
  ); then
    echo "curator installer: SHA-256 verification failed for $archive against checksums.txt; refusing to install" >&2
    exit 1
  fi
fi

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
