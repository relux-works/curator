package environments

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/pkgversion"
)

// contextVersionsVector mirrors vectors/context-versions.json (environments
// §1.4). The resolution and lock families of the same file are consumed by
// context_resolution_test.go.
type contextVersionsVector struct {
	VersionCases []struct {
		Tag        string   `json:"tag"`
		Candidate  bool     `json:"candidate"`
		Version    string   `json:"version"`
		Major      int64    `json:"major"`
		Minor      int64    `json:"minor"`
		Patch      int64    `json:"patch"`
		Prerelease []string `json:"prerelease"`
	} `json:"version_cases"`
	OrderingCases []struct {
		Name              string   `json:"name"`
		Input             []string `json:"input"`
		ExpectedAscending []string `json:"expected_ascending"`
	} `json:"ordering_cases"`
	RangeCases []struct {
		Range          string     `json:"range"`
		Valid          bool       `json:"valid"`
		ComparatorSets [][]string `json:"comparator_sets"`
		Error          string     `json:"error"`
	} `json:"range_cases"`
	SatisfiesCases []struct {
		Range     string `json:"range"`
		Version   string `json:"version"`
		Satisfies bool   `json:"satisfies"`
	} `json:"satisfies_cases"`
}

func loadContextVersionsVector(t *testing.T) (string, contextVersionsVector) {
	t.Helper()
	root := suiteRoot(t)
	vectorPath := filepath.Join(root, "vectors", "context-versions.json")
	payload := requireFamily(t, root, "vectors/context-versions.json")
	var vector contextVersionsVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatalf("decoding %s: %v", vectorPath, err)
	}
	return vectorPath, vector
}

// TestConformanceContextVersions drives pkgversion — the production version,
// tag, range, and satisfaction reader of internal/contextresolve — against the
// version, ordering, range, and satisfies families of the suite's vector.
func TestConformanceContextVersions(t *testing.T) {
	vectorPath, vector := loadContextVersionsVector(t)
	if len(vector.VersionCases) == 0 || len(vector.RangeCases) == 0 || len(vector.SatisfiesCases) == 0 || len(vector.OrderingCases) == 0 {
		t.Fatalf("%s declares an empty family", vectorPath)
	}
	t.Run("version_cases", func(t *testing.T) {
		for _, tc := range vector.VersionCases {
			version, candidate := pkgversion.ParseTag(tc.Tag)
			if candidate != tc.Candidate {
				t.Fatalf("tag %q: candidate=%v, want %v", tc.Tag, candidate, tc.Candidate)
			}
			if !tc.Candidate {
				continue
			}
			prerelease := version.Prerelease
			if prerelease == nil {
				prerelease = []string{}
			}
			if version.String() != tc.Version || version.Major != tc.Major || version.Minor != tc.Minor || version.Patch != tc.Patch || !reflect.DeepEqual(prerelease, tc.Prerelease) {
				t.Fatalf("tag %q parsed as %+v, want %s", tc.Tag, version, tc.Version)
			}
		}
	})
	t.Run("ordering_cases", func(t *testing.T) {
		for _, tc := range vector.OrderingCases {
			var versions []pkgversion.Version
			for _, raw := range tc.Input {
				version, err := pkgversion.ParseVersion(raw)
				if err != nil {
					t.Fatalf("%s: %v", tc.Name, err)
				}
				versions = append(versions, version)
			}
			pkgversion.Sort(versions)
			var got []string
			for _, version := range versions {
				got = append(got, version.String())
			}
			if !reflect.DeepEqual(got, tc.ExpectedAscending) {
				t.Fatalf("%s: sorted %v, want %v", tc.Name, got, tc.ExpectedAscending)
			}
		}
	})
	t.Run("range_cases", func(t *testing.T) {
		for _, tc := range vector.RangeCases {
			parsed, err := pkgversion.ParseRange(tc.Range)
			if !tc.Valid {
				if err == nil {
					t.Fatalf("range %q parsed; want %s", tc.Range, tc.Error)
				}
				if tc.Error != "profile_source_invalid" || !errors.Is(err, pkgversion.ErrInvalidRange) {
					t.Fatalf("range %q: error %v, want %s", tc.Range, err, tc.Error)
				}
				continue
			}
			if err != nil {
				t.Fatalf("range %q: %v", tc.Range, err)
			}
			if got := parsed.ComparatorSets(); !reflect.DeepEqual(got, tc.ComparatorSets) {
				t.Fatalf("range %q: comparator sets %v, want %v", tc.Range, got, tc.ComparatorSets)
			}
		}
	})
	t.Run("satisfies_cases", func(t *testing.T) {
		for _, tc := range vector.SatisfiesCases {
			parsed, err := pkgversion.ParseRange(tc.Range)
			if err != nil {
				t.Fatalf("range %q: %v", tc.Range, err)
			}
			version, err := pkgversion.ParseVersion(tc.Version)
			if err != nil {
				t.Fatalf("version %q: %v", tc.Version, err)
			}
			if got := parsed.Satisfies(version); got != tc.Satisfies {
				t.Fatalf("%q satisfies %q = %v, want %v", tc.Version, tc.Range, got, tc.Satisfies)
			}
		}
	})
}
