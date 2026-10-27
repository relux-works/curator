#!/usr/bin/env bash
# Node 22 is provisioned before both Unix test entrypoints.
set -euo pipefail
exec node "$(dirname "${BASH_SOURCE[0]}")/isolated-test.cjs" "$@"
