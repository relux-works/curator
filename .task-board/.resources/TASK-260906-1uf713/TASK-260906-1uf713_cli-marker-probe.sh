#!/usr/bin/env bash
# Rework-2 CLI probe: the three marker outcomes through the production binary.
set -u
SRC=/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c
W="$(mktemp -d)"; mkdir -p "$W"
CUR="$W/curator"
(cd "$SRC" && go build -o "$CUR" ./cmd/curator) || exit 1
echo "built from $(git -C "$SRC" rev-parse --short HEAD)"

pkg() { mkdir -p "$1/context"
  printf '{"schema_version": 1, "name": "%s", "version": "1.0.0", "context": {"modules": [{"path": "a.md"}]}}\n' "$2" > "$1/agent-context.json"
  printf 'body\n' > "$1/context/a.md"; }
home() { H="$W/$1"; rm -rf "$H"; mkdir -p "$H"/native/{claude,codex,xdg/opencode,pi} "$H/opsys"
  printf '{"schema_version": 2, "skills_root": "%s/skills", "projects": {}}\n' "$H" > "$H/config.json"
  export CURATOR_CONFIG="$H/config.json" CLAUDE_CONFIG_DIR="$H/native/claude" CODEX_HOME="$H/native/codex" \
         XDG_CONFIG_HOME="$H/native/xdg" PI_CODING_AGENT_DIR="$H/native/pi" HOME="$H/opsys" USERPROFILE="$H/opsys"
  unset CURATOR_SYSTEM_CONFIG GIT_CONFIG_GLOBAL; }
pkg "$W/root" acme

echo "===== absent marker: native file is detected ====="
home absent
printf 'hello native\n' > "$H/native/claude/CLAUDE.md"
"$CUR" profile import --as absentr >/tmp/rework2-evidence/cli-absent.log 2>&1; echo "  import exit: $? (profile_use_partial: installed, activation needs --takeover)"
"$CUR" profile list 2>/dev/null | awk -F'	' '$1=="absentr" {print "  installed: " $1 " " $3 " " $6}' | head -2

echo "===== intact marker: managed root is skipped ====="
home intact
"$CUR" profile install "$W/root" --use >/dev/null 2>&1
"$CUR" profile import --as clean >/tmp/rework2-evidence/cli-intact.log 2>&1; echo "  import exit: $?"
M=$(find "$H/contexts/context/clean" -name 'claude_code.md' 2>/dev/null | head -1)
{ [ -n "$M" ] && head -2 "$M" | grep -q curator-root-context && echo "  managed root imported as native: YES (BAD)"; } || echo "  managed root imported as native: no (gate holds)"

echo "===== corrupt marker: loss naming the marker, managed root never native ====="
printf 'not json at all' > "$H/native/claude/.agent-environment.json"
"$CUR" profile import --as corrupt >/tmp/rework2-evidence/cli-corrupt.log 2>&1; echo "  import exit: $? (want non-zero)"
grep -o "environment_import_lossy" /tmp/rework2-evidence/cli-corrupt.log | head -1 | sed 's/^/  /'
grep -o "\.agent-environment\.json" /tmp/rework2-evidence/cli-corrupt.log | head -1 | sed 's/^/  loss names: /'
M2=$(find "$H/contexts/context/corrupt" -name 'claude_code.md' 2>/dev/null | head -1)
{ [ -n "$M2" ] && head -2 "$M2" | grep -q curator-root-context && echo "  managed root imported as native: YES (BYPASS)"; } || echo "  managed root imported as native: no"

echo "===== unreadable marker ====="
home unread
"$CUR" profile install "$W/root" --use >/dev/null 2>&1
chmod 000 "$H/native/claude/.agent-environment.json" 2>/dev/null || echo "  chmod refused, skipping"
if [ -r "$H/native/claude/.agent-environment.json" ]; then echo "  host reads through mode 000: untestable here"; else
"$CUR" profile import --as locked >/tmp/rework2-evidence/cli-unreadable.log 2>&1; echo "  import exit: $? (want non-zero)"
grep -o "environment_import_lossy" /tmp/rework2-evidence/cli-unreadable.log | head -1 | sed 's/^/  /'
fi
chmod 644 "$H/native/claude/.agent-environment.json" 2>/dev/null
echo "scratch: $W"
