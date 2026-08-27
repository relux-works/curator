#!/bin/sh
# Capture one macOS hardened-capability evidence packet.
#
# Every command is run as its own process and its real exit status is recorded.
# Nothing is piped through tee: a pipeline would report the status of the last
# stage, and an evidence harness that misreports its own exit code is worse than
# no harness at all.
#
# Usage: ./capture-evidence.sh <output-directory>
#
# Exit status:
#   0  the packet was captured (whatever the host turned out to support)
#   2  the packet could not be captured

set -eu

if [ $# -ne 1 ]; then
	echo "usage: $0 <output-directory>" >&2
	exit 2
fi

OUT=$1
HERE=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$OUT"
BIN="$OUT/hardened-probe"

log() { printf '%s\n' "$*"; }

# ---------------------------------------------------------------- host facts

{
	log "captured-by: $0"
	log "prototype-dir: $HERE"
	log "date-utc: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
	log "uname: $(uname -a)"
	log "sw_vers-productName: $(sw_vers --productName)"
	log "sw_vers-productVersion: $(sw_vers --productVersion)"
	log "sw_vers-buildVersion: $(sw_vers --buildVersion)"
	log "arch: $(arch)"
	log "uid: $(id -u)"
	log "csrutil: $(csrutil status 2>&1 || true)"
	log "go: $(go version)"
	log "sandbox-exec: $(ls -l /usr/bin/sandbox-exec)"
} > "$OUT/host.txt"

# ------------------------------------------------------------------- build

cd "$HERE"
go build -o "$BIN" ./cmd/hardened-probe
BUILD_STATUS=$?
log "build ./cmd/hardened-probe -> exit $BUILD_STATUS"

# ---------------------------------------------------------------- the runs
#
# run_case <name> <expected-exit> <args...>
# It records stdout, stderr and the real exit status of one invocation, and
# fails the capture when the status is not the one the case is documented to
# produce. An unexpected status is a finding, not something to smooth over.

STATUS_FILE="$OUT/exit-codes.txt"
: > "$STATUS_FILE"
FAILED=0

run_case() {
	name=$1
	expected=$2
	shift 2
	set +e
	"$BIN" "$@" > "$OUT/$name.stdout.log" 2> "$OUT/$name.stderr.log"
	actual=$?
	set -e
	printf '%s\texit=%d\texpected=%d\tcmd=hardened-probe %s\n' \
		"$name" "$actual" "$expected" "$*" >> "$STATUS_FILE"
	if [ "$actual" -ne "$expected" ]; then
		echo "$name: exit $actual, expected $expected" >&2
		FAILED=1
	fi
}

# 1. The inventory, which needs no host capability at all.
run_case list-classes 0 --list-classes

# 2. The measurement run. Exit 1 is the fail-closed outcome and the expected
#    result on an unqualified platform; a host that established every guarantee
#    would exit 0 and this case would report the difference.
set +e
"$BIN" --work-dir "$OUT/work-measure" \
	--evidence "$OUT/evidence.json" \
	--report "$OUT/report.json" \
	> "$OUT/measure.stdout.log" 2> "$OUT/measure.stderr.log"
MEASURE_STATUS=$?
set -e
printf 'measure\texit=%d\texpected=0-or-1\tcmd=hardened-probe --work-dir ... --evidence ... --report ...\n' \
	"$MEASURE_STATUS" >> "$STATUS_FILE"
if [ "$MEASURE_STATUS" -ne 0 ] && [ "$MEASURE_STATUS" -ne 1 ]; then
	echo "measure: exit $MEASURE_STATUS is a harness error, not an outcome" >&2
	FAILED=1
fi

# 3. The fail-closed sweep: every capability class forced unavailable in turn.
#    The sweep itself must not report a harness error, which is exit 2.
set +e
"$BIN" --work-dir "$OUT/work-sweep" \
	--report "$OUT/report-fail-closed.json" \
	--fail-closed-sweep --quiet \
	> "$OUT/sweep.stdout.log" 2> "$OUT/sweep.stderr.log"
SWEEP_STATUS=$?
set -e
printf 'fail-closed-sweep\texit=%d\texpected=0-or-1\tcmd=hardened-probe --fail-closed-sweep --quiet ...\n' \
	"$SWEEP_STATUS" >> "$STATUS_FILE"
if [ "$SWEEP_STATUS" -eq 2 ]; then
	echo "fail-closed-sweep: exit 2, the sweep found a class that did not fail closed" >&2
	FAILED=1
fi

# 4. Negative control on the exit contract itself: asserting the outcome this
#    host cannot produce must fail, and asserting the one it does must pass.
run_case assert-rejected 0 --work-dir "$OUT/work-assert-rejected" \
	--force-unavailable network-syscall-denial --expect rejected --quiet
run_case assert-established 2 --work-dir "$OUT/work-assert-established" \
	--force-unavailable network-syscall-denial --expect established --quiet

# 5. The harness's own hygiene. Every case above created probe domains and
#    descendants, and one of the platform findings is that macOS cannot tear all
#    of them down. Whatever the platform did, nothing of this capture may still
#    be running, so the answer is recorded rather than assumed.

LEFTOVER_FILE="$OUT/leftover-processes.txt"
{
	printf 'checked-at: %s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
	printf 'command: pgrep -f %s\n' "$BIN"
	pgrep -f "$BIN" || printf '(no process matched)\n'
} > "$LEFTOVER_FILE"

LEFTOVER=$(pgrep -f "$BIN" | wc -l | tr -d ' ')
printf 'leftover-processes\tcount=%s\texpected=0\tcmd=pgrep -f %s\n' \
	"$LEFTOVER" "$BIN" >> "$STATUS_FILE"
if [ "$LEFTOVER" -ne 0 ]; then
	echo "leftover-processes: $LEFTOVER probe process(es) survived the capture" >&2
	FAILED=1
fi

# ---------------------------------------------------------------- artifacts

rm -rf "$OUT/work-measure" "$OUT/work-sweep" \
	"$OUT/work-assert-rejected" "$OUT/work-assert-established"

log "wrote $OUT"
cat "$STATUS_FILE"

if [ "$FAILED" -ne 0 ]; then
	echo "capture-evidence: at least one case did not produce its documented exit status" >&2
	exit 2
fi
