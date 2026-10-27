#!/usr/bin/env bash
# GoReleaser runs checksum signing after archives/packages/notarization and
# before publishing. A missing companion must fail before any release upload.
set -euo pipefail
python3 release/bundle.py artifacts "$(dirname "$1")"
exec cosign sign-blob --output-certificate="$2" --output-signature="$3" "$1" --yes
