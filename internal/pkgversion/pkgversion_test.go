package pkgversion

import (
	"strings"
	"testing"
)

// The production entry points under test: ParseTag, ParseVersion,
// ParseRange, Range.Satisfies, Compare, Sort.

// TestStrictTagsAreCandidates parses the happy path: v-prefixed strict
// SemVer without build metadata.
func TestStrictTagsAreCandidates(t *testing.T) {
	version, ok := ParseTag("v1.2.3")
	if !ok {
		t.Fatal("v1.2.3 is not a candidate")
	}
	if version.Major != 1 || version.Minor != 2 || version.Patch != 3 {
		t.Fatalf("parsed %+v", version)
	}
	prerelease, ok := ParseTag("v2.0.0-rc.1")
	if !ok || !prerelease.IsPrerelease() {
		t.Fatalf("v2.0.0-rc.1 candidate=%v prerelease=%v", ok, prerelease)
	}
}

// TestBuildMetadataIsNotACandidate narrows the tag gate: build metadata
// must not parse, so "v1.2.3+build" is silently outside every range.
func TestBuildMetadataIsNotACandidate(t *testing.T) {
	if _, ok := ParseTag("v1.2.3+build"); ok {
		t.Fatal("v1.2.3+build must not be a candidate")
	}
	if _, err := ParseVersion("1.2.3+build"); err == nil {
		t.Fatal("ParseVersion must reject build metadata")
	}
}

// TestHyphenRangesAreRejected narrows the range gate: hyphen ranges are
// outside the closed grammar. A mutant that admits "1.2.3 - 2.0.0" must
// fail this test.
func TestHyphenRangesAreRejected(t *testing.T) {
	for _, text := range []string{"1.2.3 - 2.3.4", "1.2.x - 2.3.x", "v1.2.3 - v2.0.0"} {
		if _, err := ParseRange(text); err == nil {
			t.Fatalf("range %q must not parse", text)
		}
	}
}

// TestVersionPrefixInsideRangesIsRejected narrows the range gate on the
// "v" rule: no "v" may appear inside a range.
func TestVersionPrefixInsideRangesIsRejected(t *testing.T) {
	for _, text := range []string{"v1.2.3", "^v1.2", ">=v1.0.0 <2.0.0"} {
		if _, err := ParseRange(text); err == nil {
			t.Fatalf("range %q must not parse", text)
		}
	}
}

// TestLatestIsStar checks that "latest" spells "*".
func TestLatestIsStar(t *testing.T) {
	r, err := ParseRange("latest")
	if err != nil {
		t.Fatal(err)
	}
	version, _ := ParseVersion("0.0.1")
	if !r.Satisfies(version) {
		t.Fatal("latest must satisfy 0.0.1")
	}
}

// TestCaretZeroSemantics pins the 0.x/0.0.x caret upper bounds.
func TestCaretZeroSemantics(t *testing.T) {
	r, err := ParseRange("^0.2.3")
	if err != nil {
		t.Fatal(err)
	}
	inside, _ := ParseVersion("0.2.9")
	outside, _ := ParseVersion("0.3.0")
	if !r.Satisfies(inside) || r.Satisfies(outside) {
		t.Fatalf("^0.2.3 admits 0.2.9=%v 0.3.0=%v", r.Satisfies(inside), r.Satisfies(outside))
	}
	r, err = ParseRange("^0.0.3")
	if err != nil {
		t.Fatal(err)
	}
	patch, _ := ParseVersion("0.0.4")
	if r.Satisfies(patch) {
		t.Fatal("^0.0.3 must not admit 0.0.4")
	}
}

// TestExclusiveUpperBoundSpelling checks the -0 bound: <2.0.0 admits
// 2.0.0-0's predecessors only, and 2.0.0 itself is outside.
func TestExclusiveUpperBoundSpelling(t *testing.T) {
	r, err := ParseRange(">=1.0.0 <2.0.0-0")
	if err != nil {
		t.Fatal(err)
	}
	inside, _ := ParseVersion("1.9.9")
	edge, _ := ParseVersion("2.0.0")
	if !r.Satisfies(inside) || r.Satisfies(edge) {
		t.Fatal("exclusive upper bound misbehaves")
	}
}

// TestPrereleaseNeedsSameTriple narrows the prerelease gate: a prerelease
// version satisfies a set only when some comparator names a prerelease on
// the same major.minor.patch. A mutant that drops the same-triple rule
// must fail this test.
func TestPrereleaseNeedsSameTriple(t *testing.T) {
	r, err := ParseRange(">=1.0.0-alpha <2.0.0-0")
	if err != nil {
		t.Fatal(err)
	}
	admitted, _ := ParseVersion("1.0.0-beta")
	excluded, _ := ParseVersion("1.5.0-alpha")
	if !r.Satisfies(admitted) {
		t.Fatal("1.0.0-beta must satisfy a set naming 1.0.0 prereleases")
	}
	if r.Satisfies(excluded) {
		t.Fatal("1.5.0-alpha must not satisfy a set with no 1.5.0 prerelease comparator")
	}
	star, err := ParseRange("*")
	if err != nil {
		t.Fatal(err)
	}
	if star.Satisfies(excluded) {
		t.Fatal("* must not admit a prerelease")
	}
}

// TestTotalOrder pins Compare and Sort to SemVer precedence.
func TestTotalOrder(t *testing.T) {
	tags := []string{"v2.0.0", "v1.0.0-alpha", "v1.0.0", "v1.0.0-alpha.1", "v0.9.9"}
	var versions []Version
	for _, tag := range tags {
		version, ok := ParseTag(tag)
		if !ok {
			t.Fatalf("%s is not a candidate", tag)
		}
		versions = append(versions, version)
	}
	Sort(versions)
	var got []string
	for _, version := range versions {
		got = append(got, "v"+version.String())
	}
	want := []string{"v0.9.9", "v1.0.0-alpha", "v1.0.0-alpha.1", "v1.0.0", "v2.0.0"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order %v, want %v", got, want)
		}
	}
}

// TestMalformedRangesAreRejected covers the negative shape: empty sets,
// double spaces, and tab separators never parse.
func TestMalformedRangesAreRejected(t *testing.T) {
	for _, text := range []string{"", " ", "^1.2.3 ||", "|| ^1.0.0", ">=1.0.0\t<2.0.0", ">=1.0.0  <2.0.0"} {
		if _, err := ParseRange(text); err == nil {
			t.Fatalf("range %q must not parse", text)
		} else if !strings.Contains(err.Error(), "profile_source_invalid") {
			t.Fatalf("range %q error %q carries no profile_source_invalid class", text, err)
		}
	}
}
