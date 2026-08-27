package yarnmodernsource

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/nodesource"
	"github.com/relux-works/curator/internal/privatedir"
)

// RawArchive binds either raw npm tgz bytes or an exact normalized Yarn cache ZIP.
// YarnChecksum is the complete cacheKey/checksum value from yarn.lock. CacheName
// is required for normalized ZIPs so the manager-visible identity is retained.
type RawArchive struct {
	Path, Format, SHA256, YarnChecksum, CacheName string
}

// RawTarball remains an input spelling for callers that capture raw npm archives.
type RawTarball = RawArchive

// CaptureRequest supplies only raw source inputs and manager-owned stores.
// WorkRoot is scratch authority, never closure authority.
type CaptureRequest struct {
	Graph              Graph
	ProjectRoot        string
	Tarballs           map[string]RawTarball
	Archives           map[string]RawArchive
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
	cacheBytes []byte
	cacheName  string
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
	stage, err := os.MkdirTemp(workRoot, "yarn-modern-project-stage-")
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
	for patchPath, expected := range request.Graph.patchBytes {
		actual, readErr := os.ReadFile(filepath.Join(stage, filepath.FromSlash(patchPath))) // #nosec G304 -- validated contained patch path.
		if readErr != nil || !bytes.Equal(actual, expected) {
			return nil, fail(CodeIntegrityMismatch, "captured Yarn patch differs from parsed patch", map[string]string{"path": patchPath})
		}
	}
	project, err := admitProjectTree(ctx, request, stage)
	if err != nil {
		return nil, err
	}

	archives := request.Archives
	if len(archives) == 0 {
		archives = request.Tarballs
	}
	external := []Package{}
	for _, pkg := range request.Graph.Packages {
		if pkg.BaseKey == "" && pkg.Resolved != "" {
			external = append(external, pkg)
		}
	}
	if len(archives) != len(external) {
		return nil, fail(CodeOfflineInputMissing, "modern Yarn archive set is not bijective with the lock graph", map[string]string{"expected": fmt.Sprint(len(external)), "observed": fmt.Sprint(len(archives))})
	}
	captured := make(map[string]capturedInput, len(external))
	evidence := make([]ArtifactEvidence, 0, len(external))
	for _, pkg := range external {
		raw, ok := archives[pkg.Key]
		if !ok {
			return nil, fail(CodeOfflineInputMissing, "required raw Yarn tarball is absent", map[string]string{"package": pkg.Key})
		}
		item, itemEvidence, metadata, inspection, admitErr := admitArchive(ctx, request, pkg, raw)
		if admitErr != nil {
			return nil, admitErr
		}
		if err = reconcileEmbeddedMetadata(pkg, metadata, inspection); err != nil {
			return nil, err
		}
		index := request.Graph.packageIndex[pkg.Key]
		request.Graph.Packages[index].manifest = metadata
		request.Graph.Packages[index].OS = sortedStrings(metadata.OS)
		request.Graph.Packages[index].CPU = sortedStrings(metadata.CPU)
		request.Graph.Packages[index].Libc = sortedStrings(metadata.Libc)
		captured[pkg.Key] = item
		evidence = append(evidence, itemEvidence)
	}
	for key := range archives {
		if _, ok := captured[key]; !ok {
			return nil, fail(CodeGraphIncomplete, "raw Yarn tarball has no lock instance", map[string]string{"package": key})
		}
	}
	basePackages := make([]Package, 0, len(request.Graph.Packages))
	for _, pkg := range request.Graph.Packages {
		if pkg.BaseKey == "" {
			basePackages = append(basePackages, pkg)
		}
	}
	request.Graph.Packages, request.Graph.Edges, request.Graph.packageIndex, err = buildPackageGraph(basePackages, request.Graph.selectorIndex, request.Graph.workspaceByName)
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
	discoveredPatches := map[string]bool{}
	patchRoot := filepath.Join(root, ".yarn", "patches")
	if _, statErr := os.Lstat(patchRoot); statErr == nil {
		if walkErr := filepath.WalkDir(patchRoot, func(current string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			rel, relErr := filepath.Rel(root, current)
			if relErr != nil {
				return relErr
			}
			discoveredPatches[filepath.ToSlash(rel)] = true
			return nil
		}); walkErr != nil {
			return walkErr
		}
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return statErr
	}
	if err := requirePatchBijection(discoveredPatches, graph.patchBytes); err != nil {
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

func requirePatchBijection(discovered map[string]bool, expected map[string][]byte) error {
	for logical := range discovered {
		if _, ok := expected[logical]; !ok {
			return fail(CodeManagerPluginUndeclared, "captured Yarn patch is absent from lock authority", map[string]string{"path": logical})
		}
	}
	for logical := range expected {
		if !discovered[logical] {
			return fail(CodeManagerPluginUndeclared, "lock-declared Yarn patch is absent from captured tree", map[string]string{"path": logical})
		}
	}
	return nil
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
	probeDescriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "yarn-modern", PackageName: request.Graph.RootName, PackageVersion: request.Graph.RootVersion}
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

func admitArchive(ctx context.Context, request CaptureRequest, pkg Package, raw RawArchive) (capturedInput, ArtifactEvidence, packageManifest, tarInspection, error) {
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
	if raw.SHA256 != "" && raw.SHA256 != string(digestID(payload)) {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeIntegrityMismatch, "raw archive differs from captured SHA-256", map[string]string{"package": pkg.Key})
	}
	format := raw.Format
	if format == "" {
		if bytes.HasPrefix(payload, []byte{'P', 'K', 3, 4}) {
			format = "zip"
		} else {
			format = "tgz"
		}
	}
	if raw.YarnChecksum != "" && raw.YarnChecksum != pkg.Checksum {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeIntegrityMismatch, "archive Yarn checksum differs from yarn.lock", map[string]string{"package": pkg.Key})
	}
	var metadata packageManifest
	var inspection tarInspection
	var files []packageFile
	var cacheBytes []byte
	switch format {
	case "zip":
		metadata, inspection, files, err = inspectZip(payload)
		cacheBytes = append([]byte(nil), payload...)
	case "tgz":
		metadata, inspection, files, err = inspectTarball(payload)
		if err == nil {
			cacheBytes, err = normalizeCacheZip(pkg, payload, request.Graph.Layout.CompressionLevel)
		}
	default:
		err = fail(CodeLockFormatUnsupported, "unsupported modern Yarn archive format", map[string]string{"format": format})
	}
	if err != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, err
	}
	observedYarnChecksum := CacheChecksum(cacheBytes, request.Graph.Layout.CacheKey)
	if observedYarnChecksum != pkg.Checksum {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeIntegrityMismatch, "normalized Yarn cache ZIP differs from lock checksum", map[string]string{"package": pkg.Key, "expected": pkg.Checksum, "observed": observedYarnChecksum})
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
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "yarn-modern", PackageName: pkg.Name, PackageVersion: pkg.Version, Origin: artifactpolicy.OriginEvidence{Locator: pkg.Resolution, ImmutableID: pkg.Checksum, LockRecord: request.Graph.LockDigest, ChecksumSHA256: string(digest), Verified: true}}
	admitted, err := request.Policy.AdmitDependency(ctx, artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: "yarn-modern/" + safeArchiveName(pkg) + "." + format, Size: int64(len(payload)), Reader: reader}})
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
	expectedCacheName, nameErr := yarnCacheName(pkg)
	if nameErr != nil {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, nameErr
	}
	cacheName := raw.CacheName
	if cacheName == "" {
		cacheName = expectedCacheName
	}
	if path.Base(cacheName) != cacheName || cacheName != expectedCacheName {
		return capturedInput{}, ArtifactEvidence{}, packageManifest{}, tarInspection{}, fail(CodeIntegrityMismatch, "cache archive name differs from pinned Yarn identity", map[string]string{"expected": expectedCacheName, "observed": cacheName})
	}
	item := capturedInput{handle: handle, input: closureexec.AdmittedInput{Receipt: receipt, Handle: handle}, receiptID: receiptID, manifestID: manifestID, digest: digest, size: int64(len(payload)), files: files, cacheBytes: cacheBytes, cacheName: cacheName}
	evidence := ArtifactEvidence{PackagePath: pkg.Key, Origin: pkg.Resolved, Integrity: pkg.Integrity, SHA256: string(digest), Size: int64(len(payload)), ArtifactManifestID: manifestID, IntakeReceiptID: receiptID}
	return item, evidence, metadata, inspection, nil
}

func inspectZip(payload []byte) (packageManifest, tarInspection, []packageFile, error) {
	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return packageManifest{}, tarInspection{}, nil, fail(CodeMetadataMismatch, "Yarn cache ZIP is invalid", nil)
	}
	var manifest packageManifest
	inspection := tarInspection{}
	files := []packageFile{}
	found := false
	seen := map[string]bool{}
	for _, member := range reader.File {
		name := path.Clean(member.Name)
		if path.IsAbs(name) || strings.HasPrefix(name, "../") || member.FileInfo().IsDir() {
			continue
		}
		marker := "/node_modules/"
		position := strings.Index("/"+name, marker)
		if position < 0 {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cache ZIP member is outside node_modules package root", map[string]string{"path": name})
		}
		relRoot := strings.TrimPrefix(("/" + name)[position+len(marker):], "/")
		parts := strings.Split(relRoot, "/")
		rootParts := 1
		if strings.HasPrefix(relRoot, "@") {
			rootParts = 2
		}
		if len(parts) <= rootParts {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cache ZIP package member is malformed", map[string]string{"path": name})
		}
		rel := strings.Join(parts[rootParts:], "/")
		if seen[rel] {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cache ZIP has duplicate package member", map[string]string{"path": rel})
		}
		seen[rel] = true
		stream, openErr := member.Open()
		if openErr != nil {
			return packageManifest{}, inspection, nil, openErr
		}
		data, readErr := io.ReadAll(io.LimitReader(stream, member.FileInfo().Size()+1))
		_ = stream.Close()
		if readErr != nil || int64(len(data)) != member.FileInfo().Size() {
			return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cannot read cache ZIP member", map[string]string{"path": name})
		}
		files = append(files, packageFile{Path: rel, SHA256: digestID(data), Size: int64(len(data)), Executable: member.Mode()&0o111 != 0})
		if rel == "binding.gyp" {
			inspection.bindingGYP = true
		}
		if strings.HasPrefix(rel, "node_modules/") {
			inspection.bundled = true
		}
		if rel == "package.json" {
			if found {
				return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cache ZIP contains duplicate package.json", nil)
			}
			found = true
			if validateJSON(data) != nil || json.Unmarshal(data, &manifest) != nil {
				return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cache ZIP package.json is invalid", nil)
			}
		}
	}
	if !found {
		return packageManifest{}, inspection, nil, fail(CodeMetadataMismatch, "cache ZIP lacks package.json", nil)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return manifest, inspection, files, nil
}

func normalizeCacheZip(pkg Package, payload []byte, compression int) ([]byte, error) {
	input, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer func() { _ = input.Close() }()
	tarReader := tar.NewReader(input)
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	method := uint16(zip.Store)
	if compression > 0 {
		method = zip.Deflate
	}
	prefix := "node_modules/" + pkg.Name + "/"
	for {
		header, readErr := tarReader.Next()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
		name := path.Clean(header.Name)
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if !strings.HasPrefix(name, "package/") || !header.FileInfo().Mode().IsRegular() {
			return nil, fail(CodeMetadataMismatch, "raw archive member cannot be normalized", map[string]string{"path": name})
		}
		data, readErr := io.ReadAll(io.LimitReader(tarReader, header.Size+1))
		if readErr != nil || int64(len(data)) != header.Size {
			return nil, fail(CodeMetadataMismatch, "raw archive member is truncated", map[string]string{"path": name})
		}
		zipHeader := &zip.FileHeader{Name: prefix + strings.TrimPrefix(name, "package/"), Method: method}
		zipHeader.SetMode(header.FileInfo().Mode())
		entry, createErr := writer.CreateHeader(zipHeader)
		if createErr != nil {
			return nil, createErr
		}
		if _, createErr = entry.Write(data); createErr != nil {
			return nil, createErr
		}
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func shortChecksum(value string) string {
	_, checksum, ok := strings.Cut(value, "/")
	if !ok || len(checksum) < 10 {
		return ""
	}
	return checksum[:10]
}

func yarnCacheName(pkg Package) (string, error) {
	prefix := pkg.Name + "@"
	if !strings.HasPrefix(pkg.Resolution, prefix) {
		return "", fail(CodeGraphIncomplete, "package resolution cannot form Yarn cache identity", map[string]string{"package": pkg.Key})
	}
	reference := strings.TrimPrefix(pkg.Resolution, prefix)
	protocol, _, ok := strings.Cut(reference, ":")
	if !ok || protocol == "" {
		return "", fail(CodeGraphIncomplete, "package reference lacks Yarn cache protocol", map[string]string{"package": pkg.Key})
	}
	scope, name := "", pkg.Name
	slug := pkg.Name
	if strings.HasPrefix(pkg.Name, "@") {
		parts := strings.SplitN(strings.TrimPrefix(pkg.Name, "@"), "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", fail(CodeGraphIncomplete, "scoped package name is malformed", map[string]string{"package": pkg.Key})
		}
		scope, name = parts[0], parts[1]
		slug = "@" + scope + "-" + name
	}
	identSeed := scope + name
	identHash := sha512.Sum512([]byte(identSeed))
	locatorHash := sha512.Sum512([]byte(hex.EncodeToString(identHash[:]) + reference))
	protocolSlug := protocol
	if protocol == "npm" {
		protocolSlug += "-" + pkg.Version
	}
	checksum := shortChecksum(pkg.Checksum)
	if checksum == "" {
		return "", fail(CodeIntegrityMissing, "package checksum cannot form Yarn cache identity", map[string]string{"package": pkg.Key})
	}
	return fmt.Sprintf("%s-%s-%s-%s.zip", slug, protocolSlug, hex.EncodeToString(locatorHash[:5]), checksum), nil
}

// CacheChecksum reproduces the modern Yarn v4 cache checksum representation.
func CacheChecksum(payload []byte, cacheKey string) string {
	sum := sha512.Sum512(payload)
	return cacheKey + "/" + hex.EncodeToString(sum[:])
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
	}{{"dependencies", withoutKeys(pkg.Dependencies, pkg.OptionalDependencies), withoutKeys(manifest.Dependencies, manifest.OptionalDependencies)}, {"optionalDependencies", pkg.OptionalDependencies, manifest.OptionalDependencies}, {"peerDependencies", pkg.PeerDependencies, manifest.PeerDependencies}} {
		if !equalStringMap(item.lock, item.embedded) {
			return fail(CodeMetadataMismatch, "embedded dependency metadata differs from yarn.lock", map[string]string{"package": pkg.Key, "field": item.name})
		}
	}
	if embeddedPeerOptional := peerOptional(manifest.PeerDependenciesMeta); !equalBoolMap(pkg.PeerOptional, embeddedPeerOptional) {
		return fail(CodeMetadataMismatch, "embedded dependency metadata differs from yarn.lock", map[string]string{"package": pkg.Key, "field": "peerDependenciesMeta"})
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
		return fail(CodeHookUndeclared, "dependency lifecycle execution is outside yarn-modern-source-v1", map[string]string{"package": pkg.Key, "script": lifecycle[0]})
	}
	return nil
}

func buildNodeCapture(graph Graph, project capturedInput, tarballs map[string]capturedInput) (nodesource.Capture, error) {
	instances := []nodesource.PackageInstance{}
	for _, pkg := range graph.Packages {
		sourceKey := packageSourceKey(pkg)
		item := project
		origin := "workspace:" + pkg.WorkspacePath
		checksum := string(project.digest)
		workspace := pkg.WorkspacePath
		if pkg.Key == "workspace:." {
			workspace = ""
		}
		if pkg.Resolved != "" {
			item = tarballs[sourceKey]
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
	capture, err := nodesource.BuildCapture(nodesource.CaptureInput{Manager: nodesource.ManagerYarnModern, RootKeys: []string{"workspace:."}, Packages: instances, PolicyIDs: []string{artifactpolicy.PolicyID, ProfileID}})
	if err != nil {
		return nodesource.Capture{}, err
	}
	if len(capture.Graph.RootNodeIDs) != 1 {
		return nodesource.Capture{}, fail(CodeGraphIncomplete, "Yarn capture requires one command product", nil)
	}
	owner := capture.Graph.RootNodeIDs[0]
	invokeTemplate := []string{"{{entrypoint}}", "{{args}}"}
	if graph.Layout.NodeLinker == "pnp" {
		invokeTemplate = []string{"--require", "{{pnp_loader}}", "{{entrypoint}}", "{{args}}"}
	}
	return nodesource.AddRuntimeActions(capture, []nodesource.RuntimeAction{
		{Name: "yarn-install", Subtype: "yarn-install", OwnerNodeID: owner, ToolRole: "package-manager", ArgvTemplate: []string{"{{manager_entrypoint}}", "install", "--immutable", "--immutable-cache", "--mode=skip-build"}, WorkingDirectory: "{{project}}", EnvironmentPolicyID: "yarn-modern-private-environment-v1", ProcessPolicyID: "yarn-modern-install-process-v1"},
		{Name: "node-invoke", Subtype: "node-invoke", OwnerNodeID: owner, ToolRole: "node-runtime", ArgvTemplate: invokeTemplate, WorkingDirectory: "{{project}}", EnvironmentPolicyID: "node-private-environment-v1", ProcessPolicyID: "node-runtime-process-v1"},
	})
}

func copyProjectSource(source, destination, _ string) ([]string, error) {
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
		if logical == ".pnp.cjs" || logical == ".pnp.loader.mjs" || logical == ".yarn/install-state.gz" {
			discarded[logical] = true
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() && (base == "node_modules" || base == ".cache" || logical == ".yarn/cache" || logical == ".yarn/unplugged" || logical == ".yarn/sdks") {
			discarded[logical] = true
			return filepath.SkipDir
		}
		if logical == ".yarn/plugins" || logical == ".yarn/releases" {
			return fail(CodeManagerPluginUndeclared, "local or downloaded Yarn extensions are forbidden", map[string]string{"path": logical})
		}
		if strings.HasPrefix(logical, ".yarn/") && !strings.HasPrefix(logical, ".yarn/patches/") && logical != ".yarn/patches" {
			return fail(CodeManagerPluginUndeclared, "unsupported Yarn project state is present", map[string]string{"path": logical})
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
	return slices.Equal(a, b)
}
