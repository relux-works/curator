package interop

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/relux-works/curator/internal/contextlock"
	"github.com/relux-works/curator/internal/contextresolve"
	"github.com/relux-works/curator/internal/pkgversion"
	"github.com/relux-works/curator/internal/protocoljson"
)

// vectorLock is the context-lock-v1 object as the vectors spell it.
type vectorLock struct {
	SchemaVersion int              `json:"schema_version"`
	Root          string           `json:"root"`
	Members       []map[string]any `json:"members"`
}

type vectorRequirement struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Range     string `json:"range"`
	Tag       string `json:"tag"`
	Revision  string `json:"revision"`
	Directory string `json:"directory"`
	Weight    *int64 `json:"weight"`
}

type vectorCommit struct {
	Version  string              `json:"version"`
	Weight   int64               `json:"weight"`
	Weights  map[string]int64    `json:"weights"`
	Requires []vectorRequirement `json:"requires"`
}

type vectorPackage struct {
	Kind    string                   `json:"kind"`
	Source  string                   `json:"source"`
	Tags    map[string]string        `json:"tags"`
	Commits map[string]*vectorCommit `json:"commits"`
}

type contextResolutionVector struct {
	ResolutionCases []struct {
		Name     string `json:"name"`
		Expected struct {
			Lock       *vectorLock      `json:"lock"`
			LockSHA256 string           `json:"lock_sha256"`
			Warnings   []map[string]any `json:"warnings"`
			Error      string           `json:"error"`
			Detail     map[string]any   `json:"detail"`
		} `json:"expected"`
		Input struct {
			Install struct {
				Name     string `json:"name"`
				Range    string `json:"range"`
				Tag      string `json:"tag"`
				Revision string `json:"revision"`
			} `json:"install"`
			OverlayDefaultWeight int64 `json:"overlay_default_weight"`
			Overlays             []struct {
				Name     string `json:"name"`
				Range    string `json:"range"`
				Tag      string `json:"tag"`
				Revision string `json:"revision"`
				Weight   *int64 `json:"weight"`
				Path     *struct {
					StateSHA256 string       `json:"state_sha256"`
					Manifest    vectorCommit `json:"manifest"`
				} `json:"path"`
			} `json:"overlays"`
			Packages map[string]vectorPackage `json:"packages"`
		} `json:"input"`
	} `json:"resolution_cases"`
	LockCases []struct {
		Name       string     `json:"name"`
		Lock       vectorLock `json:"lock"`
		CCJ1Bytes  string     `json:"ccj1_bytes"`
		ByteLength int        `json:"byte_length"`
		LockSHA256 string     `json:"lock_sha256"`
	} `json:"lock_cases"`
}

// vectorSource serves the in-memory package graph of one resolution case.
type vectorSource struct {
	packages map[string]vectorPackage
}

func (s vectorSource) Identity(_, name, _ string) (string, error) {
	pkg, ok := s.packages[name]
	if !ok {
		return "", fmt.Errorf("vector declares no package %s", name)
	}
	return pkg.Source, nil
}

func (s vectorSource) Candidates(_, name, _ string) ([]contextresolve.Candidate, error) {
	pkg := s.packages[name]
	var out []contextresolve.Candidate
	for tag, commit := range pkg.Tags {
		version, ok := pkgversion.ParseTag(tag)
		if !ok {
			continue
		}
		out = append(out, contextresolve.Candidate{Tag: tag, Version: version, Commit: commit})
	}
	return out, nil
}

func (s vectorSource) ResolveTag(_, name, _, tag string) (string, error) {
	commit, ok := s.packages[name].Tags[tag]
	if !ok {
		return "", fmt.Errorf("%s has no tag %s", name, tag)
	}
	return commit, nil
}

func (s vectorSource) Manifest(_, name, _, _, commit string) (*contextresolve.Package, error) {
	info := s.packages[name].Commits[commit]
	if info == nil {
		return &contextresolve.Package{}, nil
	}
	return convertVectorPackage(*info), nil
}

func convertVectorPackage(info vectorCommit) *contextresolve.Package {
	pkg := &contextresolve.Package{Version: info.Version, Weight: info.Weight, Weights: info.Weights}
	for _, requirement := range info.Requires {
		pkg.Requires = append(pkg.Requires, contextresolve.Requirement{
			Kind: requirement.Kind, Name: requirement.Name, Range: requirement.Range, Tag: requirement.Tag,
			Revision: requirement.Revision, Directory: requirement.Directory, Weight: requirement.Weight,
		})
	}
	return pkg
}

func lockToVector(t *testing.T, lock *contextlock.Lock) vectorLock {
	t.Helper()
	canonical, err := lock.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	var out vectorLock
	if err := json.Unmarshal(canonical, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func vectorLockToLock(t *testing.T, in vectorLock) *contextlock.Lock {
	t.Helper()
	payload, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := contextlock.Parse(payload)
	if err != nil {
		t.Fatalf("vector lock does not parse: %v", err)
	}
	return lock
}

// TestConformanceContextResolution drives contextresolve.Resolve — the
// production resolver behind `curator profile install` and `profile update` —
// against the resolution family, and contextlock's canonical bytes and lock
// hash against the lock family of vectors/context-versions.json.
func TestConformanceContextResolution(t *testing.T) {
	root := suiteRoot(t)
	vectorPath := filepath.Join(root, "vectors", "context-versions.json")
	payload, err := os.ReadFile(vectorPath)
	if errors.Is(err, os.ErrNotExist) {
		t.Skipf("conformance root %s publishes no vectors/context-versions.json (pre-environments suite; root-content)", root)
	}
	if err != nil {
		t.Fatal(err)
	}
	var vector contextResolutionVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatalf("decoding %s: %v", vectorPath, err)
	}
	if len(vector.ResolutionCases) == 0 || len(vector.LockCases) == 0 {
		t.Fatalf("%s declares an empty family", vectorPath)
	}

	t.Run("lock_cases", func(t *testing.T) {
		for _, tc := range vector.LockCases {
			t.Run(tc.Name, func(t *testing.T) {
				lock := vectorLockToLock(t, tc.Lock)
				canonical, err := lock.Canonical()
				if err != nil {
					t.Fatal(err)
				}
				if string(canonical) != tc.CCJ1Bytes || len(canonical) != tc.ByteLength {
					t.Fatalf("CCJ-1 bytes differ:\n got %s\nwant %s", canonical, tc.CCJ1Bytes)
				}
				if err := protocoljson.RequireCanonical(canonical); err != nil {
					t.Fatal(err)
				}
				hash, err := lock.Hash()
				if err != nil {
					t.Fatal(err)
				}
				if hash != tc.LockSHA256 {
					t.Fatalf("lock hash %s, want %s", hash, tc.LockSHA256)
				}
			})
		}
	})

	t.Run("resolution_cases", func(t *testing.T) {
		for _, tc := range vector.ResolutionCases {
			t.Run(tc.Name, func(t *testing.T) {
				input := contextresolve.Input{
					Root: contextresolve.Requirement{Name: tc.Input.Install.Name, Range: tc.Input.Install.Range,
						Tag: tc.Input.Install.Tag, Revision: tc.Input.Install.Revision},
					OverlayDefaultWeight: tc.Input.OverlayDefaultWeight,
				}
				for _, overlay := range tc.Input.Overlays {
					declaration := contextresolve.Overlay{Name: overlay.Name, Range: overlay.Range, Tag: overlay.Tag,
						Revision: overlay.Revision, Weight: overlay.Weight}
					if overlay.Path != nil {
						declaration.State = &contextresolve.StatePackage{StateHash: overlay.Path.StateSHA256,
							Manifest: convertVectorPackage(overlay.Path.Manifest)}
					}
					input.Overlays = append(input.Overlays, declaration)
				}
				result, err := contextresolve.Resolve(vectorSource{packages: tc.Input.Packages}, input)
				if tc.Expected.Error != "" {
					var resolveErr *contextresolve.Error
					if err == nil || !errors.As(err, &resolveErr) {
						t.Fatalf("resolved; want %s (err=%v)", tc.Expected.Error, err)
					}
					if resolveErr.Diagnostic != tc.Expected.Error {
						t.Fatalf("diagnostic %s, want %s (%v)", resolveErr.Diagnostic, tc.Expected.Error, err)
					}
					assertDetail(t, resolveErr, tc.Expected.Detail)
					return
				}
				if err != nil {
					t.Fatalf("resolve: %v", err)
				}
				got := lockToVector(t, result.Lock)
				if !reflect.DeepEqual(got, *tc.Expected.Lock) {
					gotJSON, _ := json.MarshalIndent(got, "", " ")
					wantJSON, _ := json.MarshalIndent(tc.Expected.Lock, "", " ")
					t.Fatalf("lock differs:\n got %s\nwant %s", gotJSON, wantJSON)
				}
				if result.LockHash != tc.Expected.LockSHA256 {
					t.Fatalf("lock hash %s, want %s", result.LockHash, tc.Expected.LockSHA256)
				}
				var gotWarnings []map[string]any
				for _, warning := range result.Warnings {
					entry := map[string]any{"diagnostic": warning.Diagnostic, "name": warning.Name}
					var requirers []any
					for _, requirer := range warning.Requirers {
						requirers = append(requirers, map[string]any{"requirer": requirer.Requirer, "weight": float64(requirer.Weight)})
					}
					entry["requirers"] = requirers
					gotWarnings = append(gotWarnings, entry)
				}
				if len(gotWarnings) != len(tc.Expected.Warnings) || (len(gotWarnings) > 0 && !reflect.DeepEqual(gotWarnings, tc.Expected.Warnings)) {
					t.Fatalf("warnings %v, want %v", gotWarnings, tc.Expected.Warnings)
				}
			})
		}
	})
}

func assertDetail(t *testing.T, err *contextresolve.Error, want map[string]any) {
	t.Helper()
	if want == nil {
		return
	}
	if name, ok := want["name"].(string); ok && err.Name != name {
		t.Fatalf("detail name %q, want %q", err.Name, name)
	}
	if raw, ok := want["candidates"].([]any); ok {
		var candidates []string
		for _, candidate := range raw {
			candidates = append(candidates, candidate.(string))
		}
		got := append([]string(nil), err.Candidates...)
		sort.Strings(got)
		sort.Strings(candidates)
		if !reflect.DeepEqual(got, candidates) && (len(got) != 0 || len(candidates) != 0) {
			t.Fatalf("detail candidates %v, want %v", err.Candidates, candidates)
		}
	}
	if raw, ok := want["requirers"].([]any); ok {
		var got []map[string]any
		for _, requirer := range err.Requirers {
			got = append(got, map[string]any{"requirer": requirer.Requirer, "constraint": requirer.Constraint})
		}
		for _, requirer := range err.WeightRequirers {
			got = append(got, map[string]any{"requirer": requirer.Requirer, "weight": float64(requirer.Weight)})
		}
		var wantRequirers []map[string]any
		for _, requirer := range raw {
			wantRequirers = append(wantRequirers, requirer.(map[string]any))
		}
		if !reflect.DeepEqual(got, wantRequirers) {
			t.Fatalf("detail requirers %v, want %v", got, wantRequirers)
		}
	}
	if tag, ok := want["tag"].(string); ok && err.Tag != tag {
		t.Fatalf("detail tag %q, want %q", err.Tag, tag)
	}
	if version, ok := want["manifest_version"].(string); ok && err.ManifestVersion != version {
		t.Fatalf("detail manifest_version %q, want %q", err.ManifestVersion, version)
	}
}
