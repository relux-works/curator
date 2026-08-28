package yarnclassicsource

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/nodesource"
	"github.com/relux-works/curator/internal/privatedir"
)

// RawTarball names exact caller-captured registry bytes for one lock instance.
type RawTarball struct{ Path string }

// CaptureRequest supplies only raw source inputs and manager-owned stores.
// WorkRoot is scratch authority, never closure authority.
type CaptureRequest struct {
	Graph              Graph
	ProjectRoot        string
	Tarballs           map[string]RawTarball
	WorkRoot           string
	Store              *closureexec.CaptureStore
	Policy             *artifactpolicy.Service
	PreviousCausalHead string
}

// ArtifactEvidence is detached audit evidence for one exact raw input.
type ArtifactEvidence struct {
	PackagePath, Origin, Integrity, SHA256 string
	Size                                   int64
	ArtifactManifestID, IntakeReceiptID    closuregraph.ID
}

// Evidence is the complete Yarn Classic capture/admission result.
type Evidence struct {
	ProfileID, LockDigest string
	Project               ArtifactEvidence
	Tarballs              []ArtifactEvidence
	DiscardedDerivedPaths []string
	NodeCaptureGraphID    closuregraph.ID
}

// Capture owns admitted handles required for cache derivation and replay.
// Detached Evidence is safe to persist; raw protected paths are intentionally private.
type Capture struct {
	Graph       Graph
	NodeCapture nodesource.Capture
	Evidence    Evidence
	project     capturedInput
	tarballs    map[string]capturedInput
	policy      *artifactpolicy.Service
	store       *closureexec.CaptureStore
}

type capturedInput struct {
	handle     *closureexec.CaptureHandle
	tree       *closureexec.SourceTreeHandle
	input      closureexec.AdmittedInput
	receiptID  closuregraph.ID
	manifestID closuregraph.ID
	digest     closuregraph.ID
	size       int64
	files      []packageFile
}

// packageFile is the manager-neutral extraction evidence derived directly
// from one admitted raw tgz. Yarn's cache and installed tree are never allowed
// to become the authority for these bytes.
type packageFile struct {
	Path       string
	SHA256     closuregraph.ID
	Size       int64
	Executable bool
}

// CaptureAndAdmit snapshots the project without derived manager state, verifies
// every exact tarball against lock SRI, recursively admits all bytes, reconciles
// embedded package metadata, and emits the common Node capture graph.
func CaptureAndAdmit(ctx context.Context, request CaptureRequest) (*Capture, error) {
	if request.Store == nil || request.Policy == nil || request.PreviousCausalHead == "" {
		return nil, fail(CodeInputUndeclared, "Yarn capture authority is incomplete", nil)
	}
	root, err := filepath.Abs(request.ProjectRoot)
	if err != nil || root != filepath.Clean(request.ProjectRoot) {
		return nil, fail(CodeLocalPathEscape, "project root must be an absolute clean path", nil)
	}
	workRoot, err := filepath.Abs(request.WorkRoot)
	if err != nil || request.WorkRoot == "" {
		return nil, fail(CodeInputUndeclared, "Yarn work root is invalid", nil)
	}
	if err = privatedir.MakeAll(workRoot); err != nil {
		return nil, err
	}
	stage, err := os.MkdirTemp(workRoot, "yarn-classic-project-stage-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	discarded, err := copyProjectSource(root, stage, request.Graph.Layout.ModulesFolder)
	if err != nil {
		return nil, err
	}
	lockPath := filepath.Join(stage, request.Graph.LockName)
	lockBytes, err := os.ReadFile(lockPath) // #nosec G304 -- exact contained staged lock path.
	if err != nil || digestID(lockBytes) != closuregraph.ID(request.Graph.RawLockSHA256) {
		return nil, fail(CodeLockStale, "captured project lock differs from parsed lock", map[string]string{"lock": request.Graph.LockName})
	}
	if err = reconcileCapturedAuthorities(stage, request.Graph); err != nil {
		return nil, err
	}
	for manifestPath, expected := range request.Graph.manifestBytes {
		actual, readErr := os.ReadFile(filepath.Join(stage, filepath.FromSlash(manifestPath))) // #nosec G304 -- validated manifest path.
		if readErr != nil || !bytes.Equal(actual, expected) {
			return nil, fail(CodeLockStale, "captured project manifest differs from parsed manifest", map[string]string{"path": manifestPath})
		}
	}
	for configPath, expected := range request.Graph.configurationBytes {
		actual, readErr := os.ReadFile(filepath.Join(stage, filepath.FromSlash(configPath))) // #nosec G304 -- validated root config path.
		if readErr != nil || !bytes.Equal(actual, expected) {
			return nil, fail(CodeLockStale, "captured Yarn configuration differs from parsed configuration", map[string]string{"path": configPath})
		}
	}
	project, err := admitProjectTree(ctx, request, stage)
	if err != nil {
		return nil, err
	}

	external := []Package{}
	for _, pkg := range request.Graph.Packages {
		if pkg.Resolved != "" {
			external = append(external, pkg)
		}
	}
	if len(request.Tarballs) != len(external) {
		return nil, fail(CodeOfflineInputMissing, "raw Yarn tarball set is not bijective with the lock graph", map[string]string{"expected": fmt.Sprint(len(external)), "observed": fmt.Sprint(len(request.Tarballs))})
	}
	captured := make(map[string]capturedInput, len(external))
	evidence := make([]ArtifactEvidence, 0, len(external))
	for _, pkg := range external {
		raw, ok := request.Tarballs[pkg.Key]
		if !ok {
			return nil, fail(CodeOfflineInputMissing, "required raw Yarn tarball is absent", map[string]string{"package": pkg.Key})
		}
		item, itemEvidence, metadata, inspection, admitErr := admitTarball(ctx, request, pkg, raw)
		if admitErr != nil {
			return nil, admitErr
		}
		if err = reconcileEmbeddedMetadata(pkg, metadata, inspection); err != nil {
			return nil, err
		}
		index := request.Graph.packageIndex[pkg.Key]
		request.Graph.Packages[index].manifest = metadata
		request.Graph.Packages[index].PeerDependencies = cloneMap(metadata.PeerDependencies)
		request.Graph.Packages[index].PeerOptional = peerOptional(metadata.PeerDependenciesMeta)
		request.Graph.Packages[index].OS = sortedStrings(metadata.OS)
		request.Graph.Packages[index].CPU = sortedStrings(metadata.CPU)
		request.Graph.Packages[index].Libc = sortedStrings(metadata.Libc)
		captured[pkg.Key] = item
		evidence = append(evidence, itemEvidence)
	}
	for key := range request.Tarballs {
		if _, ok := captured[key]; !ok {
			return nil, fail(CodeGraphIncomplete, "raw Yarn tarball has no lock instance", map[string]string{"package": key})
		}
	}
	request.Graph.Edges, err = buildEdges(request.Graph.Packages, request.Graph.packageIndex, request.Graph.selectorIndex, request.Graph.workspaceByName, request.Graph.Target)
	if err != nil {
		return nil, err
	}
	if err = markSelection(request.Graph.Packages, request.Graph.Edges, request.Graph.packageIndex, request.Graph.Target); err != nil {
		return nil, err
	}
	nodeCapture, err := buildNodeCapture(request.Graph, project, captured)
	if err != nil {
		return nil, err
	}
	nodeID, err := nodeCapture.Graph.ID()
	if err != nil {
		return nil, err
	}
	projectEvidence := ArtifactEvidence{PackagePath: ".", Origin: "workspace:.", SHA256: string(project.digest), Size: project.size, ArtifactManifestID: project.manifestID, IntakeReceiptID: project.receiptID}
	return &Capture{Graph: request.Graph, NodeCapture: nodeCapture, Evidence: Evidence{ProfileID: ProfileID, LockDigest: request.Graph.LockDigest, Project: projectEvidence, Tarballs: evidence, DiscardedDerivedPaths: discarded, NodeCaptureGraphID: nodeID}, project: project, tarballs: captured, policy: request.Policy, store: request.Store}, nil
}

// reconcileCapturedAuthorities discovers every Yarn-visible workspace and
// project configuration file from the immutable staged tree. Caller-selected
// manifests/configuration cannot narrow the authority Yarn will observe.
func reconcileCapturedAuthorities(root string, graph Graph) error {
	rootManifest, ok := graph.manifestBytes["package.json"]
	if !ok {
		return fail(CodeLockStale, "captured root manifest authority is absent", nil)
	}
	var manifest packageManifest
	if err := json.Unmarshal(rootManifest, &manifest); err != nil {
		return fail(CodeLockStale, "captured root manifest authority is invalid", nil)
	}
	patterns, _, err := workspacePatterns(manifest.Workspaces)
	if err != nil {
		return err
	}
	discoveredManifests := map[string]bool{"package.json": true}
	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/*") {
			parent := strings.TrimSuffix(pattern, "/*")
			entries, readErr := os.ReadDir(filepath.Join(root, filepath.FromSlash(parent)))
			if readErr != nil {
				return fail(CodeLockStale, "workspace pattern has no captured directory", map[string]string{"workspace": pattern})
			}
			for _, entry := range entries {
				if !entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
					continue
				}
				logical := path.Join(parent, entry.Name(), "package.json")
				if info, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(logical))); statErr == nil && info.Mode().IsRegular() && info.Mode()&fs.ModeSymlink == 0 {
					discoveredManifests[logical] = true
				}
			}
			continue
		}
		logical := path.Join(pattern, "package.json")
		info, statErr := os.Lstat(filepath.Join(root, filepath.FromSlash(logical)))
		if statErr != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
			return fail(CodeLockStale, "declared workspace manifest is absent", map[string]string{"path": logical})
		}
		discoveredManifests[logical] = true
	}
	if err := requireAuthorityBijection(discoveredManifests, graph.manifestBytes, "workspace manifest"); err != nil {
		return err
	}

	discoveredConfig := map[string]bool{}
	err = filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() != ".yarnrc" && entry.Name() != ".npmrc" && entry.Name() != ".yarnrc.yml" {
			return nil
		}
		rel, relErr := filepath.Rel(root, current)
		if relErr != nil {
			return relErr
		}
		discoveredConfig[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		return err
	}
	return requireAuthorityBijection(discoveredConfig, graph.configurationBytes, "manager configuration")
}

func requireAuthorityBijection(discovered map[string]bool, expected map[string][]byte, authority string) error {
	for logical := range discovered {
		if _, ok := expected[logical]; !ok {
			return fail(CodeLockStale, "captured "+authority+" is absent from parsed authority", map[string]string{"path": logical})
		}
	}
	for logical := range expected {
		if !discovered[logical] {
			return fail(CodeLockStale, "parsed "+authority+" is absent from captured tree", map[string]string{"path": logical})
		}
	}
	return nil
}

func admitProjectTree(ctx context.Context, request CaptureRequest, stage string) (capturedInput, error) {
	probeDescriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "yarn-classic", PackageName: request.Graph.RootName, PackageVersion: request.Graph.RootVersion}
	probe, probeErr := request.Policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: probeDescriptor, Root: stage, VirtualRoot: "workspace"})
	if probeErr != nil && artifactpolicy.ErrorCode(probeErr) != artifactpolicy.CodeOriginUnverified {
		return capturedInput{}, probeErr
	}
	digest := probe.Manifest.RawPayload.SHA256
	if !closuregraph.ID(digest).Valid() {
		return capturedInput{}, fail(CodeIntegrityMismatch, "project snapshot identity is unavailable", nil)
	}
	tree, err := request.Store.CaptureTree("workspace:.", stage)
	if err != nil {
		return capturedInput{}, err
	}
	protected, err := tree.ProtectedPath()
	if err != nil {
		return capturedInput{}, err
	}
	descriptor := probeDescriptor
	descriptor.Origin = artifactpolicy.OriginEvidence{Locator: "workspace:.", ImmutableID: digest, LockRecord: request.Graph.LockDigest, ChecksumSHA256: digest, Verified: true}
	admitted, err := request.Policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: protected, VirtualRoot: "workspace"})
	if err != nil {
		return capturedInput{}, err
	}
	manifestID := closuregraph.ID(admitted.Manifest.ManifestDigest)
	receipt, err := request.Store.AdmitTree(tree, "workspace:.", closureexec.AdmissionEvidence{PreviousCausalHead: request.PreviousCausalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return capturedInput{}, err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return capturedInput{}, err
	}
	return capturedInput{tree: tree, input: closureexec.AdmittedInput{Receipt: receipt, Tree: tree}, receiptID: receiptID, manifestID: manifestID, digest: closuregraph.ID(digest), size: probe.Manifest.RawPayload.Size}, nil
}

type tarInspection struct{ bindingGYP, bundled bool }

func admitTarball(ctx context.Context, request CaptureRequest, pkg Package, raw RawTarball) (capturedInput, ArtifactEvidence, packageManifest, tarInspection, error) {
	if !filepath.IsAbs(raw.Path) {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeOfflineInputMissing, "raw Yarn tarball path must be absolute", map[string]string{"package": pkg.Key})
	}
	info, err := os.Lstat(raw.Path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeOfflineInputMissing, "raw Yarn tarball is absent or not a regular file", map[string]string{"package": pkg.Key})
	}
	payload, err := os.ReadFile(raw.Path) // #nosec G304 -- caller supplied explicit raw input.
	if err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	if err = verifySRI(pkg.Integrity, payload); err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	digest := digestID(payload)
	handle, err := request.Store.Capture(pkg.Resolved, int64(len(payload)), bytes.NewReader(payload))
	if err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	reader, err := handle.Open()
	if err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	defer func() { _ = reader.Close() }()
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "yarn-classic", PackageName: pkg.Name, PackageVersion: pkg.Version, Origin: artifactpolicy.OriginEvidence{Locator: pkg.Resolved, ImmutableID: pkg.Integrity, LockRecord: request.Graph.LockDigest, ChecksumSHA256: string(digest), Verified: true}}
	admitted, err := request.Policy.AdmitDependency(ctx, artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: "yarn-classic/" + safeArchiveName(pkg) + ".tgz", Size: int64(len(payload)), Reader: reader}})
	if err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	metadata, inspection, files, err := inspectTarball(payload)
	if err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	manifestID := closuregraph.ID(admitted.Manifest.ManifestDigest)
	receipt, err := request.Store.Admit(handle, pkg.Resolved, closureexec.AdmissionEvidence{PreviousCausalHead: request.PreviousCausalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	item := capturedInput{handle: handle, input: closureexec.AdmittedInput{Receipt: receipt, Handle: handle}, receiptID: receiptID, manifestID: manifestID, digest: digest, size: int64(len(payload)), files: files}
	evidence := ArtifactEvidence{PackagePath: pkg.Key, Origin: pkg.Resolved, Integrity: pkg.Integrity, SHA256: string(digest), Size: int64(len(payload)), ArtifactManifestID: manifestID, IntakeReceiptID: receiptID}
	return item, evidence, metadata, inspection, nil
}

func inspectTarball(payload []byte) (packageManifest, tarInspection, []packageFile, error) {
	gz, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return packageManifest{}, tarInspection{}, nil, fail(CodeMetadataMismatch, "Yarn tarball gzip envelope is invalid", nil)
	}
	defer func() { _ = gz.Close() }()
	reader := tar.NewReader(gz)
	var manifest packageManifest
	found := false
	inspection := tarInspection{}
	files := []packageFile{}
	seen := map[string]bool{}
	for {
		header, readErr := reader.Next()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "Yarn tarball is invalid: "+readErr.Error(), nil)
		}
		name := path.Clean(header.Name)
		if path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "Yarn tarball member escapes package root", map[string]string{"path": header.Name})
		}
		if name != "package" && !strings.HasPrefix(name, "package/") {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "Yarn tarball member is outside the package root", map[string]string{"path": header.Name})
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if !header.FileInfo().Mode().IsRegular() {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "Yarn tarball contains a non-regular package member", map[string]string{"path": header.Name})
		}
		rel := strings.TrimPrefix(name, "package/")
		if rel == "" || seen[rel] {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "Yarn tarball contains a duplicate or empty package member", map[string]string{"path": header.Name})
		}
		seen[rel] = true
		if header.Size < 0 {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "Yarn tarball member has an invalid size", map[string]string{"path": header.Name})
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, header.Size+1))
		if readErr != nil || int64(len(data)) != header.Size {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cannot read Yarn tarball member", map[string]string{"path": header.Name})
		}
		files = append(files, packageFile{Path: rel, SHA256: digestID(data), Size: header.Size, Executable: header.Mode&0o111 != 0})
		if name == "package/binding.gyp" {
			inspection.bindingGYP = true
		}
		if strings.HasPrefix(name, "package/node_modules/") {
			inspection.bundled = true
		}
		if name != "package/package.json" {
			continue
		}
		if found {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "Yarn tarball contains duplicate package.json", nil)
		}
		found = true
		if header.Size < 0 || header.Size > 1<<20 {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "embedded package.json exceeds closed limit", nil)
		}
		if err := validateJSON(data); err != nil {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "embedded package.json is not closed JSON: "+err.Error(), nil)
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cannot decode embedded package.json", nil)
		}
	}
	if !found {
		return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "Yarn tarball lacks package/package.json", nil)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return manifest, inspection, files, nil
}

func reconcileEmbeddedMetadata(pkg Package, manifest packageManifest, inspection tarInspection) error {
	if manifest.Name != pkg.Name || manifest.Version != pkg.Version {
		return fail(CodeMetadataMismatch, "embedded package identity differs from yarn.lock", map[string]string{"package": pkg.Key, "lock_name": pkg.Name, "artifact_name": manifest.Name, "lock_version": pkg.Version, "artifact_version": manifest.Version})
	}
	for _, item := range []struct {
		name           string
		lock, embedded map[string]string
	}{{"dependencies", withoutKeys(pkg.Dependencies, pkg.OptionalDependencies), withoutKeys(manifest.Dependencies, manifest.OptionalDependencies)}, {"optionalDependencies", pkg.OptionalDependencies, manifest.OptionalDependencies}} {
		if !equalStringMap(item.lock, item.embedded) {
			return fail(CodeMetadataMismatch, "embedded dependency metadata differs from yarn.lock", map[string]string{"package": pkg.Key, "field": item.name})
		}
	}
	lifecycle := lifecycleScripts(manifest.Scripts)
	if inspection.bundled || rawBundlePresent(manifest.BundleDependencies) || rawBundlePresent(manifest.BundledDependencies) {
		return fail(CodeBundledDependencyUnsupported, "Yarn tarball embeds bundled dependencies", map[string]string{"package": pkg.Key})
	}
	gypEnabled := manifest.Gypfile == nil || *manifest.Gypfile
	if inspection.bindingGYP && gypEnabled && manifest.Scripts["install"] == "" && manifest.Scripts["preinstall"] == "" {
		return fail(CodeNativeBuildUnsupported, "binding.gyp would trigger implicit node-gyp rebuild", map[string]string{"package": pkg.Key})
	}
	if len(lifecycle) > 0 {
		return fail(CodeHookUndeclared, "dependency lifecycle execution is outside yarn-classic-source-v1", map[string]string{"package": pkg.Key, "script": lifecycle[0]})
	}
	return nil
}

func buildNodeCapture(graph Graph, project capturedInput, tarballs map[string]capturedInput) (nodesource.Capture, error) {
	instances := []nodesource.PackageInstance{}
	for _, pkg := range graph.Packages {
		item := project
		origin := "workspace:" + pkg.WorkspacePath
		checksum := string(project.digest)
		workspace := pkg.WorkspacePath
		if pkg.Key == "workspace:." {
			workspace = ""
		}
		if pkg.Resolved != "" {
			item = tarballs[pkg.Key]
			origin = pkg.Resolved
			checksum = pkg.Integrity
			workspace = ""
		}
		dependencies := []nodesource.Dependency{}
		peerContext := []string{}
		for _, edge := range graph.Edges {
			if edge.From != pkg.Key || edge.To == "" {
				continue
			}
			target := graph.Packages[graph.packageIndex[edge.To]]
			scope := closuregraph.ScopeRuntime
			switch edge.Scope {
			case "development":
				scope = closuregraph.ScopeBuild
			case "optional":
				scope = closuregraph.ScopeOptional
			case "peer":
				scope = closuregraph.ScopePeer
				peerContext = append(peerContext, edge.Name+"="+edge.To)
			}
			dependencies = append(dependencies, nodesource.Dependency{PackageKey: edge.To, Scope: scope, Condition: targetCondition(target), DeclarationField: edge.Scope})
		}
		sort.Strings(peerContext)
		instances = append(instances, nodesource.PackageInstance{Key: pkg.Key, Name: pkg.Name, Version: pkg.Version, Origin: origin, Checksum: checksum, PeerKey: strings.Join(peerContext, ","), WorkspacePath: workspace, ArtifactManifestID: item.manifestID, SnapshotDigest: item.digest, Dependencies: dependencies})
	}
	capture, err := nodesource.BuildCapture(nodesource.CaptureInput{Manager: nodesource.ManagerYarnClassic, RootKeys: []string{"workspace:."}, Packages: instances, PolicyIDs: []string{artifactpolicy.PolicyID, ProfileID}})
	if err != nil {
		return nodesource.Capture{}, err
	}
	if len(capture.Graph.RootNodeIDs) != 1 {
		return nodesource.Capture{}, fail(CodeGraphIncomplete, "Yarn capture requires one command product", nil)
	}
	owner := capture.Graph.RootNodeIDs[0]
	return nodesource.AddRuntimeActions(capture, []nodesource.RuntimeAction{
		{Name: "yarn-install", Subtype: "yarn-install", OwnerNodeID: owner, ToolRole: "package-manager", ArgvTemplate: []string{"{{manager_entrypoint}}", "install", "--frozen-lockfile", "--offline", "--ignore-scripts", "--non-interactive", "--no-default-rc", "--cache-folder", "{{cache}}", "--use-yarnrc", "{{yarnrc}}", "{{production}}", "{{modules_folder}}"}, WorkingDirectory: "{{project}}", EnvironmentPolicyID: "yarn-classic-private-environment-v1", ProcessPolicyID: "yarn-classic-install-process-v1"},
		{Name: "node-invoke", Subtype: "node-invoke", OwnerNodeID: owner, ToolRole: "node-runtime", ArgvTemplate: []string{"{{entrypoint}}", "{{args}}"}, WorkingDirectory: "{{project}}", EnvironmentPolicyID: "node-private-environment-v1", ProcessPolicyID: "node-runtime-process-v1"},
	})
}

func copyProjectSource(source, destination, modulesFolder string) ([]string, error) {
	discarded := map[string]bool{}
	err := filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == source {
			return nil
		}
		rel, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}
		logical := filepath.ToSlash(rel)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeInputUndeclared, "project source contains a link", map[string]string{"path": logical})
		}
		base := entry.Name()
		if base == "yarn.lock" && logical != "yarn.lock" {
			return fail(CodeLockStale, "dependency-subtree yarn.lock cannot become resolution authority", map[string]string{"path": logical})
		}
		if entry.IsDir() && (base == "node_modules" || logical == modulesFolder || base == ".cache" || base == ".yarn") {
			discarded[logical] = true
			return filepath.SkipDir
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return privatedir.Make(target)
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return fail(CodeInputUndeclared, "project source contains a special node", map[string]string{"path": logical})
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a contained project-source member.
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, 0o600)
	})
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(discarded))
	for value := range discarded {
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func verifySRI(integrity string, payload []byte) error {
	encoded := strings.TrimPrefix(integrity, "sha512-")
	expected, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(expected) != sha512.Size {
		return fail(CodeIntegrityMissing, "Yarn integrity is not canonical sha512", nil)
	}
	actual := sha512.Sum512(payload)
	if subtle.ConstantTimeCompare(expected, actual[:]) != 1 {
		return fail(CodeIntegrityMismatch, "raw Yarn tarball differs from lock integrity", map[string]string{"integrity": integrity})
	}
	return nil
}
func digestID(payload []byte) closuregraph.ID {
	sum := sha256.Sum256(payload)
	return closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
}
func safeArchiveName(pkg Package) string {
	value := strings.NewReplacer("@", "", "/", "-", "\\", "-", "..", "-").Replace(pkg.Name + "-" + pkg.Version)
	return value
}
func lifecycleScripts(scripts map[string]string) []string {
	names := []string{}
	for name, value := range scripts {
		if value == "" {
			continue
		}
		switch name {
		case "preinstall", "install", "postinstall", "prepare":
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
func withoutKeys(source, remove map[string]string) map[string]string {
	result := cloneMap(source)
	for key := range remove {
		delete(result, key)
	}
	return result
}
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
