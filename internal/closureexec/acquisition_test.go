package closureexec

import (
	"context"
	"reflect"
	"testing"

	"github.com/relux-works/curator/internal/closuregraph"
)

type acquisitionRunnerFixture struct {
	starts int
	result AcquisitionRunResult
}

func (runner *acquisitionRunnerFixture) RunSourceAcquisition(context.Context, SourceAcquisitionPermit) (AcquisitionRunResult, error) {
	runner.starts++
	return runner.result, nil
}

func acquisitionPermitFixture(head string) SourceAcquisitionPermit {
	limits := ResourceLimits{OutputBytes: 1024, ReadBytes: 1024, WriteBytes: 4096, WallTimeMillis: 1000, ProcessCount: 1}
	limitID, _ := limits.ID()
	evidence := []EvidenceRequirement{{Path: "quarantine/evidence.json", SchemaID: "source-control-object-evidence-v1", ArtifactManifestID: xid('7')}}
	evidenceID, _ := evidenceSchemaID(evidence)
	return SourceAcquisitionPermit{
		SchemaID: SchemaSourceAcquisitionPermit, PreviousCausalHead: head,
		SourceProfileID: "swiftpm-source-v1", CanonicalOrigin: "https://example.invalid/pkg.git", RequestedRevision: "0123456789012345678901234567890123456789",
		C0CheckpointID: xid('1'), ToolchainNodeID: xid('2'), ToolchainFingerprint: xid('3'), ExecutableSHA256: xid('4'),
		Executable: "bin/git", Argv: []string{"clone", "--mirror", "--", "https://example.invalid/pkg.git", "quarantine/repo.git"}, CWD: "work",
		Environment: map[string]string{"HOME": "empty/home"}, HostID: xid('5'), TargetID: xid('5'), AllowedProcesses: []string{"bin/git"},
		ReadRoots: []string{"toolchain"}, QuarantineWriteRoots: []string{"quarantine"}, NetworkPolicy: "exact-origin-only", ExpectedEvidence: evidence,
		ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID, RecheckRule: "immediate-exact-v1",
	}
}

func TestSourceAcquisitionCanonicalRoundTripAndPortableEvidence(t *testing.T) {
	head := string(xid('0'))
	permit := acquisitionPermitFixture(head)
	runner := &acquisitionRunnerFixture{result: AcquisitionRunResult{ExitCode: 0, Evidence: []DerivationOutput{{Path: permit.ExpectedEvidence[0].Path, SchemaID: permit.ExpectedEvidence[0].SchemaID, ArtifactManifestID: permit.ExpectedEvidence[0].ArtifactManifestID, SHA256: xid('8'), Size: 12}}, ResolvedRevision: permit.RequestedRevision, GitTree: "abcdefabcdefabcdefabcdefabcdefabcdefabcd", ObjectIDs: []string{"abcdefabcdefabcdefabcdefabcdefabcdefabcd"}}}
	executor, err := NewSourceAcquisitionExecutor(AssuranceConfig{Mode: AssurancePortable}, runner, nil, head)
	if err != nil {
		t.Fatal(err)
	}
	permitID, err := executor.Commit(t.Context(), permit)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := executor.Execute(t.Context(), permitID, func(context.Context) (ToolchainIdentity, error) {
		return ToolchainIdentity{Fingerprint: permit.ToolchainFingerprint, ExecutableSHA256: permit.ExecutableSHA256}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if runner.starts != 1 || receipt.Observation.Network != "not-observed" || !reflect.DeepEqual(receipt.ActualCapabilities, portableCapabilities) {
		t.Fatalf("portable acquisition evidence = starts:%d network:%q capabilities:%+v", runner.starts, receipt.Observation.Network, receipt.ActualCapabilities)
	}
	if err = executor.VerifyIssuedReceipt(receipt); err != nil {
		t.Fatal(err)
	}
	permitBytes, err := permitWithPortableBinding(permit).CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = DecodeSourceAcquisitionPermit(permitBytes); err != nil {
		t.Fatal(err)
	}
	receiptBytes, err := receipt.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeSourceAcquisitionReceipt(receiptBytes)
	if err != nil || !reflect.DeepEqual(decoded, receipt) {
		t.Fatalf("receipt round trip: decoded=%+v err=%v", decoded, err)
	}
}

func TestSourceAcquisitionRejectsStaleAndToolDriftBeforeStart(t *testing.T) {
	head := string(xid('0'))
	runner := &acquisitionRunnerFixture{}
	executor, err := NewSourceAcquisitionExecutor(AssuranceConfig{Mode: AssurancePortable}, runner, nil, head)
	if err != nil {
		t.Fatal(err)
	}
	stale := acquisitionPermitFixture(string(xid('9')))
	if _, err = executor.Commit(t.Context(), stale); err == nil || runner.starts != 0 {
		t.Fatalf("stale permit err=%v starts=%d", err, runner.starts)
	}
	permit := acquisitionPermitFixture(head)
	permitID, err := executor.Commit(t.Context(), permit)
	if err != nil {
		t.Fatal(err)
	}
	_, err = executor.Execute(t.Context(), permitID, func(context.Context) (ToolchainIdentity, error) {
		return ToolchainIdentity{Fingerprint: xid('f'), ExecutableSHA256: permit.ExecutableSHA256}, nil
	})
	if err == nil || runner.starts != 0 {
		t.Fatalf("tool drift err=%v starts=%d", err, runner.starts)
	}
}

func permitWithPortableBinding(permit SourceAcquisitionPermit) SourceAcquisitionPermit {
	permit.AssuranceMode, permit.PolicyID, permit.ExecutionPolicyID = AssurancePortable, PortablePolicyID, PortableExecutionPolicyID
	permit.ActualCapabilities = append([]CapabilityEvidence(nil), portableCapabilities...)
	return permit
}

func TestVerifiedSourceAcquisitionHasNoPortableFallback(t *testing.T) {
	_, err := NewSourceAcquisitionExecutor(AssuranceConfig{Mode: AssuranceVerified, ProviderID: "provider", ProviderVersion: "1", ProviderBinarySHA256: closuregraph.ID("sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), ProviderTrustEvidence: "signed"}, &acquisitionRunnerFixture{}, nil, string(xid('0')))
	if err == nil {
		t.Fatal("verified acquisition accepted a missing provider")
	}
}
