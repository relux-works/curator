package yarnmodernsource

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

// invocation is a non-authoritative projection used to construct the canonical
// shared executor permit.
type invocation struct {
	Tool        nodesource.ToolIdentity
	CWD         string
	Args        []string
	Environment map[string]string
	ReadRoots   []string
	WriteRoots  []string
	Inputs      map[closuregraph.ID]closureexec.AdmittedInput
	InputMounts map[closuregraph.ID]string
	WorkCopies  []closureexec.WorkCopy
	Template    map[string][]string
}

// ExecutionContext supplies common executor and C0/C5 authority. Process
// evidence can only come back as an executor-issued DerivationReceipt.
type ExecutionContext struct {
	Executor  *closureexec.Executor
	Selection closuregraph.SelectionContext
	Runtime   nodesource.RuntimeBinding
	BuildPlan closuregraph.BuildPlan
	Recheck   func(context.Context, nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error)
	// ExecutionRoot is the task-private physical namespace used by the common
	// executor for the logical paths committed in permits.
	ExecutionRoot string
}

type runtimeAction struct {
	ID      closuregraph.ID
	Payload closuregraph.ActionPayload
}

type runnerSession struct {
	authority *ExecutionContext
	operation *closureexec.AssuredOperation
	binding   closureexec.AssuranceBinding
	bundle    closuregraph.GraphBundle
	c0ID      closuregraph.ID
	planID    closuregraph.ID
	actions   map[string]runtimeAction
	inputs    map[closuregraph.ID]closureexec.AdmittedInput
}

// MirrorFile is one deterministic member of a derived private source mirror.
type MirrorFile struct {
	Path   string
	SHA256 closuregraph.ID
	Size   int64
}

// MirrorReceipt proves mirror bytes were derived only from admitted tarballs.
type MirrorReceipt struct {
	SchemaID        string
	InputReceiptIDs []closuregraph.ID
	Files           []MirrorFile
	ID              closuregraph.ID
}

// PrivateMirror is receipted derived state and is never source authority.
type PrivateMirror struct {
	Receipt MirrorReceipt
	root    string
	capture *Capture
	input   closureexec.AdmittedInput
	inputID closuregraph.ID
}

// MaterializeRequest selects an absent project destination and task-private scratch root.
type MaterializeRequest struct{ Destination, WorkRoot string }

// Materialization records the exact scripts-disabled offline replay result.
type Materialization struct {
	Root                 string
	MirrorReceiptID      closuregraph.ID
	Receipt              closureexec.DerivationReceipt
	MaterializedPackages []string
	capture              *Capture
	input                closureexec.AdmittedInput
	inputID              closuregraph.ID
}

type pnpLocator struct {
	Name      string `json:"name"`
	Reference string `json:"reference"`
}

type pnpRuntimeState struct {
	DependencyTreeRoots []pnpLocator      `json:"dependencyTreeRoots"`
	PackageRegistryData []json.RawMessage `json:"packageRegistryData"`
}

type pnpPackageInformation struct {
	PackageLocation     string            `json:"packageLocation"`
	PackageDependencies []json.RawMessage `json:"packageDependencies"`
	PackagePeers        []string          `json:"packagePeers"`
}

type pnpObservedPackage struct {
	Location     string
	Dependencies map[string]string
	Peers        []string
}

// CacheFile is one deterministic member of the immutable private cache input.
type CacheFile = MirrorFile

// CacheReceipt is the canonical private-cache derivation receipt.
type CacheReceipt = MirrorReceipt

// PrivateCache is immutable derived state accepted by Materialize.
type PrivateCache = PrivateMirror

// BuildPrivateCache deterministically derives the only manager-visible cache
// from admitted archives. Ambient cache, PnP, install-state, and node_modules
// bytes never participate.
func BuildPrivateCache(ctx context.Context, capture *Capture, destination, workRoot string, authority *ExecutionContext) (*PrivateCache, error) {
	return DerivePrivateMirror(ctx, capture, destination, workRoot, authority)
}

// DerivePrivateMirror is the compatibility spelling for BuildPrivateCache.
func DerivePrivateMirror(ctx context.Context, capture *Capture, destination, workRoot string, authority *ExecutionContext) (*PrivateMirror, error) {
	if capture == nil || capture.project.tree == nil || capture.store == nil || authority == nil {
		return nil, fail(CodeInputUndeclared, "Yarn mirror derivation authority is incomplete", nil)
	}
	session, err := newRunnerSession(ctx, capture, authority)
	if err != nil {
		return nil, err
	}
	dest, err := cleanAbsentAbsolute(destination)
	if err != nil {
		return nil, err
	}
	if _, err = filepath.Abs(workRoot); err != nil || workRoot == "" {
		return nil, fail(CodeInputUndeclared, "Yarn mirror work root is invalid", nil)
	}
	if err = privatedir.Make(dest); err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(dest)
		}
	}()
	projectSource, err := capture.project.tree.ProtectedPath()
	if err != nil {
		return nil, err
	}
	if err = copyContainedTreeDereferencingLinks(projectSource, projectSource, dest, map[string]bool{}); err != nil {
		return nil, err
	}
	cacheDir := filepath.Join(dest, ".yarn", "cache")
	if err = privatedir.MakeAll(cacheDir); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(capture.tarballs))
	for key := range capture.tarballs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	usedNames := map[string]string{}
	for _, key := range keys {
		item := capture.tarballs[key]
		if err = item.handle.Recheck(); err != nil {
			return nil, fail(CodeIntegrityMismatch, "captured Yarn tarball changed before mirror derivation", map[string]string{"package": key})
		}
		name := item.cacheName
		if path.Base(name) != name || !strings.HasSuffix(strings.ToLower(name), ".zip") {
			return nil, fail(CodeIntegrityMismatch, "captured Yarn cache member name is invalid", map[string]string{"package": key})
		}
		if prior := usedNames[name]; prior != "" && prior != key {
			return nil, fail(CodeGraphIncomplete, "Yarn offline mirror filename collision", map[string]string{"filename": name, "first": prior, "second": key})
		}
		usedNames[name] = key
		output, createErr := os.OpenFile(filepath.Join(cacheDir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- validated cache basename.
		if createErr != nil {
			return nil, createErr
		}
		_, copyErr := output.Write(item.cacheBytes)
		closeErr := output.Close()
		if copyErr != nil {
			return nil, copyErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		if err = item.handle.Recheck(); err != nil {
			return nil, fail(CodeIntegrityMismatch, "captured Yarn tarball changed during mirror derivation", map[string]string{"package": key})
		}
	}
	files, err := inventoryFiles(cacheDir)
	if err != nil {
		return nil, err
	}
	if len(keys) > 0 && len(files) != len(keys) {
		return nil, fail(CodeOfflineInputMissing, "Yarn private cache derivation is incomplete", nil)
	}
	receipts := []closuregraph.ID{capture.project.receiptID}
	for _, key := range keys {
		receipts = append(receipts, capture.tarballs[key].receiptID)
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i] < receipts[j] })
	payloadFiles := make([]any, len(files))
	for i, file := range files {
		payloadFiles[i] = map[string]any{"path": file.Path, "sha256": string(file.SHA256), "size": file.Size}
	}
	id, err := closuregraph.DomainID("yarn-modern-private-cache-receipt-v1", map[string]any{"files": payloadFiles, "input_receipt_ids": idsToStrings(receipts), "schema_id": "yarn-modern-private-cache-receipt-v1", "cache_key": capture.Graph.Layout.CacheKey, "compression": capture.Graph.Layout.CompressionLevel, "linker": capture.Graph.Layout.NodeLinker})
	if err != nil {
		return nil, err
	}
	if err = makeTreeReadOnly(dest); err != nil {
		return nil, err
	}
	tree, err := capture.store.CaptureTree("derived:yarn-modern-cache:"+string(id), dest)
	if err != nil {
		return nil, err
	}
	derivedReceipt, err := capture.store.AdmitTree(tree, "derived:yarn-modern-cache:"+string(id), closureexec.AdmissionEvidence{PreviousCausalHead: session.operation.CurrentCausalHead(), ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: "yarn-modern-private-cache-v1", DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: id})
	if err != nil {
		return nil, err
	}
	derivedID, err := derivedReceipt.ID()
	if err != nil {
		return nil, err
	}
	success = true
	return &PrivateMirror{Receipt: MirrorReceipt{SchemaID: "yarn-modern-private-cache-receipt-v1", InputReceiptIDs: receipts, Files: files, ID: id}, root: dest, capture: capture, input: closureexec.AdmittedInput{Receipt: derivedReceipt, Tree: tree}, inputID: derivedID}, nil
}

// Materialize copies the admitted source snapshot, starts with an admitted
// empty ordinary cache, and runs frozen Yarn offline with lifecycle scripts
// disabled. It rejects missing, extra, and substituted installed packages.
func Materialize(ctx context.Context, mirror *PrivateMirror, request MaterializeRequest, authority *ExecutionContext) (*Materialization, error) {
	if mirror == nil || mirror.capture == nil || mirror.input.Tree == nil || authority == nil {
		return nil, fail(CodeInputUndeclared, "Yarn materialization authority is incomplete", nil)
	}
	session, err := newRunnerSession(ctx, mirror.capture, authority)
	if err != nil {
		return nil, err
	}
	destination, err := cleanAbsentAbsolute(request.Destination)
	if err != nil {
		return nil, err
	}
	if _, err = filepath.Abs(request.WorkRoot); err != nil || request.WorkRoot == "" {
		return nil, fail(CodeInputUndeclared, "Yarn materialization work root is invalid", nil)
	}
	if err = mirror.capture.project.tree.VerifyAtUse(); err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(destination)
		}
	}()
	observed, err := inventoryFiles(filepath.Join(mirror.root, ".yarn", "cache"))
	if err != nil {
		return nil, err
	}
	if len(observed) < len(mirror.Receipt.Files) {
		return nil, fail(CodeOfflineInputMissing, "private Yarn mirror is missing a receipted member", nil)
	}
	if !equalMirrorFiles(observed, mirror.Receipt.Files) {
		return nil, fail(CodeIntegrityMismatch, "private Yarn mirror differs from its derivation receipt", nil)
	}
	plan, err := nodesource.PlanFreshMaterialization(mirror.capture.Evidence.DiscardedDerivedPaths)
	if err != nil {
		return nil, err
	}
	if plan.LifecycleMode != "disabled" || plan.AmbientAuthority || !plan.RequiresDerivationReceipt {
		return nil, fail(CodeInputUndeclared, "common Node materialization plan is not closed", nil)
	}
	projectPath := "work/yarn-modern-project"
	cacheInputPath := "capture/yarn-modern-private-replay"
	homePath := ".curator-home"
	relativeCache := ".yarn/cache"
	args := []string{"install", "--immutable", "--immutable-cache", "--mode=skip-build"}
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{mirror.inputID: mirror.input}
	works := []closureexec.WorkCopy{{ReceiptID: mirror.inputID, Path: projectPath, Retain: true}}
	invocation := invocation{
		Tool: authority.Runtime.Manager, CWD: projectPath, Args: args,
		Environment: yarnEnvironment(homePath, relativeCache), Inputs: inputs,
		InputMounts: map[closuregraph.ID]string{mirror.inputID: cacheInputPath},
		ReadRoots:   append([]string{cacheInputPath}, authority.Runtime.Manager.ReadRoots...), WriteRoots: []string{projectPath}, WorkCopies: works,
		Template: map[string][]string{"manager_entrypoint": {relativeLogical(projectPath, authority.Runtime.Manager.EntrypointRelativePath)}, "project": {projectPath}},
	}
	invocation.Args = append([]string{invocation.Template["manager_entrypoint"][0]}, invocation.Args...)
	receipt, runErr := session.run(ctx, invocation, "yarn-install")
	if runErr != nil {
		return nil, runErr
	}
	retained := filepath.Join(authority.ExecutionRoot, filepath.FromSlash(projectPath))
	defer func() { _ = os.RemoveAll(retained) }()
	// Yarn writes task-local logs below the private HOME. They are
	// disposable manager evidence, not source materialization output, and carry
	// wall-clock names/content. Remove the closed scratch root before admitting
	// and comparing the materialized project tree.
	if err = os.RemoveAll(filepath.Join(retained, filepath.FromSlash(homePath))); err != nil {
		return nil, err
	}
	cacheFiles, cacheErr := inventoryFiles(filepath.Join(retained, ".yarn", "cache"))
	if cacheErr != nil || !equalMirrorFiles(cacheFiles, mirror.Receipt.Files) {
		return nil, fail(CodeIntegrityMismatch, "Yarn changed immutable private cache", nil)
	}
	lockBytes, readErr := os.ReadFile(filepath.Join(retained, mirror.capture.Graph.LockName)) // #nosec G304 -- fixed root yarn.lock path below replay root.
	if readErr != nil || digestID(lockBytes) != closuregraph.ID(mirror.capture.Graph.RawLockSHA256) {
		return nil, fail(CodeLockStale, "Yarn changed or removed the frozen lock", nil)
	}
	installed, err := validateInstalledTree(ctx, retained, mirror.capture)
	if err != nil {
		return nil, err
	}
	if err = mirror.capture.project.tree.VerifyAtUse(); err != nil {
		return nil, err
	}
	if err = copyContainedTreeDereferencingLinks(retained, retained, destination, map[string]bool{}); err != nil {
		return nil, err
	}
	materializedInput, materializedID, err := admitMaterializedTree(mirror.capture, destination, receipt)
	if err != nil {
		return nil, err
	}
	success = true
	return &Materialization{Root: destination, MirrorReceiptID: mirror.Receipt.ID, Receipt: receipt, MaterializedPackages: installed, capture: mirror.capture, input: materializedInput, inputID: materializedID}, nil
}

// Invoke runs an admitted materialized JavaScript entry point through the
// protected Node runner while retaining the same zero-network/zero-ambient gate.
func Invoke(ctx context.Context, materialized *Materialization, entrypoint string, args []string, authority *ExecutionContext) (closureexec.DerivationReceipt, error) {
	if materialized == nil || materialized.capture == nil || authority == nil {
		return closureexec.DerivationReceipt{}, fail(CodeInputUndeclared, "Yarn invocation authority is incomplete", nil)
	}
	session, err := newRunnerSession(ctx, materialized.capture, authority)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	if entrypoint == "" || filepath.IsAbs(entrypoint) || filepath.Clean(entrypoint) != entrypoint || strings.HasPrefix(entrypoint, ".."+string(filepath.Separator)) {
		return closureexec.DerivationReceipt{}, fail(CodeLocalPathEscape, "Node entry point escapes materialized closure", map[string]string{"path": entrypoint})
	}
	entry := filepath.Join(materialized.Root, entrypoint)
	info, err := os.Lstat(entry)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return closureexec.DerivationReceipt{}, fail(CodeInputUndeclared, "Node entry point is absent or not regular", map[string]string{"path": entrypoint})
	}
	logicalRoot := "materialized"
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{materialized.inputID: materialized.input}
	invokeArgs := append([]string{entrypoint}, args...)
	template := map[string][]string{"entrypoint": {entrypoint}, "args": args, "project": {logicalRoot}}
	if materialized.capture.Graph.Layout.NodeLinker == "pnp" {
		loader := "./.pnp.cjs"
		invokeArgs = append([]string{"--require", loader}, invokeArgs...)
		template["pnp_loader"] = []string{loader}
	}
	invocation := invocation{Tool: authority.Runtime.Node, CWD: logicalRoot, Args: invokeArgs, Environment: map[string]string{"HOME": "home", "NO_PROXY": "*", "no_proxy": "*"}, ReadRoots: append([]string{logicalRoot}, authority.Runtime.Node.ReadRoots...), WriteRoots: []string{}, Inputs: inputs, InputMounts: map[closuregraph.ID]string{materialized.inputID: logicalRoot}, Template: template}
	receipt, runErr := session.run(ctx, invocation, "node-invoke")
	if runErr != nil {
		return closureexec.DerivationReceipt{}, runErr
	}
	return receipt, nil
}

func validateExecutionContext(authority *ExecutionContext) error {
	if authority == nil || authority.Executor == nil || authority.Recheck == nil {
		return fail(CodeDerivationUnauthorized, "Yarn shared executor authority is incomplete", nil)
	}
	if authority.Runtime.C0Checkpoint == nil {
		return fail(CodeDerivationUnauthorized, "Yarn runner lacks exact C0 Node/manager authority", nil)
	}
	if authority.Runtime.Manager.Role == "" || authority.Runtime.Node.Role == "" {
		return fail(CodeDerivationUnauthorized, "Yarn runner tool identities are incomplete", nil)
	}
	if authority.Runtime.Manager.ExecutableRelativePath != authority.Runtime.Node.ExecutableRelativePath || authority.Runtime.Manager.ExecutableSHA256 != authority.Runtime.Node.ExecutableSHA256 || authority.Runtime.Manager.EntrypointRelativePath == "" {
		return fail(CodeDerivationUnauthorized, "Yarn must execute its exact bound CLI through the exact bound Node runtime", nil)
	}
	if strings.TrimSpace(authority.Runtime.Manager.VersionOutput) != SupportedYarnVersion {
		return fail(CodeRuntimeIdentityChanged, "bound manager is not the exact supported modern Yarn release", map[string]string{"expected": SupportedYarnVersion, "observed": strings.TrimSpace(authority.Runtime.Manager.VersionOutput)})
	}
	executionRoot, err := filepath.Abs(authority.ExecutionRoot)
	if err != nil || authority.ExecutionRoot == "" || executionRoot != filepath.Clean(authority.ExecutionRoot) {
		return fail(CodeDerivationUnauthorized, "Yarn task-private execution root is invalid", nil)
	}
	rootInfo, err := os.Lstat(executionRoot)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&fs.ModeSymlink != 0 {
		return fail(CodeDerivationUnauthorized, "Yarn task-private execution root is not a real directory", nil)
	}
	return nil
}

func newRunnerSession(ctx context.Context, capture *Capture, authority *ExecutionContext) (*runnerSession, error) {
	if capture == nil || authority == nil {
		return nil, fail(CodeDerivationUnauthorized, "Yarn execution session authority is absent", nil)
	}
	if err := validateExecutionContext(authority); err != nil {
		return nil, err
	}
	operation, err := authority.Executor.Preflight(ctx)
	if err != nil {
		return nil, fail(CodeDerivationUnauthorized, "Yarn executor assurance preflight failed: "+err.Error(), nil)
	}
	binding := operation.Binding()
	wantC0, err := nodesource.NewC0Checkpoint(capture.NodeCapture, authority.Selection, authority.Runtime)
	if err != nil {
		return nil, fail(CodeDerivationUnauthorized, "Yarn C0 authority cannot be rederived: "+err.Error(), nil)
	}
	wantC0ID, err := wantC0.ID()
	if err != nil {
		return nil, err
	}
	observedC0ID, err := authority.Runtime.C0Checkpoint.ID()
	if err != nil || observedC0ID != wantC0ID {
		return nil, fail(CodeDerivationUnauthorized, "Yarn runner C0 checkpoint differs from the common Node profile", nil)
	}
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "yarn-modern-platform-v1", EvaluateFunc: evaluateYarnPlatformCondition}
	bundle, plan, err := nodesource.Close(capture.NodeCapture, authority.Selection, authority.Runtime, []closuregraph.ConditionEvaluator{evaluator}, binding.ExecutionPolicyID)
	if err != nil {
		return nil, fail(CodeDerivationUnauthorized, "Yarn common C5 plan cannot be rederived: "+err.Error(), nil)
	}
	wantPlanID, err := plan.ID()
	if err != nil {
		return nil, err
	}
	observedPlanID, err := authority.BuildPlan.ID()
	if err != nil || observedPlanID != wantPlanID {
		return nil, fail(CodeDerivationUnauthorized, "Yarn runner build plan differs from the exact common Node projection", nil)
	}
	actions := map[string]runtimeAction{}
	for _, actionID := range plan.ActionNodeIDs {
		for _, node := range bundle.Records.CaptureNodes {
			nodeID, nodeErr := node.ID()
			if nodeErr != nil || nodeID != actionID || node.Kind != closuregraph.NodeAction {
				continue
			}
			payload := node.Payload.(closuregraph.ActionPayload)
			subtype := payload.ActionSubtype
			if actions[subtype].ID != "" {
				return nil, fail(CodeDerivationUnauthorized, "Yarn C5 plan contains duplicate runtime action subtype", map[string]string{"subtype": subtype})
			}
			actions[subtype] = runtimeAction{ID: actionID, Payload: payload}
		}
	}
	for _, subtype := range []string{"yarn-install", "node-invoke"} {
		if actions[subtype].ID == "" {
			return nil, fail(CodeDerivationUnauthorized, "Yarn operation is absent from C5", map[string]string{"subtype": subtype})
		}
	}
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{capture.project.receiptID: capture.project.input}
	for _, item := range capture.tarballs {
		inputs[item.receiptID] = item.input
	}
	return &runnerSession{authority: authority, operation: operation, binding: binding, bundle: bundle, c0ID: wantC0ID, planID: wantPlanID, actions: actions, inputs: inputs}, nil
}

func evaluateYarnPlatformCondition(condition closuregraph.Condition, input closuregraph.EvaluationInput) (bool, error) {
	matched, _, err := evaluateYarnCondition(condition.Expression, input.Selection.Markers)
	if err != nil {
		return false, fail(CodeGraphIncomplete, "Yarn platform condition is malformed", map[string]string{"condition": condition.Expression})
	}
	return matched, nil
}

func (session *runnerSession) run(ctx context.Context, call invocation, subtype string) (closureexec.DerivationReceipt, error) {
	action := session.actions[subtype]
	if action.ID == "" {
		return closureexec.DerivationReceipt{}, fail(CodeDerivationUnauthorized, "Yarn operation is absent from C5", map[string]string{"subtype": subtype})
	}
	if err := validateConcreteAction(action.Payload, call, subtype); err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	outputRoot := filepath.Join(session.authority.ExecutionRoot, "output")
	if err := os.RemoveAll(outputRoot); err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	if err := session.operation.Revalidate(ctx); err != nil {
		return closureexec.DerivationReceipt{}, fail(CodeDerivationUnauthorized, subtype+" provider binding drifted before process start: "+err.Error(), nil)
	}
	toolNodeID, err := session.toolNodeID(call.Tool, action.ID)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	stdoutPath := "evidence/" + subtype + ".stdout"
	stdoutID, err := closuregraph.DomainID("yarn-modern-operation-stdout-v1", map[string]any{"action_node_id": string(action.ID), "build_plan_id": string(session.planID), "subtype": subtype})
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	requirements := []closureexec.EvidenceRequirement{{Path: stdoutPath, SchemaID: "yarn-modern-operation-stdout-v1", ArtifactManifestID: stdoutID}}
	evidenceID, err := yarnEvidenceSchemaID(requirements)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 64 << 20, ReadBytes: 2 << 30, WriteBytes: 2 << 30, WallTimeMillis: 300_000, ProcessCount: 256}
	limitID, err := limits.ID()
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	ids := make([]closuregraph.ID, 0, len(call.Inputs))
	for id := range call.Inputs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	mounts := make([]closureexec.InputMount, len(ids))
	reads := append([]string(nil), call.ReadRoots...)
	sort.Strings(reads)
	for index, id := range ids {
		path := call.InputMounts[id]
		if path == "" {
			return closureexec.DerivationReceipt{}, fail(CodeDerivationUnauthorized, "Yarn concrete action omits an admitted input mount", nil)
		}
		mounts[index] = closureexec.InputMount{ReceiptID: id, Path: path}
	}
	environment := cloneStrings(call.Environment)
	environment["CURATOR_OUTPUT_ROOT"] = "output"
	concreteID, err := closuregraph.DomainID("yarn-modern-c5-concrete-action-v1", map[string]any{"action_node_id": string(action.ID), "argv": runnerStringsToAny(call.Args), "cwd": call.CWD, "environment": runnerStringMapAny(environment), "read_roots": runnerStringsToAny(reads), "write_roots": runnerStringsToAny(call.WriteRoots)})
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	invocationKey := fmt.Sprintf("yarn-modern-c5:%s:%s:%s:%s", session.planID, action.ID, subtype, concreteID)
	writes := append([]string(nil), call.WriteRoots...)
	writes = append(writes, stdoutPath)
	sort.Strings(writes)
	processes := []string{call.Tool.ExecutableRelativePath}
	permit := closureexec.DerivationPermit{
		SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: session.operation.CurrentCausalHead(), InvocationKey: invocationKey,
		InvocationSubtype: closureexec.DerivationMetadata, AdmittedInputReceiptIDs: ids, InputMounts: mounts,
		C0CheckpointID: session.c0ID, ToolchainNodeID: toolNodeID, ToolchainFingerprint: call.Tool.Fingerprint,
		ExecutableSHA256: call.Tool.ExecutableSHA256, Executable: call.Tool.ExecutableRelativePath, CWD: call.CWD, Argv: append([]string(nil), call.Args...), Environment: environment,
		HostID: session.authority.Selection.PlatformRoles[closuregraph.PlatformTarget], TargetID: session.authority.Selection.PlatformRoles[closuregraph.PlatformTarget],
		AllowedProcesses: processes, ReadRoots: reads, WriteRoots: writes, WorkCopies: append([]closureexec.WorkCopy(nil), call.WorkCopies...), StdoutEvidencePath: stdoutPath, ExpectedEvidence: requirements,
		Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID,
	}
	permitID, err := session.operation.Commit(permit)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	receipt, err := session.operation.Execute(ctx, permitID, func(recheckCtx context.Context) (closureexec.ToolchainIdentity, error) {
		return session.authority.Recheck(recheckCtx, call.Tool)
	}, call.Inputs)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	if err = session.operation.VerifyIssuedDerivationReceipt(receipt); err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	return receipt, nil
}

func (session *runnerSession) toolNodeID(tool nodesource.ToolIdentity, actionID closuregraph.ID) (closuregraph.ID, error) {
	c0Tools := map[closuregraph.ID]bool{}
	for _, id := range session.authority.Runtime.C0Checkpoint.Payload.(closuregraph.C0ProfilePayload).EvidenceToolchainNodeIDs {
		c0Tools[id] = true
	}
	var matched closuregraph.ID
	for _, edge := range session.bundle.Records.BindingEdges {
		if edge.Kind != closuregraph.EdgeUsesTool || edge.FromNodeID != actionID {
			continue
		}
		payload := edge.Payload.(closuregraph.UsesToolPayload)
		if payload.ExecutableRelativePath != tool.ExecutableRelativePath || !c0Tools[edge.ToNodeID] || matched != "" {
			return "", fail(CodeDerivationUnauthorized, "Yarn C5 action tool differs from exact C0 authority", nil)
		}
		matched = edge.ToNodeID
	}
	if matched == "" {
		return "", fail(CodeDerivationUnauthorized, "Yarn C5 action lacks exact C0 tool binding", nil)
	}
	for _, node := range session.bundle.Records.BindingNodes {
		id, err := node.ID()
		if err != nil || id != matched || node.Kind != closuregraph.NodeToolchainComponent {
			continue
		}
		payload := node.Payload.(closuregraph.ToolchainComponentPayload)
		if payload.ContentFingerprint != tool.Fingerprint || len(payload.LinkFingerprintIDs) != 1 || payload.LinkFingerprintIDs[0] != tool.ExecutableSHA256 || payload.TimeOfUseRecheckRule != "immediate-exact-v1" {
			return "", fail(CodeRuntimeIdentityChanged, "Yarn physical tool differs from C0 binding", nil)
		}
		return matched, nil
	}
	return "", fail(CodeDerivationUnauthorized, "Yarn C0 tool node is absent", nil)
}

func yarnEvidenceSchemaID(requirements []closureexec.EvidenceRequirement) (closuregraph.ID, error) {
	values := make([]any, len(requirements))
	for index, requirement := range requirements {
		values[index] = map[string]any{"artifact_manifest_id": string(requirement.ArtifactManifestID), "path": requirement.Path, "schema_id": requirement.SchemaID}
	}
	return closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": values})
}

func cloneStrings(values map[string]string) map[string]string {
	result := make(map[string]string, len(values)+2)
	for key, value := range values {
		result[key] = value
	}
	return result
}

func validateConcreteAction(payload closuregraph.ActionPayload, call invocation, subtype string) error {
	if payload.ActionSubtype != subtype || payload.Network != "none" || len(payload.ArgvTemplate) < 2 || payload.ArgvTemplate[0] != "$TOOL(executor)" {
		return fail(CodeDerivationUnauthorized, "Yarn C5 action contract is incomplete", map[string]string{"subtype": subtype})
	}
	if err := validateActionBindings(subtype, call.Template, call.Tool); err != nil {
		return err
	}
	wantArgs, err := expandActionTemplate(payload.ArgvTemplate[1:], call.Template)
	if err != nil || !equalStrings(wantArgs, call.Args) {
		return fail(CodeDerivationUnauthorized, "Yarn concrete argv differs from C5 action", map[string]string{"subtype": subtype})
	}
	wantCWD, err := expandActionTemplate([]string{payload.WorkingDirectoryTemplate}, call.Template)
	if err != nil || len(wantCWD) != 1 || wantCWD[0] != call.CWD {
		return fail(CodeDerivationUnauthorized, "Yarn concrete cwd differs from C5 action", map[string]string{"subtype": subtype})
	}
	var wantEnvironment map[string]string
	switch subtype {
	case "yarn-install":
		wantEnvironment = yarnEnvironment(".curator-home", ".yarn/cache")
	case "node-invoke":
		wantEnvironment = map[string]string{"HOME": "home", "NO_PROXY": "*", "no_proxy": "*"}
	default:
		return fail(CodeDerivationUnauthorized, "Yarn C5 action subtype is unsupported", map[string]string{"subtype": subtype})
	}
	if !equalStringMaps(wantEnvironment, call.Environment) {
		return fail(CodeDerivationUnauthorized, "Yarn concrete environment differs from C5 policy", map[string]string{"subtype": subtype})
	}
	wantReads := append([]string(nil), call.Tool.ReadRoots...)
	for id := range call.Inputs {
		path := call.InputMounts[id]
		if path == "" {
			return fail(CodeDerivationUnauthorized, "Yarn concrete input mount is absent", map[string]string{"subtype": subtype})
		}
		wantReads = append(wantReads, path)
	}
	wantWrites := make([]string, len(call.WorkCopies))
	for index, work := range call.WorkCopies {
		wantWrites[index] = work.Path
	}
	sort.Strings(wantReads)
	sort.Strings(wantWrites)
	gotReads := append([]string(nil), call.ReadRoots...)
	gotWrites := append([]string(nil), call.WriteRoots...)
	sort.Strings(gotReads)
	sort.Strings(gotWrites)
	if !equalStrings(wantReads, gotReads) || !equalStrings(wantWrites, gotWrites) {
		return fail(CodeDerivationUnauthorized, "Yarn concrete read/write roots differ from typed operation roots", map[string]string{"subtype": subtype})
	}
	return nil
}

func validateActionBindings(subtype string, bindings map[string][]string, tool nodesource.ToolIdentity) error {
	one := func(name, expected string) bool {
		values := bindings[name]
		return len(values) == 1 && values[0] == expected
	}
	switch subtype {
	case "yarn-install":
		if !one("project", "work/yarn-modern-project") || !one("manager_entrypoint", relativeLogical("work/yarn-modern-project", tool.EntrypointRelativePath)) {
			return fail(CodeDerivationUnauthorized, "Yarn install action bindings differ from the closed C5 profile", nil)
		}
	case "node-invoke":
		if !one("project", "materialized") || len(bindings["entrypoint"]) != 1 {
			return fail(CodeDerivationUnauthorized, "Node invocation bindings differ from the closed C5 profile", nil)
		}
		if loader, ok := bindings["pnp_loader"]; ok && (len(loader) != 1 || loader[0] != "./.pnp.cjs") {
			return fail(CodeDerivationUnauthorized, "Node PnP loader binding differs from the closed C5 profile", nil)
		}
	default:
		return fail(CodeDerivationUnauthorized, "Yarn C5 action subtype is unsupported", map[string]string{"subtype": subtype})
	}
	return nil
}

func expandActionTemplate(template []string, bindings map[string][]string) ([]string, error) {
	result := []string{}
	for _, token := range template {
		if strings.HasPrefix(token, "{{") && strings.HasSuffix(token, "}}") && strings.Count(token, "{{") == 1 {
			name := strings.TrimSuffix(strings.TrimPrefix(token, "{{"), "}}")
			values, ok := bindings[name]
			if !ok {
				return nil, fmt.Errorf("action template binding %q is absent", name)
			}
			result = append(result, values...)
			continue
		}
		if strings.Contains(token, "{{") || strings.Contains(token, "}}") {
			return nil, fmt.Errorf("embedded action template placeholders are unsupported")
		}
		result = append(result, token)
	}
	return result, nil
}

func equalStringMaps(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func runnerStringsToAny(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func runnerStringMapAny(values map[string]string) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func relativeLogical(from, to string) string {
	rel, err := filepath.Rel(filepath.FromSlash(from), filepath.FromSlash(to))
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

func validateInstalledTree(ctx context.Context, root string, capture *Capture) ([]string, error) {
	graph := capture.Graph
	if graph.Layout.NodeLinker == "pnp" {
		loader := filepath.Join(root, ".pnp.cjs")
		info, err := os.Lstat(loader)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
			return nil, fail(CodeOfflineInputMissing, "Yarn did not regenerate the PnP loader", nil)
		}
		payload, err := os.ReadFile(loader) // #nosec G304 -- exact generated PnP path.
		if err != nil || len(payload) == 0 || bytes.IndexByte(payload, 0) >= 0 {
			return nil, fail(CodeMetadataMismatch, "generated PnP loader is not source text", nil)
		}
		return reconcilePnPRuntimeState(payload, graph, capture.tarballs)
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	realRoot, err = filepath.Abs(realRoot)
	if err != nil {
		return nil, err
	}
	observed, err := installedPackagePaths(root, graph.Layout.ModulesFolder)
	if err != nil {
		return nil, err
	}
	selectedPaths := make([]string, 0, len(observed))
	for installedPath := range observed {
		selectedPaths = append(selectedPaths, installedPath)
	}
	sort.Strings(selectedPaths)
	seen := map[string]bool{"workspace:.": true}
	for _, installedPath := range selectedPaths {
		packageRoot := filepath.Join(root, filepath.FromSlash(installedPath))
		resolvedRoot, resolveErr := filepath.EvalSymlinks(packageRoot)
		if resolveErr != nil {
			return nil, fail(CodeGraphIncomplete, "Yarn materialized an unresolved package link", map[string]string{"path": installedPath})
		}
		resolvedRoot, resolveErr = filepath.Abs(resolvedRoot)
		if resolveErr != nil || (resolvedRoot != realRoot && !strings.HasPrefix(resolvedRoot, realRoot+string(filepath.Separator))) {
			return nil, fail(CodeLocalPathEscape, "Yarn package link escapes the materialized project", map[string]string{"path": installedPath})
		}
		files, manifest, inspection, err := inventoryMaterializedPackage(resolvedRoot, installedPath, selectedPaths)
		if err != nil {
			return nil, err
		}
		candidates := []Package{}
		for _, pkg := range graph.Packages {
			if pkg.Selected && pkg.Name == manifest.Name && pkg.Version == manifest.Version {
				candidates = append(candidates, pkg)
			}
		}
		matched := ""
		for _, pkg := range candidates {
			if pkg.Resolved == "" {
				workspaceRoot := filepath.Join(realRoot, filepath.FromSlash(pkg.WorkspacePath))
				workspaceRoot, _ = filepath.EvalSymlinks(workspaceRoot)
				workspaceRoot, _ = filepath.Abs(workspaceRoot)
				if resolvedRoot == workspaceRoot {
					matched = pkg.Key
					break
				}
				continue
			}
			item, ok := capture.tarballs[packageSourceKey(pkg)]
			if !ok || !equalPackageFiles(files, item.files) {
				continue
			}
			if err = reconcileEmbeddedMetadata(pkg, manifest, inspection); err != nil {
				return nil, err
			}
			if err = admitMaterializedPackage(ctx, capture.policy, graph, pkg, resolvedRoot, installedPath); err != nil {
				return nil, err
			}
			matched = pkg.Key
			break
		}
		if matched == "" {
			return nil, fail(CodeGraphIncomplete, "Yarn materialized an extra or substituted package", map[string]string{"path": installedPath, "name": manifest.Name, "version": manifest.Version})
		}
		seen[matched] = true
	}
	for _, pkg := range graph.Packages {
		if pkg.Selected && pkg.Key != "workspace:." && !seen[pkg.Key] {
			return nil, fail(CodeOfflineInputMissing, "Yarn omitted a selected lock/workspace package", map[string]string{"package": pkg.Key})
		}
	}
	result := make([]string, 0, len(observed))
	for value := range observed {
		result = append(result, value)
	}
	sort.Strings(result)
	return result, nil
}

func reconcilePnPRuntimeState(loader []byte, graph Graph, archives map[string]capturedInput) ([]string, error) {
	state, err := parsePnPRuntimeState(loader)
	if err != nil {
		return nil, fail(CodeMetadataMismatch, "generated PnP loader lacks the pinned Yarn runtime state: "+err.Error(), nil)
	}
	observed, err := decodePnPPackageRegistry(state.PackageRegistryData)
	if err != nil {
		return nil, fail(CodeMetadataMismatch, "generated PnP package registry is malformed", nil)
	}
	baseLocatorByKey := map[string]string{}
	baseKeyByLocator := map[string]string{}
	selected := []string{}
	virtualKeys := []string{}
	aliases := map[string]string{}
	for _, pkg := range graph.Packages {
		if pkg.BaseKey != "" {
			if pkg.Selected {
				virtualKeys = append(virtualKeys, pkg.Key)
				selected = append(selected, pkg.Key)
			}
			continue
		}
		locator, locatorErr := pnpPackageLocator(pkg)
		if locatorErr != nil {
			return nil, locatorErr
		}
		baseLocatorByKey[pkg.Key] = locator
		baseKeyByLocator[locator] = pkg.Key
		if pkg.Selected {
			aliases[pkg.Key] = locator
			selected = append(selected, pkg.Key)
			info, ok := observed[locator]
			if !ok {
				return nil, fail(CodeGraphIncomplete, "generated PnP state omits a selected package", map[string]string{"package": pkg.Key})
			}
			if err := validatePnPLocation(pkg, info.Location, archives); err != nil {
				return nil, err
			}
		} else if _, ok := observed[locator]; ok {
			return nil, fail(CodeGraphIncomplete, "generated PnP state contains pruned package "+pkg.Key, map[string]string{"package": pkg.Key})
		}
	}

	observedVirtual := []string{}
	virtualBase := map[string]string{}
	for locator := range observed {
		if _, base := baseKeyByLocator[locator]; base {
			continue
		}
		baseLocator, parseErr := yarnVirtualBaseLocator(locator)
		if parseErr != nil {
			return nil, fail(CodeGraphIncomplete, "generated PnP state contains an unauthorized locator", map[string]string{"locator": locator})
		}
		baseKey, ok := baseKeyByLocator[baseLocator]
		if !ok {
			return nil, fail(CodeGraphIncomplete, "generated PnP virtual locator retargets outside the lock graph", map[string]string{"locator": locator})
		}
		virtualBase[locator] = baseKey
		observedVirtual = append(observedVirtual, locator)
	}
	if len(observed) != len(selected) || len(observedVirtual) != len(virtualKeys) {
		return nil, fail(CodeGraphIncomplete, "generated PnP package set differs from the selected graph", map[string]string{"expected": fmt.Sprint(len(selected)), "observed": fmt.Sprint(len(observed))})
	}
	sort.Strings(virtualKeys)
	sort.Strings(observedVirtual)
	candidates := map[string][]string{}
	for _, key := range virtualKeys {
		pkg := graph.Packages[graph.packageIndex[key]]
		for _, locator := range observedVirtual {
			if virtualBase[locator] != pkg.BaseKey {
				continue
			}
			if err := validateVirtualPnPLocation(pkg, locator, observed[locator].Location, archives); err != nil {
				continue
			}
			if normalizedPnPDependencies(observed[locator].Dependencies, virtualBase, baseKeyByLocator) == normalizedExpectedPnPDependencies(pkg, graph) && equalStrings(observed[locator].Peers, expectedPnPPeers(pkg, graph)) {
				candidates[key] = append(candidates[key], locator)
			}
		}
		if len(candidates[key]) == 0 {
			return nil, fail(CodeGraphIncomplete, "generated PnP virtual context has no authorized peer instance", map[string]string{"package": key})
		}
	}
	assignments := 0
	var solution map[string]string
	var assign func(int, map[string]string, map[string]bool)
	assign = func(position int, current map[string]string, used map[string]bool) {
		if assignments > 1 {
			return
		}
		if position == len(virtualKeys) {
			all := cloneMap(aliases)
			for key, locator := range current {
				all[key] = locator
			}
			for key, locator := range current {
				pkg := graph.Packages[graph.packageIndex[key]]
				if !equalStringMaps(expectedPnPDependencies(pkg, graph, all), observed[locator].Dependencies) {
					return
				}
			}
			for key, locator := range aliases {
				pkg := graph.Packages[graph.packageIndex[key]]
				if !equalStringMaps(expectedPnPDependencies(pkg, graph, all), observed[locator].Dependencies) {
					return
				}
			}
			assignments++
			solution = all
			return
		}
		key := virtualKeys[position]
		for _, locator := range candidates[key] {
			if used[locator] {
				continue
			}
			current[key], used[locator] = locator, true
			assign(position+1, current, used)
			delete(current, key)
			delete(used, locator)
		}
	}
	assign(0, map[string]string{}, map[string]bool{})
	if assignments != 1 {
		return nil, fail(CodeGraphIncomplete, "generated PnP virtual contexts are missing, ambiguous, or cross-wired", map[string]string{"matches": fmt.Sprint(assignments)})
	}
	_ = solution
	sort.Strings(selected)
	return selected, nil
}

func parsePnPRuntimeState(loader []byte) (pnpRuntimeState, error) {
	const prefix = "const RAW_RUNTIME_STATE =\n'"
	const suffix = "';\n\nfunction $$SETUP_STATE"
	start := bytes.Index(loader, []byte(prefix))
	if start < 0 {
		return pnpRuntimeState{}, fmt.Errorf("runtime state prefix is absent")
	}
	encoded := loader[start+len(prefix):]
	end := bytes.Index(encoded, []byte(suffix))
	if end < 0 {
		return pnpRuntimeState{}, fmt.Errorf("runtime state suffix is absent")
	}
	encoded = encoded[:end]
	encoded = bytes.ReplaceAll(encoded, []byte("\\\r\n"), nil)
	encoded = bytes.ReplaceAll(encoded, []byte("\\\n"), nil)
	var state pnpRuntimeState
	if err := json.Unmarshal(encoded, &state); err != nil {
		return pnpRuntimeState{}, err
	}
	if len(state.DependencyTreeRoots) == 0 || len(state.PackageRegistryData) == 0 {
		return pnpRuntimeState{}, fmt.Errorf("runtime state is incomplete")
	}
	return state, nil
}

func decodePnPPackageRegistry(entries []json.RawMessage) (map[string]pnpObservedPackage, error) {
	result := map[string]pnpObservedPackage{}
	for _, encodedEntry := range entries {
		var entry []json.RawMessage
		if json.Unmarshal(encodedEntry, &entry) != nil || len(entry) != 2 {
			return nil, fmt.Errorf("invalid registry entry")
		}
		if bytes.Equal(bytes.TrimSpace(entry[0]), []byte("null")) {
			continue
		}
		var name string
		var stores []json.RawMessage
		if json.Unmarshal(entry[0], &name) != nil || name == "" || json.Unmarshal(entry[1], &stores) != nil {
			return nil, fmt.Errorf("invalid registry package")
		}
		for _, encodedStore := range stores {
			var store []json.RawMessage
			if json.Unmarshal(encodedStore, &store) != nil || len(store) != 2 {
				return nil, fmt.Errorf("invalid registry store")
			}
			var reference string
			var info pnpPackageInformation
			if json.Unmarshal(store[0], &reference) != nil || reference == "" || json.Unmarshal(store[1], &info) != nil {
				return nil, fmt.Errorf("invalid package information")
			}
			locator := name + "\x00" + reference
			if _, exists := result[locator]; exists {
				return nil, fmt.Errorf("duplicate package locator")
			}
			dependencies, dependencyErr := decodePnPDependencies(info.PackageDependencies)
			if dependencyErr != nil {
				return nil, dependencyErr
			}
			sort.Strings(info.PackagePeers)
			result[locator] = pnpObservedPackage{Location: info.PackageLocation, Dependencies: dependencies, Peers: info.PackagePeers}
		}
	}
	return result, nil
}

func decodePnPDependencies(entries []json.RawMessage) (map[string]string, error) {
	result := map[string]string{}
	for _, encoded := range entries {
		var entry []json.RawMessage
		if json.Unmarshal(encoded, &entry) != nil || len(entry) != 2 {
			return nil, fmt.Errorf("invalid dependency entry")
		}
		var declaredName string
		if json.Unmarshal(entry[0], &declaredName) != nil || declaredName == "" {
			return nil, fmt.Errorf("invalid dependency name")
		}
		if bytes.Equal(bytes.TrimSpace(entry[1]), []byte("null")) {
			result[declaredName] = ""
			continue
		}
		var reference string
		if json.Unmarshal(entry[1], &reference) == nil {
			result[declaredName] = declaredName + "\x00" + reference
			continue
		}
		var alias []string
		if json.Unmarshal(entry[1], &alias) != nil || len(alias) != 2 || alias[0] == "" || alias[1] == "" {
			return nil, fmt.Errorf("invalid dependency target")
		}
		result[declaredName] = alias[0] + "\x00" + alias[1]
	}
	return result, nil
}

func pnpPackageLocator(pkg Package) (string, error) {
	reference := "workspace:" + pkg.WorkspacePath
	if pkg.Key == "workspace:." {
		reference = "workspace:."
	}
	if pkg.Resolution != "" {
		prefix := pkg.Name + "@"
		if !strings.HasPrefix(pkg.Resolution, prefix) || len(pkg.Resolution) == len(prefix) {
			return "", fail(CodeGraphIncomplete, "Yarn resolution cannot be projected to a PnP locator", map[string]string{"package": pkg.Key})
		}
		reference = strings.TrimPrefix(pkg.Resolution, prefix)
	}
	return pkg.Name + "\x00" + reference, nil
}

func validatePnPLocation(pkg Package, location string, archives map[string]capturedInput) error {
	if location == "" || filepath.IsAbs(location) || strings.Contains(location, "\\") || strings.Contains(location, "../") {
		return fail(CodeLocalPathEscape, "generated PnP package location escapes materialized closure", map[string]string{"package": pkg.Key})
	}
	if pkg.Resolution == "" {
		expected := "./"
		if pkg.Key != "workspace:." {
			expected = "./" + strings.TrimSuffix(pkg.WorkspacePath, "/") + "/"
		}
		if location != expected {
			return fail(CodeGraphIncomplete, "generated PnP workspace location differs from captured authority", map[string]string{"package": pkg.Key})
		}
		return nil
	}
	archive, ok := archives[pkg.Key]
	want := "./.yarn/cache/" + archive.cacheName + "/node_modules/" + pkg.Name + "/"
	if !ok || archive.cacheName == "" || location != want {
		return fail(CodeGraphIncomplete, "generated PnP archive location differs from private cache authority", map[string]string{"package": pkg.Key})
	}
	return nil
}

func yarnVirtualBaseLocator(locator string) (string, error) {
	parts := strings.SplitN(locator, "\x00", 2)
	if len(parts) != 2 || parts[0] == "" || !strings.HasPrefix(parts[1], "virtual:") {
		return "", fmt.Errorf("not a virtual Yarn locator")
	}
	hashAndBase := strings.TrimPrefix(parts[1], "virtual:")
	hash, base, ok := strings.Cut(hashAndBase, "#")
	if !ok || len(hash) != sha512.Size*2 || base == "" {
		return "", fmt.Errorf("malformed virtual Yarn locator")
	}
	decoded, err := hex.DecodeString(hash)
	if err != nil || len(decoded) != sha512.Size || strings.ToLower(hash) != hash {
		return "", fmt.Errorf("malformed virtual Yarn hash")
	}
	return parts[0] + "\x00" + base, nil
}

func validateVirtualPnPLocation(pkg Package, locator, location string, archives map[string]capturedInput) error {
	base := pkg
	base.Key = pkg.BaseKey
	base.BaseKey = ""
	base.PeerContext = nil
	baseLocation, err := expectedPnPBaseLocation(base, archives)
	if err != nil {
		return err
	}
	component, err := yarnVirtualLocatorSlug(locator)
	if err != nil {
		return err
	}
	depth := "1"
	suffix := strings.TrimPrefix(baseLocation, "./")
	if strings.HasPrefix(baseLocation, "./.yarn/") {
		depth = "0"
		suffix = strings.TrimPrefix(baseLocation, "./.yarn/")
	}
	want := "./.yarn/__virtual__/" + component + "/" + depth + "/" + suffix
	if location != want {
		return fail(CodeGraphIncomplete, "generated PnP virtual location differs from its admitted base locator", map[string]string{"package": pkg.Key})
	}
	return nil
}

func expectedPnPBaseLocation(pkg Package, archives map[string]capturedInput) (string, error) {
	if pkg.Resolution == "" {
		if pkg.Key == "workspace:." {
			return "./", nil
		}
		return "./" + strings.TrimSuffix(pkg.WorkspacePath, "/") + "/", nil
	}
	archive, ok := archives[packageSourceKey(pkg)]
	if !ok || archive.cacheName == "" {
		return "", fail(CodeGraphIncomplete, "generated PnP virtual package lacks private cache authority", map[string]string{"package": pkg.Key})
	}
	return "./.yarn/cache/" + archive.cacheName + "/node_modules/" + pkg.Name + "/", nil
}

func yarnVirtualLocatorSlug(locator string) (string, error) {
	parts := strings.SplitN(locator, "\x00", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("malformed virtual locator")
	}
	if _, err := yarnVirtualBaseLocator(locator); err != nil {
		return "", err
	}
	scope, name := "", parts[0]
	slugName := name
	if strings.HasPrefix(name, "@") {
		ident := strings.SplitN(strings.TrimPrefix(name, "@"), "/", 2)
		if len(ident) != 2 || ident[0] == "" || ident[1] == "" {
			return "", fmt.Errorf("malformed scoped virtual locator")
		}
		scope, name = ident[0], ident[1]
		slugName = "@" + scope + "-" + name
	}
	identHash := sha512.Sum512([]byte(scope + name))
	locatorHash := sha512.Sum512([]byte(hex.EncodeToString(identHash[:]) + parts[1]))
	return slugName + "-virtual-" + hex.EncodeToString(locatorHash[:])[:10], nil
}

func normalizedPnPDependencies(dependencies map[string]string, virtualBase, baseKeyByLocator map[string]string) string {
	values := make([]string, 0, len(dependencies))
	for name, locator := range dependencies {
		target := ""
		if locator != "" {
			if key, ok := baseKeyByLocator[locator]; ok {
				target = key
			} else {
				target = virtualBase[locator]
				if target == "" {
					target = "invalid:" + locator
				}
			}
		}
		values = append(values, name+"="+target)
	}
	sort.Strings(values)
	return strings.Join(values, "\n")
}

func normalizedExpectedPnPDependencies(pkg Package, graph Graph) string {
	values := []string{pkg.Name + "=" + packageSourceKey(pkg)}
	for _, edge := range graph.Edges {
		if edge.From != pkg.Key {
			continue
		}
		if edge.To == "" {
			if edge.Scope == "peer" && edge.Reason == "optional_peer_unresolved" {
				values = append(values, edge.Name+"=")
			}
			continue
		}
		if edge.Selected {
			values = append(values, edge.Name+"="+packageSourceKey(graph.Packages[graph.packageIndex[edge.To]]))
		}
	}
	sort.Strings(values)
	return strings.Join(values, "\n")
}

func expectedPnPDependencies(pkg Package, graph Graph, aliases map[string]string) map[string]string {
	dependencies := map[string]string{pkg.Name: aliases[pkg.Key]}
	for _, edge := range graph.Edges {
		if edge.From != pkg.Key {
			continue
		}
		if edge.To == "" {
			if edge.Scope == "peer" && edge.Reason == "optional_peer_unresolved" {
				dependencies[edge.Name] = ""
			}
			continue
		}
		if edge.Selected {
			dependencies[edge.Name] = aliases[edge.To]
		}
	}
	return dependencies
}

func expectedPnPPeers(pkg Package, graph Graph) []string {
	if pkg.BaseKey == "" {
		return nil
	}
	peers := []string{}
	for _, edge := range graph.Edges {
		if edge.From == pkg.Key && edge.Scope == "peer" {
			peers = append(peers, edge.Name)
		}
	}
	sort.Strings(peers)
	return peers
}

func admitMaterializedPackage(ctx context.Context, policy *artifactpolicy.Service, graph Graph, pkg Package, root, installedPath string) error {
	if policy == nil {
		return fail(CodeInputUndeclared, "artifact policy is absent during materialized-tree admission", nil)
	}
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "yarn-modern", PackageName: pkg.Name, PackageVersion: pkg.Version}
	probe, err := policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: "materialized/" + installedPath})
	if err != nil && artifactpolicy.ErrorCode(err) != artifactpolicy.CodeOriginUnverified {
		return err
	}
	digest := probe.Manifest.RawPayload.SHA256
	if !closuregraph.ID(digest).Valid() {
		return fail(CodeIntegrityMismatch, "materialized Yarn package identity is unavailable", map[string]string{"package": pkg.Key})
	}
	descriptor.Origin = artifactpolicy.OriginEvidence{Locator: pkg.Resolved, ImmutableID: pkg.Integrity, LockRecord: graph.LockDigest, ChecksumSHA256: digest, Verified: true}
	_, err = policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: root, VirtualRoot: "materialized/" + installedPath})
	return err
}

func inventoryMaterializedPackage(root, installPath string, selectedPaths []string) ([]packageFile, packageManifest, tarInspection, error) {
	descendants := map[string]bool{}
	for _, candidate := range selectedPaths {
		if candidate == installPath || !strings.HasPrefix(candidate, installPath+"/") {
			continue
		}
		descendants[strings.TrimPrefix(candidate, installPath+"/")] = true
	}
	files := []packageFile{}
	inspection := tarInspection{}
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
		logical := filepath.ToSlash(rel)
		if descendants[logical] && entry.IsDir() {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fail(CodeInputUndeclared, "materialized Yarn package contains a non-regular member", map[string]string{"package": installPath, "path": logical})
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- path is discovered beneath an owned materialized package root.
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		files = append(files, packageFile{Path: logical, SHA256: digestID(payload), Size: int64(len(payload)), Executable: info.Mode()&0o111 != 0})
		if logical == "binding.gyp" {
			inspection.bindingGYP = true
		}
		if strings.HasPrefix(logical, "node_modules/") {
			inspection.bundled = true
		}
		return nil
	})
	if err != nil {
		return nil, packageManifest{}, inspection, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	manifestBytes, err := os.ReadFile(filepath.Join(root, "package.json")) // #nosec G304 -- exact contained package metadata path.
	if err != nil || validateJSON(manifestBytes) != nil {
		return nil, packageManifest{}, inspection, fail(CodeMetadataMismatch, "materialized Yarn package metadata is absent or invalid", map[string]string{"package": installPath})
	}
	var manifest packageManifest
	if err = json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, packageManifest{}, inspection, fail(CodeMetadataMismatch, "materialized Yarn package metadata cannot be decoded", map[string]string{"package": installPath})
	}
	return files, manifest, inspection, nil
}

func equalPackageFiles(left, right []packageFile) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
func installedPackagePaths(root, modulesFolder string) (map[string]bool, error) {
	result := map[string]bool{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, current)
		logical := filepath.ToSlash(rel)
		isRootModules := logical == modulesFolder
		if !entry.IsDir() || (entry.Name() != "node_modules" && !isRootModules) {
			return nil
		}
		children, err := os.ReadDir(current)
		if err != nil {
			return err
		}
		for _, child := range children {
			if strings.HasPrefix(child.Name(), ".") {
				continue
			}
			if strings.HasPrefix(child.Name(), "@") {
				scoped, err := os.ReadDir(filepath.Join(current, child.Name()))
				if err != nil {
					return err
				}
				for _, leaf := range scoped {
					if leaf.IsDir() || leaf.Type()&fs.ModeSymlink != 0 {
						rel, _ := filepath.Rel(root, filepath.Join(current, child.Name(), leaf.Name()))
						result[filepath.ToSlash(rel)] = true
					}
				}
				continue
			}
			if child.IsDir() || child.Type()&fs.ModeSymlink != 0 {
				rel, _ := filepath.Rel(root, filepath.Join(current, child.Name()))
				result[filepath.ToSlash(rel)] = true
			}
		}
		return nil
	})
	return result, err
}

func yarnEnvironment(home, cache string) map[string]string {
	return map[string]string{"HOME": home, "YARN_CACHE_FOLDER": cache, "YARN_ENABLE_SCRIPTS": "0", "YARN_ENABLE_NETWORK": "0", "NO_PROXY": "*", "no_proxy": "*"}
}
func cleanAbsentAbsolute(value string) (string, error) {
	abs, err := filepath.Abs(value)
	if err != nil || value == "" || abs != filepath.Clean(value) {
		return "", fail(CodeInputUndeclared, "destination must be an absolute clean path", map[string]string{"path": value})
	}
	if _, err = os.Lstat(abs); !errors.Is(err, fs.ErrNotExist) {
		return "", fail(CodeInputUndeclared, "destination must be absent", map[string]string{"path": abs})
	}
	return abs, nil
}
func admitMaterializedTree(capture *Capture, root string, derivation closureexec.DerivationReceipt) (closureexec.AdmittedInput, closuregraph.ID, error) {
	if capture == nil || capture.store == nil {
		return closureexec.AdmittedInput{}, "", fail(CodeDerivationUnauthorized, "materialized tree store is absent", nil)
	}
	stage, err := os.MkdirTemp(filepath.Dir(root), ".yarn-modern-materialized-replay-")
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	_ = os.Remove(stage)
	defer func() { _ = os.RemoveAll(stage) }()
	if err = copyContainedTreeDereferencingLinks(root, root, stage, map[string]bool{}); err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	files, err := inventoryFiles(stage)
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	values := make([]any, len(files))
	for index, file := range files {
		values[index] = map[string]any{"path": file.Path, "sha256": string(file.SHA256), "size": file.Size}
	}
	manifestID, err := closuregraph.DomainID("yarn-modern-materialized-tree-v1", map[string]any{"files": values, "profile": ProfileID})
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	tree, err := capture.store.CaptureTree("derived:yarn-modern-materialized:"+string(manifestID), stage)
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	receipt, err := capture.store.AdmitTree(tree, "derived:yarn-modern-materialized:"+string(manifestID), closureexec.AdmissionEvidence{PreviousCausalHead: string(derivation.NextCausalHead), ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: "yarn-modern-materialized-v1", DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	receiptID, err := receipt.ID()
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	return closureexec.AdmittedInput{Receipt: receipt, Tree: tree}, receiptID, nil
}

func copyContainedTreeDereferencingLinks(root, source, destination string, active map[string]bool) error {
	realSource, err := filepath.EvalSymlinks(source)
	if err != nil {
		return err
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	inside, err := filepath.Rel(realRoot, realSource)
	if err != nil || inside == ".." || strings.HasPrefix(inside, ".."+string(filepath.Separator)) {
		return fail(CodeLocalPathEscape, "materialized Yarn link escapes the project", map[string]string{"path": source})
	}
	if active[realSource] {
		return fail(CodeInputUndeclared, "materialized Yarn link cycle is unsupported", map[string]string{"path": source})
	}
	active[realSource] = true
	defer delete(active, realSource)
	info, err := os.Stat(realSource)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		payload, readErr := os.ReadFile(realSource) // #nosec G304 -- evaluated path is contained by the validated materialized root.
		if readErr != nil {
			return readErr
		}
		if err = privatedir.MakeAll(filepath.Dir(destination)); err != nil {
			return err
		}
		return os.WriteFile(destination, payload, 0o600)
	}
	if err = privatedir.MakeAll(destination); err != nil {
		return err
	}
	entries, err := os.ReadDir(realSource)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err = copyContainedTreeDereferencingLinks(root, filepath.Join(realSource, entry.Name()), filepath.Join(destination, entry.Name()), active); err != nil {
			return err
		}
	}
	return nil
}
func inventoryFiles(root string) ([]MirrorFile, error) {
	files := []MirrorFile{}
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
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeInputUndeclared, "private Yarn mirror contains a link", map[string]string{"path": filepath.ToSlash(rel)})
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return fail(CodeInputUndeclared, "private Yarn mirror contains a special node", map[string]string{"path": filepath.ToSlash(rel)})
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a contained private-cache member.
		if err != nil {
			return err
		}
		files = append(files, MirrorFile{Path: filepath.ToSlash(rel), SHA256: digestID(payload), Size: int64(len(payload))})
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, err
}
func makeTreeReadOnly(root string) error {
	return filepath.WalkDir(root, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Chmod(current, 0o500) // #nosec G302 -- immutable cache directories require execute permission for traversal.
		}
		return os.Chmod(current, 0o400)
	})
}
func equalMirrorFiles(a, b []MirrorFile) bool {
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
func idsToStrings(values []closuregraph.ID) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = string(value)
	}
	return out
}
