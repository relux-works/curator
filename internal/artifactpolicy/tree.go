package artifactpolicy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
)

type capturedTreeEntry struct {
	relative string
	path     VirtualPath
	kind     NodeKind
	mode     int64
	data     *blob
	size     int64
	sha256   string
	reason   string
}

type treeIdentityRecord struct {
	Path       string `json:"path"`
	Kind       string `json:"kind"`
	Executable bool   `json:"executable"`
	Size       int64  `json:"size"`
	SHA256     string `json:"sha256"`
}

// AdmitDependencyDirectory recursively inspects an immutable captured tree.
// It never follows a link and it rechecks every opened regular file against
// the lstat record used during enumeration.
func (service *Service) AdmitDependencyDirectory(ctx context.Context, request DirectoryRequest) (Result, error) {
	configured, configurationErr := configuredService(service)
	if configurationErr != nil {
		return Result{}, configurationErr
	}
	service = configured
	if err := validateDescriptor(request.Descriptor); err != nil {
		payload := Payload{Path: request.VirtualRoot}
		return service.failureBeforeCapture(RoleDependencyInput, request.Descriptor, payload, CodePolicyInternalError, "invalid_descriptor:"+err.Error())
	}
	virtualRoot, err := validateVirtualPath(request.VirtualRoot, service.limits)
	if err != nil {
		payload := Payload{Path: request.VirtualRoot}
		return service.failureBeforeCapture(RoleDependencyInput, request.Descriptor, payload, CodeArchiveUnsafePath, err.Error())
	}
	rootInfo, err := os.Lstat(request.Root)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&fs.ModeSymlink != 0 {
		payload := Payload{Path: request.VirtualRoot}
		return service.failureBeforeCapture(RoleDependencyInput, request.Descriptor, payload, CodeArchiveUnsafeEntry, "directory_root_is_not_an_ordinary_directory")
	}
	root, err := os.OpenRoot(request.Root)
	if err != nil {
		payload := Payload{Path: request.VirtualRoot}
		return service.failureBeforeCapture(RoleDependencyInput, request.Descriptor, payload, CodeInspectionUnavailable, "open_directory_root")
	}
	defer func() { _ = root.Close() }()
	openedRoot, err := root.Stat(".")
	if err != nil || !openedRoot.IsDir() || !os.SameFile(rootInfo, openedRoot) {
		payload := Payload{Path: request.VirtualRoot}
		return service.failureBeforeCapture(RoleDependencyInput, request.Descriptor, payload, CodeInspectionUnavailable, "directory_root_changed_while_opening")
	}
	store, err := newBlobStore()
	if err != nil {
		payload := Payload{Path: request.VirtualRoot}
		return service.failureBeforeCapture(RoleDependencyInput, request.Descriptor, payload, CodeInspectionUnavailable, "create_private_inspection_store")
	}
	defer store.close()
	account, _ := newLimitAccountant(service.limits, 0)
	worker := &inspector{
		ctx: ctx, role: RoleDependencyInput, roleAuthorized: true,
		descriptor: request.Descriptor, limits: service.limits,
		account: account, store: store,
		findings:          newFindingAccumulator(service.limits.MaxRecordedFindings),
		authorizationSeal: service.authorizationSeal,
	}
	if err := worker.account.addEntry(1); err != nil {
		payload := Payload{Path: request.VirtualRoot}
		return service.failureBeforeCapture(RoleDependencyInput, request.Descriptor, payload, CodeInspectionLimitExceeded, err.Error())
	}
	entries, treeInspectionComplete := worker.captureTree(root, virtualRoot.Canonical)
	identity, totalSize, identityErr := canonicalTreeIdentity(entries)
	if identityErr != nil {
		worker.addDiagnostic(unavailableDiagnostic(virtualRoot.Canonical, nil, "canonical_tree_identity", identityErr))
		identity = digestBytes(nil)
	}
	if reason := validateTreeOrigin(request.Descriptor.Origin, identity); reason != "" {
		worker.addDiagnostic(Diagnostic{
			Code: CodeOriginUnverified, Path: virtualRoot.Canonical,
			OriginalNameBase64: originalNameBase64(request.VirtualRoot),
			CollisionKey:       virtualRoot.CollisionKey, SHA256: identity,
			Size: totalSize, Reason: reason,
		})
	}
	rootNode := ManifestNode{
		Path: virtualRoot.Canonical, OriginalNameBase64: originalNameBase64(request.VirtualRoot),
		CollisionKey: virtualRoot.CollisionKey, Kind: NodeDirectory,
		Size: 0, Class: ClassDirectory, Decision: DecisionDescend, Rule: "descend_directory",
		InspectionComplete: treeInspectionComplete,
	}
	worker.nodes = append(worker.nodes, rootNode)
	rejected := worker.findingCount() > 0
	for _, entry := range entries {
		fullPath := joinTreePath(virtualRoot.Canonical, entry.path.Canonical)
		parentPath := virtualRoot.Canonical
		if parent := path.Dir(entry.path.Canonical); parent != "." {
			parentPath = joinTreePath(virtualRoot.Canonical, parent)
		}
		switch entry.kind {
		case NodeDirectory:
			class := ClassDirectory
			decision := DecisionDescend
			rule := "descend_directory"
			selectedDetector := ""
			observations := []Observation{}
			if bundleClass, bundle := appleBundleClass(fullPath); bundle {
				class = bundleClass
				decision = DecisionReject
				rule = "apple_bundle_forbidden"
				selectedDetector = "apple-bundle-path-v1"
				observations = []Observation{{
					DetectorID: selectedDetector, Result: "MATCH",
					Facts: []Fact{{Key: "bundle_path", Value: fullPath}},
				}}
			}
			node := ManifestNode{
				Path: fullPath, OriginalNameBase64: originalNameBase64(entry.relative),
				CollisionKey: portableCollisionKey(fullPath), Kind: NodeDirectory,
				Parent: parentPath, Size: entry.size, Mode: entry.mode, Observations: observations,
				SelectedDetectorID: selectedDetector,
				Class:              class, Decision: decision, Rule: rule, InspectionComplete: true,
			}
			worker.nodes = append(worker.nodes, node)
			if decision == DecisionReject {
				worker.addDiagnostic(worker.classDiagnostic(node, "apple_bundle_forbidden"))
				rejected = true
			}
		case NodeLink, NodeSpecial:
			class := ClassLink
			if entry.kind == NodeSpecial {
				class = ClassSpecial
			}
			node := ManifestNode{
				Path: fullPath, OriginalNameBase64: originalNameBase64(entry.relative),
				CollisionKey: portableCollisionKey(fullPath), Kind: entry.kind,
				Parent: parentPath, Mode: entry.mode, Size: entry.size,
				Class: class, Decision: DecisionReject, Rule: "unsafe_tree_entry", InspectionComplete: true,
			}
			worker.nodes = append(worker.nodes, node)
			worker.addDiagnostic(Diagnostic{
				Code: CodeArchiveUnsafeEntry, Path: fullPath,
				OriginalNameBase64: node.OriginalNameBase64,
				CollisionKey:       node.CollisionKey, Class: class, Size: entry.size,
				Reason: entry.reason,
			})
			rejected = true
		case NodeRegularFile:
			if entry.data == nil {
				worker.addDiagnostic(unavailableDiagnostic(fullPath, nil, "captured_tree_file_missing", nil))
				rejected = true
				continue
			}
			if worker.inspectBlob(fullPath, entry.relative, parentPath, nil, *entry.data, 1, entry.mode, nil) {
				rejected = true
			}
		default:
			worker.addDiagnostic(Diagnostic{Code: CodePolicyInternalError, Path: fullPath, Reason: "unknown_tree_entry_kind"})
			rejected = true
		}
	}
	if rejected || worker.findingCount() > 0 {
		worker.forceRootReject(virtualRoot.Canonical, "tree_member_rejected")
	}
	accounting, accountingErr := bindTraversalAccounting(worker.account.snapshot(), "canonical_tree", worker.nodes)
	if accountingErr != nil {
		return Result{}, fmt.Errorf("bind tree traversal accounting evidence: %w", accountingErr)
	}
	manifest := Manifest{
		SchemaID: SchemaID, PolicyID: PolicyID, PolicyVersion: PolicyVersion,
		LimitVector: service.limits, DetectorRegistryID: DetectorRegistryID,
		Detectors: append([]DetectorIdentity(nil), service.detectors...),
		AdapterID: request.Descriptor.AdapterID, ProfileID: request.Descriptor.ProfileID,
		Manager: request.Descriptor.Manager, PackageName: request.Descriptor.PackageName,
		PackageVersion: request.Descriptor.PackageVersion, Origin: request.Descriptor.Origin,
		TrustRole: RoleDependencyInput, RoleEvidence: dependencyRoleFacts(request.Descriptor.Origin),
		RawPayload: RawPayloadEvidence{
			Path: virtualRoot.Canonical, Size: totalSize, SHA256: identity, Kind: "canonical_tree",
		},
		Accounting: accounting,
		Nodes:      worker.nodes, Decision: DecisionAdmitInput,
	}
	if rejected || worker.findingCount() > 0 {
		manifest.Decision = DecisionReject
	}
	return finishInspectorResult(manifest, worker)
}

func (inspector *inspector) captureTree(root *os.Root, virtualRoot string) ([]capturedTreeEntry, bool) {
	entries := make([]capturedTreeEntry, 0)
	inspectionComplete := true
	collisions := map[string]string{}
	err := fs.WalkDir(root.FS(), ".", func(relative string, directory fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			inspector.addDiagnostic(unavailableDiagnostic(virtualRoot, nil, "walk_captured_tree", walkErr))
			inspectionComplete = false
			return fs.SkipAll
		}
		if relative == "." {
			return nil
		}
		if accountErr := inspector.account.addEntry(1); accountErr != nil {
			inspector.addDiagnostic(diagnosticFromError(virtualRoot, nil, accountErr))
			inspectionComplete = false
			return fs.SkipAll
		}
		if err := contextError(inspector.ctx); err != nil {
			inspector.addDiagnostic(unavailableDiagnostic(virtualRoot, nil, "walk_cancelled", err))
			inspectionComplete = false
			return fs.SkipAll
		}
		portable := strings.TrimPrefix(filepathToSlash(relative), "./")
		validated, err := validateVirtualPath(portable, inspector.limits)
		if err != nil {
			inspector.addDiagnostic(treePathDiagnostic(virtualRoot, portable, err))
			if directory.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if _, err := validateVirtualPath(joinTreePath(virtualRoot, validated.Canonical), inspector.limits); err != nil {
			inspector.addDiagnostic(treePathDiagnostic(virtualRoot, portable, err))
			if directory.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if existing, collision := collisions[validated.CollisionKey]; collision && existing != validated.Canonical {
			inspector.addDiagnostic(Diagnostic{
				Code: CodeArchiveUnsafePath, Path: joinTreePath(virtualRoot, validated.Canonical),
				OriginalNameBase64: originalNameBase64(portable), CollisionKey: validated.CollisionKey,
				Reason: "portable_path_collision", Details: []Fact{{Key: "collides_with", Value: existing}},
			})
			if directory.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		collisions[validated.CollisionKey] = validated.Canonical
		info, err := directory.Info()
		if err != nil {
			inspector.addDiagnostic(unavailableDiagnostic(joinTreePath(virtualRoot, portable), nil, "lstat_tree_entry", err))
			inspectionComplete = false
			if directory.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		entry := capturedTreeEntry{relative: portable, path: validated, mode: int64(info.Mode()), size: info.Size()}
		switch {
		case info.IsDir():
			entry.kind = NodeDirectory
		case info.Mode()&fs.ModeSymlink != 0:
			entry.kind = NodeLink
			entry.reason = "filesystem_symlink"
		case info.Mode().IsRegular():
			entry.kind = NodeRegularFile
			if limitErr := inspector.account.checkLeaf(info.Size()); limitErr != nil {
				inspector.addDiagnostic(diagnosticFromError(
					joinTreePath(virtualRoot, portable), nil, limitErr,
				))
				inspectionComplete = false
				return fs.SkipAll
			}
			captured, captureErr := inspector.captureTreeFile(root, portable, info)
			if captureErr != nil {
				var limit *limitFailure
				if errorAs(captureErr, &limit) {
					inspector.addDiagnostic(diagnosticFromError(joinTreePath(virtualRoot, portable), nil, captureErr))
				} else {
					inspector.addDiagnostic(unavailableDiagnostic(joinTreePath(virtualRoot, portable), nil, "capture_tree_file", captureErr))
				}
				inspectionComplete = false
				return fs.SkipAll
			}
			if captured.kind == NodeLink {
				entry.kind = NodeLink
				entry.reason = captured.reason
			} else {
				entry.data = captured.data
				entry.sha256 = captured.sha256
				entry.size = captured.size
			}
		default:
			entry.kind = NodeSpecial
			entry.reason = "filesystem_special_node"
		}
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		inspector.addDiagnostic(unavailableDiagnostic(virtualRoot, nil, "walk_captured_tree", err))
		inspectionComplete = false
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].path.Canonical < entries[right].path.Canonical
	})
	return entries, inspectionComplete
}

func (inspector *inspector) captureTreeFile(root *os.Root, relative string, expected fs.FileInfo) (capturedTreeEntry, error) {
	file, err := root.Open(relative)
	if err != nil {
		return capturedTreeEntry{}, err
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(expected, opened) || opened.Size() != expected.Size() {
		return capturedTreeEntry{}, fmt.Errorf("tree file changed while opening")
	}
	links, known, err := regularFileLinkCount(file, opened)
	if err != nil {
		return capturedTreeEntry{}, err
	}
	if !known {
		return capturedTreeEntry{}, fmt.Errorf("hard-link count inspection unavailable")
	}
	if links != 1 {
		return capturedTreeEntry{kind: NodeLink, size: opened.Size(), reason: "filesystem_hardlink"}, nil
	}
	if err := inspector.account.addRaw(opened.Size()); err != nil {
		return capturedTreeEntry{}, err
	}
	captured, err := inspector.store.captureRoot(inspector.ctx, Payload{
		Path: relative, Size: opened.Size(), Reader: file,
	}, inspector.limits)
	if err != nil {
		return capturedTreeEntry{}, err
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(opened, after) || after.Size() != opened.Size() || after.ModTime() != opened.ModTime() {
		return capturedTreeEntry{}, fmt.Errorf("tree file changed while reading")
	}
	return capturedTreeEntry{kind: NodeRegularFile, data: &captured, size: captured.size, sha256: captured.sha256}, nil
}

func canonicalTreeIdentity(entries []capturedTreeEntry) (string, int64, error) {
	records := make([]treeIdentityRecord, 0, len(entries))
	for _, entry := range entries {
		records = append(records, treeIdentityRecord{
			Path: entry.path.Canonical, Kind: string(entry.kind),
			Executable: entry.mode&0o111 != 0, Size: entry.size, SHA256: entry.sha256,
		})
	}
	return canonicalTreeRecordIdentity(records)
}

func canonicalTreeManifestIdentity(root string, nodes []ManifestNode) (string, int64, error) {
	prefix := root + "/"
	records := make([]treeIdentityRecord, 0, len(nodes))
	for _, node := range nodes {
		if node.Path == root || len(node.ContainerChain) > 0 {
			continue
		}
		if !strings.HasPrefix(node.Path, prefix) {
			return "", 0, fmt.Errorf("canonical tree node %q is outside root %q", node.Path, root)
		}
		relative := strings.TrimPrefix(node.Path, prefix)
		if relative == "" || strings.Contains(relative, "!/") {
			return "", 0, fmt.Errorf("canonical tree node %q has an invalid physical path", node.Path)
		}
		kind := node.Kind
		switch kind {
		case NodeArchive, NodeCompressedStream:
			kind = NodeRegularFile
		case NodeRegularFile, NodeDirectory, NodeLink, NodeSpecial:
		default:
			return "", 0, fmt.Errorf("canonical tree node %q has unsupported kind %q", node.Path, node.Kind)
		}
		records = append(records, treeIdentityRecord{
			Path: relative, Kind: string(kind), Executable: node.Mode&0o111 != 0,
			Size: node.Size, SHA256: node.SHA256,
		})
	}
	sort.Slice(records, func(left, right int) bool { return records[left].Path < records[right].Path })
	return canonicalTreeRecordIdentity(records)
}

func canonicalTreeRecordIdentity(records []treeIdentityRecord) (string, int64, error) {
	var total int64
	for _, record := range records {
		if record.Kind != string(NodeRegularFile) {
			continue
		}
		updated, ok := checkedAdd(total, record.Size)
		if !ok {
			return "", 0, fmt.Errorf("tree size overflow")
		}
		total = updated
	}
	payload, err := marshalCanonicalStruct(records)
	if err != nil {
		return "", 0, err
	}
	digest := sha256.New()
	_, _ = io.WriteString(digest, "curator-artifact-tree-v1\x00")
	_, _ = digest.Write(payload)
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), total, nil
}

func validateTreeOrigin(origin OriginEvidence, identity string) string {
	if !origin.Verified || origin.Locator == "" || origin.ImmutableID == "" || origin.LockRecord == "" {
		return "immutable_origin_binding_missing"
	}
	if !sha256Identity.MatchString(origin.ChecksumSHA256) {
		return "origin_checksum_invalid"
	}
	if origin.ChecksumSHA256 != identity {
		return "origin_checksum_mismatch"
	}
	return ""
}

func filepathToSlash(value string) string {
	return strings.ReplaceAll(value, string(os.PathSeparator), "/")
}
