package godriver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/buildsource"
)

const (
	defaultBuildTimeout  = 2 * time.Minute
	defaultBuildOutput   = int64(8 * 1024 * 1024)
	defaultArtifactLimit = int64(128 * 1024 * 1024)
	defaultFileLimit     = int64(512 * 1024 * 1024)
	defaultDiskLimit     = int64(1024 * 1024 * 1024)
	defaultMemoryLimit   = int64(2 * 1024 * 1024 * 1024)
	defaultProcessLimit  = 64
)

var listArguments = []string{"list", "-mod=vendor", "-deps", "-json", "-buildvcs=false", "-compiler=gc", "-pgo=off", "."}

var buildArgumentPrefix = []string{
	"build", "-mod=vendor", "-trimpath", "-buildvcs=false", "-buildmode=exe", "-compiler=gc", "-pgo=off",
	"-ldflags=-linkmode=internal -libgcc=none", "-o",
}

// ResourceLimits is the bounded execution request applied to one operation.
// Timeout, OutputBytes, and ArtifactBytes are parent-enforced over the whole
// worker domain. FileBytes, MemoryBytes, and Processes are applied only where
// rc5-native-control-inventory-v1 marks the corresponding control available;
// the portable profile claims no hard aggregate descendant bound.
type ResourceLimits struct {
	Timeout       time.Duration
	OutputBytes   int64
	ArtifactBytes int64
	FileBytes     int64
	DiskBytes     int64
	MemoryBytes   int64
	Processes     int
}

// BuildCommand is the exact package-controlled build-command surface. Protocol
// Core section 4.2 admits exactly type, driver, and source_dir; anything else
// is an attempt to influence the execution boundary.
type BuildCommand map[string]any

// BuildRequest identifies one already-validated command in a frozen source
// snapshot. BuildRoot and SourceDir use protocol (slash-separated) paths.
type BuildRequest struct {
	Session *Session
	Source  *buildsource.Token
	// Command surface exactly as the package declared it.
	CommandObject BuildCommand
	BuildRoot     string
	SourceDir     string
	Command       string
	Limits        ResourceLimits
}

// Artifact is a verified private output. StagedPath is manager-private and is
// returned only so the cache publisher can copy it; godriver never starts it.
type Artifact struct {
	StagedPath string
	Metadata   buildmeta.Artifact
}

// Result is the complete outcome of one portable go-v1 build operation. The
// capability-evidence record is result-only and never enters a cache key,
// receipt input, install marker, or conformance claim.
type Result struct {
	Artifact Artifact
	Evidence CapabilityEvidence
}

// packageInfluenceSurfaces names the recognized package attempts to reach the
// execution boundary, so the diagnostic identifies the surface rather than the
// spelling.
var packageInfluenceSurfaces = map[string]string{
	"executable": "worker, Go launcher, or GOROOT tool program",
	"program":    "worker, Go launcher, or GOROOT tool program",
	"argv":       "Go list or build argument vector",
	"args":       "Go list or build argument vector",
	"arguments":  "Go list or build argument vector",
	"env":        "worker or compiler environment value",
	"environ":    "worker or compiler environment value",
	"output":     "staging, artifact, shim, or install destination",
	"out":        "staging, artifact, shim, or install destination",
	"flags":      "compiler, linker, tag, or toolchain flag",
	"tags":       "compiler, linker, tag, or toolchain flag",
	"ldflags":    "compiler, linker, tag, or toolchain flag",
	"gcflags":    "compiler, linker, tag, or toolchain flag",
	"toolchain":  "compiler, linker, tag, or toolchain flag",
	"hooks":      "pre-build, post-build, or lifecycle hook",
	"pre_build":  "pre-build, post-build, or lifecycle hook",
	"post_build": "pre-build, post-build, or lifecycle hook",
	"plugins":    "compiler, linker, or manager plugin",
	"plugin":     "compiler, linker, or manager plugin",
	"generate":   "source generator, macro, or code-producing step",
	"generators": "source generator, macro, or code-producing step",
	"scripts":    "source generator, macro, or code-producing step",
}

// Build runs one portable manager-worker-v1 operation: fixed package-independent
// checks, a per-operation native-control probe, an identity-verified worker,
// exactly one fixed go list, complete package-graph validation, exactly one
// authenticated build permit, exactly one fixed go build, artifact verification,
// post-exec identity re-verification, and worker-domain teardown. The built
// output is never started.
func Build(ctx context.Context, request BuildRequest) (_ Result, resultErr error) {
	if request.Session == nil || request.Source == nil {
		return Result{}, diagnostic("invalid_build_request", "trusted session and frozen source are required")
	}
	if err := validatePackageCommandSurface(request); err != nil {
		return Result{}, err
	}
	limits, err := normalizeBuildLimits(request.Limits)
	if err != nil {
		return Result{}, err
	}
	input := buildmeta.Input{
		SchemaVersion: buildmeta.SchemaVersion,
		Driver:        buildmeta.DriverGoV1,
		BuildSource:   request.Source.Identity(),
		BuildRoot:     request.BuildRoot,
		Command:       request.Command,
		SourceDir:     request.SourceDir,
		Target:        request.Session.Target(),
		Toolchain:     request.Session.Toolchain(),
		Policy:        buildmeta.FixedPolicy(),
	}
	if err := input.Validate(); err != nil {
		return Result{}, diagnosticErr("invalid_build_request", err, "invalid go-v1 build input")
	}
	if err := request.Source.Recheck(); err != nil {
		return Result{}, diagnosticErr("source_mutated", err, "frozen source changed before preflight")
	}
	if err := request.Session.VerifyToolchain(ctx); err != nil {
		return Result{}, err
	}

	sourceRoot, buildRoot, sourceDir, err := canonicalBuildDirectories(request.Source.Path(), request.BuildRoot, request.SourceDir)
	if err != nil {
		return Result{}, err
	}

	// Step 2: probe, once for this operation and before the worker exists,
	// which inventory controls this platform provides for exactly these limits.
	// Each probe performs the operation the control will perform, so a control
	// that is not applicable rejects here rather than after the worker starts.
	platform, probes, err := probeNativeControls(limits)
	if err != nil {
		return Result{}, err
	}
	// Step 3: resolve and hash the installed manager executable.
	executable, err := resolveExecutableIdentity(managerExecutable())
	if err != nil {
		return Result{}, err
	}

	stage, err := os.MkdirTemp(request.Session.operation, ".curator-go-build-")
	if err != nil {
		return Result{}, diagnosticErr("private_build_failed", err, "cannot create private build staging")
	}
	defer func() {
		if resultErr != nil {
			if cleanupErr := os.RemoveAll(stage); cleanupErr != nil {
				resultErr = errors.Join(resultErr, diagnosticErr("private_build_cleanup_failed", cleanupErr, "cannot remove failed build staging"))
			}
		}
	}()
	artifactRel, err := buildmeta.ArtifactPath(request.Command, request.Session.Target().GOOS)
	if err != nil {
		return Result{}, diagnosticErr("invalid_build_request", err, "cannot derive artifact path")
	}
	artifactPath := filepath.Join(stage, filepath.FromSlash(artifactRel))
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o700); err != nil {
		return Result{}, diagnosticErr("private_build_failed", err, "cannot create private artifact directory")
	}

	target := request.Session.Target()
	plan := workerPlan{
		Executable:    executable,
		GoExecutable:  request.Session.Executable(),
		GOROOT:        request.Session.GOROOT(),
		ToolDirectory: filepath.Join(request.Session.GOROOT(), "pkg", "tool", target.GOOS+"_"+target.GOARCH),
		Directory:     sourceDir,
		Environment:   request.Session.Environment(),
		ListArgv:      append([]string(nil), listArguments...),
		BuildArgv:     append(append([]string(nil), buildArgumentPrefix...), artifactPath, "."),
		ArtifactPath:  artifactPath,
		ReadOnlyRoots: []string{sourceRoot, request.Session.GOROOT()},
		PrivateRoots:  privateRoots(request.Session.Environment(), stage),
		Platform:      platform,
		Probes:        probes,
		Limits:        limits,
	}

	client, err := launchWorker(ctx, plan)
	if err != nil {
		return Result{}, err
	}
	defer client.teardown()

	bracket := func(output Output, runErr error) (Output, error) {
		if sourceErr := request.Source.Recheck(); sourceErr != nil {
			runErr = errors.Join(runErr, diagnosticErr("source_mutated", sourceErr, "frozen source changed during a Go child"))
		}
		if toolErr := request.Session.VerifyToolchain(ctx); toolErr != nil {
			runErr = errors.Join(runErr, toolErr)
		}
		if stateErr := verifyPrivateState(request.Session.operation, stage, limits.DiskBytes); stateErr != nil {
			runErr = errors.Join(runErr, stateErr)
		}
		return output, runErr
	}

	listOutput, err := bracket(client.list())
	if err != nil {
		return Result{}, classifyListFailure(err, listOutput.Stderr)
	}
	if err := validatePackageGraph(listOutput.Stdout, graphValidation{
		BuildRoot: buildRoot, SourceDir: sourceDir, GOROOT: request.Session.GOROOT(),
	}); err != nil {
		return Result{}, err
	}

	buildOutput, err := bracket(client.build())
	if err != nil {
		return Result{}, classifyBuildFailure(err, buildOutput.Stderr)
	}
	metadata, err := verifyArtifact(stage, artifactPath, artifactRel, limits.ArtifactBytes)
	if err != nil {
		return Result{}, err
	}

	// Step 11: re-verify the worker, source-snapshot, and fingerprinted
	// toolchain identities, then terminate and join the complete worker domain
	// before anything is published.
	if err := request.Source.Recheck(); err != nil {
		return Result{}, diagnosticErr("source_mutated", err, "frozen source changed before artifact acceptance")
	}
	if err := request.Session.VerifyToolchain(ctx); err != nil {
		return Result{}, err
	}
	if err := executable.Verify(); err != nil {
		return Result{}, err
	}
	client.teardown()
	return Result{Artifact: Artifact{StagedPath: artifactPath, Metadata: metadata}, Evidence: client.Evidence()}, nil
}

// managerExecutable is the path of the running manager. It is resolved from the
// process itself and never from a package byte, manifest value, environment
// value, or PATH lookup.
var managerExecutable = func() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return path
}

// validatePackageCommandSurface rejects every package attempt to reach the
// execution boundary before any private state is created.
func validatePackageCommandSurface(request BuildRequest) error {
	object := request.CommandObject
	if object == nil {
		return diagnostic(CodePackageInfluenceForbidden, "the package build-command surface was not presented for validation")
	}
	extra := make([]string, 0, len(object))
	for key := range object {
		switch key {
		case "type", "driver", "source_dir":
		default:
			extra = append(extra, key)
		}
	}
	if len(extra) != 0 {
		sort.Strings(extra)
		surface := packageInfluenceSurfaces[extra[0]]
		if surface == "" {
			surface = "execution boundary, worker, controls, limits, permits, or publication"
		}
		return diagnostic(CodePackageInfluenceForbidden,
			"package build command declares %q, which selects the %s", extra[0], surface)
	}
	if value, _ := object["type"].(string); value != "build" {
		return diagnostic(CodePackageInfluenceForbidden, "package build command type %q is not the closed build type", value)
	}
	if value, _ := object["driver"].(string); value != buildmeta.DriverGoV1 {
		return diagnostic(CodePackageInfluenceForbidden, "package build command driver %q is not the closed go-v1 driver", value)
	}
	if value, _ := object["source_dir"].(string); value != request.SourceDir {
		return diagnostic(CodePackageInfluenceForbidden, "package build command source_dir %q does not match the validated command", value)
	}
	return nil
}

func normalizeBuildLimits(limits ResourceLimits) (ResourceLimits, error) {
	if limits.Timeout == 0 {
		limits.Timeout = defaultBuildTimeout
	}
	if limits.OutputBytes == 0 {
		limits.OutputBytes = defaultBuildOutput
	}
	if limits.ArtifactBytes == 0 {
		limits.ArtifactBytes = defaultArtifactLimit
	}
	if limits.FileBytes == 0 {
		limits.FileBytes = defaultFileLimit
	}
	if limits.DiskBytes == 0 {
		limits.DiskBytes = defaultDiskLimit
	}
	if limits.MemoryBytes == 0 {
		limits.MemoryBytes = defaultMemoryLimit
	}
	if limits.Processes == 0 {
		limits.Processes = defaultProcessLimit
	}
	if limits.Timeout < time.Millisecond || limits.Timeout > defaultBuildTimeout ||
		limits.OutputBytes < 1 || limits.OutputBytes > defaultBuildOutput ||
		limits.ArtifactBytes < 1 || limits.ArtifactBytes > defaultArtifactLimit ||
		limits.FileBytes < limits.ArtifactBytes || limits.FileBytes > defaultFileLimit ||
		limits.DiskBytes < limits.ArtifactBytes || limits.DiskBytes > defaultDiskLimit ||
		limits.MemoryBytes < 1 || limits.MemoryBytes > defaultMemoryLimit ||
		limits.Processes < 1 || limits.Processes > defaultProcessLimit {
		return ResourceLimits{}, diagnostic("invalid_resource_limits", "go-v1 limits are outside manager bounds")
	}
	return limits, nil
}

func canonicalBuildDirectories(snapshot, buildRootRel, sourceDirRel string) (string, string, string, error) {
	physicalSnapshot, err := filepath.EvalSymlinks(snapshot)
	if err != nil {
		return "", "", "", diagnosticErr("invalid_build_request", err, "cannot resolve frozen source root")
	}
	buildRoot := filepath.Join(physicalSnapshot, filepath.FromSlash(buildRootRel))
	sourceDir := filepath.Join(physicalSnapshot, filepath.FromSlash(sourceDirRel))
	for _, item := range []struct{ path, label string }{{buildRoot, "build root"}, {sourceDir, "source directory"}} {
		info, err := os.Lstat(item.path)
		if err != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
			return "", "", "", diagnosticErr("invalid_build_request", err, "%s is not a real directory", item.label)
		}
		physical, err := filepath.EvalSymlinks(item.path)
		if err != nil || physical != item.path {
			return "", "", "", diagnosticErr("invalid_build_request", err, "%s is not canonical and link-free", item.label)
		}
	}
	goMod := filepath.Join(buildRoot, "go.mod")
	info, err := os.Lstat(goMod)
	if err != nil || !info.Mode().IsRegular() {
		return "", "", "", diagnosticErr("build_module_missing", err, "build root lacks a regular go.mod")
	}
	if err := rejectWorkspaceAndToolchainDirectives(buildRoot, goMod); err != nil {
		return "", "", "", err
	}
	return physicalSnapshot, buildRoot, sourceDir, nil
}

func rejectWorkspaceAndToolchainDirectives(buildRoot, goMod string) error {
	var workspace string
	err := filepath.WalkDir(buildRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Name() == "go.work" {
			workspace = path
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return diagnosticErr("workspace_dependency_forbidden", err, "cannot prove workspace exclusion")
	}
	if workspace != "" {
		return diagnostic("workspace_dependency_forbidden", "build root contains forbidden workspace file %q", workspace)
	}
	payload, err := os.ReadFile(goMod) // #nosec G304 -- direct go.mod below validated build root
	if err != nil {
		return diagnosticErr("build_module_missing", err, "cannot inspect build root go.mod")
	}
	for _, line := range bytes.Split(payload, []byte{'\n'}) {
		fields := bytes.Fields(line)
		if len(fields) != 0 && bytes.Equal(fields[0], []byte("toolchain")) {
			return diagnostic("toolchain_switch_forbidden", "build root go.mod contains a package-selected toolchain directive")
		}
	}
	return nil
}

// privateRoots lists the operation-private write targets the manager resolved
// independently of package data.
func privateRoots(environment []string, stage string) []string {
	values := environmentMap(environment)
	keys := []string{"GOPATH", "GOMODCACHE", "GOCACHE", "GOTMPDIR", "HOME", "XDG_CONFIG_HOME", "TMPDIR", "APPDATA", "LOCALAPPDATA", "USERPROFILE", "TEMP", "TMP"}
	result := make([]string, 0, len(keys)+1)
	seen := map[string]bool{}
	for _, key := range keys {
		if value := values[key]; value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return append(result, stage)
}

// verifyPrivateState is a post-child integrity check over manager-private
// state. It is not a kernel-enforced write confinement or an aggregate disk
// quota; the portable profile claims neither.
func verifyPrivateState(root, staging string, limit int64) error {
	var total int64
	err := filepath.WalkDir(root, func(path string, _ fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		switch {
		case info.IsDir():
			return nil
		case info.Mode().IsRegular():
			if info.Size() > limit-total {
				return diagnostic("process_disk_limit", "Go private state exceeded its disk bound")
			}
			total += info.Size()
			return nil
		default:
			if isWithin(path, staging) {
				return nil
			}
			return diagnostic("private_build_special_file", "Go created a link or special file in private state")
		}
	})
	if err != nil {
		if DiagnosticCode(err) != "" {
			return err
		}
		return diagnosticErr("private_build_unreadable", err, "cannot inspect private Go state")
	}
	return nil
}

func verifyArtifact(stage, artifactPath, artifactRel string, limit int64) (buildmeta.Artifact, error) {
	entries, err := os.ReadDir(stage)
	if err != nil || len(entries) != 1 || entries[0].Name() != "bin" || !entries[0].IsDir() {
		return buildmeta.Artifact{}, diagnosticErr("artifact_output_invalid", err, "build staging contains an unexpected output")
	}
	binEntries, err := os.ReadDir(filepath.Join(stage, "bin"))
	if err != nil || len(binEntries) != 1 || filepath.Join(stage, "bin", binEntries[0].Name()) != artifactPath {
		return buildmeta.Artifact{}, diagnosticErr("artifact_output_invalid", err, "build did not produce exactly one derived output")
	}
	info, err := os.Lstat(artifactPath)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return buildmeta.Artifact{}, diagnosticErr("artifact_special_file", err, "staged output is not a regular file")
	}
	if info.Size() < 0 || info.Size() > limit {
		return buildmeta.Artifact{}, diagnostic("artifact_size_limit", "staged output exceeds its artifact bound")
	}
	if err := applyArtifactPermissions(artifactPath); err != nil {
		return buildmeta.Artifact{}, diagnosticErr("artifact_permissions_failed", err, "cannot apply manager artifact permissions")
	}
	permissioned, err := os.Lstat(artifactPath)
	if err != nil || !permissioned.Mode().IsRegular() || !os.SameFile(info, permissioned) || permissioned.Size() != info.Size() {
		return buildmeta.Artifact{}, diagnosticErr("artifact_mutated", err, "staged output changed while applying permissions")
	}
	multiple, err := artifactHasMultipleLinks(artifactPath, permissioned)
	if err != nil {
		return buildmeta.Artifact{}, diagnosticErr("artifact_unreadable", err, "cannot inspect staged output link count")
	}
	if multiple {
		return buildmeta.Artifact{}, diagnostic("artifact_link", "staged output has multiple filesystem links")
	}
	file, err := os.Open(artifactPath) // #nosec G304 -- path is manager-derived inside private staging
	if err != nil {
		return buildmeta.Artifact{}, diagnosticErr("artifact_unreadable", err, "cannot open staged output")
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return buildmeta.Artifact{}, diagnosticErr("artifact_special_file", err, "staged output changed while opening")
	}
	digest := sha256.New()
	written, copyErr := io.CopyN(digest, file, opened.Size())
	var extra [1]byte
	extraCount, extraErr := file.Read(extra[:])
	if copyErr != nil || written != opened.Size() || extraCount != 0 || (extraErr != nil && extraErr != io.EOF) {
		return buildmeta.Artifact{}, diagnosticErr("artifact_mutated", errors.Join(copyErr, extraErr), "staged output changed while hashing")
	}
	after, err := os.Lstat(artifactPath)
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(info, after) || after.Size() != opened.Size() {
		return buildmeta.Artifact{}, diagnosticErr("artifact_mutated", err, "staged output changed during verification")
	}
	return buildmeta.Artifact{Path: artifactRel, SHA256: "sha256:" + hex.EncodeToString(digest.Sum(nil)), Size: written}, nil
}

func classifyListFailure(err error, stderr []byte) error {
	if DiagnosticCode(err) != "" {
		return err
	}
	text := string(stderr)
	switch {
	case containsAny(text, "inconsistent vendoring", "is explicitly required in go.mod, but not marked as explicit in vendor/modules.txt"):
		return diagnosticErr("vendor_metadata_inconsistent", err, "go list rejected inconsistent vendor metadata")
	case containsAny(text, "cannot find module providing package", "import lookup disabled by -mod=vendor"):
		return diagnosticErr("vendor_dependency_missing", err, "go list could not resolve a vendored dependency")
	case containsAny(text, "go.mod requires go >=", "requires go >=", "toolchain not available"):
		return diagnosticErr("toolchain_switch_forbidden", err, "module requests another toolchain")
	case containsAny(text, "go.work", "workspace"):
		return diagnosticErr("workspace_dependency_forbidden", err, "module depends on a forbidden workspace")
	case containsAny(text, "no Go files", "build constraints exclude all Go files"):
		return diagnosticErr("cgo_required", err, "package has no buildable non-cgo files")
	default:
		return diagnosticErr("go_list_failed", err, "fixed go list command failed")
	}
}

func classifyBuildFailure(err error, stderr []byte) error {
	if DiagnosticCode(err) != "" {
		return err
	}
	text := string(stderr)
	switch {
	case containsAny(text, "requires external linking", "external linking required"):
		return diagnosticErr("external_link_forbidden", err, "internal-only Go build failed")
	case containsAny(text, "libgcc", "gcc"):
		return diagnosticErr("libgcc_fallback_forbidden", err, "Go build attempted a forbidden host linker fallback")
	default:
		return diagnosticErr("go_build_failed", err, "fixed go build command failed")
	}
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if len(needle) > 0 && contains(value, needle) {
			return true
		}
	}
	return false
}

func contains(value, needle string) bool {
	for index := 0; index+len(needle) <= len(value); index++ {
		if value[index:index+len(needle)] == needle {
			return true
		}
	}
	return false
}

func containsExact(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func (limits ResourceLimits) String() string {
	return fmt.Sprintf("timeout=%s output=%d artifact=%d file=%d disk=%d memory=%d processes=%d",
		limits.Timeout, limits.OutputBytes, limits.ArtifactBytes, limits.FileBytes, limits.DiskBytes, limits.MemoryBytes, limits.Processes)
}
