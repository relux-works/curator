package yarnclassicsource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/nodesource"
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
	Receipt    MirrorReceipt
	root       string
	capture    *Capture
	input      closureexec.AdmittedInput
	inputID    closuregraph.ID
	cacheInput closureexec.AdmittedInput
	cacheID    closuregraph.ID
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

// DerivePrivateMirror copies exact admitted tgz bytes into a task-private Yarn
// source mirror and emits a fixed mirror-only yarnrc. No manager cache or
// ambient installed tree participates in this derivation.
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
	if err = os.Mkdir(dest, 0o700); err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(dest)
		}
	}()
	mirrorDir := filepath.Join(dest, "mirror")
	if err = os.Mkdir(mirrorDir, 0o700); err != nil {
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
		pkg := capture.Graph.Packages[capture.Graph.packageIndex[key]]
		parsed, parseErr := url.Parse(pkg.Resolved)
		if parseErr != nil {
			return nil, fail(CodeOriginUnpinned, "captured Yarn locator cannot name a mirror member", map[string]string{"package": key})
		}
		name := path.Base(parsed.Path)
		if name == "." || name == "/" || !strings.HasSuffix(strings.ToLower(name), ".tgz") {
			return nil, fail(CodeOriginUnpinned, "captured Yarn locator has no source tarball name", map[string]string{"package": key})
		}
		if prior := usedNames[name]; prior != "" && prior != key {
			return nil, fail(CodeGraphIncomplete, "Yarn offline mirror filename collision", map[string]string{"filename": name, "first": prior, "second": key})
		}
		usedNames[name] = key
		reader, openErr := item.handle.Open()
		if openErr != nil {
			return nil, openErr
		}
		output, createErr := os.OpenFile(filepath.Join(mirrorDir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- name is the validated basename of an admitted immutable .tgz locator.
		if createErr != nil {
			_ = reader.Close()
			return nil, createErr
		}
		_, copyErr := io.Copy(output, reader)
		closeErr := output.Close()
		_ = reader.Close()
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
	if err = os.WriteFile(filepath.Join(dest, "yarnrc"), []byte("yarn-offline-mirror \"./mirror\"\nyarn-offline-mirror-pruning false\n"), 0o600); err != nil {
		return nil, err
	}
	files, err := inventoryFiles(dest)
	if err != nil {
		return nil, err
	}
	if len(keys) > 0 && len(files) == 0 {
		return nil, fail(CodeOfflineInputMissing, "Yarn mirror derivation produced no source tarballs", nil)
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
	id, err := closuregraph.DomainID("yarn-classic-private-mirror-receipt-v1", map[string]any{"files": payloadFiles, "input_receipt_ids": idsToStrings(receipts), "schema_id": "yarn-classic-private-mirror-receipt-v1"})
	if err != nil {
		return nil, err
	}
	if err = makeTreeReadOnly(dest); err != nil {
		return nil, err
	}
	tree, err := capture.store.CaptureTree("derived:yarn-classic-mirror:"+string(id), dest)
	if err != nil {
		return nil, err
	}
	derivedReceipt, err := capture.store.AdmitTree(tree, "derived:yarn-classic-mirror:"+string(id), closureexec.AdmissionEvidence{PreviousCausalHead: session.operation.CurrentCausalHead(), ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: "yarn-classic-private-mirror-v1", DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: id})
	if err != nil {
		return nil, err
	}
	derivedID, err := derivedReceipt.ID()
	if err != nil {
		return nil, err
	}
	emptyRoot := filepath.Join(workRoot, "empty-yarn-cache")
	if err = os.MkdirAll(workRoot, 0o700); err != nil {
		return nil, err
	}
	if err = os.Mkdir(emptyRoot, 0o700); err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(emptyRoot) }()
	emptyManifestID, err := closuregraph.DomainID("yarn-classic-empty-cache-v1", map[string]any{"entries": []any{}, "schema_id": "yarn-classic-empty-cache-v1"})
	if err != nil {
		return nil, err
	}
	emptyTree, err := capture.store.CaptureTree("derived:yarn-classic-empty-cache:"+string(emptyManifestID), emptyRoot)
	if err != nil {
		return nil, err
	}
	emptyReceipt, err := capture.store.AdmitTree(emptyTree, "derived:yarn-classic-empty-cache:"+string(emptyManifestID), closureexec.AdmissionEvidence{PreviousCausalHead: string(derivedID), ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: "yarn-classic-empty-cache-v1", DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: emptyManifestID})
	if err != nil {
		return nil, err
	}
	emptyID, err := emptyReceipt.ID()
	if err != nil {
		return nil, err
	}
	success = true
	return &PrivateMirror{Receipt: MirrorReceipt{SchemaID: "yarn-classic-private-mirror-receipt-v1", InputReceiptIDs: receipts, Files: files, ID: id}, root: dest, capture: capture, input: closureexec.AdmittedInput{Receipt: derivedReceipt, Tree: tree}, inputID: derivedID, cacheInput: closureexec.AdmittedInput{Receipt: emptyReceipt, Tree: emptyTree}, cacheID: emptyID}, nil
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
	observed, err := inventoryFiles(mirror.root)
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
	projectPath := "work/yarn-classic-project"
	cachePath := "work/yarn-classic-empty-cache"
	mirrorPath := "capture/yarn-classic-private-mirror"
	homePath := ".curator-home"
	relativeCache := relativeLogical(projectPath, cachePath)
	yarnrc := relativeLogical(projectPath, mirrorPath+"/yarnrc")
	args := []string{"install", "--frozen-lockfile", "--offline", "--ignore-scripts", "--non-interactive", "--no-default-rc", "--cache-folder", relativeCache, "--use-yarnrc", yarnrc}
	production := []string{}
	if !mirror.capture.Graph.Target.IncludeDev {
		args = append(args, "--production=true")
		production = []string{"--production=true"}
	}
	modulesFolder := []string{}
	if mirror.capture.Graph.Layout.ModulesFolder != "node_modules" {
		args = append(args, "--modules-folder", mirror.capture.Graph.Layout.ModulesFolder)
		modulesFolder = []string{"--modules-folder", mirror.capture.Graph.Layout.ModulesFolder}
	}
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{mirror.capture.project.receiptID: mirror.capture.project.input, mirror.inputID: mirror.input, mirror.cacheID: mirror.cacheInput}
	works := []closureexec.WorkCopy{{ReceiptID: mirror.capture.project.receiptID, Path: projectPath, Retain: true}, {ReceiptID: mirror.cacheID, Path: cachePath}}
	sort.Slice(works, func(i, j int) bool { return works[i].ReceiptID < works[j].ReceiptID })
	invocation := invocation{
		Tool: authority.Runtime.Manager, CWD: projectPath, Args: args,
		Environment: yarnEnvironment(homePath, relativeCache), Inputs: inputs,
		InputMounts: map[closuregraph.ID]string{mirror.capture.project.receiptID: "capture/yarn-classic-project", mirror.inputID: mirrorPath, mirror.cacheID: "capture/yarn-classic-empty-cache"},
		ReadRoots:   append([]string{"capture/yarn-classic-project", mirrorPath, "capture/yarn-classic-empty-cache"}, authority.Runtime.Manager.ReadRoots...), WriteRoots: []string{cachePath, projectPath}, WorkCopies: works,
		Template: map[string][]string{"manager_entrypoint": {relativeLogical(projectPath, authority.Runtime.Manager.EntrypointRelativePath)}, "cache": {relativeCache}, "yarnrc": {yarnrc}, "production": production, "modules_folder": modulesFolder, "project": {projectPath}},
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
	invocation := invocation{Tool: authority.Runtime.Node, CWD: logicalRoot, Args: append([]string{entrypoint}, args...), Environment: map[string]string{"HOME": "home", "NO_PROXY": "*", "no_proxy": "*"}, ReadRoots: append([]string{logicalRoot}, authority.Runtime.Node.ReadRoots...), WriteRoots: []string{}, Inputs: inputs, InputMounts: map[closuregraph.ID]string{materialized.inputID: logicalRoot}, Template: map[string][]string{"entrypoint": {entrypoint}, "args": args, "project": {logicalRoot}}}
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
		return fail(CodeRuntimeIdentityChanged, "bound manager is not the exact supported Yarn Classic release", map[string]string{"expected": SupportedYarnVersion, "observed": strings.TrimSpace(authority.Runtime.Manager.VersionOutput)})
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
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "yarn-classic-platform-v1", EvaluateFunc: evaluateYarnPlatformCondition}
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
	for _, clause := range strings.Split(condition.Expression, ";") {
		parts := strings.SplitN(clause, "=", 2)
		if len(parts) != 2 {
			return false, fail(CodeGraphIncomplete, "Yarn platform condition is malformed", map[string]string{"condition": condition.Expression})
		}
		actual := input.Selection.Markers[parts[0]]
		if actual == "" || !selectorMatches(strings.Split(parts[1], ","), actual) {
			return false, nil
		}
	}
	return true, nil
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
	stdoutID, err := closuregraph.DomainID("yarn-classic-operation-stdout-v1", map[string]any{"action_node_id": string(action.ID), "build_plan_id": string(session.planID), "subtype": subtype})
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	requirements := []closureexec.EvidenceRequirement{{Path: stdoutPath, SchemaID: "yarn-classic-operation-stdout-v1", ArtifactManifestID: stdoutID}}
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
	concreteID, err := closuregraph.DomainID("yarn-classic-c5-concrete-action-v1", map[string]any{"action_node_id": string(action.ID), "argv": runnerStringsToAny(call.Args), "cwd": call.CWD, "environment": runnerStringMapAny(environment), "read_roots": runnerStringsToAny(reads), "write_roots": runnerStringsToAny(call.WriteRoots)})
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	invocationKey := fmt.Sprintf("yarn-classic-c5:%s:%s:%s:%s", session.planID, action.ID, subtype, concreteID)
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
		wantEnvironment = yarnEnvironment(".curator-home", templateOne(call.Template, "cache", ""))
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
		production := bindings["production"]
		modules := bindings["modules_folder"]
		validProduction := len(production) == 0 || (len(production) == 1 && production[0] == "--production=true")
		validModules := len(modules) == 0 || (len(modules) == 2 && modules[0] == "--modules-folder" && modules[1] != "" && !path.IsAbs(modules[1]) && path.Clean(modules[1]) == modules[1])
		if !one("project", "work/yarn-classic-project") || !one("manager_entrypoint", relativeLogical("work/yarn-classic-project", tool.EntrypointRelativePath)) || !one("cache", "../yarn-classic-empty-cache") || !one("yarnrc", relativeLogical("work/yarn-classic-project", "capture/yarn-classic-private-mirror/yarnrc")) || !validProduction || !validModules {
			return fail(CodeDerivationUnauthorized, "Yarn install action bindings differ from the closed C5 profile", nil)
		}
	case "node-invoke":
		if !one("project", "materialized") || len(bindings["entrypoint"]) != 1 {
			return fail(CodeDerivationUnauthorized, "Node invocation bindings differ from the closed C5 profile", nil)
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

func templateOne(bindings map[string][]string, name, suffix string) string {
	values := bindings[name]
	if len(values) != 1 {
		return ""
	}
	if suffix == "" {
		return values[0]
	}
	return filepath.ToSlash(filepath.Join(values[0], suffix))
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
			item, ok := capture.tarballs[pkg.Key]
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

func admitMaterializedPackage(ctx context.Context, policy *artifactpolicy.Service, graph Graph, pkg Package, root, installedPath string) error {
	if policy == nil {
		return fail(CodeInputUndeclared, "artifact policy is absent during materialized-tree admission", nil)
	}
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "yarn-classic", PackageName: pkg.Name, PackageVersion: pkg.Version}
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
	return map[string]string{"HOME": home, "YARN_CACHE_FOLDER": cache, "YARN_IGNORE_SCRIPTS": "true", "YARN_ENABLE_NETWORK": "0", "NO_PROXY": "*", "no_proxy": "*"}
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
	stage, err := os.MkdirTemp(filepath.Dir(root), ".yarn-classic-materialized-replay-")
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
	manifestID, err := closuregraph.DomainID("yarn-classic-materialized-tree-v1", map[string]any{"files": values, "profile": ProfileID})
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	tree, err := capture.store.CaptureTree("derived:yarn-classic-materialized:"+string(manifestID), stage)
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	receipt, err := capture.store.AdmitTree(tree, "derived:yarn-classic-materialized:"+string(manifestID), closureexec.AdmissionEvidence{PreviousCausalHead: string(derivation.NextCausalHead), ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: "yarn-classic-materialized-v1", DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
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
		if err = os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			return err
		}
		return os.WriteFile(destination, payload, 0o600)
	}
	if err = os.MkdirAll(destination, 0o700); err != nil {
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
