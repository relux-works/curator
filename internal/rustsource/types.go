// Package rustsource implements the fail-closed rust-source-v1 Cargo capture
// and cargo-vendor-transform-v1 boundary. It deliberately stops before Rust
// compilation; the offline build adapter consumes the closure emitted here.
package rustsource

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// ProfileID identifies the closed Rust adapter profile.
	ProfileID = "rust-source-v1"
	// TransformID pins the exact accepted Cargo vendoring implementation.
	TransformID = "cargo-vendor-transform-v1:cargo-0.92.0@ea2d97820c16195b0ca3fadb4319fe512c199a43"
	// NormalizerID pins Cargo's Git manifest normalizer.
	NormalizerID = "cargo-git-manifest-normalizer-v1:cargo-0.92.0@ea2d97820c16195b0ca3fadb4319fe512c199a43"
)

// Code is a stable Rust adapter diagnostic code.
type Code string

const (
	// CodeLockRequired reports a missing or unsupported lockfile.
	CodeLockRequired Code = "rust_lock_required"
	// CodeLockMismatch reports inconsistent lock declarations.
	CodeLockMismatch Code = "rust_lock_mismatch"
	// CodeRegistryIdentityInvalid reports registry origin or transform drift.
	CodeRegistryIdentityInvalid Code = "rust_registry_identity_invalid"
	// CodeGitIdentityInvalid reports Git object, projection, or transform drift.
	CodeGitIdentityInvalid Code = "rust_git_identity_invalid"
	// CodePathDependencyEscape reports an uncontained local dependency.
	CodePathDependencyEscape Code = "rust_path_dependency_escape"
	// CodeVendorTransformUnsupported reports an unpinned Cargo implementation.
	CodeVendorTransformUnsupported Code = "rust_vendor_transform_unsupported"
	// CodeVendorIncomplete reports non-bijective or incomplete vendor output.
	CodeVendorIncomplete Code = "rust_vendor_incomplete"
	// CodeGraphIncomplete reports malformed or contradictory graph evidence.
	CodeGraphIncomplete Code = "rust_graph_incomplete"
	// CodeFeatureProfileMismatch reports selected/resolved feature drift.
	CodeFeatureProfileMismatch Code = "rust_feature_profile_mismatch"
	// CodeTargetUnsupported reports a non-native or mismatched target.
	CodeTargetUnsupported Code = "rust_target_unsupported"
	// CodeBuildScriptUnsupported reports active build-time code.
	CodeBuildScriptUnsupported Code = "rust_build_script_unsupported"
	// CodeProcMacroUnsupported reports an active procedural macro.
	CodeProcMacroUnsupported Code = "rust_proc_macro_unsupported"
	// CodeNativeLinkUnsupported reports package-selected native linking.
	CodeNativeLinkUnsupported Code = "rust_native_link_unsupported"
	// CodeConfigUntrusted reports package or ambient Cargo configuration.
	CodeConfigUntrusted Code = "rust_config_untrusted"
	// CodeUndeclaredInput reports a build-time process, read, environment, or
	// write outside the closed rust-source-v1 plan.
	CodeUndeclaredInput Code = "rust_undeclared_input"
	// CodeOfflineRebuildFailed reports a frozen metadata or build failure after
	// the admitted closure and exact toolchain have been rechecked.
	CodeOfflineRebuildFailed Code = "rust_offline_rebuild_failed"
	// CodeGraphReferenceInvalid is the shared diagnostic for malformed C4
	// binding references consumed by this adapter.
	CodeGraphReferenceInvalid Code = "closure_graph_reference_invalid"
	// CodeToolchainIdentityChanged is the shared immediate time-of-use drift
	// diagnostic.
	CodeToolchainIdentityChanged Code = "artifact_toolchain_identity_changed"
	// CodeLocalOutputUnreceipted reports bytes planted in a build output root
	// before a permitted Cargo action starts.
	CodeLocalOutputUnreceipted Code = "artifact_local_output_unreceipted"
)

// Error retains a stable code and machine-readable fields.
type Error struct {
	Code   Code
	Detail string
	Fields map[string]string
}

func (e *Error) Error() string {
	if e.Detail == "" {
		return string(e.Code)
	}
	return string(e.Code) + ": " + e.Detail
}

func fail(code Code, detail string, fields map[string]string) error {
	return &Error{Code: code, Detail: detail, Fields: fields}
}

// ErrorCode extracts a Rust-specific diagnostic code.
func ErrorCode(err error) Code {
	if typed, ok := err.(*Error); ok {
		return typed.Code
	}
	return ""
}

// SourceKind is the closed Cargo source set supported by rust-source-v1.
type SourceKind string

const (
	// SourcePath identifies a contained root/workspace/path package.
	SourcePath SourceKind = "path"
	// SourceRegistry identifies a captured index-record/raw-archive origin.
	SourceRegistry SourceKind = "registry"
	// SourceGit identifies an exact commit/tree/submodule origin.
	SourceGit SourceKind = "git"
)

// PackageKey is the selection-neutral lock identity.
type PackageKey struct {
	Name, Version, Source string
}

func (key PackageKey) String() string { return key.Name + " " + key.Version + " " + key.Source }

// LockDependency is an unevaluated dependency reference retained from Cargo.lock.
type LockDependency struct{ Name, Version, Source string }

// LockPackage is one authoritative lock-superset member.
type LockPackage struct {
	Key          PackageKey
	Kind         SourceKind
	Checksum     string
	Dependencies []LockDependency
}

// LockFile is a supported Cargo lockfile and its complete package superset.
type LockFile struct {
	Version  int
	Packages []LockPackage
	Digest   string
}

// DependencyDeclaration preserves conditions without evaluating them.
type DependencyDeclaration struct {
	Name, Package, Version, Registry, Git, Rev, Branch, Tag, Path string
	Optional, DefaultFeatures                                     bool
	Features                                                      []string
	Kind                                                          string
	Target                                                        string
}

// Manifest is the closed security-relevant declaration projection.
type Manifest struct {
	Path, PackageName, PackageVersion string
	WorkspaceMembers                  []string
	Dependencies                      []DependencyDeclaration
	Features                          map[string][]string
	Bins                              []string
	HasBuildScript, ProcMacro         bool
	Links                             string
	Digest                            string
}

// OriginLeaf is one immutable regular source leaf.
type OriginLeaf struct {
	Path, SHA256 string
	Size         int64
	Bytes        []byte
}

// registryOrigin binds the raw archive to its registry index and lock record.
type registryOrigin struct {
	Package     PackageKey
	IndexRecord []byte
	Archive     []byte
	Checksum    string
}

// ProjectionMode is the closed Cargo 0.92 Git PathSource branch.
type ProjectionMode string

const (
	// ProjectionGitIndexNoInclude selects Cargo's tracked-index branch.
	ProjectionGitIndexNoInclude ProjectionMode = "git_index_no_include"
	// ProjectionFilesystemInclude selects Cargo's explicit-include branch.
	ProjectionFilesystemInclude ProjectionMode = "filesystem_include"
)

// SubmoduleEvidence binds one recursively captured Git submodule.
type SubmoduleEvidence struct {
	Path       string `json:"path"`
	Gitlink    string `json:"gitlink"`
	Commit     string `json:"commit"`
	TreeDigest string `json:"tree_digest"`
}

// gitOrigin binds exact object identity and complete clean tree bytes. Selected
// paths are independently derived by the capture manager; every other leaf is
// retained and receives omit_unselected.
type gitOrigin struct {
	Package                          PackageKey
	DeclaredURL, Selector            string
	Commit, Tree                     string
	Root, AdminRoot, PackagePath     string
	Include                          []string
	ManifestTracked                  bool
	Leaves                           []OriginLeaf
	Submodules                       []SubmoduleEvidence
	Dirty, UsesFilter, IndexConflict bool
}

// gitDerivation is manager-sealed cargo-0.92 projection and normalization
// evidence. It is deliberately package-private and is never accepted in the
// public capture request, so an adapter caller cannot supply projection or
// normalized-manifest authority alongside an origin.
type gitDerivation struct {
	mode                       ProjectionMode
	selected, normalizerInputs []string
	normalizerID               string
	normalizedManifest         []byte
	receiptID                  string
	commit, tree, packagePath  string
	include                    []string
	submodules                 []SubmoduleEvidence
	manifestTracked            bool
	seal                       string
}

// pathOrigin is a root/workspace/path dependency captured inside one boundary.
type pathOrigin struct {
	Package PackageKey
	Root    string
}

// Disposition is the closed per-leaf vendor transform outcome.
type Disposition string

const (
	// CopyIdentical maps one origin leaf byte-for-byte.
	CopyIdentical Disposition = "copy_identical"
	// OmitReserved removes one exact package-root reserved name.
	OmitReserved Disposition = "omit_reserved"
	// OmitRegistryCargoOK removes a registry .cargo-ok at any depth.
	OmitRegistryCargoOK Disposition = "omit_registry_cargo_ok"
	// OmitUnselected retains evidence for a Git leaf outside its projection.
	OmitUnselected Disposition = "omit_unselected"
	// ReplaceNormalizedManifest maps Cargo.toml to pinned normalized bytes.
	ReplaceNormalizedManifest Disposition = "replace_normalized_manifest"
	// GenerateChecksum adds Cargo's exact compact checksum metadata.
	GenerateChecksum Disposition = "generate_checksum"
)

// TransformEntry accounts for exactly one origin or generated leaf.
type TransformEntry struct {
	OriginPath, VendorPath string
	Disposition            Disposition
	OriginSHA256           string
	ExpectedSHA256         string
	Size                   int64
	Rule                   string
}

// VendorPackage is the complete expected result for one lock origin.
type VendorPackage struct {
	Package       PackageKey
	Directory     string
	Kind          SourceKind
	Entries       []TransformEntry
	Files         []OriginLeaf
	ChecksumBytes []byte
}

func packageDirectory(key PackageKey) string { return key.Name + "-" + key.Version }

func validatePackageKey(key PackageKey) error {
	if key.Name == "" || key.Version == "" || strings.ContainsAny(key.Name+key.Version+key.Source, "\x00\r\n") {
		return fmt.Errorf("invalid Cargo package identity")
	}
	return nil
}

func sortedUnique(values []string) ([]string, bool) {
	copyOf := append([]string(nil), values...)
	sort.Strings(copyOf)
	for i := 1; i < len(copyOf); i++ {
		if copyOf[i] == copyOf[i-1] {
			return nil, false
		}
	}
	return copyOf, true
}
