package pnpmsource

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/closuregraph"
	"github.com/relux-works/curator/internal/protocoljson"
	"gopkg.in/yaml.v3"
)

// Target is the exact selection context used for optional package pruning.
type Target struct {
	OS, Architecture, Libc string
	IncludeDev             bool
}

// ParseRequest supplies the frozen lock and all project declarations that can
// influence pnpm resolution. Paths are project-relative slash paths.
type ParseRequest struct {
	LockBytes              []byte
	Manifests              map[string][]byte
	ConfigFiles            map[string][]byte
	PatchFiles             map[string][]byte
	AllowedRegistryOrigins []string
	Target                 Target
}

// DependencyEdge retains exact importer/snapshot, scope and selection facts.
type DependencyEdge struct {
	From, Name, Specifier, Reference, Scope, To string
	Selected                                    bool
	Reason                                      string
}

// Package is one immutable raw registry artifact from the packages table.
type Package struct {
	Key, Name, Version, Resolved, Integrity string
	PeerDependencies                        map[string]string
	PeerOptional                            map[string]bool
	OS, CPU, Libc                           []string
	HasBin                                  bool
}

// Snapshot is one pnpm package instance. Key includes the exact peer context.
type Snapshot struct {
	Key, PackageKey, Name, Version, PeerContext, PatchHash string
	Dependencies, OptionalDependencies                     map[string]string
	TransitivePeerDependencies                             []string
	Selected                                               bool
	Reachable                                              bool
	PruneReason                                            string
}

// Importer is one root or workspace lock projection.
type Importer struct {
	Path                                                string
	Dependencies, DevDependencies, OptionalDependencies map[string]DependencyReference
}

// DependencyReference binds the manifest specifier and exact lock reference.
type DependencyReference struct{ Specifier, Version string }

// LocalRoot is one importer or file/link dependency captured independently.
type LocalRoot struct{ Path, Name, Version string }

// Patch is one exact declared patch input.
type Patch struct {
	Selector, Path, ManagerHash, SHA256 string
	SnapshotKeys                        []string
}

// Settings are lock-shaping pnpm settings committed by this profile.
type Settings struct {
	AutoInstallPeers         bool
	ExcludeLinksFromLockfile bool
	InjectWorkspacePackages  bool
}

// Graph is the frozen pnpm resolution and selection result before admission.
type Graph struct {
	LockDigest, RawLockSHA256 string
	LockfileVersion           string
	Target                    Target
	Settings                  Settings
	Overrides                 map[string]string
	Packages                  []Package
	Snapshots                 []Snapshot
	Importers                 []Importer
	Edges                     []DependencyEdge
	LocalRoots                []LocalRoot
	Patches                   []Patch
	ConfigurationDigests      map[string]string
	manifestBytes             map[string][]byte
	configBytes               map[string][]byte
	patchBytes                map[string][]byte
	packageIndex              map[string]int
	snapshotIndex             map[string]int
	localIndex                map[string]int
}

type lockWire struct {
	LockfileVersion     any                     `yaml:"lockfileVersion"`
	Settings            settingsWire            `yaml:"settings"`
	Overrides           map[string]string       `yaml:"overrides"`
	PatchedDependencies map[string]patchWire    `yaml:"patchedDependencies"`
	Importers           map[string]importerWire `yaml:"importers"`
	Packages            map[string]packageWire  `yaml:"packages"`
	Snapshots           map[string]snapshotWire `yaml:"snapshots"`
}

type settingsWire struct {
	AutoInstallPeers         *bool `yaml:"autoInstallPeers"`
	ExcludeLinksFromLockfile *bool `yaml:"excludeLinksFromLockfile"`
	InjectWorkspacePackages  *bool `yaml:"injectWorkspacePackages"`
}

type patchWire struct {
	Hash string `yaml:"hash"`
	Path string `yaml:"path"`
}
type resolutionWire struct {
	Integrity string `yaml:"integrity"`
	Tarball   string `yaml:"tarball"`
}
type packageWire struct {
	Resolution           resolutionWire          `yaml:"resolution"`
	PeerDependencies     map[string]string       `yaml:"peerDependencies"`
	PeerDependenciesMeta map[string]peerMetaWire `yaml:"peerDependenciesMeta"`
	Engines              map[string]string       `yaml:"engines"`
	CPU                  []string                `yaml:"cpu"`
	OS                   []string                `yaml:"os"`
	Libc                 []string                `yaml:"libc"`
	HasBin               bool                    `yaml:"hasBin"`
}
type peerMetaWire struct{ Optional bool }
type snapshotWire struct {
	Dependencies               map[string]string `yaml:"dependencies"`
	OptionalDependencies       map[string]string `yaml:"optionalDependencies"`
	TransitivePeerDependencies []string          `yaml:"transitivePeerDependencies"`
}
type importerWire struct {
	Dependencies         map[string]dependencyReferenceWire `yaml:"dependencies"`
	DevDependencies      map[string]dependencyReferenceWire `yaml:"devDependencies"`
	OptionalDependencies map[string]dependencyReferenceWire `yaml:"optionalDependencies"`
}
type dependencyReferenceWire struct {
	Specifier string `yaml:"specifier"`
	Version   string `yaml:"version"`
}

type packageManifest struct {
	Name, Version                                       string
	Dependencies, DevDependencies, OptionalDependencies map[string]string
	PeerDependencies                                    map[string]string
	OS, CPU, Libc                                       []string
	PeerDependenciesMeta                                map[string]struct{ Optional bool }
	Scripts                                             map[string]string
	BundleDependencies, BundledDependencies             json.RawMessage
	Gypfile                                             *bool
	PNPM                                                json.RawMessage `json:"pnpm"`
}

type rootPNPMConfig struct {
	Overrides                map[string]string          `json:"overrides"`
	PatchedDependencies      map[string]string          `json:"patchedDependencies"`
	SupportedArchitectures   supportedArchitecturesWire `json:"supportedArchitectures"`
	PackageExtensions        json.RawMessage            `json:"packageExtensions"`
	OnlyBuiltDependencies    json.RawMessage            `json:"onlyBuiltDependencies"`
	IgnoredBuiltDependencies json.RawMessage            `json:"ignoredBuiltDependencies"`
}
type supportedArchitecturesWire struct{ OS, CPU, Libc []string }

// Parse accepts only pnpm lockfile 9.0 and reconciles every importer, snapshot,
// peer context, override, patch, target selector and configuration input.
func Parse(request ParseRequest) (Graph, error) {
	if len(request.LockBytes) == 0 {
		return Graph{}, fail(CodeLockMissing, "an authoritative pnpm-lock.yaml is required", map[string]string{"expected": "pnpm-lock.yaml"})
	}
	if request.Target.OS == "" || request.Target.Architecture == "" {
		return Graph{}, fail(CodeGraphIncomplete, "pnpm target OS and architecture are required", nil)
	}
	canonicalValue, err := decodeClosedYAML(request.LockBytes)
	if err != nil {
		return Graph{}, fail(CodeLockFormatUnsupported, "pnpm lock is not closed YAML: "+err.Error(), nil)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(request.LockBytes))
	decoder.KnownFields(true)
	var wire lockWire
	if err = decoder.Decode(&wire); err != nil {
		return Graph{}, fail(CodeLockFormatUnsupported, "unsupported pnpm lock schema: "+err.Error(), nil)
	}
	if err = requireYAMLEOF(decoder); err != nil {
		return Graph{}, fail(CodeLockFormatUnsupported, "pnpm lock must contain exactly one YAML document: "+err.Error(), nil)
	}
	version := fmt.Sprint(wire.LockfileVersion)
	if version != "9.0" {
		return Graph{}, fail(CodeLockFormatUnsupported, "only pnpm lockfile 9.0 is supported", map[string]string{"lockfile_version": version})
	}
	if len(wire.Importers) == 0 || len(wire.Packages) == 0 || len(wire.Snapshots) == 0 {
		return Graph{}, fail(CodeGraphIncomplete, "pnpm lock requires importers, packages, and snapshots", nil)
	}
	if _, ok := wire.Importers["."]; !ok {
		return Graph{}, fail(CodeGraphIncomplete, "pnpm lock has no root importer", nil)
	}
	manifests, manifestBytes, err := parsePackageManifests(request.Manifests)
	if err != nil {
		return Graph{}, err
	}
	if _, ok := manifests["package.json"]; !ok {
		return Graph{}, fail(CodeLockStale, "root package.json is missing", map[string]string{"path": "package.json"})
	}
	configBytes, configDigests, workspacePatterns, err := parseConfiguration(request.ConfigFiles, len(wire.Importers) > 1)
	if err != nil {
		return Graph{}, err
	}
	settings := Settings{AutoInstallPeers: boolValue(wire.Settings.AutoInstallPeers, true), ExcludeLinksFromLockfile: boolValue(wire.Settings.ExcludeLinksFromLockfile, false), InjectWorkspacePackages: boolValue(wire.Settings.InjectWorkspacePackages, false)}
	if !settings.AutoInstallPeers {
		return Graph{}, fail(CodeLockFormatUnsupported, "pnpm-source-v1 requires autoInstallPeers=true", map[string]string{"field": "settings.autoInstallPeers"})
	}
	packages, packageIndex, err := parsePackages(wire.Packages, request.AllowedRegistryOrigins)
	if err != nil {
		return Graph{}, err
	}
	snapshots, snapshotIndex, err := parseSnapshots(wire.Snapshots, packageIndex, packages, request.Target)
	if err != nil {
		return Graph{}, err
	}
	importers, localRoots, localIndex, edges, err := parseImporters(wire.Importers, manifests, workspacePatterns, snapshotIndex)
	if err != nil {
		return Graph{}, err
	}
	snapshotEdges, err := buildSnapshotEdges(snapshots, snapshotIndex, localIndex)
	if err != nil {
		return Graph{}, err
	}
	edges = append(edges, snapshotEdges...)
	sort.Slice(edges, func(i, j int) bool { return edgeSortKey(edges[i]) < edgeSortKey(edges[j]) })
	markSelection(importers, snapshots, edges, request.Target)
	patches, patchBytes, err := reconcilePatches(wire.PatchedDependencies, request.PatchFiles, packageIndex, snapshots)
	if err != nil {
		return Graph{}, err
	}
	if err = reconcileRootPNPM(manifests["package.json"], wire.Overrides, patches, request.Target); err != nil {
		return Graph{}, err
	}
	canonical, err := protocoljson.MarshalCanonical(canonicalValue)
	if err != nil {
		return Graph{}, fail(CodeLockFormatUnsupported, "pnpm lock cannot be canonicalized: "+err.Error(), nil)
	}
	lockID, err := closuregraph.IDFromCanonical("pnpm-lock-v9", canonical)
	if err != nil {
		return Graph{}, err
	}
	raw := sha256.Sum256(request.LockBytes)
	return Graph{LockDigest: string(lockID), RawLockSHA256: "sha256:" + hex.EncodeToString(raw[:]), LockfileVersion: version, Target: request.Target, Settings: settings,
		Overrides: cloneMap(wire.Overrides), Packages: packages, Snapshots: snapshots, Importers: importers, Edges: edges, LocalRoots: localRoots, Patches: patches,
		ConfigurationDigests: configDigests, manifestBytes: manifestBytes, configBytes: configBytes, patchBytes: patchBytes,
		packageIndex: packageIndex, snapshotIndex: snapshotIndex, localIndex: localIndex}, nil
}

func decodeClosedYAML(payload []byte) (any, error) {
	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(payload))
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	if len(document.Content) != 1 {
		return nil, fmt.Errorf("exactly one YAML document is required")
	}
	if err := requireYAMLEOF(decoder); err != nil {
		return nil, err
	}
	return yamlNodeValue(document.Content[0])
}

func requireYAMLEOF(decoder *yaml.Decoder) error {
	var trailing yaml.Node
	err := decoder.Decode(&trailing)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("trailing YAML document is unsupported")
}

func yamlNodeValue(node *yaml.Node) (any, error) {
	if node.Anchor != "" || node.Alias != nil || (node.Tag != "" && node.Tag != "!!map" && node.Tag != "!!seq" && node.Tag != "!!str" && node.Tag != "!!bool" && node.Tag != "!!int" && node.Tag != "!!float" && node.Tag != "!!null") {
		return nil, fmt.Errorf("aliases, anchors, and custom tags are unsupported")
	}
	switch node.Kind {
	case yaml.MappingNode:
		result := map[string]any{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" || key.Value == "" {
				return nil, fmt.Errorf("mapping keys must be nonempty strings")
			}
			if _, exists := result[key.Value]; exists {
				return nil, fmt.Errorf("duplicate mapping key %q", key.Value)
			}
			value, err := yamlNodeValue(node.Content[i+1])
			if err != nil {
				return nil, err
			}
			result[key.Value] = value
		}
		return result, nil
	case yaml.SequenceNode:
		result := make([]any, len(node.Content))
		for i, child := range node.Content {
			value, err := yamlNodeValue(child)
			if err != nil {
				return nil, err
			}
			result[i] = value
		}
		return result, nil
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str":
			return node.Value, nil
		case "!!bool":
			return node.Value == "true", nil
		case "!!null":
			return nil, nil
		case "!!int":
			var value int64
			if err := node.Decode(&value); err != nil {
				return nil, err
			}
			return value, nil
		case "!!float":
			var value float64
			if err := node.Decode(&value); err != nil {
				return nil, err
			}
			return value, nil
		}
	}
	return nil, fmt.Errorf("unsupported YAML node kind")
}

func parsePackages(values map[string]packageWire, allowed []string) ([]Package, map[string]int, error) {
	keys := sortedKeys(values)
	packages := make([]Package, 0, len(keys))
	index := make(map[string]int, len(keys))
	for _, key := range keys {
		name, version, err := splitPackageKey(key)
		if err != nil {
			return nil, nil, err
		}
		wire := values[key]
		if wire.Resolution.Integrity == "" {
			return nil, nil, fail(CodeIntegrityMissing, "pnpm package has no registry integrity", map[string]string{"package": key})
		}
		resolved := wire.Resolution.Tarball
		if resolved == "" {
			resolved = registryTarballLocator(name, version, allowed)
		}
		if err = validateRegistryLocator(resolved, allowed); err != nil {
			return nil, nil, err
		}
		pkg := Package{Key: key, Name: name, Version: version, Resolved: resolved, Integrity: wire.Resolution.Integrity, PeerDependencies: cloneMap(wire.PeerDependencies), PeerOptional: peerOptional(wire.PeerDependenciesMeta), OS: sortedCopy(wire.OS), CPU: sortedCopy(wire.CPU), Libc: sortedCopy(wire.Libc), HasBin: wire.HasBin}
		index[key] = len(packages)
		packages = append(packages, pkg)
	}
	return packages, index, nil
}

func parseSnapshots(values map[string]snapshotWire, packageIndex map[string]int, packages []Package, target Target) ([]Snapshot, map[string]int, error) {
	keys := sortedKeys(values)
	result := make([]Snapshot, 0, len(keys))
	index := make(map[string]int, len(keys))
	for _, key := range keys {
		base, peers, patchHash, err := splitSnapshotKey(key)
		if err != nil {
			return nil, nil, err
		}
		pkgIndex, ok := packageIndex[base]
		if !ok {
			return nil, nil, fail(CodeGraphIncomplete, "snapshot has no packages-table artifact", map[string]string{"snapshot": key, "package": base})
		}
		pkg := packages[pkgIndex]
		selected, reason := matchesTarget(pkg, target)
		wire := values[key]
		result = append(result, Snapshot{Key: key, PackageKey: base, Name: pkg.Name, Version: pkg.Version, PeerContext: peers, PatchHash: patchHash, Dependencies: cloneMap(wire.Dependencies), OptionalDependencies: cloneMap(wire.OptionalDependencies), TransitivePeerDependencies: sortedCopy(wire.TransitivePeerDependencies), Selected: selected, PruneReason: reason})
		index[key] = len(result) - 1
	}
	return result, index, nil
}

func parseImporters(values map[string]importerWire, manifests map[string]packageManifest, patterns []string, snapshots map[string]int) ([]Importer, []LocalRoot, map[string]int, []DependencyEdge, error) {
	keys := sortedKeys(values)
	result := make([]Importer, 0, len(keys))
	locals := make([]LocalRoot, 0, len(keys))
	localIndex := map[string]int{}
	edges := []DependencyEdge{}
	// Establish every local identity before resolving importer edges so a root
	// may point at a workspace whose importer sorts later.
	for _, importerPath := range keys {
		manifestPath := "package.json"
		if importerPath != "." {
			manifestPath = path.Join(importerPath, "package.json")
		}
		manifest, ok := manifests[manifestPath]
		if !ok {
			return nil, nil, nil, nil, fail(CodeLockStale, "pnpm importer has no captured manifest", map[string]string{"importer": importerPath})
		}
		localIndex[importerPath] = len(locals)
		locals = append(locals, LocalRoot{Path: importerPath, Name: manifest.Name, Version: manifest.Version})
	}
	for _, importerPath := range keys {
		if err := validateProjectPath(importerPath, true); err != nil {
			return nil, nil, nil, nil, err
		}
		manifestPath := "package.json"
		if importerPath != "." {
			manifestPath = path.Join(importerPath, "package.json")
		}
		manifest, ok := manifests[manifestPath]
		if !ok {
			return nil, nil, nil, nil, fail(CodeLockStale, "pnpm importer has no captured manifest", map[string]string{"importer": importerPath})
		}
		if importerPath != "." && !matchesAnyWorkspace(patterns, importerPath) {
			return nil, nil, nil, nil, fail(CodeLockStale, "pnpm importer is outside pnpm-workspace.yaml", map[string]string{"importer": importerPath})
		}
		wire := values[importerPath]
		if err := reconcileImporter(manifest, wire, importerPath); err != nil {
			return nil, nil, nil, nil, err
		}
		importer := Importer{Path: importerPath, Dependencies: convertRefs(wire.Dependencies), DevDependencies: convertRefs(wire.DevDependencies), OptionalDependencies: convertRefs(wire.OptionalDependencies)}
		result = append(result, importer)
		from := localRootKey(importerPath)
		for _, set := range []struct {
			scope string
			refs  map[string]dependencyReferenceWire
		}{{"runtime", wire.Dependencies}, {"development", wire.DevDependencies}, {"optional", wire.OptionalDependencies}} {
			for _, name := range sortedKeys(set.refs) {
				ref := set.refs[name]
				to, err := resolveReference(name, ref.Version, snapshots, localIndex, importerPath)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				edges = append(edges, DependencyEdge{From: from, Name: name, Specifier: ref.Specifier, Reference: ref.Version, Scope: set.scope, To: to})
			}
		}
	}
	for manifestPath := range manifests {
		dir := path.Dir(manifestPath)
		if manifestPath == "package.json" {
			dir = "."
		}
		if _, ok := localIndex[dir]; !ok {
			return nil, nil, nil, nil, fail(CodeLockStale, "captured manifest is absent from pnpm importers", map[string]string{"path": manifestPath})
		}
	}
	return result, locals, localIndex, edges, nil
}

func buildSnapshotEdges(snapshots []Snapshot, index, locals map[string]int) ([]DependencyEdge, error) {
	edges := []DependencyEdge{}
	for _, snapshot := range snapshots {
		for _, set := range []struct {
			scope string
			refs  map[string]string
		}{{"runtime", snapshot.Dependencies}, {"optional", snapshot.OptionalDependencies}} {
			for _, name := range sortedKeys(set.refs) {
				ref := set.refs[name]
				to, err := resolveReference(name, ref, index, locals, ".")
				if err != nil {
					if set.scope == "optional" {
						edges = append(edges, DependencyEdge{From: snapshot.Key, Name: name, Reference: ref, Scope: set.scope, Reason: "optional_missing"})
						continue
					}
					return nil, err
				}
				edges = append(edges, DependencyEdge{From: snapshot.Key, Name: name, Reference: ref, Scope: set.scope, To: to})
			}
		}
		for _, peer := range parsePeerContext(snapshot.PeerContext) {
			to, err := resolveReference(peer.name, peer.reference, index, locals, ".")
			if err != nil {
				return nil, fail(CodeGraphIncomplete, "peer context cannot be resolved", map[string]string{"snapshot": snapshot.Key, "peer": peer.name, "reference": peer.reference})
			}
			edges = append(edges, DependencyEdge{From: snapshot.Key, Name: peer.name, Reference: peer.reference, Scope: "peer", To: to})
		}
	}
	return edges, nil
}

func markSelection(importers []Importer, snapshots []Snapshot, edges []DependencyEdge, target Target) {
	markGraphReachability(importers, snapshots, edges)
	selectedSnapshots := map[string]bool{}
	selectedLocals := map[string]bool{}
	queue := []string{localRootKey(".")}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if strings.HasPrefix(current, "local:") {
			if selectedLocals[current] {
				continue
			}
			selectedLocals[current] = true
		}
		if !strings.HasPrefix(current, "local:") {
			if selectedSnapshots[current] {
				continue
			}
			selectedSnapshots[current] = true
		}
		for i := range edges {
			if edges[i].From != current {
				continue
			}
			if edges[i].Scope == "development" && !target.IncludeDev {
				edges[i].Reason = "development_excluded"
				continue
			}
			if edges[i].To == "" {
				continue
			}
			if index := snapshotPosition(snapshots, edges[i].To); index >= 0 && !snapshots[index].Selected {
				edges[i].Reason = snapshots[index].PruneReason
				continue
			}
			edges[i].Selected = true
			queue = append(queue, edges[i].To)
		}
	}
	for i := range snapshots {
		snapshots[i].Selected = snapshots[i].Selected && selectedSnapshots[snapshots[i].Key]
		if !snapshots[i].Selected && snapshots[i].PruneReason == "" {
			snapshots[i].PruneReason = "unreachable"
		}
	}
}

// markGraphReachability records whether a snapshot belongs to the complete
// lock graph rooted at any importer. It deliberately ignores target and active
// selection: pnpm can materialize a target-pruned snapshot that remains
// lock-reachable, while omitting a wholly unreachable snapshot even when a
// target selector already supplied its visible prune reason.
func markGraphReachability(importers []Importer, snapshots []Snapshot, edges []DependencyEdge) {
	reachable := make(map[string]bool, len(importers)+len(snapshots))
	queue := make([]string, 0, len(importers))
	for _, importer := range importers {
		queue = append(queue, localRootKey(importer.Path))
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if reachable[current] {
			continue
		}
		reachable[current] = true
		for _, edge := range edges {
			if edge.From == current && edge.To != "" {
				queue = append(queue, edge.To)
			}
		}
	}
	for i := range snapshots {
		snapshots[i].Reachable = reachable[snapshots[i].Key]
	}
}

func localRootKey(importerPath string) string { return "local:" + importerPath }

func reconcilePatches(declared map[string]patchWire, supplied map[string][]byte, packages map[string]int, snapshots []Snapshot) ([]Patch, map[string][]byte, error) {
	if len(declared) != len(supplied) {
		return nil, nil, fail(CodeManagerPluginUndeclared, "declared and supplied pnpm patches are not bijective", map[string]string{"declared": fmt.Sprint(len(declared)), "supplied": fmt.Sprint(len(supplied))})
	}
	patches := []Patch{}
	bytesByPath := map[string][]byte{}
	for _, selector := range sortedKeys(declared) {
		item := declared[selector]
		if _, _, err := splitPackageKey(selector); err != nil {
			return nil, nil, fail(CodeLockFormatUnsupported, "patch selector is not an exact package identity", map[string]string{"selector": selector})
		}
		if _, ok := packages[selector]; !ok {
			return nil, nil, fail(CodeGraphIncomplete, "patch selector has no package", map[string]string{"selector": selector})
		}
		if err := validateProjectPath(item.Path, false); err != nil {
			return nil, nil, err
		}
		if item.Hash == "" {
			return nil, nil, fail(CodeIntegrityMissing, "pnpm patch has no manager hash", map[string]string{"selector": selector})
		}
		payload, ok := supplied[item.Path]
		if !ok {
			return nil, nil, fail(CodeOfflineInputMissing, "declared pnpm patch is absent", map[string]string{"path": item.Path})
		}
		managerHash, err := managerPatchHash(payload)
		if err != nil {
			return nil, nil, fail(CodeIntegrityMismatch, "pnpm patch cannot be hashed by the pinned manager profile: "+err.Error(), map[string]string{"selector": selector})
		}
		if item.Hash != managerHash {
			return nil, nil, fail(CodeIntegrityMismatch, "pnpm manager patch hash differs from admitted patch content", map[string]string{"selector": selector, "expected": item.Hash, "observed": managerHash})
		}
		digest := sha256.Sum256(payload)
		matchedSnapshots := []string{}
		for _, snapshot := range snapshots {
			if snapshot.PackageKey != selector {
				continue
			}
			if snapshot.PatchHash != managerHash {
				return nil, nil, fail(CodeLockStale, "patched package snapshot does not bind the exact manager patch hash", map[string]string{"selector": selector, "snapshot": snapshot.Key, "expected": managerHash, "observed": snapshot.PatchHash})
			}
			matchedSnapshots = append(matchedSnapshots, snapshot.Key)
		}
		if len(matchedSnapshots) == 0 {
			return nil, nil, fail(CodeGraphIncomplete, "declared pnpm patch has no exact snapshot context", map[string]string{"selector": selector})
		}
		patches = append(patches, Patch{Selector: selector, Path: item.Path, ManagerHash: item.Hash, SHA256: "sha256:" + hex.EncodeToString(digest[:]), SnapshotKeys: matchedSnapshots})
		bytesByPath[item.Path] = append([]byte(nil), payload...)
	}
	for name := range supplied {
		if _, ok := bytesByPath[name]; !ok {
			return nil, nil, fail(CodeManagerPluginUndeclared, "supplied pnpm patch is undeclared", map[string]string{"path": name})
		}
	}
	for _, snapshot := range snapshots {
		if snapshot.PatchHash == "" {
			continue
		}
		item, ok := declared[snapshot.PackageKey]
		if !ok || item.Hash != snapshot.PatchHash {
			return nil, nil, fail(CodeManagerPluginUndeclared, "snapshot contains an undeclared pnpm patch context", map[string]string{"snapshot": snapshot.Key, "patch_hash": snapshot.PatchHash})
		}
	}
	return patches, bytesByPath, nil
}

func managerPatchHash(payload []byte) (string, error) {
	if !utf8.Valid(payload) {
		return "", fmt.Errorf("patch is not valid UTF-8")
	}
	normalized := bytes.ReplaceAll(payload, []byte("\r\n"), []byte("\n"))
	digest := sha256.Sum256(normalized)
	return hex.EncodeToString(digest[:]), nil
}

func reconcileRootPNPM(manifest packageManifest, overrides map[string]string, patches []Patch, target Target) error {
	if len(bytes.TrimSpace(manifest.PNPM)) == 0 {
		if len(overrides)+len(patches) == 0 {
			return nil
		}
		return fail(CodeLockStale, "lock contains pnpm policy absent from root manifest", nil)
	}
	var config rootPNPMConfig
	decoder := json.NewDecoder(bytes.NewReader(manifest.PNPM))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return fail(CodeManagerPluginUndeclared, "unsupported root package.json pnpm configuration: "+err.Error(), nil)
	}
	if rawPresent(config.PackageExtensions) {
		return fail(CodeManagerPluginUndeclared, "pnpm packageExtensions are unsupported", map[string]string{"field": "pnpm.packageExtensions"})
	}
	if rawPresent(config.OnlyBuiltDependencies) || rawPresent(config.IgnoredBuiltDependencies) {
		return fail(CodeHookUndeclared, "pnpm build policy extensions are unsupported", nil)
	}
	if !equalMap(config.Overrides, overrides) {
		return fail(CodeLockStale, "pnpm overrides differ between manifest and lock", nil)
	}
	patchMap := map[string]string{}
	for _, patch := range patches {
		patchMap[patch.Selector] = patch.Path
	}
	if !equalMap(config.PatchedDependencies, patchMap) {
		return fail(CodeLockStale, "pnpm patches differ between manifest and lock", nil)
	}
	for _, item := range []struct {
		name, actual string
		allowed      []string
	}{{"os", target.OS, config.SupportedArchitectures.OS}, {"cpu", target.Architecture, config.SupportedArchitectures.CPU}, {"libc", target.Libc, config.SupportedArchitectures.Libc}} {
		if len(item.allowed) > 0 && !selectorMatches(item.allowed, item.actual) {
			return fail(CodeLockStale, "target is outside pnpm supportedArchitectures", map[string]string{"field": item.name, "target": item.actual})
		}
	}
	return nil
}

func parseConfiguration(files map[string][]byte, workspaces bool) (map[string][]byte, map[string]string, []string, error) {
	result, digests := map[string][]byte{}, map[string]string{}
	patterns := []string{}
	for _, name := range sortedKeys(files) {
		payload := files[name]
		if strings.EqualFold(path.Base(name), ".pnpmfile.cjs") || strings.EqualFold(path.Base(name), ".pnpmfile.mjs") {
			return nil, nil, nil, fail(CodeManagerPluginUndeclared, "pnpm hook file is unsupported", map[string]string{"path": name})
		}
		if name != ".npmrc" && name != "pnpm-workspace.yaml" {
			return nil, nil, nil, fail(CodeManagerPluginUndeclared, "unsupported pnpm configuration file", map[string]string{"path": name})
		}
		if name == ".npmrc" {
			if err := validateNPMRC(payload); err != nil {
				return nil, nil, nil, err
			}
		}
		if name == "pnpm-workspace.yaml" {
			var wire struct {
				Packages []string `yaml:"packages"`
			}
			dec := yaml.NewDecoder(bytes.NewReader(payload))
			dec.KnownFields(true)
			if err := dec.Decode(&wire); err != nil {
				return nil, nil, nil, fail(CodeLockFormatUnsupported, "unsupported pnpm-workspace.yaml: "+err.Error(), nil)
			}
			if err := requireYAMLEOF(dec); err != nil {
				return nil, nil, nil, fail(CodeLockFormatUnsupported, "pnpm-workspace.yaml must contain exactly one YAML document: "+err.Error(), nil)
			}
			for _, p := range wire.Packages {
				if err := validateWorkspacePattern(p); err != nil {
					return nil, nil, nil, err
				}
			}
			patterns = sortedCopy(wire.Packages)
		}
		sum := sha256.Sum256(payload)
		result[name] = append([]byte(nil), payload...)
		digests[name] = "sha256:" + hex.EncodeToString(sum[:])
	}
	if workspaces && len(patterns) == 0 {
		return nil, nil, nil, fail(CodeLockStale, "multiple pnpm importers require pnpm-workspace.yaml", nil)
	}
	return result, digests, patterns, nil
}

func validateNPMRC(payload []byte) error {
	allowed := map[string]bool{"auto-install-peers": true, "dedupe-peer-dependents": true, "ignore-scripts": true, "inject-workspace-packages": true, "link-workspace-packages": true, "prefer-workspace-packages": true, "resolve-peers-from-workspace-root": true, "shared-workspace-lockfile": true, "side-effects-cache": true, "strict-peer-dependencies": true, "verify-store-integrity": true}
	for number, line := range strings.Split(string(payload), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fail(CodeLockFormatUnsupported, "invalid .npmrc line", map[string]string{"line": fmt.Sprint(number + 1)})
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if !allowed[key] {
			return fail(CodeManagerPluginUndeclared, "unsupported pnpm configuration key", map[string]string{"key": key})
		}
		if key == "side-effects-cache" && value != "false" {
			return fail(CodeHookUndeclared, "pnpm side-effects cache must be disabled", map[string]string{"key": key})
		}
		if key == "ignore-scripts" && value != "true" {
			return fail(CodeHookUndeclared, "pnpm lifecycle scripts must be disabled", map[string]string{"key": key})
		}
	}
	return nil
}

func parsePackageManifests(values map[string][]byte) (map[string]packageManifest, map[string][]byte, error) {
	result, raw := map[string]packageManifest{}, map[string][]byte{}
	for _, name := range sortedKeys(values) {
		if name != "package.json" && !strings.HasSuffix(name, "/package.json") {
			return nil, nil, fail(CodeLocalPathEscape, "manifest path escapes the project", map[string]string{"path": name})
		}
		if err := validateProjectPath(strings.TrimSuffix(name, "/package.json"), name == "package.json"); err != nil {
			return nil, nil, err
		}
		var manifest packageManifest
		if err := json.Unmarshal(values[name], &manifest); err != nil {
			return nil, nil, fail(CodeLockStale, "invalid package manifest: "+err.Error(), map[string]string{"path": name})
		}
		if manifest.Name == "" || manifest.Version == "" {
			return nil, nil, fail(CodeLockStale, "package manifest identity is incomplete", map[string]string{"path": name})
		}
		result[name] = manifest
		raw[name] = append([]byte(nil), values[name]...)
	}
	return result, raw, nil
}

func reconcileImporter(manifest packageManifest, wire importerWire, importer string) error {
	for _, item := range []struct {
		name     string
		manifest map[string]string
		lock     map[string]dependencyReferenceWire
	}{{"dependencies", manifest.Dependencies, wire.Dependencies}, {"devDependencies", manifest.DevDependencies, wire.DevDependencies}, {"optionalDependencies", manifest.OptionalDependencies, wire.OptionalDependencies}} {
		if len(item.manifest) != len(item.lock) {
			return fail(CodeLockStale, "importer dependency count differs from manifest", map[string]string{"importer": importer, "field": item.name})
		}
		for name, spec := range item.manifest {
			ref, ok := item.lock[name]
			if !ok || ref.Specifier != spec || ref.Version == "" {
				return fail(CodeLockStale, "importer dependency differs from manifest", map[string]string{"importer": importer, "field": item.name, "dependency": name})
			}
		}
	}
	return nil
}

func resolveReference(name, reference string, snapshots, locals map[string]int, importer string) (string, error) {
	if strings.HasPrefix(reference, "link:") || strings.HasPrefix(reference, "workspace:") {
		value := strings.TrimPrefix(strings.TrimPrefix(reference, "link:"), "workspace:")
		if strings.HasPrefix(reference, "workspace:") {
			return "", fail(CodeLockStale, "workspace protocol must be resolved to an exact link in pnpm lock", map[string]string{"dependency": name})
		}
		resolved := path.Clean(path.Join(importer, value))
		if importer == "." {
			resolved = path.Clean(value)
		}
		if err := validateProjectPath(resolved, false); err != nil {
			return "", err
		}
		if _, ok := locals[resolved]; !ok {
			return "", fail(CodeGraphIncomplete, "local dependency has no captured importer", map[string]string{"dependency": name, "path": resolved})
		}
		return "local:" + resolved, nil
	}
	if strings.HasPrefix(reference, "file:") {
		value := strings.TrimPrefix(reference, "file:")
		if err := validateProjectPath(value, false); err != nil {
			return "", err
		}
		if _, ok := locals[value]; !ok {
			return "", fail(CodeOfflineInputMissing, "file dependency root is not independently captured", map[string]string{"dependency": name, "path": value})
		}
		return "local:" + value, nil
	}
	if strings.Contains(reference, ":") {
		return "", fail(CodeOriginUnpinned, "unsupported pnpm dependency locator", map[string]string{"dependency": name, "scheme": strings.SplitN(reference, ":", 2)[0]})
	}
	candidate := name + "@" + reference
	if _, ok := snapshots[candidate]; !ok {
		return "", fail(CodeGraphIncomplete, "pnpm dependency reference has no exact snapshot", map[string]string{"dependency": name, "reference": reference, "snapshot": candidate})
	}
	return candidate, nil
}

type peerReference struct{ name, reference string }

func parsePeerContext(value string) []peerReference {
	result := []peerReference{}
	for len(value) > 0 {
		start := strings.Index(value, "(")
		if start < 0 {
			break
		}
		depth, end := 0, -1
		for i := start; i < len(value); i++ {
			if value[i] == '(' {
				depth++
			}
			if value[i] == ')' {
				depth--
				if depth == 0 {
					end = i
					break
				}
			}
		}
		if end < 0 {
			break
		}
		token := value[start+1 : end]
		at := strings.LastIndex(token, "@")
		if at > 0 {
			result = append(result, peerReference{name: token[:at], reference: token[at+1:]})
		}
		value = value[end+1:]
	}
	return result
}
func splitSnapshotKey(key string) (string, string, string, error) {
	pos := strings.Index(key, "(")
	base, suffix := key, ""
	if pos >= 0 {
		base, suffix = key[:pos], key[pos:]
		depth := 0
		for _, r := range suffix {
			if r == '(' {
				depth++
			}
			if r == ')' {
				depth--
				if depth < 0 {
					return "", "", "", fail(CodeLockFormatUnsupported, "invalid pnpm peer context", map[string]string{"snapshot": key})
				}
			}
		}
		if depth != 0 {
			return "", "", "", fail(CodeLockFormatUnsupported, "invalid pnpm peer context", map[string]string{"snapshot": key})
		}
	}
	if _, _, err := splitPackageKey(base); err != nil {
		return "", "", "", err
	}
	peers := strings.Builder{}
	patchHash := ""
	peerNames := map[string]bool{}
	for len(suffix) > 0 {
		end := matchingContextEnd(suffix)
		if end < 0 {
			return "", "", "", fail(CodeLockFormatUnsupported, "invalid pnpm snapshot context", map[string]string{"snapshot": key})
		}
		group := suffix[:end+1]
		token := group[1 : len(group)-1]
		if strings.HasPrefix(token, "patch_hash=") {
			if patchHash != "" || len(token) != len("patch_hash=")+64 {
				return "", "", "", fail(CodeLockFormatUnsupported, "invalid pnpm patch context", map[string]string{"snapshot": key})
			}
			patchHash = strings.TrimPrefix(token, "patch_hash=")
			if _, err := hex.DecodeString(patchHash); err != nil {
				return "", "", "", fail(CodeLockFormatUnsupported, "invalid pnpm patch hash encoding", map[string]string{"snapshot": key})
			}
		} else {
			if strings.ContainsAny(token, "()") {
				return "", "", "", fail(CodeLockFormatUnsupported, "nested pnpm peer contexts are unsupported by the pinned profile", map[string]string{"snapshot": key})
			}
			at := strings.LastIndex(token, "@")
			if at <= 0 || !validPackageName(token[:at]) || !validPackageVersion(token[at+1:]) || peerNames[token[:at]] {
				return "", "", "", fail(CodeLockFormatUnsupported, "invalid or duplicate pnpm peer context", map[string]string{"snapshot": key})
			}
			peerNames[token[:at]] = true
			peers.WriteString(group)
		}
		suffix = suffix[end+1:]
	}
	return base, peers.String(), patchHash, nil
}

func matchingContextEnd(value string) int {
	if value == "" || value[0] != '(' {
		return -1
	}
	depth := 0
	for i := 0; i < len(value); i++ {
		switch value[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}
func splitPackageKey(key string) (string, string, error) {
	at := strings.LastIndex(key, "@")
	if at <= 0 || at == len(key)-1 {
		return "", "", fail(CodeLockFormatUnsupported, "invalid pnpm package key", map[string]string{"package": key})
	}
	name, version := key[:at], key[at+1:]
	if !validPackageName(name) || !validPackageVersion(version) {
		return "", "", fail(CodeLockFormatUnsupported, "unsafe or unsupported pnpm package key", map[string]string{"package": key})
	}
	return name, version, nil
}

func validPackageName(value string) bool {
	if value == "" || strings.ContainsAny(value, "\\\x00\r\n") || strings.Contains(value, "..") {
		return false
	}
	if strings.HasPrefix(value, "@") {
		parts := strings.Split(value, "/")
		return len(parts) == 2 && len(parts[0]) > 1 && parts[1] != ""
	}
	return !strings.Contains(value, "/") && value != "."
}

func validPackageVersion(value string) bool {
	return value != "" && !strings.ContainsAny(value, "/\\():\x00\r\n") && value != "." && value != ".."
}
func registryTarballLocator(name, version string, allowed []string) string {
	if len(allowed) == 0 {
		return ""
	}
	base := strings.TrimSuffix(allowed[0], "/")
	escaped := strings.ReplaceAll(name, "/", "%2f")
	leaf := name
	if slash := strings.LastIndex(name, "/"); slash >= 0 {
		leaf = name[slash+1:]
	}
	return base + "/" + escaped + "/-/" + leaf + "-" + version + ".tgz"
}
func validateRegistryLocator(value string, allowed []string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return fail(CodeOriginUnpinned, "pnpm tarball locator is not immutable HTTPS registry input", map[string]string{"locator": value})
	}
	for _, origin := range allowed {
		o, err := url.Parse(origin)
		if err == nil && strings.EqualFold(o.Scheme, parsed.Scheme) && strings.EqualFold(o.Host, parsed.Host) {
			return nil
		}
	}
	return fail(CodeOriginUnpinned, "pnpm tarball registry origin is not admitted", map[string]string{"locator": value})
}
func matchesTarget(pkg Package, target Target) (bool, string) {
	if !selectorMatches(pkg.OS, target.OS) {
		return false, "os_mismatch"
	}
	if !selectorMatches(pkg.CPU, target.Architecture) {
		return false, "cpu_mismatch"
	}
	if !selectorMatches(pkg.Libc, target.Libc) {
		return false, "libc_mismatch"
	}
	return true, ""
}
func selectorMatches(rules []string, value string) bool {
	if len(rules) == 0 {
		return true
	}
	positive := false
	for _, rule := range rules {
		if strings.HasPrefix(rule, "!") {
			if strings.TrimPrefix(rule, "!") == value {
				return false
			}
			continue
		}
		positive = true
		if rule == value {
			return true
		}
	}
	return !positive
}
func validateProjectPath(value string, root bool) error {
	if root && value == "." {
		return nil
	}
	if value == "" || value == "." || path.IsAbs(value) || path.Clean(value) != value || value == ".." || strings.HasPrefix(value, "../") || strings.Contains(value, "\\") {
		return fail(CodeLocalPathEscape, "pnpm path escapes project", map[string]string{"path": value})
	}
	return nil
}
func validateWorkspacePattern(value string) error {
	if strings.HasSuffix(value, "/*") {
		return validateProjectPath(strings.TrimSuffix(value, "/*"), false)
	}
	return validateProjectPath(value, false)
}
func matchesAnyWorkspace(patterns []string, value string) bool {
	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(value, prefix) && !strings.Contains(strings.TrimPrefix(value, prefix), "/") {
				return true
			}
		} else if pattern == value {
			return true
		}
	}
	return false
}
func snapshotPosition(values []Snapshot, key string) int {
	for i := range values {
		if values[i].Key == key {
			return i
		}
	}
	return -1
}
func edgeSortKey(value DependencyEdge) string {
	return value.From + "\x00" + value.Scope + "\x00" + value.Name + "\x00" + value.To
}
func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
func sortedKeys[T any](values map[string]T) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
func cloneMap(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	for k, v := range values {
		result[k] = v
	}
	return result
}
func equalMap(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
func peerOptional(values map[string]peerMetaWire) map[string]bool {
	result := map[string]bool{}
	for k, v := range values {
		if v.Optional {
			result[k] = true
		}
	}
	return result
}
func convertRefs(values map[string]dependencyReferenceWire) map[string]DependencyReference {
	result := map[string]DependencyReference{}
	for k, v := range values {
		result[k] = DependencyReference(v)
	}
	return result
}
func rawPresent(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) > 0 && !bytes.Equal(trimmed, []byte("null")) && !bytes.Equal(trimmed, []byte("{}")) && !bytes.Equal(trimmed, []byte("[]"))
}
