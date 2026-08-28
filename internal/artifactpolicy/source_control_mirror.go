package artifactpolicy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

const sourceControlMirrorTransform = "source-control-mirror-v1"

// SourceControlMirrorAuthorization is an opaque, narrowly scoped local-output
// authorization. It cannot be implemented by adapters and is not a general
// local build-output or verified-binary admission capability.
type SourceControlMirrorAuthorization interface {
	artifactPolicySourceControlMirrorAuthorization() sourceControlMirrorAuthorizationRecord
}

type sourceControlMirrorAuthorizationRecord struct {
	seal                                                          *authorizationSeal
	profileID, kind, origin, revision, gitTree, mirrorPath        string
	mirrorDigest, artifactManifestID, admittedSourceReceiptID     closuregraph.ID
	acquisitionPermitID, acquisitionReceiptID, derivationPermitID closuregraph.ID
	derivationReceiptID                                           closuregraph.ID
}

type sealedSourceControlMirrorAuthorization struct {
	record sourceControlMirrorAuthorizationRecord
}

func (authorization sealedSourceControlMirrorAuthorization) artifactPolicySourceControlMirrorAuthorization() sourceControlMirrorAuthorizationRecord {
	return authorization.record
}

// SourceControlMirrorAuthorizationRequest names the exact issued acquisition,
// admitted source, offline mirror derivation, and resulting mirror inventory.
type SourceControlMirrorAuthorizationRequest struct {
	ProfileID, Kind, Origin, Revision, GitTree, MirrorPath    string
	MirrorDigest, ArtifactManifestID, AdmittedSourceReceiptID closuregraph.ID
	AcquisitionExecutor                                       *closureexec.SourceAcquisitionExecutor
	AcquisitionReceipt                                        closureexec.SourceAcquisitionReceipt
	DerivationExecutor                                        *closureexec.Executor
	DerivationPermit                                          closureexec.DerivationPermit
	DerivationReceipt                                         closureexec.DerivationReceipt
	TransformEvidence                                         []byte
}

// IssueSourceControlMirrorAuthorization verifies the complete causal chain and
// authorizes only its exact same-kind mirror output.
func (service *Service) IssueSourceControlMirrorAuthorization(request SourceControlMirrorAuthorizationRequest) (SourceControlMirrorAuthorization, error) {
	service, err := configuredService(service)
	if err != nil {
		return nil, err
	}
	if request.ProfileID == "" || request.Kind == "" || request.Origin == "" || request.Revision == "" || request.GitTree == "" || request.MirrorPath == "" ||
		!request.MirrorDigest.Valid() || !request.ArtifactManifestID.Valid() || !request.AdmittedSourceReceiptID.Valid() ||
		request.AcquisitionExecutor == nil || request.DerivationExecutor == nil {
		return nil, &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputUnreceipted, Reason: "source_control_mirror_authority_incomplete"}}
	}
	if request.AcquisitionReceipt.CanonicalOrigin != request.Origin || request.AcquisitionReceipt.RequestedRevision != request.Revision ||
		request.AcquisitionReceipt.Observation.ResolvedRevision != request.Revision || request.AcquisitionReceipt.Observation.GitTree != request.GitTree {
		return nil, &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputDrift, Reason: "source_control_acquisition_evidence_drift"}}
	}
	if err = request.AcquisitionExecutor.VerifyIssuedReceipt(request.AcquisitionReceipt); err != nil {
		return nil, &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputUnreceipted, Reason: "source_control_acquisition_receipt_not_issued"}}
	}
	acquisitionID, _ := request.AcquisitionReceipt.ID()
	if request.DerivationPermit.InvocationSubtype != closureexec.DerivationMirror || request.DerivationPermit.Network != "none" ||
		request.DerivationPermit.InvocationKey != sourceControlMirrorTransform+":"+string(acquisitionID) ||
		!containsClosureID(request.DerivationPermit.AdmittedInputReceiptIDs, request.AdmittedSourceReceiptID) ||
		len(request.DerivationPermit.LocalOutputs) != 1 || request.DerivationPermit.LocalOutputs[0].Path != request.MirrorPath || request.DerivationPermit.LocalOutputs[0].SchemaID != sourceControlMirrorTransform {
		return nil, &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputUnreceipted, Reason: "source_control_mirror_derivation_scope_invalid"}}
	}
	if err = request.DerivationExecutor.VerifyIssuedDerivationChain(request.DerivationPermit, request.DerivationReceipt); err != nil {
		return nil, &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputUnreceipted, Reason: "source_control_mirror_derivation_not_issued"}}
	}
	if len(request.DerivationReceipt.Outputs) != 1 {
		return nil, &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputDrift, Reason: "source_control_mirror_output_cardinality"}}
	}
	output := request.DerivationReceipt.Outputs[0]
	evidenceSum := sha256.Sum256(request.TransformEvidence)
	if output.SHA256 != closuregraph.ID("sha256:"+hex.EncodeToString(evidenceSum[:])) || output.ArtifactManifestID != request.ArtifactManifestID || output.SchemaID != sourceControlMirrorTransform {
		return nil, &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputDrift, Reason: "source_control_mirror_output_evidence_drift"}}
	}
	var evidence struct {
		SchemaID             string `json:"schema_id"`
		MirrorDigest         string `json:"mirror_digest"`
		Revision             string `json:"revision"`
		GitTree              string `json:"git_tree"`
		Kind                 string `json:"kind"`
		AcquisitionReceiptID string `json:"acquisition_receipt_id"`
	}
	if err = json.Unmarshal(request.TransformEvidence, &evidence); err != nil || evidence.SchemaID != sourceControlMirrorTransform || evidence.MirrorDigest != string(request.MirrorDigest) || evidence.Revision != request.Revision || evidence.GitTree != strings.ToLower(request.GitTree) || evidence.Kind != request.Kind || evidence.AcquisitionReceiptID != string(acquisitionID) {
		return nil, &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputDrift, Reason: "source_control_mirror_evidence_payload_drift"}}
	}
	permitID, _ := request.DerivationPermit.ID()
	receiptID, _ := request.DerivationReceipt.ID()
	record := sourceControlMirrorAuthorizationRecord{seal: service.authorizationSeal, profileID: request.ProfileID, kind: request.Kind, origin: request.Origin, revision: request.Revision, gitTree: strings.ToLower(request.GitTree), mirrorPath: request.MirrorPath, mirrorDigest: request.MirrorDigest, artifactManifestID: request.ArtifactManifestID, admittedSourceReceiptID: request.AdmittedSourceReceiptID, acquisitionPermitID: request.AcquisitionReceipt.PermitID, acquisitionReceiptID: acquisitionID, derivationPermitID: permitID, derivationReceiptID: receiptID}
	return sealedSourceControlMirrorAuthorization{record: record}, nil
}

// ValidateSourceControlMirrorAuthorization rechecks an authorization against
// the exact captured mirror tree immediately before replay.
func (service *Service) ValidateSourceControlMirrorAuthorization(authorization SourceControlMirrorAuthorization, profileID, kind, origin, revision, gitTree, mirrorPath string, mirrorDigest, artifactManifestID, admittedSourceReceiptID closuregraph.ID) error {
	service, err := configuredService(service)
	if err != nil {
		return err
	}
	if authorization == nil {
		return &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputUnreceipted, Reason: "source_control_mirror_authorization_missing"}}
	}
	record := authorization.artifactPolicySourceControlMirrorAuthorization()
	expected := sourceControlMirrorAuthorizationRecord{seal: service.authorizationSeal, profileID: profileID, kind: kind, origin: origin, revision: revision, gitTree: strings.ToLower(gitTree), mirrorPath: mirrorPath, mirrorDigest: mirrorDigest, artifactManifestID: artifactManifestID, admittedSourceReceiptID: admittedSourceReceiptID, acquisitionPermitID: record.acquisitionPermitID, acquisitionReceiptID: record.acquisitionReceiptID, derivationPermitID: record.derivationPermitID, derivationReceiptID: record.derivationReceiptID}
	if record.seal != service.authorizationSeal {
		return &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputUnreceipted, Reason: "source_control_mirror_authorization_not_manager_issued"}}
	}
	if !reflect.DeepEqual(record, expected) {
		return &PolicyError{Primary: Diagnostic{Code: CodeLocalOutputDrift, Reason: "source_control_mirror_authorization_drift"}}
	}
	return nil
}

func containsClosureID(values []closuregraph.ID, expected closuregraph.ID) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func (record sourceControlMirrorAuthorizationRecord) String() string {
	return fmt.Sprintf("%s:%s@%s", record.profileID, record.kind, record.revision)
}
