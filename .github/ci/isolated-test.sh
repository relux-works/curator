#!/usr/bin/env bash
# One boundary for Unix test discovery, execution and artifact admission.
set -euo pipefail

while IFS= read -r name; do
  case "$name" in
    GIT_*|SSH_*|XDG_*|GOFLAGS|CODEX_HOME|CLAUDE_CONFIG_DIR|PI_CODING_AGENT_DIR)
      unset "$name"
      ;;
  esac
done < <(compgen -e)

test_home="$(mktemp -d)"
test_home="$(cd "$test_home" && pwd -P)"
[[ "$test_home" = /* ]]
export HOME="$test_home"
exec "$@"
