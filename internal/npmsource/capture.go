package npmsource

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

// Evidence is the complete npm capture/admission result.
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
// from one admitted raw tgz. npm's cache and installed tree are never allowed
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
		return nil, fail(CodeInputUndeclared, "npm capture authority is incomplete", nil)
	}
	root, err := filepath.Abs(request.ProjectRoot)
	if err != nil || root != filepath.Clean(request.ProjectRoot) {
		return nil, fail(CodeLocalPathEscape, "project root must be an absolute clean path", nil)
	}
	workRoot, err := filepath.Abs(request.WorkRoot)
	if err != nil || request.WorkRoot == "" {
		return nil, fail(CodeInputUndeclared, "npm work root is invalid", nil)
	}
	if err = privatedir.MakeAll(workRoot); err != nil {
		return nil, err
	}
	stage, err := os.MkdirTemp(workRoot, "npm-project-stage-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	discarded, err := copyProjectSource(root, stage)
	if err != nil {
		return nil, err
	}
	lockPath := filepath.Join(stage, request.Graph.LockName)
	lockBytes, err := os.ReadFile(lockPath) // #nosec G304 -- exact contained staged lock path.
	if err != nil || digestID(lockBytes) != closuregraph.ID(request.Graph.RawLockSHA256) {
		return nil, fail(CodeLockStale, "captured project lock differs from parsed lock", map[string]string{"lock": request.Graph.LockName})
	}
	for manifestPath, expected := range request.Graph.manifestBytes {
		actual, readErr := os.ReadFile(filepath.Join(stage, filepath.FromSlash(manifestPath))) // #nosec G304 -- validated manifest path.
		if readErr != nil || !bytes.Equal(actual, expected) {
			return nil, fail(CodeLockStale, "captured project manifest differs from parsed manifest", map[string]string{"path": manifestPath})
		}
	}
	project, err := admitProjectTree(ctx, request, stage)
	if err != nil {
		return nil, err
	}

	external := []Package{}
	for _, pkg := range request.Graph.Packages {
		if isExternalInstallPath(pkg.InstallPath) && !pkg.Link {
			external = append(external, pkg)
		}
	}
	if len(request.Tarballs) != len(external) {
		return nil, fail(CodeOfflineInputMissing, "raw npm tarball set is not bijective with the lock graph", map[string]string{"expected": fmt.Sprint(len(external)), "observed": fmt.Sprint(len(request.Tarballs))})
	}
	captured := make(map[string]capturedInput, len(external))
	evidence := make([]ArtifactEvidence, 0, len(external))
	for _, pkg := range external {
		raw, ok := request.Tarballs[pkg.InstallPath]
		if !ok {
			return nil, fail(CodeOfflineInputMissing, "required raw npm tarball is absent", map[string]string{"package": pkg.InstallPath})
		}
		item, itemEvidence, metadata, inspection, admitErr := admitTarball(ctx, request, pkg, raw)
		if admitErr != nil {
			return nil, admitErr
		}
		if err = reconcileEmbeddedMetadata(pkg, metadata, inspection); err != nil {
			return nil, err
		}
		captured[pkg.InstallPath] = item
		evidence = append(evidence, itemEvidence)
	}
	for key := range request.Tarballs {
		if _, ok := captured[key]; !ok {
			return nil, fail(CodeGraphIncomplete, "raw npm tarball has no lock instance", map[string]string{"package": key})
		}
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

func admitProjectTree(ctx context.Context, request CaptureRequest, stage string) (capturedInput, error) {
	probeDescriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "npm", PackageName: request.Graph.RootName, PackageVersion: request.Graph.RootVersion}
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
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeOfflineInputMissing, "raw npm tarball path must be absolute", map[string]string{"package": pkg.InstallPath})
	}
	info, err := os.Lstat(raw.Path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeOfflineInputMissing, "raw npm tarball is absent or not a regular file", map[string]string{"package": pkg.InstallPath})
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
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "npm", PackageName: pkg.Name, PackageVersion: pkg.Version, Origin: artifactpolicy.OriginEvidence{Locator: pkg.Resolved, ImmutableID: pkg.Integrity, LockRecord: request.Graph.LockDigest, ChecksumSHA256: string(digest), Verified: true}}
	admitted, err := request.Policy.AdmitDependency(ctx, artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: "npm/" + safeArchiveName(pkg) + ".tgz", Size: int64(len(payload)), Reader: reader}})
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
	evidence := ArtifactEvidence{PackagePath: pkg.InstallPath, Origin: pkg.Resolved, Integrity: pkg.Integrity, SHA256: string(digest), Size: int64(len(payload)), ArtifactManifestID: manifestID, IntakeReceiptID: receiptID}
	return item, evidence, metadata, inspection, nil
}

func inspectTarball(payload []byte) (packageManifest, tarInspection, []packageFile, error) {
	gz, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return packageManifest{}, tarInspection{}, nil, fail(CodeMetadataMismatch, "npm tarball gzip envelope is invalid", nil)
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
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "npm tarball is invalid: "+readErr.Error(), nil)
		}
		name := path.Clean(header.Name)
		if path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "npm tarball member escapes package root", map[string]string{"path": header.Name})
		}
		if name != "package" && !strings.HasPrefix(name, "package/") {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "npm tarball member is outside the package root", map[string]string{"path": header.Name})
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if !header.FileInfo().Mode().IsRegular() {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "npm tarball contains a non-regular package member", map[string]string{"path": header.Name})
		}
		rel := strings.TrimPrefix(name, "package/")
		if rel == "" || seen[rel] {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "npm tarball contains a duplicate or empty package member", map[string]string{"path": header.Name})
		}
		seen[rel] = true
		if header.Size < 0 {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "npm tarball member has an invalid size", map[string]string{"path": header.Name})
		}
		data, readErr := io.ReadAll(io.LimitReader(reader, header.Size+1))
		if readErr != nil || int64(len(data)) != header.Size {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cannot read npm tarball member", map[string]string{"path": header.Name})
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
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "npm tarball contains duplicate package.json", nil)
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
		return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "npm tarball lacks package/package.json", nil)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return manifest, inspection, files, nil
}

func reconcileEmbeddedMetadata(pkg Package, manifest packageManifest, inspection tarInspection) error {
	if manifest.Name != pkg.Name || manifest.Version != pkg.Version {
		return fail(CodeMetadataMismatch, "embedded package identity differs from npm lock", map[string]string{"package": pkg.InstallPath, "lock_name": pkg.Name, "artifact_name": manifest.Name, "lock_version": pkg.Version, "artifact_version": manifest.Version})
	}
	for _, item := range []struct {
		name           string
		lock, embedded map[string]string
	}{{"dependencies", withoutKeys(pkg.Dependencies, pkg.OptionalDependencies), withoutKeys(manifest.Dependencies, manifest.OptionalDependencies)}, {"optionalDependencies", pkg.OptionalDependencies, manifest.OptionalDependencies}, {"peerDependencies", pkg.PeerDependencies, manifest.PeerDependencies}} {
		if !equalStringMap(item.lock, item.embedded) {
			return fail(CodeMetadataMismatch, "embedded dependency metadata differs from npm lock", map[string]string{"package": pkg.InstallPath, "field": item.name})
		}
	}
	if !equalBoolMap(pkg.PeerOptional, peerOptional(manifest.PeerDependenciesMeta)) {
		return fail(CodeMetadataMismatch, "embedded peer metadata differs from npm lock", map[string]string{"package": pkg.InstallPath, "field": "peerDependenciesMeta"})
	}
	if !equalStrings(pkg.OS, sortedStrings(manifest.OS)) || !equalStrings(pkg.CPU, sortedStrings(manifest.CPU)) || !equalStrings(pkg.Libc, sortedStrings(manifest.Libc)) {
		return fail(CodeMetadataMismatch, "embedded platform metadata differs from npm lock", map[string]string{"package": pkg.InstallPath})
	}
	lifecycle := lifecycleScripts(manifest.Scripts)
	if pkg.HasInstallScript != (len(lifecycle) > 0) {
		return fail(CodeMetadataMismatch, "hasInstallScript differs from embedded lifecycle metadata", map[string]string{"package": pkg.InstallPath})
	}
	if inspection.bundled || rawBundlePresent(manifest.BundleDependencies) || rawBundlePresent(manifest.BundledDependencies) {
		return fail(CodeBundledDependencyUnsupported, "npm tarball embeds bundled dependencies", map[string]string{"package": pkg.InstallPath})
	}
	gypEnabled := manifest.Gypfile == nil || *manifest.Gypfile
	if inspection.bindingGYP && gypEnabled && manifest.Scripts["install"] == "" && manifest.Scripts["preinstall"] == "" {
		return fail(CodeNativeBuildUnsupported, "binding.gyp would trigger implicit node-gyp rebuild", map[string]string{"package": pkg.InstallPath})
	}
	if len(lifecycle) > 0 {
		return fail(CodeHookUndeclared, "dependency lifecycle execution is outside npm-source-v1", map[string]string{"package": pkg.InstallPath, "script": lifecycle[0]})
	}
	return nil
}

func buildNodeCapture(graph Graph, project capturedInput, tarballs map[string]capturedInput) (nodesource.Capture, error) {
	instances := []nodesource.PackageInstance{}
	for _, pkg := range graph.Packages {
		if pkg.Link {
			continue
		}
		item := project
		origin := "workspace:."
		checksum := string(project.digest)
		workspace := pkg.WorkspacePath
		if isExternalInstallPath(pkg.InstallPath) {
			item = tarballs[pkg.InstallPath]
			origin = pkg.Resolved
			checksum = pkg.Integrity
			workspace = ""
		}
		dependencies := []nodesource.Dependency{}
		for _, edge := range graph.Edges {
			if edge.From != pkg.InstallPath || edge.To == "" {
				continue
			}
			target := graph.Packages[graph.packageIndex[edge.To]]
			if target.Link {
				continue
			}
			scope := closuregraph.ScopeRuntime
			switch edge.Scope {
			case "development":
				scope = closuregraph.ScopeBuild
			case "optional":
				scope = closuregraph.ScopeOptional
			case "peer":
				scope = closuregraph.ScopePeer
			}
			dependencies = append(dependencies, nodesource.Dependency{PackageKey: packageKey(edge.To), Scope: scope, Condition: targetCondition(target), DeclarationField: edge.Scope})
		}
		instances = append(instances, nodesource.PackageInstance{Key: pkg.Key, Name: pkg.Name, Version: pkg.Version, Origin: origin, Checksum: checksum, WorkspacePath: workspace, ArtifactManifestID: item.manifestID, SnapshotDigest: item.digest, Dependencies: dependencies})
	}
	capture, err := nodesource.BuildCapture(nodesource.CaptureInput{Manager: nodesource.ManagerNPM, RootKeys: []string{"root"}, Packages: instances, PolicyIDs: []string{artifactpolicy.PolicyID, ProfileID}})
	if err != nil {
		return nodesource.Capture{}, err
	}
	if len(capture.Graph.RootNodeIDs) != 1 {
		return nodesource.Capture{}, fail(CodeGraphIncomplete, "npm capture requires one command product", nil)
	}
	owner := capture.Graph.RootNodeIDs[0]
	return nodesource.AddRuntimeActions(capture, []nodesource.RuntimeAction{
		{Name: "npm-cache", Subtype: "npm-cache", OwnerNodeID: owner, ToolRole: "package-manager", ArgvTemplate: []string{"{{manager_entrypoint}}", "cache", "add", "{{tarball}}", "--offline", "--ignore-scripts", "--cache", "{{cache}}", "--userconfig", "{{userconfig}}", "--logs-dir", "{{logs}}", "--no-audit", "--no-fund"}, WorkingDirectory: "{{work}}", EnvironmentPolicyID: "npm-private-environment-v1", ProcessPolicyID: "npm-cache-process-v1"},
		{Name: "npm-ci", Subtype: "npm-ci", OwnerNodeID: owner, ToolRole: "package-manager", ArgvTemplate: []string{"{{manager_entrypoint}}", "ci", "--offline", "--ignore-scripts", "--cache", "{{cache}}", "--userconfig", "{{userconfig}}", "--logs-dir", "{{logs}}", "--no-audit", "--no-fund", "{{omit_dev}}"}, WorkingDirectory: "{{project}}", EnvironmentPolicyID: "npm-private-environment-v1", ProcessPolicyID: "npm-ci-process-v1"},
		{Name: "node-invoke", Subtype: "node-invoke", OwnerNodeID: owner, ToolRole: "node-runtime", ArgvTemplate: []string{"{{entrypoint}}", "{{args}}"}, WorkingDirectory: "{{project}}", EnvironmentPolicyID: "node-private-environment-v1", ProcessPolicyID: "node-runtime-process-v1"},
	})
}

func copyProjectSource(source, destination string) ([]string, error) {
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
		if entry.IsDir() && (base == "node_modules" || base == ".npm" || base == ".pnpm-store" || base == ".yarn") {
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
		return fail(CodeIntegrityMissing, "npm SRI is not canonical sha512", nil)
	}
	actual := sha512.Sum512(payload)
	if subtle.ConstantTimeCompare(expected, actual[:]) != 1 {
		return fail(CodeIntegrityMismatch, "raw npm tarball differs from lock SRI", map[string]string{"integrity": integrity})
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
func equalBoolMap(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
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
