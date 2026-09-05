package envmarker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Production entry points under test: Parse, Read, Marshal, Validate,
// SortedSurfaceKeys.

func testMarker() *Marker {
	copies := []Copy{{Path: "CLAUDE.md", Reason: ReasonClaudeCodeRootContext}}
	return &Marker{
		Version: 1,
		Profile: Profile{Name: "company", Root: "acme", Kind: "git", LockSHA256: strings.Repeat("a", 64),
			Source: "https://example.com/acme", Requirement: &Requirement{Range: "^1.0.0"}},
		Members:    []Member{{Name: "acme", Version: "1.0.0", Commit: strings.Repeat("b", 40), Weight: 3}},
		Precedence: Precedence{Winner: "higher-weight", Placement: "winner-last"},
		Mode:       ModeLinked,
		Surfaces: map[string]Surface{
			SurfaceRootContext: {Paths: []string{"CLAUDE.md"}, Form: "monolithic", ContentSHA256: "sha256:" + strings.Repeat("c", 64), Copies: &copies},
		},
	}
}

// TestRoundTrip checks Marshal validates and Parse accepts what Marshal emits.
func TestRoundTrip(t *testing.T) {
	payload, err := testMarker().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(payload), "\n") {
		t.Fatal("marker bytes carry no trailing LF")
	}
	parsed, err := Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Profile.Name != "company" || parsed.Mode != ModeLinked {
		t.Fatalf("marker %+v", parsed)
	}
	if keys := parsed.SortedSurfaceKeys(); len(keys) != 1 || keys[0] != SurfaceRootContext {
		t.Fatalf("surface keys %v", keys)
	}
}

// TestUnsupportedVersionIsRejected narrows the version gate: readers reject
// an unsupported marker version and never infer newer semantics. A mutant
// that accepts version 2 must fail this test.
func TestUnsupportedVersionIsRejected(t *testing.T) {
	payload, err := testMarker().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	future := strings.Replace(string(payload), `"version": 1`, `"version": 2`, 1)
	if _, err := Parse([]byte(future)); err == nil {
		t.Fatal("version 2 must be rejected")
	} else if !strings.Contains(err.Error(), DiagMarkerInvalid) {
		t.Fatalf("error %v carries no %s", err, DiagMarkerInvalid)
	}
}

// TestUnknownFieldsAreRejected narrows the strict-reader gate: unknown
// fields must fail, so newer semantics are never silently inferred.
func TestUnknownFieldsAreRejected(t *testing.T) {
	payload, err := testMarker().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	withExtra := strings.Replace(string(payload), `"mode": "linked"`, `"mode": "linked", "future": true`, 1)
	if _, err := Parse([]byte(withExtra)); err == nil {
		t.Fatal("unknown fields must be rejected")
	}
}

// TestReadDistinguishesAbsence checks Read reports absence as (nil, nil)
// and an invalid marker as a fail-closed environment_marker_invalid error.
func TestReadDistinguishesAbsence(t *testing.T) {
	home := t.TempDir()
	marker, err := Read(home)
	if err != nil || marker != nil {
		t.Fatalf("absent marker = (%v, %v)", marker, err)
	}
	if err := os.WriteFile(filepath.Join(home, Name), []byte("{nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(home); err == nil || !strings.Contains(err.Error(), DiagMarkerInvalid) {
		t.Fatalf("invalid marker must fail as %s, got %v", DiagMarkerInvalid, err)
	}
}

// TestCopiedModeOmitsCopies checks the copied-mode marker shape: no copies
// member is recorded.
func TestCopiedModeOmitsCopies(t *testing.T) {
	marker := testMarker()
	marker.Mode = ModeCopied
	marker.Surfaces = map[string]Surface{
		SurfaceRootContext: {Paths: []string{"AGENTS.md"}, Form: "monolithic", ContentSHA256: "sha256:" + strings.Repeat("c", 64)},
	}
	payload, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "copies") {
		t.Fatalf("copied marker must not record copies:\n%s", payload)
	}
}
