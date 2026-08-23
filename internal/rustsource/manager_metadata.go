package rustsource

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

type metadataExecution struct {
	permitID closuregraph.ID
	inputs   map[closuregraph.ID]closureexec.AdmittedInput
	output   string
}

type managerMetadataRunner struct {
	manager    *managerState
	workspace  string
	executions map[string]metadataExecution
}

func (runner *managerMetadataRunner) CommitMetadata(ctx context.Context, invocation metadataInvocation) (permit, error) {
	if runner == nil || runner.manager == nil || runner.manager.workspaceID == "" {
		return permit{}, fail(CodeGraphIncomplete, "manager metadata authority is incomplete", nil)
	}
	if runner.executions == nil {
		runner.executions = map[string]metadataExecution{}
	}
	output := filepath.Join(runner.manager.session, "metadata-output-"+invocation.View)
	if err := os.RemoveAll(output); err != nil {
		return permit{}, err
	}
	runner.manager.processRunner.OutputRoot = output
	configRoot := filepath.Join(runner.manager.session, "metadata-home-"+invocation.View)
	if err := os.RemoveAll(configRoot); err != nil {
		return permit{}, err
	}
	if err := os.Mkdir(configRoot, 0o700); err != nil {
		return permit{}, err
	}
	configBytes, err := DeriveSourceReplacementConfig(filepath.Join(runner.manager.execRoot, "vendor"), runner.manager.lock.Packages)
	if err != nil {
		return permit{}, err
	}
	if err = os.WriteFile(filepath.Join(configRoot, "config.toml"), configBytes, 0o600); err != nil {
		return permit{}, err
	}
	helper := &managerVendorRunner{manager: runner.manager}
	homeInput, homeID, err := helper.admitTree(ctx, "rust-cargo-home-v1", configRoot)
	if err != nil {
		return permit{}, err
	}
	vendorInput, vendorID, err := helper.admitTree(ctx, "rust-vendor-tree-v1", runner.manager.vendor)
	if err != nil {
		return permit{}, err
	}
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{runner.manager.workspaceID: runner.manager.workspaceInput, homeID: homeInput, vendorID: vendorInput}
	paths := map[closuregraph.ID]string{runner.manager.workspaceID: "workspace", homeID: "cargo-home-seed", vendorID: "vendor"}
	ids := []closuregraph.ID{runner.manager.workspaceID, homeID, vendorID}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	mounts := make([]closureexec.InputMount, len(ids))
	reads := make([]string, len(ids))
	for index, id := range ids {
		mounts[index] = closureexec.InputMount{ReceiptID: id, Path: paths[id]}
		reads[index] = paths[id]
	}
	sort.Strings(reads)
	artifactID, err := closuregraph.DomainID("rust-metadata-output-v1", map[string]any{"capture": invocation.CaptureID, "target": invocation.Target, "view": invocation.View})
	if err != nil {
		return permit{}, err
	}
	requirement := closureexec.EvidenceRequirement{Path: "metadata.json", SchemaID: "cargo-metadata-format-1", ArtifactManifestID: artifactID}
	evidenceID, err := evidenceRequirementsID([]closureexec.EvidenceRequirement{requirement})
	if err != nil {
		return permit{}, err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 64 << 20, ReadBytes: 1 << 30, WriteBytes: 128 << 20, WallTimeMillis: 120_000, ProcessCount: 4}
	limitID, err := limits.ID()
	if err != nil {
		return permit{}, err
	}
	argv := metadataArgvForReplay(invocation, filepath.Join(runner.manager.execRoot, "cargo-home", "config.toml"), filepath.Join(runner.manager.execRoot, "workspace", "Cargo.toml"))
	c0, _ := closuregraph.DomainID("rust-c0-cargo-v1", map[string]any{"fingerprint": invocation.Toolchain.Fingerprint, "transform": TransformID})
	tool, _ := closuregraph.DomainID("rust-cargo-tool-v1", map[string]any{"executable_sha256": invocation.Toolchain.BinarySHA256})
	host, _ := closuregraph.DomainID("rust-native-host-v1", map[string]any{"native": true})
	record := closureexec.DerivationPermit{
		SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: runner.manager.causalHead,
		InvocationKey: "cargo-metadata-v1:" + invocation.View + ":" + invocation.CaptureID, InvocationSubtype: closureexec.DerivationMetadata,
		AdmittedInputReceiptIDs: ids, InputMounts: mounts, WorkCopies: []closureexec.WorkCopy{{ReceiptID: homeID, Path: "cargo-home"}}, C0CheckpointID: c0, ToolchainNodeID: tool,
		ToolchainFingerprint: closuregraph.ID(invocation.Toolchain.Fingerprint), ExecutableSHA256: closuregraph.ID(invocation.Toolchain.BinarySHA256),
		Executable: "bin/cargo", CWD: "workspace", Argv: argv,
		Environment: map[string]string{"CARGO_HOME": filepath.Join(runner.manager.execRoot, "cargo-home"), "CARGO_NET_OFFLINE": "true", "CURATOR_OUTPUT_ROOT": output, "HOME": filepath.Join(runner.manager.execRoot, "home"), "LANG": "C", "LC_ALL": "C", "RUSTC": filepath.Join(runner.manager.cargoRegistry.root, "bin", "rustc"), "TZ": "UTC"},
		HostID:      host, TargetID: host, AllowedProcesses: []string{}, ReadRoots: reads, WriteRoots: []string{"cargo-home", "metadata.json"},
		ExpectedEvidence: []closureexec.EvidenceRequirement{requirement}, StdoutEvidencePath: "metadata.json",
		Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID,
	}
	permitID, err := runner.manager.executor.Commit(record)
	if err != nil {
		return permit{}, err
	}
	legacyID, err := invocation.ID()
	if err != nil {
		return permit{}, err
	}
	runner.executions[invocation.View] = metadataExecution{permitID: permitID, inputs: inputs, output: output}
	return permit{ID: string(permitID), InvocationID: legacyID}, nil
}

func (runner *managerMetadataRunner) RunMetadata(ctx context.Context, authority permit, invocation metadataInvocation, recheck func() error) ([]byte, string, error) {
	execution, ok := runner.executions[invocation.View]
	if !ok || authority.ID != string(execution.permitID) {
		return nil, "", fail(CodeGraphIncomplete, "metadata permit is absent or foreign", nil)
	}
	if err := recheck(); err != nil {
		return nil, "", err
	}
	runner.manager.processRunner.OutputRoot = execution.output
	receipt, err := runner.manager.executor.Execute(ctx, execution.permitID, func(ctx context.Context) (closureexec.ToolchainIdentity, error) {
		tool, checkErr := runner.manager.recheckCargo(ctx)
		if checkErr != nil {
			return closureexec.ToolchainIdentity{}, checkErr
		}
		return closureexec.ToolchainIdentity{Fingerprint: closuregraph.ID(tool.Fingerprint), ExecutableSHA256: closuregraph.ID(tool.BinarySHA256)}, nil
	}, execution.inputs)
	if err != nil {
		return nil, "", err
	}
	if err = runner.manager.executor.VerifyIssuedDerivationReceipt(receipt); err != nil {
		return nil, "", err
	}
	runner.manager.causalHead = string(receipt.NextCausalHead)
	payload, err := os.ReadFile(filepath.Join(execution.output, "metadata.json")) // #nosec G304 -- exact typed executor output.
	if err != nil {
		return nil, "", err
	}
	receiptID, err := receipt.ID()
	return payload, string(receiptID), err
}

func metadataArgvForReplay(invocation metadataInvocation, config, manifest string) []string {
	result := append([]string(nil), invocation.Argv...)
	for index := range result {
		if result[index] == invocation.ConfigPath {
			result[index] = config
		}
		if result[index] == invocation.ManifestPath {
			result[index] = manifest
		}
	}
	return result
}
