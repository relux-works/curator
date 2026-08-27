package rustsource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/privatedir"
)

type cargoBuildExecution struct {
	receipt    closureexec.DerivationReceipt
	outputRoot string
	stdoutPath string
	executable string
	cleanup    func()
}

func executionFromIssuedCargoReceipts(metadata, build closureexec.DerivationReceipt, closureID closuregraph.ID, actionIDs, observationIDs []closuregraph.ID, writeSet []string) (closuregraph.ExecutionReceipt, error) {
	for _, receipt := range []closureexec.DerivationReceipt{metadata, build} {
		if err := receipt.Validate(); err != nil {
			return closuregraph.ExecutionReceipt{}, err
		}
		if receipt.BeforeFingerprint != receipt.AfterFingerprint {
			return closuregraph.ExecutionReceipt{}, fail(CodeOfflineRebuildFailed, "issued Cargo receipt lacks exact enforced execution evidence", nil)
		}
		switch receipt.AssuranceMode {
		case closureexec.AssuranceVerified:
			if receipt.Audit.Network != "none" || receipt.Audit.Processes == nil || receipt.Audit.Reads == nil || receipt.Audit.Writes == nil {
				return closuregraph.ExecutionReceipt{}, fail(CodeOfflineRebuildFailed, "verified Cargo receipt lacks lossless execution evidence", nil)
			}
		case closureexec.AssurancePortable:
			if receipt.Audit.Network != "not-observed" || len(receipt.Audit.Processes) != 0 || len(receipt.Audit.Reads) != 0 || len(receipt.Audit.Writes) != 0 {
				return closuregraph.ExecutionReceipt{}, fail(CodeOfflineRebuildFailed, "portable Cargo receipt inflates execution assurance", nil)
			}
		default:
			return closuregraph.ExecutionReceipt{}, fail(CodeOfflineRebuildFailed, "Cargo receipt assurance mode is unsupported", nil)
		}
	}
	return closuregraph.ExecutionReceipt{
		SchemaID: closuregraph.SchemaExecutionReceipt, ActionOrder: append([]closuregraph.ID(nil), actionIDs...),
		// The compact C6 record retains the committed offline policy. Detailed
		// portable evidence remains explicitly network=not-observed in the issued
		// derivation receipts and is never promoted to lossless observation.
		ClosureID: closureID, Decision: "success", Network: "none",
		ProducedObservationIDs: append([]closuregraph.ID(nil), observationIDs...), ToolchainRechecks: "match", WriteSet: append([]string(nil), writeSet...),
	}, nil
}

func verifyIssuedExecutable(receipt closureexec.DerivationReceipt, logical string, payload []byte) error {
	for _, output := range receipt.Outputs {
		if output.Path != logical {
			continue
		}
		sum := sha256.Sum256(payload)
		if output.SHA256 != closuregraph.ID("sha256:"+hex.EncodeToString(sum[:])) || output.Size != int64(len(payload)) || output.SchemaID != "rust-native-executable-v1" {
			return fail(CodeOfflineRebuildFailed, "built executable differs from issued enforcement receipt", nil)
		}
		return nil
	}
	return fail(CodeOfflineRebuildFailed, "issued enforcement receipt omits the built executable", nil)
}

func (toolchain rustBuildToolchain) identity() (closuregraph.ID, error) {
	values := make([]any, 0, len(requiredBuildToolRoles))
	for _, item := range toolchain.evidence() {
		values = append(values, map[string]any{
			"fingerprint": string(item.ContentFingerprint), "relative_path": item.ExecutableRelativePath,
			"role": string(item.Role), "version": item.VersionOutput,
		})
	}
	return closuregraph.DomainID("rust-native-build-toolchain-v1", map[string]any{"target": toolchain.target, "tools": values})
}

func (state *managerState) executeBuildCargo(
	ctx context.Context,
	operation *closureexec.AssuredOperation,
	invocationKey, phase string,
	argv []string,
	executableLogical string,
) (cargoBuildExecution, error) {
	if err := stageCargoExecutable(state); err != nil {
		return cargoBuildExecution{}, err
	}
	config, err := DeriveSourceReplacementConfig(filepath.Join(state.execRoot, phase, "vendor"), state.lock.Packages)
	if err != nil {
		return cargoBuildExecution{}, err
	}
	seedRoot, err := os.MkdirTemp(state.session, "build-seeds-")
	if err != nil {
		return cargoBuildExecution{}, err
	}
	defer func() { _ = os.RemoveAll(seedRoot) }()
	for _, name := range []string{"cargo-home", "target", "tmp"} {
		if err = privatedir.Make(filepath.Join(seedRoot, name)); err != nil {
			return cargoBuildExecution{}, err
		}
	}
	if err = os.WriteFile(filepath.Join(seedRoot, "cargo-home", "config.toml"), config, 0o600); err != nil {
		return cargoBuildExecution{}, err
	}

	helper := &managerVendorRunner{manager: state}
	vendorInput, vendorID, err := helper.admitTree(ctx, "rust-build-vendor-v1", state.vendor)
	if err != nil {
		return cargoBuildExecution{}, err
	}
	type seed struct {
		input       closureexec.AdmittedInput
		id          closuregraph.ID
		mount, work string
	}
	seeds := []seed{{input: state.workspaceInput, id: state.workspaceID, mount: phase + "/workspace"}, {input: vendorInput, id: vendorID, mount: phase + "/vendor"}}
	for _, name := range []string{"cargo-home", "target", "tmp"} {
		input, id, admitErr := helper.admitTree(ctx, "rust-build-"+name+"-seed-v1", filepath.Join(seedRoot, name))
		if admitErr != nil {
			return cargoBuildExecution{}, admitErr
		}
		seeds = append(seeds, seed{input: input, id: id, mount: phase + "/seeds/" + name, work: phase + "/" + name})
	}
	sort.Slice(seeds, func(i, j int) bool { return seeds[i].id < seeds[j].id })
	ids := make([]closuregraph.ID, len(seeds))
	mounts := make([]closureexec.InputMount, len(seeds))
	inputs := make(map[closuregraph.ID]closureexec.AdmittedInput, len(seeds))
	reads := make([]string, 0, len(seeds)+len(requiredBuildToolRoles))
	works := []closureexec.WorkCopy{}
	for index, item := range seeds {
		ids[index], mounts[index], inputs[item.id] = item.id, closureexec.InputMount{ReceiptID: item.id, Path: item.mount}, item.input
		reads = append(reads, item.mount)
		if item.work != "" {
			works = append(works, closureexec.WorkCopy{ReceiptID: item.id, Path: item.work})
		}
	}
	for _, role := range requiredBuildToolRoles {
		reads = append(reads, "toolchain/"+string(role))
	}
	sort.Strings(reads)
	sort.Slice(works, func(i, j int) bool { return works[i].ReceiptID < works[j].ReceiptID })

	outputRoot := filepath.Join(state.session, phase+"-output")
	if _, statErr := os.Lstat(outputRoot); statErr == nil {
		return cargoBuildExecution{}, fail(CodeLocalOutputUnreceipted, "Cargo output root existed before the permitted build action", map[string]string{"phase": phase})
	} else if !os.IsNotExist(statErr) {
		return cargoBuildExecution{}, statErr
	}
	removeOutput := true
	defer func() {
		if removeOutput {
			_ = os.RemoveAll(outputRoot)
		}
	}()
	executionRoot := state.execRoot
	environment := buildEnvironment(state.buildTools,
		filepath.Join(executionRoot, phase, "cargo-home"),
		filepath.Join(executionRoot, phase, "target"),
		filepath.Join(executionRoot, phase, "tmp"), executionRoot)
	environment["HOME"] = filepath.Join(executionRoot, phase, "cargo-home")
	environment["CURATOR_EXECUTION_ROOT"] = executionRoot
	environment["CURATOR_EVIDENCE_ROOT"] = executionRoot
	environment["CURATOR_OUTPUT_ROOT"] = outputRoot
	manifest := filepath.Join(executionRoot, phase, "workspace", "Cargo.toml")
	configPath := filepath.Join(executionRoot, phase, "cargo-home", "config.toml")
	for index := range argv {
		argv[index] = strings.ReplaceAll(argv[index], "{manifest}", manifest)
		argv[index] = strings.ReplaceAll(argv[index], "{config}", configPath)
	}

	requirements := []closureexec.EvidenceRequirement{}
	stdoutPath := phase + "/cargo-events.json"
	if phase == "build-metadata" {
		stdoutPath = phase + "/metadata.json"
	}
	stdoutID, idErr := closuregraph.DomainID("rust-build-stdout-v1", map[string]any{"invocation": invocationKey, "path": stdoutPath})
	if idErr != nil {
		return cargoBuildExecution{}, idErr
	}
	requirements = append(requirements, closureexec.EvidenceRequirement{Path: stdoutPath, SchemaID: "cargo-json-v1", ArtifactManifestID: stdoutID})
	if executableLogical != "" {
		artifactID, artifactErr := closuregraph.DomainID("rust-build-executable-v1", map[string]any{"invocation": invocationKey, "path": executableLogical})
		if artifactErr != nil {
			return cargoBuildExecution{}, artifactErr
		}
		requirements = append(requirements, closureexec.EvidenceRequirement{Path: executableLogical, SchemaID: "rust-native-executable-v1", ArtifactManifestID: artifactID})
	}
	sort.Slice(requirements, func(i, j int) bool { return requirements[i].Path < requirements[j].Path })
	writes := make([]string, 0, len(requirements)+len(works))
	for _, requirement := range requirements {
		writes = append(writes, requirement.Path)
	}
	for _, work := range works {
		writes = append(writes, work.Path)
	}
	sort.Strings(writes)
	evidenceID, err := evidenceRequirementsID(requirements)
	if err != nil {
		return cargoBuildExecution{}, err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 512 << 20, ReadBytes: 2 << 30, WriteBytes: 2 << 30, WallTimeMillis: 300_000, ProcessCount: 256}
	limitID, err := limits.ID()
	if err != nil {
		return cargoBuildExecution{}, err
	}
	toolchainID, err := state.buildTools.identity()
	if err != nil {
		return cargoBuildExecution{}, err
	}
	c0ID, _ := closuregraph.DomainID("rust-build-c0-v1", map[string]any{"target": state.buildTools.target, "toolchain": string(toolchainID)})
	cargoNodeID, _ := closuregraph.DomainID("rust-cargo-tool-v1", map[string]any{"executable_sha256": string(state.buildTools.items[BuildToolCargo].ContentFingerprint)})
	hostID, _ := closuregraph.DomainID("rust-native-host-v1", map[string]any{"target": state.buildTools.target})
	processes := []string{"toolchain/cargo", "toolchain/linker", "toolchain/rustc"}
	sort.Strings(processes)
	permit := closureexec.DerivationPermit{
		SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: state.causalHead,
		InvocationKey: invocationKey, InvocationSubtype: closureexec.DerivationMetadata,
		AdmittedInputReceiptIDs: ids, InputMounts: mounts, WorkCopies: works,
		C0CheckpointID: c0ID, ToolchainNodeID: cargoNodeID, ToolchainFingerprint: toolchainID,
		ExecutableSHA256: state.buildTools.items[BuildToolCargo].ContentFingerprint,
		Executable:       "bin/cargo", CWD: phase + "/workspace", Argv: append([]string(nil), argv...), Environment: environment,
		HostID: hostID, TargetID: hostID, AllowedProcesses: processes, ReadRoots: reads, WriteRoots: writes,
		StdoutEvidencePath: stdoutPath, ExpectedEvidence: requirements, Network: "none", RecheckRule: "immediate-exact-v1",
		ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID,
	}
	permitID, err := operation.Commit(permit)
	if err != nil {
		return cargoBuildExecution{}, err
	}
	state.processRunner.OutputRoot = outputRoot
	receipt, err := state.executor.Execute(ctx, permitID, func(context.Context) (closureexec.ToolchainIdentity, error) {
		if checkErr := state.buildTools.recheck(state.cargoRegistry); checkErr != nil {
			return closureexec.ToolchainIdentity{}, checkErr
		}
		current, identityErr := state.buildTools.identity()
		return closureexec.ToolchainIdentity{Fingerprint: current, ExecutableSHA256: state.buildTools.items[BuildToolCargo].ContentFingerprint}, identityErr
	}, inputs)
	if err != nil {
		return cargoBuildExecution{}, mapRustExecutionError(err)
	}
	if err = state.executor.VerifyIssuedDerivationReceipt(receipt); err != nil {
		return cargoBuildExecution{}, err
	}
	state.causalHead = string(receipt.NextCausalHead)
	removeOutput = false
	return cargoBuildExecution{
		receipt: receipt, outputRoot: outputRoot, stdoutPath: filepath.Join(outputRoot, filepath.FromSlash(stdoutPath)),
		executable: filepath.Join(outputRoot, filepath.FromSlash(executableLogical)),
		cleanup:    func() { _ = os.RemoveAll(outputRoot) },
	}, nil
}

func mapRustExecutionError(err error) error {
	var diagnostic *closureexec.DiagnosticError
	if !errors.As(err, &diagnostic) {
		return err
	}
	switch diagnostic.Code {
	case "closure_input_undeclared", "closure_write_undeclared", "closure_process_undeclared", "closure_network_attempted":
		return fail(CodeUndeclaredInput, "Cargo attempted an input, process, network, or write outside the committed Rust build permit", map[string]string{"cause": diagnostic.Code})
	case string(CodeToolchainIdentityChanged):
		return fail(CodeToolchainIdentityChanged, "physical Rust toolchain identity changed at the executor time-of-use boundary", nil)
	default:
		return err
	}
}
