#!/bin/sh
# Expected-red controls: each must fail the intended invariant, never setup.
set -eu
: "${MUSE_EXCL_BINARY:?set the qualified binary privately}"
: "${MUSE_EXCL_RUN_ROOT:?set a fresh task-local .temp directory}"
base=$MUSE_EXCL_RUN_ROOT
mkdir -p "$base"
run_control() {
 name=$1 pattern=$2 reason=$3
 set +e
 MUSE_EXCL_MUTANT=$name MUSE_EXCL_RUN_ROOT="$base/$name" \
  env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-muse -count=1 -timeout 1m -v -run "$pattern" > "$base/$name.log" 2>&1
 code=$?
 set -e
 test "$code" -eq 1
 grep -F "$reason" "$base/$name.log" > /dev/null
 printf '%s: expected exit=1; intended invariant failed\n' "$name"
}
run_control trust '^TestP1RootAndAncestors$/^AGENTS.md$/^root$/^trusted-false$' 'rules.context_load project_sources: [1]; want 0'
run_control network '^TestOfflineBoundary$' 'network deny exit=0 hits=2'
run_control keychains '^TestOfflineBoundary$' 'synthetic Keychains denial failed'
# These setup guards must refuse before launching an unqualified vendor process.
set +e
MUSE_EXCL_BINARY=/usr/bin/true MUSE_EXCL_RUN_ROOT="$base/digest" \
 env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-muse -run '^TestChecker$' -count=1 -v > "$base/digest.log" 2>&1
code=$?
set -e
test "$code" -eq 1
grep -F 'binary digest mismatch' "$base/digest.log" > /dev/null
printf 'digest guard: expected exit=1 before vendor launch\n'
set +e
MUSE_EXCL_RUN_ROOT="$PWD/non-temp-control" \
 env -u GIT_DIR -u GIT_WORK_TREE go test ./.research/project-exclusion-muse -run '^TestChecker$' -count=1 -v > "$base/run-root.log" 2>&1
code=$?
set -e
test "$code" -eq 1
grep -F 'explicit absolute .temp run root required' "$base/run-root.log" > /dev/null
test ! -e "$PWD/non-temp-control"
printf 'run-root guard: expected exit=1 before fixture writes\n'
