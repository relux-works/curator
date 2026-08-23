#!/usr/bin/env bash
# Candidate protocol-suite identity and evidence recorder.
#
# The committed `SPEC_PIN` in `.github/workflows/ci.yml` is the only default
# protocol input and always stays on the currently qualified released
# revision.  A schema v6 candidate suite is never committed and never becomes a
# default: it enters exactly one non-default job, through an explicit
# caller-supplied full revision or a pre-materialised root, and everything it
# produces is stamped as candidate evidence.
#
# This script:
#   * requires exactly one candidate source, rejecting an ambiguous revision
#     plus root before either can be checked out or recorded;
#   * rejects any candidate input that is not immutable -- a branch, a tag, a
#     short hash, a placeholder or a guess;
#   * refuses a candidate root that is byte-identical to the committed pin's
#     root, so candidate evidence can never be passed off as pin evidence;
#   * records the exact revision, the manifest digest, the whole-tree digest
#     and the file count;
#   * verifies those against caller-supplied expected digests when given;
#   * writes an evidence file whose header states, in the artifact itself, that
#     it is neither a published release nor a conformance claim.
#
# Usage:
#   candidate-suite.sh verify-inputs <revision> <root>
#   candidate-suite.sh verify-ref  <revision>
#   candidate-suite.sh record      <root> <evidence-dir>
#
# Environment (optional, all fail-closed when set and mismatched):
#   CANDIDATE_REF                     full 40-hex revision the root came from
#   CANDIDATE_EXPECTED_MANIFEST_SHA256  expected sha256 of <root>/manifest.json
#   CANDIDATE_EXPECTED_TREE_SHA256      expected aggregate digest of the tree
#   SPEC_PIN                          committed released pin, for the anti-confusion check

set -u

die() { echo "candidate-suite: $*" >&2; exit 1; }

sha256_of() {
	local file="$1" output digest
	if command -v shasum >/dev/null 2>&1; then
		output="$(shasum -a 256 <"$file")" || die "sha256 failed for $file"
	elif command -v sha256sum >/dev/null 2>&1; then
		output="$(sha256sum <"$file")" || die "sha256 failed for $file"
	else
		die 'neither shasum nor sha256sum is available'
	fi

	# Hash stdin rather than a filename. Git for Windows shasum prefixes its
	# output line with a backslash when the printed filename needs escaping;
	# parsing that line's first field would turn the escape marker into part of
	# the digest. Keep the shape check as a fail-closed guard for any hash tool
	# that still returns non-canonical output.
	read -r digest _ <<<"$output"
	case "$digest" in
		*[!0-9a-f]*) die "hash tool returned a non-canonical sha256 digest for $file: '$digest'" ;;
	esac
	[ "${#digest}" -eq 64 ] || die "hash tool returned a non-canonical sha256 digest for $file: '$digest'"
	printf '%s\n' "$digest"
}

# A revision and a pre-materialised root are independent identities. Accepting
# both would let evidence name the revision while measuring the unrelated root,
# so the caller must supply exactly one source.
verify_inputs() {
	ref="$1"
	root="$2"
	if [ -n "$ref" ] && [ -n "$root" ]; then
		die 'candidate revision and candidate root are mutually exclusive; supply exactly one'
	fi
	if [ -z "$ref" ] && [ -z "$root" ]; then
		die 'candidate revision and candidate root are both empty; supply exactly one'
	fi
	echo 'candidate-suite: candidate input accepted (exactly one revision or root)'
}

# A candidate revision must be a full, lowercase, 40-hex commit id.  Branches,
# tags, `HEAD`, short hashes and placeholders are all rejected by shape, before
# anything is fetched.
verify_ref() {
	ref="$1"
	[ -n "$ref" ] || die 'candidate revision is empty'
	case "$ref" in
		*[!0-9a-f]*) die "candidate revision is not lowercase hex: '$ref' (a branch, tag or placeholder is never a valid candidate pin)" ;;
	esac
	[ "${#ref}" -eq 40 ] || die "candidate revision must be a full 40-character commit id, got ${#ref} characters: '$ref'"
	case "$ref" in
		0000000000000000000000000000000000000000) die 'candidate revision is the null commit' ;;
	esac
	if [ -n "${SPEC_PIN:-}" ] && [ "$ref" = "$SPEC_PIN" ]; then
		die "candidate revision equals the committed released pin $SPEC_PIN; a candidate run must not impersonate the qualified pin"
	fi
	echo "candidate-suite: revision accepted (immutable, full 40-hex): $ref"
}

record() {
	root="$1"
	evidence="$2"
	[ -d "$root" ] || die "candidate root is not a directory: $root"
	[ -f "$root/manifest.json" ] || die "candidate root has no manifest.json: $root"
	mkdir -p "$evidence" || die "cannot create evidence directory: $evidence"

	manifest_sha="$(sha256_of "$root/manifest.json")" || exit 1

	# Whole-tree identity.  Enumerated in three separately status-checked
	# stages: a partial-then-failing `find` piped into a digest would produce
	# a short but internally consistent answer, and that answer is exactly
	# the thing this digest exists to make impossible.
	paths="$evidence/candidate-paths.txt"
	sorted="$evidence/candidate-paths.sorted.txt"
	digests="$evidence/candidate-file-digests.txt"

	( cd "$root" && find . -type f -print ) >"$paths" || die 'candidate enumeration failed (find)'
	sort "$paths" >"$sorted" || die 'candidate enumeration failed (sort)'
	files="$(wc -l <"$sorted" | tr -d ' ')"
	[ "$files" -gt 0 ] || die 'candidate root enumerated zero files'

	: >"$digests"
	while IFS= read -r rel; do
		printf '%s  %s\n' "$(sha256_of "$root/$rel")" "$rel" >>"$digests" || die "digest failed for $rel"
	done <"$sorted"
	[ "$(wc -l <"$digests" | tr -d ' ')" -eq "$files" ] || die 'digest count does not match the enumerated file count'

	tree_sha="$(sha256_of "$digests")" || exit 1

	protocol="$(awk -F'"' '/"protocol_version"[ \t]*:/{print $4; exit}' "$root/manifest.json")"
	[ -n "$protocol" ] || protocol='<unreadable>'

	ref="${CANDIDATE_REF:-<pre-materialised root, no revision supplied>}"

	# Anti-confusion: a candidate whose manifest is byte-identical to the
	# committed pin's manifest is the pin, not a candidate.
	if [ -n "${PIN_MANIFEST_SHA256:-}" ] && [ "$manifest_sha" = "$PIN_MANIFEST_SHA256" ]; then
		die "candidate manifest digest equals the committed pin's manifest digest ($manifest_sha); this is the released suite, not a candidate"
	fi

	if [ -n "${CANDIDATE_EXPECTED_MANIFEST_SHA256:-}" ]; then
		exp="${CANDIDATE_EXPECTED_MANIFEST_SHA256#sha256:}"
		[ "$manifest_sha" = "$exp" ] || die "candidate manifest digest mismatch: expected $exp, measured $manifest_sha -- this is a different candidate, never re-baseline it silently"
		echo "candidate-suite: manifest digest matches the supplied expectation"
	fi
	if [ -n "${CANDIDATE_EXPECTED_TREE_SHA256:-}" ]; then
		exp="${CANDIDATE_EXPECTED_TREE_SHA256#sha256:}"
		[ "$tree_sha" = "$exp" ] || die "candidate tree digest mismatch: expected $exp, measured $tree_sha"
		echo "candidate-suite: tree digest matches the supplied expectation"
	fi

	out="$evidence/candidate-suite-identity.txt"
	{
		echo '# CANDIDATE PROTOCOL SUITE EVIDENCE -- NOT A RELEASE'
		echo '#'
		echo '# This file records the identity of an explicitly supplied, non-default'
		echo '# candidate protocol suite. It is NOT a published release, NOT a release'
		echo '# claim and NOT a conformance claim, and it must not be cited as any of'
		echo '# them. The committed default pin is unchanged by this run; promoting it'
		echo '# is owned by TASK-260720-38l1sy after TASK-260720-25d05o qualifies the'
		echo '# release.'
		echo '#'
		echo "candidate_revision      $ref"
		echo "candidate_root          $root"
		echo "protocol_version        $protocol"
		echo "manifest_sha256         sha256:$manifest_sha"
		echo "tree_sha256             sha256:$tree_sha"
		echo "file_count              $files"
		echo "committed_released_pin  ${SPEC_PIN:-<unset>}"
		echo "runner_goos             $(go env GOOS 2>/dev/null || echo '<unknown>')"
		echo "runner_goarch           $(go env GOARCH 2>/dev/null || echo '<unknown>')"
		echo 'evidence_class          candidate-only'
		echo 'release_claim           none'
		echo 'conformance_claim       none'
	} >"$out" || die "cannot write $out"

	cat "$out"
	echo "candidate-suite: identity recorded at $out"
}

[ "$#" -ge 1 ] || die 'usage: candidate-suite.sh verify-inputs <revision> <root> | verify-ref <revision> | record <root> <evidence-dir>'
cmd="$1"
shift
case "$cmd" in
	verify-inputs) [ "$#" -eq 2 ] || die 'usage: candidate-suite.sh verify-inputs <revision> <root>'; verify_inputs "$1" "$2" ;;
	verify-ref) [ "$#" -eq 1 ] || die 'usage: candidate-suite.sh verify-ref <revision>'; verify_ref "$1" ;;
	record)     [ "$#" -eq 2 ] || die 'usage: candidate-suite.sh record <root> <evidence-dir>'; record "$1" "$2" ;;
	*)          die "unknown subcommand: $cmd" ;;
esac
