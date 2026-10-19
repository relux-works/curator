#!/bin/sh
# Task-local sensitivity controls; wrappers narrow one real CLI contract each.
# They are not source mutations of the vendor executable or vendor security proofs.
set -eu
: "${CODEX162_BINARY:?set the qualified vendor binary locator}"
: "${CODEX162_MUTANTS_ROOT:?set a fresh absolute task-local output directory}"
root=$(git rev-parse --show-toplevel)
cd "$root"
mkdir "$CODEX162_MUTANTS_ROOT"
printf '%s\n' "$CODEX162_BINARY" > "$CODEX162_MUTANTS_ROOT/real-binary"
printf 'mutant\ttest\texit\n' > "$CODEX162_MUTANTS_ROOT/results.tsv"

for mutant in skip-git repeated-profile auth-other-home root-git root-git-extra-root trust-second; do
  cat > "$CODEX162_MUTANTS_ROOT/$mutant" <<'WRAPPER'
#!/bin/sh
set -eu
read -r real < "${0%/*}/real-binary"
last=
for arg do last=$arg; done
case "${0##*/}" in
  skip-git)
    if [ "$*" = 'exec --ephemeral probe' ]; then
      exec "$real" exec --ephemeral --skip-git-repo-check probe
    fi ;;
  repeated-profile)
    if [ "$*" = 'exec --skip-git-repo-check -p one -p two probe' ]; then
      exec "$real" exec --skip-git-repo-check -p one probe
    fi ;;
  auth-other-home)
    if [ "$*" = 'login status' ] && [ "${CODEX_HOME##*/}" = other-codex ]; then
      printf '{"OPENAI_API_KEY":"synthetic-wrong-home"}\n' > "$CODEX_HOME/auth.json"
    fi ;;
  root-git)
    if [ "$1" = sandbox ] && [ "$last" = .git/sentinel ]; then
      shift
      exec "$real" sandbox -P narrow \
        -c "permissions={narrow={extends=\":workspace\",filesystem={\"$PWD/.git/sentinel\"=\"write\"}}}" "$@"
    fi ;;
  root-git-extra-root)
    if [ "$1" = sandbox ] && [ "$last" = .git/sentinel ]; then
      shift
      exec "$real" sandbox -c "sandbox_workspace_write.writable_roots=[\"$PWD/.git\"]" "$@"
    fi ;;
  trust-second)
    if [ "$last" = "${last%/second-work}/second-work" ] && [ "$last" != "${last%/second-work}" ]; then
      exec "$real" --dangerously-bypass-hook-trust "$@"
    fi ;;
esac
exec "$real" "$@"
WRAPPER
  chmod 700 "$CODEX162_MUTANTS_ROOT/$mutant"
done

run() {
  mutant=$1 test=$2 diagnostic=$3
  set +e
  CODEX162_BINARY="$CODEX162_MUTANTS_ROOT/$mutant" \
    CODEX162_EVIDENCE="$CODEX162_MUTANTS_ROOT/$mutant-evidence" \
    go test ./.research/codex162 -run "^$test\$" -count=1 -v \
      > "$CODEX162_MUTANTS_ROOT/$mutant.log" 2>&1
  result=$?
  set -e
  printf '%s\t%s\t%s\n' "$mutant" "$test" "$result" >> "$CODEX162_MUTANTS_ROOT/results.tsv"
  if [ "$result" != 1 ] || ! rg -q -- "--- FAIL: $test" "$CODEX162_MUTANTS_ROOT/$mutant.log" || \
    ! rg -q -- "$diagnostic" "$CODEX162_MUTANTS_ROOT/$mutant.log"; then
    printf '%s: survivor or unrelated failure; inspect log (exit %s)\n' "$mutant" "$result" >&2
    return 1
  fi
}
run skip-git TestCodex162ExecSeedProfile 'exit=0 want=1'
run repeated-profile TestCodex162ExecSeedProfile 'exit=0 want=2'
run auth-other-home TestCodex162LinkedAuth 'exit=0 want=1'
run root-git TestCodex162GitSandbox 'exit=0 want=1'
run trust-second TestCodex162ProjectHookTrust 'second entered main UI without hook consent'
run root-git-extra-root TestCodex162GitSandbox 'exit=0 want=1'

# The registry attestation retains every row and changes only one release to
# its unqualified neighbor. Go's overlay leaves the candidate source untouched.
source_file="$root/internal/envregistry/envregistry.go"
sed 's/VerifiedRelease:   "0.162.1"/VerifiedRelease:   "0.162.0"/' "$source_file" > "$CODEX162_MUTANTS_ROOT/envregistry.go"
printf '{"Replace":{"%s":"%s"}}\n' "$source_file" "$CODEX162_MUTANTS_ROOT/envregistry.go" > "$CODEX162_MUTANTS_ROOT/overlay.json"
set +e
go test -overlay "$CODEX162_MUTANTS_ROOT/overlay.json" ./internal/envregistry \
  -run '^TestCodexCLI162VerifiedRelease$' -count=1 -v > "$CODEX162_MUTANTS_ROOT/release.log" 2>&1
result=$?
set -e
printf 'release\tTestCodexCLI162VerifiedRelease\t%s\n' "$result" >> "$CODEX162_MUTANTS_ROOT/results.tsv"
test "$result" = 1
rg -q -- '--- FAIL: TestCodexCLI162VerifiedRelease' "$CODEX162_MUTANTS_ROOT/release.log"
rg -q -- 'verified release = "0.162.0"' "$CODEX162_MUTANTS_ROOT/release.log"
cat "$CODEX162_MUTANTS_ROOT/results.tsv"
