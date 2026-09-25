package environments

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
	"github.com/relux-works/curator/internal/pkgversion"
)

// contextVersionsVector mirrors vectors/context-versions.json (environments
// §1.4). The resolution and lock families of the same file are consumed by
// context_resolution_test.go.
type contextVersionCase struct {
	Tag        string   `json:"tag"`
	Candidate  bool     `json:"candidate"`
	Version    string   `json:"version"`
	Major      int64    `json:"major"`
	Minor      int64    `json:"minor"`
	Patch      int64    `json:"patch"`
	Prerelease []string `json:"prerelease"`
}

type contextOrderingCase struct {
	Name              string   `json:"name"`
	Input             []string `json:"input"`
	ExpectedAscending []string `json:"expected_ascending"`
}

type contextRangeCase struct {
	Range          string     `json:"range"`
	Valid          bool       `json:"valid"`
	ComparatorSets [][]string `json:"comparator_sets"`
	Error          string     `json:"error"`
}

type contextSatisfiesCase struct {
	Range     string `json:"range"`
	Version   string `json:"version"`
	Satisfies bool   `json:"satisfies"`
}

type contextVersionsVector struct {
	VersionCases   []contextVersionCase   `json:"version_cases"`
	OrderingCases  []contextOrderingCase  `json:"ordering_cases"`
	RangeCases     []contextRangeCase     `json:"range_cases"`
	SatisfiesCases []contextSatisfiesCase `json:"satisfies_cases"`
}

func loadContextVersionsVector(t *testing.T) (string, contextVersionsVector) {
	t.Helper()
	root := suiteRoot(t)
	vectorPath := rootPath(t, root, "vectors/context-versions.json")
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
	conformancecoverage.Run(t, "context-versions/version-cases", vector.VersionCases,
		func(tc contextVersionCase) string { return "version/" + tc.Tag }, func(t *testing.T, tc contextVersionCase) {
			version, candidate := pkgversion.ParseTag(tc.Tag)
			if candidate != tc.Candidate {
				t.Fatalf("tag %q: candidate=%v, want %v", tc.Tag, candidate, tc.Candidate)
			}
			if !tc.Candidate {
				return
			}
			prerelease := version.Prerelease
			if prerelease == nil {
				prerelease = []string{}
			}
			if version.String() != tc.Version || version.Major != tc.Major || version.Minor != tc.Minor || version.Patch != tc.Patch || !reflect.DeepEqual(prerelease, tc.Prerelease) {
				t.Fatalf("tag %q parsed as %+v, want %s", tc.Tag, version, tc.Version)
			}
		})
	conformancecoverage.Run(t, "context-versions/ordering-cases", vector.OrderingCases,
		func(tc contextOrderingCase) string { return "ordering/" + tc.Name }, func(t *testing.T, tc contextOrderingCase) {
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
		})
	conformancecoverage.Run(t, "context-versions/range-cases", vector.RangeCases,
		func(tc contextRangeCase) string { return "range/" + tc.Range }, func(t *testing.T, tc contextRangeCase) {
			parsed, err := pkgversion.ParseRange(tc.Range)
			if !tc.Valid {
				if err == nil {
					t.Fatalf("range %q parsed; want %s", tc.Range, tc.Error)
				}
				if tc.Error != "profile_source_invalid" || !errors.Is(err, pkgversion.ErrInvalidRange) {
					t.Fatalf("range %q: error %v, want %s", tc.Range, err, tc.Error)
				}
				return
			}
			if err != nil {
				t.Fatalf("range %q: %v", tc.Range, err)
			}
			if got := parsed.ComparatorSets(); !reflect.DeepEqual(got, tc.ComparatorSets) {
				t.Fatalf("range %q: comparator sets %v, want %v", tc.Range, got, tc.ComparatorSets)
			}
		})
	conformancecoverage.Run(t, "context-versions/satisfies-cases", vector.SatisfiesCases,
		func(tc contextSatisfiesCase) string { return "satisfies/" + tc.Range + "@" + tc.Version }, func(t *testing.T, tc contextSatisfiesCase) {
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
		})
}
