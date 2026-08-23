package rustsource

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
)

// ManagerConfig contains operator policy and its preselected host execution
// provider. Build and capture requests cannot replace either authority.
type ManagerConfig struct {
	WorkRoot  string
	Assurance closureexec.AssuranceConfig
	// Provider is operator-selected execution authority. Build/capture request
	// data cannot replace it after manager construction.
	Provider closureexec.VerifiedProvider
	// ProcessStartObserver receives every manager-owned portable launch attempt.
	// It is installed before C0 registration and proves that discovery,
	// assurance preflight, and permit construction do not cross the launch seam.
	ProcessStartObserver closureexec.ProcessStartObserver
}

// RawFile names a caller-supplied data file; it carries no execution authority.
type RawFile struct{ Path string }

// RawTree names a caller-supplied source tree.
type RawTree struct{ Root string }

// RawManifest binds a workspace-relative manifest name to its raw file.
type RawManifest struct {
	Path string
	File RawFile
}

// RawRegistryOrigin supplies immutable registry index and archive bytes.
type RawRegistryOrigin struct {
	SourceLocator string
	IndexRecord   RawFile
	CrateArchive  RawFile
}

// RawGitOrigin supplies a repository and exact locked Git selector.
type RawGitOrigin struct {
	DeclaredURL  string
	Selector     string
	LockedCommit string
	Repository   RawTree
}

// RawPathOrigin supplies one workspace-contained local package tree.
type RawPathOrigin struct {
	DeclaredPath string
	Tree         RawTree
}

// RawCaptureRequest is intentionally data-only.
type RawCaptureRequest struct {
	Workspace RawTree
	Lock      RawFile
	Manifests []RawManifest
	Registry  []RawRegistryOrigin
	Git       []RawGitOrigin
	Paths     []RawPathOrigin
}

// CaptureEvidence is the detached immutable evidence emitted by Capture.
type CaptureEvidence struct {
	LockDigest, TransformID, VendorReceipt, CargoHomeDigest, ConfigSHA256 string
	ArtifactManifestIDs                                                   []string
	GitObjectReceipts                                                     []string
	GitProjectionReceipts                                                 []string
	VendorPackages                                                        []VendorPackage
}

// Capture is manager-bound evidence for later metadata derivation.
type Capture struct {
	Evidence CaptureEvidence
	state    *captureState
}

type captureState struct {
	owner        *managerState
	workspace    string
	manifestPath string
	result       captureResult
	graph        CaptureGraph
}

// Manager is a concrete opaque operation authority.
type Manager struct{ state *managerState }

type managerState struct {
	mu             sync.Mutex
	seal           *managerSeal
	session        string
	cargoHome      string
	vendor         string
	cargo          cargoToolchain
	cargoRegistry  cargoRegistration
	executor       *closureexec.Executor
	buildOperation *closureexec.AssuredOperation
	processRunner  *closureexec.ManagerProcessRunner
	intake         *closureexec.CaptureStore
	execRoot       string
	outputRoot     string
	oraclePath     string
	oracleSHA      closuregraph.ID
	causalHead     string
	gitInputs      map[string]admittedGitInput
	workspaceInput closureexec.AdmittedInput
	workspaceID    closuregraph.ID
	lock           LockFile
	oracleReceipts []string
	buildTools     rustBuildToolchain
	closed         bool
	captureMade    bool
}

type admittedGitInput struct {
	input          closureexec.AdmittedInput
	adminInput     closureexec.AdmittedInput
	receiptID      closuregraph.ID
	adminReceiptID closuregraph.ID
}
type managerSeal struct{}

var productionManagerSeal = &managerSeal{}

// NewManager creates one sealed Rust source-capture authority.
func NewManager(ctx context.Context, config ManagerConfig) (*Manager, error) {
	if _, err := closureexec.PreflightAssurance(ctx, config.Assurance, config.Provider); err != nil {
		return nil, err
	}
	if !filepath.IsAbs(config.WorkRoot) {
		return nil, fail(CodeConfigUntrusted, "manager work root must be absolute", nil)
	}
	if err := os.MkdirAll(config.WorkRoot, 0o700); err != nil {
		return nil, err
	}
	session, err := os.MkdirTemp(config.WorkRoot, "rust-source-v1-")
	if err != nil {
		return nil, err
	}
	registration := registerCargoAtC0(ctx)
	tool, err := registration.recheck(ctx)
	if err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	buildTools, err := registerRustBuildToolchain(registration)
	if err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	execRoot := filepath.Join(session, "execution")
	outputRoot := filepath.Join(execRoot, "output")
	if err = os.MkdirAll(filepath.Join(execRoot, "bin"), 0o700); err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	if err = os.Mkdir(filepath.Join(execRoot, "work"), 0o700); err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	runningExecutable, err := os.Executable()
	if err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	executableBytes, err := os.ReadFile(runningExecutable) // #nosec G304 -- current trusted Curator executable.
	if err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	oraclePath := filepath.Join(execRoot, "bin", "curator")
	if err = os.WriteFile(oraclePath, executableBytes, 0o500); err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	oracleSum := sha256.Sum256(executableBytes)
	oracleSHA := closuregraph.ID("sha256:" + hex.EncodeToString(oracleSum[:]))
	runner, err := closureexec.NewManagerProcessRunner(execRoot, outputRoot)
	if err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	runner.ProcessStartObserver = config.ProcessStartObserver
	head := "rust-source-manager-v1:" + filepath.Base(session)
	executor, err := closureexec.NewAssuredExecutor(config.Assurance, runner, config.Provider, head)
	if err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	intake, err := closureexec.NewCaptureStore(filepath.Join(session, "intake"))
	if err != nil {
		_ = os.RemoveAll(session)
		return nil, err
	}
	state := &managerState{seal: productionManagerSeal, session: session, cargoHome: filepath.Join(session, "cargo-home"), vendor: filepath.Join(outputRoot, "vendor"), cargo: tool, cargoRegistry: registration, executor: executor, processRunner: runner, intake: intake, execRoot: execRoot, outputRoot: outputRoot, oraclePath: oraclePath, oracleSHA: oracleSHA, causalHead: head, gitInputs: map[string]admittedGitInput{}, buildTools: buildTools}
	return &Manager{state: state}, nil
}

func (state *managerState) recheckCargo(ctx context.Context) (cargoToolchain, error) {
	return state.cargoRegistry.recheck(ctx)
}

func (m *Manager) authority() (*managerState, error) {
	if m == nil || m.state == nil || m.state.seal != productionManagerSeal {
		return nil, fail(CodeConfigUntrusted, "manager authority is absent or foreign", nil)
	}
	return m.state, nil
}

// Capture admits all raw origins, runs the permitted vendor transform, and
// returns exact detached evidence plus an opaque manager-owned handle.
func (m *Manager) Capture(ctx context.Context, request RawCaptureRequest) (*Capture, error) {
	state, err := m.authority()
	if err != nil {
		return nil, err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.closed || state.captureMade {
		return nil, fail(CodeConfigUntrusted, "capture phase is unavailable", nil)
	}
	workspace, err := filepath.Abs(request.Workspace.Root)
	if err != nil || request.Workspace.Root == "" {
		return nil, fail(CodePathDependencyEscape, "workspace is invalid", nil)
	}
	if request.Lock.Path != filepath.Join(workspace, "Cargo.lock") {
		return nil, fail(CodeLockMismatch, "lock file is not the workspace Cargo.lock", nil)
	}
	lockBytes, err := rawBytes(request.Lock)
	if err != nil {
		return nil, err
	}
	lock, err := ParseLock(lockBytes)
	if err != nil {
		return nil, err
	}
	state.lock = lock
	manifestPath := filepath.Join(workspace, "Cargo.toml")
	manifests := make([]Manifest, 0, len(request.Manifests))
	for _, raw := range request.Manifests {
		if raw.Path == "" || pathpkg.IsAbs(raw.Path) || pathpkg.Clean(raw.Path) != raw.Path || strings.HasPrefix(raw.Path, "../") || raw.File.Path != filepath.Join(workspace, filepath.FromSlash(raw.Path)) {
			return nil, fail(CodePathDependencyEscape, "manifest path is not workspace-contained", map[string]string{"path": raw.Path})
		}
		payload, readErr := rawBytes(raw.File)
		if readErr != nil {
			return nil, readErr
		}
		manifest, parseErr := ParseManifest(raw.Path, payload)
		if parseErr != nil {
			return nil, parseErr
		}
		manifest.Path = pathpkg.Join("workspace", raw.Path)
		manifests = append(manifests, manifest)
		if raw.Path == "Cargo.toml" {
			manifestPath = raw.File.Path
		}
	}
	registry, err := bindRegistry(lock, request.Registry)
	if err != nil {
		return nil, err
	}
	if err = m.preAdmitGitSourceTrees(ctx, lock, request.Git); err != nil {
		return nil, err
	}
	gitOrigins, derivations, err := m.bindGit(ctx, lock, request.Git)
	if err != nil {
		return nil, err
	}
	paths, err := bindPaths(lock, workspace, request.Paths)
	if err != nil {
		return nil, err
	}
	runner := &managerVendorRunner{manager: state, workspace: workspace, registry: registry, git: gitOrigins}
	result, err := captureAndVendor(ctx, captureRequest{Lock: lock, Registry: registry, Git: gitOrigins, Paths: paths, WorkspaceRoot: workspace, VendorDestination: state.vendor, CargoHome: state.cargoHome, ConfigPath: filepath.Join(state.cargoHome, "config.toml"), StageCargoHome: runner.stageCargoHome, Toolchain: state.cargo, RecheckToolchain: func() (cargoToolchain, error) { return state.recheckCargo(ctx) }, Runner: runner, gitDerivations: derivations})
	if err != nil {
		return nil, err
	}
	state.captureMade = true
	gitObjectReceipts := make([]string, 0, len(state.gitInputs))
	for _, admitted := range state.gitInputs {
		gitObjectReceipts = append(gitObjectReceipts, string(admitted.adminReceiptID))
	}
	sort.Strings(gitObjectReceipts)
	evidence := CaptureEvidence{LockDigest: result.LockDigest, TransformID: result.TransformID, VendorReceipt: result.VendorReceipt, CargoHomeDigest: result.CargoHomeDigest, ConfigSHA256: result.ConfigSHA256, ArtifactManifestIDs: append([]string(nil), result.ArtifactManifestIDs...), GitObjectReceipts: gitObjectReceipts, GitProjectionReceipts: append([]string(nil), state.oracleReceipts...), VendorPackages: append([]VendorPackage(nil), result.VendorPackages...)}
	graph, err := NewCaptureGraph(lock, manifests, evidence.ArtifactManifestIDs)
	if err != nil {
		return nil, err
	}
	return &Capture{Evidence: evidence, state: &captureState{owner: state, workspace: workspace, manifestPath: manifestPath, result: result, graph: graph}}, nil
}

func (m *Manager) preAdmitGitSourceTrees(ctx context.Context, lock LockFile, values []RawGitOrigin) error {
	service := artifactpolicy.NewService()
	for index, raw := range values {
		staged := filepath.Join(m.state.session, "pre-admit", "git-"+strconv.Itoa(index))
		if err := copySourceTree(raw.Repository.Root, staged); err != nil {
			return err
		}
		virtual := "git/raw-" + strconv.Itoa(index)
		probeDescriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileRustV1, Manager: "cargo-1.91.0", PackageName: "git-origin", PackageVersion: "0"}
		probe, probeErr := service.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: probeDescriptor, Root: staged, VirtualRoot: virtual})
		if probeErr != nil && artifactpolicy.ErrorCode(probeErr) != artifactpolicy.CodeOriginUnverified {
			return probeErr
		}
		identity := probe.Manifest.RawPayload.SHA256
		if identity == "" {
			return fail(CodeGitIdentityInvalid, "raw Git tree identity is unavailable", nil)
		}
		descriptor := probeDescriptor
		descriptor.Origin = artifactpolicy.OriginEvidence{Locator: raw.DeclaredURL, ImmutableID: raw.LockedCommit, LockRecord: lock.Digest, ChecksumSHA256: identity, Verified: true}
		admitted, err := service.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: staged, VirtualRoot: virtual})
		if err != nil {
			return err
		}
		tree, err := m.state.intake.CaptureTree(raw.DeclaredURL+"#"+raw.LockedCommit, staged)
		if err != nil {
			return err
		}
		manifestID := closuregraph.ID(admitted.Manifest.ManifestDigest)
		receipt, err := m.state.intake.AdmitTree(tree, raw.DeclaredURL+"#"+raw.LockedCommit, closureexec.AdmissionEvidence{PreviousCausalHead: m.state.causalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
		if err != nil {
			return err
		}
		receiptID, err := receipt.ID()
		if err != nil {
			return err
		}
		if err = copyAdministrativeTree(filepath.Join(raw.Repository.Root, ".git"), filepath.Join(staged, ".git")); err != nil {
			return err
		}
		if err = copyGitMarkers(raw.Repository.Root, staged); err != nil {
			return err
		}
		adminTree, err := m.state.intake.CaptureTree(raw.DeclaredURL+"#"+raw.LockedCommit+"#git-repository", staged)
		if err != nil {
			return err
		}
		adminDigest, err := directoryDigest(staged)
		if err != nil {
			return err
		}
		adminManifestID, err := closuregraph.DomainID("rust-git-object-provenance-v1", map[string]any{"commit": raw.LockedCommit, "digest": adminDigest, "url": raw.DeclaredURL})
		if err != nil {
			return err
		}
		adminReceipt, err := m.state.intake.AdmitTree(adminTree, raw.DeclaredURL+"#"+raw.LockedCommit+"#git-repository", closureexec.AdmissionEvidence{PreviousCausalHead: m.state.causalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: adminManifestID})
		if err != nil {
			return err
		}
		adminReceiptID, err := adminReceipt.ID()
		if err != nil {
			return err
		}
		root, _ := filepath.Abs(raw.Repository.Root)
		m.state.gitInputs[root] = admittedGitInput{input: closureexec.AdmittedInput{Receipt: receipt, Tree: tree}, adminInput: closureexec.AdmittedInput{Receipt: adminReceipt, Tree: adminTree}, receiptID: receiptID, adminReceiptID: adminReceiptID}
	}
	return nil
}

func copySourceTree(source, destination string) error {
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == source {
			return nil
		}
		if entry.Name() == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.Mkdir(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return fail(CodeGitIdentityInvalid, "raw Git source contains non-regular member", map[string]string{"path": filepath.ToSlash(rel)})
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- contained source walk.
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, 0o600)
	})
}

func rawBytes(raw RawFile) ([]byte, error) {
	if !filepath.IsAbs(raw.Path) {
		return nil, fail(CodeConfigUntrusted, "raw file path must be absolute", map[string]string{"path": raw.Path})
	}
	info, err := os.Lstat(raw.Path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return nil, fail(CodeConfigUntrusted, "raw file is not regular", map[string]string{"path": raw.Path})
	}
	return os.ReadFile(raw.Path) // #nosec G304 -- explicit data-only input.
}

func bindRegistry(lock LockFile, values []RawRegistryOrigin) ([]registryOrigin, error) {
	result := []registryOrigin{}
	for _, raw := range values {
		index, err := rawBytes(raw.IndexRecord)
		if err != nil {
			return nil, err
		}
		archive, err := rawBytes(raw.CrateArchive)
		if err != nil {
			return nil, err
		}
		var record struct {
			Name, Version, Checksum string
		}
		var fields map[string]json.RawMessage
		if err = json.Unmarshal(index, &fields); err != nil || json.Unmarshal(fields["name"], &record.Name) != nil || json.Unmarshal(fields["vers"], &record.Version) != nil || json.Unmarshal(fields["cksum"], &record.Checksum) != nil {
			return nil, fail(CodeRegistryIdentityInvalid, "registry index identity is malformed", nil)
		}
		if record.Checksum != digest(archive) {
			return nil, fail(CodeRegistryIdentityInvalid, "registry archive checksum differs from index", nil)
		}
		matched := false
		for _, pkg := range lock.Packages {
			if pkg.Kind == SourceRegistry && pkg.Key.Source == raw.SourceLocator && pkg.Key.Name == record.Name && pkg.Key.Version == record.Version && pkg.Checksum == record.Checksum {
				if matched {
					return nil, fail(CodeRegistryIdentityInvalid, "registry origin maps to multiple lock packages", nil)
				}
				result = append(result, registryOrigin{Package: pkg.Key, IndexRecord: index, Archive: archive, Checksum: pkg.Checksum})
				matched = true
			}
		}
		if !matched {
			return nil, fail(CodeRegistryIdentityInvalid, "registry origin does not match lock/index/archive", nil)
		}
	}
	return result, nil
}

func bindPaths(lock LockFile, workspace string, values []RawPathOrigin) ([]pathOrigin, error) {
	result := []pathOrigin{}
	for _, raw := range values {
		root, err := filepath.Abs(raw.Tree.Root)
		if err != nil || !contained(workspace, root) {
			return nil, fail(CodePathDependencyEscape, "path origin escapes workspace", nil)
		}
		declared := filepath.Clean(filepath.Join(workspace, filepath.FromSlash(raw.DeclaredPath)))
		if filepath.IsAbs(raw.DeclaredPath) || declared != root {
			return nil, fail(CodePathDependencyEscape, "declared path does not identify the supplied tree", map[string]string{"path": raw.DeclaredPath})
		}
		payload, err := os.ReadFile(filepath.Join(root, "Cargo.toml")) // #nosec G304 -- contained root.
		if err != nil {
			return nil, err
		}
		manifest, err := ParseManifest("Cargo.toml", payload)
		if err != nil {
			return nil, err
		}
		found := false
		for _, pkg := range lock.Packages {
			if pkg.Kind == SourcePath && pkg.Key.Name == manifest.PackageName && pkg.Key.Version == manifest.PackageVersion {
				result = append(result, pathOrigin{Package: pkg.Key, Root: root})
				found = true
				break
			}
		}
		if !found {
			return nil, fail(CodeLockMismatch, "path origin has no lock package", map[string]string{"path": raw.DeclaredPath})
		}
	}
	return result, nil
}

type packageDoc struct {
	Package struct {
		Name    string   `toml:"name"`
		Version string   `toml:"version"`
		Include []string `toml:"include"`
	} `toml:"package"`
}

func (m *Manager) bindGit(ctx context.Context, lock LockFile, values []RawGitOrigin) ([]gitOrigin, map[string]gitDerivation, error) {
	origins := []gitOrigin{}
	derivations := map[string]gitDerivation{}
	for _, raw := range values {
		if !validLowerHex(raw.LockedCommit, 40) {
			return nil, nil, fail(CodeGitIdentityInvalid, "Git commit must be full lowercase hex", nil)
		}
		root, _ := filepath.Abs(raw.Repository.Root)
		admitted, ok := m.state.gitInputs[root]
		if !ok {
			return nil, nil, fail(CodeGitIdentityInvalid, "admitted Git source authority is absent", nil)
		}
		repositoryRoot, err := admitted.adminInput.Tree.ProtectedPath()
		if err != nil {
			return nil, nil, err
		}
		repository, err := inspectGitRepository(repositoryRoot, raw.LockedCommit)
		if err != nil {
			return nil, nil, err
		}
		commit, tree, tracked := repository.commit, repository.tree, repository.tracked
		docs, err := packageDocuments(repositoryRoot)
		if err != nil {
			return nil, nil, err
		}
		matchedOrigin := false
		for manifestPath, doc := range docs {
			packageRoot := filepath.Dir(filepath.Join(repositoryRoot, filepath.FromSlash(manifestPath)))
			workspacePackage, _, inheritanceErr := findWorkspaceInheritance(packageRoot)
			if inheritanceErr != nil {
				return nil, nil, inheritanceErr
			}
			if doc.Package.Version == "" {
				doc.Package.Version, _ = workspacePackage["version"].(string)
			}
			for _, pkg := range lock.Packages {
				if pkg.Kind != SourceGit || pkg.Key.Name != doc.Package.Name || pkg.Key.Version != doc.Package.Version || !strings.HasSuffix(pkg.Key.Source, "#"+commit) || !gitDeclarationMatches(pkg.Key.Source, raw.DeclaredURL, raw.Selector) {
					continue
				}
				leaves, inventoryErr := managerInventory(repositoryRoot)
				if inventoryErr != nil {
					return nil, nil, inventoryErr
				}
				packagePath, _ := filepath.Rel(repositoryRoot, packageRoot)
				packagePath = filepath.ToSlash(packagePath)
				if packagePath == "." {
					packagePath = ""
				}
				origin := gitOrigin{Package: pkg.Key, DeclaredURL: raw.DeclaredURL, Selector: raw.Selector, Commit: commit, Tree: tree, Root: packageRoot, AdminRoot: repositoryRoot, PackagePath: packagePath, Include: append([]string(nil), doc.Package.Include...), ManifestTracked: tracked[manifestPath], Leaves: leaves, Submodules: append([]SubmoduleEvidence(nil), repository.submodules...)}
				derived, deriveErr := m.gitProjection(ctx, root, origin, tracked)
				if deriveErr != nil {
					return nil, nil, deriveErr
				}
				stagedRoot := filepath.Join(m.state.session, "captured-git", digest([]byte(pkg.Key.String())))
				if copyErr := copySourceTree(repositoryRoot, stagedRoot); copyErr != nil {
					return nil, nil, copyErr
				}
				origin.Root = stagedRoot
				origins = append(origins, origin)
				derivations[pkg.Key.String()] = derived
				matchedOrigin = true
			}
		}
		if !matchedOrigin {
			return nil, nil, fail(CodeGitIdentityInvalid, "Git declaration has no exact lock package", map[string]string{"url": raw.DeclaredURL})
		}
	}
	sort.Slice(origins, func(i, j int) bool { return origins[i].Package.String() < origins[j].Package.String() })
	return origins, derivations, nil
}

func gitDeclarationMatches(lockSource, declaredURL, selector string) bool {
	value := strings.TrimPrefix(lockSource, "git+")
	if value == lockSource {
		return false
	}
	if index := strings.LastIndexByte(value, '#'); index >= 0 {
		value = value[:index]
	}
	lockedSelector := ""
	if index := strings.IndexByte(value, '?'); index >= 0 {
		lockedSelector = value[index+1:]
		value = value[:index]
	}
	return value == declaredURL && lockedSelector == selector
}

func packageDocuments(root string) (map[string]packageDoc, error) {
	result := map[string]packageDoc{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "target") {
			return filepath.SkipDir
		}
		if entry.IsDir() || entry.Name() != "Cargo.toml" {
			return nil
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- contained walk.
		if err != nil {
			return err
		}
		var raw map[string]any
		if _, err = toml.Decode(string(payload), &raw); err != nil {
			return err
		}
		packageTable, _ := stringAnyMap(raw["package"])
		doc := packageDoc{}
		doc.Package.Name, _ = packageTable["name"].(string)
		doc.Package.Version, _ = packageTable["version"].(string)
		if values, ok := packageTable["include"].([]any); ok {
			for _, value := range values {
				if item, stringOK := value.(string); stringOK {
					doc.Package.Include = append(doc.Package.Include, item)
				}
			}
		}
		if doc.Package.Name != "" {
			rel, _ := filepath.Rel(root, current)
			result[filepath.ToSlash(rel)] = doc
		}
		return nil
	})
	return result, err
}

func managerInventory(root string) ([]OriginLeaf, error) {
	result := []OriginLeaf{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == root {
			return nil
		}
		if entry.Name() == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fail(CodeGitIdentityInvalid, "Git input contains non-regular member", nil)
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- contained walk.
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, current)
		result = append(result, OriginLeaf{Path: filepath.ToSlash(rel), SHA256: digest(payload), Size: int64(len(payload)), Bytes: payload})
		return nil
	})
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, err
}

func (m *Manager) gitProjection(ctx context.Context, repositoryRoot string, origin gitOrigin, tracked map[string]bool) (gitDerivation, error) {
	trackedPaths := make([]string, 0, len(tracked))
	for item, present := range tracked {
		if present && item != "" {
			trackedPaths = append(trackedPaths, item)
		}
	}
	sort.Strings(trackedPaths)
	contextRecord := gitOracleContext{SchemaID: "rust-git-oracle-context-v1", Package: origin.Package, DeclaredURL: origin.DeclaredURL, Selector: origin.Selector, Commit: origin.Commit, Tree: origin.Tree, PackagePath: origin.PackagePath, Include: append([]string(nil), origin.Include...), ManifestTracked: origin.ManifestTracked, TrackedPaths: trackedPaths, Submodules: append([]SubmoduleEvidence(nil), origin.Submodules...)}
	contextBytes, err := protocoljson.MarshalCanonical(map[string]any{"commit": contextRecord.Commit, "declared_url": contextRecord.DeclaredURL, "include": stringsAny(contextRecord.Include), "manifest_tracked": contextRecord.ManifestTracked, "package": map[string]any{"name": origin.Package.Name, "source": origin.Package.Source, "version": origin.Package.Version}, "package_path": contextRecord.PackagePath, "schema_id": contextRecord.SchemaID, "selector": contextRecord.Selector, "submodules": submoduleValues(contextRecord.Submodules), "tracked_paths": stringsAny(contextRecord.TrackedPaths), "tree": contextRecord.Tree})
	if err != nil {
		return gitDerivation{}, err
	}
	service := artifactpolicy.NewService()
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileRustV1, Manager: "cargo-1.91.0", PackageName: origin.Package.Name, PackageVersion: origin.Package.Version, Origin: artifactpolicy.OriginEvidence{Locator: origin.DeclaredURL + "#package-context", ImmutableID: origin.Commit, LockRecord: origin.Package.String(), ChecksumSHA256: "sha256:" + digest(contextBytes), Verified: true}}
	admitted, err := service.AdmitDependency(ctx, artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: "rust-git-oracle-context-v1.json", Size: int64(len(contextBytes)), Reader: bytes.NewReader(contextBytes)}})
	if err != nil {
		return gitDerivation{}, err
	}
	handle, err := m.state.intake.Capture(origin.Package.String()+"#context", int64(len(contextBytes)), bytes.NewReader(contextBytes))
	if err != nil {
		return gitDerivation{}, err
	}
	contextReceipt, err := m.state.intake.Admit(handle, origin.Package.String()+"#context", closureexec.AdmissionEvidence{PreviousCausalHead: m.state.causalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: closuregraph.ID(admitted.Manifest.ManifestDigest)})
	if err != nil {
		return gitDerivation{}, err
	}
	contextReceiptID, err := contextReceipt.ID()
	if err != nil {
		return gitDerivation{}, err
	}
	repositoryRoot, _ = filepath.Abs(repositoryRoot)
	sourceAuthority, ok := m.state.gitInputs[repositoryRoot]
	if !ok {
		return gitDerivation{}, fail(CodeGitIdentityInvalid, "admitted Git source authority is absent", nil)
	}
	type mountedInput struct {
		id    closuregraph.ID
		path  string
		input closureexec.AdmittedInput
	}
	mounted := []mountedInput{{id: sourceAuthority.receiptID, path: "capture/source", input: sourceAuthority.input}, {id: sourceAuthority.adminReceiptID, path: "capture/git-objects", input: sourceAuthority.adminInput}, {id: contextReceiptID, path: "capture/context.json", input: closureexec.AdmittedInput{Receipt: contextReceipt, Handle: handle}}}
	sort.Slice(mounted, func(i, j int) bool { return mounted[i].id < mounted[j].id })
	ids := make([]closuregraph.ID, len(mounted))
	mounts := make([]closureexec.InputMount, len(mounted))
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{}
	readRoots := make([]string, len(mounted))
	for index, item := range mounted {
		ids[index] = item.id
		mounts[index] = closureexec.InputMount{ReceiptID: item.id, Path: item.path}
		inputs[item.id] = item.input
		readRoots[index] = item.path
	}
	sort.Strings(readRoots)
	artifactID, err := closuregraph.DomainID("rust-git-projection-declaration-v1", map[string]any{"context_sha256": "sha256:" + digest(contextBytes), "normalizer_id": NormalizerID, "package": origin.Package.String()})
	if err != nil {
		return gitDerivation{}, err
	}
	requirement := closureexec.EvidenceRequirement{Path: "rust-git-projection-v1.json", SchemaID: gitDerivationSchemaID, ArtifactManifestID: artifactID}
	evidenceSchemaID, err := closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": []any{map[string]any{"artifact_manifest_id": string(artifactID), "path": requirement.Path, "schema_id": requirement.SchemaID}}})
	if err != nil {
		return gitDerivation{}, err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 4 << 20, ReadBytes: 256 << 20, WriteBytes: 4 << 20, WallTimeMillis: 30_000, ProcessCount: 1}
	limitID, err := limits.ID()
	if err != nil {
		return gitDerivation{}, err
	}
	c0ID, _ := closuregraph.DomainID("rust-c0-oracle-v1", map[string]any{"descriptor": "cargo-git-oracle-v1:cargo-0.92.0@ea2d97820c16195b0ca3fadb4319fe512c199a43", "executable_sha256": string(m.state.oracleSHA)})
	toolID, _ := closuregraph.DomainID("rust-oracle-tool-v1", map[string]any{"executable_sha256": string(m.state.oracleSHA)})
	hostID, _ := closuregraph.DomainID("rust-oracle-host-v1", map[string]any{"native": true})
	permitRecord := closureexec.DerivationPermit{SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: m.state.causalHead, InvocationKey: "rust-git-projection-v1:" + origin.Package.String(), InvocationSubtype: closureexec.DerivationManifest, AdmittedInputReceiptIDs: ids, InputMounts: mounts, C0CheckpointID: c0ID, ToolchainNodeID: toolID, ToolchainFingerprint: m.state.oracleSHA, ExecutableSHA256: m.state.oracleSHA, Executable: "bin/curator", CWD: "work", Argv: []string{rustGitOracleWorkerMode}, Environment: map[string]string{"CURATOR_OUTPUT_ROOT": m.state.outputRoot, "RUST_GIT_CONTEXT": filepath.Join(m.state.execRoot, "capture", "context.json"), "RUST_GIT_SOURCE": filepath.Join(m.state.execRoot, "capture", "source"), "LANG": "C", "LC_ALL": "C", "TZ": "UTC"}, HostID: hostID, TargetID: hostID, AllowedProcesses: []string{}, ReadRoots: readRoots, WriteRoots: []string{"rust-git-projection-v1.json"}, ExpectedEvidence: []closureexec.EvidenceRequirement{requirement}, Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceSchemaID}
	permitID, err := m.state.executor.Commit(permitRecord)
	if err != nil {
		return gitDerivation{}, err
	}
	recheck := func(context.Context) (closureexec.ToolchainIdentity, error) {
		payload, readErr := os.ReadFile(m.state.oraclePath) // #nosec G304 -- manager-owned staged executable.
		if readErr != nil {
			return closureexec.ToolchainIdentity{}, readErr
		}
		sum := sha256.Sum256(payload)
		observed := closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
		return closureexec.ToolchainIdentity{Fingerprint: observed, ExecutableSHA256: observed}, nil
	}
	receipt, err := m.state.executor.Execute(ctx, permitID, recheck, inputs)
	if err != nil {
		return gitDerivation{}, err
	}
	payload, err := os.ReadFile(filepath.Join(m.state.outputRoot, "rust-git-projection-v1.json")) // #nosec G304 -- exact declared manager output.
	if err != nil {
		return gitDerivation{}, err
	}
	result, err := bindGitDerivation(m.state.executor, receipt, payload)
	if err == nil {
		m.state.causalHead = string(receipt.NextCausalHead)
		if receiptID, idErr := receipt.ID(); idErr == nil {
			m.state.oracleReceipts = append(m.state.oracleReceipts, string(receiptID))
		}
	}
	_ = os.RemoveAll(filepath.Join(m.state.execRoot, "capture"))
	_ = os.RemoveAll(m.state.outputRoot)
	return result, err
}

func includePath(patterns []string, value string) bool {
	for _, pattern := range patterns {
		if ok, _ := filepath.Match(strings.ReplaceAll(pattern, "**", "*"), value); ok || strings.HasSuffix(pattern, "/**") && strings.HasPrefix(value, strings.TrimSuffix(pattern, "**")) {
			return true
		}
	}
	return false
}

func stringSliceContains(values []string, value string) bool {
	index := sort.SearchStrings(values, value)
	return index < len(values) && values[index] == value
}

// DeriveMetadata derives unfiltered and selection-active Cargo metadata under
// the same causal executor that issued the capture's vendor receipt.
func (m *Manager) DeriveMetadata(ctx context.Context, capture *Capture, selection SelectionContext) (MetadataResult, error) {
	state, err := m.authority()
	if err != nil {
		return MetadataResult{}, err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.closed || capture == nil || capture.state == nil || capture.state.owner != state {
		return MetadataResult{}, fail(CodeGraphIncomplete, "capture is absent, closed, or foreign", nil)
	}
	config, err := os.ReadFile(capture.state.result.ConfigPath) // #nosec G304 -- private capture path.
	if err != nil {
		return MetadataResult{}, err
	}
	runner := &managerMetadataRunner{manager: state, workspace: capture.state.workspace}
	result, err := runPermittedMetadata(ctx, metadataRequest{CaptureID: "sha256:" + digest([]byte(capture.Evidence.LockDigest+"\x00"+capture.Evidence.VendorReceipt)), WorkspaceRoot: capture.state.workspace, ManifestPath: capture.state.manifestPath, CargoHome: state.cargoHome, CargoHomeDigest: capture.state.result.CargoHomeDigest, ConfigPath: capture.state.result.ConfigPath, ConfigBytes: config, Selection: selection, Toolchain: state.cargo, RecheckToolchain: func() (cargoToolchain, error) { return state.recheckCargo(ctx) }, Runner: runner, NormalizeRoots: map[string]string{filepath.Join(state.execRoot, "workspace"): "workspace", filepath.Join(state.execRoot, "vendor"): "vendor"}})
	if err == nil {
		result.owner = state
		result.capture = capture.state
		result.selection = selection
		if result.selection.ResolvedFeatures == nil {
			result.selection.ResolvedFeatures = map[string][]string{}
			for _, node := range result.Active.Resolve {
				result.selection.ResolvedFeatures[node.ID] = append([]string(nil), node.Features...)
			}
		}
	}
	return result, err
}

// Close invalidates manager-owned handles and removes the exact private session.
func (m *Manager) Close() error {
	state, err := m.authority()
	if err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.closed {
		return nil
	}
	state.closed = true
	_ = filepath.WalkDir(state.session, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			_ = os.Chmod(current, 0o700) // #nosec G302 -- cleanup restores owner traversal only within the sealed private session.
		} else {
			_ = os.Chmod(current, 0o600)
		}
		return nil
	})
	return os.RemoveAll(state.session)
}

// Compile-time assertion that the public raw request remains free of behavior.
var _ = reflect.TypeOf(RawCaptureRequest{})
