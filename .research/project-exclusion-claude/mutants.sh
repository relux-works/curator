#!/bin/sh
# Expected-red controls require the exact failed claim, not any nonzero exit.
set -eu
: "${CLAUDE_EXCL_BINARY:?set the exact pinned vendor copy}"
: "${CLAUDE_EXCL_RUN_ROOT:?set a fresh absolute task-local .temp directory}"
go_tool=${CURATOR_GO:-go}
mkdir -p "$CLAUDE_EXCL_RUN_ROOT"
run_control() {
 name=$1
 pattern=$2
 cause=$3
 omit=$4
 boundary=$5
 log="$CLAUDE_EXCL_RUN_ROOT/$name.log"
 set +e
 env -u GIT_DIR -u GIT_WORK_TREE CLAUDE_EXCL_OMIT="$omit" CLAUDE_EXCL_BOUNDARY_MUTANT="$boundary" CLAUDE_EXCL_RUN_ROOT="$CLAUDE_EXCL_RUN_ROOT/$name" "$go_tool" test ./.research/project-exclusion-claude -count=1 -timeout=90s -run "$pattern" -v > "$log" 2>&1
 result=$?
 set -e
 if [ "$result" -ne 1 ] || ! /usr/bin/grep -F "$cause" "$log" >/dev/null; then
  printf 'FAIL %s: exit=%s, intended failure absent; see %s\n' "$name" "$result" "$log"
  exit 1
 fi
 printf '%s: test exit=%s; intended failure: %s\n' "$name" "$result" "$cause"
}
run_control source-control '^TestP1Instructions$/^root$/^exclude-true$' 'observed EXCL_FOREIGN_ROOT_7352=true; want false' setting-sources ''
run_control mcp-control '^TestP5MCP$/^project-empty-false$/^exclude-false$/^strict-true$/^approved-true$' 'forbidden foreign-mcp startup:' strict-mcp ''
run_control network-boundary '^TestOfflineBoundary$' 'unrelated loopback not denied:' '' network
run_control keychains-boundary '^TestOfflineBoundary$' 'synthetic Keychains read not denied:' '' keychains
