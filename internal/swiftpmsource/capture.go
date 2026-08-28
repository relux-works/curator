package swiftpmsource

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/privatedir"
)

var versionManifestPattern = regexp.MustCompile(`\APackage@swift-([0-9]+(?:\.[0-9]+){0,2})\.swift\z`)
var swiftVersionPattern = regexp.MustCompile(`(?i)swift(?: version)?[^0-9]*([0-9]+(?:\.[0-9]+){0,2})`)

// CaptureAndClose performs intake before each executable manifest, freezes the
// root lock, acquires every exact pin, and closes the selected graph through C4.
func CaptureAndClose(ctx context.Context, config Config, request Request) (*Capture, error) {
	if config.Store == nil || config.Policy == nil || config.Evaluator == nil || config.Broker == nil || config.MirrorVerifier == nil || config.CausalHead == "" {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM capture authority is incomplete")
	}
	root, err := cleanAbsoluteRoot(request.Root)
	if err != nil || request.Product == "" || strings.ContainsAny(request.Product, "\x00\r\n") {
		return nil, fail(CodeGraphIncomplete, "root and one executable product are required")
	}
	if err = validateDestination(config.Destination); err != nil {
		return nil, err
	}
	if err = validateToolchain(config.Toolchain); err != nil {
		return nil, err
	}

	rootInput, rootArtifact, rootInventory, rootFiles, err := admitTree(ctx, config, "root", root, "workspace:root", "root", "")
	if err != nil {
		return nil, err
	}
	productNode, productID, err := provisionalProduct(request.Product, rootInput.Receipt.ContentSHA256)
	if err != nil {
		return nil, err
	}
	platformNode, platformID, toolNodes, toolIDs, err := bindingRecords(config)
	if err != nil {
		return nil, err
	}
	selection, err := closuregraph.NewSelectionContext([]closuregraph.ID{productID}, map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID, closuregraph.PlatformHost: platformID}, []string{}, true, cloneMap(config.Destination.Markers), map[string]string{}, []string{ConditionEvaluatorID})
	if err != nil {
		return nil, err
	}
	selectionID, _ := selection.ID()
	toolIDs = sortedIDs(toolIDs)
	c0Payload := closuregraph.C0ProfilePayload{
		AdapterProfileID: ProfileID, SchemaIDs: []string{"closure-graph-v1", ManifestSchemaID, "swiftpm-source-manifest-v1"},
		ArtifactPolicyID: artifactpolicy.PolicyID, DetectorRegistryID: artifactpolicy.DetectorRegistryID,
		SourceGrammarIDs: []string{"c-lexer-v1", "clang-modulemap-lexer-v1", "cxx-lexer-v1", "objective-c-lexer-v1", "swift-lexer-v1"}, LimitVectorID: artifactpolicy.LimitVectorID,
		SelectionContextID: selectionID, PlatformNodeIDs: []closuregraph.ID{platformID}, PlatformRoles: map[closuregraph.PlatformRole]closuregraph.ID{closuregraph.PlatformTarget: platformID, closuregraph.PlatformHost: platformID},
		ManagerSchemaIDs: []string{"Package.resolved-v2", "Package.resolved-v3"}, ConfigurationPolicyID: "swiftpm-configuration-v1",
		CapabilityIDs: []string{"force-resolved-versions", "kind-preserving-mirrors", "native-build-system", "network-denied", "prebuilts-disabled", "source-only"}, EvidenceToolchainNodeIDs: toolIDs,
	}
	c0, err := closuregraph.NewCheckpoint(c0Payload, nil, nil)
	if err != nil {
		return nil, err
	}
	c0ID, _ := c0.ID()
	if err = bindGitAuthority(config, c0ID); err != nil {
		return nil, err
	}

	rootPackage, err := evaluateAdmitted(ctx, config, c0ID, "root", rootInput, rootFiles, rootInventory)
	if err != nil {
		return nil, err
	}
	rootPackage.ArtifactManifestID = rootArtifact
	rootPackage.Origin = "workspace:root"
	rootPackage.Kind = SourcePath
	rootPackage.Identity = strings.ToLower(rootPackage.Manifest.PackageName)
	if rootPackage.Identity == "" {
		return nil, fail(CodeManifestReplayDrift, "root manifest omitted package identity")
	}
	if err = validateRootSelection(rootPackage.Manifest, request.Product, config.Destination.Markers); err != nil {
		return nil, err
	}

	lockBytes := append([]byte(nil), request.Resolved...)
	var resolutionPermitID, resolutionReceiptID closuregraph.ID
	var resolutionJournalIDs, resolutionDerivationIDs []closuregraph.ID
	if len(lockBytes) != 0 {
		if err = verifySuppliedRootLock(rootInput, lockBytes); err != nil {
			return nil, err
		}
	}
	if len(lockBytes) == 0 {
		if config.Resolver == nil {
			return nil, fail(CodeResolutionUnfrozen, "root lock is absent and no controlled resolver is configured")
		}
		permit := resolutionPermit(c0ID, rootInput, config.Toolchain.Git, config.Destination)
		if err = recheckTool(ctx, config.Toolchain, config.Toolchain.Git); err != nil {
			return nil, err
		}
		result, resolveErr := config.Resolver.Resolve(ctx, mustProtectedPath(rootInput), permit, rootPackage.Manifest)
		if resolveErr != nil {
			return nil, fail(CodeResolutionUnfrozen, "controlled root resolution failed: %v", resolveErr)
		}
		if !result.ReceiptID.Valid() {
			return nil, fail(CodeDerivationUnauthorized, "controlled root resolution omitted its derivation receipt")
		}
		if err = config.Resolver.VerifyResult(permit, result); err != nil {
			return nil, err
		}
		lockBytes = append([]byte(nil), result.Lock...)
		resolutionPermitID, resolutionReceiptID = permit.ID, result.ReceiptID
		resolutionJournalIDs = append([]closuregraph.ID(nil), result.JournalEntryIDs...)
		resolutionDerivationIDs = append([]closuregraph.ID(nil), result.DerivationReceiptIDs...)
	}
	lock, err := ParseResolved(lockBytes)
	if err != nil {
		return nil, err
	}
	lockID, err := lockRecordID(lock)
	if err != nil {
		return nil, err
	}

	packages := []PackageEvidence{rootPackage}
	pinsByIdentity := map[string]Pin{}
	for _, pin := range lock.Pins {
		pinsByIdentity[pin.Identity] = pin
	}
	if err = reconcileOneManifest(rootPackage, pinsByIdentity); err != nil {
		return nil, err
	}
	packageByIdentity := map[string]int{rootPackage.Identity: 0}
	if err = captureLocalDependencies(ctx, config, c0ID, mustProtectedPath(rootInput), rootPackage, rootArtifact, rootInput, &packages, packageByIdentity); err != nil {
		return nil, err
	}
	mirrors := make([]Mirror, 0, len(lock.Pins))
	visitedPins := map[string]bool{}
	queuedPackages := 0
	for queuedPackages < len(packages) {
		pkgIndex := queuedPackages
		queuedPackages++
		for _, dependency := range packages[pkgIndex].Manifest.Dependencies {
			if dependency.Kind == SourcePath || visitedPins[dependency.Identity] {
				continue
			}
			pin, exists := pinsByIdentity[dependency.Identity]
			if !exists {
				return nil, failFields(CodeResolvedFileOutOfDate, map[string]string{"package": packages[pkgIndex].Identity, "dependency": dependency.Identity}, "manifest dependency is absent from root lock")
			}
			visitedPins[pin.Identity] = true
			if err = recheckTool(ctx, config.Toolchain, config.Toolchain.Git); err != nil {
				return nil, err
			}
			snapshot, acquireErr := config.Broker.Acquire(ctx, pin)
			if acquireErr != nil {
				return nil, failFields(CodeDependencyMirrorMissing, map[string]string{"identity": pin.Identity}, "source-control acquisition failed: %v", acquireErr)
			}
			if err = validateSnapshot(pin, snapshot); err != nil {
				return nil, err
			}
			if verifier, ok := config.Broker.(acquisitionEvidenceVerifier); ok {
				if err = verifier.VerifySnapshot(pin, snapshot); err != nil {
					return nil, err
				}
			}
			input, artifactID, inventoryID, files, captureErr := admitTree(ctx, config, pin.Identity, snapshot.Root, pin.CanonicalLocation, pin.Revision, lock.Digest)
			if captureErr != nil {
				return nil, captureErr
			}
			pkg, evaluateErr := evaluateAdmitted(ctx, config, c0ID, pin.Identity, input, files, inventoryID)
			if evaluateErr != nil {
				return nil, evaluateErr
			}
			if identity := strings.ToLower(pkg.Manifest.PackageName); identity != pin.Identity {
				return nil, failFields(CodeDependencyPinMismatch, map[string]string{"expected": pin.Identity, "observed": identity}, "captured package identity differs from lock")
			}
			mirror, mirrorErr := captureMirror(ctx, config, pin, snapshot, input, input.Receipt.ContentSHA256)
			if mirrorErr != nil {
				return nil, mirrorErr
			}
			pkg.Identity, pkg.Origin, pkg.Revision, pkg.GitTree, pkg.Kind, pkg.ArtifactManifestID, pkg.Mirror = pin.Identity, pin.CanonicalLocation, pin.Revision, strings.ToLower(snapshot.GitTree), pin.Kind, artifactID, &mirror
			pkg.BrokerReceiptID = snapshot.BrokerReceiptID
			pkg.BrokerPermitIDs = append([]closuregraph.ID(nil), snapshot.BrokerPermitIDs...)
			pkg.BrokerProcessReceiptIDs = append([]closuregraph.ID(nil), snapshot.BrokerProcessReceiptIDs...)
			if err = reconcileOneManifest(pkg, pinsByIdentity); err != nil {
				return nil, err
			}
			if _, duplicate := packageByIdentity[pkg.Identity]; duplicate {
				return nil, failFields(CodeDependencyPinMismatch, map[string]string{"identity": pkg.Identity}, "multiple snapshots map to one lock pin")
			}
			packageByIdentity[pkg.Identity] = len(packages)
			packages = append(packages, pkg)
			if err = captureLocalDependencies(ctx, config, c0ID, mustProtectedPath(input), pkg, artifactID, input, &packages, packageByIdentity); err != nil {
				return nil, err
			}
			mirrors = append(mirrors, mirror)
		}
	}
	for _, pin := range lock.Pins {
		if !visitedPins[pin.Identity] {
			return nil, failFields(CodeResolvedFileOutOfDate, map[string]string{"identity": pin.Identity}, "root lock contains a dangling pin")
		}
	}
	sort.Slice(mirrors, func(i, j int) bool { return mirrors[i].Identity < mirrors[j].Identity })

	if err = reconcileManifestDependencies(packages, lock); err != nil {
		return nil, err
	}
	graphResult, err := buildGraph(packages, lock, productNode, productID)
	if err != nil {
		return nil, err
	}
	activeTargetIDs, err := validateSelectedReachability(packages, request.Product, config.Destination.Markers)
	if err != nil {
		return nil, err
	}
	for _, targetName := range activeTargetIDs {
		id, ok := graphResult.targetIDs[targetName]
		if !ok {
			return nil, fail(CodeGraphIncomplete, "selected target %s is absent from capture", targetName)
		}
		graphResult.selectedTargetIDs = append(graphResult.selectedTargetIDs, id)
	}
	graphResult.selectedTargetIDs = sortedIDs(graphResult.selectedTargetIDs)

	binding, bindingNodes, bindingEdges, authority, err := bindSelection(graphResult, selection, c0, platformNode, toolNodes, toolIDs)
	if err != nil {
		return nil, err
	}
	records := closuregraph.NewRecordTables(graphResult.nodes, graphResult.edges, bindingNodes, bindingEdges)
	bundle, err := closuregraph.ProjectActive(graphResult.graph, selection, binding, records, authority, []closuregraph.ConditionEvaluator{swiftPMConditionEvaluator{markers: config.Destination.Markers}})
	if err != nil {
		return nil, err
	}

	graphID, _ := graphResult.graph.ID()
	activeID, _ := bundle.Active.ID()
	bindingID, _ := binding.ID()
	manifestIDs, intakeIDs, originIDs, handleIDs, journalIDs, derivationIDs := evidenceIDs(packages)
	if resolutionPermitID.Valid() {
		journalIDs = sortedIDs(append(journalIDs, append(resolutionJournalIDs, resolutionPermitID, resolutionReceiptID)...))
		derivationIDs = sortedIDs(append(derivationIDs, append(resolutionDerivationIDs, resolutionReceiptID)...))
	}
	c1, err := closuregraph.NewCheckpoint(closuregraph.C1ResolvePayload{RootDeclarationIDs: []closuregraph.ID{rootPackage.ManifestDigest}, WorkspaceDeclarationIDs: []closuregraph.ID{}, LockCandidateID: lockID, ConditionEdgeIDs: graphResult.conditionEdgeIDs, ParserEvaluatorIDs: []string{ConditionEvaluatorID}, CandidateNodeIDs: graphResult.graph.NodeIDs, CandidateEdgeIDs: graphResult.graph.EdgeIDs, SelectionContextID: selectionID, JournalEntryIDs: journalIDs}, &c0, nil)
	if err != nil {
		return nil, err
	}
	c2, err := closuregraph.NewCheckpoint(closuregraph.C2CapturePayload{IntakeReceiptIDs: intakeIDs, OriginIDs: originIDs, ProtectedHandleIDs: handleIDs, BrokerReceiptIDs: graphResult.brokerIDs}, &c1, nil)
	if err != nil {
		return nil, err
	}
	c3, err := closuregraph.NewCheckpoint(closuregraph.C3AdmitPayload{Phase: "main", IntakeReceiptIDs: intakeIDs, ArtifactManifestIDs: manifestIDs, DerivationReceiptIDs: derivationIDs}, &c2, nil)
	if err != nil {
		return nil, err
	}
	c4, err := closuregraph.NewCheckpoint(closuregraph.C4ClosePayload{ActiveGraphID: activeID, CapturedGraphID: graphID, SelectionBindingID: bindingID, SelectionContextID: selectionID}, &c3, nil)
	if err != nil {
		return nil, err
	}
	inventoryDigest, _ := closuregraph.DomainID("swiftpm-source-inventory-v1", map[string]any{"package_inventory_ids": idsAny(packageInventoryIDs(packages))})
	return &Capture{Lock: lock, RootLock: lock, Packages: packages, Mirrors: mirrors, Graph: graphResult.graph, Selection: selection, Binding: binding, Active: bundle.Active, Records: records, Authority: authority, C0: c0, C1: c1, C2: c2, C3: c3, C4: c4, ResolutionPermitID: resolutionPermitID, ResolutionReceiptID: resolutionReceiptID, GraphDigest: graphID, InventoryDigest: inventoryDigest, ProductNodeID: productID, TargetNodeIDs: graphResult.selectedTargetIDs, CausalHead: string(activeID), rootInput: rootInput, config: config}, nil
}

func captureMirror(ctx context.Context, config Config, pin Pin, snapshot Snapshot, sourceInput closureexec.AdmittedInput, snapshotDigest closuregraph.ID) (Mirror, error) {
	broker, brokerOK := config.Broker.(*GitBroker)
	if !brokerOK || broker.authority == nil || !snapshot.AcquisitionReceipt.PermitID.Valid() || len(snapshot.CommitObject) == 0 {
		// Test-only/fake brokers retain the closed seam without gaining production
		// authority; their verifier fixtures continue to exercise graph behavior.
		return captureFixtureMirror(ctx, config, pin, snapshot, sourceInput, snapshotDigest)
	}
	commitInput, err := admitCommitEvidence(ctx, config, pin, snapshot.CommitObject)
	if err != nil {
		return Mirror{}, err
	}
	sourceReceiptID, _ := sourceInput.Receipt.ID()
	commitReceiptID, _ := commitInput.Receipt.ID()
	acquisitionID, _ := snapshot.AcquisitionReceipt.ID()
	manifestID, _ := closuregraph.DomainID("source-control-mirror-transform-declaration-v1", map[string]any{"acquisition_receipt_id": string(acquisitionID), "git_tree": snapshot.GitTree, "kind": string(pin.Kind), "revision": pin.Revision, "source_receipt_id": string(sourceReceiptID)})
	evidencePath := "mirror-evidence.json"
	expected := []closureexec.EvidenceRequirement{{Path: evidencePath, SchemaID: "source-control-mirror-v1", ArtifactManifestID: manifestID}}
	evidenceSchemaID, _ := closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": []any{map[string]any{"artifact_manifest_id": string(manifestID), "path": evidencePath, "schema_id": "source-control-mirror-v1"}}})
	limits := closureexec.ResourceLimits{OutputBytes: 64 << 20, ReadBytes: 1 << 30, WriteBytes: 1 << 30, WallTimeMillis: 120000, ProcessCount: 1}
	limitID, _ := limits.ID()
	receiptIDs := []closuregraph.ID{sourceReceiptID, commitReceiptID}
	sort.Slice(receiptIDs, func(i, j int) bool { return receiptIDs[i] < receiptIDs[j] })
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{sourceReceiptID: sourceInput, commitReceiptID: commitInput}
	mounts := make([]closureexec.InputMount, len(receiptIDs))
	for index, receiptID := range receiptIDs {
		mounts[index] = closureexec.InputMount{ReceiptID: receiptID, Path: fmt.Sprintf("mirror-input/%d", index)}
	}
	mirrorRelative := filepath.ToSlash(filepath.Join(".swiftpm-authority", "mirrors", pin.Identity+"-"+pin.Revision+".git"))
	mirrorRoot := filepath.Join(broker.authority.executionRoot, filepath.FromSlash(mirrorRelative))
	permit := closureexec.DerivationPermit{SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: broker.authority.derivation.CurrentCausalHead(), InvocationKey: "source-control-mirror-v1:" + string(acquisitionID), InvocationSubtype: closureexec.DerivationMirror, AdmittedInputReceiptIDs: receiptIDs, InputMounts: mounts, C0CheckpointID: broker.authority.c0ID, ToolchainNodeID: mustToolNodeID(config.Toolchain.Git), ToolchainFingerprint: config.Toolchain.Git.Fingerprint, ExecutableSHA256: config.Toolchain.Git.ExecutableSHA256, Executable: config.Toolchain.Git.ExecutableRelativePath, CWD: filepath.ToSlash(filepath.Join(".swiftpm-authority", "work")), Argv: []string{"curator-source-control-mirror-v1"}, Environment: map[string]string{"CURATOR_GIT_REVISION": pin.Revision, "CURATOR_GIT_TREE": snapshot.GitTree, "CURATOR_MIRROR_ROOT": mirrorRelative, "CURATOR_OUTPUT_ROOT": filepath.ToSlash(filepath.Join(".swiftpm-authority", "mirror-output")), "CURATOR_SOURCE_CONTROL_KIND": string(pin.Kind), "HOME": "empty/home", "TZ": "UTC"}, HostID: mustPlatformID(config.Toolchain.Git.PlatformABI), TargetID: mustPlatformID(config.Toolchain.Git.PlatformABI), AllowedProcesses: []string{config.Toolchain.Git.ExecutableRelativePath}, ReadRoots: sortedUniqueStrings([]string{filepath.ToSlash(filepath.Dir(config.Toolchain.Git.ExecutableRelativePath)), mounts[0].Path, mounts[1].Path}), WriteRoots: sortedUniqueStrings([]string{evidencePath, mirrorRelative}), ExpectedEvidence: expected, LocalOutputs: []closureexec.LocalOutputDeclaration{{Path: mirrorRelative, SchemaID: "source-control-mirror-v1"}}, Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceSchemaID}
	if err = privatedir.MakeAll(filepath.Join(broker.authority.executionRoot, filepath.FromSlash(permit.CWD))); err != nil {
		return Mirror{}, err
	}
	permitID, err := broker.authority.derivation.Commit(permit)
	if err != nil {
		return Mirror{}, err
	}
	issuedPermit, err := broker.authority.derivation.IssuedDerivationPermit(permitID)
	if err != nil {
		return Mirror{}, err
	}
	receipt, err := broker.authority.derivation.Execute(ctx, permitID, func(checkCtx context.Context) (closureexec.ToolchainIdentity, error) {
		return broker.authority.recheck(checkCtx)
	}, inputs)
	if err != nil {
		return Mirror{}, err
	}
	derivationReceiptID, err := receipt.ID()
	if err != nil {
		return Mirror{}, err
	}
	protectedNodes, mirrorDigest, err := inventoryMirrorTree(mirrorRoot)
	if err != nil {
		return Mirror{}, err
	}
	evidenceBytes, err := os.ReadFile(filepath.Join(broker.authority.executionRoot, ".swiftpm-authority", "mirror-output", evidencePath))
	if err != nil {
		return Mirror{}, err
	}
	authorization, err := config.Policy.IssueSourceControlMirrorAuthorization(artifactpolicy.SourceControlMirrorAuthorizationRequest{ProfileID: ProfileID, Kind: string(pin.Kind), Origin: pin.CanonicalLocation, Revision: pin.Revision, GitTree: snapshot.GitTree, MirrorPath: mirrorRelative, MirrorDigest: mirrorDigest, ArtifactManifestID: manifestID, AdmittedSourceReceiptID: sourceReceiptID, AcquisitionExecutor: broker.authority.acquisition, AcquisitionReceipt: snapshot.AcquisitionReceipt, DerivationExecutor: broker.authority.derivation, DerivationPermit: issuedPermit, DerivationReceipt: receipt, TransformEvidence: evidenceBytes})
	if err != nil {
		if policyErr, ok := err.(*artifactpolicy.PolicyError); ok {
			return Mirror{}, fail(CodeDerivationUnauthorized, "mirror authorization rejected: %s", policyErr.Primary.Reason)
		}
		return Mirror{}, err
	}
	if err = os.RemoveAll(filepath.Join(broker.authority.executionRoot, ".swiftpm-authority", "mirror-output")); err != nil {
		return Mirror{}, err
	}
	handle, err := config.Store.CaptureTree("swiftpm-mirror:"+pin.Identity+":"+pin.Revision, mirrorRoot)
	if err != nil {
		return Mirror{}, err
	}
	protected, err := handle.ProtectedPath()
	if err != nil {
		return Mirror{}, err
	}
	nodes, observedMirrorDigest, err := inventoryMirrorTree(protected)
	if err != nil {
		return Mirror{}, err
	}
	if observedMirrorDigest != mirrorDigest || !reflect.DeepEqual(nodes, protectedNodes) {
		return Mirror{}, fail(CodeDerivationUnauthorized, "protected mirror differs from receipted transform")
	}
	if err = config.Policy.ValidateSourceControlMirrorAuthorization(authorization, ProfileID, string(pin.Kind), pin.CanonicalLocation, pin.Revision, snapshot.GitTree, mirrorRelative, mirrorDigest, manifestID, sourceReceiptID); err != nil {
		return Mirror{}, err
	}
	intakeReceipt, err := admitMirrorTree(config, handle, pin, snapshot, GitVerificationEvidence{}, nodes, mirrorDigest, manifestID)
	if err != nil {
		return Mirror{}, err
	}
	input := closureexec.AdmittedInput{Receipt: intakeReceipt, Tree: handle}
	verification, verifyErr := verifyCapturedMirror(ctx, config, protected, pin, snapshot, input)
	if verifyErr != nil {
		err = verifyErr
		return Mirror{}, failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "captured mirror does not prove the pinned revision and tree: %v", err)
	}
	if !verificationIDsIssued(config, verification) {
		return Mirror{}, fail(CodeDerivationUnauthorized, "mirror verification receipts are not authority-issued")
	}
	receiptID, err := intakeReceipt.ID()
	if err != nil {
		return Mirror{}, err
	}
	return Mirror{Identity: pin.Identity, Original: pin.CanonicalLocation, Local: protected, Revision: pin.Revision, GitTree: strings.ToLower(snapshot.GitTree), OriginalKind: pin.Kind, LocalKind: snapshot.Kind, SnapshotDigest: snapshotDigest, MirrorDigest: mirrorDigest, BrokerReceiptID: snapshot.BrokerReceiptID, MirrorIntakeReceiptID: receiptID, ArtifactManifestID: manifestID, CommitEvidenceIntakeReceiptID: commitReceiptID, CommitEvidenceArtifactManifestID: commitInput.Receipt.ArtifactManifestID, MirrorDerivationPermitID: permitID, MirrorDerivationReceiptID: derivationReceiptID, VerificationPermitIDs: append([]closuregraph.ID(nil), verification.PermitIDs...), VerificationReceiptIDs: append([]closuregraph.ID(nil), verification.ReceiptIDs...), AuthorizedOutputPath: mirrorRelative, authorization: authorization, input: input, commitInput: commitInput}, nil
}

func captureFixtureMirror(ctx context.Context, config Config, pin Pin, snapshot Snapshot, _ closureexec.AdmittedInput, snapshotDigest closuregraph.ID) (Mirror, error) {
	handle, err := config.Store.CaptureTree("swiftpm-mirror:"+pin.Identity+":"+pin.Revision, snapshot.MirrorRoot)
	if err != nil {
		return Mirror{}, err
	}
	protected, err := handle.ProtectedPath()
	if err != nil {
		return Mirror{}, err
	}
	nodes, mirrorDigest, err := inventoryMirrorTree(protected)
	if err != nil {
		return Mirror{}, err
	}
	verification, err := config.MirrorVerifier.Verify(ctx, protected, pin, snapshot)
	if err != nil {
		return Mirror{}, failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "captured fixture mirror does not prove the pinned revision and tree: %v", err)
	}
	manifestID, err := mirrorArtifactManifestID(pin, snapshot, mirrorDigest, nodes, verification)
	if err != nil {
		return Mirror{}, err
	}
	intakeReceipt, err := admitMirrorTree(config, handle, pin, snapshot, verification, nodes, mirrorDigest, manifestID)
	if err != nil {
		return Mirror{}, err
	}
	receiptID, _ := intakeReceipt.ID()
	input := closureexec.AdmittedInput{Receipt: intakeReceipt, Tree: handle}
	return Mirror{Identity: pin.Identity, Original: pin.CanonicalLocation, Local: protected, Revision: pin.Revision, GitTree: strings.ToLower(snapshot.GitTree), OriginalKind: pin.Kind, LocalKind: snapshot.Kind, SnapshotDigest: snapshotDigest, MirrorDigest: mirrorDigest, BrokerReceiptID: snapshot.BrokerReceiptID, MirrorIntakeReceiptID: receiptID, ArtifactManifestID: manifestID, VerificationPermitIDs: append([]closuregraph.ID(nil), verification.PermitIDs...), VerificationReceiptIDs: append([]closuregraph.ID(nil), verification.ReceiptIDs...), input: input}, nil
}

func admitCommitEvidence(ctx context.Context, config Config, pin Pin, payload []byte) (closureexec.AdmittedInput, error) {
	if len(payload) == 0 {
		return closureexec.AdmittedInput{}, fail(CodeDerivationUnauthorized, "acquisition commit evidence is empty")
	}
	digest := sha256.Sum256(payload)
	digestID := "sha256:" + hex.EncodeToString(digest[:])
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileSwiftPMV1, Manager: "swiftpm", PackageName: pin.Identity, PackageVersion: pin.Revision, Origin: artifactpolicy.OriginEvidence{Locator: pin.CanonicalLocation + "#commit-object", ImmutableID: pin.Revision, LockRecord: pin.Revision, ChecksumSHA256: digestID, Verified: true}, DeclaredText: map[string]artifactpolicy.TextDeclaration{"commit-evidence.txt": {Grammar: artifactpolicy.GrammarPlain, Class: artifactpolicy.ClassTextMetadata}}}
	result, err := config.Policy.AdmitDependency(ctx, artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: "commit-evidence.txt", Size: int64(len(payload)), Reader: bytes.NewReader(payload)}})
	if err != nil {
		return closureexec.AdmittedInput{}, err
	}
	manifestID := closuregraph.ID(result.Manifest.ManifestDigest)
	if result.Admission == nil || !manifestID.Valid() {
		return closureexec.AdmittedInput{}, fail(CodeDerivationUnauthorized, "commit evidence artifact admission is absent")
	}
	handle, err := config.Store.Capture(pin.CanonicalLocation+"#commit-object", int64(len(payload)), bytes.NewReader(payload))
	if err != nil {
		return closureexec.AdmittedInput{}, err
	}
	receipt, err := config.Store.Admit(handle, pin.CanonicalLocation+"#commit-object", closureexec.AdmissionEvidence{PreviousCausalHead: config.CausalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: string(artifactpolicy.ProfileSwiftPMV1), DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return closureexec.AdmittedInput{}, err
	}
	return closureexec.AdmittedInput{Receipt: receipt, Handle: handle}, nil
}

func verifyCapturedMirror(ctx context.Context, config Config, protected string, pin Pin, snapshot Snapshot, input closureexec.AdmittedInput) (GitVerificationEvidence, error) {
	if verifier, ok := config.MirrorVerifier.(*GitMirrorVerifier); ok {
		return verifier.VerifyAdmitted(ctx, pin, snapshot, input)
	}
	return config.MirrorVerifier.Verify(ctx, protected, pin, snapshot)
}

type mirrorArtifactNode struct {
	Path, SHA256 string
	Size         int64
	Executable   bool
}

func inventoryMirrorTree(root string) ([]mirrorArtifactNode, closuregraph.ID, error) {
	nodes := []mirrorArtifactNode{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == root || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
			return fail(CodeDependencyOriginUnsupported, "mirror contains a linked or special node")
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies an exact member below the captured mirror root.
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, current)
		nodes = append(nodes, mirrorArtifactNode{Path: filepath.ToSlash(relative), SHA256: string(sha256Bytes(payload)), Size: int64(len(payload)), Executable: info.Mode()&0o111 != 0})
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Path < nodes[j].Path })
	values := make([]any, len(nodes))
	for index, node := range nodes {
		values[index] = map[string]any{"executable": node.Executable, "path": node.Path, "sha256": node.SHA256, "size": node.Size}
	}
	id, err := closuregraph.DomainID("source-control-mirror-tree-v1", map[string]any{"nodes": values})
	return nodes, id, err
}

func mirrorArtifactManifestID(pin Pin, snapshot Snapshot, mirrorDigest closuregraph.ID, nodes []mirrorArtifactNode, verification GitVerificationEvidence) (closuregraph.ID, error) {
	values := make([]any, len(nodes))
	for index, node := range nodes {
		values[index] = map[string]any{"executable": node.Executable, "path": node.Path, "sha256": node.SHA256, "size": node.Size}
	}
	return closuregraph.DomainID("swiftpm-git-mirror-artifact-manifest-v1", map[string]any{"decision": "ADMIT_MIRROR_CONTAINER", "git_tree": strings.ToLower(snapshot.GitTree), "identity": pin.Identity, "kind": string(pin.Kind), "mirror_digest": string(mirrorDigest), "nodes": values, "policy_id": artifactpolicy.PolicyID, "revision": pin.Revision, "schema_id": "swiftpm-git-mirror-artifact-manifest-v1", "verification_permit_ids": idsAny(verification.PermitIDs), "verification_receipt_ids": idsAny(verification.ReceiptIDs)})
}

func admitMirrorTree(config Config, handle *closureexec.SourceTreeHandle, pin Pin, snapshot Snapshot, verification GitVerificationEvidence, nodes []mirrorArtifactNode, actualDigest, manifestID closuregraph.ID) (closureexec.IntakeAdmissionReceipt, error) {
	protected, err := handle.ProtectedPath()
	if err != nil {
		return closureexec.IntakeAdmissionReceipt{}, err
	}
	observedNodes, observed, err := inventoryMirrorTree(protected)
	if err != nil {
		return closureexec.IntakeAdmissionReceipt{}, err
	}
	if observed != actualDigest || !reflect.DeepEqual(observedNodes, nodes) || !manifestID.Valid() {
		return closureexec.IntakeAdmissionReceipt{}, fail(CodeDerivationUnauthorized, "mirror artifact evidence does not bind the exact bare repository bytes")
	}
	if !snapshot.AcquisitionReceipt.PermitID.Valid() {
		expected, expectedErr := mirrorArtifactManifestID(pin, snapshot, actualDigest, nodes, verification)
		if expectedErr != nil || expected != manifestID {
			return closureexec.IntakeAdmissionReceipt{}, fail(CodeDerivationUnauthorized, "fixture mirror artifact evidence was substituted")
		}
	}
	return config.Store.AdmitTree(handle, pin.CanonicalLocation, closureexec.AdmissionEvidence{PreviousCausalHead: config.CausalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: "swiftpm-git-mirror-v1", DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
}

func bindGitAuthority(config Config, c0ID closuregraph.ID) error {
	broker, brokerOK := config.Broker.(*GitBroker)
	verifier, verifierOK := config.MirrorVerifier.(*GitMirrorVerifier)
	resolver, resolverOK := config.Resolver.(*BrokeredResolver)
	resolverBroker, resolverUsesGit := any(nil), false
	if resolverOK {
		resolverBroker, resolverUsesGit = resolver.Broker.(*GitBroker)
	}
	if !brokerOK && !verifierOK && !resolverUsesGit {
		return nil
	}
	toolRoot := config.GitToolRoot
	if toolRoot == "" {
		toolRoot = config.GitExecutionRoot
	}
	authority, err := newSharedGitAuthority(c0ID, config.Toolchain, config.GitExecutionRoot, toolRoot)
	if err != nil {
		return err
	}
	if brokerOK {
		if err = broker.bindGitAuthority(authority); err != nil {
			return err
		}
	}
	if resolverUsesGit {
		if err = resolverBroker.(*GitBroker).bindGitAuthority(authority); err != nil {
			return err
		}
	}
	if verifierOK {
		if err = verifier.bindGitAuthority(authority); err != nil {
			return err
		}
	}
	return nil
}

func verificationIDsIssued(config Config, evidence GitVerificationEvidence) bool {
	if verifier, ok := config.MirrorVerifier.(*GitMirrorVerifier); ok {
		return verifier.authority != nil && verifier.authority.verifyDerivations(evidence.PermitIDs, evidence.ReceiptIDs) == nil
	}
	return true
}

func admitTree(ctx context.Context, config Config, identity, source, origin, immutable string, lockDigest closuregraph.ID) (closureexec.AdmittedInput, closuregraph.ID, closuregraph.ID, []string, error) {
	handle, err := config.Store.CaptureTree("swiftpm:"+identity+":"+immutable, source)
	if err != nil {
		return closureexec.AdmittedInput{}, "", "", nil, err
	}
	protected, err := handle.ProtectedPath()
	if err != nil {
		return closureexec.AdmittedInput{}, "", "", nil, err
	}
	files, inventoryID, err := inventoryTree(protected)
	if err != nil {
		return closureexec.AdmittedInput{}, "", "", nil, err
	}
	lockRecord := string(lockDigest)
	if lockRecord == "" {
		lockRecord = "root-intake-before-lock"
	}
	packageVersion := immutable
	if packageVersion == "" {
		packageVersion = "workspace"
	}
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileSwiftPMV1, Manager: "swiftpm", PackageName: identity, PackageVersion: packageVersion}
	probe, probeErr := config.Policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: protected, VirtualRoot: "package"})
	if probeErr != nil && artifactpolicy.ErrorCode(probeErr) != artifactpolicy.CodeOriginUnverified {
		return closureexec.AdmittedInput{}, "", "", nil, probeErr
	}
	treeDigest := closuregraph.ID(probe.Manifest.RawPayload.SHA256)
	if !treeDigest.Valid() {
		return closureexec.AdmittedInput{}, "", "", nil, fail(CodeDerivationUnauthorized, "artifact policy did not derive a canonical tree identity")
	}
	immutableID := immutable
	if immutableID == "" {
		immutableID = string(treeDigest)
	}
	descriptor.Origin = artifactpolicy.OriginEvidence{Locator: origin, ImmutableID: immutableID, LockRecord: lockRecord, ChecksumSHA256: string(treeDigest), Verified: true}
	result, policyErr := config.Policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: protected, VirtualRoot: "package"})
	if policyErr != nil {
		return closureexec.AdmittedInput{}, "", "", nil, policyErr
	}
	manifestID := closuregraph.ID(result.Manifest.ManifestDigest)
	if result.Admission == nil || !manifestID.Valid() {
		return closureexec.AdmittedInput{}, "", "", nil, fail(CodeDerivationUnauthorized, "artifact admission did not issue source authority")
	}
	receipt, err := config.Store.AdmitTree(handle, origin, closureexec.AdmissionEvidence{PreviousCausalHead: config.CausalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: string(artifactpolicy.ProfileSwiftPMV1), DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return closureexec.AdmittedInput{}, "", "", nil, err
	}
	return closureexec.AdmittedInput{Receipt: receipt, Tree: handle}, manifestID, inventoryID, files, nil
}

func evaluateAdmitted(ctx context.Context, config Config, c0ID closuregraph.ID, identity string, input closureexec.AdmittedInput, files []string, inventoryID closuregraph.ID) (PackageEvidence, error) {
	return evaluateAdmittedAt(ctx, config, c0ID, identity, input, mustProtectedPath(input), ".", files, inventoryID)
}

func evaluateAdmittedAt(ctx context.Context, config Config, c0ID closuregraph.ID, identity string, input closureexec.AdmittedInput, evaluationRoot, evaluationSubpath string, files []string, inventoryID closuregraph.ID) (PackageEvidence, error) {
	selected, err := SelectManifest(files, config.Toolchain.Swift.VersionOutput)
	if err != nil {
		return PackageEvidence{}, err
	}
	permit := manifestPermit(c0ID, identity, input, selected, config.Toolchain.SwiftPM, config.Destination)
	if err = recheckTool(ctx, config.Toolchain, config.Toolchain.SwiftPM); err != nil {
		return PackageEvidence{}, err
	}
	if err = recheckTool(ctx, config.Toolchain, config.Toolchain.PackageDescription); err != nil {
		return PackageEvidence{}, err
	}
	if config.ProcessStartObserver != nil {
		config.ProcessStartObserver(permit)
	}
	result, err := config.Evaluator.Evaluate(ctx, evaluationRoot, permit)
	if err != nil {
		return PackageEvidence{}, failFields(CodeManifestReplayDrift, map[string]string{"package": identity}, "manifest evaluation failed: %v", err)
	}
	if !result.ReceiptID.Valid() {
		return PackageEvidence{}, failFields(CodeDerivationUnauthorized, map[string]string{"package": identity}, "manifest evaluator omitted its issued derivation receipt")
	}
	manifest := result.Manifest
	manifest.SelectedManifest = selected
	manifest, err = normalizeManifest(manifest)
	if err != nil {
		return PackageEvidence{}, err
	}
	if err = reconcileManifestSources(manifest, files); err != nil {
		return PackageEvidence{}, err
	}
	digest, err := manifestDigest(manifest)
	if err != nil {
		return PackageEvidence{}, err
	}
	return PackageEvidence{SelectedManifest: selected, SnapshotDigest: input.Receipt.ContentSHA256, IntakeReceiptID: permit.IntakeReceiptID, ManifestPermitID: permit.ID, ManifestReceiptID: result.ReceiptID, ManifestDigest: digest, SourceInventoryDigest: inventoryID, Manifest: manifest, input: input, evaluationSubpath: evaluationSubpath}, nil
}

func captureLocalDependencies(ctx context.Context, config Config, c0ID closuregraph.ID, parentRoot string, parent PackageEvidence, artifactID closuregraph.ID, input closureexec.AdmittedInput, packages *[]PackageEvidence, byIdentity map[string]int) error {
	handleRoot, err := input.Tree.ProtectedPath()
	if err != nil {
		return err
	}
	for _, dependency := range parent.Manifest.Dependencies {
		if dependency.Kind != SourcePath {
			continue
		}
		candidate := filepath.Clean(filepath.Join(parentRoot, filepath.FromSlash(dependency.LocalPath)))
		relative, relErr := filepath.Rel(handleRoot, candidate)
		if relErr != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return failFields(CodeLocalDependencyOutside, map[string]string{"identity": dependency.Identity}, "local dependency escapes its admitted tree")
		}
		info, statErr := os.Lstat(candidate)
		if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return failFields(CodeLocalDependencyOutside, map[string]string{"identity": dependency.Identity}, "local dependency root is absent or linked")
		}
		logical := filepath.ToSlash(relative)
		if existing, exists := byIdentity[dependency.Identity]; exists {
			if (*packages)[existing].evaluationSubpath != logical {
				return failFields(CodeDependencyPinMismatch, map[string]string{"identity": dependency.Identity}, "local dependency identity maps to multiple roots")
			}
			continue
		}
		files, inventoryID, inventoryErr := inventoryTree(candidate)
		if inventoryErr != nil {
			return inventoryErr
		}
		pkg, evaluateErr := evaluateAdmittedAt(ctx, config, c0ID, dependency.Identity, input, candidate, logical, files, inventoryID)
		if evaluateErr != nil {
			return evaluateErr
		}
		if strings.ToLower(pkg.Manifest.PackageName) != dependency.Identity {
			return failFields(CodeDependencyPinMismatch, map[string]string{"expected": dependency.Identity, "observed": strings.ToLower(pkg.Manifest.PackageName)}, "local package identity differs from declaration")
		}
		pkg.Identity, pkg.Origin, pkg.Kind, pkg.ArtifactManifestID, pkg.SnapshotDigest = dependency.Identity, "workspace:"+logical, SourcePath, artifactID, inventoryID
		byIdentity[pkg.Identity] = len(*packages)
		*packages = append(*packages, pkg)
		if err = captureLocalDependencies(ctx, config, c0ID, candidate, pkg, artifactID, input, packages, byIdentity); err != nil {
			return err
		}
	}
	return nil
}

func manifestPermit(c0ID closuregraph.ID, identity string, input closureexec.AdmittedInput, selected string, tool ToolIdentity, destination Destination) ManifestPermit {
	receiptID, _ := input.Receipt.ID()
	argv := manifestArgv()
	environment := map[string]string{"HOME": "empty/home", "SWIFTPM_CONFIG_DIR": "empty/config", "SWIFTPM_SECURITY_DIR": "empty/security", "TZ": "UTC"}
	id, _ := closuregraph.DomainID("swiftpm-manifest-derivation-permit-v1", map[string]any{"argv": stringsAny(argv), "c0_checkpoint_id": string(c0ID), "environment": stringMapAny(environment), "intake_receipt_id": string(receiptID), "network": "none", "package_identity": identity, "selected_manifest": selected, "toolchain_fingerprint": string(tool.Fingerprint)})
	toolNode, _ := toolNodeRecord(tool)
	toolNodeID, _ := toolNode.ID()
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "swiftpm.platform.target", Payload: destination.Platform}
	platformID, _ := platform.ID()
	return ManifestPermit{ID: id, C0CheckpointID: c0ID, IntakeReceiptID: receiptID, PackageIdentity: identity, SelectedManifest: selected, ToolchainFingerprint: tool.Fingerprint, Argv: argv, Environment: environment, Network: "none", input: input, ToolchainNodeID: toolNodeID, HostID: platformID, TargetID: platformID}
}

func resolutionPermit(c0ID closuregraph.ID, input closureexec.AdmittedInput, tool ToolIdentity, destination Destination) ResolutionPermit {
	receiptID, _ := input.Receipt.ID()
	algorithmID := "swiftpm-brokered-resolution-v1"
	environment := map[string]string{"HOME": "empty/home", "SWIFTPM_CONFIG_DIR": "empty/config", "SWIFTPM_SECURITY_DIR": "empty/security", "TZ": "UTC"}
	id, _ := closuregraph.DomainID("swiftpm-resolution-derivation-permit-v1", map[string]any{"algorithm_id": algorithmID, "broker": "swiftpm-source-control-broker-v1", "c0_checkpoint_id": string(c0ID), "environment": stringMapAny(environment), "intake_receipt_id": string(receiptID), "network": "broker-only", "toolchain_fingerprint": string(tool.Fingerprint)})
	toolNode, _ := toolNodeRecord(tool)
	toolNodeID, _ := toolNode.ID()
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "swiftpm.platform.target", Payload: destination.Platform}
	platformID, _ := platform.ID()
	return ResolutionPermit{ID: id, C0CheckpointID: c0ID, IntakeReceiptID: receiptID, ToolchainFingerprint: tool.Fingerprint, AlgorithmID: algorithmID, Environment: environment, Network: "broker-only", input: input, ToolchainNodeID: toolNodeID, HostID: platformID, TargetID: platformID}
}

func manifestArgv() []string {
	return []string{"package", "--disable-experimental-prebuilts", "dump-package"}
}

func verifySuppliedRootLock(input closureexec.AdmittedInput, expected []byte) error {
	root, err := input.Tree.ProtectedPath()
	if err != nil {
		return err
	}
	candidates := []string{filepath.Join(root, "Package.resolved"), filepath.Join(root, ".swiftpm", "configuration", "Package.resolved")}
	found := 0
	for _, candidate := range candidates {
		payload, readErr := os.ReadFile(candidate) // #nosec G304 -- candidate is one exact contained root-lock location below the admitted protected tree.
		if readErr != nil {
			if os.IsNotExist(readErr) {
				continue
			}
			return readErr
		}
		found++
		if string(payload) != string(expected) {
			return fail(CodeResolvedFileOutOfDate, "supplied lock differs from admitted root Package.resolved")
		}
	}
	if found != 1 {
		return fail(CodeResolutionUnfrozen, "admitted root must contain exactly one supplied Package.resolved")
	}
	return nil
}

// SelectManifest deterministically applies SwiftPM's version-specific manifest
// rule and records the selected file before evaluation.
func SelectManifest(files []string, swiftVersion string) (string, error) {
	version := swiftVersionPattern.FindStringSubmatch(swiftVersion)
	if version == nil {
		return "", fail(CodeTargetPlatformUnsupported, "Swift version output cannot select a manifest")
	}
	tool, ok := parseVersion(version[1])
	if !ok {
		return "", fail(CodeTargetPlatformUnsupported, "Swift version is invalid")
	}
	base := false
	selected := ""
	selectedVersion := [3]int{-1, -1, -1}
	for _, file := range files {
		name := path.Base(file)
		if name == "Package.swift" && file == name {
			base = true
			continue
		}
		match := versionManifestPattern.FindStringSubmatch(name)
		if match == nil || file != name {
			continue
		}
		candidate, valid := parseVersion(match[1])
		if valid && compareVersion(candidate, tool) <= 0 && compareVersion(candidate, selectedVersion) > 0 {
			selected, selectedVersion = name, candidate
		}
	}
	if selected != "" {
		return selected, nil
	}
	if base {
		return "Package.swift", nil
	}
	return "", fail(CodeManifestReplayDrift, "package has no selectable manifest")
}

func inventoryTree(root string) ([]string, closuregraph.ID, error) {
	type leaf struct {
		Path, SHA256 string
		Size         int64
	}
	leaves := []leaf{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == root || entry.IsDir() {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeSourceInventoryDrift, "source inventory contains a link")
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return fail(CodeSourceInventoryDrift, "source inventory contains a special node")
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the rechecked protected package root.
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, current)
		logical := filepath.ToSlash(rel)
		sum := sha256.Sum256(payload)
		leaves = append(leaves, leaf{Path: logical, SHA256: "sha256:" + hex.EncodeToString(sum[:]), Size: int64(len(payload))})
		return nil
	})
	if err != nil {
		return nil, "", err
	}
	sort.Slice(leaves, func(i, j int) bool { return leaves[i].Path < leaves[j].Path })
	values, files := make([]any, len(leaves)), make([]string, len(leaves))
	for i, item := range leaves {
		files[i] = item.Path
		values[i] = map[string]any{"path": item.Path, "sha256": item.SHA256, "size": item.Size}
	}
	id, err := closuregraph.DomainID("swiftpm-package-tree-inventory-v1", map[string]any{"files": values})
	return files, id, err
}

func validateSnapshot(pin Pin, snapshot Snapshot) error {
	root, err := cleanAbsoluteRoot(snapshot.Root)
	rootInfo, rootStatErr := os.Lstat(snapshot.Root)
	if err != nil || root != snapshot.Root || rootStatErr != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "broker snapshot root is invalid")
	}
	mirror, err := cleanAbsoluteRoot(snapshot.MirrorRoot)
	mirrorInfo, mirrorStatErr := os.Lstat(snapshot.MirrorRoot)
	if err != nil || mirror != snapshot.MirrorRoot || mirrorStatErr != nil || !mirrorInfo.IsDir() || mirrorInfo.Mode()&os.ModeSymlink != 0 {
		return failFields(CodeDependencyMirrorMissing, map[string]string{"identity": pin.Identity}, "broker mirror root is invalid")
	}
	if strings.ToLower(snapshot.Identity) != pin.Identity || snapshot.Kind != pin.Kind || strings.ToLower(snapshot.Revision) != pin.Revision || !validRevision(snapshot.GitTree) || !snapshot.BrokerReceiptID.Valid() {
		return failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "broker snapshot identity, kind, revision, or tree differs from lock")
	}
	if snapshot.UsesSubmodules || snapshot.UsesLFS || snapshot.UsesCheckoutFilter || snapshot.RequiresHook {
		return failFields(CodeDependencyOriginUnsupported, map[string]string{"identity": pin.Identity}, "source-control snapshot requires an unsupported fetch or execution plane")
	}
	if err = scanUnsupportedGitShape(snapshot.Root); err != nil {
		return err
	}
	return nil
}

func scanUnsupportedGitShape(root string) error {
	return filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, current)
		logical := filepath.ToSlash(rel)
		if logical == ".git" || strings.HasPrefix(logical, ".git/") {
			return fail(CodeDependencyOriginUnsupported, "captured package tree contains Git administration data")
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if name == ".gitmodules" || name == ".lfsconfig" {
			return fail(CodeDependencyOriginUnsupported, "captured package requires submodule or LFS metadata")
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the validated broker snapshot root.
		if err != nil {
			return err
		}
		text := string(payload)
		if strings.HasPrefix(text, "version https://git-lfs.github.com/spec/v1") || name == ".gitattributes" && (strings.Contains(text, "filter=") || strings.Contains(text, "smudge=")) {
			return fail(CodeDependencyOriginUnsupported, "captured package requires LFS or checkout filters")
		}
		return nil
	})
}

func cleanAbsoluteRoot(value string) (string, error) {
	absolute, err := filepath.Abs(value)
	if err != nil || value == "" || absolute != filepath.Clean(value) {
		return "", fmt.Errorf("path must be absolute and clean")
	}
	return absolute, nil
}
func mustProtectedPath(input closureexec.AdmittedInput) string {
	value, err := input.Tree.ProtectedPath()
	if err != nil {
		panic(err)
	}
	return value
}
func parseVersion(value string) ([3]int, bool) {
	out := [3]int{}
	parts := strings.Split(value, ".")
	if len(parts) > 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
func compareVersion(a, b [3]int) int {
	for i := range a {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
func sortedIDs(values []closuregraph.ID) []closuregraph.ID {
	out := append([]closuregraph.ID{}, values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	unique := out[:0]
	for _, value := range out {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}
func stringsAny(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}
func idsAny(values []closuregraph.ID) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return out
}
func stringMapAny(values map[string]string) map[string]any {
	out := map[string]any{}
	for k, v := range values {
		out[k] = v
	}
	return out
}
func cloneMap(values map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range values {
		out[k] = v
	}
	return out
}
