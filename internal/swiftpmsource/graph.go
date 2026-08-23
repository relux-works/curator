package swiftpmsource

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closuregraph"
)

type graphBuild struct {
	graph             closuregraph.CaptureGraph
	nodes             []closuregraph.Node
	edges             []closuregraph.Edge
	targetIDs         map[string]closuregraph.ID
	selectedTargetIDs []closuregraph.ID
	conditionEdgeIDs  []closuregraph.ID
	brokerIDs         []closuregraph.ID
}

func provisionalProduct(name string, rootDigest closuregraph.ID) (closuregraph.Node, closuregraph.ID, error) {
	declaration, err := closuregraph.DomainID("swiftpm-product-declaration-v1", map[string]any{"name": name, "root_tree_digest": string(rootDigest), "type": "executable"})
	if err != nil {
		return closuregraph.Node{}, "", err
	}
	node := closuregraph.Node{Kind: closuregraph.NodeCommandProduct, LogicalKey: "swiftpm.product.root." + name, Payload: closuregraph.CommandProductPayload{Profile: ProfileID, SkillKey: "root", CommandKey: name, EntryPointContract: "swiftpm-executable-product-v1", DeclarationDigest: declaration, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget}}}
	id, err := node.ID()
	return node, id, err
}

func buildGraph(packages []PackageEvidence, _ Lock, selectedProduct closuregraph.Node, selectedProductID closuregraph.ID) (graphBuild, error) {
	nodes := []closuregraph.Node{}
	edges := []closuregraph.Edge{}
	packageIDs := map[string]closuregraph.ID{}
	sourceIDs := map[string]closuregraph.ID{}
	targetIDs := map[string]closuregraph.ID{}
	productIDs := map[string]closuregraph.ID{}
	manifestIDs := []closuregraph.ID{}
	roots := []closuregraph.ID{}
	for packageIndex, pkg := range packages {
		version := pkg.Revision
		if version == "" {
			version = "workspace"
		}
		if pkg.Manifest.PackageName == "" {
			return graphBuild{}, fail(CodeGraphIncomplete, "package manifest identity is absent")
		}
		packageNode := closuregraph.Node{Kind: closuregraph.NodePackageInstance, LogicalKey: "swiftpm.package." + pkg.Identity, Payload: closuregraph.PackageInstancePayload{Profile: ProfileID, Ecosystem: "swift", Manager: "swiftpm", NormalizedSourceID: pkg.Origin + "#" + pkg.Revision, Origin: pkg.Origin, LockInstanceKey: pkg.Identity + "@" + version, Name: pkg.Identity, Version: version, ArtifactManifestID: pkg.ArtifactManifestID, SnapshotDigest: pkg.SnapshotDigest, TrustRole: closuregraph.TrustDependencyInput}}
		packageID, err := packageNode.ID()
		if err != nil {
			return graphBuild{}, err
		}
		packageIDs[pkg.Identity] = packageID
		nodes = append(nodes, packageNode)
		sourceNode := closuregraph.Node{Kind: closuregraph.NodeSourceSet, LogicalKey: "swiftpm.source." + pkg.Identity, Payload: closuregraph.SourceSetPayload{Profile: ProfileID, Origin: pkg.Origin, ArtifactManifestID: pkg.ArtifactManifestID, Projection: []string{}, Grammar: "swiftpm-package-tree-v1", TrustRole: closuregraph.TrustDependencyInput, SourceClass: "source.tree", TreeDigest: pkg.SnapshotDigest}}
		sourceID, err := sourceNode.ID()
		if err != nil {
			return graphBuild{}, err
		}
		sourceIDs[pkg.Identity] = sourceID
		nodes = append(nodes, sourceNode)
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeResolvesTo, EdgeKey: "swiftpm.resolves." + pkg.Identity, FromNodeID: packageID, ToNodeID: sourceID, Payload: closuregraph.ResolvesToPayload{LockField: "pins." + pkg.Identity, Origin: closuregraph.EvidenceOrigin{Field: "Package.resolved.pins." + pkg.Identity, ManifestDigest: pkg.ManifestDigest}, Checksum: string(pkg.SnapshotDigest), ArtifactManifestID: pkg.ArtifactManifestID}})
		manifestIDs = append(manifestIDs, pkg.ArtifactManifestID)
		for _, target := range pkg.Manifest.Targets {
			declaration, _ := targetDeclarationDigest(pkg, target)
			languages := targetLanguages(target)
			node := closuregraph.Node{Kind: closuregraph.NodeTargetUnit, LogicalKey: "swiftpm.target." + pkg.Identity + "." + target.Name, Payload: closuregraph.TargetUnitPayload{Profile: ProfileID, TargetName: target.Name, TargetKind: target.Type, DeclarationDigest: declaration, Languages: languages, ExecutionDomain: targetDomain(target), ConditionExpressions: []closuregraph.Condition{}, ExpectedOutputClass: "native.object", PlatformRoleNames: []closuregraph.PlatformRole{targetPlatformRole(target)}}}
			id, err := node.ID()
			if err != nil {
				return graphBuild{}, err
			}
			targetIDs[pkg.Identity+":"+target.Name] = id
			nodes = append(nodes, node)
			edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeDeclares, EdgeKey: "swiftpm.package-target." + pkg.Identity + "." + target.Name, FromNodeID: packageID, ToNodeID: id, Payload: closuregraph.DeclaresPayload{Origin: closuregraph.EvidenceOrigin{Field: "targets." + target.Name, ManifestDigest: pkg.ManifestDigest}}})
		}
		for _, product := range pkg.Manifest.Products {
			var node closuregraph.Node
			if packageIndex == 0 && product.Name == selectedProduct.Payload.(closuregraph.CommandProductPayload).CommandKey {
				node = selectedProduct
			} else {
				declaration, _ := closuregraph.DomainID("swiftpm-product-declaration-v1", map[string]any{"manifest_digest": string(pkg.ManifestDigest), "name": product.Name, "package": pkg.Identity, "targets": stringsAny(product.Targets), "type": product.Type})
				node = closuregraph.Node{Kind: closuregraph.NodeCommandProduct, LogicalKey: "swiftpm.product." + pkg.Identity + "." + product.Name, Payload: closuregraph.CommandProductPayload{Profile: ProfileID, SkillKey: pkg.Identity, CommandKey: product.Name, EntryPointContract: "swiftpm-" + product.Type + "-product-v1", DeclarationDigest: declaration, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformTarget}}}
			}
			id, err := node.ID()
			if err != nil {
				return graphBuild{}, err
			}
			if packageIndex == 0 && product.Name == selectedProduct.Payload.(closuregraph.CommandProductPayload).CommandKey && id != selectedProductID {
				return graphBuild{}, fail(CodeGraphReferenceInvalid, "provisional selected product identity drifted")
			}
			productIDs[pkg.Identity+":"+product.Name] = id
			nodes = append(nodes, node)
			if packageIndex == 0 {
				roots = append(roots, id)
			}
			for _, targetName := range product.Targets {
				targetID, ok := targetIDs[pkg.Identity+":"+targetName]
				if !ok {
					return graphBuild{}, fail(CodeGraphIncomplete, "product %s names absent target %s", product.Name, targetName)
				}
				edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: "swiftpm.product-target." + pkg.Identity + "." + product.Name + "." + targetName, FromNodeID: id, ToNodeID: targetID, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeRuntime, Origin: closuregraph.EvidenceOrigin{Field: "products." + product.Name + ".targets", ManifestDigest: pkg.ManifestDigest}, DependencyKind: "target"}})
			}
		}
	}
	conditionIDs := []closuregraph.ID{}
	for _, pkg := range packages {
		fromPackage := packageIDs[pkg.Identity]
		for index, dependency := range pkg.Manifest.Dependencies {
			target, ok := packageIDs[dependency.Identity]
			if !ok {
				return graphBuild{}, fail(CodeGraphIncomplete, "package dependency %s is absent", dependency.Identity)
			}
			kind := "source-control"
			scope := closuregraph.ScopeRuntime
			if dependency.Kind == SourcePath {
				kind = "local-path"
				scope = closuregraph.ScopeWorkspace
			}
			edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: fmt.Sprintf("swiftpm.package-dependency.%s.%04d", pkg.Identity, index), FromNodeID: fromPackage, ToNodeID: target, Payload: closuregraph.RequiresPayload{Scope: scope, Origin: closuregraph.EvidenceOrigin{Field: "dependencies." + dependency.Identity, ManifestDigest: pkg.ManifestDigest}, DependencyKind: kind}})
		}
		for _, target := range pkg.Manifest.Targets {
			from := targetIDs[pkg.Identity+":"+target.Name]
			for index, dependency := range target.Dependencies {
				packageID := pkg.Identity
				if dependency.Package != "" {
					packageID = strings.ToLower(dependency.Package)
				}
				if dependency.Product != "" {
					manifestIndex, exists := packageIndexByIdentity(packages, packageID)
					if !exists {
						return graphBuild{}, fail(CodeGraphIncomplete, "product dependency package %s is absent", packageID)
					}
					product, exists := findProduct(packages[manifestIndex].Manifest, dependency.Product)
					if !exists {
						return graphBuild{}, fail(CodeGraphIncomplete, "product dependency %s:%s is absent", packageID, dependency.Product)
					}
					for productIndex, targetName := range product.Targets {
						to, present := targetIDs[packageID+":"+targetName]
						if !present {
							return graphBuild{}, fail(CodeGraphIncomplete, "product target %s:%s is absent", packageID, targetName)
						}
						edge := closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: fmt.Sprintf("swiftpm.target-product-dependency.%s.%s.%04d.%04d", pkg.Identity, target.Name, index, productIndex), FromNodeID: from, ToNodeID: to, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeBuild, Condition: dependency.Condition, Origin: closuregraph.EvidenceOrigin{Field: "targets." + target.Name + ".dependencies", ManifestDigest: pkg.ManifestDigest}, DependencyKind: "product-target"}}
						edges = append(edges, edge)
						if dependency.Condition != nil {
							id, _ := edge.ID()
							conditionIDs = append(conditionIDs, id)
						}
					}
					continue
				}
				to, ok := targetIDs[packageID+":"+dependency.Name]
				if !ok {
					return graphBuild{}, fail(CodeGraphIncomplete, "target dependency %s:%s is absent", packageID, dependency.Name)
				}
				edge := closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: fmt.Sprintf("swiftpm.target-dependency.%s.%s.%04d", pkg.Identity, target.Name, index), FromNodeID: from, ToNodeID: to, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeBuild, Condition: dependency.Condition, Origin: closuregraph.EvidenceOrigin{Field: "targets." + target.Name + ".dependencies", ManifestDigest: pkg.ManifestDigest}, DependencyKind: "target"}}
				edges = append(edges, edge)
				if dependency.Condition != nil {
					id, _ := edge.ID()
					conditionIDs = append(conditionIDs, id)
				}
			}
		}
	}
	graph, err := closuregraph.NewCaptureGraph(ProfileID, []string{"curator-artifact-policy-v1", "swiftpm-source-closure-v1"}, roots, nodes, edges, sortedIDs(manifestIDs))
	if err != nil {
		return graphBuild{}, err
	}
	brokerIDs := []closuregraph.ID{}
	for _, pkg := range packages {
		if pkg.BrokerReceiptID != "" {
			brokerIDs = append(brokerIDs, pkg.BrokerReceiptID)
		}
	}
	return graphBuild{graph: graph, nodes: nodes, edges: edges, targetIDs: targetIDs, conditionEdgeIDs: sortedIDs(conditionIDs), brokerIDs: sortedIDs(brokerIDs)}, nil
}

func bindingRecords(config Config) (closuregraph.Node, closuregraph.ID, []closuregraph.Node, []closuregraph.ID, error) {
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "swiftpm.platform.target", Payload: config.Destination.Platform}
	platformID, err := platform.ID()
	if err != nil {
		return closuregraph.Node{}, "", nil, nil, err
	}
	tools := []ToolIdentity{config.Toolchain.Swift, config.Toolchain.SwiftPM, config.Toolchain.PackageDescription, config.Toolchain.Git}
	nodes := make([]closuregraph.Node, len(tools))
	ids := make([]closuregraph.ID, len(tools))
	for index, tool := range tools {
		node, nodeErr := toolNodeRecord(tool)
		if nodeErr != nil {
			return closuregraph.Node{}, "", nil, nil, nodeErr
		}
		id, nodeErr := node.ID()
		if nodeErr != nil {
			return closuregraph.Node{}, "", nil, nil, nodeErr
		}
		nodes[index], ids[index] = node, id
	}
	return platform, platformID, nodes, ids, nil
}

func toolNodeRecord(tool ToolIdentity) (closuregraph.Node, error) {
	fingerprints := []closuregraph.ID{tool.ExecutableSHA256}
	for _, process := range tool.ProcessFamily {
		fingerprints = append(fingerprints, process.ExecutableSHA256)
	}
	fingerprints = sortedIDs(fingerprints)
	node := closuregraph.Node{Kind: closuregraph.NodeToolchainComponent, LogicalKey: "swiftpm.tool." + tool.Role, Payload: closuregraph.ToolchainComponentPayload{ComponentRole: tool.Role, ContentFingerprint: tool.Fingerprint, ExecutableRelativePath: tool.ExecutableRelativePath, PlatformABI: tool.PlatformABI, PolicySelector: tool.PolicySelector, VersionOutput: tool.VersionOutput, LinkFingerprintIDs: fingerprints, TimeOfUseRecheckRule: "immediate-exact-v1", ExecutionDomain: closuregraph.ExecutionHost, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformHost}}}
	_, err := node.ID()
	return node, err
}

func bindSelection(graph graphBuild, selection closuregraph.SelectionContext, c0 closuregraph.Checkpoint, platform closuregraph.Node, tools []closuregraph.Node, toolIDs []closuregraph.ID) (closuregraph.SelectionBinding, []closuregraph.Node, []closuregraph.Edge, closuregraph.BindingAuthority, error) {
	captureID, _ := graph.graph.ID()
	selectionID, _ := selection.ID()
	platformID, _ := platform.ID()
	nodes := append([]closuregraph.Node{platform}, tools...)
	edges := []closuregraph.Edge{}
	targets := append([]closuregraph.ID{selection.ProductNodeIDs[0]}, graph.selectedTargetIDs...)
	for index, id := range targets {
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("swiftpm.selection-target.%04d", index), FromNodeID: id, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformTarget, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.target"}}})
	}
	for index, toolID := range toolIDs {
		edges = append(edges, closuregraph.Edge{Kind: closuregraph.EdgeRequires, EdgeKey: fmt.Sprintf("swiftpm.selection-tool.%04d", index), FromNodeID: selection.ProductNodeIDs[0], ToNodeID: toolID, Payload: closuregraph.RequiresPayload{Scope: closuregraph.ScopeToolchain, Origin: closuregraph.EvidenceOrigin{Field: "toolchain." + tools[index].LogicalKey}, DependencyKind: "external-toolchain"}}, closuregraph.Edge{Kind: closuregraph.EdgeTargets, EdgeKey: fmt.Sprintf("swiftpm.selection-tool-target.%04d", index), FromNodeID: toolID, ToNodeID: platformID, Payload: closuregraph.TargetsPayload{BindingRole: closuregraph.PlatformHost, Origin: closuregraph.EvidenceOrigin{Field: "selection.platform_roles.host"}}})
	}
	binding, err := closuregraph.NewSelectionBinding(captureID, selectionID, nodes, edges)
	if err != nil {
		return closuregraph.SelectionBinding{}, nil, nil, closuregraph.BindingAuthority{}, err
	}
	c0ID, _ := c0.ID()
	authority := closuregraph.BindingAuthority{C0Checkpoint: &c0}
	for _, id := range toolIDs {
		authority.Toolchains = append(authority.Toolchains, closuregraph.ToolchainBindingEvidence{NodeID: id, FirstBound: closuregraph.ToolchainBoundAtC0, EvidenceID: c0ID})
	}
	sort.Slice(authority.Toolchains, func(i, j int) bool { return authority.Toolchains[i].NodeID < authority.Toolchains[j].NodeID })
	return binding, nodes, edges, authority, nil
}

func evidenceIDs(packages []PackageEvidence) (manifests, intakes, origins, handles, journal, derivations []closuregraph.ID) {
	for _, pkg := range packages {
		manifests = append(manifests, pkg.ArtifactManifestID)
		intakes = append(intakes, pkg.IntakeReceiptID)
		handles = append(handles, closuregraph.ID(pkg.input.Receipt.ProtectedHandleID))
		journal = append(journal, pkg.ManifestPermitID, pkg.ManifestReceiptID)
		derivations = append(derivations, pkg.ManifestReceiptID)
		journal = append(journal, pkg.BrokerPermitIDs...)
		journal = append(journal, pkg.BrokerProcessReceiptIDs...)
		derivations = append(derivations, pkg.BrokerProcessReceiptIDs...)
		if pkg.Mirror != nil {
			manifests = append(manifests, pkg.Mirror.ArtifactManifestID)
			intakes = append(intakes, pkg.Mirror.MirrorIntakeReceiptID)
			handles = append(handles, closuregraph.ID(pkg.Mirror.input.Receipt.ProtectedHandleID))
			if pkg.Mirror.CommitEvidenceIntakeReceiptID.Valid() {
				manifests = append(manifests, pkg.Mirror.CommitEvidenceArtifactManifestID)
				intakes = append(intakes, pkg.Mirror.CommitEvidenceIntakeReceiptID)
				handles = append(handles, closuregraph.ID(pkg.Mirror.commitInput.Receipt.ProtectedHandleID))
				journal = append(journal, pkg.Mirror.MirrorDerivationPermitID, pkg.Mirror.MirrorDerivationReceiptID)
				derivations = append(derivations, pkg.Mirror.MirrorDerivationReceiptID)
			}
			journal = append(journal, pkg.Mirror.VerificationPermitIDs...)
			journal = append(journal, pkg.Mirror.VerificationReceiptIDs...)
			derivations = append(derivations, pkg.Mirror.VerificationReceiptIDs...)
		}
		origin, _ := closuregraph.DomainID("swiftpm-immutable-origin-v1", map[string]any{"git_tree": pkg.GitTree, "identity": pkg.Identity, "kind": string(pkg.Kind), "origin": pkg.Origin, "revision": pkg.Revision, "snapshot_digest": string(pkg.SnapshotDigest)})
		origins = append(origins, origin)
	}
	return sortedIDs(manifests), sortedIDs(intakes), sortedIDs(origins), sortedIDs(handles), sortedIDs(journal), sortedIDs(derivations)
}

func targetDeclarationDigest(pkg PackageEvidence, target Target) (closuregraph.ID, error) {
	return closuregraph.DomainID("swiftpm-target-declaration-v1", map[string]any{"manifest_digest": string(pkg.ManifestDigest), "name": target.Name, "package": pkg.Identity, "path": target.Path, "sources": stringsAny(target.Sources), "type": target.Type})
}
func targetLanguages(target Target) []string {
	values := []string{}
	for _, source := range target.Sources {
		switch strings.ToLower(filepath.Ext(source)) {
		case ".swift":
			values = append(values, "swift")
		case ".c":
			values = append(values, "c")
		case ".cc", ".cpp", ".cxx":
			values = append(values, "c++")
		case ".m":
			values = append(values, "objective-c")
		case ".mm":
			values = append(values, "objective-c++")
		}
	}
	return sortedUnique(values)
}
func targetDomain(target Target) closuregraph.ExecutionDomain {
	if target.Type == "plugin" || target.Type == "macro" {
		return closuregraph.ExecutionHost
	}
	return closuregraph.ExecutionTarget
}
func targetPlatformRole(target Target) closuregraph.PlatformRole {
	if targetDomain(target) == closuregraph.ExecutionHost {
		return closuregraph.PlatformHost
	}
	return closuregraph.PlatformTarget
}

func packageIndexByIdentity(packages []PackageEvidence, identity string) (int, bool) {
	for index, pkg := range packages {
		if pkg.Identity == identity {
			return index, true
		}
	}
	return 0, false
}
