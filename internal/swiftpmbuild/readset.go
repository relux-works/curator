package swiftpmbuild

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/swiftpminterop"
	"github.com/relux-works/curator/internal/swiftpmsource"
)

// ObservationSchemaID identifies the offline dependency-scan evidence.
const ObservationSchemaID = "swiftpm-observed-reads-v1"

// ReadSetObserver is the verified compiler read-set provider. It runs one
// offline native SwiftPM build from the accepted capture, under an isolated
// home with network denied and resolution frozen, and answers every interop
// read-set request from the dependency files the compilers themselves emitted.
// This is what makes the C-family header closure adversarially complete: the
// proof stops depending on statically reproducing Clang's search behaviour and
// starts depending on what the selected compiler actually read.
//
// Portable assurance cannot confine reads at the operating-system boundary, so
// a compiler-emitted dependency file is corroboration rather than proof there.
// In that mode the observer honestly reports not-observed and the interop
// stage keeps its reject-by-default portable verdict.
type ReadSetObserver struct {
	Config        Config
	Capture       *swiftpmsource.Capture
	Triple        string
	Configuration string
	// Driver is the exact SwiftPM build driver identity. It must be the same
	// physical component the build binding resolves for the driver slot.
	Driver swiftpmsource.ToolIdentity

	observed  map[string][]swiftpminterop.ObservedRead
	receiptID closuregraph.ID
	failure   error
	ran       bool
}

// NewReadSetObserver binds one observation pass to the accepted capture.
func NewReadSetObserver(config Config, capture *swiftpmsource.Capture, driver swiftpmsource.ToolIdentity, triple string) (*ReadSetObserver, error) {
	if capture == nil || driver.ExecutableRelativePath == "" || triple == "" {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM read-set observer authority is incomplete")
	}
	if config.Executor == nil || config.Store == nil || config.Policy == nil || config.Recheck == nil || config.Configuration == "" {
		return nil, fail(CodeDerivationUnauthorized, "SwiftPM read-set observer services are incomplete")
	}
	return &ReadSetObserver{Config: config, Capture: capture, Triple: triple, Configuration: config.Configuration, Driver: driver}, nil
}

// ObserveReads implements swiftpminterop.ReadSetProvider.
func (observer *ReadSetObserver) ObserveReads(ctx context.Context, request swiftpminterop.ReadSetRequest) (swiftpminterop.ReadSetResult, error) {
	if observer == nil {
		return swiftpminterop.ReadSetResult{}, fail(CodeDerivationUnauthorized, "SwiftPM read-set observer is absent")
	}
	if observer.Config.Assurance != closureexec.AssuranceVerified {
		// Portable execution cannot prove the compiler read nothing else, so
		// claiming observation here would inflate the assurance level.
		return swiftpminterop.ReadSetResult{Observed: false}, nil
	}
	if err := observer.run(ctx); err != nil {
		return swiftpminterop.ReadSetResult{}, err
	}
	// SwiftPM requires globally unique module names in one build graph, so the
	// target name alone identifies the emitted dependency evidence.
	reads, present := observer.observed[request.Target]
	if !present {
		return swiftpminterop.ReadSetResult{}, failFields(CodeHeaderInputUndeclared, map[string]string{"target": request.Package + ":" + request.Target}, "offline observation produced no dependency evidence for a selected target")
	}
	return swiftpminterop.ReadSetResult{Observed: true, Reads: reads, ReceiptID: observer.receiptID}, nil
}

// run performs the single offline observation build and harvests every
// compiler-emitted dependency file exactly once.
func (observer *ReadSetObserver) run(ctx context.Context) error {
	if observer.ran {
		return observer.failure
	}
	observer.ran = true
	observer.failure = observer.observe(ctx)
	return observer.failure
}

func (observer *ReadSetObserver) observe(ctx context.Context) error {
	config := observer.Config
	key := "swiftpm-observe:" + string(observer.Capture.GraphDigest) + ":" + observer.Triple + ":" + observer.Configuration
	buildRoot, cleanupRoot, err := materializeCaptureRoot(config, observer.Capture)
	if err != nil {
		return err
	}
	defer cleanupRoot()
	rootInput, rootReceiptID, err := admitDerivedRoot(ctx, config, observer.Capture, key, buildRoot)
	if err != nil {
		return err
	}
	mirrors, err := observer.Capture.OfflineMirrors()
	if err != nil {
		return err
	}
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{rootReceiptID: rootInput}
	mounts := []closureexec.InputMount{{ReceiptID: rootReceiptID, Path: buildRootMount}}
	for _, mirror := range mirrors {
		inputs[mirror.ReceiptID] = mirror.Input
		mounts = append(mounts, closureexec.InputMount{ReceiptID: mirror.ReceiptID, Path: mirror.Mount})
	}
	sort.Slice(mounts, func(i, j int) bool { return mounts[i].ReceiptID < mounts[j].ReceiptID })
	receiptIDs := make([]closuregraph.ID, len(mounts))
	readRoots := make([]string, len(mounts))
	for index, mount := range mounts {
		receiptIDs[index], readRoots[index] = mount.ReceiptID, mount.Path
	}
	sort.Strings(readRoots)

	scratch := swiftpmScratchDirectory(observer.Triple, observer.Configuration)
	evidencePath := path.Join(buildWorkMount, scratch, "description.json")
	writeRoots := []string{buildWorkMount, evidencePath}
	sort.Strings(writeRoots)
	artifactID, err := closuregraph.DomainID("swiftpm-observation-evidence-v1", map[string]any{"key": key, "path": evidencePath, "schema": ObservationSchemaID})
	if err != nil {
		return err
	}
	requirements := []closureexec.EvidenceRequirement{{Path: evidencePath, SchemaID: ObservationSchemaID, ArtifactManifestID: artifactID}}
	evidenceSchemaID, err := buildEvidenceSchemaID(requirements)
	if err != nil {
		return err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 64 << 20, ReadBytes: 4 << 30, WriteBytes: 4 << 30, WallTimeMillis: 900_000, ProcessCount: 512}
	limitID, err := limits.ID()
	if err != nil {
		return err
	}
	if err = requireEmptyOutputRoot(config.OutputRoot); err != nil {
		return err
	}
	c0ID, err := observer.Capture.C0.ID()
	if err != nil {
		return err
	}
	platform := closuregraph.Node{Kind: closuregraph.NodeTargetPlatform, LogicalKey: "swiftpm.platform.target", Payload: observer.Capture.Destination().Platform}
	platformID, err := platform.ID()
	if err != nil {
		return err
	}
	toolNodeID, err := observationToolNodeID(observer.Driver)
	if err != nil {
		return err
	}
	command := observationCommand(observer)
	operation, err := config.Executor.Preflight(ctx)
	if err != nil {
		return err
	}
	permit := closureexec.DerivationPermit{
		SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: operation.CurrentCausalHead(),
		InvocationKey: key, InvocationSubtype: closureexec.DerivationMetadata,
		AdmittedInputReceiptIDs: receiptIDs, InputMounts: mounts,
		WorkCopies:     []closureexec.WorkCopy{{ReceiptID: rootReceiptID, Path: buildWorkMount, Retain: true}},
		C0CheckpointID: c0ID, ToolchainNodeID: toolNodeID, ToolchainFingerprint: observer.Driver.Fingerprint,
		ExecutableSHA256: observer.Driver.ExecutableSHA256, Executable: command.Executable, CWD: command.CWD,
		Argv: append([]string(nil), command.Argv...), Environment: resolveEnvironment(config, command.Environment),
		HostID: platformID, TargetID: platformID, AllowedProcesses: observationProcesses(config, observer.Driver),
		ReadRoots: readRoots, WriteRoots: writeRoots, ExpectedEvidence: requirements, Network: "none",
		RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceSchemaID,
	}
	permitID, err := operation.Commit(permit)
	if err != nil {
		return err
	}
	workRoot := filepath.Join(config.ExecutionRoot, filepath.FromSlash(buildWorkMount))
	defer func() { _ = os.RemoveAll(workRoot) }()
	receipt, err := operation.Execute(ctx, permitID, func(recheckCtx context.Context) (closureexec.ToolchainIdentity, error) {
		return config.Recheck(recheckCtx, observer.Driver)
	}, inputs)
	if err != nil {
		return mapExecutionError(err)
	}
	if err = operation.VerifyIssuedDerivationReceipt(receipt); err != nil {
		return err
	}
	if receipt.AssuranceMode != closureexec.AssuranceVerified || receipt.Audit.Network != "none" {
		return fail(CodeHeaderInputUndeclared, "offline read observation lacks a verified network-denied boundary")
	}
	observer.receiptID, err = receipt.ID()
	if err != nil {
		return err
	}
	observer.observed, err = harvestDependencyFiles(observer, filepath.Join(workRoot, filepath.FromSlash(scratch)))
	return err
}

// observationCommand is the exact dependency-scan build. It differs from the
// publication build only in that it builds every selected target rather than
// exactly one product, so every target emits its dependency file.
func observationCommand(observer *ReadSetObserver) Command {
	isolated := executionRootPlaceholder + "/" + buildWorkMount + "/" + scratchRoot
	return Command{
		Executable: observer.Driver.ExecutableRelativePath, CWD: buildWorkMount,
		Argv: []string{
			"build", "--package-path", ".",
			"--cache-path", path.Join(scratchRoot, "cache"),
			"--config-path", path.Join(scratchRoot, "config"),
			"--security-path", path.Join(scratchRoot, "security"),
			"--scratch-path", path.Join(scratchRoot, "scratch"),
			"--disable-netrc", "--disable-experimental-prebuilts", "--force-resolved-versions",
			"--build-system", "native", "--configuration", observer.Configuration, "--triple", observer.Triple,
		},
		Environment: map[string]string{
			"HOME": isolated + "/home", "SWIFTPM_CACHE_DIR": isolated + "/cache",
			"SWIFTPM_CONFIG_DIR": isolated + "/config", "SWIFTPM_SCRATCH_DIR": isolated + "/scratch",
			"SWIFTPM_SECURITY_DIR": isolated + "/security", "TZ": "UTC",
		},
	}
}

func observationProcesses(config Config, driver swiftpmsource.ToolIdentity) []string {
	processes := append([]string(nil), config.AllowedProcesses...)
	if len(processes) == 0 {
		processes = []string{driver.ExecutableRelativePath}
	}
	sort.Strings(processes)
	return processes
}

func observationToolNodeID(driver swiftpmsource.ToolIdentity) (closuregraph.ID, error) {
	fingerprints := []closuregraph.ID{driver.ExecutableSHA256}
	node := closuregraph.Node{Kind: closuregraph.NodeToolchainComponent, LogicalKey: "swiftpm.tool." + driver.Role, Payload: closuregraph.ToolchainComponentPayload{
		ComponentRole: driver.Role, ContentFingerprint: driver.Fingerprint, ExecutableRelativePath: driver.ExecutableRelativePath,
		PlatformABI: driver.PlatformABI, PolicySelector: driver.PolicySelector, VersionOutput: driver.VersionOutput,
		LinkFingerprintIDs: fingerprints, TimeOfUseRecheckRule: "immediate-exact-v1",
		ExecutionDomain: closuregraph.ExecutionHost, PlatformRoleNames: []closuregraph.PlatformRole{closuregraph.PlatformHost},
	}}
	return node.ID()
}

// harvestDependencyFiles parses every compiler-emitted dependency file in the
// retained private build tree and maps each observed read back to the admitted
// closure or leaves it verbatim for the interop binding resolver.
func harvestDependencyFiles(observer *ReadSetObserver, scratchRootPath string) (map[string][]swiftpminterop.ObservedRead, error) {
	packageRoots, err := admittedPackageRoots(observer.Capture)
	if err != nil {
		return nil, err
	}
	workPackageRoot := filepath.Join(observer.Config.ExecutionRoot, filepath.FromSlash(buildWorkMount))
	if len(observer.Capture.Packages) == 0 {
		return nil, fail(CodeHeaderInputUndeclared, "capture admitted no package root for the observed read set")
	}
	rootIdentity := observer.Capture.Packages[0].Identity
	observed := map[string][]swiftpminterop.ObservedRead{}
	entries, err := os.ReadDir(scratchRootPath)
	if err != nil {
		return nil, fail(CodeHeaderInputUndeclared, "offline observation produced no build tree")
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".build") {
			continue
		}
		target := strings.TrimSuffix(entry.Name(), ".build")
		directory := filepath.Join(scratchRootPath, entry.Name())
		files, readErr := os.ReadDir(directory)
		if readErr != nil {
			return nil, readErr
		}
		reads := map[string]string{}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".d") {
				continue
			}
			payload, fileErr := os.ReadFile(filepath.Join(directory, file.Name())) // #nosec G304 -- retained private work copy below the validated execution root.
			if fileErr != nil {
				return nil, fileErr
			}
			for _, dependency := range parseDependencyFile(string(payload)) {
				reads[dependency] = classifyRead(dependency)
			}
		}
		mapped := []swiftpminterop.ObservedRead{}
		for observedPath := range reads {
			resolved, inBuildTree, mapErr := mapObservedRead(observedPath, workPackageRoot, rootIdentity, packageRoots)
			if mapErr != nil {
				return nil, mapErr
			}
			if inBuildTree {
				// Reads of derived build state — the module cache and the
				// module maps SwiftPM generates below the scratch tree — are
				// covered by the same permitted, network-denied derivation.
				// Dependency source is not derived build state and is never
				// dropped here; mapObservedRead rewrites it instead.
				continue
			}
			mapped = append(mapped, swiftpminterop.ObservedRead{Path: resolved, Class: reads[observedPath]})
		}
		sort.Slice(mapped, func(i, j int) bool { return mapped[i].Path < mapped[j].Path })
		if _, duplicate := observed[target]; duplicate {
			return nil, failFields(CodeHeaderInputUndeclared, map[string]string{"target": target}, "offline observation emitted two build trees for one target name")
		}
		observed[target] = mapped
	}
	if len(observed) == 0 {
		return nil, fail(CodeHeaderInputUndeclared, "offline observation emitted no compiler dependency file")
	}
	return observed, nil
}

func admittedPackageRoots(capture *swiftpmsource.Capture) (map[string]string, error) {
	roots := map[string]string{}
	for _, pkg := range capture.Packages {
		root, err := pkg.ProtectedRoot()
		if err != nil {
			return nil, err
		}
		roots[pkg.Identity] = root
	}
	return roots, nil
}

// checkoutsSegments is SwiftPM's dependency checkout prefix inside the
// isolated scratch tree. Every source-control dependency materializes its
// admitted source and headers below it, so a read there is closure source and
// not derived build state.
var checkoutsSegments = []string{scratchRoot, "scratch", "checkouts"}

// mapObservedRead rewrites a read of the private work copy back to the exact
// admitted protected tree it was copied from. A read of a dependency checkout
// is rewritten to that dependency's admitted root, so the verified header
// proof covers every package and not only the root one. Only a genuinely
// derived read — the module cache and the module maps SwiftPM generates below
// the scratch tree — is reported as locally produced build state. A checkout
// read that matches no admitted package identity fails closed.
func mapObservedRead(observed, workPackageRoot, rootIdentity string, packageRoots map[string]string) (string, bool, error) {
	cleaned, err := filepath.Abs(observed)
	if err != nil {
		return "", false, err
	}
	relative, err := filepath.Rel(workPackageRoot, cleaned)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return cleaned, false, nil
	}
	segments := strings.Split(filepath.ToSlash(relative), "/")
	if segments[0] != scratchRoot {
		return joinAdmitted(packageRoots, rootIdentity, segments)
	}
	if !hasPrefixSegments(segments, checkoutsSegments) {
		return cleaned, true, nil
	}
	identity := ""
	if len(segments) > len(checkoutsSegments) {
		identity = segments[len(checkoutsSegments)]
	}
	return joinAdmitted(packageRoots, identity, segments[min(len(segments), len(checkoutsSegments)+1):])
}

// joinAdmitted resolves one admitted package root and rebinds the observed
// remainder to it. An unknown identity is undeclared closure input.
func joinAdmitted(packageRoots map[string]string, identity string, remainder []string) (string, bool, error) {
	root, admitted := packageRoots[identity]
	if !admitted || root == "" {
		return "", false, failFields(CodeHeaderInputUndeclared, map[string]string{"package": identity}, "observed read names no admitted package root")
	}
	return filepath.Join(append([]string{root}, remainder...)...), false, nil
}

// hasPrefixSegments reports whether path segments start with the given prefix.
func hasPrefixSegments(segments, prefix []string) bool {
	if len(segments) < len(prefix) {
		return false
	}
	for index, value := range prefix {
		if segments[index] != value {
			return false
		}
	}
	return true
}

// classifyRead labels an observed read by its concrete grammar so the interop
// resolver records a typed resolution rather than an opaque path.
func classifyRead(value string) string {
	switch strings.ToLower(filepath.Ext(value)) {
	case ".h", ".hh", ".hpp", ".hxx", ".inc", ".pch":
		return "header"
	case ".modulemap":
		return "module-map"
	case ".swiftinterface", ".swiftmodule":
		return "swift-module"
	case ".c", ".cc", ".cpp", ".cxx", ".m", ".mm", ".swift", ".s":
		return "source"
	default:
		return "input"
	}
}

// parseDependencyFile decodes the Make-style dependency grammar every C-family
// and Swift driver emits: one or more `target: prerequisite...` rules with
// backslash line continuations and backslash-escaped spaces.
func parseDependencyFile(payload string) []string {
	joined := strings.ReplaceAll(payload, "\\\n", " ")
	joined = strings.ReplaceAll(joined, "\\\r\n", " ")
	dependencies := []string{}
	for _, line := range strings.Split(joined, "\n") {
		colon := indexRuleSeparator(line)
		if colon < 0 {
			continue
		}
		for _, token := range splitDependencyTokens(line[colon+1:]) {
			if token != "" {
				dependencies = append(dependencies, token)
			}
		}
	}
	return dependencies
}

// indexRuleSeparator finds the rule separator, skipping a Windows drive letter
// and any backslash-escaped colon inside a path.
func indexRuleSeparator(line string) int {
	for index := 0; index < len(line); index++ {
		if line[index] != ':' {
			continue
		}
		if index > 0 && line[index-1] == '\\' {
			continue
		}
		if index == 1 && len(line) > 2 && (line[2] == '/' || line[2] == '\\') {
			continue
		}
		return index
	}
	return -1
}

// splitDependencyTokens splits on unescaped whitespace and unescapes the
// backslash-space and backslash-colon sequences the grammar defines.
func splitDependencyTokens(value string) []string {
	tokens := []string{}
	current := strings.Builder{}
	escaped := false
	for index := 0; index < len(value); index++ {
		character := value[index]
		if escaped {
			current.WriteByte(character)
			escaped = false
			continue
		}
		switch character {
		case '\\':
			escaped = true
		case ' ', '\t', '\r':
			if current.Len() != 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(character)
		}
	}
	if current.Len() != 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

var _ swiftpminterop.ReadSetProvider = (*ReadSetObserver)(nil)
