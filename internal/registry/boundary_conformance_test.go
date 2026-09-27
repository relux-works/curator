package registry

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type pageBoundaryVector struct {
	Name               string  `json:"name"`
	Accepted           bool    `json:"accepted"`
	BoundaryPresent    bool    `json:"boundary_present"`
	BoundaryVersion    int     `json:"boundary_version"`
	ChainBoundaryEqual bool    `json:"chain_boundary_equal"`
	Diagnostic         *string `json:"diagnostic"`
	HighWaterAdvanced  bool    `json:"high_water_advanced"`
	RegistryExcluded   bool    `json:"registry_excluded"`
	SameBody           bool    `json:"same_body"`
	SignatureValid     bool    `json:"signature_valid"`
	StoredVersion      int     `json:"stored_version"`
}

// TestPageBoundaryConformanceVectors drives every pinned registry-client
// page_boundary_cases vector through the production HTTP fetch entry point.
// It skips only when CURATOR_CONFORMANCE_ROOT is unset; a root without the
// family or with a different case set is a failure.
func TestPageBoundaryConformanceVectors(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	payload, err := os.ReadFile(filepath.Join(root, "vectors", "registry-client.json")) // #nosec G304 -- pinned conformance root
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Cases []pageBoundaryVector `json:"page_boundary_cases"`
	}
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"fresh-boundary-advances-high-water",
		"equal-version-same-body-accepted",
		"equal-version-different-body-rejected",
		"below-high-water-rejected",
		"chain-boundary-mismatch-rejected",
		"missing-boundary-excluded",
		"bad-signature-rejected",
		"stale-and-mismatch-reports-mismatch",
		"higher-and-mismatch-never-advances",
	}
	got := make([]string, 0, len(document.Cases))
	for _, testCase := range document.Cases {
		got = append(got, testCase.Name)
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("page_boundary_cases names = %v, want %v", got, want)
	}
	for _, testCase := range document.Cases {
		t.Run(testCase.Name, func(t *testing.T) { runPageBoundaryVector(t, testCase) })
	}
}

func runPageBoundaryVector(t *testing.T, testCase pageBoundaryVector) {
	t.Helper()
	w := newBoundaryWorld(t)
	cacheDir, stateDir := t.TempDir(), t.TempDir()
	stored := w.storedBody(testCase.StoredVersion)
	subject := w.otherBody(testCase.BoundaryVersion)
	if testCase.SameBody {
		subject = w.storedBody(testCase.BoundaryVersion)
	}
	signedBy := w.signer
	if !testCase.SignatureValid {
		signedBy = w.rogue
	}
	var pages []string
	if testCase.ChainBoundaryEqual {
		var boundary map[string]any
		if testCase.BoundaryPresent {
			boundary = signedBy.sign(subject)
		}
		pages = []string{boundaryPage(t, boundary, []any{signedBoundaryRecord(t, w, StatusAudited)}, nil)}
	} else {
		next := "page-2"
		pages = []string{
			boundaryPage(t, w.signer.sign(stored), []any{signedBoundaryRecord(t, w, StatusAudited)}, next),
			boundaryPage(t, signedBy.sign(subject), []any{signedBoundaryRecord(t, w, StatusDeprecated)}, nil),
		}
	}
	server, _ := boundaryServer(t, pages...)
	regs := w.registries(server.URL)
	w.seed(t, stateDir, regs[0], stored)
	statePath := boundaryStateFile(stateDir, server.URL)
	before := readBoundaryState(t, statePath)
	fetch := fetchBoundary(t, cacheDir, stateDir, regs, FetchPolicy{PersistCache: true, PersistState: true})
	records, fetchErr := fetch(server.URL, "vector", testCommit, testContentSHA256)
	if testCase.Accepted {
		if fetchErr != nil || len(records) == 0 {
			t.Fatalf("vector accepted=%v but fetch returned %d records and %v", testCase.Accepted, len(records), fetchErr)
		}
	} else {
		if fetchErr == nil || len(records) != 0 {
			t.Fatalf("rejected vector returned %d records and error %v", len(records), fetchErr)
		}
		if testCase.Diagnostic != nil {
			var boundaryErr *PageBoundaryError
			if !errors.As(fetchErr, &boundaryErr) || boundaryErr.Diagnostic != *testCase.Diagnostic {
				t.Fatalf("error = %v, want diagnostic %s", fetchErr, *testCase.Diagnostic)
			}
			if !strings.Contains(fetchErr.Error(), server.URL) {
				t.Fatalf("error does not name registry URL %s: %v", server.URL, fetchErr)
			}
		}
	}
	if testCase.RegistryExcluded != (fetchErr != nil) {
		t.Fatalf("registry_excluded=%v, fetch error=%v", testCase.RegistryExcluded, fetchErr)
	}
	after := readBoundaryState(t, statePath)
	if testCase.HighWaterAdvanced {
		var state snapshotState
		if err := json.Unmarshal(after, &state); err != nil {
			t.Fatal(err)
		}
		if state.HighestVersion != testCase.BoundaryVersion || !state.BoundaryVerified {
			t.Fatalf("vector high-water = %+v", state)
		}
	} else if string(before) != string(after) {
		t.Fatalf("vector changed high-water unexpectedly:\nbefore %s\nafter %s", before, after)
	}
	cache := boundaryCacheEntries(t, cacheDir)
	if testCase.Accepted && len(cache) != 1 {
		t.Fatalf("accepted vector cache entries = %v, want one", cache)
	}
	if !testCase.Accepted && len(cache) != 0 {
		t.Fatalf("rejected vector wrote cache entries: %v", cache)
	}
}
