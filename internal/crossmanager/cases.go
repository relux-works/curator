package crossmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// ExpectedCaseCount pins the complete accepted external-repository corpus.
const ExpectedCaseCount = 60

// Case is one implementation-neutral case from case-manifest.json. Expected
// remains JSON so the runner cannot accidentally narrow the normative corpus.
type Case struct {
	ID       string          `json:"id"`
	Category string          `json:"category"`
	Source   string          `json:"source"`
	Expected json.RawMessage `json:"expected"`
}

type caseManifest struct {
	SchemaVersion         int             `json:"schema_version"`
	CorpusVersion         string          `json:"corpus_version"`
	ProtocolVersion       string          `json:"protocol_version"`
	ImplementationNeutral bool            `json:"implementation_neutral"`
	ManagerAdapter        any             `json:"manager_adapter"`
	PhysicalPaths         string          `json:"physical_paths"`
	ArchitectureCoverage  json.RawMessage `json:"architecture_v6_coverage"`
	ArchitectureMatrix    json.RawMessage `json:"architecture_v6_threat_matrix"`
	LifecycleBoundaries   json.RawMessage `json:"lifecycle_boundaries"`
	LifecycleMatrix       json.RawMessage `json:"lifecycle_matrix"`
	Cases                 []Case          `json:"cases"`
}

// Cases authenticates and strictly decodes the accepted 60-case manifest.
func (c *Corpus) Cases() ([]Case, error) {
	if c == nil {
		return nil, fmt.Errorf("corpus is required")
	}
	payload, _, err := c.Read("case-manifest.json")
	if err != nil {
		return nil, err
	}
	var manifest caseManifest
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode case manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("decode case manifest: trailing JSON value")
	}
	if manifest.SchemaVersion != 1 || manifest.CorpusVersion != CorpusRC5 || manifest.ProtocolVersion != ProtocolRC5 {
		return nil, fmt.Errorf("case manifest identity does not match accepted rc.5 corpus")
	}
	if !manifest.ImplementationNeutral || manifest.PhysicalPaths != "implementation-specific" {
		return nil, fmt.Errorf("case manifest implementation boundary is invalid")
	}
	if len(manifest.Cases) != ExpectedCaseCount {
		return nil, fmt.Errorf("case manifest has %d cases, want %d", len(manifest.Cases), ExpectedCaseCount)
	}
	seen := make(map[string]struct{}, len(manifest.Cases))
	result := append([]Case(nil), manifest.Cases...)
	for index := range result {
		item := &result[index]
		if item.ID == "" || item.Category == "" || item.Source == "" || len(item.Expected) == 0 {
			return nil, fmt.Errorf("case %d is incomplete", index)
		}
		if _, duplicate := seen[item.ID]; duplicate {
			return nil, fmt.Errorf("duplicate case id %q", item.ID)
		}
		seen[item.ID] = struct{}{}
		file, _, _ := strings.Cut(item.Source, "#")
		if err := validateCorpusPath(file); err != nil {
			return nil, fmt.Errorf("case %q source: %w", item.ID, err)
		}
		if !strings.HasPrefix(file, "conformance/v1/") {
			if _, _, err := c.Read(file); err != nil {
				return nil, fmt.Errorf("case %q source: %w", item.ID, err)
			}
		}
		var expected map[string]any
		if err := json.Unmarshal(item.Expected, &expected); err != nil || expected["outcome"] == nil {
			return nil, fmt.Errorf("case %q expected object is invalid", item.ID)
		}
	}
	return result, nil
}

// SortedCaseIDs is useful for proving both adapters receive exactly the same set.
func SortedCaseIDs(cases []Case) []string {
	ids := make([]string, len(cases))
	for index, item := range cases {
		ids[index] = item.ID
	}
	sort.Strings(ids)
	return ids
}
