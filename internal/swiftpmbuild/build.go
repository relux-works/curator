package swiftpmbuild

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// Manager is the production entry point for SwiftPM planning, offline build,
// and protected publication.
type Manager struct{ config Config }

// NewManager binds the central services, isolated roots, protected store,
// selected linker, and assurance mode exactly once.
func NewManager(config Config) (*Manager, error) {
	if config.Executor == nil || config.Store == nil || config.Policy == nil || config.Recheck == nil {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM build authority is incomplete")
	}
	if !filepath.IsAbs(config.ExecutionRoot) || !filepath.IsAbs(config.OutputRoot) || !filepath.IsAbs(config.StoreRoot) {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM build roots must be absolute")
	}
	if err := validateIsolatedRoots(config); err != nil {
		return nil, err
	}
	if config.Configuration == "" || config.Linker.Role == "" || config.CausalHead == "" {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM build selection is incomplete")
	}
	return &Manager{config: config}, nil
}

// Plan validates the accepted closure and derives the immutable C5 plan.
func (manager *Manager) Plan(ctx context.Context, capture *swiftpmsource.Capture, interop *swiftpminterop.Result) (*Plan, error) {
	if manager == nil {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM build manager is absent")
	}
	return NewPlan(ctx, manager.config, capture, interop)
}

// Build executes exactly one offline native SwiftPM build from the accepted
// C5 plan and publishes its sorted observations and receipts. An exact
// protected-cache hit derived independently from the expected input short-
// circuits before any process starts.
func (manager *Manager) Build(ctx context.Context, plan *Plan) (Result, error) {
	if manager == nil || plan == nil || plan.capture == nil || plan.interop == nil {
		return Result{}, fail(CodeDerivationUnauthorized, "SwiftPM build requires an accepted plan")
	}
	config := manager.config
	if err := recheckSlots(ctx, config, plan.Binding); err != nil {
		return Result{}, err
	}
	operation, err := config.Executor.Preflight(ctx)
	if err != nil {
		return Result{}, err
	}
	cacheInput, err := operation.CacheInput(plan.Expected)
	if err != nil {
		return Result{}, err
	}
	cacheInputID, err := cacheInput.ID()
	if err != nil {
		return Result{}, err
	}
	store, err := closureexec.NewProtectedStore(config.StoreRoot)
	if err != nil {
		return Result{}, err
	}
	if hit, inspectErr := store.Inspect(cacheInput); inspectErr == nil {
		return Result{CacheHit: true, ArtifactPath: hit.Paths[plan.OutputPath], ActiveGraphID: plan.Graph.Active.CapturedGraphID, CommandID: plan.CommandID, Publication: hit.Publication, AssuredCacheInput: cacheInputID}, nil
	} else if !errors.Is(inspectErr, fs.ErrNotExist) {
		return Result{}, inspectErr
	}

	offline, err := manager.runOffline(ctx, operation, plan)
	if err != nil {
		return Result{}, err
	}
	defer offline.cleanup()
	execution, observations := offline.execution, offline.observations
	staging, err := os.MkdirTemp(filepath.Dir(config.OutputRoot), "swiftpm-build-publish-")
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = os.RemoveAll(staging) }()
	if err = stageDeclaredWrites(offline, staging); err != nil {
		return Result{}, err
	}
	publication, err := store.Publish(plan.Publication, cacheInput, execution, observations, staging)
	if err != nil {
		return Result{}, err
	}
	hit, err := store.Inspect(cacheInput)
	if err != nil {
		return Result{}, err
	}
	executionID, err := execution.ID()
	if err != nil {
		return Result{}, err
	}
	c6, err := closuregraph.NewCheckpoint(closuregraph.C6OfflinePayload{ExecutionReceiptID: executionID}, &plan.C5, nil)
	if err != nil {
		return Result{}, err
	}
	publicationID, err := publication.ID()
	if err != nil {
		return Result{}, err
	}
	c7, err := closuregraph.NewCheckpoint(closuregraph.C7PublishPayload{PublicationReceiptID: publicationID}, &c6, nil)
	if err != nil {
		return Result{}, err
	}
	return Result{
		ArtifactPath: hit.Paths[plan.OutputPath], ActiveGraphID: plan.Graph.Active.CapturedGraphID, CommandID: plan.CommandID,
		Execution: execution, Publication: publication, Observations: observations, AssuredCacheInput: cacheInputID,
		ReadSet: offline.readSet, WriteSet: execution.WriteSet, C6: c6, C7: c7,
	}, nil
}

// offlineEvidence is the detached result of one permitted offline build.
type offlineEvidence struct {
	execution    closuregraph.ExecutionReceipt
	observations []closuregraph.ProducedArtifactObservation
	payloads     map[string][]byte
	readSet      []string
	workRoot     string
	cleanup      func()
}

// runOffline materializes the frozen offline build root, commits the exact
// build permit, executes it through the shared seam, and reconciles the issued
// evidence against the planned command and declared write set.
func (manager *Manager) runOffline(ctx context.Context, operation *closureexec.AssuredOperation, plan *Plan) (offlineEvidence, error) {
	config := manager.config
	buildRoot, cleanupRoot, err := materializeCaptureRoot(config, plan.capture)
	if err != nil {
		return offlineEvidence{}, err
	}
	defer cleanupRoot()
	rootInput, rootReceiptID, err := admitDerivedRoot(ctx, config, plan.capture, string(plan.CommandID), buildRoot)
	if err != nil {
		return offlineEvidence{}, err
	}
	mirrors, err := plan.capture.OfflineMirrors()
	if err != nil {
		return offlineEvidence{}, err
	}
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{rootReceiptID: rootInput}
	mounts := []closureexec.InputMount{{ReceiptID: rootReceiptID, Path: buildRootMount}}
	for _, mirror := range mirrors {
		if _, duplicate := inputs[mirror.ReceiptID]; duplicate {
			return offlineEvidence{}, failFields(CodeMirrorMissing, map[string]string{"identity": mirror.Identity}, "two admitted mirrors share one intake receipt")
		}
		inputs[mirror.ReceiptID] = mirror.Input
		mounts = append(mounts, closureexec.InputMount{ReceiptID: mirror.ReceiptID, Path: mirror.Mount})
	}
	sort.Slice(mounts, func(i, j int) bool { return mounts[i].ReceiptID < mounts[j].ReceiptID })
	receiptIDs := make([]closuregraph.ID, len(mounts))
	readRoots := make([]string, len(mounts))
	for index, mount := range mounts {
		receiptIDs[index], readRoots[index] = mount.ReceiptID, mount.Path
	}
	sort.Strings(readRoots)

	evidencePath := path.Join(buildWorkMount, plan.OutputPath)
	writeRoots := []string{buildWorkMount, evidencePath}
	sort.Strings(writeRoots)
	artifactID, err := closuregraph.DomainID("swiftpm-build-evidence-v1", map[string]any{"command": string(plan.CommandID), "path": evidencePath, "schema": ExecutablePayloadSchemaID})
	if err != nil {
		return offlineEvidence{}, err
	}
	requirements := []closureexec.EvidenceRequirement{{Path: evidencePath, SchemaID: ExecutablePayloadSchemaID, ArtifactManifestID: artifactID}}
	evidenceSchemaID, err := buildEvidenceSchemaID(requirements)
	if err != nil {
		return offlineEvidence{}, err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 64 << 20, ReadBytes: 4 << 30, WriteBytes: 4 << 30, WallTimeMillis: 900_000, ProcessCount: 512}
	limitID, err := limits.ID()
	if err != nil {
		return offlineEvidence{}, err
	}
	if err = requireEmptyOutputRoot(config.OutputRoot); err != nil {
		return offlineEvidence{}, err
	}
	c0ID, err := plan.capture.C0.ID()
	if err != nil {
		return offlineEvidence{}, err
	}
	driver := plan.Binding.Slots[SlotSwiftPM]
	permit := closureexec.DerivationPermit{
		SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: operation.CurrentCausalHead(),
		InvocationKey: "swiftpm-offline-build:" + string(plan.CommandID), InvocationSubtype: closureexec.DerivationMetadata,
		AdmittedInputReceiptIDs: receiptIDs, InputMounts: mounts,
		WorkCopies:     []closureexec.WorkCopy{{ReceiptID: rootReceiptID, Path: buildWorkMount, Retain: true}},
		C0CheckpointID: c0ID, ToolchainNodeID: driver.NodeID, ToolchainFingerprint: driver.Payload.ContentFingerprint,
		ExecutableSHA256: executableDigest(driver), Executable: plan.Command.Executable, CWD: plan.Command.CWD,
		Argv: append([]string(nil), plan.Command.Argv...), Environment: resolveEnvironment(config, plan.Command.Environment),
		HostID: plan.Binding.PlatformNodeID, TargetID: plan.Binding.PlatformNodeID, AllowedProcesses: allowedProcesses(config, plan),
		ReadRoots: readRoots, WriteRoots: writeRoots, ExpectedEvidence: requirements, Network: "none",
		RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceSchemaID,
	}
	permitID, err := operation.Commit(permit)
	if err != nil {
		return offlineEvidence{}, err
	}
	workRoot := filepath.Join(config.ExecutionRoot, filepath.FromSlash(buildWorkMount))
	discard := func() { _ = os.RemoveAll(workRoot) }
	receipt, err := operation.Execute(ctx, permitID, func(recheckCtx context.Context) (closureexec.ToolchainIdentity, error) {
		if checkErr := recheckSlots(recheckCtx, config, plan.Binding); checkErr != nil {
			return closureexec.ToolchainIdentity{}, checkErr
		}
		return config.Recheck(recheckCtx, toolIdentity(driver))
	}, inputs)
	if err != nil {
		discard()
		return offlineEvidence{}, mapExecutionError(err)
	}
	if err = operation.VerifyIssuedDerivationReceipt(receipt); err != nil {
		discard()
		return offlineEvidence{}, err
	}
	if err = reconcileCommand(permit, receipt); err != nil {
		discard()
		return offlineEvidence{}, err
	}
	payloads := map[string][]byte{}
	observation, err := observeProduct(config, plan, receipt, evidencePath, payloads)
	if err != nil {
		discard()
		return offlineEvidence{}, err
	}
	observations, err := observeDeclaredObjects(plan, workRoot, payloads)
	if err != nil {
		discard()
		return offlineEvidence{}, err
	}
	observations = append(observations, observation)
	// The publication contract compares the observation slice and the receipt
	// identity list element by element, so both are ordered by exact identity.
	if err = sortObservations(observations); err != nil {
		discard()
		return offlineEvidence{}, err
	}
	writeSet, err := selectedWritePaths(plan)
	if err != nil {
		discard()
		return offlineEvidence{}, err
	}
	observationIDs := make([]closuregraph.ID, len(observations))
	for index := range observations {
		if observationIDs[index], err = observations[index].ID(); err != nil {
			discard()
			return offlineEvidence{}, err
		}
	}
	closureID, err := plan.Closure.ID()
	if err != nil {
		discard()
		return offlineEvidence{}, err
	}
	execution := closuregraph.ExecutionReceipt{
		SchemaID: closuregraph.SchemaExecutionReceipt, ActionOrder: planExecutionOrder(plan.BuildPlan),
		ClosureID: closureID, Decision: "success", Network: "none",
		ProducedObservationIDs: observationIDs, ToolchainRechecks: "match", WriteSet: writeSet,
	}
	if err = execution.Validate(); err != nil {
		discard()
		return offlineEvidence{}, err
	}
	return offlineEvidence{
		execution: execution, observations: observations, payloads: payloads,
		readSet: append([]string(nil), readRoots...), workRoot: workRoot, cleanup: discard,
	}, nil
}

// observeDeclaredObjects reinspects the object each selected compile action
// produced for one source file and binds the exact observed bytes to its
// immutable declared output node. SwiftPM's native build system emits one
// object per source, so every declared object slot resolves to real produced
// bytes; a slot with no object, or an ambiguous one, fails closed. The
// produced set must also be exhausted: an object no declared slot claims is
// undeclared local generation and fails closed before any publication.
func observeDeclaredObjects(plan *Plan, workRoot string, payloads map[string][]byte) ([]closuregraph.ProducedArtifactObservation, error) {
	observations := make([]closuregraph.ProducedArtifactObservation, 0, len(plan.Objects))
	produced := map[string][]string{}
	claimed := map[string]map[string]bool{}
	for _, slot := range plan.Objects {
		candidates, directory, err := targetObjects(plan, workRoot, slot, produced)
		if err != nil {
			return nil, err
		}
		match, err := resolveProducedObject(slot, candidates)
		if err != nil {
			return nil, err
		}
		payload, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(match))) // #nosec G304 -- retained private work copy below the validated execution root.
		if err != nil {
			return nil, err
		}
		if claimed[slot.Target] == nil {
			claimed[slot.Target] = map[string]bool{}
		}
		if claimed[slot.Target][match] {
			return nil, failFields(CodeOutputUnreceipted, map[string]string{"path": slot.Path, "object": match}, "two declared object slots resolve to one produced object")
		}
		claimed[slot.Target][match] = true
		sum := sha256.Sum256(payload)
		observation := closuregraph.ProducedArtifactObservation{
			Class: "native.object", ExpectedOutputNodeID: slot.NodeID, Path: slot.Path,
			ProducerActionID: slot.ActionNodeID, ProducesEdgeID: slot.ProducesEdgeID,
			SHA256: closuregraph.ID("sha256:" + hex.EncodeToString(sum[:])), Size: int64(len(payload)),
		}
		if err = observation.ValidateAgainst(plan.Graph.Records); err != nil {
			return nil, err
		}
		payloads[slot.Path] = payload
		observations = append(observations, observation)
	}
	if err := requireNoUndeclaredObject(produced, claimed); err != nil {
		return nil, err
	}
	return observations, nil
}

// targetObjects lists one selected target's produced objects exactly once and
// returns them with the build directory they were listed from.
func targetObjects(plan *Plan, workRoot string, slot ObjectSlot, produced map[string][]string) ([]string, string, error) {
	directory := filepath.Join(workRoot, filepath.FromSlash(plan.ScratchDirectory), slot.Target+".build")
	if candidates, listed := produced[slot.Target]; listed {
		return candidates, directory, nil
	}
	candidates, err := collectObjectFiles(directory)
	if err != nil {
		return nil, "", failFields(CodeOutputUnreceipted, map[string]string{"target": slot.Package + ":" + slot.Target}, "declared object slot produced no build directory")
	}
	produced[slot.Target] = candidates
	return candidates, directory, nil
}

// resolveProducedObject resolves one declared object slot to the exact file
// SwiftPM produced. A Clang target mirrors the source path below the target
// build directory, relative to the target's declared source root; a Swift
// target flattens it to the source base name. The resolution must be unique,
// so an ambiguous or absent object fails closed.
func resolveProducedObject(slot ObjectSlot, candidates []string) (string, error) {
	base := path.Base(slot.Source) + ".o"
	matches := []string{}
	for _, candidate := range candidates {
		if path.Base(candidate) == base {
			matches = append(matches, candidate)
		}
	}
	if len(matches) > 1 {
		// Two sources of one target can share a base name in different
		// directories. SwiftPM mirrors the tree below the target's declared
		// source root, so the target-relative source path disambiguates them.
		relative := targetRelativeSource(slot)
		exact := []string{}
		for _, candidate := range matches {
			if candidate == relative+".o" {
				exact = append(exact, candidate)
			}
		}
		matches = exact
	}
	if len(matches) != 1 {
		return "", failFields(CodeOutputUnreceipted, map[string]string{"path": slot.Path, "matches": strconv.Itoa(len(matches))}, "declared object slot did not resolve to exactly one produced object")
	}
	return matches[0], nil
}

// targetRelativeSource is the declared source path relative to the target's
// source root, which is exactly what SwiftPM mirrors below <Target>.build.
func targetRelativeSource(slot ObjectSlot) string {
	if slot.SourceRoot == "" {
		return path.Base(slot.Source)
	}
	relative, err := filepath.Rel(filepath.FromSlash(slot.SourceRoot), filepath.FromSlash(slot.Source))
	if err != nil {
		return path.Base(slot.Source)
	}
	relative = filepath.ToSlash(relative)
	if relative == ".." || strings.HasPrefix(relative, "../") {
		return path.Base(slot.Source)
	}
	return relative
}

// requireNoUndeclaredObject proves the produced object set is exactly the
// declared one. An object below a selected target build directory that no
// declared slot claims is undeclared local generation.
func requireNoUndeclaredObject(produced map[string][]string, claimed map[string]map[string]bool) error {
	targets := make([]string, 0, len(produced))
	for target := range produced {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	for _, target := range targets {
		for _, candidate := range produced[target] {
			if !claimed[target][candidate] {
				return failFields(CodeOutputUnreceipted, map[string]string{"target": target, "object": candidate}, "selected target produced an object no declared slot claims")
			}
		}
	}
	return nil
}

// collectObjectFiles lists every object file below one target build directory,
// as slash-separated paths relative to that directory.
func collectObjectFiles(directory string) ([]string, error) {
	if _, err := os.Stat(directory); err != nil {
		return nil, err
	}
	objects := []string{}
	err := filepath.WalkDir(directory, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".o") {
			return nil
		}
		relative, relErr := filepath.Rel(directory, current)
		if relErr != nil {
			return relErr
		}
		objects = append(objects, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(objects)
	return objects, nil
}

// reconcileCommand proves the issued receipt describes exactly the committed
// command and evidence, with no undeclared read, write, process, or network.
func reconcileCommand(permit closureexec.DerivationPermit, receipt closureexec.DerivationReceipt) error {
	if receipt.Audit.Executable != permit.Executable || receipt.Audit.CWD != permit.CWD || !equalStrings(receipt.Audit.Argv, permit.Argv) {
		return fail(CodeBuildGraphDrift, "observed SwiftPM command differs from the committed build plan")
	}
	if receipt.BeforeFingerprint != receipt.AfterFingerprint {
		return fail(CodeToolchainChanged, "SwiftPM build receipt lacks exact enforced toolchain evidence")
	}
	switch receipt.AssuranceMode {
	case closureexec.AssuranceVerified:
		if receipt.Audit.Network != "none" {
			return fail(CodeNetworkAttempted, "verified SwiftPM build observed a network attempt")
		}
		if !equalStrings(receipt.Audit.Reads, permit.ReadRoots) {
			return fail(CodeInputUndeclared, "verified SwiftPM build read outside the committed read set")
		}
		if !equalStrings(receipt.Audit.Writes, permit.WriteRoots) {
			return fail(CodeWriteUndeclared, "verified SwiftPM build wrote outside the committed write set")
		}
		if !equalStrings(receipt.Audit.Processes, permit.AllowedProcesses) {
			return fail(CodeProcessUndeclared, "verified SwiftPM build started an undeclared process")
		}
	case closureexec.AssurancePortable:
		if receipt.Audit.Network != "not-observed" || len(receipt.Audit.Reads) != 0 || len(receipt.Audit.Writes) != 0 || len(receipt.Audit.Processes) != 0 {
			return fail(CodeOfflineRebuildFailed, "portable SwiftPM build receipt inflates execution assurance")
		}
	default:
		return fail(CodeOfflineRebuildFailed, "SwiftPM build receipt assurance mode is unsupported")
	}
	if len(receipt.Outputs) != 1 || receipt.Outputs[0].Path != permit.ExpectedEvidence[0].Path || receipt.Outputs[0].SchemaID != ExecutablePayloadSchemaID {
		return fail(CodeOutputUnreceipted, "issued SwiftPM build receipt omits the exact declared product")
	}
	return nil
}

// observeProduct reinspects the produced bytes and binds them to the immutable
// expected output node without mutating any graph record.
func observeProduct(config Config, plan *Plan, receipt closureexec.DerivationReceipt, evidencePath string, payloads map[string][]byte) (closuregraph.ProducedArtifactObservation, error) {
	payload, err := os.ReadFile(filepath.Join(config.OutputRoot, filepath.FromSlash(evidencePath))) // #nosec G304 -- exact declared evidence path below the validated output root.
	if err != nil {
		return closuregraph.ProducedArtifactObservation{}, failFields(CodeOutputUnreceipted, map[string]string{"path": evidencePath}, "declared SwiftPM product is absent from the protected output root")
	}
	sum := sha256.Sum256(payload)
	digest := closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
	if receipt.Outputs[0].SHA256 != digest || receipt.Outputs[0].Size != int64(len(payload)) {
		return closuregraph.ProducedArtifactObservation{}, fail(CodeOutputDrift, "SwiftPM product bytes differ from the issued enforcement receipt")
	}
	producesEdgeID, err := producesEdgeIdentity(plan)
	if err != nil {
		return closuregraph.ProducedArtifactObservation{}, err
	}
	observation := closuregraph.ProducedArtifactObservation{
		Class: "native.executable", ExpectedOutputNodeID: plan.OutputNodeID, Path: plan.OutputPath,
		ProducerActionID: plan.LinkActionNodeID, ProducesEdgeID: producesEdgeID, SHA256: digest, Size: int64(len(payload)),
	}
	if err = observation.ValidateAgainst(plan.Graph.Records); err != nil {
		return closuregraph.ProducedArtifactObservation{}, err
	}
	payloads[plan.OutputPath] = payload
	return observation, nil
}

// stageDeclaredWrites materializes exactly the C4/C5 write set in a private
// staging tree from the reinspected bytes of every declared output, so the
// protected store publishes the product and every per-source object it
// causally observed and nothing else.
func stageDeclaredWrites(offline offlineEvidence, staging string) error {
	for _, observation := range offline.observations {
		payload, present := offline.payloads[observation.Path]
		if !present {
			return failFields(CodeOutputUnreceipted, map[string]string{"path": observation.Path}, "declared output has no reinspected bytes")
		}
		mode := fs.FileMode(0o600)
		if observation.Class == "native.executable" {
			mode = 0o500
		}
		if err := writeStagedFile(staging, observation.Path, payload, mode); err != nil {
			return err
		}
	}
	return nil
}

func writeStagedFile(staging, logical string, payload []byte, mode fs.FileMode) error {
	target := filepath.Join(staging, filepath.FromSlash(logical))
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	return os.WriteFile(target, payload, mode)
}

// selectedWritePaths derives the exact C4/C5 write set: every path a selected
// action declares it produces, sorted and unique.
func selectedWritePaths(plan *Plan) ([]string, error) {
	actions := map[closuregraph.ID]bool{}
	for _, id := range plan.BuildPlan.ActionNodeIDs {
		actions[id] = true
	}
	pruned := map[closuregraph.ID]bool{}
	for _, activation := range plan.Graph.Active.EdgeActivations {
		if activation.State != closuregraph.ActivationSelected {
			pruned[activation.EdgeID] = true
		}
	}
	seen := map[string]bool{}
	paths := []string{}
	for _, edge := range plan.Graph.Records.CaptureEdges {
		if edge.Kind != closuregraph.EdgeProduces || !actions[edge.FromNodeID] {
			continue
		}
		id, err := edge.ID()
		if err != nil {
			return nil, err
		}
		if pruned[id] {
			continue
		}
		value := edge.Payload.(closuregraph.ProducesPayload).Path
		if !seen[value] {
			seen[value] = true
			paths = append(paths, value)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

// planExecutionOrder flattens the exact deterministic C5 waves. One SwiftPM
// process performs every action, so the reported order is the plan's own
// stable topological order rather than a rediscovered one.
func planExecutionOrder(plan closuregraph.BuildPlan) []closuregraph.ID {
	order := make([]closuregraph.ID, 0, len(plan.ActionNodeIDs))
	for _, wave := range plan.Waves {
		order = append(order, wave...)
	}
	return order
}

// sortObservations orders produced observations by their exact canonical
// identity, which is the order the shared publication contract requires.
func sortObservations(observations []closuregraph.ProducedArtifactObservation) error {
	identities := make(map[string]closuregraph.ID, len(observations))
	for _, observation := range observations {
		id, err := observation.ID()
		if err != nil {
			return err
		}
		identities[observation.Path] = id
	}
	sort.Slice(observations, func(i, j int) bool {
		return identities[observations[i].Path] < identities[observations[j].Path]
	})
	return nil
}

func producesEdgeIdentity(plan *Plan) (closuregraph.ID, error) {
	for _, edge := range plan.Graph.Records.CaptureEdges {
		if edge.Kind == closuregraph.EdgeProduces && edge.FromNodeID == plan.LinkActionNodeID && edge.ToNodeID == plan.OutputNodeID {
			return edge.ID()
		}
	}
	return "", fail(CodeGraphReferenceInvalid, "planned product has no produces edge")
}

// mapExecutionError renames no shared cause: undeclared reads, writes,
// processes, and network attempts keep the vocabulary the executor raised.
func mapExecutionError(err error) error {
	var diagnostic *closureexec.DiagnosticError
	if !errors.As(err, &diagnostic) {
		return err
	}
	switch Code(diagnostic.Code) {
	case CodeInputUndeclared, CodeWriteUndeclared, CodeProcessUndeclared, CodeNetworkAttempted:
		return failFields(Code(diagnostic.Code), map[string]string{"cause": diagnostic.Code}, "SwiftPM build attempted an input, process, network, or write outside the committed permit")
	case CodeToolchainChanged:
		return fail(CodeToolchainChanged, "physical SwiftPM toolchain identity changed at the executor time-of-use boundary")
	case "closure_derivation_drift":
		return fail(CodeOfflineRebuildFailed, "offline SwiftPM build failed or drifted from its committed permit: %s", diagnostic.Detail)
	default:
		return err
	}
}

const (
	// buildRootMount is the immutable admitted offline build root.
	buildRootMount = "inputs/build-root"
	// buildWorkMount is the private writable work copy of that root.
	buildWorkMount = "work/package"
)

// materializeCaptureRoot derives an offline build root: the admitted root tree
// plus the frozen lock and the generated kind-preserving mirror configuration.
func materializeCaptureRoot(config Config, capture *swiftpmsource.Capture) (string, func(), error) {
	rootInput, _, err := capture.RootInput()
	if err != nil {
		return "", func() {}, err
	}
	if err = rootInput.Tree.VerifyAtUse(); err != nil {
		return "", func() {}, fail(CodeOfflineRebuildFailed, "admitted root tree changed before the offline build")
	}
	source, err := rootInput.Tree.ProtectedPath()
	if err != nil {
		return "", func() {}, err
	}
	temporary, err := os.MkdirTemp(filepath.Dir(config.ExecutionRoot), "swiftpm-build-root-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(temporary) }
	target := filepath.Join(temporary, "root")
	if err = copyRegularTree(source, target); err != nil {
		cleanup()
		return "", func() {}, err
	}
	if err = os.WriteFile(filepath.Join(target, "Package.resolved"), capture.Lock.Bytes, 0o600); err != nil {
		cleanup()
		return "", func() {}, err
	}
	mirrors, err := capture.OfflineMirrors()
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	type mirrorEntry struct {
		Original string `json:"original"`
		Mirror   string `json:"mirror"`
	}
	entries := make([]mirrorEntry, len(mirrors))
	for index, mirror := range mirrors {
		location := filepath.Join(config.ExecutionRoot, filepath.FromSlash(mirror.Mount))
		if mirror.Kind == "remoteSourceControl" {
			location = "file://" + location
		}
		entries[index] = mirrorEntry{Original: mirror.Original, Mirror: location}
	}
	payload, err := json.Marshal(map[string]any{"object": entries, "version": 1})
	if err != nil {
		cleanup()
		return "", func() {}, err
	}
	for _, directory := range []string{"cache", "config", "home", "scratch", "security"} {
		if err = os.MkdirAll(filepath.Join(target, scratchRoot, directory), 0o700); err != nil {
			cleanup()
			return "", func() {}, err
		}
	}
	if err = os.WriteFile(filepath.Join(target, scratchRoot, "config", "mirrors.json"), payload, 0o600); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return target, cleanup, nil
}

func requireEmptyOutputRoot(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(root, 0o700)
		}
		return err
	}
	if len(entries) != 0 {
		return failFields(CodeOutputUnreceipted, map[string]string{"root": root}, "SwiftPM output root was not empty before the permitted build action")
	}
	return nil
}

func validateIsolatedRoots(config Config) error {
	execution := filepath.Clean(config.ExecutionRoot)
	output := filepath.Clean(config.OutputRoot)
	store := filepath.Clean(config.StoreRoot)
	if execution == output || execution == store || output == store {
		return fail(CodeDerivationUnauthorized, "SwiftPM execution, output, and protected store roots must be distinct")
	}
	relative, err := filepath.Rel(execution, output)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fail(CodeDerivationUnauthorized, "SwiftPM output root must be a child of the execution root")
	}
	if storeRelative, relErr := filepath.Rel(execution, store); relErr == nil && storeRelative != ".." && !strings.HasPrefix(storeRelative, ".."+string(filepath.Separator)) {
		return fail(CodeDerivationUnauthorized, "protected store must live outside the SwiftPM execution root")
	}
	return nil
}

func resolveEnvironment(config Config, logical map[string]string) map[string]string {
	resolved := make(map[string]string, len(logical)+3)
	for key, value := range logical {
		resolved[key] = strings.ReplaceAll(value, executionRootPlaceholder, config.ExecutionRoot)
	}
	resolved["CURATOR_OUTPUT_ROOT"] = config.OutputRoot
	resolved["CURATOR_EVIDENCE_ROOT"] = config.ExecutionRoot
	resolved["PATH"] = filepath.Join(config.ExecutionRoot, "bin")
	return resolved
}

func allowedProcesses(config Config, plan *Plan) []string {
	processes := append([]string(nil), config.AllowedProcesses...)
	if len(processes) == 0 {
		seen := map[string]bool{}
		slots := make([]ToolSlot, 0, len(plan.Binding.Slots))
		for slot := range plan.Binding.Slots {
			slots = append(slots, slot)
		}
		sort.Slice(slots, func(i, j int) bool { return slots[i] < slots[j] })
		for _, slot := range slots {
			value := plan.Binding.Slots[slot].Payload.ExecutableRelativePath
			if value != "" && !seen[value] {
				seen[value] = true
				processes = append(processes, value)
			}
		}
	}
	sort.Strings(processes)
	return processes
}

func executableDigest(bound SlotBinding) closuregraph.ID {
	if len(bound.Payload.LinkFingerprintIDs) == 1 {
		return bound.Payload.LinkFingerprintIDs[0]
	}
	return bound.Payload.ContentFingerprint
}

func buildEvidenceSchemaID(requirements []closureexec.EvidenceRequirement) (closuregraph.ID, error) {
	values := make([]any, len(requirements))
	for index, requirement := range requirements {
		values[index] = map[string]any{"artifact_manifest_id": string(requirement.ArtifactManifestID), "path": requirement.Path, "schema_id": requirement.SchemaID}
	}
	return closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": values})
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// copyRegularTree copies only regular files and directories. Links, devices,
// and other special nodes never reach the offline build root.
func copyRegularTree(source, target string) error {
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.IsDir():
			return os.MkdirAll(destination, 0o700)
		case info.Mode().IsRegular():
			payload, readErr := os.ReadFile(current) // #nosec G304 -- admitted protected tree walked below its own root.
			if readErr != nil {
				return readErr
			}
			mode := fs.FileMode(0o600)
			if info.Mode().Perm()&0o100 != 0 {
				mode = 0o700
			}
			return os.WriteFile(destination, payload, mode)
		default:
			return failFields(CodeOfflineRebuildFailed, map[string]string{"path": relative}, "admitted root contains a non-regular node")
		}
	})
}

// admitDerivedRoot admits a derived offline build root through the shared
// recursive artifact classifier and the capture store. The derivative is bound
// to the exact frozen lock, graph, and command identities so a different plan
// can never reuse this intake receipt.
func admitDerivedRoot(ctx context.Context, config Config, capture *swiftpmsource.Capture, key, root string) (closureexec.AdmittedInput, closuregraph.ID, error) {
	handle, err := config.Store.CaptureTree("swiftpm-build-root:"+key, root)
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	protected, err := handle.ProtectedPath()
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	descriptor := artifactpolicy.Descriptor{
		AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileSwiftPMV1, Manager: "swiftpm",
		PackageName: "offline-build-root", PackageVersion: key,
	}
	probe, probeErr := config.Policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: protected, VirtualRoot: "package"})
	if probeErr != nil && artifactpolicy.ErrorCode(probeErr) != artifactpolicy.CodeOriginUnverified {
		return closureexec.AdmittedInput{}, "", probeErr
	}
	treeDigest := closuregraph.ID(probe.Manifest.RawPayload.SHA256)
	if !treeDigest.Valid() {
		return closureexec.AdmittedInput{}, "", fail(CodeDerivationUnauthorized, "artifact policy did not derive a canonical build-root identity")
	}
	descriptor.Origin = artifactpolicy.OriginEvidence{
		Locator: "swiftpm-build-root", ImmutableID: key,
		LockRecord: string(capture.Lock.Digest), ChecksumSHA256: string(treeDigest), Verified: true,
	}
	result, policyErr := config.Policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: protected, VirtualRoot: "package"})
	if policyErr != nil {
		return closureexec.AdmittedInput{}, "", policyErr
	}
	manifestID := closuregraph.ID(result.Manifest.ManifestDigest)
	if result.Admission == nil || !manifestID.Valid() {
		return closureexec.AdmittedInput{}, "", fail(CodeDerivationUnauthorized, "artifact admission did not issue build-root authority")
	}
	receipt, err := config.Store.AdmitTree(handle, "swiftpm-build-root", closureexec.AdmissionEvidence{
		PreviousCausalHead: config.CausalHead, ArtifactPolicyID: artifactpolicy.PolicyID,
		SourceProfileID: string(artifactpolicy.ProfileSwiftPMV1), DetectorRegistryID: artifactpolicy.DetectorRegistryID,
		LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID,
	})
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	return closureexec.AdmittedInput{Receipt: receipt, Tree: handle}, receiptID, nil
}
