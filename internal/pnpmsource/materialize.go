package pnpmsource

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/artifactpolicy"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/nodesource"
)

// ExecutionContext supplies the shared protected executor and exact C0/C5 authority.
type ExecutionContext struct {
	Executor      *closureexec.Executor
	Selection     closuregraph.SelectionContext
	Runtime       nodesource.RuntimeBinding
	BuildPlan     closuregraph.BuildPlan
	Recheck       func(context.Context, nodesource.ToolIdentity) (closureexec.ToolchainIdentity, error)
	ExecutionRoot string
}

// StoreFile is one deterministic member of the derived private store.
type StoreFile struct {
	Path   string
	SHA256 closuregraph.ID
	Size   int64
}

// StoreReceipt binds the complete private store to admitted input receipts.
type StoreReceipt struct {
	SchemaID        string
	InputReceiptIDs []closuregraph.ID
	Files           []StoreFile
	ID              closuregraph.ID
}

// PrivateStore is derived, receipted, read-only state, never source authority.
type PrivateStore struct {
	Receipt StoreReceipt
	root    string
	capture *Capture
	input   closureexec.AdmittedInput
	inputID closuregraph.ID
}

// MaterializeRequest selects an absent destination and task-private work root.
type MaterializeRequest struct{ Destination, WorkRoot string }

// Materialization records an exact frozen offline replay.
type Materialization struct {
	Root                 string
	StoreReceiptID       closuregraph.ID
	Receipt              closureexec.DerivationReceipt
	MaterializedPackages []string
	capture              *Capture
	input                closureexec.AdmittedInput
	inputID              closuregraph.ID
}

// Invoke runs one admitted JavaScript entry point through the exact bound Node
// runtime with the same networkless, ambient-free executor authority.
func Invoke(ctx context.Context, materialized *Materialization, entrypoint string, args []string, authority *ExecutionContext) (closureexec.DerivationReceipt, error) {
	if materialized == nil || materialized.capture == nil || materialized.input.Tree == nil {
		return closureexec.DerivationReceipt{}, fail(CodeInputUndeclared, "pnpm invocation authority is incomplete", nil)
	}
	if entrypoint == "" || filepath.IsAbs(entrypoint) || filepath.Clean(entrypoint) != entrypoint || entrypoint == ".." || strings.HasPrefix(entrypoint, ".."+string(filepath.Separator)) {
		return closureexec.DerivationReceipt{}, fail(CodeLocalPathEscape, "Node entry point escapes materialized pnpm closure", map[string]string{"path": entrypoint})
	}
	info, err := os.Lstat(filepath.Join(materialized.Root, entrypoint))
	if err != nil || !info.Mode().IsRegular() || info.Mode()&fs.ModeSymlink != 0 {
		return closureexec.DerivationReceipt{}, fail(CodeInputUndeclared, "Node entry point is absent or not regular", map[string]string{"path": entrypoint})
	}
	session, err := newRunnerSession(ctx, materialized.capture, authority)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{materialized.inputID: materialized.input}
	mounts := map[closuregraph.ID]string{materialized.inputID: "materialized"}
	call := invocation{Tool: authority.Runtime.Node, CWD: "materialized", Args: append([]string{entrypoint}, args...), Environment: map[string]string{"HOME": "home", "NO_PROXY": "*", "no_proxy": "*"}, Inputs: inputs, InputMounts: mounts, ReadRoots: append(values(mounts), authority.Runtime.Node.ReadRoots...), WriteRoots: []string{}}
	return session.run(ctx, call, "node-invoke")
}

type invocation struct {
	Tool                  nodesource.ToolIdentity
	CWD                   string
	Args                  []string
	Environment           map[string]string
	ReadRoots, WriteRoots []string
	Inputs                map[closuregraph.ID]closureexec.AdmittedInput
	InputMounts           map[closuregraph.ID]string
	WorkCopies            []closureexec.WorkCopy
}
type runtimeAction struct {
	ID      closuregraph.ID
	Payload closuregraph.ActionPayload
}
type runnerSession struct {
	authority    *ExecutionContext
	operation    *closureexec.AssuredOperation
	bundle       closuregraph.GraphBundle
	c0ID, planID closuregraph.ID
	actions      map[string]runtimeAction
}

// DerivePrivateStore runs exactly one protected `pnpm store add` over the full
// admitted tarball set, rejects side-effect output, receipts the result, and
// freezes it read-only before replay.
func DerivePrivateStore(ctx context.Context, capture *Capture, destination string, authority *ExecutionContext) (*PrivateStore, error) {
	if capture == nil || capture.project.tree == nil || capture.store == nil {
		return nil, fail(CodeInputUndeclared, "pnpm store derivation authority is incomplete", nil)
	}
	session, err := newRunnerSession(ctx, capture, authority)
	if err != nil {
		return nil, err
	}
	dest, err := cleanAbsentAbsolute(destination)
	if err != nil {
		return nil, err
	}
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{capture.project.receiptID: capture.project.input}
	mounts := map[closuregraph.ID]string{capture.project.receiptID: "capture/project"}
	keys := sortedKeys(capture.tarballs)
	tarballArgs := []string{}
	for index, key := range keys {
		item := capture.tarballs[key]
		if err = item.handle.Recheck(); err != nil {
			return nil, fail(CodeIntegrityMismatch, "captured pnpm tarball changed before store derivation", map[string]string{"package": key})
		}
		mount := fmt.Sprintf("capture/tarballs/%04d.tgz", index)
		inputs[item.receiptID] = item.input
		mounts[item.receiptID] = mount
		tarballArgs = append(tarballArgs, relativeLogical("work/pnpm-store-project", mount))
	}
	// Keep derivation inside the sole writable admitted work-copy. The store is
	// detached, admitted, and frozen immediately after the process returns.
	storePath := "work/pnpm-store-project/.pnpm-store"
	manager := relativeLogical("work/pnpm-store-project", authority.Runtime.Manager.EntrypointRelativePath)
	args := []string{manager, "--store-dir", relativeLogical("work/pnpm-store-project", storePath), "--config.side-effects-cache=false", "store", "add"}
	args = append(args, tarballArgs...)
	call := invocation{Tool: authority.Runtime.Manager, CWD: "work/pnpm-store-project", Args: args, Environment: pnpmEnvironment(".home", relativeLogical("work/pnpm-store-project", storePath), ".home/npmrc"), Inputs: inputs, InputMounts: mounts, ReadRoots: append(values(mounts), authority.Runtime.Manager.ReadRoots...), WriteRoots: []string{"work/pnpm-store-project"}, WorkCopies: []closureexec.WorkCopy{{ReceiptID: capture.project.receiptID, Path: "work/pnpm-store-project", Retain: true}}}
	receipt, err := session.run(ctx, call, "pnpm-store-add")
	if err != nil {
		return nil, err
	}
	_ = receipt
	physical := filepath.Join(authority.ExecutionRoot, filepath.FromSlash(storePath))
	defer func() { _ = os.RemoveAll(filepath.Join(authority.ExecutionRoot, "work")) }()
	if err = scanAmbientSideEffects(physical); err != nil {
		return nil, err
	}
	if err = reconcilePNPMStoreIndexes(physical, capture.Graph.Packages); err != nil {
		return nil, err
	}
	if err = copyWritableTree(physical, dest); err != nil {
		return nil, err
	}
	files, err := inventoryStore(dest)
	if err != nil {
		return nil, err
	}
	if len(keys) > 0 && len(files) == 0 {
		return nil, fail(CodeOfflineInputMissing, "pnpm store derivation produced no bytes", nil)
	}
	receipts := make([]closuregraph.ID, 0, len(inputs))
	for id := range inputs {
		receipts = append(receipts, id)
	}
	sort.Slice(receipts, func(i, j int) bool { return receipts[i] < receipts[j] })
	fileValues := make([]any, len(files))
	for i, file := range files {
		fileValues[i] = map[string]any{"path": file.Path, "sha256": string(file.SHA256), "size": file.Size}
	}
	id, err := closuregraph.DomainID("pnpm-private-store-receipt-v1", map[string]any{"schema_id": "pnpm-private-store-receipt-v1", "input_receipt_ids": idsToStrings(receipts), "files": fileValues})
	if err != nil {
		return nil, err
	}
	if err = makeTreeReadOnly(dest); err != nil {
		return nil, err
	}
	tree, err := capture.store.CaptureTree("derived:pnpm-store:"+string(id), dest)
	if err != nil {
		return nil, err
	}
	derived, err := capture.store.AdmitTree(tree, "derived:pnpm-store:"+string(id), closureexec.AdmissionEvidence{PreviousCausalHead: session.operation.CurrentCausalHead(), ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: "pnpm-private-store-v1", DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: id})
	if err != nil {
		return nil, err
	}
	derivedID, err := derived.ID()
	if err != nil {
		return nil, err
	}
	return &PrivateStore{Receipt: StoreReceipt{SchemaID: "pnpm-private-store-receipt-v1", InputReceiptIDs: receipts, Files: files, ID: id}, root: dest, capture: capture, input: closureexec.AdmittedInput{Receipt: derived, Tree: tree}, inputID: derivedID}, nil
}

func reconcilePNPMStoreIndexes(storeRoot string, packages []Package) error {
	indexRoot := filepath.Join(storeRoot, "v10", "index")
	available := map[string][]string{}
	err := filepath.WalkDir(indexRoot, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == indexRoot || entry.IsDir() {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fail(CodeInputUndeclared, "pnpm store index contains a non-regular member", nil)
		}
		rel, relErr := filepath.Rel(indexRoot, current)
		if relErr != nil {
			return relErr
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) != 2 || len(parts[0]) != 2 || !strings.HasSuffix(parts[1], ".json") {
			return fail(CodeInputUndeclared, "pnpm store index path is outside the pinned grammar", map[string]string{"path": filepath.ToSlash(rel)})
		}
		prefix := parts[0] + strings.SplitN(parts[1], "-", 2)[0]
		available[prefix] = append(available[prefix], current)
		return nil
	})
	if err != nil {
		return fail(CodeOfflineInputMissing, "pnpm store add produced no closed package index", nil)
	}
	expected := map[string]bool{}
	for _, pkg := range packages {
		encoded := strings.TrimPrefix(pkg.Integrity, "sha512-")
		digest, decodeErr := base64.StdEncoding.DecodeString(encoded)
		if decodeErr != nil || len(digest) != 64 {
			return fail(CodeIntegrityMissing, "pnpm package integrity is not canonical sha512", map[string]string{"package": pkg.Key})
		}
		hexDigest := hex.EncodeToString(digest)[:64]
		candidates := available[hexDigest]
		if len(candidates) != 1 {
			return fail(CodeOfflineInputMissing, "pnpm store add did not bind one exact admitted package index", map[string]string{"package": pkg.Key})
		}
		payload, readErr := os.ReadFile(candidates[0]) // #nosec G304 -- exact store-add index below the task-private derived store.
		if readErr != nil {
			return readErr
		}
		var index map[string]any
		if json.Unmarshal(payload, &index) != nil || len(index) == 0 {
			return fail(CodeMetadataMismatch, "pnpm store-add package index is malformed", map[string]string{"package": pkg.Key})
		}
		packageID := strings.NewReplacer("\\", "+", "/", "+", ":", "+", "*", "+", "?", "+", "\"", "+", "<", "+", ">", "+", "|", "+").Replace(pkg.Key)
		target := filepath.Join(indexRoot, hexDigest[:2], hexDigest[2:]+"-"+packageID+".json")
		if err = os.WriteFile(target, payload, 0o600); err != nil {
			return err
		}
		expected[target] = true
	}
	for _, candidates := range available {
		for _, candidate := range candidates {
			if !expected[candidate] {
				if err = os.Remove(candidate); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Materialize replays the admitted project with a read-only private store and
// frozen/offline/scripts-disabled pnpm flags, then reconciles the result.
func Materialize(ctx context.Context, store *PrivateStore, request MaterializeRequest, authority *ExecutionContext) (*Materialization, error) {
	if store == nil || store.capture == nil || store.input.Tree == nil {
		return nil, fail(CodeInputUndeclared, "pnpm materialization authority is incomplete", nil)
	}
	// pnpm 10.33.0 materializes target-pruned snapshots that remain reachable
	// from the lock graph, but omits wholly unreachable snapshots. Reject that
	// narrower shape before starting install rather than pretending an absent
	// physical instance was reconciled.
	for _, snapshot := range store.capture.Graph.Snapshots {
		if !snapshot.Reachable {
			return nil, fail(CodeGraphIncomplete, "pnpm 10.33.0 cannot materialize an unreachable lock-superset snapshot", map[string]string{"snapshot": snapshot.Key})
		}
	}
	session, err := newRunnerSession(ctx, store.capture, authority)
	if err != nil {
		return nil, err
	}
	destination, err := cleanAbsentAbsolute(request.Destination)
	if err != nil {
		return nil, err
	}
	if request.WorkRoot == "" || !filepath.IsAbs(request.WorkRoot) {
		return nil, fail(CodeInputUndeclared, "pnpm materialization work root must be absolute", nil)
	}
	observed, err := inventoryStore(store.root)
	if err != nil {
		return nil, err
	}
	if !equalStoreFiles(observed, store.Receipt.Files) {
		return nil, fail(CodeIntegrityMismatch, "private pnpm store differs from its derivation receipt", nil)
	}
	if err = scanAmbientSideEffects(store.root); err != nil {
		return nil, err
	}
	if err = store.capture.project.tree.VerifyAtUse(); err != nil {
		return nil, err
	}
	projectPath := "work/pnpm-install-project"
	frozenStorePath := "capture/pnpm-private-store"
	storePath := "work/pnpm-install-store"
	manager := relativeLogical(projectPath, authority.Runtime.Manager.EntrypointRelativePath)
	relativeStore := relativeLogical(projectPath, storePath)
	args := []string{manager, "install", "--frozen-lockfile", "--offline", "--ignore-scripts", "--store-dir", relativeStore, "--config.side-effects-cache=false", "--config.hoist=false", "--config.public-hoist-pattern=", "--package-import-method=copy"}
	prod := []string{}
	if !store.capture.Graph.Target.IncludeDev {
		args = append(args, "--prod")
		prod = []string{"--prod"}
	}
	_ = prod
	inputs := map[closuregraph.ID]closureexec.AdmittedInput{store.capture.project.receiptID: store.capture.project.input, store.inputID: store.input}
	mounts := map[closuregraph.ID]string{store.capture.project.receiptID: "capture/project", store.inputID: frozenStorePath}
	readRoots := append(values(mounts), authority.Runtime.Manager.ReadRoots...)
	call := invocation{Tool: authority.Runtime.Manager, CWD: projectPath, Args: args, Environment: pnpmEnvironment(".home", relativeStore, ".home/npmrc"), Inputs: inputs, InputMounts: mounts, ReadRoots: readRoots, WriteRoots: []string{projectPath, storePath}, WorkCopies: []closureexec.WorkCopy{{ReceiptID: store.capture.project.receiptID, Path: projectPath, Retain: true}, {ReceiptID: store.inputID, Path: storePath, Retain: true}}}
	sort.Slice(call.WorkCopies, func(i, j int) bool { return call.WorkCopies[i].ReceiptID < call.WorkCopies[j].ReceiptID })
	receipt, err := session.run(ctx, call, "pnpm-install")
	if err != nil {
		return nil, err
	}
	retained := filepath.Join(authority.ExecutionRoot, filepath.FromSlash(projectPath))
	retainedStore := filepath.Join(authority.ExecutionRoot, filepath.FromSlash(storePath))
	defer func() { _ = os.RemoveAll(filepath.Join(authority.ExecutionRoot, "work")) }()
	if err = os.RemoveAll(filepath.Join(retained, ".home")); err != nil {
		return nil, err
	}
	if err = rejectGeneratedSideEffects(retained); err != nil {
		return nil, err
	}
	if err = reconcileWritableStoreOverlay(retainedStore, retained, store.Receipt.Files); err != nil {
		return nil, err
	}
	lock, err := os.ReadFile(filepath.Join(retained, "pnpm-lock.yaml")) // #nosec G304 -- exact lock basename below the retained task-private work copy.
	if err != nil || digestID(lock) != closuregraph.ID(store.capture.Graph.RawLockSHA256) {
		return nil, fail(CodeLockStale, "pnpm changed or removed frozen lock", nil)
	}
	if err = validateLocalRoots(retained, store.capture); err != nil {
		return nil, err
	}
	installed, err := validateMaterializedTree(ctx, retained, store.capture)
	if err != nil {
		return nil, err
	}
	if err = copyContainedTree(retained, destination); err != nil {
		return nil, err
	}
	if err = normalizePortableRuntimeLayout(destination, store.capture); err != nil {
		return nil, err
	}
	input, inputID, err := admitMaterializedTree(store.capture, destination, receipt)
	if err != nil {
		return nil, err
	}
	return &Materialization{Root: destination, StoreReceiptID: store.Receipt.ID, Receipt: receipt, MaterializedPackages: installed, capture: store.capture, input: input, inputID: inputID}, nil
}

func normalizePortableRuntimeLayout(projectRoot string, capture *Capture) error {
	snapshotRoots := map[string]string{}
	for _, snapshot := range capture.Graph.Snapshots {
		snapshotRoots[snapshot.Key] = filepath.Join(projectRoot, "node_modules", ".pnpm", pnpmSnapshotDirectory(snapshot.Key), "node_modules", filepath.FromSlash(snapshot.Name))
	}
	localRoots := map[string]string{}
	for _, local := range capture.Graph.LocalRoots {
		root := projectRoot
		if local.Path != "." {
			root = filepath.Join(projectRoot, filepath.FromSlash(local.Path))
		}
		localRoots[localRootKey(local.Path)] = root
	}
	for key, root := range localRoots {
		if err := hydratePortableDependencies(root, key, capture, snapshotRoots, localRoots, map[string]bool{key: true}); err != nil {
			return err
		}
	}
	return nil
}

func hydratePortableDependencies(packageRoot, ownerKey string, capture *Capture, snapshotRoots, localRoots map[string]string, ancestors map[string]bool) error {
	for _, edge := range capture.Graph.Edges {
		if edge.From != ownerKey || !edge.Selected {
			continue
		}
		// A runtime cycle resolves through the nearest already materialized
		// ancestor, matching Node's upward node_modules lookup without links.
		if ancestors[edge.To] {
			continue
		}
		source := snapshotRoots[edge.To]
		if strings.HasPrefix(edge.To, "local:") {
			source = localRoots[edge.To]
		}
		if source == "" {
			return fail(CodeGraphIncomplete, "portable pnpm runtime dependency has no exact source", map[string]string{"owner": ownerKey, "dependency": edge.Name, "target": edge.To})
		}
		target := filepath.Join(packageRoot, "node_modules", filepath.FromSlash(edge.Name))
		if err := os.RemoveAll(target); err != nil {
			return err
		}
		if err := copyContainedTree(source, target); err != nil {
			return err
		}
		next := copyBoolMap(ancestors)
		next[edge.To] = true
		if err := hydratePortableDependencies(target, edge.To, capture, snapshotRoots, localRoots, next); err != nil {
			return err
		}
	}
	return nil
}

func copyBoolMap(input map[string]bool) map[string]bool {
	result := make(map[string]bool, len(input)+1)
	for key, value := range input {
		result[key] = value
	}
	return result
}

func reconcileWritableStoreOverlay(storeRoot, projectRoot string, expected []StoreFile) error {
	registry := filepath.Join(storeRoot, "v10", "projects")
	entries, err := os.ReadDir(registry)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fail(CodeInputUndeclared, "pnpm writable store registry cannot be read", nil)
	}
	for _, entry := range entries {
		link := filepath.Join(registry, entry.Name())
		info, statErr := os.Lstat(link)
		if statErr != nil || info.Mode()&fs.ModeSymlink == 0 {
			return fail(CodeInputUndeclared, "pnpm writable store registry contains an undeclared member", map[string]string{"entry": entry.Name()})
		}
		target, evalErr := filepath.EvalSymlinks(link)
		if evalErr != nil {
			return fail(CodeInputUndeclared, "pnpm writable store registry link cannot be resolved", map[string]string{"entry": entry.Name()})
		}
		project, evalErr := filepath.EvalSymlinks(projectRoot)
		if evalErr != nil || target != project {
			return fail(CodeInputUndeclared, "pnpm writable store registry targets an undeclared project", map[string]string{"entry": entry.Name()})
		}
	}
	if err = os.RemoveAll(registry); err != nil {
		return err
	}
	if err = scanAmbientSideEffects(storeRoot); err != nil {
		return err
	}
	observed, err := inventoryStore(storeRoot)
	if err != nil {
		return err
	}
	if !equalStoreFiles(observed, expected) {
		return fail(CodeIntegrityMismatch, "pnpm writable store overlay changed frozen content", nil)
	}
	return nil
}

func newRunnerSession(ctx context.Context, capture *Capture, authority *ExecutionContext) (*runnerSession, error) {
	if capture == nil || authority == nil || authority.Executor == nil || authority.Recheck == nil {
		return nil, fail(CodeDerivationUnauthorized, "pnpm executor authority is absent", nil)
	}
	if authority.Runtime.C0Checkpoint == nil || authority.Runtime.Manager.EntrypointRelativePath == "" || authority.Runtime.Manager.ExecutableRelativePath != authority.Runtime.Node.ExecutableRelativePath || authority.Runtime.Manager.ExecutableSHA256 != authority.Runtime.Node.ExecutableSHA256 {
		return nil, fail(CodeDerivationUnauthorized, "pnpm must execute through exact bound Node runtime", nil)
	}
	if strings.TrimSpace(authority.Runtime.Manager.VersionOutput) != SupportedPNPMVersion {
		return nil, fail(CodeRuntimeIdentityChanged, "pnpm manager release is outside the pinned profile", map[string]string{"expected": SupportedPNPMVersion, "observed": strings.TrimSpace(authority.Runtime.Manager.VersionOutput)})
	}
	root, err := filepath.Abs(authority.ExecutionRoot)
	if err != nil || root != filepath.Clean(authority.ExecutionRoot) {
		return nil, fail(CodeDerivationUnauthorized, "pnpm execution root is invalid", nil)
	}
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return nil, fail(CodeDerivationUnauthorized, "pnpm execution root is not a real directory", nil)
	}
	operation, err := authority.Executor.Preflight(ctx)
	if err != nil {
		return nil, fail(CodeDerivationUnauthorized, "pnpm executor preflight failed: "+err.Error(), nil)
	}
	wantC0, err := nodesource.NewC0Checkpoint(capture.NodeCapture, authority.Selection, authority.Runtime)
	if err != nil {
		return nil, err
	}
	c0ID, err := wantC0.ID()
	if err != nil {
		return nil, err
	}
	observed, err := authority.Runtime.C0Checkpoint.ID()
	if err != nil || observed != c0ID {
		return nil, fail(CodeDerivationUnauthorized, "pnpm C0 authority drifted", nil)
	}
	evaluator := closuregraph.ConditionEvaluatorFunc{EvaluatorID: "pnpm-platform-selector-v1", EvaluateFunc: evaluatePlatform}
	bundle, plan, err := nodesource.Close(capture.NodeCapture, authority.Selection, authority.Runtime, []closuregraph.ConditionEvaluator{evaluator}, operation.Binding().ExecutionPolicyID)
	if err != nil {
		return nil, err
	}
	planID, err := plan.ID()
	if err != nil {
		return nil, err
	}
	gotPlanID, err := authority.BuildPlan.ID()
	if err != nil || gotPlanID != planID {
		return nil, fail(CodeDerivationUnauthorized, "pnpm C5 plan drifted", nil)
	}
	actions := map[string]runtimeAction{}
	for _, id := range plan.ActionNodeIDs {
		for _, node := range bundle.Records.CaptureNodes {
			nodeID, _ := node.ID()
			if nodeID == id && node.Kind == closuregraph.NodeAction {
				payload := node.Payload.(closuregraph.ActionPayload)
				actions[payload.ActionSubtype] = runtimeAction{ID: id, Payload: payload}
			}
		}
	}
	for _, name := range []string{"pnpm-store-add", "pnpm-install", "node-invoke"} {
		if actions[name].ID == "" {
			return nil, fail(CodeDerivationUnauthorized, "pnpm operation is absent from C5", map[string]string{"subtype": name})
		}
	}
	return &runnerSession{authority: authority, operation: operation, bundle: bundle, c0ID: c0ID, planID: planID, actions: actions}, nil
}
func evaluatePlatform(condition closuregraph.Condition, input closuregraph.EvaluationInput) (bool, error) {
	for _, clause := range strings.Split(condition.Expression, ";") {
		parts := strings.SplitN(clause, "=", 2)
		if len(parts) != 2 {
			return false, fail(CodeGraphIncomplete, "pnpm platform condition is malformed", nil)
		}
		if !selectorMatches(strings.Split(parts[1], ","), input.Selection.Markers[parts[0]]) {
			return false, nil
		}
	}
	return true, nil
}

func (session *runnerSession) run(ctx context.Context, call invocation, subtype string) (closureexec.DerivationReceipt, error) {
	action := session.actions[subtype]
	if action.ID == "" || action.Payload.Network != "none" || action.Payload.ActionSubtype != subtype {
		return closureexec.DerivationReceipt{}, fail(CodeDerivationUnauthorized, "pnpm concrete action is absent from C5", nil)
	}
	if err := session.operation.Revalidate(ctx); err != nil {
		return closureexec.DerivationReceipt{}, fail(CodeDerivationUnauthorized, "pnpm provider drifted before execution: "+err.Error(), nil)
	}
	toolID, err := session.toolNodeID(call.Tool, action.ID)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	ids := make([]closuregraph.ID, 0, len(call.Inputs))
	for id := range call.Inputs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	mounts := make([]closureexec.InputMount, len(ids))
	reads := uniqueSorted(call.ReadRoots)
	for i, id := range ids {
		mount := call.InputMounts[id]
		if mount == "" {
			return closureexec.DerivationReceipt{}, fail(CodeDerivationUnauthorized, "pnpm input mount is absent", nil)
		}
		mounts[i] = closureexec.InputMount{ReceiptID: id, Path: mount}
	}
	stdoutPath := "evidence/" + subtype + ".stdout"
	stdoutID, err := closuregraph.DomainID("pnpm-operation-stdout-v1", map[string]any{"action_node_id": string(action.ID), "plan": string(session.planID), "subtype": subtype})
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	requirements := []closureexec.EvidenceRequirement{{Path: stdoutPath, SchemaID: "pnpm-operation-stdout-v1", ArtifactManifestID: stdoutID}}
	evidenceID, err := evidenceSchemaID(requirements)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	limits := closureexec.ResourceLimits{OutputBytes: 64 << 20, ReadBytes: 2 << 30, WriteBytes: 2 << 30, WallTimeMillis: 300000, ProcessCount: 256}
	limitID, err := limits.ID()
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	environment := copyMap(call.Environment)
	environment["CURATOR_OUTPUT_ROOT"] = "output"
	writes := uniqueSorted(append(append([]string{}, call.WriteRoots...), stdoutPath))
	permit := closureexec.DerivationPermit{SchemaID: closureexec.SchemaDerivationPermit, PreviousCausalHead: session.operation.CurrentCausalHead(), InvocationKey: "pnpm-c5:" + string(session.planID) + ":" + string(action.ID) + ":" + subtype, InvocationSubtype: closureexec.DerivationMetadata, AdmittedInputReceiptIDs: ids, InputMounts: mounts, WorkCopies: call.WorkCopies, C0CheckpointID: session.c0ID, ToolchainNodeID: toolID, ToolchainFingerprint: call.Tool.Fingerprint, ExecutableSHA256: call.Tool.ExecutableSHA256, Executable: call.Tool.ExecutableRelativePath, CWD: call.CWD, Argv: call.Args, Environment: environment, HostID: session.authority.Selection.PlatformRoles[closuregraph.PlatformTarget], TargetID: session.authority.Selection.PlatformRoles[closuregraph.PlatformTarget], AllowedProcesses: []string{call.Tool.ExecutableRelativePath}, ReadRoots: reads, WriteRoots: writes, StdoutEvidencePath: stdoutPath, ExpectedEvidence: requirements, Network: "none", RecheckRule: "immediate-exact-v1", ResourceLimits: limits, ResourceLimitID: limitID, EvidenceSchemaID: evidenceID}
	permitID, err := session.operation.Commit(permit)
	if err != nil {
		return closureexec.DerivationReceipt{}, err
	}
	receipt, err := session.operation.Execute(ctx, permitID, func(ctx context.Context) (closureexec.ToolchainIdentity, error) {
		return session.authority.Recheck(ctx, call.Tool)
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
	c0 := map[closuregraph.ID]bool{}
	for _, id := range session.authority.Runtime.C0Checkpoint.Payload.(closuregraph.C0ProfilePayload).EvidenceToolchainNodeIDs {
		c0[id] = true
	}
	for _, edge := range session.bundle.Records.BindingEdges {
		if edge.Kind != closuregraph.EdgeUsesTool || edge.FromNodeID != actionID || !c0[edge.ToNodeID] {
			continue
		}
		for _, node := range session.bundle.Records.BindingNodes {
			id, _ := node.ID()
			if id != edge.ToNodeID || node.Kind != closuregraph.NodeToolchainComponent {
				continue
			}
			payload := node.Payload.(closuregraph.ToolchainComponentPayload)
			if payload.ContentFingerprint != tool.Fingerprint || len(payload.LinkFingerprintIDs) != 1 || payload.LinkFingerprintIDs[0] != tool.ExecutableSHA256 {
				return "", fail(CodeRuntimeIdentityChanged, "pnpm tool identity changed", nil)
			}
			return id, nil
		}
	}
	return "", fail(CodeDerivationUnauthorized, "pnpm C5 action lacks exact C0 tool", nil)
}

func validateMaterializedTree(ctx context.Context, root string, capture *Capture) ([]string, error) {
	virtual := filepath.Join(root, "node_modules", ".pnpm")
	entries, err := os.ReadDir(virtual)
	if err != nil {
		return nil, fail(CodeOfflineInputMissing, "pnpm virtual store is absent", nil)
	}
	expected := map[string]Snapshot{}
	roots := map[string]string{}
	for _, snapshot := range capture.Graph.Snapshots {
		directory := pnpmSnapshotDirectory(snapshot.Key)
		if previous, exists := expected[directory]; exists {
			return nil, fail(CodeGraphIncomplete, "selected pnpm snapshots collide in the virtual-store layout", map[string]string{"snapshot": snapshot.Key, "other": previous.Key, "directory": directory})
		}
		expected[directory] = snapshot
		roots[snapshot.Key] = filepath.Join(virtual, directory, "node_modules", filepath.FromSlash(snapshot.Name))
	}
	observed := map[string]bool{}
	for _, entry := range entries {
		if entry.Name() == "lock.yaml" && entry.Type().IsRegular() {
			payload, readErr := os.ReadFile(filepath.Join(virtual, entry.Name())) // #nosec G304 -- entry name is exactly lock.yaml below the owned virtual store.
			if readErr != nil {
				return nil, fail(CodeGraphIncomplete, "pnpm virtual-store metadata cannot be read", map[string]string{"path": entry.Name()})
			}
			if _, parseErr := decodeClosedYAML(payload); parseErr != nil {
				return nil, fail(CodeGraphIncomplete, "pnpm virtual-store metadata is not closed YAML", map[string]string{"path": entry.Name()})
			}
			continue
		}
		snapshot, ok := expected[entry.Name()]
		if !ok || !entry.IsDir() || entry.Type()&fs.ModeSymlink != 0 {
			return nil, fail(CodeGraphIncomplete, "pnpm virtual store contains an unclaimed entry", map[string]string{"entry": entry.Name()})
		}
		if observed[entry.Name()] {
			return nil, fail(CodeGraphIncomplete, "pnpm virtual store contains a duplicate snapshot entry", map[string]string{"entry": entry.Name()})
		}
		observed[entry.Name()] = true
		if err := validateSnapshotInstance(ctx, root, virtual, snapshot, roots, capture); err != nil {
			return nil, err
		}
	}
	for directory, snapshot := range expected {
		if !observed[directory] {
			return nil, fail(CodeGraphIncomplete, "pnpm virtual store is missing an exact snapshot entry", map[string]string{"snapshot": snapshot.Key, "directory": directory})
		}
	}
	if err := validateImporterLayouts(root, capture, roots); err != nil {
		return nil, err
	}
	result := []string{}
	for _, snapshot := range expected {
		if snapshot.Selected {
			result = append(result, snapshot.Key)
		}
	}
	sort.Strings(result)
	return result, nil
}

func validateImporterLayouts(projectRoot string, capture *Capture, snapshotRoots map[string]string) error {
	for _, local := range capture.Graph.LocalRoots {
		ownerRoot := projectRoot
		if local.Path != "." {
			ownerRoot = filepath.Join(projectRoot, filepath.FromSlash(local.Path))
		}
		expected := map[string]string{}
		for _, edge := range capture.Graph.Edges {
			if edge.From != localRootKey(local.Path) || !edge.Selected {
				continue
			}
			target := snapshotRoots[edge.To]
			if strings.HasPrefix(edge.To, "local:") {
				path := strings.TrimPrefix(edge.To, "local:")
				target = projectRoot
				if path != "." {
					target = filepath.Join(projectRoot, filepath.FromSlash(path))
				}
			}
			if target == "" {
				return fail(CodeGraphIncomplete, "pnpm importer dependency has no exact materialized target", map[string]string{"importer": local.Path, "dependency": edge.Name, "target": edge.To})
			}
			expected[edge.Name] = target
		}
		if err := validateDirectNodeModules(filepath.Join(ownerRoot, "node_modules"), local.Path == ".", expected); err != nil {
			return fail(CodeGraphIncomplete, "pnpm importer node_modules layout is not closed: "+err.Error(), map[string]string{"importer": local.Path})
		}
	}
	return nil
}

func validateDirectNodeModules(nodeModules string, root bool, expected map[string]string) error {
	entries, err := os.ReadDir(nodeModules)
	if errors.Is(err, fs.ErrNotExist) && !root && len(expected) == 0 {
		return nil
	}
	if err != nil {
		return err
	}
	observed := map[string]string{}
	modulesMetadata := false
	workspaceMetadata := false
	for _, entry := range entries {
		memberPath := filepath.Join(nodeModules, entry.Name())
		switch entry.Name() {
		case ".pnpm":
			info, statErr := os.Lstat(memberPath)
			if !root || statErr != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
				return fmt.Errorf("unowned virtual-store member %q", entry.Name())
			}
			continue
		case ".modules.yaml":
			if !root || modulesMetadata || !entry.Type().IsRegular() {
				return fmt.Errorf("unowned manager metadata %q", entry.Name())
			}
			payload, readErr := os.ReadFile(memberPath) // #nosec G304 -- exact manager metadata basename below the owned root node_modules.
			if readErr != nil {
				return readErr
			}
			if _, parseErr := decodeClosedYAML(payload); parseErr != nil {
				return fmt.Errorf("manager metadata is not closed YAML: %w", parseErr)
			}
			modulesMetadata = true
			continue
		case ".pnpm-workspace-state-v1.json":
			if !root || workspaceMetadata || !entry.Type().IsRegular() {
				return fmt.Errorf("unowned manager metadata %q", entry.Name())
			}
			payload, readErr := os.ReadFile(memberPath) // #nosec G304 -- exact manager metadata basename below the owned root node_modules.
			if readErr != nil {
				return readErr
			}
			var state map[string]any
			if json.Unmarshal(payload, &state) != nil || state == nil {
				return fmt.Errorf("workspace metadata is not closed JSON")
			}
			workspaceMetadata = true
			continue
		}
		if strings.HasPrefix(entry.Name(), "@") {
			info, statErr := os.Lstat(memberPath)
			if statErr != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
				return fmt.Errorf("scope container %q is not an owned directory", entry.Name())
			}
			children, readErr := os.ReadDir(memberPath)
			if readErr != nil {
				return readErr
			}
			if len(children) == 0 {
				return fmt.Errorf("scope container %q is empty", entry.Name())
			}
			for _, child := range children {
				observed[entry.Name()+"/"+child.Name()] = filepath.Join(memberPath, child.Name())
			}
			continue
		}
		observed[entry.Name()] = memberPath
	}
	if root && (!modulesMetadata || !workspaceMetadata) {
		return fmt.Errorf("required pnpm manager metadata is missing")
	}
	if len(observed) != len(expected) {
		return fmt.Errorf("direct dependency members differ: expected %v observed %v", sortedKeys(expected), sortedKeys(observed))
	}
	for name, target := range expected {
		link, ok := observed[name]
		if !ok {
			return fmt.Errorf("direct dependency link %q is missing", name)
		}
		info, statErr := os.Lstat(link)
		if statErr != nil || info.Mode()&fs.ModeSymlink == 0 {
			return fmt.Errorf("direct dependency member %q is not a link", name)
		}
		actual, evalErr := filepath.EvalSymlinks(link)
		if evalErr != nil {
			return fmt.Errorf("direct dependency link %q cannot be resolved", name)
		}
		want, evalErr := filepath.EvalSymlinks(target)
		if evalErr != nil || actual != want {
			return fmt.Errorf("direct dependency link %q targets the wrong snapshot/local identity", name)
		}
	}
	return nil
}

func validateSnapshotInstance(ctx context.Context, projectRoot, virtual string, snapshot Snapshot, roots map[string]string, capture *Capture) error {
	base := filepath.Join(virtual, pnpmSnapshotDirectory(snapshot.Key), "node_modules")
	members, err := readVirtualPackageMembers(base)
	if err != nil {
		return fail(CodeGraphIncomplete, "pnpm snapshot node_modules layout cannot be read: "+err.Error(), map[string]string{"snapshot": snapshot.Key})
	}
	expectedLinks := map[string]string{}
	for _, edge := range capture.Graph.Edges {
		// The virtual store is a physical projection of the frozen lock
		// superset, not of the active target graph. pnpm retains declared links
		// for every materialized snapshot, including target-pruned instances, so
		// reconcile every resolved snapshot edge here. Selection remains
		// authoritative for importer links, runtime reachability, and the returned
		// active package set.
		if edge.From != snapshot.Key || edge.To == "" {
			continue
		}
		if strings.HasPrefix(edge.To, "local:") {
			expectedLinks[edge.Name] = filepath.Join(projectRoot, filepath.FromSlash(strings.TrimPrefix(edge.To, "local:")))
			continue
		}
		target, ok := roots[edge.To]
		if !ok {
			return fail(CodeGraphIncomplete, "pnpm lock-superset snapshot edge has no exact materialized target", map[string]string{"snapshot": snapshot.Key, "dependency": edge.Name, "target": edge.To})
		}
		expectedLinks[edge.Name] = target
	}
	packageRoot, ok := members[snapshot.Name]
	if !ok {
		return fail(CodeGraphIncomplete, "pnpm snapshot package root is missing", map[string]string{"snapshot": snapshot.Key, "package": snapshot.Name})
	}
	info, err := os.Lstat(packageRoot)
	if err != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return fail(CodeGraphIncomplete, "pnpm snapshot package root is not an owned directory", map[string]string{"snapshot": snapshot.Key, "package": snapshot.Name})
	}
	delete(members, snapshot.Name)
	if len(members) != len(expectedLinks) {
		return fail(CodeGraphIncomplete, "pnpm snapshot contains missing or unclaimed dependency links", map[string]string{"snapshot": snapshot.Key, "expected": fmt.Sprint(len(expectedLinks)), "observed": fmt.Sprint(len(members))})
	}
	for name, expectedTarget := range expectedLinks {
		link, exists := members[name]
		if !exists {
			return fail(CodeGraphIncomplete, "pnpm snapshot dependency link is missing", map[string]string{"snapshot": snapshot.Key, "dependency": name})
		}
		linkInfo, err := os.Lstat(link)
		if err != nil || linkInfo.Mode()&fs.ModeSymlink == 0 {
			return fail(CodeGraphIncomplete, "pnpm snapshot dependency member is not an exact link", map[string]string{"snapshot": snapshot.Key, "dependency": name})
		}
		actualTarget, err := filepath.EvalSymlinks(link)
		if err != nil {
			return fail(CodeGraphIncomplete, "pnpm snapshot dependency link cannot be resolved", map[string]string{"snapshot": snapshot.Key, "dependency": name})
		}
		expectedReal, err := filepath.EvalSymlinks(expectedTarget)
		if err != nil || actualTarget != expectedReal {
			return fail(CodeGraphIncomplete, "pnpm snapshot dependency link targets the wrong peer/package context", map[string]string{"snapshot": snapshot.Key, "dependency": name})
		}
	}
	payload, err := os.ReadFile(filepath.Join(packageRoot, "package.json")) // #nosec G304 -- exact metadata basename below a validated package root.
	if err != nil {
		return fail(CodeGraphIncomplete, "pnpm snapshot package metadata is missing", map[string]string{"snapshot": snapshot.Key})
	}
	var manifest packageManifest
	if err = json.Unmarshal(payload, &manifest); err != nil || manifest.Name != snapshot.Name || manifest.Version != snapshot.Version {
		return fail(CodeMetadataMismatch, "pnpm snapshot package metadata differs from exact snapshot identity", map[string]string{"snapshot": snapshot.Key})
	}
	item, ok := capture.tarballs[snapshot.PackageKey]
	if !ok {
		return fail(CodeGraphIncomplete, "pnpm snapshot has no admitted package inventory", map[string]string{"snapshot": snapshot.Key})
	}
	files, err := inventoryPackage(packageRoot)
	if err != nil {
		return err
	}
	if !equalPackageFiles(files, item.files) {
		return fail(CodeIntegrityMismatch, "materialized pnpm package differs from its admitted and patched expected inventory", map[string]string{"snapshot": snapshot.Key, "package": snapshot.PackageKey})
	}
	descriptor := artifactpolicy.Descriptor{AdapterID: ProfileID, ProfileID: artifactpolicy.ProfileNodeV1, Manager: "pnpm", PackageName: manifest.Name, PackageVersion: manifest.Version}
	probe, probeErr := capture.policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: packageRoot, VirtualRoot: "materialized/" + snapshot.Key})
	if probeErr != nil && artifactpolicy.ErrorCode(probeErr) != artifactpolicy.CodeOriginUnverified {
		return probeErr
	}
	treeDigest := closuregraph.ID(probe.Manifest.RawPayload.SHA256)
	if !treeDigest.Valid() {
		return fail(CodeIntegrityMismatch, "materialized pnpm package tree identity is unavailable", map[string]string{"snapshot": snapshot.Key})
	}
	descriptor.Origin = artifactpolicy.OriginEvidence{Locator: "materialized:" + snapshot.Key, ImmutableID: string(treeDigest), LockRecord: capture.Graph.LockDigest, ChecksumSHA256: string(treeDigest), Verified: true}
	_, err = capture.policy.AdmitDependencyDirectory(ctx, artifactpolicy.DirectoryRequest{Descriptor: descriptor, Root: packageRoot, VirtualRoot: "materialized/" + snapshot.Key})
	return err
}

func readVirtualPackageMembers(root string) (map[string]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, entry := range entries {
		memberPath := filepath.Join(root, entry.Name())
		if strings.HasPrefix(entry.Name(), "@") {
			info, err := os.Lstat(memberPath)
			if err != nil || !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
				return nil, fmt.Errorf("scope container %q is not an owned directory", entry.Name())
			}
			scoped, err := os.ReadDir(memberPath)
			if err != nil {
				return nil, err
			}
			if len(scoped) == 0 {
				return nil, fmt.Errorf("scope container %q is empty", entry.Name())
			}
			for _, child := range scoped {
				name := entry.Name() + "/" + child.Name()
				if _, exists := result[name]; exists {
					return nil, fmt.Errorf("duplicate package member %q", name)
				}
				result[name] = filepath.Join(memberPath, child.Name())
			}
			continue
		}
		if _, exists := result[entry.Name()]; exists {
			return nil, fmt.Errorf("duplicate package member %q", entry.Name())
		}
		result[entry.Name()] = memberPath
	}
	return result, nil
}

const pnpmVirtualStoreDirMaxLength = 120

func pnpmSnapshotDirectory(snapshotKey string) string {
	filename := strings.NewReplacer("\\", "+", "/", "+", ":", "+", "*", "+", "?", "+", "\"", "+", "<", "+", ">", "+", "|", "+", "#", "+").Replace(snapshotKey)
	if strings.Contains(filename, "(") {
		filename = strings.TrimSuffix(filename, ")")
		filename = strings.NewReplacer(")(", "_", "(", "_", ")", "_").Replace(filename)
	}
	if len(filename) > pnpmVirtualStoreDirMaxLength || (filename != strings.ToLower(filename) && !strings.HasPrefix(filename, "file+")) {
		digest := sha256.Sum256([]byte(filename))
		prefixLength := min(len(filename), pnpmVirtualStoreDirMaxLength-33)
		filename = filename[:prefixLength] + "_" + hex.EncodeToString(digest[:])[:32]
	}
	return filename
}

func inventoryPackage(root string) ([]packageFile, error) {
	files := []packageFile{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if current == root || entry.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, current)
		if entry.Type()&fs.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fail(CodeInputUndeclared, "materialized package contains non-regular member", map[string]string{"path": filepath.ToSlash(rel)})
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the owned materialized package root.
		if err != nil {
			return err
		}
		info, _ := entry.Info()
		files = append(files, packageFile{Path: filepath.ToSlash(rel), SHA256: digestID(payload), Size: int64(len(payload)), Executable: info.Mode()&0o111 != 0})
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, err
}
func rejectGeneratedSideEffects(root string) error {
	sideEffects := filepath.Join(root, "node_modules", ".pnpm", "side-effects")
	if _, err := os.Lstat(sideEffects); err == nil {
		return fail(CodeHookUndeclared, "pnpm materialization produced side-effects cache", nil)
	}
	return nil
}
func pnpmEnvironment(home, store, userConfig string) map[string]string {
	return map[string]string{"HOME": home, "PNPM_HOME": home, "CI": "true", "NPM_CONFIG_USERCONFIG": userConfig, "NPM_CONFIG_OFFLINE": "true", "NPM_CONFIG_IGNORE_SCRIPTS": "true", "NPM_CONFIG_SIDE_EFFECTS_CACHE": "false", "COREPACK_ENABLE_DOWNLOAD_PROMPT": "0", "NO_PROXY": "*", "no_proxy": "*", "CURATOR_PNPM_STORE": store}
}
func cleanAbsentAbsolute(value string) (string, error) {
	absolute, err := filepath.Abs(value)
	if err != nil || value == "" || absolute != filepath.Clean(value) {
		return "", fail(CodeInputUndeclared, "destination must be absolute and clean", map[string]string{"path": value})
	}
	if _, err = os.Lstat(absolute); !errors.Is(err, fs.ErrNotExist) {
		return "", fail(CodeInputUndeclared, "destination must be absent", map[string]string{"path": value})
	}
	return absolute, nil
}
func copyWritableTree(source, destination string) error {
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if current == source {
			return nil
		}
		rel, _ := filepath.Rel(source, current)
		target := filepath.Join(destination, rel)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeInputUndeclared, "derived pnpm store contains link", map[string]string{"path": filepath.ToSlash(rel)})
		}
		if entry.IsDir() {
			return os.Mkdir(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return fail(CodeInputUndeclared, "derived pnpm store contains special node", nil)
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the protected local-root comparison tree.
		if err != nil {
			return err
		}
		return os.WriteFile(target, payload, 0o600)
	})
}
func copyContainedTree(source, destination string) error {
	realRoot, err := filepath.EvalSymlinks(source)
	if err != nil {
		return err
	}
	return copyContainedNode(realRoot, source, destination, map[string]bool{})
}

func copyContainedNode(realRoot, source, destination string, active map[string]bool) error {
	realSource, err := filepath.EvalSymlinks(source)
	if err != nil {
		return err
	}
	inside, err := filepath.Rel(realRoot, realSource)
	if err != nil || inside == ".." || strings.HasPrefix(inside, ".."+string(filepath.Separator)) {
		return fail(CodeLocalPathEscape, "materialized pnpm link escapes project", map[string]string{"path": source})
	}
	if active[realSource] {
		return fail(CodeInputUndeclared, "materialized pnpm link cycle is unsupported", map[string]string{"path": source})
	}
	active[realSource] = true
	defer delete(active, realSource)
	info, err := os.Stat(realSource)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		payload, err := os.ReadFile(realSource)
		if err != nil {
			return err
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
		if err = copyContainedNode(realRoot, filepath.Join(realSource, entry.Name()), filepath.Join(destination, entry.Name()), active); err != nil {
			return err
		}
	}
	return nil
}

func validateLocalRoots(materialized string, capture *Capture) error {
	for _, local := range capture.Graph.LocalRoots {
		item := capture.localRoots[local.Path]
		protected, err := item.tree.ProtectedPath()
		if err != nil {
			return err
		}
		actualRoot := materialized
		if local.Path != "." {
			actualRoot = filepath.Join(materialized, filepath.FromSlash(local.Path))
		}
		expected, err := inventorySourceFiles(protected)
		if err != nil {
			return err
		}
		actual, err := inventorySourceFiles(actualRoot)
		if err != nil {
			return err
		}
		if !equalStoreFiles(expected, actual) {
			return fail(CodeIntegrityMismatch, "materialized pnpm local root differs from admitted source", map[string]string{"path": local.Path})
		}
	}
	return nil
}
func inventorySourceFiles(root string) ([]StoreFile, error) {
	files := []StoreFile{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if current == root {
			return nil
		}
		rel, _ := filepath.Rel(root, current)
		logical := filepath.ToSlash(rel)
		if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == ".pnpm" || entry.Name() == ".pnpm-store") {
			return filepath.SkipDir
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fail(CodeInputUndeclared, "local pnpm source contains link after replay", map[string]string{"path": logical})
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fail(CodeInputUndeclared, "local pnpm source contains special node after replay", map[string]string{"path": logical})
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the private store root.
		if err != nil {
			return err
		}
		files = append(files, StoreFile{Path: logical, SHA256: digestID(payload), Size: int64(len(payload))})
		return nil
	})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, err
}
func inventoryStore(root string) ([]StoreFile, error) {
	files := []StoreFile{}
	err := filepath.WalkDir(root, func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if current == root || entry.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, current)
		if entry.Type()&fs.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fail(CodeInputUndeclared, "private pnpm store contains non-regular member", map[string]string{"path": filepath.ToSlash(rel)})
		}
		payload, err := os.ReadFile(current) // #nosec G304 -- WalkDir supplies a member below the private derived store root.
		if err != nil {
			return err
		}
		files = append(files, StoreFile{Path: filepath.ToSlash(rel), SHA256: digestID(payload), Size: int64(len(payload))})
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
			return os.Chmod(current, 0o500) // #nosec G302 -- read-only directories require owner execute for traversal.
		}
		return os.Chmod(current, 0o400)
	})
}
func admitMaterializedTree(capture *Capture, root string, derivation closureexec.DerivationReceipt) (closureexec.AdmittedInput, closuregraph.ID, error) {
	files, err := inventoryStore(root)
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	values := make([]any, len(files))
	for i, file := range files {
		values[i] = map[string]any{"path": file.Path, "sha256": string(file.SHA256), "size": file.Size}
	}
	manifestID, err := closuregraph.DomainID("pnpm-materialized-tree-v1", map[string]any{"profile": ProfileID, "files": values})
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	tree, err := capture.store.CaptureTree("derived:pnpm-materialized:"+string(manifestID), root)
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	receipt, err := capture.store.AdmitTree(tree, "derived:pnpm-materialized:"+string(manifestID), closureexec.AdmissionEvidence{PreviousCausalHead: string(derivation.NextCausalHead), ArtifactPolicyID: artifactpolicy.PolicyID, SourceProfileID: "pnpm-materialized-v1", DetectorRegistryID: artifactpolicy.DetectorRegistryID, LimitVectorID: artifactpolicy.LimitVectorID, ArtifactManifestID: manifestID})
	if err != nil {
		return closureexec.AdmittedInput{}, "", err
	}
	id, err := receipt.ID()
	return closureexec.AdmittedInput{Receipt: receipt, Tree: tree}, id, err
}
func evidenceSchemaID(values []closureexec.EvidenceRequirement) (closuregraph.ID, error) {
	items := make([]any, len(values))
	for i, v := range values {
		items[i] = map[string]any{"artifact_manifest_id": string(v.ArtifactManifestID), "path": v.Path, "schema_id": v.SchemaID}
	}
	return closuregraph.DomainID("curator-derivation-evidence-schema-v1", map[string]any{"requirements": items})
}
func relativeLogical(from, to string) string {
	value, err := filepath.Rel(filepath.FromSlash(from), filepath.FromSlash(to))
	if err != nil {
		return ""
	}
	return filepath.ToSlash(value)
}
func values(input map[closuregraph.ID]string) []string {
	result := make([]string, 0, len(input))
	for _, value := range input {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
func uniqueSorted(input []string) []string {
	result := append([]string(nil), input...)
	sort.Strings(result)
	out := result[:0]
	for _, value := range result {
		if len(out) == 0 || out[len(out)-1] != value {
			out = append(out, value)
		}
	}
	return out
}
func copyMap(input map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range input {
		result[key] = value
	}
	return result
}
func idsToStrings(input []closuregraph.ID) []string {
	result := make([]string, len(input))
	for i, value := range input {
		result[i] = string(value)
	}
	return result
}
func equalStoreFiles(a, b []StoreFile) bool {
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
func equalPackageFiles(a, b []packageFile) bool {
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
