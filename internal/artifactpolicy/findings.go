package artifactpolicy

import (
	"bytes"
	"container/heap"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
)

type findingAccumulator struct {
	maximum  int
	total    int64
	records  diagnosticMaxHeap
	evidence []FindingEvidence
}

func newFindingAccumulator(maximum int64) *findingAccumulator {
	return &findingAccumulator{maximum: int(maximum)}
}

func (accumulator *findingAccumulator) add(diagnostic Diagnostic) error {
	canonicalizeDiagnostic(&diagnostic)
	evidence, err := findingEvidenceFromDiagnostic(diagnostic)
	if err != nil {
		return err
	}
	if accumulator.total == math.MaxInt64 {
		return fmt.Errorf("finding count overflows")
	}
	accumulator.total++
	accumulator.evidence = append(accumulator.evidence, evidence)
	if accumulator.maximum <= 0 {
		return nil
	}
	if len(accumulator.records) < accumulator.maximum {
		heap.Push(&accumulator.records, diagnostic)
		return nil
	}
	if diagnosticLess(diagnostic, accumulator.records[0]) {
		heap.Pop(&accumulator.records)
		heap.Push(&accumulator.records, diagnostic)
	}
	return nil
}

func (accumulator *findingAccumulator) recorded() []Diagnostic {
	result := make([]Diagnostic, len(accumulator.records))
	copy(result, accumulator.records)
	sort.SliceStable(result, func(left, right int) bool {
		return diagnosticLess(result[left], result[right])
	})
	return result
}

func (accumulator *findingAccumulator) summary() FindingsSummary {
	evidence := make([]FindingEvidence, len(accumulator.evidence))
	copy(evidence, accumulator.evidence)
	sort.Slice(evidence, func(left, right int) bool { return findingEvidenceLess(evidence[left], evidence[right]) })
	return summarizeFindingEvidence(evidence, int64(len(accumulator.records)))
}

func findingEvidenceFromDiagnostic(diagnostic Diagnostic) (FindingEvidence, error) {
	canonicalizeDiagnostic(&diagnostic)
	payload, err := marshalCanonicalStruct(diagnostic)
	if err != nil {
		return FindingEvidence{}, fmt.Errorf("canonicalize diagnostic: %w", err)
	}
	detailPayload, err := marshalCanonicalStruct(diagnostic.Details)
	if err != nil {
		return FindingEvidence{}, fmt.Errorf("canonicalize diagnostic details: %w", err)
	}
	return FindingEvidence{
		DiagnosticSHA256: digestBytes(payload), Code: diagnostic.Code, Path: diagnostic.Path,
		OriginalNameBase64: diagnostic.OriginalNameBase64, CollisionKey: diagnostic.CollisionKey,
		Class: diagnostic.Class, Variant: diagnostic.Variant, DetectorID: diagnostic.DetectorID,
		Reason: diagnostic.Reason, ContainerChain: append([]string{}, diagnostic.ContainerChain...),
		SHA256: diagnostic.SHA256, Size: diagnostic.Size, LimitName: diagnostic.LimitName,
		Limit: diagnostic.Limit, Observed: diagnostic.Observed,
		Details: append([]Fact{}, diagnostic.Details...), DetailsSHA256: digestBytes(detailPayload),
	}, nil
}

func summarizeFindingEvidence(evidence []FindingEvidence, recorded int64) FindingsSummary {
	digest := sha256.New()
	_, _ = digest.Write([]byte(findingsDigestAlgorithm))
	_, _ = digest.Write([]byte{0})
	var count [8]byte
	binary.BigEndian.PutUint64(count[:], uint64(len(evidence))) // #nosec G115 -- slice lengths are nonnegative.
	_, _ = digest.Write(count[:])
	for _, item := range evidence {
		identity, err := hex.DecodeString(stringsTrimSHA256(item.DiagnosticSHA256))
		if err != nil || len(identity) != sha256.Size {
			// Validation rejects this state. Keep summary total/digest
			// deterministic for callers constructing negative fixtures.
			identity = make([]byte, sha256.Size)
		}
		_, _ = digest.Write(identity)
	}
	return FindingsSummary{
		Algorithm: findingsDigestAlgorithm,
		Total:     int64(len(evidence)),
		Recorded:  recorded,
		Evidence:  evidence,
		SHA256:    "sha256:" + hex.EncodeToString(digest.Sum(nil)),
	}
}

func stringsTrimSHA256(value string) string {
	if len(value) == len("sha256:")+sha256.Size*2 && bytes.HasPrefix([]byte(value), []byte("sha256:")) {
		return value[len("sha256:"):]
	}
	return value
}

func findingEvidenceLess(left, right FindingEvidence) bool {
	leftDiagnostic := diagnosticFromFindingEvidence(left)
	rightDiagnostic := diagnosticFromFindingEvidence(right)
	if diagnosticLess(leftDiagnostic, rightDiagnostic) {
		return true
	}
	if diagnosticLess(rightDiagnostic, leftDiagnostic) {
		return false
	}
	return left.DiagnosticSHA256 < right.DiagnosticSHA256
}

func diagnosticFromFindingEvidence(evidence FindingEvidence) Diagnostic {
	return Diagnostic{
		Code: evidence.Code, Path: evidence.Path, OriginalNameBase64: evidence.OriginalNameBase64,
		CollisionKey: evidence.CollisionKey, Class: evidence.Class, Variant: evidence.Variant,
		DetectorID: evidence.DetectorID, Reason: evidence.Reason,
		ContainerChain: append([]string{}, evidence.ContainerChain...), SHA256: evidence.SHA256,
		Size: evidence.Size, LimitName: evidence.LimitName, Limit: evidence.Limit, Observed: evidence.Observed,
		Details: append([]Fact{}, evidence.Details...),
	}
}

type diagnosticMaxHeap []Diagnostic

func (values diagnosticMaxHeap) Len() int { return len(values) }

func (values diagnosticMaxHeap) Less(left, right int) bool {
	return diagnosticLess(values[right], values[left])
}

func (values diagnosticMaxHeap) Swap(left, right int) {
	values[left], values[right] = values[right], values[left]
}

func (values *diagnosticMaxHeap) Push(value any) {
	*values = append(*values, value.(Diagnostic))
}

func (values *diagnosticMaxHeap) Pop() any {
	current := *values
	last := current[len(current)-1]
	*values = current[:len(current)-1]
	return last
}

func findingsFromDiagnostics(maximum int64, diagnostics []Diagnostic) (*findingAccumulator, error) {
	accumulator := newFindingAccumulator(maximum)
	for _, diagnostic := range diagnostics {
		if err := accumulator.add(diagnostic); err != nil {
			return nil, err
		}
	}
	return accumulator, nil
}
