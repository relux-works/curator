// Package pnpmsource implements the pinned, fail-closed pnpm source profile.
package pnpmsource

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

// RawTarball names exact caller-captured registry bytes by packages-table key.
type RawTarball struct{ Path string }

// CaptureRequest supplies manager-neutral raw inputs and protected stores.
type CaptureRequest struct {
	Graph              Graph
	ProjectRoot        string
	Tarballs           map[string]RawTarball
	WorkRoot           string
	Store              *closureexec.CaptureStore
	Policy             *artifactpolicy.Service
	PreviousCausalHead string
}

// ArtifactEvidence is detached immutable evidence for one admitted input.
type ArtifactEvidence struct {
	Key, Origin, Integrity, SHA256 string
	Size                           int64
	ArtifactManifestID             closuregraph.ID
	IntakeReceiptID                closuregraph.ID
}

// Evidence is the persistable pnpm capture/admission record.
type Evidence struct {
	ProfileID, LockDigest string
	LocalRoots, Tarballs  []ArtifactEvidence
	Patches               []ArtifactEvidence
	PatchTransforms       []PatchTransformEvidence
	DiscardedAmbientPaths []string
	NodeCaptureGraphID    closuregraph.ID
}

// PatchTransformEvidence binds admitted patch inputs to the exact expected
// package inventory for the snapshot contexts that consume them.
type PatchTransformEvidence struct {
	Selector                         string
	SnapshotKeys                     []string
	ManagerHash                      string
	TarballReceiptID, PatchReceiptID closuregraph.ID
	ExpectedFiles                    []packageFile
	ReceiptID                        closuregraph.ID
}

// Capture retains protected handles used for private-store derivation.
type Capture struct {
	Graph       Graph
	NodeCapture nodesource.Capture
	Evidence    Evidence
	project     capturedTree
	localRoots  map[string]capturedTree
	tarballs    map[string]capturedBlob
	patches     map[string]capturedBlob
	policy      *artifactpolicy.Service
	store       *closureexec.CaptureStore
}

type capturedTree struct {
	tree       *closureexec.SourceTreeHandle
	input      closureexec.AdmittedInput
	receiptID  closuregraph.ID
	manifestID closuregraph.ID
	digest     closuregraph.ID
	size       int64
}
type capturedBlob struct {
	handle     *closureexec.CaptureHandle
	input      closureexec.AdmittedInput
	receiptID  closuregraph.ID
	manifestID closuregraph.ID
	digest     closuregraph.ID
	size       int64
	files      []packageFile
	contents   map[string][]byte
}
type packageFile struct {
	Path       string
	SHA256     closuregraph.ID
	Size       int64
	Executable bool
}
type tarInspection struct{ bindingGYP, bundled bool }

// CaptureAndAdmit snapshots the project without installed/store state, admits
// each local root independently, verifies all raw tarballs and patches, and
// emits the common selection-neutral Node graph.
func CaptureAndAdmit(ctx context.Context, request CaptureRequest) (*Capture, error) {
	if request.Store == nil || request.Policy == nil || request.PreviousCausalHead == "" {
		return nil, fail(CodeInputUndeclared, "pnpm capture authority is incomplete", nil)
	}
	root, err := filepath.Abs(request.ProjectRoot)
	if err != nil || request.ProjectRoot == "" || root != filepath.Clean(request.ProjectRoot) {
		return nil, fail(CodeLocalPathEscape, "project root must be an absolute clean path", nil)
	}
	workRoot, err := filepath.Abs(request.WorkRoot)
	if err != nil || request.WorkRoot == "" || workRoot != filepath.Clean(request.WorkRoot) {
		return nil, fail(CodeInputUndeclared, "pnpm work root must be absolute and clean", nil)
	}
	if err = privatedir.MakeAll(workRoot); err != nil {
		return nil, err
	}
	stage, err := os.MkdirTemp(workRoot, "pnpm-project-stage-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	discarded, err := copyProjectSource(root, stage)
	if err != nil {
		return nil, err
	}
	if err = reconcileCapturedDeclarations(stage, request.Graph); err != nil {
		return nil, err
	}
	project, err := admitTree(ctx, request, stage, ".", "workspace:.")
	if err != nil {
		return nil, err
	}
	locals := map[string]capturedTree{".": project}
	localEvidence := []ArtifactEvidence{treeEvidence(".", "workspace:.", project)}
	for _, local := range request.Graph.LocalRoots {
		if local.Path == "." {
			continue
		}
		localPath := filepath.Join(stage, filepath.FromSlash(local.Path))
		item, err := admitTree(ctx, request, localPath, local.Path, "workspace:"+local.Path)
		if err != nil {
			return nil, err
		}
		locals[local.Path] = item
		localEvidence = append(localEvidence, treeEvidence(local.Path, "workspace:"+local.Path, item))
	}
	if len(request.Tarballs) != len(request.Graph.Packages) {
		return nil, fail(CodeOfflineInputMissing, "raw pnpm tarball set is not bijective with packages table", map[string]string{"expected": fmt.Sprint(len(request.Graph.Packages)), "observed": fmt.Sprint(len(request.Tarballs))})
	}
	tarballs := map[string]capturedBlob{}
	tarballEvidence := []ArtifactEvidence{}
	for _, pkg := range request.Graph.Packages {
		raw, ok := request.Tarballs[pkg.Key]
		if !ok {
			return nil, fail(CodeOfflineInputMissing, "required raw pnpm tarball is absent", map[string]string{"package": pkg.Key})
		}
		item, evidence, manifest, inspection, err := admitTarball(ctx, request, pkg, raw)
		if err != nil {
			return nil, err
		}
		if err = reconcileEmbeddedMetadata(request.Graph, pkg, manifest, inspection); err != nil {
			return nil, err
		}
		tarballs[pkg.Key] = item
		tarballEvidence = append(tarballEvidence, evidence)
	}
	for key := range request.Tarballs {
		if _, ok := tarballs[key]; !ok {
			return nil, fail(CodeGraphIncomplete, "raw pnpm tarball has no packages-table member", map[string]string{"package": key})
		}
	}
	patches, patchEvidence, err := admitPatches(ctx, request)
	if err != nil {
		return nil, err
	}
	patchTransforms, err := derivePatchedInventories(request.Graph, tarballs, patches)
	if err != nil {
		return nil, err
	}
	nodeCapture, err := buildNodeCapture(request.Graph, locals, tarballs)
	if err != nil {
		return nil, err
	}
	nodeID, err := nodeCapture.Graph.ID()
	if err != nil {
		return nil, err
	}
	sort.Slice(localEvidence, func(i, j int) bool { return localEvidence[i].Key < localEvidence[j].Key })
	sort.Slice(tarballEvidence, func(i, j int) bool { return tarballEvidence[i].Key < tarballEvidence[j].Key })
	sort.Slice(patchEvidence, func(i, j int) bool { return patchEvidence[i].Key < patchEvidence[j].Key })
	return &Capture{Graph: request.Graph, NodeCapture: nodeCapture, Evidence: Evidence{ProfileID: ProfileID, LockDigest: request.Graph.LockDigest, LocalRoots: localEvidence, Tarballs: tarballEvidence, Patches: patchEvidence, PatchTransforms: patchTransforms, DiscardedAmbientPaths: discarded, NodeCaptureGraphID: nodeID}, project: project, localRoots: locals, tarballs: tarballs, patches: patches, policy: request.Policy, store: request.Store}, nil
}

func reconcileCapturedDeclarations(stage string, graph Graph) error {
	lock, err := os.ReadFile(filepath.Join(stage, "pnpm-lock.yaml")) // #nosec G304 -- exact lock basename below the owned stage.
	if err != nil || digestID(lock) != closuregraph.ID(graph.RawLockSHA256) {
		return fail(CodeLockStale, "captured pnpm-lock.yaml differs from parsed lock", nil)
	}
	for name, expected := range graph.manifestBytes {
		actual, err := os.ReadFile(filepath.Join(stage, filepath.FromSlash(name))) // #nosec G304 -- parser validated the contained manifest path.
		if err != nil || !bytes.Equal(actual, expected) {
			return fail(CodeLockStale, "captured package manifest differs from parsed manifest", map[string]string{"path": name})
		}
	}
	for name, expected := range graph.configBytes {
		actual, err := os.ReadFile(filepath.Join(stage, filepath.FromSlash(name))) // #nosec G304 -- parser restricts configuration to closed contained basenames.
		if err != nil || !bytes.Equal(actual, expected) {
			return fail(CodeLockStale, "captured pnpm configuration differs from parsed configuration", map[string]string{"path": name})
		}
	}
	for name, expected := range graph.patchBytes {
		actual, err := os.ReadFile(filepath.Join(stage, filepath.FromSlash(name))) // #nosec G304 -- parser validated the declared contained patch path.
		if err != nil || !bytes.Equal(actual, expected) {
			return fail(CodeIntegrityMismatch, "captured pnpm patch differs from parsed patch", map[string]string{"path": name})
		}
	}
	for _, hook := range []string{".pnpmfile.cjs", ".pnpmfile.mjs"} {
		if _, err := os.Lstat(filepath.Join(stage, hook)); err == nil {
			return fail(CodeManagerPluginUndeclared, "pnpm hook file is present in captured project", map[string]string{"path": hook})
		}
	}
	return nil
}

func admitTree(ctx context.Context, request CaptureRequest, root, logical, origin string) (capturedTree, error) {
	virtualRoot := "workspace"
	if logical != "." {
		virtualRoot = "workspace/" + logical
	}
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "pnpm", PackageName: "workspace", PackageVersion: request.Graph.LockDigest}
	probe, err := request.Policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: virtualRoot})
	if err != nil && artifactpolicy.ErrorCode(err) != artifactpolicy.CodeOriginUnverified {
		return capturedTree{}, err
	}
	digest := closuregraph.ID(probe.Manifest.RawPayload.SHA256)
	if !digest.Valid() {
		return capturedTree{}, fail(CodeIntegrityMismatch, "local pnpm root identity is unavailable", map[string]string{"path": logical})
	}
	tree, err := request.Store.CaptureTree(origin, root)
	if err != nil {
		return capturedTree{}, err
	}
	protected, err := tree.ProtectedPath()
	if err != nil {
		return capturedTree{}, err
	}
	descriptor.Origin = artifactpolicy.OriginEvidence{Locator: origin, ImmutableID: string(digest), LockRecord: request.Graph.LockDigest, ChecksumSHA256: string(digest), Verified: true}
	admitted, err := request.Policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: protected, VirtualRoot: virtualRoot})
	if err != nil {
		return capturedTree{}, err
	}
	manifestID := closuregraph.ID(admitted.Manifest.ManifestDigest)
	receipt, err := request.Store.AdmitTree(tree, origin, closureexec.AdmissionEvidence{PreviousCausalHead: request.PreviousCausalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return capturedTree{}, err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return capturedTree{}, err
	}
	return capturedTree{tree: tree, input: closureexec.AdmittedInput{Receipt: receipt, Tree: tree}, receiptID: receiptID, manifestID: manifestID, digest: digest, size: probe.Manifest.RawPayload.Size}, nil
}

func admitTarball(ctx context.Context, request CaptureRequest, pkg Package, raw RawTarball) (capturedBlob, ArtifactEvidence, packageManifest, tarInspection, error) {
	if !filepath.IsAbs(raw.Path) {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeOfflineInputMissing, "raw pnpm tarball path must be absolute", map[string]string{"package": pkg.Key})
	}
	info, err := os.Lstat(raw.Path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeOfflineInputMissing, "raw pnpm tarball is absent or not regular", map[string]string{"package": pkg.Key})
	}
	payload, err := os.ReadFile(raw.Path)
	if err != nil {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	if err = verifySRI(pkg.Integrity, payload); err != nil {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	digest := digestID(payload)
	handle, err := request.Store.Capture(pkg.Resolved, int64(len(payload)), bytes.NewReader(payload))
	if err != nil {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	reader, err := handle.Open()
	if err != nil {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	defer func() { _ = reader.Close() }()
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "pnpm", PackageName: pkg.Name, PackageVersion: pkg.Version, Origin: artifactpolicy.OriginEvidence{Locator: pkg.Resolved, ImmutableID: pkg.Integrity, LockRecord: request.Graph.LockDigest, ChecksumSHA256: string(digest), Verified: true}}
	admitted, err := request.Policy.AdmitDependency(ctx, artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: "pnpm/" + safeArchiveName(pkg) + ".tgz", Size: int64(len(payload)), Reader: reader}})
	if err != nil {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	manifest, inspection, files, contents, err := inspectTarball(payload)
	if err != nil {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	manifestID := closuregraph.ID(admitted.Manifest.ManifestDigest)
	receipt, err := request.Store.Admit(handle, pkg.Resolved, closureexec.AdmissionEvidence{PreviousCausalHead: request.PreviousCausalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return capturedBlob{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	item := capturedBlob{handle: handle, input: closureexec.AdmittedInput{Receipt: receipt, Handle: handle}, receiptID: receiptID, manifestID: manifestID, digest: digest, size: int64(len(payload)), files: files, contents: contents}
	return item, ArtifactEvidence{Key: pkg.Key, Origin: pkg.Resolved, Integrity: pkg.Integrity, SHA256: string(digest), Size: int64(len(payload)), ArtifactManifestID: manifestID, IntakeReceiptID: receiptID}, manifest, inspection, nil
}

func admitPatches(ctx context.Context, request CaptureRequest) (map[string]capturedBlob, []ArtifactEvidence, error) {
	result := map[string]capturedBlob{}
	evidence := []ArtifactEvidence{}
	for _, patch := range request.Graph.Patches {
		payload := request.Graph.patchBytes[patch.Path]
		digest := digestID(payload)
		handle, err := request.Store.Capture("patch:"+patch.Path, int64(len(payload)), bytes.NewReader(payload))
		if err != nil {
			return nil, nil, err
		}
		reader, err := handle.Open()
		if err != nil {
			return nil, nil, err
		}
		descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "pnpm", PackageName: patch.Selector, PackageVersion: "patch", Origin: artifactpolicy.OriginEvidence{Locator: "patch:" + patch.Path, ImmutableID: patch.SHA256, LockRecord: request.Graph.LockDigest, ChecksumSHA256: string(digest), Verified: true}}
		admitted, err := request.Policy.AdmitDependency(ctx, artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: patch.Path, Size: int64(len(payload)), Reader: reader}})
		_ = reader.Close()
		if err != nil {
			return nil, nil, err
		}
		manifestID := closuregraph.ID(admitted.Manifest.ManifestDigest)
		receipt, err := request.Store.Admit(handle, "patch:"+patch.Path, closureexec.AdmissionEvidence{PreviousCausalHead: request.PreviousCausalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
		if err != nil {
			return nil, nil, err
		}
		receiptID, err := receipt.ID()
		if err != nil {
			return nil, nil, err
		}
		item := capturedBlob{handle: handle, input: closureexec.AdmittedInput{Receipt: receipt, Handle: handle}, receiptID: receiptID, manifestID: manifestID, digest: digest, size: int64(len(payload))}
		result[patch.Path] = item
		evidence = append(evidence, ArtifactEvidence{Key: patch.Selector, Origin: "patch:" + patch.Path, Integrity: patch.ManagerHash, SHA256: string(digest), Size: int64(len(payload)), ArtifactManifestID: manifestID, IntakeReceiptID: receiptID})
	}
	return result, evidence, nil
}

func inspectTarball(payload []byte) (packageManifest, tarInspection, []packageFile, map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return packageManifest{}, tarInspection{}, nil, nil, fail(CodeMetadataMismatch, "pnpm tarball gzip envelope is invalid", nil)
	}
	defer func() { _ = gz.Close() }()
	reader := tar.NewReader(gz)
	manifest := packageManifest{}
	inspection := tarInspection{}
	files := []packageFile{}
	contents := map[string][]byte{}
	seen := map[string]bool{}
	found := false
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return packageManifest{}, inspection, nil, nil, fail(CodeMetadataMismatch, "pnpm tarball is invalid: "+err.Error(), nil)
		}
		name := path.Clean(header.Name)
		if path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") || (!strings.HasPrefix(name, "package/") && name != "package") {
			return packageManifest{}, inspection, nil, nil, fail(CodeMetadataMismatch, "pnpm tarball member escapes package root", map[string]string{"path": header.Name})
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if !header.FileInfo().Mode().IsRegular() {
			return packageManifest{}, inspection, nil, nil, fail(CodeMetadataMismatch, "pnpm tarball contains non-regular member", map[string]string{"path": header.Name})
		}
		rel := strings.TrimPrefix(name, "package/")
		if rel == "" || seen[rel] {
			return packageManifest{}, inspection, nil, nil, fail(CodeMetadataMismatch, "pnpm tarball contains duplicate or empty member", map[string]string{"path": header.Name})
		}
		seen[rel] = true
		data, err := io.ReadAll(io.LimitReader(reader, header.Size+1))
		if err != nil || int64(len(data)) != header.Size {
			return packageManifest{}, inspection, nil, nil, fail(CodeMetadataMismatch, "cannot read pnpm tarball member", map[string]string{"path": header.Name})
		}
		files = append(files, packageFile{Path: rel, SHA256: digestID(data), Size: header.Size, Executable: header.Mode&0o111 != 0})
		contents[rel] = append([]byte(nil), data...)
		if rel == "binding.gyp" {
			inspection.bindingGYP = true
		}
		if strings.HasPrefix(rel, "node_modules/") {
			inspection.bundled = true
		}
		if rel == "package.json" {
			if err = json.Unmarshal(data, &manifest); err != nil {
				return packageManifest{}, inspection, nil, nil, fail(CodeMetadataMismatch, "embedded package.json is invalid", nil)
			}
			found = true
		}
	}
	if !found {
		return packageManifest{}, inspection, nil, nil, fail(CodeMetadataMismatch, "pnpm tarball has no package.json", nil)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return manifest, inspection, files, contents, nil
}

func reconcileEmbeddedMetadata(graph Graph, pkg Package, manifest packageManifest, inspection tarInspection) error {
	if manifest.Name != pkg.Name || manifest.Version != pkg.Version {
		return fail(CodeMetadataMismatch, "embedded package identity differs from pnpm lock", map[string]string{"package": pkg.Key})
	}
	if !equalMap(manifest.PeerDependencies, pkg.PeerDependencies) || !equalBoolMap(manifest.PeerDependenciesMeta, pkg.PeerOptional) {
		return fail(CodeMetadataMismatch, "embedded peer metadata differs from pnpm lock", map[string]string{"package": pkg.Key})
	}
	if strings.Join(sortedCopy(manifest.OS), "\x00") != strings.Join(pkg.OS, "\x00") || strings.Join(sortedCopy(manifest.CPU), "\x00") != strings.Join(pkg.CPU, "\x00") || strings.Join(sortedCopy(manifest.Libc), "\x00") != strings.Join(pkg.Libc, "\x00") {
		return fail(CodeMetadataMismatch, "embedded target metadata differs from pnpm lock", map[string]string{"package": pkg.Key})
	}
	dependencyNames := map[string]bool{}
	for _, snapshot := range graph.Snapshots {
		if snapshot.PackageKey != pkg.Key {
			continue
		}
		for name := range snapshot.Dependencies {
			dependencyNames[name] = true
		}
		for name := range snapshot.OptionalDependencies {
			dependencyNames[name] = true
		}
	}
	manifestNames := map[string]bool{}
	for name := range manifest.Dependencies {
		manifestNames[name] = true
	}
	for name := range manifest.OptionalDependencies {
		manifestNames[name] = true
	}
	if !equalSet(dependencyNames, manifestNames) {
		return fail(CodeMetadataMismatch, "embedded dependency names differ from pnpm snapshots", map[string]string{"package": pkg.Key})
	}
	if inspection.bundled || rawPresent(manifest.BundleDependencies) || rawPresent(manifest.BundledDependencies) {
		return fail(CodeBundledDependencyUnsupported, "pnpm tarball embeds bundled dependencies", map[string]string{"package": pkg.Key})
	}
	gyp := manifest.Gypfile == nil || *manifest.Gypfile
	if inspection.bindingGYP && gyp && manifest.Scripts["install"] == "" && manifest.Scripts["preinstall"] == "" {
		return fail(CodeNativeBuildUnsupported, "binding.gyp would trigger implicit native build", map[string]string{"package": pkg.Key})
	}
	if scripts := lifecycleScripts(manifest.Scripts); len(scripts) > 0 {
		return fail(CodeHookUndeclared, "dependency lifecycle execution is outside pnpm-source-v1", map[string]string{"package": pkg.Key, "script": scripts[0]})
	}
	return nil
}

func buildNodeCapture(graph Graph, locals map[string]capturedTree, tarballs map[string]capturedBlob) (nodesource.Capture, error) {
	instances := []nodesource.PackageInstance{}
	rootKeys := []string{}
	for _, local := range graph.LocalRoots {
		item := locals[local.Path]
		key := localRootKey(local.Path)
		workspacePath := local.Path
		if workspacePath == "." {
			workspacePath = ""
		}
		deps := []nodesource.Dependency{}
		from := key
		for _, edge := range graph.Edges {
			if edge.From != from || edge.To == "" {
				continue
			}
			deps = append(deps, nodeDependency(edge, graph))
		}
		instances = append(instances, nodesource.PackageInstance{Key: key, Name: local.Name, Version: local.Version, Origin: "workspace:" + local.Path, Checksum: string(item.digest), WorkspacePath: workspacePath, ArtifactManifestID: item.manifestID, SnapshotDigest: item.digest, Dependencies: deps})
		if local.Path == "." {
			rootKeys = []string{key}
		}
	}
	for _, snapshot := range graph.Snapshots {
		pkg := graph.Packages[graph.packageIndex[snapshot.PackageKey]]
		item := tarballs[pkg.Key]
		deps := []nodesource.Dependency{}
		for _, edge := range graph.Edges {
			if edge.From != snapshot.Key || edge.To == "" {
				continue
			}
			deps = append(deps, nodeDependency(edge, graph))
		}
		instances = append(instances, nodesource.PackageInstance{Key: snapshot.Key, Name: snapshot.Name, Version: snapshot.Version, Origin: pkg.Resolved, Checksum: pkg.Integrity, PeerKey: snapshot.PeerContext, ArtifactManifestID: item.manifestID, SnapshotDigest: item.digest, Dependencies: deps})
	}
	capture, err := nodesource.BuildCapture(nodesource.CaptureInput{Manager: nodesource.ManagerPNPM, RootKeys: rootKeys, Packages: instances, PolicyIDs: []string{artifactpolicy.PolicyID, ProfileID}})
	if err != nil {
		return nodesource.Capture{}, err
	}
	if len(capture.Graph.RootNodeIDs) != 1 {
		return nodesource.Capture{}, fail(CodeGraphIncomplete, "pnpm capture requires one command product", nil)
	}
	owner := capture.Graph.RootNodeIDs[0]
	return nodesource.AddRuntimeActions(capture, []nodesource.RuntimeAction{{Name: "pnpm-store-add", Subtype: "pnpm-store-add", OwnerNodeID: owner, ToolRole: "package-manager", ArgvTemplate: []string{"{{manager_entrypoint}}", "--store-dir", "{{store}}", "--config.side-effects-cache=false", "store", "add", "{{tarball}}"}, WorkingDirectory: "{{work}}", EnvironmentPolicyID: "pnpm-private-environment-v1", ProcessPolicyID: "pnpm-store-process-v1"}, {Name: "pnpm-install", Subtype: "pnpm-install", OwnerNodeID: owner, ToolRole: "package-manager", ArgvTemplate: []string{"{{manager_entrypoint}}", "install", "--frozen-lockfile", "--offline", "--ignore-scripts", "--store-dir", "{{store}}", "--config.side-effects-cache=false", "--package-import-method=copy", "{{prod}}"}, WorkingDirectory: "{{project}}", EnvironmentPolicyID: "pnpm-private-environment-v1", ProcessPolicyID: "pnpm-install-process-v1"}, {Name: "node-invoke", Subtype: "node-invoke", OwnerNodeID: owner, ToolRole: "node-runtime", ArgvTemplate: []string{"{{entrypoint}}", "{{args}}"}, WorkingDirectory: "{{project}}", EnvironmentPolicyID: "node-private-environment-v1", ProcessPolicyID: "node-runtime-process-v1"}})
}

func nodeDependency(edge DependencyEdge, graph Graph) nodesource.Dependency {
	scope := closuregraph.ScopeRuntime
	switch edge.Scope {
	case "development":
		scope = closuregraph.ScopeBuild
	case "optional":
		scope = closuregraph.ScopeOptional
	case "peer":
		scope = closuregraph.ScopePeer
	}
	key := edge.To
	condition := (*closuregraph.Condition)(nil)
	if index, ok := graph.snapshotIndex[key]; ok {
		snapshot := graph.Snapshots[index]
		pkg := graph.Packages[graph.packageIndex[snapshot.PackageKey]]
		condition = targetCondition(pkg)
	}
	return nodesource.Dependency{PackageKey: key, Scope: scope, Condition: condition, DeclarationField: edge.Scope}
}

func copyProjectSource(source, destination string) ([]string, error) {
	discarded := []string{}
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
			return fail(CodeLocalPathEscape, "project contains a symlink", map[string]string{"path": logical})
		}
		if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == ".pnpm-store" || entry.Name() == ".pnpm") {
			if err := scanAmbientSideEffects(current); err != nil {
				return err
			}
			discarded = append(discarded, logical)
			return filepath.SkipDir
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return privatedir.Make(target)
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return fail(CodeInputUndeclared, "project contains special node", map[string]string{"path": logical})
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the exact project root.
		if err != nil {
			return err
		}
		if strings.EqualFold(entry.Name(), ".pnpmfile.cjs") || strings.EqualFold(entry.Name(), ".pnpmfile.mjs") {
			return fail(CodeManagerPluginUndeclared, "pnpm hook file is unsupported", map[string]string{"path": logical})
		}
		return os.WriteFile(target, payload, 0o600)
	})
	sort.Strings(discarded)
	return discarded, err
}
func scanAmbientSideEffects(root string) error {
	return filepath.WalkDir(root, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.Contains(strings.ToLower(filepath.ToSlash(current)), "side_effect") || strings.Contains(strings.ToLower(filepath.ToSlash(current)), "side-effects") {
			return fail(CodeHookUndeclared, "pnpm side-effects cache is forbidden", map[string]string{"path": current})
		}
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), "-index.json") {
			payload, _ := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the rejected ambient store root.
			if bytes.Contains(payload, []byte("\"sideEffects\"")) {
				return fail(CodeHookUndeclared, "pnpm side-effects cache is forbidden", map[string]string{"path": current})
			}
		}
		return nil
	})
}

func verifySRI(integrity string, payload []byte) error {
	parts := strings.SplitN(integrity, "-", 2)
	if len(parts) != 2 || parts[0] != "sha512" {
		return fail(CodeIntegrityMissing, "pnpm integrity must be sha512 SRI", map[string]string{"integrity": integrity})
	}
	want, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(want) != sha512.Size {
		return fail(CodeIntegrityMissing, "pnpm integrity is malformed", nil)
	}
	actual := sha512.Sum512(payload)
	if subtle.ConstantTimeCompare(want, actual[:]) != 1 {
		return fail(CodeIntegrityMismatch, "raw pnpm tarball differs from lock integrity", nil)
	}
	return nil
}
func digestID(payload []byte) closuregraph.ID {
	sum := sha256.Sum256(payload)
	return closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
}
func treeEvidence(key, origin string, item capturedTree) ArtifactEvidence {
	return ArtifactEvidence{Key: key, Origin: origin, SHA256: string(item.digest), Size: item.size, ArtifactManifestID: item.manifestID, IntakeReceiptID: item.receiptID}
}
func safeArchiveName(pkg Package) string {
	return strings.NewReplacer("/", "_", "@", "_").Replace(pkg.Key)
}
func lifecycleScripts(values map[string]string) []string {
	result := []string{}
	for _, name := range []string{"preinstall", "install", "postinstall", "prepare", "prepublish", "prepublishOnly"} {
		if values[name] != "" {
			result = append(result, name)
		}
	}
	return result
}
func equalBoolMap(values map[string]struct{ Optional bool }, expected map[string]bool) bool {
	actual := map[string]bool{}
	for key, value := range values {
		if value.Optional {
			actual[key] = true
		}
	}
	if len(actual) != len(expected) {
		return false
	}
	for key, value := range actual {
		if expected[key] != value {
			return false
		}
	}
	return true
}
func equalSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for key := range a {
		if !b[key] {
			return false
		}
	}
	return true
}
func targetCondition(pkg Package) *closuregraph.Condition {
	if len(pkg.OS)+len(pkg.CPU)+len(pkg.Libc) == 0 {
		return nil
	}
	return &closuregraph.Condition{EvaluatorID: "pnpm-platform-selector-v1", Expression: strings.Join([]string{"os=" + strings.Join(pkg.OS, ","), "cpu=" + strings.Join(pkg.CPU, ","), "libc=" + strings.Join(pkg.Libc, ",")}, ";")}
}
