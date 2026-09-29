#!/usr/bin/env bash
# Employer name gate.
#
# Curator is an open protocol implementation. The employer under whose roof
# some of it was written is not part of that protocol and must not appear
# anywhere in the repository -- not in source, docs, task records or commit
# messages. Historical records that named a local checkout path or an internal
# remote were masked to `intranet` rather than deleted.
#
# Both patterns are assembled from parts so this file does not trip the gate
# it defines.
#
# The one exemption: inside a git binary-patch block (from a line exactly
# `GIT binary patch` to the next `diff --git ` line or EOF), `literal N` /
# `delta N` lines, empty lines and base85 data lines are encoding noise, not
# mentions. Every other line, including any other line inside a block, is
# scanned.
#
# Reports carry `path:line` only, never the line content: CI job logs are
# excerpted into immutable board records, so a report that echoed the line
# would plant fresh hits in the next run. A second narrow exemption covers the
# echoes the old content-printing format already left behind: a line carrying
# the GitHub Actions log prefix of this step -- the step name, a TAB (literal,
# or `\t` inside a JSON string), an RFC3339 UTC timestamp and ` ./` -- is a
# machine echo and is skipped. Excerpts the runner truncated inside the
# prefix keep only its tail: the `…` truncation marker directly followed by
# the rest of the timestamp and ` ./` is the same echo. Stated bound: a line forged with that exact
# signature is exempt too; prose without it never is.
#
# Usage:
#   naming-gate.sh [root]    (default: the current directory)

set -u

root="${1:-.}"
cd "$root" || exit 2

full="$(printf '%s%s' wild berries)"
short="$(printf '%s%s' w b)"

hits="$(mktemp "${TMPDIR:-/tmp}/naming-gate.XXXXXX")" || exit 2
trap 'rm -f "$hits"' EXIT

scan() {
	# $1 = ERE; grep lists candidate files, python re-reads each file and
	# reports every matching line that is not binary-patch noise.
	pat="$1"
	grep -rIliE --exclude-dir=.git -- "$pat" . >"$hits.g"
	rc=$?
	if [ "$rc" -gt 1 ]; then
		echo "naming gate: grep failed (exit $rc)" >&2
		exit 2
	fi
	python3 - "$pat" "$hits.g" >"$hits" <<'PY' || exit 2
import re, sys
pat = re.compile(sys.argv[1].encode(), re.I)
B85 = re.compile(rb'^[A-Za-z][0-9A-Za-z!#$%&()*+;<=>?@^_`{|}~-]+$')
LIT = re.compile(rb'^(literal|delta) [0-9]+$')
ECHO = re.compile(rb'Employer name gate(\t|\\t)[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9:.]+Z \./'
                  rb'|\xe2\x80\xa6[0-9-]*T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?Z \./')
out = sys.stdout.buffer
for f in open(sys.argv[2], 'rb').read().split(b'\n'):
    if not f:
        continue
    inblock = False
    with open(f, 'rb') as fh:
        for n, raw in enumerate(fh, 1):
            line = raw.rstrip(b'\n').rstrip(b'\r')
            if line.startswith(b'diff --git '):
                inblock = False
            elif line == b'GIT binary patch':
                inblock = True
            elif inblock and (line == b'' or LIT.match(line) or B85.match(line)):
                continue
            if pat.search(line) and not ECHO.search(line):
                out.write(f + b':' + str(n).encode() + b'\n')
PY
	rm -f "$hits.g"
}

status=0
scan "$full"
if [ -s "$hits" ]; then
	cat "$hits"
	echo "naming gate: the employer name must not appear in this repository"
	status=1
fi
scan "\\b$short\\b"
if [ -s "$hits" ]; then
	cat "$hits"
	echo "naming gate: the employer's short name must not appear in this repository"
	status=1
fi
exit "$status"
