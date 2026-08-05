package conformanceconsumer

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// ProtocolObservation contains only normative, implementation-neutral facts.
// Adapters must translate private manager paths into logical names before
// constructing this value.
type ProtocolObservation struct {
	CaseID            string            `json:"case_id"`
	ExitCode          int               `json:"exit_code"`
	Stdout            string            `json:"stdout"`
	Stderr            string            `json:"stderr"`
	TypedError        string            `json:"typed_error,omitempty"`
	Identities        map[string]string `json:"identities,omitempty"`
	Digests           map[string]string `json:"digests,omitempty"`
	Receipts          map[string]string `json:"receipts,omitempty"`
	Markers           map[string]string `json:"markers,omitempty"`
	States            map[string]string `json:"states,omitempty"`
	Files             []FileObservation `json:"files,omitempty"`
	UnexpectedProcess []string          `json:"unexpected_processes,omitempty"`
	UnexpectedNetwork []string          `json:"unexpected_network,omitempty"`
	UnexpectedWrites  []string          `json:"unexpected_writes,omitempty"`
	MutationOnFailure bool              `json:"mutation_on_failure,omitempty"`
}

// Comparison is a deterministic parity result. Mismatch names are stable and
// intentionally exclude implementation-private physical paths.
type Comparison struct {
	CaseID      string   `json:"case_id"`
	Equal       bool     `json:"equal"`
	Mismatches  []string `json:"mismatches,omitempty"`
	LeftSHA256  string   `json:"left_sha256"`
	RightSHA256 string   `json:"right_sha256"`
}

// CompareProtocol compares all protocol-required observations exactly.
func CompareProtocol(left, right ProtocolObservation) (Comparison, error) {
	if left.CaseID == "" || right.CaseID == "" || left.CaseID != right.CaseID {
		return Comparison{}, fmt.Errorf("matching non-empty case ids are required")
	}
	normalizeProtocolObservation(&left)
	normalizeProtocolObservation(&right)
	leftPayload, err := json.Marshal(left)
	if err != nil {
		return Comparison{}, err
	}
	rightPayload, err := json.Marshal(right)
	if err != nil {
		return Comparison{}, err
	}
	comparison := Comparison{CaseID: left.CaseID, LeftSHA256: digestBytes(leftPayload), RightSHA256: digestBytes(rightPayload)}
	fields := []struct {
		name        string
		left, right any
	}{
		{"exit_code", left.ExitCode, right.ExitCode},
		{"stdout", left.Stdout, right.Stdout},
		{"stderr", left.Stderr, right.Stderr},
		{"typed_error", left.TypedError, right.TypedError},
		{"identities", left.Identities, right.Identities},
		{"digests", left.Digests, right.Digests},
		{"receipts", left.Receipts, right.Receipts},
		{"markers", left.Markers, right.Markers},
		{"states", left.States, right.States},
		{"files", left.Files, right.Files},
		{"unexpected_processes", left.UnexpectedProcess, right.UnexpectedProcess},
		{"unexpected_network", left.UnexpectedNetwork, right.UnexpectedNetwork},
		{"unexpected_writes", left.UnexpectedWrites, right.UnexpectedWrites},
		{"mutation_on_failure", left.MutationOnFailure, right.MutationOnFailure},
	}
	for _, field := range fields {
		if !reflect.DeepEqual(field.left, field.right) {
			comparison.Mismatches = append(comparison.Mismatches, field.name)
		}
	}
	comparison.Equal = len(comparison.Mismatches) == 0
	return comparison, nil
}

func normalizeProtocolObservation(observation *ProtocolObservation) {
	sort.Slice(observation.Files, func(i, j int) bool { return observation.Files[i].Path < observation.Files[j].Path })
	observation.UnexpectedProcess = sortedUnique(observation.UnexpectedProcess)
	observation.UnexpectedNetwork = sortedUnique(observation.UnexpectedNetwork)
	observation.UnexpectedWrites = sortedUnique(observation.UnexpectedWrites)
}
