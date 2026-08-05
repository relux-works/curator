package crossmanager

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// Comparison is a stable cross-manager result over normalized normative data.
type Comparison struct {
	CaseID      string   `json:"case_id"`
	Equal       bool     `json:"equal"`
	Mismatches  []string `json:"mismatches,omitempty"`
	LeftSHA256  string   `json:"left_sha256"`
	RightSHA256 string   `json:"right_sha256"`
}

// CompareReports compares case status, exit code, expected corpus bytes, and
// adapter-normalized protocol observations. Artifact and physical paths are
// deliberately excluded.
func CompareReports(left, right Report) ([]Comparison, error) {
	if left.Manager.Name == "" || right.Manager.Name == "" || left.Manager.Name == right.Manager.Name {
		return nil, fmt.Errorf("reports from two distinct managers are required")
	}
	if !reflect.DeepEqual(left.Corpus, right.Corpus) {
		return nil, fmt.Errorf("reports use different corpus evidence")
	}
	leftCases, err := indexCases(left.Cases)
	if err != nil {
		return nil, fmt.Errorf("left report: %w", err)
	}
	rightCases, err := indexCases(right.Cases)
	if err != nil {
		return nil, fmt.Errorf("right report: %w", err)
	}
	if len(leftCases) != len(rightCases) {
		return nil, fmt.Errorf("reports contain different case counts")
	}
	ids := make([]string, 0, len(leftCases))
	for id := range leftCases {
		if _, exists := rightCases[id]; !exists {
			return nil, fmt.Errorf("right report is missing case %q", id)
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]Comparison, 0, len(ids))
	for _, id := range ids {
		leftCase, rightCase := leftCases[id], rightCases[id]
		leftCanonical, err := canonicalCaseComparison(leftCase)
		if err != nil {
			return nil, fmt.Errorf("left case %q: %w", id, err)
		}
		rightCanonical, err := canonicalCaseComparison(rightCase)
		if err != nil {
			return nil, fmt.Errorf("right case %q: %w", id, err)
		}
		comparison := Comparison{CaseID: id, LeftSHA256: digestBytes(leftCanonical), RightSHA256: digestBytes(rightCanonical)}
		if leftCase.Status != rightCase.Status {
			comparison.Mismatches = append(comparison.Mismatches, "status")
		}
		if leftCase.ExitCode != rightCase.ExitCode {
			comparison.Mismatches = append(comparison.Mismatches, "exit_code")
		}
		if !jsonEqual(leftCase.Expected, rightCase.Expected) {
			comparison.Mismatches = append(comparison.Mismatches, "expected")
		}
		if !jsonEqual(leftCase.Observed, rightCase.Observed) {
			comparison.Mismatches = append(comparison.Mismatches, "observed")
		}
		if !reflect.DeepEqual(leftCase.Violations, rightCase.Violations) {
			comparison.Mismatches = append(comparison.Mismatches, "violations")
		}
		comparison.Equal = len(comparison.Mismatches) == 0
		result = append(result, comparison)
	}
	return result, nil
}

func indexCases(cases []CaseResult) (map[string]CaseResult, error) {
	result := make(map[string]CaseResult, len(cases))
	for _, item := range cases {
		if item.ID == "" {
			return nil, fmt.Errorf("case id is required")
		}
		if _, duplicate := result[item.ID]; duplicate {
			return nil, fmt.Errorf("duplicate case %q", item.ID)
		}
		result[item.ID] = item
	}
	return result, nil
}

func canonicalCaseComparison(item CaseResult) ([]byte, error) {
	value := struct {
		Status     CaseStatus      `json:"status"`
		ExitCode   int             `json:"exit_code"`
		Expected   json.RawMessage `json:"expected"`
		Observed   json.RawMessage `json:"observed"`
		Violations []string        `json:"violations,omitempty"`
	}{item.Status, item.ExitCode, item.Expected, item.Observed, item.Violations}
	return json.Marshal(value)
}

func jsonEqual(left, right json.RawMessage) bool {
	var leftValue, rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return string(left) == string(right)
	}
	return reflect.DeepEqual(leftValue, rightValue)
}
