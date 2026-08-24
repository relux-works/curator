package swiftpmsource

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
)

// ExecutorSwiftPM is the production manifest evaluator backed by the shared
// commit-before-start derivation executor. The selected swift executable must
// already be staged at Tool.ExecutableRelativePath below ExecutionRoot.
type ExecutorSwiftPM struct {
	Executor         *closureexec.Executor
	ExecutionRoot    string
	OutputRoot       string
	Tool             ToolIdentity
	Recheck          func(context.Context, ToolIdentity) (closureexec.ToolchainIdentity, error)
	AllowedProcesses []string
}

// Evaluate implements ManifestEvaluator using the exact argv committed by the
// adapter. It never accepts evaluator-authored identity or receipt evidence.
func (runtime *ExecutorSwiftPM) Evaluate(ctx context.Context, root string, sourcePermit ManifestPermit) (ManifestResult, error) {
	if runtime == nil || runtime.Executor == nil || runtime.Recheck == nil || sourcePermit.input.Tree == nil {
		return ManifestResult{}, fail(CodeDerivationUnauthorized, "executor-backed manifest runtime is incomplete")
	}
	payload, receipt, err := runtime.run(ctx, executorCall{
		key: "swiftpm-manifest:" + sourcePermit.PackageIdentity + ":" + string(sourcePermit.ID), subtype: closureexec.DerivationManifest,
		inputs: []executorInput{{input: sourcePermit.input, receiptID: sourcePermit.IntakeReceiptID, mount: "inputs/package"}}, workReceiptID: sourcePermit.IntakeReceiptID, c0ID: sourcePermit.C0CheckpointID,
		toolNodeID: sourcePermit.ToolchainNodeID, hostID: sourcePermit.HostID, targetID: sourcePermit.TargetID,
		argv: sourcePermit.Argv, environment: sourcePermit.Environment, evidencePath: "evidence/manifest.json", evidenceSchema: ManifestSchemaID,
		workPath: "work/package", stdoutEvidence: true,
	})
	if err != nil {
		return ManifestResult{}, err
	}
	manifest, err := decodeDumpPackage(payload, root)
	if err != nil {
		return ManifestResult{}, err
	}
	manifest.SelectedManifest = sourcePermit.SelectedManifest
	receiptID, err := receipt.ID()
	if err != nil {
		return ManifestResult{}, err
	}
	return ManifestResult{Manifest: manifest, ReceiptID: receiptID}, nil
}

// Replay implements the production offline metadata boundary. It freezes the
// accepted lock into a newly admitted root derivative, mounts every admitted
// mirror read-only, and executes the exact supported SwiftPM metadata argv.
func (runtime *ExecutorSwiftPM) Replay(ctx context.Context, capture *Capture) (OfflineMetadataResult, error) {
	if runtime == nil || capture == nil || capture.config.Store == nil || capture.config.Policy == nil {
		return OfflineMetadataResult{}, fail(CodeDerivationUnauthorized, "executor-backed offline replay authority is incomplete")
	}
	temporary, err := os.MkdirTemp(filepath.Dir(runtime.ExecutionRoot), "swiftpm-offline-replay-")
	if err != nil {
		return OfflineMetadataResult{}, err
	}
	defer func() { _ = os.RemoveAll(temporary) }()
	rootSource, err := capture.rootInput.Tree.ProtectedPath()
	if err != nil {
		return OfflineMetadataResult{}, err
	}
	replayRoot := filepath.Join(temporary, "root")
	if err = copyRegularTree(rootSource, replayRoot); err != nil {
		return OfflineMetadataResult{}, err
	}
	if err = os.WriteFile(filepath.Join(replayRoot, "Package.resolved"), capture.Lock.Bytes, 0o600); err != nil {
		return OfflineMetadataResult{}, err
	}
	type mirrorEntry struct {
		Original string `json:"original"`
		Mirror   string `json:"mirror"`
	}
	entries := make([]mirrorEntry, len(capture.Mirrors))
	for index, mirror := range capture.Mirrors {
		mount, mountErr := mirrorMount(mirror.Identity)
		if mountErr != nil {
			return OfflineMetadataResult{}, mountErr
		}
		target := filepath.Join(runtime.ExecutionRoot, filepath.FromSlash(mount))
		if mirror.LocalKind == SourceRemote {
			target = "file://" + target
		}
		entries[index] = mirrorEntry{Original: mirror.Original, Mirror: target}
	}
	mirrorPayload, err := json.Marshal(map[string]any{"object": entries, "version": 1})
	if err != nil {
		return OfflineMetadataResult{}, err
	}
	configRoot := filepath.Join(replayRoot, ".curator", "config")
	if err = os.MkdirAll(configRoot, 0o700); err != nil {
		return OfflineMetadataResult{}, err
	}
	if err = os.WriteFile(filepath.Join(configRoot, "mirrors.json"), mirrorPayload, 0o600); err != nil {
		return OfflineMetadataResult{}, err
	}
	rootInput, _, _, _, err := admitTree(ctx, capture.config, "offline-root", replayRoot, "swiftpm-offline-root", string(capture.GraphDigest), capture.Lock.Digest)
	if err != nil {
		return OfflineMetadataResult{}, err
	}
	inputs := []executorInput{{input: rootInput, receiptID: mustReceiptID(rootInput), mount: "inputs/root"}}
	for _, mirror := range capture.Mirrors {
		mount, _ := mirrorMount(mirror.Identity)
		inputs = append(inputs, executorInput{input: mirror.input, receiptID: mirror.MirrorIntakeReceiptID, mount: mount})
	}
	c0ID, err := capture.C0.ID()
	if err != nil {
		return OfflineMetadataResult{}, err
	}
	toolNode, _ := toolNodeRecord(capture.config.Toolchain.SwiftPM)
	toolNodeID, _ := toolNode.ID()
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "swiftpm.platform.target", Payload: capture.config.Destination.Platform}
	platformID, _ := platform.ID()
	argv := offlineMetadataArgv()
	payload, receipt, err := runtime.run(ctx, executorCall{
		key: "swiftpm-offline-metadata:" + string(capture.GraphDigest), subtype: closureexec.DerivationMetadata,
		inputs: inputs, workReceiptID: mustReceiptID(rootInput), c0ID: c0ID, toolNodeID: toolNodeID, hostID: platformID, targetID: platformID,
		argv: argv, environment: map[string]string{"HOME": "work/package/.curator/home", "SWIFTPM_CONFIG_DIR": "work/package/.curator/config", "SWIFTPM_SECURITY_DIR": "work/package/.curator/security", "TZ": "UTC"}, evidencePath: "evidence/show-dependencies.json", evidenceSchema: "swiftpm-show-dependencies-v1", workPath: "work/package", stdoutEvidence: true,
	})
	if err != nil {
		return OfflineMetadataResult{}, fail(CodeOfflineReplayFailed, "protected SwiftPM show-dependencies replay failed: %v", err)
	}
	identities, err := decodeDependencyIdentities(payload)
	if err != nil {
		return OfflineMetadataResult{}, err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return OfflineMetadataResult{}, err
	}
	return OfflineMetadataResult{ReceiptID: receiptID, PackageIdentities: identities}, nil
}

func mirrorMount(identity string) (string, error) {
	if identity == "" || identity == "." || identity == ".." || filepath.Base(identity) != identity || strings.ContainsAny(identity, "\\\x00\r\n") {
		return "", failFields(CodeDependencyPinMismatch, map[string]string{"identity": identity}, "package identity cannot name an isolated mirror mount")
	}
	return "inputs/mirrors/" + identity, nil
}

func offlineMetadataArgv() []string {
	return []string{"package", "--package-path", ".", "--cache-path", ".curator/cache", "--config-path", ".curator/config", "--security-path", ".curator/security", "--scratch-path", ".curator/scratch", "--disable-netrc", "--disable-experimental-prebuilts", "--force-resolved-versions", "show-dependencies", "--format", "json"}
}

func mustReceiptID(input closureexec.AdmittedInput) closuregraph.ID {
	id, _ := input.Receipt.ID()
	return id
}

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
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeOfflineReplayFailed, "offline root contains a link")
		}
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o700)
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return fail(CodeOfflineReplayFailed, "offline root contains a special node")
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir confines the admitted root.
		if err != nil {
			return err
		}
		return os.WriteFile(destination, payload, 0o600)
	})
}

type dependencyEnvelope struct {
	Identity     string               `json:"identity"`
	Dependencies []dependencyEnvelope `json:"dependencies"`
}

func decodeDependencyIdentities(payload []byte) ([]string, error) {
	var root dependencyEnvelope
	if err := json.Unmarshal(payload, &root); err != nil || root.Identity == "" {
		return nil, fail(CodeBuildGraphDrift, "show-dependencies JSON is malformed")
	}
	seen := map[string]bool{}
	var walk func([]dependencyEnvelope) error
	walk = func(items []dependencyEnvelope) error {
		for _, item := range items {
			identity := strings.ToLower(item.Identity)
			if identity == "" || seen[identity] {
				return fail(CodeBuildGraphDrift, "show-dependencies contains a missing or duplicate package identity")
			}
			seen[identity] = true
			if err := walk(item.Dependencies); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root.Dependencies); err != nil {
		return nil, err
	}
	identities := make([]string, 0, len(seen))
	for identity := range seen {
		identities = append(identities, identity)
	}
	sort.Strings(identities)
	return identities, nil
}

type executorCall struct {
	key, workPath, evidencePath, evidenceSchema string
	subtype                                     closureexec.DerivationKind
	inputs                                      []executorInput
	workReceiptID                               closuregraph.ID
	c0ID, toolNodeID, hostID, targetID          closuregraph.ID
	argv                                        []string
	environment                                 map[string]string
	stdoutEvidence                              bool
}

type executorInput struct {
	input     closureexec.AdmittedInput
	receiptID closuregraph.ID
	mount     string
}

func (runtime *ExecutorSwiftPM) run(ctx context.Context, call executorCall) ([]byte, closureexec.DerivationReceipt, error) {
	if err := validateRuntimeRoots(runtime.ExecutionRoot, runtime.OutputRoot); err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	if err := os.RemoveAll(runtime.OutputRoot); err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	operation, err := runtime.Executor.Preflight(ctx)
	if err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 32 << 20, ReadBytes: 2 << 30, WriteBytes: 512 << 20, WallTimeMillis: 120_000, ProcessCount: 64}
	limitID, err := limits.ID()
	if err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	artifactID, err := closuregraph.DomainID("swiftpm-executor-evidence-v1", map[string]any{"invocation_key": call.key, "path": call.evidencePath, "schema": call.evidenceSchema})
	if err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	requirements := []closureexec.EvidenceRequirement{{Path: call.evidencePath, SchemaID: call.evidenceSchema, ArtifactManifestID: artifactID}}
	evidenceID, err := swiftPMEvidenceSchemaID(requirements)
	if err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	environment := cloneMap(call.environment)
	environment["CURATOR_OUTPUT_ROOT"] = "output"
	environment["PATH"] = filepath.Join(runtime.ExecutionRoot, "bin")
	if call.workPath != "" {
		environment["HOME"] = call.workPath + "/.curator/home"
		environment["SWIFTPM_CACHE_DIR"] = call.workPath + "/.curator/cache"
		environment["SWIFTPM_CONFIG_DIR"] = call.workPath + "/.curator/config"
		environment["SWIFTPM_SCRATCH_DIR"] = call.workPath + "/.curator/scratch"
		environment["SWIFTPM_SECURITY_DIR"] = call.workPath + "/.curator/security"
	}
	if len(call.inputs) == 0 {
		return nil, closureexec.DerivationReceipt{}, fail(CodeDerivationUnauthorized, "SwiftPM executor has no admitted inputs")
	}
	sort.Slice(call.inputs, func(i, j int) bool { return call.inputs[i].receiptID < call.inputs[j].receiptID })
	reads := make([]string, len(call.inputs))
	inputIDs := make([]closuregraph.ID, len(call.inputs))
	mounts := make([]closureexec.InputMount, len(call.inputs))
	admitted := make(map[closuregraph.ID]closureexec.AdmittedInput, len(call.inputs))
	for index, input := range call.inputs {
		reads[index] = input.mount
		inputIDs[index] = input.receiptID
		mounts[index] = closureexec.InputMount{ReceiptID: input.receiptID, Path: input.mount}
		admitted[input.receiptID] = input.input
	}
	sort.Strings(reads)
	writes := []string{call.evidencePath}
	workCopies := []closureexec.WorkCopy{}
	cwd := call.inputs[0].mount
	if call.workPath != "" {
		writes = append(writes, call.workPath)
		workCopies = append(workCopies, closureexec.WorkCopy{ReceiptID: call.workReceiptID, Path: call.workPath})
		cwd = call.workPath
	}
	sort.Strings(writes)
	processes := append([]string(nil), runtime.AllowedProcesses...)
	if len(processes) == 0 {
		processes = []string{runtime.Tool.ExecutableRelativePath}
	}
	sort.Strings(processes)
	stdoutPath := ""
	if call.stdoutEvidence {
		stdoutPath = call.evidencePath
	}
	permit := closureexec.DerivationPermit{
		SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: operation.CurrentCausalHead(), InvocationKey: call.key, InvocationSubtype: call.subtype,
		AdmittedInputReceiptIDs: inputIDs, InputMounts: mounts,
		C0CheckpointID: call.c0ID, ToolchainNodeID: call.toolNodeID, ToolchainFingerprint: runtime.Tool.Fingerprint, ExecutableSHA256: runtime.Tool.ExecutableSHA256,
		Executable: runtime.Tool.ExecutableRelativePath, CWD: cwd, Argv: append([]string(nil), call.argv...), Environment: environment,
		HostID: call.hostID, TargetID: call.targetID, AllowedProcesses: processes, ReadRoots: reads, WriteRoots: writes,
		WorkCopies:         workCopies,
		StdoutEvidencePath: stdoutPath, ExpectedEvidence: requirements, Network: "none", RecheckRule: "immediate-exact-v1",
		ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID,
	}
	permitID, err := operation.Commit(permit)
	if err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	receipt, err := operation.Execute(ctx, permitID, func(recheckCtx context.Context) (closureexec.ToolchainIdentity, error) {
		return runtime.Recheck(recheckCtx, runtime.Tool)
	}, admitted)
	if err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	if err = operation.VerifyIssuedDerivationReceipt(receipt); err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	evidence := filepath.Join(runtime.OutputRoot, filepath.FromSlash(call.evidencePath))
	payload, err := os.ReadFile(evidence) // #nosec G304 -- fixed declared evidence path below validated runtime output root.
	if err != nil {
		return nil, closureexec.DerivationReceipt{}, err
	}
	return payload, receipt, nil
}

func validateRuntimeRoots(executionRoot, outputRoot string) error {
	execution, err := cleanAbsoluteRoot(executionRoot)
	if err != nil {
		return err
	}
	output, err := filepath.Abs(outputRoot)
	if err != nil || output == execution {
		return fmt.Errorf("SwiftPM executor output root is invalid")
	}
	relative, err := filepath.Rel(execution, output)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("SwiftPM executor output root escapes execution root")
	}
	return nil
}

func swiftPMEvidenceSchemaID(requirements []closureexec.EvidenceRequirement) (closuregraph.ID, error) {
	values := make([]any, len(requirements))
	for index, requirement := range requirements {
		values[index] = map[string]any{"artifact_manifest_id": string(requirement.ArtifactManifestID), "path": requirement.Path, "schema_id": requirement.SchemaID}
	}
	return closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": values})
}

type dumpPackageEnvelope struct {
	Name         string           `json:"name"`
	Dependencies []map[string]any `json:"dependencies"`
	Products     []dumpProduct    `json:"products"`
	Targets      []dumpTarget     `json:"targets"`
	ToolsVersion struct {
		Version string `json:"_version"`
	} `json:"toolsVersion"`
	Traits []struct {
		Name string `json:"name"`
	} `json:"traits"`
}

type dumpProduct struct {
	Name    string                     `json:"name"`
	Targets []string                   `json:"targets"`
	Type    map[string]json.RawMessage `json:"type"`
}

type dumpTarget struct {
	Name              string                       `json:"name"`
	Type              string                       `json:"type"`
	Path              string                       `json:"path"`
	PublicHeadersPath string                       `json:"publicHeadersPath"`
	Sources           []string                     `json:"sources"`
	Exclude           []string                     `json:"exclude"`
	Dependencies      []map[string]json.RawMessage `json:"dependencies"`
	Settings          []map[string]json.RawMessage `json:"settings"`
}

func decodeDumpPackage(payload []byte, root string) (Manifest, error) {
	var envelope dumpPackageEnvelope
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		// SwiftPM adds fields across compatible releases. Decode the closed fields
		// again while retaining explicit validation of every consumed variant.
		if err = json.Unmarshal(payload, &envelope); err != nil {
			return Manifest{}, fail(CodeManifestReplayDrift, "dump-package JSON is malformed: %v", err)
		}
	}
	manifest := Manifest{PackageName: envelope.Name, ToolsVersion: envelope.ToolsVersion.Version}
	for _, product := range envelope.Products {
		kind, err := singletonKey(product.Type)
		if err != nil {
			return Manifest{}, err
		}
		manifest.Products = append(manifest.Products, Product{Name: product.Name, Type: kind, Targets: product.Targets})
	}
	for _, target := range envelope.Targets {
		sources, err := enumerateTargetSources(root, target)
		if err != nil {
			return Manifest{}, err
		}
		item := Target{Name: target.Name, Type: target.Type, Path: target.Path, PublicHeadersPath: target.PublicHeadersPath, Sources: sources}
		for _, raw := range target.Dependencies {
			dependency, err := decodeTargetDependency(raw)
			if err != nil {
				return Manifest{}, err
			}
			item.Dependencies = append(item.Dependencies, dependency)
		}
		for _, raw := range target.Settings {
			setting, err := decodeBuildSetting(raw)
			if err != nil {
				return Manifest{}, err
			}
			item.Settings = append(item.Settings, setting)
		}
		manifest.Targets = append(manifest.Targets, item)
	}
	for _, raw := range envelope.Dependencies {
		dependency, err := decodeManifestDependency(raw, root)
		if err != nil {
			return Manifest{}, err
		}
		manifest.Dependencies = append(manifest.Dependencies, dependency)
	}
	for _, trait := range envelope.Traits {
		if trait.Name == "" {
			return Manifest{}, fail(CodeManifestReplayDrift, "manifest trait is unnamed")
		}
		manifest.Traits = append(manifest.Traits, trait.Name)
	}
	return manifest, nil
}

func singletonKey(value map[string]json.RawMessage) (string, error) {
	if len(value) != 1 {
		return "", fail(CodeManifestReplayDrift, "SwiftPM product type is unsupported")
	}
	for key := range value {
		return key, nil
	}
	return "", fail(CodeManifestReplayDrift, "SwiftPM product type is absent")
}

func enumerateTargetSources(root string, target dumpTarget) ([]string, error) {
	targetPath := targetSourceRoot(target.Name, target.Type, target.Path)
	if filepath.IsAbs(targetPath) || filepath.Clean(targetPath) != filepath.FromSlash(targetPath) || targetPath == ".." || strings.HasPrefix(targetPath, "../") {
		return nil, fail(CodeSourceInventoryDrift, "target path escapes package")
	}
	excluded := map[string]bool{}
	for _, value := range target.Exclude {
		excluded[filepath.ToSlash(filepath.Clean(filepath.Join(targetPath, value)))] = true
	}
	requested := map[string]bool{}
	for _, value := range target.Sources {
		requested[filepath.ToSlash(filepath.Clean(filepath.Join(targetPath, value)))] = true
	}
	rootPath := filepath.Join(root, filepath.FromSlash(targetPath))
	result := []string{}
	err := filepath.WalkDir(rootPath, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) && current == rootPath && target.Type == "system" {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		logical := filepath.ToSlash(relative)
		if excluded[logical] || len(requested) > 0 && !requested[logical] || !swiftPMSourceExtension(filepath.Ext(logical)) {
			return nil
		}
		result = append(result, logical)
		return nil
	})
	if err != nil {
		return nil, fail(CodeSourceInventoryDrift, "target source enumeration failed: %v", err)
	}
	sort.Strings(result)
	return result, nil
}

// swiftPMSourceExtension admits the target-source suffixes SwiftPM enumerates.
// Admission stays case-insensitive because the driver compiles `impl.C` and
// `impl.M` as readily as `impl.c` and `impl.m` — it selects a *different*
// language for the upper-case pair, which targetLanguages and the interop
// stage record, but the bytes are target source either way and must be
// admitted, hashed, and inventoried rather than silently dropped.
func swiftPMSourceExtension(extension string) bool {
	switch strings.ToLower(extension) {
	case ".swift", ".c", ".cc", ".cpp", ".cxx", ".m", ".mm", ".s":
		return true
	default:
		return false
	}
}

func decodeTargetDependency(raw map[string]json.RawMessage) (TargetDependency, error) {
	if len(raw) != 1 {
		return TargetDependency{}, fail(CodeManifestReplayDrift, "target dependency shape is unsupported")
	}
	for kind, payload := range raw {
		var values []json.RawMessage
		if err := json.Unmarshal(payload, &values); err != nil || len(values) == 0 {
			return TargetDependency{}, fail(CodeManifestReplayDrift, "target dependency payload is unsupported")
		}
		var name string
		if err := json.Unmarshal(values[0], &name); err != nil || name == "" {
			return TargetDependency{}, fail(CodeManifestReplayDrift, "target dependency name is invalid")
		}
		dependency := TargetDependency{}
		switch kind {
		case "target", "byName":
			dependency.Name = name
		case "product":
			dependency.Product = name
			if len(values) > 1 && string(values[1]) != "null" {
				_ = json.Unmarshal(values[1], &dependency.Package)
			}
		default:
			return TargetDependency{}, fail(CodeManifestReplayDrift, "target dependency kind %s is unsupported", kind)
		}
		return dependency, nil
	}
	return TargetDependency{}, fail(CodeManifestReplayDrift, "target dependency is absent")
}

func decodeBuildSetting(raw map[string]json.RawMessage) (BuildSetting, error) {
	payload, err := json.Marshal(raw)
	if err != nil {
		return BuildSetting{}, err
	}
	setting := BuildSetting{Kind: "swiftpm-setting", Value: string(payload), Unsafe: bytesContainJSONKey(payload, "unsafeFlags")}
	return setting, nil
}

func bytesContainJSONKey(payload []byte, key string) bool {
	return strings.Contains(string(payload), `"`+key+`"`)
}

func decodeManifestDependency(raw map[string]any, root string) (ManifestDependency, error) {
	if values, ok := raw["fileSystem"].([]any); ok && len(values) == 1 {
		item, ok := values[0].(map[string]any)
		if !ok {
			return ManifestDependency{}, fail(CodeDependencyOriginUnsupported, "SwiftPM file-system dependency is malformed")
		}
		identity, _ := item["identity"].(string)
		location, _ := item["path"].(string)
		relative, err := filepath.Rel(root, location)
		if identity == "" || location == "" || err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return ManifestDependency{}, failFields(CodeLocalDependencyOutside, map[string]string{"identity": identity}, "local dependency escapes package")
		}
		return ManifestDependency{Identity: strings.ToLower(identity), Kind: SourcePath, LocalPath: filepath.ToSlash(relative), Requirement: "path"}, nil
	}
	values, ok := raw["sourceControl"].([]any)
	if !ok || len(values) != 1 {
		return ManifestDependency{}, fail(CodeDependencyOriginUnsupported, "SwiftPM dependency kind is unsupported")
	}
	item, ok := values[0].(map[string]any)
	if !ok {
		return ManifestDependency{}, fail(CodeDependencyOriginUnsupported, "SwiftPM source-control dependency is malformed")
	}
	identity, _ := item["identity"].(string)
	locations, ok := item["location"].(map[string]any)
	if identity == "" || !ok || len(locations) != 1 {
		return ManifestDependency{}, fail(CodeDependencyOriginUnsupported, "SwiftPM source-control identity or location is malformed")
	}
	kind := SourceRemote
	var location string
	for locationKind, rawLocations := range locations {
		list, listOK := rawLocations.([]any)
		if !listOK || len(list) != 1 {
			return ManifestDependency{}, fail(CodeDependencyOriginUnsupported, "SwiftPM source-control location is malformed")
		}
		location, _ = list[0].(string)
		if encoded, encodedOK := list[0].(map[string]any); encodedOK {
			location, _ = encoded["urlString"].(string)
			if location == "" {
				location, _ = encoded["pathString"].(string)
			}
		}
		if location == "" {
			return ManifestDependency{}, fail(CodeDependencyOriginUnsupported, "SwiftPM source-control location value is malformed")
		}
		switch locationKind {
		case "remote":
			kind = SourceRemote
		case "local":
			kind = SourceLocal
		default:
			return ManifestDependency{}, fail(CodeDependencyOriginUnsupported, "SwiftPM source-control location kind is unsupported")
		}
	}
	requirement, err := decodeRequirement(item["requirement"])
	if err != nil {
		return ManifestDependency{}, err
	}
	return ManifestDependency{Identity: strings.ToLower(identity), Kind: kind, Location: location, Requirement: requirement}, nil
}

func decodeRequirement(raw any) (string, error) {
	value, ok := raw.(map[string]any)
	if !ok || len(value) != 1 {
		return "", fail(CodeDependencyOriginUnsupported, "SwiftPM requirement is malformed")
	}
	for kind, rawValues := range value {
		values, ok := rawValues.([]any)
		if !ok || len(values) == 0 {
			return "", fail(CodeDependencyOriginUnsupported, "SwiftPM requirement values are malformed")
		}
		parts := make([]string, len(values))
		for index, item := range values {
			parts[index], ok = item.(string)
			if !ok || parts[index] == "" {
				return "", fail(CodeDependencyOriginUnsupported, "SwiftPM requirement value is malformed")
			}
		}
		switch kind {
		case "exact", "revision", "branch":
			if len(parts) != 1 {
				return "", fail(CodeDependencyOriginUnsupported, "SwiftPM %s requirement is malformed", kind)
			}
			return kind + ":" + parts[0], nil
		case "range":
			if len(parts) != 2 {
				return "", fail(CodeDependencyOriginUnsupported, "SwiftPM range requirement is malformed")
			}
			return "range:" + parts[0] + "..<" + parts[1], nil
		default:
			return "", fail(CodeDependencyOriginUnsupported, "SwiftPM requirement kind %s is unsupported", kind)
		}
	}
	return "", fail(CodeDependencyOriginUnsupported, "SwiftPM requirement is absent")
}
