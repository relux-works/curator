package rustsource

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/protocoljson"
)

// cargoToolchain is the C0-bound physical Cargo implementation admitted by
// cargo-vendor-transform-v1. It is rechecked immediately before each spawn.
type cargoToolchain struct {
	CargoPath, Version, ImplementationCommit, BinarySHA256, Fingerprint, C0CheckpointID string
}

func (tool cargoToolchain) validate() error {
	if !filepath.IsAbs(tool.CargoPath) || tool.Version != "1.91.0" || tool.ImplementationCommit != "ea2d97820c16195b0ca3fadb4319fe512c199a43" || !strings.HasPrefix(tool.BinarySHA256, "sha256:") || !strings.HasPrefix(tool.Fingerprint, "sha256:") || !strings.HasPrefix(tool.C0CheckpointID, "sha256:") {
		return fail(CodeVendorTransformUnsupported, "Cargo toolchain does not match the pinned transform", map[string]string{"version": tool.Version, "commit": tool.ImplementationCommit})
	}
	return nil
}

// vendorInvocation is the exact committed C3a derivation authority.
type vendorInvocation struct {
	TransformID, Executable, CWD, Destination, CargoHome, CargoHomeDigest, ConfigPath, ConfigSHA256 string
	Argv, AdmittedManifestIDs                                                                       []string
	ConfigBytes                                                                                     []byte
	Environment                                                                                     map[string]string
	Toolchain                                                                                       cargoToolchain
	Network                                                                                         string
}

// permit identifies a committed manager-owned derivation authority.
type permit struct{ ID, InvocationID string }

// vendorRunner commits an exact permit before starting Cargo and returns a
// causal receipt. Implementations bridge this seam to closureexec.
type vendorRunner interface {
	CommitVendor(context.Context, vendorInvocation) (permit, error)
	RunVendor(context.Context, permit, vendorInvocation, func() error) (string, error)
}

// captureRequest supplies immutable acquired origins; it contains no ambient
// Cargo cache or unpacked registry source authority.
type captureRequest struct {
	Lock              LockFile
	Registry          []registryOrigin
	Git               []gitOrigin
	Paths             []pathOrigin
	WorkspaceRoot     string
	VendorDestination string
	CargoHome         string
	ConfigPath        string
	StageCargoHome    func(context.Context, string, []string) error
	Toolchain         cargoToolchain
	RecheckToolchain  func() (cargoToolchain, error)
	Runner            vendorRunner
	gitDerivations    map[string]gitDerivation
}

// captureResult is C3b evidence consumed by selection closure.
type captureResult struct {
	LockDigest, TransformID, VendorReceipt, CargoHomeDigest, ConfigPath, ConfigSHA256 string
	ArtifactManifestIDs                                                               []string
	VendorPackages                                                                    []VendorPackage
}

// captureAndVendor admits every origin before committing or starting Cargo,
// runs the one exact vendor transform into an absent destination, and admits
// and verifies every resulting package before returning evidence.
func captureAndVendor(ctx context.Context, request captureRequest) (captureResult, error) {
	if err := request.Toolchain.validate(); err != nil {
		return captureResult{}, err
	}
	if request.Runner == nil || request.RecheckToolchain == nil {
		return captureResult{}, fail(CodeVendorTransformUnsupported, "vendor runner or toolchain recheck is missing", nil)
	}
	if err := rejectAmbientCargoConfig(request.WorkspaceRoot); err != nil {
		return captureResult{}, err
	}
	if request.StageCargoHome == nil || !filepath.IsAbs(request.CargoHome) || !filepath.IsAbs(request.ConfigPath) {
		return captureResult{}, fail(CodeConfigUntrusted, "private Cargo home/config staging is missing", nil)
	}
	if filepath.Clean(request.ConfigPath) != filepath.Join(filepath.Clean(request.CargoHome), "config.toml") {
		return captureResult{}, fail(CodeConfigUntrusted, "source replacement config must be private CARGO_HOME/config.toml", map[string]string{"path": request.ConfigPath})
	}
	if _, err := os.Lstat(request.CargoHome); !os.IsNotExist(err) {
		return captureResult{}, fail(CodeConfigUntrusted, "private Cargo home must be absent before admission", nil)
	}
	if _, err := os.Lstat(request.ConfigPath); !os.IsNotExist(err) {
		return captureResult{}, fail(CodeConfigUntrusted, "source replacement config must be absent before admission", nil)
	}
	if _, err := os.Lstat(request.VendorDestination); !os.IsNotExist(err) {
		return captureResult{}, fail(CodeVendorIncomplete, "vendor destination must be absent", map[string]string{"path": request.VendorDestination})
	}
	remote := map[string]LockPackage{}
	local := map[string]LockPackage{}
	for _, pkg := range request.Lock.Packages {
		if pkg.Kind != SourcePath {
			remote[pkg.Key.String()] = pkg
		} else {
			local[pkg.Key.String()] = pkg
		}
	}
	service := artifactpolicy.NewService()
	manifests := []string{}
	expected := []VendorPackage{}
	mapped := map[string]bool{}
	admitBytes := func(key PackageKey, locator, immutable, lockRecord, virtual string, payload []byte) (string, error) {
		descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileRustV1, Manager: "cargo-1.91.0", PackageName: key.Name, PackageVersion: key.Version, Origin: artifactpolicy.OriginEvidence{Locator: locator, ImmutableID: immutable, LockRecord: lockRecord, ChecksumSHA256: "sha256:" + digest(payload), Verified: true}}
		result, err := service.AdmitDependency(ctx, artifactpolicy.DependencyRequest{Descriptor: descriptor, Payload: artifactpolicy.Payload{Path: virtual, Size: int64(len(payload)), Reader: bytes.NewReader(payload)}})
		if err != nil {
			return "", err
		}
		return result.Manifest.ManifestDigest, nil
	}
	admitDir := func(key PackageKey, locator, immutable, lockRecord, root, virtual string) (string, error) {
		probeDescriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileRustV1, Manager: "cargo-1.91.0", PackageName: key.Name, PackageVersion: key.Version}
		probe, probeErr := service.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: probeDescriptor, Root: root, VirtualRoot: virtual})
		if probeErr != nil && artifactpolicy.ErrorCode(probeErr) != artifactpolicy.CodeOriginUnverified {
			return "", probeErr
		}
		identity := probe.Manifest.RawPayload.SHA256
		if identity == "" {
			return "", fail(CodeGraphIncomplete, "source tree identity could not be captured", map[string]string{"path": virtual})
		}
		descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileRustV1, Manager: "cargo-1.91.0", PackageName: key.Name, PackageVersion: key.Version, Origin: artifactpolicy.OriginEvidence{Locator: locator, ImmutableID: immutable, LockRecord: lockRecord, ChecksumSHA256: identity, Verified: true}}
		result, err := service.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: virtual})
		if err != nil {
			return "", err
		}
		return result.Manifest.ManifestDigest, nil
	}
	for _, origin := range request.Registry {
		pkg, ok := remote[origin.Package.String()]
		if !ok || pkg.Kind != SourceRegistry || pkg.Checksum != origin.Checksum {
			return captureResult{}, fail(CodeRegistryIdentityInvalid, "registry origin has no exact lock package", map[string]string{"package": origin.Package.String()})
		}
		if mapped[origin.Package.String()] {
			return captureResult{}, fail(CodeVendorIncomplete, "registry origin is duplicated", map[string]string{"package": origin.Package.String()})
		}
		id, err := admitBytes(origin.Package, origin.Package.Source, origin.Checksum, request.Lock.Digest, packageDirectory(origin.Package)+".crate", origin.Archive)
		if err != nil {
			return captureResult{}, err
		}
		manifests = append(manifests, id)
		indexID, err := admitBytes(origin.Package, origin.Package.Source+"#index", digest(origin.IndexRecord), request.Lock.Digest, "registry-index/"+packageDirectory(origin.Package)+".json", origin.IndexRecord)
		if err != nil {
			return captureResult{}, err
		}
		manifests = append(manifests, indexID)
		mapped[origin.Package.String()] = true
	}
	for _, origin := range request.Git {
		pkg, ok := remote[origin.Package.String()]
		if !ok || pkg.Kind != SourceGit {
			return captureResult{}, fail(CodeGitIdentityInvalid, "Git origin has no exact lock package", map[string]string{"package": origin.Package.String()})
		}
		if mapped[origin.Package.String()] {
			return captureResult{}, fail(CodeVendorIncomplete, "Git origin is duplicated", map[string]string{"package": origin.Package.String()})
		}
		id, err := admitDir(origin.Package, origin.DeclaredURL, origin.Commit, request.Lock.Digest, origin.Root, "git/"+packageDirectory(origin.Package))
		if err != nil {
			return captureResult{}, err
		}
		manifests = append(manifests, id)
		mapped[origin.Package.String()] = true
	}
	for _, origin := range request.Paths {
		if _, ok := local[origin.Package.String()]; !ok {
			return captureResult{}, fail(CodeLockMismatch, "path origin has no exact lock package", map[string]string{"package": origin.Package.String()})
		}
		if mapped[origin.Package.String()] {
			return captureResult{}, fail(CodeLockMismatch, "package origin is duplicated", map[string]string{"package": origin.Package.String()})
		}
		if !contained(request.WorkspaceRoot, origin.Root) {
			return captureResult{}, fail(CodePathDependencyEscape, "path package escapes frozen workspace", map[string]string{"path": origin.Root})
		}
		id, err := admitDir(origin.Package, "path:"+origin.Root, digest([]byte(origin.Root)), request.Lock.Digest, origin.Root, "path/"+packageDirectory(origin.Package))
		if err != nil {
			return captureResult{}, err
		}
		manifests = append(manifests, id)
		mapped[origin.Package.String()] = true
	}
	for key := range remote {
		if !mapped[key] {
			return captureResult{}, fail(CodeVendorIncomplete, "remote lock package has no admitted origin", map[string]string{"package": key})
		}
	}
	for key := range local {
		if !mapped[key] {
			return captureResult{}, fail(CodeGraphIncomplete, "path lock package has no admitted snapshot", map[string]string{"package": key})
		}
	}
	// Transform parsing begins only after every registry archive/index, complete
	// Git tree, root, workspace, and path origin has passed shared admission.
	for _, origin := range request.Registry {
		transformed, err := deriveRegistryTransform(origin)
		if err != nil {
			return captureResult{}, err
		}
		expected = append(expected, transformed)
	}
	for _, origin := range request.Git {
		if err := validateGitRoot(origin); err != nil {
			return captureResult{}, err
		}
		derivation, ok := request.gitDerivations[origin.Package.String()]
		if !ok {
			return captureResult{}, fail(CodeGitIdentityInvalid, "manager-owned Git derivation is missing", map[string]string{"package": origin.Package.String()})
		}
		transformed, err := deriveGitTransform(origin, derivation)
		if err != nil {
			return captureResult{}, err
		}
		expected = append(expected, transformed)
	}
	sort.Strings(manifests)
	sort.Slice(expected, func(i, j int) bool { return expected[i].Directory < expected[j].Directory })
	if binder, ok := request.Runner.(interface{ bindExpectedPackages([]VendorPackage) }); ok {
		binder.bindExpectedPackages(expected)
	}
	if err := request.StageCargoHome(ctx, request.CargoHome, append([]string(nil), manifests...)); err != nil {
		return captureResult{}, err
	}
	homeDigest, err := directoryDigest(request.CargoHome)
	if err != nil {
		return captureResult{}, fail(CodeConfigUntrusted, "private Cargo home staging is invalid", map[string]string{"detail": err.Error()})
	}
	configBytes, err := DeriveSourceReplacementConfig(request.VendorDestination, request.Lock.Packages)
	if err != nil {
		return captureResult{}, err
	}
	invocation := vendorInvocation{TransformID: TransformID, Executable: request.Toolchain.CargoPath, CWD: request.WorkspaceRoot, Destination: request.VendorDestination, CargoHome: request.CargoHome, CargoHomeDigest: homeDigest, ConfigPath: request.ConfigPath, ConfigBytes: append([]byte(nil), configBytes...), ConfigSHA256: "sha256:" + digest(configBytes), Environment: map[string]string{"CARGO_HOME": request.CargoHome, "CARGO_NET_OFFLINE": "true"}, Argv: []string{"vendor", "--locked", "--offline", "--versioned-dirs", request.VendorDestination}, AdmittedManifestIDs: append([]string(nil), manifests...), Toolchain: request.Toolchain, Network: "none"}
	if err := invocation.validate(); err != nil {
		return captureResult{}, fail(CodeVendorIncomplete, err.Error(), nil)
	}
	permit, err := request.Runner.CommitVendor(ctx, invocation)
	if err != nil {
		return captureResult{}, err
	}
	invocationID, idErr := invocation.ID()
	if idErr != nil {
		return captureResult{}, idErr
	}
	if permit.ID == "" || permit.InvocationID != invocationID {
		return captureResult{}, fail(CodeVendorIncomplete, "vendor permit is missing", nil)
	}
	recheck := func() error {
		if checkErr := recheckCargo(request.Toolchain, request.RecheckToolchain); checkErr != nil {
			return checkErr
		}
		if _, statErr := os.Lstat(request.VendorDestination); !os.IsNotExist(statErr) {
			return fail(CodeVendorIncomplete, "vendor destination appeared before use", nil)
		}
		observedHome, homeErr := directoryDigest(request.CargoHome)
		if homeErr != nil || observedHome != homeDigest {
			return fail(CodeConfigUntrusted, "private Cargo home changed before use", nil)
		}
		if _, configErr := os.Lstat(request.ConfigPath); !os.IsNotExist(configErr) {
			return fail(CodeConfigUntrusted, "source replacement config appeared before use", nil)
		}
		return nil
	}
	if err := recheck(); err != nil {
		return captureResult{}, err
	}
	receipt, err := request.Runner.RunVendor(ctx, permit, invocation, recheck)
	if err != nil {
		return captureResult{}, err
	}
	if receipt == "" {
		return captureResult{}, fail(CodeVendorIncomplete, "vendor receipt is missing", nil)
	}
	observedConfig, err := os.ReadFile(request.ConfigPath) // #nosec G304 -- absolute manager-owned permit path is bound before launch.
	if err != nil || !bytes.Equal(observedConfig, configBytes) {
		return captureResult{}, fail(CodeConfigUntrusted, "source replacement config differs from permit", nil)
	}
	postHomeDigest, err := directoryDigest(request.CargoHome)
	if err != nil {
		return captureResult{}, err
	}
	if err = VerifyVendor(request.VendorDestination, expected); err != nil {
		return captureResult{}, err
	}
	for _, item := range expected {
		id, admitErr := admitDir(item.Package, "vendor:"+item.Directory, TransformID, request.Lock.Digest, filepath.Join(request.VendorDestination, item.Directory), "vendor/"+item.Directory)
		if admitErr != nil {
			return captureResult{}, admitErr
		}
		manifests = append(manifests, id)
	}
	sort.Strings(manifests)
	return captureResult{LockDigest: request.Lock.Digest, TransformID: TransformID, VendorReceipt: receipt, CargoHomeDigest: postHomeDigest, ConfigPath: request.ConfigPath, ConfigSHA256: "sha256:" + digest(configBytes), ArtifactManifestIDs: manifests, VendorPackages: expected}, nil
}

func recheckCargo(expected cargoToolchain, recheck func() (cargoToolchain, error)) error {
	observed, err := recheck()
	if err != nil {
		return err
	}
	if observed != expected {
		return fail(CodeVendorTransformUnsupported, "Cargo toolchain changed before use", nil)
	}
	return nil
}

func validateGitRoot(origin gitOrigin) error {
	observed := map[string]string{}
	err := filepath.WalkDir(origin.Root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == origin.Root || entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(origin.Root, current)
		if err != nil {
			return err
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a contained member under the captured Git root.
		if err != nil {
			return err
		}
		observed[filepath.ToSlash(rel)] = digest(payload)
		return nil
	})
	if err != nil {
		return err
	}
	if len(observed) != len(origin.Leaves) {
		return fail(CodeGitIdentityInvalid, "captured Git tree leaf cardinality differs", nil)
	}
	for _, leaf := range origin.Leaves {
		if observed[leaf.Path] != leaf.SHA256 {
			return fail(CodeGitIdentityInvalid, "captured Git tree differs from object evidence", map[string]string{"path": leaf.Path})
		}
		delete(observed, leaf.Path)
	}
	return nil
}

func contained(root, child string) bool {
	base, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	candidate, err := filepath.Abs(child)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(base, candidate)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func rejectAmbientCargoConfig(workspace string) error {
	current, err := filepath.Abs(workspace)
	if err != nil {
		return err
	}
	for {
		for _, name := range []string{"config", "config.toml"} {
			candidate := filepath.Join(current, ".cargo", name)
			if _, statErr := os.Lstat(candidate); statErr == nil {
				return fail(CodeConfigUntrusted, "package or ancestor Cargo config is forbidden", map[string]string{"path": candidate})
			} else if !os.IsNotExist(statErr) {
				return fail(CodeConfigUntrusted, "Cargo config presence cannot be determined", map[string]string{"path": candidate})
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return nil
}

func (invocation vendorInvocation) validate() error {
	if invocation.TransformID != TransformID || invocation.Network != "none" || len(invocation.Argv) != 5 || invocation.Argv[0] != "vendor" || invocation.Argv[1] != "--locked" || invocation.Argv[2] != "--offline" || invocation.Argv[3] != "--versioned-dirs" || invocation.Argv[4] != invocation.Destination || invocation.Environment["CARGO_HOME"] != invocation.CargoHome || invocation.Environment["CARGO_NET_OFFLINE"] != "true" || invocation.ConfigSHA256 != "sha256:"+digest(invocation.ConfigBytes) || invocation.CargoHomeDigest == "" {
		return fmt.Errorf("widened vendor invocation")
	}
	return nil
}

// ID returns the exact portable vendor authority identity.
func (invocation vendorInvocation) ID() (string, error) {
	if err := invocation.validate(); err != nil {
		return "", err
	}
	encoded, err := protocoljson.MarshalCanonical(map[string]any{"admitted_manifest_ids": stringsAny(invocation.AdmittedManifestIDs), "argv": stringsAny(invocation.Argv), "cargo_home": invocation.CargoHome, "cargo_home_digest": invocation.CargoHomeDigest, "config_path": invocation.ConfigPath, "config_sha256": invocation.ConfigSHA256, "cwd": invocation.CWD, "destination": invocation.Destination, "environment": invocation.Environment, "executable": invocation.Executable, "network": invocation.Network, "toolchain": toolchainValue(invocation.Toolchain), "transform_id": invocation.TransformID})
	if err != nil {
		return "", err
	}
	return "sha256:" + digest(append([]byte("rust-vendor-invocation-v1\x00"), encoded...)), nil
}

func toolchainValue(tool cargoToolchain) map[string]any {
	return map[string]any{"binary_sha256": tool.BinarySHA256, "c0_checkpoint_id": tool.C0CheckpointID, "cargo_path": tool.CargoPath, "fingerprint": tool.Fingerprint, "implementation_commit": tool.ImplementationCommit, "version": tool.Version}
}

func directoryDigest(root string) (string, error) {
	records := []any{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == root {
			return nil
		}
		rel, err := filepath.Rel(root, current)
		if err != nil {
			return err
		}
		record := map[string]any{"path": filepath.ToSlash(rel), "kind": "directory", "size": int64(0), "sha256": ""}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeConfigUntrusted, "Cargo home contains a link", map[string]string{"path": filepath.ToSlash(rel)})
		}
		if !entry.IsDir() {
			if !entry.Type().IsRegular() {
				return fail(CodeConfigUntrusted, "Cargo home contains a special node", map[string]string{"path": filepath.ToSlash(rel)})
			}
			payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member contained by the private Cargo-home root.
			if err != nil {
				return err
			}
			record["kind"] = "regular_file"
			record["size"] = int64(len(payload))
			record["sha256"] = "sha256:" + digest(payload)
		}
		records = append(records, record)
		return nil
	})
	if err != nil {
		return "", err
	}
	encoded, err := protocoljson.MarshalCanonical(records)
	if err != nil {
		return "", err
	}
	return "sha256:" + digest(append([]byte("rust-private-cargo-home-v1\x00"), encoded...)), nil
}
