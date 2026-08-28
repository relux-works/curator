package rustsource

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/privatedir"
)

type managerVendorRunner struct {
	manager       *managerState
	workspace     string
	registry      []registryOrigin
	git           []gitOrigin
	expected      []VendorPackage
	expectedBound bool
	permitID      closuregraph.ID
	inputs        map[closuregraph.ID]closureexec.AdmittedInput
}

func (runner *managerVendorRunner) bindExpectedPackages(expected []VendorPackage) {
	runner.expected = append([]VendorPackage{}, expected...)
	runner.expectedBound = true
}

func (runner *managerVendorRunner) stageCargoHome(_ context.Context, root string, _ []string) error {
	if runner == nil || runner.manager == nil || root != runner.manager.cargoHome {
		return fail(CodeConfigUntrusted, "Cargo home staging authority differs", nil)
	}
	if err := privatedir.Make(root); err != nil {
		return err
	}
	for _, origin := range runner.registry {
		sourceID, err := cargoRegistrySourceID(origin.Package.Source)
		if err != nil {
			return err
		}
		cacheRoot := filepath.Join(root, "registry", "cache", sourceID)
		if err = privatedir.MakeAll(cacheRoot); err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(cacheRoot, packageDirectory(origin.Package)+".crate"), origin.Archive, 0o400); err != nil {
			return err
		}
		indexRoot := filepath.Join(root, "registry", "index", sourceID)
		indexPath := filepath.Join(indexRoot, ".cache", filepath.FromSlash(registryIndexPath(origin.Package.Name)))
		if err = privatedir.MakeAll(filepath.Dir(indexPath)); err != nil {
			return err
		}
		cacheRecord := append([]byte{3, 2, 0, 0, 0}, []byte("etag: \"curator-admitted\"\x00")...)
		cacheRecord = append(cacheRecord, []byte(origin.Package.Version)...)
		cacheRecord = append(cacheRecord, 0)
		cacheRecord = append(cacheRecord, origin.IndexRecord...)
		cacheRecord = append(cacheRecord, 0)
		if err = os.WriteFile(indexPath, cacheRecord, 0o400); err != nil {
			return err
		}
		configPath := filepath.Join(indexRoot, "config.json")
		configBytes := []byte(`{"dl":"https://static.crates.io/crates","api":"https://crates.io"}`)
		if existing, readErr := os.ReadFile(configPath); readErr == nil { // #nosec G304 -- configPath is the fixed registry source path below the manager-owned Cargo home.
			if string(existing) != string(configBytes) {
				return fail(CodeRegistryIdentityInvalid, "registry source config differs across admitted packages", map[string]string{"source": origin.Package.Source})
			}
		} else if !os.IsNotExist(readErr) {
			return readErr
		} else if err = os.WriteFile(configPath, configBytes, 0o400); err != nil {
			return err
		}
		pkg, ok := expectedPackage(runner.expected, origin.Package)
		if !ok {
			return fail(CodeVendorIncomplete, "registry staging expectation is absent", nil)
		}
		if err = materializePackage(filepath.Join(root, "registry", "src", sourceID, pkg.Directory), pkg, true); err != nil {
			return err
		}
	}
	for _, origin := range runner.git {
		name := gitCacheName(origin.DeclaredURL)
		db := filepath.Join(root, "git", "db", name)
		if err := copyAdministrativeTree(filepath.Join(origin.AdminRoot, ".git"), db); err != nil {
			return err
		}
		checkout := filepath.Join(root, "git", "checkouts", name, origin.Commit[:7])
		if err := copySourceTree(origin.Root, checkout); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(checkout, ".git"), []byte("gitdir: "+db+"\n"), 0o400); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(checkout, ".cargo-ok"), nil, 0o400); err != nil {
			return err
		}
	}
	return nil
}

func cargoRegistrySourceID(source string) (string, error) {
	value := strings.TrimPrefix(source, "registry+")
	if value == "https://github.com/rust-lang/crates.io-index" {
		return "index.crates.io-1949cf8c6b5b557f", nil
	}
	return "", fail(CodeRegistryIdentityInvalid, "registry cache layout is unsupported", map[string]string{"source": source})
}

func registryIndexPath(name string) string {
	name = strings.ToLower(name)
	switch len(name) {
	case 1:
		return "1/" + name
	case 2:
		return "2/" + name
	case 3:
		return "3/" + name[:1] + "/" + name
	default:
		return name[:2] + "/" + name[2:4] + "/" + name
	}
}

func gitCacheName(url string) string {
	trimmed := strings.TrimSuffix(url, "/")
	base := filepath.Base(strings.TrimPrefix(trimmed, "file://"))
	base = strings.TrimSuffix(base, ".git")
	if base == "" || base == "." || base == string(filepath.Separator) {
		base = "repository"
	}
	return base + "-" + cargoShortHash(url)
}

func expectedPackage(packages []VendorPackage, key PackageKey) (VendorPackage, bool) {
	for _, pkg := range packages {
		if pkg.Package == key {
			return pkg, true
		}
	}
	return VendorPackage{}, false
}

func materializePackage(root string, pkg VendorPackage, cargoOK bool) error {
	if err := privatedir.MakeAll(root); err != nil {
		return err
	}
	for _, leaf := range pkg.Files {
		if leaf.Path == ".cargo-checksum.json" {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(leaf.Path))
		if err := privatedir.MakeAll(filepath.Dir(path)); err != nil {
			return err
		}
		if err := os.WriteFile(path, leaf.Bytes, 0o400); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".cargo-checksum.json"), pkg.ChecksumBytes, 0o400); err != nil {
		return err
	}
	if cargoOK {
		return os.WriteFile(filepath.Join(root, ".cargo-ok"), nil, 0o400)
	}
	return nil
}

func copyAdministrativeTree(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil || !info.IsDir() {
		return fail(CodeGitIdentityInvalid, "Git administrative object store is absent", nil)
	}
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, current)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return privatedir.MakeAll(target)
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeGitIdentityInvalid, "Git administrative store contains a link", nil)
		}
		if !entry.Type().IsRegular() {
			return fail(CodeGitIdentityInvalid, "Git administrative store contains a special node", nil)
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- contained administrative provenance tree.
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, 0o400)
	})
}

func copyGitMarkers(source, destination string) error {
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == source {
			return nil
		}
		if entry.Name() == ".git" && entry.Type().IsRegular() {
			rel, err := filepath.Rel(source, current)
			if err != nil {
				return err
			}
			payload, err := os.ReadFile(current) // #nosec G304 -- contained repository marker.
			if err != nil {
				return err
			}
			target := filepath.Join(destination, rel)
			if err = privatedir.MakeAll(filepath.Dir(target)); err != nil {
				return err
			}
			return os.WriteFile(target, payload, 0o400)
		}
		if entry.Name() == ".git" && entry.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
}

func (runner *managerVendorRunner) CommitVendor(ctx context.Context, invocation vendorInvocation) (permit, error) {
	if runner == nil || runner.manager == nil || !runner.expectedBound {
		return permit{}, fail(CodeVendorIncomplete, "manager vendor authority is incomplete", nil)
	}
	workspaceInput, workspaceID, err := runner.admitTree(ctx, "rust-workspace-v1", runner.workspace)
	if err != nil {
		return permit{}, err
	}
	homeInput, homeID, err := runner.admitTree(ctx, "rust-cargo-home-v1", runner.manager.cargoHome)
	if err != nil {
		return permit{}, err
	}
	runner.inputs = map[closuregraph.ID]closureexec.AdmittedInput{workspaceID: workspaceInput, homeID: homeInput}
	runner.manager.workspaceInput = workspaceInput
	runner.manager.workspaceID = workspaceID
	ids := []closuregraph.ID{workspaceID, homeID}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	mountPath := map[closuregraph.ID]string{workspaceID: "workspace", homeID: "cargo-home-seed"}
	mounts := make([]closureexec.InputMount, len(ids))
	reads := make([]string, len(ids))
	for index, id := range ids {
		mounts[index] = closureexec.InputMount{ReceiptID: id, Path: mountPath[id]}
		reads[index] = mountPath[id]
	}
	sort.Strings(reads)
	requirements, err := vendorEvidenceRequirements(runner.expected)
	if err != nil {
		return permit{}, err
	}
	stdoutEvidence := ""
	if len(requirements) == 0 {
		markerID, idErr := closuregraph.DomainID("rust-empty-vendor-output-v1", map[string]any{"config_sha256": invocation.ConfigSHA256})
		if idErr != nil {
			return permit{}, idErr
		}
		stdoutEvidence = "vendor-empty-config.toml"
		requirements = []closureexec.EvidenceRequirement{{Path: stdoutEvidence, SchemaID: "rust-empty-vendor-v1", ArtifactManifestID: markerID}}
	}
	writes := make([]string, len(requirements))
	for index := range requirements {
		writes[index] = requirements[index].Path
	}
	writes = append(writes, "cargo-home")
	sort.Strings(writes)
	evidenceID, err := evidenceRequirementsID(requirements)
	if err != nil {
		return permit{}, err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 512 << 20, ReadBytes: 1 << 30, WriteBytes: 512 << 20, WallTimeMillis: 120_000, ProcessCount: 1}
	limitID, err := limits.ID()
	if err != nil {
		return permit{}, err
	}
	c0, _ := closuregraph.DomainID("rust-c0-cargo-v1", map[string]any{"fingerprint": invocation.Toolchain.Fingerprint, "transform": TransformID})
	tool, _ := closuregraph.DomainID("rust-cargo-tool-v1", map[string]any{"executable_sha256": invocation.Toolchain.BinarySHA256})
	host, _ := closuregraph.DomainID("rust-native-host-v1", map[string]any{"native": true})
	record := closureexec.DerivationPermit{
		SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: runner.manager.causalHead,
		InvocationKey: "cargo-vendor-transform-v1:" + invocationIDMust(invocation), InvocationSubtype: closureexec.DerivationVendor,
		AdmittedInputReceiptIDs: ids, InputMounts: mounts, WorkCopies: []closureexec.WorkCopy{{ReceiptID: homeID, Path: "cargo-home"}}, C0CheckpointID: c0, ToolchainNodeID: tool,
		ToolchainFingerprint: closuregraph.ID(invocation.Toolchain.Fingerprint), ExecutableSHA256: closuregraph.ID(invocation.Toolchain.BinarySHA256),
		Executable: "bin/cargo", CWD: "workspace", Argv: []string{"vendor", "--locked", "--offline", "--versioned-dirs", filepath.Join(runner.manager.outputRoot, "vendor")},
		Environment: map[string]string{"CARGO_HOME": filepath.Join(runner.manager.execRoot, "cargo-home"), "CARGO_NET_OFFLINE": "true", "CARGO_REGISTRIES_CRATES_IO_PROTOCOL": "sparse", "CURATOR_OUTPUT_ROOT": runner.manager.outputRoot, "HOME": filepath.Join(runner.manager.execRoot, "home"), "LANG": "C", "LC_ALL": "C", "TZ": "UTC"},
		HostID:      host, TargetID: host, AllowedProcesses: []string{}, ReadRoots: reads, WriteRoots: writes, ExpectedEvidence: requirements, StdoutEvidencePath: stdoutEvidence,
		Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID,
	}
	if err = stageCargoExecutable(runner.manager); err != nil {
		return permit{}, err
	}
	runner.permitID, err = runner.manager.executor.Commit(record)
	if err != nil {
		return permit{}, err
	}
	id, err := invocation.ID()
	return permit{ID: string(runner.permitID), InvocationID: id}, err
}

func invocationIDMust(invocation vendorInvocation) string {
	id, _ := invocation.ID()
	return id
}

func (runner *managerVendorRunner) RunVendor(ctx context.Context, legacy permit, invocation vendorInvocation, recheck func() error) (string, error) {
	if legacy.ID != string(runner.permitID) || legacy.ID == "" {
		return "", fail(CodeVendorIncomplete, "vendor permit is absent or foreign", nil)
	}
	if err := recheck(); err != nil {
		return "", err
	}
	receipt, err := runner.manager.executor.Execute(ctx, runner.permitID, func(ctx context.Context) (closureexec.ToolchainIdentity, error) {
		tool, checkErr := runner.manager.recheckCargo(ctx)
		if checkErr != nil {
			return closureexec.ToolchainIdentity{}, checkErr
		}
		return closureexec.ToolchainIdentity{Fingerprint: closuregraph.ID(tool.Fingerprint), ExecutableSHA256: closuregraph.ID(tool.BinarySHA256)}, nil
	}, runner.inputs)
	if err != nil {
		return "", err
	}
	if err = runner.manager.executor.VerifyIssuedDerivationReceipt(receipt); err != nil {
		return "", err
	}
	runner.manager.causalHead = string(receipt.NextCausalHead)
	receiptID, err := receipt.ID()
	if err != nil {
		return "", err
	}
	if len(runner.expected) == 0 {
		if err = privatedir.Make(invocation.Destination); err != nil && !os.IsExist(err) {
			return "", err
		}
	}
	if err = os.WriteFile(invocation.ConfigPath, invocation.ConfigBytes, 0o600); err != nil {
		return "", err
	}
	return string(receiptID), nil
}

func (runner *managerVendorRunner) admitTree(ctx context.Context, origin, root string) (closureexec.AdmittedInput, closuregraph.ID, error) {
	manifestID := closuregraph.ID("")
	if origin == "rust-cargo-home-v1" {
		directoryID, digestErr := directoryDigest(root)
		if digestErr != nil {
			return closureexec.AdmittedInput{}, "", digestErr
		}
		manifestID, digestErr = closuregraph.DomainID("rust-derived-cargo-home-v1", map[string]any{"directory_sha256": directoryID})
		if digestErr != nil {
			return closureexec.AdmittedInput{}, "", digestErr
		}
	} else {
		service := artifactpolicy.NewService()
		probeDescriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileRustV1, Manager: "cargo-1.91.0", PackageName: origin, PackageVersion: "1"}
		probe, probeErr := service.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: probeDescriptor, Root: root, VirtualRoot: origin})
		if probeErr != nil && artifactpolicy.ErrorCode(probeErr) != artifactpolicy.CodeOriginUnverified {
			return closureexec.AdmittedInput{}, "", probeErr
		}
		descriptor := probeDescriptor
		descriptor.Origin = artifactpolicy.OriginEvidence{Locator: origin, ImmutableID: probe.Manifest.RawPayload.SHA256, LockRecord: origin, ChecksumSHA256: probe.Manifest.RawPayload.SHA256, Verified: true}
		admitted, admitErr := service.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: origin})
		if admitErr != nil {
			return closureexec.AdmittedInput{}, "", admitErr
		}
		manifestID = closuregraph.ID(admitted.Manifest.ManifestDigest)
	}
	tree, err := runner.manager.intake.CaptureTree(origin, root)
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	receipt, err := runner.manager.intake.AdmitTree(tree, origin, closureexec.AdmissionEvidence{PreviousCausalHead: runner.manager.causalHead, ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: ProfileID, DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	id, err := receipt.ID()
	return closureexec.AdmittedInput{Receipt: receipt, Tree: tree}, id, err
}

func stageCargoExecutable(state *managerState) error {
	payload, err := os.ReadFile(state.cargoRegistry.executable) // #nosec G304 -- manager-sealed C0 Cargo registration.
	if err != nil {
		return err
	}
	path := filepath.Join(state.execRoot, "bin", "cargo")
	if chmodErr := os.Chmod(path, 0o700); chmodErr != nil && !os.IsNotExist(chmodErr) { // #nosec G302 -- private staged executable must be made owner-writable before exact-byte refresh.
		return chmodErr
	}
	if err = os.WriteFile(path, payload, 0o500); err != nil {
		return err
	}
	return os.Chmod(path, 0o500) // #nosec G302 -- manager-staged trusted executable must retain owner execute permission.
}

func vendorEvidenceRequirements(packages []VendorPackage) ([]closureexec.EvidenceRequirement, error) {
	requirements := []closureexec.EvidenceRequirement{}
	for _, pkg := range packages {
		for _, leaf := range pkg.Files {
			if leaf.Path == ".cargo-checksum.json" {
				continue
			}
			path := filepath.ToSlash(filepath.Join("vendor", pkg.Directory, leaf.Path))
			id, err := closuregraph.DomainID("rust-vendor-output-v1", map[string]any{"path": path, "sha256": "sha256:" + leaf.SHA256, "size": leaf.Size})
			if err != nil {
				return nil, err
			}
			requirements = append(requirements, closureexec.EvidenceRequirement{Path: path, SchemaID: "rust-vendor-leaf-v1", ArtifactManifestID: id})
		}
		path := filepath.ToSlash(filepath.Join("vendor", pkg.Directory, ".cargo-checksum.json"))
		id, err := closuregraph.DomainID("rust-vendor-output-v1", map[string]any{"path": path, "sha256": "sha256:" + digest(pkg.ChecksumBytes), "size": len(pkg.ChecksumBytes)})
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, closureexec.EvidenceRequirement{Path: path, SchemaID: "rust-vendor-checksum-v1", ArtifactManifestID: id})
	}
	sort.Slice(requirements, func(i, j int) bool { return requirements[i].Path < requirements[j].Path })
	return requirements, nil
}

func evidenceRequirementsID(requirements []closureexec.EvidenceRequirement) (closuregraph.ID, error) {
	values := make([]any, len(requirements))
	for index, requirement := range requirements {
		values[index] = map[string]any{"artifact_manifest_id": string(requirement.ArtifactManifestID), "path": requirement.Path, "schema_id": requirement.SchemaID}
	}
	return closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": values})
}
