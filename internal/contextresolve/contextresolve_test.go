package contextresolve

import (
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/pkgversion"
)

// stubSource is an in-memory Source for unit tests. commits maps tag to
// commit; manifests maps commit to the package manifest.
type stubSource struct {
	commits   map[string]string
	manifests map[string]*Package
}

func (s *stubSource) Identity(_ string, name, declared string) (string, error) {
	if declared != "" {
		return declared, nil
	}
	return "https://example.com/" + name, nil
}

func (s *stubSource) Candidates(_ string, name, _ string) ([]Candidate, error) {
	var out []Candidate
	for tag, commit := range s.commits {
		if !strings.HasPrefix(tag, name+"@") {
			continue
		}
		version, ok := pkgversion.ParseTag(tag[len(name)+1:])
		if !ok {
			continue
		}
		out = append(out, Candidate{Tag: tag[len(name)+1:], Version: version, Commit: commit})
	}
	return out, nil
}

func (s *stubSource) ResolveTag(_ string, name, _ string, tag string) (string, error) {
	commit, ok := s.commits[name+"@"+tag]
	if !ok {
		return "", &Error{Diagnostic: DiagVersionMismatch, Name: name, Tag: tag}
	}
	return commit, nil
}

func (s *stubSource) Manifest(_, name, _, _ string, commit string) (*Package, error) {
	manifest, ok := s.manifests[commit]
	if !ok {
		return nil, &Error{Diagnostic: DiagVersionMismatch, Name: name}
	}
	return manifest, nil
}

func resolveError(t *testing.T, err error) *Error {
	t.Helper()
	resolutionErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error %v is not a *Error", err)
	}
	return resolutionErr
}

// TestMinimalResolution pins the happy path through the production Resolve:
// one root with one range dependency resolves to a sorted lock with a hash.
func TestMinimalResolution(t *testing.T) {
	source := &stubSource{
		commits: map[string]string{"root@v1.0.0": strings.Repeat("1", 40), "lib@v1.2.0": strings.Repeat("2", 40)},
		manifests: map[string]*Package{
			strings.Repeat("1", 40): {Version: "1.0.0", Requires: []Requirement{
				{Kind: contextlock.KindContext, Name: "lib", Source: "https://example.com/lib", Range: "^1.0.0"},
			}},
			strings.Repeat("2", 40): {Version: "1.2.0"},
		},
	}
	result, err := Resolve(source, Input{
		Root: Requirement{Name: "root", Source: "https://example.com/root", Tag: "v1.0.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Lock.Root != "root" || len(result.Lock.Members) != 2 {
		t.Fatalf("lock %+v", result.Lock)
	}
	if err := result.Lock.Validate(); err != nil {
		t.Fatalf("resolved lock invalid: %v", err)
	}
	if len(result.LockHash) != len("sha256:")+64 {
		t.Fatalf("lock hash %q", result.LockHash)
	}
	if len(result.Members) != 2 {
		t.Fatalf("members %+v", result.Members)
	}
}

// TestRangeConflict narrows the conflict gate: disjoint ranges on one name
// must fail with context_range_conflict. A mutant that picks the higher
// candidate despite the conflict must fail this test.
func TestRangeConflict(t *testing.T) {
	source := &stubSource{
		commits: map[string]string{
			"root@v1.0.0": strings.Repeat("1", 40),
			"mid@v1.0.0":  strings.Repeat("3", 40),
			"lib@v1.0.0":  strings.Repeat("2", 40),
			"lib@v2.0.0":  strings.Repeat("4", 40),
		},
		manifests: map[string]*Package{
			strings.Repeat("1", 40): {Version: "1.0.0", Requires: []Requirement{
				{Kind: contextlock.KindContext, Name: "lib", Source: "https://example.com/lib", Range: "^1.0.0"},
				{Kind: contextlock.KindContext, Name: "mid", Source: "https://example.com/mid", Range: "^1.0.0"},
			}},
			strings.Repeat("3", 40): {Version: "1.0.0", Requires: []Requirement{
				{Kind: contextlock.KindContext, Name: "lib", Source: "https://example.com/lib", Range: "^2.0.0"},
			}},
			strings.Repeat("2", 40): {Version: "1.0.0"},
			strings.Repeat("4", 40): {Version: "2.0.0"},
		},
	}
	_, err := Resolve(source, Input{
		Root: Requirement{Name: "root", Source: "https://example.com/root", Tag: "v1.0.0"},
	})
	if err == nil {
		t.Fatal("disjoint ranges must conflict")
	}
	if resolutionErr := resolveError(t, err); resolutionErr.Diagnostic != DiagRangeConflict {
		t.Fatalf("diagnostic %q, want %s", resolutionErr.Diagnostic, DiagRangeConflict)
	}
}

// TestWeightConflict narrows the weights gate: two non-root direct
// requirers disagreeing on a member's edge weight, with the root map naming
// neither, must fail with context_weight_conflict. (A root edge of its own
// would name the member into the root map and downgrade this to a warning.)
func TestWeightConflict(t *testing.T) {
	heavy, light := int64(9), int64(1)
	source := &stubSource{
		commits: map[string]string{
			"root@v1.0.0": strings.Repeat("1", 40),
			"a@v1.0.0":    strings.Repeat("5", 40),
			"b@v1.0.0":    strings.Repeat("6", 40),
			"lib@v1.0.0":  strings.Repeat("2", 40),
		},
		manifests: map[string]*Package{
			strings.Repeat("1", 40): {Version: "1.0.0", Requires: []Requirement{
				{Kind: contextlock.KindContext, Name: "a", Source: "https://example.com/a", Range: "^1.0.0"},
				{Kind: contextlock.KindContext, Name: "b", Source: "https://example.com/b", Range: "^1.0.0"},
			}},
			strings.Repeat("5", 40): {Version: "1.0.0", Requires: []Requirement{
				{Kind: contextlock.KindContext, Name: "lib", Source: "https://example.com/lib", Range: "^1.0.0", Weight: &heavy},
			}},
			strings.Repeat("6", 40): {Version: "1.0.0", Requires: []Requirement{
				{Kind: contextlock.KindContext, Name: "lib", Source: "https://example.com/lib", Range: "^1.0.0", Weight: &light},
			}},
			strings.Repeat("2", 40): {Version: "1.0.0"},
		},
	}
	_, err := Resolve(source, Input{
		Root: Requirement{Name: "root", Source: "https://example.com/root", Tag: "v1.0.0"},
	})
	if err == nil {
		t.Fatal("disagreeing edge weights must conflict")
	}
	if resolutionErr := resolveError(t, err); resolutionErr.Diagnostic != DiagWeightConflict {
		t.Fatalf("diagnostic %q, want %s", resolutionErr.Diagnostic, DiagWeightConflict)
	}
}

// TestNonRootWeightsAreRejected narrows the root-only weights gate: a
// non-root manifest carrying a weights map must fail with
// context_weights_not_root.
func TestNonRootWeightsAreRejected(t *testing.T) {
	source := &stubSource{
		commits: map[string]string{"root@v1.0.0": strings.Repeat("1", 40), "lib@v1.0.0": strings.Repeat("2", 40)},
		manifests: map[string]*Package{
			strings.Repeat("1", 40): {Version: "1.0.0", Requires: []Requirement{
				{Kind: contextlock.KindContext, Name: "lib", Source: "https://example.com/lib", Range: "^1.0.0"},
			}},
			strings.Repeat("2", 40): {Version: "1.0.0", Weights: map[string]int64{"x": 1}},
		},
	}
	_, err := Resolve(source, Input{
		Root: Requirement{Name: "root", Source: "https://example.com/root", Tag: "v1.0.0"},
	})
	if err == nil {
		t.Fatal("non-root weights must fail")
	}
	if resolutionErr := resolveError(t, err); resolutionErr.Diagnostic != DiagWeightsNotRoot {
		t.Fatalf("diagnostic %q, want %s", resolutionErr.Diagnostic, DiagWeightsNotRoot)
	}
}

// TestDuplicateOverlayIsCompositionInvalid checks the overlay gate.
func TestDuplicateOverlayIsCompositionInvalid(t *testing.T) {
	source := &stubSource{}
	_, err := Resolve(source, Input{
		Root:     Requirement{Name: "root", Source: "https://example.com/root", Tag: "v1.0.0"},
		Overlays: []Overlay{{Name: "root", Tag: "v1.0.0"}},
	})
	if err == nil {
		t.Fatal("overlay repeating the root must fail")
	}
	if resolutionErr := resolveError(t, err); resolutionErr.Diagnostic != DiagCompositionInvalid {
		t.Fatalf("diagnostic %q, want %s", resolutionErr.Diagnostic, DiagCompositionInvalid)
	}
}
