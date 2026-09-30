package conformancecoverage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type publishedSchemaEntry struct {
	Instance string `json:"instance"`
}

type candidateSchemaCase struct {
	Name string
}

func isContentHashV2Gap(gap Gap) bool {
	switch gap.Family {
	case "agent-environment-marker-v3/schema-cases",
		"audit-record-v2/schema-cases",
		"context-lock-v2/schema-cases",
		"marker/install-marker-v5/schema-cases",
		"log-response-v3/schema-cases",
		"manager-config-v3/schema-cases",
		"registry-bundle-v2/schema-cases",
		"registry-log-entry-v2/schema-cases",
		"content-hashes-v2/vectors":
		return true
	case "skillfile-sources-v1/schema-cases":
		return strings.HasPrefix(gap.CaseID, "schema-cases/install-marker-v6/") ||
			strings.HasPrefix(gap.CaseID, "schema-cases/skillfile-lock-v2/") ||
			strings.HasPrefix(gap.CaseID, "schema-cases/source-audit-v2/")
	default:
		return false
	}
}

func TestContentHashV2GapRowsHaveTheAssignedOwner(t *testing.T) {
	suiteID, err := selectedSuiteIdentity()
	if err != nil {
		t.Fatal(err)
	}
	_, gaps, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	ownedRows := 0
	for _, gap := range gaps {
		if !isContentHashV2Gap(gap) {
			continue
		}
		if !IsContentHashV2Candidate(suiteID) {
			t.Errorf("content-hash-v2 gap %s/%s leaked into non-candidate suite %s", gap.Family, gap.CaseID, suiteID)
			continue
		}
		ownedRows++
		if gap.Owner != "TASK-260917-2tx81l" {
			t.Errorf("content-hash-v2 gap %s/%s owner = %q, want TASK-260917-2tx81l", gap.Family, gap.CaseID, gap.Owner)
		}
	}
	if IsContentHashV2Candidate(suiteID) && ownedRows != 87 {
		t.Errorf("candidate content-hash-v2 owned gap rows = %d, want 87", ownedRows)
	}
}

// TestContentHashV2FamiliesAreKnownGapsWhenPublished accounts for every case
// in the not-yet-implemented v2 schema families. The families are discovered
// from index.json and only activated when the selected root publishes them.
func TestContentHashV2FamiliesAreKnownGapsWhenPublished(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	suiteID, err := selectedSuiteIdentity()
	if err != nil {
		t.Fatal(err)
	}
	wantCandidate := IsContentHashV2Candidate(suiteID)
	indexPath := filepath.Join(root, "schema-cases", "index.json")
	indexBytes, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read published schema-case index: %v", err)
	}
	var index []publishedSchemaEntry
	if err := json.Unmarshal(indexBytes, &index); err != nil {
		t.Fatalf("parse published schema-case index: %v", err)
	}
	indexed := map[string][]string{}
	for _, entry := range index {
		family, name, ok := strings.Cut(entry.Instance, "/")
		if !ok || family == "" || name == "" || strings.Contains(name, "/") {
			t.Fatalf("schema-case index has unsupported instance path %q", entry.Instance)
		}
		indexed[family] = append(indexed[family], name)
	}

	families := []struct {
		indexFamily string
		coverageID  string
	}{
		{indexFamily: "agent-environment-marker-v3", coverageID: "agent-environment-marker-v3/schema-cases"},
		{indexFamily: "audit-record-v2", coverageID: "audit-record-v2/schema-cases"},
		{indexFamily: "context-lock-v2", coverageID: "context-lock-v2/schema-cases"},
		{indexFamily: "install-marker-v5", coverageID: "marker/install-marker-v5/schema-cases"},
		{indexFamily: "log-response-v3", coverageID: "log-response-v3/schema-cases"},
		{indexFamily: "manager-config-v3", coverageID: "manager-config-v3/schema-cases"},
		{indexFamily: "registry-bundle-v2", coverageID: "registry-bundle-v2/schema-cases"},
		{indexFamily: "registry-log-entry-v2", coverageID: "registry-log-entry-v2/schema-cases"},
	}
	for _, family := range families {
		family := family
		caseNames := indexed[family.indexFamily]
		dir := filepath.Join(root, "schema-cases", family.indexFamily)
		_, statErr := os.Stat(dir)
		if os.IsNotExist(statErr) {
			if len(caseNames) != 0 {
				t.Fatalf("schema-case index lists %s cases but its directory is absent", family.indexFamily)
			}
			if wantCandidate {
				t.Fatalf("candidate suite lost schema-cases/%s", family.indexFamily)
			}
			continue
		}
		if statErr != nil {
			t.Fatalf("inspect schema-cases/%s: %v", family.indexFamily, statErr)
		}
		if len(caseNames) == 0 {
			t.Fatalf("schema-cases/%s exists without indexed cases", family.indexFamily)
		}
		sort.Strings(caseNames)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read schema-cases/%s: %v", family.indexFamily, err)
		}
		var diskNames []string
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			diskNames = append(diskNames, entry.Name())
		}
		sort.Strings(diskNames)
		if fmt.Sprint(diskNames) != fmt.Sprint(caseNames) {
			t.Fatalf("schema-cases/%s disk/index cases differ: disk=%v index=%v", family.indexFamily, diskNames, caseNames)
		}
		cases := make([]candidateSchemaCase, 0, len(caseNames))
		for _, name := range caseNames {
			if _, err := os.ReadFile(filepath.Join(dir, name)); err != nil {
				t.Fatalf("read schema-cases/%s/%s: %v", family.indexFamily, name, err)
			}
			cases = append(cases, candidateSchemaCase{Name: name})
		}
		RunOutcomes(t, family.coverageID, cases, func(testCase candidateSchemaCase) string {
			return testCase.Name
		}, func(*testing.T, candidateSchemaCase) Observation {
			return Observation{FailureReason: "content-hash-v2 reader is owned by TASK-260917-2tx81l and is not implemented in this task"}
		})
	}
}

type contentHashVector struct {
	ID string
}

// TestContentHashV2VectorsAreKnownGapsWhenPublished counts and classifies the
// five executable vectors in content-hashes-v2.json. The schema_version and
// framing members define the vector document; they are metadata, not cases.
func TestContentHashV2VectorsAreKnownGapsWhenPublished(t *testing.T) {
	root := os.Getenv("CURATOR_CONFORMANCE_ROOT")
	if root == "" {
		t.Skip("CURATOR_CONFORMANCE_ROOT is not set")
	}
	suiteID, err := selectedSuiteIdentity()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "vectors", "content-hashes-v2.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if IsContentHashV2Candidate(suiteID) {
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
	if IsContentHashV2Candidate(suiteID) && len(cases) != 5 {
		t.Fatalf("candidate content-hash-v2 executable vectors = %d, want 5", len(cases))
	}
	RunOutcomes(t, "content-hashes-v2/vectors", cases, func(testCase contentHashVector) string {
		return testCase.ID
	}, func(*testing.T, contentHashVector) Observation {
		return Observation{FailureReason: "content-hash-v2 hashing is owned by TASK-260917-2tx81l and is not implemented in this task"}
	})
}
