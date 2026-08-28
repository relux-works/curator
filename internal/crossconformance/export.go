package crossconformance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

// ExportSchemaID versions the cross-adapter integration contract this package
// publishes to independent implementations.
const ExportSchemaID = "curator-cross-adapter-conformance-export-v1"

// Requirement returns the normative sentence one obligation states. It is the
// exported form of the contract, so an implementation in another language can
// state the same requirement without reading Go.
func (obligation Obligation) Requirement() string {
	switch obligation {
	case ObligationSelectionNeutralCapture:
		return "The selection-neutral capture contains no exact target platform, no toolchain component, and no targets or uses_tool edge."
	case ObligationCaptureStableAcrossTargets:
		return "Two different requested targets over the same source inputs produce one exact capture identity."
	case ObligationBindingOwnsTargetAuthority:
		return "The exact target platform and every concrete tool identity enter only through the selection overlay, bound by an explicit targets edge where the path emits edges."
	case ObligationSelectionDivergesPerTarget:
		return "The selection-bound identity, the active projection, and the build plan all differ between two targets."
	case ObligationDeterministicProjection:
		return "Repeated projection of the same inputs produces byte-identical identities."
	case ObligationCausalEvidenceChain:
		return "Every emitted checkpoint names its exact predecessor, C5 adds no graph record, and every pre-C5 evidence derivation answers with a causal receipt."
	case ObligationSharedArtifactAdmission:
		return "One deny-class dependency payload produces one shared class, decision, primary diagnostic, and leaf digest through every adapter, and each profile admits exactly its own source grammars."
	default:
		return ""
	}
}

// ProtocolExport renders the complete cross-adapter integration contract as
// exact CCJ-1 bytes: the accepted corpus with its independently derived
// identities, the counters this package derived from it, the normative
// obligations, the delivered paths, and the published rejection matrix.
//
// The export deliberately contains no Go type, no file path, and no host
// value. An independent implementation consumes it as data.
func ProtocolExport(corpus Corpus, report Report) ([]byte, error) {
	records := make([]Value, 0, len(corpus.Records))
	for _, name := range corpus.Names() {
		record, ok := corpus.ByName(name)
		if !ok {
			return nil, fmt.Errorf("corpus lost record %q", name)
		}
		payload, err := RequireCanonical(record.Payload)
		if err != nil {
			return nil, fmt.Errorf("record %q: %w", name, err)
		}
		records = append(records, Object(
			Field("name", Text(record.Name)),
			Field("label", Text(record.Label)),
			Field("id", Text(record.Derived)),
			Field("payload", payload),
		))
	}

	obligations := make([]Value, 0, len(Obligations()))
	for _, obligation := range Obligations() {
		requirement := obligation.Requirement()
		if requirement == "" {
			return nil, fmt.Errorf("obligation %q publishes no requirement", obligation)
		}
		obligations = append(obligations, Object(
			Field("id", Text(string(obligation))),
			Field("requirement", Text(requirement)),
		))
	}

	vectors := make([]Value, 0, len(RejectionVectors()))
	for _, vector := range RejectionVectors() {
		codes := append([]string(nil), vector.Codes...)
		sort.Strings(codes)
		owners := append([]string(nil), vector.OwnedBy...)
		sort.Strings(owners)
		vectors = append(vectors, Object(
			Field("id", Text(vector.ID)),
			Field("family", Text(vector.Family)),
			Field("requirement", Text(vector.Requirement)),
			Field("codes", TextArray(codes)),
			Field("owned_by", TextArray(owners)),
			Field("cross_drivable", Boolean(vector.CrossDrivable())),
		))
	}

	paths := make([]string, 0, len(DeliveredPaths()))
	for _, path := range DeliveredPaths() {
		paths = append(paths, string(path))
	}
	sort.Strings(paths)

	document := Object(
		Field("schema_id", Text(ExportSchemaID)),
		Field("corpus", Object(
			Field("sha256", Text(AcceptedCorpusSHA256)),
			Field("record_count", Integer(int64(len(corpus.Records)))),
			Field("records", Array(records...)),
		)),
		Field("derived", Object(
			Field("labeled_records", Integer(int64(report.LabeledRecords))),
			Field("resolved_references", Integer(int64(report.ResolvedReferences))),
			Field("artifact_manifest_references", Integer(int64(report.ArtifactManifestRefs))),
			Field("cgp05_capture_reused", Boolean(report.CGP05CaptureReused)),
			Field("cgp05_target_branches", Integer(int64(report.CGP05TargetBranches))),
			Field("explicit_target_bindings", Integer(int64(report.ExplicitTargetEdges))),
			Field("cgp05_divergent_kinds", TextArray(report.CGP05DivergentKinds)),
			Field("cgp10_observation_branches", Integer(int64(report.CGP10ObservationSets))),
			Field("cgp10_stable_records", TextArray(report.CGP10StableRecords)),
			Field("cgp10_all_refs_resolve", Boolean(report.CGP10AllRefsResolve)),
		)),
		Field("delivered_paths", TextArray(paths)),
		Field("obligations", Array(obligations...)),
		Field("rejection_matrix", Array(vectors...)),
	)
	return Canonical(document)
}

// ExportDigest is the domain-separated identity of an export document.
func ExportDigest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
