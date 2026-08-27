#!/usr/bin/env bash
# Ledger consistency check.
#
# `platform-cases.tsv` claims, per case, which platforms must execute it. That
# claim is only worth something if the case is actually compiled into those
# platforms' builds. This script proves it WITHOUT running anything on those
# platforms: `go list` reports the exact test file set for a target GOOS, and a
# case exists on that GOOS iff it is declared in one of those files.
#
# It fails when:
#   * a row names a package that is not in the module;
#   * a row requires a case on a GOOS whose build does not compile that case
#     (a stale name, a moved build tag, a deleted test);
#   * a case IS compiled in on a GOOS that the row neither requires nor
#     tolerates, and that GOOS is not excluded for the package by the supplied
#     conformance root -- i.e. an undeclared platform.
#
# Usage:
#   ledger-consistency.sh <evidence-dir>
#
# The per-platform exclusion set comes from `excluded-packages.sh`, the same
# helper `suite-plan.sh` uses, so this gate and the run it describes can never
# disagree about what this platform executes.
#
# Environment:
#   CI_PLATFORM_CASES        ledger path      (default: .github/ci/platform-cases.tsv)
#   CI_LEDGER_GOOS           platforms to check (default: "linux darwin windows")
#   CURATOR_CONFORMANCE_ROOT optional; when set, its qualification vector is
#                            authoritative for the exclusions
#   GO                       go launcher      (default: go)

set -u

GO="${GO:-go}"
LEDGER="${CI_PLATFORM_CASES:-.github/ci/platform-cases.tsv}"
PLATFORMS="${CI_LEDGER_GOOS:-linux darwin windows}"

if [ "$#" -ne 1 ]; then
	echo 'ledger-consistency: usage: ledger-consistency.sh <evidence-dir>' >&2
	exit 2
fi
EVIDENCE="$1"
[ -f "$LEDGER" ] || { echo "ledger-consistency: ledger not found: $LEDGER" >&2; exit 2; }
mkdir -p "$EVIDENCE" || exit 2

MODULE="$(awk '/^module[ \t]/{print $2; exit}' go.mod 2>/dev/null)"
[ -n "$MODULE" ] || { echo 'ledger-consistency: cannot determine the module path' >&2; exit 2; }

INVENTORY="$EVIDENCE/case-inventory.tsv"
: >"$INVENTORY" || exit 2

# --- build the per-GOOS case inventory -------------------------------------
for goos in $PLATFORMS; do
	listing="$EVIDENCE/go-list-$goos.txt"
	GOOS="$goos" "$GO" list -e -f '{{.ImportPath}}	{{range .TestGoFiles}}{{.}} {{end}}{{range .XTestGoFiles}}{{.}} {{end}}' ./... >"$listing" 2>"$EVIDENCE/go-list-$goos.err"
	rc=$?
	if [ "$rc" -ne 0 ]; then
		echo "ledger-consistency: go list failed for GOOS=$goos (exit $rc)" >&2
		sed -n '1,20p' "$EVIDENCE/go-list-$goos.err" >&2
		exit 2
	fi
	[ -s "$listing" ] || { echo "ledger-consistency: go list produced nothing for GOOS=$goos" >&2; exit 2; }

	while IFS="$(printf '\t')" read -r importpath files; do
		[ -n "$importpath" ] || continue
		pkg="$importpath"
		if [ "$pkg" = "$MODULE" ]; then pkg="."
		else pkg="${pkg#"$MODULE"/}"
		fi
		dir=".${importpath#"$MODULE"}"
		for f in $files; do
			[ -f "$dir/$f" ] || continue
			# Top-level `func TestXxx(` declarations only; that is exactly what
			# `go test` turns into a case name.
			sed -n 's/^func \(Test[A-Za-z0-9_]*\)(.*/\1/p' "$dir/$f" |
				while IFS= read -r name; do
					[ -n "$name" ] && printf '%s\t%s\t%s\n' "$pkg" "$name" "$goos"
				done
		done
	done <"$listing" >>"$INVENTORY"
done

sort -u -o "$INVENTORY" "$INVENTORY" || exit 2
[ -s "$INVENTORY" ] || { echo 'ledger-consistency: empty case inventory' >&2; exit 2; }

# --- the per-platform exclusion set, from the shared helper ----------------
EXCLUSIONS_TSV="$EVIDENCE/exclusions.tsv"
: >"$EXCLUSIONS_TSV" || exit 2
for goos in $PLATFORMS; do
	bash "$(dirname "$0")/excluded-packages.sh" "$goos" "${CURATOR_CONFORMANCE_ROOT:-}" |
		while IFS="$(printf '\t')" read -r xpkg _case _src; do
			[ -n "$xpkg" ] && printf '%s\t%s\n' "$xpkg" "$goos"
		done
done >>"$EXCLUSIONS_TSV"

REPORT="$EVIDENCE/ledger-consistency.txt"

awk -v ledger="$LEDGER" -v inventory="$INVENTORY" -v platforms="$PLATFORMS" \
    -v exclusions="$EXCLUSIONS_TSV" '
function listed(set, want,   n, i, a) {
	if (set == "-" || set == "") return 0
	n = split(set, a, ",")
	for (i = 1; i <= n; i++) if (a[i] == want) return 1
	return 0
}
BEGIN {
	fail = 0
	np = split(platforms, P, " ")

	while ((getline line < exclusions) > 0) {
		split(line, e, "\t")
		if (e[1] != "") isexcluded[e[1] "\t" e[2]] = 1
	}
	close(exclusions)

	while ((getline line < inventory) > 0) {
		split(line, c, "\t")
		exists[c[1] "\t" c[2] "\t" c[3]] = 1
		pkgseen[c[1]] = 1
	}
	close(inventory)

	while ((getline line < ledger) > 0) {
		if (line ~ /^[ \t]*#/ || line ~ /^[ \t]*$/) continue
		split(line, f, "\t")
		pkg = f[1]; test = f[2]; must = f[3] ""; skipok = f[4] ""
		rows++

		# `Parent/*` and `Parent/child` rows describe a subtest of a case that
		# has its own row; only the parent name is checkable against the source.
		base = test
		sub("/.*$", "", base)
		subtest = (base != test)

		if (!(pkg in pkgseen)) {
			printf "FAIL  ledger row names a package that is not in the module: %s :: %s\n", pkg, test
			fail = 1
			continue
		}

		for (i = 1; i <= np; i++) {
			goos = P[i]
			here = exists[pkg "\t" base "\t" goos]

			if (listed(must, goos) && !here) {
				printf "FAIL  required on %s but not compiled into that build: %s :: %s\n", goos, pkg, test
				print  "      a rename, a moved build tag or a deleted test looks exactly like this."
				fail = 1
			}

			if (here && !listed(must, goos) && !listed(skipok, goos) && !subtest) {
				if (!isexcluded[pkg "\t" goos]) {
					printf "FAIL  compiled into the %s build but undeclared there: %s :: %s\n", goos, pkg, test
					print  "      add the platform to must_run_on, or to skip_allowed_on with a class and a reason."
					fail = 1
				} else {
					printf "excl  %s :: %s (package excluded on %s by the supplied root)\n", pkg, test, goos
				}
			}
		}
		if (!fail || 1) printf "ok    %s :: %s [must=%s skip=%s]\n", pkg, test, must, skipok
	}
	close(ledger)

	printf "\nledger-consistency: %d rows checked across %s\n", rows, platforms
	print (fail ? "ledger-consistency: FAILED" : "ledger-consistency: ok")
	exit fail
}
' </dev/null >"$REPORT" 2>&1
RC=$?
cat "$REPORT"
exit "$RC"
