#!/bin/sh
# Curator installer: detects os/arch, downloads the latest release from
# GitHub, verifies the checksum, and installs the binary.
#
#   curl -fsSL https://raw.githubusercontent.com/relux-works/curator/main/install.sh | sh
#
# Environment:
#   CURATOR_VERSION   install a specific version (default: latest)
#   CURATOR_BIN_DIR   install directory (default: /usr/local/bin, falls back to ~/.local/bin)
#   CURATOR_YES=1     skip confirmation (also: --yes)
#   --plan            show the plan without downloading or installing artifacts
#   CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1
#                     bypass release attestation, signature, and checksum verification
set -eu

REPO="relux-works/curator"
SIGNER_IDENTITY_REGEX='^https://github.com/relux-works/curator/\.github/workflows/release\.yml@refs/tags/v'
SIGNER_WORKFLOW="$REPO/.github/workflows/release.yml"
OIDC_ISSUER="https://token.actions.githubusercontent.com"

# Keep translations together. Labels are padded inside each block because
# POSIX printf field widths do not consistently count UTF-8 display columns.
case "${LC_ALL:-${LC_MESSAGES:-${LANG:-}}}" in
  ru|ru_*|ru.*|ru@*)
    msg_title='План установки'
    msg_component='  Компонент:      %s\n'
    msg_version='  Версия:         %s\n'
    msg_target='  Путь:           %s\n'
    msg_sudo='  Нужен sudo:     нет\n'
    msg_source='  Исходники:      %s\n'
    msg_readme='  README:         %s\n'
    msg_book='  Книга:          %s\n'
    msg_spec='  Спецификация:   %s\n'
    msg_verify='Проверка: GitHub build attestation или подпись cosign, затем checksum SHA-256.\n'
    msg_insecure='Проверка ОТКЛЮЧЕНА: CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1. Файл будет установлен без проверки целостности.\n'
    msg_prompt='Продолжить? [д/Н] '
    msg_cancel='Установка отменена. Изменений нет.\n'
    msg_terminal='Для подтверждения нужен терминал. Для установки без вопроса используйте --yes или CURATOR_YES=1.\n'
    msg_option='curator installer: неизвестный параметр: %s (используйте --plan или --yes)\n'
    yes_word='да'; yes_upper='ДА'; yes_title='Да'; yes_short='д'; yes_short_upper='Д'
    ;;
  *)
    msg_title='Installation plan'
    msg_component='  Component:      %s\n'
    msg_version='  Version:        %s\n'
    msg_target='  Path:           %s\n'
    msg_sudo='  Needs sudo:     no\n'
    msg_source='  Source:         %s\n'
    msg_readme='  README:         %s\n'
    msg_book='  Book:           %s\n'
    msg_spec='  Specification:  %s\n'
    msg_verify='Verification: GitHub build attestation or a cosign signature, then the SHA-256 checksum.\n'
    msg_insecure='Verification DISABLED: CURATOR_INSTALL_INSECURE_SKIP_VERIFY=1. The file will be installed without integrity verification.\n'
    msg_prompt='Continue? [y/N] '
    msg_cancel='Installation cancelled. Nothing changed.\n'
    msg_terminal='Confirmation needs a terminal. Use --yes or CURATOR_YES=1 to install without a prompt.\n'
    msg_option='curator installer: unknown option: %s (use --plan or --yes)\n'
    yes_word='yes'; yes_upper='YES'; yes_title='Yes'; yes_short='y'; yes_short_upper='Y'
    ;;
esac

assume_yes=${CURATOR_YES:-0}
plan_only=0
for option in "$@"; do
  case "$option" in
    --yes) assume_yes=1 ;;
    --plan) plan_only=1 ;;
    *) printf "$msg_option" "$option" >&2; exit 1 ;;
  esac
done

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

# Resolve exactly the existing destination rule, without creating it yet.
bin_dir="${CURATOR_BIN_DIR:-/usr/local/bin}"
create_bin_dir=0
if [ ! -w "$bin_dir" ]; then
  bin_dir="$HOME/.local/bin"
  create_bin_dir=1
fi

# One row group per component; a future bundle can add another call here.
print_component() {
  printf "$msg_component" "$1"
  printf "$msg_version" "$2"
  printf "$msg_target" "$3"
  printf "$msg_sudo"
  printf "$msg_source" "$4"
  printf "$msg_readme" "$5"
  printf "$msg_book" "$6"
  printf "$msg_spec" "$7"
}

colour=''
reset=''
if [ -t 1 ] && [ "${NO_COLOR+x}" != x ]; then
  colour=$(printf '\033[1;36m')
  reset=$(printf '\033[0m')
fi
printf '\n%s%s%s\n\n' "$colour" "$msg_title" "$reset"
print_component curator "$version" "$bin_dir/curator" \
  "https://github.com/$REPO" "https://github.com/$REPO#readme" \
  "https://github.com/$REPO/blob/main/docs/book/README.md" \
  'https://github.com/relux-works/curator-spec'
printf '\n'
if [ "${CURATOR_INSTALL_INSECURE_SKIP_VERIFY:-}" = 1 ]; then
  printf "$msg_insecure"
else
  printf "$msg_verify"
fi
printf '\n'

[ "$plan_only" -eq 0 ] || exit 0
if [ "$assume_yes" != 1 ]; then
  printf '%s' "$msg_prompt"
  # stdin can contain this script (curl | sh); never consume it as an answer.
  if ! { IFS= read -r answer </dev/tty; } 2>/dev/null; then
    printf '\n' >&2
    printf "$msg_terminal" >&2
    exit 1
  fi
  case "$answer" in
    "$yes_word"|"$yes_upper"|"$yes_title"|"$yes_short"|"$yes_short_upper"|y|Y|yes|Yes|YES) ;;
    *) printf "$msg_cancel"; exit 1 ;;
  esac
fi

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

if [ "$create_bin_dir" -eq 1 ]; then
  mkdir -p "$bin_dir"
fi
install -m 0755 "$tmp/curator" "$bin_dir/curator"

echo "installed $("$bin_dir/curator" --version) to $bin_dir/curator"
case ":$PATH:" in
  *":$bin_dir:"*) ;;
  *) echo "note: add $bin_dir to your PATH" ;;
esac
