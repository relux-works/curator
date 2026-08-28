package buildrepo

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/buildmeta"
	"github.com/relux-works/curator/internal/closureexec"
	"github.com/relux-works/curator/internal/registry"
)

const (
	// CodeDescriptorInvalid reports a rejected skill-build.json document.
	CodeDescriptorInvalid = "build_repository_descriptor_invalid"
	// CodeAuditBlocked reports an external source refused by policy audit.
	CodeAuditBlocked = "build_repository_audit_blocked"
	// CodeReceiptInvalid reports a non-canonical or mismatched receipt v2.
	CodeReceiptInvalid = "build_repository_receipt_invalid"
	// CodeArtifactInvalid reports an invalid protected artifact.
	CodeArtifactInvalid = "build_repository_artifact_invalid"
	// CodeProtectedBoundaryUntrusted reports cache ownership or shape failure.
	CodeProtectedBoundaryUntrusted = "build_repository_protected_boundary_untrusted"
	// CodeUnverifiedOffline reports absent proof for offline source reuse.
	CodeUnverifiedOffline = "build_repository_unverified_offline"
	// CodeSignerPolicyUnsupported reports an unreviewed operator signer policy.
	CodeSignerPolicyUnsupported = "build_repository_signer_policy_unsupported"
	// CodePackageSigningForbidden reports signing credentials in package input.
	CodePackageSigningForbidden = "build_repository_package_signing_forbidden"
)

// Operation identifies the source-coverage and mutation promises of a run.
type Operation string

const (
	// OperationInstall performs a mutating install pipeline.
	OperationInstall Operation = "install"
	// OperationDryRun performs the complete read-only validation pipeline.
	OperationDryRun Operation = "dry-run"
	// OperationRepair safely reacquires and republishes invalid cache state.
	OperationRepair Operation = "repair"
	// OperationAudit validates source and policy without publication.
	OperationAudit Operation = "audit"
	// OperationSyntax validates only closed document syntax.
	OperationSyntax Operation = "syntax"
)

// DeclaredState is immutable package state. It is retained even when an
// operator substitution selects different effective bytes.
type DeclaredState struct {
	Repository, Identity, Transport, ObjectFormat, Commit, Tag string
}

// SubstitutionState records the selected non-committed developer override.
type SubstitutionState struct {
	Type, RefKind, RefValue string
}

// EffectiveState identifies the bytes actually admitted for compilation.
type EffectiveState struct {
	IdentityKind, Identity, Transport, ObjectFormat, Commit string
	Substituted                                             bool
	Substitution                                            *SubstitutionState
}

// AuditSubject binds declared and effective repository state for policy audit.
type AuditSubject struct {
	Declared                                    DeclaredState
	Effective                                   EffectiveState
	BuildSource, DescriptorTarget, SnapshotRoot string
	TagVerified                                 bool
}

// ToolchainIdentity is the exact compiler identity entering receipt v2.
type ToolchainIdentity struct {
	ContentSHA256, GoVersion, GoRelpath string
	GOOS, GOARCH                        string
	Tuning                              map[string]string
}

// GoSession is deliberately shared with the local go-v1 driver. The external
// pipeline adds receipt-v2 source state but does not alter the session or any
// receipt-v1 representation owned by go-v1.
type GoSession interface {
	Identity() ToolchainIdentity
	BuildInput(CompileRequest) (buildmeta.Input, error)
	Compile(context.Context, CompileRequest) (CompileResult, error)
}

// CompileRequest is the restricted target view passed to the compiler session.
type CompileRequest struct {
	Root, SourceDir, Command string
}

// CompileResult carries the artifact together with the exact session receipt
// produced by the authority-gated compiler dispatch.
type CompileResult struct {
	Artifact         []byte
	ExecutionReceipt closureexec.BuildSessionReceipt
}

// ArtifactHit is a protected receipt-v2 cache hit.
type ArtifactHit struct {
	Bytes, Receipt, ExecutionReceipt []byte
}

// ProtectedStore is the fail-closed snapshot and artifact cache boundary.
type ProtectedStore interface {
	LoadSnapshot(key string, mutate bool) (*Snapshot, error)
	StoreSnapshot(key string, snapshot *Snapshot) error
	LookupArtifact(key string, input map[string]any, mutate bool) (*ArtifactHit, error)
	StoreArtifact(key string, input map[string]any, command string, artifact, executionReceipt []byte) ([]byte, error)
}

// PipelineRequest contains manager-derived inputs to one external build.
type PipelineRequest struct {
	Operation       Operation
	Command, Target string
	Declared        DeclaredState
	Effective       EffectiveState
	Acquire         func(context.Context) (*Snapshot, error)
	Audit           func(context.Context, AuditSubject) error
	AuditWarnings   func(context.Context, AuditSubject) ([]string, error)
	Store           ProtectedStore
	Go              GoSession
	// OfflineSnapshotKey is an explicitly retained exact snapshot reference.
	// It is never inferred from declared state and cannot prove a declared tag.
	OfflineSnapshotKey string
	Trace              func(string)
	// SigningPolicy is manager/operator state, never package input. The rc.5
	// driver admits only the empty/"none" policy because no reviewed local
	// signing profile exists yet.
	SigningPolicy string
	// PackageSigningRequested is an explicit parser-to-driver refusal witness.
	// Closed schema-7 inputs cannot normally set it; keeping the gate here
	// prevents future parser expansion from silently crossing the boundary.
	PackageSigningRequested bool
	// Assurance is the explicit operation binding selected before any cache
	// lookup. Its zero value is never interpreted as portable.
	Assurance closureexec.AssuranceBinding
	// AssuranceCheck revalidates the same operation authority immediately at
	// cache lookup/adoption, dispatch, and publication boundaries.
	AssuranceCheck func(context.Context) error
}

// PipelineResult is the audited receipt-v2 result ready for transaction staging.
type PipelineResult struct {
	State, Code, SnapshotKey, CacheKey string
	BuildSource                        string
	Artifact, Receipt                  []byte
	ExecutionReceipt                   closureexec.BuildSessionReceipt
	Subject                            AuditSubject
	Warnings                           []string
}

func (r PipelineRequest) trace(phase string) {
	if r.Trace != nil {
		r.Trace(phase)
	}
}

// RunPipeline executes the pre-publication external repository state machine.
// Marker and consumer transaction publication intentionally remain outside it.
func RunPipeline(ctx context.Context, request PipelineRequest) (PipelineResult, error) {
	result := PipelineResult{}
	if request.PackageSigningRequested {
		return result, admissionError(CodePackageSigningForbidden, "package data cannot request artifact signing")
	}
	if request.SigningPolicy != "" && request.SigningPolicy != "none" {
		return result, admissionError(CodeSignerPolicyUnsupported, "no reviewed operator signer profile is configured")
	}
	mutate := request.Operation == OperationInstall || request.Operation == OperationRepair
	request.trace("exact-source-acquisition")
	if request.Acquire == nil {
		return result, admissionError(CodeSourceUnavailable, "exact source acquisition is not configured")
	}
	snapshot, acquireErr := request.Acquire(ctx)
	if acquireErr != nil {
		if request.Operation == OperationSyntax {
			return PipelineResult{State: "unverified-offline", Code: CodeUnverifiedOffline}, nil
		}
		// An untagged exact protected snapshot may support offline reinstall,
		// but a declared tag requires a same-operation exact-tag assertion.
		if request.Declared.Tag == "" && request.OfflineSnapshotKey != "" && request.Store != nil {
			snapshot, acquireErr = request.Store.LoadSnapshot(request.OfflineSnapshotKey, mutate)
		}
		if acquireErr != nil || snapshot == nil {
			return result, admissionError(CodeSourceUnavailable, "exact external source is unavailable")
		}
	}
	if snapshot == nil {
		return result, admissionError(CodeIncompleteSource, "source admission returned no snapshot")
	}
	request.trace("raw-object-identity-and-graph-proof")
	request.trace("all-blob-lfs-scan")
	request.trace("immutable-snapshot-materialization")
	root, cleanup, err := materializeValidated(snapshot)
	if err != nil {
		return result, err
	}
	defer cleanup()
	request.trace("whole-snapshot-validation")
	if err := validateMaterialized(root, snapshot); err != nil {
		return result, err
	}
	request.trace("build-source-digest")
	result.BuildSource = snapshot.Digest
	descriptor, err := LoadDescriptor(root)
	if err != nil {
		return result, admissionError(CodeDescriptorInvalid, "%v", err)
	}
	target, ok := descriptor.Targets[request.Target]
	if !ok {
		return result, admissionError(CodeDescriptorInvalid, "descriptor target %q is absent", request.Target)
	}
	if err := validateTargetModule(root, target); err != nil {
		return result, admissionError(CodeDescriptorInvalid, "%v", err)
	}
	// A local, tag, or branch substitution selects bytes before it has an
	// immutable commit field. Admission is the authority that derives that
	// exact effective commit; the declared lock remains unchanged alongside it.
	if request.Effective.Substituted && request.Effective.Substitution != nil &&
		(request.Effective.Substitution.Type == "local-path" || request.Effective.Substitution.RefKind != "revision") {
		request.Effective.Commit = snapshot.Commit
		request.Effective.ObjectFormat = snapshot.ObjectFormat
	}
	request.trace("descriptor-and-target-validation")
	if err := validateSourceBinding(request.Declared, request.Effective, snapshot); err != nil {
		return result, err
	}
	result.SnapshotKey, err = SnapshotKey(request.Effective, snapshot.Digest)
	if err != nil {
		return result, err
	}
	result.Subject = AuditSubject{Declared: request.Declared, Effective: request.Effective, BuildSource: snapshot.Digest, DescriptorTarget: request.Target, SnapshotRoot: root, TagVerified: request.Declared.Tag != "" && snapshot.TagVerified}
	request.trace("independent-external-audit")
	var auditErr error
	if request.AuditWarnings != nil {
		result.Warnings, auditErr = request.AuditWarnings(ctx, result.Subject)
	} else if request.Audit != nil {
		auditErr = request.Audit(ctx, result.Subject)
	} else {
		auditErr = fmt.Errorf("audit is not configured")
	}
	if auditErr != nil {
		return result, admissionError(CodeAuditBlocked, "independent external repository audit failed")
	}
	if err := validateMaterialized(root, snapshot); err != nil {
		return result, err
	}
	if request.Operation == OperationAudit || request.Operation == OperationSyntax {
		result.State = "audited"
		if request.Operation == OperationSyntax {
			result.State = "source-covered"
		}
		return result, nil
	}
	if request.Store == nil || request.Go == nil {
		return result, fmt.Errorf("external pipeline requires protected store and shared Go session")
	}
	if err := request.Assurance.Validate(); err != nil {
		return result, fmt.Errorf("build assurance is invalid: %w", err)
	}
	if request.AssuranceCheck == nil {
		return result, fmt.Errorf("build assurance recheck is absent")
	}
	if err := request.AssuranceCheck(ctx); err != nil {
		return result, err
	}
	if mutate {
		if err := request.Store.StoreSnapshot(result.SnapshotKey, snapshot); err != nil {
			return result, err
		}
	}
	input := receiptInput(request, target, snapshot.Digest, request.Go.Identity())
	result.CacheKey, err = cacheKey(input)
	if err != nil {
		return result, err
	}
	compilerRoot, sourceDir, removeCompilerRoot, err := compilerView(root, target)
	if err != nil {
		return result, err
	}
	defer removeCompilerRoot()
	compileRequest := CompileRequest{Root: compilerRoot, SourceDir: sourceDir, Command: request.Command}
	buildInput, err := request.Go.BuildInput(compileRequest)
	if err != nil {
		return result, fmt.Errorf("derive assured compiler input: %w", err)
	}
	request.trace("artifact-cache-lookup")
	if err := request.AssuranceCheck(ctx); err != nil {
		return result, err
	}
	hit, cacheErr := request.Store.LookupArtifact(result.CacheKey, input, mutate)
	if err := request.AssuranceCheck(ctx); err != nil {
		return result, err
	}
	if cacheErr == nil && hit != nil {
		execution, receiptErr := closureexec.DecodeBuildSessionReceipt(hit.ExecutionReceipt)
		if receiptErr == nil {
			receiptErr = execution.ValidateFor(request.Assurance, buildInput, artifactMetadata(input, hit.Bytes))
		}
		if receiptErr == nil {
			if err := validateMaterialized(root, snapshot); err != nil {
				return result, err
			}
			result.State, result.Artifact, result.Receipt = "cache-hit", hit.Bytes, hit.Receipt
			result.ExecutionReceipt = execution
			return result, nil
		}
		cacheErr = admissionError(CodeReceiptInvalid, "execution receipt mismatch: %v", receiptErr)
	}
	if request.Operation == OperationDryRun {
		result.State = "would-preflight-and-build"
		if cacheErr != nil {
			result.State, result.Code = "corrupt", ErrorCode(cacheErr)
		}
		return result, nil
	}
	request.trace("compiler")
	if err := request.AssuranceCheck(ctx); err != nil {
		return result, err
	}
	compiled, err := request.Go.Compile(ctx, compileRequest)
	if err != nil {
		return result, err
	}
	artifact := compiled.Artifact
	if len(artifact) == 0 {
		return result, admissionError(CodeArtifactInvalid, "compiler returned an empty artifact")
	}
	if err := compiled.ExecutionReceipt.ValidateFor(request.Assurance, buildInput, artifactMetadata(input, artifact)); err != nil {
		return result, admissionError(CodeReceiptInvalid, "compiler execution receipt mismatch: %v", err)
	}
	executionBytes, err := compiled.ExecutionReceipt.CanonicalBytes()
	if err != nil {
		return result, admissionError(CodeReceiptInvalid, "encode compiler execution receipt: %v", err)
	}
	if err := validateMaterialized(root, snapshot); err != nil {
		return result, err
	}
	if err := request.AssuranceCheck(ctx); err != nil {
		return result, err
	}
	receipt, err := request.Store.StoreArtifact(result.CacheKey, input, request.Command, artifact, executionBytes)
	if err != nil {
		return result, err
	}
	result.State, result.Artifact, result.Receipt = "would-preflight-and-build", artifact, receipt
	result.ExecutionReceipt = compiled.ExecutionReceipt
	if cacheErr != nil {
		result.State, result.Code = "would-rebuild-untrusted-cache", ErrorCode(cacheErr)
	}
	return result, nil
}

func validateSourceBinding(declared DeclaredState, effective EffectiveState, snapshot *Snapshot) error {
	if declared.Repository == "" || declared.Identity == "" || declared.ObjectFormat == "" || declared.Commit == "" {
		return admissionError(CodeIdentityInvalid, "declared source is incomplete")
	}
	if effective.IdentityKind != "network-git" && effective.IdentityKind != "operator-local-git" {
		return admissionError(CodeIdentityInvalid, "effective identity kind is invalid")
	}
	if effective.ObjectFormat != snapshot.ObjectFormat || effective.Commit != snapshot.Commit {
		return admissionError(CodeIdentityInvalid, "effective source does not bind admitted snapshot")
	}
	if !effective.Substituted {
		if effective.Substitution != nil || effective.IdentityKind != "network-git" || effective.Identity != declared.Identity || effective.Transport != declared.Transport || effective.ObjectFormat != declared.ObjectFormat || effective.Commit != declared.Commit {
			return admissionError(CodeIdentityInvalid, "unsubstituted declared/effective source mismatch")
		}
		if declared.Tag != "" && !snapshot.TagVerified {
			return admissionError(CodeRefMoved, "declared tag was not proved in this operation")
		}
	} else if effective.Substitution == nil || (effective.Substitution.Type != "local-path" && effective.Substitution.Type != "network-git") {
		return admissionError(CodeIdentityInvalid, "substituted source lacks typed substitution state")
	}
	return nil
}

func materializeValidated(snapshot *Snapshot) (string, func(), error) {
	root, err := os.MkdirTemp("", "curator-external-snapshot-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(root) }
	if err := snapshot.Materialize(root); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return root, cleanup, nil
}

func validateMaterialized(root string, expected *Snapshot) error {
	files, err := readRegularTree(root)
	if err != nil {
		return admissionError(CodeObjectSemanticsInvalid, "snapshot validation: %v", err)
	}
	canonical := frameSnapshot(files)
	digest := sha256.Sum256(canonical)
	if expected.Digest != "sha256:"+hex.EncodeToString(digest[:]) || !bytes.Equal(expected.CanonicalBytes, canonical) || !reflect.DeepEqual(files, expected.Files) {
		return admissionError(CodeObjectSemanticsInvalid, "materialized snapshot differs from admitted bytes")
	}
	return nil
}

func readRegularTree(root string) ([]File, error) {
	var files []File
	err := filepath.WalkDir(root, func(name string, _ os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == root {
			return nil
		}
		info, err := os.Lstat(name)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || (!info.IsDir() && !info.Mode().IsRegular()) || !regularFileIdentity(info, info.IsDir()) {
			return fmt.Errorf("non-regular snapshot entry")
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		payload, err := os.ReadFile(name) // #nosec G304 -- WalkDir proved this entry is a regular file below the admitted snapshot root.
		if err != nil {
			return err
		}
		files = append(files, File{Path: rel, Content: payload, Executable: info.Mode()&0o100 != 0})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func validateTargetModule(root string, target Target) error {
	buildRoot := filepath.Join(root, filepath.FromSlash(target.BuildRoot))
	info, err := os.Lstat(buildRoot)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("build_root is not a real directory")
	}
	mod := filepath.Join(buildRoot, "go.mod")
	info, err = os.Lstat(mod)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("build_root must contain a real go.mod")
	}
	source := filepath.Join(root, filepath.FromSlash(target.SourceDir))
	info, err = os.Lstat(source)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source_dir is not a real directory")
	}
	return nil
}

func compilerView(snapshotRoot string, target Target) (string, string, func(), error) {
	view, err := os.MkdirTemp("", "curator-external-build-root-")
	if err != nil {
		return "", "", func() {}, err
	}
	remove := func() { _ = os.RemoveAll(view) }
	sourceRoot := filepath.Join(snapshotRoot, filepath.FromSlash(target.BuildRoot))
	viewBuildRoot := filepath.Join(view, "build")
	if err := os.Mkdir(viewBuildRoot, 0o700); err != nil {
		remove()
		return "", "", func() {}, err
	}
	err = filepath.WalkDir(sourceRoot, func(name string, _ os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(sourceRoot, name)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		destination := filepath.Join(viewBuildRoot, rel)
		info, err := os.Lstat(name)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.Mkdir(destination, 0o700)
		}
		payload, err := os.ReadFile(name) // #nosec G304 -- WalkDir proved this entry is regular and below the restricted compiler view.
		if err != nil {
			return err
		}
		mode := os.FileMode(0o600)
		if info.Mode()&0o100 != 0 {
			mode = 0o700
		}
		return os.WriteFile(destination, payload, mode)
	})
	if err != nil {
		remove()
		return "", "", func() {}, err
	}
	sourceDir := "build"
	if target.SourceDir != target.BuildRoot {
		sourceDir += "/" + strings.TrimPrefix(target.SourceDir, target.BuildRoot+"/")
	}
	return view, sourceDir, remove, nil
}

// SnapshotKey derives the canonical protected-snapshot key.
func SnapshotKey(e EffectiveState, digest string) (string, error) {
	payload, err := registry.CanonicalBytesChecked(map[string]any{"identity": map[string]any{"kind": e.IdentityKind, "value": e.Identity}, "object_format": e.ObjectFormat, "commit": e.Commit, "build_source": digest})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func cacheKey(input map[string]any) (string, error) {
	payload, err := registry.CanonicalBytesChecked(input)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func artifactMetadata(input map[string]any, artifact []byte) buildmeta.Artifact {
	sum := sha256.Sum256(artifact)
	path, _ := artifactPathFromInput(input)
	return buildmeta.Artifact{Path: path, SHA256: "sha256:" + hex.EncodeToString(sum[:]), Size: int64(len(artifact))}
}

func receiptInput(r PipelineRequest, target Target, digest string, tool ToolchainIdentity) map[string]any {
	declared := map[string]any{"identity": map[string]any{"kind": "network-git", "value": r.Declared.Identity}, "transport": r.Declared.Transport, "locked_commit": map[string]any{"object_format": r.Declared.ObjectFormat, "hex": r.Declared.Commit}}
	if r.Declared.Tag != "" {
		declared["tag"] = r.Declared.Tag
	}
	effective := map[string]any{"identity": map[string]any{"kind": r.Effective.IdentityKind, "value": r.Effective.Identity}, "object_format": r.Effective.ObjectFormat, "commit": r.Effective.Commit, "substituted": r.Effective.Substituted, "build_source": map[string]any{"algorithm": "curator-build-source-v1", "content_sha256": digest}}
	if r.Effective.Transport != "" {
		effective["transport"] = r.Effective.Transport
	}
	if r.Effective.Substitution != nil {
		sub := map[string]any{"type": r.Effective.Substitution.Type}
		if r.Effective.Substitution.RefKind != "" {
			sub["ref"] = map[string]any{"kind": r.Effective.Substitution.RefKind, "value": r.Effective.Substitution.RefValue}
		}
		effective["substitution"] = sub
	}
	tuning := map[string]any{}
	for k, v := range tool.Tuning {
		tuning[k] = v
	}
	input := map[string]any{"schema_version": 2, "driver": "go-repository-v1", "command": r.Command, "build_root": target.BuildRoot, "source_dir": target.SourceDir, "source": map[string]any{"repository": r.Declared.Repository, "declared": declared, "effective": effective, "descriptor": map[string]any{"path": DescriptorName, "target": r.Target}}, "target": map[string]any{"goos": tool.GOOS, "goarch": tool.GOARCH, "tuning": tuning}, "toolchain": map[string]any{"algorithm": "curator-go-toolchain-v1", "content_sha256": tool.ContentSHA256, "go_version": tool.GoVersion, "go_relpath": tool.GoRelpath}, "policy": map[string]any{"module_mode": "vendor", "network": "none", "workspace": false, "cgo": false, "compiler_directives": "reject-nonstandard-cgo-import-dynamic-v1", "target_mode": "native", "link_mode": "internal", "libgcc": "none", "package_assembly": false, "host_objects": false, "telemetry": "off-private", "execution_policy": "manager-worker-v1", "source_kind": "locked-external-git-v1"}}
	input["assurance"] = r.Assurance.CanonicalValue()
	input["policy"].(map[string]any)["execution_policy"] = r.Assurance.ExecutionPolicyID
	return input
}
