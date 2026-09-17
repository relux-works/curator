#!/bin/bash
# Throwaway MSYS simulation probe for TASK-260910-1952mz rev5 (kept in /tmp,
# not committed). Stubs uname/cygpath so the emitted POSIX hook believes it
# runs under Git Bash, then drives approved/unapproved/changed/cygpath-absent
# activations under both profiles.
set -u
ROOT=/tmp/msys-sim
rm -rf "$ROOT"
mkdir -p "$ROOT/fakebin" "$ROOT/project/.agents" "$ROOT/home"
PROJ="$ROOT/project"
CAND="$PROJ/.agents/env.sh"
printf 'export PROBE_MSYS=1\n' > "$CAND"

# Fake uname: report Git Bash.
cat > "$ROOT/fakebin/uname" <<'EOF'
#!/bin/sh
if [ "${1:-}" = "-s" ]; then echo "MINGW64_NT-10.0"; else /usr/bin/uname "$@"; fi
EOF
chmod +x "$ROOT/fakebin/uname"

# Fake cygpath -w: map /tmp/msys-sim/... (or its /private/tmp resolution) to
# C:\msys-sim\... (native form).
cat > "$ROOT/fakebin/cygpath" <<'EOF'
#!/bin/sh
# usage: cygpath -w <msys-path>
p="${2:-$1}"
case "$p" in
  /tmp/msys-sim/*) rest="${p#/tmp/msys-sim/}" ;;
  /private/tmp/msys-sim/*) rest="${p#/private/tmp/msys-sim/}" ;;
  *) printf '%s\n' "$p"; exit 0 ;;
esac
printf 'C:\\msys-sim\\%s\n' "$(printf '%s' "$rest" | tr '/' '\\')"
EOF
chmod +x "$ROOT/fakebin/cygpath"

NATIVE='C:\msys-sim\project\.agents\env.sh'
DIGEST="$(shasum -a 256 "$CAND" | awk '{print $1}')"
STAMP="2026-09-16T00:00:00Z"
CFG="$ROOT/home/config.json"
printf '{}\n' > "$CFG"

run_activation() { # $1=hook $2=proj
  PATH="$ROOT/fakebin:$PATH" CURATOR_CONFIG="$CFG" PROJ="$2" HOOK="$1" \
    sh -c 'cd "$PROJ"; . "$HOOK"; printf "sourced1=%s\n" "${PROBE_MSYS:-no}"; printf "PROBE-1\n" >&2; _curator_auto_env; printf "sourced2=%s\n" "${PROBE_MSYS:-no}"' 2> "$ROOT/stderr.txt" > "$ROOT/stdout.txt"
  echo "--- stdout:"; cat "$ROOT/stdout.txt"; echo "--- stderr:"; cat "$ROOT/stderr.txt"
}

PASS=0; FAIL=0
check() { # $1=desc $2=condition-command...
  if eval "$2"; then PASS=$((PASS+1)); echo "PASS: $1";
  else FAIL=$((FAIL+1)); echo "FAIL: $1"; fi
}

echo "=== 1. approved native record, B-enforcing: sourced silent ==="
printf '%s\t%s\tmanager\t%s\n' "$NATIVE" "$DIGEST" "$STAMP" > "$ROOT/home/hook-approvals.tsv"
run_activation /tmp/msys-hook-B-enforcing.sh "$PROJ"
check "B approved sourced" 'grep -q "sourced1=1" "$ROOT/stdout.txt" && grep -q "sourced2=1" "$ROOT/stdout.txt"'
check "B approved silent" '! grep -q "shell_hook_env_" "$ROOT/stderr.txt"'

echo "=== 2. approved native record, A-warning: sourced silent ==="
run_activation /tmp/msys-hook-A-warning.sh "$PROJ"
check "A approved sourced" 'grep -q "sourced1=1" "$ROOT/stdout.txt"'
check "A approved silent" '! grep -q "shell_hook_env_" "$ROOT/stderr.txt"'

echo "=== 3. lower-drive + upper-component record still matches (fold) ==="
printf '%s\t%s\tmanager\t%s\n' 'c:\MSYS-SIM\PROJECT\.agents\ENV.SH' "$DIGEST" "$STAMP" > "$ROOT/home/hook-approvals.tsv"
run_activation /tmp/msys-hook-B-enforcing.sh "$PROJ"
check "B folded record sourced" 'grep -q "sourced1=1" "$ROOT/stdout.txt"'
check "B folded record silent" '! grep -q "shell_hook_env_" "$ROOT/stderr.txt"'

echo "=== 4. unapproved, B-enforcing: refused + warned once ==="
rm "$ROOT/home/hook-approvals.tsv"
run_activation /tmp/msys-hook-B-enforcing.sh "$PROJ"
check "B unapproved refused" 'grep -q "sourced1=no" "$ROOT/stdout.txt" && grep -q "sourced2=no" "$ROOT/stdout.txt"'
check "B unapproved diagnostic" 'grep -q "shell_hook_env_unapproved" "$ROOT/stderr.txt"'
check "B unapproved names native path" 'grep -q "C:\\\\msys-sim" "$ROOT/stderr.txt"'
check "B unapproved names approve cmd" 'grep -q "curator hook approve" "$ROOT/stderr.txt"'
check "B unapproved warned once" '[ "$(grep -c "curator hook approve" "$ROOT/stderr.txt")" = "1" ]'

echo "=== 5. unapproved, A-warning: sourced + warned once ==="
run_activation /tmp/msys-hook-A-warning.sh "$PROJ"
check "A unapproved sourced" 'grep -q "sourced1=1" "$ROOT/stdout.txt"'
check "A unapproved warned once" '[ "$(grep -c "curator hook approve" "$ROOT/stderr.txt")" = "1" ]'

echo "=== 6. changed bytes, B-enforcing: refused with changed diagnostic ==="
printf '%s\t%s\tmanager\t%s\n' "$NATIVE" "$(printf '0%.0s' $(seq 1 64))" "$STAMP" > "$ROOT/home/hook-approvals.tsv"
run_activation /tmp/msys-hook-B-enforcing.sh "$PROJ"
check "B changed refused" 'grep -q "sourced1=no" "$ROOT/stdout.txt"'
check "B changed diagnostic" 'grep -q "shell_hook_env_changed" "$ROOT/stderr.txt"'

echo "=== 7. cygpath absent: refused with warning under BOTH profiles (never silent) ==="
mv "$ROOT/fakebin/cygpath" "$ROOT/fakebin/cygpath.hidden"
printf '%s\t%s\tmanager\t%s\n' "$NATIVE" "$DIGEST" "$STAMP" > "$ROOT/home/hook-approvals.tsv"
run_activation /tmp/msys-hook-B-enforcing.sh "$PROJ"
check "B no-cygpath refused" 'grep -q "sourced1=no" "$ROOT/stdout.txt"'
check "B no-cygpath warned" 'grep -q "shell_hook_env_unapproved" "$ROOT/stderr.txt"'
run_activation /tmp/msys-hook-A-warning.sh "$PROJ"
check "A no-cygpath refused (fail-closed)" 'grep -q "sourced1=no" "$ROOT/stdout.txt"'
check "A no-cygpath warned" 'grep -q "shell_hook_env_unapproved" "$ROOT/stderr.txt"'
mv "$ROOT/fakebin/cygpath.hidden" "$ROOT/fakebin/cygpath"

echo "=== RESULT: PASS=$PASS FAIL=$FAIL ==="
[ "$FAIL" = "0" ]
