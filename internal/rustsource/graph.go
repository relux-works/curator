package rustsource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/protocoljson"
)

// CaptureGraph is selection-neutral: it intentionally contains no requested
// features, target facts, platform nodes, toolchains, or targets edges.
type CaptureGraph struct {
	SchemaID, LockDigest, Identity string
	Packages                       []LockPackage
	Declarations                   []DependencyDeclaration
	CapturedManifestPaths          []string
	ManifestDigests                []string
	ArtifactManifestIDs            []string
}

// NewCaptureGraph creates the reusable lock-superset identity.
func NewCaptureGraph(lock LockFile, manifests []Manifest, artifactManifestIDs []string) (CaptureGraph, error) {
	declarations := []DependencyDeclaration{}
	paths := []string{}
	digests := []string{}
	for _, manifest := range manifests {
		declarations = append(declarations, manifest.Dependencies...)
		paths = append(paths, manifest.Path)
		digests = append(digests, manifest.Digest)
	}
	sort.Slice(declarations, func(i, j int) bool {
		a, b := declarations[i], declarations[j]
		return a.Name+"\x00"+a.Kind+"\x00"+a.Target+"\x00"+a.Package < b.Name+"\x00"+b.Kind+"\x00"+b.Target+"\x00"+b.Package
	})
	ids, unique := sortedUnique(artifactManifestIDs)
	if !unique {
		return CaptureGraph{}, fail(CodeGraphIncomplete, "duplicate artifact manifest identity", nil)
	}
	paths, unique = sortedUnique(paths)
	if !unique {
		return CaptureGraph{}, fail(CodeGraphIncomplete, "duplicate captured manifest path", nil)
	}
	digests, unique = sortedUnique(digests)
	if !unique || len(digests) != len(manifests) || (len(digests) > 0 && digests[0] == "") {
		return CaptureGraph{}, fail(CodeGraphIncomplete, "manifest digest is missing or duplicated", nil)
	}
	graph := CaptureGraph{SchemaID: "rust-capture-graph-v1", LockDigest: lock.Digest, Packages: append([]LockPackage(nil), lock.Packages...), Declarations: declarations, CapturedManifestPaths: paths, ManifestDigests: digests, ArtifactManifestIDs: ids}
	value := map[string]any{"artifact_manifest_ids": stringsAny(ids), "captured_manifest_paths": stringsAny(paths), "manifest_digests": stringsAny(digests), "declarations": declarationsValue(declarations), "lock_digest": lock.Digest, "packages": packagesValue(lock.Packages), "schema_id": graph.SchemaID}
	encoded, err := protocoljson.MarshalCanonical(value)
	if err != nil {
		return CaptureGraph{}, err
	}
	graph.Identity = "sha256:" + digest(append([]byte("rust-capture-graph-v1\x00"), encoded...))
	return graph, nil
}

// SelectionContext contains exact requested values, separate from capture.
type SelectionContext struct {
	Package, Binary, Target string
	DefaultFeatures         bool
	Features, TargetCFG     []string
	ResolvedFeatures        map[string][]string
}

// Metadata is the normalized security projection of Cargo metadata format 1.
type Metadata struct {
	Packages []MetadataPackage
	Resolve  []MetadataNode
}

// MetadataPackage is one normalized Cargo metadata package.
type MetadataPackage struct {
	ID, Name, Version, Source, ManifestPath, Links string
	Targets                                        []MetadataTarget
}

// MetadataTarget is one closed Cargo target declaration.
type MetadataTarget struct {
	Name, SrcPath     string
	Kinds, CrateTypes []string
}

// MetadataNode is one active Cargo resolution node.
type MetadataNode struct {
	ID           string
	Features     []string
	Dependencies []MetadataDependency
}

// MetadataDependency is one typed active dependency edge.
type MetadataDependency struct{ ID, Name, Kind, Target string }

// ActiveGraph binds selection and C0 platform/tool identities without changing capture.
type ActiveGraph struct {
	SchemaID, CaptureID, Identity, Target, ToolchainID, Package, Binary string
	DefaultFeatures                                                     bool
	RequestedFeatures                                                   []string
	Packages                                                            []MetadataPackage
	Nodes                                                               []MetadataNode
}

// ParseMetadata accepts the closed subset of Cargo metadata format 1 used by
// rust-source-v1 and rejects unknown security-relevant target/dependency kinds.
func ParseMetadata(payload []byte) (Metadata, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil {
		return Metadata{}, fail(CodeGraphIncomplete, "Cargo metadata is malformed", nil)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Metadata{}, fail(CodeGraphIncomplete, "Cargo metadata has trailing values", nil)
	}
	for key := range root {
		if key != "packages" && key != "resolve" && key != "workspace_root" && key != "target_directory" && key != "build_directory" && key != "workspace_members" && key != "workspace_default_members" && key != "version" && key != "metadata" && key != "workspace_metadata" {
			return Metadata{}, fail(CodeGraphIncomplete, "unknown metadata root field", map[string]string{"field": key})
		}
	}
	version, ok := root["version"].(json.Number)
	if !ok || string(version) != "1" {
		return Metadata{}, fail(CodeGraphIncomplete, "Cargo metadata format version is unsupported", nil)
	}
	packageValues, ok := root["packages"].([]any)
	if !ok {
		return Metadata{}, fail(CodeGraphIncomplete, "metadata packages missing", nil)
	}
	result := Metadata{}
	for _, value := range packageValues {
		object, ok := value.(map[string]any)
		if !ok {
			return Metadata{}, fail(CodeGraphIncomplete, "metadata package malformed", nil)
		}
		if err := allowFields(object, "metadata package", []string{"name", "version", "id", "license", "license_file", "description", "source", "dependencies", "targets", "features", "manifest_path", "metadata", "publish", "authors", "categories", "keywords", "readme", "repository", "homepage", "documentation", "edition", "links", "default_run", "rust_version"}); err != nil {
			return Metadata{}, err
		}
		if err := validateMetadataPackageDeclarations(object); err != nil {
			return Metadata{}, err
		}
		pkg := MetadataPackage{}
		var err error
		if pkg.ID, err = requiredString(object, "id"); err != nil {
			return Metadata{}, err
		}
		if pkg.Name, err = requiredString(object, "name"); err != nil {
			return Metadata{}, err
		}
		if pkg.Version, err = requiredString(object, "version"); err != nil {
			return Metadata{}, err
		}
		if pkg.Source, err = nullableString(object, "source"); err != nil {
			return Metadata{}, err
		}
		if pkg.ManifestPath, err = requiredString(object, "manifest_path"); err != nil {
			return Metadata{}, err
		}
		if pkg.Links, err = nullableString(object, "links"); err != nil {
			return Metadata{}, err
		}
		targetValues, ok := object["targets"].([]any)
		if !ok {
			return Metadata{}, fail(CodeGraphIncomplete, "metadata targets missing", map[string]string{"package": pkg.ID})
		}
		for _, targetValue := range targetValues {
			targetObject, ok := targetValue.(map[string]any)
			if !ok {
				return Metadata{}, fail(CodeGraphIncomplete, "metadata target malformed", nil)
			}
			if err := allowFields(targetObject, "metadata target", []string{"kind", "crate_types", "name", "src_path", "edition", "doc", "doctest", "test", "required-features"}); err != nil {
				return Metadata{}, err
			}
			target := MetadataTarget{}
			if target.Name, err = requiredString(targetObject, "name"); err != nil {
				return Metadata{}, err
			}
			if target.SrcPath, err = requiredString(targetObject, "src_path"); err != nil {
				return Metadata{}, err
			}
			target.Kinds, err = stringSlice(targetObject["kind"])
			if err != nil {
				return Metadata{}, err
			}
			target.CrateTypes, err = stringSlice(targetObject["crate_types"])
			if err != nil {
				return Metadata{}, err
			}
			for _, kind := range target.Kinds {
				if kind != "bin" && kind != "lib" && kind != "custom-build" && kind != "proc-macro" {
					return Metadata{}, fail(CodeGraphIncomplete, "unknown Cargo target kind", map[string]string{"kind": kind})
				}
			}
			pkg.Targets = append(pkg.Targets, target)
		}
		result.Packages = append(result.Packages, pkg)
	}
	resolveObject, ok := root["resolve"].(map[string]any)
	if !ok {
		return Metadata{}, fail(CodeGraphIncomplete, "metadata resolve missing", nil)
	}
	if err := allowFields(resolveObject, "metadata resolve", []string{"nodes", "root"}); err != nil {
		return Metadata{}, err
	}
	if _, err := nullableString(resolveObject, "root"); err != nil {
		return Metadata{}, err
	}
	nodeValues, ok := resolveObject["nodes"].([]any)
	if !ok {
		return Metadata{}, fail(CodeGraphIncomplete, "metadata resolve nodes missing", nil)
	}
	for _, value := range nodeValues {
		object, ok := value.(map[string]any)
		if !ok {
			return Metadata{}, fail(CodeGraphIncomplete, "metadata resolve node malformed", nil)
		}
		if err := allowFields(object, "metadata resolve node", []string{"id", "dependencies", "deps", "features"}); err != nil {
			return Metadata{}, err
		}
		if _, err := stringSlice(object["dependencies"]); err != nil {
			return Metadata{}, err
		}
		node := MetadataNode{}
		var err error
		if node.ID, err = requiredString(object, "id"); err != nil {
			return Metadata{}, err
		}
		if node.Features, err = stringSlice(object["features"]); err != nil {
			return Metadata{}, err
		}
		sort.Strings(node.Features)
		deps, ok := object["deps"].([]any)
		if !ok {
			return Metadata{}, fail(CodeGraphIncomplete, "metadata node deps missing", map[string]string{"package": node.ID})
		}
		for _, depValue := range deps {
			depObject, ok := depValue.(map[string]any)
			if !ok {
				return Metadata{}, fail(CodeGraphIncomplete, "metadata dependency malformed", nil)
			}
			if err := allowFields(depObject, "metadata dependency", []string{"name", "pkg", "dep_kinds"}); err != nil {
				return Metadata{}, err
			}
			dep := MetadataDependency{}
			if dep.ID, err = requiredString(depObject, "pkg"); err != nil {
				return Metadata{}, err
			}
			if dep.Name, err = requiredString(depObject, "name"); err != nil {
				return Metadata{}, err
			}
			kinds, ok := depObject["dep_kinds"].([]any)
			if !ok {
				return Metadata{}, fail(CodeGraphIncomplete, "dependency kinds missing", nil)
			}
			for _, kindValue := range kinds {
				kindObject, ok := kindValue.(map[string]any)
				if !ok {
					return Metadata{}, fail(CodeGraphIncomplete, "metadata dependency kind malformed", nil)
				}
				if err := allowFields(kindObject, "metadata dependency kind", []string{"kind", "target"}); err != nil {
					return Metadata{}, err
				}
				kind, err := nullableString(kindObject, "kind")
				if err != nil {
					return Metadata{}, err
				}
				if kind == "" {
					kind = "normal"
				}
				if kind != "normal" && kind != "dev" && kind != "build" {
					return Metadata{}, fail(CodeGraphIncomplete, "unknown dependency kind", map[string]string{"kind": kind})
				}
				copyOf := dep
				copyOf.Kind = kind
				copyOf.Target, err = nullableString(kindObject, "target")
				if err != nil {
					return Metadata{}, err
				}
				node.Dependencies = append(node.Dependencies, copyOf)
			}
		}
		sort.Slice(node.Dependencies, func(i, j int) bool { return fmt.Sprint(node.Dependencies[i]) < fmt.Sprint(node.Dependencies[j]) })
		result.Resolve = append(result.Resolve, node)
	}
	sort.Slice(result.Packages, func(i, j int) bool { return result.Packages[i].ID < result.Packages[j].ID })
	sort.Slice(result.Resolve, func(i, j int) bool { return result.Resolve[i].ID < result.Resolve[j].ID })
	return result, nil
}

// Reconcile verifies lock-superset coverage and produces an exact active identity.
func Reconcile(capture CaptureGraph, selection SelectionContext, metadata Metadata, hostTarget, toolchainID string) (ActiveGraph, error) {
	if selection.Target == "" || selection.Target != hostTarget {
		return ActiveGraph{}, fail(CodeTargetUnsupported, "rust-source-v1 requires one native target", map[string]string{"host": hostTarget, "target": selection.Target})
	}
	requested, unique := sortedUnique(selection.Features)
	if !unique {
		return ActiveGraph{}, fail(CodeFeatureProfileMismatch, "requested features contain duplicates", nil)
	}
	lockKeys := map[string]bool{}
	for _, item := range capture.Packages {
		lockKeys[item.Key.Name+"\x00"+item.Key.Version+"\x00"+item.Key.Source] = true
	}
	activeIDs := map[string]bool{}
	for _, node := range metadata.Resolve {
		activeIDs[node.ID] = true
	}
	packageIDs := map[string]MetadataPackage{}
	activePackages := []MetadataPackage{}
	for _, pkg := range metadata.Packages {
		if !activeIDs[pkg.ID] {
			continue
		}
		key := pkg.Name + "\x00" + pkg.Version + "\x00" + pkg.Source
		if !lockKeys[key] && pkg.Source != "" {
			return ActiveGraph{}, fail(CodeGraphIncomplete, "active package is absent from lock superset", map[string]string{"package": pkg.ID})
		}
		if !containedPackagePath(capture, pkg) {
			return ActiveGraph{}, fail(CodeGraphIncomplete, "metadata manifest path is outside captured roots", map[string]string{"path": pkg.ManifestPath})
		}
		if pkg.Links != "" {
			return ActiveGraph{}, fail(CodeNativeLinkUnsupported, "active package declares links", map[string]string{"package": pkg.ID})
		}
		for _, target := range pkg.Targets {
			for _, kind := range target.Kinds {
				if kind == "custom-build" {
					return ActiveGraph{}, fail(CodeBuildScriptUnsupported, "active custom-build target", map[string]string{"package": pkg.ID, "target": target.Name})
				}
				if kind == "proc-macro" {
					return ActiveGraph{}, fail(CodeProcMacroUnsupported, "active proc-macro target", map[string]string{"package": pkg.ID, "target": target.Name})
				}
			}
		}
		packageIDs[pkg.ID] = pkg
		activePackages = append(activePackages, pkg)
	}
	for _, node := range metadata.Resolve {
		if _, ok := packageIDs[node.ID]; !ok {
			return ActiveGraph{}, fail(CodeGraphIncomplete, "resolve node has no package", map[string]string{"package": node.ID})
		}
		expected, ok := selection.ResolvedFeatures[node.ID]
		if !ok {
			return ActiveGraph{}, fail(CodeFeatureProfileMismatch, "resolved feature checkpoint is missing", map[string]string{"package": node.ID})
		}
		expected, _ = sortedUnique(expected)
		if strings.Join(expected, "\x00") != strings.Join(node.Features, "\x00") {
			return ActiveGraph{}, fail(CodeFeatureProfileMismatch, "resolved feature vector differs", map[string]string{"package": node.ID})
		}
		for _, dep := range node.Dependencies {
			if _, ok := packageIDs[dep.ID]; !ok {
				return ActiveGraph{}, fail(CodeGraphIncomplete, "dependency has no active package", map[string]string{"package": dep.ID})
			}
			if dep.Kind == "build" {
				return ActiveGraph{}, fail(CodeBuildScriptUnsupported, "active build dependency", map[string]string{"package": node.ID})
			}
		}
	}
	rootCount := 0
	for _, pkg := range activePackages {
		if pkg.Name == selection.Package && pkg.Source == "" {
			for _, target := range pkg.Targets {
				if target.Name == selection.Binary {
					for _, kind := range target.Kinds {
						if kind == "bin" {
							rootCount++
						}
					}
				}
			}
		}
	}
	if rootCount != 1 {
		return ActiveGraph{}, fail(CodeGraphIncomplete, "exact selected package/bin is absent", nil)
	}
	active := ActiveGraph{SchemaID: "rust-active-graph-v1", CaptureID: capture.Identity, Target: selection.Target, ToolchainID: toolchainID, Package: selection.Package, Binary: selection.Binary, DefaultFeatures: selection.DefaultFeatures, RequestedFeatures: requested, Packages: activePackages, Nodes: metadata.Resolve}
	value := map[string]any{"binary": active.Binary, "capture_id": active.CaptureID, "default_features": active.DefaultFeatures, "package": active.Package, "requested_features": stringsAny(requested), "schema_id": active.SchemaID, "target": active.Target, "toolchain_id": active.ToolchainID, "metadata": metadataValue(Metadata{Packages: activePackages, Resolve: metadata.Resolve})}
	encoded, err := protocoljson.MarshalCanonical(value)
	if err != nil {
		return ActiveGraph{}, err
	}
	active.Identity = "sha256:" + digest(append([]byte("rust-active-graph-v1\x00"), encoded...))
	return active, nil
}

func containedPath(capture CaptureGraph, value string) bool {
	index := sort.SearchStrings(capture.CapturedManifestPaths, value)
	return index < len(capture.CapturedManifestPaths) && capture.CapturedManifestPaths[index] == value
}

func containedPackagePath(capture CaptureGraph, pkg MetadataPackage) bool {
	if containedPath(capture, pkg.ManifestPath) {
		return true
	}
	if pkg.Source == "" {
		return false
	}
	for _, locked := range capture.Packages {
		if locked.Key.Name == pkg.Name && locked.Key.Version == pkg.Version && locked.Key.Source == pkg.Source {
			return pkg.ManifestPath == "vendor/"+packageDirectory(locked.Key)+"/Cargo.toml"
		}
	}
	return false
}
func requiredString(object map[string]any, key string) (string, error) {
	value, ok := object[key].(string)
	if !ok || value == "" {
		return "", fail(CodeGraphIncomplete, "required metadata string missing", map[string]string{"field": key})
	}
	return value, nil
}
func nullableString(object map[string]any, key string) (string, error) {
	value, present := object[key]
	if !present || value == nil {
		return "", nil
	}
	text, ok := value.(string)
	if !ok {
		return "", fail(CodeGraphIncomplete, "metadata nullable string malformed", map[string]string{"field": key})
	}
	return text, nil
}
func allowFields(object map[string]any, context string, allowed []string) error {
	set := map[string]bool{}
	for _, key := range allowed {
		set[key] = true
	}
	for key := range object {
		if !set[key] {
			return fail(CodeGraphIncomplete, "unknown security-relevant metadata field", map[string]string{"context": context, "field": key})
		}
	}
	return nil
}
func validateMetadataPackageDeclarations(object map[string]any) error {
	if raw, present := object["dependencies"]; present {
		items, ok := raw.([]any)
		if !ok {
			return fail(CodeGraphIncomplete, "metadata package dependencies malformed", nil)
		}
		for _, item := range items {
			dependency, ok := item.(map[string]any)
			if !ok {
				return fail(CodeGraphIncomplete, "metadata package dependency malformed", nil)
			}
			if err := allowFields(dependency, "metadata package dependency", []string{"name", "source", "req", "kind", "rename", "optional", "uses_default_features", "features", "target", "registry", "path"}); err != nil {
				return err
			}
			kind, err := nullableString(dependency, "kind")
			if err != nil {
				return err
			}
			if kind != "" && kind != "normal" && kind != "dev" && kind != "build" {
				return fail(CodeGraphIncomplete, "unknown package dependency kind", map[string]string{"kind": kind})
			}
		}
	}
	if raw, present := object["features"]; present {
		features, ok := raw.(map[string]any)
		if !ok {
			return fail(CodeGraphIncomplete, "metadata package features malformed", nil)
		}
		for _, members := range features {
			if _, err := stringSlice(members); err != nil {
				return err
			}
		}
	}
	return nil
}
func stringSlice(value any) ([]string, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fail(CodeGraphIncomplete, "metadata string array malformed", nil)
	}
	result := make([]string, len(items))
	for i, item := range items {
		var ok bool
		result[i], ok = item.(string)
		if !ok {
			return nil, fail(CodeGraphIncomplete, "metadata string array malformed", nil)
		}
	}
	return result, nil
}
func stringsAny(values []string) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}
func packagesValue(values []LockPackage) []any {
	result := make([]any, len(values))
	for i, item := range values {
		deps := make([]any, len(item.Dependencies))
		for j, dep := range item.Dependencies {
			deps[j] = map[string]any{"name": dep.Name, "source": dep.Source, "version": dep.Version}
		}
		result[i] = map[string]any{"checksum": item.Checksum, "dependencies": deps, "kind": string(item.Kind), "name": item.Key.Name, "source": item.Key.Source, "version": item.Key.Version}
	}
	return result
}
func declarationsValue(values []DependencyDeclaration) []any {
	result := make([]any, len(values))
	for i, item := range values {
		result[i] = map[string]any{"branch": item.Branch, "default_features": item.DefaultFeatures, "features": stringsAny(item.Features), "git": item.Git, "kind": item.Kind, "name": item.Name, "optional": item.Optional, "package": item.Package, "path": item.Path, "registry": item.Registry, "rev": item.Rev, "tag": item.Tag, "target": item.Target, "version": item.Version}
	}
	return result
}
func metadataValue(value Metadata) map[string]any {
	payload, _ := json.Marshal(value)
	var result map[string]any
	_ = json.Unmarshal(payload, &result)
	return result
}

// ContainsSelectionFacts reports forbidden selection-binding contamination.
func (graph CaptureGraph) ContainsSelectionFacts() bool {
	encoded, _ := protocoljson.MarshalCanonical(map[string]any{"declarations": declarationsValue(graph.Declarations), "packages": packagesValue(graph.Packages)})
	return strings.Contains(string(encoded), "toolchain_id") || strings.Contains(string(encoded), "requested_features") || strings.Contains(string(encoded), "target_platform")
}
