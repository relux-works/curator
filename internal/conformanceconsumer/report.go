package conformanceconsumer

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

const ReportSchemaV1 = "urn:relux-works:curator:black-box-consumer-report:v1"

//go:embed report.schema.json
var reportSchema []byte

// CorpusEvidence identifies input bytes without promoting them to a release.
type CorpusEvidence struct {
	Boundary        string `json:"boundary"`
	ProtocolVersion string `json:"protocol_version"`
	ManifestSHA256  string `json:"manifest_sha256"`
}

// CaseStatus is an observation state, not a conformance verdict.
type CaseStatus string

const (
	CaseObserved CaseStatus = "observed"
	CaseMismatch CaseStatus = "mismatch"
	CaseError    CaseStatus = "error"
	CaseNotRun   CaseStatus = "not-run"
)

// CaseResult is one deterministic black-box observation.
type CaseResult struct {
	ID       string     `json:"id"`
	Status   CaseStatus `json:"status"`
	ExitCode int        `json:"exit_code"`
	Detail   string     `json:"detail,omitempty"`
}

// Report is the machine-readable provisional consumer report.
type Report struct {
	Schema  string         `json:"schema"`
	Corpus  CorpusEvidence `json:"corpus"`
	Adapter string         `json:"adapter"`
	Cases   []CaseResult   `json:"cases"`
}

// ReportJSON validates and deterministically encodes a report. Cases are
// sorted by ID so process scheduling cannot affect output bytes.
func ReportJSON(report Report) ([]byte, error) {
	if report.Schema != ReportSchemaV1 {
		return nil, fmt.Errorf("report schema must be %q", ReportSchemaV1)
	}
	if report.Corpus.Boundary != CorpusBoundaryV1 || report.Corpus.ProtocolVersion == "" {
		return nil, fmt.Errorf("report corpus boundary is invalid")
	}
	if _, err := parseDigest(report.Corpus.ManifestSHA256); err != nil {
		return nil, fmt.Errorf("report manifest digest: %w", err)
	}
	if report.Adapter == "" {
		return nil, fmt.Errorf("report adapter is required")
	}
	copyReport := report
	copyReport.Cases = append([]CaseResult(nil), report.Cases...)
	sort.Slice(copyReport.Cases, func(i, j int) bool { return copyReport.Cases[i].ID < copyReport.Cases[j].ID })
	for index, result := range copyReport.Cases {
		if result.ID == "" {
			return nil, fmt.Errorf("report case %d has no id", index)
		}
		if index > 0 && copyReport.Cases[index-1].ID == result.ID {
			return nil, fmt.Errorf("report contains duplicate case %q", result.ID)
		}
		switch result.Status {
		case CaseObserved, CaseMismatch, CaseError, CaseNotRun:
		default:
			return nil, fmt.Errorf("report case %q has invalid status %q", result.ID, result.Status)
		}
	}
	payload, err := json.MarshalIndent(copyReport, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// ReportSchema returns a defensive copy of the embedded JSON Schema.
func ReportSchema() []byte { return append([]byte(nil), reportSchema...) }
