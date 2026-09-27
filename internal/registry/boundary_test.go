package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/relux-works/curator/internal/stateread"
)

type boundaryWorld struct {
	signer *signer
	rogue  *signer
	now    time.Time
}

func newBoundaryWorld(t *testing.T) *boundaryWorld {
	t.Helper()
	return &boundaryWorld{signer: newSigner(t), rogue: newSigner(t), now: time.Now().UTC().Truncate(time.Second)}
}

func (w *boundaryWorld) registries(url string) []Registry {
	return []Registry{{Name: "one", URL: url, PublicKeys: []string{w.signer.pinned}}}
}

func (w *boundaryWorld) storedBody(version int) map[string]any {
	body := snapshotBody(version, w.now)
	body["head"] = strings.Repeat("b", 64)
	body["merkle_root"] = strings.Repeat("a", 64)
	body["log_size"] = version
	return body
}

func (w *boundaryWorld) otherBody(version int) map[string]any {
	body := snapshotBody(version, w.now)
	body["head"] = strings.Repeat("c", 64)
	body["merkle_root"] = strings.Repeat("d", 64)
	body["log_size"] = version
	return body
}

func (w *boundaryWorld) seed(t *testing.T, stateDir string, reg Registry, body map[string]any) {
	t.Helper()
	fetch := func(string) (map[string]any, error) { return w.signer.sign(body), nil }
	tampered, warnings := CheckSnapshotsWithPolicy(
		[]Registry{reg}, stateDir, fetch, w.now, DefaultSnapshotMaxAge, DefaultSnapshotClockSkew,
	)
	if len(tampered) != 0 || len(warnings) != 0 {
		t.Fatalf("seed snapshot state: tampered=%v warnings=%v", tampered, warnings)
	}
}

func boundaryStateFile(stateDir, registryURL string) string {
	sum := sha256.Sum256([]byte(registryURL))
	return filepath.Join(stateDir, "snapshot-"+hex.EncodeToString(sum[:])[:16]+".json")
}

func readBoundaryState(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path) // #nosec G304 -- test-owned state path
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func boundaryCacheEntries(t *testing.T, cacheDir string) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(cacheDir, "records-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func boundaryPage(t *testing.T, boundary map[string]any, records []any, next any) string {
	t.Helper()
	page := map[string]any{"records": records, "next_cursor": next}
	if boundary != nil {
		page["boundary"] = boundary
	}
	payload, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func boundaryServer(t *testing.T, pages ...string) (*httptest.Server, *int) {
	t.Helper()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		page := pages[len(pages)-1]
		if calls <= len(pages) {
			page = pages[calls-1]
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(page))
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func signedBoundaryRecord(t *testing.T, w *boundaryWorld, status string) map[string]any {
	t.Helper()
	return w.signer.sign(record(status))
}

func fetchBoundary(t *testing.T, cacheDir, stateDir string, regs []Registry, policy FetchPolicy) FetchFn {
	t.Helper()
	return NewHTTPFetchWithPolicy(cacheDir, stateDir, regs, time.Minute, time.Hour, time.Now, policy)
}

type boundaryFetchBuilder struct {
	name  string
	build func(cacheDir, stateDir string, regs []Registry) FetchFn
}

func boundaryFetchBuilders() []boundaryFetchBuilder {
	return []boundaryFetchBuilder{
		{
			name: "NewHTTPFetch persistent",
			build: func(cacheDir, stateDir string, regs []Registry) FetchFn {
				return NewHTTPFetch(cacheDir, stateDir, regs, time.Minute, time.Hour, time.Now)
			},
		},
		{
			name: "NewHTTPFetchWithPolicy persistent",
			build: func(cacheDir, stateDir string, regs []Registry) FetchFn {
				return NewHTTPFetchWithPolicy(cacheDir, stateDir, regs, time.Minute, time.Hour, time.Now,
					FetchPolicy{PersistCache: true, PersistState: true})
			},
		},
		{
			name: "NewHTTPFetchWithPolicy read-only",
			build: func(cacheDir, stateDir string, regs []Registry) FetchFn {
				return NewHTTPFetchWithPolicy(cacheDir, stateDir, regs, time.Minute, time.Hour, time.Now, FetchPolicy{})
			},
		},
		{
			name: "NewHTTPFetchWithPolicyReadOnly",
			build: func(cacheDir, stateDir string, regs []Registry) FetchFn {
				return NewHTTPFetchWithPolicyReadOnly(cacheDir, stateDir, regs, time.Minute, time.Hour, time.Now)
			},
		},
	}
}

func requirePageStateError(t *testing.T, records []map[string]any, err error) {
	t.Helper()
	var stateErr *pageStateError
	if !errors.As(err, &stateErr) {
		t.Fatalf("error = %v, want fail-closed pageStateError", err)
	}
	if len(records) != 0 {
		t.Fatalf("unavailable high-water state served %d records", len(records))
	}
}

func TestFetchFnRejectsUnreadableOrCorruptHighWaterBeforeNetwork(t *testing.T) {
	for _, builder := range boundaryFetchBuilders() {
		t.Run(builder.name, func(t *testing.T) {
			for _, stateShape := range []string{"directory", "corrupt-json"} {
				t.Run(stateShape, func(t *testing.T) {
					w := newBoundaryWorld(t)
					staleBoundary := w.signer.sign(w.otherBody(7))
					server, calls := boundaryServer(t, boundaryPage(t, staleBoundary,
						[]any{signedBoundaryRecord(t, w, StatusAudited)}, nil))
					regs := w.registries(server.URL)
					cacheDir, stateDir := filepath.Join(t.TempDir(), "cache"), t.TempDir()
					w.seed(t, stateDir, regs[0], w.storedBody(8))

					statePath := boundaryStateFile(stateDir, server.URL)
					catalogPath := filepath.Join(stateDir, snapshotStateCatalogName)
					// Model interruption after the state-file write but before its
					// catalog update, so the unreadable-state check itself must fail
					// closed rather than relying on the catalog's missing-file guard.
					catalogBefore := []byte(`{"schema_version":1,"states":[]}`)
					if err := os.WriteFile(catalogPath, catalogBefore, 0o600); err != nil {
						t.Fatal(err)
					}
					var stateBefore []byte
					var directoryMarker string
					switch stateShape {
					case "directory":
						if err := os.Remove(statePath); err != nil {
							t.Fatal(err)
						}
						if err := os.Mkdir(statePath, 0o700); err != nil {
							t.Fatal(err)
						}
						directoryMarker = filepath.Join(statePath, "keep-me")
						stateBefore = []byte("still here")
						if err := os.WriteFile(directoryMarker, stateBefore, 0o600); err != nil {
							t.Fatal(err)
						}
					case "corrupt-json":
						stateBefore = []byte("{not a snapshot state")
						if err := os.WriteFile(statePath, stateBefore, 0o600); err != nil {
							t.Fatal(err)
						}
					}

					fetch := builder.build(cacheDir, stateDir, regs)
					records, err := fetch(server.URL, "id", testCommit, testContentSHA256)
					requirePageStateError(t, records, err)
					if *calls != 0 {
						t.Fatalf("fetch contacted registry %d times after the high-water read failed", *calls)
					}
					if entries := boundaryCacheEntries(t, cacheDir); len(entries) != 0 {
						t.Fatalf("unavailable high-water state wrote record cache: %v", entries)
					}
					if after := readBoundaryState(t, catalogPath); string(after) != string(catalogBefore) {
						t.Fatalf("unavailable high-water state changed catalog:\nbefore %s\nafter %s", catalogBefore, after)
					}
					switch stateShape {
					case "directory":
						info, statErr := os.Stat(statePath)
						if statErr != nil || !info.IsDir() {
							t.Fatalf("state path after fetch = (%v, %v), want unchanged directory", info, statErr)
						}
						if after := readBoundaryState(t, directoryMarker); string(after) != string(stateBefore) {
							t.Fatalf("unavailable state directory contents changed: before %q, after %q", stateBefore, after)
						}
					case "corrupt-json":
						if after := readBoundaryState(t, statePath); string(after) != string(stateBefore) {
							t.Fatalf("corrupt rollback state changed:\nbefore %s\nafter %s", stateBefore, after)
						}
					}
				})
			}
		})
	}
}

func TestFetchFnRejectsCatalogListedMissingHighWater(t *testing.T) {
	for _, builder := range boundaryFetchBuilders() {
		t.Run(builder.name, func(t *testing.T) {
			w := newBoundaryWorld(t)
			staleBoundary := w.signer.sign(w.otherBody(7))
			server, calls := boundaryServer(t, boundaryPage(t, staleBoundary,
				[]any{signedBoundaryRecord(t, w, StatusAudited)}, nil))
			regs := w.registries(server.URL)
			cacheDir, stateDir := filepath.Join(t.TempDir(), "cache"), t.TempDir()
			w.seed(t, stateDir, regs[0], w.storedBody(8))

			statePath := boundaryStateFile(stateDir, server.URL)
			catalogPath := filepath.Join(stateDir, snapshotStateCatalogName)
			catalogBefore := readBoundaryState(t, catalogPath)
			if err := os.Remove(statePath); err != nil {
				t.Fatal(err)
			}

			fetch := builder.build(cacheDir, stateDir, regs)
			records, err := fetch(server.URL, "id", testCommit, testContentSHA256)
			requirePageStateError(t, records, err)
			if *calls != 0 {
				t.Fatalf("fetch contacted registry %d times after catalog-listed state disappeared", *calls)
			}
			if _, err := os.Lstat(statePath); !os.IsNotExist(err) {
				t.Fatalf("catalog-listed high-water path was recreated: %v", err)
			}
			if after := readBoundaryState(t, catalogPath); string(after) != string(catalogBefore) {
				t.Fatalf("missing high-water state changed catalog:\nbefore %s\nafter %s", catalogBefore, after)
			}
			if entries := boundaryCacheEntries(t, cacheDir); len(entries) != 0 {
				t.Fatalf("missing catalog-listed state wrote record cache: %v", entries)
			}
		})
	}
}

func TestReadOnlyStateDirFallsBackOnlyForProvenAbsence(t *testing.T) {
	legacyDir := t.TempDir()
	presentDir := t.TempDir()
	blockedParent := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blockedParent, []byte("file blocks child inspection"), 0o600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		stateDir   string
		wantKind   stateread.Kind
		wantDir    string
		wantReport bool
	}{
		{
			name:     "absent state uses legacy directory",
			stateDir: filepath.Join(t.TempDir(), "state"),
			wantKind: stateread.KindAbsent,
			wantDir:  legacyDir,
		},
		{
			name:     "present state uses protected directory",
			stateDir: presentDir,
			wantKind: stateread.KindPresent,
			wantDir:  presentDir,
		},
		{
			name:       "unreadable state does not fall back",
			stateDir:   filepath.Join(blockedParent, "state"),
			wantKind:   stateread.KindUnreadable,
			wantDir:    filepath.Join(blockedParent, "state"),
			wantReport: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			metadata, err := stateread.Stat(test.stateDir)
			if metadata.Kind != test.wantKind || (test.wantKind == stateread.KindUnreadable) != (err != nil) {
				t.Fatalf("stateread.Stat(%q) = (%+v, %v), want kind %q", test.stateDir, metadata, err, test.wantKind)
			}
			selected := ReadOnlyStateDir(test.stateDir, legacyDir)
			if selected != test.wantDir {
				t.Fatalf("ReadOnlyStateDir(%q) = %q, want %q", test.stateDir, selected, test.wantDir)
			}
			if test.wantReport {
				rows := ReadBoundaryPosture(selected, []Registry{{Name: "one", URL: "https://one.example.test", PublicKeys: []string{"pinned"}}})
				if len(rows) != 1 || !strings.Contains(rows[0].Diagnostic, stateread.DiagUnreadable) {
					t.Fatalf("unreadable selected state posture = %+v, want a fail-closed %s diagnostic", rows, stateread.DiagUnreadable)
				}
			}
		})
	}
}

func requireBoundaryError(t *testing.T, records []map[string]any, err error, code, registryURL string) {
	t.Helper()
	if err == nil {
		t.Fatalf("wanted %s, received %d records", code, len(records))
	}
	var boundaryErr *PageBoundaryError
	if !errors.As(err, &boundaryErr) || boundaryErr.Diagnostic != code {
		t.Fatalf("error = %v, want page diagnostic %s", err, code)
	}
	if !strings.HasPrefix(err.Error(), code) || !strings.Contains(err.Error(), registryURL) {
		t.Fatalf("error %q must start with %s and name %s", err, code, registryURL)
	}
	if len(records) != 0 {
		t.Fatalf("rejected page contributed %d records", len(records))
	}
}

func TestRecordsPageBoundaryCases(t *testing.T) {
	tests := []struct {
		name          string
		storedVersion int
		first         func(*boundaryWorld) map[string]any
		second        func(*boundaryWorld) map[string]any
		missingFirst  bool
		accepted      bool
		wantError     string
		wantAdvanced  int
	}{
		{name: "fresh boundary advances", storedVersion: 7, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.otherBody(8)) }, accepted: true, wantAdvanced: 8},
		{name: "equal version same body", storedVersion: 8, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.storedBody(8)) }, accepted: true, wantAdvanced: 8},
		{name: "equal version different body", storedVersion: 8, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.otherBody(8)) }, wantError: PageBoundaryStale, wantAdvanced: 8},
		{name: "below high-water", storedVersion: 8, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.otherBody(7)) }, wantError: PageBoundaryStale, wantAdvanced: 8},
		{name: "later mismatch", storedVersion: 7, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.storedBody(7)) }, second: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.otherBody(8)) }, wantError: PageBoundaryMismatch, wantAdvanced: 7},
		{name: "missing boundary", storedVersion: 7, missingFirst: true, wantError: PageBoundaryMissing, wantAdvanced: 7},
		{name: "bad signature", storedVersion: 7, first: func(w *boundaryWorld) map[string]any { return w.rogue.sign(w.otherBody(8)) }, wantError: PageBoundaryMissing, wantAdvanced: 7},
		{name: "stale and mismatch reports mismatch", storedVersion: 8, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.storedBody(8)) }, second: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.otherBody(7)) }, wantError: PageBoundaryMismatch, wantAdvanced: 8},
		{name: "higher and mismatch never advances", storedVersion: 7, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.storedBody(7)) }, second: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.otherBody(9)) }, wantError: PageBoundaryMismatch, wantAdvanced: 7},
		{name: "missing wins over later mismatch", storedVersion: 7, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.storedBody(7)) }, second: func(*boundaryWorld) map[string]any { return nil }, wantError: PageBoundaryMissing, wantAdvanced: 7},
		{name: "bad signature wins over later mismatch", storedVersion: 7, first: func(w *boundaryWorld) map[string]any { return w.signer.sign(w.storedBody(7)) }, second: func(w *boundaryWorld) map[string]any { return w.rogue.sign(w.otherBody(8)) }, wantError: PageBoundaryMissing, wantAdvanced: 7},
		{name: "bad signature wins over stale", storedVersion: 8, first: func(w *boundaryWorld) map[string]any { return w.rogue.sign(w.otherBody(7)) }, wantError: PageBoundaryMissing, wantAdvanced: 8},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newBoundaryWorld(t)
			cacheDir, stateDir := t.TempDir(), t.TempDir()
			var pages []string
			if test.second != nil {
				next := "page-2"
				pages = []string{
					boundaryPage(t, test.first(w), []any{signedBoundaryRecord(t, w, StatusAudited)}, next),
					boundaryPage(t, test.second(w), []any{signedBoundaryRecord(t, w, StatusDeprecated)}, nil),
				}
			} else {
				var first map[string]any
				if !test.missingFirst {
					first = test.first(w)
				}
				pages = []string{boundaryPage(t, first, []any{signedBoundaryRecord(t, w, StatusAudited)}, nil)}
			}
			server, calls := boundaryServer(t, pages...)
			regs := w.registries(server.URL)
			w.seed(t, stateDir, regs[0], w.storedBody(test.storedVersion))
			before := readBoundaryState(t, boundaryStateFile(stateDir, server.URL))

			fetch := fetchBoundary(t, cacheDir, stateDir, regs, FetchPolicy{PersistCache: true, PersistState: true})
			records, err := fetch(server.URL, "id", testCommit, testContentSHA256)
			if test.accepted {
				if err != nil || len(records) == 0 {
					t.Fatalf("accepted case: records=%v error=%v", records, err)
				}
			} else {
				requireBoundaryError(t, records, err, test.wantError, server.URL)
			}
			if test.second != nil && *calls != 2 {
				t.Fatalf("later-page case made %d requests, want 2", *calls)
			}
			stateBytes := readBoundaryState(t, boundaryStateFile(stateDir, server.URL))
			if test.accepted && test.wantAdvanced > test.storedVersion {
				var state snapshotState
				if err := json.Unmarshal(stateBytes, &state); err != nil {
					t.Fatal(err)
				}
				if state.HighestVersion != test.wantAdvanced || !state.BoundaryVerified {
					t.Fatalf("higher boundary did not persist with posture: %+v", state)
				}
			} else if string(stateBytes) != string(before) {
				t.Fatalf("accepted equal or rejected chain changed rollback state:\nbefore %s\nafter %s", before, stateBytes)
			}
			entries := boundaryCacheEntries(t, cacheDir)
			if test.accepted && len(entries) != 1 {
				t.Fatalf("accepted chain cache entries = %v, want one", entries)
			}
			if !test.accepted && len(entries) != 0 {
				t.Fatalf("rejected chain wrote record cache: %v", entries)
			}
		})
	}
}

func TestRecordsPageBoundaryEqualVersionRejectsEachChangedBodyField(t *testing.T) {
	changes := []struct {
		name   string
		change func(map[string]any)
	}{
		{name: "head", change: func(body map[string]any) { body["head"] = strings.Repeat("c", 64) }},
		{name: "merkle_root", change: func(body map[string]any) { body["merkle_root"] = strings.Repeat("d", 64) }},
		{name: "log_size", change: func(body map[string]any) { body["log_size"] = 7 }},
	}
	for _, test := range changes {
		t.Run(test.name, func(t *testing.T) {
			w := newBoundaryWorld(t)
			cacheDir, stateDir := t.TempDir(), t.TempDir()
			body := w.storedBody(8)
			test.change(body)
			boundary := w.signer.sign(body)
			server, _ := boundaryServer(t, boundaryPage(t, boundary, []any{signedBoundaryRecord(t, w, StatusAudited)}, nil))
			reg := w.registries(server.URL)[0]
			w.seed(t, stateDir, reg, w.storedBody(8))
			statePath := boundaryStateFile(stateDir, server.URL)
			before := readBoundaryState(t, statePath)
			fetch := fetchBoundary(t, cacheDir, stateDir, []Registry{reg}, FetchPolicy{PersistCache: true, PersistState: true})
			records, err := fetch(server.URL, "id", testCommit, testContentSHA256)
			requireBoundaryError(t, records, err, PageBoundaryStale, server.URL)
			if after := readBoundaryState(t, statePath); string(after) != string(before) {
				t.Fatalf("equal-version %s conflict changed rollback state:\nbefore %s\nafter %s", test.name, before, after)
			}
			if entries := boundaryCacheEntries(t, cacheDir); len(entries) != 0 {
				t.Fatalf("equal-version %s conflict wrote cache: %v", test.name, entries)
			}
		})
	}
}

func TestRecordsPageBoundaryRejectsNonV2EnvelopeBeforeStateAdvance(t *testing.T) {
	tests := []struct {
		name string
		page func(*testing.T, *boundaryWorld, map[string]any) map[string]any
	}{
		{
			name: "unknown top-level member",
			page: func(t *testing.T, w *boundaryWorld, boundary map[string]any) map[string]any {
				return map[string]any{
					"records": []any{signedBoundaryRecord(t, w, StatusAudited)}, "next_cursor": nil,
					"boundary": boundary, "extension": true,
				}
			},
		},
		{
			name: "missing records member",
			page: func(_ *testing.T, _ *boundaryWorld, boundary map[string]any) map[string]any {
				return map[string]any{"next_cursor": nil, "boundary": boundary}
			},
		},
		{
			name: "missing next_cursor member",
			page: func(t *testing.T, w *boundaryWorld, boundary map[string]any) map[string]any {
				return map[string]any{"records": []any{signedBoundaryRecord(t, w, StatusAudited)}, "boundary": boundary}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newBoundaryWorld(t)
			cacheDir, stateDir := t.TempDir(), t.TempDir()
			body := w.otherBody(8)
			boundary := w.signer.sign(body)
			page, err := json.Marshal(test.page(t, w, boundary))
			if err != nil {
				t.Fatal(err)
			}
			server, _ := boundaryServer(t, string(page))
			regs := w.registries(server.URL)
			w.seed(t, stateDir, regs[0], w.storedBody(7))
			statePath := boundaryStateFile(stateDir, server.URL)
			before := readBoundaryState(t, statePath)
			fetch := fetchBoundary(t, cacheDir, stateDir, regs, FetchPolicy{PersistCache: true, PersistState: true})
			records, fetchErr := fetch(server.URL, "id", testCommit, testContentSHA256)
			if fetchErr == nil || !strings.Contains(fetchErr.Error(), "unknown or missing fields") || len(records) != 0 {
				t.Fatalf("non-v2 page returned records=%v error=%v", records, fetchErr)
			}
			if after := readBoundaryState(t, statePath); string(after) != string(before) {
				t.Fatalf("non-v2 page advanced rollback state:\nbefore %s\nafter %s", before, after)
			}
			if entries := boundaryCacheEntries(t, cacheDir); len(entries) != 0 {
				t.Fatalf("non-v2 page wrote cache: %v", entries)
			}
		})
	}
}

func TestRecordsPageBoundaryRequiresByteIdenticalChain(t *testing.T) {
	w := newBoundaryWorld(t)
	cacheDir, stateDir := t.TempDir(), t.TempDir()
	boundary := w.signer.sign(w.storedBody(7))
	first, err := json.Marshal(map[string]any{
		"records": []any{signedBoundaryRecord(t, w, StatusAudited)}, "next_cursor": "page-2", "boundary": boundary,
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.MarshalIndent(map[string]any{
		"records": []any{signedBoundaryRecord(t, w, StatusDeprecated)}, "next_cursor": nil, "boundary": boundary,
	}, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	server, _ := boundaryServer(t, string(first), string(second))
	regs := w.registries(server.URL)
	w.seed(t, stateDir, regs[0], w.storedBody(7))
	fetch := fetchBoundary(t, cacheDir, stateDir, regs, FetchPolicy{PersistCache: true, PersistState: true})
	records, fetchErr := fetch(server.URL, "id", testCommit, testContentSHA256)
	requireBoundaryError(t, records, fetchErr, PageBoundaryMismatch, server.URL)
}

func TestRecordsPageBoundaryReadOnlyNeverPersists(t *testing.T) {
	w := newBoundaryWorld(t)
	t.Run("first use creates no state or cache", func(t *testing.T) {
		root := t.TempDir()
		stateDir := filepath.Join(root, "missing-state")
		cacheDir := filepath.Join(root, "missing-cache")
		boundary := w.signer.sign(w.otherBody(8))
		server, _ := boundaryServer(t, boundaryPage(t, boundary, []any{signedBoundaryRecord(t, w, StatusAudited)}, nil))
		regs := w.registries(server.URL)
		fetch := fetchBoundary(t, cacheDir, stateDir, regs, FetchPolicy{})
		if records, err := fetch(server.URL, "id", testCommit, testContentSHA256); err != nil || len(records) != 1 {
			t.Fatalf("read-only page failed: records=%v error=%v", records, err)
		}
		if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
			t.Fatalf("read-only fetch created high-water state: %v", err)
		}
		if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
			t.Fatalf("read-only fetch created a record cache: %v", err)
		}
	})

	t.Run("higher boundary is verified without advancing", func(t *testing.T) {
		cacheDir, stateDir := t.TempDir(), t.TempDir()
		boundary := w.signer.sign(w.otherBody(8))
		server, _ := boundaryServer(t, boundaryPage(t, boundary, []any{signedBoundaryRecord(t, w, StatusAudited)}, nil))
		regs := w.registries(server.URL)
		w.seed(t, stateDir, regs[0], w.storedBody(7))
		before := readBoundaryState(t, boundaryStateFile(stateDir, server.URL))
		fetch := fetchBoundary(t, cacheDir, stateDir, regs, FetchPolicy{})
		if records, err := fetch(server.URL, "id", testCommit, testContentSHA256); err != nil || len(records) != 1 {
			t.Fatalf("read-only higher boundary failed: records=%v error=%v", records, err)
		}
		if after := readBoundaryState(t, boundaryStateFile(stateDir, server.URL)); string(after) != string(before) {
			t.Fatalf("read-only fetch changed rollback state:\nbefore %s\nafter %s", before, after)
		}
	})
}

func TestReadBoundaryPostureReportsPersistedHighWaterWithoutWrites(t *testing.T) {
	w := newBoundaryWorld(t)
	stateDir := t.TempDir()
	reg := w.registries("https://registry.example.test")[0]
	w.seed(t, stateDir, reg, w.storedBody(7))
	statePath := boundaryStateFile(stateDir, reg.URL)
	state, exists, err := readSnapshotState(statePath)
	if err != nil || !exists {
		t.Fatalf("seeded snapshot state exists=%v error=%v", exists, err)
	}
	state.BoundaryVerified = true
	if err := writeSnapshotState(statePath, state); err != nil {
		t.Fatal(err)
	}
	before := readBoundaryState(t, statePath)
	entriesBefore, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatal(err)
	}

	rows := ReadBoundaryPosture(stateDir, []Registry{reg})
	if len(rows) != 1 || rows[0].Diagnostic != "" || rows[0].Name != reg.Name || rows[0].URL != reg.URL ||
		rows[0].HighWaterVersion == nil || *rows[0].HighWaterVersion != 7 ||
		rows[0].HighWaterLogSize == nil || *rows[0].HighWaterLogSize != 7 || !rows[0].LastBoundaryVerified {
		t.Fatalf("persisted boundary posture = %+v", rows)
	}
	if after := readBoundaryState(t, statePath); string(after) != string(before) {
		t.Fatalf("read-only posture changed state:\nbefore %s\nafter %s", before, after)
	}
	entriesAfter, err := os.ReadDir(stateDir)
	if err != nil || len(entriesAfter) != len(entriesBefore) {
		t.Fatalf("read-only posture changed state files: before=%v after=%v error=%v", entriesBefore, entriesAfter, err)
	}
}

func TestReadBoundaryPostureOnlyListsRegistriesWithPinnedKeys(t *testing.T) {
	w := newBoundaryWorld(t)
	registries := []Registry{
		{Name: "untrusted", URL: "https://untrusted.example.test"},
		w.registries("https://trusted.example.test")[0],
	}
	rows := ReadBoundaryPosture(t.TempDir(), registries)
	if len(rows) != 1 || rows[0].Name != "one" || rows[0].URL != "https://trusted.example.test" {
		t.Fatalf("registry posture rows = %+v, want only the pinned trusted registry", rows)
	}
}

func TestRecordsPageBoundaryPersistsBeforeContribution(t *testing.T) {
	w := newBoundaryWorld(t)
	cacheDir, stateDir := t.TempDir(), t.TempDir()
	firstBoundary := w.signer.sign(w.otherBody(8))
	var calls int
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		calls++
		resp.Header().Set("Content-Type", "application/json")
		if req.URL.Query().Get("cursor") == "" {
			next := "page-2"
			_, _ = resp.Write([]byte(boundaryPage(t, firstBoundary, []any{signedBoundaryRecord(t, w, StatusAudited)}, next)))
			return
		}
		var persisted snapshotState
		stateBytes, err := os.ReadFile(boundaryStateFile(stateDir, serverURL))
		if err != nil || json.Unmarshal(stateBytes, &persisted) != nil || persisted.HighestVersion != 8 || !persisted.BoundaryVerified {
			resp.WriteHeader(http.StatusInternalServerError)
			_, _ = resp.Write([]byte(`{"error":"boundary was not persisted before page contribution"}`))
			return
		}
		_, _ = resp.Write([]byte(boundaryPage(t, firstBoundary, []any{signedBoundaryRecord(t, w, StatusDeprecated)}, nil)))
	}))
	serverURL = server.URL
	t.Cleanup(server.Close)
	regs := w.registries(serverURL)
	w.seed(t, stateDir, regs[0], w.storedBody(7))
	fetch := fetchBoundary(t, cacheDir, stateDir, regs, FetchPolicy{PersistCache: true, PersistState: true})
	records, err := fetch(serverURL, "id", testCommit, testContentSHA256)
	if err != nil || len(records) != 2 || calls != 2 {
		t.Fatalf("page chain = %d records, %d calls, error %v; want persisted page-1 state before page 2", len(records), calls, err)
	}
}

func TestRecordsPageBoundaryRejectsRefreshWithoutServingOrReplacingCache(t *testing.T) {
	w := newBoundaryWorld(t)
	cacheDir, stateDir := t.TempDir(), t.TempDir()
	currentBoundary := w.signer.sign(w.otherBody(8))
	var stale atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
		resp.Header().Set("Content-Type", "application/json")
		boundary := currentBoundary
		if stale.Load() {
			boundary = w.signer.sign(w.storedBody(7))
		}
		_, _ = resp.Write([]byte(boundaryPage(t, boundary, []any{signedBoundaryRecord(t, w, StatusAudited)}, nil)))
	}))
	t.Cleanup(server.Close)
	regs := w.registries(server.URL)
	w.seed(t, stateDir, regs[0], w.storedBody(7))
	fetch := NewHTTPFetchWithPolicy(cacheDir, stateDir, regs, 0, time.Hour, time.Now,
		FetchPolicy{PersistCache: true, PersistState: true})
	if records, err := fetch(server.URL, "id", testCommit, testContentSHA256); err != nil || len(records) != 1 {
		t.Fatalf("initial current page failed: records=%v error=%v", records, err)
	}
	entries := boundaryCacheEntries(t, cacheDir)
	if len(entries) != 1 {
		t.Fatalf("initial accepted page cache entries = %v, want one", entries)
	}
	cacheBefore := readBoundaryState(t, entries[0])
	stale.Store(true)
	if records, err := fetch(server.URL, "id", testCommit, testContentSHA256); err == nil || len(records) != 0 {
		t.Fatalf("stale refresh returned %d cached/current records with error %v", len(records), err)
	} else {
		var boundaryErr *PageBoundaryError
		if !errors.As(err, &boundaryErr) || boundaryErr.Diagnostic != PageBoundaryStale || !strings.Contains(err.Error(), server.URL) {
			t.Fatalf("stale refresh error = %v, want URL-scoped %s", err, PageBoundaryStale)
		}
	}
	if after := readBoundaryState(t, entries[0]); string(after) != string(cacheBefore) {
		t.Fatalf("rejected refresh replaced the accepted cache:\nbefore %s\nafter %s", cacheBefore, after)
	}
}

func TestAttestRootKeepsRejectedBoundaryUnknownAndExplainsWhy(t *testing.T) {
	w := newBoundaryWorld(t)
	cacheDir, stateDir := t.TempDir(), t.TempDir()
	server, _ := boundaryServer(t, boundaryPage(t, w.signer.sign(w.otherBody(7)), []any{signedBoundaryRecord(t, w, StatusAudited)}, nil))
	regs := w.registries(server.URL)
	w.seed(t, stateDir, regs[0], w.storedBody(8))
	fetch := fetchBoundary(t, cacheDir, stateDir, regs, FetchPolicy{})
	root := t.TempDir()
	writeAttestMarker(t, root, "net", v5AttestNetworkMarker())
	results := attestResultsBySkill(AttestRoot("test", root, regs, fetch))
	got := results["net"]
	if got.Result != ResultUnknown || got.Registry != "" {
		t.Fatalf("stale boundary attestation = %+v, want unknown without an authorizing registry", got)
	}
	if !strings.Contains(got.Detail, PageBoundaryStale) || !strings.Contains(got.Detail, server.URL) {
		t.Fatalf("attestation detail %q must name diagnostic and registry URL", got.Detail)
	}
}
