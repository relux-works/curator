// Package closuregraph defines Curator's language-neutral source-closure
// graph, selection overlay, deterministic build projection, and checkpoint
// evidence records. It owns portable schemas and identities, not acquisition,
// artifact byte detection, or sandbox execution.
package closuregraph

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/relux-works/curator/internal/protocoljson"
)

// ID is a domain-separated SHA-256 identity encoded as sha256:<lower-hex>.
type ID string

const (
	// LabelNode and the related labels domain-separate every portable record
	// identity owned by this package.
	LabelNode = "curator-node-v1"
	// LabelEdge domain-separates typed edge identities.
	LabelEdge = "curator-edge-v1"
	// LabelCaptureGraph domain-separates selection-neutral capture identities.
	LabelCaptureGraph = "curator-capture-graph-v1"
	// LabelSelectionContext domain-separates requested-selection identities.
	LabelSelectionContext = "curator-selection-context-v1"
	// LabelSelectionBinding domain-separates concrete binding identities.
	LabelSelectionBinding = "curator-selection-binding-v1"
	// LabelActiveGraph domain-separates selected active-graph identities.
	LabelActiveGraph = "curator-active-graph-v1"
	// LabelBuildPlan domain-separates deterministic build-plan identities.
	LabelBuildPlan = "curator-build-plan-v1"
	// LabelOrderingEdge domain-separates derived action-ordering identities.
	LabelOrderingEdge = "curator-ordering-edge-v1"
	// LabelNonOrderingSCC domain-separates retained non-ordering cycle identities.
	LabelNonOrderingSCC = "curator-non-ordering-scc-v1"
	// LabelBuildCycle domain-separates rejected build-cycle identities.
	LabelBuildCycle = "curator-build-cycle-v1"
	// LabelToolchainSelector domain-separates C4 build-tool selector records.
	LabelToolchainSelector = "curator-toolchain-selector-v1"
	// LabelCheckpoint domain-separates C0-C7 checkpoint identities.
	LabelCheckpoint = "curator-checkpoint-v1"
	// LabelIntakeAdmissionReceipt domain-separates intake admission receipts.
	LabelIntakeAdmissionReceipt = "curator-intake-admission-receipt-v1"
	// LabelDerivationPermit domain-separates pre-C5 derivation permits.
	LabelDerivationPermit = "curator-derivation-permit-v1"
	// LabelDerivationReceipt domain-separates pre-C5 derivation receipts.
	LabelDerivationReceipt = "curator-derivation-receipt-v1"
	// LabelSourceClosure domain-separates C5-bound closure identities.
	LabelSourceClosure = "curator-source-closure-v1"
	// LabelExpectedCacheInput domain-separates expected cache-input identities.
	LabelExpectedCacheInput = "curator-expected-cache-input-v1"
	// LabelProducedArtifactObservation domain-separates C6 output observations.
	LabelProducedArtifactObservation = "curator-produced-artifact-observation-v1"
	// LabelExecutionReceipt domain-separates protected execution receipts.
	LabelExecutionReceipt = "curator-execution-receipt-v1"
	// LabelPublicationReceipt domain-separates protected publication receipts.
	LabelPublicationReceipt = "curator-publication-receipt-v1"
)

const (
	// SchemaCaptureGraph and the related schema IDs version the graph,
	// projection, checkpoint, and receipt wire formats.
	SchemaCaptureGraph = "closure-capture-graph-v1"
	// SchemaSelectionContext identifies the requested-selection wire schema.
	SchemaSelectionContext = "closure-selection-context-v1"
	// SchemaSelectionBinding identifies the concrete-binding wire schema.
	SchemaSelectionBinding = "closure-selection-binding-v1"
	// SchemaActiveGraph identifies the active-projection wire schema.
	SchemaActiveGraph = "closure-active-graph-v1"
	// SchemaBuildPlan identifies the deterministic build-plan wire schema.
	SchemaBuildPlan = "closure-build-plan-v1"
	// SchemaCheckpoint identifies the C0-C7 envelope wire schema.
	SchemaCheckpoint = "closure-checkpoint-v1"
	// SchemaSourceClosure identifies the C5-bound closure wire schema.
	SchemaSourceClosure = "curator-source-closure-v1"
	// SchemaExpectedCacheInput identifies the expected-cache-input wire schema.
	SchemaExpectedCacheInput = "closure-expected-cache-input-v1"
	// SchemaExecutionReceipt identifies the C6 execution-receipt wire schema.
	SchemaExecutionReceipt = "closure-execution-receipt-v1"
	// SchemaPublicationReceipt identifies the C7 publication-receipt wire schema.
	SchemaPublicationReceipt = "closure-publication-receipt-v1"
)

var idPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// Valid reports whether id has the canonical SHA-256 identity shape.
func (id ID) Valid() bool { return idPattern.MatchString(string(id)) }

// DomainID derives ID(label, payload) from a JSON-domain value using strict
// CCJ-1 and SHA-256(label || NUL || canonical payload).
func DomainID(label string, payload any) (ID, error) {
	canonical, err := protocoljson.MarshalCanonical(payload)
	if err != nil {
		return "", fmt.Errorf("canonicalize %s payload: %w", label, err)
	}
	return IDFromCanonical(label, canonical)
}

// IDFromCanonical derives a domain-separated identity from exact canonical
// CCJ-1 bytes. Noncanonical payload bytes are rejected rather than normalized.
func IDFromCanonical(label string, payload []byte) (ID, error) {
	if label == "" || strings.ContainsRune(label, '\x00') {
		return "", fmt.Errorf("identity label must be non-empty and contain no NUL")
	}
	if err := protocoljson.RequireCanonical(payload); err != nil {
		return "", fmt.Errorf("identity payload: %w", err)
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte(label))
	_, _ = digest.Write([]byte{0})
	_, _ = digest.Write(payload)
	return ID("sha256:" + hex.EncodeToString(digest.Sum(nil))), nil
}

type canonicalRecord interface {
	Validate() error
	canonicalValue() map[string]any
	domainLabel() string
}

func canonicalBytes(record canonicalRecord) ([]byte, error) {
	if err := record.Validate(); err != nil {
		return nil, err
	}
	return protocoljson.MarshalCanonical(record.canonicalValue())
}

func recordID(record canonicalRecord) (ID, error) {
	payload, err := canonicalBytes(record)
	if err != nil {
		return "", err
	}
	return IDFromCanonical(record.domainLabel(), payload)
}
