package crossmanager

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

// ReportSchemaV1 identifies the stable machine-readable report contract.
const ReportSchemaV1 = "urn:relux-works:curator:cross-manager-report:v1"

//go:embed report.schema.json
var reportSchema []byte

// CorpusEvidence identifies input bytes without promoting them to a release.
type CorpusEvidence struct {
	Boundary        string `json:"boundary"`
	ProtocolVersion string `json:"protocol_version"`
	Revision        string `json:"revision"`
	ManifestSHA256  string `json:"manifest_sha256"`
}

// CaseStatus is an observation state, not a conformance verdict.
type CaseStatus string

const (
	// CaseObserved means the process completed and matched the normative subset.
	CaseObserved CaseStatus = "observed"
	// CaseMismatch means the observation differed from normative expectations.
	CaseMismatch CaseStatus = "mismatch"
	// CaseError means the runner could not produce a complete observation.
	CaseError CaseStatus = "error"
	// CaseNotRun means no process execution was attempted.
	CaseNotRun CaseStatus = "not-run"
)

// CaseResult is one deterministic black-box observation.
type CaseResult struct {
	ID            string          `json:"id"`
	Status        CaseStatus      `json:"status"`
	ExitCode      int             `json:"exit_code"`
	Expected      json.RawMessage `json:"expected,omitempty"`
	Observed      json.RawMessage `json:"observed,omitempty"`
	StdoutSHA256  string          `json:"stdout_sha256,omitempty"`
	StderrSHA256  string          `json:"stderr_sha256,omitempty"`
	ChangedPaths  []string        `json:"changed_paths,omitempty"`
	OutsideWrites []string        `json:"outside_writes,omitempty"`
	Boundary      BoundaryEvents  `json:"boundary,omitempty"`
	Violations    []string        `json:"violations,omitempty"`
	ArtifactDir   string          `json:"artifact_dir,omitempty"`
	Detail        string          `json:"detail,omitempty"`
}

// Report is the machine-readable provisional consumer report.
type Report struct {
	Schema  string          `json:"schema"`
	Corpus  CorpusEvidence  `json:"corpus"`
	Adapter string          `json:"adapter"`
	Manager ManagerEvidence `json:"manager"`
	Cases   []CaseResult    `json:"cases"`
}

// ReportJSON validates and deterministically encodes a report. Cases are
// sorted by ID so process scheduling cannot affect output bytes.
func ReportJSON(report Report) ([]byte, error) {
	if report.Schema != ReportSchemaV1 {
		return nil, fmt.Errorf("report schema must be %q", ReportSchemaV1)
	}
	if report.Corpus.Boundary != CorpusBoundaryV1 || report.Corpus.ProtocolVersion == "" || report.Corpus.Revision == "" {
		return nil, fmt.Errorf("report corpus boundary is invalid")
	}
	if _, err := parseDigest(report.Corpus.ManifestSHA256); err != nil {
		return nil, fmt.Errorf("report manifest digest: %w", err)
	}
	if report.Adapter == "" {
		return nil, fmt.Errorf("report adapter is required")
	}
	if err := validateManagerEvidence(report.Manager); err != nil {
		return nil, err
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
