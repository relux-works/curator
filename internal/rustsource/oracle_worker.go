package rustsource

import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/privatedir"
	"github.com/relux-works/curator/internal/protocoljson"
)

// rustGitOracleWorkerMode is the sole hidden process mode for the pinned Go
// implementation of Cargo 0.92 Git PathSource projection and normalization.
const rustGitOracleWorkerMode = "__curator_rust_git_oracle_v1"

type gitOracleContext struct {
	SchemaID        string              `json:"schema_id"`
	Package         PackageKey          `json:"package"`
	DeclaredURL     string              `json:"declared_url"`
	Selector        string              `json:"selector"`
	Commit          string              `json:"commit"`
	Tree            string              `json:"tree"`
	PackagePath     string              `json:"package_path"`
	Include         []string            `json:"include"`
	ManifestTracked bool                `json:"manifest_tracked"`
	TrackedPaths    []string            `json:"tracked_paths"`
	Submodules      []SubmoduleEvidence `json:"submodules"`
}

// DispatchInternalWorker handles the sole fixed Rust worker mode. The trusted
// CLI composition root calls it before ordinary command parsing. Importing
// rustsource has no execution side effects.
func DispatchInternalWorker(args []string, _ io.Reader, _ io.Writer) (bool, int) {
	if len(args) != 1 || args[0] != rustGitOracleWorkerMode {
		return false, 0
	}
	return true, runRustGitOracleWorker()
}

func runRustGitOracleWorker() int {
	contextPath := os.Getenv("RUST_GIT_CONTEXT")
	sourceRoot := os.Getenv("RUST_GIT_SOURCE")
	outputRoot := os.Getenv("CURATOR_OUTPUT_ROOT")
	if contextPath == "" || sourceRoot == "" || outputRoot == "" {
		return 2
	}
	contextBytes, err := os.ReadFile(contextPath) // #nosec G304 -- exact permit-bound replay path.
	if err != nil {
		return 1
	}
	var contextRecord gitOracleContext
	if err = protocoljson.UnmarshalCanonical(contextBytes, &contextRecord); err != nil || contextRecord.SchemaID != "rust-git-oracle-context-v1" || !validLowerHex(contextRecord.Commit, 40) || contextRecord.Tree == "" {
		return 1
	}
	tracked, unique := sortedUnique(contextRecord.TrackedPaths)
	include, includeUnique := sortedUnique(contextRecord.Include)
	if !unique || !includeUnique {
		return 1
	}
	trackedSet := map[string]bool{}
	for _, item := range tracked {
		trackedSet[item] = true
	}
	leaves, err := managerInventory(sourceRoot)
	if err != nil {
		return 1
	}
	selected := []string{}
	for _, leaf := range leaves {
		packageRelative, inside := trimPackagePath(contextRecord.PackagePath, leaf.Path)
		if !inside || packageRelative == "" {
			continue
		}
		trackedBySubmodule := false
		for _, submodule := range contextRecord.Submodules {
			if leaf.Path != submodule.Path && strings.HasPrefix(leaf.Path, submodule.Path+"/") {
				trackedBySubmodule = true
				break
			}
		}
		if len(include) == 0 && (trackedSet[leaf.Path] || trackedBySubmodule) {
			selected = append(selected, packageRelative)
		}
		if len(include) != 0 && !strings.HasPrefix(packageRelative, "target/") && includePath(include, packageRelative) {
			selected = append(selected, packageRelative)
		}
	}
	selected, unique = sortedUnique(selected)
	if !unique || !stringSliceContains(selected, "Cargo.toml") {
		return 1
	}
	packageRoot := sourceRoot
	if contextRecord.PackagePath != "" {
		packageRoot = filepath.Join(sourceRoot, filepath.FromSlash(contextRecord.PackagePath))
	}
	normalized, err := normalizeGitManifestV1(packageRoot, selected)
	if err != nil {
		return 1
	}
	normalizerInputs := []string{}
	for _, item := range selected {
		if item != ".gitignore" && item != ".gitattributes" && item != ".cargo-ok" {
			normalizerInputs = append(normalizerInputs, item)
		}
	}
	mode := ProjectionGitIndexNoInclude
	if len(include) != 0 {
		mode = ProjectionFilesystemInclude
	}
	encoded, err := protocoljson.MarshalCanonical(map[string]any{"commit": contextRecord.Commit, "include": stringsAny(include), "manifest_tracked": contextRecord.ManifestTracked, "mode": string(mode), "normalized_manifest_base64": base64.StdEncoding.EncodeToString(normalized), "normalizer_id": NormalizerID, "normalizer_inputs": stringsAny(normalizerInputs), "package_path": contextRecord.PackagePath, "schema_id": gitDerivationSchemaID, "selected": stringsAny(selected), "submodules": submoduleValues(contextRecord.Submodules), "tree": contextRecord.Tree})
	if err != nil {
		return 1
	}
	if err = privatedir.MakeAll(outputRoot); err != nil {
		return 1
	}
	if err = os.WriteFile(filepath.Join(outputRoot, "rust-git-projection-v1.json"), encoded, 0o600); err != nil {
		return 1
	}
	return 0
}

var _ = sort.Strings
