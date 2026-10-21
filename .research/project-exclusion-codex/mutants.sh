#!/bin/sh
# Each child must fail for the intended leaked surface, not setup or timeout.
set -eu
: "${CODEX_EXCL_BINARY:?select the exact qualified binary}"
: "${CODEX_EXCL_MUTANTS_ROOT:?set a fresh absolute .temp directory}"
excl_go=${CURATOR_GO:-go}
mkdir -p "$CODEX_EXCL_MUTANTS_ROOT"
excl_summary="$CODEX_EXCL_MUTANTS_ROOT/results.tsv"
printf 'control\texit\tverdict\n' > "$excl_summary"
check_mutant() {
    excl_name=$1
    excl_test=$2
    excl_diagnostic=$3
    excl_log="$CODEX_EXCL_MUTANTS_ROOT/$excl_name.log"
    set +e
    env -u GIT_DIR -u GIT_WORK_TREE CODEX_EXCL_MUTANT="$excl_name" \
        CODEX_EXCL_RUN_ROOT="$CODEX_EXCL_MUTANTS_ROOT/$excl_name" \
        "$excl_go" test ./.research/project-exclusion-codex -count=1 -v -run "$excl_test" > "$excl_log" 2>&1
    excl_code=$?
    set -e
    if [ "$excl_code" -eq 0 ] || ! grep -F "$excl_diagnostic" "$excl_log" > /dev/null; then
        printf '%s\t%s\tINVALID_OR_SURVIVED\n' "$excl_name" "$excl_code" >> "$excl_summary"
        cat "$excl_log"
        exit 1
    fi
    printf '%s\t%s\tKILLED\n' "$excl_name" "$excl_code" >> "$excl_summary"
}
check_mutant drop-document-cap '^TestInstructions$' 'excluded marker "PROJECT_ROOT" returned'
check_mutant drop-skill-disable '^TestSkills$' 'excluded marker "PROJECT_AGENTS_SKILL" returned'
check_mutant trust-project '^TestConfigLayers/untrusted$' 'excluded marker "PROJECT_HOOK" returned'
check_mutant allow-loopback '^TestOfflineBoundary$' 'expected connect denial exit 7, got 0'
cat "$excl_summary"
