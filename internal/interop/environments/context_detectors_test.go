package environments

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/contextaudit"
	"github.com/relux-works/curator/internal/contextpkg"
)

type contextDetectorsVector struct {
	PatternClasses []struct {
		Pattern           string `json:"pattern"`
		Regexp            string `json:"regexp"`
		Group             int    `json:"group"`
		PlaceholderPrefix string `json:"placeholder_prefix"`
	} `json:"pattern_classes"`
	Scope []string `json:"scope"`
	Cases []struct {
		Name           string            `json:"name"`
		PackageKind    string            `json:"package_kind"`
		Pin            string            `json:"pin"`
		ContentHashPin bool              `json:"content_hash_pin"`
		Files          map[string]string `json:"files"`
		Waivers        []struct {
			Pin    string `json:"pin"`
			File   string `json:"file"`
			Span   [2]int `json:"span"`
			Reason string `json:"reason"`
		} `json:"waivers"`
		Expected struct {
			Installs bool `json:"installs"`
			Findings []struct {
				Class        string `json:"class"`
				File         string `json:"file"`
				Pattern      string `json:"pattern"`
				Severity     string `json:"severity"`
				Span         [2]int `json:"span"`
				Waived       bool   `json:"waived"`
				WaiverReason string `json:"waiver_reason"`
			} `json:"findings"`
			Warnings []map[string]any `json:"warnings"`
		} `json:"expected"`
	} `json:"cases"`
}

// TestConformanceContextDetectors drives contextaudit.Detect — the production
// detector behind every profile install and update — against
// vectors/context-detectors.json, over a snapshot written to disk exactly as
// the store would hold it.
func TestConformanceContextDetectors(t *testing.T) {
	root := suiteRoot(t)
	vectorPath := filepath.Join(root, "vectors", "context-detectors.json")
	payload := requireFamily(t, root, "vectors/context-detectors.json")
	var vector contextDetectorsVector
	if err := json.Unmarshal(payload, &vector); err != nil {
		t.Fatalf("decoding %s: %v", vectorPath, err)
	}
	if len(vector.Cases) == 0 {
		t.Fatalf("%s declares no cases", vectorPath)
	}
	t.Run("pattern_classes", func(t *testing.T) {
		if len(vector.PatternClasses) != len(contextaudit.Classes) {
			t.Fatalf("the vector declares %d classes, the detector %d", len(vector.PatternClasses), len(contextaudit.Classes))
		}
		for index, class := range vector.PatternClasses {
			got := contextaudit.Classes[index]
			if got.Pattern != class.Pattern || got.Regexp.String() != class.Regexp || got.Group != class.Group || got.PlaceholderPrefix != class.PlaceholderPrefix {
				t.Fatalf("class %d: detector %+v, vector %+v", index, got, class)
			}
		}
		for _, path := range vector.Scope {
			if path == "context/**" {
				if !contextaudit.InScope("context/deep/file.md") {
					t.Fatal("context/** is not in scope")
				}
				continue
			}
			if !contextaudit.InScope(path) {
				t.Fatalf("%s is not in scope", path)
			}
		}
	})
	for _, tc := range vector.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			snapshot := t.TempDir()
			for path, content := range tc.Files {
				full := filepath.Join(snapshot, filepath.FromSlash(path))
				if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			var waivers []contextaudit.Waiver
			for _, waiver := range tc.Waivers {
				waivers = append(waivers, contextaudit.Waiver{Pin: waiver.Pin, File: waiver.File, Span: waiver.Span, Reason: waiver.Reason})
			}
			report, err := contextaudit.Detect(snapshot, tc.Pin, waivers)
			if err != nil {
				t.Fatal(err)
			}
			if got := !report.Blocking(); got != tc.Expected.Installs {
				t.Fatalf("installs %v, want %v (findings %+v)", got, tc.Expected.Installs, report.Findings)
			}
			if len(report.Findings) != len(tc.Expected.Findings) {
				t.Fatalf("findings %+v, want %+v", report.Findings, tc.Expected.Findings)
			}
			for index, want := range tc.Expected.Findings {
				got := report.Findings[index]
				if got.Class != want.Class || got.File != want.File || got.Pattern != want.Pattern || got.Severity != want.Severity || got.Span != want.Span || got.Waived != want.Waived || got.WaiverReason != want.WaiverReason {
					t.Fatalf("finding %d: %+v, want %+v", index, got, want)
				}
			}
			var gotWarnings []map[string]any
			for _, warning := range report.Waivers {
				entry := map[string]any{"diagnostic": warning.Diagnostic, "file": warning.File, "span": []any{float64(warning.Span[0]), float64(warning.Span[1])}}
				if warning.Reason != "" {
					entry["reason"] = warning.Reason
				}
				if warning.Pin != "" {
					entry["pin"] = warning.Pin
				}
				gotWarnings = append(gotWarnings, entry)
			}
			if tc.PackageKind == "context" {
				manifest, err := contextpkg.LoadManifest(snapshot)
				if err != nil {
					t.Fatal(err)
				}
				for _, module := range contextaudit.SystemModules(manifest.Name, manifest) {
					entry := map[string]any{"class": contextaudit.ClassSystemModulePresent, "package": module.Package, "path": module.Path}
					if module.Selector == nil {
						entry["selector"] = nil
					} else {
						var selector []any
						for _, id := range module.Selector {
							selector = append(selector, id)
						}
						entry["selector"] = selector
					}
					gotWarnings = append(gotWarnings, entry)
				}
			}
			if len(gotWarnings) != len(tc.Expected.Warnings) || (len(gotWarnings) > 0 && !reflect.DeepEqual(gotWarnings, tc.Expected.Warnings)) {
				t.Fatalf("warnings %v, want %v", gotWarnings, tc.Expected.Warnings)
			}
		})
	}
}
