package artifactpolicy

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type inspector struct {
	ctx               context.Context
	role              TrustRole
	roleAuthorized    bool
	descriptor        Descriptor
	limits            LimitVector
	account           *limitAccountant
	store             *blobStore
	nodes             []ManifestNode
	findings          *findingAccumulator
	findingErr        error
	authorizationSeal *authorizationSeal
}

// AdmitDependency recursively inspects immutable package bytes under the
// dependency_input role. A policy rejection returns both canonical evidence
// and a PolicyError; no admission token is returned.
func (service *Service) AdmitDependency(ctx context.Context, request DependencyRequest) (Result, error) {
	return service.inspectPayload(ctx, RoleDependencyInput, request.Descriptor, request.Payload, nil, nil)
}

// AdmitToolchain inspects an independently selected and fingerprinted
// toolchain component. It cannot be called with dependency role metadata.
func (service *Service) AdmitToolchain(ctx context.Context, request ToolchainRequest) (Result, error) {
	return service.inspectPayload(ctx, RoleExternalToolchain, request.Descriptor, request.Payload, request.Authorization, nil)
}

// AdmitLocalOutput inspects a causally observed build output and reconciles it
// with its independently derived protected receipt expectation.
func (service *Service) AdmitLocalOutput(ctx context.Context, request LocalOutputRequest) (Result, error) {
	return service.inspectPayload(ctx, RoleLocalBuildOutput, request.Descriptor, request.Payload, nil, request.Authorization)
}

// AdmitVerifiedBinary always rejects because verified-binary-v1 is not a
// current Curator capability.
func (service *Service) AdmitVerifiedBinary(ctx context.Context, request VerifiedBinaryRequest) (Result, error) {
	return service.inspectPayload(ctx, RoleVerifiedBinaryCandidate, request.Descriptor, request.Payload, nil, nil)
}

func (service *Service) inspectPayload(
	ctx context.Context,
	role TrustRole,
	descriptor Descriptor,
	payload Payload,
	toolchain ToolchainAuthorization,
	output LocalOutputAuthorization,
) (Result, error) {
	configured, err := configuredService(service)
	if err != nil {
		return Result{}, err
	}
	service = configured
	if err := validateDescriptor(descriptor); err != nil {
		return service.failureBeforeCapture(role, descriptor, payload, CodePolicyInternalError, "invalid_descriptor:"+err.Error())
	}
	rootPath, pathErr := validateVirtualPath(payload.Path, service.limits)
	if pathErr != nil {
		return service.failureBeforeCapture(role, descriptor, payload, CodeArchiveUnsafePath, pathErr.Error())
	}
	store, err := newBlobStore()
	if err != nil {
		return service.failureBeforeCapture(role, descriptor, payload, CodeInspectionUnavailable, "create_private_inspection_store")
	}
	defer store.close()
	captured, err := store.captureRoot(ctx, payload, service.limits)
	if err != nil {
		code := CodeInspectionUnavailable
		var limit *limitFailure
		if errorAs(err, &limit) {
			code = CodeInspectionLimitExceeded
		}
		return service.failureBeforeCapture(role, descriptor, payload, code, err.Error())
	}
	account, err := newLimitAccountant(service.limits, captured.size)
	if err != nil {
		return service.failureWithCaptured(role, descriptor, payload.Path, captured, err)
	}
	worker := &inspector{
		ctx: ctx, role: role, roleAuthorized: true, descriptor: descriptor,
		limits: service.limits, account: account, store: store,
		findings:          newFindingAccumulator(service.limits.MaxRecordedFindings),
		authorizationSeal: service.authorizationSeal,
	}
	roleEvidence := []Fact{}
	outputExpectation := ArtifactExpectation{}
	if role == RoleDependencyInput {
		roleEvidence = dependencyRoleFacts(descriptor.Origin)
		if reason := validateOrigin(descriptor.Origin, captured); reason != "" {
			worker.addDiagnostic(Diagnostic{
				Code: CodeOriginUnverified, Path: rootPath.Canonical,
				OriginalNameBase64: originalNameBase64(payload.Path),
				CollisionKey:       rootPath.CollisionKey, SHA256: captured.sha256,
				Size: captured.size, Reason: reason,
			})
		}
	}
	switch role {
	case RoleExternalToolchain:
		code, reason, evidence := validateToolchainAuthorization(toolchain, rootPath.Canonical, captured, service.authorizationSeal)
		roleEvidence = evidence
		if code != "" {
			worker.roleAuthorized = false
			worker.addDiagnostic(Diagnostic{
				Code: code, Path: rootPath.Canonical,
				OriginalNameBase64: originalNameBase64(payload.Path),
				CollisionKey:       rootPath.CollisionKey, SHA256: captured.sha256,
				Size: captured.size, Reason: reason,
			})
		}
	case RoleLocalBuildOutput:
		code, reason, evidence, expectation := validateLocalOutputAuthorization(output, rootPath.Canonical, captured, service.authorizationSeal)
		roleEvidence = evidence
		outputExpectation = expectation
		if code != "" {
			worker.roleAuthorized = false
			worker.addDiagnostic(Diagnostic{
				Code: code, Path: rootPath.Canonical,
				OriginalNameBase64: originalNameBase64(payload.Path),
				CollisionKey:       rootPath.CollisionKey, SHA256: captured.sha256,
				Size: captured.size, Reason: reason,
			})
		}
	case RoleVerifiedBinaryCandidate:
		roleEvidence = verifiedBinaryRoleFacts()
		worker.roleAuthorized = false
		worker.addDiagnostic(Diagnostic{
			Code: CodeBinaryAdmissionUnavailable, Path: rootPath.Canonical,
			OriginalNameBase64: originalNameBase64(payload.Path),
			CollisionKey:       rootPath.CollisionKey, SHA256: captured.sha256,
			Size: captured.size, Reason: "verified_binary_v1_unavailable",
		})
	}

	rejected := worker.inspectBlob(
		rootPath.Canonical, payload.Path, "", nil, captured, 1, 0, nil,
	)
	if role == RoleLocalBuildOutput && output != nil && worker.roleAuthorized {
		rootNode := worker.rootNode(rootPath.Canonical)
		if reason := validateLocalOutputExpectation(outputExpectation, rootNode); reason != "" {
			worker.roleAuthorized = false
			worker.addDiagnostic(Diagnostic{
				Code: CodeLocalOutputDrift, Path: rootPath.Canonical,
				OriginalNameBase64: originalNameBase64(payload.Path),
				CollisionKey:       rootPath.CollisionKey, Class: rootNode.Class,
				Variant: rootNode.Variant, SHA256: captured.sha256,
				Size: captured.size, Reason: reason,
			})
			rejected = true
		}
	}
	if !worker.roleAuthorized {
		worker.forceRootReject(rootPath.Canonical, "trust_role_not_authorized")
		rejected = true
	}
	if worker.findingCount() > 0 {
		worker.forceRootReject(rootPath.Canonical, "artifact_finding_rejected")
		rejected = true
	}
	accounting, accountingErr := bindTraversalAccounting(worker.account.snapshot(), "file", worker.nodes)
	if accountingErr != nil {
		return Result{}, fmt.Errorf("bind traversal accounting evidence: %w", accountingErr)
	}
	manifest := Manifest{
		SchemaID: SchemaID, PolicyID: PolicyID, PolicyVersion: PolicyVersion,
		LimitVector: service.limits, DetectorRegistryID: DetectorRegistryID,
		Detectors: append([]DetectorIdentity(nil), service.detectors...),
		AdapterID: descriptor.AdapterID, ProfileID: descriptor.ProfileID,
		Manager: descriptor.Manager, PackageName: descriptor.PackageName,
		PackageVersion: descriptor.PackageVersion, Origin: descriptor.Origin,
		TrustRole: role, RoleEvidence: roleEvidence,
		RawPayload: RawPayloadEvidence{
			Path: rootPath.Canonical, Size: captured.size, SHA256: captured.sha256, Kind: "file",
		},
		Accounting: accounting,
		Nodes:      worker.nodes, Decision: allowDecision(role),
	}
	if rejected || worker.findingCount() > 0 {
		manifest.Decision = DecisionReject
	}
	return finishInspectorResult(manifest, worker)
}

func (inspector *inspector) inspectBlob(
	virtualPath, originalName, parent string,
	chain []string,
	item blob,
	depth, mode int64,
	baseObservations []Observation,
) bool {
	if err := contextError(inspector.ctx); err != nil {
		node := ManifestNode{
			Path: virtualPath, OriginalNameBase64: originalNameBase64(originalName),
			CollisionKey: portableCollisionKey(virtualPath), Kind: NodeRegularFile,
			Parent: parent, ContainerChain: append([]string(nil), chain...),
			Size: item.size, SHA256: item.sha256, Mode: mode,
			Class: ClassOpaqueUnknown, Decision: DecisionReject, Rule: "inspection_cancelled",
			InspectionComplete: false,
		}
		inspector.nodes = append(inspector.nodes, node)
		inspector.addDiagnostic(unavailableDiagnostic(virtualPath, chain, "context_cancelled", err))
		return true
	}
	descriptorPath := descriptorPathForNode(virtualPath, baseObservations)
	uses := append([]UseEdge(nil), inspector.descriptor.ResolvedUses[descriptorPath]...)
	detected := inspector.detect(item, virtualPath, descriptorPath, uses)
	kind := NodeRegularFile
	switch detected.format {
	case formatZIP, formatTar, formatAR:
		kind = NodeArchive
	case formatGZIP:
		kind = NodeCompressedStream
	}
	decision := decisionForClass(inspector.role, detected.class)
	rule := "class_role_decision"
	if detected.format == formatZIP || detected.format == formatTar || detected.format == formatGZIP {
		decision = DecisionDescend
		rule = "recursive_container_descent"
	}
	if detected.diagnostic != "" {
		decision = DecisionReject
		rule = "detector_rejection"
	}
	if detected.format == formatAR && inspector.role == RoleDependencyInput {
		decision = DecisionReject
		rule = "compiled_native_archive"
	}
	// A root archive is governed by the raw-payload limit. Every ordinary root
	// file and every byte-emitting member at a nested layer is also a bounded
	// single leaf at that layer, even when its bytes decode as another container.
	if kind == NodeRegularFile || parent != "" {
		if err := inspector.account.checkLeaf(item.size); err != nil {
			decision = DecisionReject
			detected.diagnostic = CodeInspectionLimitExceeded
			detected.reason = err.Error()
		}
	}
	node := ManifestNode{
		Path: virtualPath, OriginalNameBase64: originalNameBase64(originalName),
		CollisionKey: portableCollisionKey(virtualPath), Kind: kind,
		Parent: parent, ContainerChain: append([]string(nil), chain...),
		Size: item.size, SHA256: item.sha256, Mode: mode,
		DeclaredUses:       uses,
		Observations:       append(append([]Observation(nil), baseObservations...), detected.observations...),
		SelectedDetectorID: detected.detectorID,
		Class:              detected.class, Variant: detected.variant,
		Decision: decision, Rule: rule,
		InspectionComplete: detected.diagnostic == "",
	}
	inspector.nodes = append(inspector.nodes, node)
	nodeIndex := len(inspector.nodes) - 1
	rejected := decision == DecisionReject
	deferNativeArchiveClassFinding := detected.format == formatAR &&
		detected.diagnostic == "" && decision == DecisionReject
	if detected.diagnostic != "" {
		diagnostic := inspector.classDiagnostic(node, detected.reason)
		diagnostic.Code = detected.diagnostic
		diagnostic.DetectorID = detected.detectorID
		inspector.addDiagnostic(diagnostic)
	} else if decision == DecisionReject && !deferNativeArchiveClassFinding {
		inspector.addDiagnostic(inspector.classDiagnostic(node, detected.reason))
	}
	containerRejected := false
	if detected.format != formatNone && detected.format != formatUnsupported {
		containerRejected = inspector.walkContainer(nodeIndex, item, detected.format, depth)
		if containerRejected {
			rejected = true
		}
	}
	if deferNativeArchiveClassFinding && !containerRejected {
		inspector.addDiagnostic(inspector.classDiagnostic(node, detected.reason))
	}
	return rejected
}

func descriptorPathForNode(virtualPath string, observations []Observation) string {
	for _, observation := range observations {
		if observation.Result != "ENTRY" {
			continue
		}
		for _, fact := range observation.Facts {
			if fact.Key == "logical_path" && fact.Value != "" {
				return fact.Value
			}
		}
	}
	return virtualPath
}

func decisionForClass(role TrustRole, class ArtifactClass) Decision {
	if class == ClassArchive || class == ClassCompressedStream || class == ClassDirectory {
		return DecisionDescend
	}
	if class == ClassOpaqueUnknown || class == ClassDataKnownInert || class == ClassLink || class == ClassSpecial {
		return DecisionReject
	}
	switch role {
	case RoleDependencyInput:
		if sourceClass(class) {
			return DecisionAdmitInput
		}
		return DecisionReject
	case RoleExternalToolchain:
		return DecisionAllowToolchain
	case RoleLocalBuildOutput:
		return DecisionAllowOutput
	default:
		return DecisionReject
	}
}

func sourceClass(class ArtifactClass) bool {
	return class == ClassSourceAuthoredText || class == ClassSourceGeneratedText || class == ClassTextMetadata
}

func compiledClass(class ArtifactClass) bool {
	switch class {
	case ClassNativeExecutable, ClassNativeObject, ClassNativeLibraryStatic,
		ClassNativeLibraryDynamic, ClassELFETDYNAmbiguous, ClassAppleFramework,
		ClassAppleXCFramework, ClassNodeExtension, ClassPythonExtension,
		ClassJVMBytecode, ClassPythonBytecode, ClassJavaScriptCodeCache,
		ClassWebAssembly, ClassCompilerSerialized:
		return true
	default:
		return false
	}
}

func (inspector *inspector) classDiagnostic(node ManifestNode, reason string) Diagnostic {
	if reason == "" {
		reason = "class_forbidden_for_trust_role"
	}
	var code DiagnosticCode
	switch {
	case inspector.role == RoleVerifiedBinaryCandidate:
		code = CodeBinaryAdmissionUnavailable
	case compiledClass(node.Class) && inspector.role == RoleDependencyInput:
		code = CodeCompiledDependency
	case node.Class == ClassLink || node.Class == ClassSpecial:
		code = CodeArchiveUnsafeEntry
	case node.Class == ClassOpaqueUnknown || node.Class == ClassDataKnownInert:
		code = CodeOpaqueDependency
	default:
		code = CodePolicyInternalError
	}
	return Diagnostic{
		Code: code, Path: node.Path, OriginalNameBase64: node.OriginalNameBase64,
		CollisionKey: node.CollisionKey, Class: node.Class, Variant: node.Variant,
		ContainerChain: append([]string(nil), node.ContainerChain...),
		SHA256:         node.SHA256, Size: node.Size, Reason: reason,
		DetectorID: node.SelectedDetectorID,
		Details:    observationFacts(node.Observations),
	}
}

func observationFacts(observations []Observation) []Fact {
	result := make([]Fact, 0)
	for _, observation := range observations {
		for _, fact := range observation.Facts {
			result = append(result, Fact{Key: observation.DetectorID + "." + fact.Key, Value: fact.Value})
		}
	}
	sortFacts(result)
	return result
}

func allowDecision(role TrustRole) Decision {
	switch role {
	case RoleDependencyInput:
		return DecisionAdmitInput
	case RoleExternalToolchain:
		return DecisionAllowToolchain
	case RoleLocalBuildOutput:
		return DecisionAllowOutput
	default:
		return DecisionReject
	}
}

func validateDescriptor(descriptor Descriptor) error {
	if descriptor.AdapterID == "" || descriptor.Manager == "" || descriptor.PackageName == "" || descriptor.PackageVersion == "" {
		return fmt.Errorf("adapter, manager, package name, and package version are required")
	}
	if _, ok := profileIDs[descriptor.ProfileID]; !ok {
		return fmt.Errorf("unsupported profile %q", descriptor.ProfileID)
	}
	for pathValue, declaration := range descriptor.DeclaredText {
		if err := validateManifestVirtualPath(pathValue); err != nil {
			return fmt.Errorf("declared text path %q: %w", pathValue, err)
		}
		if _, ok := grammarIDs[declaration.Grammar]; !ok || !textClass(declaration.Class) {
			return fmt.Errorf("declared text path %q has an unsupported grammar or class", pathValue)
		}
	}
	for pathValue, uses := range descriptor.ResolvedUses {
		if err := validateManifestVirtualPath(pathValue); err != nil {
			return fmt.Errorf("resolved use path %q: %w", pathValue, err)
		}
		for _, use := range uses {
			if _, ok := useKinds[use.Kind]; !ok || use.Origin == "" {
				return fmt.Errorf("resolved use path %q has an invalid use edge", pathValue)
			}
		}
	}
	return nil
}

func validateManifestVirtualPath(value string) error {
	return validateManifestVirtualPathWithLimits(value, DefaultLimits())
}

func validateManifestVirtualPathWithLimits(value string, limits LimitVector) error {
	if int64(len(value)) > limits.MaxPathBytes {
		return &pathFailure{
			reason: "path_too_long", limitName: "max_path_bytes",
			limit: limits.MaxPathBytes, observed: int64(len(value)),
		}
	}
	parts := strings.Split(value, "!/")
	if len(parts) == 0 {
		return fmt.Errorf("empty path")
	}
	for _, part := range parts {
		if _, err := validateVirtualPath(part, limits); err != nil {
			return err
		}
	}
	return nil
}

func validateOrigin(origin OriginEvidence, captured blob) string {
	if !origin.Verified || origin.Locator == "" || origin.ImmutableID == "" || origin.LockRecord == "" {
		return "immutable_origin_binding_missing"
	}
	if !sha256Identity.MatchString(origin.ChecksumSHA256) {
		return "origin_checksum_invalid"
	}
	if origin.ChecksumSHA256 != captured.sha256 {
		return "origin_checksum_mismatch"
	}
	return ""
}

func validateLocalOutputExpectation(expectation ArtifactExpectation, node ManifestNode) string {
	if expectation.Path != node.Path {
		return "output_path_drift"
	}
	if expectation.Size != node.Size {
		return "output_size_drift"
	}
	if expectation.SHA256 != node.SHA256 || !sha256Identity.MatchString(expectation.SHA256) {
		return "output_digest_drift"
	}
	if expectation.Class != node.Class {
		return "output_class_drift"
	}
	return ""
}

func (inspector *inspector) rootNode(pathValue string) ManifestNode {
	for _, node := range inspector.nodes {
		if node.Path == pathValue {
			return node
		}
	}
	return ManifestNode{Path: pathValue, Class: ClassOpaqueUnknown, Decision: DecisionReject}
}

func (inspector *inspector) forceRootReject(pathValue, rule string) {
	for index := range inspector.nodes {
		if inspector.nodes[index].Path == pathValue {
			inspector.nodes[index].Decision = DecisionReject
			inspector.nodes[index].Rule = rule
			return
		}
	}
}

func (inspector *inspector) rejectNode(index int, diagnostic Diagnostic) {
	inspector.nodes[index].Decision = DecisionReject
	inspector.nodes[index].Rule = diagnostic.Reason
	if inspector.nodes[index].Kind == NodeArchive || inspector.nodes[index].Kind == NodeCompressedStream {
		inspector.nodes[index].InspectionComplete = false
	}
	inspector.addDiagnostic(diagnostic)
}

func (inspector *inspector) addDiagnostic(diagnostic Diagnostic) {
	if inspector.findingErr != nil {
		return
	}
	if inspector.findings == nil {
		inspector.findings = newFindingAccumulator(inspector.limits.MaxRecordedFindings)
	}
	inspector.findingErr = inspector.findings.add(diagnostic)
}

func (inspector *inspector) findingCount() int64 {
	if inspector.findings == nil {
		return 0
	}
	return inspector.findings.total
}

func finishInspectorResult(manifest Manifest, inspector *inspector) (Result, error) {
	if inspector.findingErr != nil {
		return Result{}, fmt.Errorf("accumulate artifact finding: %w", inspector.findingErr)
	}
	result, err := finishResultWithFindings(manifest, inspector.findings)
	if result.Admission != nil {
		result.Admission.seal = inspector.authorizationSeal
	}
	return result, err
}

func finishResult(manifest Manifest, diagnostics []Diagnostic) (Result, error) {
	findings, err := findingsFromDiagnostics(manifest.LimitVector.MaxRecordedFindings, diagnostics)
	if err != nil {
		return Result{}, fmt.Errorf("accumulate artifact findings: %w", err)
	}
	return finishResultWithFindings(manifest, findings)
}

func finishResultWithFindings(manifest Manifest, findings *findingAccumulator) (Result, error) {
	propagateRejectDecisions(&manifest)
	if err := sealManifest(&manifest, findings); err != nil {
		return Result{}, fmt.Errorf("seal artifact manifest: %w", err)
	}
	canonical, err := EncodeManifest(manifest)
	if err != nil {
		return Result{}, fmt.Errorf("encode artifact manifest: %w", err)
	}
	result := Result{Manifest: manifest, CanonicalBytes: canonical}
	if manifest.Decision != DecisionReject {
		result.Admission = &Admission{
			role: manifest.TrustRole, decision: manifest.Decision,
			digest: manifest.ManifestDigest, schema: SchemaID, policy: PolicyID,
			seal: managerAuthorizationSeal,
		}
		return result, nil
	}
	primary := Diagnostic{Code: CodePolicyInternalError, Path: manifest.RawPayload.Path, Reason: "rejection_without_diagnostic"}
	if len(manifest.Diagnostics) > 0 {
		primary = manifest.Diagnostics[0]
	}
	return result, &PolicyError{Primary: primary}
}

func propagateRejectDecisions(manifest *Manifest) {
	indexes := make(map[string]int, len(manifest.Nodes))
	for index := range manifest.Nodes {
		indexes[manifest.Nodes[index].Path] = index
		if !manifest.Nodes[index].InspectionComplete {
			manifest.Nodes[index].Decision = DecisionReject
			if manifest.Nodes[index].Rule == "" {
				manifest.Nodes[index].Rule = "inspection_incomplete"
			}
		}
	}
	anyRejected := false
	for index := range manifest.Nodes {
		if manifest.Nodes[index].Decision != DecisionReject {
			continue
		}
		anyRejected = true
		parent := manifest.Nodes[index].Parent
		for parent != "" {
			parentIndex, ok := indexes[parent]
			if !ok {
				break
			}
			if manifest.Nodes[parentIndex].Decision != DecisionReject {
				manifest.Nodes[parentIndex].Decision = DecisionReject
				manifest.Nodes[parentIndex].Rule = "descendant_rejected"
			}
			parent = manifest.Nodes[parentIndex].Parent
		}
	}
	if anyRejected {
		manifest.Decision = DecisionReject
	}
}

func (service *Service) failureBeforeCapture(
	role TrustRole,
	descriptor Descriptor,
	payload Payload,
	code DiagnosticCode,
	reason string,
) (Result, error) {
	profile := descriptor.ProfileID
	if _, ok := profileIDs[profile]; !ok {
		profile = ProfileCommonV1
	}
	checksum := descriptor.Origin.ChecksumSHA256
	if !sha256Identity.MatchString(checksum) {
		checksum = digestBytes(nil)
	}
	size := payload.Size
	if size < 0 {
		size = 0
	}
	pathValue := payload.Path
	if pathValue == "" {
		pathValue = "invalid-payload"
	}
	manifest := Manifest{
		SchemaID: SchemaID, PolicyID: PolicyID, PolicyVersion: PolicyVersion,
		LimitVector: service.limits, DetectorRegistryID: DetectorRegistryID,
		Detectors: append([]DetectorIdentity(nil), service.detectors...),
		AdapterID: descriptor.AdapterID, ProfileID: profile, Manager: descriptor.Manager,
		PackageName: descriptor.PackageName, PackageVersion: descriptor.PackageVersion,
		Origin: descriptor.Origin, TrustRole: role, RoleEvidence: initialRoleEvidence(role, descriptor.Origin),
		RawPayload: RawPayloadEvidence{Path: pathValue, Size: size, SHA256: checksum, Kind: "incomplete"},
		Decision:   DecisionReject,
	}
	if validated, pathErr := validateVirtualPath(pathValue, service.limits); pathErr == nil {
		manifest.RawPayload.Path = validated.Canonical
		manifest.Nodes = []ManifestNode{incompleteRootNode(
			validated, payload.Path, size, "", "capture_incomplete",
		)}
	}
	diagnostic := Diagnostic{
		Code: code, Path: pathValue, OriginalNameBase64: originalNameBase64(payload.Path),
		Reason: reason, Size: size, SHA256: checksum,
	}
	if code == CodeInspectionLimitExceeded {
		diagnostic.LimitName = "max_raw_payload_bytes"
		diagnostic.Limit = service.limits.MaxRawPayloadBytes
		diagnostic.Observed = payload.Size
	}
	return finishResult(manifest, []Diagnostic{diagnostic})
}

func (service *Service) failureWithCaptured(
	role TrustRole,
	descriptor Descriptor,
	pathValue string,
	captured blob,
	err error,
) (Result, error) {
	manifest := Manifest{
		SchemaID: SchemaID, PolicyID: PolicyID, PolicyVersion: PolicyVersion,
		LimitVector: service.limits, DetectorRegistryID: DetectorRegistryID,
		Detectors: append([]DetectorIdentity(nil), service.detectors...),
		AdapterID: descriptor.AdapterID, ProfileID: descriptor.ProfileID, Manager: descriptor.Manager,
		PackageName: descriptor.PackageName, PackageVersion: descriptor.PackageVersion,
		Origin: descriptor.Origin, TrustRole: role, RoleEvidence: initialRoleEvidence(role, descriptor.Origin),
		RawPayload: RawPayloadEvidence{Path: pathValue, Size: captured.size, SHA256: captured.sha256, Kind: "file"},
		Accounting: TraversalAccounting{RawPayloadBytes: captured.size},
		Decision:   DecisionReject,
	}
	if validated, pathErr := validateVirtualPath(pathValue, service.limits); pathErr == nil {
		manifest.RawPayload.Path = validated.Canonical
		manifest.Nodes = []ManifestNode{incompleteRootNode(
			validated, pathValue, captured.size, captured.sha256, "inspection_incomplete",
		)}
	}
	return finishResult(manifest, []Diagnostic{diagnosticFromError(pathValue, nil, err)})
}

func incompleteRootNode(
	pathValue VirtualPath,
	original string,
	size int64,
	sha256Value, rule string,
) ManifestNode {
	return ManifestNode{
		Path: pathValue.Canonical, OriginalNameBase64: originalNameBase64(original),
		CollisionKey: pathValue.CollisionKey, Kind: NodeRegularFile,
		Size: size, SHA256: sha256Value, Class: ClassOpaqueUnknown,
		Decision: DecisionReject, Rule: rule, InspectionComplete: false,
	}
}

func initialRoleEvidence(role TrustRole, origin OriginEvidence) []Fact {
	switch role {
	case RoleDependencyInput:
		return dependencyRoleFacts(origin)
	case RoleVerifiedBinaryCandidate:
		return verifiedBinaryRoleFacts()
	default:
		return []Fact{}
	}
}

func diagnosticFromError(pathValue string, chain []string, err error) Diagnostic {
	diagnostic := Diagnostic{
		Code: CodeInspectionUnavailable, Path: pathValue,
		ContainerChain: append([]string(nil), chain...), Reason: "inspection_failure",
	}
	var limit *limitFailure
	if errorAs(err, &limit) {
		diagnostic.Code = CodeInspectionLimitExceeded
		diagnostic.Reason = limit.Error()
		diagnostic.LimitName = limit.name
		diagnostic.Limit = limit.limit
		diagnostic.Observed = limit.observed
	} else if err != nil {
		diagnostic.Reason = err.Error()
	}
	return diagnostic
}

func unavailableDiagnostic(pathValue string, chain []string, reason string, err error) Diagnostic {
	if err != nil {
		reason += ":" + err.Error()
	}
	return Diagnostic{
		Code: CodeInspectionUnavailable, Path: pathValue,
		ContainerChain: append([]string(nil), chain...), Reason: reason,
	}
}

func containerReadDiagnostic(pathValue string, chain []string, detector string, err error) Diagnostic {
	if contextErrorCode(err) {
		diagnostic := unavailableDiagnostic(pathValue, chain, "container_read_cancelled", err)
		diagnostic.DetectorID = detector
		return diagnostic
	}
	var limit *limitFailure
	if errorAs(err, &limit) {
		diagnostic := diagnosticFromError(pathValue, chain, err)
		diagnostic.DetectorID = detector
		return diagnostic
	}
	return Diagnostic{
		Code: CodeArchiveInvalid, Path: pathValue, DetectorID: detector,
		ContainerChain: append([]string(nil), chain...), Reason: "container_member_invalid:" + err.Error(),
	}
}

func contextErrorCode(err error) bool {
	return errorIs(err, context.Canceled) || errorIs(err, context.DeadlineExceeded)
}

func init() {
	// Ensure closed detector identity order cannot accidentally contain a
	// duplicate that would make canonical evidence dependent on sort stability.
	identities := detectorIdentities()
	sort.Slice(identities, func(left, right int) bool { return identities[left].ID < identities[right].ID })
	for index := 1; index < len(identities); index++ {
		if identities[index-1].ID == identities[index].ID {
			panic("duplicate artifact detector identity")
		}
	}
}
