package rustsource

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

const buildProfileID = "curator"

// BuildBinding is the exact C4 overlay consumed by rust-source-v1. It repeats
// no tool identity: every physical component is resolved from canonical
// selection-binding nodes and typed edges.
type BuildBinding struct {
	C4            closuregraph.Checkpoint
	Selection     closuregraph.SelectionBinding
	Nodes         []closuregraph.Node
	Edges         []closuregraph.Edge
	ProductNodeID closuregraph.ID
	ActionNodeID  closuregraph.ID
}

// BuildRequest executes one exact package/bin selection over an admitted
// capture and its manager-issued active metadata.
type BuildRequest struct {
	Capture     *Capture
	Metadata    MetadataResult
	Selection   SelectionContext
	Binding     BuildBinding
	Publication closuregraph.PublicationEvidence
	StoreRoot   string
}

// CargoEvent is the normalized security projection of one accepted Cargo JSON
// message. Rendered diagnostics and transient filesystem paths are excluded.
type CargoEvent struct {
	Reason, PackageID, TargetName string
	TargetKinds                   []string
	Executable                    bool
}

// BuildResult contains detached C6/C7 evidence and the protected executable.
type BuildResult struct {
	CacheHit          bool
	ArtifactPath      string
	ActiveGraphID     string
	CommandID         closuregraph.ID
	CargoEventsID     closuregraph.ID
	Execution         closuregraph.ExecutionReceipt
	Publication       closuregraph.PublicationReceipt
	Observation       closuregraph.ProducedArtifactObservation
	AssuredCacheInput closuregraph.ID
}

// Build performs fail-closed binding/unit validation before any compiler
// process, reconstructs metadata and build state from fresh roots, validates
// Cargo JSON, and publishes only through the shared protected store.
func (m *Manager) Build(ctx context.Context, request BuildRequest) (BuildResult, error) {
	state, err := m.authority()
	if err != nil {
		return BuildResult{}, err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.closed || request.Capture == nil || request.Capture.state == nil || request.Capture.state.owner != state || request.Metadata.owner != state || request.Metadata.capture != request.Capture.state {
		return BuildResult{}, fail(CodeGraphIncomplete, "build capture or metadata authority is absent, closed, or foreign", nil)
	}
	if !reflect.DeepEqual(request.Metadata.selection, request.Selection) {
		return BuildResult{}, fail(CodeFeatureProfileMismatch, "build selection differs from metadata selection", nil)
	}
	if request.StoreRoot == "" || !filepath.IsAbs(request.StoreRoot) {
		return BuildResult{}, fail(CodeConfigUntrusted, "protected store root must be absolute", nil)
	}
	if err = validateBuildBinding(request.Binding, request.Publication, request.Selection, state.buildTools); err != nil {
		return BuildResult{}, err
	}
	active, err := Reconcile(request.Capture.state.graph, request.Selection, request.Metadata.Active, state.buildTools.target, state.cargo.Fingerprint)
	if err != nil {
		return BuildResult{}, err
	}
	if err = validateBuildUnits(request.Metadata.Active, request.Selection); err != nil {
		return BuildResult{}, err
	}
	if err = rejectBuildConfiguration(request.Capture.state.workspace); err != nil {
		return BuildResult{}, err
	}
	if err = state.buildTools.recheck(state.cargoRegistry); err != nil {
		return BuildResult{}, err
	}

	closureID, expected, outputNode, actionNode, producesEdge, err := publicationSlots(request.Publication)
	if err != nil {
		return BuildResult{}, err
	}
	operation := state.buildOperation
	if operation == nil {
		operation, err = state.executor.Preflight(ctx)
		if err != nil {
			return BuildResult{}, err
		}
		state.buildOperation = operation
	} else if err = operation.Revalidate(ctx); err != nil {
		return BuildResult{}, err
	}
	cacheInput, err := operation.CacheInput(expected)
	if err != nil {
		return BuildResult{}, err
	}
	cacheInputID, _ := cacheInput.ID()
	store, err := closureexec.NewProtectedStore(request.StoreRoot)
	if err != nil {
		return BuildResult{}, err
	}
	if hit, inspectErr := store.Inspect(cacheInput); inspectErr == nil {
		return BuildResult{CacheHit: true, ArtifactPath: hit.Paths[outputNode.Payload.(closuregraph.OutputArtifactPayload).LogicalPath], ActiveGraphID: active.Identity, Publication: hit.Publication, AssuredCacheInput: cacheInputID}, nil
	} else if !errors.Is(inspectErr, fs.ErrNotExist) {
		return BuildResult{}, inspectErr
	}

	metadataArgv := []string{"metadata", "--config", "{config}", "--format-version", "1", "--frozen", "--filter-platform", request.Selection.Target, "--manifest-path", "{manifest}"}
	metadataArgv = appendFeatureArguments(metadataArgv, request.Selection)
	metadataRun, err := state.executeBuildCargo(ctx, operation, "rust-build-fresh-metadata-v1", "build-metadata", metadataArgv, "")
	if err != nil {
		return BuildResult{}, err
	}
	defer metadataRun.cleanup()
	metadataBytes, err := os.ReadFile(metadataRun.stdoutPath) // #nosec G304 -- issued executor receipt binds this exact evidence path.
	if err != nil {
		return BuildResult{}, err
	}
	freshMetadata, err := ParseMetadata(metadataBytes)
	if err == nil {
		err = normalizeMetadataPaths(&freshMetadata, map[string]string{filepath.Join(state.execRoot, "build-metadata", "workspace"): "workspace", filepath.Join(state.execRoot, "build-metadata", "vendor"): "vendor"})
	}
	if err != nil || !reflect.DeepEqual(freshMetadata, request.Metadata.Active) {
		return BuildResult{}, fail(CodeGraphIncomplete, "fresh-home metadata differs from C4 active graph", map[string]string{"stage": "metadata"})
	}

	buildArgv := []string{"build", "--config", "{config}", "--frozen", "--profile", buildProfileID, "--package", request.Selection.Package, "--bin", request.Selection.Binary, "--target", request.Selection.Target, "--message-format", "json-render-diagnostics", "--manifest-path", "{manifest}"}
	buildArgv = appendFeatureArguments(buildArgv, request.Selection)
	commandID, err := closuregraph.DomainID("rust-build-command-v1", map[string]any{"argv": stringValues(buildArgv), "profile": buildProfileID, "target": request.Selection.Target, "toolchain": string(state.buildTools.items[BuildToolCargo].ContentFingerprint)})
	if err != nil {
		return BuildResult{}, err
	}
	executableLogical := filepath.ToSlash(filepath.Join("build", "target", request.Selection.Target, buildProfileID, request.Selection.Binary))
	buildRun, err := state.executeBuildCargo(ctx, operation, "rust-build-frozen-v1:"+string(commandID), "build", buildArgv, executableLogical)
	if err != nil {
		return BuildResult{}, err
	}
	defer buildRun.cleanup()
	stdout, err := os.ReadFile(buildRun.stdoutPath) // #nosec G304 -- issued executor receipt binds this exact evidence path.
	if err != nil {
		return BuildResult{}, err
	}
	selectedPackage, err := selectedMetadataPackage(request.Metadata.Active, request.Selection)
	if err != nil {
		return BuildResult{}, err
	}
	targetDir := filepath.Join(state.execRoot, "build", "target")
	events, executable, err := validateCargoEvents(stdout, request.Selection, selectedPackage, targetDir, buildRun.executable)
	if err != nil {
		return BuildResult{}, err
	}
	if err = state.buildTools.recheck(state.cargoRegistry); err != nil {
		return BuildResult{}, err
	}
	if err = state.executor.VerifyIssuedDerivationReceipt(metadataRun.receipt); err != nil {
		return BuildResult{}, err
	}
	if err = state.executor.VerifyIssuedDerivationReceipt(buildRun.receipt); err != nil {
		return BuildResult{}, err
	}

	output := outputNode.Payload.(closuregraph.OutputArtifactPayload)
	staging, err := os.MkdirTemp(state.session, "build-publish-")
	if err != nil {
		return BuildResult{}, err
	}
	defer func() { _ = os.RemoveAll(staging) }()
	artifactPath := filepath.Join(staging, filepath.FromSlash(output.LogicalPath))
	if err = os.MkdirAll(filepath.Dir(artifactPath), 0o700); err != nil {
		return BuildResult{}, err
	}
	payload, err := os.ReadFile(executable) // #nosec G304 -- executable path is validated from Cargo JSON below the private target root.
	if err != nil {
		return BuildResult{}, err
	}
	if err = verifyIssuedExecutable(buildRun.receipt, executableLogical, payload); err != nil {
		return BuildResult{}, err
	}
	if err = os.WriteFile(artifactPath, payload, 0o500); err != nil {
		return BuildResult{}, err
	}
	sum := sha256.Sum256(payload)
	observation := closuregraph.ProducedArtifactObservation{Class: output.ExpectedClass, ExpectedOutputNodeID: mustNodeID(outputNode), Path: output.LogicalPath, ProducerActionID: mustNodeID(actionNode), ProducesEdgeID: mustEdgeID(producesEdge), SHA256: closuregraph.ID("sha256:" + hex.EncodeToString(sum[:])), Size: int64(len(payload))}
	observationID, err := observation.ID()
	if err != nil {
		return BuildResult{}, err
	}
	eventID, err := closuregraph.DomainID("rust-cargo-events-v1", cargoEventValues(events))
	if err != nil {
		return BuildResult{}, err
	}
	execution, err := executionFromIssuedCargoReceipts(metadataRun.receipt, buildRun.receipt, closureID, request.Publication.Plan.ActionNodeIDs, []closuregraph.ID{observationID}, []string{output.LogicalPath})
	if err != nil {
		return BuildResult{}, err
	}
	publication, err := store.Publish(request.Publication, cacheInput, execution, []closuregraph.ProducedArtifactObservation{observation}, staging)
	if err != nil {
		return BuildResult{}, err
	}
	hit, err := store.Inspect(cacheInput)
	if err != nil {
		return BuildResult{}, err
	}
	return BuildResult{ArtifactPath: hit.Paths[output.LogicalPath], ActiveGraphID: active.Identity, CommandID: commandID, CargoEventsID: eventID, Execution: execution, Publication: publication, Observation: observation, AssuredCacheInput: cacheInputID}, nil
}

func validateBuildBinding(binding BuildBinding, publication closuregraph.PublicationEvidence, selection SelectionContext, tools rustBuildToolchain) error {
	if err := binding.C4.Validate(); err != nil || binding.C4.Name != closuregraph.CheckpointC4 {
		return fail(CodeGraphReferenceInvalid, "C4 checkpoint is absent or invalid", nil)
	}
	if err := binding.Selection.Validate(); err != nil {
		return fail(CodeGraphReferenceInvalid, "selection binding is invalid", nil)
	}
	c4, ok := binding.C4.Payload.(closuregraph.C4ClosePayload)
	if !ok {
		return fail(CodeGraphReferenceInvalid, "C4 payload kind is invalid", nil)
	}
	bindingID, _ := binding.Selection.ID()
	if c4.SelectionBindingID != bindingID || binding.Selection.CapturedGraphID != c4.CapturedGraphID || binding.Selection.SelectionContextID != c4.SelectionContextID {
		return fail(CodeGraphReferenceInvalid, "C4 does not name the supplied selection binding", nil)
	}
	publicationC4ID, _ := publication.C4.ID()
	bindingC4ID, _ := binding.C4.ID()
	if publicationC4ID != bindingC4ID {
		return fail(CodeGraphReferenceInvalid, "publication authority uses another C4 checkpoint", nil)
	}
	nodes := map[closuregraph.ID]closuregraph.Node{}
	roles := map[BuildToolRole]closuregraph.ID{}
	var platformID closuregraph.ID
	for _, node := range binding.Nodes {
		if err := node.Validate(); err != nil {
			return fail(CodeGraphReferenceInvalid, "binding contains an invalid node", map[string]string{"detail": err.Error()})
		}
		id := mustNodeID(node)
		if nodes[id].Kind != "" {
			return fail(CodeGraphReferenceInvalid, "binding node is duplicated", map[string]string{"node": string(id)})
		}
		nodes[id] = node
		switch node.Kind {
		case closuregraph.NodeTargetPlatform:
			if platformID != "" {
				return fail(CodeGraphReferenceInvalid, "target platform binding is duplicated", nil)
			}
			payload := node.Payload.(closuregraph.TargetPlatformPayload)
			if payload.TargetTriple != selection.Target || selection.Target != tools.target {
				return fail(CodeTargetUnsupported, "bound target is not the selected native target", map[string]string{"target": payload.TargetTriple, "host": tools.target})
			}
			platformID = id
		case closuregraph.NodeToolchainComponent:
			payload := node.Payload.(closuregraph.ToolchainComponentPayload)
			role := BuildToolRole(payload.ComponentRole)
			expected, exists := tools.items[role]
			if !exists || roles[role] != "" || payload.ContentFingerprint != expected.ContentFingerprint || payload.ExecutableRelativePath != expected.ExecutableRelativePath || payload.VersionOutput != expected.VersionOutput || payload.TimeOfUseRecheckRule != "immediate-exact-v1" {
				return fail(CodeGraphReferenceInvalid, "toolchain component is missing, duplicated, or drifted", map[string]string{"role": payload.ComponentRole})
			}
			roles[role] = id
		default:
			return fail(CodeGraphReferenceInvalid, "binding contains a wrong-kind node", map[string]string{"kind": string(node.Kind)})
		}
	}
	if platformID == "" || len(roles) != len(requiredBuildToolRoles) {
		return fail(CodeGraphReferenceInvalid, "target platform or required toolchain component is missing", nil)
	}
	nodeIDs := append([]closuregraph.ID(nil), binding.Selection.BindingNodeIDs...)
	sort.Slice(nodeIDs, func(i, j int) bool { return nodeIDs[i] < nodeIDs[j] })
	actualNodeIDs := make([]closuregraph.ID, 0, len(nodes))
	for id := range nodes {
		actualNodeIDs = append(actualNodeIDs, id)
	}
	sort.Slice(actualNodeIDs, func(i, j int) bool { return actualNodeIDs[i] < actualNodeIDs[j] })
	if !reflect.DeepEqual(nodeIDs, actualNodeIDs) {
		return fail(CodeGraphReferenceInvalid, "selection binding node references are dangling or incomplete", nil)
	}
	uses, targets := map[BuildToolRole]int{}, map[closuregraph.ID]int{}
	edgeIDs := make([]closuregraph.ID, 0, len(binding.Edges))
	for _, edge := range binding.Edges {
		if err := edge.Validate(); err != nil {
			return fail(CodeGraphReferenceInvalid, "binding contains an invalid edge", nil)
		}
		edgeIDs = append(edgeIDs, mustEdgeID(edge))
		switch edge.Kind {
		case closuregraph.EdgeUsesTool:
			payload := edge.Payload.(closuregraph.UsesToolPayload)
			role := BuildToolRole(payload.ToolSlot)
			if edge.FromNodeID != binding.ActionNodeID || edge.ToNodeID != roles[role] || payload.ExecutableRelativePath != tools.items[role].ExecutableRelativePath {
				return fail(CodeGraphReferenceInvalid, "uses_tool edge has a dangling or wrong-kind slot", map[string]string{"role": payload.ToolSlot})
			}
			uses[role]++
		case closuregraph.EdgeTargets:
			if edge.ToNodeID != platformID || edge.Payload.(closuregraph.TargetsPayload).BindingRole != closuregraph.PlatformTarget {
				return fail(CodeGraphReferenceInvalid, "targets edge does not resolve to the native target", nil)
			}
			targets[edge.FromNodeID]++
		default:
			return fail(CodeGraphReferenceInvalid, "binding edge kind is unsupported by rust-source-v1", map[string]string{"kind": string(edge.Kind)})
		}
	}
	sort.Slice(edgeIDs, func(i, j int) bool { return edgeIDs[i] < edgeIDs[j] })
	expectedEdgeIDs := append([]closuregraph.ID(nil), binding.Selection.BindingEdgeIDs...)
	sort.Slice(expectedEdgeIDs, func(i, j int) bool { return expectedEdgeIDs[i] < expectedEdgeIDs[j] })
	if !reflect.DeepEqual(edgeIDs, expectedEdgeIDs) {
		return fail(CodeGraphReferenceInvalid, "selection binding edge references are dangling or incomplete", nil)
	}
	for _, role := range requiredBuildToolRoles {
		if uses[role] != 1 || targets[roles[role]] != 1 {
			return fail(CodeGraphReferenceInvalid, "toolchain slot must have one uses_tool and one targets edge", map[string]string{"role": string(role)})
		}
	}
	if targets[binding.ProductNodeID] != 1 || targets[binding.ActionNodeID] != 1 {
		return fail(CodeGraphReferenceInvalid, "product and build action require exact target bindings", nil)
	}
	return nil
}

func validateBuildUnits(metadata Metadata, selection SelectionContext) error {
	selected := 0
	for _, pkg := range metadata.Packages {
		if pkg.Links != "" {
			return fail(CodeNativeLinkUnsupported, "active package declares native links", map[string]string{"package": pkg.ID})
		}
		for _, target := range pkg.Targets {
			for _, kind := range target.Kinds {
				switch kind {
				case "custom-build":
					return fail(CodeBuildScriptUnsupported, "active custom-build target", map[string]string{"package": pkg.ID, "target": target.Name})
				case "proc-macro":
					return fail(CodeProcMacroUnsupported, "active proc-macro target", map[string]string{"package": pkg.ID, "target": target.Name})
				}
			}
			if pkg.Name == selection.Package && pkg.Source == "" && target.Name == selection.Binary && reflect.DeepEqual(target.Kinds, []string{"bin"}) && reflect.DeepEqual(target.CrateTypes, []string{"bin"}) {
				selected++
			}
		}
	}
	for _, node := range metadata.Resolve {
		for _, dep := range node.Dependencies {
			if dep.Kind == "build" {
				return fail(CodeBuildScriptUnsupported, "active build dependency", map[string]string{"package": node.ID})
			}
		}
	}
	if selected != 1 {
		return fail(CodeGraphIncomplete, "exactly one package/bin build unit is required", nil)
	}
	return nil
}

func rejectBuildConfiguration(workspace string) error {
	for _, name := range []string{"rust-toolchain", "rust-toolchain.toml"} {
		if _, err := os.Lstat(filepath.Join(workspace, name)); err == nil {
			return fail(CodeConfigUntrusted, "package-selected Rust toolchain is unsupported", map[string]string{"path": name})
		} else if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return rejectAmbientCargoConfig(workspace)
}

func appendFeatureArguments(argv []string, selection SelectionContext) []string {
	if !selection.DefaultFeatures {
		argv = append(argv, "--no-default-features")
	}
	features := append([]string(nil), selection.Features...)
	sort.Strings(features)
	if len(features) > 0 {
		argv = append(argv, "--features", strings.Join(features, ","))
	}
	return argv
}

func buildEnvironment(tools rustBuildToolchain, home, target, temp, root string) map[string]string {
	tripleKey := strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(tools.target))
	env := map[string]string{
		"CARGO_HOME": home, "CARGO_NET_OFFLINE": "true", "CARGO_TARGET_DIR": target,
		"HOME": filepath.Join(root, "home"), "LANG": "C", "LC_ALL": "C", "TMPDIR": temp, "TZ": "UTC",
		"RUSTC":                          tools.items[BuildToolRustc].PhysicalPath,
		"RUSTFLAGS":                      "--remap-path-prefix=" + root + "=/curator/rust-source-v1",
		"CARGO_PROFILE_CURATOR_INHERITS": "release", "CARGO_PROFILE_CURATOR_DEBUG": "0",
		"CARGO_PROFILE_CURATOR_INCREMENTAL": "false", "CARGO_PROFILE_CURATOR_STRIP": "none",
		"CARGO_TARGET_" + tripleKey + "_LINKER": tools.items[BuildToolLinker].PhysicalPath,
	}
	_ = os.Mkdir(env["HOME"], 0o700)
	if tools.items[BuildToolSDK].PhysicalPath != tools.items[BuildToolSysroot].PhysicalPath {
		env["SDKROOT"] = tools.items[BuildToolSDK].PhysicalPath
	}
	return env
}

func validateCargoEvents(payload []byte, selection SelectionContext, selectedPackage MetadataPackage, targetRoot, evidenceExecutable string) ([]CargoEvent, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(payload))
	scanner.Buffer(make([]byte, 64<<10), 16<<20)
	events := []CargoEvent{}
	executable := ""
	finished := false
	for scanner.Scan() {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			return nil, "", fail(CodeGraphIncomplete, "Cargo emitted malformed JSON", nil)
		}
		var reason string
		if json.Unmarshal(raw["reason"], &reason) != nil {
			return nil, "", fail(CodeGraphIncomplete, "Cargo event reason is missing", nil)
		}
		event := CargoEvent{Reason: reason}
		switch reason {
		case "compiler-message":
		case "build-script-executed":
			return nil, "", fail(CodeBuildScriptUnsupported, "Cargo executed a build script", nil)
		case "compiler-artifact":
			_ = json.Unmarshal(raw["package_id"], &event.PackageID)
			var target struct {
				Name string   `json:"name"`
				Kind []string `json:"kind"`
			}
			if json.Unmarshal(raw["target"], &target) != nil || target.Name == "" || len(target.Kind) == 0 {
				return nil, "", fail(CodeGraphIncomplete, "Cargo artifact target is malformed", nil)
			}
			event.TargetName, event.TargetKinds = target.Name, append([]string(nil), target.Kind...)
			var path *string
			if value, ok := raw["executable"]; ok && string(value) != "null" {
				var candidate string
				if json.Unmarshal(value, &candidate) != nil {
					return nil, "", fail(CodeGraphIncomplete, "Cargo executable path is malformed", nil)
				}
				path = &candidate
			}
			if path != nil {
				if !cargoPackageIDMatches(event.PackageID, selectedPackage) || target.Name != selection.Binary || !reflect.DeepEqual(target.Kind, []string{"bin"}) || executable != "" || !contained(targetRoot, *path) {
					return nil, "", fail(CodeGraphIncomplete, "Cargo executable artifact differs from selected package/bin/target", nil)
				}
				validatedPath := *path
				if evidenceExecutable != "" {
					validatedPath = evidenceExecutable
				}
				info, err := os.Lstat(validatedPath)
				if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
					return nil, "", fail(CodeGraphIncomplete, "Cargo executable is absent or non-regular", nil)
				}
				executable, event.Executable = validatedPath, true
			}
		case "build-finished":
			var success bool
			if json.Unmarshal(raw["success"], &success) != nil || !success || finished {
				return nil, "", fail(CodeOfflineRebuildFailed, "Cargo build-finished event is unsuccessful or duplicated", nil)
			}
			finished = true
		default:
			return nil, "", fail(CodeGraphIncomplete, "Cargo event reason is unsupported", map[string]string{"reason": reason})
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	if executable == "" || !finished {
		return nil, "", fail(CodeOfflineRebuildFailed, "Cargo did not report the selected executable and successful completion", nil)
	}
	return events, executable, nil
}

func selectedMetadataPackage(metadata Metadata, selection SelectionContext) (MetadataPackage, error) {
	selected := MetadataPackage{}
	for _, pkg := range metadata.Packages {
		if pkg.Name == selection.Package && pkg.Source == "" {
			if selected.ID != "" {
				return MetadataPackage{}, fail(CodeGraphIncomplete, "selected package identity is duplicated", nil)
			}
			selected = pkg
		}
	}
	if selected.ID == "" {
		return MetadataPackage{}, fail(CodeGraphIncomplete, "selected package identity is missing", nil)
	}
	return selected, nil
}

func cargoPackageIDMatches(observed string, expected MetadataPackage) bool {
	if observed == expected.ID {
		return true
	}
	if expected.Source != "" {
		return false
	}
	prefix, fragment, ok := strings.Cut(observed, "#")
	if !ok {
		return false
	}
	if fragment == expected.Name+"@"+expected.Version {
		return true
	}
	if fragment != expected.Version || !strings.HasPrefix(prefix, "path+") {
		return false
	}
	parsed, err := url.Parse(strings.TrimPrefix(prefix, "path+"))
	return err == nil && filepath.Base(parsed.Path) == expected.Name
}

func publicationSlots(authority closuregraph.PublicationEvidence) (closuregraph.ID, closuregraph.ExpectedCacheInput, closuregraph.Node, closuregraph.Node, closuregraph.Edge, error) {
	if err := authority.Graph.Validate(); err != nil {
		return "", closuregraph.ExpectedCacheInput{}, closuregraph.Node{}, closuregraph.Node{}, closuregraph.Edge{}, err
	}
	if err := authority.Plan.Validate(); err != nil || len(authority.Plan.DeclaredOutputNodeIDs) != 1 || len(authority.Plan.ActionNodeIDs) != 1 {
		return "", closuregraph.ExpectedCacheInput{}, closuregraph.Node{}, closuregraph.Node{}, closuregraph.Edge{}, fail(CodeGraphIncomplete, "C5 must declare exactly one Rust executable output", nil)
	}
	closureID, err := authority.Closure.ID()
	if err != nil {
		return "", closuregraph.ExpectedCacheInput{}, closuregraph.Node{}, closuregraph.Node{}, closuregraph.Edge{}, err
	}
	expected := closuregraph.ExpectedCacheInput{SchemaID: closuregraph.SchemaExpectedCacheInput, ClosureID: closureID, ExpectedOutputNodeIDs: append([]closuregraph.ID(nil), authority.Plan.DeclaredOutputNodeIDs...)}
	outputID := expected.ExpectedOutputNodeIDs[0]
	var output, action closuregraph.Node
	var produces closuregraph.Edge
	produceCount := 0
	for _, node := range authority.Graph.Records.CaptureNodes {
		id := mustNodeID(node)
		if id == outputID {
			output = node
		}
		for _, actionID := range authority.Plan.ActionNodeIDs {
			if id == actionID {
				action = node
			}
		}
	}
	for _, edge := range authority.Graph.Records.CaptureEdges {
		if edge.Kind == closuregraph.EdgeProduces && edge.FromNodeID == mustNodeID(action) && edge.ToNodeID == outputID {
			produces = edge
			produceCount++
		}
	}
	if output.Kind != closuregraph.NodeOutputArtifact || action.Kind != closuregraph.NodeAction || produces.Kind != closuregraph.EdgeProduces || produceCount != 1 {
		return "", closuregraph.ExpectedCacheInput{}, closuregraph.Node{}, closuregraph.Node{}, closuregraph.Edge{}, fail(CodeGraphReferenceInvalid, "C5 output/action/produces slots do not resolve uniquely", nil)
	}
	return closureID, expected, output, action, produces, nil
}

func stringValues(values []string) []any {
	result := make([]any, len(values))
	for i := range values {
		result[i] = values[i]
	}
	return result
}
func cargoEventValues(events []CargoEvent) []any {
	result := make([]any, len(events))
	for i, event := range events {
		result[i] = map[string]any{"executable": event.Executable, "package_id": event.PackageID, "reason": event.Reason, "target_kinds": stringValues(event.TargetKinds), "target_name": event.TargetName}
	}
	return result
}
func mustNodeID(node closuregraph.Node) closuregraph.ID { id, _ := node.ID(); return id }
func mustEdgeID(edge closuregraph.Edge) closuregraph.ID { id, _ := edge.ID(); return id }
