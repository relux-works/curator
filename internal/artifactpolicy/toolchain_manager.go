package artifactpolicy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"go/build"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

const dependencyBoundaryAlgorithm = "artifact-admission-boundary-v1"

// ToolchainSelectorID is a closed central-manager selection policy. It names
// policy, never a caller-provided filesystem root or executable.
type ToolchainSelectorID string

const (
	// ToolchainSelectorRuntimeGoV1 selects the Go toolchain that built the
	// running Curator process. The root is captured during package
	// initialization, before an adapter or caller can mutate go/build.Default.
	ToolchainSelectorRuntimeGoV1 ToolchainSelectorID = "curator-runtime-go-toolchain-v1"
)

var centrallySelectedGoRoot = build.Default.GOROOT

type selectedToolchainSpec struct {
	selector               ToolchainSelectorID
	root                   string
	executableRelativePath string
	version                string
	platform               string
	rootInfo               fs.FileInfo
}

type selectedToolchainState struct {
	seal                     *authorizationSeal
	spec                     selectedToolchainSpec
	checkpointFingerprint    string
	dependencyBoundaryDigest string
	dependencyCount          int

	mu              sync.Mutex
	admissionDigest string
}

// SelectedToolchain is an opaque manager-owned selection checkpoint. Its zero
// value and caller copies with a nil private state cannot authorize anything.
type SelectedToolchain struct {
	state *selectedToolchainState
}

// SelectExternalToolchain resolves a closed central selector, binds the exact
// admitted dependency boundary, and fingerprints the complete selected root.
// No public argument can supply a root, executable, fingerprint, seal, or
// trust assertion.
func (service *Service) SelectExternalToolchain(
	ctx context.Context,
	selector ToolchainSelectorID,
	dependencies []*Admission,
) (*SelectedToolchain, error) {
	configured, err := configuredService(service)
	if err != nil {
		return nil, err
	}
	boundaryDigest, dependencyCount, err := admittedDependencyBoundary(dependencies, configured.authorizationSeal)
	if err != nil {
		return nil, err
	}
	spec, err := resolveCentralToolchain(selector)
	if err != nil {
		return nil, err
	}
	fingerprint, err := fingerprintSelectedToolchain(ctx, spec)
	if err != nil {
		return nil, fmt.Errorf("select external toolchain: %w", err)
	}
	return &SelectedToolchain{state: &selectedToolchainState{
		seal: configured.authorizationSeal, spec: spec,
		checkpointFingerprint:    fingerprint,
		dependencyBoundaryDigest: boundaryDigest,
		dependencyCount:          dependencyCount,
	}}, nil
}

// AdmitSelectedToolchain rechecks the complete selected root around a bounded
// read of the selected executable, constructs the opaque authorization inside
// the manager boundary, and submits those exact bytes to artifact admission.
func (service *Service) AdmitSelectedToolchain(
	ctx context.Context,
	selection *SelectedToolchain,
	dependencies []*Admission,
	descriptor Descriptor,
) (Result, error) {
	configured, err := configuredService(service)
	if err != nil {
		return Result{}, err
	}
	state, err := validateSelectedToolchain(selection, configured.authorizationSeal, dependencies)
	if err != nil {
		return Result{}, err
	}
	payload, payloadDigest, payloadSize, err := readSelectedExecutable(ctx, state.spec)
	if err != nil {
		return Result{}, fmt.Errorf("read selected toolchain executable: %w", err)
	}
	timeOfUseFingerprint, err := fingerprintSelectedToolchain(ctx, state.spec)
	if err != nil {
		return Result{}, fmt.Errorf("recheck selected toolchain: %w", err)
	}
	if timeOfUseFingerprint != state.checkpointFingerprint {
		return Result{}, fmt.Errorf("%s: selected toolchain changed before admission", CodeToolchainIdentityChanged)
	}
	authorization := sealedToolchainAuthorization{record: toolchainAuthorizationRecord{
		seal:                        configured.authorizationSeal,
		policySelector:              string(state.spec.selector),
		resolvedRoot:                state.spec.root,
		executableRelativePath:      state.spec.executableRelativePath,
		environmentSearchResolution: filepath.Join(state.spec.root, filepath.FromSlash(state.spec.executableRelativePath)),
		version:                     state.spec.version,
		platform:                    state.spec.platform,
		fingerprintAlgorithm:        toolchainFingerprintAlgorithm,
		checkpointFingerprintSHA256: state.checkpointFingerprint,
		timeOfUseFingerprintSHA256:  timeOfUseFingerprint,
		payloadPath:                 state.spec.executableRelativePath,
		payloadSHA256:               payloadDigest,
		payloadSize:                 payloadSize,
		outsideDependencyClosure:    true,
		containedLinksValidated:     true,
		ordinaryNodesValidated:      true,
	}}
	result, admitErr := configured.AdmitToolchain(ctx, ToolchainRequest{
		Descriptor: descriptor,
		Payload: Payload{
			Path: state.spec.executableRelativePath, Size: payloadSize,
			Reader: bytes.NewReader(payload),
		},
		Authorization: authorization,
	})
	if result.Admission != nil {
		result.Admission.toolchain = state
		state.mu.Lock()
		state.admissionDigest = result.Admission.ManifestDigest()
		state.mu.Unlock()
	}
	return result, admitErr
}

// AuthorizeSelectedAdapterExecution performs the immediate complete-tree
// time-of-use recheck, rebinds the exact dependency admissions, validates the
// toolchain admission, and only then returns the centrally selected executable
// path to the manager.
func (service *Service) AuthorizeSelectedAdapterExecution(
	ctx context.Context,
	selection *SelectedToolchain,
	dependencies []*Admission,
	toolchain *Admission,
) (string, error) {
	configured, err := configuredService(service)
	if err != nil {
		return "", err
	}
	state, err := validateSelectedToolchain(selection, configured.authorizationSeal, dependencies)
	if err != nil {
		return "", err
	}
	return authorizeSelectedToolchainExecution(ctx, state, dependencies, toolchain)
}

func authorizeSelectedToolchainExecution(
	ctx context.Context,
	state *selectedToolchainState,
	dependencies []*Admission,
	toolchain *Admission,
) (string, error) {
	if state == nil || toolchain == nil || toolchain.toolchain != state {
		return "", fmt.Errorf("external toolchain selection is absent or does not match admission")
	}
	boundaryDigest, dependencyCount, err := admittedDependencyBoundary(dependencies, state.seal)
	if err != nil {
		return "", err
	}
	if boundaryDigest != state.dependencyBoundaryDigest || dependencyCount != state.dependencyCount {
		return "", fmt.Errorf("captured dependency boundary changed before toolchain use")
	}
	state.mu.Lock()
	admissionDigest := state.admissionDigest
	state.mu.Unlock()
	if admissionDigest == "" || admissionDigest != toolchain.ManifestDigest() {
		return "", fmt.Errorf("external toolchain admission is not bound to the selected checkpoint")
	}
	if err := authorizeAdmissionRoles(dependencies, toolchain); err != nil {
		return "", err
	}
	fingerprint, err := fingerprintSelectedToolchain(ctx, state.spec)
	if err != nil {
		return "", fmt.Errorf("recheck selected toolchain before execution: %w", err)
	}
	if fingerprint != state.checkpointFingerprint {
		return "", fmt.Errorf("%s: selected toolchain changed before execution", CodeToolchainIdentityChanged)
	}
	return filepath.Join(state.spec.root, filepath.FromSlash(state.spec.executableRelativePath)), nil
}

func validateSelectedToolchain(
	selection *SelectedToolchain,
	expectedSeal *authorizationSeal,
	dependencies []*Admission,
) (*selectedToolchainState, error) {
	if selection == nil || selection.state == nil || expectedSeal == nil || selection.state.seal != expectedSeal {
		return nil, fmt.Errorf("external toolchain selection is absent or foreign")
	}
	boundaryDigest, dependencyCount, err := admittedDependencyBoundary(dependencies, expectedSeal)
	if err != nil {
		return nil, err
	}
	if boundaryDigest != selection.state.dependencyBoundaryDigest || dependencyCount != selection.state.dependencyCount {
		return nil, fmt.Errorf("captured dependency boundary does not match toolchain selection")
	}
	return selection.state, nil
}

func admittedDependencyBoundary(dependencies []*Admission, expectedSeal *authorizationSeal) (string, int, error) {
	if len(dependencies) == 0 {
		return "", 0, fmt.Errorf("captured dependency boundary is empty")
	}
	digests := make([]string, len(dependencies))
	for index, dependency := range dependencies {
		if !validAdmission(dependency, RoleDependencyInput, DecisionAdmitInput) || dependency.seal != expectedSeal {
			return "", 0, fmt.Errorf("dependency admission %d is absent or invalid", index)
		}
		digests[index] = dependency.ManifestDigest()
	}
	sort.Strings(digests)
	digest := sha256.New()
	_, _ = io.WriteString(digest, dependencyBoundaryAlgorithm)
	_, _ = digest.Write([]byte{0})
	writeToolchainUint64(digest, uint64(len(digests)))
	for _, value := range digests {
		writeToolchainBytes(digest, []byte(value))
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), len(digests), nil
}

func resolveCentralToolchain(selector ToolchainSelectorID) (selectedToolchainSpec, error) {
	if selector != ToolchainSelectorRuntimeGoV1 {
		return selectedToolchainSpec{}, fmt.Errorf("unsupported central toolchain selector %q", selector)
	}
	if centrallySelectedGoRoot == "" || !filepath.IsAbs(centrallySelectedGoRoot) {
		return selectedToolchainSpec{}, fmt.Errorf("central Go toolchain root is unavailable")
	}
	root, err := filepath.EvalSymlinks(centrallySelectedGoRoot)
	if err != nil {
		return selectedToolchainSpec{}, fmt.Errorf("canonicalize central Go toolchain root: %w", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return selectedToolchainSpec{}, fmt.Errorf("resolve central Go toolchain root: %w", err)
	}
	rootInfo, err := os.Stat(root)
	if err != nil || !rootInfo.IsDir() {
		return selectedToolchainSpec{}, fmt.Errorf("central Go toolchain root is not a stable directory")
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return selectedToolchainSpec{}, fmt.Errorf("open central Go toolchain root: %w", err)
	}
	defer func() { _ = rootHandle.Close() }()
	versionBytes, err := readRootFileBounded(rootHandle, "VERSION", 4096)
	if err != nil {
		return selectedToolchainSpec{}, fmt.Errorf("read central Go toolchain version: %w", err)
	}
	version := strings.TrimSpace(strings.SplitN(string(versionBytes), "\n", 2)[0])
	if version == "" || version != runtime.Version() {
		return selectedToolchainSpec{}, fmt.Errorf("central Go toolchain version %q does not match manager runtime %q", version, runtime.Version())
	}
	executable := "bin/go"
	if runtime.GOOS == "windows" {
		executable = "bin/go.exe"
	}
	if _, _, _, err := readSelectedExecutable(context.Background(), selectedToolchainSpec{
		root: root, executableRelativePath: executable,
	}); err != nil {
		return selectedToolchainSpec{}, fmt.Errorf("validate central Go executable: %w", err)
	}
	return selectedToolchainSpec{
		selector: selector, root: root, executableRelativePath: executable,
		version: version, platform: runtime.GOOS + "/" + runtime.GOARCH,
		rootInfo: rootInfo,
	}, nil
}

type toolchainTreeRecord struct {
	path string
	kind byte
	info fs.FileInfo
	link string
}

func fingerprintSelectedToolchain(ctx context.Context, spec selectedToolchainSpec) (string, error) {
	if err := contextError(ctx); err != nil {
		return "", err
	}
	currentRoot, err := os.Stat(spec.root)
	if err != nil || !currentRoot.IsDir() || spec.rootInfo == nil || !os.SameFile(spec.rootInfo, currentRoot) {
		return "", fmt.Errorf("selected toolchain root identity changed")
	}
	root, err := os.OpenRoot(spec.root)
	if err != nil {
		return "", err
	}
	defer func() { _ = root.Close() }()
	records := make([]toolchainTreeRecord, 0, 4096)
	err = fs.WalkDir(root.FS(), ".", func(pathValue string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := contextError(ctx); err != nil {
			return err
		}
		relative := pathValue
		if relative == "." {
			relative = ""
		} else {
			if _, err := ValidateVirtualPath(filepath.ToSlash(relative)); err != nil {
				return fmt.Errorf("selected toolchain path %q is not portable: %w", relative, err)
			}
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		record := toolchainTreeRecord{path: filepath.ToSlash(relative), info: info}
		switch {
		case info.IsDir():
			record.kind = 'D'
		case info.Mode().IsRegular():
			record.kind = 'F'
		case info.Mode()&fs.ModeSymlink != 0:
			record.kind = 'L'
			target, err := root.Readlink(relative)
			if err != nil || !utf8.ValidString(target) || strings.ContainsRune(target, 0) {
				return fmt.Errorf("selected toolchain link %q is unreadable or invalid", relative)
			}
			if filepath.IsAbs(target) || filepath.VolumeName(target) != "" ||
				strings.HasPrefix(target, "/") || strings.HasPrefix(target, `\`) {
				return fmt.Errorf("selected toolchain link %q is absolute", relative)
			}
			resolved := filepath.Clean(filepath.Join(filepath.Dir(relative), target))
			if resolved == ".." || strings.HasPrefix(resolved, ".."+string(filepath.Separator)) || filepath.IsAbs(resolved) {
				return fmt.Errorf("selected toolchain link %q escapes its root", relative)
			}
			if _, err := root.Stat(relative); err != nil {
				return fmt.Errorf("selected toolchain link %q does not resolve inside its root", relative)
			}
			record.link = target
		default:
			return fmt.Errorf("selected toolchain path %q is a special node", relative)
		}
		records = append(records, record)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Slice(records, func(left, right int) bool { return records[left].path < records[right].path })
	digest := sha256.New()
	_, _ = io.WriteString(digest, toolchainFingerprintAlgorithm)
	_, _ = digest.Write([]byte{0})
	writeToolchainBytes(digest, []byte(spec.version))
	writeToolchainBytes(digest, []byte(spec.platform))
	for _, record := range records {
		if err := contextError(ctx); err != nil {
			return "", err
		}
		_, _ = digest.Write([]byte{record.kind})
		writeToolchainBytes(digest, []byte(record.path))
		writeToolchainUint64(digest, uint64(record.info.Mode()))
		switch record.kind {
		case 'D':
			writeToolchainUint64(digest, 0)
		case 'L':
			writeToolchainBytes(digest, []byte(record.link))
		case 'F':
			name := record.path
			if name == "" {
				return "", fmt.Errorf("selected toolchain root cannot be a regular file")
			}
			file, err := root.Open(name)
			if err != nil {
				return "", err
			}
			opened, statErr := file.Stat()
			if statErr != nil || !opened.Mode().IsRegular() || !os.SameFile(record.info, opened) || opened.Size() < 0 {
				_ = file.Close()
				return "", fmt.Errorf("selected toolchain file %q changed while opening", name)
			}
			writeToolchainBytes(digest, []byte(strconv.FormatInt(opened.Size(), 10)))
			written, copyErr := io.Copy(digest, &contextReader{ctx: ctx, reader: file})
			closeErr := file.Close()
			if copyErr != nil || closeErr != nil || written != opened.Size() {
				return "", fmt.Errorf("selected toolchain file %q changed while hashing", name)
			}
		}
	}
	currentRoot, err = os.Stat(spec.root)
	if err != nil || !os.SameFile(spec.rootInfo, currentRoot) {
		return "", fmt.Errorf("selected toolchain root changed while hashing")
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

func readSelectedExecutable(ctx context.Context, spec selectedToolchainSpec) ([]byte, string, int64, error) {
	root, err := os.OpenRoot(spec.root)
	if err != nil {
		return nil, "", 0, err
	}
	defer func() { _ = root.Close() }()
	file, err := root.Open(filepath.FromSlash(spec.executableRelativePath))
	if err != nil {
		return nil, "", 0, err
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > DefaultLimits().MaxSingleLeafBytes {
		return nil, "", 0, fmt.Errorf("selected executable is not a bounded regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o111 == 0 {
		return nil, "", 0, fmt.Errorf("selected executable has no execute permission")
	}
	payload := make([]byte, info.Size())
	if _, err := io.ReadFull(&contextReader{ctx: ctx, reader: file}, payload); err != nil {
		return nil, "", 0, err
	}
	var sentinel [1]byte
	if read, err := file.Read(sentinel[:]); read != 0 || err != io.EOF {
		return nil, "", 0, fmt.Errorf("selected executable size changed while reading")
	}
	digest := sha256.Sum256(payload)
	return payload, "sha256:" + hex.EncodeToString(digest[:]), int64(len(payload)), nil
}

func readRootFileBounded(root *os.Root, pathValue string, maximum int64) ([]byte, error) {
	file, err := root.Open(pathValue)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	payload, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(payload)) > maximum {
		return nil, fmt.Errorf("file exceeds the central selector bound")
	}
	return payload, nil
}

func writeToolchainBytes(writer io.Writer, payload []byte) {
	writeToolchainUint64(writer, uint64(len(payload)))
	_, _ = writer.Write(payload)
}

func writeToolchainUint64(writer io.Writer, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = writer.Write(encoded[:])
}
