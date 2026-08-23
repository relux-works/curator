package swiftpmsource

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/relux-works/curator/internal/closuregraph"
)

// ReplayOffline rechecks every admitted tree and C0 tool, repeats manifest
// selection/evaluation, and asks the protected SwiftPM metadata runner to
// resolve only from kind-preserving mirrors under network=none.
func (capture *Capture) ReplayOffline(ctx context.Context) error {
	if capture == nil {
		return fail(CodeOfflineReplayFailed, "offline SwiftPM capture is absent")
	}
	c0ID, err := capture.C0.ID()
	if err != nil {
		return err
	}
	for _, expected := range []ToolIdentity{capture.config.Toolchain.Swift, capture.config.Toolchain.SwiftPM, capture.config.Toolchain.PackageDescription, capture.config.Toolchain.Git} {
		if err = recheckTool(ctx, capture.config.Toolchain, expected); err != nil {
			return err
		}
	}
	replayedPackages := make([]PackageEvidence, len(capture.Packages))
	for index := range capture.Packages {
		pkg := &capture.Packages[index]
		if err = pkg.input.Tree.VerifyAtUse(); err != nil {
			return failFields(CodeManifestReplayDrift, map[string]string{"package": pkg.Identity}, "captured package changed before replay")
		}
		root, _ := pkg.input.Tree.ProtectedPath()
		if pkg.evaluationSubpath != "" && pkg.evaluationSubpath != "." {
			root = filepath.Join(root, filepath.FromSlash(pkg.evaluationSubpath))
		}
		files, inventoryID, inventoryErr := inventoryTree(root)
		if inventoryErr != nil || inventoryID != pkg.SourceInventoryDigest {
			return failFields(CodeSourceInventoryDrift, map[string]string{"package": pkg.Identity}, "package source inventory changed before replay")
		}
		replayed, replayErr := evaluateAdmittedAt(ctx, capture.config, c0ID, pkg.Identity, pkg.input, root, pkg.evaluationSubpath, files, inventoryID)
		if replayErr != nil {
			return replayErr
		}
		if replayed.SelectedManifest != pkg.SelectedManifest || replayed.ManifestDigest != pkg.ManifestDigest {
			return failFields(CodeManifestReplayDrift, map[string]string{"package": pkg.Identity}, "selected manifest or normalized dump-package evidence drifted")
		}
		replayedPackages[index] = *pkg
		replayedPackages[index].Manifest = replayed.Manifest
		replayedPackages[index].ManifestDigest = replayed.ManifestDigest
		replayedPackages[index].SourceInventoryDigest = replayed.SourceInventoryDigest
	}
	mirrors := append([]Mirror(nil), capture.Mirrors...)
	sort.Slice(mirrors, func(i, j int) bool { return mirrors[i].Identity < mirrors[j].Identity })
	if len(mirrors) != len(capture.Lock.Pins) {
		return fail(CodeDependencyMirrorMissing, "mirror set is not bijective with root lock")
	}
	for index, mirror := range mirrors {
		pin := capture.Lock.Pins[index]
		if mirror.Identity != pin.Identity || mirror.OriginalKind != pin.Kind || mirror.LocalKind != pin.Kind || mirror.Revision != pin.Revision || !mirror.BrokerReceiptID.Valid() || !mirror.MirrorIntakeReceiptID.Valid() || !mirror.MirrorDigest.Valid() || !packageBrokerReceiptMatches(capture.Packages, mirror.Identity, mirror.BrokerReceiptID) {
			return failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "offline mirror changed kind or revision")
		}
		if err = mirror.input.Tree.VerifyAtUse(); err != nil {
			return failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "captured mirror changed before replay")
		}
		protected, protectedErr := mirror.input.Tree.ProtectedPath()
		absolute, cleanErr := filepath.Abs(protected)
		info, statErr := os.Lstat(protected)
		if cleanErr != nil || absolute != filepath.Clean(mirror.Local) || statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return failFields(CodeDependencyMirrorMissing, map[string]string{"identity": pin.Identity}, "offline mirror is absent or mutable by indirection")
		}
		_, observedDigest, inventoryErr := inventoryMirrorTree(protected)
		if protectedErr != nil || inventoryErr != nil || observedDigest != mirror.MirrorDigest {
			return failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "offline mirror bytes drifted")
		}
		if mirror.authorization != nil {
			sourceReceiptID := packageIntakeReceiptID(capture.Packages, mirror.Identity)
			if err = capture.config.Policy.ValidateSourceControlMirrorAuthorization(mirror.authorization, ProfileID, string(pin.Kind), pin.CanonicalLocation, pin.Revision, mirror.GitTree, mirror.AuthorizedOutputPath, mirror.MirrorDigest, mirror.ArtifactManifestID, sourceReceiptID); err != nil {
				return failFields(CodeDerivationUnauthorized, map[string]string{"identity": pin.Identity}, "offline mirror authorization drifted")
			}
		}
		snapshot := Snapshot{Identity: mirror.Identity, MirrorRoot: protected, Revision: mirror.Revision, GitTree: mirror.GitTree, Kind: mirror.LocalKind, BrokerReceiptID: mirror.BrokerReceiptID}
		verification, verifyErr := verifyCapturedMirror(ctx, capture.config, protected, pin, snapshot, mirror.input)
		if verifyErr != nil {
			return failFields(CodeDependencyPinMismatch, map[string]string{"identity": pin.Identity}, "offline mirror no longer proves pinned revision and tree")
		}
		if !verificationIDsIssued(capture.config, verification) {
			return fail(CodeDerivationUnauthorized, "offline mirror verification receipts are not authority-issued")
		}
		mirrors[index].Local = protected
	}
	productNode, productFound := captureNode(capture, capture.ProductNodeID)
	if !productFound {
		return fail(CodeBuildGraphDrift, "selected product node is absent during offline replay")
	}
	rebuilt, err := buildGraph(replayedPackages, capture.Lock, productNode, capture.ProductNodeID)
	if err != nil {
		return fail(CodeBuildGraphDrift, "offline graph reconstruction failed: %v", err)
	}
	rebuiltID, err := rebuilt.graph.ID()
	if err != nil || rebuiltID != capture.GraphDigest {
		return fail(CodeBuildGraphDrift, "offline package/product/target/condition graph differs from C4 capture")
	}
	activeTargets, err := validateSelectedReachability(replayedPackages, capture.SelectionProduct(), capture.config.Destination.Markers)
	if err != nil {
		return err
	}
	rebuiltTargets := make([]closuregraph.ID, 0, len(activeTargets))
	for _, target := range activeTargets {
		id, exists := rebuilt.targetIDs[target]
		if !exists {
			return fail(CodeBuildGraphDrift, "offline selected target is absent")
		}
		rebuiltTargets = append(rebuiltTargets, id)
	}
	if got := sortedIDs(rebuiltTargets); !equalIDLists(got, capture.TargetNodeIDs) {
		return fail(CodeBuildGraphDrift, "offline selected target reachability differs from C4 capture")
	}
	inventoryID, err := closuregraph.DomainID("swiftpm-source-inventory-v1", map[string]any{"package_inventory_ids": idsAny(packageInventoryIDs(replayedPackages))})
	if err != nil || inventoryID != capture.InventoryDigest {
		return fail(CodeSourceInventoryDrift, "offline source inventory differs from captured package trees")
	}
	if capture.config.OfflineRunner == nil {
		return fail(CodeDerivationUnauthorized, "offline SwiftPM metadata runner is absent")
	}
	metadata, err := capture.config.OfflineRunner.Replay(ctx, capture)
	if err != nil {
		return err
	}
	if !metadata.ReceiptID.Valid() {
		return fail(CodeDerivationUnauthorized, "offline SwiftPM metadata replay omitted its issued receipt")
	}
	expectedIdentities := make([]string, len(capture.Lock.Pins))
	for index, pin := range capture.Lock.Pins {
		expectedIdentities[index] = pin.Identity
	}
	sort.Strings(expectedIdentities)
	if !equalStrings(expectedIdentities, metadata.PackageIdentities) {
		return fail(CodeBuildGraphDrift, "offline SwiftPM dependency graph differs from the frozen root lock")
	}
	return nil
}

func packageIntakeReceiptID(packages []PackageEvidence, identity string) closuregraph.ID {
	for _, pkg := range packages {
		if pkg.Identity == identity {
			id, _ := pkg.input.Receipt.ID()
			return id
		}
	}
	return ""
}

func captureNode(capture *Capture, id closuregraph.ID) (closuregraph.Node, bool) {
	for _, node := range capture.Records.CaptureNodes {
		nodeID, err := node.ID()
		if err == nil && nodeID == id {
			return node, true
		}
	}
	return closuregraph.Node{}, false
}

func equalIDLists(left, right []closuregraph.ID) bool {
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

func packageBrokerReceiptMatches(packages []PackageEvidence, identity string, receipt closuregraph.ID) bool {
	for _, pkg := range packages {
		if pkg.Identity == identity {
			return pkg.BrokerReceiptID == receipt
		}
	}
	return false
}

// SelectionProduct returns the exact selected command product name.
func (capture *Capture) SelectionProduct() string {
	if capture == nil {
		return ""
	}
	for _, node := range capture.Records.CaptureNodes {
		id, _ := node.ID()
		if id == capture.ProductNodeID {
			if payload, ok := node.Payload.(closuregraph.CommandProductPayload); ok {
				return payload.CommandKey
			}
		}
	}
	return ""
}
