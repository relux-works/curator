package envmarker

import (
	"encoding/json"
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
// that accepts version 3 must fail this test.
func TestUnsupportedVersionIsRejected(t *testing.T) {
	payload, err := testMarker().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	future := strings.Replace(string(payload), `"version": 1`, `"version": 3`, 1)
	if _, err := Parse([]byte(future)); err == nil {
		t.Fatal("version 3 must be rejected")
	} else if !strings.Contains(err.Error(), DiagMarkerInvalid) {
		t.Fatalf("error %v carries no %s", err, DiagMarkerInvalid)
	}
}

func TestSchema2PathlessCredentialRecordOmitsPath(t *testing.T) {
	marker := testMarker()
	marker.Version = VersionV2
	marker.Mode = ModeManagedHome
	entries := []Passthrough{{
		Isolation: "shared", Strategy: "keyring-preferred", SourceRole: "native",
		Backend: "ambient", BackendVersion: "0.153.2", Provenance: "provisioned",
	}}
	marker.Passthrough = &entries
	seeds := []string{}
	marker.Seeds = &seeds
	payload, err := marker.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	passthrough, _ := decoded["passthrough"].([]any)
	entry, _ := passthrough[0].(map[string]any)
	if _, exists := entry["path"]; exists {
		t.Fatalf("ambient credentials are recorded without path: %s", payload)
	}
	if parsed.Version != VersionV2 || (*parsed.Passthrough)[0].Backend != "ambient" {
		t.Fatalf("parsed schema-2 credential record: %+v", parsed.Passthrough)
	}
	decoded["passthrough"].([]any)[0].(map[string]any)["path"] = ""
	invalid, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(invalid); err == nil {
		t.Fatal("a present empty path is not a pathless record; linkless records omit the member")
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
