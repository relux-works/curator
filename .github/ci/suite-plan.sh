#!/usr/bin/env bash
# Suite plan: decides, from the supplied conformance root alone, how this
# runner's package set is executed.
#
# Two questions, both answered by the root and never by a hand-maintained
# per-lane list:
#
#  1. PLATFORM EXCLUSION. `vectors/conformance-claim-v3-qualification.json` is
#     the protocol's own statement of which platforms it qualifies. A platform
#     it marks `excluded` does not execute the packages listed in
#     `.github/ci/platform-exclusions.tsv` there -- and the exclusion is
#     ASSERTED, not merely obeyed: the named rejection case must still run and
#     pass on that runner, proving the package fails closed rather than being
#     quietly dropped. When the root stops excluding the platform, the packages
#     come back automatically; nothing here needs editing.
#
#  2. ROOT COVERAGE. `.github/ci/root-artifacts.tsv` declares which root
#     artefacts each package's conformance tests read without a guard. A
#     package whose artefacts the root does not publish runs with
#     CURATOR_CONFORMANCE_ROOT unset, taking the "root is not set" path its own
#     tests implement, and the missing artefact is named in the report.
#
# `CI_REQUIRE_FULL_ROOT=1` makes any deferral fatal. The candidate lane sets
# it, so an explicitly supplied schema v6 candidate must serve the entire
# package set or fail -- a candidate run can never be green while a package
# silently ran without its vectors.
#
# Usage:
#   suite-plan.sh <conformance-root> <evidence-dir>
#
# Writes into <evidence-dir>:
#   plan-served.txt    packages to run WITH the root exported
#   plan-deferred.txt  packages to run with the root UNSET
#   plan-excluded.txt  packages this platform does not execute
#   plan-assert.txt    `package<TAB>case` rejection assertions to run instead
#   suite-plan.txt     the human-readable report (also echoed to stdout)
#
# Environment:
#   CI_GATE_GOOS          GOOS to plan for   (default: `go env GOOS`)
#   CI_ROOT_ARTIFACTS     default: .github/ci/root-artifacts.tsv
#   CI_PLATFORM_EXCLUSIONS default: .github/ci/platform-exclusions.tsv
#   CI_REQUIRE_FULL_ROOT  1 = a deferral is fatal
#   GO                    go launcher        (default: go)

set -u

GO="${GO:-go}"
ARTEFACTS="${CI_ROOT_ARTIFACTS:-.github/ci/root-artifacts.tsv}"
EXCLUSIONS="${CI_PLATFORM_EXCLUSIONS:-.github/ci/platform-exclusions.tsv}"
QUALIFICATION_VECTOR='vectors/conformance-claim-v3-qualification.json'

if [ "$#" -ne 2 ]; then
	echo 'suite-plan: usage: suite-plan.sh <conformance-root> <evidence-dir>' >&2
	exit 2
fi
ROOT="$1"
EVIDENCE="$2"

[ -d "$ROOT" ]     || { echo "suite-plan: conformance root is not a directory: $ROOT" >&2; exit 2; }
[ -f "$ARTEFACTS" ]  || { echo "suite-plan: artefact table not found: $ARTEFACTS" >&2; exit 2; }
[ -f "$EXCLUSIONS" ] || { echo "suite-plan: exclusion table not found: $EXCLUSIONS" >&2; exit 2; }
mkdir -p "$EVIDENCE" || exit 2

GOOS="${CI_GATE_GOOS:-$("$GO" env GOOS)}"
[ -n "$GOOS" ] || { echo 'suite-plan: cannot determine GOOS' >&2; exit 2; }

MODULE="$(awk '/^module[ \t]/{print $2; exit}' go.mod 2>/dev/null)"
[ -n "$MODULE" ] || { echo 'suite-plan: cannot determine the module path' >&2; exit 2; }

SERVED="$EVIDENCE/plan-served.txt"
DEFERRED="$EVIDENCE/plan-deferred.txt"
EXCLUDED="$EVIDENCE/plan-excluded.txt"
ASSERT="$EVIDENCE/plan-assert.txt"
REPORT="$EVIDENCE/suite-plan.txt"
: >"$SERVED"; : >"$DEFERRED"; : >"$EXCLUDED"; : >"$ASSERT"

# --- the module's real package set -----------------------------------------
ALL="$EVIDENCE/plan-all-packages.txt"
"$GO" list ./... >"$ALL" 2>"$EVIDENCE/plan-go-list.err"
rc=$?
[ "$rc" -eq 0 ] || { echo "suite-plan: go list ./... failed (exit $rc)" >&2; sed -n '1,20p' "$EVIDENCE/plan-go-list.err" >&2; exit 2; }
[ -s "$ALL" ]   || { echo 'suite-plan: go list ./... produced no packages' >&2; exit 2; }

# --- which packages does this platform not execute? ------------------------
# Resolved by the one helper both this gate and `ledger-consistency.sh` use,
# so the two can never disagree about what is excluded here.
QV="$ROOT/$QUALIFICATION_VECTOR"
EXCLUSION_SET="$EVIDENCE/plan-exclusion-set.tsv"
bash "$(dirname "$0")/excluded-packages.sh" "$GOOS" "$ROOT" >"$EXCLUSION_SET"
rc=$?
[ "$rc" -eq 0 ] || { echo "suite-plan: cannot resolve the platform exclusions (exit $rc)" >&2; exit 2; }

{
	echo "suite-plan: GOOS=$GOOS"
	echo "suite-plan: root=$ROOT"
	if [ -f "$QV" ]; then
		echo "suite-plan: platform qualification read from the root's own $QUALIFICATION_VECTOR"
	else
		echo "suite-plan: the root publishes no $QUALIFICATION_VECTOR; the recorded default_excluded_on applies"
	fi
	echo ''
} >"$REPORT"

# --- partition -------------------------------------------------------------
FAIL=0
while IFS= read -r importpath; do
	[ -n "$importpath" ] || continue
	pkg="$importpath"
	if [ "$pkg" = "$MODULE" ]; then pkg="."
	else pkg="${pkg#"$MODULE"/}"
	fi

	# 1. platform exclusion, as resolved by excluded-packages.sh
	excl_case=''; source_of_truth=''
	while IFS="$(printf '\t')" read -r xpkg xcase xsource; do
		[ "$xpkg" = "$pkg" ] || continue
		excl_case="$xcase"; source_of_truth="$xsource"
	done <"$EXCLUSION_SET"

	if [ -n "$excl_case" ]; then
		printf '%s\n' "$importpath" >>"$EXCLUDED"
		printf '%s\t%s\n' "$importpath" "$excl_case" >>"$ASSERT"
		{
			echo "excl  $pkg"
			echo "      excluded on $GOOS by $source_of_truth"
			echo "      the exclusion is asserted by $excl_case, which still runs and must pass"
		} >>"$REPORT"
		continue
	fi

	# 2. root coverage
	missing=''
	while IFS="$(printf '\t')" read -r apkg artefacts _note; do
		case "$apkg" in ''|\#*) continue ;; esac
		[ "$apkg" = "$pkg" ] || continue
		oldifs="$IFS"; IFS=','
		for artefact in $artefacts; do
			[ -e "$ROOT/$artefact" ] || missing="$missing $artefact"
		done
		IFS="$oldifs"
	done <"$ARTEFACTS"

	if [ -n "$missing" ]; then
		printf '%s\n' "$importpath" >>"$DEFERRED"
		{
			echo "defer $pkg"
			echo "      the supplied root publishes none of:$missing"
			echo "      it runs with CURATOR_CONFORMANCE_ROOT unset, taking the path its own tests implement"
		} >>"$REPORT"
		if [ "${CI_REQUIRE_FULL_ROOT:-0}" = 1 ]; then
			echo "FAIL  $pkg was deferred, but this lane requires a root that serves the whole module" >>"$REPORT"
			echo "      missing:$missing" >>"$REPORT"
			FAIL=1
		fi
		continue
	fi

	printf '%s\n' "$importpath" >>"$SERVED"
done <"$ALL"

{
	echo ''
	echo "suite-plan: served=$(wc -l <"$SERVED" | tr -d ' ') deferred=$(wc -l <"$DEFERRED" | tr -d ' ') excluded=$(wc -l <"$EXCLUDED" | tr -d ' ')"
	if [ "${CI_REQUIRE_FULL_ROOT:-0}" = 1 ]; then
		echo 'suite-plan: CI_REQUIRE_FULL_ROOT=1 -- every package must be served by this root'
	fi
	echo $([ "$FAIL" -eq 0 ] && echo 'suite-plan: ok' || echo 'suite-plan: FAILED')
} >>"$REPORT"

cat "$REPORT"
exit "$FAIL"
