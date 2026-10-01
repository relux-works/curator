package conformancecoverage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/hashing"
	"github.com/relux-works/curator/internal/registry"
)

func isContentHashV2Gap(gap Gap) bool {
	if gap.Family == "skillfile-sources-v1/schema-cases" {
		return strings.HasPrefix(gap.CaseID, "schema-cases/install-marker-v6/") ||
			strings.HasPrefix(gap.CaseID, "schema-cases/skillfile-lock-v2/") ||
			strings.HasPrefix(gap.CaseID, "schema-cases/source-audit-v2/")
	}
	return false
}

func TestDeferredContentHashV2CarrierCasesAreAbsentAndUnowned(t *testing.T) {
	suiteID, err := selectedSuiteIdentity()
	if err != nil {
		t.Fatal(err)
	}
	_, gaps, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	deferredRows := 0
	for _, gap := range gaps {
		if !isContentHashV2Gap(gap) {
			continue
		}
		if suiteID != ContentHashV2CandidateManifestSHA256 {
			t.Errorf("content-hash-v2 gap %s/%s leaked into non-candidate suite %s", gap.Family, gap.CaseID, suiteID)
			continue
		}
		deferredRows++
	}
	if deferredRows != 0 {
		t.Errorf("deferred carrier gap rows = %d, want 0 for selected suite %s", deferredRows, suiteID)
	}
	if suiteID != ContentHashV2CandidateManifestSHA256 {
		return
	}
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Fatal("candidate suite identity requires CURATOR_CONFORMANCE_ROOT")
	}
	data, err := os.ReadFile(filepath.Join(root, "schema-cases", "index.json"))
	if err != nil {
		t.Fatalf("read schema-case index: %v", err)
	}
	var index []struct {
		Instance string `json:"instance"`
	}
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatalf("parse schema-case index: %v", err)
	}
	deferredPaths := []string{
		"install-marker-v6/",
		"skillfile-lock-v2/",
		"source-audit-v2/",
	}
	for _, entry := range index {
		for _, path := range deferredPaths {
			if strings.Contains(entry.Instance, path) {
				t.Errorf("deferred schema case %q is unexpectedly published in suite %s", entry.Instance, suiteID)
			}
		}
	}
}

type contentHashVector struct {
	ID string
}

type contentHashVectorFile struct {
	Path      string `json:"path"`
	BytesUTF8 string `json:"bytes_utf8"`
}

type contentHashVectorTree struct {
	Files    []contentHashVectorFile `json:"files"`
	V1SHA256 string                  `json:"v1_sha256"`
	V2SHA256 string                  `json:"v2_sha256"`
}

func driveContentHashVector(t *testing.T, root string, vector contentHashVector) {
	t.Helper()
	readJSON := func(path string, target any) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(root, "vectors", "content-hashes-v2.json"))
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]json.RawMessage
		if err := json.Unmarshal(data, &document); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(document[path], target); err != nil {
			t.Fatalf("decode vector %s: %v", path, err)
		}
	}
	writeTree := func(files []contentHashVectorFile) string {
		t.Helper()
		dir := t.TempDir()
		for _, file := range files {
			path := filepath.Join(dir, filepath.FromSlash(file.Path))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(file.BytesUTF8), 0o600); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	assertHashes := func(dir string, v1, v2 string) {
		t.Helper()
		gotV1, err := hashing.ContentSHA256WithVersion(dir, nil, hashing.VersionV1)
		if err != nil {
			t.Fatal(err)
		}
		gotV2, err := hashing.ContentSHA256WithVersion(dir, nil, hashing.VersionV2)
		if err != nil {
			t.Fatal(err)
		}
		if v1 != "" && gotV1 != v1 {
			t.Fatalf("v1 hash = %s, want %s", gotV1, v1)
		}
		if gotV2 != v2 {
			t.Fatalf("v2 hash = %s, want %s", gotV2, v2)
		}
	}

	switch vector.ID {
	case "colliding_v1_pair":
		var payload struct {
			Trees []contentHashVectorTree `json:"trees"`
		}
		readJSON(vector.ID, &payload)
		if len(payload.Trees) != 2 {
			t.Fatalf("collision vector tree count = %d, want 2", len(payload.Trees))
		}
		for _, tree := range payload.Trees {
			assertHashes(writeTree(tree.Files), tree.V1SHA256, tree.V2SHA256)
		}
		if payload.Trees[0].V1SHA256 != payload.Trees[1].V1SHA256 || payload.Trees[0].V2SHA256 == payload.Trees[1].V2SHA256 {
			t.Fatal("published collision construction does not collide under v1 and separate under v2")
		}
	case "empty_tree":
		var payload struct {
			V2SHA256 string `json:"v2_sha256"`
		}
		readJSON(vector.ID, &payload)
		assertHashes(writeTree(nil), "", payload.V2SHA256)
	case "nested_nul":
		var payload struct {
			File     contentHashVectorFile `json:"file"`
			V2SHA256 string                `json:"v2_sha256"`
		}
		readJSON(vector.ID, &payload)
		assertHashes(writeTree([]contentHashVectorFile{payload.File}), "", payload.V2SHA256)
	case "ordinary_tree":
		var payload contentHashVectorTree
		readJSON(vector.ID, &payload)
		assertHashes(writeTree(payload.Files), payload.V1SHA256, payload.V2SHA256)
	case "registry_version_mismatch":
		var payload struct {
			Computed struct {
				ContentSHA256 string `json:"content_sha256"`
				HashVersion   int    `json:"hash_version"`
			} `json:"computed"`
			Record struct {
				ContentSHA256 string `json:"content_sha256"`
				HashVersion   int    `json:"hash_version"`
			} `json:"record"`
		}
		readJSON(vector.ID, &payload)
		record := registry.Record{
			ContentSHA256:  payload.Record.ContentSHA256,
			HashVersion:    hashing.Version(payload.Record.HashVersion),
			SourceIdentity: "same-source",
			Commit:         "same-commit",
		}
		if registry.MatchesVersioned(record, record.SourceIdentity, record.Commit, payload.Computed.ContentSHA256, hashing.Version(payload.Computed.HashVersion)) {
			t.Fatal("registry matching admitted equal digest/source across unequal hash versions")
		}
	default:
		t.Fatalf("unhandled content hash vector %q", vector.ID)
	}
}

// TestContentHashV2VectorsWhenPublished counts the five executable vectors in
// content-hashes-v2.json and drives each through internal/hashing and
// registry matching. The schema_version and framing members define the vector
// document; they are metadata, not cases.
func TestContentHashV2VectorsWhenPublished(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	suiteID, err := selectedSuiteIdentity()
	if err != nil {
		t.Fatal(err)
	}
	if suiteID != ContentHashV2CandidateManifestSHA256 {
		return
	}
	path := filepath.Join(root, "vectors", "content-hashes-v2.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if suiteID == ContentHashV2CandidateManifestSHA256 {
			t.Fatal("candidate suite lost vectors/content-hashes-v2.json")
		}
		return
	}
	if err != nil {
		t.Fatalf("read content-hashes-v2 vector: %v", err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("parse content-hashes-v2 vector: %v", err)
	}
	cases := make([]contentHashVector, 0, len(document)-2)
	for id := range document {
		if id == "schema_version" || id == "framing" {
			continue
		}
		cases = append(cases, contentHashVector{ID: id})
	}
	if suiteID == ContentHashV2CandidateManifestSHA256 && len(cases) != 5 {
		t.Fatalf("candidate content-hash-v2 executable vectors = %d, want 5", len(cases))
	}
	RunOutcomes(t, "content-hashes-v2/vectors", cases, func(testCase contentHashVector) string {
		return testCase.ID
	}, func(t *testing.T, testCase contentHashVector) Observation {
		driveContentHashVector(t, root, testCase)
		return Observation{}
	})
}
