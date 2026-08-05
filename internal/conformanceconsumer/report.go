package conformanceconsumer

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

const ReportSchemaV1 = "urn:relux-works:curator:cross-manager-report:v1"

//go:embed report.schema.json
var reportSchema []byte

// CorpusEvidence identifies input bytes without promoting them to a release.
type CorpusEvidence struct {
	Boundary        string `json:"boundary"`
	ProtocolVersion string `json:"protocol_version"`
	ManifestSHA256  string `json:"manifest_sha256"`
}

// RevisionEvidence pins every independently supplied input that can affect a
// native black-box result. Revision values are caller-provided because a
// released binary cannot reliably recover its source revision from argv.
type RevisionEvidence struct {
	Manager         string `json:"manager"`
	Version         string `json:"version"`
	Revision        string `json:"revision"`
	BinarySHA256    string `json:"binary_sha256"`
	SpecRevision    string `json:"spec_revision"`
	Toolchain       string `json:"toolchain"`
	OperatingSystem string `json:"operating_system"`
	Architecture    string `json:"architecture"`
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
	ID                  string     `json:"id"`
	Status              CaseStatus `json:"status"`
	ExitCode            int        `json:"exit_code"`
	Detail              string     `json:"detail,omitempty"`
	ObservationSHA256   string     `json:"observation_sha256,omitempty"`
	FailureArtifact     string     `json:"failure_artifact,omitempty"`
	UnexpectedProcesses []string   `json:"unexpected_processes,omitempty"`
	UnexpectedNetwork   []string   `json:"unexpected_network,omitempty"`
	UnexpectedWrites    []string   `json:"unexpected_writes,omitempty"`
	MutationOnFailure   bool       `json:"mutation_on_failure,omitempty"`
}

// Report is the machine-readable provisional consumer report.
type Report struct {
	Schema    string             `json:"schema"`
	Corpus    CorpusEvidence     `json:"corpus"`
	Adapter   string             `json:"adapter"`
	Revisions []RevisionEvidence `json:"revisions"`
	Cases     []CaseResult       `json:"cases"`
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
	if len(report.Revisions) == 0 {
		return nil, fmt.Errorf("report revision evidence is required")
	}
	copyReport := report
	copyReport.Revisions = append([]RevisionEvidence(nil), report.Revisions...)
	sort.Slice(copyReport.Revisions, func(i, j int) bool { return copyReport.Revisions[i].Manager < copyReport.Revisions[j].Manager })
	for index, revision := range copyReport.Revisions {
		if revision.Manager == "" || revision.Version == "" || revision.Revision == "" || revision.SpecRevision == "" || revision.Toolchain == "" || revision.OperatingSystem == "" || revision.Architecture == "" {
			return nil, fmt.Errorf("report revision %d is incomplete", index)
		}
		if _, err := parseDigest(revision.BinarySHA256); err != nil {
			return nil, fmt.Errorf("report revision %q binary digest: %w", revision.Manager, err)
		}
		if index > 0 && copyReport.Revisions[index-1].Manager == revision.Manager {
			return nil, fmt.Errorf("report contains duplicate manager revision %q", revision.Manager)
		}
	}
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
		if result.ObservationSHA256 != "" {
			if _, err := parseDigest(result.ObservationSHA256); err != nil {
				return nil, fmt.Errorf("report case %q observation digest: %w", result.ID, err)
			}
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
