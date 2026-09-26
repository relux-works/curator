package crossconformance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/relux-works/curator/internal/conformancecoverage"
)

// skillfileSourcesPin is the exact curator-spec v1.0.0-rc.13 tag commit
// carrying Decision 0022 and the released skillfile-sources-v1 corpus.
const skillfileSourcesPin = "23435129ebc4c29e5b7f75ec72a0aa0cd3f16065"

const skillfileSourcesTestdata = "testdata/skillfile-sources-v1"

func draftCorpusDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(skillfileSourcesTestdata, "corpus", "conformance", "skillfile-sources-v1")
	if _, err := os.Stat(filepath.Join(dir, "index.json")); err != nil {
		t.Fatalf("vendored draft corpus missing: %v", err)
	}
	return dir
}

func readDraftFile(t *testing.T, path string) []byte {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return payload
}

type draftSchemaEntry struct {
	Schema   string `json:"schema"`
	Instance string `json:"instance"`
	Valid    bool   `json:"valid"`
}

func loadDraftIndex(t *testing.T) []draftSchemaEntry {
	t.Helper()
	var entries []draftSchemaEntry
	if err := json.Unmarshal(readDraftFile(t, filepath.Join(draftCorpusDir(t), "index.json")), &entries); err != nil {
		t.Fatalf("parse index.json: %v", err)
	}
	return entries
}

type draftSemanticCase struct {
	ID       string         `json:"id"`
	Input    map[string]any `json:"input"`
	Expected string         `json:"expected"`
}

func loadDraftSemantic(t *testing.T) []draftSemanticCase {
	t.Helper()
	var cases []draftSemanticCase
	if err := json.Unmarshal(readDraftFile(t, filepath.Join(draftCorpusDir(t), "semantic-cases.json")), &cases); err != nil {
		t.Fatalf("parse semantic-cases.json: %v", err)
	}
	return cases
}

type draftSnapshotFile struct {
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Executable bool   `json:"executable"`
}

type draftSnapshotInventory struct {
	SchemaVersion int                 `json:"schema_version"`
	Algorithm     string              `json:"algorithm"`
	Files         []draftSnapshotFile `json:"files"`
	Snapshot      string              `json:"snapshot"`
}

type draftSnapshotVector struct {
	ID        string                 `json:"id"`
	UTF8Files map[string]string      `json:"utf8_files"`
	Inventory draftSnapshotInventory `json:"inventory"`
}

func loadDraftSnapshots(t *testing.T) []draftSnapshotVector {
	t.Helper()
	var vectors []draftSnapshotVector
	if err := json.Unmarshal(readDraftFile(t, filepath.Join(draftCorpusDir(t), "snapshot-cases.json")), &vectors); err != nil {
		t.Fatalf("parse snapshot-cases.json: %v", err)
	}
	return vectors
}

// TestSkillfileSourcesPin pins the vendored corpus to the exact spec
// revision: the pin file must equal the constant, and every vendored
// byte must match MANIFEST.sha256 with no missing or extra files.
func TestSkillfileSourcesPin(t *testing.T) {
	pin := strings.TrimSpace(string(readDraftFile(t, filepath.Join(skillfileSourcesTestdata, "SKILLFILE_SOURCES_PIN"))))
	if pin != skillfileSourcesPin {
		t.Fatalf("SKILLFILE_SOURCES_PIN = %q, want %q", pin, skillfileSourcesPin)
	}
	manifest := string(readDraftFile(t, filepath.Join(skillfileSourcesTestdata, "MANIFEST.sha256")))
	want := map[string]string{}
	for _, line := range strings.Split(manifest, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			t.Fatalf("malformed manifest line %q", line)
		}
		want[filepath.Clean(parts[1])] = parts[0]
	}
	root := filepath.Join(skillfileSourcesTestdata, "corpus")
	var seen []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		seen = append(seen, rel)
		payload, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(payload)
		key := filepath.Clean("./" + rel)
		if want[key] == "" {
			t.Errorf("vendored file %q is not in MANIFEST.sha256", rel)
			return nil
		}
		if hex.EncodeToString(sum[:]) != want[key] {
			t.Errorf("vendored file %q drifted from the pinned revision", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != len(want) {
		t.Fatalf("vendored files = %d, manifest entries = %d", len(seen), len(want))
	}
}

// TestSkillfileSourcesCorpusCounts asserts the exact released counts and that
// every indexed instance file exists.
func TestSkillfileSourcesCorpusCounts(t *testing.T) {
	dir := draftCorpusDir(t)
	countPins, _, err := conformancecoverage.Load()
	if err != nil {
		t.Fatal(err)
	}
	entries := loadDraftIndex(t)
	if len(entries) != countPins["skillfile-sources-v1/schema-cases"] {
		t.Fatalf("schema cases = %d, want %d at pin %s", len(entries), countPins["skillfile-sources-v1/schema-cases"], skillfileSourcesPin)
	}
	for _, entry := range entries {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(entry.Instance))); err != nil {
			t.Errorf("indexed instance missing: %s: %v", entry.Instance, err)
		}
	}
	semantic := loadDraftSemantic(t)
	wantSemantic := countPins["skillfile-sources-v1/semantic-cases/0"] +
		countPins["skillfile-sources-v1/semantic-cases/1"] +
		countPins["skillfile-sources-v1/semantic-cases/2"] +
		countPins["skillfile-sources-v1/semantic-cases/3"] +
		countPins["skillfile-sources-v1/semantic-cases/4"]
	if len(semantic) != wantSemantic {
		t.Fatalf("semantic cases = %d, want %d at pin %s", len(semantic), wantSemantic, skillfileSourcesPin)
	}
	ids := map[string]bool{}
	for _, c := range semantic {
		if ids[c.ID] {
			t.Errorf("duplicate semantic id %q", c.ID)
		}
		ids[c.ID] = true
	}
	vectors := loadDraftSnapshots(t)
	if len(vectors) != countPins["skillfile-sources-v1/snapshot-cases"] {
		t.Fatalf("snapshot vectors = %d, want %d at pin %s", len(vectors), countPins["skillfile-sources-v1/snapshot-cases"], skillfileSourcesPin)
	}
}
