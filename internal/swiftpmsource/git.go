package swiftpmsource

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/privatedir"
)

// GitBroker is the production source-control acquisition boundary. It fetches
// one exact origin into a task-owned bare repository, verifies the pinned
// object graph, and materializes only the committed tree bytes.
type GitBroker struct {
	WorkRoot              string
	GitExecutable         string
	ProcessStartObserver  func([]string)
	ProcessLaunchObserver closureexec.ProcessLaunchObserver
	authority             *sharedGitAuthority
}

// sharedGitAuthority delegates every authority decision to closureexec. The
// SwiftPM adapter owns only Git command/profile construction.
type sharedGitAuthority struct {
	mu                  sync.Mutex
	c0ID                closuregraph.ID
	toolchain           Toolchain
	executionRoot       string
	toolRoot            string
	executable          string
	allowedProcesses    []string
	acquisition         *closureexec.SourceAcquisitionExecutor
	acquisitionRunner   *closureexec.ManagerSourceAcquisitionRunner
	derivation          *closureexec.Executor
	derivationRunner    *closureexec.ManagerProcessRunner
	acquisitionReceipts map[closuregraph.ID]closureexec.SourceAcquisitionReceipt
	derivationPermits   map[closuregraph.ID]closureexec.DerivationPermit
	derivationReceipts  map[closuregraph.ID]closureexec.DerivationReceipt
	sequence            int64
}

func newSharedGitAuthority(c0ID closuregraph.ID, toolchain Toolchain, executionRoot, toolRoot string) (*sharedGitAuthority, error) {
	root, err := cleanAbsoluteRoot(executionRoot)
	if err != nil {
		return nil, fail(CodeDerivationUnauthorized, "Git execution root is not exact")
	}
	trustedRoot, err := cleanAbsoluteRoot(toolRoot)
	if err != nil {
		return nil, fail(CodeDerivationUnauthorized, "C0 Git tool root is not exact")
	}
	executable := filepath.Join(trustedRoot, filepath.FromSlash(toolchain.Git.ExecutableRelativePath))
	resolvedRoot, rootErr := filepath.EvalSymlinks(trustedRoot)
	resolvedExecutable, executableErr := filepath.EvalSymlinks(executable)
	if rootErr != nil || executableErr != nil || !pathWithinRoot(resolvedRoot, resolvedExecutable) {
		return nil, fail(CodeDerivationUnauthorized, "C0 Git executable escapes its declared root")
	}
	if _, err = cleanAbsoluteFile(resolvedExecutable); err != nil {
		return nil, fail(CodeDerivationUnauthorized, "C0 Git executable is unavailable: %v", err)
	}
	if len(toolchain.Git.ProcessFamily) > 0 {
		expectedFingerprint, fingerprintErr := toolProcessFamilyFingerprint(toolchain.Git)
		if fingerprintErr != nil || expectedFingerprint != toolchain.Git.Fingerprint {
			return nil, fail(CodeDerivationUnauthorized, "C0 Git process-family fingerprint is invalid")
		}
	}
	allowedProcesses := []string{toolchain.Git.ExecutableRelativePath}
	for _, member := range toolchain.Git.ProcessFamily {
		candidate := filepath.Join(resolvedRoot, filepath.FromSlash(member.ExecutableRelativePath))
		resolvedCandidate, resolveErr := filepath.EvalSymlinks(candidate)
		if resolveErr != nil || !pathWithinRoot(resolvedRoot, resolvedCandidate) {
			return nil, fail(CodeDerivationUnauthorized, "C0 Git process-family executable escapes its declared root")
		}
		payload, readErr := os.ReadFile(resolvedCandidate) // #nosec G304 -- candidate is symlink-resolved below the exact C0 tool root.
		if readErr != nil || sha256Bytes(payload) != member.ExecutableSHA256 {
			return nil, fail(CodeDerivationUnauthorized, "C0 Git process-family executable identity changed")
		}
		allowedProcesses = append(allowedProcesses, member.ExecutableRelativePath)
	}
	allowedProcesses = sortedUniqueStrings(allowedProcesses)
	acquisitionRunner, err := closureexec.NewManagerSourceAcquisitionRunnerWithExecutableRoot(root, resolvedRoot)
	if err != nil {
		return nil, err
	}
	acquisition, err := closureexec.NewSourceAcquisitionExecutor(closureexec.AssuranceConfig{Mode: closureexec.AssurancePortable}, acquisitionRunner, nil, string(c0ID))
	if err != nil {
		return nil, err
	}
	verificationOutputRoot := filepath.Join(root, ".swiftpm-authority", "verification-output")
	processRunner, err := closureexec.NewManagerProcessRunnerWithExecutableRoot(root, resolvedRoot, verificationOutputRoot)
	if err != nil {
		return nil, err
	}
	mirrorRunner, err := closureexec.NewSourceControlMirrorRunner(root, filepath.Join(root, ".swiftpm-authority", "mirror-output"))
	if err != nil {
		return nil, err
	}
	mirrorRunner.Delegate = processRunner
	derivation, err := closureexec.NewAssuredExecutor(closureexec.AssuranceConfig{Mode: closureexec.AssurancePortable}, mirrorRunner, nil, string(c0ID))
	if err != nil {
		return nil, err
	}
	return &sharedGitAuthority{c0ID: c0ID, toolchain: toolchain, executionRoot: root, toolRoot: resolvedRoot, executable: resolvedExecutable, allowedProcesses: allowedProcesses, acquisition: acquisition, acquisitionRunner: acquisitionRunner, derivation: derivation, derivationRunner: processRunner, acquisitionReceipts: map[closuregraph.ID]closureexec.SourceAcquisitionReceipt{}, derivationPermits: map[closuregraph.ID]closureexec.DerivationPermit{}, derivationReceipts: map[closuregraph.ID]closureexec.DerivationReceipt{}}, nil
}

func toolProcessFamilyFingerprint(tool ToolIdentity) (closuregraph.ID, error) {
	members := make([]ToolProcessIdentity, len(tool.ProcessFamily))
	copy(members, tool.ProcessFamily)
	sort.Slice(members, func(i, j int) bool { return members[i].ExecutableRelativePath < members[j].ExecutableRelativePath })
	values := make([]any, len(members))
	for index, member := range members {
		values[index] = map[string]any{"executable_relative_path": member.ExecutableRelativePath, "executable_sha256": string(member.ExecutableSHA256)}
	}
	return closuregraph.DomainID("swiftpm-toolchain-process-family-v1", map[string]any{"executable_relative_path": tool.ExecutableRelativePath, "executable_sha256": string(tool.ExecutableSHA256), "members": values})
}

func sha256Bytes(payload []byte) closuregraph.ID {
	sum := sha256.Sum256(payload)
	return closuregraph.ID("sha256:" + hex.EncodeToString(sum[:]))
}

func (authority *sharedGitAuthority) recheck(ctx context.Context) (closureexec.ToolchainIdentity, error) {
	identity, err := authority.toolchain.Recheck(ctx, authority.toolchain.Git)
	if err != nil {
		return closureexec.ToolchainIdentity{}, err
	}
	for _, member := range authority.toolchain.Git.ProcessFamily {
		pathValue := filepath.Join(authority.toolRoot, filepath.FromSlash(member.ExecutableRelativePath))
		resolved, resolveErr := filepath.EvalSymlinks(pathValue)
		if resolveErr != nil || !pathWithinRoot(authority.toolRoot, resolved) {
			return closureexec.ToolchainIdentity{}, fail(CodeDerivationUnauthorized, "C0 Git process family escaped before use")
		}
		payload, readErr := os.ReadFile(resolved) // #nosec G304 -- path is re-resolved below the exact C0 tool root immediately before use.
		if readErr != nil || sha256Bytes(payload) != member.ExecutableSHA256 {
			return closureexec.ToolchainIdentity{}, fail("artifact_toolchain_identity_changed", "C0 Git process-family executable changed before use")
		}
	}
	return identity, nil
}

func (broker *GitBroker) bindGitAuthority(authority *sharedGitAuthority) error {
	if broker == nil || authority == nil {
		return fail(CodeDerivationUnauthorized, "Git broker executable differs from C0 Git")
	}
	resolved, _ := filepath.EvalSymlinks(broker.GitExecutable)
	if resolved != authority.executable {
		return fail(CodeDerivationUnauthorized, "Git broker executable differs from C0 Git")
	}
	broker.authority = authority
	if broker.ProcessLaunchObserver != nil {
		authority.acquisitionRunner.ProcessLaunchObserver = broker.ProcessLaunchObserver
		authority.derivationRunner.ProcessLaunchObserver = broker.ProcessLaunchObserver
	}
	return nil
}

func (authority *sharedGitAuthority) acquire(ctx context.Context, cwd, phase string, argv []string, canonicalOrigin, resolvedRevision, gitTree string, observer func([]string)) ([]byte, closuregraph.ID, closuregraph.ID, error) {
	if authority == nil || authority.acquisition == nil || authority.toolchain.Recheck == nil {
		return nil, "", "", fail(CodeDerivationUnauthorized, "shared Git acquisition authority is absent")
	}
	authority.mu.Lock()
	defer authority.mu.Unlock()
	authority.sequence++
	cwdRelative, err := filepath.Rel(authority.executionRoot, cwd)
	if err != nil || cwdRelative == "." || filepath.IsAbs(cwdRelative) || strings.HasPrefix(cwdRelative, ".."+string(filepath.Separator)) {
		return nil, "", "", fail(CodeDerivationUnauthorized, "Git acquisition cwd escapes its execution root")
	}
	cwdLogical := filepath.ToSlash(cwdRelative)
	evidencePath := filepath.ToSlash(filepath.Join(".swiftpm-authority", "acquisition-evidence", fmt.Sprintf("%06d-%s.bin", authority.sequence, phase)))
	manifestID, _ := closuregraph.DomainID("source-acquisition-expected-evidence-v1", map[string]any{"phase": phase, "schema_id": "source-control-acquisition-output-v1"})
	expected := []closureexec.EvidenceRequirement{{Path: evidencePath, SchemaID: "source-control-acquisition-output-v1", ArtifactManifestID: manifestID}}
	limits := closureexec.ResourceLimits{OutputBytes: 256 << 20, ReadBytes: 1 << 30, WriteBytes: 1 << 30, WallTimeMillis: 120000, ProcessCount: 1}
	limitID, _ := limits.ID()
	evidenceID, _ := closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": []any{map[string]any{"artifact_manifest_id": string(manifestID), "path": evidencePath, "schema_id": "source-control-acquisition-output-v1"}}})
	origin := canonicalOrigin
	if origin == "" {
		origin = gitOriginFromArgs(argv)
	}
	requested := gitRevisionFromArgs(argv)
	if requested == "" {
		requested = "constraint:" + string(sha256Bytes([]byte(strings.Join(argv, "\x00"))))
	}
	environment := gitEnvironmentMap(cwd, authority.executable)
	if validRevision(resolvedRevision) {
		environment["CURATOR_ACQUISITION_REVISION"] = strings.ToLower(resolvedRevision)
	}
	if validRevision(gitTree) {
		environment["CURATOR_ACQUISITION_GIT_TREE"] = strings.ToLower(gitTree)
	}
	permit := closureexec.SourceAcquisitionPermit{SchemaID: closureexec.SchemaSourceAcquisitionPermit, PreviousCausalHead: authority.acquisition.CurrentCausalHead(), SourceProfileID: ProfileID, CanonicalOrigin: origin, RequestedRevision: requested, C0CheckpointID: authority.c0ID, ToolchainNodeID: mustToolNodeID(authority.toolchain.Git), ToolchainFingerprint: authority.toolchain.Git.Fingerprint, ExecutableSHA256: authority.toolchain.Git.ExecutableSHA256, Executable: authority.toolchain.Git.ExecutableRelativePath, Argv: append([]string(nil), argv...), CWD: cwdLogical, Environment: environment, HostID: mustPlatformID(authority.toolchain.Git.PlatformABI), TargetID: mustPlatformID(authority.toolchain.Git.PlatformABI), AllowedProcesses: append([]string(nil), authority.allowedProcesses...), ReadRoots: []string{filepath.ToSlash(filepath.Dir(authority.toolchain.Git.ExecutableRelativePath))}, QuarantineWriteRoots: []string{cwdLogical, evidencePath}, NetworkPolicy: "exact-origin-only", ExpectedEvidence: expected, StdoutEvidencePath: evidencePath, ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID, RecheckRule: "immediate-exact-v1"}
	permit.QuarantineWriteRoots = sortedUniqueStrings(permit.QuarantineWriteRoots)
	permit.ReadRoots = sortedUniqueStrings(permit.ReadRoots)
	authority.acquisitionRunner.ProcessStartObserver = func(closureexec.SourceAcquisitionPermit) {
		if observer != nil {
			observer(append([]string(nil), argv...))
		}
	}
	permitID, err := authority.acquisition.Commit(ctx, permit)
	if err != nil {
		return nil, "", "", err
	}
	receipt, err := authority.acquisition.Execute(ctx, permitID, func(checkCtx context.Context) (closureexec.ToolchainIdentity, error) {
		return authority.recheck(checkCtx)
	})
	if err != nil {
		return nil, permitID, "", err
	}
	receiptID, _ := receipt.ID()
	authority.acquisitionReceipts[receiptID] = receipt
	outputPath := filepath.Join(authority.executionRoot, filepath.FromSlash(evidencePath))
	output, err := os.ReadFile(outputPath) // #nosec G304 -- exact manager-declared acquisition evidence.
	return output, permitID, receiptID, err
}

func (authority *sharedGitAuthority) verify(permitIDs, receiptIDs []closuregraph.ID) error {
	if authority == nil || len(permitIDs) == 0 || len(permitIDs) != len(receiptIDs) {
		return fail(CodeDerivationUnauthorized, "shared Git journal is incomplete")
	}
	for index, receiptID := range receiptIDs {
		receipt, ok := authority.acquisitionReceipts[receiptID]
		if !ok || receipt.PermitID != permitIDs[index] || authority.acquisition.VerifyIssuedReceipt(receipt) != nil {
			return fail(CodeDerivationUnauthorized, "Git acquisition receipt was not issued")
		}
	}
	return nil
}

func (authority *sharedGitAuthority) verifyDerivations(permitIDs, receiptIDs []closuregraph.ID) error {
	if authority == nil || len(permitIDs) == 0 || len(permitIDs) != len(receiptIDs) {
		return fail(CodeDerivationUnauthorized, "shared Git derivation journal is incomplete")
	}
	for index, receiptID := range receiptIDs {
		permit, permitOK := authority.derivationPermits[permitIDs[index]]
		receipt, receiptOK := authority.derivationReceipts[receiptID]
		if !permitOK || !receiptOK || receipt.PermitID != permitIDs[index] || authority.derivation.VerifyIssuedDerivationChain(permit, receipt) != nil {
			return fail(CodeDerivationUnauthorized, "Git derivation permit or receipt was not issued")
		}
	}
	return nil
}

func gitOriginFromArgs(argv []string) string {
	for index, value := range argv {
		if value == "--" && index+1 < len(argv) {
			return argv[index+1]
		}
	}
	return "local-quarantine:" + string(sha256Bytes([]byte(strings.Join(argv, "\x00"))))
}
func gitRevisionFromArgs(argv []string) string {
	for index := len(argv) - 1; index >= 0; index-- {
		value := strings.TrimSuffix(argv[index], "^{commit}")
		if validRevision(strings.ToLower(value)) {
			return strings.ToLower(value)
		}
	}
	return ""
}
func sortedUniqueStrings(values []string) []string {
	sort.Strings(values)
	output := values[:0]
	for _, value := range values {
		if len(output) == 0 || output[len(output)-1] != value {
			output = append(output, value)
		}
	}
	return output
}
func mustToolNodeID(tool ToolIdentity) closuregraph.ID {
	node, _ := toolNodeRecord(tool)
	id, _ := node.ID()
	return id
}
func mustPlatformID(abi string) closuregraph.ID {
	id, _ := closuregraph.DomainID("swiftpm-git-host-platform-v1", map[string]any{"abi": abi})
	return id
}

// ResolvePin resolves one closed manifest requirement through the broker
// without executing package code. Branch/tag names remain metadata; the
// returned Pin always carries one immutable commit.
func (broker *GitBroker) ResolvePin(ctx context.Context, dependency ManifestDependency) (Pin, error) {
	pin, _, err := broker.ResolvePinWithEvidence(ctx, dependency)
	return pin, err
}

// ResolvePinWithEvidence retains the exact Git permit/receipt pairs used to
// turn a mutable version, range, or branch declaration into one commit.
func (broker *GitBroker) ResolvePinWithEvidence(ctx context.Context, dependency ManifestDependency) (Pin, GitVerificationEvidence, error) {
	journal := GitVerificationEvidence{}
	if dependency.Kind != SourceRemote && dependency.Kind != SourceLocal {
		return Pin{}, journal, fail(CodeDependencyOriginUnsupported, "only source-control requirements have pins")
	}
	git, err := cleanAbsoluteFile(broker.GitExecutable)
	if err != nil {
		return Pin{}, journal, err
	}
	workRoot, err := cleanAbsoluteRoot(broker.WorkRoot)
	if err != nil {
		return Pin{}, journal, err
	}
	if err = privatedir.MakeAll(workRoot); err != nil {
		return Pin{}, journal, err
	}
	prefix, value, ok := strings.Cut(dependency.Requirement, ":")
	if !ok || value == "" {
		return Pin{}, journal, failFields(CodeDependencyOriginUnsupported, map[string]string{"identity": dependency.Identity}, "source-control requirement is unsupported")
	}
	pin := Pin{Identity: dependency.Identity, Kind: dependency.Kind, RawLocation: dependency.Location, CanonicalLocation: dependency.Location}
	if prefix == "revision" {
		if !validRevision(value) {
			return Pin{}, journal, failFields(CodeDependencyOriginUnsupported, map[string]string{"identity": dependency.Identity}, "revision requirement is mutable or malformed")
		}
		pin.Revision = strings.ToLower(value)
		return pin, journal, nil
	}
	output, err := broker.run(ctx, git, workRoot, "resolve-pin", "broker", &journal, "ls-remote", "--tags", "--heads", "--", dependency.Location)
	if err != nil {
		return Pin{}, journal, err
	}
	refs := parseRemoteRefs(output)
	switch prefix {
	case "exact":
		pin.Revision = resolvedTag(refs, value)
		pin.Version = value
	case "branch":
		pin.Revision = refs["refs/heads/"+value]
		pin.Branch = value
	case "range":
		lower, upper, found := strings.Cut(value, "..<")
		if !found {
			return Pin{}, journal, failFields(CodeDependencyOriginUnsupported, map[string]string{"identity": dependency.Identity}, "range requirement is malformed")
		}
		pin.Version, pin.Revision = highestTagInRange(refs, lower, upper)
	default:
		return Pin{}, journal, failFields(CodeDependencyOriginUnsupported, map[string]string{"identity": dependency.Identity}, "source-control requirement kind is unsupported")
	}
	if !validRevision(pin.Revision) {
		return Pin{}, journal, failFields(CodeResolutionUnfrozen, map[string]string{"identity": dependency.Identity}, "source-control requirement resolved to no immutable commit")
	}
	return pin, journal, nil
}

// Acquire implements AcquisitionBroker without executing checkout hooks,
// filters, submodules, or package code.
func (broker *GitBroker) Acquire(ctx context.Context, pin Pin) (Snapshot, error) {
	if broker == nil {
		return Snapshot{}, fmt.Errorf("git broker is absent")
	}
	workRoot, err := cleanAbsoluteRoot(broker.WorkRoot)
	if err != nil {
		return Snapshot{}, err
	}
	git, err := cleanAbsoluteFile(broker.GitExecutable)
	if err != nil {
		return Snapshot{}, err
	}
	journal := GitVerificationEvidence{}
	if err = privatedir.MakeAll(workRoot); err != nil {
		return Snapshot{}, err
	}
	runRoot, err := os.MkdirTemp(workRoot, "acquire-"+pin.Identity+"-")
	if err != nil {
		return Snapshot{}, err
	}
	mirrorRoot := filepath.Join(runRoot, "mirror.git")
	snapshotRoot := filepath.Join(runRoot, "tree")
	cloneArgs := []string{"-c", "core.hooksPath=/dev/null", "-c", "filter.lfs.smudge=", "clone", "--mirror", "--no-local", "--", pin.CanonicalLocation, mirrorRoot}
	if _, err = broker.run(ctx, git, runRoot, "acquire-clone", "broker", &journal, cloneArgs...); err != nil {
		return Snapshot{}, fmt.Errorf("clone exact origin: %w", err)
	}
	verifier := GitMirrorVerifier{GitExecutable: git, EnvironmentRoot: runRoot, ProcessStartObserver: broker.ProcessStartObserver, authority: broker.authority}
	probe := Snapshot{Identity: pin.Identity, MirrorRoot: mirrorRoot, Revision: pin.Revision, Kind: pin.Kind}
	tree, err := verifier.tree(ctx, mirrorRoot, pin.Revision, &journal)
	if err != nil {
		return Snapshot{}, err
	}
	probe.GitTree = tree
	commitObject, acquisitionReceipt, evidenceErr := broker.runExactEvidence(ctx, runRoot, "capture-commit-object", pin.CanonicalLocation, pin.Revision, tree, &journal, "--git-dir", mirrorRoot, "cat-file", "commit", pin.Revision)
	if evidenceErr != nil {
		return Snapshot{}, fmt.Errorf("capture exact commit object: %w", evidenceErr)
	}
	verified, verifyErr := verifier.Verify(ctx, mirrorRoot, pin, probe)
	if verifyErr != nil {
		return Snapshot{}, verifyErr
	}
	journal.PermitIDs = append(journal.PermitIDs, verified.PermitIDs...)
	journal.ReceiptIDs = append(journal.ReceiptIDs, verified.ReceiptIDs...)
	if err = broker.authority.verify(journal.PermitIDs, journal.ReceiptIDs); err != nil {
		return Snapshot{}, err
	}
	archive, err := broker.run(ctx, git, runRoot, "materialize-archive", "broker", &journal, "--git-dir", mirrorRoot, "archive", "--format=tar", pin.Revision)
	if err != nil {
		return Snapshot{}, fmt.Errorf("archive pinned tree: %w", err)
	}
	if err = extractSourceArchive(snapshotRoot, archive); err != nil {
		return Snapshot{}, err
	}
	if err = scanUnsupportedGitShape(snapshotRoot); err != nil {
		return Snapshot{}, err
	}
	if err = broker.authority.verify(journal.PermitIDs, journal.ReceiptIDs); err != nil {
		return Snapshot{}, err
	}
	receiptID, err := acquisitionReceipt.ID()
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Identity: pin.Identity, Root: snapshotRoot, MirrorRoot: mirrorRoot, Revision: pin.Revision, GitTree: tree, Kind: pin.Kind, BrokerReceiptID: receiptID, AcquisitionReceipt: acquisitionReceipt, CommitObject: append([]byte(nil), commitObject...), BrokerPermitIDs: append([]closuregraph.ID(nil), journal.PermitIDs...), BrokerProcessReceiptIDs: append([]closuregraph.ID(nil), journal.ReceiptIDs...)}, nil
}

// VerifySnapshot proves that the aggregate broker receipt resolves to every
// authority-issued Git process and the exact returned mirror/checkout bytes.
func (broker *GitBroker) VerifySnapshot(pin Pin, snapshot Snapshot) error {
	if broker == nil || broker.authority == nil || broker.authority.verify(snapshot.BrokerPermitIDs, snapshot.BrokerProcessReceiptIDs) != nil {
		return fail(CodeDerivationUnauthorized, "broker snapshot Git journal is not authority-issued")
	}
	_, _, err := inventoryTree(snapshot.MirrorRoot)
	if err != nil {
		return err
	}
	_, _, err = inventoryTree(snapshot.Root)
	if err != nil {
		return err
	}
	receiptID, err := snapshot.AcquisitionReceipt.ID()
	if err != nil || receiptID != snapshot.BrokerReceiptID || snapshot.AcquisitionReceipt.CanonicalOrigin != pin.CanonicalLocation || snapshot.AcquisitionReceipt.RequestedRevision != pin.Revision || snapshot.AcquisitionReceipt.Observation.ResolvedRevision != pin.Revision || snapshot.AcquisitionReceipt.Observation.GitTree != snapshot.GitTree || len(snapshot.CommitObject) == 0 {
		return fail(CodeDerivationUnauthorized, "broker snapshot receipt differs from exact process or byte evidence")
	}
	return nil
}

func (broker *GitBroker) run(ctx context.Context, executable, cwd, phase, network string, journal *GitVerificationEvidence, args ...string) ([]byte, error) {
	resolvedExecutable, _ := filepath.EvalSymlinks(executable)
	if network != "broker" || broker.authority == nil || resolvedExecutable != broker.authority.executable {
		return nil, fail(CodeDerivationUnauthorized, "Git acquisition did not use the shared exact-origin broker")
	}
	output, permitID, receiptID, err := broker.authority.acquire(ctx, cwd, phase, args, "", gitRevisionFromArgs(args), "", broker.ProcessStartObserver)
	if permitID.Valid() {
		journal.PermitIDs = append(journal.PermitIDs, permitID)
	}
	if receiptID.Valid() {
		journal.ReceiptIDs = append(journal.ReceiptIDs, receiptID)
	}
	return output, err
}

func (broker *GitBroker) runExactEvidence(ctx context.Context, cwd, phase, origin, revision, gitTree string, journal *GitVerificationEvidence, args ...string) ([]byte, closureexec.SourceAcquisitionReceipt, error) {
	output, permitID, receiptID, err := broker.authority.acquire(ctx, cwd, phase, args, origin, revision, gitTree, broker.ProcessStartObserver)
	if permitID.Valid() {
		journal.PermitIDs = append(journal.PermitIDs, permitID)
	}
	if receiptID.Valid() {
		journal.ReceiptIDs = append(journal.ReceiptIDs, receiptID)
	}
	if err != nil {
		return nil, closureexec.SourceAcquisitionReceipt{}, err
	}
	receipt, ok := broker.authority.acquisitionReceipts[receiptID]
	if !ok {
		return nil, closureexec.SourceAcquisitionReceipt{}, fail(CodeDerivationUnauthorized, "exact acquisition receipt is absent")
	}
	return output, receipt, nil
}

// GitMirrorVerifier is the production exact-object verifier used both after
// acquisition and immediately before offline replay.
type GitMirrorVerifier struct {
	GitExecutable         string
	EnvironmentRoot       string
	ProcessStartObserver  func([]string)
	ProcessLaunchObserver closureexec.ProcessLaunchObserver
	authority             *sharedGitAuthority
}

func (verifier *GitMirrorVerifier) bindGitAuthority(authority *sharedGitAuthority) error {
	if verifier == nil || authority == nil {
		return fail(CodeDerivationUnauthorized, "Git verifier executable differs from C0 Git")
	}
	resolved, _ := filepath.EvalSymlinks(verifier.GitExecutable)
	if resolved != authority.executable {
		return fail(CodeDerivationUnauthorized, "Git verifier executable differs from C0 Git")
	}
	verifier.authority = authority
	if verifier.ProcessLaunchObserver != nil {
		authority.derivationRunner.ProcessLaunchObserver = verifier.ProcessLaunchObserver
	}
	return nil
}

// Verify implements MirrorVerifier.
func (verifier GitMirrorVerifier) Verify(ctx context.Context, mirrorRoot string, pin Pin, snapshot Snapshot) (GitVerificationEvidence, error) {
	journal := GitVerificationEvidence{}
	git, err := cleanAbsoluteFile(verifier.GitExecutable)
	if err != nil {
		return journal, err
	}
	bare, err := verifier.run(ctx, git, mirrorRoot, "verify-bare", &journal, "--git-dir", mirrorRoot, "rev-parse", "--is-bare-repository")
	if err != nil || strings.TrimSpace(string(bare)) != "true" {
		return journal, fmt.Errorf("mirror is not a bare repository")
	}
	if _, err = verifier.run(ctx, git, mirrorRoot, "verify-commit", &journal, "--git-dir", mirrorRoot, "cat-file", "-e", pin.Revision+"^{commit}"); err != nil {
		return journal, fmt.Errorf("pinned commit is absent: %w", err)
	}
	tree, err := verifier.tree(ctx, mirrorRoot, pin.Revision, &journal)
	if err != nil {
		return journal, err
	}
	if tree != strings.ToLower(snapshot.GitTree) {
		return journal, fmt.Errorf("pinned tree differs: %s != %s", tree, snapshot.GitTree)
	}
	if _, err = verifier.run(ctx, git, mirrorRoot, "verify-object-graph", &journal, "--git-dir", mirrorRoot, "fsck", "--full", "--no-reflogs"); err != nil {
		return journal, fmt.Errorf("mirror object graph is incomplete: %w", err)
	}
	if err = verifier.authority.verify(journal.PermitIDs, journal.ReceiptIDs); err != nil {
		return journal, err
	}
	return journal, nil
}

// VerifyAdmitted replays the authorized mirror from its immutable intake
// handle and runs the exact C0 Git executable only through ordinary C4
// network-none derivation permits. The adapter never receives direct process
// authority for the protected mirror path.
func (verifier GitMirrorVerifier) VerifyAdmitted(ctx context.Context, pin Pin, snapshot Snapshot, input closureexec.AdmittedInput) (GitVerificationEvidence, error) {
	journal := GitVerificationEvidence{}
	if snapshot.Revision != pin.Revision || !validRevision(snapshot.GitTree) {
		return journal, fmt.Errorf("snapshot identity is invalid before admitted replay")
	}
	if _, err := verifier.runAdmitted(ctx, pin, input, "verify-object-graph", &journal, "fsck", "--full", "--no-reflogs"); err != nil {
		return journal, fmt.Errorf("mirror object graph is incomplete: %w", err)
	}
	if err := verifier.authority.verifyDerivations(journal.PermitIDs, journal.ReceiptIDs); err != nil {
		return journal, err
	}
	return journal, nil
}

func (verifier GitMirrorVerifier) runAdmitted(ctx context.Context, pin Pin, input closureexec.AdmittedInput, phase string, journal *GitVerificationEvidence, commandArgs ...string) ([]byte, error) {
	if verifier.authority == nil || verifier.authority.derivation == nil {
		return nil, fail(CodeDerivationUnauthorized, "Git verifier is not bound to shared derivation authority")
	}
	authority := verifier.authority
	authority.mu.Lock()
	defer authority.mu.Unlock()
	authority.sequence++
	receiptID, err := input.Receipt.ID()
	if err != nil {
		return nil, err
	}
	mount := filepath.ToSlash(filepath.Join(".swiftpm-authority", "replay", pin.Identity+"-"+pin.Revision+".git"))
	cwd := filepath.ToSlash(filepath.Join(".swiftpm-authority", "work"))
	evidencePath := filepath.ToSlash(filepath.Join(".swiftpm-authority", "verification-evidence", fmt.Sprintf("%06d-%s.txt", authority.sequence, phase)))
	manifestID, _ := closuregraph.DomainID("swiftpm-git-verification-evidence-v1", map[string]any{"phase": phase, "schema_id": "swiftpm-git-verification-output-v1"})
	expected := []closureexec.EvidenceRequirement{{Path: evidencePath, SchemaID: "swiftpm-git-verification-output-v1", ArtifactManifestID: manifestID}}
	limits := closureexec.ResourceLimits{OutputBytes: 16 << 20, ReadBytes: 1 << 30, WriteBytes: 32 << 20, WallTimeMillis: 120000, ProcessCount: 1}
	limitID, _ := limits.ID()
	evidenceID, _ := closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": []any{map[string]any{"artifact_manifest_id": string(manifestID), "path": evidencePath, "schema_id": "swiftpm-git-verification-output-v1"}}})
	environment := gitEnvironmentMap(authority.executionRoot, authority.executable)
	environment["CURATOR_OUTPUT_ROOT"] = filepath.Join(authority.executionRoot, ".swiftpm-authority", "verification-output")
	environment["CURATOR_EVIDENCE_ROOT"] = authority.executionRoot
	environment["GIT_OPTIONAL_LOCKS"] = "0"
	relativeMirror, _ := filepath.Rel(filepath.Join(authority.executionRoot, filepath.FromSlash(cwd)), filepath.Join(authority.executionRoot, filepath.FromSlash(mount)))
	argv := append([]string{"--git-dir", filepath.ToSlash(relativeMirror)}, commandArgs...)
	permit := closureexec.DerivationPermit{SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: authority.derivation.CurrentCausalHead(), InvocationKey: "swiftpm-git-mirror-verification-v1:" + phase + ":" + pin.Identity, InvocationSubtype: closureexec.DerivationMetadata, AdmittedInputReceiptIDs: []closuregraph.ID{receiptID}, InputMounts: []closureexec.InputMount{{ReceiptID: receiptID, Path: mount}}, C0CheckpointID: authority.c0ID, ToolchainNodeID: mustToolNodeID(authority.toolchain.Git), ToolchainFingerprint: authority.toolchain.Git.Fingerprint, ExecutableSHA256: authority.toolchain.Git.ExecutableSHA256, Executable: authority.toolchain.Git.ExecutableRelativePath, CWD: cwd, Argv: argv, Environment: environment, HostID: mustPlatformID(authority.toolchain.Git.PlatformABI), TargetID: mustPlatformID(authority.toolchain.Git.PlatformABI), AllowedProcesses: append([]string(nil), authority.allowedProcesses...), ReadRoots: sortedUniqueStrings([]string{filepath.ToSlash(filepath.Dir(authority.toolchain.Git.ExecutableRelativePath)), mount}), WriteRoots: []string{evidencePath}, StdoutEvidencePath: evidencePath, ExpectedEvidence: expected, Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID}
	permitID, err := authority.derivation.Commit(permit)
	if err != nil {
		return nil, err
	}
	issuedPermit, err := authority.derivation.IssuedDerivationPermit(permitID)
	if err != nil {
		return nil, err
	}
	receipt, err := authority.derivation.Execute(ctx, permitID, func(checkCtx context.Context) (closureexec.ToolchainIdentity, error) {
		return authority.recheck(checkCtx)
	}, map[closuregraph.ID]closureexec.AdmittedInput{receiptID: input})
	if err != nil {
		return nil, err
	}
	derivationReceiptID, _ := receipt.ID()
	authority.derivationPermits[permitID] = issuedPermit
	authority.derivationReceipts[derivationReceiptID] = receipt
	journal.PermitIDs = append(journal.PermitIDs, permitID)
	journal.ReceiptIDs = append(journal.ReceiptIDs, derivationReceiptID)
	if len(receipt.Outputs) != 1 {
		return nil, fail(CodeDerivationUnauthorized, "Git verification output receipt is incomplete")
	}
	verificationRoot := filepath.Join(authority.executionRoot, ".swiftpm-authority", "verification-output")
	output, readErr := os.ReadFile(filepath.Join(verificationRoot, filepath.FromSlash(evidencePath))) // #nosec G304 -- exact receipted output path.
	removeErr := os.RemoveAll(verificationRoot)
	if readErr != nil {
		return nil, readErr
	}
	return output, removeErr
}

func (verifier GitMirrorVerifier) tree(ctx context.Context, mirrorRoot, revision string, journal *GitVerificationEvidence) (string, error) {
	git, err := cleanAbsoluteFile(verifier.GitExecutable)
	if err != nil {
		return "", err
	}
	output, err := verifier.run(ctx, git, mirrorRoot, "verify-tree", journal, "--git-dir", mirrorRoot, "show", "-s", "--format=%T", revision)
	if err != nil {
		return "", err
	}
	tree := strings.ToLower(strings.TrimSpace(string(output)))
	if !validRevision(tree) {
		return "", fmt.Errorf("git tree identity is invalid")
	}
	return tree, nil
}

func (verifier GitMirrorVerifier) run(ctx context.Context, executable, cwd, phase string, journal *GitVerificationEvidence, args ...string) ([]byte, error) {
	resolvedExecutable, _ := filepath.EvalSymlinks(executable)
	if verifier.authority == nil || resolvedExecutable != verifier.authority.executable {
		return nil, fail(CodeDerivationUnauthorized, "Git verifier is not bound to shared C0 authority")
	}
	output, permitID, receiptID, err := verifier.authority.acquire(ctx, cwd, phase, args, "", gitRevisionFromArgs(args), "", verifier.ProcessStartObserver)
	if permitID.Valid() {
		journal.PermitIDs = append(journal.PermitIDs, permitID)
	}
	if receiptID.Valid() {
		journal.ReceiptIDs = append(journal.ReceiptIDs, receiptID)
	}
	return output, err
}

func extractSourceArchive(root string, payload []byte) error {
	if err := privatedir.Make(root); err != nil {
		return err
	}
	reader := tar.NewReader(bytes.NewReader(payload))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := path.Clean(header.Name)
		if name == "." || path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") {
			return fail(CodeDependencyOriginUnsupported, "Git archive contains an unsafe path")
		}
		destination := filepath.Join(root, filepath.FromSlash(name))
		relative, relErr := filepath.Rel(root, destination)
		if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fail(CodeDependencyOriginUnsupported, "Git archive path escapes snapshot")
		}
		switch header.Typeflag {
		case tar.TypeXHeader, tar.TypeXGlobalHeader:
			continue
		case tar.TypeDir:
			if err = privatedir.MakeAll(destination); err != nil {
				return err
			}
		case tar.TypeReg:
			if err = privatedir.MakeAll(filepath.Dir(destination)); err != nil {
				return err
			}
			mode := fs.FileMode(0o600)
			if header.FileInfo().Mode().Perm()&0o111 != 0 {
				mode = 0o700
			}
			file, openErr := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode) // #nosec G304 -- destination passed containment checks.
			if openErr != nil {
				return openErr
			}
			_, copyErr := io.CopyN(file, reader, header.Size)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			return fail(CodeDependencyOriginUnsupported, "Git archive contains a linked or special node")
		}
	}
}

func gitEnvironment(root, executable string) []string {
	home := filepath.Join(root, "empty-home")
	_ = privatedir.MakeAll(home)
	return []string{"GIT_CONFIG_NOSYSTEM=1", "GIT_LFS_SKIP_SMUDGE=1", "GIT_TERMINAL_PROMPT=0", "HOME=" + home, "LANG=C", "LC_ALL=C", "PATH=" + filepath.Dir(executable), "TZ=UTC"}
}

func gitEnvironmentMap(root, executable string) map[string]string {
	values := map[string]string{}
	for _, entry := range gitEnvironment(root, executable) {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			values[key] = value
		}
	}
	return values
}

func pathWithinRoot(root, candidate string) bool {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(candidate))
	return err == nil && relative != ".." && !filepath.IsAbs(relative) && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func cleanAbsoluteFile(value string) (string, error) {
	absolute, err := filepath.Abs(value)
	if err != nil || value == "" || absolute != filepath.Clean(value) {
		return "", fmt.Errorf("executable path must be absolute and clean")
	}
	info, err := os.Lstat(absolute)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return "", fmt.Errorf("executable is absent, linked, or non-regular")
	}
	return absolute, nil
}

func mirrorRootFromArgs(args []string) string {
	for index, value := range args {
		if value == "--git-dir" && index+1 < len(args) {
			return args[index+1]
		}
	}
	return "."
}

func parseRemoteRefs(payload []byte) map[string]string {
	refs := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(payload)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && validRevision(fields[0]) {
			refs[fields[1]] = strings.ToLower(fields[0])
		}
	}
	return refs
}

func resolvedTag(refs map[string]string, version string) string {
	for _, tag := range []string{version, "v" + version} {
		if peeled := refs["refs/tags/"+tag+"^{}"]; peeled != "" {
			return peeled
		}
		if revision := refs["refs/tags/"+tag]; revision != "" {
			return revision
		}
	}
	return ""
}

func highestTagInRange(refs map[string]string, lower, upper string) (string, string) {
	lowerVersion, lowerOK := parseVersion(lower)
	upperVersion, upperOK := parseVersion(upper)
	if !lowerOK || !upperOK {
		return "", ""
	}
	selected := [3]int{-1, -1, -1}
	selectedText, selectedRevision := "", ""
	for ref := range refs {
		if !strings.HasPrefix(ref, "refs/tags/") || strings.HasSuffix(ref, "^{}") {
			continue
		}
		name := strings.TrimPrefix(ref, "refs/tags/")
		versionText := strings.TrimPrefix(name, "v")
		version, ok := parseVersion(versionText)
		if !ok || compareVersion(version, lowerVersion) < 0 || compareVersion(version, upperVersion) >= 0 || compareVersion(version, selected) <= 0 {
			continue
		}
		selected, selectedText, selectedRevision = version, versionText, resolvedTag(refs, versionText)
	}
	return selectedText, selectedRevision
}
