package artifactpolicy

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/protocoljson"
)

var sha256Identity = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// EncodeManifest validates and returns exact CCJ-1 artifact-manifest-v1 bytes.
// The supplied digest must already match the canonical digest projection.
func EncodeManifest(manifest Manifest) ([]byte, error) {
	if err := validateManifest(manifest); err != nil {
		return nil, err
	}
	want, err := manifestDigest(manifest)
	if err != nil {
		return nil, err
	}
	if manifest.ManifestDigest != want {
		return nil, fmt.Errorf("manifest_digest is %q, want %q", manifest.ManifestDigest, want)
	}
	return marshalCanonicalStruct(manifest)
}

// DecodeManifest requires exact canonical bytes, rejects unknown fields, and
// verifies every closed enum and the manifest digest.
func DecodeManifest(payload []byte) (Manifest, error) {
	var manifest Manifest
	if err := protocoljson.UnmarshalCanonical(payload, &manifest); err != nil {
		return Manifest{}, err
	}
	if err := validateManifest(manifest); err != nil {
		return Manifest{}, err
	}
	want, err := manifestDigest(manifest)
	if err != nil {
		return Manifest{}, err
	}
	if manifest.ManifestDigest != want {
		return Manifest{}, fmt.Errorf("manifest_digest mismatch")
	}
	return manifest, nil
}

func sealManifest(manifest *Manifest, findings *findingAccumulator) error {
	if findings == nil {
		findings = newFindingAccumulator(manifest.LimitVector.MaxRecordedFindings)
	}
	recorded := findings.recorded()
	failureEvidence, err := traversalFailureEvidence(manifest.Accounting, recorded)
	if err != nil {
		return err
	}
	manifest.TraversalFailure = failureEvidence
	canonicalizeManifest(manifest, recorded)
	manifest.Diagnostics = recorded
	manifest.Findings = findings.summary()
	manifest.ManifestDigest = ""
	digest, err := manifestDigest(*manifest)
	if err != nil {
		return err
	}
	manifest.ManifestDigest = digest
	return validateManifest(*manifest)
}

func traversalFailureEvidence(accounting TraversalAccounting, diagnostics []Diagnostic) (TraversalFailureEvidence, error) {
	if accounting.UnmanifestedEntryCount == 0 && accounting.UnmanifestedEmittedBytes == 0 &&
		accounting.MaxUnmanifestedLeafBytes == 0 {
		return TraversalFailureEvidence{}, nil
	}
	for _, diagnostic := range diagnostics {
		if !diagnosticSupportsTraversalFailure(diagnostic.Code) {
			continue
		}
		return TraversalFailureEvidence{
			Code: diagnostic.Code, Path: diagnostic.Path, Reason: diagnostic.Reason,
			UnmanifestedEntryCount:     accounting.UnmanifestedEntryCount,
			UnmanifestedEmittedBytes:   accounting.UnmanifestedEmittedBytes,
			MaxUnmanifestedLeafBytes:   accounting.MaxUnmanifestedLeafBytes,
			MaxObservedStreamInput:     accounting.MaxStreamInputBytes,
			MaxObservedStreamEmitted:   accounting.MaxStreamEmittedBytes,
			MaxObservedStreamExpansion: accounting.MaxStreamExpansionRatio,
		}, nil
	}
	return TraversalFailureEvidence{}, fmt.Errorf("unmanifested traversal work has no recorded structural cause")
}

func diagnosticSupportsTraversalFailure(code DiagnosticCode) bool {
	switch code {
	case CodeArchiveInvalid, CodeArchiveUnsupported, CodeArchiveEncrypted,
		CodeArchiveUnsafePath, CodeArchiveUnsafeEntry, CodeInspectionLimitExceeded,
		CodeInspectionUnavailable, CodePolicyInternalError:
		return true
	default:
		return false
	}
}

func manifestDigest(manifest Manifest) (string, error) {
	manifest.ManifestDigest = ""
	payload, err := marshalCanonicalStruct(manifest)
	if err != nil {
		return "", fmt.Errorf("canonicalize manifest digest projection: %w", err)
	}
	return digestBytes(payload), nil
}

func digestBytes(payload []byte) string {
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func marshalCanonicalStruct(value any) ([]byte, error) {
	ordinary, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(ordinary))
	decoder.UseNumber()
	var domain any
	if err := decoder.Decode(&domain); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(domain)
}

func canonicalizeManifest(manifest *Manifest, diagnostics []Diagnostic) {
	if manifest.RoleEvidence == nil {
		manifest.RoleEvidence = []Fact{}
	}
	if manifest.Detectors == nil {
		manifest.Detectors = []DetectorIdentity{}
	}
	if manifest.Nodes == nil {
		manifest.Nodes = []ManifestNode{}
	}
	sortFacts(manifest.RoleEvidence)
	sort.Slice(manifest.Detectors, func(left, right int) bool {
		if manifest.Detectors[left].ID != manifest.Detectors[right].ID {
			return manifest.Detectors[left].ID < manifest.Detectors[right].ID
		}
		return manifest.Detectors[left].Version < manifest.Detectors[right].Version
	})
	for index := range manifest.Nodes {
		canonicalizeNode(&manifest.Nodes[index])
	}
	sort.Slice(manifest.Nodes, func(left, right int) bool {
		return manifest.Nodes[left].Path < manifest.Nodes[right].Path
	})
	for index := range diagnostics {
		canonicalizeDiagnostic(&diagnostics[index])
	}
	sort.SliceStable(diagnostics, func(left, right int) bool {
		return diagnosticLess(diagnostics[left], diagnostics[right])
	})
}

func canonicalizeNode(node *ManifestNode) {
	if node.ContainerChain == nil {
		node.ContainerChain = []string{}
	}
	if node.DeclaredUses == nil {
		node.DeclaredUses = []UseEdge{}
	}
	if node.Observations == nil {
		node.Observations = []Observation{}
	}
	sort.Slice(node.DeclaredUses, func(left, right int) bool {
		if node.DeclaredUses[left].Kind != node.DeclaredUses[right].Kind {
			return node.DeclaredUses[left].Kind < node.DeclaredUses[right].Kind
		}
		return node.DeclaredUses[left].Origin < node.DeclaredUses[right].Origin
	})
	for index := range node.Observations {
		if node.Observations[index].Facts == nil {
			node.Observations[index].Facts = []Fact{}
		}
		sortFacts(node.Observations[index].Facts)
	}
	sort.Slice(node.Observations, func(left, right int) bool {
		if node.Observations[left].DetectorID != node.Observations[right].DetectorID {
			return node.Observations[left].DetectorID < node.Observations[right].DetectorID
		}
		return node.Observations[left].Result < node.Observations[right].Result
	})
}

func canonicalizeDiagnostic(diagnostic *Diagnostic) {
	if diagnostic.ContainerChain == nil {
		diagnostic.ContainerChain = []string{}
	}
	if diagnostic.Details == nil {
		diagnostic.Details = []Fact{}
	}
	sortFacts(diagnostic.Details)
}

func diagnosticLess(leftDiagnostic, rightDiagnostic Diagnostic) bool {
	if leftDiagnostic.Path != rightDiagnostic.Path {
		return leftDiagnostic.Path < rightDiagnostic.Path
	}
	leftPriority := diagnosticPriority(leftDiagnostic.Code)
	rightPriority := diagnosticPriority(rightDiagnostic.Code)
	if leftPriority != rightPriority {
		return leftPriority < rightPriority
	}
	if leftDiagnostic.Class != rightDiagnostic.Class {
		return leftDiagnostic.Class < rightDiagnostic.Class
	}
	if leftDiagnostic.Code != rightDiagnostic.Code {
		return leftDiagnostic.Code < rightDiagnostic.Code
	}
	if leftDiagnostic.Reason != rightDiagnostic.Reason {
		return leftDiagnostic.Reason < rightDiagnostic.Reason
	}
	if leftDiagnostic.DetectorID != rightDiagnostic.DetectorID {
		return leftDiagnostic.DetectorID < rightDiagnostic.DetectorID
	}
	if leftDiagnostic.OriginalNameBase64 != rightDiagnostic.OriginalNameBase64 {
		return leftDiagnostic.OriginalNameBase64 < rightDiagnostic.OriginalNameBase64
	}
	if leftDiagnostic.CollisionKey != rightDiagnostic.CollisionKey {
		return leftDiagnostic.CollisionKey < rightDiagnostic.CollisionKey
	}
	if leftDiagnostic.Variant != rightDiagnostic.Variant {
		return leftDiagnostic.Variant < rightDiagnostic.Variant
	}
	if comparison := slices.Compare(leftDiagnostic.ContainerChain, rightDiagnostic.ContainerChain); comparison != 0 {
		return comparison < 0
	}
	if leftDiagnostic.SHA256 != rightDiagnostic.SHA256 {
		return leftDiagnostic.SHA256 < rightDiagnostic.SHA256
	}
	if leftDiagnostic.Size != rightDiagnostic.Size {
		return leftDiagnostic.Size < rightDiagnostic.Size
	}
	if leftDiagnostic.LimitName != rightDiagnostic.LimitName {
		return leftDiagnostic.LimitName < rightDiagnostic.LimitName
	}
	if leftDiagnostic.Limit != rightDiagnostic.Limit {
		return leftDiagnostic.Limit < rightDiagnostic.Limit
	}
	if leftDiagnostic.Observed != rightDiagnostic.Observed {
		return leftDiagnostic.Observed < rightDiagnostic.Observed
	}
	leftIdentity, leftErr := marshalCanonicalStruct(leftDiagnostic)
	rightIdentity, rightErr := marshalCanonicalStruct(rightDiagnostic)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return digestBytes(leftIdentity) < digestBytes(rightIdentity)
}

func sortFacts(facts []Fact) {
	sort.Slice(facts, func(left, right int) bool {
		if facts[left].Key != facts[right].Key {
			return facts[left].Key < facts[right].Key
		}
		return facts[left].Value < facts[right].Value
	})
}

func diagnosticPriority(code DiagnosticCode) int {
	switch code {
	case CodeOriginUnverified:
		return 10
	case CodeInspectionLimitExceeded, CodeInspectionUnavailable:
		// A closed limit or unavailable traversal is an ancestor-level
		// structural inability to continue. Keep it visible even after a
		// bounded flood of member-local findings.
		return 20
	case CodeCompiledDependency:
		return 21
	case CodeTypeAmbiguous, CodeOpaqueDependency,
		CodeArchiveInvalid, CodeArchiveUnsupported, CodeArchiveEncrypted,
		CodeArchiveUnsafePath, CodeArchiveUnsafeEntry:
		return 22
	case CodeToolchainUntrusted, CodeToolchainIdentityChanged:
		return 30
	case CodeGeneratedInputUndeclared:
		return 40
	case CodeBinaryAdmissionUnavailable:
		return 60
	case CodeLocalOutputUnreceipted, CodeLocalOutputDrift:
		return 80
	default:
		return 90
	}
}

func validateManifest(manifest Manifest) error {
	if manifest.SchemaID != SchemaID || manifest.PolicyID != PolicyID || manifest.PolicyVersion != PolicyVersion {
		return fmt.Errorf("unsupported artifact manifest schema or policy")
	}
	if err := validateLimits(manifest.LimitVector); err != nil {
		return err
	}
	if manifest.DetectorRegistryID != DetectorRegistryID {
		return fmt.Errorf("unsupported detector registry %q", manifest.DetectorRegistryID)
	}
	if manifest.AdapterID == "" || manifest.Manager == "" ||
		manifest.PackageName == "" || manifest.PackageVersion == "" {
		return fmt.Errorf("manifest adapter, manager, package name, and package version are required")
	}
	if manifest.Detectors == nil || manifest.RoleEvidence == nil || manifest.Nodes == nil || manifest.Diagnostics == nil {
		return fmt.Errorf("manifest repeated fields must be canonical arrays")
	}
	if !slices.Equal(manifest.Detectors, detectorIdentities()) {
		return fmt.Errorf("manifest detector identities do not match %s", DetectorRegistryID)
	}
	if !sortedFacts(manifest.RoleEvidence) {
		return fmt.Errorf("role evidence is not in canonical order")
	}
	if _, ok := profileIDs[manifest.ProfileID]; !ok {
		return fmt.Errorf("unsupported profile %q", manifest.ProfileID)
	}
	if _, ok := trustRoles[manifest.TrustRole]; !ok {
		return fmt.Errorf("unsupported trust role %q", manifest.TrustRole)
	}
	if _, ok := decisions[manifest.Decision]; !ok || manifest.Decision == DecisionDescend {
		return fmt.Errorf("invalid final decision %q", manifest.Decision)
	}
	if manifest.Decision != DecisionReject && manifest.Decision != allowDecision(manifest.TrustRole) {
		return fmt.Errorf("final decision %q is invalid for trust role %q", manifest.Decision, manifest.TrustRole)
	}
	if !sha256Identity.MatchString(manifest.RawPayload.SHA256) || manifest.RawPayload.Size < 0 {
		return fmt.Errorf("invalid raw payload identity")
	}
	switch manifest.RawPayload.Kind {
	case "file", "canonical_tree", "incomplete":
	default:
		return fmt.Errorf("invalid raw payload kind %q", manifest.RawPayload.Kind)
	}
	if err := validateTraversalAccounting(manifest); err != nil {
		return err
	}
	if err := validateRoleEvidence(manifest); err != nil {
		return err
	}
	if manifest.ManifestDigest != "" && !sha256Identity.MatchString(manifest.ManifestDigest) {
		return fmt.Errorf("invalid manifest digest")
	}
	if manifest.Findings.Algorithm != findingsDigestAlgorithm || manifest.Findings.Evidence == nil ||
		manifest.Findings.Total < 0 || manifest.Findings.Recorded < 0 ||
		manifest.Findings.Recorded > manifest.Findings.Total ||
		manifest.Findings.Recorded != int64(len(manifest.Diagnostics)) ||
		manifest.Findings.Total != int64(len(manifest.Findings.Evidence)) ||
		!sha256Identity.MatchString(manifest.Findings.SHA256) {
		return fmt.Errorf("invalid findings summary")
	}
	wantRecorded := manifest.Findings.Total
	if wantRecorded > manifest.LimitVector.MaxRecordedFindings {
		wantRecorded = manifest.LimitVector.MaxRecordedFindings
	}
	if manifest.Findings.Recorded != wantRecorded {
		return fmt.Errorf("findings are neither fully recorded nor cap-saturated")
	}
	if manifest.Decision == DecisionReject && manifest.Findings.Total == 0 {
		return fmt.Errorf("rejected manifest has no diagnostic findings")
	}
	if manifest.Decision != DecisionReject && manifest.Findings.Total != 0 {
		return fmt.Errorf("admitted manifest has diagnostic findings")
	}
	seenFindingIdentities := make(map[string]struct{}, len(manifest.Findings.Evidence))
	for index, evidence := range manifest.Findings.Evidence {
		if err := validateFindingEvidenceShape(evidence); err != nil {
			return fmt.Errorf("finding evidence %d: %w", index, err)
		}
		if _, duplicate := seenFindingIdentities[evidence.DiagnosticSHA256]; duplicate {
			return fmt.Errorf("finding evidence contains a duplicate diagnostic identity")
		}
		seenFindingIdentities[evidence.DiagnosticSHA256] = struct{}{}
		if index > 0 && findingEvidenceLess(evidence, manifest.Findings.Evidence[index-1]) {
			return fmt.Errorf("finding evidence is not in canonical order")
		}
	}
	wantFindings := summarizeFindingEvidence(manifest.Findings.Evidence, manifest.Findings.Recorded)
	if manifest.Findings.SHA256 != wantFindings.SHA256 {
		return fmt.Errorf("complete finding-set digest mismatch")
	}
	for index, diagnostic := range manifest.Diagnostics {
		evidence, err := findingEvidenceFromDiagnostic(diagnostic)
		if err != nil {
			return fmt.Errorf("canonicalize recorded finding %d: %w", index, err)
		}
		if !findingEvidenceEqual(evidence, manifest.Findings.Evidence[index]) {
			return fmt.Errorf("recorded finding %d is not the canonical evidence prefix", index)
		}
	}
	previousPath := ""
	known := make(map[string]ManifestNode, len(manifest.Nodes))
	knownContainers := make(map[string]struct{}, len(manifest.Nodes))
	directChildren := make(map[string]int, len(manifest.Nodes))
	rootCount := 0
	for index, node := range manifest.Nodes {
		if err := validateNode(node); err != nil {
			return fmt.Errorf("node %d: %w", index, err)
		}
		if err := validateNodeSemantics(manifest, node); err != nil {
			return fmt.Errorf("node %d: %w", index, err)
		}
		if index > 0 && node.Path <= previousPath {
			return fmt.Errorf("manifest nodes are not in unique canonical path order")
		}
		if node.Parent != "" {
			parent, ok := known[node.Parent]
			if !ok {
				return fmt.Errorf("node %q has unknown or later parent %q", node.Path, node.Parent)
			}
			if !containerCapableNode(parent.Kind) {
				return fmt.Errorf("node %q has non-container parent %q", node.Path, node.Parent)
			}
			if !immediateNodeChild(node.Path, parent) {
				return fmt.Errorf("node %q is not an immediate child of parent %q", node.Path, node.Parent)
			}
			expectedChain := append([]string(nil), parent.ContainerChain...)
			if parent.Kind == NodeArchive || parent.Kind == NodeCompressedStream {
				expectedChain = append(expectedChain, parent.Path)
			}
			if !slices.Equal(node.ContainerChain, expectedChain) {
				return fmt.Errorf("node %q container chain is not derived from parent %q", node.Path, node.Parent)
			}
			directChildren[parent.Path]++
		} else {
			rootCount++
			if node.Path != manifest.RawPayload.Path {
				return fmt.Errorf("root node %q does not match raw payload path %q", node.Path, manifest.RawPayload.Path)
			}
			if len(node.ContainerChain) != 0 {
				return fmt.Errorf("root node %q has a nonempty container chain", node.Path)
			}
			if manifest.RawPayload.Kind == "file" &&
				(node.Size != manifest.RawPayload.Size || node.SHA256 != manifest.RawPayload.SHA256) {
				return fmt.Errorf("root node content identity does not match raw payload evidence")
			}
		}
		for _, container := range node.ContainerChain {
			if _, ok := knownContainers[container]; !ok {
				return fmt.Errorf("node %q has unknown or later container %q", node.Path, container)
			}
		}
		known[node.Path] = node
		if node.Kind == NodeArchive || node.Kind == NodeCompressedStream {
			knownContainers[node.Path] = struct{}{}
		}
		previousPath = node.Path
	}
	if len(manifest.Nodes) > 0 && rootCount != 1 {
		return fmt.Errorf("manifest must contain exactly one root node")
	}
	if manifest.RawPayload.Kind == "canonical_tree" {
		identity, totalSize, err := canonicalTreeManifestIdentity(manifest.RawPayload.Path, manifest.Nodes)
		if err != nil {
			return fmt.Errorf("derive canonical tree identity: %w", err)
		}
		if identity != manifest.RawPayload.SHA256 || totalSize != manifest.RawPayload.Size {
			return fmt.Errorf("canonical tree nodes do not match raw payload identity")
		}
	}
	if err := validateTarMetadataBindings(manifest.Nodes, known); err != nil {
		return err
	}
	for _, node := range manifest.Nodes {
		if node.Kind == NodeCompressedStream && directChildren[node.Path] != 1 && node.InspectionComplete {
			return fmt.Errorf("completely inspected compressed stream %q must have exactly one child", node.Path)
		}
	}
	if manifest.Decision != DecisionReject && len(manifest.Nodes) == 0 {
		return fmt.Errorf("admitted manifest has no classified nodes")
	}
	if err := validateDecisionClosure(manifest, known); err != nil {
		return err
	}
	for index, diagnostic := range manifest.Diagnostics {
		if _, ok := diagnosticCodes[diagnostic.Code]; !ok {
			return fmt.Errorf("diagnostic %d has unsupported code %q", index, diagnostic.Code)
		}
		if diagnostic.Class != "" {
			if _, ok := artifactClasses[diagnostic.Class]; !ok {
				return fmt.Errorf("diagnostic %d has unsupported class %q", index, diagnostic.Class)
			}
		}
		if diagnostic.ContainerChain == nil || diagnostic.Details == nil {
			return fmt.Errorf("diagnostic %d repeated fields must be canonical arrays", index)
		}
		if !sortedFacts(diagnostic.Details) {
			return fmt.Errorf("diagnostic %d details are not in canonical order", index)
		}
		if diagnostic.Reason == "" {
			return fmt.Errorf("diagnostic %d has no stable reason", index)
		}
		if diagnostic.DetectorID != "" && !knownDetectorID(diagnostic.DetectorID) {
			return fmt.Errorf("diagnostic %d has unsupported detector %q", index, diagnostic.DetectorID)
		}
		if index > 0 && diagnosticLess(diagnostic, manifest.Diagnostics[index-1]) {
			return fmt.Errorf("diagnostics are not in canonical order")
		}
		if err := validateDiagnosticSemantics(manifest.TrustRole, diagnostic); err != nil {
			return fmt.Errorf("diagnostic %d: %w", index, err)
		}
		if err := validateDiagnosticEvidence(manifest, diagnostic, known); err != nil {
			return fmt.Errorf("diagnostic %d: %w", index, err)
		}
	}
	if err := validateCompleteFindingEvidence(manifest, known); err != nil {
		return err
	}
	for _, node := range manifest.Nodes {
		if node.Decision == DecisionReject && !hasFindingEvidenceForNode(node, manifest.Findings.Evidence) {
			return fmt.Errorf("rejecting node %q has no finding evidence in its subtree", node.Path)
		}
	}
	return nil
}

type tarManifestEntry struct {
	node          ManifestNode
	container     string
	index         uint64
	metadataKind  string
	metadataNames string
	presence      map[string]bool
}

func validateTarMetadataBindings(nodes []ManifestNode, known map[string]ManifestNode) error {
	byContainer := make(map[string]map[uint64]tarManifestEntry)
	metadataReferences := make(map[string]int)
	for _, node := range nodes {
		observation, ok := findObservation(node.Observations, "archive-tar-v1", "ENTRY")
		if !ok {
			continue
		}
		if len(node.ContainerChain) == 0 {
			return fmt.Errorf("tar entry %q has no container ancestry", node.Path)
		}
		container := node.ContainerChain[len(node.ContainerChain)-1]
		metadataNames, ok := singleFactValue(observation.Facts, "metadata_members")
		if !ok {
			return fmt.Errorf("tar entry %q has no metadata-member binding", node.Path)
		}
		metadataKind, _ := singleFactValue(observation.Facts, "metadata_kind")
		index, indexOK := requiredDecimalFact(observation.Facts, "physical_header_index", 1, ^uint64(0)>>1)
		if metadataKind == "" && metadataNames == "" {
			if indexOK {
				return fmt.Errorf("ordinary tar entry %q has an order-dependent physical index", node.Path)
			}
			continue
		}
		if !indexOK {
			return fmt.Errorf("tar metadata binding %q has no physical header index", node.Path)
		}
		entry := tarManifestEntry{
			node: node, container: container, index: index,
			metadataNames: metadataNames, presence: tarPresenceFromFacts(observation.Facts),
		}
		entry.metadataKind = metadataKind
		entries := byContainer[container]
		if entries == nil {
			entries = make(map[uint64]tarManifestEntry)
			byContainer[container] = entries
		}
		if _, duplicate := entries[index]; duplicate {
			return fmt.Errorf("tar container %q repeats physical header index %d", container, index)
		}
		entries[index] = entry
	}

	for container, entries := range byContainer {
		for _, entry := range entries {
			if entry.metadataKind != "" {
				if entry.metadataNames != "" {
					return fmt.Errorf("tar metadata node %q recursively references metadata", entry.node.Path)
				}
				continue
			}
			expectedPresence := make(map[string]bool)
			previousIndex := uint64(0)
			if entry.metadataNames != "" {
				for _, relative := range strings.Split(entry.metadataNames, ",") {
					if relative == "" {
						return fmt.Errorf("tar entry %q has an empty metadata reference", entry.node.Path)
					}
					metadataPath := joinContainerPath(container, relative)
					metadataNode, exists := known[metadataPath]
					if !exists {
						return fmt.Errorf("tar entry %q references unknown metadata %q", entry.node.Path, metadataPath)
					}
					metadataObservation, exists := findObservation(metadataNode.Observations, "archive-tar-v1", "ENTRY")
					kind, kindOK := singleFactValue(metadataObservation.Facts, "metadata_kind")
					metadataIndex, indexOK := requiredDecimalFact(metadataObservation.Facts, "physical_header_index", 1, ^uint64(0)>>1)
					if !exists || !kindOK || kind == "pax-global" || !indexOK || metadataIndex >= entry.index || metadataIndex <= previousIndex {
						return fmt.Errorf("tar entry %q has an invalid or reordered metadata reference %q", entry.node.Path, metadataPath)
					}
					previousIndex = metadataIndex
					metadataReferences[metadataPath]++
					if metadataReferences[metadataPath] > 1 {
						return fmt.Errorf("tar metadata node %q is referenced more than once", metadataPath)
					}
					for key, value := range tarPresenceFromFacts(metadataObservation.Facts) {
						expectedPresence[key] = expectedPresence[key] || value
					}
				}
			}
			for _, key := range tarAssociationPresenceKeys() {
				if entry.presence[key] != expectedPresence[key] {
					return fmt.Errorf("tar entry %q metadata presence %q does not match referenced headers", entry.node.Path, key)
				}
			}
		}
	}

	for _, entries := range byContainer {
		for _, entry := range entries {
			if entry.metadataKind == "" || entry.metadataKind == "pax-global" {
				continue
			}
			if metadataReferences[entry.node.Path] != 1 {
				return fmt.Errorf("local tar metadata node %q is not bound to exactly one logical member", entry.node.Path)
			}
		}
	}
	return nil
}

func tarPresenceFromFacts(input []Fact) map[string]bool {
	result := make(map[string]bool)
	for _, key := range tarAssociationPresenceKeys() {
		value, ok := booleanFact(input, key)
		result[key] = ok && value
	}
	return result
}

func tarAssociationPresenceKeys() []string {
	return []string{
		"atime_present", "ctime_present", "gnu_long_link_present", "gnu_long_name_present",
		"pax_charset_present", "pax_comment_present", "pax_linkpath_present",
		"pax_path_present", "xattr_present",
	}
}

func validateNode(node ManifestNode) error {
	if node.Path == "" || node.CollisionKey == "" {
		return fmt.Errorf("path and collision key are required")
	}
	if err := validateManifestVirtualPath(node.Path); err != nil {
		return fmt.Errorf("invalid canonical path: %w", err)
	}
	if node.CollisionKey != portableCollisionKey(node.Path) {
		return fmt.Errorf("collision key does not match canonical path")
	}
	if node.OriginalNameBase64 == "" {
		return fmt.Errorf("original encoded name is required")
	}
	if node.ContainerChain == nil || node.DeclaredUses == nil || node.Observations == nil {
		return fmt.Errorf("node repeated fields must be canonical arrays")
	}
	if _, err := base64.StdEncoding.DecodeString(node.OriginalNameBase64); err != nil {
		return fmt.Errorf("invalid original encoded name")
	}
	if _, ok := nodeKinds[node.Kind]; !ok {
		return fmt.Errorf("unsupported kind %q", node.Kind)
	}
	if _, ok := artifactClasses[node.Class]; !ok {
		return fmt.Errorf("unsupported class %q", node.Class)
	}
	if _, ok := decisions[node.Decision]; !ok {
		return fmt.Errorf("unsupported decision %q", node.Decision)
	}
	if node.SelectedDetectorID != "" && !knownDetectorID(node.SelectedDetectorID) {
		return fmt.Errorf("unsupported selected detector %q", node.SelectedDetectorID)
	}
	if node.Size < 0 || (node.SHA256 != "" && !sha256Identity.MatchString(node.SHA256)) {
		return fmt.Errorf("invalid content identity")
	}
	if node.InspectionComplete && byteBearingNode(node.Kind) {
		if !sha256Identity.MatchString(node.SHA256) {
			return fmt.Errorf("completely inspected byte-bearing node lacks a content digest")
		}
	}
	for index, use := range node.DeclaredUses {
		if _, ok := useKinds[use.Kind]; !ok || use.Origin == "" {
			return fmt.Errorf("invalid declared use")
		}
		if index > 0 {
			previous := node.DeclaredUses[index-1]
			if use.Kind < previous.Kind || (use.Kind == previous.Kind && use.Origin < previous.Origin) {
				return fmt.Errorf("declared uses are not in canonical order")
			}
		}
	}
	for index, observation := range node.Observations {
		if observation.Facts == nil {
			return fmt.Errorf("observation facts must be a canonical array")
		}
		if !knownDetectorID(observation.DetectorID) {
			return fmt.Errorf("observation has unsupported detector %q", observation.DetectorID)
		}
		if !knownObservationResult(observation.Result) {
			return fmt.Errorf("observation has unsupported result %q", observation.Result)
		}
		if observation.Result == "ERROR" && (node.InspectionComplete || node.Decision != DecisionReject) {
			return fmt.Errorf("detector error appears on a completely inspected or non-rejecting node")
		}
		if !sortedFacts(observation.Facts) {
			return fmt.Errorf("observation facts are not in canonical order")
		}
		if index > 0 {
			previous := node.Observations[index-1]
			if observation.DetectorID < previous.DetectorID ||
				(observation.DetectorID == previous.DetectorID && observation.Result < previous.Result) {
				return fmt.Errorf("observations are not in canonical order")
			}
		}
	}
	if node.SelectedDetectorID != "" && !hasDetectorObservation(node.Observations, node.SelectedDetectorID) {
		return fmt.Errorf("selected detector has no recorded observation")
	}
	if node.InspectionComplete && node.Kind != NodeDirectory && node.Kind != NodeLink && node.Kind != NodeSpecial &&
		node.SelectedDetectorID == "" {
		return fmt.Errorf("completely inspected classifying node has no selected detector")
	}
	return nil
}

func validateTraversalAccounting(manifest Manifest) error {
	accounting := manifest.Accounting
	failure := manifest.TraversalFailure
	values := []int64{
		accounting.RawPayloadBytes, accounting.TotalEmittedBytes,
		accounting.ManifestedEmittedBytes, accounting.UnmanifestedEmittedBytes,
		accounting.ContainerCount, accounting.EntryCount,
		accounting.ManifestedEntryCount, accounting.UnmanifestedEntryCount,
		accounting.MaxObservedArchiveDepth, accounting.MaxObservedLeafBytes,
		accounting.MaxManifestedLeafBytes, accounting.MaxUnmanifestedLeafBytes,
		accounting.MaxStreamInputBytes, accounting.MaxStreamEmittedBytes,
		accounting.MaxStreamExpansionRatio, accounting.AggregateExpansionRatio,
	}
	for _, value := range values {
		if value < 0 {
			return fmt.Errorf("traversal accounting contains a negative value")
		}
	}
	if manifest.RawPayload.Kind != "incomplete" && accounting.RawPayloadBytes != manifest.RawPayload.Size {
		return fmt.Errorf("raw payload accounting does not match the captured payload")
	}
	aggregateRatio := int64(0)
	if accounting.TotalEmittedBytes > 0 {
		aggregateRatio = ceilingRatio(accounting.TotalEmittedBytes, accounting.RawPayloadBytes)
	}
	if accounting.AggregateExpansionRatio != aggregateRatio {
		return fmt.Errorf("aggregate expansion accounting is inconsistent")
	}
	streamRatio := int64(0)
	if accounting.MaxStreamEmittedBytes > 0 {
		streamRatio = ceilingRatio(accounting.MaxStreamEmittedBytes, accounting.MaxStreamInputBytes)
	}
	if accounting.MaxStreamExpansionRatio != streamRatio {
		return fmt.Errorf("stream expansion accounting is inconsistent")
	}
	if manifest.Decision != DecisionReject {
		limits := manifest.LimitVector
		if accounting.RawPayloadBytes > limits.MaxRawPayloadBytes ||
			accounting.TotalEmittedBytes > limits.MaxTotalEmittedBytes ||
			accounting.ContainerCount > limits.MaxContainerCount ||
			accounting.EntryCount > limits.MaxEntryCount ||
			accounting.MaxObservedArchiveDepth > limits.MaxArchiveDepth ||
			accounting.MaxObservedLeafBytes > limits.MaxSingleLeafBytes ||
			accounting.MaxStreamExpansionRatio > limits.MaxExpansionRatio ||
			accounting.AggregateExpansionRatio > limits.MaxExpansionRatio {
			return fmt.Errorf("admitted manifest exceeds the bound limit vector")
		}
	}
	bound, err := bindTraversalAccounting(accounting, manifest.RawPayload.Kind, manifest.Nodes)
	if err != nil {
		return fmt.Errorf("bind traversal evidence: %w", err)
	}
	if bound != accounting {
		return fmt.Errorf("traversal accounting evidence is not exact")
	}
	if manifest.Decision != DecisionReject &&
		(accounting.UnmanifestedEntryCount != 0 || accounting.UnmanifestedEmittedBytes != 0 ||
			accounting.MaxUnmanifestedLeafBytes != 0) {
		return fmt.Errorf("admitted traversal contains unmanifested work")
	}
	if accounting.UnmanifestedEntryCount > 0 &&
		!hasUnmanifestedEntryEvidence(manifest.Diagnostics) {
		return fmt.Errorf("unmanifested entry accounting lacks explicit failure evidence")
	}
	if accounting.UnmanifestedEmittedBytes > 0 &&
		!hasUnmanifestedEmissionEvidence(manifest.Diagnostics) {
		return fmt.Errorf("unmanifested emitted-byte accounting lacks explicit failure evidence")
	}
	if accounting.UnmanifestedEntryCount == 0 && accounting.UnmanifestedEmittedBytes == 0 &&
		accounting.MaxUnmanifestedLeafBytes == 0 {
		if failure != (TraversalFailureEvidence{}) {
			return fmt.Errorf("traversal failure evidence exists without unmanifested work")
		}
	} else {
		if !diagnosticSupportsTraversalFailure(failure.Code) || failure.Path == "" || failure.Reason == "" {
			return fmt.Errorf("unmanifested traversal work lacks a structural failure binding")
		}
		if failure.UnmanifestedEntryCount != accounting.UnmanifestedEntryCount ||
			failure.UnmanifestedEmittedBytes != accounting.UnmanifestedEmittedBytes ||
			failure.MaxUnmanifestedLeafBytes != accounting.MaxUnmanifestedLeafBytes ||
			failure.MaxObservedStreamInput != accounting.MaxStreamInputBytes ||
			failure.MaxObservedStreamEmitted != accounting.MaxStreamEmittedBytes ||
			failure.MaxObservedStreamExpansion != accounting.MaxStreamExpansionRatio {
			return fmt.Errorf("traversal failure evidence does not exactly bind accounting")
		}
		matched := false
		for _, diagnostic := range manifest.Diagnostics {
			if diagnostic.Code == failure.Code && diagnostic.Path == failure.Path && diagnostic.Reason == failure.Reason {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("traversal failure evidence has no exact recorded diagnostic")
		}
	}

	var observedContainers, observedDepth, observedLeafBytes int64
	for _, node := range manifest.Nodes {
		emittedSize, emittedErr := manifestedNodeEmittedSize(node)
		if emittedErr != nil {
			return emittedErr
		}
		if node.Kind == NodeArchive || node.Kind == NodeCompressedStream {
			observedContainers++
			depth := int64(len(node.ContainerChain) + 1)
			if depth > observedDepth {
				observedDepth = depth
			}
		}
		if byteBearingNode(node.Kind) && node.SHA256 != "" &&
			(node.Kind == NodeRegularFile || node.Parent != "") &&
			emittedSize > observedLeafBytes {
			observedLeafBytes = emittedSize
		}
	}
	if accounting.ContainerCount != observedContainers || accounting.MaxObservedArchiveDepth != observedDepth {
		return fmt.Errorf("container traversal accounting does not match manifest nodes")
	}
	if accounting.MaxManifestedLeafBytes != observedLeafBytes {
		return fmt.Errorf("manifested leaf-size accounting is not exact")
	}
	if accounting.MaxUnmanifestedLeafBytes == 0 {
		if accounting.MaxObservedLeafBytes != accounting.MaxManifestedLeafBytes {
			return fmt.Errorf("observed leaf maximum has no manifested or unmanifested evidence")
		}
	} else if accounting.MaxUnmanifestedLeafBytes != accounting.MaxObservedLeafBytes ||
		accounting.MaxUnmanifestedLeafBytes <= accounting.MaxManifestedLeafBytes {
		return fmt.Errorf("unmanifested leaf maximum is not an exact maximum")
	} else if accounting.MaxUnmanifestedLeafBytes > manifest.LimitVector.MaxSingleLeafBytes {
		if !hasLeafLimitEvidence(manifest.Diagnostics, accounting.MaxUnmanifestedLeafBytes) {
			return fmt.Errorf("over-limit unmanifested leaf maximum lacks exact limit evidence")
		}
	} else if accounting.UnmanifestedEmittedBytes < accounting.MaxUnmanifestedLeafBytes &&
		!hasUnmanifestedLeafEvidence(manifest.Diagnostics, accounting.MaxUnmanifestedLeafBytes) {
		return fmt.Errorf("unmanifested leaf maximum lacks emitted-byte evidence")
	}
	streamInput, streamEmitted, streamRatio, streamErr := manifestedStreamMaximum(manifest.Nodes)
	if streamErr != nil {
		return streamErr
	}
	if accounting.UnmanifestedEmittedBytes == 0 &&
		(accounting.MaxStreamInputBytes != streamInput ||
			accounting.MaxStreamEmittedBytes != streamEmitted ||
			accounting.MaxStreamExpansionRatio != streamRatio) {
		return fmt.Errorf("stream expansion accounting does not match manifested entries")
	}
	return nil
}

func hasUnmanifestedEntryEvidence(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		switch diagnostic.Code {
		case CodeArchiveInvalid, CodeArchiveUnsupported, CodeArchiveEncrypted,
			CodeArchiveUnsafePath, CodeArchiveUnsafeEntry,
			CodeInspectionLimitExceeded, CodeInspectionUnavailable,
			CodePolicyInternalError:
			return true
		}
	}
	return false
}

func hasUnmanifestedEmissionEvidence(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		switch diagnostic.Code {
		case CodeArchiveInvalid, CodeArchiveUnsupported, CodeArchiveUnsafePath, CodeArchiveUnsafeEntry,
			CodeInspectionLimitExceeded,
			CodeInspectionUnavailable, CodePolicyInternalError:
			return true
		}
	}
	return false
}

func hasLeafLimitEvidence(diagnostics []Diagnostic, observed int64) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == CodeInspectionLimitExceeded &&
			diagnostic.LimitName == "max_single_leaf_bytes" && diagnostic.Observed == observed {
			return true
		}
	}
	return false
}

func hasUnmanifestedLeafEvidence(diagnostics []Diagnostic, observed int64) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Size != observed {
			continue
		}
		switch diagnostic.Code {
		case CodeArchiveInvalid, CodeArchiveUnsupported, CodeArchiveEncrypted,
			CodeArchiveUnsafePath, CodeArchiveUnsafeEntry,
			CodeInspectionLimitExceeded, CodeInspectionUnavailable:
			return true
		}
	}
	return false
}

func manifestedStreamMaximum(nodes []ManifestNode) (int64, int64, int64, error) {
	byPath := make(map[string]ManifestNode, len(nodes))
	for _, node := range nodes {
		byPath[node.Path] = node
	}
	var maximumInput, maximumEmitted, maximumRatio int64
	for _, node := range nodes {
		if node.Parent == "" || len(node.ContainerChain) == 0 || !byteBearingNode(node.Kind) || node.SHA256 == "" {
			continue
		}
		parent, ok := byPath[node.Parent]
		if !ok {
			return 0, 0, 0, fmt.Errorf("stream accounting node %q has no parent", node.Path)
		}
		compressed, ok, err := compressedInputForNode(node, parent)
		if err != nil {
			return 0, 0, 0, err
		}
		if !ok {
			return 0, 0, 0, fmt.Errorf("stream accounting node %q lacks compressed-size evidence", node.Path)
		}
		emittedSize, err := manifestedNodeEmittedSize(node)
		if err != nil {
			return 0, 0, 0, err
		}
		ratio := int64(0)
		if emittedSize > 0 {
			ratio = ceilingRatio(emittedSize, compressed)
		}
		if ratio > maximumRatio || (ratio == maximumRatio && emittedSize > maximumEmitted) {
			maximumInput = compressed
			maximumEmitted = emittedSize
			maximumRatio = ratio
		}
	}
	return maximumInput, maximumEmitted, maximumRatio, nil
}

func compressedInputForNode(node, parent ManifestNode) (int64, bool, error) {
	if parent.Kind == NodeCompressedStream {
		return parent.Size, true, nil
	}
	candidates := []struct {
		detector string
		key      string
	}{
		{detector: "archive-zip-v1", key: "compressed_size"},
		{detector: "archive-tar-v1", key: "size"},
		{detector: "archive-ar-v1", key: "declared_size"},
	}
	for _, candidate := range candidates {
		for _, observation := range node.Observations {
			if observation.DetectorID != candidate.detector || observation.Result != "ENTRY" {
				continue
			}
			for _, fact := range observation.Facts {
				if fact.Key != candidate.key {
					continue
				}
				value, err := strconv.ParseInt(fact.Value, 10, 64)
				if err != nil || value < 0 {
					return 0, false, fmt.Errorf("node %q has invalid %s.%s evidence", node.Path, candidate.detector, candidate.key)
				}
				return value, true, nil
			}
		}
	}
	return 0, false, nil
}

func validateRoleEvidence(manifest Manifest) error {
	values, err := uniqueFactMap(manifest.RoleEvidence)
	if err != nil {
		return fmt.Errorf("invalid role evidence: %w", err)
	}
	switch manifest.TrustRole {
	case RoleDependencyInput:
		if err := requireFactKeys(values,
			"origin_checksum_sha256", "origin_immutable_id", "origin_locator",
			"origin_lock_record", "origin_verified"); err != nil {
			return fmt.Errorf("dependency role evidence: %w", err)
		}
		if values["origin_checksum_sha256"] != manifest.Origin.ChecksumSHA256 ||
			values["origin_immutable_id"] != manifest.Origin.ImmutableID ||
			values["origin_locator"] != manifest.Origin.Locator ||
			values["origin_lock_record"] != manifest.Origin.LockRecord ||
			values["origin_verified"] != strconv.FormatBool(manifest.Origin.Verified) {
			return fmt.Errorf("dependency role evidence does not match immutable origin evidence")
		}
		if manifest.Decision == DecisionAdmitInput &&
			(!manifest.Origin.Verified || manifest.Origin.Locator == "" ||
				manifest.Origin.ImmutableID == "" || manifest.Origin.LockRecord == "" ||
				manifest.Origin.ChecksumSHA256 != manifest.RawPayload.SHA256) {
			return fmt.Errorf("admitted dependency lacks an exact verified immutable origin")
		}
	case RoleVerifiedBinaryCandidate:
		if err := requireFactKeys(values, "capability", "capability_available"); err != nil {
			return fmt.Errorf("verified-binary role evidence: %w", err)
		}
		if values["capability"] != "verified-binary-v1" || values["capability_available"] != "false" ||
			manifest.Decision != DecisionReject {
			return fmt.Errorf("verified-binary-v1 must remain unavailable")
		}
	case RoleExternalToolchain:
		if len(values) == 0 && manifest.Decision == DecisionReject {
			return requireRoleDiagnostic(manifest, CodeToolchainUntrusted, "toolchain_checkpoint_missing")
		}
		if err := requireFactKeys(values,
			"authorization", "checkpoint_fingerprint", "contained_links_validated",
			"environment_search_resolution", "executable_relative_path", "fingerprint_algorithm",
			"ordinary_nodes_validated", "outside_dependency_closure", "payload_path",
			"payload_sha256", "payload_size", "platform", "policy_selector", "resolved_root",
			"time_of_use_fingerprint", "version"); err != nil {
			return fmt.Errorf("toolchain role evidence: %w", err)
		}
		containedLinks, err := parseRoleBool(values, "contained_links_validated")
		if err != nil {
			return err
		}
		ordinaryNodes, err := parseRoleBool(values, "ordinary_nodes_validated")
		if err != nil {
			return err
		}
		outsideClosure, err := parseRoleBool(values, "outside_dependency_closure")
		if err != nil {
			return err
		}
		payloadSize, sizeErr := strconv.ParseInt(values["payload_size"], 10, 64)
		code, reason := DiagnosticCode(""), ""
		switch {
		case values["authorization"] != "manager-issued-v1":
			code, reason = CodeToolchainUntrusted, "toolchain_checkpoint_not_manager_issued"
		case values["fingerprint_algorithm"] != toolchainFingerprintAlgorithm ||
			!containedLinks || !ordinaryNodes || !outsideClosure ||
			!filepath.IsAbs(values["resolved_root"]) || values["policy_selector"] == "" ||
			values["version"] == "" || values["platform"] == "" || sizeErr != nil || payloadSize < 0 ||
			!sha256Identity.MatchString(values["payload_sha256"]) ||
			!sha256Identity.MatchString(values["checkpoint_fingerprint"]) ||
			!sha256Identity.MatchString(values["time_of_use_fingerprint"]):
			code, reason = CodeToolchainUntrusted, "toolchain_role_evidence_incomplete"
		default:
			validated, pathErr := ValidateVirtualPath(filepath.ToSlash(values["executable_relative_path"]))
			if pathErr != nil || validated.Canonical != filepath.ToSlash(values["executable_relative_path"]) {
				code, reason = CodeToolchainUntrusted, "toolchain_executable_path_invalid"
			} else {
				payloadPath, payloadPathErr := ValidateVirtualPath(values["payload_path"])
				switch {
				case payloadPathErr != nil || payloadPath.Canonical != values["payload_path"]:
					code, reason = CodeToolchainUntrusted, "toolchain_payload_path_invalid"
				default:
					resolved := filepath.Clean(filepath.Join(values["resolved_root"], filepath.FromSlash(validated.Canonical)))
					if filepath.Clean(values["environment_search_resolution"]) != resolved {
						code, reason = CodeToolchainUntrusted, "toolchain_environment_resolution_mismatch"
					} else if values["checkpoint_fingerprint"] != values["time_of_use_fingerprint"] {
						code, reason = CodeToolchainIdentityChanged, "toolchain_fingerprint_changed_before_use"
					} else if values["payload_path"] != manifest.RawPayload.Path ||
						values["executable_relative_path"] != manifest.RawPayload.Path {
						code, reason = CodeToolchainUntrusted, "toolchain_checkpoint_path_replay"
					} else if values["payload_sha256"] != manifest.RawPayload.SHA256 ||
						payloadSize != manifest.RawPayload.Size {
						code, reason = CodeToolchainIdentityChanged, "toolchain_checkpoint_payload_replay"
					}
				}
			}
		}
		if code != "" {
			return requireRoleDiagnostic(manifest, code, reason)
		}
		if err := rejectUnexpectedRoleDiagnostics(manifest, CodeToolchainUntrusted, CodeToolchainIdentityChanged); err != nil {
			return err
		}
		if manifest.Decision == DecisionAllowToolchain {
			if values["payload_path"] != manifest.RawPayload.Path ||
				values["payload_sha256"] != manifest.RawPayload.SHA256 ||
				values["payload_size"] != strconv.FormatInt(manifest.RawPayload.Size, 10) ||
				values["executable_relative_path"] != manifest.RawPayload.Path ||
				values["checkpoint_fingerprint"] != values["time_of_use_fingerprint"] {
				return fmt.Errorf("allowed toolchain evidence is replayed or drifted")
			}
		}
	case RoleLocalBuildOutput:
		if len(values) == 0 && manifest.Decision == DecisionReject {
			return requireRoleDiagnostic(manifest, CodeLocalOutputUnreceipted, "local_output_receipt_missing")
		}
		if err := requireFactKeys(values,
			"artifact_manifest_digest", "authorization", "build_plan_digest", "complete_input_matched",
			"declared_action_id", "execution_receipt", "expectation_class", "expectation_digest",
			"expectation_independently_derived", "expectation_path", "expectation_size",
			"hardlink_source_excluded", "observed_production", "payload_path", "payload_sha256",
			"payload_size", "preexisting_input_excluded", "protected_publication_validated",
			"protected_receipt", "protected_store_identity", "source_closure_digest",
			"staging_root_identity", "staging_started_empty", "write_set_matched"); err != nil {
			return fmt.Errorf("local-output role evidence: %w", err)
		}
		boolValues := make(map[string]bool, 8)
		for _, key := range []string{
			"complete_input_matched", "expectation_independently_derived", "hardlink_source_excluded",
			"observed_production", "preexisting_input_excluded", "protected_publication_validated",
			"staging_started_empty", "write_set_matched",
		} {
			parsed, parseErr := parseRoleBool(values, key)
			if parseErr != nil {
				return parseErr
			}
			boolValues[key] = parsed
		}
		payloadSize, payloadSizeErr := strconv.ParseInt(values["payload_size"], 10, 64)
		expectationSize, expectationSizeErr := strconv.ParseInt(values["expectation_size"], 10, 64)
		expectedClass := ArtifactClass(values["expectation_class"])
		_, classSupported := artifactClasses[expectedClass]
		validIdentities := true
		for _, key := range []string{
			"artifact_manifest_digest", "build_plan_digest", "execution_receipt",
			"expectation_digest", "payload_sha256", "protected_receipt", "source_closure_digest",
		} {
			validIdentities = validIdentities && sha256Identity.MatchString(values[key])
		}
		code, reason := DiagnosticCode(""), ""
		switch {
		case values["authorization"] != "manager-issued-v1":
			code, reason = CodeLocalOutputUnreceipted, "local_output_receipt_not_manager_issued"
		case values["declared_action_id"] == "" || values["protected_store_identity"] == "" ||
			values["staging_root_identity"] == "" || values["payload_path"] == "" ||
			!validIdentities || payloadSizeErr != nil || payloadSize < 0 ||
			!boolValues["staging_started_empty"] || !boolValues["observed_production"] ||
			!boolValues["write_set_matched"] || !boolValues["preexisting_input_excluded"] ||
			!boolValues["hardlink_source_excluded"] || !boolValues["expectation_independently_derived"] ||
			!boolValues["protected_publication_validated"]:
			code, reason = CodeLocalOutputUnreceipted, "local_output_causal_receipt_incomplete"
		case !boolValues["complete_input_matched"]:
			code, reason = CodeLocalOutputDrift, "local_output_complete_input_drift"
		case !classSupported || expectationSizeErr != nil || expectationSize < 0 ||
			values["expectation_path"] != values["payload_path"] ||
			values["expectation_digest"] != values["payload_sha256"] || expectationSize != payloadSize:
			code, reason = CodeLocalOutputDrift, "local_output_expectation_binding_invalid"
		default:
			if _, pathErr := ValidateVirtualPath(values["payload_path"]); pathErr != nil {
				code, reason = CodeLocalOutputDrift, "local_output_path_invalid"
			} else if values["payload_path"] != manifest.RawPayload.Path ||
				values["expectation_path"] != manifest.RawPayload.Path {
				code, reason = CodeLocalOutputDrift, "output_path_replay"
			} else if values["payload_sha256"] != manifest.RawPayload.SHA256 || payloadSize != manifest.RawPayload.Size {
				code, reason = CodeLocalOutputDrift, "output_payload_replay"
			}
		}
		if code != "" {
			return requireRoleDiagnostic(manifest, code, reason)
		}
		if err := rejectUnexpectedRoleDiagnostics(manifest, CodeLocalOutputUnreceipted, CodeLocalOutputDrift); err != nil {
			return err
		}
		if manifest.Decision == DecisionAllowOutput {
			if values["payload_path"] != manifest.RawPayload.Path ||
				values["payload_sha256"] != manifest.RawPayload.SHA256 ||
				values["payload_size"] != strconv.FormatInt(manifest.RawPayload.Size, 10) ||
				values["expectation_path"] != manifest.RawPayload.Path ||
				values["expectation_digest"] != manifest.RawPayload.SHA256 ||
				values["expectation_size"] != strconv.FormatInt(manifest.RawPayload.Size, 10) {
				return fmt.Errorf("allowed output evidence is replayed or drifted")
			}
			root := rootManifestNode(manifest.Nodes)
			if root == nil || root.Class != expectedClass {
				return fmt.Errorf("allowed output does not match its independent class expectation")
			}
		}
	}
	return nil
}

func parseRoleBool(values map[string]string, key string) (bool, error) {
	value, err := strconv.ParseBool(values[key])
	if err != nil || (values[key] != "true" && values[key] != "false") {
		return false, fmt.Errorf("role evidence %q is not a canonical boolean", key)
	}
	return value, nil
}

func requireRoleDiagnostic(manifest Manifest, code DiagnosticCode, reason string) error {
	if manifest.Decision != DecisionReject {
		return fmt.Errorf("invalid role evidence did not reject the manifest")
	}
	for _, diagnostic := range manifest.Diagnostics {
		if diagnostic.Code == code && diagnostic.Reason == reason && diagnostic.Path == manifest.RawPayload.Path {
			return nil
		}
	}
	return fmt.Errorf("role evidence lacks exact %s/%s diagnostic", code, reason)
}

func rejectUnexpectedRoleDiagnostics(manifest Manifest, codes ...DiagnosticCode) error {
	for _, diagnostic := range manifest.Diagnostics {
		if slices.Contains(codes, diagnostic.Code) {
			return fmt.Errorf("role diagnostic %s/%s is not supported by role evidence", diagnostic.Code, diagnostic.Reason)
		}
	}
	return nil
}

func uniqueFactMap(input []Fact) (map[string]string, error) {
	result := make(map[string]string, len(input))
	for _, fact := range input {
		if _, exists := result[fact.Key]; exists {
			return nil, fmt.Errorf("duplicate key %q", fact.Key)
		}
		result[fact.Key] = fact.Value
	}
	return result, nil
}

func requireFactKeys(values map[string]string, keys ...string) error {
	if len(values) != len(keys) {
		return fmt.Errorf("field set has %d keys, want %d", len(values), len(keys))
	}
	for _, key := range keys {
		if _, ok := values[key]; !ok {
			return fmt.Errorf("required field %q is absent", key)
		}
	}
	return nil
}

func validateNodeSemantics(manifest Manifest, node ManifestNode) error {
	if node.Rule == "" {
		return fmt.Errorf("decision rule is required")
	}
	if !node.InspectionComplete && node.Decision != DecisionReject {
		return fmt.Errorf("incompletely inspected node is not rejected")
	}
	if err := validateKindClass(node.Kind, node.Class); err != nil {
		return err
	}
	expected, err := deriveNodeClassification(node)
	if err != nil {
		return fmt.Errorf("derive classification: %w", err)
	}
	if node.Class != expected.class || node.Variant != expected.variant ||
		node.SelectedDetectorID != expected.detectorID {
		return fmt.Errorf(
			"classification (%q, %q, %q) does not match detector evidence (%q, %q, %q)",
			node.Class, node.Variant, node.SelectedDetectorID,
			expected.class, expected.variant, expected.detectorID,
		)
	}
	if err := validateResolvedNodeRule(manifest, node, expected); err != nil {
		return err
	}
	return nil
}

func validateKindClass(kind NodeKind, class ArtifactClass) error {
	valid := false
	switch kind {
	case NodeRegularFile:
		valid = class != ClassArchive && class != ClassCompressedStream && class != ClassDirectory &&
			class != ClassLink && class != ClassSpecial
	case NodeArchive:
		valid = class == ClassArchive || class == ClassNativeLibraryStatic
	case NodeCompressedStream:
		valid = class == ClassCompressedStream
	case NodeDirectory:
		valid = class == ClassDirectory || class == ClassAppleFramework || class == ClassAppleXCFramework
	case NodeLink:
		valid = class == ClassLink
	case NodeSpecial:
		valid = class == ClassSpecial
	}
	if !valid {
		return fmt.Errorf("node kind %q and class %q are incompatible", kind, class)
	}
	return nil
}

func validateDecisionClosure(manifest Manifest, nodes map[string]ManifestNode) error {
	anyRejected := false
	for _, node := range manifest.Nodes {
		if node.Decision != DecisionReject {
			continue
		}
		anyRejected = true
		parent := node.Parent
		for parent != "" {
			ancestor := nodes[parent]
			if ancestor.Decision != DecisionReject {
				return fmt.Errorf("rejecting node %q has admitted ancestor %q", node.Path, parent)
			}
			parent = ancestor.Parent
		}
	}
	if anyRejected && manifest.Decision != DecisionReject {
		return fmt.Errorf("rejecting node has an admitted final decision")
	}
	if len(manifest.Nodes) > 0 && manifest.Decision == DecisionReject && !anyRejected {
		return fmt.Errorf("rejected final decision has no rejecting node")
	}
	return nil
}

func validateDiagnosticSemantics(role TrustRole, diagnostic Diagnostic) error {
	switch diagnostic.Code {
	case CodeOriginUnverified:
		if role != RoleDependencyInput {
			return fmt.Errorf("origin diagnostic is invalid for role %q", role)
		}
	case CodeCompiledDependency:
		if role != RoleDependencyInput || !compiledClass(diagnostic.Class) {
			return fmt.Errorf("compiled-dependency diagnostic has an invalid role or class")
		}
	case CodeBinaryAdmissionUnavailable:
		if role != RoleVerifiedBinaryCandidate {
			return fmt.Errorf("binary-capability diagnostic is invalid for role %q", role)
		}
	case CodeToolchainUntrusted, CodeToolchainIdentityChanged:
		if role != RoleExternalToolchain {
			return fmt.Errorf("toolchain diagnostic is invalid for role %q", role)
		}
	case CodeLocalOutputUnreceipted, CodeLocalOutputDrift:
		if role != RoleLocalBuildOutput {
			return fmt.Errorf("local-output diagnostic is invalid for role %q", role)
		}
	}
	return nil
}

func validateDiagnosticEvidence(
	manifest Manifest,
	diagnostic Diagnostic,
	nodes map[string]ManifestNode,
) error {
	for index, containerPath := range diagnostic.ContainerChain {
		container, ok := nodes[containerPath]
		if !ok || (container.Kind != NodeArchive && container.Kind != NodeCompressedStream) {
			return fmt.Errorf("container chain references non-container %q", containerPath)
		}
		if !slices.Equal(container.ContainerChain, diagnostic.ContainerChain[:index]) {
			return fmt.Errorf("container chain is not an exact ancestry sequence")
		}
	}

	node, exactNode := nodes[diagnostic.Path]
	nodeBound := exactNode && diagnostic.Code != CodeArchiveUnsafePath
	if diagnosticRequiresNode(diagnostic.Code) && !nodeBound {
		return fmt.Errorf("code %q is not bound to a manifest node", diagnostic.Code)
	}
	if nodeBound {
		if node.Decision != DecisionReject {
			return fmt.Errorf("diagnostic path %q does not identify a rejecting node", diagnostic.Path)
		}
		if !slices.Equal(diagnostic.ContainerChain, node.ContainerChain) {
			return fmt.Errorf("diagnostic container chain does not match node %q", node.Path)
		}
		if diagnostic.SHA256 != "" && node.SHA256 != "" && diagnostic.SHA256 != node.SHA256 {
			return fmt.Errorf("diagnostic digest does not match node %q", node.Path)
		}
		if diagnostic.Class != "" && diagnostic.Class != node.Class {
			return fmt.Errorf("diagnostic class does not match node %q", node.Path)
		}
		if diagnostic.Variant != "" && diagnostic.Variant != node.Variant {
			return fmt.Errorf("diagnostic variant does not match node %q", node.Path)
		}
		if diagnostic.DetectorID != "" && !hasDetectorObservation(node.Observations, diagnostic.DetectorID) {
			return fmt.Errorf("diagnostic detector has no node observation")
		}
	} else if len(diagnostic.ContainerChain) > 0 {
		container := nodes[diagnostic.ContainerChain[len(diagnostic.ContainerChain)-1]]
		if container.Decision != DecisionReject {
			return fmt.Errorf("unmanifested diagnostic entry has a non-rejecting container")
		}
		if !strings.HasPrefix(diagnostic.Path, container.Path+"!/") {
			return fmt.Errorf("unmanifested diagnostic path is outside its container chain")
		}
	} else if manifest.RawPayload.Kind == "canonical_tree" {
		root, ok := nodes[manifest.RawPayload.Path]
		if !ok || root.Decision != DecisionReject || !strings.HasPrefix(diagnostic.Path, root.Path+"/") {
			return fmt.Errorf("unmanifested tree diagnostic is outside the rejecting root")
		}
	} else if manifest.RawPayload.Kind != "incomplete" {
		return fmt.Errorf("diagnostic path %q has no manifest or container binding", diagnostic.Path)
	}

	switch diagnostic.Code {
	case CodeOriginUnverified:
		if !nodeBound || diagnostic.Path != manifest.RawPayload.Path ||
			diagnostic.OriginalNameBase64 != node.OriginalNameBase64 ||
			diagnostic.CollisionKey != node.CollisionKey ||
			diagnostic.SHA256 != manifest.RawPayload.SHA256 || diagnostic.Size != manifest.RawPayload.Size {
			return fmt.Errorf("origin diagnostic is not bound to the captured root")
		}
	case CodeCompiledDependency, CodeTypeAmbiguous, CodeOpaqueDependency, CodeGeneratedInputUndeclared:
		if !nodeBound || diagnostic.OriginalNameBase64 != node.OriginalNameBase64 ||
			diagnostic.CollisionKey != node.CollisionKey || diagnostic.Class != node.Class ||
			diagnostic.Variant != node.Variant || diagnostic.SHA256 != node.SHA256 ||
			diagnostic.Size != node.Size || diagnostic.DetectorID == "" ||
			diagnostic.DetectorID != node.SelectedDetectorID ||
			!slices.Equal(diagnostic.Details, observationFacts(node.Observations)) {
			return fmt.Errorf("classification diagnostic is not exact node evidence")
		}
	case CodeBinaryAdmissionUnavailable, CodeToolchainUntrusted, CodeToolchainIdentityChanged,
		CodeLocalOutputUnreceipted, CodeLocalOutputDrift:
		if !nodeBound || diagnostic.Path != manifest.RawPayload.Path ||
			diagnostic.OriginalNameBase64 != node.OriginalNameBase64 ||
			diagnostic.CollisionKey != node.CollisionKey ||
			diagnostic.SHA256 != manifest.RawPayload.SHA256 || diagnostic.Size != manifest.RawPayload.Size {
			return fmt.Errorf("trust-role diagnostic is not bound to the captured root")
		}
	case CodeInspectionLimitExceeded:
		if err := validateLimitDiagnostic(manifest.LimitVector, diagnostic); err != nil {
			return err
		}
	case CodeArchiveUnsafePath:
		if diagnostic.OriginalNameBase64 == "" {
			return fmt.Errorf("unsafe-path diagnostic lacks the original encoded name")
		}
		if diagnostic.LimitName != "" {
			if err := validateLimitDiagnostic(manifest.LimitVector, diagnostic); err != nil {
				return err
			}
		}
	}
	return nil
}

func diagnosticRequiresNode(code DiagnosticCode) bool {
	switch code {
	case CodeOriginUnverified, CodeCompiledDependency, CodeBinaryAdmissionUnavailable,
		CodeTypeAmbiguous, CodeOpaqueDependency, CodeGeneratedInputUndeclared,
		CodeToolchainUntrusted, CodeToolchainIdentityChanged,
		CodeLocalOutputUnreceipted, CodeLocalOutputDrift:
		return true
	default:
		return false
	}
}

func validateLimitDiagnostic(limits LimitVector, diagnostic Diagnostic) error {
	var want int64
	switch diagnostic.LimitName {
	case "max_raw_payload_bytes":
		want = limits.MaxRawPayloadBytes
	case "max_single_leaf_bytes":
		want = limits.MaxSingleLeafBytes
	case "max_total_emitted_bytes":
		want = limits.MaxTotalEmittedBytes
	case "max_archive_depth":
		want = limits.MaxArchiveDepth
	case "max_container_count":
		want = limits.MaxContainerCount
	case "max_entry_count":
		want = limits.MaxEntryCount
	case "max_expansion_ratio":
		want = limits.MaxExpansionRatio
	case "max_path_bytes":
		want = limits.MaxPathBytes
	case "max_component_bytes":
		want = limits.MaxComponentBytes
	case "max_recorded_findings":
		want = limits.MaxRecordedFindings
	default:
		return fmt.Errorf("limit diagnostic has unsupported limit name %q", diagnostic.LimitName)
	}
	if diagnostic.Limit != want || diagnostic.Observed <= want {
		return fmt.Errorf("limit diagnostic does not record the exact exceeded bound")
	}
	return nil
}

func validateFindingEvidenceShape(evidence FindingEvidence) error {
	if !sha256Identity.MatchString(evidence.DiagnosticSHA256) ||
		!sha256Identity.MatchString(evidence.DetailsSHA256) {
		return fmt.Errorf("finding evidence lacks canonical diagnostic identities")
	}
	if _, ok := diagnosticCodes[evidence.Code]; !ok {
		return fmt.Errorf("unsupported diagnostic code %q", evidence.Code)
	}
	if evidence.Path == "" || evidence.Reason == "" || evidence.ContainerChain == nil || evidence.Details == nil {
		return fmt.Errorf("finding evidence lacks its path, reason, or canonical chain")
	}
	if evidence.OriginalNameBase64 != "" {
		if _, err := base64.StdEncoding.DecodeString(evidence.OriginalNameBase64); err != nil {
			return fmt.Errorf("finding evidence has invalid original encoded name")
		}
	}
	if evidence.Class != "" {
		if _, ok := artifactClasses[evidence.Class]; !ok {
			return fmt.Errorf("unsupported artifact class %q", evidence.Class)
		}
	}
	if evidence.DetectorID != "" && !knownDetectorID(evidence.DetectorID) {
		return fmt.Errorf("unsupported detector %q", evidence.DetectorID)
	}
	if evidence.SHA256 != "" && !sha256Identity.MatchString(evidence.SHA256) {
		return fmt.Errorf("invalid finding payload digest")
	}
	if evidence.Size < 0 || evidence.Limit < 0 || evidence.Observed < 0 {
		return fmt.Errorf("negative finding evidence value")
	}
	if !sortedFacts(evidence.Details) {
		return fmt.Errorf("finding evidence details are not canonical")
	}
	detailPayload, err := marshalCanonicalStruct(evidence.Details)
	if err != nil || evidence.DetailsSHA256 != digestBytes(detailPayload) {
		return fmt.Errorf("finding evidence detail digest mismatch")
	}
	diagnosticPayload, err := marshalCanonicalStruct(diagnosticFromFindingEvidence(evidence))
	if err != nil || evidence.DiagnosticSHA256 != digestBytes(diagnosticPayload) {
		return fmt.Errorf("finding evidence diagnostic digest mismatch")
	}
	return nil
}

func findingEvidenceEqual(left, right FindingEvidence) bool {
	return left.DiagnosticSHA256 == right.DiagnosticSHA256 && left.Code == right.Code &&
		left.Path == right.Path && left.OriginalNameBase64 == right.OriginalNameBase64 &&
		left.CollisionKey == right.CollisionKey && left.Class == right.Class &&
		left.Variant == right.Variant && left.DetectorID == right.DetectorID &&
		left.Reason == right.Reason && slices.Equal(left.ContainerChain, right.ContainerChain) &&
		left.SHA256 == right.SHA256 && left.Size == right.Size &&
		left.LimitName == right.LimitName && left.Limit == right.Limit &&
		left.Observed == right.Observed && slices.Equal(left.Details, right.Details) &&
		left.DetailsSHA256 == right.DetailsSHA256
}

func validateCompleteFindingEvidence(manifest Manifest, nodes map[string]ManifestNode) error {
	for index, evidence := range manifest.Findings.Evidence {
		diagnostic := diagnosticFromFindingEvidence(evidence)
		if err := validateDiagnosticSemantics(manifest.TrustRole, diagnostic); err != nil {
			return fmt.Errorf("finding evidence %d: %w", index, err)
		}
		for chainIndex, containerPath := range evidence.ContainerChain {
			container, ok := nodes[containerPath]
			if !ok || (container.Kind != NodeArchive && container.Kind != NodeCompressedStream) {
				return fmt.Errorf("finding evidence %d references non-container %q", index, containerPath)
			}
			if !slices.Equal(container.ContainerChain, evidence.ContainerChain[:chainIndex]) {
				return fmt.Errorf("finding evidence %d has a noncanonical container chain", index)
			}
		}

		node, exactNode := nodes[evidence.Path]
		nodeBound := exactNode && evidence.Code != CodeArchiveUnsafePath
		if diagnosticRequiresNode(evidence.Code) {
			if !nodeBound || node.Decision != DecisionReject {
				return fmt.Errorf("finding evidence %d is not bound to a rejecting node", index)
			}
			switch evidence.Code {
			case CodeCompiledDependency, CodeTypeAmbiguous, CodeOpaqueDependency, CodeGeneratedInputUndeclared:
				if evidence.OriginalNameBase64 != node.OriginalNameBase64 ||
					evidence.CollisionKey != node.CollisionKey || evidence.Class != node.Class ||
					evidence.Variant != node.Variant || evidence.SHA256 != node.SHA256 ||
					evidence.Size != node.Size || evidence.DetectorID != node.SelectedDetectorID ||
					!slices.Equal(evidence.ContainerChain, node.ContainerChain) ||
					!slices.Equal(evidence.Details, observationFacts(node.Observations)) {
					return fmt.Errorf("finding evidence %d does not exactly bind classification node %q", index, node.Path)
				}
				if err := validateClassificationReason(node, evidence.Code, evidence.Reason); err != nil {
					return fmt.Errorf("finding evidence %d: %w", index, err)
				}
			case CodeOriginUnverified, CodeBinaryAdmissionUnavailable,
				CodeToolchainUntrusted, CodeToolchainIdentityChanged,
				CodeLocalOutputUnreceipted, CodeLocalOutputDrift:
				if evidence.Path != manifest.RawPayload.Path ||
					evidence.OriginalNameBase64 != node.OriginalNameBase64 ||
					evidence.CollisionKey != node.CollisionKey || evidence.SHA256 != manifest.RawPayload.SHA256 ||
					evidence.Size != manifest.RawPayload.Size || !slices.Equal(evidence.ContainerChain, node.ContainerChain) {
					return fmt.Errorf("finding evidence %d does not exactly bind trust-role root %q", index, node.Path)
				}
			}
		} else if nodeBound {
			if node.Decision != DecisionReject || !slices.Equal(evidence.ContainerChain, node.ContainerChain) {
				return fmt.Errorf("finding evidence %d is not bound to its rejecting node", index)
			}
			if (evidence.Class != "" && evidence.Class != node.Class) ||
				(evidence.Variant != "" && evidence.Variant != node.Variant) ||
				(evidence.DetectorID != "" && !hasDetectorObservation(node.Observations, evidence.DetectorID)) ||
				(evidence.SHA256 != "" && node.SHA256 != "" && evidence.SHA256 != node.SHA256) {
				return fmt.Errorf("finding evidence %d contradicts rejecting node %q", index, node.Path)
			}
		} else if len(evidence.ContainerChain) > 0 {
			container := nodes[evidence.ContainerChain[len(evidence.ContainerChain)-1]]
			if container.Decision != DecisionReject || !strings.HasPrefix(evidence.Path, container.Path+"!/") {
				return fmt.Errorf("finding evidence %d is outside its rejecting container", index)
			}
		} else if manifest.RawPayload.Kind == "canonical_tree" {
			root, ok := nodes[manifest.RawPayload.Path]
			if !ok || root.Decision != DecisionReject || !strings.HasPrefix(evidence.Path, root.Path+"/") {
				return fmt.Errorf("finding evidence %d is outside its rejecting tree", index)
			}
		} else if manifest.RawPayload.Kind != "incomplete" && !exactNode {
			return fmt.Errorf("finding evidence %d has no manifest or container binding", index)
		}
		if evidence.Code == CodeInspectionLimitExceeded {
			if err := validateLimitDiagnostic(manifest.LimitVector, diagnostic); err != nil {
				return fmt.Errorf("finding evidence %d: %w", index, err)
			}
		}
		if evidence.Code == CodeArchiveUnsafePath && evidence.OriginalNameBase64 == "" {
			return fmt.Errorf("finding evidence %d lacks its unsafe original name", index)
		}
		if evidence.DetailsSHA256 == "" {
			return fmt.Errorf("finding evidence %d lacks a detail projection", index)
		}
	}
	return nil
}

func validateClassificationReason(node ManifestNode, code DiagnosticCode, reason string) error {
	allowed := false
	switch code {
	case CodeCompiledDependency:
		allowed = reason == "class_forbidden_for_trust_role" ||
			(reason == "native_archive_metadata_forbidden" && node.Rule == "native_archive_metadata") ||
			(reason == "apple_bundle_forbidden" &&
				(node.Class == ClassAppleFramework || node.Class == ClassAppleXCFramework)) ||
			(reason == "compiled_match_with_detector_error" && !node.InspectionComplete)
	case CodeTypeAmbiguous:
		switch reason {
		case "deny_indicating_name_with_noncompiled_bytes",
			"resolved_link_or_load_with_noncompiled_bytes",
			"deny_indicating_name_with_container_bytes",
			"deny_indicating_name_with_invalid_structural_candidate":
			allowed = true
		}
	case CodeGeneratedInputUndeclared:
		allowed = reason == "missing_generated_input_lineage"
	case CodeOpaqueDependency:
		// Opaque parser failures retain their closed detector observations as
		// the semantic authority; the human-stable reason is additionally
		// constrained to a nonempty value by the finding schema.
		allowed = node.Class == ClassOpaqueUnknown && reason != ""
	}
	if !allowed {
		return fmt.Errorf("classification reason %q is not valid for code %q", reason, code)
	}
	return nil
}

func hasFindingEvidenceForNode(node ManifestNode, evidence []FindingEvidence) bool {
	for _, finding := range evidence {
		if finding.Path == node.Path || stringsHasNodeParent(finding.Path, node.Path) {
			return true
		}
	}
	return false
}

func rootManifestNode(nodes []ManifestNode) *ManifestNode {
	for index := range nodes {
		if nodes[index].Parent == "" {
			return &nodes[index]
		}
	}
	return nil
}

func stringsHasNodeParent(nodePath, parentPath string) bool {
	return strings.HasPrefix(nodePath, parentPath+"/") || strings.HasPrefix(nodePath, parentPath+"!/")
}

func containerCapableNode(kind NodeKind) bool {
	return kind == NodeArchive || kind == NodeCompressedStream || kind == NodeDirectory
}

func immediateNodeChild(child string, parent ManifestNode) bool {
	separator := "/"
	if parent.Kind == NodeArchive || parent.Kind == NodeCompressedStream {
		separator = "!/"
	}
	prefix := parent.Path + separator
	if !strings.HasPrefix(child, prefix) {
		return false
	}
	relative := strings.TrimPrefix(child, prefix)
	return relative != "" && !strings.Contains(relative, "/") && !strings.Contains(relative, "!/")
}

func sortedFacts(values []Fact) bool {
	for index, value := range values {
		if value.Key == "" {
			return false
		}
		if index == 0 {
			continue
		}
		previous := values[index-1]
		if value.Key < previous.Key || (value.Key == previous.Key && value.Value < previous.Value) {
			return false
		}
	}
	return true
}

func knownDetectorID(value string) bool {
	for _, detector := range detectorIdentities() {
		if detector.ID == value {
			return true
		}
	}
	return false
}

func knownObservationResult(value string) bool {
	switch value {
	case "ENTRY", "ERROR", "MATCH", "NO_MATCH", "NO_PROFILE_MATCH":
		return true
	default:
		return false
	}
}

func hasDetectorObservation(observations []Observation, detectorID string) bool {
	for _, observation := range observations {
		if observation.DetectorID == detectorID {
			return true
		}
	}
	return false
}

func validateLimits(limits LimitVector) error {
	if limits != DefaultLimits() {
		return fmt.Errorf("limit vector is not %s", LimitVectorID)
	}
	return nil
}
